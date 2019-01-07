package mbd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingletonProvider(t *testing.T) {
	ctx := SingletonProvider("key", "value")(context.Background())
	require.Equal(t, "value", ctx.Value("key"))
}

func TestRequestProvider(t *testing.T) {
	var i int
	provider := RequestProvider("key", func() interface{} { i++; return i })

	ctx := provider(context.Background())
	require.Equal(t, 1, ctx.Value("key"))

	ctx = provider(context.Background())
	require.Equal(t, 2, ctx.Value("key"))
}
