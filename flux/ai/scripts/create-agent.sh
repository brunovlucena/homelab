#!/bin/bash
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Agent Template Generator
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Creates a new agent following best practices:
# - Standard directory structure
# - Makefile with version-bump
# - Kustomization files with image tags
# - VERSION file
#
# Usage: ./scripts/create-agent.sh <agent-name> [agent-type]
#   agent-type: lambda (default) or standard
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AI_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

AGENT_NAME="${1:-}"
AGENT_TYPE="${2:-lambda}"

if [ -z "$AGENT_NAME" ]; then
    echo "Usage: $0 <agent-name> [agent-type]"
    echo "  agent-type: lambda (default) or standard"
    exit 1
fi

AGENT_DIR="$AI_DIR/$AGENT_NAME"

if [ -d "$AGENT_DIR" ]; then
    echo "Error: Agent directory already exists: $AGENT_DIR"
    exit 1
fi

echo "Creating agent: $AGENT_NAME (type: $AGENT_TYPE)"
echo ""

# Create directory structure
mkdir -p "$AGENT_DIR"/{src,k8s/kustomize/{base,pro,studio},tests}

# Create VERSION file
echo "0.1.0" > "$AGENT_DIR/VERSION"

# Create README
cat > "$AGENT_DIR/README.md" <<EOF
# $AGENT_NAME

Agent description goes here.

## Quick Start

