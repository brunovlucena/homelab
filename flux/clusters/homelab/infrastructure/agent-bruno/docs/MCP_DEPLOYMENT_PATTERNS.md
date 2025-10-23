# Agent MCP Server - Deployment Patterns

**[← Back to README](../README.md)** | **[MCP Workflows](MCP_WORKFLOWS.md)** | **[Architecture](ARCHITECTURE.md)** | **[Multi-Tenancy](MULTI-TENANCY.md)**

---

## Overview

This document outlines three deployment patterns for the Agent Bruno MCP Server, each with different security and isolation characteristics.

## Pattern 1: Local Access (Default, Recommended) 🔒

### Use Cases
- Development and testing
- Same-cluster service communication  
- CI/CD pipelines with cluster access
- Admin/operator access
- Single-user deployments

### Architecture
```
┌─────────────────────────────────────────────────────────┐
│  Developer/Service with kubectl access                  │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
         kubectl port-forward -n agent-bruno \
           svc/agent-mcp-server 8080:80
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Kubernetes Service (agent-mcp-server)                  │
│  - Type: ClusterIP                                      │
│  - No external exposure                                 │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Knative Service (agent-mcp-server)                     │
│  - Min: 0, Max: 5 replicas                              │
│  - Auto-scaling based on requests                       │
└─────────────────────────────────────────────────────────┘
```

### Setup

```bash
# 1. Deploy MCP server (ClusterIP only)
kubectl apply -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-mcp-server
  namespace: agent-bruno
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "0"
        autoscaling.knative.dev/max-scale: "5"
    spec:
      containers:
      - image: ghcr.io/brunolucena/agent-bruno-mcp:latest
        ports:
        - containerPort: 8080
        env:
        - name: MCP_AUTH_ENABLED
          value: "false"  # No auth needed - kubectl controls access
EOF

# 2. Access via port-forward
kubectl port-forward -n agent-bruno svc/agent-mcp-server 8080:80

# 3. Test MCP connection
curl http://localhost:8080/mcp/tools/list
```

### Security Features
| Feature | Status | Notes |
|---------|--------|-------|
| Internet exposure | ❌ None | ClusterIP only |
| Authentication | ⚠️ Kubernetes RBAC | kubectl access required |
| API key management | ✅ Not needed | Managed by k8s |
| Rate limiting | ⚠️ Optional | Can add if needed |
| TLS | ⚠️ Optional | kubectl handles tunnel |
| Audit logging | ✅ K8s audit logs | Via API server |

### Advantages
- ✅ **Zero attack surface**: No internet exposure
- ✅ **Simple**: No API key rotation or management
- ✅ **Secure by default**: Kubernetes RBAC enforces access
- ✅ **Cost-effective**: No ingress/load balancer costs
- ✅ **Fast**: Direct cluster access

### Disadvantages
- ⚠️ Requires kubectl access (not suitable for external clients)
- ⚠️ Manual port-forward setup per session
- ⚠️ Not suitable for multi-agent scenarios across clusters

### Best For
- 🎯 Personal deployments
- 🎯 Development environments
- 🎯 Testing and debugging
- 🎯 Same-cluster integrations

---

## Pattern 2: Remote Access via Cloudflare Tunnel (Optional) 🌐

### Use Cases
- External agent-to-agent communication
- Cross-cluster deployments
- Third-party integrations
- Multi-agent orchestration
- Trusted partner access

### Architecture
```
┌─────────────────────────────────────────────────────────┐
│  External MCP Client (Claude, other agents)             │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼ HTTPS (TLS 1.3)
┌─────────────────────────────────────────────────────────┐
│  Cloudflare Tunnel (mcp.bruno.dev)                      │
│  - WAF (Web Application Firewall)                       │
│  - DDoS Protection                                      │
│  - Rate Limiting (global)                               │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Kubernetes Ingress / Service                           │
│  - Type: LoadBalancer or ClusterIP with tunnel          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Knative Service (agent-mcp-server)                     │
│  - Authentication middleware (API keys)                 │
│  - Per-client rate limiting                             │
│  - Request validation                                   │
└─────────────────────────────────────────────────────────┘
```

### Setup

