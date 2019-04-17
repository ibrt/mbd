package testrunner

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gorilla/schema"
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
	f := mbd.NewFunction(c.ReqTemplate, c.Handler).
		SetDebug(!c.DisableDebug).
		AddProviders(func(ctx context.Context) context.Context {
			return testcontext.WithTestingT(ctx, t)
		}).
		AddProviders(c.Providers...).
		AddCheckers(c.Checkers...)

	if c.FormReqParser {
		f.SetRequestParser(mbd.FormRequestParser())
	}

	out, err := f.Handler(context.Background(), *r.makeInput(t, c.FormReqParser, c.Name, c.Request))
	require.NoError(t, err)
	resp := r.parseResponse(t, c.RespTemplate, &out)
	c.Assertion(t, out.StatusCode, out.MultiValueHeaders, resp)
}

func (r *localRunner) makeInput(t *testing.T, form bool, name string, req interface{}) *events.APIGatewayProxyRequest {
	contentType := "application/json; charset=utf-8"
	if form {
		contentType = "application/x-www-form-urlencoded"
	}

	var body []byte
	if req != nil {
		if form {
			v := url.Values{}
			require.NoError(t, schema.NewEncoder().Encode(req, v))
			body = []byte(v.Encode())
		} else {
			contentType = "application/json; charset=utf-8"
			var err error
			body, err = json.MarshalIndent(req, "", "  ")
			require.NoError(t, err)
		}
	}

	in := &events.APIGatewayProxyRequest{
		Resource:   "/" + name,
		Path:       "/" + name,
		HTTPMethod: "POST",
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		MultiValueHeaders: map[string][]string{
			"Content-Type": {contentType},
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

	if _, ok := respTemplate.(mbd.SerializedResponse); ok {
		resp := &mbd.SerializedResponse{
			ContentType:     out.Headers["Content-Type"],
			IsBase64Encoded: out.IsBase64Encoded,
			Body:            out.Body,
		}
		r.printValue("Response", resp)
		return resp
	}

	resp := reflect.New(reflect.TypeOf(respTemplate)).Interface()
	require.NoError(t, json.Unmarshal([]byte(out.Body), resp))
	r.printValue("Response", resp)
	return resp
}
