// +build remote

package mbd_test

import (
	"testing"

	"github.com/ibrt/mbd/internal/testrunner"
)

func TestRemote(t *testing.T) {
	testrunner.RunTests(t, testrunner.NewRemoteRunner())
}
