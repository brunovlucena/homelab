# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🏥 Medical Agent Command Center - Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Configuration
VERSION := 1.0.0
IMAGE_NAME := agent-medical/command-center
LOCAL_REGISTRY := localhost:5001
REMOTE_REGISTRY := ghcr.io/brunovlucena/homelab

.PHONY: help
help: ## Show this help
	@echo "Medical Agent Command Center - Make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: dev
dev: ## Run development server
	npm run dev

.PHONY: install
install: ## Install dependencies
	npm ci

.PHONY: build-app
build-app: ## Build Next.js app
	npm run build

.PHONY: typecheck
typecheck: ## Run TypeScript type checking
	npm run typecheck

.PHONY: lint
lint: ## Run linter
	npm run lint

# Docker
.PHONY: build
build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) .
	docker tag $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest

.PHONY: push-local
push-local: build ## Push to local registry
	@echo "📤 Pushing to local registry..."
	docker push $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest

.PHONY: push-remote
push-remote: ## Push to remote registry
	@echo "📤 Pushing to remote registry..."
	docker tag $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker tag $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest $(REMOTE_REGISTRY)/$(IMAGE_NAME):latest
	docker push $(REMOTE_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE_NAME):latest

# Kubernetes
.PHONY: deploy
deploy: ## Deploy to Kubernetes
	@echo "☸️  Deploying to Kubernetes..."
	kubectl apply -k k8s/kustomize/base

.PHONY: delete
delete: ## Delete from Kubernetes
	@echo "🗑️  Deleting from Kubernetes..."
	kubectl delete -k k8s/kustomize/base

.PHONY: logs
logs: ## Show logs
	kubectl logs -n agent-medical -l app.kubernetes.io/name=medical-command-center -f

.PHONY: status
status: ## Show deployment status
	@echo "📊 Deployment Status:"
	@kubectl get deployment -n agent-medical medical-command-center
	@echo "\n📦 Pods:"
	@kubectl get pods -n agent-medical -l app.kubernetes.io/name=medical-command-center
	@echo "\n🌐 Service:"
	@kubectl get service -n agent-medical medical-command-center
	@echo "\n☁️  Cloudflare Tunnel:"
	@kubectl get cloudflaretunnelingress -n agent-medical medical-command-center

.PHONY: test-local
test-local: ## Test locally with Docker
	@echo "🧪 Testing locally..."
	docker run --rm -p 3002:8080 $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(VERSION)

# Clean
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf .next
	rm -rf node_modules

.PHONY: all
all: install typecheck lint build push-local deploy ## Run full pipeline
