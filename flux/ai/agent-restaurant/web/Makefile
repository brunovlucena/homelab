# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🍽️ Restaurant Command Center - Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Configuration
REGISTRY ?= ghcr.io/brunovlucena
LOCAL_REGISTRY ?= localhost:5001
IMAGE_NAME ?= agent-restaurant/command-center
VERSION ?= $(shell cat ../VERSION 2>/dev/null || echo "v1.0.0")
PORT ?= 3005

.PHONY: help dev install build lint docker-build docker-push docker-build-local docker-run clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Development
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

dev: ## Run local development server
	npm run dev

install: ## Install dependencies
	npm install

build: ## Build for production
	npm run build

lint: ## Run linter
	npm run lint

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Docker
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

docker-build: ## Build Docker image
	@echo "🍽️ Building Restaurant Command Center..."
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(VERSION) .
	docker tag $(REGISTRY)/$(IMAGE_NAME):$(VERSION) $(REGISTRY)/$(IMAGE_NAME):latest
	@echo "✅ Built $(REGISTRY)/$(IMAGE_NAME):$(VERSION)"

docker-push: docker-build ## Build and push to registry
	@echo "📤 Pushing to $(REGISTRY)..."
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(REGISTRY)/$(IMAGE_NAME):latest
	@echo "✅ Pushed to $(REGISTRY)/$(IMAGE_NAME):$(VERSION)"

docker-build-local: ## Build and push to local registry
	@echo "🍽️ Building for local registry..."
	docker build -t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) .
	docker tag $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest
	docker push $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest
	@echo "✅ Pushed to $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION)"

docker-run: docker-build ## Run Docker container locally
	@echo "🚀 Running container on port $(PORT)..."
	docker run -p $(PORT):$(PORT) --rm $(REGISTRY)/$(IMAGE_NAME):$(VERSION)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Utilities
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

clean: ## Clean build artifacts
	rm -rf .next node_modules

version: ## Show version
	@echo $(VERSION)
