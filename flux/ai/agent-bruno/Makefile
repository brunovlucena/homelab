# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🤖 AGENT-BRUNO - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the AI-Powered Chatbot for Homelab
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make build-local      - Build and push Docker image to local registry
#	  make test             - Run tests
#	  make deploy-studio    - Deploy to studio environment
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.DEFAULT_GOAL := help

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 PATH & PLATFORM CONFIGURATION                                       │
# └─────────────────────────────────────────────────────────────────────────┘

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)/src

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
	ifeq ($(UNAME_M),arm64)
		DEFAULT_LOAD_PLATFORM := linux/arm64
	endif
endif

DEFAULT_LOAD_PLATFORM ?= linux/amd64
LOAD_PLATFORM ?= $(DEFAULT_LOAD_PLATFORM)
PUSH_PLATFORM ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?= homelab

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ PROJECT CONFIGURATION                                              │
# └─────────────────────────────────────────────────────────────────────────┘

PROJECT_NAME := agent-bruno
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
PYTHON ?= python3

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY ?= localhost:5001
IMAGE_PREFIX := $(REGISTRY)/agent-bruno
CHATBOT_IMAGE := $(IMAGE_PREFIX)/chatbot

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/agent-bruno
ifeq ($(GIT_BRANCH),main)
	DOCKER_TAG := $(VERSION)
else ifeq ($(GIT_BRANCH),develop)
	DOCKER_TAG := $(VERSION)-beta.$(shell date +%s)
else
	DOCKER_TAG := $(VERSION)-dev.$(GIT_COMMIT)
endif

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES CONFIGURATION                                            │
# └─────────────────────────────────────────────────────────────────────────┘

ENV ?= studio
NAMESPACE ?= agent-bruno

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎨 COLORS                                                              │
# └─────────────────────────────────────────────────────────────────────────┘

COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m
COLOR_RED := \033[31m

define print_status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)$(1)$(COLOR_RESET)"
endef

define print_success
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ $(1)$(COLOR_RESET)"
endef

define print_warning
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)⚠️  $(1)$(COLOR_RESET)"
endef

define print_error
	@echo "$(COLOR_BOLD)$(COLOR_RED)❌ $(1)$(COLOR_RESET)"
endef

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  📋 HELP
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: help
help: ## 📋 Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🤖 Agent-Bruno$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🏷️ VERSION MANAGEMENT (DRY Pattern)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

