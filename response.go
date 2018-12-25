package mbd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/ibrt/errors"

	"github.com/aws/aws-lambda-go/events"
)

// NewResponse is a utility for generating Handler responses.
// By default the status code is 200 and no headers are set.
// Pass ResponseOption(s) to set status code, headers and include a response body.
// The given ResponseOption(s) are applied in the same order as provided.
func NewResponse(ctx context.Context, options ...ResponseOption) events.APIGatewayProxyResponse {
	resp := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
	}

	for _, option := range options {
		option(ctx, &resp)
	}

	return resp
}

// StatusCode returns ResponseOption that sets the given status code on the response.
func StatusCode(statusCode int) ResponseOption {
	return func(_ context.Context, resp *events.APIGatewayProxyResponse) {
		resp.StatusCode = statusCode
	}
}

// JSONBody returns a ResponseOption that marshals the given body to JSON and sets it on the response.
// It panics in case json.Marshal returns an error.
// The "Content-Type" header is set to "application/json; charset=utf-8".
func JSONBody(body interface{}) ResponseOption {
	return func(_ context.Context, resp *events.APIGatewayProxyResponse) {
		buf, err := json.MarshalIndent(body, "", "  ")
		errors.MaybeMustWrap(err)
		resp.Headers["Content-Type"] = "application/json; charset=utf-8"
		resp.Body = string(buf)
		resp.IsBase64Encoded = false
	}
}

// BinaryBody returns a ResponseOption that encodes the given body to base64 and sets it on the response.
// The "Content-Type" header is set to the given string.
func BinaryBody(body []byte, contentType string) ResponseOption {
	return func(_ context.Context, resp *events.APIGatewayProxyResponse) {
		resp.Headers["Content-Type"] = contentType
		resp.Body = base64.StdEncoding.EncodeToString(body)
		resp.IsBase64Encoded = true
	}
}

// StringBody returns a ResponseOption that sets the given body on the response.
// The "Content-Type" header is set to the given string.
// The IsBase64Encoded flag is also set as provided.
func StringBody(body, contentType string, isBase64Encoded bool) ResponseOption {
	return func(_ context.Context, resp *events.APIGatewayProxyResponse) {
		resp.Headers["Content-Type"] = contentType
		resp.Body = body
		resp.IsBase64Encoded = isBase64Encoded
	}
}

// Header returns a ResponseOption that sets the given header on the response.
func Header(name, value string) ResponseOption {
	return func(_ context.Context, resp *events.APIGatewayProxyResponse) {
		resp.Headers[name] = value
	}
}
