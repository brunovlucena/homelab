# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏗️ PROMETHEUS EVENTS - MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy prometheus-events service
#
#	🔧 USAGE:
#	  make                  - Show help
#	  make build-local      - Build and push to localhost:5001
#	  make build-ghcr       - Build and push to ghcr.io
#	  make build-all        - Build and push to both registries
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.DEFAULT_GOAL := help

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 PATH & PLATFORM CONFIGURATION                                       │
# └─────────────────────────────────────────────────────────────────────────┘

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null | tr -d '[:space:]' || echo "0.1.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

BUILDX_BUILDER ?= homelab
PUSH_PLATFORM ?= linux/amd64,linux/arm64
LOAD_PLATFORM ?= linux/arm64

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

# Local registry
LOCAL_REGISTRY := localhost:5001
LOCAL_IMAGE := $(LOCAL_REGISTRY)/prometheus-events

# Remote registry
GHCR_REGISTRY := ghcr.io
GHCR_IMAGE := $(GHCR_REGISTRY)/brunovlucena/prometheus-events

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
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏗️ Prometheus Events$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔧 BUILDX BUILDER                                                      │
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

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER BUILD TARGETS                                                │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build-local build-ghcr build-all
build-local: ensure-buildx-builder ## 🐳 Build and push to localhost:5001 (arm64)
	@$(call print_status,🐳 Building prometheus-events for local registry...)
	@VERSION=$(VERSION); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(LOCAL_IMAGE):$$DOCKER_TAG"; \
	IMAGE_LATEST="$(LOCAL_IMAGE):latest"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️ Platform: $(LOAD_PLATFORM)"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(LOAD_PLATFORM) \
		--push \
		-t $$IMAGE_TAG \
		-t $$IMAGE_LATEST \
		.
	@$(call print_success,Built and pushed to $(LOCAL_IMAGE))

build-ghcr: ensure-buildx-builder ## 🐳 Build and push to ghcr.io (multi-arch)
	@$(call print_status,🐳 Building prometheus-events for GHCR...)
	@VERSION=$(VERSION); \
	DOCKER_TAG="v$$VERSION"; \
	IMAGE_TAG="$(GHCR_IMAGE):$$DOCKER_TAG"; \
	IMAGE_LATEST="$(GHCR_IMAGE):latest"; \
	echo "📋 Version: $$VERSION"; \
	echo "🐳 Image: $$IMAGE_TAG"; \
	echo "🏗️ Platforms: $(PUSH_PLATFORM)"; \
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f Dockerfile \
		--platform $(PUSH_PLATFORM) \
		--push \
		-t $$IMAGE_TAG \
		-t $$IMAGE_LATEST \
		.
	@$(call print_success,Built and pushed to $(GHCR_IMAGE))

build-all: build-local build-ghcr ## 🐳 Build and push to both localhost:5001 and ghcr.io
	@$(call print_success,All images built and pushed)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES DEPLOYMENT                                               │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: deploy deploy-studio
deploy: ## ☸️ Deploy to Kubernetes (uses base)
	@$(call print_status,☸️ Deploying prometheus-events...)
	kubectl apply -k k8s/base/
	@$(call print_success,Deployed to Kubernetes)

deploy-studio: ## ☸️ Deploy to Kubernetes (uses studio overlay)
	@$(call print_status,☸️ Deploying prometheus-events (studio overlay)...)
	kubectl apply -k k8s/overlays/studio/
	@$(call print_success,Deployed to Kubernetes (studio))

.PHONY: version version-bump release-patch release-minor release-major
version: ## 🏷️ Show current version
	@echo "$(COLOR_BOLD)Current version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)VERSION file:$(COLOR_RESET) $(VERSION_FILE)"
	@echo "$(COLOR_BOLD)Git commit:$(COLOR_RESET) $(GIT_COMMIT)"
	@echo "$(COLOR_BOLD)Git branch:$(COLOR_RESET) $(GIT_BRANCH)"
	@echo "$(COLOR_BOLD)Image tag:$(COLOR_RESET) v$(VERSION)"
	@echo "$(COLOR_BOLD)Local image:$(COLOR_RESET) $(LOCAL_IMAGE):v$(VERSION)"
	@echo "$(COLOR_BOLD)GHCR image:$(COLOR_RESET) $(GHCR_IMAGE):v$(VERSION)"

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
	# Update base kustomization.yaml
	@sed -i.bak 's|newTag: v[0-9.]*|newTag: v$(NEW_VERSION)|g' k8s/base/kustomization.yaml && rm -f k8s/base/kustomization.yaml.bak; \
	echo "  ✅ Updated k8s/base/kustomization.yaml"; \
	# Update overlay kustomizations
	@for overlay in k8s/overlays/*/kustomization.yaml; do \
		if [ -f "$$overlay" ]; then \
			sed -i.bak 's|newTag: v[0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay" && rm -f "$$overlay.bak"; \
			echo "  ✅ Updated $$overlay"; \
		fi; \
	done
	@echo ""
	@echo "✅ Version bumped to v$(NEW_VERSION)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): prometheus-events v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release-patch: ## 🏷️ Bump patch version (x.y.Z → x.y.(Z+1))
	@CURRENT=$$(echo $(VERSION) | cut -d. -f1-2); \
	PATCH=$$(echo $(VERSION) | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_VERSION="$${CURRENT}.$${NEW_PATCH}"; \
	$(MAKE) version-bump NEW_VERSION=$$NEW_VERSION

release-minor: ## 🏷️ Bump minor version (x.Y.z → x.(Y+1).0)
	@MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	CURRENT_MINOR=$$(echo $(VERSION) | cut -d. -f2); \
	NEW_MINOR=$$((CURRENT_MINOR + 1)); \
	NEW_VERSION="$${MAJOR}.$${NEW_MINOR}.0"; \
	$(MAKE) version-bump NEW_VERSION=$$NEW_VERSION

release-major: ## 🏷️ Bump major version (X.y.z → (X+1).0.0)
	@CURRENT_MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	NEW_MAJOR=$$((CURRENT_MAJOR + 1)); \
	NEW_VERSION="$${NEW_MAJOR}.0.0"; \
	$(MAKE) version-bump NEW_VERSION=$$NEW_VERSION

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔥 PREWARM                                                              │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: prewarm
prewarm: build-local ## 🔥 Prewarm: Build and push to localhost:5001
	@$(call print_success,Image prewarmed at $(LOCAL_IMAGE):v$(VERSION))
