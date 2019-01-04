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

var (
	badEncoding              = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("bad-encoding"))
	badJSON                  = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("bad-json"))
	badContentType           = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("bad-content-type"))
	badPathParameters        = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("bad-path-parameters"))
	badQueryStringParameters = errors.Behaviors(errors.HTTPStatusBadRequest, errors.PublicMessage("bad-query-string-parameters"))
)

type errorResponse struct {
	StatusCode    int    `json:"statusCode"`
	PublicMessage string `json:"publicMessage"`
	RequestID     string `json:"requestId"`

	// included only if debug context value is set to true
	Error      string   `json:"error,omitempty"`
	StackTrace []string `json:"stackTrace,omitempty"`
}

func adaptError(ctx context.Context, err error) events.APIGatewayProxyResponse {
	statusCode := errors.GetHTTPStatusOrDefault(err, http.StatusInternalServerError)

	resp := &errorResponse{
		StatusCode:    statusCode,
		PublicMessage: errors.GetPublicMessageOrDefault(err, getDefaultPublicMessage(statusCode)),
		RequestID:     GetRequestID(ctx),
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

func parseRequest(_ context.Context, reqType reflect.Type, in events.APIGatewayProxyRequest) (interface{}, error) {
	if in.IsBase64Encoded {
		return nil, errors.Errorf("bad IsBase64Encoded: expected 'false', got 'true'", badEncoding)
	}

	req := reflect.New(reqType).Interface()

	dec := json.NewDecoder(strings.NewReader(in.Body))
	dec.DisallowUnknownFields()
	dec.UseNumber()

	if err := dec.Decode(req); err != nil {
		return nil, errors.Wrap(err, errors.Prefix("bad Body"), badJSON)
	}

	return req, nil
}

func adaptResponse(_ context.Context, statusCode int, resp interface{}) events.APIGatewayProxyResponse {
	buf, err := json.MarshalIndent(resp, "", "  ")
	errors.MaybeMustWrap(err)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":  "application/json; charset=utf-8",
			"Cache-Control": "no-cache, no-store, must-revalidate",
			"Pragma":        "no-cache",
			"Expires":       "0",
		},
		Body:            string(buf),
		IsBase64Encoded: false,
	}
}

func checkContentType(_ context.Context, in events.APIGatewayProxyRequest, _ interface{}) error {
	contentType := in.Headers["Content-Type"]
	if contentType == "" {
		return nil
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return errors.Wrap(err, badContentType)
	}

	if mediaType != "application/json" {
		return errors.Errorf("bad Content-Type: expected mime type 'application/json', got '%v'", contentType, badContentType)

	}

	if charset, ok := params["charset"]; ok && charset != "utf-8" {
		return errors.Errorf("bad Content-Type: expected charset 'utf-8', got '%v'", charset, badContentType)
	}

	return nil
}

func checkParameters(_ context.Context, in events.APIGatewayProxyRequest, _ interface{}) error {
	if len(in.PathParameters) > 0 {
		return errors.Errorf("bad PathParameters: expected 'map[]', got '%v'", in.PathParameters, badPathParameters)
	}
	if len(in.QueryStringParameters) > 0 {
		return errors.Errorf("bad QueryStringParameters: expected 'map[]', got '%v'", in.QueryStringParameters, badQueryStringParameters)
	}
	if len(in.MultiValueQueryStringParameters) > 0 {
		return errors.Errorf("bad MultiValueQueryStringParameters: expected 'map[]', got '%v'", in.MultiValueQueryStringParameters, badQueryStringParameters)
	}
	return nil
}
