package testrunner

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/mbd"
	"github.com/ibrt/mbd/internal/testcases"
	"github.com/ibrt/mbd/internal/testcontext"
	"github.com/stretchr/testify/require"
)

type localRunner struct {
	baseRunner
}

// NewLocalRunner returns a Runner that runs test cases locally in-memory.
func NewLocalRunner() Runner {
	return &localRunner{}
}

// RunTest implements Runner.
func (r *localRunner) RunTest(t *testing.T, c *testcases.TestCase) {
	out, err := mbd.NewFunction(c.ReqTemplate, c.Handler).
		SetDebug(!c.DisableDebug).
		AddProviders(func(ctx context.Context) context.Context {
			return testcontext.WithTestingT(ctx, t)
		}).
		AddProviders(c.Providers...).
		AddCheckers(c.Checkers...).
		Handler(context.Background(), *r.makeInput(t, c.Name, c.Request))

	require.NoError(t, err)
	resp := r.parseResponse(t, c.RespTemplate, &out)
	c.Assertion(t, out.StatusCode, out.MultiValueHeaders, resp)
}

func (r *localRunner) makeInput(t *testing.T, name string, req interface{}) *events.APIGatewayProxyRequest {
	var body []byte
	if req != nil {
		var err error
		body, err = json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)
	}

	in := &events.APIGatewayProxyRequest{
		Resource:   "/" + name,
		Path:       "/" + name,
		HTTPMethod: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
		MultiValueHeaders: map[string][]string{
			"Content-Type": {"application/json; charset=utf-8"},
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "test-" + name,
		},
		Body: string(body),
	}

	r.printValue("Input", in)
	r.printValue("Request", req)

	return in
}

func (r *localRunner) parseResponse(t *testing.T, respTemplate interface{}, out *events.APIGatewayProxyResponse) interface{} {
	r.printValue("Output", out)

	if respTemplate == nil {
		require.Empty(t, out.Body)
		r.printValue("Response", nil)
		return nil
	}

	resp := reflect.New(reflect.TypeOf(respTemplate)).Interface()
	require.NoError(t, json.Unmarshal([]byte(out.Body), resp))
	r.printValue("Response", resp)
	return resp
}
