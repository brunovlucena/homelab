# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏗️ AGENTS WHATSAPP RUST - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the messaging platform services
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make build-images     - Build and push all Docker images
#	  make deploy           - Deploy to Kubernetes
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.DEFAULT_GOAL := help

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 PATH & PLATFORM CONFIGURATION                                       │
# └─────────────────────────────────────────────────────────────────────────┘

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)

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

PROJECT_NAME := agents-whatsapp-rust
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY := localhost:5001
MESSAGING_IMAGE := $(REGISTRY)/agents-whatsapp-rust-messaging
USER_IMAGE := $(REGISTRY)/agents-whatsapp-rust-user
AGENT_GATEWAY_IMAGE := $(REGISTRY)/agents-whatsapp-rust-agent-gateway
MESSAGE_PROCESSOR_IMAGE := $(REGISTRY)/agents-whatsapp-rust-message-processor
MESSAGE_STORAGE_IMAGE := $(REGISTRY)/agents-whatsapp-rust-message-storage

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/agents-whatsapp-rust
ifeq ($(GIT_BRANCH),main)
	DOCKER_TAG := v$(VERSION)
else ifeq ($(GIT_BRANCH),develop)
	DOCKER_TAG := v$(VERSION)-beta.$(shell date +%s)
else
	DOCKER_TAG := v$(VERSION)-dev.$(GIT_COMMIT)
endif

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES CONFIGURATION                                            │
# └─────────────────────────────────────────────────────────────────────────┘

NAMESPACE ?= homelab-services
K8S_DIR := $(ROOT_DIR)/k8s

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
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏗️ Agents WhatsApp Rust$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ VERSION & RELEASE TARGETS                                           │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: version version-check version-bump release release-patch release-minor release-major
version: ## 🏷️ Show current version
	@echo "$(COLOR_BOLD)Current version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)VERSION file:$(COLOR_RESET) $(VERSION_FILE)"
	@echo ""
	@echo "$(COLOR_BOLD)Image tags (ghcr.io):$(COLOR_RESET)"
	@echo "  Messaging:      $(DOCKER_REPO)-messaging:v$(VERSION)"
	@echo "  User:           $(DOCKER_REPO)-user:v$(VERSION)"
	@echo "  Agent Gateway:  $(DOCKER_REPO)-agent-gateway:v$(VERSION)"
	@echo "  Message Storage: $(DOCKER_REPO)-message-storage:v$(VERSION)"

version-check: ## 🔍 Verify version is set correctly
	@echo "$(COLOR_BOLD)🔍 Checking version...$(COLOR_RESET)"
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "1.0.0" ]; then \
		echo "$(COLOR_YELLOW)⚠️ Using default version. Set VERSION or update VERSION file$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_GREEN)✅ Version: $(VERSION)$(COLOR_RESET)"; \
	fi

