#!/bin/bash

echo "running backend..."

set -o allexport
source backend.secrets.sh
set +o allexport

go run ./*.go
