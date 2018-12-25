package mbd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/ibrt/mbd"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	r                     interface{}
	expectedStatus        int
	expectedPublicMessage string
	expectedError         string
}

var testCases = []testCase{
	{"test error", http.StatusInternalServerError, "internal-server-error", "test error"},
	{100, http.StatusInternalServerError, "internal-server-error", "100"},
	{fmt.Errorf("test error"), http.StatusInternalServerError, "internal-server-error", "test error"},
	{errors.Errorf("test error"), http.StatusInternalServerError, "internal-server-error", "test error"},
	{errors.Errorf("test error", mbd.BadEncoding), http.StatusBadRequest, "bad-encoding", "test error"},
}

func TestDefaultErrorAdapter(t *testing.T) {
	for _, testCase := range testCases {
		for _, debug := range []bool{false, true} {
			for _, requestID := range []string{"", "request-id"} {
				t.Run(fmt.Sprintf("%v_%v_%v", testCase.r, debug, requestID), func(t *testing.T) {
					ctx := mbd.WithDebug(context.Background(), debug)
					ctx = mbd.WithRequestID(ctx, requestID)

					out := mbd.DefaultErrorAdapter(ctx, events.APIGatewayProxyRequest{}, testCase.r)

					resp := &mbd.DefaultErrorResponse{}
					dec := json.NewDecoder(strings.NewReader(out.Body))
					dec.DisallowUnknownFields()
					require.NoError(t, dec.Decode(resp))

					require.Equal(t, testCase.expectedStatus, out.StatusCode)
					require.Equal(t, "application/json; charset=utf-8", out.Headers["Content-Type"])
					require.Len(t, out.Headers, 1)
					require.False(t, out.IsBase64Encoded)

					require.Equal(t, testCase.expectedStatus, resp.StatusCode)
					require.Equal(t, testCase.expectedPublicMessage, resp.PublicMessage)
					require.Equal(t, requestID, resp.RequestID)

					if debug {
						require.Equal(t, testCase.expectedError, resp.Error)
						require.NotNil(t, resp.StackTrace)
					} else {
						require.Equal(t, "", resp.Error)
						require.Nil(t, resp.StackTrace)
					}
				})
			}
		}
	}
}
