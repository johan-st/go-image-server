#!/bin/bash

# exit on error
set -e

docker build . -t img-server:latest

echo
echo "----------------------------------------------------------------------------"
echo "verify that the image starts and runs as expected before pushing to registry"
echo "----------------------------------------------------------------------------"
echo "starting container.."
echo
docker run --rm -p 80:80 img-server
