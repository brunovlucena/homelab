# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏥 MEDICAL-SERVICE-PLATFORM - UNIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the Medical Service Platform
#	             Integration layer for agent-medical and agents-whatsapp-rust
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

PROJECT_NAME := medical-service-platform
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER CONFIGURATION                                                │
# └─────────────────────────────────────────────────────────────────────────┘

REGISTRY ?= localhost:5001
IMAGE_PREFIX := $(REGISTRY)/medical-service-platform
MEDICAL_SERVICE_IMAGE := $(IMAGE_PREFIX)/medical-service

# Remote registry
DOCKER_REGISTRY := ghcr.io
DOCKER_REPO := $(DOCKER_REGISTRY)/brunovlucena/medical-service
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
NAMESPACE ?= medical-service-platform

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
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏥 Medical Service Platform$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

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

.PHONY: build-local
build-local: ensure-buildx-builder ## 🐳 Build and push to local registry (arm64)
	$(call print_status,📋 Building medical-service-platform version: $(VERSION))
	$(call print_status,🐳 Image: $(MEDICAL_SERVICE_IMAGE):$(VERSION))
	$(call print_status,🏗️ Platform: $(LOAD_PLATFORM))
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(LOAD_PLATFORM) \
		-t $(MEDICAL_SERVICE_IMAGE):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--no-cache \
		-f $(SRC_DIR)/medical-service/Dockerfile \
		$(SRC_DIR)/medical-service/ \
		--push
	$(call print_success,Built and pushed to local registry)
	@echo "   - $(MEDICAL_SERVICE_IMAGE):$(VERSION)"

.PHONY: build
build: ## 🐳 Build Docker image for GHCR
	$(call print_status,Building medical-service-platform v$(VERSION) for GHCR...)
	docker build -t $(DOCKER_REPO):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f $(SRC_DIR)/medical-service/Dockerfile \
		$(SRC_DIR)/medical-service/
	$(call print_success,Built $(DOCKER_REPO):$(VERSION))

.PHONY: push
push: build ## 🚀 Push to GHCR
	$(call print_status,Pushing medical-service-platform image to GHCR...)
	docker push $(DOCKER_REPO):$(VERSION)
	$(call print_success,Pushed $(DOCKER_REPO):$(VERSION))

.PHONY: push-local
push-local: build-local ## 🚀 Build and push to local registry
	$(call print_success,Already pushed to local registry via build-local)

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
	@grep -h "value:" $(K8S_KUSTOMIZE_DIR)/*/kustomization.yaml 2>/dev/null | grep tag | sed 's/^/  /' || echo "  (no tags found)"

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
	# Update base deployment.yaml tag
	@if [ -f "$(K8S_KUSTOMIZE_DIR)/base/deployment.yaml" ]; then \
		sed -i.bak 's|image: .*:.*|image: $(MEDICAL_SERVICE_IMAGE):$(NEW_VERSION)|g' "$(K8S_KUSTOMIZE_DIR)/base/deployment.yaml" && rm -f "$(K8S_KUSTOMIZE_DIR)/base/deployment.yaml.bak"; \
		echo "  ✅ Updated base deployment.yaml"; \
	fi
	# Update kustomization overlays (image tags)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "newTag:" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|newTag: .*|newTag: $(NEW_VERSION)|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (image tag)"; \
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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): medical-service-platform v$(NEW_VERSION)'"
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
#  🧪 TESTING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: test
test: ## 🧪 Run tests
	$(call print_status,Running tests...)
	cd $(SRC_DIR)/medical-service && cargo test
	$(call print_success,All tests passed)

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

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  📊 STATUS & MONITORING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: status
status: ## 📊 Show deployment status
	$(call print_status,Deployment status:)
	@echo ""
	@echo "$(COLOR_BOLD)Deployments:$(COLOR_RESET)"
	@kubectl get deployments -n $(NAMESPACE) 2>/dev/null || echo "  No deployments found"
	@echo ""
	@echo "$(COLOR_BOLD)Services:$(COLOR_RESET)"
	@kubectl get svc -n $(NAMESPACE) 2>/dev/null || echo "  No services found"
	@echo ""
	@echo "$(COLOR_BOLD)Pods:$(COLOR_RESET)"
	@kubectl get pods -n $(NAMESPACE) 2>/dev/null || echo "  No pods found"

.PHONY: logs
logs: ## 📜 Tail logs
	kubectl logs -n $(NAMESPACE) -l app=medical-service -f --tail=100

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🧹 CLEANUP
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: clean
clean: ## 🧹 Clean build artifacts
	$(call print_warning,Cleaning...)
	rm -rf target/ Cargo.lock
	find . -type d -name "target" -exec rm -rf {} + 2>/dev/null || true

.PHONY: clean-docker
clean-docker: ## 🧹 Clean Docker images
	$(call print_warning,Cleaning Docker images...)
	docker rmi $(MEDICAL_SERVICE_IMAGE):$(VERSION) 2>/dev/null || true
	docker rmi $(MEDICAL_SERVICE_IMAGE):latest 2>/dev/null || true
