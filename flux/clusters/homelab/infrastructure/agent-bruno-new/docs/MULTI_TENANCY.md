# 🏢 Agent Bruno - Multi-Tenancy Architecture

**[← Back to README](../README.md)** | **[RBAC](RBAC.md)** | **[Rate Limiting](RATELIMITING.md)** | **[MCP Deployment](MCP_DEPLOYMENT_PATTERNS.md)**

---

## Overview

This document describes the multi-tenancy strategy for Agent Bruno, enabling multiple isolated agent instances for different users, teams, or organizations. The architecture leverages **Kamaji** for Kubernetes control plane isolation, ensuring complete tenant separation at the infrastructure level while maintaining cost efficiency and operational simplicity.

---

## 📋 Table of Contents

1. [Multi-Tenancy Overview](#multi-tenancy-overview)
2. [Kamaji Architecture](#kamaji-architecture)
3. [Isolation Levels](#isolation-levels)
4. [Deployment Patterns](#deployment-patterns)
5. [Data Isolation](#data-isolation)
6. [Security & Compliance](#security--compliance)
7. [Resource Management](#resource-management)
8. [Observability](#observability)
9. [Migration Strategy](#migration-strategy)
10. [Cost Analysis](#cost-analysis)

---

## Multi-Tenancy Overview

### Why Multi-Tenancy?

Agent Bruno's current architecture is designed for **single-tenant** deployment (personal use). Multi-tenancy enables:

- **SaaS Offering**: Deploy Agent Bruno as a service for multiple customers
- **Enterprise Deployment**: Isolated instances for different departments/teams
- **Development/Testing**: Separate environments for dev, staging, production
- **Compliance**: Meet regulatory requirements for data isolation (GDPR, HIPAA, SOC2)
- **Cost Efficiency**: Share infrastructure while maintaining isolation

### Multi-Tenancy Levels

| Level | Isolation | Complexity | Cost | Use Case |
|-------|-----------|------------|------|----------|
| **Soft Multi-Tenancy** | Namespace-level | Low | Low | Internal teams, dev/staging |
| **Hard Multi-Tenancy (Kamaji)** | Control plane-level | High | Medium | External customers, compliance |
| **Cluster-per-Tenant** | Full cluster | Very High | High | Enterprise customers |

**This document focuses on Hard Multi-Tenancy with Kamaji** as it provides the optimal balance of isolation, cost, and operational complexity for a SaaS-like deployment.

---

## Kamaji Architecture

### What is Kamaji?

[Kamaji](https://github.com/clastix/kamaji) is a **Kubernetes control plane manager** that enables running multiple isolated Kubernetes control planes as lightweight tenants on a shared infrastructure. Each tenant gets:

- Dedicated API server
- Dedicated etcd datastore
- Dedicated controller manager
- Dedicated scheduler
- **Complete isolation** from other tenants

### Kamaji Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Kamaji Management Cluster                              │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  Control Plane Components (Shared Infrastructure)                  │     │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │     │
│  │  │ Kamaji       │  │ PostgreSQL/  │  │ etcd Cluster │             │     │
│  │  │ Controller   │  │ etcd Backend │  │ (Shared)     │             │     │
│  │  └──────────────┘  └──────────────┘  └──────────────┘             │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  Tenant Control Planes (Isolated)                                  │     │
│  │                                                                    │     │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐        │     │
│  │  │ Tenant A Control Plane  │  │ Tenant B Control Plane  │        │     │
│  │  │ ┌─────────────────────┐ │  │ ┌─────────────────────┐ │        │     │
│  │  │ │ API Server          │ │  │ │ API Server          │ │        │     │
│  │  │ │ (dedicated)         │ │  │ │ (dedicated)         │ │        │     │
│  │  │ └─────────────────────┘ │  │ └─────────────────────┘ │        │     │
│  │  │ ┌─────────────────────┐ │  │ ┌─────────────────────┐ │        │     │
│  │  │ │ etcd Datastore      │ │  │ │ etcd Datastore      │ │        │     │
│  │  │ │ (isolated prefix)   │ │  │ │ (isolated prefix)   │ │        │     │
│  │  │ └─────────────────────┘ │  │ └─────────────────────┘ │        │     │
│  │  │ ┌─────────────────────┐ │  │ ┌─────────────────────┐ │        │     │
│  │  │ │ Controller Manager  │ │  │ │ Controller Manager  │ │        │     │
│  │  │ └─────────────────────┘ │  │ └─────────────────────┘ │        │     │
│  │  └─────────────────────────┘  └─────────────────────────┘        │     │
│  └────────────────────────────────────────────────────────────────────┘     │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
         ▼                       ▼                       ▼
┌──────────────────┐  ┌───────────────────┐  ┌──────────────────────┐
│ Tenant A Workers │  │ Tenant B Workers  │  │ Tenant C Workers     │
│                  │  │                   │  │                      │
│ ┌──────────────┐ │  │ ┌──────────────┐  │  │ ┌──────────────────┐ │
│ │ Agent Bruno  │ │  │ │ Agent Bruno  │  │  │ │ Agent Bruno      │ │
│ │ (Tenant A)   │ │  │ │ (Tenant B)   │  │  │ │ (Tenant C)       │ │
│ │              │ │  │ │              │  │  │ │                  │ │
│ │ + LanceDB    │ │  │ │ + LanceDB    │  │  │ │ + LanceDB        │ │
│ │ + Memory     │ │  │ │ + Memory     │  │  │ │ + Memory         │ │
│ └──────────────┘ │  │ └──────────────┘  │  │ └──────────────────┘ │
│                  │  │                   │  │                      │
│ Network:         │  │ Network:          │  │ Network:             │
│ 10.100.1.0/24    │  │ 10.100.2.0/24     │  │ 10.100.3.0/24        │
└──────────────────┘  └───────────────────┘  └──────────────────────┘
```

### Kamaji Benefits for Agent Bruno

1. **Complete Isolation**: Each tenant has a separate Kubernetes control plane
   - No shared API server
   - No shared etcd
   - No namespace collision risks

2. **Resource Efficiency**: Lightweight control planes
   - ~200MB RAM per tenant control plane
   - Fast provisioning (<2 minutes)
   - Scales to hundreds of tenants

3. **Independent Lifecycle**: Per-tenant operations
   - Separate Kubernetes version
   - Independent upgrades
   - Isolated failures

4. **Network Isolation**: CNI-level separation
   - Separate pod networks per tenant
   - NetworkPolicies per tenant
   - No cross-tenant traffic

5. **Security**: Enhanced tenant isolation
   - Dedicated RBAC per tenant
   - Separate service accounts
   - Isolated secrets

---

## Isolation Levels

### 1. Control Plane Isolation

Each tenant gets a dedicated Kubernetes control plane managed by Kamaji:

```yaml
# TenantControlPlane Custom Resource
apiVersion: kamaji.clastix.io/v1alpha1
kind: TenantControlPlane
metadata:
  name: tenant-bruno-enterprise-acme
  namespace: kamaji-system
spec:
  # Control plane configuration
  controlPlane:
    deployment:
      replicas: 2  # HA control plane
      resources:
        apiServer:
          requests:
            cpu: 250m
            memory: 512Mi
          limits:
            cpu: 500m
            memory: 1Gi
    
    service:
      type: LoadBalancer
      serviceType: LoadBalancer
    
    ingress:
      enabled: true
      hostname: acme-k8s.agent-bruno.com
      className: nginx
  
  # etcd datastore (isolated)
  dataStore:
    type: PostgreSQL
    # OR: type: etcd (dedicated etcd cluster)
    postgresql:
      host: postgresql.kamaji-system
      port: 5432
      dbName: tenant_acme_cp
      secretRef:
        name: postgres-credentials
        namespace: kamaji-system
  
  # Kubernetes version
  kubernetes:
    version: v1.28.0
    admissionControllers:
    - ResourceQuota
    - LimitRanger
    - MutatingAdmissionWebhook
    - ValidatingAdmissionWebhook
  
  # Network configuration
  network:
    serviceClusterIPRange: 10.100.1.0/24
    podClusterIPRange: 10.200.1.0/24
  
  # Add-ons
  addons:
    coreDNS:
      enabled: true
    kubeProxy:
      enabled: true
```

### 2. Data Plane Isolation

Worker nodes run tenant workloads with isolation:

**Option A: Dedicated Worker Nodes** (Highest Isolation)
```yaml
# Node affinity for tenant-specific workers
apiVersion: v1
kind: Node
metadata:
  name: worker-acme-01
  labels:
    tenant: acme
    node.kubernetes.io/tenant: acme
spec:
  taints:
  - key: tenant
    value: acme
    effect: NoSchedule
```

**Option B: Shared Workers with Namespace Isolation** (Cost Efficient)
```yaml
# Namespace for Tenant A
apiVersion: v1
kind: Namespace
metadata:
  name: agent-bruno-acme
  labels:
    tenant: acme
    app.kubernetes.io/name: agent-bruno
    app.kubernetes.io/tenant: acme
```

### 3. Storage Isolation

Each tenant has isolated storage:

```yaml
# Tenant-specific StorageClass
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: lancedb-storage-acme
  labels:
    tenant: acme
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: local
  tenant: acme
---
# Tenant-specific PersistentVolume
apiVersion: v1
kind: PersistentVolume
metadata:
  name: lancedb-pv-acme
  labels:
    tenant: acme
spec:
  capacity:
    storage: 100Gi
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: lancedb-storage-acme
  hostPath:
    path: /data/tenants/acme/lancedb
    type: DirectoryOrCreate
```

### 4. Network Isolation

```yaml
# NetworkPolicy: Deny all cross-tenant traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-cross-tenant
  namespace: agent-bruno-acme
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  
  ingress:
  # Allow only from same tenant
  - from:
    - namespaceSelector:
        matchLabels:
          tenant: acme
  
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  
  # Allow Ollama (external)
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 11434
  
  # Allow only to same tenant
  - to:
    - namespaceSelector:
        matchLabels:
          tenant: acme
```

### 5. Resource Quota Isolation

```yaml
# ResourceQuota per tenant
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-acme-quota
  namespace: agent-bruno-acme
spec:
  hard:
    # Compute
    requests.cpu: "8"
    requests.memory: 16Gi
    limits.cpu: "16"
    limits.memory: 32Gi
    
    # Storage
    requests.storage: 100Gi
    persistentvolumeclaims: "5"
    
    # Objects
    pods: "50"
    services: "20"
    configmaps: "50"
    secrets: "50"
---
# LimitRange: Default resource limits
apiVersion: v1
kind: LimitRange
metadata:
  name: tenant-acme-limits
  namespace: agent-bruno-acme
spec:
  limits:
  - max:
      cpu: "4"
      memory: 8Gi
    min:
      cpu: 100m
      memory: 128Mi
    default:
      cpu: 500m
      memory: 1Gi
    defaultRequest:
      cpu: 250m
      memory: 512Mi
    type: Container
```

---

## Deployment Patterns

### Pattern 1: Full Isolation (Recommended for SaaS)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Kamaji Management Cluster                           │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  Shared Infrastructure:                                            │     │
│  │  - Kamaji Controller                                               │     │
│  │  - PostgreSQL (tenant control plane data)                          │     │
│  │  - Monitoring Stack (Prometheus, Grafana, Loki, Tempo)             │     │
│  └────────────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
        ┌─────────────────────────────┼─────────────────────────────┐
        │                             │                             │
        ▼                             ▼                             ▼
┌──────────────────┐        ┌──────────────────┐        ┌──────────────────┐
│ Tenant: ACME     │        │ Tenant: FooBar   │        │ Tenant: Initech  │
│ (Enterprise)     │        │ (Startup)        │        │ (Enterprise)     │
├──────────────────┤        ├──────────────────┤        ├──────────────────┤
│ - Dedicated CP   │        │ - Dedicated CP   │        │ - Dedicated CP   │
│ - Dedicated etcd │        │ - Dedicated etcd │        │ - Dedicated etcd │
│ - 3 Worker Nodes │        │ - 1 Worker Node  │        │ - 5 Worker Nodes │
│ - 500GB Storage  │        │ - 100GB Storage  │        │ - 1TB Storage    │
│                  │        │                  │        │                  │
│ Agent Bruno:     │        │ Agent Bruno:     │        │ Agent Bruno:     │
│ - API Server     │        │ - API Server     │        │ - API Server     │
│ - MCP Server     │        │ - MCP Server     │        │ - MCP Server     │
│ - LanceDB        │        │ - LanceDB        │        │ - LanceDB        │
│ - Memory System  │        │ - Memory System  │        │ - Memory System  │
│ - Observability  │        │ - Observability  │        │ - Observability  │
└──────────────────┘        └──────────────────┘        └──────────────────┘
```

**Isolation Guarantees**:
- ✅ Separate Kubernetes control planes
- ✅ Separate etcd datastores
- ✅ Separate worker nodes (optional)
- ✅ Network isolation (CNI-level)
- ✅ Storage isolation (dedicated PVs)
- ✅ Independent RBAC and secrets

### Pattern 2: Namespace-Level Multi-Tenancy (Cost Optimized)

For less strict isolation requirements (internal teams):

```yaml
# Single cluster, namespace-based isolation
apiVersion: v1
kind: Namespace
metadata:
  name: agent-bruno-team-a
  labels:
    team: team-a
    environment: production
---
apiVersion: v1
kind: Namespace
metadata:
  name: agent-bruno-team-b
  labels:
    team: team-b
    environment: production
---
# ResourceQuota per namespace
# NetworkPolicies for isolation
# Separate LanceDB PVs per namespace
```

**Trade-offs**:
- ⚠️ Shared Kubernetes control plane (lower isolation)
- ✅ Lower resource overhead
- ⚠️ Namespace-level RBAC (less granular)
- ✅ Faster provisioning
- ⚠️ Higher blast radius (cluster-wide issues affect all tenants)

### Pattern 3: Hybrid Multi-Tenancy

Combine Kamaji for external customers with namespace isolation for internal teams:

```
┌─────────────────────────────────────────────────────────────────────┐
│ Management Cluster (Kamaji + Internal Shared Cluster)               │
│                                                                     │
│ ┌────────────────────────┐   ┌─────────────────────────────────┐   │
│ │ External Customers     │   │ Internal Teams                  │   │
│ │ (Kamaji Control Planes)│   │ (Namespace Isolation)           │   │
│ │                        │   │                                 │   │
│ │ - Tenant: ACME Corp    │   │ - Namespace: team-ml            │   │
│ │ - Tenant: StartupXYZ   │   │ - Namespace: team-sre           │   │
│ └────────────────────────┘   └─────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Data Isolation

### 1. LanceDB Isolation

Each tenant gets a dedicated LanceDB instance:

```python
# agent_bruno/storage/multi_tenant_lancedb.py
from typing import Dict
import lancedb

class MultiTenantVectorStore:
    """Multi-tenant LanceDB wrapper with isolation."""
    
    def __init__(self, base_path: str = "/data/tenants"):
        self.base_path = base_path
        self.connections: Dict[str, lancedb.DBConnection] = {}
    
    def get_tenant_db(self, tenant_id: str) -> lancedb.DBConnection:
        """Get or create tenant-specific LanceDB connection."""
        if tenant_id not in self.connections:
            tenant_path = f"{self.base_path}/{tenant_id}/lancedb"
            
            # Ensure directory exists with proper permissions
            os.makedirs(tenant_path, mode=0o700, exist_ok=True)
            
            # Create isolated connection
            self.connections[tenant_id] = lancedb.connect(tenant_path)
        
        return self.connections[tenant_id]
    
    def add_chunks(self, tenant_id: str, table_name: str, chunks: List[Dict]):
        """Add chunks to tenant-specific table."""
        db = self.get_tenant_db(tenant_id)
        
        # Create or get table
        if table_name not in db.table_names():
            table = db.create_table(table_name, data=chunks)
        else:
            table = db.open_table(table_name)
            table.add(chunks)
    
    def search(self, tenant_id: str, table_name: str, query_vector, limit: int = 10):
        """Search in tenant-specific table."""
        db = self.get_tenant_db(tenant_id)
        table = db.open_table(table_name)
        
        return table.search(query_vector).limit(limit).to_list()
```

### 2. Memory System Isolation

Tenant-specific memory tables:

```python
# agent_bruno/memory/multi_tenant_memory.py
class MultiTenantMemorySystem:
    """Multi-tenant memory system with isolation."""
    
    def __init__(self, vector_store: MultiTenantVectorStore):
        self.vector_store = vector_store
    
    def store_episodic_memory(self, tenant_id: str, user_id: str, turn: ConversationTurn):
        """Store episodic memory for a tenant's user."""
        # Table name includes tenant_id for isolation
        table_name = f"episodic_memory_{tenant_id}"
        
        # Add tenant_id to metadata for additional safety
        turn_data = {
            **turn.to_dict(),
            "tenant_id": tenant_id,
            "user_id": user_id,
        }
        
        self.vector_store.add_chunks(tenant_id, table_name, [turn_data])
    
    def retrieve_user_context(self, tenant_id: str, user_id: str, limit: int = 5):
        """Retrieve user context within tenant boundary."""
        table_name = f"episodic_memory_{tenant_id}"
        
        # Query with tenant_id and user_id filters
        results = self.vector_store.query(
            tenant_id=tenant_id,
            table_name=table_name,
            filters=f"tenant_id = '{tenant_id}' AND user_id = '{user_id}'",
            limit=limit
        )
        
        return results
```

### 3. Secrets Isolation

```yaml
# Per-tenant secrets in separate namespaces
apiVersion: v1
kind: Secret
metadata:
  name: agent-bruno-secrets
  namespace: agent-bruno-acme
type: Opaque
stringData:
  # Tenant-specific API keys
  ollama-api-key: "tenant-acme-ollama-key"
  
  # Tenant-specific external MCP credentials
  github-mcp-token: "ghp_tenant_acme_token"
  grafana-api-key: "tenant-acme-grafana-key"
  
  # Tenant-specific observability tokens
  logfire-token: "tenant-acme-logfire-token"
  wandb-api-key: "tenant-acme-wandb-key"
```

---

## Security & Compliance

### 1. Authentication & Authorization

```python
# agent_bruno/auth/multi_tenant_auth.py
from fastapi import HTTPException, Security
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

security = HTTPBearer()

async def get_tenant_from_token(credentials: HTTPAuthorizationCredentials = Security(security)):
    """Extract tenant_id from JWT token."""
    try:
        # Decode JWT
        payload = jwt.decode(
            credentials.credentials,
            settings.JWT_SECRET,
            algorithms=["HS256"]
        )
        
        tenant_id = payload.get("tenant_id")
        user_id = payload.get("user_id")
        
        if not tenant_id:
            raise HTTPException(status_code=403, detail="No tenant_id in token")
        
        return {
            "tenant_id": tenant_id,
            "user_id": user_id,
            "permissions": payload.get("permissions", [])
        }
        
    except jwt.InvalidTokenError:
        raise HTTPException(status_code=401, detail="Invalid token")

# Usage in API endpoints
@app.post("/api/v1/query")
async def query(
    request: QueryRequest,
    tenant_context = Depends(get_tenant_from_token)
):
    """Process query within tenant context."""
    tenant_id = tenant_context["tenant_id"]
    user_id = tenant_context["user_id"]
    
    # All operations scoped to tenant_id
    response = await agent.query(
        query=request.query,
        tenant_id=tenant_id,
        user_id=user_id
    )
    
    return response
```

### 2. Audit Logging

```python
# agent_bruno/audit/tenant_audit.py
class TenantAuditLogger:
    """Audit logging with tenant context."""
    
    def log_event(
        self,
        tenant_id: str,
        user_id: str,
        event_type: str,
        resource_type: str,
        resource_id: str,
        action: str,
        result: str,
        metadata: Dict = None
    ):
        """Log tenant-scoped audit event."""
        audit_log.info(
            f"Tenant audit event: {event_type}",
            extra={
                "tenant_id": tenant_id,
                "user_id": user_id,
                "event_type": event_type,
                "resource_type": resource_type,
                "resource_id": resource_id,
                "action": action,
                "result": result,
                "metadata": metadata or {},
                "timestamp": datetime.utcnow().isoformat(),
                "compliance": {
                    "gdpr": True,
                    "soc2": True,
                    "hipaa": False,  # Configure per tenant
                }
            }
        )

# Usage
audit = TenantAuditLogger()

audit.log_event(
    tenant_id="acme",
    user_id="user_123",
    event_type="data_access",
    resource_type="conversation_history",
    resource_id="conv_abc123",
    action="read",
    result="success",
    metadata={"ip_address": request.client.host}
)
```

### 3. Data Retention & GDPR Compliance

```python
# agent_bruno/compliance/data_retention.py
class TenantDataRetention:
    """Manage tenant-specific data retention policies."""
    
    async def apply_retention_policy(self, tenant_id: str):
        """Apply retention policy for a tenant."""
        # Get tenant's retention configuration
        retention_config = await self.get_tenant_retention_config(tenant_id)
        
        # Episodic memory retention
        if retention_config.episodic_retention_days:
            await self.purge_old_episodic_memory(
                tenant_id=tenant_id,
                days=retention_config.episodic_retention_days
            )
        
        # Logs retention
        if retention_config.logs_retention_days:
            await self.purge_old_logs(
                tenant_id=tenant_id,
                days=retention_config.logs_retention_days
            )
    
    async def delete_tenant_data(self, tenant_id: str):
        """Delete all data for a tenant (GDPR right to erasure)."""
        # Delete from LanceDB
        await self.delete_lancedb_data(tenant_id)
        
        # Delete from logs
        await self.delete_logs(tenant_id)
        
        # Delete from traces
        await self.delete_traces(tenant_id)
        
        # Delete from metrics (anonymize)
        await self.anonymize_metrics(tenant_id)
        
        # Delete backups
        await self.delete_backups(tenant_id)
        
        # Audit log deletion
        audit.log_event(
            tenant_id=tenant_id,
            user_id="system",
            event_type="data_deletion",
            resource_type="tenant",
            resource_id=tenant_id,
            action="delete_all",
            result="success"
        )
```

---

## Resource Management

### 1. Tenant Provisioning

```yaml
# flux/clusters/homelab/tenants/acme/kustomization.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: tenant-acme
  namespace: flux-system
spec:
  interval: 5m
  path: ./tenants/acme
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
  
  # Health checks
  healthChecks:
    - apiVersion: kamaji.clastix.io/v1alpha1
      kind: TenantControlPlane
      name: tenant-acme
      namespace: kamaji-system
  
  # Post-deployment configuration
  postBuild:
    substitute:
      TENANT_ID: "acme"
      TENANT_NAME: "ACME Corporation"
      RESOURCE_QUOTA_CPU: "8"
      RESOURCE_QUOTA_MEMORY: "16Gi"
      RESOURCE_QUOTA_STORAGE: "100Gi"
```

### 2. Auto-Scaling Configuration

```yaml
# Per-tenant HorizontalPodAutoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: agent-bruno-api-hpa
  namespace: agent-bruno-acme
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: agent-bruno-api
  
  minReplicas: 2
  maxReplicas: 10
  
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
  
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

### 3. Cost Tracking

```python
# agent_bruno/billing/cost_tracking.py
class TenantCostTracker:
    """Track resource usage and costs per tenant."""
    
    def track_resource_usage(self, tenant_id: str):
        """Track and record resource usage."""
        # Compute costs
        cpu_hours = self.get_cpu_usage_hours(tenant_id)
        memory_gb_hours = self.get_memory_usage_gb_hours(tenant_id)
        storage_gb = self.get_storage_usage_gb(tenant_id)
        
        # LLM costs
        llm_tokens = self.get_llm_token_usage(tenant_id)
        llm_cost = self.calculate_llm_cost(llm_tokens)
        
        # Total cost
        total_cost = (
            cpu_hours * COST_PER_CPU_HOUR +
            memory_gb_hours * COST_PER_GB_HOUR +
            storage_gb * COST_PER_GB_MONTH +
            llm_cost
        )
        
        # Record in database
        await self.record_usage(
            tenant_id=tenant_id,
            period=datetime.utcnow().strftime("%Y-%m"),
            cpu_hours=cpu_hours,
            memory_gb_hours=memory_gb_hours,
            storage_gb=storage_gb,
            llm_tokens=llm_tokens,
            llm_cost_usd=llm_cost,
            total_cost_usd=total_cost
        )
        
        # Emit metrics
        tenant_cost_dollars.labels(tenant_id=tenant_id).set(total_cost)
```

---

## Observability

### Multi-Tenant Observability Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Shared Observability Stack                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ Prometheus   │  │ Grafana Loki │  │ Grafana Tempo│             │
│  │ (Metrics)    │  │ (Logs)       │  │ (Traces)     │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
│                                                                     │
│  All data tagged with tenant_id label                              │
│  PromQL/LogQL/TraceQL queries filtered by tenant_id                │
└─────────────────────────────────────────────────────────────────────┘
```

### Tenant-Scoped Queries

**Metrics (PromQL)**:
```promql
# Tenant-specific request rate
rate(http_requests_total{tenant_id="acme"}[5m])

# Tenant-specific error rate
sum(rate(http_requests_total{tenant_id="acme", status_code=~"5.."}[5m]))
  /
sum(rate(http_requests_total{tenant_id="acme"}[5m]))

# Tenant-specific resource usage
sum(container_cpu_usage_seconds_total{namespace="agent-bruno-acme"})
```

**Logs (LogQL)**:
```logql
# Tenant-specific logs
{namespace="agent-bruno-acme"} 
  | json 
  | tenant_id="acme"
  | level="ERROR"

# Tenant-specific error patterns
{namespace="agent-bruno-acme"} 
  |= "ERROR" 
  | json 
  | line_format "{{.timestamp}} {{.message}}"
```

**Traces (TraceQL)**:
```traceql
# Tenant-specific traces
{
  resource.tenant_id = "acme"
  && span.service.name = "agent-bruno-api"
  && duration > 2s
}
```

### Tenant Dashboards

```yaml
# Grafana dashboard for tenant overview
apiVersion: v1
kind: ConfigMap
metadata:
  name: tenant-dashboard-template
  namespace: monitoring
data:
  dashboard.json: |
    {
      "dashboard": {
        "title": "Tenant: ${tenant_name}",
        "templating": {
          "list": [
            {
              "name": "tenant_id",
              "type": "constant",
              "current": {
                "value": "${tenant_id}"
              }
            }
          ]
        },
        "panels": [
          {
            "title": "Request Rate",
            "targets": [
              {
                "expr": "rate(http_requests_total{tenant_id=\"$tenant_id\"}[5m])"
              }
            ]
          },
          {
            "title": "Error Rate",
            "targets": [
              {
                "expr": "sum(rate(http_requests_total{tenant_id=\"$tenant_id\", status_code=~\"5..\"}[5m])) / sum(rate(http_requests_total{tenant_id=\"$tenant_id\"}[5m]))"
              }
            ]
          },
          {
            "title": "Resource Usage",
            "targets": [
              {
                "expr": "sum(container_cpu_usage_seconds_total{namespace=~\"agent-bruno-$tenant_id.*\"})"
              }
            ]
          }
        ]
      }
    }
```

---

## Migration Strategy

### Phase 1: Single-Tenant (Current)

```
┌────────────────────────────────────────┐
│ Current: Single Kubernetes Cluster     │
│                                        │
│ - One namespace: agent-bruno           │
│ - One LanceDB instance                 │
│ - No tenant isolation                  │
└────────────────────────────────────────┘
```

### Phase 2: Soft Multi-Tenancy (Namespace Isolation)

```
┌────────────────────────────────────────┐
│ Step 1: Add tenant_id everywhere       │
│                                        │
│ - Add tenant_id to all data models     │
│ - Add tenant_id to all API requests    │
│ - Add tenant_id to all metrics/logs    │
└────────────────────────────────────────┘
        ↓
┌────────────────────────────────────────┐
│ Step 2: Namespace-per-tenant           │
│                                        │
│ - Namespace: agent-bruno-tenant-a      │
│ - Namespace: agent-bruno-tenant-b      │
│ - Separate ResourceQuotas              │
│ - NetworkPolicies for isolation        │
└────────────────────────────────────────┘
```

### Phase 3: Hard Multi-Tenancy (Kamaji)

```
┌────────────────────────────────────────┐
│ Step 1: Deploy Kamaji                  │
│                                        │
│ - Install Kamaji in management cluster │
│ - Configure PostgreSQL/etcd backend    │
│ - Test with pilot tenant               │
└────────────────────────────────────────┘
        ↓
┌────────────────────────────────────────┐
│ Step 2: Migrate Tenants to Kamaji      │
│                                        │
│ - Create TenantControlPlane CRs        │
│ - Deploy Agent Bruno to tenant CPs     │
│ - Migrate data (LanceDB, configs)      │
│ - Cutover DNS/ingress                  │
└────────────────────────────────────────┘
        ↓
┌────────────────────────────────────────┐
│ Step 3: Full Multi-Tenant Platform     │
│                                        │
│ - Self-service tenant provisioning     │
│ - Automated billing & metering         │
│ - Tenant lifecycle management          │
└────────────────────────────────────────┘
```

---

## Cost Analysis

### Infrastructure Costs (Per Tenant)

| Component | Resource Usage | Monthly Cost | Notes |
|-----------|---------------|--------------|-------|
| **Kamaji Control Plane** | 250m CPU, 512Mi RAM | $5 | Lightweight control plane |
| **Worker Nodes (Shared)** | 2 CPU, 4Gi RAM | $20 | Pro-rated across tenants |
| **Storage (LanceDB)** | 50Gi | $5 | Block storage |
| **Observability** | Shared | $2 | Pro-rated |
| **Networking** | Shared | $1 | Pro-rated |
| **LLM Inference (Ollama)** | Shared | Variable | Based on token usage |
| **Total (Base)** | | **~$33/tenant/month** | Excluding LLM costs |

### Scaling Economics

| Tier | Tenants | Cost Per Tenant | Total Monthly Cost |
|------|---------|----------------|-------------------|
| **Startup** | 1-10 | $50 | $500 |
| **Growth** | 11-50 | $40 | $2,000 |
| **Scale** | 51-100 | $35 | $3,500 |
| **Enterprise** | 100+ | $33 | $3,300+ |

**Economies of Scale**:
- Shared Ollama inference reduces per-tenant LLM costs
- Shared observability stack (Prometheus, Loki, Tempo)
- Shared management cluster overhead
- Pro-rated networking and storage costs

---

## 📚 References

- **Kamaji Documentation**: https://github.com/clastix/kamaji
- **Kubernetes Multi-Tenancy**: https://kubernetes.io/docs/concepts/security/multi-tenancy/
- **Agent Bruno Architecture**: [ARCHITECTURE.md](./ARCHITECTURE.md)
- **Agent Bruno Observability**: [OBSERVABILITY.md](./OBSERVABILITY.md)
- **Agent Bruno Testing**: [TESTING.md](./TESTING.md)
- **Agent Bruno Memory**: [MEMORY.md](./MEMORY.md)
- **Agent Bruno RAG**: [RAG.md](./RAG.md)

---

**Last Updated**: October 22, 2025  
**Next Review**: January 22, 2026  
**Owner**: Platform Engineering Team

---

## 📋 Document Review

**Review Completed By**: 
- ✅ **AI Senior Pentester (COMPLETE)** - October 22, 2025 - Security review complete, noted premature for current scale, security must be fixed first
- [AI Senior SRE (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review (1/7 complete)  
**Next Review**: After single-tenant security is complete

---


