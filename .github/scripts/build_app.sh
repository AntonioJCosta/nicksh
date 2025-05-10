#!/bin/bash
set -e 

echo "Building version: $APP_VERSION for $GOOS/$GOARCH"

# Version for ldflags (without 'v' prefix)
LD_APP_VERSION="${APP_VERSION#v}"

cd cmd/nicksh

BINARY_NAME_BASE="nicksh-${GOOS}-${GOARCH}"
ARCHIVE_NAME="${BINARY_NAME_BASE}-${APP_VERSION}.tar.gz" 
OUTPUT_BINARY_NAME="${BINARY_NAME_BASE}"

go build -v -trimpath -ldflags="-s -w -X main.Version=$LD_APP_VERSION" -o "${OUTPUT_BINARY_NAME}" .
echo "Built: ${OUTPUT_BINARY_NAME} (internal version $LD_APP_VERSION)"

tar -czvf "${ARCHIVE_NAME}" "${OUTPUT_BINARY_NAME}"
echo "Archived: ${ARCHIVE_NAME}"
echo "Build completed successfully."