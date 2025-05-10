#!/bin/bash
set -e 

# Environment variables GOOS, GOARCH, and APP_VERSION are expected to be set by the workflow.
echo "Building version: $APP_VERSION for $GOOS/$GOARCH"

cd cmd/nicksh

BINARY_NAME="nicksh-${GOOS}-${GOARCH}"
go build -v -trimpath -ldflags="-s -w -X main.Version=$APP_VERSION" -o "${BINARY_NAME}" .
echo "Built: ${BINARY_NAME}"

tar -czvf "${BINARY_NAME}.tar.gz" "${BINARY_NAME}"
echo "Archived: ${BINARY_NAME}.tar.gz"
echo "Build completed successfully."