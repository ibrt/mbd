// +build remote

package mbd_test

import (
	"testing"

	"github.com/ibrt/mbd/internal/tests"
)

func TestRemote(t *testing.T) {
	tests.RunRemoteTests(t)
}
