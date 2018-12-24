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

func GetDebug(ctx context.Context) bool {
	if debug, ok := ctx.Value(debugContextKey).(bool); ok {
		return debug
	}
	return false
}

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDContextKey).(string)
	errors.Assert(ok && requestID != "", "RequestIDMiddleware not installed")
	return requestID

}

func GetBody(ctx context.Context) interface{} {
	req := ctx.Value(bodyContextKey)
	errors.Assert(req != nil, "BodyMiddleware not installed")
	return req
}
