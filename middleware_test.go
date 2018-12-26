package mbd_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/ibrt/mbd"
	"github.com/stretchr/testify/require"
)

func TestContextMiddleware(t *testing.T) {
	contextProvider := mbd.StaticContextProvider("key", "value")

	mbd.ContextMiddleware(contextProvider)(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			require.Equal(t, "value", ctx.Value("key").(string))
			return events.APIGatewayProxyResponse{}
		})(
		context.Background(),
		events.APIGatewayProxyRequest{})
}

func TestDebugMiddleware(t *testing.T) {
	mbd.DebugMiddleware(false)(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			require.False(t, mbd.GetDebug(ctx))
			return events.APIGatewayProxyResponse{}
		})(
		context.Background(),
		events.APIGatewayProxyRequest{})

	mbd.DebugMiddleware(true)(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			require.True(t, mbd.GetDebug(ctx))
			return events.APIGatewayProxyResponse{}
		})(
		context.Background(),
		events.APIGatewayProxyRequest{})
}

func TestRequestIDMiddleware(t *testing.T) {
	mbd.RequestIDMiddleware()(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			require.Equal(t, "", mbd.GetRequestID(ctx))
			return events.APIGatewayProxyResponse{}
		})(
		context.Background(),
		events.APIGatewayProxyRequest{})

	mbd.RequestIDMiddleware()(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			require.Equal(t, "request-id", mbd.GetRequestID(ctx))
			return events.APIGatewayProxyResponse{}
		})(
		context.Background(),
		events.APIGatewayProxyRequest{
			RequestContext: events.APIGatewayProxyRequestContext{
				RequestID: "request-id",
			},
		})
}

func TestPanicMiddleware(t *testing.T) {
	out := mbd.PanicMiddleware(mbd.DefaultErrorAdapter)(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			panic("test error")
		})(
		context.Background(),
		events.APIGatewayProxyRequest{})

	require.Equal(t, http.StatusInternalServerError, out.StatusCode)

	out = mbd.PanicMiddleware(mbd.DefaultErrorAdapter)(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			panic(errors.Errorf("test error", mbd.BadEncoding))
		})(
		context.Background(),
		events.APIGatewayProxyRequest{})

	resp := &mbd.DefaultErrorResponse{}
	require.NoError(t, json.Unmarshal([]byte(out.Body), resp))

	require.Equal(t, http.StatusBadRequest, out.StatusCode)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "bad-encoding", resp.PublicMessage)
}

func TestBodyMiddleware(t *testing.T) {
	type Request struct {
		Key string `json:"key"`
	}

	mbd.BodyMiddleware(mbd.JSONBodyAdapter(reflect.TypeOf(Request{})))(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			require.Equal(t, &Request{Key: "value"}, mbd.GetBody(ctx))
			return events.APIGatewayProxyResponse{}
		})(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{ "key": "value" }`,
		})
}

func TestValidatorMiddleware(t *testing.T) {
	require.PanicsWithValue(t, "invalid", func() {
		mbd.ValidatorMiddleware(func(ctx context.Context, in events.APIGatewayProxyRequest) { panic("invalid") })(
			func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
				return events.APIGatewayProxyResponse{}
			})(
			context.Background(),
			events.APIGatewayProxyRequest{})
	})
}
