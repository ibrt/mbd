package mbd

import (
	"context"
)

type contextKey int

const (
	debugContextKey contextKey = iota
	requestIDContextKey
)

// GetDebug returns the debug flag stored in context. If missing, it returns false.
func GetDebug(ctx context.Context) bool {
	if debug, ok := ctx.Value(debugContextKey).(bool); ok {
		return debug
	}
	return false
}

// GetRequestID returns the request ID stored in context. If missing, it returns "".
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
