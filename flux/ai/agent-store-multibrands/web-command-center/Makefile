# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🏬 Multi-Brand Store Command Center - Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Configuration
REGISTRY ?= ghcr.io/brunovlucena
LOCAL_REGISTRY ?= localhost:5001
IMAGE_NAME ?= agent-store-multibrands/command-center
VERSION ?= $(shell cat ../VERSION 2>/dev/null || echo "v1.0.0")
PORT ?= 3001

.PHONY: help dev install build lint docker-build docker-push docker-build-local docker-run clean deploy-studio deploy-rollout status logs port-forward

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
	@echo "🏬 Building Multi-Brand Store Command Center..."
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(VERSION) .
	docker tag $(REGISTRY)/$(IMAGE_NAME):$(VERSION) $(REGISTRY)/$(IMAGE_NAME):latest
	@echo "✅ Built $(REGISTRY)/$(IMAGE_NAME):$(VERSION)"

docker-push: docker-build ## Build and push to registry
	@echo "📤 Pushing to $(REGISTRY)..."
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(REGISTRY)/$(IMAGE_NAME):latest
	@echo "✅ Pushed to $(REGISTRY)/$(IMAGE_NAME):$(VERSION)"

docker-build-local: ## Build and push to local registry
	@echo "🏬 Building for local registry..."
	docker build -t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) .
	docker tag $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest
	docker push $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest
	@echo "✅ Pushed to $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION)"

docker-run: docker-build ## Run Docker container locally
	@echo "🚀 Running container on port $(PORT)..."
	docker run -p $(PORT):$(PORT) --rm $(REGISTRY)/$(IMAGE_NAME):$(VERSION)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Kubernetes
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

deploy-studio: docker-push ## Deploy to studio environment
	kubectl apply -k k8s/kustomize/studio/

deploy-rollout: ## Restart deployment
	kubectl rollout restart deployment/store-command-center -n agent-store-multibrands

status: ## Check deployment status
	@echo "=== Deployment Status ==="
	kubectl get deployment store-command-center -n agent-store-multibrands
	@echo ""
	@echo "=== Pods ==="
	kubectl get pods -n agent-store-multibrands -l app.kubernetes.io/name=store-command-center
	@echo ""
	@echo "=== Service ==="
	kubectl get svc store-command-center -n agent-store-multibrands

logs: ## Tail logs
	kubectl logs -n agent-store-multibrands -l app.kubernetes.io/name=store-command-center -f

port-forward: ## Port forward for local access
	kubectl port-forward -n agent-store-multibrands svc/store-command-center $(PORT):80

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Utilities
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

clean: ## Clean build artifacts
	rm -rf .next node_modules

version: ## Show version
	@echo $(VERSION)
