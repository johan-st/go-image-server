#!/bin/bash

DATE_TAG=$(date +%Y%m%d_%H%M%S)
LATEST_TAG="registry.digitalocean.com/johan-st/img-server:latest"

docker tag img-server:latest "$LATEST_TAG"
docker tag img-server:latest "registry.digitalocean.com/johan-st/img-server:$DATE_TAG"
docker push "$LATEST_TAG"
docker push "$DATE_TAG"
