.PHONY: help install build build-local build-local-no-cache build-ghcr build-all push deploy-pro deploy-dev run-agent test clean export-ollama import-ollama train-and-deploy version version-bump clear-cache diagnose-build fix-buildx

# Version configuration
VERSION_FILE := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null | tr -d '[:space:]' || echo "0.1.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Docker configuration
REGISTRY ?= ghcr.io/brunovlucena
LOCAL_REGISTRY ?= localhost:5001
IMAGE_NAME ?= agent-sre
IMAGE_TAG ?= v$(VERSION)
NAMESPACE ?= ai
OLLAMA_MODEL_NAME ?= agent-sre:latest

# Buildx configuration
BUILDX_BUILDER ?= homelab
PUSH_PLATFORM ?= linux/amd64,linux/arm64
LOAD_PLATFORM ?= linux/arm64

help: ## 📋 Show this help message
	@echo "Agent-SRE Makefile"
	@echo ""
	@echo "Current version: $(VERSION)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-25s %s\n", $$1, $$2}'

install:
	pip install -r src/requirements.txt
	pip install -e src/

.PHONY: ensure-buildx-builder
ensure-buildx-builder: ## 🔧 Ensure buildx builder exists
	@if ! docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1; then \
		echo "🔧 Creating buildx builder '$(BUILDX_BUILDER)' with host network access..."; \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --driver-opt network=host --use || true; \
		docker buildx inspect $(BUILDX_BUILDER) --bootstrap || true; \
	else \
		echo "✅ Buildx builder '$(BUILDX_BUILDER)' already exists"; \
	fi
	@echo "🔧 Configuring builder to skip credential helper..."
	@docker buildx inspect $(BUILDX_BUILDER) --bootstrap >/dev/null 2>&1 || true

build: ## 🐳 Build Docker image (local, single arch)
	docker build -t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest -f Dockerfile .

build-local: ensure-buildx-builder ## 🐳 Build and push to localhost:5001 (arm64)
	@echo "🐳 Building agent-sre v$(VERSION) for local registry..."
	@echo "   Using DOCKER_BUILDKIT=1 and skipping credential helper..."
	@DOCKER_BUILDKIT=1 DOCKER_CONFIG=$${DOCKER_CONFIG:-$$HOME/.docker} \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(LOAD_PLATFORM) \
		--push \
		--secret id=creds,src=/dev/null \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest \
		. 2>&1 | grep -v "keychain" || \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(LOAD_PLATFORM) \
		--push \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest \
		. || \
	( \
		echo "⚠️  Standard build failed, trying with credential helper disabled..."; \
		DOCKER_CONFIG_TMP=$$(mktemp -d); \
		echo '{}' > "$$DOCKER_CONFIG_TMP/config.json"; \
		DOCKER_CONFIG="$$DOCKER_CONFIG_TMP" docker buildx build \
			--builder $(BUILDX_BUILDER) \
			-f Dockerfile \
			--platform $(LOAD_PLATFORM) \
			--push \
			-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
			-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest \
			. && \
		rm -rf "$$DOCKER_CONFIG_TMP" \
	)
	@echo "✅ Built and pushed to $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

