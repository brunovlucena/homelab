# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏗️ KNATIVE LAMBDA C2 - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the Command & Control Center
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

PROJECT_NAME := knative-lambda-c2
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY := localhost:5001
BACKEND_IMAGE := $(REGISTRY)/knative-lambda-c2-backend
FRONTEND_IMAGE := $(REGISTRY)/knative-lambda-c2-frontend

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/knative-lambda-c2
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

NAMESPACE ?= knative-lambda-c2
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
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏗️ Knative Lambda C2$(COLOR_RESET)"
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
	@echo "$(COLOR_BOLD)Image tags:$(COLOR_RESET)"
	@echo "  Backend:  $(BACKEND_IMAGE):v$(VERSION)"
	@echo "  Frontend: $(FRONTEND_IMAGE):v$(VERSION)"

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
	@# Update kustomization image tags
	@if [ -f "$(K8S_DIR)/kustomization.yaml" ]; then \
		sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$(K8S_DIR)/kustomization.yaml" && rm -f "$(K8S_DIR)/kustomization.yaml.bak"; \
		echo "  ✅ Updated kustomization.yaml image tags"; \
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
build-images: backend-build-image frontend-build-image ## 🐳 Build and push all Docker images

build-images-local: backend-build-image-local frontend-build-image-local ## 🐳 Build and push all images to local registry (arm64)
	@echo "✅ Built and pushed to local registry:"
	@echo "   - $(BACKEND_IMAGE):v$(VERSION)"
	@echo "   - $(FRONTEND_IMAGE):v$(VERSION)"

.PHONY: backend-build-image backend-build-image-local backend-build-image-nocache
backend-build-image: ensure-buildx-builder ## 🐳 Build and push backend image
	$(call build-image,knative-lambda-c2-backend,$(SRC_DIR)/backend,Dockerfile,$(PUSH_PLATFORM))

backend-build-image-local: ensure-buildx-builder ## 🐳 Build and push backend image (arm64)
	$(call build-image,knative-lambda-c2-backend,$(SRC_DIR)/backend,Dockerfile,linux/arm64)

backend-build-image-nocache: ensure-buildx-builder ## 🐳 Build and push backend image (no cache, arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(REGISTRY)/knative-lambda-c2-backend:$$DOCKER_TAG"; \
	echo "📋 Building knative-lambda-c2-backend version: $$VERSION (NO CACHE)"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/backend/Dockerfile \
		--platform linux/arm64 \
		--no-cache \
		--push \
		-t $$IMAGE_TAG \
		$(SRC_DIR)/backend

.PHONY: frontend-build-image frontend-build-image-local frontend-build-image-nocache
frontend-build-image: ensure-buildx-builder ## 🐳 Build and push frontend image
	$(call build-image,knative-lambda-c2-frontend,$(SRC_DIR)/frontend,Dockerfile,$(PUSH_PLATFORM))

frontend-build-image-local: ensure-buildx-builder ## 🐳 Build and push frontend image (arm64)
	$(call build-image,knative-lambda-c2-frontend,$(SRC_DIR)/frontend,Dockerfile,linux/arm64)

frontend-build-image-nocache: ensure-buildx-builder ## 🐳 Build and push frontend image (no cache, arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(REGISTRY)/knative-lambda-c2-frontend:$$DOCKER_TAG"; \
	echo "📋 Building knative-lambda-c2-frontend version: $$VERSION (NO CACHE)"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/frontend/Dockerfile \
		--platform linux/arm64 \
		--no-cache \
		--push \
		-t $$IMAGE_TAG \
		$(SRC_DIR)/frontend

backend-push-ghcr: backend-push-ghcr-multiarch ## 🚀 Build and push backend to ghcr.io (multi-arch) - alias

backend-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push backend to ghcr.io (multi-arch amd64+arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO)-backend:$$DOCKER_TAG"; \
	echo "🚀 Pushing backend to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️ Platforms: linux/amd64,linux/arm64"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/backend/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		$(SRC_DIR)/backend

frontend-push-ghcr: frontend-push-ghcr-multiarch ## 🚀 Build and push frontend to ghcr.io (multi-arch) - alias

frontend-push-ghcr-multiarch: ensure-buildx-builder ## 🚀 Build and push frontend to ghcr.io (multi-arch amd64+arm64)
	@VERSION=$$(cat "$(VERSION_FILE)" 2>/dev/null || echo "1.0.0"); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(DOCKER_REPO)-frontend:$$DOCKER_TAG"; \
	echo "🚀 Pushing frontend to ghcr.io (multi-arch)"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️ Platforms: linux/amd64,linux/arm64"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(SRC_DIR)/frontend/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $$IMAGE_TAG \
		$(SRC_DIR)/frontend

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔨 BUILD TARGETS                                                       │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build-frontend clean
build-frontend: ## 🔨 Build frontend (production)
	$(call print_status,🔨 Building frontend...)
	@cd $(SRC_DIR)/frontend && npm ci
	@cd $(SRC_DIR)/frontend && npm run build
	$(call print_success,Frontend built in $(SRC_DIR)/frontend/dist/)

