UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
	ifeq ($(UNAME_M),arm64)
		DEFAULT_LOAD_PLATFORM := linux/arm64
	endif
endif

DEFAULT_LOAD_PLATFORM ?= linux/amd64
DEFAULT_PUSH_PLATFORM ?= $(DEFAULT_LOAD_PLATFORM)
LOAD_PLATFORM ?= $(DEFAULT_LOAD_PLATFORM)
PUSH_PLATFORM ?= $(DEFAULT_PUSH_PLATFORM)
BUILDX_BUILDER ?= homelab

ifdef LOAD_PLATFORM
	LOAD_PLATFORM_FLAG := --platform $(LOAD_PLATFORM)
endif

ifdef PUSH_PLATFORM
	PUSH_PLATFORM_FLAG := --platform $(PUSH_PLATFORM)
endif
ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
REGISTRY ?= localhost:5001
API_IMAGE := $(REGISTRY)/homepage-api
FRONTEND_IMAGE := $(REGISTRY)/homepage-frontend

API_CONTEXT := $(ROOT_DIR)/src/api
FRONTEND_CONTEXT := $(ROOT_DIR)/src/frontend

K8S_NAMESPACE ?= homepage
API_DEPLOYMENT := homepage-api
FRONTEND_DEPLOYMENT := homepage-frontend

API_PORT ?= 8080
FRONTEND_PORT ?= 8080

.PHONY: \
	api-build-image \
	api-build-image-local \
	api-build-image-ghcr \
	api-lint \
	api-test \
	api-run \
	api-run-dev \
	api-image \
	frontend-build-image \
	frontend-build-image-local \
	frontend-build-image-ghcr \
	frontend-lint \
	frontend-test \
	frontend-run \
	frontend-run-dev \
	frontend-image \
	ensure-buildx-builder \
	build-images \
	build-images-local \
	build-images-ghcr \
	lint \
	test \
	workflow-local \
	rollout \
	verify \
	help

# GHCR registry for pulling images built by GitHub Actions
GHCR_REGISTRY ?= ghcr.io/brunovlucena
LOCAL_REGISTRY ?= localhost:5001

help: ## 📖 Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

rollout: ## 🔁 Rollout restart both API and Frontend deployments
	@echo "🔁 Rolling out homepage deployments..."
	kubectl rollout restart deployment/$(API_DEPLOYMENT) -n $(K8S_NAMESPACE)
	kubectl rollout restart deployment/$(FRONTEND_DEPLOYMENT) -n $(K8S_NAMESPACE)
	@echo "⏳ Waiting for rollouts to complete..."
	kubectl rollout status deployment/$(API_DEPLOYMENT) -n $(K8S_NAMESPACE) --timeout=300s
	kubectl rollout status deployment/$(FRONTEND_DEPLOYMENT) -n $(K8S_NAMESPACE) --timeout=300s
	@echo "✅ Rollout complete!"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔍 LINT & TEST (run before build/push)                                  │
# └─────────────────────────────────────────────────────────────────────────┘

api-lint: ## 🔍 Lint API (go vet, golangci-lint if available)
	@echo "🔍 Linting API..."
	@cd $(API_CONTEXT) && go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(API_CONTEXT) && golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not installed, skipped (go vet only)"; \
	fi
	@echo "✅ API lint passed"

api-test: ## 🧪 Test API
	@echo "🧪 Testing API..."
	@cd $(API_CONTEXT) && go test -v ./...
	@echo "✅ API tests passed"

frontend-lint: ## 🔍 Lint frontend (ESLint + tsc --noEmit)
	@echo "🔍 Linting frontend..."
	@cd $(FRONTEND_CONTEXT) && \
		if [ ! -d node_modules ]; then npm ci --legacy-peer-deps --no-audit --no-fund; fi && \
		npm run lint
	@echo "✅ Frontend lint passed"

