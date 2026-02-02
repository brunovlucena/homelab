# Pi-hole DNS Server

Network-wide ad blocking via your own Linux hardware.

## Overview

Pi-hole is a DNS sinkhole that protects your devices from unwanted content without installing any client-side software.

- **Official Docs**: https://docs.pi-hole.net/
- **Docker Image**: https://github.com/pi-hole/docker-pi-hole/

## Services

| Service | Type | Ports | Description |
|---------|------|-------|-------------|
| `pihole-dns` | ClusterIP | 53/TCP, 53/UDP | Internal DNS for cluster workloads |
| `pihole-web` | NodePort | 80 → 30053 | Web admin interface |
| `pihole-dns-external` | NodePort | 53 → 30054/TCP, 30055/UDP | External DNS access |

## Accessing Pi-hole

### Web Interface

Port forward to access the admin UI:

```bash
kubectl port-forward -n pihole svc/pihole-web 8080:80
# Access at http://localhost:8080/admin
```

Default password: `change-me-in-production` (configured in `secret.yaml`)

### DNS Resolution

From within the cluster, use:
```
pihole-dns.pihole.svc.cluster.local
```

## Configuration

### Secret Management

The `secret.yaml` contains a placeholder password. Before deploying to production:

1. **Option 1**: Use Sealed Secrets
   ```bash
   kubectl create secret generic pihole-secret \
     --namespace pihole \
     --from-literal=FTLCONF_webserver_api_password='your-secure-password' \
     --dry-run=client -o yaml | kubeseal -o yaml > sealed-secret.yaml
   ```

2. **Option 2**: Use External Secrets Operator to pull from a secrets manager

### Environment Variables

Edit `configmap.yaml` to customize:

| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone | `America/Sao_Paulo` |
| `FTLCONF_dns_listeningMode` | DNS listening mode | `all` |

## Kind Cluster Integration

To expose Pi-hole ports on your host, add these to your Kind config's `extraPortMappings`:

```yaml
# Pi-hole ports
- containerPort: 30053  # web UI
  hostPort: 30053
  protocol: TCP
- containerPort: 30054  # DNS TCP
  hostPort: 30054
  protocol: TCP
- containerPort: 30055  # DNS UDP
  hostPort: 30055
  protocol: UDP
```

Then access the web UI at `http://localhost:30053/admin`

## Troubleshooting

### Check logs
```bash
kubectl logs -n pihole -l app=pihole -f
```

### Access the container
```bash
kubectl exec -it -n pihole deployment/pihole -- bash
```

### DNS resolution test
```bash
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup google.com pihole-dns.pihole.svc.cluster.local
```
