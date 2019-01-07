package mbd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingleGet(t *testing.T) {
	singleGet := newSingleGet(map[string]string{})
	require.Equal(t, map[string]string{}, singleGet.Map())
	require.Equal(t, "", singleGet.Get("key"))

	singleGet = newSingleGet(map[string]string{"Key": "Value"})
	require.Equal(t, map[string]string{"Key": "Value"}, singleGet.Map())
	require.Equal(t, "Value", singleGet.Get("Key"))
	require.Equal(t, "Value", singleGet.Get("key"))
}

func TestMultiGet(t *testing.T) {
	multiGet := newMultiGet(map[string]string{}, map[string][]string{})
	require.Equal(t, map[string]string{}, multiGet.Map())
	require.Equal(t, map[string][]string{}, multiGet.MapMulti())
	require.Equal(t, "", multiGet.Get("Key"))
	require.Equal(t, "", multiGet.Get("key"))
	require.Equal(t, []string{}, multiGet.GetMulti("Key"))
	require.Equal(t, []string{}, multiGet.GetMulti("key"))

	multiGet = newMultiGet(map[string]string{"Key": "V2"}, map[string][]string{"Key": {"V1", "V2"}})
	require.Equal(t, map[string]string{"Key": "V2"}, multiGet.Map())
	require.Equal(t, map[string][]string{"Key": {"V1", "V2"}}, multiGet.MapMulti())
	require.Equal(t, "V2", multiGet.Get("Key"))
	require.Equal(t, "V2", multiGet.Get("key"))
	require.Equal(t, []string{"V1", "V2"}, multiGet.GetMulti("Key"))
	require.Equal(t, []string{"V1", "V2"}, multiGet.GetMulti("key"))
}
