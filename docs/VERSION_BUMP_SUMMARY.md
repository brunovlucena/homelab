# Version Bump & Build Summary - Refactoring Release

**Date:** December 10, 2025  
**Reason:** Refactoring all agents to use shared business logic between API and CloudEvent handlers, with OpenTelemetry tracing

---

## 📦 Version Bumps

All agent versions have been bumped (patch increment):

| Agent | Old Version | New Version | Change |
|-------|-------------|-------------|--------|
| **agent-bruno** | 1.2.0 | **1.2.1** | +0.0.1 |
| **agent-redteam** | 1.1.0 | **1.1.1** | +0.0.1 |
| **agent-blueteam** | v1.1.0 | **1.1.1** | +0.0.1 (normalized) |
| **agent-contracts** | 1.2.0 | **1.2.1** | +0.0.1 |
| **agent-tools** | v1.1.0 | **1.1.1** | +0.0.1 (normalized) |
| **agent-restaurant** | 0.2.0 | **0.2.1** | +0.0.1 |
| **agent-pos-edge** | 0.2.0 | **0.2.1** | +0.0.1 |
| **agent-chat** | 1.1.0 | **1.1.1** | +0.0.1 |
| **agent-store-multibrands** | 0.2.0 | **0.2.1** | +0.0.1 |
| **agent-rpg** | 1.1.0 | **1.1.1** | +0.0.1 |
| **demo-mag7-battle** | v1.1.0 | **1.1.1** | +0.0.1 (normalized) |

---

## 🏗️ Build Instructions

### Option 1: Build All Agents (Recommended)

```bash
# Build all agents
./scripts/build-all-agents.sh

# Build and push all agents
./scripts/build-all-agents.sh ghcr.io/brunovlucena true
```

### Option 2: Build Individual Agents

Each agent has a Makefile with build targets:

```bash
# Example: Build agent-bruno
cd flux/ai/agent-bruno
make build
make push  # Optional: push to registry
```

### Option 3: Use GitHub Actions

The CI/CD workflows will automatically build and push images when you commit the version changes.

---

## 📊 Grafana Dashboard

### New Dashboard: Agent Versions QA Dashboard

**Location:** `flux/infrastructure/prometheus-operator/k8s/dashboards/agent-versions-dashboard.json`

**ConfigMap:** `flux/infrastructure/prometheus-operator/k8s/dashboards/agent-versions-dashboard-configmap.yaml`

**Features:**
- ✅ Shows all agent versions in a single table
- ✅ Highlights outdated versions in **RED**
- ✅ Shows version statistics (Total, Outdated, Up-to-date, Missing)
- ✅ Filters by namespace
- ✅ Auto-refreshes every 30 seconds

**Outdated Version Detection:**
- Versions older than `1.1.1` or `0.2.1` are marked as outdated
- Versions matching `1.2.1`, `1.2.0`, `1.1.1`, `0.2.1` are considered up-to-date
- Versions like `1.1.0`, `0.2.0` are marked as outdated (highlighted in RED)

**Access:**
- Dashboard UID: `agent-versions`
- Tags: `agents`, `versions`, `qa`, `monitoring`
- Auto-discovered by Grafana via ConfigMap labels

---

## 🔍 How to Verify Versions

### 1. Check VERSION Files

```bash
# List all versions
for agent in flux/ai/agent-*/VERSION; do
    echo "$(basename $(dirname $agent)): $(cat $agent)"
done
```

### 2. Check Grafana Dashboard

1. Open Grafana
2. Navigate to "🤖 Agent Versions - QA Dashboard"
3. View the table - outdated versions will be highlighted in RED

### 3. Query Prometheus Directly

```bash
# Query all agent build_info metrics
kubectl port-forward -n prometheus svc/kube-prometheus-stack-prometheus 9090:9090

# Then query:
curl 'http://localhost:9090/api/v1/query?query={__name__=~"agent_.*_build_info"}'
```

### 4. Check Running Pods

```bash
# Check image tags in running pods
kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.namespace}{"\t"}{.metadata.name}{"\t"}{.spec.containers[*].image}{"\n"}{end}' | grep agent
```

---

## 🚀 Deployment Steps

### 1. Build Images

```bash
# Build all agents
./scripts/build-all-agents.sh ghcr.io/brunovlucena true
```

### 2. Deploy Dashboard

The dashboard ConfigMap will be automatically discovered by Grafana. To apply manually:

```bash
kubectl apply -f flux/infrastructure/prometheus-operator/k8s/dashboards/agent-versions-dashboard-configmap.yaml
```

### 3. Update Kustomizations (if needed)

If your agents use Kustomize overlays, update the image tags:

```bash
# Example for agent-bruno
cd flux/ai/agent-bruno/k8s/kustomize
# Update base/lambdaagent.yaml or overlay kustomization.yaml with new image tag
```

### 4. Apply via Flux

If using Flux CD, commit the changes and Flux will reconcile:

```bash
git add .
git commit -m "chore: bump agent versions to 1.1.1/1.2.1 after refactoring"
git push
```

---

## ✅ Verification Checklist

- [x] All VERSION files bumped
- [x] Grafana dashboard created
- [x] Dashboard ConfigMap created
- [x] Build script created
- [ ] Images built and pushed
- [ ] Dashboard deployed to Grafana
- [ ] Versions verified in Grafana
- [ ] Outdated versions highlighted in RED

---

## 📝 Notes

- **Version Format:** Normalized to remove 'v' prefix (e.g., `v1.1.0` → `1.1.1`)
- **Outdated Detection:** Based on semantic versioning - versions below `1.1.1` or `0.2.1` are considered outdated
- **Dashboard Refresh:** Auto-refreshes every 30 seconds
- **Namespace Filtering:** Dashboard supports filtering by namespace

---

*Version bump completed: December 10, 2025*
