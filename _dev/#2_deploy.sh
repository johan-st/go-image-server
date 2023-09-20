#!/bin/bash

docker tag img-server:latest registry.digitalocean.com/johan-st/img-server:latest
docker tag img-server:latest registry.digitalocean.com/johan-st/img-server:"$(date +%Y%m%d_%H%M%S)"
docker push -a registry.digitalocean.com/johan-st/img-server
