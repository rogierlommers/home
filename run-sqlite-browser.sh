#!/bin/bash

# debug output
echo "GUI will be accessible at: http://localhost:8081"

# debug
DBFILE="${PWD}/home-service.db"
echo "Using database file: ${DBFILE}"

# run container with specified database file
docker run -it --rm \
    -p 8081:8080 \
    -v ${PWD}:/data \
    -e SQLITE_DATABASE=home-service.db \
    ghcr.io/coleifer/sqlite-web:latest
