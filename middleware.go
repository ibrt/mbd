package mbd

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// ContextMiddleware returns a Middleware that enriches the context using the given ContextProvider(s)
// This is mainly used to inject singleton dependencies into the Handler. For other request-dependent context uses,
// define your own Middleware.
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

// DebugMiddleware returns a Middleware that stores a debug flag in the context.
// Note that DefaultErrorAdapter includes a detailed error and stack trace in the response if the debug flag is set.
func DebugMiddleware(debug bool) Middleware {
	return func(next Handler) Handler { // Middleware
		return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // Handler
			return next(context.WithValue(ctx, debugContextKey, debug), in)
		}

	}
}

// RequestIDMiddleware returns a Middleware that stores the Lambda request ID in the context.
func RequestIDMiddleware() Middleware {
	return func(next Handler) Handler { // Middleware
		return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // Handler
			return next(context.WithValue(ctx, requestIDContextKey, in.RequestContext.RequestID), in)
		}

	}
}

// PanicMiddleware returns a Middleware that recovers panics and returns them as error responses.
// The given ErrorAdapter is used to convert the panic into a ApiGatewayProxyResponse.
func PanicMiddleware(errorAdapter ErrorAdapter) Middleware {
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

// BodyMiddleware returns a Middleware that parses the request body using the given BodyAdapter.
func BodyMiddleware(bodyAdapter BodyAdapter) Middleware {
	return func(next Handler) Handler { // Middleware
		return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // Handler
			req := bodyAdapter(ctx, in)
			return next(context.WithValue(ctx, bodyContextKey, req), in)
		}
	}
}

// ValidatorMiddleware returns a Middleware that validates the request using the given Validator(s).
func ValidatorMiddleware(validators ...Validator) Middleware {
	return func(next Handler) Handler { // Middleware
		return func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse { // Handler
			for _, requestValidator := range validators {
				requestValidator(ctx, in)
			}
			return next(ctx, in)
		}
	}
}
