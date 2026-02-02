# Agent Management Scripts

This directory contains scripts for managing agents and ensuring compliance with best practices.

## Scripts

### `check-compliance.sh`

Checks all agents (or a specific agent) for compliance with best practices.

**Usage:**
```bash
# Check all agents
./scripts/check-compliance.sh

# Check specific agent
./scripts/check-compliance.sh agent-bruno
```

**Checks:**
- ✅ VERSION file exists and is not empty
- ✅ Makefile exists
- ✅ version-bump target exists
- ✅ release-patch/minor/major targets exist
- ✅ Image tags in kustomization overlays
- ✅ VERSION_FILE variable in Makefile

**Exit codes:**
- `0` - All agents compliant
- `1` - Issues found

---

### `validate-makefile.sh`

Validates a single Makefile for best practices.

**Usage:**
```bash
./scripts/validate-makefile.sh flux/ai/agent-bruno/Makefile
```

**Checks:**
- ✅ VERSION_FILE variable exists
- ✅ version-bump target exists
- ⚠️  release-patch/minor/major targets (recommended)

---

### `create-agent.sh`

Creates a new agent following best practices.

**Usage:**
```bash
# Create LambdaAgent-based agent (default)
./scripts/create-agent.sh agent-new-name

# Create standard K8s agent
./scripts/create-agent.sh agent-new-name standard
```

**Creates:**
- ✅ Directory structure (`src/`, `k8s/kustomize/`, `tests/`)
- ✅ `VERSION` file (0.1.0)
- ✅ `Makefile` with version-bump and release targets
- ✅ Base kustomization files
- ✅ Pro and Studio overlays with image tags
- ✅ Basic LambdaAgent CRD
- ✅ Basic Dockerfile and Python code
- ✅ README.md

**Example:**
```bash
cd flux/ai
./scripts/create-agent.sh agent-test

cd agent-test
make build-local
make deploy-pro
```

---

## Integration

### Pre-commit Hooks

The scripts are integrated with pre-commit hooks (see `.pre-commit-config.yaml`):

- **agent-version-check**: Runs `check-compliance.sh` on agent files
- **agent-makefile-check**: Runs `validate-makefile.sh` on Makefiles

**Setup:**
```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Test
pre-commit run --all-files
```

### CI/CD Integration

Add to your CI/CD pipeline:

```yaml
# .github/workflows/agent-compliance.yml
- name: Check Agent Compliance
  run: |
    cd flux/ai
    ./scripts/check-compliance.sh
```

---

## Best Practices

All scripts follow the patterns defined in `AGENT_BEST_PRACTICES.md`:

1. **DRY Principle**: Single VERSION file as source of truth
2. **KISS Principle**: Simple, consistent commands
3. **Standardization**: All agents follow same patterns

---

## Troubleshooting

### Script fails with "command not found"

Make sure scripts are executable:
```bash
chmod +x scripts/*.sh
```

### Pre-commit hook not running

Check installation:
```bash
pre-commit install --hook-type pre-commit
```

### Compliance check fails

Run manually to see detailed output:
```bash
./scripts/check-compliance.sh agent-name
```
