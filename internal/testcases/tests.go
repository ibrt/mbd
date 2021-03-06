package testcases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/ibrt/mbd"
	"github.com/ibrt/mbd/internal/testcontext"
	"github.com/stretchr/testify/require"
)

// TestRequest is a request for test functions.
type TestRequest struct {
	Value string `json:"value" schema:"value"`
}

// TestResponse is a response for test functions.
type TestResponse struct {
	Value string `json:"value"`
}

// TestCase describes a test case.
type TestCase struct {
	Name          string
	DisableDebug  bool
	ReqTemplate   interface{}
	FormReqParser bool
	RespTemplate  interface{}
	Request       interface{}
	Handler       mbd.Handler
	Providers     []mbd.Provider
	Checkers      []mbd.Checker
	Assertion     func(require.TestingT, int, map[string][]string, interface{})
}

// GetTestCase returns the TestCase with the given name, panics if not found.
func GetTestCase(name string) *TestCase {
	for _, c := range testCases {
		if c.Name == name {
			return c
		}
	}
	panic("unknown test case")
}

// GetTestCases returns a list of test cases, which can be run locally in-memory, or against a test remote Lambda
// deployment.
func GetTestCases() []*TestCase {
	return testCases
}

var testCases = []*TestCase{
	{
		Name:          "HappyPath",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  TestResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			t := testcontext.GetTestingT(ctx)

			require.True(t, mbd.GetDebug(ctx))
			require.Equal(t, "/HappyPath", mbd.GetPath(ctx).Resource)
			require.Equal(t, "/HappyPath", mbd.GetPath(ctx).Path)
			require.Equal(t, "POST", mbd.GetPath(ctx).Method)
			require.Equal(t, "application/json; charset=utf-8", mbd.GetHeaders(ctx).Get("Content-Type"))
			require.Equal(t, []string{"application/json; charset=utf-8"}, mbd.GetHeaders(ctx).GetMulti("Content-Type"))
			require.Empty(t, mbd.GetQueryString(ctx).Map())
			require.Empty(t, mbd.GetQueryString(ctx).MapMulti())
			require.Empty(t, mbd.GetPathParameters(ctx).Map())
			require.Empty(t, mbd.GetStageVariables(ctx).Map())
			require.NotEmpty(t, mbd.GetRequestContext(ctx).RequestID)

			return &TestResponse{
				Value: req.(*TestRequest).Value,
			}, nil
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			response := resp.(*TestResponse)

			require.Equal(t, http.StatusOK, statusCode)
			require.Equal(t, &TestResponse{Value: "testValue"}, response)
		},
	},
	{
		Name:          "MissingBody",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusBadRequest, statusCode)
			require.Equal(t, http.StatusBadRequest, errorResponse.StatusCode)
			require.Equal(t, "invalid-body", errorResponse.PublicMessage)
			require.Len(t, errorResponse.Errors, 1)
			require.Equal(t, "invalid Body: EOF", errorResponse.Errors[0].Error)
			require.NotEmpty(t, errorResponse.Errors[0].StackTrace)
		},
	},
	{
		Name: "EmptyRequestAndResponse",
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			t := testcontext.GetTestingT(ctx)

			require.Nil(t, req)
			return nil, nil
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			require.Equal(t, http.StatusOK, statusCode)
			require.Nil(t, resp)
		},
	},
	{
		Name:          "Error",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, errors.Errorf("test error",
				errors.HTTPStatusConflict,
				errors.PublicMessage("test-error"))
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusConflict, statusCode)
			require.Equal(t, http.StatusConflict, errorResponse.StatusCode)
			require.Equal(t, "test-error", errorResponse.PublicMessage)
			require.Len(t, errorResponse.Errors, 1)
			require.Equal(t, "test error", errorResponse.Errors[0].Error)
			require.NotEmpty(t, errorResponse.Errors[0].StackTrace)
		},
	},
	{
		Name:          "DefaultError",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("test error")
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusInternalServerError, statusCode)
			require.Equal(t, http.StatusInternalServerError, errorResponse.StatusCode)
			require.Equal(t, "internal-server-error", errorResponse.PublicMessage)
			require.Len(t, errorResponse.Errors, 1)
			require.Equal(t, "test error", errorResponse.Errors[0].Error)
			require.NotEmpty(t, errorResponse.Errors[0].StackTrace)
		},
	},
	{
		Name:          "NoDebugError",
		DisableDebug:  true,
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("test error")
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusInternalServerError, statusCode)
			require.Equal(t, http.StatusInternalServerError, errorResponse.StatusCode)
			require.Equal(t, "internal-server-error", errorResponse.PublicMessage)
			require.Empty(t, errorResponse.Errors)
		},
	},
	{
		Name:          "Panic",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			panic(errors.Errorf("test error",
				errors.HTTPStatusConflict,
				errors.PublicMessage("test-error")))
		},
		Providers: nil,
		Checkers:  nil,
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusConflict, statusCode)
			require.Equal(t, http.StatusConflict, errorResponse.StatusCode)
			require.Equal(t, "test-error", errorResponse.PublicMessage)
			require.Len(t, errorResponse.Errors, 1)
			require.Equal(t, "test error", errorResponse.Errors[0].Error)
			require.NotEmpty(t, errorResponse.Errors[0].StackTrace)
		},
	},
	{
		Name:          "DefaultPanic",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("test error")
		},
		Providers: nil,
		Checkers:  nil,
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusInternalServerError, statusCode)
			require.Equal(t, http.StatusInternalServerError, errorResponse.StatusCode)
			require.Equal(t, "internal-server-error", errorResponse.PublicMessage)
			require.Len(t, errorResponse.Errors, 1)
			require.Equal(t, "test error", errorResponse.Errors[0].Error)
			require.NotEmpty(t, errorResponse.Errors[0].StackTrace)
		},
	},
	{
		Name:          "Checker",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.ErrorResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		},
		Providers: nil,
		Checkers: []mbd.Checker{
			func(ctx context.Context, in *events.APIGatewayProxyRequest, req interface{}) (context.Context, error) {
				return nil, errors.Errorf("check failed", errors.HTTPStatusBadRequest)
			},
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			errorResponse := resp.(*mbd.ErrorResponse)

			require.Equal(t, http.StatusBadRequest, statusCode)
			require.Equal(t, http.StatusBadRequest, errorResponse.StatusCode)
			require.Equal(t, "bad-request", errorResponse.PublicMessage)
			require.Len(t, errorResponse.Errors, 1)
			require.Equal(t, "check failed", errorResponse.Errors[0].Error)
			require.NotEmpty(t, errorResponse.Errors[0].StackTrace)
		},
	},
	{
		Name:          "SerializedResponse",
		ReqTemplate:   TestRequest{},
		FormReqParser: false,
		RespTemplate:  mbd.SerializedResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			t := testcontext.GetTestingT(ctx)

			require.True(t, mbd.GetDebug(ctx))
			require.Equal(t, "/SerializedResponse", mbd.GetPath(ctx).Resource)
			require.Equal(t, "/SerializedResponse", mbd.GetPath(ctx).Path)
			require.Equal(t, "POST", mbd.GetPath(ctx).Method)
			require.Equal(t, "application/json; charset=utf-8", mbd.GetHeaders(ctx).Get("Content-Type"))
			require.Equal(t, []string{"application/json; charset=utf-8"}, mbd.GetHeaders(ctx).GetMulti("Content-Type"))
			require.Empty(t, mbd.GetQueryString(ctx).Map())
			require.Empty(t, mbd.GetQueryString(ctx).MapMulti())
			require.Empty(t, mbd.GetPathParameters(ctx).Map())
			require.Empty(t, mbd.GetStageVariables(ctx).Map())
			require.NotEmpty(t, mbd.GetRequestContext(ctx).RequestID)

			return &mbd.SerializedResponse{
				ContentType:     "text/plain",
				IsBase64Encoded: false,
				Body:            "Hello world!",
			}, nil
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			response := resp.(*mbd.SerializedResponse)

			require.Equal(t, http.StatusOK, statusCode)
			require.Equal(t, &mbd.SerializedResponse{ContentType: "text/plain", IsBase64Encoded: false, Body: "Hello world!"}, response)
		},
	},
	{
		Name:          "FormRequestParser",
		ReqTemplate:   TestRequest{},
		FormReqParser: true,
		RespTemplate:  TestResponse{},
		Request: &TestRequest{
			Value: "testValue",
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			t := testcontext.GetTestingT(ctx)

			require.True(t, mbd.GetDebug(ctx))
			require.Equal(t, "/FormRequestParser", mbd.GetPath(ctx).Resource)
			require.Equal(t, "/FormRequestParser", mbd.GetPath(ctx).Path)
			require.Equal(t, "POST", mbd.GetPath(ctx).Method)
			require.Equal(t, "application/x-www-form-urlencoded", mbd.GetHeaders(ctx).Get("Content-Type"))
			require.Equal(t, []string{"application/x-www-form-urlencoded"}, mbd.GetHeaders(ctx).GetMulti("Content-Type"))
			require.Empty(t, mbd.GetQueryString(ctx).Map())
			require.Empty(t, mbd.GetQueryString(ctx).MapMulti())
			require.Empty(t, mbd.GetPathParameters(ctx).Map())
			require.Empty(t, mbd.GetStageVariables(ctx).Map())
			require.NotEmpty(t, mbd.GetRequestContext(ctx).RequestID)

			return &TestResponse{
				Value: req.(*TestRequest).Value,
			}, nil
		},
		Assertion: func(t require.TestingT, statusCode int, headers map[string][]string, resp interface{}) {
			response := resp.(*TestResponse)

			require.Equal(t, http.StatusOK, statusCode)
			require.Equal(t, &TestResponse{Value: "testValue"}, response)
		},
	},
}
