package mbd

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultPublicMessage(t *testing.T) {
	require.Equal(t, "i-am-a-teapot", getDefaultPublicMessage(http.StatusTeapot))
	require.Equal(t, "internal-server-error", getDefaultPublicMessage(http.StatusInternalServerError))
	require.Equal(t, "unknown", getDefaultPublicMessage(1))
}
