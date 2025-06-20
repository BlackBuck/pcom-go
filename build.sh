#!/bin/bash

set -eo pipefail

go clean
go clean -testcache
go test ./...
go build ./...