.PHONY: version
version: ## 🏷️ Show current version
	@echo "$(COLOR_BOLD)Version: v$(VERSION)$(COLOR_RESET)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo ""
	@echo "$(COLOR_BOLD)Kustomization tags:$(COLOR_RESET)"
	@grep -h "newTag:\|tag:" $(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml $(K8S_KUSTOMIZE_DIR)/*/kustomization.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no tags found)"

.PHONY: version-bump
version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z) - DRY pattern
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "$(COLOR_RED)❌ Usage: make version-bump NEW_VERSION=x.y.z$(COLOR_RESET)"; \
		exit 1; \
	fi
	@OLD_VERSION=$(VERSION); \
	echo "$(COLOR_BOLD)$(COLOR_BLUE)🏷️ Bumping version: $$OLD_VERSION → $(NEW_VERSION)$(COLOR_RESET)"; \
	echo ""; \
	echo "$(NEW_VERSION)" > $(VERSION_FILE); \
	echo "  ✅ Updated VERSION file"; \
	# Update base lambdaagent.yaml tag
	@if [ -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" ]; then \
		sed -i.bak 's|tag: "v[0-9.]*[-a-zA-Z0-9.]*"|tag: "v$(NEW_VERSION)"|g' "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" && rm -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml.bak"; \
		echo "  ✅ Updated base lambdaagent.yaml"; \
	fi
	# Update kustomization overlays (images section or patches)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "newTag:" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (images section)"; \
			elif grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|value: "v[0-9.]*[-a-zA-Z0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (LambdaAgent patches)"; \
			fi; \
		fi; \
	done
	@echo ""
	$(call print_success,Version bumped to $(NEW_VERSION))
	@echo ""
	@echo "$(COLOR_BOLD)Next steps:$(COLOR_RESET)"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-bruno v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

.PHONY: release release-patch release-minor release-major
release: version-bump build-local ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
	$(call print_status,🚀 Deploying v$(NEW_VERSION)...)
	@$(MAKE) deploy-$(ENV)
	$(call print_success,Released v$(NEW_VERSION) to $(ENV))

release-patch: ## 🏷️ Bump patch version (x.y.Z)
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	$(MAKE) version-bump NEW_VERSION=$$MAJOR.$$MINOR.$$NEW_PATCH

release-minor: ## 🏷️ Bump minor version (x.Y.0)
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	NEW_MINOR=$$((MINOR + 1)); \
	$(MAKE) version-bump NEW_VERSION=$$MAJOR.$$NEW_MINOR.0

release-major: ## 🏷️ Bump major version (X.0.0)
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	NEW_MAJOR=$$((MAJOR + 1)); \
	$(MAKE) version-bump NEW_VERSION=$$NEW_MAJOR.0.0

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🔨 BUILD
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: ensure-buildx-builder
ensure-buildx-builder: ## 🔧 Ensure buildx builder exists
	@if ! docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1; then \
		echo "$(COLOR_YELLOW)Creating buildx builder '$(BUILDX_BUILDER)'...$(COLOR_RESET)"; \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --bootstrap; \
	else \
		echo "$(COLOR_GREEN)✅ Buildx builder '$(BUILDX_BUILDER)' already exists$(COLOR_RESET)"; \
	fi

.PHONY: copy-shared-lib
copy-shared-lib: ## 📦 Copy shared-lib for Docker build
	@rm -rf $(SRC_DIR)/shared-lib
	@cp -r $(ROOT_DIR)/../shared-lib $(SRC_DIR)/shared-lib
	$(call print_success,Copied shared-lib)

.PHONY: build
build: copy-shared-lib ## 🐳 Build Docker image for GHCR
	$(call print_status,Building agent-bruno chatbot v$(VERSION) for GHCR...)
	docker build -t $(DOCKER_REPO):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/chatbot/Dockerfile $(SRC_DIR)/
	$(call print_success,Built $(DOCKER_REPO):$(VERSION))

.PHONY: build-local
build-local: ensure-buildx-builder copy-shared-lib ## 🐳 Build and push to local registry (arm64)
	$(call print_status,📋 Building agent-bruno chatbot version: $(VERSION))
	$(call print_status,🐳 Image: $(CHATBOT_IMAGE):v$(VERSION))
	$(call print_status,🏗️ Platform: $(LOAD_PLATFORM))
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(LOAD_PLATFORM) \
		-t $(CHATBOT_IMAGE):v$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/chatbot/Dockerfile $(SRC_DIR)/ \
		--push
	$(call print_success,Built and pushed to local registry)
	@echo "   - $(CHATBOT_IMAGE):v$(VERSION)"

.PHONY: push
push: build ## 🚀 Push to GHCR
	$(call print_status,Pushing agent-bruno image to GHCR...)
	docker push $(DOCKER_REPO):$(VERSION)
	$(call print_success,Pushed $(DOCKER_REPO):$(VERSION))

.PHONY: push-local
push-local: build-local ## 🚀 Build and push to local registry
	$(call print_success,Already pushed to local registry via build-local)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🧪 TESTING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: test
test: lint test-unit ## 🧪 Run lint + unit tests
	$(call print_success,All tests passed)

.PHONY: test-unit
test-unit: ## 🧪 Run unit tests
	$(call print_status,Running unit tests...)
	PYTHONPATH=$(SRC_DIR) $(PYTHON) -m pytest $(ROOT_DIR)/tests/ -v --cov=$(SRC_DIR) --cov-report=term-missing

.PHONY: test-k6
test-k6: ## 📊 Run k6 load tests
	$(call print_status,Running k6 load tests...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/

.PHONY: test-k6-smoke
test-k6-smoke: ## 🔥 Run k6 smoke test
	$(call print_status,Running k6 smoke test...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/k6-smoke.yaml

.PHONY: test-k6-chat
test-k6-chat: ## 💬 Run k6 chat load test
	$(call print_status,Running k6 chat load test...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/k6-chat-load.yaml

.PHONY: test-ollama
test-ollama: ## 🦙 Test Ollama connectivity
	$(call print_status,Testing Ollama...)
	@curl -s -X POST http://localhost:11434/api/generate \
		-H "Content-Type: application/json" \
		-d '{"model": "llama3.2:3b", "prompt": "Say hello in one word", "stream": false}' | jq -r '.response' 2>/dev/null || \
		echo "$(COLOR_YELLOW)Ollama not accessible. Try: kubectl port-forward -n ai-inference svc/ollama 11434:11434$(COLOR_RESET)"

.PHONY: test-chat
test-chat: ## 💬 Test chat endpoint
	$(call print_status,Testing chat endpoint...)
	@curl -s -X POST http://localhost:8080/chat \
		-H "Content-Type: application/json" \
		-d '{"message": "Hello! What can you help me with?"}' | jq 2>/dev/null || \
		echo "$(COLOR_YELLOW)Chat not accessible. Port-forward the service first.$(COLOR_RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🔍 LINTING & FORMATTING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: lint
lint: ## 🔍 Run linting
	$(call print_status,Running linters...)
	$(PYTHON) -m ruff check $(SRC_DIR)/ $(ROOT_DIR)/tests/ || true
	$(PYTHON) -m mypy $(SRC_DIR)/ --ignore-missing-imports || true

.PHONY: fmt
fmt: ## 🎨 Format code
	$(call print_status,Formatting code...)
	$(PYTHON) -m ruff check --fix $(SRC_DIR)/ $(ROOT_DIR)/tests/ || true
	$(PYTHON) -m black $(SRC_DIR)/ $(ROOT_DIR)/tests/ 2>/dev/null || true

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🚀 DEPLOYMENT
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: deploy-studio
deploy-studio: ## 🚀 Deploy to studio environment
	$(call print_status,🚀 Deploying to studio...)
	kubectl apply -k $(ROOT_DIR)/k8s/kustomize/studio/
	$(call print_success,Deployed to studio)

.PHONY: deploy-pro
deploy-pro: ## 🚀 Deploy to pro environment
	$(call print_status,🚀 Deploying to pro...)
	kubectl apply -k $(ROOT_DIR)/k8s/kustomize/pro/
	$(call print_success,Deployed to pro)

.PHONY: undeploy
undeploy: ## 🗑️ Remove deployment
	$(call print_warning,Removing deployment from $(ENV)...)
	kubectl delete -k $(ROOT_DIR)/k8s/kustomize/$(ENV)/ --ignore-not-found

.PHONY: rollout-restart
rollout-restart: ## ♻️ Restart deployment
	$(call print_status,Restarting agent-bruno...)
	kubectl delete revision -n $(NAMESPACE) -l serving.knative.dev/service=agent-bruno --ignore-not-found
	flux reconcile kustomization studio-07-apps --with-source
	$(call print_success,Rollout restart triggered)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  📊 STATUS & MONITORING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: status
status: ## 📊 Show deployment status
	$(call print_status,Deployment status:)
	@echo ""
	@echo "$(COLOR_BOLD)Knative Services:$(COLOR_RESET)"
	@kubectl get ksvc -n $(NAMESPACE) 2>/dev/null || echo "  No Knative services found"
	@echo ""
	@echo "$(COLOR_BOLD)Lambda Agents:$(COLOR_RESET)"
	@kubectl get lambdaagent -n $(NAMESPACE) 2>/dev/null || echo "  No Lambda agents found"
	@echo ""
	@echo "$(COLOR_BOLD)Pods:$(COLOR_RESET)"
	@kubectl get pods -n $(NAMESPACE) 2>/dev/null || echo "  No pods found"

.PHONY: logs
logs: ## 📜 Tail logs
	kubectl logs -n $(NAMESPACE) -l serving.knative.dev/service=agent-bruno -f --tail=100

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🔧 LOCAL DEVELOPMENT
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: install
install: ## 📦 Install Python dependencies
	$(call print_status,Installing dependencies...)
	$(PYTHON) -m pip install -r $(SRC_DIR)/requirements.txt

.PHONY: install-dev
install-dev: install ## 📦 Install development dependencies
	$(call print_status,Installing dev dependencies...)
	$(PYTHON) -m pip install -r $(ROOT_DIR)/tests/requirements.txt

.PHONY: run-dev
run-dev: ## 🏃 Run locally with hot-reload
	$(call print_status,Starting agent-bruno locally...)
	cd $(SRC_DIR) && OLLAMA_URL=http://localhost:11434 uvicorn chatbot.main:app --reload --host 0.0.0.0 --port 8080

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🧹 CLEANUP
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: clean
clean: ## 🧹 Clean build artifacts
	$(call print_warning,Cleaning...)
	rm -rf __pycache__ .pytest_cache .mypy_cache .ruff_cache .coverage htmlcov
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true

.PHONY: clean-docker
clean-docker: ## 🧹 Clean Docker images
	$(call print_warning,Cleaning Docker images...)
	docker rmi $(CHATBOT_IMAGE):v$(VERSION) 2>/dev/null || true
	docker rmi $(CHATBOT_IMAGE):latest 2>/dev/null || true
