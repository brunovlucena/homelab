# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Agent Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: help build build-local push push-local deploy deploy-pro deploy-studio test clean version version-bump release release-patch release-minor release-major

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")
LOCAL_REGISTRY := localhost:5001
GHCR_REGISTRY := ghcr.io/brunovlucena
NAMESPACE := agent-screenshot
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Build
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

build: ## Build Docker image for GHCR
	@echo "Building image for GHCR v$(VERSION)..."
	docker build -t $(GHCR_REGISTRY)/agent-screenshot:$(VERSION) -f src/Dockerfile src/

build-local: ## Build Docker image for local registry
	@echo "Building image for local registry v$(VERSION)..."
	docker build -t $(LOCAL_REGISTRY)/agent-screenshot:$(VERSION) -f src/Dockerfile src/

push-local: build-local ## Push to local registry
	docker push $(LOCAL_REGISTRY)/agent-screenshot:$(VERSION)

push: build ## Push to GHCR
	docker push $(GHCR_REGISTRY)/agent-screenshot:$(VERSION)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Deploy
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

deploy-pro: ## Deploy to pro environment
	kubectl apply -k $(K8S_KUSTOMIZE_DIR)/pro

deploy-studio: ## Deploy to studio environment
	kubectl apply -k $(K8S_KUSTOMIZE_DIR)/studio

deploy: deploy-pro ## Default deploy (pro)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Testing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

test: ## Run tests
	@echo "Running tests..."
	# Add your test commands here

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Cleanup
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

clean: ## Clean up resources
	kubectl delete -k $(K8S_KUSTOMIZE_DIR)/pro --ignore-not-found
	kubectl delete -k $(K8S_KUSTOMIZE_DIR)/studio --ignore-not-found

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🏷️ VERSION MANAGEMENT (DRY Pattern)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

version: ## 🏷️ Show current version
	@echo "Current version: $(VERSION)"
	@echo "VERSION file: $(VERSION_FILE)"
	@echo ""
	@echo "Kustomization tags:"
	@grep -h "tag:" $(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no tags found)"

version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z) - DRY pattern
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "❌ Usage: make version-bump NEW_VERSION=x.y.z"; \
		exit 1; \
	fi
	@OLD_VERSION=$(VERSION); \
	echo "🏷️ Bumping version: $$OLD_VERSION → $(NEW_VERSION)"; \
	echo ""; \
	echo "$(NEW_VERSION)" > $(VERSION_FILE); \
	echo "  ✅ Updated VERSION file"; \
	# Update base lambdaagent.yaml tag
	@if [ -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" ]; then \
		sed -i.bak 's|tag: "[0-9.]*"|tag: "$(NEW_VERSION)"|g' "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" && rm -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml.bak"; \
		echo "  ✅ Updated base lambdaagent.yaml"; \
	fi
	# Update kustomization overlays (LambdaAgent patches)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|value: "[0-9.]*"|value: "$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (LambdaAgent patches)"; \
			fi; \
		fi; \
	done
	@echo ""
	@echo "✅ Version bumped to $(NEW_VERSION)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-screenshot v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump build-local ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
	@echo "🚀 Deploying v$(NEW_VERSION)..."
	@$(MAKE) deploy-$(ENV)
	@echo "✅ Released v$(NEW_VERSION) to $(ENV)"

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
