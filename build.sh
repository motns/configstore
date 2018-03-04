#!/usr/bin/env bash

mkdir -p bin/linux
mkdir -p bin/darwin
mkdir -p bin/windows

echo "=== Building 64bit MacOS binary..."
GOOS=darwin GOARCH=amd64 go build -v -o bin/darwin/configstore cmd/configstore/**
echo ""

echo "=== Building 64bit Linux binary..."
GOOS=linux GOARCH=amd64 go build -v -o bin/linux/configstore cmd/configstore/**
echo ""

echo "=== Building 64bit Windows binary..."
GOOS=windows GOARCH=amd64 go build -v -o bin/windows/configstore cmd/configstore/**
echo ""
