package mbd

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type contextKey int

const (
	debugContextKey contextKey = iota
	pathContextKey
	headersContextKey
	queryStringContextKey
	pathParametersContextKey
	stageVariablesContextKey
	requestContextContextKey
)

func populateContext(ctx context.Context, debug Debug, in *events.APIGatewayProxyRequest) context.Context {
	ctx = context.WithValue(ctx, debugContextKey, debug)
	ctx = context.WithValue(ctx, pathContextKey, newPath(in))
	ctx = context.WithValue(ctx, headersContextKey, &Headers{newMultiGet(in.Headers, in.MultiValueHeaders)})
	ctx = context.WithValue(ctx, queryStringContextKey, &QueryString{newMultiGet(in.QueryStringParameters, in.MultiValueQueryStringParameters)})
	ctx = context.WithValue(ctx, pathParametersContextKey, &PathParameters{newSingleGet(in.PathParameters)})
	ctx = context.WithValue(ctx, stageVariablesContextKey, &StageVariables{newSingleGet(in.StageVariables)})
	ctx = context.WithValue(ctx, requestContextContextKey, &in.RequestContext)
	return ctx
}

// Provider is a function that populates the Context with some values.
type Provider func(ctx context.Context) context.Context

// SingletonProvider returns a Provider that adds the given k/v pair to the context.
func SingletonProvider(k, v interface{}) Provider {
	return func(ctx context.Context) context.Context { // Provider
		return context.WithValue(ctx, k, v)
	}
}

// RequestProvider returns a Provider that generates a new k/v pair for every requests, obtaining v from f.
func RequestProvider(k interface{}, f func() interface{}) Provider {
	return func(ctx context.Context) context.Context { // Provider
		return context.WithValue(ctx, k, f())
	}
}

// Debug describes whether the framework should run in debug mode.
type Debug = bool

// GetDebug returns the debug flag stored in context. If missing, it returns false.
func GetDebug(ctx context.Context) Debug {
	if debug, ok := ctx.Value(debugContextKey).(Debug); ok {
		return debug
	}
	return false
}

// Path provides some metadata about the original request path and method.
type Path struct {
	Resource string
	Path     string
	Method   string
}

func newPath(in *events.APIGatewayProxyRequest) *Path {
	return &Path{
		Resource: in.Resource,
		Path:     in.Path,
		Method:   in.HTTPMethod,
	}
}

// GetPath returns the Path stored in context.
func GetPath(ctx context.Context) *Path {
	return ctx.Value(pathContextKey).(*Path)
}

// Headers provides access to request headers, as original map or case-insensitive getters.
type Headers struct {
	*multiGet
}

// GetHeaders returns the Headers stored in context.
func GetHeaders(ctx context.Context) *Headers {
	return ctx.Value(headersContextKey).(*Headers)
}

// QueryString provides access to query string parameters, as original map or case-insensitive getters.
type QueryString struct {
	*multiGet
}

// GetQueryString returns the QueryString stored in context.
func GetQueryString(ctx context.Context) *QueryString {
	return ctx.Value(queryStringContextKey).(*QueryString)
}

// PathParameters provides access to path parameters, as original map or case-insensitive getter.
type PathParameters struct {
	*singleGet
}

// GetPathParameters returns the PathParameters stored in context.
func GetPathParameters(ctx context.Context) *PathParameters {
	return ctx.Value(pathParametersContextKey).(*PathParameters)
}

// StageVariables provides access to stage variables, as original map or case-insensitive getter.
type StageVariables struct {
	*singleGet
}

// GetStageVariables returns the GetStageVariables stored in context.
func GetStageVariables(ctx context.Context) *StageVariables {
	return ctx.Value(stageVariablesContextKey).(*StageVariables)
}

// RequestContext is an alias for events.APIGatewayProxyRequestContext.
type RequestContext = events.APIGatewayProxyRequestContext

// GetRequestContext returns the RequestContext stored in context.
func GetRequestContext(ctx context.Context) *RequestContext {
	return ctx.Value(requestContextContextKey).(*RequestContext)
}
