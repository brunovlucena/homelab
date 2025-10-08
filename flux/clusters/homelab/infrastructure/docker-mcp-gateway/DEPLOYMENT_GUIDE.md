# Docker MCP Gateway - Deployment Guide

## Prerequisites

- Kind cluster running (named `homelab`)
- Flux installed and configured
- kubectl configured to connect to your cluster

## Deployment Steps

### 1. Verify Your Kind Cluster

```bash
kind get clusters
kubectl cluster-info
```

### 2. Deploy via Flux (Recommended)

Since this is integrated into your Flux setup, it will be automatically deployed when you commit and push:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab
git add flux/clusters/homelab/infrastructure/docker-mcp-gateway/
git commit -m "Add Docker MCP Gateway deployment"
git push
```

Flux will automatically reconcile and deploy the gateway.

### 3. Monitor Deployment

Watch Flux reconciliation:
```bash
flux get kustomizations -w
```

Check gateway deployment:
```bash
kubectl get all -n mcp-gateway
```

Watch pods come up:
```bash
watch kubectl get pods -n mcp-gateway
```

### 4. Verify Deployment

Check pod status:
```bash
kubectl get pods -n mcp-gateway
```

Check logs:
```bash
kubectl logs -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway
```

Test the health endpoint:
```bash
# Via NodePort (host access)
curl http://localhost:31140/health

# Via port-forward
kubectl port-forward -n mcp-gateway svc/mcp-gateway 8000:8000
curl http://localhost:8000/health
```

## Manual Deployment (Alternative)

If you want to test without Flux:

```bash
cd flux/clusters/homelab/infrastructure/docker-mcp-gateway
make apply
```

Or directly with kubectl:
```bash
kubectl apply -k flux/clusters/homelab/infrastructure/docker-mcp-gateway
```

## Configuration

### Environment Variables

You can add environment variables to the deployment by editing `deployment.yaml`:

```yaml
env:
- name: PORT
  value: "8000"
- name: LOG_LEVEL
  value: "info"
# Add more as needed
```

### Resource Limits

Current configuration:
- **Requests**: 256Mi memory, 100m CPU
- **Limits**: 1Gi memory, 500m CPU

Adjust in `deployment.yaml` if needed.

### ConfigMap (Optional)

If you need to add configuration files, create a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-gateway-config
  namespace: mcp-gateway
data:
  config.json: |
    {
      "servers": []
    }
```

## Connecting MCP Clients

### Visual Studio Code

Add to your `settings.json`:

```json
{
  "mcpGateway.url": "http://localhost:31140"
}
```

### Claude Desktop / Cursor

Configure in your MCP client settings:

```json
{
  "mcpServers": {
    "gateway": {
      "url": "http://localhost:31140"
    }
  }
}
```

## Troubleshooting

### Pod Not Starting

Check events:
```bash
kubectl describe pod -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway
```

Check logs:
```bash
kubectl logs -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway
```

### Port Already in Use

If port 31140 is already in use, change the hostPort in `kind.yaml`:

```yaml
extraPortMappings:
- containerPort: 30140
  hostPort: 31141  # Change this
  protocol: TCP
```

### Image Pull Issues

The docker/mcp-gateway image is public, but if you have rate limiting issues:

```bash
# Check image pull status
kubectl get events -n mcp-gateway
```

### Node Selector Issues

If pods are pending, check node labels:

```bash
kubectl get nodes --show-labels | grep mcp-servers
```

If the label is missing:
```bash
kubectl label node <node-name> role=mcp-servers
```

## Uninstall

### Via Make
```bash
cd flux/clusters/homelab/infrastructure/docker-mcp-gateway
make delete
```

### Via kubectl
```bash
kubectl delete -k flux/clusters/homelab/infrastructure/docker-mcp-gateway
```

### Via Flux
Remove from `infrastructure/kustomization.yaml`:
```yaml
# Remove this line:
# - docker-mcp-gateway
```

Then commit and push.

## Useful Commands

```bash
# Check status
make status

# View logs
make logs

# Port forward
make port-forward

# Test endpoints
make test
```

## Next Steps

1. Configure MCP servers to connect to the gateway
2. Set up monitoring with Prometheus/Grafana
3. Configure ingress for external access (optional)
4. Add authentication if needed
5. Configure persistent storage for gateway state

## Additional Resources

- [Docker MCP Gateway Docs](https://docs.docker.com/ai/mcp-gateway/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [GitHub Repository](https://github.com/docker/mcp-gateway)

