package mbd

import (
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"reflect"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
)

type noRequestBodyType struct {
	// intentionally empty
}

type errorResponse struct {
	StatusCode    int    `json:"statusCode"`
	PublicMessage string `json:"publicMessage"`
	RequestID     string `json:"requestId"`

	// included only if debug context value is set to true
	Error      string   `json:"error,omitempty"`
	StackTrace []string `json:"stackTrace,omitempty"`
}

var (
	invalidBody        = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("invalid-body"))
	invalidContentType = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("invalid-content-type"))
	unexpectedBody     = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("unexpected-body"))
	noRequestBody      = reflect.TypeOf(noRequestBodyType{})
)

func adaptError(ctx context.Context, err error) *events.APIGatewayProxyResponse {
	statusCode := errors.GetHTTPStatusOrDefault(err, http.StatusInternalServerError)

	resp := &errorResponse{
		StatusCode:    statusCode,
		PublicMessage: errors.GetPublicMessageOrDefault(err, getDefaultPublicMessage(statusCode)),
		RequestID:     GetRequestContext(ctx).RequestID,
	}

	if GetDebug(ctx) {
		resp.Error = err.Error()
		resp.StackTrace = errors.FormatCallers(errors.GetCallersOrCurrent(err))
	}

	return adaptResponse(ctx, statusCode, resp)
}

func getDefaultPublicMessage(statusCode int) string {
	if statusCode == http.StatusTeapot {
		return "i-am-a-teapot"
	}
	if statusText := strings.ToLower(http.StatusText(statusCode)); statusText != "" {
		return strings.Replace(strings.ToLower(statusText), " ", "-", -1)
	}
	return "unknown"
}

func parseRequest(_ context.Context, reqType reflect.Type, in *events.APIGatewayProxyRequest) (interface{}, error) {
	if in.IsBase64Encoded {
		return nil, errors.Errorf("invalid IsBase64Encoded: expected 'false', got 'true'", invalidBody)
	}

	if reqType == noRequestBody {
		if in.Body != "" {
			return nil, errors.Errorf("unexpected Body", unexpectedBody)
		}

		return nil, nil
	}

	req := reflect.New(reqType).Interface()

	dec := json.NewDecoder(strings.NewReader(in.Body))
	dec.DisallowUnknownFields()
	dec.UseNumber()

	if err := dec.Decode(req); err != nil {
		return nil, errors.Wrap(err, errors.Prefix("invalid Body"), invalidBody)
	}

	return req, nil
}

func adaptResponse(_ context.Context, statusCode int, resp interface{}) *events.APIGatewayProxyResponse {
	out := &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":  "application/json; charset=utf-8",
			"Cache-Control": "no-cache, no-store, must-revalidate",
			"Pragma":        "no-cache",
			"Expires":       "0",
		},
	}

	if resp != nil {
		buf, err := json.MarshalIndent(resp, "", "  ")
		errors.MaybeMustWrap(err)
		out.Body = string(buf)
		out.IsBase64Encoded = false
	}

	return out
}

func checkContentType(_ context.Context, in *events.APIGatewayProxyRequest, _ interface{}) error {
	contentType := in.Headers["Content-Type"]
	if contentType == "" {
		return nil
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return errors.Wrap(err, invalidContentType)
	}

	if strings.ToLower(mediaType) != "application/json" {
		return errors.Errorf("bad Content-Type: expected mime type 'application/json', got '%v'", contentType, invalidContentType)

	}

	if charset, ok := params["charset"]; ok && strings.ToLower(charset) != "utf-8" {
		return errors.Errorf("bad Content-Type: expected charset 'utf-8', got '%v'", charset, invalidContentType)
	}

	return nil
}