version-bump: ## 🏷️ Bump version and update kustomizations (NEW_VERSION=x.y.z)
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "$(COLOR_RED)❌ Usage: make version-bump NEW_VERSION=x.y.z$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏷️ Bumping version: $(VERSION) → $(NEW_VERSION)$(COLOR_RESET)"
	@echo ""
	@# Update VERSION file
	@echo "$(NEW_VERSION)" > $(VERSION_FILE)
	@echo "  ✅ Updated VERSION file"
	@# Update kustomization image tags in base
	@if [ -f "$(K8S_DIR)/base/kustomization.yaml" ]; then \
		sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$(K8S_DIR)/base/kustomization.yaml" && rm -f "$(K8S_DIR)/base/kustomization.yaml.bak"; \
		sed -i.bak 's|value: "v[0-9.]*[-a-zA-Z0-9.]*"|value: "v$(NEW_VERSION)"|g' "$(K8S_DIR)/base/kustomization.yaml" && rm -f "$(K8S_DIR)/base/kustomization.yaml.bak"; \
		echo "  ✅ Updated base/kustomization.yaml image tags and patches"; \
	fi
	@# Update kustomization image tags in overlays
	@for overlay in $(K8S_DIR)/overlays/*/kustomization.yaml; do \
		if [ -f "$$overlay" ]; then \
			sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay" && rm -f "$$overlay.bak"; \
			sed -i.bak 's|value: "v[0-9.]*[-a-zA-Z0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay" && rm -f "$$overlay.bak"; \
			echo "  ✅ Updated $$overlay image tags and patches"; \
		fi \
	done
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
	@$(MAKE) deploy
	@echo ""
	$(call print_success,Released v$(NEW_VERSION))
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
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	IMAGE_TAG="$(REGISTRY)/$(1):v$$VERSION"; \
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
		$(SRC_DIR)
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
build-images: messaging-build-image user-build-image agent-gateway-build-image message-processor-build-image message-storage-build-image ## 🐳 Build and push all Docker images

build-images-local: messaging-build-image-local user-build-image-local agent-gateway-build-image-local message-processor-build-image-local message-storage-build-image-local ## 🐳 Build and push all images to local registry (arm64)
	@echo "✅ Built and pushed to local registry:"
	@echo "   - $(MESSAGING_IMAGE):v$(VERSION)"
	@echo "   - $(USER_IMAGE):v$(VERSION)"
	@echo "   - $(AGENT_GATEWAY_IMAGE):v$(VERSION)"
	@echo "   - $(MESSAGE_PROCESSOR_IMAGE):v$(VERSION)"
	@echo "   - $(MESSAGE_STORAGE_IMAGE):v$(VERSION)"

.PHONY: messaging-build-image messaging-build-image-local
messaging-build-image: ensure-buildx-builder ## 🐳 Build and push messaging service image
	$(call build-image,agents-whatsapp-rust-messaging,$(SRC_DIR)/messaging-service,Dockerfile,$(PUSH_PLATFORM))

messaging-build-image-local: ensure-buildx-builder ## 🐳 Build and push messaging service image (arm64)
	$(call build-image,agents-whatsapp-rust-messaging,$(SRC_DIR)/messaging-service,Dockerfile,linux/arm64)

.PHONY: user-build-image user-build-image-local
user-build-image: ensure-buildx-builder ## 🐳 Build and push user service image
	$(call build-image,agents-whatsapp-rust-user,$(SRC_DIR)/user-service,Dockerfile,$(PUSH_PLATFORM))

user-build-image-local: ensure-buildx-builder ## 🐳 Build and push user service image (arm64)
	$(call build-image,agents-whatsapp-rust-user,$(SRC_DIR)/user-service,Dockerfile,linux/arm64)

.PHONY: agent-gateway-build-image agent-gateway-build-image-local
agent-gateway-build-image: ensure-buildx-builder ## 🐳 Build and push agent gateway image
	$(call build-image,agents-whatsapp-rust-agent-gateway,$(SRC_DIR)/agent-gateway,Dockerfile,$(PUSH_PLATFORM))

agent-gateway-build-image-local: ensure-buildx-builder ## 🐳 Build and push agent gateway image (arm64)
	$(call build-image,agents-whatsapp-rust-agent-gateway,$(SRC_DIR)/agent-gateway,Dockerfile,linux/arm64)

.PHONY: message-processor-build-image message-processor-build-image-local
message-processor-build-image: ensure-buildx-builder ## 🐳 Build and push message processor image
	$(call build-image,agents-whatsapp-rust-message-processor,$(SRC_DIR)/message-processor,Dockerfile,$(PUSH_PLATFORM))

message-processor-build-image-local: ensure-buildx-builder ## 🐳 Build and push message processor image (arm64)
	$(call build-image,agents-whatsapp-rust-message-processor,$(SRC_DIR)/message-processor,Dockerfile,linux/arm64)

.PHONY: message-storage-build-image message-storage-build-image-local
message-storage-build-image: ensure-buildx-builder ## 🐳 Build and push message storage service image
	$(call build-image,agents-whatsapp-rust-message-storage,$(SRC_DIR)/message-storage-service,Dockerfile,$(PUSH_PLATFORM))

message-storage-build-image-local: ensure-buildx-builder ## 🐳 Build and push message storage service image (arm64)
	$(call build-image,agents-whatsapp-rust-message-storage,$(SRC_DIR)/message-storage-service,Dockerfile,linux/arm64)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🚀 GHCR.IO PUSH TARGETS                                               │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: push-ghcr push-ghcr-multiarch
push-ghcr: push-ghcr-multiarch ## 🚀 Build and push all images to ghcr.io (multi-arch) - alias

push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push all images to ghcr.io (multi-arch amd64+arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	echo "🚀 Pushing all images to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🏗️ Platforms: linux/amd64,linux/arm64"; \
	$(MAKE) messaging-push-ghcr-multiarch; \
	$(MAKE) user-push-ghcr-multiarch; \
	$(MAKE) agent-gateway-push-ghcr-multiarch; \
	$(MAKE) message-storage-push-ghcr-multiarch; \
	echo ""; \
	$(call print_success,All images pushed to ghcr.io)

.PHONY: messaging-push-ghcr-multiarch
messaging-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push messaging service to ghcr.io (multi-arch)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO)-messaging:$$DOCKER_TAG"; \
	echo "🚀 Pushing messaging service to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/messaging-service/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		-t $(DOCKER_REPO)-messaging:latest \
		$(SRC_DIR)

.PHONY: user-push-ghcr-multiarch
user-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push user service to ghcr.io (multi-arch)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO)-user:$$DOCKER_TAG"; \
	echo "🚀 Pushing user service to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/user-service/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		-t $(DOCKER_REPO)-user:latest \
		$(SRC_DIR)

.PHONY: agent-gateway-push-ghcr-multiarch
agent-gateway-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push agent gateway to ghcr.io (multi-arch)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO)-agent-gateway:$$DOCKER_TAG"; \
	echo "🚀 Pushing agent gateway to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/agent-gateway/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		-t $(DOCKER_REPO)-agent-gateway:latest \
		$(SRC_DIR)

.PHONY: message-storage-push-ghcr-multiarch
message-storage-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push message storage service to ghcr.io (multi-arch)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO)-message-storage:$$DOCKER_TAG"; \
	echo "🚀 Pushing message storage service to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/message-storage-service/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		-t $(DOCKER_REPO)-message-storage:latest \
		$(SRC_DIR)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🚀 GITOPS DEPLOYMENT TARGETS                                           │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: deploy deploy-diff deploy-studio deploy-pro studio pro kustomize-build
deploy: ## 🚀 Deploy to Kubernetes (base)
	$(call print_status,🚀 Deploying to Kubernetes...)
	@kubectl apply -k $(K8S_DIR)/base
	$(call print_success,Deployment triggered)

deploy-studio: ## 🚀 Deploy to studio environment
	$(call print_status,🚀 Deploying to studio environment...)
	@kubectl apply -k $(K8S_DIR)/overlays/studio
	$(call print_success,Deployed to studio)

deploy-pro: ## 🚀 Deploy to pro environment
	$(call print_status,🚀 Deploying to pro environment...)
	@kubectl apply -k $(K8S_DIR)/overlays/pro
	$(call print_success,Deployed to pro)

studio: build-images-local deploy-studio ## 🎬 Build and deploy to studio (shorthand)
	$(call print_success,Studio deployment complete)

pro: build-images-local deploy-pro ## 🎬 Build and deploy to pro (shorthand)
	$(call print_success,Pro deployment complete)

deploy-diff: ## 🔍 Show diff for deployment
	$(call print_status,🔍 Showing deployment diff...)
	@kubectl diff -k $(K8S_DIR)/base || true

kustomize-build: ## 🔨 Build kustomize output
	@kubectl kustomize $(K8S_DIR)/base

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📊 STATUS TARGETS                                                      │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: status status-pods status-services
status: ## 📊 Show full system status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  📊 AGENTS WHATSAPP RUST STATUS$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo ""
	@$(MAKE) --no-print-directory status-pods
	@echo ""
	@$(MAKE) --no-print-directory status-services
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Status check complete$(COLOR_RESET)"

status-pods: ## 🐳 Show pod status
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)┌─────────────────────────────────────────────────────────────────────────┐$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)│  🐳 PODS                                                                 │$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)└─────────────────────────────────────────────────────────────────────────┘$(COLOR_RESET)"
	@echo ""
	@kubectl get pods -n $(NAMESPACE) -l app.kubernetes.io/part-of=agents-whatsapp-rust -o wide 2>/dev/null || echo "$(COLOR_YELLOW)  ⚠️ No pods found$(COLOR_RESET)"

status-services: ## 🔌 Show service status
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)┌─────────────────────────────────────────────────────────────────────────┐$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)│  🔌 SERVICES                                                             │$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)└─────────────────────────────────────────────────────────────────────────┘$(COLOR_RESET)"
	@echo ""
	@kubectl get svc -n $(NAMESPACE) -l app.kubernetes.io/part-of=agents-whatsapp-rust -o wide 2>/dev/null || echo "$(COLOR_YELLOW)  ⚠️ No services found$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)🌐 Knative Services:$(COLOR_RESET)"
	@kubectl get ksvc -n $(NAMESPACE) -l app.kubernetes.io/part-of=agents-whatsapp-rust -o wide 2>/dev/null || echo "$(COLOR_YELLOW)  ⚠️ No Knative services found$(COLOR_RESET)"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏗️ RUST BUILD TARGETS                                                 │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build build-release test
build: lint ## 🏗️ Build all Rust services (with lint check)
	$(call print_status,🏗️ Building all Rust services...)
	@if command -v cargo >/dev/null 2>&1; then \
		. "$$HOME/.cargo/env" 2>/dev/null || true; \
		cargo build --all; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Build complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Cargo not found. Install Rust: https://rustup.rs/$(COLOR_RESET)"; \
		exit 1; \
	fi

build-release: lint ## 🏗️ Build all Rust services in release mode (with lint check)
	$(call print_status,🏗️ Building all Rust services (release mode)...)
	@if command -v cargo >/dev/null 2>&1; then \
		. "$$HOME/.cargo/env" 2>/dev/null || true; \
		cargo build --release --all; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Release build complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Cargo not found. Install Rust: https://rustup.rs/$(COLOR_RESET)"; \
		exit 1; \
	fi

test: ## 🧪 Run all tests
	$(call print_status,🧪 Running tests...)
	@if command -v cargo >/dev/null 2>&1; then \
		. "$$HOME/.cargo/env" 2>/dev/null || true; \
		cargo test --all; \
		$(call print_success,Tests complete); \
	else \
		echo "$(COLOR_YELLOW)⚠️ Cargo not found. Install Rust: https://rustup.rs/$(COLOR_RESET)"; \
		exit 1; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔍 LINT & FORMAT TARGETS                                               │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: lint lint-fmt lint-clippy fmt
lint: lint-fmt lint-clippy ## 🔍 Run all lint checks (format + clippy)

lint-fmt: ## 🔍 Check code formatting
	$(call print_status,🔍 Checking code formatting...)
	@if command -v cargo >/dev/null 2>&1; then \
		. "$$HOME/.cargo/env" 2>/dev/null || true; \
		if cargo fmt --check --all; then \
			echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Formatting check passed$(COLOR_RESET)"; \
		else \
			echo "$(COLOR_RED)❌ Formatting check failed. Run 'make fmt' to fix.$(COLOR_RESET)"; \
			exit 1; \
		fi; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Cargo not found. Install Rust: https://rustup.rs/$(COLOR_RESET)"; \
		exit 1; \
	fi

lint-clippy: ## 🔍 Run clippy linter
	$(call print_status,🔍 Running clippy linter...)
	@if command -v cargo >/dev/null 2>&1; then \
		. "$$HOME/.cargo/env" 2>/dev/null || true; \
		if cargo clippy --all-targets --all-features -- -D warnings; then \
			echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Clippy check passed$(COLOR_RESET)"; \
		else \
			echo "$(COLOR_RED)❌ Clippy check failed$(COLOR_RESET)"; \
			exit 1; \
		fi; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Cargo not found. Install Rust: https://rustup.rs/$(COLOR_RESET)"; \
		exit 1; \
	fi

fmt: ## 🎨 Format code with rustfmt
	$(call print_status,🎨 Formatting code...)
	@if command -v cargo >/dev/null 2>&1; then \
		. "$$HOME/.cargo/env" 2>/dev/null || true; \
		cargo fmt --all; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Code formatted$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Cargo not found. Install Rust: https://rustup.rs/$(COLOR_RESET)"; \
		exit 1; \
	fi
