package testcontext

import (
	"context"

	"github.com/stretchr/testify/require"
)

type contextKey int

const (
	testingContextKey contextKey = iota
)

// WithTestingT adds the given require.TestingT to context.
func WithTestingT(ctx context.Context, t require.TestingT) context.Context {
	return context.WithValue(ctx, testingContextKey, t)
}

// GetTestingT returns the require.TestingT stored in context.
func GetTestingT(ctx context.Context) require.TestingT {
	return ctx.Value(testingContextKey).(require.TestingT)
}
