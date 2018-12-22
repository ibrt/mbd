package mbd

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// Handler describes a function invoked via the API Gateway in proxy mode.
type Handler func(ctx context.Context, in events.APIGatewayProxyRequest) events.APIGatewayProxyResponse

// Middleware describes Handler wrapper that performs some computation and calls the given next Handler.
type Middleware func(next Handler) Handler

// ContextProvider is a function that populates the Context with some values.
type ContextProvider func(ctx context.Context) context.Context

// ErrorAdapter describes a function that converts an error or recovered panic value into an ApiGatewayProxyResponse.
// In most cases the interface{} parameter will implement error, although it might be anything recovered from panic.
type ErrorAdapter func(ctx context.Context, in events.APIGatewayProxyRequest, r interface{}) events.APIGatewayProxyResponse

// BodyAdapter describes a function that parses the body found in an ApiGatewayProxyRequest, returning it as interface{}.
type BodyAdapter func(ctx context.Context, in events.APIGatewayProxyRequest) interface{}

// ResponseOption describes a mutation on a response.
type ResponseOption func(*events.APIGatewayProxyResponse)