```bash
# 1. Create API key secret
kubectl create secret generic mcp-api-keys -n agent-bruno \
  --from-literal=client-a="$(uuidgen)" \
  --from-literal=client-b="$(uuidgen)"

# 2. Deploy MCP server with authentication
kubectl apply -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-mcp-server
  namespace: agent-bruno
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "1"
        autoscaling.knative.dev/max-scale: "10"
    spec:
      containers:
      - image: ghcr.io/brunolucena/agent-bruno-mcp:latest
        ports:
        - containerPort: 8080
        env:
        - name: MCP_AUTH_ENABLED
          value: "true"
        - name: MCP_API_KEYS_SECRET
          value: "mcp-api-keys"
        - name: RATE_LIMIT_PER_CLIENT
          value: "100"  # 100 requests per minute per client
        volumeMounts:
        - name: api-keys
          mountPath: /secrets/api-keys
          readOnly: true
      volumes:
      - name: api-keys
        secret:
          secretName: mcp-api-keys
EOF

# 3. Configure Cloudflare Tunnel
cloudflared tunnel route dns homelab-tunnel mcp.bruno.dev

# 4. Test remote access
curl -H "Authorization: Bearer <API_KEY>" \
  https://mcp.bruno.dev/mcp/tools/list
```

### Security Features
| Feature | Status | Notes |
|---------|--------|-------|
| Internet exposure | ✅ Controlled | Via Cloudflare only |
| Authentication | ✅ API Keys | Rotated monthly |
| API key management | ✅ Automated | K8s secrets + rotation |
| Rate limiting | ✅ Two-tier | Cloudflare + app-level |
| TLS | ✅ TLS 1.3 | Cloudflare-managed |
| Audit logging | ✅ Full | All requests logged |
| IP allowlisting | ⚠️ Optional | Can configure in CF |
| WAF | ✅ Cloudflare | OWASP rules enabled |

### API Key Rotation

```bash
# Automated monthly rotation via CronJob
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mcp-api-key-rotation
  namespace: agent-bruno
spec:
  schedule: "0 0 1 * *"  # Monthly on 1st
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: mcp-key-rotator
          containers:
          - name: rotate-keys
            image: bitnami/kubectl:latest
            command:
            - /bin/sh
            - -c
            - |
              # Generate new keys
              NEW_KEY_A=$(uuidgen)
              NEW_KEY_B=$(uuidgen)
              
              # Update secret
              kubectl create secret generic mcp-api-keys-new \
                --from-literal=client-a="$NEW_KEY_A" \
                --from-literal=client-b="$NEW_KEY_B" \
                --dry-run=client -o yaml | kubectl apply -f -
              
              # Notify clients (send new keys via secure channel)
              # Implementation depends on your notification system
              
              # Grace period: keep old keys for 24h
              sleep 86400
              
              # Remove old secret
              kubectl delete secret mcp-api-keys-old --ignore-not-found
              kubectl create secret generic mcp-api-keys-old \
                --from-literal=client-a="$(kubectl get secret mcp-api-keys -o jsonpath='{.data.client-a}' | base64 -d)" \
                --dry-run=client -o yaml | kubectl apply -f -
          restartPolicy: OnFailure
EOF
```

### Advantages
- ✅ **Accessible remotely**: Works from anywhere
- ✅ **Multi-agent support**: Multiple AI agents can connect
- ✅ **Cloudflare protection**: WAF, DDoS, rate limiting
- ✅ **Scalable**: Auto-scales based on demand
- ✅ **Audit trail**: Complete request logging

### Disadvantages
- ⚠️ **Complexity**: Requires API key management
- ⚠️ **Attack surface**: Internet-exposed (mitigated by Cloudflare)
- ⚠️ **Cost**: Cloudflare tunnel, potential data transfer costs
- ⚠️ **Maintenance**: Key rotation, monitoring

### Best For
- 🎯 Multi-agent orchestration
- 🎯 External integrations
- 🎯 Cross-organization collaboration
- 🎯 Production multi-tenant scenarios

---

## Pattern 3: Multi-Tenancy with Kamaji (Future) 🏢

### Use Cases
- SaaS deployment of Agent Bruno
- Enterprise multi-tenant scenarios
- Compliance requirements (data isolation)
- Independent agent instances per customer
- Strict resource and security boundaries

