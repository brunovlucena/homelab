# Makefile for Flyte Sandbox

.PHONY: help build build-local push push-local build-and-push register test clean

# Configuration
IMAGE_NAME ?= ghcr.io/brunovlucena/flyte-sandbox-training
IMAGE_TAG ?= latest
REGISTRY_IMAGE_NAME ?= ghcr.io/brunovlucena/flyte-workflow-registry
REGISTRY_IMAGE_TAG ?= latest
PROJECT ?= homelab
DOMAIN ?= production
FLYTE_ENDPOINT ?= flyte.ml-platform.svc.forge.remote:81

help:
	@echo "Flyte Sandbox - ML Training Pipeline"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build Docker image"
	@echo "  make build-local    - Build Docker image locally (with confirmation)"
	@echo "  make push           - Build and push Docker image to registry"
	@echo "  make push-local     - Build and push Docker image locally"
	@echo "  make build-and-push - Build and push in one command (local)"
	@echo "  make register       - Register workflow with Flyte"
	@echo "  make test           - Test workflow locally"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Configuration:"
	@echo "  IMAGE_NAME=$(IMAGE_NAME)"
	@echo "  IMAGE_TAG=$(IMAGE_TAG)"
	@echo "  PROJECT=$(PROJECT)"
	@echo "  DOMAIN=$(DOMAIN)"
	@echo ""
	@echo "Quick start (local build and push):"
	@echo "  make build-and-push"

build:
	@echo "Building Docker image: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

build-local:
	@echo "Building Docker image locally: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "✅ Image built: $(IMAGE_NAME):$(IMAGE_TAG)"

build-registry:
	@echo "Building registry Docker image: $(REGISTRY_IMAGE_NAME):$(REGISTRY_IMAGE_TAG)"
	docker build -f Dockerfile.registry -t $(REGISTRY_IMAGE_NAME):$(REGISTRY_IMAGE_TAG) .

push: build
	@echo "Pushing Docker image: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

push-local: build-local
	@echo "Pushing Docker image to registry: $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "🔐 Make sure you're logged in to GHCR:"
	@echo "   docker login ghcr.io -u YOUR_USERNAME"
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@echo "✅ Image pushed: $(IMAGE_NAME):$(IMAGE_TAG)"

build-and-push: build-local
	@echo ""
	@echo "🚀 Building and pushing Docker image..."
	@echo "   Image: $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo ""
	@echo "🔐 Checking Docker login status..."
	@docker info | grep -q "Username" || (echo "❌ Not logged in to Docker registry. Run: docker login ghcr.io" && exit 1)
	@echo "✅ Logged in to Docker registry"
	@echo ""
	@echo "📤 Pushing image..."
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@echo ""
	@echo "✅ Successfully built and pushed: $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "   You can now register the workflow with: make register"

push-registry: build-registry
	@echo "Pushing registry Docker image: $(REGISTRY_IMAGE_NAME):$(REGISTRY_IMAGE_TAG)"
	docker push $(REGISTRY_IMAGE_NAME):$(REGISTRY_IMAGE_TAG)

register: push
	@echo "Registering workflow with Flyte..."
	pyflyte register workflows/agent_training.py \
		--project $(PROJECT) \
		--domain $(DOMAIN) \
		--image $(IMAGE_NAME):$(IMAGE_TAG)

test:
	@echo "Testing workflow locally..."
	python -m workflows.agent_training

clean:
	@echo "Cleaning build artifacts..."
	rm -rf __pycache__/
	rm -rf workflows/__pycache__/
	rm -rf scripts/__pycache__/
	rm -rf .flyte/
	rm -f flyte-package.tar.gz

install:
	@echo "Installing dependencies..."
	pip install -r requirements.txt
	pip install flytekit flytectl

validate:
	@echo "Validating Flyte configuration..."
	flytectl config validate

list:
	@echo "Listing workflows..."
	flytectl get workflows --project $(PROJECT) --domain $(DOMAIN)

