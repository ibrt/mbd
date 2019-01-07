package testcases

import (
	"context"
	"net/http"

	"github.com/ibrt/mbd"

	"github.com/stretchr/testify/require"
)

// TestRequest is a request for test functions.
type TestRequest struct {
	Value string `json:"value"`
}

// TestResponse is a response for test functions.
type TestResponse struct {
	Value string `json:"value"`
}

// TestCase describes a test case.
type TestCase struct {
	Name         string
	ReqTemplate  interface{}
	RespTemplate interface{}
	Request      interface{}
	Handler      mbd.Handler
	Providers    []mbd.Provider
	Checkers     []mbd.Checker
	Assertion    func(require.TestingT, int, map[string][]string, interface{})
}

// GetTestCase returns the TestCase with the given name, panics if not found.
func GetTestCase(name string) *TestCase {
	for _, c := range TestCases {
		if c.Name == name {
			return c
		}
	}
	panic("unknown test case")
}

// TestCases is a list of test cases, which can be run locally in-memory, or against a test remote Lambda deployment.
var TestCases = []*TestCase{
	{
		Name:         "MissingBody",
		ReqTemplate:  TestRequest{},
		RespTemplate: mbd.ErrorResponse{},
		Request:      nil,
		Handler: func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		},
		Providers: nil,
		Checkers:  nil,
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
}
