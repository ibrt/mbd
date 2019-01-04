package mbd

import (
	"context"
	"strings"

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
	requestContextContexKey
)

// Provider is a function that populates the Context with some values.
type Provider func(context.Context) context.Context

// StaticProvider returns a Provider that adds the given k/v pair to the context.
func StaticProvider(k, v interface{}) Provider {
	return func(ctx context.Context) context.Context { // Provider
		return context.WithValue(ctx, k, v)
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
	return ctx.Value(requestContextContexKey).(*RequestContext)
}

type singleGet struct {
	original  map[string]string
	lowercase map[string]string
}

func newSingleGet(original map[string]string) *singleGet {
	lowercase := make(map[string]string, len(original))
	for k, v := range original {
		lowercase[strings.ToLower(k)] = v
	}
	return &singleGet{
		original:  original,
		lowercase: lowercase,
	}
}

// Map returns the original values as a map.
func (s *singleGet) Map() map[string]string {
	return s.original
}

// Get returns the value corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns "".
func (s *singleGet) Get(k string) string {
	return s.lowercase[strings.ToLower(k)]
}

type multiGet struct {
	original       map[string]string
	originalMulti  map[string][]string
	lowercaseMulti map[string][]string
}

func newMultiGet(original map[string]string, originalMulti map[string][]string) *multiGet {
	lowercaseMulti := make(map[string][]string, len(originalMulti))
	for k, v := range originalMulti {
		lowercaseMulti[strings.ToLower(k)] = v
	}
	return &multiGet{
		original:       original,
		originalMulti:  originalMulti,
		lowercaseMulti: lowercaseMulti,
	}
}

// Map returns the original values as single-value map.
func (m *multiGet) Map() map[string]string {
	return m.original
}

// MapMulti returns the original values as a multi-value map.
func (m *multiGet) MapMulti() map[string][]string {
	return m.originalMulti
}

// Get returns a single value corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns "". If the key has multiple values, it returns the last one.
func (m *multiGet) Get(k string) string {
	v := m.GetMulti(k)
	if len(v) == 0 {
		return ""
	}
	return v[len(v)-1]
}

// Get returns the values corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns []string{}. If the key has multiple values, it returns all of them.
func (m *multiGet) GetMulti(k string) []string {
	v, ok := m.lowercaseMulti[strings.ToLower(k)]
	if !ok {
		return []string{}
	}
	return v
}

func populateContext(ctx context.Context, debug Debug, in *events.APIGatewayProxyRequest) context.Context {
	ctx = context.WithValue(ctx, debugContextKey, debug)
	ctx = context.WithValue(ctx, pathContextKey, newPath(in))
	ctx = context.WithValue(ctx, headersContextKey, &Headers{newMultiGet(in.Headers, in.MultiValueHeaders)})
	ctx = context.WithValue(ctx, queryStringContextKey, &QueryString{newMultiGet(in.QueryStringParameters, in.MultiValueQueryStringParameters)})
	ctx = context.WithValue(ctx, pathParametersContextKey, &PathParameters{newSingleGet(in.PathParameters)})
	ctx = context.WithValue(ctx, stageVariablesContextKey, &StageVariables{newSingleGet(in.StageVariables)})
	ctx = context.WithValue(ctx, requestContextContexKey, &in.RequestContext)

	return ctx
}
