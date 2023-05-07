#!/bin/bash
VERSION=$(git describe --tags --always --dirty)
# Exit on error
set -e

# Build the project
echo "Building the project..."
go build -o bin/go-image-server_"$VERSION" .

# Exit normally
exit 0
