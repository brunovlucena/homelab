.PHONY: help build push deploy delete logs test clean docker-login

# 🎨 Configuration
IMAGE_NAME := jamie-slack-bot
REGISTRY := ghcr.io/brunovlucena
TAG := latest
IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(TAG)
NAMESPACE := jamie

help: ## 📖 Show this help message
	@echo "🤖 Jamie Slack Bot - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## 🏗️ Build Docker image
	@echo "🏗️ Building Jamie Slack Bot image..."
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE) --load .
	@echo "✅ Build complete: $(IMAGE)"

build-local: ## 🏗️ Build Docker image for local platform only
	@echo "🏗️ Building Jamie Slack Bot image (local platform)..."
	docker build -t $(IMAGE) .
	@echo "✅ Build complete: $(IMAGE)"

docker-login: ## 🔑 Login to GitHub Container Registry
	@echo "🔑 Logging into GitHub Container Registry..."
	@echo $(GITHUB_TOKEN) | docker login ghcr.io -u $(GITHUB_USER) --password-stdin
	@echo "✅ Login successful"

push: build-local ## 📤 Push image to GitHub Container Registry
	@echo "📤 Pushing image to GHCR..."
	@echo "⚠️  Make sure you're logged in: echo \$$GITHUB_TOKEN | docker login ghcr.io -u brunovlucena --password-stdin"
	docker push $(IMAGE)
	@echo "✅ Push complete: $(IMAGE)"
	@echo "🔄 Trigger Flux reconciliation with: flux reconcile kustomization infrastructure"

deploy: ## 🚀 Deploy to Kubernetes
	@echo "🚀 Deploying Jamie Slack Bot to Kubernetes..."
	kubectl apply -k k8s/
	@echo "✅ Deployment complete"

flux-reconcile: ## 🔄 Trigger Flux reconciliation
	@echo "🔄 Reconciling Flux..."
	flux reconcile source git homelab
	flux reconcile kustomization infrastructure
	@echo "✅ Flux reconciliation triggered"

delete: ## 🗑️ Delete from Kubernetes
	@echo "🗑️ Deleting Jamie Slack Bot from Kubernetes..."
	kubectl delete -f k8s/ --ignore-not-found=true
	@echo "✅ Deletion complete"

logs: ## 📋 Show logs
	@echo "📋 Showing Jamie Slack Bot logs..."
	kubectl logs -n $(NAMESPACE) -l app=jamie-slack-bot -f --tail=100

restart: ## 🔄 Restart deployment
	@echo "🔄 Restarting Jamie Slack Bot..."
	kubectl rollout restart deployment/jamie-slack-bot -n $(NAMESPACE)
	@echo "✅ Restart initiated"

status: ## 📊 Check deployment status
	@echo "📊 Checking Jamie Slack Bot status..."
	@kubectl get pods -n $(NAMESPACE) -l app=jamie-slack-bot
	@kubectl get deployment -n $(NAMESPACE) jamie-slack-bot 2>/dev/null || echo "⚠️  Deployment not found"
	@kubectl get secret -n $(NAMESPACE) jamie-slack-secrets 2>/dev/null || echo "⚠️  Secret not found"

test: ## 🧪 Run local tests
	@echo "🧪 Running tests..."
	python -m pytest tests/ -v

clean: ## 🧹 Clean up local resources
	@echo "🧹 Cleaning up..."
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	@echo "✅ Cleanup complete"

dev: ## 🔧 Run locally with docker-compose
	@echo "🔧 Starting Jamie Slack Bot locally..."
	docker-compose up --build

dev-down: ## 🛑 Stop local development
	@echo "🛑 Stopping local development..."
	docker-compose down

all: clean push flux-reconcile ## 🎯 Build, push, and let Flux deploy
	@echo "✅ All steps complete!"
	@echo "⏳ Wait for Flux to reconcile and deploy..."
	@echo "📊 Check status with: make status"

