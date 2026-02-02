.PHONY: help deploy deploy-studio deploy-pro undeploy logs url test build version version-bump release release-patch release-minor release-major

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.1.1")
NAMESPACE := agent-rpg
REGISTRY := localhost:5001
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deploy-studio: ## Deploy to studio (local) environment
	kubectl apply -k k8s/kustomize/studio

deploy-pro: ## Deploy to production environment
	kubectl apply -k k8s/kustomize/pro

deploy: deploy-studio ## Default deploy (studio)

undeploy: ## Remove deployment
	kubectl delete -k k8s/kustomize/studio --ignore-not-found

logs: ## Watch all game logs
	kubectl logs -f -l app.kubernetes.io/part-of=agent-rpg -n $(NAMESPACE) --max-log-requests=10

logs-gm: ## Watch game master logs
	kubectl logs -f -l app.kubernetes.io/name=game-master -n $(NAMESPACE)

logs-combat: ## Watch combat-related logs
	kubectl logs -f -l app.kubernetes.io/component=playable-character -n $(NAMESPACE) --max-log-requests=6

url: ## Get game server URL
	@kubectl get ksvc rpg-game-server -n $(NAMESPACE) -o jsonpath='{.status.url}'
	@echo

status: ## Show all game resources
	@echo "=== Agents ==="
	kubectl get lambdaagents -n $(NAMESPACE)
	@echo "\n=== Functions ==="
	kubectl get lambdafunctions -n $(NAMESPACE)
	@echo "\n=== Knative Services ==="
	kubectl get ksvc -n $(NAMESPACE)
	@echo "\n=== Pods ==="
	kubectl get pods -n $(NAMESPACE)

start-game: ## Start a new game
	@URL=$$(kubectl get ksvc rpg-game-server -n $(NAMESPACE) -o jsonpath='{.status.url}') && \
	curl -X POST $$URL/game/new -H "Content-Type: application/json" -d '{"playerId": "bruno"}' | jq

characters: ## List available characters
	@URL=$$(kubectl get ksvc rpg-game-server -n $(NAMESPACE) -o jsonpath='{.status.url}') && \
	curl $$URL/characters | jq

test: ## Run k6 tests
	kubectl apply -f k8s/tests/

clean: ## Clean up test resources
	kubectl delete -f k8s/tests/ --ignore-not-found

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
	@echo "  2. Build: make build"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-rpg v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump ## 🚀 Full release: bump version + deploy (NEW_VERSION=x.y.z)
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
