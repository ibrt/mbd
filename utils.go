package mbd

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// ContextProviders combines the given ContextProvider(s) into a single one.
func ContextProviders(ctxProviders ...ContextProvider) ContextProvider {
	return func(ctx context.Context) context.Context {
		for _, ctxProvider := range ctxProviders {
			ctx = ctxProvider(ctx)
		}
		return ctx
	}
}

// Middlewares combines the given Middleware(s) into single one.
func Middlewares(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler { // Middleware
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Validators combines the given Validator(s) into a single one.
func Validators(validators ...Validator) Validator {
	return func(ctx context.Context, in events.APIGatewayProxyRequest) {
		for _, validator := range validators {
			validator(ctx, in)
		}
	}
}
