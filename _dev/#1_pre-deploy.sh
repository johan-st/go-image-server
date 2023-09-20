#!/bin/bash

docker build . -t img-server:latest
docker tag img-server:latest registry.digitalocean.com/johan-st/img-server:latest
docker tag img-server:latest registry.digitalocean.com/johan-st/img-server:"$(date +%Y%m%d_%H%M%S)"

echo "verify that the image starts and runs as expected before pushing to registry"
echo "starting container.."
docker run --rm -p 80:80 img-server
