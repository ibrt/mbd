package mbd

import (
	"context"
	"net/http"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ibrt/errors"
)

// Checker implements a check on a request, usually for authentication or validation.
// It can optionally also add values to the context.
type Checker func(ctx context.Context, in *events.APIGatewayProxyRequest, req interface{}) (context.Context, error)

// Handler implements a Lambda function handler.
type Handler func(ctx context.Context, req interface{}) (resp interface{}, err error)

// RequestParser describes a custom request parser. The default is JSON.
type RequestParser func(context.Context, reflect.Type, *events.APIGatewayProxyRequest) (interface{}, error)

// Function sets up a Lambda function handler.
type Function struct {
	reqType   reflect.Type
	reqParser RequestParser
	handler   Handler
	debug     Debug
	providers []Provider
	checkers  []Checker
}

// NewFunction initializes a new Function.
func NewFunction(reqTemplate interface{}, handler Handler) *Function {
	reqType := noRequestBody
	if reqTemplate != nil {
		reqType = reflect.TypeOf(reqTemplate)
	}

	errors.Assert(handler != nil, "handler must not be nil")
	errors.Assert(reqType.Kind() == reflect.Struct, "reqTemplate must be nil or struct value")

	return &Function{
		reqType:   reqType,
		reqParser: JSONRequestParser(),
		handler:   handler,
		debug:     false,
		providers: make([]Provider, 0),
	}
}

// SetDebug enables or disables additional debug information. Default is disabled.
func (e *Function) SetDebug(debug Debug) *Function {
	e.debug = debug
	return e
}

// SetRequestParser sets a custom RequestParser. Default is JSON.
func (e *Function) SetRequestParser(reqParser RequestParser) *Function {
	e.reqParser = reqParser
	return e
}

// AddProviders adds one or more Provider(s) to the Function.
func (e *Function) AddProviders(providers ...Provider) *Function {
	e.providers = append(e.providers, providers...)
	return e
}

// AddCheckers adds one or more Checker(s) to the Function.
func (e *Function) AddCheckers(checkers ...Checker) *Function {
	e.checkers = append(e.checkers, checkers...)
	return e
}

// Handler provides a handler function suitable for lambda.Start().
func (e *Function) Handler(ctx context.Context, in events.APIGatewayProxyRequest) (out events.APIGatewayProxyResponse, _ error) {
	ctx = populateContext(ctx, e.debug, &in)

	defer func() {
		if err := errors.MaybeWrapRecover(recover()); err != nil {
			out = *adaptError(ctx, err)
		}
	}()

	for _, provider := range e.providers {
		ctx = provider(ctx)
	}

	req, err := e.reqParser(ctx, e.reqType, &in)
	if err != nil {
		return *adaptError(ctx, err), nil
	}

	for _, checker := range e.checkers {
		newCtx, err := checker(ctx, &in, req)
		if err != nil {
			return *adaptError(ctx, err), nil
		}
		if newCtx != nil {
			ctx = newCtx
		}
	}

	resp, err := e.handler(ctx, req)
	if err != nil {
		return *adaptError(ctx, err), nil
	}

	return *adaptResponse(ctx, http.StatusOK, resp), nil
}

// Start invokes lambda.Start() passing the Function handler as argument.
func (e *Function) Start() {
	lambda.Start(e.Handler)
}
