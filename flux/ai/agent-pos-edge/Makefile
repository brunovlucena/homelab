# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🏪 Agent-POS-Edge Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.2.1")
LOCAL_REGISTRY := localhost:5001
GHCR_REGISTRY := ghcr.io/brunovlucena/agent-pos-edge
NAMESPACE := agent-pos-edge
SRC_DIR := $(ROOT_DIR)/src
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

# Agent images
AGENTS := command-center pos-edge kitchen-agent pump-agent

.PHONY: help build build-local push push-local deploy test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Shared Library
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: copy-shared-lib
copy-shared-lib: ## 📦 Copy shared-lib for Docker build
	@rm -rf $(SRC_DIR)/shared-lib
	@cp -r $(ROOT_DIR)/../shared-lib $(SRC_DIR)/shared-lib
	@echo "✅ Copied shared-lib"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Build
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

build: copy-shared-lib ## Build all agent images for GHCR
	@echo "🏪 Building POS-Edge Images for GHCR v$(VERSION)..."
	@for agent in $(AGENTS); do \
		echo "Building $$agent..."; \
		docker build -t $(GHCR_REGISTRY)/$$agent:v$(VERSION) -f src/$$(echo $$agent | tr '-' '_')/Dockerfile src/; \
	done
	@echo "✅ All images built for GHCR"

build-local: copy-shared-lib ## Build all agent images for local registry
	@echo "🏪 Building POS-Edge Images for local registry v$(VERSION)..."
	@for agent in $(AGENTS); do \
		echo "Building $$agent..."; \
		docker build -t $(LOCAL_REGISTRY)/agent-pos-edge/$$agent:v$(VERSION) -f src/$$(echo $$agent | tr '-' '_')/Dockerfile src/; \
	done
	@echo "✅ All images built for local registry"

build-%: ## Build specific agent (e.g., make build-command-center)
	docker build -t $(GHCR_REGISTRY)/$*:v$(VERSION) -f src/$(subst -,_,$*)/Dockerfile src/

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Push
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

push: build ## Push all agent images to GHCR
	@echo "📤 Pushing to GHCR..."
	@for agent in $(AGENTS); do \
		echo "Pushing $$agent..."; \
		docker push $(GHCR_REGISTRY)/$$agent:v$(VERSION); \
	done
	@echo "✅ Pushed to GHCR"

push-local: build-local ## Push all agent images to local registry
	@echo "📤 Pushing to local registry..."
	@for agent in $(AGENTS); do \
		echo "Pushing $$agent..."; \
		docker push $(LOCAL_REGISTRY)/agent-pos-edge/$$agent:v$(VERSION); \
	done
	@echo "✅ Pushed to local registry"

push-%: build-% ## Push specific agent to GHCR
	docker push $(GHCR_REGISTRY)/$*:v$(VERSION)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Deploy
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

deploy-studio: ## Deploy to studio environment
	kubectl apply -k k8s/kustomize/studio

deploy-pro: ## Deploy to production environment
	kubectl apply -k k8s/kustomize/pro

undeploy: ## Remove deployment
	kubectl delete -k k8s/kustomize/studio --ignore-not-found

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Test
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

test: ## Run unit tests
	cd tests && python -m pytest unit/ -v

test-k6: ## Run k6 load tests
	kubectl apply -k k8s/tests

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Development
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

dev: ## Run locally with docker-compose
	docker-compose -f docker-compose.dev.yaml up -d

dev-down: ## Stop local environment
	docker-compose -f docker-compose.dev.yaml down

logs: ## Tail agent logs
	kubectl logs -f -l app.kubernetes.io/part-of=agent-pos-edge -n $(NAMESPACE)

status: ## Check agent status
	kubectl get lambdaagents -n $(NAMESPACE)
	kubectl get ksvc -n $(NAMESPACE)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Utilities
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

clean: ## Clean build artifacts
	docker rmi $(REGISTRY)/*:$(VERSION) 2>/dev/null || true

version: ## 🏷️ Show current version
	@echo "Current version: $(VERSION)"
	@echo "VERSION file: $(VERSION_FILE)"
	@echo ""
	@echo "Kustomization tags:"
	@grep -h "tag:" $(K8S_KUSTOMIZE_DIR)/base/lambdaagent-*.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no tags found)"

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
	# Update all base LambdaAgent files (multiple files)
	@for file in $(K8S_KUSTOMIZE_DIR)/base/lambdaagent-*.yaml; do \
		if [ -f "$$file" ]; then \
			sed -i.bak 's|tag: "v[0-9.]*"|tag: "v$(NEW_VERSION)"|g' "$$file" && rm -f "$$file.bak"; \
			echo "  ✅ Updated $$file"; \
		fi; \
	done
	# Update kustomization overlays (LambdaAgent patches)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|value: "v[0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-pos-edge v$(NEW_VERSION)'"
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

bump-minor: ## Bump minor version
	@echo $$(awk -F. '{print $$1"."$$2+1".0"}' VERSION) > VERSION
	@cat VERSION
