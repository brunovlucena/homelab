# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#  🔐 AGENT DEVSECOPS - MAKEFILE
#
#  Purpose: Build, test, and deploy the DevSecOps agent
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")

# Registry configuration
REGISTRY ?= localhost:5001
GHCR_REGISTRY ?= ghcr.io/brunovlucena
IMAGE_NAME := agent-devsecops/scanner

# Build configuration
BUILDX_BUILDER ?= homelab
LOAD_PLATFORM ?= linux/arm64
PUSH_PLATFORM ?= linux/amd64,linux/arm64

# Kubernetes
K8S_NAMESPACE ?= agent-devsecops
ENV ?= pro

# Colors
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m

.PHONY: help
help: ## 📖 Show this help
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🔐 Agent DevSecOps$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)Current version:$(COLOR_RESET) $(VERSION)"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ VERSION & RELEASE TARGETS                                           │
# │                                                                          │
# │  DRY Principle: VERSION file is the single source of truth              │
# │  These targets update VERSION and all dependent files                    │
# └─────────────────────────────────────────────────────────────────────────┘

K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

.PHONY: version version-bump release release-patch release-minor release-major
version: ## 🏷️ Show current version
	@echo "$(COLOR_BOLD)Current version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)VERSION file:$(COLOR_RESET) $(VERSION_FILE)"
	@echo ""
	@echo "$(COLOR_BOLD)Kustomization tags:$(COLOR_RESET)"
	@grep -h "newTag:" $(K8S_KUSTOMIZE_DIR)/*/kustomization.yaml 2>/dev/null | sed 's/^/  /' || echo "  (no image tags defined)"

version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "$(COLOR_RED)❌ Usage: make version-bump NEW_VERSION=x.y.z$(COLOR_RESET)"; \
		exit 1; \
	fi
	@OLD_VERSION=$(VERSION); \
	echo "$(COLOR_BOLD)$(COLOR_BLUE)🏷️ Bumping version: $$OLD_VERSION → $(NEW_VERSION)$(COLOR_RESET)"; \
	echo ""; \
	echo "$(NEW_VERSION)" > $(VERSION_FILE); \
	echo "  ✅ Updated VERSION file"; \
	# Update base lambdaagent.yaml tag
	@if [ -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" ]; then \
		sed -i.bak 's|tag: "v[0-9.]*[-a-zA-Z0-9.]*"|tag: "v$(NEW_VERSION)"|g' "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml" && rm -f "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml.bak"; \
		echo "  ✅ Updated base lambdaagent.yaml"; \
	fi
	# Update kustomization overlays
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "newTag:" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|newTag: v[0-9.]*[-a-zA-Z0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (images section)"; \
			elif grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|value: "v[0-9.]*[-a-zA-Z0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml (LambdaAgent patches)"; \
			fi; \
		fi; \
		done
	@echo ""
	@echo "$(COLOR_GREEN)✅ Version bumped to $(NEW_VERSION)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Next steps:$(COLOR_RESET)"
	@echo "  1. Review changes: git diff"
	@echo "  2. Build: make build-local"
	@echo "  3. Test locally"
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-devsecops v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump build-local ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🚀 Deploying v$(NEW_VERSION)...$(COLOR_RESET)"
	@$(MAKE) deploy-$(ENV)
	@echo ""
	@echo "$(COLOR_GREEN)✅ Released v$(NEW_VERSION) to $(ENV)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Don't forget to:$(COLOR_RESET)"
	@echo "  git add -A && git commit -m 'chore(release): agent-devsecops v$(NEW_VERSION)' && git push"

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
# │  BUILD                                                                   │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: ensure-buildx-builder
ensure-buildx-builder:
	@if ! docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1; then \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --driver-opt network=host --use || true; \
	fi

.PHONY: build build-local build-push
build: build-local ## 🔨 Build Docker image (local registry)

build-local: ensure-buildx-builder ## 🔨 Build and push to local registry
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🔨 Building agent-devsecops v$(VERSION)$(COLOR_RESET)"
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(ROOT_DIR)/src/scanner/Dockerfile \
		--platform $(LOAD_PLATFORM) \
		--push \
		-t $(REGISTRY)/$(IMAGE_NAME):v$(VERSION) \
		$(ROOT_DIR)/src/scanner
	@echo "$(COLOR_GREEN)✅ Built $(REGISTRY)/$(IMAGE_NAME):v$(VERSION)$(COLOR_RESET)"

build-push: ensure-buildx-builder ## 🚀 Build and push to GHCR (multi-arch)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🚀 Pushing to GHCR v$(VERSION)$(COLOR_RESET)"
	docker buildx build \
		--builder $(BUILDX_BUILDER) \
		-f $(ROOT_DIR)/src/scanner/Dockerfile \
		--platform $(PUSH_PLATFORM) \
		--push \
		-t $(GHCR_REGISTRY)/$(IMAGE_NAME):v$(VERSION) \
		-t $(GHCR_REGISTRY)/$(IMAGE_NAME):latest \
		$(ROOT_DIR)/src/scanner
	@echo "$(COLOR_GREEN)✅ Pushed $(GHCR_REGISTRY)/$(IMAGE_NAME):v$(VERSION)$(COLOR_RESET)"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  TEST                                                                    │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: test lint
test: ## 🧪 Run tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🧪 Running tests$(COLOR_RESET)"
	cd $(ROOT_DIR)/src && python -m pytest -v || true

lint: ## 🔍 Run linting
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🔍 Running linting$(COLOR_RESET)"
	cd $(ROOT_DIR)/src && python -m ruff check . || true

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  DEPLOY                                                                  │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: deploy deploy-pro deploy-studio
deploy: deploy-$(ENV) ## 🚀 Deploy to cluster (ENV=pro|studio)

deploy-pro: ## 🚀 Deploy to pro environment
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🚀 Deploying to pro$(COLOR_RESET)"
	kubectl apply -k $(ROOT_DIR)/k8s/kustomize/pro

deploy-studio: ## 🚀 Deploy to studio environment
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🚀 Deploying to studio$(COLOR_RESET)"
	kubectl apply -k $(ROOT_DIR)/k8s/kustomize/studio

.PHONY: rollout
rollout: ## 🔁 Rollout restart
	kubectl rollout restart deployment/agent-devsecops -n $(K8S_NAMESPACE) || true
	kubectl rollout status deployment/agent-devsecops -n $(K8S_NAMESPACE) --timeout=300s || true

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  RBAC MANAGEMENT                                                         │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: rbac-readonly rbac-operator rbac-admin rbac-status
rbac-status: ## 📋 Show current RBAC level
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)📋 Current RBAC Bindings$(COLOR_RESET)"
	@kubectl get clusterrolebindings -l app.kubernetes.io/name=agent-devsecops -o wide

rbac-readonly: ## 🔒 Set agent to readonly mode (default)
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)🔒 Setting RBAC to readonly$(COLOR_RESET)"
	@kubectl patch lambdaagent agent-devsecops -n $(K8S_NAMESPACE) \
		--type=merge -p '{"spec":{"serviceAccountName":"devsecops-readonly-sa"}}' || \
		echo "$(COLOR_YELLOW)⚠️  LambdaAgent CRD not available, patch deployment manually$(COLOR_RESET)"

rbac-operator: ## 🔧 Set agent to operator mode (limited write)
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)🔧 Setting RBAC to operator$(COLOR_RESET)"
	@kubectl patch lambdaagent agent-devsecops -n $(K8S_NAMESPACE) \
		--type=merge -p '{"spec":{"serviceAccountName":"devsecops-operator-sa"}}' || \
		echo "$(COLOR_YELLOW)⚠️  LambdaAgent CRD not available, patch deployment manually$(COLOR_RESET)"

rbac-admin: ## ⚠️  Set agent to admin mode (EMERGENCY ONLY)
	@echo "$(COLOR_BOLD)\033[31m⚠️  WARNING: Admin mode grants full cluster access$(COLOR_RESET)"
	@read -p "Are you sure? Type 'yes-emergency' to confirm: " confirm; \
	if [ "$$confirm" = "yes-emergency" ]; then \
		kubectl patch lambdaagent agent-devsecops -n $(K8S_NAMESPACE) \
			--type=merge -p '{"spec":{"serviceAccountName":"devsecops-admin-sa"}}'; \
		echo "$(COLOR_BOLD)\033[31m🚨 Admin mode activated - remember to revert!$(COLOR_RESET)"; \
	else \
		echo "Cancelled."; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  SECURITY PROPOSALS                                                      │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: proposals proposals-pending proposals-approve
proposals: ## 📋 List all security proposals
	@kubectl get securityproposals -n $(K8S_NAMESPACE) -o wide 2>/dev/null || \
		echo "No SecurityProposals found (CRD may not be installed)"

proposals-pending: ## 📋 List pending proposals
	@kubectl get securityproposals -n $(K8S_NAMESPACE) \
		-o jsonpath='{range .items[?(@.status.phase=="Pending")]}{.metadata.name}{"\t"}{.spec.severity}{"\t"}{.spec.action}{"\n"}{end}' 2>/dev/null

proposals-approve: ## ✅ Approve a proposal (PROPOSAL=name)
	@if [ -z "$(PROPOSAL)" ]; then \
		echo "Usage: make proposals-approve PROPOSAL=<name>"; \
		exit 1; \
	fi
	kubectl patch securityproposal $(PROPOSAL) -n $(K8S_NAMESPACE) \
		--type=merge -p '{"status":{"phase":"Approved","approvedBy":"manual","approvedAt":"'$$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}}'

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  STATUS                                                                  │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: status logs
status: ## 📊 Show agent status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)📊 Agent DevSecOps Status$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Deployment:$(COLOR_RESET)"
	@kubectl get deployment -n $(K8S_NAMESPACE) -l app.kubernetes.io/name=agent-devsecops -o wide 2>/dev/null || echo "  Not deployed"
	@echo ""
	@echo "$(COLOR_BOLD)Pods:$(COLOR_RESET)"
	@kubectl get pods -n $(K8S_NAMESPACE) -l app.kubernetes.io/name=agent-devsecops -o wide 2>/dev/null || echo "  No pods"
	@echo ""
	@echo "$(COLOR_BOLD)Service Accounts:$(COLOR_RESET)"
	@kubectl get sa -n $(K8S_NAMESPACE) -l app.kubernetes.io/name=agent-devsecops 2>/dev/null || echo "  Not found"

logs: ## 📜 Tail agent logs
	kubectl logs -f -n $(K8S_NAMESPACE) -l app.kubernetes.io/name=agent-devsecops --tail=100
