package mbd

import "context"

// ContextProviders combines the given ContextProvider(s) into a single one.
func ContextProviders(ctxProvider ContextProvider, ctxProviders ...ContextProvider) ContextProvider {
	ctxProviders = append([]ContextProvider{ctxProvider}, ctxProviders...)

	return func(ctx context.Context) context.Context {
		for _, ctxProvider := range ctxProviders {
			ctx = ctxProvider(ctx)
		}
		return ctx
	}
}

// Middlewares combines the given Middleware(s) into single one.
func Middlewares(middleware Middleware, middlewares ...Middleware) Middleware {
	middlewares = append([]Middleware{middleware}, middlewares...)

	return func(next Handler) Handler { // Middleware
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
