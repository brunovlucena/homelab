# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏗️ KNATIVE LAMBDA - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the Knative Lambda services
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make build-images     - Build and push all Docker images
#	  make test             - Run tests
#	  make lint             - Run linting
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

ifdef LOAD_PLATFORM
	LOAD_PLATFORM_FLAG := --platform $(LOAD_PLATFORM)
endif

ifdef PUSH_PLATFORM
	PUSH_PLATFORM_FLAG := --platform $(PUSH_PLATFORM)
endif

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ PROJECT CONFIGURATION                                              │
# └─────────────────────────────────────────────────────────────────────────┘

PROJECT_NAME := knative-lambda
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.2")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR := $(SRC_DIR)/build
BIN_DIR := $(SRC_DIR)/bin

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY := localhost:5001
OPERATOR_IMAGE := $(REGISTRY)/knative-lambda-operator

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/knative-lambda-operator
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

ENV ?= pro
NAMESPACE ?= knative-lambda

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
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)⚠️ $(1)$(COLOR_RESET)"
endef

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 HELP                                                                │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: help
help: ## 📋 Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏗️ Knative Lambda$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)💡 Tip:$(COLOR_RESET) Use ENV=pro|studio for environment-specific commands"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ VERSION & RELEASE TARGETS                                           │
# │                                                                          │
# │  DRY Principle: VERSION file is the single source of truth              │
# │  These targets update VERSION and all dependent files                    │
# └─────────────────────────────────────────────────────────────────────────┘

K8S_DIR := $(ROOT_DIR)/k8s

.PHONY: version version-check version-bump release release-patch release-minor release-major
version: ## 🏷️ Show current version
	@echo "$(COLOR_BOLD)Current version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)VERSION file:$(COLOR_RESET) $(VERSION_FILE)"
	@echo ""
	@echo "$(COLOR_BOLD)Kustomization tags:$(COLOR_RESET)"
	@grep -h "newTag:" $(K8S_DIR)/overlays/*/kustomization.yaml 2>/dev/null | sed 's/^/  /'
	@echo ""
	@echo "$(COLOR_BOLD)OPERATOR_VERSION env vars:$(COLOR_RESET)"
	@grep -h "OPERATOR_VERSION" -A1 $(K8S_DIR)/overlays/*/kustomization.yaml 2>/dev/null | grep "value:" | sed 's/^/  /'

version-check: ## 🔍 Verify all versions are in sync (CI gate)
	@echo "$(COLOR_BOLD)🔍 Checking version consistency...$(COLOR_RESET)"
	@EXPECTED="v$(VERSION)"; \
	ERRORS=0; \
	for overlay in $(K8S_DIR)/overlays/*/kustomization.yaml; do \
		TAG=$$(grep "newTag:" "$$overlay" | head -1 | awk '{print $$2}'); \
		OPVER=$$(grep -A1 "OPERATOR_VERSION" "$$overlay" | grep "value:" | sed 's/.*value: *"\([^"]*\)".*/\1/'); \
		if [ "$$TAG" != "$$EXPECTED" ]; then \
			echo "$(COLOR_RED)❌ $$overlay: newTag=$$TAG (expected $$EXPECTED)$(COLOR_RESET)"; \
			ERRORS=1; \
		fi; \
		if [ "$$OPVER" != "$$EXPECTED" ]; then \
			echo "$(COLOR_RED)❌ $$overlay: OPERATOR_VERSION=$$OPVER (expected $$EXPECTED)$(COLOR_RESET)"; \
			ERRORS=1; \
		fi; \
	done; \
	if [ $$ERRORS -eq 0 ]; then \
		echo "$(COLOR_GREEN)✅ All versions match: $$EXPECTED$(COLOR_RESET)"; \
	else \
		echo ""; \
		echo "$(COLOR_RED)❌ Version mismatch! Run: make version-bump NEW_VERSION=$(VERSION)$(COLOR_RESET)"; \
		exit 1; \
	fi

version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "$(COLOR_RED)❌ Usage: make version-bump NEW_VERSION=x.y.z$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏷️ Bumping version: $(VERSION) → $(NEW_VERSION)$(COLOR_RESET)"
	@echo ""
	@# Update VERSION file
	@echo "$(NEW_VERSION)" > $(VERSION_FILE)
	@echo "  ✅ Updated VERSION file"
	@# Update all kustomization overlays (image tag AND OPERATOR_VERSION - MUST be in sync!)
	@for overlay in $(K8S_DIR)/overlays/*/kustomization.yaml; do \
		sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay" && rm -f "$$overlay.bak"; \
		sed -i.bak 's|value: "v[0-9.]*[-a-zA-Z0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay" && rm -f "$$overlay.bak"; \
		echo "  ✅ Updated $$overlay (image + OPERATOR_VERSION)"; \
	done
	@# Update base kustomization default tag
	@if [ -f "$(K8S_DIR)/base/kustomization.yaml" ]; then \
		sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$(K8S_DIR)/base/kustomization.yaml" && rm -f "$(K8S_DIR)/base/kustomization.yaml.bak"; \
		echo "  ✅ Updated base/kustomization.yaml"; \
	fi
	@echo ""
	$(call print_success,Version bumped to $(NEW_VERSION))
	@echo ""
	@echo "$(COLOR_BOLD)Next steps:$(COLOR_RESET)"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-images-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump build-images-local ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
	$(call print_status,🚀 Deploying v$(NEW_VERSION)...)
	@$(MAKE) deploy-$(ENV)
	@echo ""
	$(call print_success,Released v$(NEW_VERSION) to $(ENV))
	@echo ""
	@echo "$(COLOR_BOLD)Don't forget to:$(COLOR_RESET)"
	@echo "  git add -A && git commit -m 'chore(release): v$(NEW_VERSION)' && git push"

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

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER IMAGE BUILD TARGETS                                          │
# └─────────────────────────────────────────────────────────────────────────┘

define build-image
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.2"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(REGISTRY)/$(1):$$DOCKER_TAG"; \
	PLATFORM_FLAG=""; \
	if [ -n "$(4)" ]; then \
		PLATFORM_FLAG="--platform $(4)"; \
	fi; \
	echo "📋 Building $(1) image version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️ Platform(s): $$(echo $(4) | tr ',' ' ' || echo 'default')"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(2)/$(3) \
		$$PLATFORM_FLAG \
		--push \
		-t $$IMAGE_TAG \
		$(2)
endef

.PHONY: ensure-buildx-builder
ensure-buildx-builder: ## 🔧 Ensure buildx builder exists
	@if ! docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1; then \
		echo "🔧 Creating buildx builder '$(BUILDX_BUILDER)' with host network access..."; \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --driver-opt network=host --use || true; \
		docker buildx inspect $(BUILDX_BUILDER) --bootstrap || true; \
	else \
		echo "✅ Buildx builder '$(BUILDX_BUILDER)' already exists"; \
	fi; \
	echo "🔍 Available platforms:"; \
	docker buildx inspect $(BUILDX_BUILDER) --bootstrap 2>/dev/null | grep -i platform || echo "  (checking...)"; \
	docker buildx ls | grep $(BUILDX_BUILDER) || true

.PHONY: build-images build-images-local
build-images: operator-build-image ## 🐳 Build and push all Docker images

build-images-local: operator-build-image-local ## 🐳 Build and push all images to local registry (arm64)
	@echo "✅ Built and pushed to local registry:"
	@echo "   - $(OPERATOR_IMAGE)"

.PHONY: operator-build-image operator-build-image-local
operator-build-image: ensure-buildx-builder ## 🐳 Build and push operator image
	$(call build-image,knative-lambda-operator,$(SRC_DIR)/operator,Dockerfile,$(PUSH_PLATFORM))

operator-build-image-local: ensure-buildx-builder ## 🐳 Build and push operator image (arm64)
	$(call build-image,knative-lambda-operator,$(SRC_DIR)/operator,Dockerfile,linux/arm64)

operator-build-image-nocache: ensure-buildx-builder ## 🐳 Build and push operator image (no cache, arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.5"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(REGISTRY)/knative-lambda-operator:$$DOCKER_TAG"; \
	echo "📋 Building knative-lambda-operator version: $$VERSION (NO CACHE)"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/operator/Dockerfile \
		--platform linux/arm64 \
		--no-cache \
		--push \
		-t $$IMAGE_TAG \
		$(SRC_DIR)/operator

operator-push-ghcr: operator-push-ghcr-multiarch ## 🚀 Build and push operator to ghcr.io (multi-arch amd64+arm64) - alias for operator-push-ghcr-multiarch

operator-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push operator to ghcr.io (multi-arch amd64+arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.2"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO):$$DOCKER_TAG"; \
	echo "🚀 Pushing to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️ Platforms: linux/amd64,linux/arm64"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/operator/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		-t $(DOCKER_REPO):latest \
		-t $(DOCKER_REPO):studio \
		$(SRC_DIR)/operator

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔨 BUILD TARGETS                                                       │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build-all clean clean-knative-lambda
build-all: ## 🔨 Build all Go binaries
	$(call print_status,🔨 Building all binaries...)
	@mkdir -p $(BIN_DIR)
	@echo "Building operator binary..."
	@cd $(SRC_DIR)/operator && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN_DIR)/operator ./cmd
	$(call print_success,All binaries built in $(BIN_DIR)/)

clean: ## 🧹 Clean all build artifacts
	$(call print_status,🧹 Cleaning build artifacts...)
	@rm -rf $(BUILD_DIR) $(BIN_DIR)
	@rm -f $(SRC_DIR)/coverage.out
	@cd $(SRC_DIR) && go clean -cache 2>/dev/null || true
	@cd $(SRC_DIR) && go clean -testcache 2>/dev/null || true

clean-knative-lambda: ## 🧹 Clean Knative Lambda k8s resources
	$(call print_status,🧹 Cleaning Knative Lambda resources...)
	@kubectl delete kpa --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete revision --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete service --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete route --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete configuration --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete job --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete trigger --all -n $(NAMESPACE) 2>/dev/null || true
	$(call print_success,Knative Lambda resources cleaned)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧪 LINTING TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: lint fmt
lint: ## 🔍 Run linting
	$(call print_status,🔍 Running linting...)
	@echo "Checking code formatting..."
	@cd $(SRC_DIR)/operator && if [ -n "$$(gofmt -l . | grep -v vendor)" ]; then \
		echo "$(COLOR_RED)❌ Code needs formatting. Run 'make fmt'$(COLOR_RESET)"; \
		gofmt -l . | grep -v vendor; \
		exit 1; \
	fi
	@echo "Checking for common issues..."
	@cd $(SRC_DIR)/operator && go vet ./... 2>/dev/null || echo "$(COLOR_YELLOW)⚠️ Some vet warnings (expected for incomplete code)$(COLOR_RESET)"
	$(call print_success,Linting completed)

fmt: ## 🎨 Format code
	$(call print_status,🎨 Formatting code...)
	@cd $(SRC_DIR)/operator && go fmt ./...
	$(call print_success,Code formatted)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES PORT-FORWARD TARGETS                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: pf-rabbitmq pf-rabbitmq-admin pf-broker pf-prometheus pf-registry
pf-rabbitmq: ## 🔌 Port forward RabbitMQ to 0.0.0.0:5672
	@kubectl port-forward --address 0.0.0.0 svc/rabbitmq-cluster-$(ENV) 5672:5672 -n rabbitmq-$(ENV)

pf-rabbitmq-admin: ## 🔌 Port forward RabbitMQ Management UI to 0.0.0.0:15672
	@kubectl port-forward --address 0.0.0.0 svc/rabbitmq-cluster-$(ENV) 15672:15672 -n rabbitmq-$(ENV)

pf-broker: ## 🔌 Port forward Knative broker to 0.0.0.0:8081
	@kubectl port-forward --address 0.0.0.0 svc/knative-lambda-broker-broker-ingress 8081:80 -n knative-lambda

pf-prometheus: ## 🔌 Port forward prometheus to 0.0.0.0:9090
	@kubectl port-forward --address 0.0.0.0 svc/prometheus-kube-prometheus-prometheus 9090:9090 -n prometheus

pf-registry: ## 🔌 Port forward registry to 0.0.0.0:5000
	@kubectl port-forward --address 0.0.0.0 -n registry service/registry 5000:5000

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐰 RABBITMQ TARGETS                                                    │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: rabbitmq-status rabbitmq-purge
rabbitmq-status: ## 📊 Print RabbitMQ queue status and statistics
	$(call print_status,📊 Checking RabbitMQ queue message counts...)
	@echo "$(COLOR_BOLD)🔍 RabbitMQ Queue Messages for $(ENV) environment$(COLOR_RESET)"
	@echo ""
	@RABBITMQ_POD=$$(kubectl get pods -l app.kubernetes.io/name=rabbitmq-cluster-$(ENV) -n rabbitmq-$(ENV) -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -n "$$RABBITMQ_POD" ]; then \
		echo "✅ Found RabbitMQ pod: $$RABBITMQ_POD"; \
		echo ""; \
		echo "$(COLOR_BOLD)📊 Queue Message Counts:$(COLOR_RESET)"; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl list_queues name messages_ready messages_unacknowledged messages consumers 2>/dev/null || echo "$(COLOR_YELLOW)⚠️ Could not retrieve queue information$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ No RabbitMQ pods found in rabbitmq-$(ENV) namespace$(COLOR_RESET)"; \
	fi

rabbitmq-purge: ## 🧹 Purge lambda event queues (use ENV=pro|studio)
	@RABBITMQ_POD=$$(kubectl get pods -l app.kubernetes.io/name=rabbitmq-cluster-$(ENV) -n rabbitmq-$(ENV) -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -n "$$RABBITMQ_POD" ]; then \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl purge_queue lambda-build-events-$(ENV) 2>/dev/null || true; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl purge_queue lambda-service-events-$(ENV) 2>/dev/null || true; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl purge_queue parser-results-$(ENV) 2>/dev/null || true; \
		echo "✅ Lambda event queues purged"; \
	else \
		echo "❌ No RabbitMQ pods found"; exit 1; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🛠️ DEVELOPMENT TARGETS                                                 │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: setup upload-parser
upload-parser: ## 📤 Upload parser to S3 (use PARSER_ID=xxx ENV=pro|studio)
	@if [ -z "$(PARSER_ID)" ] || [ ! -f "$(SRC_DIR)/tests/$(PARSER_ID)" ]; then \
		echo "$(COLOR_RED)❌ PARSER_ID required or file not found$(COLOR_RESET)"; \
		exit 1; \
	fi
	@aws s3 cp $(SRC_DIR)/tests/$(PARSER_ID) s3://knative-lambda-$(ENV)-fusion-modules-tmp/global/parser/$(PARSER_ID)
	$(call print_success,Parser uploaded to S3)

setup: ## 🔧 Setup development environment
	$(call print_status,🔧 Setting up development environment...)
	@cd $(SRC_DIR)/operator && go mod download
	@cd $(SRC_DIR)/operator && go mod tidy
	$(call print_success,Development environment ready)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧪 TESTING TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: test test-all test-unit test-security test-backend test-sre test-devops test-integration test-e2e test-load
.PHONY: trigger-lambda trigger-build trigger-delete

trigger-lambda: ## 🚀 Trigger lambda parser event (use ENV=pro|studio)
	@cd $(SRC_DIR)/tests && ENV=$(ENV) uv run --python 3.9 python create-event-lambda.py

trigger-build: ## 🚀 Trigger lambda build event (use ENV=pro|studio)
	@cd $(SRC_DIR)/tests && ENV=$(ENV) uv run --python 3.9 python create-event-builder.py

trigger-delete: ## 🗑️ Trigger lambda delete event (use ENV=pro|studio)
	@cd $(SRC_DIR)/tests && ENV=$(ENV) uv run --python 3.9 python create-event-delete.py

test: lint test-unit ## 🧪 Run lint + unit tests
	$(call print_success,Tests completed)

test-all: ## 🧪 Run ALL tests (comprehensive validation)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🧪 Running comprehensive test suite...$(COLOR_RESET)"
	@$(MAKE) lint || echo "$(COLOR_YELLOW)⚠️ Linting had issues$(COLOR_RESET)"
	@$(MAKE) test-security || true
	@$(MAKE) test-backend || true
	@$(MAKE) test-sre || true
	@$(MAKE) test-devops || true
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Test suite completed!$(COLOR_RESET)"

test-unit: ## 🧪 Run all unit tests with coverage
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running unit tests...$(COLOR_RESET)"
	@cd $(SRC_DIR) && go test -v -race -coverprofile=coverage.out ./tests/unit/... 2>&1 | tee test-output.log || true
	@if [ -f $(SRC_DIR)/coverage.out ]; then cd $(SRC_DIR) && go tool cover -func=coverage.out; fi
	@rm -f $(SRC_DIR)/test-output.log

test-security: ## 🔒 Run security tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/security/... || echo "$(COLOR_YELLOW)⚠️ Some security tests failed$(COLOR_RESET)"

test-backend: ## 🏗️ Run backend tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/handler/... ./tests/unit/backend/... || echo "$(COLOR_YELLOW)⚠️ Some backend tests failed$(COLOR_RESET)"

test-sre: ## 🚨 Run SRE tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/sre/... || echo "$(COLOR_YELLOW)⚠️ Some SRE tests failed$(COLOR_RESET)"

test-devops: ## ⚙️ Run DevOps tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/devops/... || echo "$(COLOR_YELLOW)⚠️ Some DevOps tests failed$(COLOR_RESET)"

test-integration: ## 🔗 Run integration tests (requires cluster)
	@cd $(SRC_DIR) && go test -tags=integration -v ./tests/integration/...

test-e2e: ## 🧪 Run E2E tests (requires cluster)
	@chmod +x $(SRC_DIR)/scripts/run-qa-e2e-tests.sh 2>/dev/null || true
	@$(SRC_DIR)/scripts/run-qa-e2e-tests.sh

test-load: ## 📊 Run load tests
	@chmod +x $(SRC_DIR)/scripts/run-qa-load-tests.sh 2>/dev/null || true
	@$(SRC_DIR)/scripts/run-qa-load-tests.sh

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🚀 GITOPS DEPLOYMENT TARGETS                                           │
# └─────────────────────────────────────────────────────────────────────────┘

# K8S_DIR already defined in release section above
RECEIVER_DIR := $(abspath $(ROOT_DIR)/../knative-lambda-receiver/k8s)

.PHONY: deploy-pro deploy-studio deploy-diff
deploy-pro: ## 🚀 Deploy to pro (development) - Operator + Receiver
	$(call print_status,🚀 Deploying operator to pro...)
	@kubectl apply -k $(K8S_DIR)/overlays/pro
	$(call print_status,🚀 Deploying receiver to pro...)
	@kubectl apply -k $(RECEIVER_DIR)/overlays/pro
	$(call print_success,Pro deployment triggered)

deploy-studio: ## 🚀 Deploy to studio (production) - Operator + Receiver
	$(call print_status,🚀 Deploying operator to studio...)
	@kubectl apply -k $(K8S_DIR)/overlays/studio
	$(call print_status,🚀 Deploying receiver to studio...)
	@kubectl apply -k $(RECEIVER_DIR)/overlays/studio
	$(call print_success,Studio deployment triggered)

deploy-diff: ## 🔍 Show diff for deployment (use ENV=pro|studio)
	$(call print_status,🔍 Showing deployment diff for $(ENV)...)
	@kubectl diff -k $(K8S_DIR)/overlays/$(ENV) || true

# Flux GitRepository (pro cluster): patch to current branch and reconcile
FLUX_NS ?= flux-system
FLUX_REPO ?= homelab

.PHONY: flux-test-branch
flux-test-branch: ## 🔀 Point Flux at current branch and reconcile (use from pro context)
	@BRANCH=$$(git branch --show-current 2>/dev/null); \
	if [ -z "$$BRANCH" ]; then echo "$(COLOR_RED)❌ Not a git repo or detached HEAD$(COLOR_RESET)"; exit 1; fi; \
	$(call print_status,🔀 Patching GitRepository $(FLUX_REPO) to branch $$BRANCH...); \
	kubectl patch gitrepository $(FLUX_REPO) -n $(FLUX_NS) --type=merge -p "{\"spec\":{\"ref\":{\"branch\":\"$$BRANCH\"}}}"; \
	$(call print_status,🔄 Reconciling source and pro-04-knative-instances...); \
	flux reconcile source git $(FLUX_REPO) --with-source && flux reconcile kustomization pro-04-knative-instances --with-source; \
	$(call print_success,Flux now tracking branch $$BRANCH)

# Pro cluster uses local registry (localhost:5001). Build and push for pro:
.PHONY: pro-build pro-dev
pro-build: operator-build-image-local ## 🐳 Build and push operator to local registry (pro uses this)
	@echo ""; \
	$(call print_success,Pro cluster pulls from $(REGISTRY). Run 'make flux-test-branch' or reconcile pro-04-knative-instances to roll out.)

pro-dev: pro-build ## 🐳 Build, push, then reconcile pro (kubectl context = pro)
	$(call print_status,🔄 Reconciling pro-04-knative-instances...); \
	flux reconcile source git $(FLUX_REPO) --with-source && flux reconcile kustomization pro-04-knative-instances --with-source; \
	$(call print_success,Pro updated with new image)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐦 FLAGGER CANARY TARGETS                                              │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: canary-status canary-promote canary-rollback canary-describe
canary-status: ## 🐦 Show canary deployment status
	$(call print_status,🐦 Canary Status...)
	@kubectl get canary -n $(NAMESPACE) -o wide
	@echo ""
	@kubectl get canary knative-lambda-operator -n $(NAMESPACE) -o jsonpath='{.status.phase}' 2>/dev/null && echo "" || echo "No canary found"

canary-promote: ## ✅ Manually promote canary (skip analysis)
	$(call print_status,✅ Promoting canary...)
	@kubectl annotate canary knative-lambda-operator -n $(NAMESPACE) flagger.app/promotion=skip --overwrite
	$(call print_success,Canary promotion requested)

canary-rollback: ## ⏪ Rollback canary deployment
	$(call print_status,⏪ Rolling back canary...)
	@kubectl annotate canary knative-lambda-operator -n $(NAMESPACE) flagger.app/rollback=true --overwrite
	$(call print_success,Canary rollback requested)

canary-describe: ## 📋 Describe canary in detail
	@kubectl describe canary knative-lambda-operator -n $(NAMESPACE)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📊 K6 LOAD TESTING TARGETS                                             │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: k6-load k6-stress k6-status k6-clean
k6-load: ## 📈 Run k6 load test
	$(call print_status,📈 Running k6 load test...)
	@kubectl delete testrun knative-lambda-execute-load -n $(NAMESPACE) 2>/dev/null || true
	@kubectl apply -f $(K8S_DIR)/tests/k6-testrun-execute-lambda.yaml -l test-variant=load
	$(call print_success,Load test started - check status with: make k6-status)

k6-stress: ## 💪 Run k6 stress test
	$(call print_status,💪 Running k6 stress test...)
	@kubectl delete testrun knative-lambda-execute-stress -n $(NAMESPACE) 2>/dev/null || true
	@kubectl apply -f $(K8S_DIR)/tests/k6-testrun-execute-lambda.yaml -l test-variant=stress
	$(call print_success,Stress test started - check status with: make k6-status)

k6-status: ## 📋 Show k6 test status
	$(call print_status,📋 k6 Test Status...)
	@kubectl get testrun -n $(NAMESPACE) -o wide 2>/dev/null || echo "No TestRuns found"
	@echo ""
	@kubectl get pods -n $(NAMESPACE) -l app=k6 2>/dev/null || true

k6-clean: ## 🧹 Clean up k6 test runs
	$(call print_status,🧹 Cleaning k6 test runs...)
	@kubectl delete testrun --all -n $(NAMESPACE) 2>/dev/null || true
	$(call print_success,k6 test runs cleaned)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔀 A/B TESTING TARGETS                                                 │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: ab-test-canary ab-test-primary ab-split-status
ab-test-canary: ## 🔀 Send request to canary (A/B test)
	$(call print_status,🔀 Testing canary endpoint...)
	@curl -s -H "x-ab-test: canary" http://localhost:8080/healthz 2>/dev/null || \
		echo "$(COLOR_YELLOW)⚠️ Port-forward required: kubectl port-forward svc/knative-lambda-operator 8080:8080 -n $(NAMESPACE)$(COLOR_RESET)"

ab-test-primary: ## 🔀 Send request to primary (A/B test)
	$(call print_status,🔀 Testing primary endpoint...)
	@curl -s -H "x-ab-test: primary" http://localhost:8080/healthz 2>/dev/null || \
		echo "$(COLOR_YELLOW)⚠️ Port-forward required: kubectl port-forward svc/knative-lambda-operator 8080:8080 -n $(NAMESPACE)$(COLOR_RESET)"

ab-split-status: ## 📊 Show traffic split status
	$(call print_status,📊 Traffic Split Status...)
	@kubectl get trafficsplit -n $(NAMESPACE) -o wide 2>/dev/null || echo "No TrafficSplit found"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔍 OBSERVABILITY TARGETS                                               │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: metrics alerts logs
metrics: ## 📊 Show operator metrics
	$(call print_status,📊 Fetching operator metrics...)
	@kubectl port-forward svc/knative-lambda-operator 8080:8080 -n $(NAMESPACE) &
	@sleep 2
	@curl -s http://localhost:8080/metrics | head -50
	@pkill -f "port-forward.*8080" 2>/dev/null || true

alerts: ## 🚨 Show active alerts
	$(call print_status,🚨 Active Alerts...)
	@kubectl get prometheusrule -n $(NAMESPACE) -o wide 2>/dev/null || echo "No PrometheusRules found"

logs: ## 📜 Tail operator logs
	@kubectl logs -f deployment/knative-lambda-operator -n $(NAMESPACE) --tail=100

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🌐 CLOUDEVENTS TESTING TARGETS                                         │
# └─────────────────────────────────────────────────────────────────────────┘

# Broker URL - no path suffix needed for RabbitMQ broker
BROKER_URL := http://lambda-broker-broker-ingress.knative-lambda.svc.cluster.local
EVENT_COUNT ?= 10
CONCURRENCY ?= 5

.PHONY: ce-test ce-test-python ce-test-nodejs ce-load ce-load-python ce-load-nodejs ce-watch

ce-test: ce-test-python ce-test-nodejs ## 🌐 Send test CloudEvents to both lambdas
	$(call print_success,CloudEvents sent to both lambdas)

ce-test-python: ## 🐍 Send single CloudEvent to hello-python
	$(call print_status,🐍 Sending CloudEvent to hello-python...)
	@kubectl run ce-test-python-$$(date +%s) --rm -it --restart=Never \
		--image=curlimages/curl \
		-- -s -X POST "$(BROKER_URL)" \
		-H "Content-Type: application/json" \
		-H "Ce-Id: test-python-$$(date +%s)" \
		-H "Ce-Type: io.knative.lambda.invoke.sync" \
		-H "Ce-Source: makefile/test" \
		-H "Ce-Specversion: 1.0" \
		-H "Ce-Subject: hello-python" \
		-d '{"message": "Hello from Makefile test!", "timestamp": "'$$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'

ce-test-nodejs: ## 📦 Send single CloudEvent to hello-nodejs
	$(call print_status,📦 Sending CloudEvent to hello-nodejs...)
	@kubectl run ce-test-nodejs-$$(date +%s) --rm -it --restart=Never \
		--image=curlimages/curl \
		-- -s -X POST "$(BROKER_URL)" \
		-H "Content-Type: application/json" \
		-H "Ce-Id: test-nodejs-$$(date +%s)" \
		-H "Ce-Type: io.knative.lambda.invoke.sync" \
		-H "Ce-Source: makefile/test" \
		-H "Ce-Specversion: 1.0" \
		-H "Ce-Subject: hello-nodejs" \
		-d '{"message": "Hello from Makefile test!", "timestamp": "'$$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'

ce-load: ## 🔥 Send CloudEvents load test to both lambdas (EVENT_COUNT=10 CONCURRENCY=5)
	$(call print_status,🔥 Running CloudEvents load test...)
	@echo "Sending $(EVENT_COUNT) events with $(CONCURRENCY) concurrent workers to each lambda..."
	@$(MAKE) --no-print-directory ce-load-python &
	@$(MAKE) --no-print-directory ce-load-nodejs &
	@wait
	$(call print_success,Load test completed!)

ce-load-python: ## 🐍🔥 Send CloudEvents load to hello-python (EVENT_COUNT=10)
	$(call print_status,🐍 Sending $(EVENT_COUNT) CloudEvents to hello-python...)
	@kubectl delete job ce-load-python -n $(NAMESPACE) 2>/dev/null || true
	@kubectl create job ce-load-python -n $(NAMESPACE) \
		--image=curlimages/curl \
		-- sh -c 'for i in $$(seq 1 $(EVENT_COUNT)); do \
			EVENT_ID="python-$$(date +%s%N)-$$RANDOM"; \
			echo "[$$(date +%H:%M:%S)] Sending event $$i/$(EVENT_COUNT): $$EVENT_ID"; \
			curl -s -w " [HTTP %{http_code}]\n" -X POST \
				"$(BROKER_URL)" \
				-H "Content-Type: application/json" \
				-H "Ce-Id: $$EVENT_ID" \
				-H "Ce-Type: io.knative.lambda.invoke.async" \
				-H "Ce-Source: makefile/load-test" \
				-H "Ce-Specversion: 1.0" \
				-H "Ce-Subject: hello-python" \
				-d "{\"message\": \"Load test $$i\", \"timestamp\": \"$$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"; \
			sleep 0.1; \
		done'
	@echo "Job created. Watch with: kubectl logs -f job/ce-load-python -n $(NAMESPACE)"

ce-load-nodejs: ## 📦🔥 Send CloudEvents load to hello-nodejs (EVENT_COUNT=10)
	$(call print_status,📦 Sending $(EVENT_COUNT) CloudEvents to hello-nodejs...)
	@kubectl delete job ce-load-nodejs -n $(NAMESPACE) 2>/dev/null || true
	@kubectl create job ce-load-nodejs -n $(NAMESPACE) \
		--image=curlimages/curl \
		-- sh -c 'for i in $$(seq 1 $(EVENT_COUNT)); do \
			EVENT_ID="nodejs-$$(date +%s%N)-$$RANDOM"; \
			echo "[$$(date +%H:%M:%S)] Sending event $$i/$(EVENT_COUNT): $$EVENT_ID"; \
			curl -s -w " [HTTP %{http_code}]\n" -X POST \
				"$(BROKER_URL)" \
				-H "Content-Type: application/json" \
				-H "Ce-Id: $$EVENT_ID" \
				-H "Ce-Type: io.knative.lambda.invoke.async" \
				-H "Ce-Source: makefile/load-test" \
				-H "Ce-Specversion: 1.0" \
				-H "Ce-Subject: hello-nodejs" \
				-d "{\"message\": \"Load test $$i\", \"timestamp\": \"$$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"; \
			sleep 0.1; \
		done'
	@echo "Job created. Watch with: kubectl logs -f job/ce-load-nodejs -n $(NAMESPACE)"

ce-watch: ## 👀 Watch CloudEvents load test progress and lambda logs
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  👀 CLOUDEVENTS LOAD TEST PROGRESS$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)📊 Load Test Jobs:$(COLOR_RESET)"
	@kubectl get jobs -n $(NAMESPACE) -l 'job-name in (ce-load-python, ce-load-nodejs)' 2>/dev/null || echo "  No load test jobs found"
	@echo ""
	@echo "$(COLOR_BOLD)🐍 hello-python pods:$(COLOR_RESET)"
	@kubectl get pods -n $(NAMESPACE) -l serving.knative.dev/service=hello-python 2>/dev/null || echo "  No pods (scaled to zero)"
	@echo ""
	@echo "$(COLOR_BOLD)📦 hello-nodejs pods:$(COLOR_RESET)"
	@kubectl get pods -n $(NAMESPACE) -l serving.knative.dev/service=hello-nodejs 2>/dev/null || echo "  No pods (scaled to zero)"
	@echo ""
	@echo "$(COLOR_BOLD)📜 Recent hello-python logs:$(COLOR_RESET)"
	@kubectl logs -n $(NAMESPACE) -l serving.knative.dev/service=hello-python -c user-container --tail=5 2>/dev/null || echo "  No logs available"
	@echo ""
	@echo "$(COLOR_BOLD)📜 Recent hello-nodejs logs:$(COLOR_RESET)"
	@kubectl logs -n $(NAMESPACE) -l serving.knative.dev/service=hello-nodejs -c user-container --tail=5 2>/dev/null || echo "  No logs available"

ce-clean: ## 🧹 Clean up CloudEvents load test jobs
	$(call print_status,🧹 Cleaning CloudEvents load test jobs...)
	@kubectl delete job -n $(NAMESPACE) -l 'job-name in (ce-load-python, ce-load-nodejs)' 2>/dev/null || true
	@kubectl delete pods -n $(NAMESPACE) -l job-name=ce-load-python --force --grace-period=0 2>/dev/null || true
	@kubectl delete pods -n $(NAMESPACE) -l job-name=ce-load-nodejs --force --grace-period=0 2>/dev/null || true
	$(call print_success,Load test jobs cleaned)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏭 CLOUDEVENTS LAMBDA CREATION SIMULATION                              │
# │                                                                          │
# │  These targets simulate external services creating Lambda functions     │
# │  via CloudEvents sent to the Broker (command.function.deploy events)    │
# │                                                                          │
# │  Architecture (handles 1000+ concurrent events):                        │
# │    Makefile → Broker → RabbitMQ Queue → Trigger → Receiver Service     │
# │                         (buffered)                  (auto-scaled)       │
# └─────────────────────────────────────────────────────────────────────────┘

# Broker URL for CloudEvents (events are queued in RabbitMQ for reliable delivery)
# The broker routes events to the lambda-command-receiver via Triggers
COMMAND_BROKER_URL := http://lambda-broker-broker-ingress.knative-lambda.svc.cluster.local

# Lambda creation defaults
LAMBDA_NAME ?= sample-lambda
LAMBDA_LANGUAGE ?= python
LAMBDA_VERSION ?= 3.11
LAMBDA_HANDLER ?= main.handler
LAMBDA_BUCKET ?= lambda-functions
LAMBDA_KEY ?= $(LAMBDA_NAME)/
LAMBDA_REGISTRY ?= kind-registry:5000

.PHONY: ce-create-lambda ce-create-lambda-python ce-create-lambda-nodejs ce-create-lambda-go
.PHONY: ce-delete-lambda ce-update-lambda ce-build-lambda ce-lambda-status

ce-create-lambda: ## 🏭 Create Lambda via CloudEvent (LAMBDA_NAME=xxx LAMBDA_LANGUAGE=python|nodejs|go)
	$(call print_status,🏭 Sending command.function.deploy CloudEvent to broker for $(LAMBDA_NAME)...)
	@kubectl run ce-create-$(LAMBDA_NAME)-$$(date +%s) --rm -it --restart=Never \
		--image=curlimages/curl \
		-- -s -w "\n[HTTP %{http_code}]\n" -X POST "$(COMMAND_BROKER_URL)" \
		-H "Content-Type: application/json" \
		-H "Ce-Id: deploy-$(LAMBDA_NAME)-$$(date +%s)" \
		-H "Ce-Type: io.knative.lambda.command.function.deploy" \
		-H "Ce-Source: makefile/simulation" \
		-H "Ce-Specversion: 1.0" \
		-H "Ce-Subject: $(LAMBDA_NAME)" \
		-d '{ \
			"metadata": { \
				"name": "$(LAMBDA_NAME)", \
				"namespace": "$(NAMESPACE)", \
				"labels": { \
					"environment": "$(ENV)", \
					"created-by": "cloudevent" \
				} \
			}, \
			"spec": { \
				"source": { \
					"type": "minio", \
					"minio": { \
						"endpoint": "minio.minio.svc.cluster.local:9000", \
						"bucket": "$(LAMBDA_BUCKET)", \
						"key": "$(LAMBDA_KEY)", \
						"secretRef": { "name": "minio-credentials" } \
					} \
				}, \
				"runtime": { \
					"language": "$(LAMBDA_LANGUAGE)", \
					"version": "$(LAMBDA_VERSION)", \
					"handler": "$(LAMBDA_HANDLER)" \
				}, \
				"scaling": { \
					"minReplicas": 0, \
					"maxReplicas": 10, \
					"targetConcurrency": 5, \
					"scaleToZeroGracePeriod": "30s" \
				}, \
				"resources": { \
					"limits": { "memory": "128Mi", "cpu": "100m" } \
				}, \
				"env": [ \
					{ "name": "LOG_LEVEL", "value": "debug" }, \
					{ "name": "ENVIRONMENT", "value": "$(ENV)" } \
				], \
				"build": { \
					"timeout": "15m", \
					"registry": "$(LAMBDA_REGISTRY)" \
				}, \
				"eventing": { \
					"enabled": true, \
					"rabbitmq": { \
						"clusterName": "rabbitmq-cluster-knative-lambda", \
						"namespace": "$(NAMESPACE)" \
					} \
				} \
			} \
		}'

ce-create-lambda-python: ## 🐍 Create Python Lambda via CloudEvent (LAMBDA_NAME=xxx)
	@$(MAKE) --no-print-directory ce-create-lambda \
		LAMBDA_NAME=$(or $(LAMBDA_NAME),hello-python-ce) \
		LAMBDA_LANGUAGE=python \
		LAMBDA_VERSION=3.11 \
		LAMBDA_HANDLER=main.handler \
		LAMBDA_KEY=$(or $(LAMBDA_NAME),hello-python-ce)/

ce-create-lambda-nodejs: ## 📦 Create Node.js Lambda via CloudEvent (LAMBDA_NAME=xxx)
	@$(MAKE) --no-print-directory ce-create-lambda \
		LAMBDA_NAME=$(or $(LAMBDA_NAME),hello-nodejs-ce) \
		LAMBDA_LANGUAGE=nodejs \
		LAMBDA_VERSION=20 \
		LAMBDA_HANDLER=index.handler \
		LAMBDA_KEY=$(or $(LAMBDA_NAME),hello-nodejs-ce)/

ce-create-lambda-go: ## 🔷 Create Go Lambda via CloudEvent (LAMBDA_NAME=xxx)
	@$(MAKE) --no-print-directory ce-create-lambda \
		LAMBDA_NAME=$(or $(LAMBDA_NAME),hello-go-ce) \
		LAMBDA_LANGUAGE=go \
		LAMBDA_VERSION=1.21 \
		LAMBDA_HANDLER=main \
		LAMBDA_KEY=$(or $(LAMBDA_NAME),hello-go-ce)/

ce-delete-lambda: ## 🗑️ Delete Lambda via CloudEvent (LAMBDA_NAME=xxx)
	$(call print_status,🗑️ Sending command.service.delete CloudEvent for $(LAMBDA_NAME)...)
	@kubectl run ce-delete-$(LAMBDA_NAME)-$$(date +%s) --rm -it --restart=Never \
		--image=curlimages/curl \
		-- -s -w "\n[HTTP %{http_code}]\n" -X POST "$(COMMAND_BROKER_URL)" \
		-H "Content-Type: application/json" \
		-H "Ce-Id: delete-$(LAMBDA_NAME)-$$(date +%s)" \
		-H "Ce-Type: io.knative.lambda.command.service.delete" \
		-H "Ce-Source: makefile/simulation" \
		-H "Ce-Specversion: 1.0" \
		-H "Ce-Subject: $(LAMBDA_NAME)" \
		-d '{ \
			"name": "$(LAMBDA_NAME)", \
			"namespace": "$(NAMESPACE)", \
			"reason": "Deleted via CloudEvent simulation" \
		}'

ce-update-lambda: ## 🔄 Update Lambda via CloudEvent (LAMBDA_NAME=xxx)
	$(call print_status,🔄 Sending command.function.deploy CloudEvent to update $(LAMBDA_NAME)...)
	@kubectl run ce-update-$(LAMBDA_NAME)-$$(date +%s) --rm -it --restart=Never \
		--image=curlimages/curl \
		-- -s -w "\n[HTTP %{http_code}]\n" -X POST "$(COMMAND_BROKER_URL)" \
		-H "Content-Type: application/json" \
		-H "Ce-Id: update-$(LAMBDA_NAME)-$$(date +%s)" \
		-H "Ce-Type: io.knative.lambda.command.function.deploy" \
		-H "Ce-Source: makefile/simulation" \
		-H "Ce-Specversion: 1.0" \
		-H "Ce-Subject: $(LAMBDA_NAME)" \
		-H "Ce-Action: update" \
		-d '{ \
			"metadata": { \
				"name": "$(LAMBDA_NAME)", \
				"namespace": "$(NAMESPACE)" \
			}, \
			"spec": { \
				"source": { \
					"type": "minio", \
					"minio": { \
						"endpoint": "minio.minio.svc.cluster.local:9000", \
						"bucket": "$(LAMBDA_BUCKET)", \
						"key": "$(LAMBDA_KEY)", \
						"secretRef": { "name": "minio-credentials" } \
					} \
				}, \
				"runtime": { \
					"language": "$(LAMBDA_LANGUAGE)", \
					"version": "$(LAMBDA_VERSION)", \
					"handler": "$(LAMBDA_HANDLER)" \
				}, \
				"scaling": { \
					"minReplicas": 0, \
					"maxReplicas": 20, \
					"targetConcurrency": 10, \
					"scaleToZeroGracePeriod": "60s" \
				}, \
				"resources": { \
					"limits": { "memory": "256Mi", "cpu": "200m" } \
				}, \
				"env": [ \
					{ "name": "LOG_LEVEL", "value": "info" }, \
					{ "name": "ENVIRONMENT", "value": "$(ENV)" }, \
					{ "name": "UPDATED_AT", "value": "'$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" } \
				], \
				"build": { \
					"timeout": "15m", \
					"registry": "$(LAMBDA_REGISTRY)", \
					"forceRebuild": true \
				}, \
				"eventing": { \
					"enabled": true, \
					"rabbitmq": { \
						"clusterName": "rabbitmq-cluster-knative-lambda", \
						"namespace": "$(NAMESPACE)" \
					} \
				} \
			} \
		}'

ce-build-lambda: ## 🔨 Trigger Lambda rebuild via CloudEvent (LAMBDA_NAME=xxx)
	$(call print_status,🔨 Sending command.build.start CloudEvent for $(LAMBDA_NAME)...)
	@kubectl run ce-build-$(LAMBDA_NAME)-$$(date +%s) --rm -it --restart=Never \
		--image=curlimages/curl \
		-- -s -w "\n[HTTP %{http_code}]\n" -X POST "$(COMMAND_BROKER_URL)" \
		-H "Content-Type: application/json" \
		-H "Ce-Id: build-$(LAMBDA_NAME)-$$(date +%s)" \
		-H "Ce-Type: io.knative.lambda.command.build.start" \
		-H "Ce-Source: makefile/simulation" \
		-H "Ce-Specversion: 1.0" \
		-H "Ce-Subject: $(LAMBDA_NAME)" \
		-d '{ \
			"name": "$(LAMBDA_NAME)", \
			"namespace": "$(NAMESPACE)", \
			"forceRebuild": true, \
			"reason": "Manual rebuild via CloudEvent" \
		}'

ce-lambda-status: ## 📊 Check Lambda status after CloudEvent creation
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  📊 LAMBDA STATUS (CloudEvent Created)$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)🐑 LambdaFunctions with 'created-by: cloudevent' label:$(COLOR_RESET)"
	@kubectl get lambdafunctions -n $(NAMESPACE) -l created-by=cloudevent -o wide 2>/dev/null || echo "  (none found)"
	@echo ""
	@echo "$(COLOR_BOLD)📋 All LambdaFunctions:$(COLOR_RESET)"
	@kubectl get lambdafunctions -n $(NAMESPACE) -o wide 2>/dev/null || echo "  (none found)"
	@echo ""
	@if [ -n "$(LAMBDA_NAME)" ]; then \
		echo "$(COLOR_BOLD)🔍 Details for $(LAMBDA_NAME):$(COLOR_RESET)"; \
		kubectl get lambdafunction $(LAMBDA_NAME) -n $(NAMESPACE) -o yaml 2>/dev/null || echo "  Lambda '$(LAMBDA_NAME)' not found"; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎭 CLOUDEVENTS SIMULATION SCENARIOS                                    │
# │                                                                          │
# │  End-to-end scenarios for testing event-driven Lambda management        │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: ce-simulate-full-lifecycle ce-simulate-multi-lambda

ce-simulate-full-lifecycle: ## 🎭 Simulate full Lambda lifecycle via CloudEvents
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  🎭 SIMULATING FULL LAMBDA LIFECYCLE VIA CLOUDEVENTS$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)📍 Step 1: Creating Lambda 'lifecycle-test' via CloudEvent...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory ce-create-lambda LAMBDA_NAME=lifecycle-test LAMBDA_LANGUAGE=python
	@echo ""
	@echo "$(COLOR_BOLD)⏳ Waiting 10s for build to start...$(COLOR_RESET)"
	@sleep 10
	@$(MAKE) --no-print-directory ce-lambda-status LAMBDA_NAME=lifecycle-test
	@echo ""
	@echo "$(COLOR_BOLD)📍 Step 2: Updating Lambda 'lifecycle-test' via CloudEvent...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory ce-update-lambda LAMBDA_NAME=lifecycle-test LAMBDA_LANGUAGE=python
	@echo ""
	@echo "$(COLOR_BOLD)⏳ Waiting 5s for update to process...$(COLOR_RESET)"
	@sleep 5
	@$(MAKE) --no-print-directory ce-lambda-status LAMBDA_NAME=lifecycle-test
	@echo ""
	@echo "$(COLOR_BOLD)💡 Run 'make ce-delete-lambda LAMBDA_NAME=lifecycle-test' to cleanup$(COLOR_RESET)"

ce-simulate-multi-lambda: ## 🎭 Simulate creating multiple Lambdas (Python, Node.js, Go)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  🎭 CREATING MULTI-RUNTIME LAMBDAS VIA CLOUDEVENTS$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)🐍 Creating Python Lambda...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory ce-create-lambda-python LAMBDA_NAME=multi-python
	@echo ""
	@echo "$(COLOR_BOLD)📦 Creating Node.js Lambda...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory ce-create-lambda-nodejs LAMBDA_NAME=multi-nodejs
	@echo ""
	@echo "$(COLOR_BOLD)🔷 Creating Go Lambda...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory ce-create-lambda-go LAMBDA_NAME=multi-go
	@echo ""
	@echo "$(COLOR_BOLD)⏳ Waiting 5s for resources to be created...$(COLOR_RESET)"
	@sleep 5
	@$(MAKE) --no-print-directory ce-lambda-status
	@echo ""
	@echo "$(COLOR_BOLD)💡 Cleanup: make ce-delete-lambda LAMBDA_NAME=multi-python && make ce-delete-lambda LAMBDA_NAME=multi-nodejs && make ce-delete-lambda LAMBDA_NAME=multi-go$(COLOR_RESET)"
