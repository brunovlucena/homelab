# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#  📱 AgentChat - Makefile
#
#  Build, deploy, and manage AgentChat platform
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: help deploy-dev deploy-prod build-images test clean version version-bump release release-patch release-minor release-major

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.1.1")
LOCAL_REGISTRY := localhost:5001
GHCR_REGISTRY := ghcr.io/brunovlucena
NAMESPACE := agent-chat
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

# ═══════════════════════════════════════════════════════════════════════════
# Help
# ═══════════════════════════════════════════════════════════════════════════

help: ## Show this help
	@echo "AgentChat v$(VERSION) - Makefile Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ═══════════════════════════════════════════════════════════════════════════
# Deployment
# ═══════════════════════════════════════════════════════════════════════════

deploy-dev: ## Deploy to development (pro cluster)
	@echo "🚀 Deploying AgentChat to development..."
	kubectl apply -k k8s/kustomize/pro/
	@echo "✅ Deployment complete"

deploy-prod: ## Deploy to production (studio cluster)
	@echo "🚀 Deploying AgentChat to production..."
	kubectl apply -k k8s/kustomize/studio/
	@echo "✅ Deployment complete"

deploy-diff: ## Show diff before deploying
	@echo "📋 Showing deployment diff..."
	kubectl diff -k k8s/kustomize/$(ENV)/

undeploy: ## Remove all AgentChat resources
	@echo "🗑️ Removing AgentChat resources..."
	kubectl delete -k k8s/kustomize/base/ --ignore-not-found
	@echo "✅ Cleanup complete"

# ═══════════════════════════════════════════════════════════════════════════
# Building
# ═══════════════════════════════════════════════════════════════════════════

AGENTS := messaging-hub voice-agent media-agent location-agent command-center

build: ## Build all agent images for GHCR
	@echo "📱 Building AgentChat Images for GHCR v$(VERSION)..."
	@for agent in $(AGENTS); do \
		echo "🏗️ Building $$agent..."; \
		docker build -t $(GHCR_REGISTRY)/$$agent:v$(VERSION) src/$$agent/; \
	done
	@echo "✅ All images built for GHCR"

build-local: ## Build and push to local registry
	@echo "📱 Building AgentChat Images for local registry v$(VERSION)..."
	@for agent in $(AGENTS); do \
		echo "🏗️ Building $$agent..."; \
		docker build -t $(LOCAL_REGISTRY)/agent-chat/$$agent:v$(VERSION) src/$$agent/; \
	done
	@echo "✅ All images built for local registry"

push-local: build-local ## Push all images to local registry
	@echo "📤 Pushing images to local registry..."
	@for agent in $(AGENTS); do \
		docker push $(LOCAL_REGISTRY)/agent-chat/$$agent:v$(VERSION); \
	done
	@echo "✅ Images pushed to local registry"

push: build ## Push all images to GHCR
	@echo "📤 Pushing images to GHCR..."
	@for agent in $(AGENTS); do \
		docker push $(GHCR_REGISTRY)/$$agent:v$(VERSION); \
	done
	@echo "✅ Images pushed to GHCR"

push-%: ## Push specific image (e.g., make push-messaging-hub)
	@echo "📤 Pushing $*..."
	docker push $(GHCR_REGISTRY)/agent-chat/$*:v$(VERSION)

# ═══════════════════════════════════════════════════════════════════════════
# Development
# ═══════════════════════════════════════════════════════════════════════════

web-dev: ## Run web command center locally
	@echo "🌐 Starting web command center..."
	cd web-command-center && npm run dev

web-install: ## Install web dependencies
	cd web-command-center && npm install

# ═══════════════════════════════════════════════════════════════════════════
# Monitoring
# ═══════════════════════════════════════════════════════════════════════════

status: ## Show deployment status
	@echo "📊 AgentChat Status"
	@echo ""
	@echo "=== LambdaAgents ==="
	kubectl get lambdaagents -n $(NAMESPACE) 2>/dev/null || echo "No agents deployed"
	@echo ""
	@echo "=== Knative Services ==="
	kubectl get ksvc -n $(NAMESPACE) 2>/dev/null || echo "No services deployed"
	@echo ""
	@echo "=== Pods ==="
	kubectl get pods -n $(NAMESPACE) 2>/dev/null || echo "No pods running"

logs: ## Follow agent logs
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/part-of=agent-chat -f --all-containers

logs-hub: ## Follow messaging hub logs
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=messaging-hub -f

logs-voice: ## Follow voice agent logs
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=voice-agent -f

logs-media: ## Follow media agent logs
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=media-agent -f

logs-location: ## Follow location agent logs
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=location-agent -f

# ═══════════════════════════════════════════════════════════════════════════
# Testing
# ═══════════════════════════════════════════════════════════════════════════

test: ## Run all tests
	@echo "🧪 Running tests..."
	cd src && go test ./...

test-k6: ## Run k6 load tests
	@echo "📈 Running k6 tests..."
	kubectl apply -f k8s/tests/
	kubectl get testruns -n $(NAMESPACE) -w

# ═══════════════════════════════════════════════════════════════════════════
# Cleanup
# ═══════════════════════════════════════════════════════════════════════════

clean: ## Clean up local artifacts
	@echo "🧹 Cleaning up..."
	rm -rf web-command-center/.next
	rm -rf web-command-center/node_modules
	@echo "✅ Cleanup complete"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🏷️ VERSION MANAGEMENT (DRY Pattern)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-chat v$(NEW_VERSION)'"
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
