package mbd

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestJSONBodyAdapter(t *testing.T) {
	type Request struct {
		Key    string      `json:"key"`
		Number interface{} `json:"number"`
	}

	require.Panics(t, func() { JSONBodyAdapter(reflect.TypeOf("bad")) })
	require.Panics(t, func() { JSONBodyAdapter(reflect.TypeOf(&Request{})) })

	body := JSONBodyAdapter(reflect.TypeOf(Request{}))(context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{ "key": "value", "number": 1, "unknown": "unknown" }`,
		})
	require.Equal(t, &Request{Key: "value", Number: float64(1)}, body)

	body = JSONBodyAdapter(reflect.TypeOf(Request{}), JSONDecoderUseNumber())(context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{ "key": "value", "number": 1, "unknown": "unknown" }`,
		})
	require.Equal(t, &Request{Key: "value", Number: json.Number("1")}, body)

	require.Panics(t, func() {
		JSONBodyAdapter(reflect.TypeOf(Request{}), JSONDecoderDisallowUnknownFields())(context.Background(),
			events.APIGatewayProxyRequest{
				Body: `{ "key": "value", "number": 1, "unknown": "unknown" }`,
			})
	})
}
