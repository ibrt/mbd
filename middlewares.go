package mbd

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// ContextMiddleware returns a Middleware that enriches the context using the given ContextProvider(s)
// This is mainly used to inject singleton dependencies into the Handler. For other request-dependent context uses, just
// use or create a Middleware (see BodyMiddleware as example).
func ContextMiddleware(contextProviders ...ContextProvider) Middleware {
	return func(next Handler) Handler { // Middleware
		return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // Handler
			for _, contextProvider := range contextProviders {
				ctx = contextProvider(ctx)
			}
			return next(ctx, in)
		}
	}
}

// PanicMiddleware returns a Middleware that recovers panics and returns them as error responses.
// The given ErrorAdapter is used to convert the panic into a ApiGatewayProxyResponse.
func PanicMiddleware(errorAdapter ErrorAdapter) func(Handler) Handler {
	return func(next Handler) Handler { // Middleware
		return func(ctx context.Context, in events.APIGatewayProxyRequest) (out events.APIGatewayProxyResponse) { // Handler
			defer func() {
				if r := recover(); r != nil {
					out = errorAdapter(ctx, in, r)
				}
			}()

			return next(ctx, in)
		}
	}
}