build-local-no-cache: ensure-buildx-builder ## 🐳 Build and push to localhost:5001 (arm64) without cache
	@echo "🐳 Building agent-sre v$(VERSION) for local registry (NO CACHE)..."
	@echo "   Using DOCKER_BUILDKIT=1 and skipping credential helper..."
	@DOCKER_BUILDKIT=1 DOCKER_CONFIG=$${DOCKER_CONFIG:-$$HOME/.docker} \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(LOAD_PLATFORM) \
		--push \
		--no-cache \
		--secret id=creds,src=/dev/null \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest \
		. 2>&1 | grep -v "keychain" || \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(LOAD_PLATFORM) \
		--push \
		--no-cache \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest \
		. || \
	( \
		echo "⚠️  Standard build failed, trying with credential helper disabled..."; \
		DOCKER_CONFIG_TMP=$$(mktemp -d); \
		echo '{}' > "$$DOCKER_CONFIG_TMP/config.json"; \
		DOCKER_CONFIG="$$DOCKER_CONFIG_TMP" docker buildx build \
			--builder $(BUILDX_BUILDER) \
			-f Dockerfile \
			--platform $(LOAD_PLATFORM) \
			--push \
			--no-cache \
			-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
			-t $(LOCAL_REGISTRY)/$(IMAGE_NAME):latest \
			. && \
		rm -rf "$$DOCKER_CONFIG_TMP" \
	)
	@echo "✅ Built and pushed to $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) (no cache)"

build-ghcr: ensure-buildx-builder ## 🐳 Build and push to ghcr.io (multi-arch)
	@echo "🐳 Building agent-sre v$(VERSION) for GHCR..."
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(PUSH_PLATFORM) \
		--push \
		-t $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(REGISTRY)/$(IMAGE_NAME):latest \
		.
	@echo "✅ Built and pushed to $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

build-all: build-local build-ghcr ## 🐳 Build and push to both localhost:5001 and ghcr.io

push: build-ghcr ## 🚀 Push image to ghcr.io (alias for build-ghcr)

deploy-pro:
	kubectl apply -k k8s/kustomize/pro

deploy-dev:
	kubectl apply -k k8s/kustomize/dev

run-agent:
	python -m src.sre_agent.main

generate-report:
	@if [ -z "$(COMPONENT)" ]; then \
		echo "Usage: make generate-report COMPONENT=loki|prometheus|infrastructure|all"; \
	else \
		python -m src.sre_agent.main --component $(COMPONENT); \
	fi

test:
	pytest tests/ -v

clean:
	rm -rf dist/ build/ *.egg-info
	find . -type d -name __pycache__ -exec rm -r {} +
	find . -type f -name "*.pyc" -delete

clear-cache: ## 🗑️ Clear Docker buildx cache
	@echo "🗑️ Clearing Docker buildx cache..."
	@docker buildx prune -af
	@echo "✅ Cache cleared"

diagnose-build: ## 🔍 Diagnose build environment issues
	@./scripts/diagnose-build.sh

fix-buildx: ## 🔧 Fix buildx builder configuration (recreate builder)
	@if [ -f "scripts/fix-buildx-keychain.sh" ]; then \
		./scripts/fix-buildx-keychain.sh; \
	else \
		echo "🔧 Recreating buildx builder '$(BUILDX_BUILDER)'..."; \
		docker buildx rm $(BUILDX_BUILDER) 2>/dev/null || true; \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --driver-opt network=host --use; \
		docker buildx inspect $(BUILDX_BUILDER) --bootstrap; \
		echo "✅ Builder recreated!"; \
	fi

export-ollama:
	@if [ -z "$(BASE_MODEL)" ] || [ -z "$(ADAPTER_PATH)" ]; then \
		echo "Usage: make export-ollama BASE_MODEL=/path/to/mlx-model ADAPTER_PATH=/path/to/adapters [OUTPUT=./agent-sre-ollama] [MODEL_NAME=agent-sre:latest]"; \
		exit 1; \
	fi
	@python export_ollama.py \
		--base-model "$(BASE_MODEL)" \
		--adapter-path "$(ADAPTER_PATH)" \
		--output "$(OUTPUT)" \
		--model-name "$(OLLAMA_MODEL_NAME)"

