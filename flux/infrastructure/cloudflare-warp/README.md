# Cloudflare WARP

Cloudflare WARP connector for Kubernetes cluster networking through Cloudflare Zero Trust.

## 🎯 Overview

This deployment runs Cloudflare WARP as a DaemonSet to provide secure connectivity through Cloudflare's network for all nodes in the cluster.

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│         Kubernetes Cluster               │
│                                          │
│  ┌────────────────────────────────┐     │
│  │  DaemonSet (on each node)      │     │
│  │  ├── Register with connector   │     │
│  │  ├── Connect to WARP           │     │
│  │  └── Monitor status            │     │
│  └────────────────────────────────┘     │
│                                          │
└─────────────────────────────────────────┘
                    │
                    │ WARP Tunnel
                    ▼
         ┌──────────────────────┐
         │  Cloudflare Network   │
         │  Zero Trust Gateway   │
         └──────────────────────┘
```

## 🚀 Deployment

### Prerequisites

1. **Cloudflare WARP Connector Token**:
   - Get token from: `warp-cli connector new`

2. **Store Token in GitHub Secrets**:
   - All secrets in this homelab are managed via **External Secrets Operator** (ESO)
   - Add `CLOUDFLARE_WARP_TOKEN` to your GitHub repository secrets
   - ESO will automatically pull it from GitHub and create the secret in the `cloudflare-warp` namespace
   - See `external-secret.yaml` for the ExternalSecret configuration

### Deploy via Flux

The service is automatically deployed via Flux in the 01-foundation layer.

Reconcile:
```bash
flux reconcile kustomization studio-01-foundation --with-source
```

## 🔍 Monitoring

### Check DaemonSet Status

```bash
# View DaemonSet
kubectl get daemonset -n cloudflare-warp

# View pods on each node
kubectl get pods -n cloudflare-warp -o wide

# View logs
kubectl logs -l app=cloudflare-warp -n cloudflare-warp --tail=100
```

### Check WARP Connection Status

```bash
# Get status from any pod
kubectl exec -it -n cloudflare-warp $(kubectl get pod -n cloudflare-warp -l app=cloudflare-warp -o jsonpath='{.items[0].metadata.name}') -- warp-cli status
```

## 🔧 Configuration

### Environment Variables

- `CLOUDFLARE_WARP_TOKEN`: Connector token from Cloudflare Zero Trust dashboard

## 🚨 Troubleshooting

### WARP Not Connecting

Check pod logs:
```bash
kubectl logs -l app=cloudflare-warp -n cloudflare-warp
```

Common issues:
- **"Command not found: warp-cli"**: Wrong container image
- **"Invalid token"**: Check secret has correct token
- **"Permission denied"**: Check securityContext allows privileged mode

### DaemonSet Not Scheduling

Check DaemonSet status:
```bash
kubectl describe daemonset cloudflare-warp -n cloudflare-warp
```

Common issues:
- Node taints preventing scheduling
- Insufficient node resources
- Image pull errors

## 🔐 Security

### Required Permissions

WARP requires privileged mode because it:
- Modifies network routing tables
- Creates network interfaces
- Manages DNS settings

### Capabilities

- `NET_ADMIN`: Modify network configuration
- `NET_RAW`: Use raw sockets
- `SYS_ADMIN`: Mount filesystems, modify namespaces

## 📝 Notes

- Runs as DaemonSet to ensure WARP on every node
- Uses hostNetwork for direct network access
- Persists data in /var/lib/cloudflare-warp on host
- Status check runs every 5 minutes

## 🔗 Related Documentation

- [Cloudflare WARP Connector Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-devices/warp/deployment/mdm-deployment/)
- [Cloudflare Zero Trust](https://developers.cloudflare.com/cloudflare-one/)

