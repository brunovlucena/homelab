.PHONY: build push deploy test version version-bump release release-patch release-minor release-major copy-shared-lib

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "1.0.0")
REGISTRY ?= localhost:5001
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize
SRC_DIR := $(ROOT_DIR)/src
IMAGE := $(REGISTRY)/agent-tools/k8s-tools

copy-shared-lib:
	@rm -rf $(SRC_DIR)/shared-lib
	@cp -r $(ROOT_DIR)/../shared-lib $(SRC_DIR)/shared-lib

build: copy-shared-lib
	docker build -t $(IMAGE):v$(VERSION) -f src/k8s_tools/Dockerfile src/

push: build
	docker push $(IMAGE):v$(VERSION)

deploy:
	kubectl apply -k k8s/kustomize/base

test:
	cd tests && pytest -v

clean:
	kubectl delete -k k8s/kustomize/base --ignore-not-found

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  🏷️ VERSION MANAGEMENT (DRY Pattern)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

version: ## 🏷️ Show current version
	@echo "Current version: $(VERSION)"
	@echo "VERSION file: $(VERSION_FILE)"

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
	# Update kustomization overlays if they exist
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio $(K8S_KUSTOMIZE_DIR)/base; do \
		if [ -f "$$overlay/kustomization.yaml" ]; then \
			if grep -q "newTag:" "$$overlay/kustomization.yaml"; then \
				sed -i.bak 's|newTag: v[0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay/kustomization.yaml" && rm -f "$$overlay/kustomization.yaml.bak"; \
				echo "  ✅ Updated $$overlay/kustomization.yaml"; \
			elif grep -q "path: /spec/image/tag" "$$overlay/kustomization.yaml"; then \
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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): agent-tools v$(NEW_VERSION)'"
	@echo "  5. Push: git push origin main"

release: version-bump build ## 🚀 Full release: bump version + build + deploy (NEW_VERSION=x.y.z)
	@echo "🚀 Deploying v$(NEW_VERSION)..."
	@$(MAKE) deploy
	@echo "✅ Released v$(NEW_VERSION)"

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
