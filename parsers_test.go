package mbd

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/stretchr/testify/require"
)

func TestParseJSONRequest_InvalidBody(t *testing.T) {
	_, err := JSONRequestParser()(context.Background(), reflect.TypeOf(struct{}{}), &events.APIGatewayProxyRequest{IsBase64Encoded: true})
	require.EqualError(t, err, "invalid IsBase64Encoded: expected 'false', got 'true'")
	require.Equal(t, http.StatusBadRequest, errors.GetHTTPStatus(err))
	require.Equal(t, "invalid-body", errors.GetPublicMessage(err))
}

func TestParseJSONRequest_UnexpectedBody(t *testing.T) {
	_, err := JSONRequestParser()(context.Background(), noRequestBody, &events.APIGatewayProxyRequest{Body: "unexpected"})
	require.EqualError(t, err, "unexpected Body")
	require.Equal(t, http.StatusBadRequest, errors.GetHTTPStatus(err))
	require.Equal(t, "unexpected-body", errors.GetPublicMessage(err))
}

func TestParseFormRequest_InvalidBody(t *testing.T) {
	_, err := FormRequestParser()(context.Background(), reflect.TypeOf(struct{}{}), &events.APIGatewayProxyRequest{IsBase64Encoded: true})
	require.EqualError(t, err, "invalid IsBase64Encoded: expected 'false', got 'true'")
	require.Equal(t, http.StatusBadRequest, errors.GetHTTPStatus(err))
	require.Equal(t, "invalid-body", errors.GetPublicMessage(err))
}

func TestParseFormRequest_UnexpectedBody(t *testing.T) {
	_, err := FormRequestParser()(context.Background(), noRequestBody, &events.APIGatewayProxyRequest{Body: "unexpected"})
	require.EqualError(t, err, "unexpected Body")
	require.Equal(t, http.StatusBadRequest, errors.GetHTTPStatus(err))
	require.Equal(t, "unexpected-body", errors.GetPublicMessage(err))
}
