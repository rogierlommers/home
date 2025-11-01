# syntax=docker/dockerfile:1

FROM golang:1.21-bullseye
WORKDIR /workspace
COPY . /workspace

# install common dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    sudo \
    unzip \
    wget \
    && rm -rf /var/lib/apt/lists/*

# install wgo to watch file system changes and rebuild/restart the app
# RUN go install github.com/bokwoon95/wgo@latest

# create entrypoint script
COPY --chown=root:root <<EOF /entrypoint.sh
#!/usr/bin/env bash
set -euo pipefail

cd /workspace

if [ "\${BUILD_AND_RUN_IN_ENTRYPOINT:-false}" = "true" ]; then
  exec just run-dev
else
  sleep infinity
fi
EOF

RUN sudo chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
