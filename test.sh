#!/usr/bin/env bash

set -e

echo "=== Running Go tests..."
go test -v client/**

echo ""
echo "=== Building latest version of configstore binary..."
mkdir -p bin/darwin/amd64
GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/configstore cmd/configstore/**

echo "=== Running BATS tests against configstore binary..."
bats configstore.bats

echo "=== Done!"