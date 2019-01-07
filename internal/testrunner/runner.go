package testrunner

import (
	"fmt"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibrt/mbd/internal/testcases"
)

// Runner describes a test runner that can run a TestCase locally in-memory, or against a test remote Lambda deployment.
type Runner interface {
	Setup(t *testing.T)
	Teardown(t *testing.T)
	RunTest(t *testing.T, c *testcases.TestCase)
}

type baseRunner struct {
	// intentionally empty
}

// Setup implements Runner.
func (r *baseRunner) Setup(t *testing.T) {
	// intentionally empty
}

// Teardown implements Runner.
func (r *baseRunner) Teardown(t *testing.T) {
	// intentionally empty
}

func (r *baseRunner) printHeader(header string) {
	fmt.Println("┌" + strings.Repeat("─", len(header)+2) + "┐")
	fmt.Println("│", strings.ToUpper(header), "│")
	fmt.Println("└" + strings.Repeat("─", len(header)+2) + "┘")
}

func (r *baseRunner) printValue(title string, value interface{}) {
	r.printHeader(title)
	spew.Dump(value)
}

// RunTests runs all TestCases using the given Runner.
func RunTests(t *testing.T, r Runner) {
	r.Setup(t)
	defer r.Teardown(t)

	for _, c := range testcases.GetTestCases() {
		t.Run(c.Name, func(t *testing.T) {
			r.RunTest(t, c)
		})
	}
}