frontend-test: ## 🧪 Test frontend (Vitest with coverage)
	@echo "🧪 Testing frontend..."
	@cd $(FRONTEND_CONTEXT) && \
		if [ ! -d node_modules ]; then npm ci --legacy-peer-deps --no-audit --no-fund; fi && \
		CI=1 npm run test:coverage
	@echo "✅ Frontend tests passed"

lint: api-lint frontend-lint ## 🔍 Lint API and frontend
	@echo "✅ All lint passed"

test: api-test frontend-test ## 🧪 Test API and frontend
	@echo "✅ All tests passed"

# Simulate homepage CI workflow locally with act (https://github.com/nektos/act). Requires: act, Docker.
REPO_ROOT := $(shell cd $(ROOT_DIR) && git rev-parse --show-toplevel 2>/dev/null || echo "")
WORKFLOW_FILE := .github/workflows/infra-homepage.yml
ACT_PLATFORM ?= self-hosted=catthehacker/ubuntu:act-latest

workflow-local: ## 🔄 Run homepage GitHub workflow locally (act). Requires: act, Docker.
	@if ! command -v act >/dev/null 2>&1; then \
		echo "❌ act not found. Install: https://github.com/nektos/act (e.g. brew install act)"; \
		exit 1; \
	fi
	@if [ -z "$(REPO_ROOT)" ] || [ ! -f "$(REPO_ROOT)/$(WORKFLOW_FILE)" ]; then \
		echo "❌ Repo root or $(WORKFLOW_FILE) not found (REPO_ROOT=$(REPO_ROOT))"; \
		exit 1; \
	fi
	@echo "🔄 Running homepage workflow locally (act)..."
	@cd "$(REPO_ROOT)" && act workflow_dispatch -W $(WORKFLOW_FILE) -P $(ACT_PLATFORM) --verbose
	@echo "✅ Workflow simulation passed"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ VERSION & RELEASE TARGETS                                           │
# │                                                                          │
# │  DRY Principle: VERSION file is the single source of truth              │
# │  These targets update VERSION and all dependent files                    │
# └─────────────────────────────────────────────────────────────────────────┘

K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

