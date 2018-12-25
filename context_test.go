package mbd_test

import (
	"context"
	"testing"

	"github.com/ibrt/mbd"
	"github.com/stretchr/testify/require"
)

func TestDebug(t *testing.T) {
	require.False(t, mbd.GetDebug(context.Background()))
	require.False(t, mbd.GetDebug(mbd.WithDebug(context.Background(), false)))
	require.True(t, mbd.GetDebug(mbd.WithDebug(context.Background(), true)))
}

func TestRequestID(t *testing.T) {
	require.Equal(t, "", mbd.GetRequestID(context.Background()))
	require.Equal(t, "request-id", mbd.GetRequestID(mbd.WithRequestID(context.Background(), "request-id")))
}

func TestBody(t *testing.T) {
	require.Equal(t, "body", mbd.GetBody(mbd.WithBody(context.Background(), "body")))
	require.Panics(t, func() { mbd.GetBody(context.Background()) })
}
