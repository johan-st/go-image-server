#!/bin/bash

# Exit on error
set -e

# Build the project
echo "Building the project..."
go build -o bin/ ./...

# Run the project
echo "Running the project..."
./bin/go-image-server

# Exit normally
exit 0
