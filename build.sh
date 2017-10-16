#!/usr/bin/env bash

mkdir -p bin/linux
mkdir -p bin/darwin
mkdir -p bin/windows

GOOS=darwin GOARCH=amd64 go build -v -o bin/darwin/configstore cmd/configstore/**
GOOS=linux GOARCH=amd64 go build -v -o bin/linux/configstore cmd/configstore/**
GOOS=windows GOARCH=amd64 go build -v -o bin/windows/configstore cmd/configstore/**
