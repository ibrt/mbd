// +build e2e

package main

import (
	"context"
	"reflect"

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
	m := mbd.Middlewares(
		mbd.ContextMiddleware(mbd.StaticContextProvider("key", "value")),
		mbd.DebugMiddleware(true),
		mbd.RequestIDMiddleware(),
		mbd.PanicMiddleware(mbd.DefaultErrorAdapter),
		mbd.BodyMiddleware(mbd.JSONBodyAdapter(reflect.TypeOf(Request{}), mbd.JSONDecoderDisallowUnknownFields())),
		mbd.ValidatorMiddleware(validator))

	mbd.Start(m(handler))
}

func validator(ctx context.Context, _ events.APIGatewayProxyRequest) {
	if req := mbd.GetBody(ctx).(*Request); req.Behavior != "ok" && req.Behavior != "error" && req.Behavior != "panic" {
		panic(errors.Errorf("missing or unknown behavior '%v'", req.Behavior, errors.HTTPStatusBadRequest))
	}
}

func handler(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	errors.Assert(ctx.Value("key").(string) == "value", "missing context key")
	req := mbd.GetBody(ctx).(*Request)

	switch req.Behavior {
	case "ok":
		return mbd.NewResponse(ctx, mbd.JSONBody(&Response{Value: req.Value}))
	case "error":
		return mbd.DefaultErrorAdapter(ctx, in, errors.Errorf("test error", errors.HTTPStatus(req.Value)))
	case "panic":
		panic(errors.Errorf("test error", errors.HTTPStatus(req.Value)))
	default:
		panic("unexpected")
	}
}
