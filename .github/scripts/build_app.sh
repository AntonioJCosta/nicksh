#!/bin/bash
set -e 

echo "Building version: $APP_VERSION for $GOOS/$GOARCH"

# Remove version prefix from APP_VERSION
APP_VERSION="${APP_VERSION#v}"

cd cmd/nicksh

BINARY_NAME_BASE="nicksh-${GOOS}-${GOARCH}"
BINARY_NAME_WITH_VERSION="nicksh-${GOOS}-${GOARCH}-${APP_VERSION}"
OUTPUT_BINARY_NAME="${BINARY_NAME_BASE}"
ARCHIVE_NAME="${BINARY_NAME_BASE}-${APP_VERSION}.tar.gz" 

go build -v -trimpath -ldflags="-s -w -X main.Version=$APP_VERSION" -o "${OUTPUT_BINARY_NAME}" .
echo "Built: ${OUTPUT_BINARY_NAME}"

tar -czvf "${ARCHIVE_NAME}" "${OUTPUT_BINARY_NAME}"
echo "Archived: ${ARCHIVE_NAME}"
echo "Build completed successfully."