clean: ## 🧹 Clean all build artifacts
	$(call print_status,🧹 Cleaning build artifacts...)
	@rm -rf $(SRC_DIR)/frontend/dist $(SRC_DIR)/frontend/node_modules
	@rm -f $(SRC_DIR)/frontend/package-lock.json
	$(call print_success,Build artifacts cleaned)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧪 TESTING TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: test test-backend test-frontend
test: test-backend test-frontend ## 🧪 Run all tests

test-backend: ## 🧪 Test backend API
	$(call print_status,🧪 Testing backend API...)
	@cd $(SRC_DIR)/backend && python3 test_api.py || echo "$(COLOR_YELLOW)⚠️ Backend server must be running$(COLOR_RESET)"

test-frontend: build-frontend ## 🧪 Test frontend build
	$(call print_status,🧪 Testing frontend build...)
	@if [ -d "$(SRC_DIR)/frontend/dist" ]; then \
		echo "$(COLOR_GREEN)✅ Frontend build successful$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_RED)❌ Frontend build failed$(COLOR_RESET)"; \
		exit 1; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🚀 GITOPS DEPLOYMENT TARGETS                                           │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: deploy deploy-diff
deploy: ## 🚀 Deploy to Kubernetes
	$(call print_status,🚀 Deploying to Kubernetes...)
	@kubectl apply -k $(K8S_DIR)
	$(call print_success,Deployment triggered)

deploy-diff: ## 🔍 Show diff for deployment
	$(call print_status,🔍 Showing deployment diff...)
	@kubectl diff -k $(K8S_DIR) || true

.PHONY: kustomize-build
kustomize-build: ## 🔨 Build kustomize output
	@kubectl kustomize $(K8S_DIR)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES PORT-FORWARD TARGETS                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: pf-backend pf-frontend pf
pf-backend: ## 🔌 Port forward backend to 0.0.0.0:8080
	@kubectl port-forward --address 0.0.0.0 svc/c2-backend 8080:8080 -n $(NAMESPACE)

pf-frontend: ## 🔌 Port forward frontend to 0.0.0.0:3000
	@kubectl port-forward --address 0.0.0.0 deployment/c2-frontend 3000:80 -n $(NAMESPACE)

pf: pf-frontend ## 🔌 Port forward frontend (alias)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📊 STATUS TARGETS                                                      │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: status status-pods status-services status-logs
status: ## 📊 Show full C2 system status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)  📊 KNATIVE LAMBDA C2 STATUS$(COLOR_RESET)"
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
	@kubectl get pods -n $(NAMESPACE) -o wide 2>/dev/null || echo "$(COLOR_YELLOW)  ⚠️ Namespace $(NAMESPACE) not found$(COLOR_RESET)"

status-services: ## 🔌 Show service status
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)┌─────────────────────────────────────────────────────────────────────────┐$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)│  🔌 SERVICES                                                             │$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)└─────────────────────────────────────────────────────────────────────────┘$(COLOR_RESET)"
	@echo ""
	@kubectl get svc -n $(NAMESPACE) -o wide 2>/dev/null || echo "$(COLOR_YELLOW)  ⚠️ No services found$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)🌐 Ingress:$(COLOR_RESET)"
	@kubectl get ingress -n $(NAMESPACE) -o wide 2>/dev/null || echo "$(COLOR_YELLOW)  ⚠️ No ingress found$(COLOR_RESET)"

status-logs: ## 📜 Tail backend logs
	@kubectl logs -f deployment/c2-backend -n $(NAMESPACE) --tail=100

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🛠️ DEVELOPMENT TARGETS                                                 │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: setup dev-backend dev-frontend
setup: ## 🔧 Setup development environment
	$(call print_status,🔧 Setting up development environment...)
	@cd $(SRC_DIR)/backend && pip install -r requirements.txt
	@cd $(SRC_DIR)/frontend && npm install
	$(call print_success,Development environment ready)

dev-backend: ## 🚀 Run backend in development mode
	$(call print_status,🚀 Starting backend dev server...)
	@cd $(SRC_DIR)/backend && uvicorn main:app --reload --host 0.0.0.0 --port 8080

dev-frontend: ## 🚀 Run frontend in development mode
	$(call print_status,🚀 Starting frontend dev server...)
	@cd $(SRC_DIR)/frontend && npm run dev
