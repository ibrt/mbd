package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ibrt/mbd"
)

// RunLocalTests runs all Function tests locally.
func RunLocalTests(t *testing.T) {
	for _, c := range testCases {
		t.Run(c.FunctionName+"_"+c.TestName, func(t *testing.T) {
			in := makeTestCaseInput(t, c)
			dumpValue("INPUT", in)
			dumpValue("REQUEST", c.Request)

			out, err := GetTestFunction(c.FunctionName).
				AddProviders(mbd.SingletonProvider(testingContextKey, t)).
				Handler(context.Background(), *in)

			require.NoError(t, err)
			dumpValue("OUTPUT", out)
			assertTestCase(t, c, &out)
		})
	}
}
