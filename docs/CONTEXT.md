# 🏠 Homelab Context for AI Assistants

## What is this project?

A production-grade GitOps homelab running multiple Kubernetes clusters with:
- **Flux CD** for GitOps automation
- **Pulumi** for infrastructure as code
- **Knative** for serverless workloads
- **Prometheus/Grafana** for observability

## Key Components

| Component | Location | Purpose |
|-----------|----------|---------|
| Homepage | `flux/infrastructure/homepage/` | Portfolio website (Go API + React) |
| Lambda Operator | `flux/infrastructure/knative-lambda-operator/` | Serverless functions on Knative |
| Prometheus Stack | `flux/infrastructure/prometheus/` | Monitoring & alerting |
| Cloudflare Tunnel | `flux/infrastructure/cloudflare-tunnel-operator/` | Secure ingress |

## Clusters

| Cluster | Purpose | Config |
|---------|---------|--------|
| `pro` | Production workloads | `flux/clusters/pro/` |
| `studio` | Development/testing | `flux/clusters/studio/` |
| `air` | Lightweight workloads | `flux/clusters/air/` |
| `forge` | CI/CD runners | `flux/clusters/forge/` |
| `pi` | Raspberry Pi cluster | `flux/clusters/pi/` |

## Common Tasks

### Deploy a component
```bash
# Homepage
cd flux/infrastructure/homepage && make deploy

# Lambda Operator  
cd flux/infrastructure/knative-lambda-operator/src/operator && make deploy

# Any component via Flux
make reconcile-all
```

### Check cluster health
```bash
make observe           # Flux status
kubectl get pods -A    # All pods
make pf-grafana        # Grafana dashboards
```

### Run tests
```bash
# Homepage
cd flux/infrastructure/homepage/src/api && go test ./...
cd flux/infrastructure/homepage/src/frontend && npm test

# Lambda Operator
cd flux/infrastructure/knative-lambda-operator/src && go test ./...
```

## Architecture Decisions

1. **GitOps-first**: All changes go through Git, Flux reconciles
2. **Multi-cluster**: Workloads isolated by cluster purpose
3. **Local registry**: `localhost:5001` for fast image pulls
4. **Sealed Secrets**: Encrypted secrets in Git

## Files AI Should Read First

When working on a component, read in this order:
1. `README.md` - Overview and quick start
2. `.cursor/rules` - AI-specific instructions
3. `Makefile` - Available commands
4. `docs/` - Architecture and decisions
