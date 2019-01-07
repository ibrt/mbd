package mbd_test

import (
	"testing"

	"github.com/ibrt/mbd/internal/testrunner"
)

func TestLocal(t *testing.T) {
	testrunner.RunTests(t, testrunner.NewLocalRunner())
}
