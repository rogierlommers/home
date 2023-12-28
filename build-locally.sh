#!/bin/bash

# set to exit when something fails
set -e 
set -o pipefail

echo "building binary"
GOOS=linux GOARCH=amd64 go build -o ./bin/home *.go

echo "building docker image..."
docker buildx build --platform linux/amd64 -t rogierlommers/home .
# docker build . -t rogierlommers/home
docker push rogierlommers/home
