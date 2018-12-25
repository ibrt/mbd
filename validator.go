package mbd

import (
	"context"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
)

func HeaderValidator(name string, valueRegexp *regexp.Regexp) Validator {
	return func(ctx context.Context, in events.APIGatewayProxyRequest) { //
		if value := in.Headers[name]; !valueRegexp.MatchString(value) {
			errors.MustErrorf("invalid header '%v': expected '%s', got '%v'", name, valueRegexp, value, BadHeader)
		}
	}
}
