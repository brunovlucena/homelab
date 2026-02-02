.PHONY: help build test deploy clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build Docker image
	docker build -t ghcr.io/brunovlucena/speech-coach-agent:latest -f src/speech_coach/Dockerfile src/

test: ## Run tests
	pytest tests/ -v

deploy: ## Deploy to cluster
	kubectl apply -k k8s/kustomize/studio/

clean: ## Clean build artifacts
	find . -type d -name __pycache__ -exec rm -r {} +
	find . -type f -name "*.pyc" -delete
