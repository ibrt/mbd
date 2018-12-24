package mbd

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// Handler describes a function invoked via the API Gateway in proxy mode.
type Handler func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse

// Middleware describes Handler wrapper that performs some computation and calls the given next Handler.
type Middleware func(next Handler) Handler

// ContextProvider is a function that populates the Context with some values. It is used by ContextMiddleware to enrich
// the context passed to downstream middlewares.
type ContextProvider func(ctx context.Context) context.Context

// ErrorAdapter describes a function that converts an error or recovered panic value into an ApiGatewayProxyResponse.
// In most cases the interface{} parameter will implement error, although it might be anything recovered from panic.
// It is used by PanicMiddleware to convert a panic value into a response.
type ErrorAdapter func(ctx context.Context, in events.APIGatewayProxyRequest, r interface{}) events.APIGatewayProxyResponse

// BodyAdapter describes a function that converts the body in an ApiGatewayProxyRequest into an interface{}. It will
// panic if parsing fails. For JSON requests it will usually return a struct, while for binary requests it will usually
// return a io.Reader. It is used by BodyMiddleware to parse the request body.
type BodyAdapter func(ctx context.Context, in events.APIGatewayProxyRequest) interface{}

// Validator describes a function that validates a request. It will panic if validation fails. It is used by
// ValidatorMiddleware. Note that if ValidatorMiddleware is installed after BodyMiddleware, the context will contain the
// parsed request body.
type Validator func(ctx context.Context, in events.APIGatewayProxyRequest)

// ResponseOption describes an option for NewResponse.
type ResponseOption func(ctx context.Context, out *events.APIGatewayProxyResponse)

// JSONBodyAdapterOption describes an option for JSONBodyAdapter.
type JSONBodyAdapterOption func(ctx context.Context, in events.APIGatewayProxyRequest, dec *json.Decoder)
