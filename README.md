# mbd [![Build Status](https://travis-ci.org/ibrt/mbd.svg?branch=master)](https://travis-ci.org/ibrt/mbd) [![Go Report Card](https://goreportcard.com/badge/github.com/ibrt/mbd)](https://goreportcard.com/report/github.com/ibrt/mbd) [![Test Coverage](https://codecov.io/gh/ibrt/mbd/branch/master/graph/badge.svg)](https://codecov.io/gh/ibrt/mbd) [![Go Docs](https://godoc.org/github.com/ibrt/mbd?status.svg)](http://godoc.org/github.com/ibrt/mbd)

MBD is a Go framework for AWS Lambda, currently focused on the API Gateway integration. It wraps the official AWS Lambda framework and provides a simplified [handler signature](https://godoc.org/github.com/ibrt/mbd#Handler), case-insensitive access to [headers](https://godoc.org/github.com/ibrt/mbd#Headers), [query string parameters](https://godoc.org/github.com/ibrt/mbd#QueryString), and other types of request metadata. It interoperates well with the [ibrt/errors](https://github.com/ibrt/errors) package to return configurable, rich [error responses](https://godoc.org/github.com/ibrt/mbd#ErrorResponse), complete with debug information.

#### Basic Usage

```go
package main

import (
  "github.com/ibrt/mbd"
)

type echoRequest struct {
  Value string `json:"value"`
}

type echoResponse struct {
  Value string `json:"value"`
} 

func main() {
  mbd.NewFunction(echoRequest{}, echoHandler).
    SetDebug(true).
    Start()
}

func(ctx context.Context, req interface{}) (interface{}, error) {
  echoRequest := req.(*echoRequest)
  return &echoResponse{Value: echoRequest.Value}  
}
```
