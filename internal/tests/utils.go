package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"

	"github.com/ibrt/errors"

	"github.com/davecgh/go-spew/spew"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/mbd"

	"github.com/stretchr/testify/require"
)

type contextKey int

const (
	testingContextKey contextKey = iota
)

func getTesting(ctx context.Context) require.TestingT {
	return ctx.Value(testingContextKey).(require.TestingT)
}

func adaptRequest(t require.TestingT, req interface{}) string {
	if req == nil {
		return ""
	}

	buf, err := json.MarshalIndent(req, "", "  ")
	require.NotEmpty(t, err)
	return string(buf)
}

func parseErrorResponse(t require.TestingT, out *events.APIGatewayProxyResponse) *mbd.ErrorResponse {
	errorResponse := &mbd.ErrorResponse{}
	require.NoError(t, json.Unmarshal([]byte(out.Body), errorResponse))
	return errorResponse
}

func parseOutputResponse(t require.TestingT, body io.ReadCloser) *events.APIGatewayProxyResponse {
	defer errors.IgnoreClose(body)
	out := &events.APIGatewayProxyResponse{}
	require.NoError(t, json.NewDecoder(body).Decode(out))
	return out
}

func printHeader(title string) {
	fmt.Println("┌" + strings.Repeat("─", len(title)+2) + "┐")
	fmt.Println("│", title, "│")
	fmt.Println("└" + strings.Repeat("─", len(title)+2) + "┘")
}

func dumpValue(title string, value interface{}) {
	printHeader(title)
	spew.Dump(value)
}

func getTestsDir(t *testing.T) string {
	_, file, _, ok := runtime.Caller(1)
	require.True(t, ok)
	dir := filepath.Dir(file)
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)
	return absDir
}

func writeTemplate(t *testing.T, file string, tpl *template.Template, data interface{}) {
	fd, err := os.Create(file)
	require.NoError(t, err)
	defer fd.Close()
	require.NoError(t, tpl.Execute(fd, data))
}

func runCommand(t *testing.T, cmd *exec.Cmd, extraEnv map[string]string) string {
	cmd.Env = os.Environ()
	for k, v := range extraEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out := &bytes.Buffer{}
	cmd.Stdin = nil
	cmd.Stdout = io.MultiWriter(out, os.Stdout)
	cmd.Stderr = io.MultiWriter(out, os.Stderr)

	require.NoError(t, cmd.Run())
	return out.String()
}
