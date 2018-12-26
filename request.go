package mbd

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
)

// JSONBodyAdapter returns a BodyAdapter that parses a JSON body into a struct of the given type.
// Note that reqType.Kind() must equal reflect.Struct, but the value returned by the adapter will be a pointer to it.
func JSONBodyAdapter(reqType reflect.Type, options ...JSONBodyAdapterOption) BodyAdapter {
	errors.Assert(reqType.Kind() == reflect.Struct, "request type kind must be struct")

	return func(ctx context.Context, in events.APIGatewayProxyRequest) interface{} { // BodyAdapter
		errors.Assert(!in.IsBase64Encoded, "bad IsBase64Encoded: expected 'false', got 'true'", BadEncoding)

		req := reflect.New(reqType).Interface()
		dec := json.NewDecoder(strings.NewReader(in.Body))

		for _, option := range options {
			option(ctx, in, dec)
		}

		errors.MaybeMustWrap(dec.Decode(req), BadJSON)
		return req
	}
}

// JSONDecoderDisallowUnknownFields returns a JSONBodyAdapterOption that calls DisallowUnknownFields().
func JSONDecoderDisallowUnknownFields() JSONBodyAdapterOption {
	return func(ctx context.Context, in events.APIGatewayProxyRequest, dec *json.Decoder) { // JSONBodyAdapterOption
		dec.DisallowUnknownFields()
	}
}

// JSONDecoderUseNumber returns a JSONBodyAdapterOption that calls UseNumber().
func JSONDecoderUseNumber() JSONBodyAdapterOption {
	return func(ctx context.Context, in events.APIGatewayProxyRequest, dec *json.Decoder) { // JSONBodyAdapterOption
		dec.UseNumber()
	}
}
