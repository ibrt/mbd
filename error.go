package mbd

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
)

// DefaultErrorResponse describes an standard error response.
// It is used by DefaultErrorAdapter and optionally by your own ErrorAdapter.
type DefaultErrorResponse struct {
	StatusCode    int    `json:"statusCode"`
	PublicMessage string `json:"publicMessage"`
	RequestID     string `json:"requestId"`

	// included only if debug context value is set to true
	Error      string   `json:"error,omitempty"`
	StackTrace []string `json:"stackTrace,omitempty"`
}

// DefaultErrorAdapter implements mbd.ErrorAdapter. It optionally accepts an error wrapped via github.com/ibrt/errors,
// in which case it extracts status code, public message, and stack trace using the corresponding error behaviors. If a
// standard error is provided, the status code is set to 500, and the public message is set to "internal-server-error".
// Any other value is first converted to error using fmt.Errorf. If the debug flag is set in the context, the Error()
// string and stack trace are included in the response. The generated response will contain a JSON body.
func DefaultErrorAdapter(ctx context.Context, _ events.APIGatewayProxyRequest, r interface{}) events.APIGatewayProxyResponse {
	err := errors.MaybeWrapRecover(r)
	statusCode := errors.GetHTTPStatusOrDefault(err, http.StatusInternalServerError)

	publicMessage := errors.GetPublicMessageOrDefault(err, http.StatusText(statusCode))
	publicMessage = strings.ToLower(publicMessage)
	publicMessage = strings.Replace(publicMessage, " ", "-", -1)
	publicMessage = strings.Replace(publicMessage, "'", "", -1)

	resp := &DefaultErrorResponse{
		StatusCode:    statusCode,
		PublicMessage: publicMessage,
		RequestID:     GetRequestID(ctx),
	}

	if GetDebug(ctx) {
		resp.Error = err.Error()
		resp.StackTrace = errors.FormatCallers(errors.GetCallersOrCurrent(err))
	}

	return NewResponse(ctx, StatusCode(statusCode), JSONBody(resp))
}
