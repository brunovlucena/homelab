# Docker MCP Gateway

This deployment runs the Docker MCP Gateway on Kubernetes, allowing you to manage and interact with MCP (Model Context Protocol) servers.

## Overview

The Docker MCP Gateway provides a unified interface for managing multiple MCP servers and connecting them to various clients like VS Code, Claude Desktop, or Cursor.

## Architecture

- **Namespace**: `mcp-gateway`
- **Image**: `docker/mcp-gateway:latest`
- **Port**: 8000
- **NodePort**: 30140 (mapped to hostPort 31140 on kind)

## Access

### From within the cluster:
```bash
curl http://mcp-gateway.mcp-gateway.svc.cluster.local:8000/health
```

### From your host machine:
```bash
curl http://localhost:31140/health
```

## Configuration

The gateway is deployed with:
- 1 replica
- 256Mi memory request, 1Gi limit
- 100m CPU request, 500m limit
- Deployed on nodes with `role: mcp-servers` label
- Persistent storage (1Gi PVC) for configuration
- ConfigMap with initial catalog templates
- InitContainer to initialize configuration on first run

### Configuration Files

The gateway uses four main configuration files stored in `~/.docker/mcp/`:
- **`docker-mcp.yaml`** - Server catalog definition
- **`registry.yaml`** - Enabled servers tracking
- **`config.yaml`** - Per-server configuration
- **`tools.yaml`** - Enabled tools per server

See [CONFIGURATION.md](./CONFIGURATION.md) for detailed configuration guide.

### Managing Configuration

Use the provided management script:

```bash
# Backup configuration
./manage-config.sh backup

# View configuration files
./manage-config.sh view docker-mcp.yaml

# Edit configuration
./manage-config.sh edit config.yaml

# Export all configs
./manage-config.sh export

# Import configs
./manage-config.sh import ./my-configs/

# Restart after changes
./manage-config.sh restart
```

Or use Makefile shortcuts:
```bash
make config-backup
make config-export
```

## Deployment

The service is automatically deployed via Flux when changes are pushed to the repository.

To manually verify:
```bash
kubectl get pods -n mcp-gateway
kubectl get svc -n mcp-gateway
kubectl logs -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway
```

## Port Forwarding (Alternative Access)

If you prefer not to use NodePort:
```bash
kubectl port-forward -n mcp-gateway svc/mcp-gateway 8000:8000
```

Then access at `http://localhost:8000`

## Connecting MCP Clients

Configure your MCP client to connect to:
- **Cluster internal**: `http://mcp-gateway.mcp-gateway.svc.cluster.local:8000`
- **Host access**: `http://localhost:31140`

## Documentation

- [Docker MCP Gateway Documentation](https://docs.docker.com/ai/mcp-gateway/)
- [MCP Gateway GitHub Repository](https://github.com/docker/mcp-gateway)
- [MCP Catalog and Toolkit](https://docs.docker.com/ai/mcp-catalog-and-toolkit/)

## Troubleshooting

Check gateway logs:
```bash
kubectl logs -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway -f
```

Verify service endpoints:
```bash
kubectl get endpoints -n mcp-gateway
```

Check pod status:
```bash
kubectl describe pod -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway
```

