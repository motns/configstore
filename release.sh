#!/usr/bin/env bash

if [ -z "$1" ]
then
    echo "ERROR: You need to pass in the version number you're releasing"
    exit 1
fi

version="$1"

app_version_check=`grep "$version" main.go`

if [ -z "$app_version_check" ]
then
    echo "ERROR: App version number in main.go doesn't match"
    exit 1
fi

run_cmd() {
    eval "$1"

    if [ $? -gt 0 ]
    then
        echo "ERROR: Failed to execute command: $1"
        exit 1
    fi
}

echo "Building archives for supported platforms..."

run_cmd "tar -czf configstore-${version}-darwin-amd64.tar.gz -C bin/darwin configstore"
run_cmd "tar -czf configstore-${version}-linux-amd64.tar.gz -C bin/linux configstore"
run_cmd "tar -czf configstore-${version}-windows-amd64.tar.gz -C bin/windows configstore"

echo "Creating Git Tag..."
run_cmd "git tag ${version}"

echo "Pushing Git Tag to remote..."
run_cmd "git push origin ${version}"

echo "Done!"