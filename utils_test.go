package mbd_test

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/mbd"
	"github.com/stretchr/testify/require"
)

func TestContextProviders(t *testing.T) {
	contextProvider := mbd.ContextProviders(mbd.StaticContextProvider("k1", "v1"), mbd.StaticContextProvider("k2", "v2"))
	ctx := contextProvider(context.Background())
	require.Equal(t, "v1", ctx.Value("k1"))
	require.Equal(t, "v2", ctx.Value("k2"))
}

func TestStaticContextProvider(t *testing.T) {
	ctx := mbd.StaticContextProvider("k", "v")(context.Background())
	require.Equal(t, "v", ctx.Value("k"))
}

func TestMiddlewares(t *testing.T) {
	middlewares := mbd.Middlewares(
		func(next mbd.Handler) mbd.Handler {
			return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // mbd.Handle
				require.Empty(t, in.Body)
				in.Body += "m1"
				out := next(ctx, in)
				out.Body += "m1"
				return out
			}
		},
		func(next mbd.Handler) mbd.Handler {
			return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // mbd.Handler
				require.Equal(t, "m1", in.Body)
				in.Body += "m2"
				out := next(ctx, in)
				out.Body += "m2"
				return out
			}
		})(
		func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // mbd.Handler
			require.Equal(t, "m1m2", in.Body)
			return events.APIGatewayProxyResponse{
				Body: "h",
			}
		})

	out := middlewares(context.Background(), events.APIGatewayProxyRequest{})
	require.Equal(t, "hm2m1", out.Body)
}

func TestValidators(t *testing.T) {
	out := ""
	validators := mbd.Validators(
		func(_ context.Context, in events.APIGatewayProxyRequest) { out += "v1" },
		func(_ context.Context, in events.APIGatewayProxyRequest) { out += "v2" })
	validators(context.Background(), events.APIGatewayProxyRequest{})
	require.Equal(t, "v1v2", out)
}
