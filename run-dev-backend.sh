#!/bin/bash

echo "running backend..."

DIST_DIRECTORY="${PWD}/frontend/dist" go run ./backend/*.go
