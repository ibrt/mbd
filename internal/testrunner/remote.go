package testrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"text/template"
	"unicode"

	"github.com/stretchr/testify/require"

	"github.com/ibrt/mbd/internal/testcontext"

	"github.com/ibrt/errors"

	"github.com/ibrt/mbd/internal/testcases"
)

const serverlessTpl = `
service: mbd

provider:
  stage: test
  name: aws
  runtime: go1.x
  memorySize: 1024
  timeout: 30
  logRetentionInDays: 7
  endpointType: regional

functions:{{ range $testCase := . }}
  {{$testCase.Name}}:
    handler: build/{{$testCase.Name}}
    events:
      - http:
          path: {{$testCase.Name}}
          method: post
{{end}}
package:
  exclude:
    - ./**
  include:
    - ./build/**
`

const mainTpl = `
// +build remote

package main

import (
	"github.com/ibrt/mbd"
	"github.com/ibrt/mbd/internal/testcases"
	"github.com/ibrt/mbd/internal/testrunner"
)

func main() {
	testCase := testcases.GetTestCase("{{.Name}}")

	mbd.NewFunction(testCase.ReqTemplate, testCase.Handler).
		SetDebug(true).
		AddProviders(testrunner.RemoteTestingTProvider).
		AddProviders(testCase.Providers...).
		AddCheckers(testCase.Checkers...).
		Start()
}
`

type remoteRunner struct {
	baseRunner
	dir     string
	baseURL string
}

// NewRemoteRunner returns a Runner that runs test cases against a test remote Lambda deployment.
func NewRemoteRunner() Runner {
	return &remoteRunner{}
}

func (r *remoteRunner) Setup(t *testing.T) {
	r.printHeader("Setup")
	r.setupDir(t)
	r.generateArtifacts(t)
	r.deploy(t)
}

func (r *remoteRunner) Teardown(t *testing.T) {
	r.printHeader("Teardown")

	if r.baseURL != "" {
		r.runCommand(t, exec.Command("sls", "remove"), nil)
	}
	if r.dir != "" {
		require.NoError(t, os.RemoveAll(r.dir))
	}
}

func (r *remoteRunner) RunTest(t *testing.T, c *testcases.TestCase) {
	httpResp, err := http.DefaultClient.Do(r.makeHTTPRequest(t, c.Name, c.Request))
	require.NoError(t, err)
	resp := r.parseHTTPResponse(t, c.RespTemplate, httpResp)
	c.Assertion(t, httpResp.StatusCode, httpResp.Header, resp)
}

func (r *remoteRunner) setupDir(t *testing.T) {
	fmt.Println("Creating .remote directory...")

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	fileDir := filepath.Dir(filepath.Dir(file))
	absDir, err := filepath.Abs(fileDir)
	require.NoError(t, err)
	r.dir = filepath.Join(absDir, ".remote")

	require.NoError(t, os.RemoveAll(r.dir))
	require.NoError(t, os.MkdirAll(r.dir, 0777))
	require.NoError(t, os.Chdir(r.dir))

	fmt.Printf("Created '%v'.\n", r.dir)
}

func (r *remoteRunner) generateArtifacts(t *testing.T) {
	fmt.Println("Generating templates...")
	r.writeTemplate(t, "serverless.yml", template.Must(template.New("").Parse(serverlessTpl)), testcases.TestCases)
	mainTpl := template.Must(template.New("").Parse(mainTpl))

	for _, c := range testcases.TestCases {
		require.NoError(t, os.MkdirAll(filepath.Join("functions", c.Name), 0777))
		r.writeTemplate(t, filepath.Join("functions", c.Name, "main.go"), mainTpl, c)
	}

	for _, c := range testcases.TestCases {
		fmt.Printf("Compiling '%v'...\n", c.Name)

		cmd := exec.Command("go", "build", "-ldflags=-s -w",
			"-o", filepath.Join("build", c.Name),
			filepath.Join("functions", c.Name, "main.go"))

		r.runCommand(t, cmd, map[string]string{"GOOS": "linux"})
	}
}

func (r *remoteRunner) deploy(t *testing.T) {
	fmt.Println("Checking environment variables...")
	require.NotEmpty(t, os.Getenv("AWS_PROFILE"), "Must set AWS_PROFILE.")
	require.NotEmpty(t, os.Getenv("AWS_DEFAULT_REGION"), "Must set AWS_DEFAULT_REGION.")

	fmt.Println("Deploying test functions...")
	slsOut := r.runCommand(t, exec.Command("sls", "deploy"), nil)
	r.baseURL = regexp.MustCompile(`https://[^\.]+\.execute-api.[^\.]+\.amazonaws\.com/[^/]+/`).FindString(slsOut)
	require.NotEmpty(t, r.baseURL)
}

func (r *remoteRunner) makeHTTPRequest(t *testing.T, name string, req interface{}) *http.Request {
	var body io.Reader
	if req != nil {
		buf, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)
		body = bytes.NewReader(buf)
	}

	httpReq, err := http.NewRequest("POST", r.baseURL+name, body)
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	r.printHeader("Input")
	buf, err := httputil.DumpRequestOut(httpReq, true)
	require.NoError(t, err)
	fmt.Println(string(buf))
	r.printValue("Request", req)

	return httpReq
}

func (r *remoteRunner) parseHTTPResponse(t *testing.T, respTemplate interface{}, httpResp *http.Response) interface{} {
	r.printHeader("Output")
	buf, err := httputil.DumpResponse(httpResp, true)
	require.NoError(t, err)
	fmt.Println(string(buf))

	body, err := ioutil.ReadAll(httpResp.Body)
	defer errors.IgnoreClose(httpResp.Body)

	if respTemplate == nil {
		require.Empty(t, body)
		r.printValue("Response", nil)
		return nil
	}

	resp := reflect.New(reflect.TypeOf(respTemplate)).Interface()
	require.NoError(t, json.Unmarshal(body, resp))
	r.printValue("Response", resp)
	return resp
}

func (r *remoteRunner) writeTemplate(t *testing.T, file string, tpl *template.Template, data interface{}) {
	buf := &bytes.Buffer{}
	require.NoError(t, tpl.Execute(buf, data))
	out := bytes.TrimLeftFunc(buf.Bytes(), unicode.IsSpace)
	require.NoError(t, ioutil.WriteFile(file, out, 0666))
}

func (r *remoteRunner) runCommand(t *testing.T, cmd *exec.Cmd, extraEnv map[string]string) string {
	cmd.Env = os.Environ()
	for k, v := range extraEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	buf := &bytes.Buffer{}
	cmd.Stdin = nil
	cmd.Stdout = io.MultiWriter(buf, os.Stdout)
	cmd.Stderr = io.MultiWriter(buf, os.Stderr)

	require.NoError(t, cmd.Run())
	return buf.String()
}

type remoteTestingT struct {
	mu  *sync.Mutex
	err error
}

// RemoteTestingTProvider provides a require.TestingT suitable to use in Lambda functions.
func RemoteTestingTProvider(ctx context.Context) context.Context {
	return testcontext.WithTestingT(ctx, &remoteTestingT{})
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
