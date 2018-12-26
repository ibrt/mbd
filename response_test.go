package mbd_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/mbd"
	"github.com/stretchr/testify/require"
)

func TestNewResponse(t *testing.T) {
	resp := mbd.NewResponse(context.Background())
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, resp.Headers)
	require.Empty(t, resp.Body)
	require.False(t, resp.IsBase64Encoded)

	resp = mbd.NewResponse(context.Background(), mbd.StatusCode(http.StatusBadRequest), mbd.StringBody("body", "text/plain", false))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "text/plain", resp.Headers["Content-Type"])
	require.Len(t, resp.Headers, 1)
	require.Equal(t, "body", resp.Body)
	require.False(t, resp.IsBase64Encoded)
}

func TestStatusCode(t *testing.T) {
	resp := &events.APIGatewayProxyResponse{}
	mbd.StatusCode(http.StatusInternalServerError)(context.Background(), resp)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Empty(t, resp.Headers)
	require.False(t, resp.IsBase64Encoded)
	require.Empty(t, resp.Body)
}

func TestJSONBody(t *testing.T) {
	resp := &events.APIGatewayProxyResponse{}
	mbd.JSONBody(map[string]interface{}{"key": "value"})(context.Background(), resp)
	require.Equal(t, 0, resp.StatusCode)
	require.Equal(t, "application/json; charset=utf-8", resp.Headers["Content-Type"])
	require.Len(t, resp.Headers, 1)
	require.Equal(t, "{\n  \"key\": \"value\"\n}", resp.Body)
	require.False(t, resp.IsBase64Encoded)
}

func TestBinaryBody(t *testing.T) {
	resp := &events.APIGatewayProxyResponse{}
	mbd.BinaryBody([]byte{1, 2, 3}, "application/octet-stream")(context.Background(), resp)
	require.Equal(t, 0, resp.StatusCode)
	require.Equal(t, "application/octet-stream", resp.Headers["Content-Type"])
	require.Len(t, resp.Headers, 1)
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), resp.Body)
	require.True(t, resp.IsBase64Encoded)
}

func TestStringBody(t *testing.T) {
	resp := &events.APIGatewayProxyResponse{}
	mbd.StringBody("body", "text/plain", false)(context.Background(), resp)
	require.Equal(t, 0, resp.StatusCode)
	require.Equal(t, "text/plain", resp.Headers["Content-Type"])
	require.Len(t, resp.Headers, 1)
	require.Equal(t, "body", resp.Body)
	require.False(t, resp.IsBase64Encoded)

	resp = &events.APIGatewayProxyResponse{}
	mbd.StringBody(base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), "application/octet-stream", true)(context.Background(), resp)
	require.Equal(t, 0, resp.StatusCode)
	require.Equal(t, "application/octet-stream", resp.Headers["Content-Type"])
	require.Len(t, resp.Headers, 1)
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), resp.Body)
	require.True(t, resp.IsBase64Encoded)
}
