#!/bin/bash

# Check if a version is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# Verify that the version is valid semver
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Version must be in semver format (e.g. 1.0.0)"
    exit 1
fi

# Build the image for linux/amd64, macOS/arm64, and linux/arm64
docker buildx build \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    --push \
    -t notifinetwork/localfusion:$VERSION \
    -t notifinetwork/localfusion:latest .
