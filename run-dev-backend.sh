#!/bin/bash

echo "running backend..."

set -o allexport
source backend.secrets.sh
set +o allexport

DIST_DIRECTORY="${PWD}/frontend/dist" go run ./backend/*.go
