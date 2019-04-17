package mbd

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"strings"

	"github.com/gorilla/schema"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ibrt/errors"
)

var (
	defaultDecoder = schema.NewDecoder()
)

// JSONRequestParser returns a RequestParser for JSON requests.
func JSONRequestParser() RequestParser {
	return func(_ context.Context, reqType reflect.Type, in *events.APIGatewayProxyRequest) (interface{}, error) { // RequestParser
		if in.IsBase64Encoded {
			return nil, errors.Errorf("invalid IsBase64Encoded: expected 'false', got 'true'", invalidBody)
		}

		if reqType == noRequestBody {
			if in.Body != "" {
				return nil, errors.Errorf("unexpected Body", unexpectedBody)
			}

			return nil, nil
		}

		req := reflect.New(reqType).Interface()

		dec := json.NewDecoder(strings.NewReader(in.Body))
		dec.DisallowUnknownFields()
		dec.UseNumber()

		if err := dec.Decode(req); err != nil {
			return nil, errors.Wrap(err, errors.Prefix("invalid Body"), invalidBody)
		}

		return req, nil
	}
}

// FormRequestParser returns a RequestParser for form encoded requests. It uses gorilla/schema to map values to a struct.
func FormRequestParser() RequestParser {
	return func(_ context.Context, reqType reflect.Type, in *events.APIGatewayProxyRequest) (interface{}, error) { // RequestParser
		if in.IsBase64Encoded {
			return nil, errors.Errorf("invalid IsBase64Encoded: expected 'false', got 'true'", invalidBody)
		}

		if reqType == noRequestBody {
			if in.Body != "" {
				return nil, errors.Errorf("unexpected Body", unexpectedBody)
			}

			return nil, nil
		}

		q, err := url.ParseQuery(in.Body)
		if err != nil {
			return nil, errors.Wrap(err, errors.Prefix("invalid Body"), invalidBody)
		}

		req := reflect.New(reqType).Interface()
		if err := defaultDecoder.Decode(req, q); err != nil {
			return nil, errors.Wrap(err, errors.Prefix("invalid Body"), invalidBody)
		}

		return req, nil
	}
}
