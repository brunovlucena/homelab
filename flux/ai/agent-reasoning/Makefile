.PHONY: help install run-dev build push deploy-studio test clean

# Default target
help:
	@echo "Agent-Reasoning Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  install      - Install dependencies"
	@echo "  run-dev      - Run development server"
	@echo "  build        - Build Docker image"
	@echo "  push         - Push image to registry"
	@echo "  deploy-studio - Deploy to studio cluster"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"

# Variables
IMAGE_NAME ?= agent-reasoning
IMAGE_TAG ?= latest
REGISTRY ?= your-registry.local
NAMESPACE ?= ai-agents

# Install dependencies
install:
	python3 -m pip install --upgrade pip
	python3 -m pip install -r src/requirements.txt

# Run development server
run-dev:
	cd src && uvicorn reasoning.main:app --host 0.0.0.0 --port 8080 --reload

# Build Docker image
build:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -f src/reasoning/Dockerfile .

# Push image
push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Deploy to studio
deploy-studio:
	kubectl apply -k k8s/kustomize/studio

# Run tests
test:
	pytest tests/ -v

# Clean
clean:
	find . -type d -name __pycache__ -exec rm -r {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete

