# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏗️ KNATIVE LAMBDA - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the Knative Lambda services
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make operator-build-image-local  - Build and push operator to local registry
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
K8S_DIR := $(ROOT_DIR)/k8s
FLUX_NS ?= flux-system
FLUX_REPO ?= homelab

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

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 TARGETS (alphabetical)                                              │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build-all clean clean-knative-lambda ensure-buildx-builder fmt
.PHONY: pro-dev-branch-flux help k6-clean k6-load k6-status k6-stress lint logs
.PHONY: metrics operator-build-image operator-build-image-local operator-build-image-nocache
.PHONY: operator-push-ghcr operator-push-ghcr-multiarch pf-broker pf-prometheus
.PHONY: pf-rabbitmq pf-rabbitmq-admin pro-dev-build pro-dev-build-push
.PHONY: rabbitmq-purge rabbitmq-status release release-major release-minor release-patch
.PHONY: test test-all test-backend test-devops test-e2e test-integration test-load
.PHONY: test-security test-sre test-unit trigger-build trigger-delete trigger-lambda
.PHONY: upload-parser version version-bump

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
	$(call print_success,Build artifacts cleaned)

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

fmt: ## 🎨 Format code
	$(call print_status,🎨 Formatting code...)
	@cd $(SRC_DIR)/operator && go fmt ./...
	$(call print_success,Code formatted)

pro-dev-branch-flux: ## 🔀 Point Flux at current branch and reconcile (use from pro context)
	@BRANCH=$$(git branch --show-current 2>/dev/null); \
	if [ -z "$$BRANCH" ]; then echo "$(COLOR_RED)❌ Not a git repo or detached HEAD$(COLOR_RESET)"; exit 1; fi; \
	$(call print_status,🔀 Patching GitRepository $(FLUX_REPO) to branch $$BRANCH...); \
	kubectl patch gitrepository $(FLUX_REPO) -n $(FLUX_NS) --type=merge -p "{\"spec\":{\"ref\":{\"branch\":\"$$BRANCH\"}}}"; \
	$(call print_status,🔄 Reconciling source and pro-04-knative-instances...); \
	flux reconcile source git $(FLUX_REPO) --with-source && flux reconcile kustomization pro-04-knative-instances --with-source; \
	$(call print_success,Flux now tracking branch $$BRANCH)

help: ## 📋 Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏗️ Knative Lambda$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -t: -k1,1 | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)💡 Tip:$(COLOR_RESET) Use ENV=pro|studio for environment-specific commands"

k6-clean: ## 🧹 Clean up k6 test runs
	$(call print_status,🧹 Cleaning k6 test runs...)
	@kubectl delete testrun --all -n $(NAMESPACE) 2>/dev/null || true
	$(call print_success,k6 test runs cleaned)

k6-load: ## 📈 Run k6 load test
	$(call print_status,📈 Running k6 load test...)
	@kubectl delete testrun knative-lambda-execute-load -n $(NAMESPACE) 2>/dev/null || true
	@kubectl apply -f $(K8S_DIR)/tests/k6-testrun-execute-lambda.yaml -l test-variant=load
	$(call print_success,Load test started - check status with: make k6-status)

k6-status: ## 📋 Show k6 test status
	$(call print_status,📋 k6 Test Status...)
	@kubectl get testrun -n $(NAMESPACE) -o wide 2>/dev/null || echo "No TestRuns found"
	@echo ""
	@kubectl get pods -n $(NAMESPACE) -l app=k6 2>/dev/null || true

k6-stress: ## 💪 Run k6 stress test
	$(call print_status,💪 Running k6 stress test...)
	@kubectl delete testrun knative-lambda-execute-stress -n $(NAMESPACE) 2>/dev/null || true
	@kubectl apply -f $(K8S_DIR)/tests/k6-testrun-execute-lambda.yaml -l test-variant=stress
	$(call print_success,Stress test started - check status with: make k6-status)

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

logs: ## 📜 Tail operator logs
	@kubectl logs -f deployment/knative-lambda-operator -n $(NAMESPACE) --tail=100

metrics: ## 📊 Show operator metrics
	$(call print_status,📊 Fetching operator metrics...)
	@kubectl port-forward svc/knative-lambda-operator 8080:8080 -n $(NAMESPACE) &
	@sleep 2
	@curl -s http://localhost:8080/metrics | head -50
	@pkill -f "port-forward.*8080" 2>/dev/null || true

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

operator-push-ghcr: operator-push-ghcr-multiarch ## 🚀 Build and push operator to ghcr.io (multi-arch) - alias

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

pf-broker: ## 🔌 Port forward Knative broker to 0.0.0.0:8081
	@kubectl port-forward --address 0.0.0.0 svc/knative-lambda-broker-broker-ingress 8081:80 -n knative-lambda

pf-prometheus: ## 🔌 Port forward prometheus to 0.0.0.0:9090
	@kubectl port-forward --address 0.0.0.0 svc/prometheus-kube-prometheus-prometheus 9090:9090 -n prometheus

pf-rabbitmq: ## 🔌 Port forward RabbitMQ to 0.0.0.0:5672
	@kubectl port-forward --address 0.0.0.0 svc/rabbitmq-cluster-$(ENV) 5672:5672 -n rabbitmq-$(ENV)

pf-rabbitmq-admin: ## 🔌 Port forward RabbitMQ Management UI to 0.0.0.0:15672
	@kubectl port-forward --address 0.0.0.0 svc/rabbitmq-cluster-$(ENV) 15672:15672 -n rabbitmq-$(ENV)

pro-dev-build: operator-build-image-local ## 🐳 Build and push operator to local registry (pro uses this)
	@echo ""; \
	$(call print_success,Pro cluster pulls from $(REGISTRY). Run 'make pro-dev-branch-flux' or reconcile pro-04-knative-instances to roll out.)

pro-dev-build-push: pro-dev-build ## 🐳 Build, push, then reconcile pro (kubectl context = pro)
	$(call print_status,🔄 Reconciling pro-04-knative-instances...); \
	flux reconcile source git $(FLUX_REPO) --with-source && flux reconcile kustomization pro-04-knative-instances --with-source; \
	$(call print_success,Pro updated with new image)

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

release: version-bump operator-build-image-local ## 🚀 Full release: bump version + build (NEW_VERSION=x.y.z)
	$(call print_success,Released v$(NEW_VERSION) - commit and push to deploy)

release-major: ## 🏷️ Bump major version (X.0.0)
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	NEW_MAJOR=$$((MAJOR + 1)); \
	$(MAKE) version-bump NEW_VERSION=$$NEW_MAJOR.0.0

release-minor: ## 🏷️ Bump minor version (x.Y.0)
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	NEW_MINOR=$$((MINOR + 1)); \
	$(MAKE) version-bump NEW_VERSION=$$MAJOR.$$NEW_MINOR.0

release-patch: ## 🏷️ Bump patch version (x.y.Z)
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	$(MAKE) version-bump NEW_VERSION=$$MAJOR.$$MINOR.$$NEW_PATCH

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

test-backend: ## 🏗️ Run backend tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/handler/... ./tests/unit/backend/... || echo "$(COLOR_YELLOW)⚠️ Some backend tests failed$(COLOR_RESET)"

test-devops: ## ⚙️ Run DevOps tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/devops/... || echo "$(COLOR_YELLOW)⚠️ Some DevOps tests failed$(COLOR_RESET)"

test-e2e: ## 🧪 Run E2E tests (requires cluster)
	@chmod +x $(SRC_DIR)/scripts/run-qa-e2e-tests.sh 2>/dev/null || true
	@$(SRC_DIR)/scripts/run-qa-e2e-tests.sh

test-integration: ## 🔗 Run integration tests (requires cluster)
	@cd $(SRC_DIR) && go test -tags=integration -v ./tests/integration/...

test-load: ## 📊 Run load tests
	@chmod +x $(SRC_DIR)/scripts/run-qa-load-tests.sh 2>/dev/null || true
	@$(SRC_DIR)/scripts/run-qa-load-tests.sh

test-security: ## 🔒 Run security tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/security/... || echo "$(COLOR_YELLOW)⚠️ Some security tests failed$(COLOR_RESET)"

test-sre: ## 🚨 Run SRE tests
	@cd $(SRC_DIR) && go test -v -race ./tests/unit/sre/... || echo "$(COLOR_YELLOW)⚠️ Some SRE tests failed$(COLOR_RESET)"

test-unit: ## 🧪 Run all unit tests with coverage
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running unit tests...$(COLOR_RESET)"
	@cd $(SRC_DIR) && go test -v -race -coverprofile=coverage.out ./tests/unit/... 2>&1 | tee test-output.log || true
	@if [ -f $(SRC_DIR)/coverage.out ]; then cd $(SRC_DIR) && go tool cover -func=coverage.out; fi
	@rm -f $(SRC_DIR)/test-output.log

trigger-build: ## 🚀 Trigger lambda build event (use ENV=pro|studio)
	@cd $(SRC_DIR)/tests && ENV=$(ENV) uv run --python 3.9 python create-event-builder.py

trigger-delete: ## 🗑️ Trigger lambda delete event (use ENV=pro|studio)
	@cd $(SRC_DIR)/tests && ENV=$(ENV) uv run --python 3.9 python create-event-delete.py

trigger-lambda: ## 🚀 Trigger lambda parser event (use ENV=pro|studio)
	@cd $(SRC_DIR)/tests && ENV=$(ENV) uv run --python 3.9 python create-event-lambda.py

upload-parser: ## 📤 Upload parser to S3 (use PARSER_ID=xxx ENV=pro|studio)
	@if [ -z "$(PARSER_ID)" ] || [ ! -f "$(SRC_DIR)/tests/$(PARSER_ID)" ]; then \
		echo "$(COLOR_RED)❌ PARSER_ID required or file not found$(COLOR_RESET)"; \
		exit 1; \
	fi
	@aws s3 cp $(SRC_DIR)/tests/$(PARSER_ID) s3://knative-lambda-$(ENV)-fusion-modules-tmp/global/parser/$(PARSER_ID)
	$(call print_success,Parser uploaded to S3)

version: ## 🏷️ Show current version
	@echo "$(COLOR_BOLD)Current version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)VERSION file:$(COLOR_RESET) $(VERSION_FILE)"
	@echo ""
	@echo "$(COLOR_BOLD)Kustomization tags:$(COLOR_RESET)"
	@grep -h "newTag:" $(K8S_DIR)/overlays/*/kustomization.yaml 2>/dev/null | sed 's/^/  /'
	@echo ""
	@echo "$(COLOR_BOLD)OPERATOR_VERSION env vars:$(COLOR_RESET)"
	@grep -h "OPERATOR_VERSION" -A1 $(K8S_DIR)/overlays/*/kustomization.yaml 2>/dev/null | grep "value:" | sed 's/^/  /'

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
	@echo "  2. Build: make operator-build-image-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"
