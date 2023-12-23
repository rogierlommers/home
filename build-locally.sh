#!/bin/bash

echo "building docker image..."


docker build . -t rogierlommers/home
docker push rogierlommers/home
