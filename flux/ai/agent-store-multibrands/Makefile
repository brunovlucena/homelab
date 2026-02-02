# Agent Store MultiBrands - Makefile
.PHONY: all build push deploy test lint clean version version-bump release release-patch release-minor release-major

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.2.1")
REGISTRY ?= ghcr.io/brunovlucena/agent-store-multibrands
NAMESPACE ?= agent-store-multibrands
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

# Agents to build
AGENTS = whatsapp_gateway ai_seller sales_assistant product_catalog order_processor

all: lint test build

# Build all agent images
build:
	@for agent in $(AGENTS); do \
		echo "Building $$agent..."; \
		docker build -t $(REGISTRY)/$$agent:$(VERSION) -f src/$$agent/Dockerfile src/; \
	done

# Build with host builder (for local testing)
build-host:
	@for agent in $(AGENTS); do \
		func build --path src/$$agent --builder host --image $(REGISTRY)/$$agent:$(VERSION); \
	done

# Push all images
push:
	@for agent in $(AGENTS); do \
		echo "Pushing $$agent..."; \
		docker push $(REGISTRY)/$$agent:$(VERSION); \
	done

# Deploy to cluster
deploy:
	kubectl apply -k k8s/kustomize/studio/

deploy-pro:
	kubectl apply -k k8s/kustomize/pro/

# Run tests
test:
	cd src && pytest ../tests/ -v --cov=. --cov-report=term-missing

test-unit:
	cd src && pytest ../tests/unit/ -v

test-integration:
	cd src && pytest ../tests/integration/ -v

# Lint code
lint:
	cd src && ruff check .
	cd src && mypy . --ignore-missing-imports

# Format code
format:
	cd src && ruff format .
	cd src && isort .

# Clean up
clean:
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type d -name .pytest_cache -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete

# Local development
dev-whatsapp:
	cd src/whatsapp_gateway && uvicorn main:app --reload --port 8081

dev-seller:
	cd src/ai_seller && uvicorn main:app --reload --port 8082

dev-assistant:
	cd src/sales_assistant && uvicorn main:app --reload --port 8083

dev-catalog:
	cd src/product_catalog && uvicorn main:app --reload --port 8084

dev-order:
	cd src/order_processor && uvicorn main:app --reload --port 8085

# Generate k6 test
k6-test:
	kubectl apply -k k8s/tests/

# View logs
logs:
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/part-of=agent-store-multibrands -f --max-log-requests=10

# Port forward for local testing
port-forward-whatsapp:
	kubectl port-forward -n $(NAMESPACE) svc/whatsapp-gateway 8081:80

# Check status
status:
	@echo "=== LambdaAgents ==="
	kubectl get lambdaagents -n $(NAMESPACE)
	@echo "\n=== Knative Services ==="
	kubectl get ksvc -n $(NAMESPACE)
	@echo "\n=== Pods ==="
	kubectl get pods -n $(NAMESPACE)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🏷️ VERSION MANAGEMENT (DRY Pattern)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

version: ## 🏷️ Show current version
	@echo "Current version: $(VERSION)"
	@echo "VERSION file: $(VERSION_FILE)"
	@echo ""
	@echo "Kustomization tags:"
	@grep -h "newTag:" $(K8S_KUSTOMIZE_DIR)/*/kustomization.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no tags found)"

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
	# Update kustomization overlays (images: section for standard K8s)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "newTag:" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|newTag: v[0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (images: section)"; \
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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-store-multibrands v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump build ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
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
