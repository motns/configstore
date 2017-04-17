#!/usr/bin/env bash

mkdir -p bin/linux
mkdir -p bin/darwin
mkdir -p bin/windows

GOOS=darwin GOARCH=amd64 go build -o bin/darwin/configstore
GOOS=linux GOARCH=amd64 go build -o bin/linux/configstore
GOOS=windows GOARCH=amd64 go build -o bin/windows/configstore
