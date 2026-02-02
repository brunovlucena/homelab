# 🛡️ Garak - Makefile

.PHONY: help build build-local buildx buildx-push push deploy test test-local scan clean run-local stop-local logs-local logs status

# Variables
IMAGE_NAME := garak
IMAGE_TAG := latest
REGISTRY := ghcr.io/brunovlucena
NAMESPACE := garak
FULL_IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
LOCAL_IMAGE := $(IMAGE_NAME):$(IMAGE_TAG)
CONTAINER_NAME := garak-local

help: ## Show this help message
	@echo "🛡️ Garak - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build-local: ## Build Docker image locally (single platform for faster local testing)
	@echo "🐳 Building Docker image locally..."
	@echo "📦 Image: $(LOCAL_IMAGE)"
	docker build -t $(LOCAL_IMAGE) .
	@echo "✅ Build complete! Image: $(LOCAL_IMAGE)"
	@echo "💡 Run 'make run-local' to test the container"

build: ## Build Docker image for current platform
	@echo "🐳 Building Docker image..."
	docker build -t $(LOCAL_IMAGE) .
	docker tag $(LOCAL_IMAGE) $(FULL_IMAGE)
	@echo "✅ Build complete! Image: $(FULL_IMAGE)"

buildx: ## Build multi-platform Docker image (amd64, arm64) using buildx
	@echo "🐳 Building multi-platform Docker image..."
	@echo "📦 Platforms: linux/amd64,linux/arm64"
	@echo "📦 Image: $(FULL_IMAGE)"
	@if ! docker buildx ls | grep -q garak-builder; then \
		echo "🔧 Creating buildx builder: garak-builder..."; \
		docker buildx create --name garak-builder --use --bootstrap || \
		docker buildx use garak-builder; \
	fi
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag $(FULL_IMAGE) \
		--tag $(REGISTRY)/$(IMAGE_NAME):latest \
		--cache-from type=local,src=/tmp/.buildx-cache \
		--cache-to type=local,dest=/tmp/.buildx-cache,mode=max \
		--push=false \
		--load=false \
		.
	@echo "✅ Multi-platform build complete!"
	@echo "💡 Use 'docker buildx imagetools inspect $(FULL_IMAGE)' to verify"

buildx-push: ## Build and push multi-platform Docker image
	@echo "🐳 Building and pushing multi-platform Docker image..."
	@echo "📦 Platforms: linux/amd64,linux/arm64"
	@echo "📦 Image: $(FULL_IMAGE)"
	@if ! docker buildx ls | grep -q garak-builder; then \
		echo "🔧 Creating buildx builder: garak-builder..."; \
		docker buildx create --name garak-builder --use --bootstrap || \
		docker buildx use garak-builder; \
	fi
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag $(FULL_IMAGE) \
		--tag $(REGISTRY)/$(IMAGE_NAME):latest \
		--cache-from type=local,src=/tmp/.buildx-cache \
		--cache-to type=local,dest=/tmp/.buildx-cache,mode=max \
		--push \
		.
	@echo "✅ Multi-platform build and push complete!"

push: build ## Build and push Docker image to registry
	@echo "📤 Pushing Docker image to registry..."
	docker push $(FULL_IMAGE)

deploy: ## Deploy to Kubernetes via kubectl
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -k .

status: ## Check deployment status
	@echo "📊 Checking deployment status..."
	kubectl get pods -n $(NAMESPACE)
	kubectl get svc -n $(NAMESPACE)

logs: ## View logs
	@echo "📋 Viewing logs..."
	kubectl logs -n $(NAMESPACE) -l app=$(IMAGE_NAME) -f

run-local: build-local ## Build and run container locally
	@echo "🚀 Running container locally..."
	@echo "📦 Container name: $(CONTAINER_NAME)"
	@echo "🌐 API will be available at: http://localhost:8080"
	@if docker ps -a --format '{{.Names}}' | grep -q "^$(CONTAINER_NAME)$$"; then \
		echo "⚠️  Container already exists. Stopping and removing it..."; \
		docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true; \
		docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true; \
	fi
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		$(LOCAL_IMAGE)
	@echo "✅ Container started!"
	@echo "💡 Run 'make test-local' to test the API"
	@echo "💡 Run 'make logs-local' to view logs"
	@echo "💡 Run 'make stop-local' to stop the container"

stop-local: ## Stop local container
	@echo "🛑 Stopping local container..."
	@if docker ps --format '{{.Names}}' | grep -q "^$(CONTAINER_NAME)$$"; then \
		docker stop $(CONTAINER_NAME); \
		docker rm $(CONTAINER_NAME); \
		echo "✅ Container stopped and removed"; \
	else \
		echo "ℹ️  Container $(CONTAINER_NAME) is not running"; \
	fi

logs-local: ## View logs from local container
	@echo "📋 Viewing logs from local container..."
	@if docker ps -a --format '{{.Names}}' | grep -q "^$(CONTAINER_NAME)$$"; then \
		docker logs -f $(CONTAINER_NAME); \
	else \
		echo "❌ Container $(CONTAINER_NAME) does not exist. Run 'make run-local' first."; \
	fi

test-local: ## Test API running locally
	@echo "🧪 Testing local API..."
	@echo "⏳ Waiting for API to be ready..."
	@for i in $$(seq 1 30); do \
		if curl -sf http://localhost:8080/health >/dev/null 2>&1; then \
			echo "✅ API is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ API did not become ready within 30 seconds"; \
			exit 1; \
		fi; \
		echo "   Attempt $$i/30..."; \
		sleep 1; \
	done
	@echo ""
	@echo "📊 Health check:"
	@curl -s http://localhost:8080/health | jq . || curl -s http://localhost:8080/health
	@echo ""
	@echo "✅ API is responding correctly!"

test: ## Test API in Kubernetes (requires port-forward)
	@echo "🧪 Testing API in Kubernetes..."
	@echo "💡 Make sure to run: kubectl port-forward -n $(NAMESPACE) svc/$(IMAGE_NAME)-service 8080:8080"
	@curl -sf http://localhost:8080/health >/dev/null && \
		echo "✅ API is responding!" || \
		echo "❌ API not available. Run port-forward first."

scan-agent-evil: ## Scan agent-evil (requires port-forward)
	@echo "🔍 Scanning agent-evil..."
	@curl -X POST http://localhost:8080/scan/agent-evil \
		-H "Content-Type: application/json" || echo "API not available. Run port-forward first."

clean: ## Clean up deployment and local resources
	@echo "🧹 Cleaning up..."
	@echo "🗑️  Removing Kubernetes resources..."
	@kubectl delete -k . 2>/dev/null || true
	@echo "🗑️  Removing local container..."
	@if docker ps -a --format '{{.Names}}' | grep -q "^$(CONTAINER_NAME)$$"; then \
		docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true; \
		docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true; \
		echo "✅ Local container removed"; \
	fi
	@echo "✅ Cleanup complete!"

