# 🤖 Agent Best Practices & Patterns

This document defines the standard patterns and best practices that all homelab agents and infrastructure components should follow.

## 📋 Table of Contents

1. [Domain Memory (Stateful Agents)](#domain-memory-stateful-agents)
2. [Version Management (DRY Principle)](#version-management-dry-principle)
3. [Kustomization Patterns](#kustomization-patterns)
4. [Makefile Structure](#makefile-structure)
5. [RBAC Patterns](#rbac-patterns)
6. [Directory Structure](#directory-structure)

---

## 🧠 Domain Memory (Stateful Agents)

Following **Nate B. Jones's Domain Memory Factory pattern**, all agents should implement persistent, structured memory to transform from "forgetful entities into disciplined workers."

### Key Concepts

1. **Domain Memory Factory** - Two-agent pattern:
   - **Initializer Agent**: Sets up structured memory (goals, requirements, constraints)
   - **Worker Agent**: Acts upon the memory, making progress in discrete steps

2. **Multi-tiered Memory System**:
   - **Short-term Memory**: Current conversation context (Redis, 1-24h TTL)
   - **Working Memory**: Current task state, goals, progress (Redis, session-based)
   - **Entity Memory**: Structured data about domain objects (PostgreSQL)
   - **User Memory**: User preferences, history, facts (PostgreSQL)
   - **Long-term Memory**: Accumulated knowledge, patterns, learnings (PostgreSQL)

### LambdaAgent Memory Configuration

```yaml
spec:
  memory:
    enabled: true
    schema: chat  # chat, restaurant, medical, security, pos, default
    shortTerm:
      enabled: true
      backend: redis
      redisSecretRef:
        name: agent-redis-secret
        key: url
      ttlSeconds: 86400
    longTerm:
      enabled: true
      backend: postgres
      postgresSecretRef:
        name: agent-postgres-secret
        key: url
    userMemory:
      enabled: true
      storePreferences: true
      storeFacts: true
    workingMemory:
      enabled: true
      trackDecisions: true
      trackProgress: true
    defaultConstraints:
      - description: "Your constraint here"
        hard: true
        category: behavior
```

### Using Domain Memory in Code

```python
from agent_memory import DomainMemoryManager

# Initialize
memory_manager = DomainMemoryManager(
    agent_id="agent-name",
    agent_type="chat",  # or restaurant, medical, etc.
    domain="conversation",
    redis_url=os.getenv("REDIS_URL"),
    postgres_url=os.getenv("POSTGRES_URL"),
)
await memory_manager.connect()

# Start conversation with memory
conv = await memory_manager.start_conversation(
    user_id="user-123",
    conversation_id="conv-456",
)

# Add messages
await memory_manager.add_message(conv, "user", "Hello")

# Build context for LLM (aggregates all memory tiers)
context = await memory_manager.build_context(
    user_id="user-123",
    conversation_id=conv.conversation_id,
    include_user_memory=True,
    include_domain_knowledge=True,
)

# Record learnings to long-term memory
await memory_manager.record_learning(
    domain="conversation",
    content="User prefers detailed explanations",
    source="interaction",
)

# Cleanup
await memory_manager.disconnect()
```

### Domain Memory Schemas

Each agent type has a specialized schema:

| Schema | Use Case | Special Fields |
|--------|----------|----------------|
| `chat` | Conversational agents | conversation_history, topics_discussed, sentiment |
| `restaurant` | Hospitality agents | active_tables, orders, guest_preferences |
| `medical` | Healthcare agents | clinical_context, hipaa_audit_trail |
| `security` | Red/Blue team agents | active_threats, defenses, attack_history |
| `pos` | Retail/POS agents | transactions, inventory_alerts, customer_queue |

### Memory Best Practices

1. **Always initialize memory** in the agent's lifespan handler
2. **Use try/except** when connecting to memory stores (graceful degradation)
3. **Record decisions** with reasoning for explainability
4. **Track progress** for long-running tasks
5. **Use constraints** to define hard boundaries
6. **Record learnings** after successful interactions
7. **Summarize long conversations** to save context window

### HIPAA-Compliant Memory (Medical Agents)

Medical agents must:
- **Hash patient IDs** before storing in memory
- **Never store unencrypted PHI** in any memory tier
- **Log all access** to audit trail
- **Set appropriate constraints** for data access

```python
default_constraints=[
    {"description": "HIPAA compliance - protect all PHI", "hard": True, "category": "privacy"},
    {"description": "Log all data access for audit", "hard": True, "category": "audit"},
]
```

---

## 🏷️ Version Management (DRY Principle)

### ✅ Single Source of Truth

**Rule:** The `VERSION` file is the **ONLY** source of truth for version numbers.

### Pattern

```makefile
# All components MUST have:
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")

# Version-bump target that updates:
# 1. VERSION file
# 2. Base resource files (lambdaagent.yaml, deployment.yaml, etc.)
# 3. All kustomization overlays (pro, studio)
.PHONY: version-bump
version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@# Updates VERSION file
	@# Updates base resources
	@# Updates all overlay kustomizations
```

### Implementation Examples

#### ✅ knative-lambda-operator (Reference Implementation)

```makefile
version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@echo "$(NEW_VERSION)" > $(VERSION_FILE)
	@# Update all kustomization overlays
	@for overlay in $(K8S_DIR)/overlays/*/kustomization.yaml; do \
		sed -i.bak 's|newTag: v[0-9.]*|newTag: v$(NEW_VERSION)|g' "$$overlay"; \
	done
	@# Update OPERATOR_VERSION env var (special case)
	@sed -i.bak 's|value: "v[0-9.]*"|value: "v$(NEW_VERSION)"|g' "$(K8S_DIR)/overlays/studio/kustomization.yaml"
```

#### ✅ agent-bruno / agent-devsecops (LambdaAgent CRD)

```makefile
version-bump: ## 🏷️ Bump version and update all kustomizations (NEW_VERSION=x.y.z)
	@echo "$(NEW_VERSION)" > $(VERSION_FILE)
	@# Update base lambdaagent.yaml
	@sed -i.bak 's|tag: "v[0-9.]*"|tag: "v$(NEW_VERSION)"|g' "$(K8S_KUSTOMIZE_DIR)/base/lambdaagent.yaml"
	@# Update overlay patches (LambdaAgent uses patches, not images: section)
	@for overlay in $(K8S_KUSTOMIZE_DIR)/pro $(K8S_KUSTOMIZE_DIR)/studio; do \
		sed -i.bak 's|value: "v[0-9.]*"|value: "v$(NEW_VERSION)"|g' "$$overlay/kustomization.yaml"; \
	done
```

### Auto-Bump Targets

All components MUST provide:

```makefile
release-patch:    # x.y.Z → x.y.(Z+1)
release-minor:    # x.Y.z → x.(Y+1).0
release-major:    # X.y.z → (X+1).0.0
```

---

## 📦 Kustomization Patterns

### Structure

```
k8s/kustomize/
├── base/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── rbac.yaml
│   └── <resource>.yaml (lambdaagent.yaml, deployment.yaml, etc.)
├── pro/
│   └── kustomization.yaml
└── studio/
    └── kustomization.yaml
```

### Image Management

#### ✅ Standard Kubernetes Resources (Deployments, etc.)

Use `images:` section in kustomization overlays:

```yaml
# pro/kustomization.yaml
images:
  - name: localhost:5001/homepage-api
    newTag: v0.1.8
  - name: localhost:5001/homepage-frontend
    newTag: v0.1.8
```

#### ✅ LambdaAgent CRD

**LambdaAgent uses `spec.image.repository` + `spec.image.tag` separately**, so kustomize's `images:` section doesn't work. Use **patches** instead:

```yaml
# pro/kustomization.yaml
patches:
  - target:
      kind: LambdaAgent
      name: agent-bruno
    patch: |-
      - op: replace
        path: /spec/image/tag
        value: "v1.2.1"
```

```yaml
# studio/kustomization.yaml
patches:
  - target:
      kind: LambdaAgent
      name: agent-bruno
    patch: |-
      - op: replace
        path: /spec/image/repository
        value: ghcr.io/brunovlucena/agent-bruno/chatbot
      - op: replace
        path: /spec/image/tag
        value: "v1.2.1"
```

### Overlay Patterns

#### Pro (Development)
- Local registry: `localhost:5001`
- Lower resource limits
- Debug logging enabled
- Single replica (for operators)

#### Studio (Production)
- GHCR registry: `ghcr.io/brunovlucena`
- Production resource limits
- Info logging
- Multiple replicas (HA)

---

## 🔧 Makefile Structure

### Required Sections

All Makefiles MUST have these sections in order:

```makefile
# 1. Header & Configuration
.DEFAULT_GOAL := help
ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION_FILE := $(ROOT_DIR)/VERSION
VERSION ?= $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")

# 2. Platform Configuration
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
LOAD_PLATFORM ?= linux/arm64
PUSH_PLATFORM ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?= homelab

# 3. Docker Configuration
REGISTRY := localhost:5001
GHCR_REGISTRY ?= ghcr.io/brunovlucena

# 4. Kubernetes Configuration
ENV ?= pro
NAMESPACE ?= <component-name>

# 5. Colors (for output)
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m

# 6. Help
help: ## 📋 Show help
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)<Component Name>$(COLOR_RESET)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'

# 7. Version Management (DRY)
version: ## 🏷️ Show current version
version-bump: ## 🏷️ Bump version (NEW_VERSION=x.y.z)
release-patch: ## 🏷️ Bump patch version
release-minor: ## 🏷️ Bump minor version
release-major: ## 🏷️ Bump major version

# 8. Build
build-local: ## 🔨 Build for local registry
build-push: ## 🚀 Build and push to GHCR

# 9. Test
test: ## 🧪 Run tests
lint: ## 🔍 Run linting

# 10. Deploy
deploy-pro: ## 🚀 Deploy to pro
deploy-studio: ## 🚀 Deploy to studio
deploy: deploy-$(ENV) ## 🚀 Deploy (ENV=pro|studio)

# 11. Status
status: ## 📊 Show status
logs: ## 📜 Tail logs
```

### Required Targets

| Target | Purpose | Required? |
|--------|---------|-----------|
| `help` | Show help message | ✅ Yes |
| `version` | Show current version | ✅ Yes |
| `version-bump NEW_VERSION=x.y.z` | Update version everywhere | ✅ Yes |
| `release-patch/minor/major` | Auto-bump versions | ✅ Yes |
| `build-local` | Build for local registry | ✅ Yes |
| `build-push` | Build and push to GHCR | ✅ Yes |
| `deploy-pro` | Deploy to pro environment | ✅ Yes |
| `deploy-studio` | Deploy to studio environment | ✅ Yes |
| `status` | Show deployment status | ✅ Yes |

---

## 🔐 RBAC Patterns

### Agent RBAC

All agents MUST have:

1. **ServiceAccount** in their namespace
2. **RoleBinding** or **ClusterRoleBinding** (depending on scope)
3. **Labels** for identification:
   ```yaml
   labels:
     app.kubernetes.io/name: <agent-name>
     app.kubernetes.io/component: <component>
     app.kubernetes.io/part-of: homelab-ai
   ```

### Multi-Level RBAC (Security Agents)

For security-focused agents (like `agent-devsecops`), provide multiple RBAC levels:

1. **readonly** - View only (default, safest)
2. **operator** - Limited write (proposals, network policies)
3. **admin** - Full access (emergency only, requires approval)

---

## 📁 Directory Structure

### Standard Structure

```
<component>/
├── VERSION                    # Single source of truth
├── Makefile                   # Standardized build/deploy
├── README.md                  # Documentation
├── k8s/
│   └── kustomize/
│       ├── base/
│       │   ├── kustomization.yaml
│       │   ├── namespace.yaml
│       │   ├── rbac.yaml
│       │   └── <resource>.yaml
│       ├── pro/
│       │   └── kustomization.yaml
│       └── studio/
│           └── kustomization.yaml
└── src/                       # Source code
```

### LambdaAgent Structure

```
agent-<name>/
├── VERSION
├── Makefile
├── README.md
├── k8s/
│   └── kustomize/
│       ├── base/
│       │   ├── kustomization.yaml
│       │   ├── namespace.yaml
│       │   ├── rbac.yaml
│       │   └── lambdaagent.yaml  # LambdaAgent CRD
│       ├── pro/
│       │   └── kustomization.yaml  # Uses patches for image tag
│       └── studio/
│           └── kustomization.yaml  # Uses patches for image tag
└── src/
    └── <component>/
        ├── Dockerfile
        ├── handler.py
        └── main.py
```

---

## ✅ Checklist for New Agents

When creating a new agent, ensure:

- [ ] `VERSION` file exists
- [ ] `Makefile` has `version-bump` target that updates:
  - [ ] VERSION file
  - [ ] Base resource files
  - [ ] All kustomization overlays
- [ ] `Makefile` has `release-patch/minor/major` targets
- [ ] Kustomization overlays use:
  - [ ] `images:` section (for standard K8s resources)
  - [ ] `patches:` (for LambdaAgent CRD)
- [ ] Pro overlay uses `localhost:5001` registry
- [ ] Studio overlay uses `ghcr.io/brunovlucena` registry
- [ ] RBAC configured with proper labels
- [ ] `README.md` documents usage
- [ ] Directory structure follows standard pattern

---

## 🔄 Migration Guide

### Migrating Existing Agents

1. **Add version-bump target** to Makefile
2. **Update kustomization overlays** to use patches (for LambdaAgent) or images: section (for standard resources)
3. **Test version-bump** works correctly
4. **Update documentation**

### Example Migration

**Before:**
```yaml
# pro/kustomization.yaml
patches:
  - patch: |-
      - op: replace
        path: /spec/template/spec/containers/0/image
        value: localhost:5001/agent-bruno/chatbot:latest
```

**After:**
```yaml
# pro/kustomization.yaml
patches:
  - target:
      kind: LambdaAgent
      name: agent-bruno
    patch: |-
      - op: replace
        path: /spec/image/tag
        value: "v1.2.1"
```

Then `make version-bump NEW_VERSION=1.2.2` automatically updates the patch!

---

## 📚 Reference Implementations

| Component | Pattern | Notes |
|-----------|---------|-------|
| `knative-lambda-operator` | ✅ Reference | Uses `images:` section (standard K8s) |
| `homepage` | ✅ Reference | Uses `images:` section (standard K8s) |
| `agent-bruno` | ✅ Reference | Uses `patches:` (LambdaAgent CRD) |
| `agent-devsecops` | ✅ Reference | Uses `patches:` (LambdaAgent CRD) + multi-level RBAC |

---

## 🎯 Key Principles

1. **DRY (Don't Repeat Yourself)**: Version in ONE place (VERSION file)
2. **KISS (Keep It Simple, Stupid)**: Simple `make release-patch` commands
3. **Consistency**: All agents follow the same patterns
4. **Security**: Least privilege RBAC by default
5. **GitOps**: All changes go through Git, Flux reconciles

---

**Last Updated:** 2025-12-10  
**Maintained By:** Homelab Platform Team