.PHONY: version version-bump release release-patch release-minor release-major deploy-current
version: ## 🏷️ Show current version
	@VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	VERSION=$$(cat "$$VERSION_FILE" 2>/dev/null || echo "0.0.0"); \
	echo "Current version: $$VERSION"; \
	echo "VERSION file: $$VERSION_FILE"; \
	echo ""; \
	echo "Kustomization tags:"; \
	grep -rh "newTag:" $(K8S_KUSTOMIZE_DIR)/*/kustomization.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no image tags defined)"

version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "❌ Usage: make version-bump NEW_VERSION=x.y.z"; \
		exit 1; \
	fi
	@VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	OLD_VERSION=$$(cat "$$VERSION_FILE" 2>/dev/null || echo "0.0.0"); \
	echo "🏷️ Bumping version: $$OLD_VERSION → $(NEW_VERSION)"; \
	echo ""; \
	echo "$(NEW_VERSION)" > "$$VERSION_FILE"; \
	echo "  ✅ Updated VERSION file"; \
	for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "newTag:" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml"; \
			fi; \
		fi; \
	done; \
	echo ""; \
	echo "✅ Version bumped to $(NEW_VERSION)"; \
	echo ""; \
	echo "Next steps:"; \
	echo "  1. Review changes: git diff"; \
	echo "  2. Build: make build-images-ghcr"; \
	echo "  3. Test locally"; \
	echo "  4. Commit: git add -A && git commit -m 'chore(release): homepage v$(NEW_VERSION)'"; \
	echo "  5. Push: git push origin main"

# Build and deploy current VERSION only (no bump, no commit/push). Used when user selects 'n' on release-patch/minor/major.
deploy-current: build-images-ghcr ## 🚀 Build, push GHCR, rollout using current VERSION (no bump, no commit/push)
	@VERSION_FILE="$(ROOT_DIR)/VERSION"; VERSION=$$(cat "$$VERSION_FILE" 2>/dev/null || echo "?"); \
	echo "🚀 Deploying current v$$VERSION (no version bump)..."; \
	$(MAKE) rollout
	@echo "✅ Deploy complete (no commit/push)"

# Single release flow: version-bump → build-images-ghcr → rollout → commit & push. Use release-patch/minor/major or pass NEW_VERSION=x.y.z
release: version-bump build-images-ghcr ## 🚀 Bump, build, push GHCR, deploy, commit & push (NEW_VERSION=x.y.z or release-patch/minor/major)
	@echo "🚀 Deploying v$(NEW_VERSION)..."
	@$(MAKE) rollout
	@echo ""
	@echo "📝 Committing and pushing v$(NEW_VERSION)..."
	@git add -A && git commit -m "chore(release): homepage v$(NEW_VERSION)" && git push
	@echo ""
	@echo "✅ Released v$(NEW_VERSION)"

release-patch: lint test ## 🚀 Lint, test, confirm; [y] bump patch + release, [n] build & deploy current version only
	@VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	CURRENT=$$(cat "$$VERSION_FILE" 2>/dev/null || echo "0.0.0"); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	NEW_VER=$$MAJOR.$$MINOR.$$((PATCH + 1)); \
	echo ""; echo "  Current: $$CURRENT  →  Bump to (patch): $$NEW_VER"; echo -n "  Bump version and release? [y/n] "; \
	read -r ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		$(MAKE) release NEW_VERSION=$$NEW_VER; \
	else \
		echo "  Skipping version bump → building and deploying current v$$CURRENT..."; \
		$(MAKE) deploy-current; \
	fi

release-minor: lint test ## 🚀 Lint, test, confirm; [y] bump minor + release, [n] build & deploy current version only
	@VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	CURRENT=$$(cat "$$VERSION_FILE" 2>/dev/null || echo "0.0.0"); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	NEW_VER=$$MAJOR.$$((MINOR + 1)).0; \
	echo ""; echo "  Current: $$CURRENT  →  Bump to (minor): $$NEW_VER"; echo -n "  Bump version and release? [y/n] "; \
	read -r ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		$(MAKE) release NEW_VERSION=$$NEW_VER; \
	else \
		echo "  Skipping version bump → building and deploying current v$$CURRENT..."; \
		$(MAKE) deploy-current; \
	fi

release-major: lint test ## 🚀 Lint, test, confirm; [y] bump major + release, [n] build & deploy current version only
	@VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	CURRENT=$$(cat "$$VERSION_FILE" 2>/dev/null || echo "0.0.0"); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	NEW_VER=$$((MAJOR + 1)).0.0; \
	echo ""; echo "  Current: $$CURRENT  →  Bump to (major): $$NEW_VER"; echo -n "  Bump version and release? [y/n] "; \
	read -r ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		$(MAKE) release NEW_VERSION=$$NEW_VER; \
	else \
		echo "  Skipping version bump → building and deploying current v$$CURRENT..."; \
		$(MAKE) deploy-current; \
	fi

# Generic function to build and push image based on VERSION file
# Usage: $(call build-image-from-version,service-name,context-path,dockerfile-path,platform)
define build-image-from-version
	VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	if [ ! -f "$$VERSION_FILE" ]; then \
		echo "❌ ERROR: VERSION file not found at $$VERSION_FILE"; \
		exit 1; \
	fi; \
	VERSION=$$(cat "$$VERSION_FILE" | tr -d '[:space:]'); \
	if [ -z "$$VERSION" ]; then \
		echo "❌ ERROR: VERSION file is empty"; \
		exit 1; \
	fi; \
	DOCKER_TAG="v$$VERSION"; \
	REGISTRY_VAL="$(REGISTRY)"; \
	IMAGE_TAG="$$REGISTRY_VAL/$(1):$$DOCKER_TAG"; \
	PLATFORM_FLAG=""; \
	if [ -n "$(4)" ]; then \
		PLATFORM_FLAG="--platform $(4)"; \
	fi; \
	echo "📋 Building $(1) image from VERSION file: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(2)/$(3) \
		$$PLATFORM_FLAG \
		--push \
		-t $$IMAGE_TAG \
		$(2)
endef

# Generic function to build and push multi-platform image based on VERSION file
# Usage: $(call build-multi-platform-image,service-name,context-path,dockerfile-path,platforms)
define build-multi-platform-image
	VERSION_FILE="$(ROOT_DIR)/VERSION"; \
	if [ ! -f "$$VERSION_FILE" ]; then \
		echo "❌ ERROR: VERSION file not found at $$VERSION_FILE"; \
		exit 1; \
	fi; \
	VERSION=$$(cat "$$VERSION_FILE" | tr -d '[:space:]'); \
	if [ -z "$$VERSION" ]; then \
		echo "❌ ERROR: VERSION file is empty"; \
		exit 1; \
	fi; \
	DOCKER_TAG="v$$VERSION"; \
	REGISTRY_VAL="$(REGISTRY)"; \
	IMAGE_TAG="$$REGISTRY_VAL/$(1):$$DOCKER_TAG"; \
	PLATFORMS="$(4)"; \
	if [ -z "$$PLATFORMS" ]; then \
		PLATFORMS="$(DEFAULT_PUSH_PLATFORM)"; \
	fi; \
	echo "📋 Building multi-platform $(1) image from VERSION file: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️  Platforms: $$PLATFORMS"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(2)/$(3) \
		--platform $$PLATFORMS \
		--push \
		-t $$IMAGE_TAG \
		$(2)
endef

ensure-buildx-builder: ## 🔧 Ensure Docker Buildx builder exists
	@if ! docker buildx ls | grep -q "$(BUILDX_BUILDER)"; then \
		echo "📦 Creating Docker Buildx builder: $(BUILDX_BUILDER)"; \
		docker buildx create --name $(BUILDX_BUILDER) --use --driver docker-container || true; \
	else \
		echo "✅ Docker Buildx builder '$(BUILDX_BUILDER)' already exists"; \
	fi
	@docker buildx use $(BUILDX_BUILDER) || true

api-docker-build-local: REGISTRY = $(LOCAL_REGISTRY)
api-docker-build-local: ensure-buildx-builder
	@echo "📦 Building and pushing API image to local registry: $(LOCAL_REGISTRY)"
	$(call build-image-from-version,homepage-api,$(API_CONTEXT),Dockerfile.dev,linux/arm64)
	@VERSION=$$(cat $(ROOT_DIR)/VERSION | tr -d '[:space:]'); \
	echo "✅ Pushed $(LOCAL_REGISTRY)/homepage-api:v$$VERSION to local registry"

frontend-docker-build-local: REGISTRY = $(LOCAL_REGISTRY)
frontend-docker-build-local: ensure-buildx-builder
	@echo "📦 Building and pushing Frontend PRODUCTION image to local registry: $(LOCAL_REGISTRY)"
	@echo "⚠️  Using Dockerfile (production with nginx), not Dockerfile.dev"
	$(call build-image-from-version,homepage-frontend,$(FRONTEND_CONTEXT),Dockerfile,linux/arm64)
	@VERSION=$$(cat $(ROOT_DIR)/VERSION | tr -d '[:space:]'); \
	echo "✅ Pushed $(LOCAL_REGISTRY)/homepage-frontend:v$$VERSION to local registry"

api-build-image: ensure-buildx-builder ## 🏗️ Build and push API image to registry
	@echo "📦 Building and pushing API image to registry: $(REGISTRY)"
	@if [ -z "$(PUSH_PLATFORM)" ]; then \
		$(call build-image-from-version,homepage-api,$(API_CONTEXT),Dockerfile.dev,) \
	else \
		for platform in $$(echo $(PUSH_PLATFORM) | tr ',' ' '); do \
			echo "📦 Building for platform: $$platform"; \
			$(call build-image-from-version,homepage-api,$(API_CONTEXT),Dockerfile,$$platform); \
		done; \
	fi
	@VERSION=$$(cat $(ROOT_DIR)/VERSION | tr -d '[:space:]'); \
	echo "✅ Pushed $(REGISTRY)/homepage-api:v$$VERSION to registry"

frontend-build-image: ensure-buildx-builder ## 🏗️ Build and push Frontend image to registry
	@echo "📦 Building and pushing Frontend image to registry: $(REGISTRY)"
	@if [ -z "$(PUSH_PLATFORM)" ]; then \
		$(call build-image-from-version,homepage-frontend,$(FRONTEND_CONTEXT),Dockerfile,) \
	else \
		for platform in $$(echo $(PUSH_PLATFORM) | tr ',' ' '); do \
			echo "📦 Building for platform: $$platform"; \
			$(call build-image-from-version,homepage-frontend,$(FRONTEND_CONTEXT),Dockerfile,$$platform); \
		done; \
	fi
	@VERSION=$$(cat $(ROOT_DIR)/VERSION | tr -d '[:space:]'); \
	echo "✅ Pushed $(REGISTRY)/homepage-frontend:v$$VERSION to registry"

frontend-build-image-ghcr: REGISTRY = $(GHCR_REGISTRY)
frontend-build-image-ghcr: frontend-lint frontend-test ensure-buildx-builder ## 🏗️ Build and push multi-platform Frontend image to GHCR (runs lint + test first)
	@echo "📦 Building and pushing multi-platform Frontend image to GHCR: $(GHCR_REGISTRY)"
	$(call build-multi-platform-image,homepage-frontend,$(FRONTEND_CONTEXT),Dockerfile,)
	@VERSION=$$(cat $(ROOT_DIR)/VERSION | tr -d '[:space:]'); \
	echo "✅ Pushed $(GHCR_REGISTRY)/homepage-frontend:v$$VERSION to GHCR (multi-platform)"

api-build-image-ghcr: REGISTRY = $(GHCR_REGISTRY)
api-build-image-ghcr: api-lint api-test ensure-buildx-builder ## 🏗️ Build and push multi-platform API image to GHCR (runs lint + test first)
	@echo "📦 Building and pushing multi-platform API image to GHCR: $(GHCR_REGISTRY)"
	$(call build-multi-platform-image,homepage-api,$(API_CONTEXT),Dockerfile,)
	@VERSION=$$(cat $(ROOT_DIR)/VERSION | tr -d '[:space:]'); \
	echo "✅ Pushed $(GHCR_REGISTRY)/homepage-api:v$$VERSION to GHCR (multi-platform)"

build-images-local: lint test ## 🏗️ Lint + test, then build and push both images to local registry
	@echo "📦 Building and pushing images to local registry: $(LOCAL_REGISTRY)"
	@$(MAKE) api-docker-build-local
	@$(MAKE) frontend-docker-build-local
	@echo "✅ All images pushed to local registry"

build-images-ghcr: frontend-build-image-ghcr api-build-image-ghcr ## 🏗️ Lint + test (per component), then build and push both images to GHCR
	@echo "✅ All images pushed to GHCR"

api-docker-run: api-build-image
	docker run --rm -it \
		-p $(API_PORT):8080 \
		$(API_IMAGE):$(API_TAG)

frontend-docker-run: frontend-build-image
	docker run --rm -it \
		-p $(FRONTEND_PORT):8080 \
		$(FRONTEND_IMAGE):$(FRONTEND_TAG)


api-run: ## 🚀 Run API in development mode
	@echo "🚀 Starting API in development mode..."
	@cd $(API_CONTEXT) && go run .

frontend-run: ## 🚀 Run frontend in development mode
	@echo "🚀 Starting frontend in development mode..."
	@cd $(FRONTEND_CONTEXT) && \
		if [ ! -d node_modules ]; then \
			echo "📦 Installing dependencies..."; \
			npm ci --legacy-peer-deps --no-audit --no-fund; \
		fi && \
		npm run dev
