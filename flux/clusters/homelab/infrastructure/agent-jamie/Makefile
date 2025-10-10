.PHONY: help build-bot build-mcp build push-bot push-mcp push deploy delete logs logs-bot logs-mcp test clean docker-login status status-bot status-mcp test-mcp

# 🎨 Configuration
BOT_IMAGE_NAME := jamie-slack-bot
MCP_IMAGE_NAME := jamie-mcp-server
REGISTRY := ghcr.io/brunovlucena
TAG := latest
BOT_IMAGE := $(REGISTRY)/$(BOT_IMAGE_NAME):$(TAG)
MCP_IMAGE := $(REGISTRY)/$(MCP_IMAGE_NAME):$(TAG)
NAMESPACE := jamie

help: ## 📖 Show this help message
	@echo "🤖 Jamie (Slack Bot + MCP Server) - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 🏗️ Build targets
build-bot: ## 🏗️ Build Slack Bot Docker image
	@echo "🏗️ Building Jamie Slack Bot image..."
	docker build -f src/slack-bot/Dockerfile -t $(BOT_IMAGE) .
	@echo "✅ Build complete: $(BOT_IMAGE)"

build: build-bot ## 🏗️ Build both images

# 📤 Push targets
docker-login: ## 🔑 Login to GitHub Container Registry
	@echo "🔑 Logging into GitHub Container Registry..."
	@echo $(GITHUB_TOKEN) | docker login ghcr.io -u $(GITHUB_USER) --password-stdin
	@echo "✅ Login successful"

push-bot: build-bot ## 📤 Push Slack Bot image to GHCR
	@echo "📤 Pushing Slack Bot image to GHCR..."
	docker push $(BOT_IMAGE)
	@echo "✅ Push complete: $(BOT_IMAGE)"

push: push-bot ## 📤 Push both images

# 🚀 Deployment targets
deploy: ## 🚀 Deploy to Kubernetes
	@echo "🚀 Deploying Jamie to Kubernetes..."
	kubectl apply -k .
	@echo "✅ Deployment complete"

flux-reconcile: ## 🔄 Trigger Flux reconciliation
	@echo "🔄 Reconciling Flux..."
	flux reconcile source git homelab
	flux reconcile kustomization infrastructure
	@echo "✅ Flux reconciliation triggered"

delete: ## 🗑️ Delete from Kubernetes
	@echo "🗑️ Deleting Jamie from Kubernetes..."
	kubectl delete -k . --ignore-not-found=true
	@echo "✅ Deletion complete"

# 📋 Logs targets
logs-bot: ## 📋 Show Slack Bot logs
	@echo "📋 Showing Jamie Slack Bot logs..."
	kubectl logs -n $(NAMESPACE) -l app=jamie-slack-bot -f --tail=100

logs-mcp: ## 📋 Show MCP Server logs
	@echo "📋 Showing Jamie MCP Server logs..."
	kubectl logs -n $(NAMESPACE) -l app=jamie-mcp-server -f --tail=100

logs: logs-bot ## 📋 Show Slack Bot logs (default)

# 🔄 Restart targets
restart-bot: ## 🔄 Restart Slack Bot
	@echo "🔄 Restarting Jamie Slack Bot..."
	kubectl rollout restart deployment/jamie-slack-bot -n $(NAMESPACE)
	@echo "✅ Restart initiated"

restart-mcp: ## 🔄 Restart MCP Server
	@echo "🔄 Restarting Jamie MCP Server..."
	kubectl rollout restart deployment/jamie-mcp-server -n $(NAMESPACE)
	@echo "✅ Restart initiated"

restart: restart-bot restart-mcp ## 🔄 Restart both deployments

# 📊 Status targets
status-bot: ## 📊 Check Slack Bot status
	@echo "📊 Checking Jamie Slack Bot status..."
	@kubectl get pods -n $(NAMESPACE) -l app=jamie-slack-bot
	@kubectl get deployment -n $(NAMESPACE) jamie-slack-bot 2>/dev/null || echo "⚠️  Deployment not found"

status-mcp: ## 📊 Check MCP Server status
	@echo "📊 Checking Jamie MCP Server status..."
	@kubectl get pods -n $(NAMESPACE) -l app=jamie-mcp-server
	@kubectl get deployment -n $(NAMESPACE) jamie-mcp-server 2>/dev/null || echo "⚠️  Deployment not found"
	@kubectl get service -n $(NAMESPACE) jamie-mcp-server-service 2>/dev/null || echo "⚠️  Service not found"

status: status-bot status-mcp ## 📊 Check both statuses

# 🧪 Test targets
test-mcp: ## 🧪 Test MCP Server endpoint
	@echo "🧪 Testing Jamie MCP Server..."
	@echo "Health check:"
	@curl -s http://192.168.0.16:30121/health | jq . || echo "❌ Health check failed"
	@echo ""
	@echo "MCP Info:"
	@curl -s http://192.168.0.16:30121/mcp | jq . || echo "❌ MCP info failed"

test: ## 🧪 Run tests
	@echo "🧪 Running tests..."
	python -m pytest tests/ -v

# 🧹 Cleanup targets
clean: ## 🧹 Clean up local resources
	@echo "🧹 Cleaning up..."
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	@echo "✅ Cleanup complete"

# 🔧 Development targets
dev-bot: ## 🔧 Run Slack Bot locally with docker-compose
	@echo "🔧 Starting Jamie Slack Bot locally..."
	docker-compose up --build jamie-slack-bot

dev: ## 🔧 Run both locally
	@echo "🔧 Starting Jamie locally..."
	docker-compose up --build

dev-down: ## 🛑 Stop local development
	@echo "🛑 Stopping local development..."
	docker-compose down

# 🎯 Combined targets
all: clean push flux-reconcile ## 🎯 Build, push, and deploy everything
	@echo "✅ All steps complete!"
	@echo "⏳ Wait for Flux to reconcile and deploy..."
	@echo "📊 Check status with: make status"