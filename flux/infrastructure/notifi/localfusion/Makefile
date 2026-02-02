# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🚀 LOCALFUSION - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy LocalFusion mock services
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make build-local      - Build and push Docker image to local registry
#	  make release-patch    - Bump patch version, build, commit, push, and deploy
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.DEFAULT_GOAL := help

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 PATH & PLATFORM CONFIGURATION                                       │
# └─────────────────────────────────────────────────────────────────────────┘

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)/src
K8S_DIR := $(ROOT_DIR)/k8s

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
BUILDX_BUILDER ?= homelab-builder

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ PROJECT CONFIGURATION                                              │
# └─────────────────────────────────────────────────────────────────────────┘

PROJECT_NAME := localfusion
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY := localhost:5001
IMAGE_NAME := localfusion
IMAGE_TAG := v$(VERSION)
FULL_IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
LOCALFUSION_SRC := src
DOCKERFILE := Dockerfile

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES CONFIGURATION                                            │
# └─────────────────────────────────────────────────────────────────────────┘

ENV ?= studio
NAMESPACE ?= localfusion

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

define print_error
	@echo "$(COLOR_BOLD)$(COLOR_RED)❌ $(1)$(COLOR_RESET)"
endef

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 HELP & INFO                                                          │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: help version info
help: ## Show this help message
	@echo "$(COLOR_BOLD)🚀 LocalFusion - Available commands:$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_BLUE)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

version: ## Show current version
	@echo "$(COLOR_BOLD)Current version: $(COLOR_GREEN)v$(VERSION)$(COLOR_RESET)"
	@echo "VERSION file: $(VERSION_FILE)"

info: ## Show build information
	@echo "$(COLOR_BOLD)Build Information:$(COLOR_RESET)"
	@echo "  Version:     v$(VERSION)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Git Branch:  $(GIT_BRANCH)"
	@echo "  Build Date:  $(BUILD_DATE)"
	@echo "  Image:      $(FULL_IMAGE)"
	@echo "  Registry:   $(REGISTRY)"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ VERSION MANAGEMENT                                                 │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: version-bump release release-patch release-minor release-major
version-bump: ## 🏷️ Bump version and update all manifests (NEW_VERSION=x.y.z)
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "$(COLOR_BOLD)$(COLOR_RED)❌ Usage: make version-bump NEW_VERSION=x.y.z$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏷️ Bumping version: $(VERSION) → $(NEW_VERSION)$(COLOR_RESET)"
	@echo ""
	@# Update VERSION file
	@echo "$(NEW_VERSION)" > $(VERSION_FILE)
	@echo "  ✅ Updated VERSION file"
	@# Update deployment.yaml image tag
	@if [ -f "$(K8S_DIR)/deployment.yaml" ]; then \
		sed -i.bak 's|image: localhost:5001/localfusion:[^[:space:]]*|image: localhost:5001/localfusion:v$(NEW_VERSION)|g' "$(K8S_DIR)/deployment.yaml" && rm -f "$(K8S_DIR)/deployment.yaml.bak"; \
		echo "  ✅ Updated k8s/deployment.yaml"; \
	fi
	@# Update kustomization.yaml if it has image tag references
	@if [ -f "$(ROOT_DIR)/kustomization.yaml" ]; then \
		if grep -q "newTag:" "$(ROOT_DIR)/kustomization.yaml"; then \
			sed -i.bak 's|newTag: v[0-9.]*|newTag: v$(NEW_VERSION)|g' "$(ROOT_DIR)/kustomization.yaml" && rm -f "$(ROOT_DIR)/kustomization.yaml.bak"; \
			echo "  ✅ Updated kustomization.yaml"; \
		fi; \
	fi
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Version bumped to $(NEW_VERSION)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Next steps:$(COLOR_RESET)"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): localfusion v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump build-local ## 🚀 Full release: bump version + build + commit + push + deploy (NEW_VERSION=x.y.z)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🚀 Creating release v$(NEW_VERSION)...$(COLOR_RESET)"
	@$(MAKE) commit-release
	@$(MAKE) push-release
	@$(MAKE) deploy
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Released v$(NEW_VERSION) to $(ENV)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Release complete!$(COLOR_RESET)"

release-patch: ## 🏷️ Bump patch version (x.y.Z) and release
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	$(MAKE) release NEW_VERSION=$$MAJOR.$$MINOR.$$NEW_PATCH

release-minor: ## 🏷️ Bump minor version (x.Y.0) and release
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	NEW_MINOR=$$((MINOR + 1)); \
	$(MAKE) release NEW_VERSION=$$MAJOR.$$NEW_MINOR.0

release-major: ## 🏷️ Bump major version (X.0.0) and release
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	NEW_MAJOR=$$((MAJOR + 1)); \
	$(MAKE) release NEW_VERSION=$$NEW_MAJOR.0.0

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER IMAGE BUILD TARGETS                                          │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build build-local buildx ensure-buildx-builder
ensure-buildx-builder: ## 🔧 Ensure buildx builder exists
	@if ! docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1; then \
		$(call print_status,🔧 Creating builder with host network access...); \
		docker buildx create --name "$(BUILDX_BUILDER)" --driver docker-container --driver-opt network=host --bootstrap || true; \
	fi
	@docker buildx use "$(BUILDX_BUILDER)" >/dev/null 2>&1 || true

