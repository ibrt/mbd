package mbd

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/ibrt/errors"

	"github.com/aws/aws-lambda-go/events"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultPublicMessage(t *testing.T) {
	require.Equal(t, "i-am-a-teapot", getDefaultPublicMessage(http.StatusTeapot))
	require.Equal(t, "internal-server-error", getDefaultPublicMessage(http.StatusInternalServerError))
	require.Equal(t, "unknown", getDefaultPublicMessage(1))
}

func TestParseRequest_InvalidBody(t *testing.T) {
	_, err := parseRequest(context.Background(), reflect.TypeOf(struct{}{}), &events.APIGatewayProxyRequest{IsBase64Encoded: true})
	require.EqualError(t, err, "invalid IsBase64Encoded: expected 'false', got 'true'")
	require.Equal(t, http.StatusBadRequest, errors.GetHTTPStatus(err))
	require.Equal(t, "invalid-body", errors.GetPublicMessage(err))
}

func TestParseRequest_UnexpectedBody(t *testing.T) {
	_, err := parseRequest(context.Background(), noRequestBody, &events.APIGatewayProxyRequest{Body: "unexpected"})
	require.EqualError(t, err, "unexpected Body")
	require.Equal(t, http.StatusBadRequest, errors.GetHTTPStatus(err))
	require.Equal(t, "unexpected-body", errors.GetPublicMessage(err))
}
