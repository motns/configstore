#!/usr/bin/env bash

#for OS in darwin linux openbsd windows
for OS in darwin
do
    echo "=== Building amd64/$OS binary..."
    mkdir -p bin/${OS}/amd64
    GOOS=${OS} GOARCH=amd64 go build -v -o bin/${OS}/amd64/configstore cmd/configstore/**
    echo ""
done

echo "== Done!"