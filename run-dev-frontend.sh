#!/bin/bash

echo "running frontend..."


cd frontend && \
VUE_APP_API_HOST="http://localhost:3000/api/send" \
npm run serve
