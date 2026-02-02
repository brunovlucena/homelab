# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🔐 AGENT-CONTRACTS - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the Smart Contract Security Agent
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make build-all        - Build and push all Docker images
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

PROJECT_NAME := agent-contracts
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY ?= localhost:5001
IMAGE_PREFIX := $(REGISTRY)/agent-contracts

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/agent-contracts
ifeq ($(GIT_BRANCH),main)
	DOCKER_TAG := $(VERSION)
else ifeq ($(GIT_BRANCH),develop)
	DOCKER_TAG := $(VERSION)-beta.$(shell date +%s)
else
	DOCKER_TAG := $(VERSION)-dev.$(GIT_COMMIT)
endif

# Images
CONTRACT_FETCHER_IMAGE := $(IMAGE_PREFIX)/contract-fetcher
VULN_SCANNER_IMAGE := $(IMAGE_PREFIX)/vuln-scanner
EXPLOIT_GENERATOR_IMAGE := $(IMAGE_PREFIX)/exploit-generator
NOTIFI_ADAPTER_IMAGE := $(IMAGE_PREFIX)/notifi-adapter

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES CONFIGURATION                                            │
# └─────────────────────────────────────────────────────────────────────────┘

ENV ?= studio
NAMESPACE ?= agent-contracts

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
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🔐 Agent-Contracts$(COLOR_RESET)"
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
	@grep -h "tag:" $(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no tags found)"

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
	# Update base lambdaagent.yaml tags (multiple LambdaAgents)
	@if [ -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" ]; then \
		sed -i.bak 's|tag: "[0-9.]*"|tag: "$(NEW_VERSION)"|g' "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" && rm -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml.bak"; \
		echo "  ✅ Updated base lambdaagent.yaml (all agents)"; \
	fi
	# Update kustomization overlays (LambdaAgent patches)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|value: "[v0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (LambdaAgent patches)"; \
			fi; \
		fi; \
	done
	@echo ""
	$(call print_success,Version bumped to $(NEW_VERSION))
	@echo ""
	@echo "$(COLOR_BOLD)Next steps:$(COLOR_RESET)"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-all-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-contracts v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

.PHONY: release release-patch release-minor release-major
release: version-bump build-all-local ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
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
	@echo "$(COLOR_BLUE)🔍 Available platforms:$(COLOR_RESET)"
	@docker buildx ls | grep -A 2 $(BUILDX_BUILDER) | head -5

.PHONY: copy-shared-lib
copy-shared-lib: ## 📦 Copy shared-lib for Docker build
	@rm -rf $(SRC_DIR)/shared-lib
	@cp -r $(ROOT_DIR)/../shared-lib $(SRC_DIR)/shared-lib
	$(call print_success,Copied shared-lib)

.PHONY: build-all
build-all: copy-shared-lib build-fetcher build-scanner build-exploit-gen build-notifi-adapter ## 🐳 Build all Docker images
	$(call print_success,All images built successfully)

.PHONY: build-all-local
build-all-local: ensure-buildx-builder copy-shared-lib ## 🐳 Build and push all images to local registry (arm64)
	$(call print_status,📋 Building all agent-contracts images version: $(VERSION))
	@$(MAKE) build-fetcher-local
	@$(MAKE) build-scanner-local
	@$(MAKE) build-exploit-gen-local
	@$(MAKE) build-notifi-adapter-local
	$(call print_success,Built and pushed to local registry)
	@echo "   - $(CONTRACT_FETCHER_IMAGE):v$(VERSION)"
	@echo "   - $(VULN_SCANNER_IMAGE):v$(VERSION)"
	@echo "   - $(EXPLOIT_GENERATOR_IMAGE):v$(VERSION)"
	@echo "   - $(NOTIFI_ADAPTER_IMAGE):v$(VERSION)"

.PHONY: build-fetcher
build-fetcher: copy-shared-lib ## 🐳 Build contract-fetcher image
	$(call print_status,Building contract-fetcher v$(VERSION)...)
	docker build -t $(CONTRACT_FETCHER_IMAGE):v$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/contract_fetcher/Dockerfile $(SRC_DIR)/

.PHONY: build-fetcher-local
build-fetcher-local: ensure-buildx-builder copy-shared-lib ## 🐳 Build and push contract-fetcher to local registry
	$(call print_status,Building contract-fetcher v$(VERSION) for $(LOAD_PLATFORM)...)
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(LOAD_PLATFORM) \
		-t $(CONTRACT_FETCHER_IMAGE):v$(VERSION) \
		-t $(CONTRACT_FETCHER_IMAGE):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/contract_fetcher/Dockerfile $(SRC_DIR)/ \
		--push

.PHONY: build-scanner
build-scanner: copy-shared-lib ## 🐳 Build vuln-scanner image
	$(call print_status,Building vuln-scanner v$(VERSION)...)
	docker build -t $(VULN_SCANNER_IMAGE):v$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/vuln_scanner/Dockerfile $(SRC_DIR)/

.PHONY: build-scanner-local
build-scanner-local: ensure-buildx-builder copy-shared-lib ## 🐳 Build and push vuln-scanner to local registry
	$(call print_status,Building vuln-scanner v$(VERSION) for $(LOAD_PLATFORM)...)
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(LOAD_PLATFORM) \
		-t $(VULN_SCANNER_IMAGE):v$(VERSION) \
		-t $(VULN_SCANNER_IMAGE):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/vuln_scanner/Dockerfile $(SRC_DIR)/ \
		--push

.PHONY: build-exploit-gen
build-exploit-gen: copy-shared-lib ## 🐳 Build exploit-generator image
	$(call print_status,Building exploit-generator v$(VERSION)...)
	docker build -t $(EXPLOIT_GENERATOR_IMAGE):v$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/exploit_generator/Dockerfile $(SRC_DIR)/

.PHONY: build-exploit-gen-local
build-exploit-gen-local: ensure-buildx-builder ## 🐳 Build and push exploit-generator to local registry
	$(call print_status,Building exploit-generator v$(VERSION) for $(LOAD_PLATFORM)...)
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(LOAD_PLATFORM) \
		-t $(EXPLOIT_GENERATOR_IMAGE):v$(VERSION) \
		-t $(EXPLOIT_GENERATOR_IMAGE):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/exploit_generator/Dockerfile $(SRC_DIR)/ \
		--push

.PHONY: build-notifi-adapter
build-notifi-adapter: copy-shared-lib ## 🐳 Build notifi-adapter image
	$(call print_status,Building notifi-adapter v$(VERSION)...)
	docker build -t $(NOTIFI_ADAPTER_IMAGE):v$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/notifi_adapter/Dockerfile $(SRC_DIR)/

.PHONY: build-notifi-adapter-local
build-notifi-adapter-local: ensure-buildx-builder copy-shared-lib ## 🐳 Build and push notifi-adapter to local registry
	$(call print_status,Building notifi-adapter v$(VERSION) for $(LOAD_PLATFORM)...)
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(LOAD_PLATFORM) \
		-t $(NOTIFI_ADAPTER_IMAGE):v$(VERSION) \
		-t $(NOTIFI_ADAPTER_IMAGE):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/notifi_adapter/Dockerfile $(SRC_DIR)/ \
		--push

.PHONY: push-all
push-all: ## 🚀 Push all images to registry
	$(call print_status,Pushing all images...)
	docker push $(CONTRACT_FETCHER_IMAGE):v$(VERSION)
	docker push $(VULN_SCANNER_IMAGE):v$(VERSION)
	docker push $(EXPLOIT_GENERATOR_IMAGE):v$(VERSION)
	docker push $(NOTIFI_ADAPTER_IMAGE):v$(VERSION)
	$(call print_success,All images pushed)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🧪 TESTING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: test
test: lint test-unit ## 🧪 Run lint + unit tests
	$(call print_success,All tests passed)

.PHONY: test-unit
test-unit: ## 🧪 Run unit tests
	$(call print_status,Running unit tests...)
	cd $(ROOT_DIR) && python -m pytest tests/unit/ -v --tb=short

.PHONY: test-integration
test-integration: ## 🧪 Run integration tests (requires cluster)
	$(call print_status,Running integration tests...)
	cd $(ROOT_DIR) && python -m pytest tests/integration/ -v --tb=short

.PHONY: test-cov
test-cov: ## 🧪 Run tests with coverage
	$(call print_status,Running tests with coverage...)
	cd $(ROOT_DIR) && python -m pytest tests/ --cov=src --cov-report=html --cov-report=term

.PHONY: test-k6
test-k6: ## 📊 Run k6 load tests
	$(call print_status,Running k6 load tests...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/

.PHONY: test-k6-smoke
test-k6-smoke: ## 🔥 Run k6 smoke test
	$(call print_status,Running k6 smoke test...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/k6-smoke.yaml

.PHONY: test-k6-sre
test-k6-sre: ## 📈 Run k6 SRE metrics validation
	$(call print_status,Running k6 SRE metrics validation...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/k6-sre-metrics.yaml

.PHONY: test-k6-e2e
test-k6-e2e: ## 🔗 Run k6 E2E pipeline test
	$(call print_status,Running k6 E2E pipeline test...)
	kubectl apply -f $(ROOT_DIR)/k8s/tests/k6-e2e-pipeline.yaml

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🔍 LINTING & FORMATTING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: lint
lint: ## 🔍 Run linting
	$(call print_status,Running linters...)
	cd $(SRC_DIR) && ruff check . || true
	cd $(SRC_DIR) && mypy . --ignore-missing-imports || true

.PHONY: fmt
fmt: ## 🎨 Format code
	$(call print_status,Formatting code...)
	cd $(SRC_DIR) && black .
	cd $(SRC_DIR) && ruff check --fix . || true

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
rollout-restart: ## ♻️ Restart all deployments
	$(call print_status,Restarting all deployments...)
	kubectl delete revision -n $(NAMESPACE) -l serving.knative.dev/service=contract-fetcher --ignore-not-found
	kubectl delete revision -n $(NAMESPACE) -l serving.knative.dev/service=vuln-scanner --ignore-not-found
	kubectl delete revision -n $(NAMESPACE) -l serving.knative.dev/service=exploit-generator --ignore-not-found
	kubectl delete revision -n $(NAMESPACE) -l serving.knative.dev/service=notifi-adapter --ignore-not-found
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
logs: ## 📜 Tail logs from all agent pods
	$(call print_status,Tailing logs...)
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/part-of=agent-contracts -f --tail=100

.PHONY: logs-fetcher
logs-fetcher: ## 📜 Tail contract-fetcher logs
	kubectl logs -n $(NAMESPACE) -l serving.knative.dev/service=contract-fetcher -f --tail=100

.PHONY: logs-scanner
logs-scanner: ## 📜 Tail vuln-scanner logs
	kubectl logs -n $(NAMESPACE) -l serving.knative.dev/service=vuln-scanner -f --tail=100

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🔧 LOCAL DEVELOPMENT
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: install
install: ## 📦 Install Python dependencies
	$(call print_status,Installing dependencies...)
	cd $(SRC_DIR) && pip install -r requirements.txt

.PHONY: install-dev
install-dev: install ## 📦 Install development dependencies
	$(call print_status,Installing dev dependencies...)
	pip install pytest pytest-asyncio pytest-cov black ruff mypy httpx

.PHONY: venv
venv: ## 🐍 Create virtual environment
	$(call print_status,Creating virtual environment...)
	python3 -m venv .venv
	$(call print_warning,Activate with: source .venv/bin/activate)

.PHONY: run-fetcher
run-fetcher: ## 🏃 Run contract-fetcher locally
	$(call print_status,Running contract-fetcher...)
	cd $(SRC_DIR) && uvicorn contract_fetcher.main:app --reload --port 8081

.PHONY: run-scanner
run-scanner: ## 🏃 Run vuln-scanner locally
	$(call print_status,Running vuln-scanner...)
	cd $(SRC_DIR) && uvicorn vuln_scanner.main:app --reload --port 8082

.PHONY: run-all
run-all: ## 🏃 Run all services locally (docker-compose)
	$(call print_status,Starting all services...)
	docker-compose -f $(ROOT_DIR)/docker-compose.dev.yaml up -d

.PHONY: stop-all
stop-all: ## 🛑 Stop all local services
	$(call print_warning,Stopping all services...)
	docker-compose -f $(ROOT_DIR)/docker-compose.dev.yaml down

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🧹 CLEANUP
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: clean
clean: ## 🧹 Clean build artifacts
	$(call print_warning,Cleaning...)
	rm -rf $(SRC_DIR)/__pycache__ $(SRC_DIR)/**/__pycache__
	rm -rf $(SRC_DIR)/.pytest_cache $(SRC_DIR)/**/.pytest_cache
	rm -rf $(SRC_DIR)/.mypy_cache $(SRC_DIR)/**/.mypy_cache
	rm -rf $(SRC_DIR)/.ruff_cache
	rm -rf $(SRC_DIR)/htmlcov
	rm -rf $(ROOT_DIR)/.venv

.PHONY: clean-docker
clean-docker: ## 🧹 Clean Docker images
	$(call print_warning,Cleaning Docker images...)
	docker rmi $(CONTRACT_FETCHER_IMAGE):v$(VERSION) 2>/dev/null || true
	docker rmi $(VULN_SCANNER_IMAGE):v$(VERSION) 2>/dev/null || true
	docker rmi $(EXPLOIT_GENERATOR_IMAGE):v$(VERSION) 2>/dev/null || true
	docker rmi $(NOTIFI_ADAPTER_IMAGE):v$(VERSION) 2>/dev/null || true
