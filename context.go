package mbd

import (
	"context"

	"github.com/ibrt/errors"
)

type contextKey int

const (
	debugContextKey contextKey = iota
	requestIDContextKey
	bodyContextKey
)

// GetDebug returns the debug flag stored in context. If missing, it returns false.
func GetDebug(ctx context.Context) bool {
	if debug, ok := ctx.Value(debugContextKey).(bool); ok {
		return debug
	}
	return false
}

// With debug adds the given debug flag to context.
// It is mainly meant to be used by DebugMiddleware.
func WithDebug(ctx context.Context, debug bool) context.Context {
	return context.WithValue(ctx, debugContextKey, debug)
}

// GetRequestID returns the request ID stored in context. If missing, it returns "".
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}

// WithRequestID adds the given request ID to context.
// It is mainly meant to be used by RequestIDMiddleware.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

// GetBody returns the request body stored in context. If missing, it panics.
// The type of the returned value will depend on the BodyAdapter used in BodyMiddleware.
func GetBody(ctx context.Context) interface{} {
	req := ctx.Value(bodyContextKey)
	errors.Assert(req != nil, "BodyMiddleware not installed")
	return req
}

// WithBody adds the given request body to the context.
// It is mainly meant to be used by BodyMiddleware.
func WithBody(ctx context.Context, body interface{}) context.Context {
	return context.WithValue(ctx, bodyContextKey, body)
}
