#!/usr/bin/env bash

set -e
cd "$(dirname "${BASH_SOURCE[0]}")"
env GOOS=linux go build -tags=e2e -ldflags="-s -w" -o build/e2e e2e.go
sls deploy --verbose