# RBAC & Access Control for Multi-Agent Systems

**[← Back to README](../README.md)** | **[Architecture](ARCHITECTURE.md)** | **[Observability](OBSERVABILITY.md)** | **[Multi-Tenancy](MULTI-TENANCY.md)**

---

## Overview

This document defines Role-Based Access Control (RBAC) strategies for managing multiple AI agents with controlled access to:
- **MCP Servers/Services** (internal Knative services)
- **Cloud Services** (AWS, GCP)
- **Kubernetes Resources** (preventing unauthorized deletions/modifications)

## Table of Contents

1. [Principles](#principles)
2. [Kubernetes RBAC](#kubernetes-rbac)
3. [MCP Server Access Control](#mcp-server-access-control)
4. [Cloud Service Access Control](#cloud-service-access-control)
5. [Agent Identity & Authentication](#agent-identity--authentication)
6. [Implementation Examples](#implementation-examples)
7. [Audit & Monitoring](#audit--monitoring)

---

## Principles

### 🔒 Least Privilege
- Each agent receives **minimum permissions** required for its role
- No agent has cluster-admin or root-level access
- Read-only by default, write access granted explicitly

### 🎭 Agent Roles
Different agents have different responsibilities:
- **SRE Agent**: Infrastructure monitoring, incident response, read-only cluster access
- **DevOps Agent**: Deployment automation, CI/CD, limited write access
- **Security Agent**: Audit logs, compliance checks, read-only security scans
- **Analytics Agent**: Metrics collection, dashboard updates, no resource modifications

### 🔐 Defense in Depth
- Multiple layers of access control
- Kubernetes RBAC + Pod Security Standards
- MCP server authentication tokens
- Cloud IAM policies
- Network policies

### 📊 Auditability
- All agent actions logged
- Trace agent identity in CloudEvents
- Regular access reviews

---

## Kubernetes RBAC

### Agent Service Accounts

Each agent runs with its own ServiceAccount:

```yaml
# agent-bruno/rbac/serviceaccounts.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-sre
  namespace: agent-bruno
  labels:
    app: agent-bruno
    role: sre
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-devops
  namespace: agent-bruno
  labels:
    app: agent-bruno
    role: devops
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-security
  namespace: agent-bruno
  labels:
    app: agent-bruno
    role: security
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-analytics
  namespace: agent-bruno
  labels:
    app: agent-bruno
    role: analytics
```

### Role Definitions

#### 1. SRE Agent Role (Read-Only Infrastructure)

```yaml
# agent-bruno/rbac/role-sre.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agent-bruno-sre-viewer
rules:
# 📖 Read-only access to cluster resources
- apiGroups: [""]
  resources:
    - pods
    - pods/log
    - pods/status
    - services
    - endpoints
    - configmaps
    - persistentvolumeclaims
    - events
  verbs: ["get", "list", "watch"]

# 📊 Metrics & monitoring
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]

# 🚫 NO DELETE permissions
- apiGroups: ["apps"]
  resources:
    - deployments
    - replicasets
    - statefulsets
    - daemonsets
  verbs: ["get", "list", "watch"]  # NO delete, update, patch

# 🔔 Alert rules (read-only)
- apiGroups: ["monitoring.coreos.com"]
  resources:
    - prometheusrules
    - servicemonitors
  verbs: ["get", "list", "watch"]

# 🎯 Knative resources (read-only)
- apiGroups: ["serving.knative.dev"]
  resources:
    - services
    - revisions
    - routes
  verbs: ["get", "list", "watch"]

# 🔒 EXPLICITLY DENY - no deletion anywhere
# This is enforced by NOT including "delete" in any verb list
```

```yaml
# agent-bruno/rbac/rolebinding-sre.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: agent-bruno-sre-viewer-binding
subjects:
- kind: ServiceAccount
  name: agent-bruno-sre
  namespace: agent-bruno
roleRef:
  kind: ClusterRole
  name: agent-bruno-sre-viewer
  apiGroup: rbac.authorization.k8s.io
```

#### 2. DevOps Agent Role (Limited Write Access)

```yaml
# agent-bruno/rbac/role-devops.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: agent-bruno-devops
  namespace: agent-bruno
rules:
# ✅ Can manage its own namespace resources
- apiGroups: [""]
  resources:
    - configmaps
    - secrets  # For MCP API keys
  verbs: ["get", "list", "watch", "create", "update", "patch"]

# ✅ Can scale deployments (but not delete)
- apiGroups: ["apps"]
  resources:
    - deployments/scale
  verbs: ["get", "update", "patch"]

# ✅ Can trigger Knative services
- apiGroups: ["serving.knative.dev"]
  resources:
    - services
  verbs: ["get", "list", "update", "patch"]

# 🚫 CANNOT delete any resources
# 🚫 CANNOT access secrets outside agent-bruno namespace
```

```yaml
# agent-bruno/rbac/rolebinding-devops.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: agent-bruno-devops-binding
  namespace: agent-bruno
subjects:
- kind: ServiceAccount
  name: agent-bruno-devops
  namespace: agent-bruno
roleRef:
  kind: Role
  name: agent-bruno-devops
  apiGroup: rbac.authorization.k8s.io
```

#### 3. Security Agent Role (Audit Access)

```yaml
# agent-bruno/rbac/role-security.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agent-bruno-security-auditor
rules:
# 🔍 Cluster-wide read for security audits
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]

# 🔒 Access to security-related resources
- apiGroups: ["policy"]
  resources:
    - podsecuritypolicies
    - poddisruptionbudgets
  verbs: ["get", "list", "watch"]

- apiGroups: ["networking.k8s.io"]
  resources:
    - networkpolicies
  verbs: ["get", "list", "watch"]

# 🚨 RBAC review (read-only)
- apiGroups: ["rbac.authorization.k8s.io"]
  resources:
    - roles
    - rolebindings
    - clusterroles
    - clusterrolebindings
  verbs: ["get", "list", "watch"]

# 🚫 ABSOLUTELY NO write/delete permissions
```

#### 4. Analytics Agent Role (Metrics Only)

```yaml
# agent-bruno/rbac/role-analytics.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agent-bruno-analytics-reader
rules:
# 📊 Metrics access only
- apiGroups: ["metrics.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list"]

# 📈 Prometheus resources
- apiGroups: ["monitoring.coreos.com"]
  resources:
    - prometheuses
    - prometheusrules
    - servicemonitors
  verbs: ["get", "list", "watch"]

# 📉 Pod metrics
- apiGroups: [""]
  resources:
    - pods
    - nodes
  verbs: ["get", "list"]

# 🚫 NO access to logs, secrets, or configuration
```

### Prevention of Unauthorized Deletions

#### Pod Security Standards

```yaml
# agent-bruno/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: agent-bruno
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

#### Resource Quotas (Prevent Resource Exhaustion)

```yaml
# agent-bruno/rbac/resource-quota.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: agent-bruno-quota
  namespace: agent-bruno
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    persistentvolumeclaims: "5"
    # Prevent creating too many resources
    count/secrets: "20"
    count/configmaps: "30"
    count/services: "10"
```

#### Limit Ranges (Per-Pod Limits)

```yaml
# agent-bruno/rbac/limit-range.yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: agent-bruno-limits
  namespace: agent-bruno
spec:
  limits:
  - max:
      cpu: "2"
      memory: "4Gi"
    min:
      cpu: "100m"
      memory: "128Mi"
    type: Container
```

#### Admission Controllers (Prevent Privileged Escalation)

```yaml
# kube-apiserver configuration (managed by cluster admin)
# Enabled admission plugins:
# - PodSecurity
# - ResourceQuota
# - LimitRanger
# - ServiceAccount
# - DefaultStorageClass
# - ValidatingAdmissionWebhook
# - MutatingAdmissionWebhook

# Custom validating webhook to prevent agent deletions
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: prevent-agent-deletions
webhooks:
- name: validate-deletions.agent-bruno.svc
  rules:
  - operations: ["DELETE"]
    apiGroups: ["*"]
    apiVersions: ["*"]
    resources: ["*"]
  clientConfig:
    service:
      name: deletion-validator
      namespace: agent-bruno
      path: "/validate-deletion"
  admissionReviewVersions: ["v1"]
  sideEffects: None
  failurePolicy: Fail  # Block deletions if webhook fails
  # Only check requests from agent service accounts
  namespaceSelector:
    matchLabels:
      enable-deletion-validation: "true"
```

---

## MCP Server Access Control

### Gateway Options Comparison

When choosing an MCP Server Gateway, you have several options:

| Feature | Ambassador Edge Stack | Kong | Raw Envoy |
|---------|----------------------|------|-----------|
| **Performance** | ⭐⭐⭐⭐⭐ (Envoy-based) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Ease of Config** | ⭐⭐⭐⭐⭐ (K8s CRDs) | ⭐⭐⭐⭐ | ⭐⭐ (complex YAML) |
| **Development Workflow** | ⭐⭐⭐⭐⭐ (Telepresence!) | ⭐⭐ | ⭐ |
| **Kubernetes Native** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **GitOps Friendly** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Rate Limiting** | ✅ Built-in | ✅ Built-in | ⚙️ Manual config |
| **JWT Auth** | ✅ Built-in | ✅ Built-in | ⚙️ Manual config |
| **Web UI** | ✅ | ✅ | ❌ |
| **Telepresence Integration** | ✅ Native | ❌ | ❌ |
| **Plugin Ecosystem** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Best For** | Internal services + dev | External APIs | Maximum control |

**Why Ambassador Edge Stack?**
- **Built on Envoy** (gets all performance benefits without complexity)
- **[Telepresence integration](https://www.getambassador.io/docs/telepresence-oss/latest/quick-start)** enables local MCP server development with remote cluster traffic
- **Kubernetes-native** configuration (fits GitOps workflow with Flux)
- **Simpler than raw Envoy**, more focused than Kong for internal services

**Development Workflow Advantage:**
With Telepresence, developers can intercept MCP server traffic and debug locally:
```bash
# Intercept github-mcp traffic to your laptop
telepresence intercept github-mcp --port 8080 --env-file ./github-mcp.env

# Start local MCP server - receives real agent requests!
python main.py
```

### Architecture

```
┌───────────────────────────────────────────────────────────────────┐
│                        Agent Instances                            │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐   │
│  │ SRE Agent  │  │DevOps Agent│  │Security Agt│  │Analytics   │   │
│  │ (token-A)  │  │ (token-B)  │  │ (token-C)  │  │ (token-D)  │   │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘   │
└────────┼───────────────┼───────────────┼───────────────┼──────────┘
         │               │               │               │
         └───────────────┴───────────────┴───────────────┘
                             │
                             ▼
         ┌────────────────────────────────────────────────┐
         │   Ambassador Edge Stack (Envoy-based)          │
         │   - JWT validation                             │
         │   - Token-based routing & RBAC                 │
         │   - Rate limiting per agent role               │
         │   - Telepresence integration (dev workflow)    │
         │   - K8s-native (CRDs, GitOps-friendly)         │
         └─────────────┬──────────────────────────────────┘
                       │
         ┌─────────────┴───────────────┐
         │                             │
         ▼                             ▼
┌─────────────────┐          ┌─────────────────┐
│ GitHub MCP      │          │ Grafana MCP     │
│ Allowed roles:  │          │ Allowed roles:  │
│ - sre (RO)      │          │ - sre (RO)      │
│ - devops (RW)   │          │ - analytics(RW) │
│ - security (RO) │          │                 │
└─────────────────┘          └─────────────────┘
         ▲                            ▲
         │                            │
         └────────────────┬───────────┘
                          │
                    Telepresence
                   (intercept for
                   local development)
                          ▲
                          │
                    ┌─────┴─────┐
                    │ Developer │
                    │  Laptop   │
                    └───────────┘
```

### Ambassador Implementation

#### Installation via Flux

```yaml
# flux/clusters/homelab/infrastructure/ambassador/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ambassador
---
# flux/clusters/homelab/infrastructure/ambassador/helmrepository.yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: datawire
  namespace: flux-system
spec:
  interval: 1h
  url: https://app.getambassador.io
---
# flux/clusters/homelab/infrastructure/ambassador/helmrelease.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ambassador-edge-stack
  namespace: ambassador
spec:
  interval: 5m
  chart:
    spec:
      chart: edge-stack
      version: 8.x.x
      sourceRef:
        kind: HelmRepository
        name: datawire
        namespace: flux-system
  values:
    replicaCount: 2
    service:
      type: ClusterIP  # Internal for homelab
    
    # Enable Telepresence for development
    telepresence:
      enabled: true
```

#### MCP Server Mappings with RBAC

```yaml
# flux/clusters/homelab/infrastructure/ambassador/mappings/github-mcp.yaml
apiVersion: getambassador.io/v3alpha1
kind: Mapping
metadata:
  name: github-mcp
  namespace: agent-bruno
spec:
  hostname: "*"
  prefix: /mcp/github/
  service: github-mcp.agent-bruno:8080
  timeout_ms: 30000
  
  # Extract agent identity from headers
  add_request_headers:
    x-gateway: "ambassador"
  
  # Rate limiting based on agent role
  labels:
    ambassador:
      - request_label:
        - agent-role:
            header: "X-Agent-Role"
---
# flux/clusters/homelab/infrastructure/ambassador/mappings/grafana-mcp.yaml
apiVersion: getambassador.io/v3alpha1
kind: Mapping
metadata:
  name: grafana-mcp
  namespace: agent-bruno
spec:
  prefix: /mcp/grafana/
  service: grafana-mcp.agent-bruno:8080
  labels:
    ambassador:
      - request_label:
        - agent-role:
            header: "X-Agent-Role"
```

#### Rate Limiting per Agent Role

```yaml
# flux/clusters/homelab/infrastructure/ambassador/ratelimits/agent-limits.yaml
apiVersion: getambassador.io/v3alpha1
kind: RateLimitService
metadata:
  name: agent-ratelimit
  namespace: agent-bruno
spec:
  service: "ratelimit.agent-bruno:8080"

---
# RateLimiting configuration
apiVersion: getambassador.io/v3alpha1
kind: RateLimit
metadata:
  name: agent-role-limits
  namespace: agent-bruno
spec:
  domain: agent-bruno
  
  # Different limits per agent role
  descriptors:
    # SRE Agent: Read-only, lower rate
    - key: agent-role
      value: sre
      rate_limit:
        unit: second
        requests_per_unit: 10
    
    # DevOps Agent: Write operations, medium rate
    - key: agent-role
      value: devops
      rate_limit:
        unit: second
        requests_per_unit: 50
    
    # Analytics Agent: Metrics collection, higher rate
    - key: agent-role
      value: analytics
      rate_limit:
        unit: second
        requests_per_unit: 100
    
    # Security Agent: Audit scans, low rate
    - key: agent-role
      value: security
      rate_limit:
        unit: second
        requests_per_unit: 5
```

#### Development Workflow with Telepresence

Create a helper script for MCP server development:

```bash
#!/bin/bash
# scripts/dev-mcp-server.sh

set -e

MCP_SERVER=${1:-github-mcp}
LOCAL_PORT=${2:-8080}

echo "🚀 Starting development session for $MCP_SERVER"
echo ""

# Connect to homelab cluster
echo "📡 Connecting to homelab cluster..."
telepresence connect

# Check if already intercepting
if telepresence list | grep -q "$MCP_SERVER.*intercepted"; then
  echo "⚠️  Already intercepting $MCP_SERVER, leaving first..."
  telepresence leave $MCP_SERVER
fi

# Intercept the MCP server
echo "🎯 Intercepting $MCP_SERVER traffic..."
telepresence intercept $MCP_SERVER \
  --namespace agent-bruno \
  --port $LOCAL_PORT \
  --env-file ./${MCP_SERVER}.env

echo ""
echo "✅ SUCCESS! Intercepting $MCP_SERVER traffic"
echo ""
echo "📋 Next steps:"
echo "  1. Check environment variables: cat ./${MCP_SERVER}.env"
echo "  2. Start your local MCP server on port $LOCAL_PORT"
echo "  3. All agent requests will hit your laptop!"
echo ""
echo "🛑 To stop intercepting:"
echo "  telepresence leave $MCP_SERVER"
echo ""
```

Usage:
```bash
# Make script executable
chmod +x scripts/dev-mcp-server.sh

# Intercept github-mcp
./scripts/dev-mcp-server.sh github-mcp 8080

# In another terminal, run your local MCP server
cd mcp-servers/github-mcp
source venv/bin/activate
python main.py  # Now receives real traffic from agents!
```

### Token Management

#### MCP API Keys as Kubernetes Secrets

```yaml
# agent-bruno/rbac/mcp-tokens.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mcp-tokens-sre
  namespace: agent-bruno
type: Opaque
stringData:
  github-token: "readonly-github-pat-for-sre"
  grafana-token: "readonly-grafana-key-for-sre"
  # NO deletion or write capabilities
---
apiVersion: v1
kind: Secret
metadata:
  name: mcp-tokens-devops
  namespace: agent-bruno
type: Opaque
stringData:
  github-token: "write-github-pat-for-devops"
  grafana-token: "write-grafana-key-for-devops"
  aws-access-key: "AKIAEXAMPLE"
  aws-secret-key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
---
apiVersion: v1
kind: Secret
metadata:
  name: mcp-tokens-security
  namespace: agent-bruno
type: Opaque
stringData:
  github-token: "audit-github-pat-for-security"
  # NO cloud credentials - security agent audits only
---
apiVersion: v1
kind: Secret
metadata:
  name: mcp-tokens-analytics
  namespace: agent-bruno
type: Opaque
stringData:
  grafana-token: "metrics-write-grafana-key"
  prometheus-token: "prometheus-read-token"
```

#### Agent Deployment with Token Injection

```yaml
# agent-bruno/deployment-sre.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-bruno-sre
  namespace: agent-bruno
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      serviceAccountName: agent-bruno-sre  # 🔑 Links to RBAC
      containers:
      - name: agent
        image: ghcr.io/your-org/agent-bruno:v1.0.0
        env:
        - name: AGENT_ROLE
          value: "sre"
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: mcp-tokens-sre
              key: github-token
        - name: GRAFANA_TOKEN
          valueFrom:
            secretKeyRef:
              name: mcp-tokens-sre
              key: grafana-token
        # 🚫 NO AWS/GCP credentials for SRE agent
```

### MCP Server Authorization Logic

```python
# mcp-servers/github-mcp/auth.py
from enum import Enum
from typing import Optional

class Permission(Enum):
    READ = "read"
    WRITE = "write"
    DELETE = "delete"
    ADMIN = "admin"

# Token -> Permissions mapping
TOKEN_PERMISSIONS = {
    "readonly-github-pat-for-sre": {Permission.READ},
    "write-github-pat-for-devops": {Permission.READ, Permission.WRITE},
    "audit-github-pat-for-security": {Permission.READ},
}

def check_permission(token: str, required: Permission) -> bool:
    """Validate if token has required permission."""
    permissions = TOKEN_PERMISSIONS.get(token, set())
    return required in permissions

def authorize_mcp_request(request, required_permission: Permission):
    """Decorator to enforce MCP server permissions."""
    token = request.headers.get("Authorization", "").replace("Bearer ", "")
    
    if not check_permission(token, required_permission):
        raise PermissionError(
            f"Token does not have {required_permission.value} permission"
        )
    
    # Log for audit
    log_access(
        token=token[:8] + "...",  # Partial token for audit
        endpoint=request.path,
        permission=required_permission.value,
        allowed=True
    )

# Usage in MCP server
@app.post("/repos/{owner}/{repo}/delete")
async def delete_repository(owner: str, repo: str, request: Request):
    authorize_mcp_request(request, Permission.DELETE)  # 🚫 Most agents blocked
    # ... deletion logic
```

### MCP Server Network Policies

```yaml
# agent-bruno/rbac/network-policy-mcp.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-server-ingress
  namespace: agent-bruno
spec:
  podSelector:
    matchLabels:
      app: mcp-server
  policyTypes:
  - Ingress
  ingress:
  # Only allow traffic from agent pods
  - from:
    - podSelector:
        matchLabels:
          app: agent-bruno
    ports:
    - protocol: TCP
      port: 8080
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: agent-egress-restriction
  namespace: agent-bruno
spec:
  podSelector:
    matchLabels:
      app: agent-bruno
  policyTypes:
  - Egress
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  # Allow internal MCP servers
  - to:
    - podSelector:
        matchLabels:
          app: mcp-server
    ports:
    - protocol: TCP
      port: 8080
  # Allow external HTTPS (for external MCP servers)
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
  # 🚫 Block all other egress
```

---

## Cloud Service Access Control

### AWS IAM Integration

#### Workload Identity (IRSA - IAM Roles for Service Accounts)

```yaml
# agent-bruno/rbac/serviceaccount-aws.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-devops
  namespace: agent-bruno
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/AgentBrunoDevOpsRole
```

#### AWS IAM Policy (DevOps Agent)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOnlyEC2",
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "ec2:Get*"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ListS3Buckets",
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::agent-bruno-artifacts/*"
      ]
    },
    {
      "Sid": "DenyDeletion",
      "Effect": "Deny",
      "Action": [
        "ec2:DeleteVolume",
        "ec2:TerminateInstances",
        "s3:DeleteBucket",
        "s3:DeleteObject",
        "rds:DeleteDBInstance",
        "dynamodb:DeleteTable"
      ],
      "Resource": "*"
    },
    {
      "Sid": "AllowCloudWatchMetrics",
      "Effect": "Allow",
      "Action": [
        "cloudwatch:GetMetricStatistics",
        "cloudwatch:ListMetrics"
      ],
      "Resource": "*"
    }
  ]
}
```

#### AWS IAM Policy (SRE Agent - Read Only)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOnlyEverything",
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "s3:Get*",
        "s3:List*",
        "cloudwatch:Get*",
        "cloudwatch:List*",
        "logs:Get*",
        "logs:Describe*",
        "rds:Describe*"
      ],
      "Resource": "*"
    },
    {
      "Sid": "DenyAllMutations",
      "Effect": "Deny",
      "Action": [
        "*:Create*",
        "*:Delete*",
        "*:Update*",
        "*:Put*",
        "*:Modify*"
      ],
      "Resource": "*"
    }
  ]
}
```

#### AWS IAM Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub": "system:serviceaccount:agent-bruno:agent-bruno-devops"
        }
      }
    }
  ]
}
```

### GCP IAM Integration

#### Workload Identity

```yaml
# agent-bruno/rbac/serviceaccount-gcp.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-analytics
  namespace: agent-bruno
  annotations:
    iam.gke.io/gcp-service-account: agent-bruno-analytics@project-id.iam.gserviceaccount.com
```

#### GCP IAM Binding

```bash
# Bind Kubernetes SA to GCP SA
gcloud iam service-accounts add-iam-policy-binding \
  agent-bruno-analytics@project-id.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:project-id.svc.id.goog[agent-bruno/agent-bruno-analytics]"
```

#### GCP IAM Roles (Analytics Agent)

```yaml
# terraform/gcp-iam.tf
resource "google_project_iam_custom_role" "agent_bruno_analytics" {
  role_id     = "agentBrunoAnalytics"
  title       = "Agent Bruno Analytics Role"
  description = "Read-only metrics and write to BigQuery"
  
  permissions = [
    # ✅ Read metrics
    "monitoring.metricDescriptors.list",
    "monitoring.timeSeries.list",
    
    # ✅ Write to BigQuery
    "bigquery.tables.get",
    "bigquery.tables.update",
    "bigquery.tables.updateData",
    
    # ✅ Read GCS
    "storage.buckets.get",
    "storage.objects.get",
    "storage.objects.list",
    
    # 🚫 NO deletion permissions (implicit - not granted)
  ]
}

resource "google_project_iam_binding" "analytics_binding" {
  project = var.project_id
  role    = google_project_iam_custom_role.agent_bruno_analytics.id
  
  members = [
    "serviceAccount:agent-bruno-analytics@project-id.iam.gserviceaccount.com"
  ]
}
```

### Azure AD Integration (AKS)

```yaml
# agent-bruno/rbac/serviceaccount-azure.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-security
  namespace: agent-bruno
  annotations:
    azure.workload.identity/client-id: "00000000-0000-0000-0000-000000000000"
```

```json
// Azure AD App Registration - Custom Role
{
  "Name": "Agent Bruno Security Auditor",
  "Description": "Read-only access for security audits",
  "Actions": [
    "Microsoft.Resources/subscriptions/resourceGroups/read",
    "Microsoft.Compute/virtualMachines/read",
    "Microsoft.Storage/storageAccounts/read",
    "Microsoft.Security/assessments/read"
  ],
  "NotActions": [
    "*/delete",
    "*/write"
  ],
  "DataActions": [],
  "NotDataActions": []
}
```

---

## Agent Identity & Authentication

### Agent Metadata in Requests

Every agent request includes identity metadata:

```python
# agent-bruno/core/identity.py
import os
from dataclasses import dataclass

@dataclass
class AgentIdentity:
    """Agent identity for RBAC and audit."""
    agent_id: str
    role: str  # sre, devops, security, analytics
    service_account: str
    namespace: str
    pod_name: str
    
    def to_headers(self) -> dict:
        """Convert to HTTP headers for MCP requests."""
        return {
            "X-Agent-ID": self.agent_id,
            "X-Agent-Role": self.role,
            "X-Agent-ServiceAccount": self.service_account,
            "X-Agent-Namespace": self.namespace,
            "X-Agent-Pod": self.pod_name,
        }

def get_agent_identity() -> AgentIdentity:
    """Extract agent identity from environment."""
    return AgentIdentity(
        agent_id=os.getenv("AGENT_ID", "unknown"),
        role=os.getenv("AGENT_ROLE", "unknown"),
        service_account=os.getenv("SERVICE_ACCOUNT", "default"),
        namespace=os.getenv("POD_NAMESPACE", "default"),
        pod_name=os.getenv("POD_NAME", "unknown"),
    )
```

### CloudEvents with Agent Identity

```python
# agent-bruno/core/events.py
from cloudevents.http import CloudEvent
from typing import Dict, Any

def create_agent_event(
    event_type: str,
    data: Dict[str, Any],
    identity: AgentIdentity
) -> CloudEvent:
    """Create CloudEvent with agent identity."""
    attributes = {
        "type": event_type,
        "source": f"agent-bruno/{identity.role}",
        # 🔑 Agent identity in extensions
        "agentid": identity.agent_id,
        "agentrole": identity.role,
        "serviceaccount": identity.service_account,
    }
    return CloudEvent(attributes, data)
```

### Audit Trail

```python
# agent-bruno/core/audit.py
import logging
from datetime import datetime
from typing import Optional

logger = logging.getLogger(__name__)

def audit_log(
    identity: AgentIdentity,
    action: str,
    resource: str,
    success: bool,
    details: Optional[str] = None
):
    """Structured audit logging."""
    log_entry = {
        "timestamp": datetime.utcnow().isoformat(),
        "agent_id": identity.agent_id,
        "agent_role": identity.role,
        "service_account": identity.service_account,
        "action": action,
        "resource": resource,
        "success": success,
        "details": details,
    }
    
    logger.info("AUDIT", extra=log_entry)
    
    # Send to centralized audit system (Loki, CloudWatch, etc.)
    # send_to_audit_system(log_entry)
```

---

## Implementation Examples

### Complete Agent Deployment (SRE Agent)

```yaml
# agent-bruno/deployments/agent-sre-complete.yaml
---
# 1️⃣ ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-bruno-sre
  namespace: agent-bruno
  labels:
    app: agent-bruno
    role: sre

---
# 2️⃣ ClusterRole (Read-Only)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agent-bruno-sre-viewer
rules:
- apiGroups: [""]
  resources: ["pods", "pods/log", "services", "events"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch"]
# NO delete, update, patch, create

---
# 3️⃣ ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: agent-bruno-sre-viewer-binding
subjects:
- kind: ServiceAccount
  name: agent-bruno-sre
  namespace: agent-bruno
roleRef:
  kind: ClusterRole
  name: agent-bruno-sre-viewer
  apiGroup: rbac.authorization.k8s.io

---
# 4️⃣ MCP Tokens Secret
apiVersion: v1
kind: Secret
metadata:
  name: mcp-tokens-sre
  namespace: agent-bruno
type: Opaque
stringData:
  github-token: "${GITHUB_READONLY_TOKEN}"
  grafana-token: "${GRAFANA_READONLY_TOKEN}"

---
# 5️⃣ Knative Service Deployment
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-bruno-sre
  namespace: agent-bruno
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "3"
    spec:
      serviceAccountName: agent-bruno-sre  # 🔑 RBAC link
      containers:
      - name: agent
        image: ghcr.io/your-org/agent-bruno:v1.0.0
        env:
        - name: AGENT_ID
          value: "agent-bruno-sre-001"
        - name: AGENT_ROLE
          value: "sre"
        - name: SERVICE_ACCOUNT
          value: "agent-bruno-sre"
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: mcp-tokens-sre
              key: github-token
        - name: GRAFANA_TOKEN
          valueFrom:
            secretKeyRef:
              name: mcp-tokens-sre
              key: grafana-token
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true

---
# 6️⃣ Network Policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: agent-bruno-sre-netpol
  namespace: agent-bruno
spec:
  podSelector:
    matchLabels:
      serving.knative.dev/service: agent-bruno-sre
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          knative-eventing-injection: enabled
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  - to:
    - podSelector:
        matchLabels:
          app: mcp-server
```

### MCP Server with RBAC Enforcement

```python
# mcp-servers/lancedb-mcp/main.py
from fastapi import FastAPI, Request, HTTPException, Header
from typing import Optional
import logging

app = FastAPI()
logger = logging.getLogger(__name__)

# Token -> Role -> Permissions
ROLE_PERMISSIONS = {
    "sre": {"read"},
    "devops": {"read", "write"},
    "security": {"read"},
    "analytics": {"read", "write"},
}

def check_permission(agent_role: str, required: str) -> bool:
    """Check if agent role has required permission."""
    permissions = ROLE_PERMISSIONS.get(agent_role, set())
    return required in permissions

def get_agent_identity(request: Request) -> dict:
    """Extract agent identity from headers."""
    return {
        "agent_id": request.headers.get("X-Agent-ID", "unknown"),
        "role": request.headers.get("X-Agent-Role", "unknown"),
        "service_account": request.headers.get("X-Agent-ServiceAccount", "unknown"),
    }

@app.get("/vectors/search")
async def search_vectors(
    query: str,
    request: Request,
    x_agent_role: Optional[str] = Header(None)
):
    """Search vectors - requires read permission."""
    identity = get_agent_identity(request)
    
    if not check_permission(x_agent_role or identity["role"], "read"):
        logger.warning(f"Access denied for {identity}")
        raise HTTPException(status_code=403, detail="Read permission required")
    
    logger.info(f"Vector search by {identity['agent_id']}")
    # ... search logic
    return {"results": []}

@app.post("/vectors/insert")
async def insert_vectors(
    data: dict,
    request: Request,
    x_agent_role: Optional[str] = Header(None)
):
    """Insert vectors - requires write permission."""
    identity = get_agent_identity(request)
    
    if not check_permission(x_agent_role or identity["role"], "write"):
        logger.warning(f"Write access denied for {identity}")
        raise HTTPException(status_code=403, detail="Write permission required")
    
    logger.info(f"Vector insertion by {identity['agent_id']}")
    # ... insertion logic
    return {"status": "inserted"}

@app.delete("/vectors/{vector_id}")
async def delete_vector(
    vector_id: str,
    request: Request,
    x_agent_role: Optional[str] = Header(None)
):
    """Delete vector - requires admin permission (none have this)."""
    identity = get_agent_identity(request)
    
    # 🚫 NO agent has delete permission
    logger.error(f"Deletion attempt denied for {identity}")
    raise HTTPException(
        status_code=403,
        detail="Deletion not allowed for any agent role"
    )
```

---

## Audit & Monitoring

### Kubernetes Audit Policy

```yaml
# kube-apiserver/audit-policy.yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
# 🔍 Log all agent actions at RequestResponse level
- level: RequestResponse
  users:
  - system:serviceaccount:agent-bruno:agent-bruno-sre
  - system:serviceaccount:agent-bruno:agent-bruno-devops
  - system:serviceaccount:agent-bruno:agent-bruno-security
  - system:serviceaccount:agent-bruno:agent-bruno-analytics
  
# 🚨 Alert on deletion attempts
- level: RequestResponse
  verbs: ["delete"]
  users:
  - system:serviceaccount:agent-bruno:*
  omitStages:
  - RequestReceived

# Log RBAC changes
- level: RequestResponse
  resources:
  - group: rbac.authorization.k8s.io
    resources: ["roles", "rolebindings", "clusterroles", "clusterrolebindings"]
```

### Prometheus Alerts

```yaml
# prometheus-rules/agent-rbac-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: agent-rbac-alerts
  namespace: agent-bruno
spec:
  groups:
  - name: agent-rbac
    interval: 30s
    rules:
    - alert: AgentUnauthorizedAccess
      expr: |
        sum(rate(http_requests_total{
          app="agent-bruno",
          status=~"403|401"
        }[5m])) > 0
      for: 1m
      labels:
        severity: warning
      annotations:
        summary: "Agent unauthorized access attempt detected"
        description: "Agent {{ $labels.agent_id }} attempted unauthorized action"
    
    - alert: AgentDeletionAttempt
      expr: |
        sum(rate(apiserver_audit_event_total{
          user=~"system:serviceaccount:agent-bruno:.*",
          verb="delete"
        }[5m])) > 0
      for: 0m  # Immediate alert
      labels:
        severity: critical
      annotations:
        summary: "🚨 Agent attempted to delete resources"
        description: "Agent {{ $labels.user }} attempted deletion"
    
    - alert: AgentEscalationAttempt
      expr: |
        sum(rate(apiserver_audit_event_total{
          user=~"system:serviceaccount:agent-bruno:.*",
          objectRef_resource=~"roles|clusterroles|rolebindings|clusterrolebindings"
        }[5m])) > 0
      for: 0m
      labels:
        severity: critical
      annotations:
        summary: "🚨 Agent attempted privilege escalation"
        description: "Agent {{ $labels.user }} attempted RBAC modification"
```

### Grafana Dashboard (Agent RBAC Monitoring)

```json
{
  "dashboard": {
    "title": "Agent RBAC & Access Control",
    "panels": [
      {
        "title": "Unauthorized Access Attempts",
        "targets": [{
          "expr": "sum(rate(http_requests_total{app=\"agent-bruno\", status=~\"403|401\"}[5m])) by (agent_id, agent_role)"
        }]
      },
      {
        "title": "Agent Actions by Permission Level",
        "targets": [{
          "expr": "sum(rate(agent_actions_total[5m])) by (agent_role, permission_required)"
        }]
      },
      {
        "title": "Cloud API Calls by Agent",
        "targets": [{
          "expr": "sum(rate(cloud_api_requests_total[5m])) by (agent_id, cloud_provider, action)"
        }]
      },
      {
        "title": "MCP Server Access Denials",
        "targets": [{
          "expr": "sum(rate(mcp_server_access_denied_total[5m])) by (server, agent_role, reason)"
        }]
      }
    ]
  }
}
```

### Loki Queries for Audit Logs

```logql
# All agent actions
{namespace="agent-bruno"} |= "AUDIT"

# Failed authorization
{namespace="agent-bruno"} 
  |= "AUDIT" 
  | json 
  | success="false"

# Deletion attempts (should be empty)
{namespace="agent-bruno"} 
  |= "AUDIT" 
  | json 
  | action=~"delete.*"

# Actions by specific agent
{namespace="agent-bruno"} 
  |= "AUDIT" 
  | json 
  | agent_id="agent-bruno-sre-001"
  
# Permission escalation attempts
{app="kube-apiserver"} 
  | json 
  | user=~"system:serviceaccount:agent-bruno:.*"
  | objectRef_resource=~"roles|clusterroles"
```

---

## Best Practices

### ✅ Do

1. **Use dedicated ServiceAccounts** for each agent role
2. **Apply least privilege** - grant minimum required permissions
3. **Rotate tokens regularly** - MCP API keys, cloud credentials
4. **Audit everything** - log all agent actions with identity
5. **Use Network Policies** - restrict agent network access
6. **Enable Pod Security Standards** - prevent privilege escalation
7. **Test RBAC policies** - verify agents cannot exceed permissions
8. **Monitor access patterns** - alert on anomalies

### ❌ Don't

1. **Never grant delete permissions** unless absolutely required (and audited)
2. **Don't share tokens** across different agent roles
3. **Don't hardcode credentials** - use Secrets and Workload Identity
4. **Don't grant cluster-admin** to any agent
5. **Don't allow privileged containers** for agent pods
6. **Don't skip audit logging** - always track actions
7. **Don't trust blindly** - verify every request, every time

### 🔐 Security Checklist

- [ ] Each agent has dedicated ServiceAccount
- [ ] RBAC policies follow least privilege
- [ ] No agent can delete critical resources
- [ ] MCP tokens are role-specific and rotated
- [ ] Cloud IAM uses Workload Identity (no static keys)
- [ ] Network Policies restrict agent communication
- [ ] Pod Security Standards enforced
- [ ] Audit logging enabled for all agent actions
- [ ] Prometheus alerts configured for unauthorized access
- [ ] Regular RBAC policy reviews scheduled
- [ ] Deletion attempts trigger immediate alerts
- [ ] Agent identity included in all events/logs

---

## References

### Kubernetes Security
- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Kubernetes Audit Logging](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/)

### Cloud IAM
- [AWS IAM Roles for Service Accounts (IRSA)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [GCP Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [Azure AD Workload Identity](https://azure.github.io/azure-workload-identity/)

### API Gateways
- [Ambassador Edge Stack Documentation](https://www.getambassador.io/docs/edge-stack/latest/)
- [Telepresence Quick Start](https://www.getambassador.io/docs/telepresence-oss/latest/quick-start)
- [Kong Documentation](https://docs.konghq.com/)
- [Envoy Proxy Documentation](https://www.envoyproxy.io/docs/envoy/latest/)

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-22  
**Maintained By**: SRE Team

---

## 📋 Document Review

**Review Completed By**: 
- ✅ **AI Senior Pentester (COMPLETE)** - October 22, 2025 - Security review complete, vulnerabilities documented
- [AI Senior SRE (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review (1/7 complete)  
**Next Review**: After security remediation Phase 1-2 complete

---