\`\`\`bash
# Build locally
make build-local

# Deploy to pro
make deploy-pro

# Deploy to studio
make deploy-studio

# Bump version
make version-bump NEW_VERSION=0.1.1
\`\`\`

## Version Management

This agent follows DRY principles for version management:

- Single source of truth: \`VERSION\` file
- Update all versions: \`make version-bump NEW_VERSION=x.y.z\`
- Auto-bump: \`make release-patch\`, \`make release-minor\`, \`make release-major\`

See \`AGENT_BEST_PRACTICES.md\` for details.
EOF

# Create Makefile based on agent type
if [ "$AGENT_TYPE" = "lambda" ]; then
    cat > "$AGENT_DIR/Makefile" <<'MAKEFILE_EOF'
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Agent Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: help build build-local push push-local deploy deploy-pro deploy-studio test clean version version-bump release release-patch release-minor release-major

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")
LOCAL_REGISTRY := localhost:5001
GHCR_REGISTRY := ghcr.io/brunovlucena
NAMESPACE := AGENT_NAME_PLACEHOLDER
K8S_KUSTOMIZE_DIR := $(ROOT_DIR)/k8s/kustomize

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Build
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

build: ## Build Docker image for GHCR
	@echo "Building image for GHCR v$(VERSION)..."
	docker build -t $(GHCR_REGISTRY)/AGENT_NAME_PLACEHOLDER:$(VERSION) -f src/Dockerfile src/

build-local: ## Build Docker image for local registry
	@echo "Building image for local registry v$(VERSION)..."
	docker build -t $(LOCAL_REGISTRY)/AGENT_NAME_PLACEHOLDER:$(VERSION) -f src/Dockerfile src/

push-local: build-local ## Push to local registry
	docker push $(LOCAL_REGISTRY)/AGENT_NAME_PLACEHOLDER:$(VERSION)

push: build ## Push to GHCR
	docker push $(GHCR_REGISTRY)/AGENT_NAME_PLACEHOLDER:$(VERSION)

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
	@echo "  4. Commit: git add -A && git commit -m 'chore(release): AGENT_NAME_PLACEHOLDER v$(NEW_VERSION)'"
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
MAKEFILE_EOF

    # Replace placeholder
    sed -i.bak "s/AGENT_NAME_PLACEHOLDER/$AGENT_NAME/g" "$AGENT_DIR/Makefile"
    rm -f "$AGENT_DIR/Makefile.bak"
else
    echo "Standard agent type not yet implemented"
    exit 1
fi

# Create base kustomization
cat > "$AGENT_DIR/k8s/kustomize/base/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: $AGENT_NAME

resources:
  - namespace.yaml
  - lambdaagent.yaml

commonLabels:
  app.kubernetes.io/name: $AGENT_NAME
  app.kubernetes.io/part-of: $AGENT_NAME
EOF

# Create namespace
cat > "$AGENT_DIR/k8s/kustomize/base/namespace.yaml" <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: $AGENT_NAME
  labels:
    app.kubernetes.io/name: $AGENT_NAME
EOF

# Create LambdaAgent
cat > "$AGENT_DIR/k8s/kustomize/base/lambdaagent.yaml" <<EOF
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: $AGENT_NAME
  namespace: $AGENT_NAME
  labels:
    app.kubernetes.io/name: $AGENT_NAME
    app.kubernetes.io/component: service
    app.kubernetes.io/part-of: homelab
spec:
  serviceAccountName: $AGENT_NAME-sa
  
  image:
    repository: localhost:5001/$AGENT_NAME
    tag: "0.1.0"
    port: 8080

  ai:
    provider: ollama
    endpoint: "http://ollama-native.ollama.svc.cluster.local:11434"
    model: "llama3.2:3b"

  behavior:
    emitEvents: true

  scaling:
    minReplicas: 0
    maxReplicas: 5
    targetConcurrency: 5
    scaleToZeroGracePeriod: 30s

  resources:
    requests:
      memory: "256Mi"
    limits:
      memory: "512Mi"

  eventing:
    enabled: true
    eventSource: "/$AGENT_NAME"
EOF

# Create pro overlay
cat > "$AGENT_DIR/k8s/kustomize/pro/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../base

# Note: LambdaAgent CRD uses spec.image.repository + tag (not container image),
# so we use patches instead of kustomize images: section
patches:
  - target:
      kind: LambdaAgent
      name: $AGENT_NAME
    patch: |-
      - op: replace
        path: /spec/image/tag
        value: "0.1.0"

# 📝 Pro-specific labels
commonLabels:
  environment: pro
EOF

# Create studio overlay
cat > "$AGENT_DIR/k8s/kustomize/studio/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../base

# Note: LambdaAgent CRD uses spec.image.repository + tag (not container image),
# so we use patches instead of kustomize images: section
patches:
  - target:
      kind: LambdaAgent
      name: $AGENT_NAME
    patch: |-
      - op: replace
        path: /spec/image/repository
        value: "ghcr.io/brunovlucena/$AGENT_NAME"
      - op: replace
        path: /spec/image/tag
        value: "0.1.0"

# 📝 Studio-specific labels
commonLabels:
  environment: studio
EOF

# Create basic Dockerfile
cat > "$AGENT_DIR/src/Dockerfile" <<EOF
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "main.py"]
EOF

# Create basic requirements.txt
cat > "$AGENT_DIR/src/requirements.txt" <<EOF
flask>=2.3.0
cloudevents>=1.10.0
kubernetes>=28.0.0
EOF

# Create basic main.py
cat > "$AGENT_DIR/src/main.py" <<EOF
#!/usr/bin/env python3
"""
$AGENT_NAME - Agent implementation
"""

from flask import Flask, request, jsonify
from cloudevents.http import from_http

app = Flask(__name__)

@app.route("/", methods=["POST"])
def handle():
    """Handle CloudEvent requests."""
    event = from_http(request.headers, request.get_data())
    
    # Process event
    result = {
        "status": "success",
        "event_type": event.get("type"),
        "data": event.get("data")
    }
    
    return jsonify(result), 200

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
EOF

chmod +x "$AGENT_DIR/src/main.py"

echo ""
echo "✅ Agent created successfully!"
echo ""
echo "Next steps:"
echo "  1. cd $AGENT_DIR"
echo "  2. Review and customize the generated files"
echo "  3. Add your agent logic to src/main.py"
echo "  4. Test: make build-local"
echo "  5. Deploy: make deploy-pro"
echo ""
