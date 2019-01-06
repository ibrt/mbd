package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/ibrt/errors"
	"github.com/stretchr/testify/require"
)

// RunRemoteTests runs all Function tests remotely on AWS Lambda.
func RunRemoteTests(t *testing.T) {
	printHeader("SETUP")

	dir := filepath.Join(getTestsDir(t), ".remote")
	fmt.Printf("Generating artifacts at '%v'...\n", dir)
	require.NoError(t, os.RemoveAll(dir))
	require.NoError(t, os.MkdirAll(dir, 0777))
	require.NoError(t, os.Chdir(dir))
	defer func() {
		fmt.Println("Removing artifacts...")
		errors.Ignore(os.RemoveAll(dir))
	}()

	fmt.Println("Writing templates...")
	writeTemplate(t, "serverless.yml", serverlessTpl, testFunctions)
	for functionName := range testFunctions {
		functionDir := filepath.Join("functions", functionName)
		require.NoError(t, os.MkdirAll(functionDir, 0777))
		writeTemplate(t, filepath.Join(functionDir, "main.go"), mainTpl, functionName)
	}

	for functionName := range testFunctions {
		fmt.Printf("Compiling '%v'...\n", functionName)
		outputFile := filepath.Join("build", functionName)
		mainFile := filepath.Join("functions", functionName, "main.go")

		runCommand(t, exec.Command("go", "build", "-ldflags=-s -w", "-o", outputFile, mainFile), map[string]string{
			"GOOS": "linux",
		})
	}

	fmt.Println("Deploying...")
	slsOut := runCommand(t, exec.Command("sls", "deploy"), map[string]string{
		"AWS_PROFILE":        "ibrt",
		"AWS_DEFAULT_REGION": "us-east-1",
	})
	defer func() {
		fmt.Println("Removing deployment...")
		runCommand(t, exec.Command("sls", "remove"), map[string]string{
			"AWS_PROFILE":        "ibrt",
			"AWS_DEFAULT_REGION": "us-east-1",
		})
	}()

	baseURL := regexp.MustCompile(`https://[^\.]+\.execute-api.[^\.]+\.amazonaws\.com/[^/]+/`).FindString(slsOut)

	for _, c := range testCases {
		t.Run(c.FunctionName+"_"+c.TestName, func(t *testing.T) {
			dumpValue("REQUEST", c.Request)

			var body io.Reader
			if c.Request != nil {
				body = strings.NewReader(adaptRequest(t, c.Request))
			}

			resp, err := http.Post(baseURL+"/"+c.FunctionName, "application/json; charset=utf-8", body)
			require.NoError(t, err)
			out := parseOutputResponse(t, resp.Body)
			dumpValue("OUTPUT", out)
			assertTestCase(t, c, out)
		})
	}

	// TODO(ibrt): Env checks.

	defer printHeader("TEARDOWN")
}

// RemoteTestingTProvider provides a require.TestingT suitable to use in Lambda functions.
func RemoteTestingTProvider(ctx context.Context) context.Context {
	return context.WithValue(ctx, testingContextKey, &remoteTestingT{
		mu: &sync.Mutex{},
	})
}

var serverlessTpl = template.Must(template.New("serverless.yml").Parse(`service: mbd

provider:
  stage: test
  name: aws
  runtime: go1.x
  memorySize: 1024
  timeout: 30
  logRetentionInDays: 7
  endpointType: regional

functions:{{ range $functionName, $testFunction := . }}
  {{$functionName}}:
    handler: build/{{$functionName}}
    events:
      - http:
          path: {{$functionName}}
          method: post
{{end}}
package:
  exclude:
    - ./**
  include:
    - ./build/**
`))

var mainTpl = template.Must(template.New("main.go").Parse(`// +build remote

package main

import "github.com/ibrt/mbd/internal/tests"

func main() {
	tests.
		GetTestFunction("{{.}}").
		AddProviders(tests.RemoteTestingTProvider).
		Start()
}
`))

type remoteTestingT struct {
	mu  *sync.Mutex
	err error
}

// Errorf implements require.TestingT
func (t *remoteTestingT) Errorf(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.err = errors.Append(t.err, errors.Errorf(fmt.Sprintf(format, args...), errors.Skip(1)))
}

// FailNow implements require.TestingT
func (t *remoteTestingT) FailNow() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.err == nil {
		t.err = errors.Errorf("test failed", errors.Skip(1))
	}
	errors.MustWrap(t.err)
}
