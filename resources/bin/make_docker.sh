#!/usr/bin/env bash
if [ $# -eq 0 ]; then
    echo "Usage: $0 <image-name>"
    exit 1
fi
IMAGE=$1
VERSION=$(git describe --always --dirty)
DATE=$(date -u +%Y-%m-%dT%H:%M:%S)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
VERSION=$(echo $VERSION | sed -e"s/^v//")

set -x
docker build --build-arg VERSION=${VERSION} --build-arg DATE=${DATE} -t ${IMAGE}:${VERSION} .
set +x

if [ "$BRANCH" = "master" ]; then
    set -x
    docker tag ${IMAGE}:${VERSION} ${IMAGE}:latest
    set +x
fi
