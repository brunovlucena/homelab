#!/bin/bash

# Build script for knative-lambda-metrics-pusher

set -e

# Configuration
REGISTRY="339954290315.dkr.ecr.us-west-2.amazonaws.com"
REPOSITORY="knative-lambdas/knative-lambda-metrics-pusher"
TAG="${1:-latest}"
IMAGE="${REGISTRY}/${REPOSITORY}:${TAG}"

echo "🏗️  Building knative-lambda-metrics-pusher..."
echo "📦 Image: ${IMAGE}"

# Build the Docker image
docker build -t "${IMAGE}" .

echo "✅ Build completed successfully!"
echo "🚀 To push to ECR:"
echo "   aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin ${REGISTRY}"
echo "   docker push ${IMAGE}" 