// +build e2e

package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
	"github.com/ibrt/mbd"
)

type Request struct {
	Behavior string `json:"behavior"`
	Value    int    `json:"value"`
}

type Response struct {
	Value int `json:"value"`
}

func main() {
	mbd.NewFunction(Request{}, handler).
		SetDebug(true).
		AddProviders(mbd.StaticProvider("key", "value")).
		AddCheckers(checker).
		Start()
}

func checker(_ context.Context, _ *events.APIGatewayProxyRequest, req interface{}) error {
	if req := req.(*Request); req.Behavior != "ok" && req.Behavior != "error" && req.Behavior != "panic" {
		return errors.Errorf("missing or unknown behavior '%v'", req.Behavior, errors.HTTPStatusBadRequest)
	}
	return nil
}

func handler(ctx context.Context, req interface{}) (interface{}, error) {
	errors.Assert(ctx.Value("key").(string) == "value", "missing context key")
	request := req.(*Request)

	switch request.Behavior {
	case "ok":
		return &Response{Value: request.Value}, nil
	case "error":
		return nil, errors.Errorf("test error", errors.HTTPStatus(request.Value))
	case "panic":
		panic(errors.Errorf("test error", errors.HTTPStatus(request.Value)))
	default:
		panic("unexpected")
	}
}