### Architecture
```
┌─────────────────────────────────────────────────────────────┐
│              Kamaji Management Cluster                      │
│  ┌────────────────────────────────────────────────────┐     │
│  │  Control Plane Tenants:                            │     │
│  │  - Tenant A: API Server + etcd (isolated)          │     │
│  │  - Tenant B: API Server + etcd (isolated)          │     │
│  │  - Tenant C: API Server + etcd (isolated)          │     │
│  └────────────────────────────────────────────────────┘     │
└──────────────────────┬──────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       │               │               │
       ▼               ▼               ▼
┌────────────┐  ┌────────────┐  ┌────────────┐
│ Tenant A   │  │ Tenant B   │  │ Tenant C   │
│ Workers    │  │ Workers    │  │ Workers    │
│            │  │            │  │            │
│ Agent      │  │ Agent      │  │ Agent      │
│ Bruno      │  │ Bruno      │  │ Bruno      │
│ + LanceDB  │  │ + LanceDB  │  │ + LanceDB  │
│            │  │            │  │            │
│ Network:   │  │ Network:   │  │ Network:   │
│ 10.0.1/24  │  │ 10.0.2/24  │  │ 10.0.3/24  │
│            │  │            │  │            │
│ Resources: │  │ Resources: │  │ Resources: │
│ CPU: 4c    │  │ CPU: 8c    │  │ CPU: 16c   │
│ Mem: 8Gi   │  │ Mem: 16Gi  │  │ Mem: 32Gi  │
└────────────┘  └────────────┘  └────────────┘
```

### Isolation Levels

| Layer | Isolation Type | Description |
|-------|---------------|-------------|
| Control Plane | ✅ Full | Dedicated K8s API server & etcd per tenant |
| Compute | ✅ Full | Dedicated worker nodes or reserved resources |
| Network | ✅ Full | Separate pod networks (CNI-level isolation) |
| Storage | ✅ Full | Dedicated PVs for LanceDB per tenant |
| Secrets | ✅ Full | Separate secret stores per tenant |
| RBAC | ✅ Full | Independent role bindings per tenant |
| Observability | ⚠️ Partial | Shared monitoring with tenant labels |

### Setup (Conceptual)

```bash
# 1. Install Kamaji on management cluster
helm install kamaji clastix/kamaji -n kamaji-system --create-namespace

# 2. Create tenant control plane
kubectl apply -f - <<EOF
apiVersion: kamaji.clastix.io/v1alpha1
kind: TenantControlPlane
metadata:
  name: tenant-a
  namespace: kamaji-system
spec:
  controlPlane:
    deployment:
      replicas: 2
      resources:
        apiServer:
          requests:
            cpu: 500m
            memory: 512Mi
        controllerManager:
          requests:
            cpu: 250m
            memory: 256Mi
  dataStore: postgres
  networkProfile:
    podCIDR: 10.0.1.0/24
    serviceCIDR: 10.96.1.0/24
EOF

# 3. Get tenant kubeconfig
kubectl get secret tenant-a-admin-kubeconfig -n kamaji-system \
  -o jsonpath='{.data.admin\.conf}' | base64 -d > tenant-a.kubeconfig

# 4. Deploy Agent Bruno to tenant cluster
kubectl --kubeconfig=tenant-a.kubeconfig apply -k \
  ./k8s/overlays/tenant-a

# 5. Expose MCP server (per-tenant ingress)
kubectl --kubeconfig=tenant-a.kubeconfig apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agent-mcp-ingress
  namespace: agent-bruno
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - tenant-a.mcp.bruno.dev
    secretName: tenant-a-tls
  rules:
  - host: tenant-a.mcp.bruno.dev
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: agent-mcp-server
            port:
              number: 80
EOF
```

### Resource Allocation

```yaml
# Tenant resource quotas (enforced at control plane level)
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-a-quota
spec:
  hard:
    requests.cpu: "16"
    requests.memory: "32Gi"
    requests.storage: "100Gi"
    persistentvolumeclaims: "10"
    services.loadbalancers: "2"
    count/deployments.apps: "20"
```

