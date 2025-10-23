# 🤖 Agent Bruno - Makefile

.PHONY: help install dev test build docker-build docker-push deploy clean

# Variables
IMAGE_NAME := agent-bruno
IMAGE_TAG := latest
REGISTRY := ghcr.io/brunovlucena
NAMESPACE := agent-bruno

help: ## Show this help message
	@echo "🤖 Agent Bruno - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies using uv
	@echo "📦 Installing dependencies..."
	uv sync

dev: ## Run development server
	@echo "🚀 Starting development server..."
	uvicorn src.main:app --reload --host 0.0.0.0 --port 8080

test: ## Run tests
	@echo "🧪 Running tests..."
	pytest -v --cov=src --cov-report=term-missing

test-watch: ## Run tests in watch mode
	@echo "👀 Running tests in watch mode..."
	pytest-watch

lint: ## Run linting
	@echo "🔍 Running linters..."
	ruff check src/
	mypy src/

format: ## Format code
	@echo "✨ Formatting code..."
	ruff format src/

build: ## Build Python package
	@echo "📦 Building package..."
	uv build

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-push: docker-build ## Push Docker image to registry
	@echo "📤 Pushing Docker image to registry..."
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

docker-run: ## Run Docker container locally
	@echo "🏃 Running Docker container..."
	docker run -p 8080:8080 \
		-e REDIS_URL=redis://localhost:6379 \
		-e MONGODB_URL=mongodb://localhost:27017 \
		$(IMAGE_NAME):$(IMAGE_TAG)

deploy: ## Deploy to Kubernetes
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -k k8s/

deploy-local: ## Deploy to local Kubernetes (kind/minikube)
	@echo "🏠 Deploying to local Kubernetes..."
	kubectl apply -k k8s/

status: ## Check deployment status
	@echo "📊 Checking deployment status..."
	kubectl get pods -n $(NAMESPACE)
	kubectl get svc -n $(NAMESPACE)
	kubectl get hpa -n $(NAMESPACE)

logs: ## Show logs
	@echo "📜 Showing logs..."
	kubectl logs -n $(NAMESPACE) -l app=agent-bruno --tail=100 -f

port-forward: ## Port forward to local machine
	@echo "🔌 Port forwarding..."
	kubectl port-forward -n $(NAMESPACE) svc/agent-bruno-service 8080:8080

clean: ## Clean generated files
	@echo "🧹 Cleaning..."
	rm -rf build/
	rm -rf dist/
	rm -rf *.egg-info
	rm -rf .pytest_cache/
	rm -rf .coverage
	rm -rf htmlcov/
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete

redis-local: ## Run Redis locally
	@echo "🔴 Starting Redis locally..."
	docker run -d -p 6379:6379 --name redis-local redis:7-alpine

mongodb-local: ## Run MongoDB locally
	@echo "🍃 Starting MongoDB locally..."
	docker run -d -p 27017:27017 --name mongodb-local mongo:7

stop-local: ## Stop local services
	@echo "🛑 Stopping local services..."
	docker stop redis-local mongodb-local 2>/dev/null || true
	docker rm redis-local mongodb-local 2>/dev/null || true

test-api: ## Test API endpoints
	@echo "🧪 Testing API endpoints..."
	@curl -s http://localhost:8080/health | jq
	@echo "\n"
	@curl -s http://localhost:8080/ready | jq
	@echo "\n"
	@curl -s http://localhost:8080/knowledge/summary | jq

test-chat: ## Test chat endpoint
	@echo "💬 Testing chat endpoint..."
	@curl -s -X POST http://localhost:8080/chat \
		-H "Content-Type: application/json" \
		-d '{"message": "How do I deploy the homepage?"}' | jq

watch: ## Watch for changes and reload
	@echo "👀 Watching for changes..."
	watchmedo auto-restart --directory=./src --pattern=*.py --recursive -- \
		uvicorn src.main:app --host 0.0.0.0 --port 8080

