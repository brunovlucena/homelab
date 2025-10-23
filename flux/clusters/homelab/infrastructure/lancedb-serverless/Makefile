# 🎯 Makefile for LanceDB Serverless

# Variables
IMAGE_NAME := ghcr.io/brunolucena/lancedb-serverless
IMAGE_TAG := latest
NAMESPACE := lancedb-serverless
SERVICE_NAME := lancedb-serverless

.PHONY: help build push build-push deploy delete logs port-forward test create-bucket

help: ## 📚 Show this help message
	@echo "🚀 LanceDB Serverless - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 🔨 Build Docker image
	@echo "🔨 Building Docker image..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "✅ Image built successfully!"

push: ## 📤 Push Docker image to registry
	@echo "📤 Pushing Docker image..."
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@echo "✅ Image pushed successfully!"

build-push: build push ## 🚀 Build and push Docker image

create-bucket: ## 🪣 Create MinIO bucket
	@echo "🪣 Creating MinIO bucket 'lancedb'..."
	@kubectl port-forward -n minio svc/minio-service 9000:9000 & \
	sleep 2 && \
	mc alias set local http://localhost:9000 minioadmin minioadmin && \
	mc mb local/lancedb --ignore-existing && \
	mc policy set public local/lancedb && \
	pkill -f "port-forward.*minio" || true
	@echo "✅ Bucket created successfully!"

deploy: ## 🚀 Deploy to Kubernetes
	@echo "🚀 Deploying LanceDB Serverless..."
	kubectl apply -k .
	@echo "✅ Deployed successfully!"
	@echo ""
	@echo "📊 Check status with: make status"

delete: ## 🗑️ Delete deployment
	@echo "🗑️ Deleting LanceDB Serverless..."
	kubectl delete -k .
	@echo "✅ Deleted successfully!"

status: ## 📊 Check deployment status
	@echo "📊 Checking deployment status..."
	@echo ""
	@echo "=== Knative Service ==="
	kubectl get ksvc -n $(NAMESPACE)
	@echo ""
	@echo "=== Pods ==="
	kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "=== Service URL ==="
	@kubectl get ksvc $(SERVICE_NAME) -n $(NAMESPACE) -o jsonpath='{.status.url}' && echo ""

logs: ## 📝 View service logs
	@echo "📝 Viewing logs..."
	kubectl logs -n $(NAMESPACE) -l serving.knative.dev/service=$(SERVICE_NAME) --tail=100 -f

port-forward: ## 🔌 Port-forward to local machine
	@echo "🔌 Port-forwarding to localhost:8080..."
	@echo "Access the service at: http://localhost:8080"
	kubectl port-forward -n $(NAMESPACE) svc/$(SERVICE_NAME) 8080:80

test: ## 🧪 Run basic API tests
	@echo "🧪 Testing API endpoints..."
	@SERVICE_URL=$$(kubectl get ksvc $(SERVICE_NAME) -n $(NAMESPACE) -o jsonpath='{.status.url}'); \
	echo "Service URL: $$SERVICE_URL"; \
	echo ""; \
	echo "=== Health Check ==="; \
	curl -s $$SERVICE_URL/health | jq .; \
	echo ""; \
	echo "=== List Tables ==="; \
	curl -s $$SERVICE_URL/tables | jq .; \
	echo ""; \
	echo "=== Create Test Table ==="; \
	curl -s -X POST $$SERVICE_URL/tables \
		-H "Content-Type: application/json" \
		-d '{"name":"test","data":[{"id":1,"vector":[0.1,0.2,0.3],"text":"hello"}]}' | jq .; \
	echo ""; \
	echo "=== Search Test ==="; \
	curl -s -X POST $$SERVICE_URL/search \
		-H "Content-Type: application/json" \
		-d '{"table":"test","query_vector":[0.1,0.2,0.3],"limit":5}' | jq .

describe: ## 🔍 Describe Knative service
	kubectl describe ksvc $(SERVICE_NAME) -n $(NAMESPACE)

restart: ## 🔄 Restart service by deleting pods
	@echo "🔄 Restarting service..."
	kubectl delete pods -n $(NAMESPACE) -l serving.knative.dev/service=$(SERVICE_NAME)
	@echo "✅ Service restarted!"

metrics: ## 📈 View metrics
	@echo "📈 Fetching metrics..."
	@SERVICE_URL=$$(kubectl get ksvc $(SERVICE_NAME) -n $(NAMESPACE) -o jsonpath='{.status.url}'); \
	curl -s $$SERVICE_URL/metrics