import-ollama:
	@if [ -z "$(MODEL_PATH)" ]; then \
		echo "Usage: make import-ollama MODEL_PATH=./agent-sre-ollama [MODEL_NAME=agent-sre:latest] [OLLAMA_HOST=localhost:11434]"; \
		exit 1; \
	fi
	@if [ -z "$(OLLAMA_HOST)" ]; then \
		OLLAMA_HOST="localhost:11434"; \
	fi
	@echo "📤 Importing model into Ollama..."
	@MODELFILE="$(MODEL_PATH)/Modelfile"; \
	if [ ! -f "$$MODELFILE" ]; then \
		MODELFILE=$$(find "$(MODEL_PATH)" -name "Modelfile" | head -1); \
	fi; \
	if [ -z "$$MODELFILE" ] || [ ! -f "$$MODELFILE" ]; then \
		echo "❌ Modelfile not found in $(MODEL_PATH)"; \
		exit 1; \
	fi; \
	echo "   Using Modelfile: $$MODELFILE"; \
	ollama create "$(OLLAMA_MODEL_NAME)" -f "$$MODELFILE" || \
		(echo "⚠️  Model might already exist, trying to update..." && \
		 ollama rm "$(OLLAMA_MODEL_NAME)" 2>/dev/null || true && \
		 ollama create "$(OLLAMA_MODEL_NAME)" -f "$$MODELFILE"); \
	echo "✅ Model $(OLLAMA_MODEL_NAME) imported successfully!"

train-and-deploy:
	@echo "🚀 Triggering Flyte training workflow..."
	@echo "   This will:"
	@echo "   1. Train the model on RUNBOOK.md"
	@echo "   2. Export to Ollama format"
	@echo "   3. Store in MinIO"
	@echo "   4. Automatically import into Ollama"
	@echo ""
	@kubectl create job --from=cronjob/flyte-trigger-training -n flyte train-agent-sre-$$(date +%s) 2>/dev/null || \
		kubectl apply -f ../infrastructure/flyte/k8s/trigger-training-job.yaml || \
		(echo "⚠️  Could not trigger training job. Please trigger manually:" && \
		 echo "   kubectl apply -f ../infrastructure/flyte/k8s/trigger-training-job.yaml"); \
	echo "✅ Training workflow triggered!"
	@echo ""
	@echo "📊 Monitor progress:"
	@echo "   kubectl get pods -n homelab-production | grep agent-sre"
	@echo "   kubectl logs -n flyte -l app=flyte-workflow-registration"
	@echo ""
	@echo "⏳ After training completes, the model will be automatically imported into Ollama"
	@echo "   Check import job: kubectl get jobs -n ai | grep import-agent-sre-ollama"

version: ## 🏷️ Show current version
	@echo "Current version: $(VERSION)"
	@echo "Version file: $(VERSION_FILE)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Git branch: $(GIT_BRANCH)"
	@echo "Image tag: $(IMAGE_TAG)"
	@echo "Local image: $(LOCAL_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "GHCR image: $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "❌ Usage: make version-bump NEW_VERSION=x.y.z"; \
		exit 1; \
	fi
	@OLD_VERSION=$(VERSION); \
	echo "🏷️ Bumping version: $$OLD_VERSION → $(NEW_VERSION)"; \
	echo ""; \
	echo "$(NEW_VERSION)" > $(VERSION_FILE); \
	echo "  ✅ Updated VERSION file"; \
	# Update base lambdaagent.yaml tag
	@if [ -f "k8s/kustomize/base/lambdaagent.yaml" ]; then \
		sed -i.bak 's|tag: "[^"]*"|tag: "v$(NEW_VERSION)"|g' k8s/kustomize/base/lambdaagent.yaml && rm -f k8s/kustomize/base/lambdaagent.yaml.bak; \
		echo "  ✅ Updated k8s/kustomize/base/lambdaagent.yaml"; \
	fi
	# Update kustomization overlays (LambdaAgent patches)
	@for overlay in k8s/kustomize/pro k8s/kustomize/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|value: "[^"]*"|value: "v$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (LambdaAgent patches)"; \
			fi; \
		fi; \
	done
	@echo ""
	@echo "✅ Version bumped to v$(NEW_VERSION)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-sre v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

