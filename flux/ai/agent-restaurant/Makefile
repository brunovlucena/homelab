# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🍽️ Agent-Restaurant Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

VERSION := $(shell cat VERSION)
GHCR_REGISTRY := ghcr.io/brunovlucena/agent-restaurant
LOCAL_REGISTRY := localhost:5001/agent-restaurant
NAMESPACE := agent-restaurant

.PHONY: help build build-local push push-local deploy test clean web-dev k6-smoke k6-load k6-dining copy-shared-lib

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Copy shared-lib for Docker build
copy-shared-lib:
	@rm -rf src/shared-lib
	@cp -r ../shared-lib src/shared-lib
	@echo "✅ Copied shared-lib to src/"

# Build
build: copy-shared-lib ## Build all agent images for GHCR
	@echo "🍽️ Building Restaurant Agent Images for GHCR v$(VERSION)..."
	docker build --build-arg AGENT_NAME=Pierre --build-arg AGENT_ROLE=waiter \
		-t $(GHCR_REGISTRY)/waiter-agent:$(VERSION) -f src/Dockerfile src/
	docker build --build-arg AGENT_NAME=Marco --build-arg AGENT_ROLE=chef \
		-t $(GHCR_REGISTRY)/chef-agent:$(VERSION) -f src/Dockerfile src/
	docker build --build-arg AGENT_NAME=Isabella --build-arg AGENT_ROLE=sommelier \
		-t $(GHCR_REGISTRY)/sommelier-agent:$(VERSION) -f src/Dockerfile src/
	docker build --build-arg AGENT_NAME=Maximilian --build-arg AGENT_ROLE=host \
		-t $(GHCR_REGISTRY)/host-agent:$(VERSION) -f src/Dockerfile src/
	@echo "✅ All agent images built for GHCR"

build-local: copy-shared-lib ## Build all agent images for local registry
	@echo "🍽️ Building Restaurant Agent Images for local registry v$(VERSION)..."
	docker build --build-arg AGENT_NAME=Pierre --build-arg AGENT_ROLE=waiter \
		-t $(LOCAL_REGISTRY)/waiter-agent:$(VERSION) -f src/Dockerfile src/
	docker build --build-arg AGENT_NAME=Marco --build-arg AGENT_ROLE=chef \
		-t $(LOCAL_REGISTRY)/chef-agent:$(VERSION) -f src/Dockerfile src/
	docker build --build-arg AGENT_NAME=Isabella --build-arg AGENT_ROLE=sommelier \
		-t $(LOCAL_REGISTRY)/sommelier-agent:$(VERSION) -f src/Dockerfile src/
	docker build --build-arg AGENT_NAME=Maximilian --build-arg AGENT_ROLE=host \
		-t $(LOCAL_REGISTRY)/host-agent:$(VERSION) -f src/Dockerfile src/
	@echo "✅ All agent images built for local registry"

push-local: build-local ## Push to local registry
	@echo "📤 Pushing to local registry..."
	docker push $(LOCAL_REGISTRY)/waiter-agent:$(VERSION)
	docker push $(LOCAL_REGISTRY)/chef-agent:$(VERSION)
	docker push $(LOCAL_REGISTRY)/sommelier-agent:$(VERSION)
	docker push $(LOCAL_REGISTRY)/host-agent:$(VERSION)
	@echo "✅ Images pushed to local registry"

build-web: ## Build command center web app
	cd web && npm run build
	docker build -t $(GHCR_REGISTRY)/command-center:$(VERSION) -f web/Dockerfile web/

push: build ## Push images to GHCR
	@echo "📤 Pushing to GHCR..."
	docker push $(GHCR_REGISTRY)/waiter-agent:$(VERSION)
	docker push $(GHCR_REGISTRY)/chef-agent:$(VERSION)
	docker push $(GHCR_REGISTRY)/sommelier-agent:$(VERSION)
	docker push $(GHCR_REGISTRY)/host-agent:$(VERSION)
	@echo "✅ Images pushed to GHCR"

# Deploy
deploy-studio: ## Deploy to studio
	kubectl apply -k k8s/kustomize/studio

deploy-pro: ## Deploy to production
	kubectl apply -k k8s/kustomize/pro

undeploy: ## Remove deployment
	kubectl delete -k k8s/kustomize/studio --ignore-not-found

# Testing
k6-smoke: ## Run k6 smoke test
	kubectl delete testrun agent-restaurant-smoke -n $(NAMESPACE) --ignore-not-found
	kubectl apply -f k8s/tests/k6-smoke.yaml
	@echo "⏳ Waiting for smoke test to complete..."
	kubectl wait --for=jsonpath='{.status.stage}'=finished testrun/agent-restaurant-smoke -n $(NAMESPACE) --timeout=180s
	kubectl logs -n $(NAMESPACE) -l k6_cr=agent-restaurant-smoke --tail=60

k6-load: ## Run k6 load test
	kubectl delete testrun agent-restaurant-load -n $(NAMESPACE) --ignore-not-found
	kubectl apply -f k8s/tests/k6-load.yaml
	@echo "⏳ Waiting for load test to complete..."
	kubectl wait --for=jsonpath='{.status.stage}'=finished testrun/agent-restaurant-load -n $(NAMESPACE) --timeout=600s
	kubectl logs -n $(NAMESPACE) -l k6_cr=agent-restaurant-load --tail=100

k6-dining: ## Run k6 dining flow test
	kubectl delete testrun agent-restaurant-dining-flow -n $(NAMESPACE) --ignore-not-found
	kubectl apply -f k8s/tests/k6-dining-flow.yaml
	@echo "⏳ Waiting for dining flow test to complete..."
	kubectl wait --for=jsonpath='{.status.stage}'=finished testrun/agent-restaurant-dining-flow -n $(NAMESPACE) --timeout=900s
	kubectl logs -n $(NAMESPACE) -l k6_cr=agent-restaurant-dining-flow --tail=100

# Development
web-dev: ## Run web dev server
	cd web && npm run dev

logs: ## Tail agent logs
	kubectl logs -f -l app.kubernetes.io/part-of=agent-restaurant -n $(NAMESPACE)

status: ## Check status
	kubectl get lambdaagents,ksvc -n $(NAMESPACE)

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
			sed -i.bak 's|tag: "[0-9.]*"|tag: "$(NEW_VERSION)"|g' "$$file" && rm -f "$$file.bak"; \
			echo "  ✅ Updated $$file"; \
		fi; \
	done
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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-restaurant v$(NEW_VERSION)'"
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
			sed -i.bak 's|tag: "[0-9.]*"|tag: "$(NEW_VERSION)"|g' "$$file" && rm -f "$$file.bak"; \
			echo "  ✅ Updated $$file"; \
		fi; \
	done
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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-restaurant v$(NEW_VERSION)'"
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
