# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#   🌐 CLOUDFLARE TUNNEL OPERATOR - MAKEFILE
#
#   🎯 Purpose: Build, test, and deploy the Cloudflare Tunnel Operator
#
#   🔧 USAGE:
#     make                  - Show help
#     make build            - Build Go binary
#     make build-image      - Build and push Docker image
#     make deploy           - Deploy to Kubernetes
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

PROJECT_NAME := cloudflare-tunnel-operator
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY ?= localhost:5001
IMAGE_NAME := $(REGISTRY)/cloudflare-tunnel-operator
DOCKER_TAG := v$(VERSION)

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/cloudflare-tunnel-operator

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES CONFIGURATION                                            │
# └─────────────────────────────────────────────────────────────────────────┘

NAMESPACE ?= cloudflare-tunnel
K8S_DIR := $(ROOT_DIR)/k8s

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎨 COLORS                                                              │
# └─────────────────────────────────────────────────────────────────────────┘

COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m

define print_status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)$(1)$(COLOR_RESET)"
endef

define print_success
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ $(1)$(COLOR_RESET)"
endef

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 HELP                                                                │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: help
help: ## 📋 Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🌐 Cloudflare Tunnel Operator$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔧 BUILD TARGETS                                                       │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build
build: ## 🔨 Build Go binary locally
	$(call print_status,🔨 Building cloudflare-tunnel-operator...)
	@cd $(SRC_DIR) && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o ../bin/manager ./cmd/main.go
	$(call print_success,Binary built: bin/manager)

.PHONY: test
test: ## 🧪 Run tests
	$(call print_status,🧪 Running tests...)
	@cd $(SRC_DIR) && go test -v -race ./...
	$(call print_success,Tests passed)

.PHONY: lint
lint: ## 🔍 Run linting
	$(call print_status,🔍 Running linting...)
	@cd $(SRC_DIR) && go vet ./...
	@cd $(SRC_DIR) && test -z "$$(gofmt -l . | tee /dev/stderr)"
	$(call print_success,Linting passed)

.PHONY: fmt
fmt: ## 🎨 Format code
	$(call print_status,🎨 Formatting code...)
	@cd $(SRC_DIR) && go fmt ./...
	$(call print_success,Code formatted)

.PHONY: tidy
tidy: ## 📦 Tidy go modules
	$(call print_status,📦 Tidying modules...)
	@cd $(SRC_DIR) && go mod tidy
	$(call print_success,Modules tidied)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER IMAGE BUILD TARGETS                                          │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: ensure-buildx-builder
ensure-buildx-builder: ## 🔧 Ensure buildx builder exists
	@if ! docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1; then \
		echo "🔧 Creating buildx builder '$(BUILDX_BUILDER)' with host network access..."; \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --driver-opt network=host --use || true; \
		docker buildx inspect $(BUILDX_BUILDER) --bootstrap || true; \
	else \
		echo "✅ Buildx builder '$(BUILDX_BUILDER)' already exists"; \
	fi

.PHONY: build-image
build-image: ensure-buildx-builder ## 🐳 Build and push Docker image (multi-arch)
	$(call print_status,🐳 Building cloudflare-tunnel-operator image v$(VERSION)...)
	@docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/Dockerfile \
		--platform $(PUSH_PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--push \
		-t $(IMAGE_NAME):$(DOCKER_TAG) \
		-t $(IMAGE_NAME):latest \
		$(SRC_DIR)
	$(call print_success,Image pushed: $(IMAGE_NAME):$(DOCKER_TAG))

.PHONY: build-image-local
build-image-local: ensure-buildx-builder ## 🐳 Build and push to local registry (arm64)
	$(call print_status,🐳 Building for local registry $(REGISTRY)...)
	@docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/Dockerfile \
		--platform linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--push \
		-t $(IMAGE_NAME):$(DOCKER_TAG) \
		$(SRC_DIR)
	$(call print_success,Image pushed: $(IMAGE_NAME):$(DOCKER_TAG))

.PHONY: prewarm
prewarm: build-image-local ## 🔥 Prewarm: Build and push to localhost:5001
	$(call print_success,Image prewarmed at $(IMAGE_NAME):$(DOCKER_TAG))

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🚀 DEPLOYMENT TARGETS                                                  │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: deploy
deploy: ## 🚀 Deploy operator to Kubernetes
	$(call print_status,🚀 Deploying cloudflare-tunnel-operator...)
	@kubectl apply -k $(K8S_DIR)/base
	$(call print_success,Operator deployed)

.PHONY: undeploy
undeploy: ## 🗑️ Remove operator from Kubernetes
	$(call print_status,🗑️ Removing cloudflare-tunnel-operator...)
	@kubectl delete -k $(K8S_DIR)/base --ignore-not-found
	$(call print_success,Operator removed)

.PHONY: restart
restart: ## 🔄 Restart operator deployment
	$(call print_status,🔄 Restarting operator...)
	@kubectl rollout restart deployment/cloudflare-tunnel-operator -n $(NAMESPACE)
	$(call print_success,Operator restarted)

.PHONY: logs
logs: ## 📜 Tail operator logs
	@kubectl logs -f deployment/cloudflare-tunnel-operator -n $(NAMESPACE)

.PHONY: status
status: ## 📊 Show operator and ingress status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  🌐 CLOUDFLARE TUNNEL OPERATOR STATUS$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)📦 Deployment:$(COLOR_RESET)"
	@kubectl get deployment cloudflare-tunnel-operator -n $(NAMESPACE) -o wide 2>/dev/null || echo "  Not found"
	@echo ""
	@echo "$(COLOR_BOLD)🐳 Pods:$(COLOR_RESET)"
	@kubectl get pods -n $(NAMESPACE) -l app=cloudflare-tunnel-operator -o wide 2>/dev/null || echo "  No pods"
	@echo ""
	@echo "$(COLOR_BOLD)🌐 CloudflareTunnelIngresses:$(COLOR_RESET)"
	@kubectl get cloudflaretunnelingresses -A 2>/dev/null || echo "  None found"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔧 DEVELOPMENT TARGETS                                                 │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: run-local
run-local: ## 🏃 Run operator locally (requires kubeconfig)
	$(call print_status,🏃 Running operator locally...)
	@cd $(SRC_DIR) && go run ./cmd/main.go

.PHONY: generate
generate: ## 🔄 Generate CRD manifests
	$(call print_status,🔄 Generating manifests...)
	@cd $(SRC_DIR) && go generate ./...
	$(call print_success,Manifests generated)

.PHONY: clean
clean: ## 🧹 Clean build artifacts
	$(call print_status,🧹 Cleaning...)
	@rm -rf $(ROOT_DIR)/bin
	@cd $(SRC_DIR) && go clean -cache
	$(call print_success,Cleaned)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📦 COMBINED TARGETS                                                    │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: all
all: tidy lint build build-image-local deploy ## 🎯 Build, push, and deploy (full cycle)
	$(call print_success,Full deployment cycle complete!)

.PHONY: release
release: tidy lint test build-image ## 🚀 Full release (test + multi-arch build)
	$(call print_success,Release build complete: $(IMAGE_NAME):$(DOCKER_TAG))