build-local: ensure-buildx-builder ## Build Docker image locally (single platform for faster local testing)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🐳 Building LocalFusion Docker image locally...$(COLOR_RESET)"
	@if [ ! -d "$(LOCALFUSION_SRC)" ]; then \
		echo "$(COLOR_BOLD)$(COLOR_RED)❌ LocalFusion source not found at $(LOCALFUSION_SRC)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@if [ ! -f "$(DOCKERFILE)" ]; then \
		echo "$(COLOR_BOLD)$(COLOR_RED)❌ Dockerfile not found at $(DOCKERFILE)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@CURRENT_VERSION=$$(cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0"); \
	CURRENT_IMAGE_TAG="v$$CURRENT_VERSION"; \
	CURRENT_FULL_IMAGE="$(REGISTRY)/$(IMAGE_NAME):$$CURRENT_IMAGE_TAG"; \
	ARCH=$$(uname -m); \
	case "$$ARCH" in \
		x86_64) PLATFORM="linux/amd64" ;; \
		arm64|aarch64) PLATFORM="linux/arm64" ;; \
		*) echo "$(COLOR_BOLD)$(COLOR_RED)❌ Unsupported architecture: $$ARCH$(COLOR_RESET)"; exit 1 ;; \
	esac; \
	echo "📦 Image: $$CURRENT_FULL_IMAGE"; \
	echo "📦 Platform: $$PLATFORM"; \
	echo "📦 Source: $(LOCALFUSION_SRC)"; \
	docker buildx build \
		--builder "$(BUILDX_BUILDER)" \
		--platform "$$PLATFORM" \
		-t $$CURRENT_FULL_IMAGE \
		-f "$(DOCKERFILE)" \
		--cache-from type=registry,ref=$(REGISTRY)/$(IMAGE_NAME):buildcache \
		--cache-to type=registry,ref=$(REGISTRY)/$(IMAGE_NAME):buildcache,mode=max \
		--push \
		"$(LOCALFUSION_SRC)" || exit 1; \
	echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Build complete! Image: $$CURRENT_FULL_IMAGE$(COLOR_RESET)"

build: build-local ## Build Docker image (alias for build-local)

buildx: ensure-buildx-builder ## Build multi-platform Docker image (amd64, arm64) using buildx
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🐳 Building multi-platform LocalFusion Docker image...$(COLOR_RESET)"
	@if [ ! -d "$(LOCALFUSION_SRC)" ]; then \
		echo "$(COLOR_BOLD)$(COLOR_RED)❌ LocalFusion source not found at $(LOCALFUSION_SRC)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@if [ ! -f "$(DOCKERFILE)" ]; then \
		echo "$(COLOR_BOLD)$(COLOR_RED)❌ Dockerfile not found at $(DOCKERFILE)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@CURRENT_VERSION=$$(cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0"); \
	CURRENT_IMAGE_TAG="v$$CURRENT_VERSION"; \
	CURRENT_FULL_IMAGE="$(REGISTRY)/$(IMAGE_NAME):$$CURRENT_IMAGE_TAG"; \
	echo "📦 Platforms: linux/amd64,linux/arm64"; \
	echo "📦 Image: $$CURRENT_FULL_IMAGE"; \
	docker buildx build \
		--builder "$(BUILDX_BUILDER)" \
		--platform linux/amd64,linux/arm64 \
		-t $$CURRENT_FULL_IMAGE \
		-f "$(DOCKERFILE)" \
		--cache-from type=registry,ref=$(REGISTRY)/$(IMAGE_NAME):buildcache \
		--cache-to type=registry,ref=$(REGISTRY)/$(IMAGE_NAME):buildcache,mode=max \
		--push \
		"$(LOCALFUSION_SRC)" || exit 1; \
	echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Multi-platform build complete!$(COLOR_RESET)"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES DEPLOYMENT                                               │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: deploy apply restart logs
deploy: ## 🚀 Deploy to cluster using kustomize
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🚀 Deploying LocalFusion to cluster...$(COLOR_RESET)"
	@kubectl apply -k $(ROOT_DIR)
	@kubectl rollout status deployment/localfusion -n $(NAMESPACE) --timeout=90s || true
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Deployed to $(NAMESPACE) namespace$(COLOR_RESET)"

apply: deploy ## Apply kustomization to cluster (alias for deploy)

restart: ## 🔄 Restart deployment
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🔄 Restarting LocalFusion deployment...$(COLOR_RESET)"
	@kubectl rollout restart deployment/localfusion -n $(NAMESPACE)
	@kubectl rollout status deployment/localfusion -n $(NAMESPACE) --timeout=60s || true
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Deployment restarted$(COLOR_RESET)"

logs: ## 📋 Show logs
	@kubectl logs -f deployment/localfusion -n $(NAMESPACE) || kubectl logs deployment/localfusion -n $(NAMESPACE)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📝 GIT OPERATIONS                                                      │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: commit-release push-release
commit-release: ## 📝 Commit release changes
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)📝 Committing release changes...$(COLOR_RESET)"
	@RELEASE_VERSION=$$(cat $(VERSION_FILE) 2>/dev/null || echo "unknown"); \
	git add -A; \
	git commit -m "chore(release): localfusion v$$RELEASE_VERSION" || echo "$(COLOR_BOLD)$(COLOR_RED)❌ Failed to commit. Changes may already be committed.$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Changes committed$(COLOR_RESET)"

push-release: ## 📤 Push release to remote
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)📤 Pushing release to remote...$(COLOR_RESET)"
	@git push origin $(GIT_BRANCH) || echo "$(COLOR_BOLD)$(COLOR_RED)❌ Failed to push. Check your git remote and permissions.$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Pushed to origin/$(GIT_BRANCH)$(COLOR_RESET)"
