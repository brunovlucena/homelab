.PHONY: help build push deploy delete logs test clean

# 🎨 Configuration
IMAGE_NAME := jamie-slack-bot
REGISTRY := 565265565115.dkr.ecr.us-west-2.amazonaws.com
TAG := dev
IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(TAG)
NAMESPACE := homepage

help: ## 📖 Show this help message
	@echo "🤖 Jamie Slack Bot - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## 🏗️ Build Docker image
	@echo "🏗️ Building Jamie Slack Bot image..."
	docker build -t $(IMAGE) .
	@echo "✅ Build complete: $(IMAGE)"

push: build ## 📤 Push image to ECR
	@echo "📤 Pushing image to ECR..."
	aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin $(REGISTRY)
	docker push $(IMAGE)
	@echo "✅ Push complete: $(IMAGE)"

deploy: ## 🚀 Deploy to Kubernetes
	@echo "🚀 Deploying Jamie Slack Bot to Kubernetes..."
	kubectl apply -f k8s/
	@echo "✅ Deployment complete"

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
	kubectl get pods -n $(NAMESPACE) -l app=jamie-slack-bot
	kubectl get deployment -n $(NAMESPACE) jamie-slack-bot
	kubectl get service -n $(NAMESPACE) jamie-slack-bot

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

all: clean build push deploy ## 🎯 Build, push, and deploy
	@echo "✅ All steps complete!"

