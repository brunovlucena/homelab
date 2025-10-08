# MCP Gateway Configuration Guide

This guide explains how to configure catalogs and persist changes in the Docker MCP Gateway running on Kubernetes.

## Overview

The MCP Gateway stores its configuration in the `~/.docker/mcp/` directory. In this Kubernetes deployment:

- **ConfigMap** (`configmap.yaml`): Provides initial configuration templates
- **PersistentVolume** (`pvc.yaml`): Stores runtime changes and persists them across pod restarts
- **InitContainer**: Copies initial config from ConfigMap to PersistentVolume on first run

## Configuration Files

Based on the [Docker MCP Gateway repository](https://github.com/docker/mcp-gateway), four main configuration files are used:

### 1. `docker-mcp.yaml` - Catalog Definition

Defines available MCP servers in your catalog:

```yaml
name: docker-mcp
description: Docker MCP Server Catalog
version: "1.0"
servers:
  - name: filesystem
    description: File system operations
    image: mcp/filesystem
    enabled: true
  - name: brave-search
    description: Web search via Brave
    image: mcp/brave-search
    enabled: false
  - name: postgres
    description: PostgreSQL database access
    image: mcp/postgres
    enabled: true
```

### 2. `registry.yaml` - Enabled Servers

Tracks which servers are currently enabled:

```yaml
enabled_servers:
  - filesystem
  - postgres
```

### 3. `config.yaml` - Server Configuration

Per-server configuration settings:

```yaml
servers:
  filesystem:
    root_path: "/data"
    allowed_operations:
      - read
      - write
      - list
  brave-search:
    api_key: "${BRAVE_API_KEY}"
    max_results: 10
    safe_search: true
  postgres:
    host: "postgres.default.svc.cluster.local"
    port: 5432
    database: "mydb"
    user: "mcpuser"
    password: "${POSTGRES_PASSWORD}"
```

### 4. `tools.yaml` - Enabled Tools

Tools available per server:

```yaml
tools:
  filesystem:
    - list_directory
    - read_file
    - write_file
    - delete_file
  brave-search:
    - search
    - suggest
  postgres:
    - query
    - execute
    - list_tables
```

## Updating Configuration

### Method 1: Update ConfigMap (Template Changes)

Edit `configmap.yaml` to change the initial configuration template:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/docker-mcp-gateway
# Edit configmap.yaml
kubectl apply -k .
```

**Note**: This only affects NEW deployments. Existing PVC data is preserved.

### Method 2: Direct File Edit in Running Pod

For runtime changes that persist:

```bash
# Get pod name
POD=$(kubectl get pod -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway -o jsonpath='{.items[0].metadata.name}')

# Edit configuration files directly
kubectl exec -n mcp-gateway $POD -- vi /home/mcp/.docker/mcp/docker-mcp.yaml
kubectl exec -n mcp-gateway $POD -- vi /home/mcp/.docker/mcp/config.yaml

# Or copy files in/out
kubectl cp my-local-config.yaml mcp-gateway/$POD:/home/mcp/.docker/mcp/config.yaml
kubectl cp mcp-gateway/$POD:/home/mcp/.docker/mcp/config.yaml ./downloaded-config.yaml

# Restart to apply changes
kubectl rollout restart deployment/mcp-gateway -n mcp-gateway
```

### Method 3: Using Docker MCP CLI (If Available)

If the gateway exposes CLI commands via API:

```bash
# Enable a server
curl -X POST http://localhost:31140/api/server/enable -d '{"server": "brave-search"}'

# Disable a server
curl -X POST http://localhost:31140/api/server/disable -d '{"server": "filesystem"}'

# Update configuration
curl -X POST http://localhost:31140/api/config/write -d @config.yaml
```

## Example Configurations

### Adding a Custom MCP Server

Edit `configmap.yaml`:

```yaml
data:
  docker-mcp.yaml: |
    name: docker-mcp
    description: Docker MCP Server Catalog
    version: "1.0"
    servers:
      - name: my-custom-server
        description: My custom MCP server
        image: ghcr.io/myorg/my-mcp-server:latest
        enabled: true
        env:
          - name: API_KEY
            value: "${MY_API_KEY}"
          - name: LOG_LEVEL
            value: "info"
```

### Configuring Server with Secrets

For sensitive data, use Kubernetes Secrets:

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mcp-secrets
  namespace: mcp-gateway
type: Opaque
stringData:
  brave-api-key: "your-brave-api-key"
  postgres-password: "your-postgres-password"
```

Then reference in deployment:

```yaml
env:
- name: BRAVE_API_KEY
  valueFrom:
    secretKeyRef:
      name: mcp-secrets
      key: brave-api-key
- name: POSTGRES_PASSWORD
  valueFrom:
    secretKeyRef:
      name: mcp-secrets
      key: postgres-password
```

### Connecting to Other Cluster Services

To connect MCP servers to other services in your cluster:

```yaml
servers:
  prometheus:
    url: "http://prometheus-operated.prometheus.svc.cluster.local:9090"
  loki:
    url: "http://loki.loki.svc.cluster.local:3100"
  grafana:
    url: "http://prometheus-operator-grafana.prometheus.svc.cluster.local:80"
```

## Backup and Restore

### Backup Configuration

```bash
# Create backup directory
mkdir -p backups/mcp-gateway/$(date +%Y%m%d)

# Get pod name
POD=$(kubectl get pod -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway -o jsonpath='{.items[0].metadata.name}')

# Backup all config files
kubectl cp mcp-gateway/$POD:/home/mcp/.docker/mcp/ backups/mcp-gateway/$(date +%Y%m%d)/
```

### Restore Configuration

```bash
# Delete existing PVC (warning: this deletes all data!)
kubectl delete pvc mcp-gateway-data -n mcp-gateway

# Recreate PVC
kubectl apply -f pvc.yaml

# Wait for new pod
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mcp-gateway -n mcp-gateway

# Get new pod name
POD=$(kubectl get pod -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway -o jsonpath='{.items[0].metadata.name}')

# Restore backup
kubectl cp backups/mcp-gateway/20241008/ mcp-gateway/$POD:/home/mcp/.docker/mcp/

# Restart
kubectl rollout restart deployment/mcp-gateway -n mcp-gateway
```

## Verification

Check configuration is loaded correctly:

```bash
# View current config
kubectl exec -n mcp-gateway deployment/mcp-gateway -- cat /home/mcp/.docker/mcp/docker-mcp.yaml

# Check logs
kubectl logs -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway -f

# Test API endpoint
curl http://localhost:31140/api/catalog/list
```

## Troubleshooting

### Configuration Not Persisting

Check PVC status:
```bash
kubectl get pvc -n mcp-gateway
kubectl describe pvc mcp-gateway-data -n mcp-gateway
```

### InitContainer Fails

Check init logs:
```bash
kubectl logs -n mcp-gateway deployment/mcp-gateway -c init-config
```

### Permission Issues

The gateway runs as user 1000. Ensure PV permissions allow this:
```bash
kubectl exec -n mcp-gateway deployment/mcp-gateway -- ls -la /home/mcp/.docker/mcp/
```

## References

- [Docker MCP Gateway Repository](https://github.com/docker/mcp-gateway)
- [Docker MCP Documentation](https://docs.docker.com/ai/mcp-gateway/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)