### Advantages
- ✅ **Complete isolation**: No shared control plane components
- ✅ **Independent operations**: Each tenant can upgrade independently  
- ✅ **Compliance-friendly**: Data residency, security boundaries
- ✅ **Fault isolation**: Tenant failures don't affect others
- ✅ **Resource guarantees**: Dedicated CPU/memory per tenant
- ✅ **Custom policies**: Per-tenant NetworkPolicies, RBAC, etc.

### Disadvantages
- ⚠️ **High overhead**: Multiple control planes consume resources
- ⚠️ **Complexity**: More complex to manage and troubleshoot
- ⚠️ **Cost**: Higher infrastructure costs (control plane per tenant)
- ⚠️ **Operational burden**: More clusters to maintain

### Cost Considerations

| Component | Single Cluster | Kamaji Multi-Tenant |
|-----------|---------------|---------------------|
| Control Plane | 1x (shared) | N x (dedicated) |
| Worker Nodes | Shared pool | Dedicated or reserved |
| Storage | Shared PVs | Isolated PVs |
| Network | Single CNI | Per-tenant networks |
| **Total Overhead** | **~5%** | **~30-40%** |

### Best For
- 🎯 SaaS platforms offering Agent Bruno as a service
- 🎯 Enterprise deployments with strict compliance
- 🎯 Multi-organization scenarios with trust boundaries
- 🎯 When data residency/sovereignty is required
- 🎯 High-value customers requiring dedicated resources

---

## Comparison Matrix

| Feature | Local | Remote | Kamaji |
|---------|-------|--------|--------|
| **Setup Complexity** | ⭐ Simple | ⭐⭐ Moderate | ⭐⭐⭐⭐ Complex |
| **Security** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent |
| **Isolation** | ⚠️ None | ⚠️ Application-level | ✅ Full (control plane) |
| **Cost** | ⭐⭐⭐⭐⭐ Minimal | ⭐⭐⭐ Moderate | ⭐ High |
| **Scalability** | ⚠️ Limited | ✅ High | ✅ Very High |
| **Multi-tenant** | ❌ No | ⚠️ Shared cluster | ✅ Full isolation |
| **External Access** | ❌ No | ✅ Yes | ✅ Yes (per-tenant) |
| **Operations** | ⭐⭐⭐⭐⭐ Easy | ⭐⭐⭐ Moderate | ⭐⭐ Complex |
| **Attack Surface** | ✅ Zero | ⚠️ Internet-exposed | ✅ Isolated |
| **Use Case** | Dev/Testing | Multi-agent | SaaS/Enterprise |

---

## Decision Tree

```
Start: Need MCP Server Access?
│
├─ Is this for development/testing?
│  └─ ✅ Use Pattern 1: Local (kubectl port-forward)
│
├─ Do you need remote access?
│  │
│  ├─ Is it for a few trusted agents/services?
│  │  └─ ✅ Use Pattern 2: Remote (Cloudflare Tunnel)
│  │
│  └─ Do you need complete tenant isolation?
│     │
│     ├─ Is this a SaaS or enterprise deployment?
│     │  └─ ✅ Use Pattern 3: Kamaji Multi-Tenancy
│     │
│     └─ Is cost a concern?
│        └─ ⚠️  Consider Pattern 2 with namespace-level isolation
│
└─ Default recommendation: Pattern 1 (Local)
```

---

## Migration Path

### From Pattern 1 → Pattern 2
```bash
# 1. Add authentication to MCP server
kubectl set env deployment/agent-mcp-server \
  -n agent-bruno MCP_AUTH_ENABLED=true

# 2. Create API keys
kubectl create secret generic mcp-api-keys -n agent-bruno \
  --from-literal=client-a="$(uuidgen)"

# 3. Configure Cloudflare Tunnel
# (See Pattern 2 setup)

# 4. Test remote access
curl -H "Authorization: Bearer <API_KEY>" \
  https://mcp.bruno.dev/mcp/tools/list

# 5. Update clients to use new endpoint
```

