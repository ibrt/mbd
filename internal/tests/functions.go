package tests

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/ibrt/mbd"
	"github.com/stretchr/testify/require"
)

type testRequest struct {
	Behavior               string              `json:"behavior"`
	ResponseValue          string              `json:"responseValue"`
	ErrorStatusCode        int                 `json:"errorStatusCode"`
	ErrorPublicMessage     string              `json:"errorPublicMessage"`
	ExpectedQueryString    map[string][]string `json:"expectedQueryString,omitempty"`
	ExpectedPathParameters map[string]string   `json:"expectedPathParameters,omitempty"`
	ExpectedStageVariables map[string]string   `json:"expectedStageVariables,omitempty"`
	ExpectedRequestID      string              `json:"expectedRequestId,omitempty"`
}

type testResponse struct {
	Value string `json:"value"`
}

type testFunction struct {
	ReqTemplate interface{}
	Checkers    []mbd.Checker
	Handler     mbd.Handler
}

var testFunctions = map[string]*testFunction{
	"FunctionWithBody": {
		ReqTemplate: testRequest{},
		Checkers: []mbd.Checker{
			func(ctx context.Context, _ *events.APIGatewayProxyRequest, req interface{}) error {
				request := req.(*testRequest)
				_ = getTesting(ctx) // panics if value is missing

				if b := request.Behavior; b != "sendResponse" && b != "sendError" && b != "panic" {
					return errors.Errorf("invalid Behavior '%v'", b, errors.HTTPStatusBadRequest)
				}

				if request.Behavior == "sendResponse" && request.ResponseValue == "" {
					return errors.Errorf("missing ResponseValue", errors.HTTPStatusBadRequest)
				}

				return nil
			},
		},
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			request := req.(*testRequest)
			t := getTesting(ctx)

			require.True(t, mbd.GetDebug(ctx))
			require.NotEmpty(t, mbd.GetPath(ctx).Path)
			require.NotEmpty(t, mbd.GetPath(ctx).Method)
			require.NotEmpty(t, mbd.GetPath(ctx).Resource)
			require.Equal(t, "application/json; charset=utf-8", mbd.GetHeaders(ctx).Get("Content-Type"))

			if request.ExpectedQueryString != nil {
				require.Equal(t, request.ExpectedQueryString, mbd.GetQueryString(ctx).MapMulti())
			}
			if request.ExpectedPathParameters != nil {
				require.Equal(t, request.ExpectedPathParameters, mbd.GetPathParameters(ctx).Map())
			}
			if request.ExpectedStageVariables != nil {
				require.Equal(t, request.ExpectedStageVariables, mbd.GetStageVariables(ctx).Map())
			}
			if request.ExpectedRequestID != "" {
				require.Equal(t, request.ExpectedRequestID, mbd.GetRequestContext(ctx).RequestID)
			}

			if request.Behavior == "sendResponse" {
				return &testResponse{Value: request.ResponseValue}, nil
			}

			err := errors.Errorf("simulated error")

			if request.ErrorStatusCode != 0 {
				err = errors.Wrap(err, errors.HTTPStatus(request.ErrorStatusCode))
			}
			if request.ErrorPublicMessage != "" {
				err = errors.Wrap(err, errors.PublicMessage(request.ErrorPublicMessage))
			}

			if request.Behavior == "sendError" {
				return nil, err
			}

			panic(err)
		},
	},
}

// GetTestFunction initializes and returns a test function.
func GetTestFunction(functionName string) *mbd.Function {
	testFunction, ok := testFunctions[functionName]
	errors.Assert(ok, "unknown test function '%v'", functionName)

	return mbd.NewFunction(testFunction.ReqTemplate, testFunction.Handler).
		SetDebug(true).
		AddProviders(mbd.SingletonProvider("key", "value")).
		AddCheckers(testFunction.Checkers...)
}
