package tests

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	TestName         string
	FunctionName     string
	Request          interface{}
	ExpectedResponse interface{}
	ExpectedError    error
}

var testCases = []*testCase{
	{
		TestName:         "MissingBody",
		FunctionName:     "FunctionWithBody",
		Request:          nil,
		ExpectedResponse: nil,
		ExpectedError:    errors.Errorf("invalid Body: EOF", errors.HTTPStatusBadRequest, errors.PublicMessage("invalid-body")),
	},
}

func makeTestCaseInput(t require.TestingT, c *testCase) *events.APIGatewayProxyRequest {
	return &events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "test-" + c.TestName,
		},
		Body: adaptRequest(t, c.Request),
	}
}

func assertTestCase(t require.TestingT, c *testCase, out *events.APIGatewayProxyResponse) {
	if c.ExpectedError != nil {
		require.Equal(t, out.StatusCode, errors.GetHTTPStatusOrDefault(c.ExpectedError, http.StatusInternalServerError))
		require.Equal(t, "application/json; charset=utf-8", out.Headers["Content-Type"])
		require.Equal(t, "no-cache, no-store, must-revalidate", out.Headers["Cache-Control"])
		require.Equal(t, "no-cache", out.Headers["Pragma"])
		require.Equal(t, "0", out.Headers["Expires"])
		require.NotEmpty(t, out.Body)
		require.False(t, out.IsBase64Encoded)

		resp := parseErrorResponse(t, out)
		dumpValue("RESPONSE", resp)

		require.Equal(t, errors.GetHTTPStatusOrDefault(c.ExpectedError, http.StatusInternalServerError), resp.StatusCode)
		require.Equal(t, errors.GetPublicMessage(c.ExpectedError), resp.PublicMessage)
		require.Equal(t, "test-"+c.TestName, resp.RequestID)
		require.Len(t, resp.Errors, 1)
		require.Equal(t, c.ExpectedError.Error(), resp.Errors[0].Error)
	}
}