### From Pattern 2 → Pattern 3
```bash
# 1. Install Kamaji
helm install kamaji clastix/kamaji -n kamaji-system

# 2. Create tenant control planes
# (See Pattern 3 setup)

# 3. Migrate workloads tenant by tenant
for tenant in tenant-a tenant-b tenant-c; do
  kubectl --kubeconfig=${tenant}.kubeconfig apply -k ./k8s/overlays/${tenant}
done

# 4. Update DNS to point to tenant-specific endpoints
# tenant-a.mcp.bruno.dev
# tenant-b.mcp.bruno.dev

# 5. Verify isolation and remove old deployment
```

---

## Monitoring & Observability

### Pattern 1: Local
```promql
# Simple cluster-local metrics
rate(mcp_requests_total[5m])
histogram_quantile(0.95, mcp_request_duration_seconds_bucket)
```

### Pattern 2: Remote
```promql
# Per-client metrics
rate(mcp_requests_total{client_id="client-a"}[5m])

# Auth failures
rate(mcp_auth_failures_total[5m])

# Rate limit hits
rate(mcp_rate_limit_exceeded_total[5m]) by (client_id)
```

### Pattern 3: Kamaji
```promql
# Per-tenant metrics with tenant label
rate(mcp_requests_total{tenant="tenant-a"}[5m])

# Tenant resource usage
sum(container_memory_usage_bytes{tenant="tenant-a"})

# Cross-tenant isolation verification
count(mcp_cross_tenant_access_attempts_total) # Should be 0
```

---

## Security Best Practices

### All Patterns
- ✅ Enable audit logging
- ✅ Implement request size limits
- ✅ Use Pydantic for input validation
- ✅ Filter PII from logs
- ✅ Regular security scans (Trivy)
- ✅ Keep dependencies updated

### Pattern 2 (Remote) Specific
- ✅ Rotate API keys monthly (automated)
- ✅ Enable Cloudflare WAF rules
- ✅ Configure rate limiting (global + per-client)
- ✅ Use TLS 1.3 only
- ✅ Implement request signing (optional)
- ✅ IP allowlisting for known clients (optional)

### Pattern 3 (Kamaji) Specific
- ✅ Network policies between tenants (deny-all default)
- ✅ Separate secrets stores per tenant
- ✅ Resource quotas enforcement
- ✅ Regular control plane backups
- ✅ Tenant-specific RBAC auditing

---

## Troubleshooting

### Pattern 1: Local
```bash
# Check port-forward
kubectl port-forward -n agent-bruno svc/agent-mcp-server 8080:80
# Test: curl http://localhost:8080/mcp/tools/list

# Check service
kubectl get svc -n agent-bruno agent-mcp-server

# Check pods
kubectl get pods -n agent-bruno -l serving.knative.dev/service=agent-mcp-server
```

### Pattern 2: Remote
```bash
# Check Cloudflare tunnel
cloudflared tunnel info homelab-tunnel

# Check authentication
curl -v -H "Authorization: Bearer <API_KEY>" \
  https://mcp.bruno.dev/mcp/tools/list
# Look for 401 (bad key) vs 200 (success)

# Check rate limiting
for i in {1..150}; do
  curl -H "Authorization: Bearer <API_KEY>" \
    https://mcp.bruno.dev/mcp/tools/list
done
# Should see 429 after quota exceeded
```

### Pattern 3: Kamaji
```bash
# Check tenant control plane health
kubectl get tenantcontrolplane -n kamaji-system

# Check tenant cluster access
kubectl --kubeconfig=tenant-a.kubeconfig get nodes

# Verify network isolation
kubectl --kubeconfig=tenant-a.kubeconfig run test-pod \
  --image=nicolaka/netshoot -it -- /bin/bash
# Try to access tenant-b services (should fail)
```

---

## Conclusion

**Default Recommendation**: Start with **Pattern 1 (Local)** for maximum security and simplicity.

**When to Upgrade**:
- Need remote access? → **Pattern 2 (Remote)**
- Need full multi-tenancy? → **Pattern 3 (Kamaji)**

**Remember**: Security and simplicity often align. Only increase complexity when the use case clearly demands it.

---

**Last Updated**: October 22, 2025  
**Next Review**: January 22, 2026  
**Owner**: SRE/Platform Team

---

## 📋 Document Review

**Review Completed By**: 
- [AI Senior SRE (Pending)]
- [AI Senior Pentester (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review  
**Next Review**: TBD

---

