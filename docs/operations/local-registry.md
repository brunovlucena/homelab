# 📦 Local Container Registry Guide

> **KISS: ONE REGISTRY FOR ALL CLUSTERS**  
> **Works with**: Air, Pro, Studio (Kind) + Forge, Pi (k3s)  
> **Registry Name**: `local-registry` (customizable via `REGISTRY_NAME`)  
> **Port**: `localhost:5001` (customizable via `REGISTRY_PORT`)

---

## Quick Start (KISS!)

### For Kind Clusters (Air, Pro, Studio)

```bash
# ONE command to rule them all (defaults to studio)
make registry

# Or specify a cluster
make registry ENV=pro
make registry ENV=air

# Check status
make registry-info
```

### For k3s Clusters (Forge, Pi)

```bash
# Setup script (provides manual configuration steps)
./scripts/mac/setup-local-registry.sh forge k3s
./scripts/mac/setup-local-registry.sh pi k3s

# Follow the on-screen instructions to configure /etc/rancher/k3s/registries.yaml
```

---

## Usage

Once the registry is set up, use it the same way across all clusters:

```bash
# 1. Tag your image
docker tag myapp:latest localhost:5001/myapp:test

# 2. Push to local registry
docker push localhost:5001/myapp:test

# 3. Use in Kubernetes
kubectl run myapp --image=localhost:5001/myapp:test

# 4. List images in registry
curl http://localhost:5001/v2/_catalog
```

---

## Architecture

### Single Registry for All Clusters

The same registry container (`local-registry`) works for **all clusters**:

```
┌─────────────────────────────────────────────────────────────────┐
│                  Local Container Registry                        │
│                    (local-registry:5000)                         │
│                                                                  │
│  Host Port: localhost:5001                                       │
│  Internal: local-registry:5000                                   │
└──────────┬──────────────────────────────────────────────────────┘
           │
           ├──────────┬──────────┬──────────┬──────────┬──────────┐
           │          │          │          │          │          │
      ┌────▼───┐ ┌────▼───┐ ┌────▼────┐ ┌──▼──┐   ┌──▼──┐
      │  Air   │ │  Pro   │ │ Studio  │ │Forge│   │ Pi  │
      │ (Kind) │ │ (Kind) │ │ (Kind)  │ │(k3s)│   │(k3s)│
      └────────┘ └────────┘ └─────────┘ └─────┘   └─────┘
```

### How It Works

**For Kind Clusters:**
- Registry connected to `kind` Docker network
- Each node configured with `/etc/containerd/certs.d/localhost:5001/hosts.toml`
- ConfigMap in `kube-public` namespace documents the registry

**For k3s Clusters:**
- Registry accessible on host network (`localhost:5001`)
- Configured via `/etc/rancher/k3s/registries.yaml`
- Requires k3s restart after configuration

---

## Customization

### Change Registry Name

```bash
# Use a different name
REGISTRY_NAME=my-registry make registry ENV=studio
REGISTRY_NAME=my-registry ./scripts/mac/setup-local-registry.sh forge k3s
```

### Change Registry Port

```bash
# Use a different port
REGISTRY_PORT=5002 make registry ENV=pro
REGISTRY_PORT=5002 ./scripts/mac/setup-local-registry.sh pro kind
```

### Combined

```bash
# Custom name AND port
REGISTRY_NAME=test-registry REGISTRY_PORT=5002 make registry ENV=air
```

---

## Management Commands

### View Registry Status

```bash
make registry-info
```

**Output:**
```
✅ Registry Status: Running
   Container:       local-registry
   Host Port:       localhost:5001
   Internal Port:   local-registry:5000
   Works for:       Air, Pro, Studio, Forge, Pi (all clusters!)

📊 Registry Catalog:
   "repositories":["myapp","nginx"]

📝 Usage:
   docker tag myapp:latest localhost:5001/myapp:test
   docker push localhost:5001/myapp:test
   kubectl run test --image=localhost:5001/myapp:test
```

### Clean Up Registry

```bash
# Remove registry container
make registry-clean

# Or with custom name
REGISTRY_NAME=my-registry make registry-clean
```

---

## k3s Configuration

For k3s clusters (Forge, Pi), you need to manually configure each node.

### Configuration File: `/etc/rancher/k3s/registries.yaml`

```yaml
mirrors:
  "localhost:5001":
    endpoint:
      - "http://localhost:5001"

configs:
  "localhost:5001":
    tls:
      insecure_skip_verify: true
```

### Apply Configuration

```bash
# On control-plane nodes
sudo systemctl restart k3s

# On worker nodes
sudo systemctl restart k3s-agent
```

### Verify Configuration

```bash
# Check if registry is configured
sudo cat /etc/rancher/k3s/registries.yaml

# Test pulling an image
sudo k3s crictl pull localhost:5001/nginx:latest
```

---

## Troubleshooting

### Registry Not Accessible

```bash
# Check if container is running
docker ps | grep local-registry

# Check registry health
curl http://localhost:5001/v2/

# Expected output: {}
```

### Kind Cluster Can't Pull Images

```bash
# Verify registry is connected to kind network
docker network inspect kind | grep local-registry

# Check node configuration
docker exec <cluster>-control-plane cat /etc/containerd/certs.d/localhost:5001/hosts.toml

# Reconnect if needed
docker network connect kind local-registry
```

### k3s Cluster Can't Pull Images

```bash
# Verify registries.yaml exists
sudo cat /etc/rancher/k3s/registries.yaml

# Check k3s logs
sudo journalctl -u k3s -f

# Restart k3s
sudo systemctl restart k3s
```

### Can't Push to Registry

```bash
# Make sure you're using the host port (5001), not internal port (5000)
docker push localhost:5001/myapp:test  # ✅ Correct

# Not this:
docker push local-registry:5000/myapp:test  # ❌ Wrong (internal only)
```

---

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Build and Test

on: [push]

jobs:
  build:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v3
      
      - name: Build image
        run: docker build -t localhost:5001/myapp:${{ github.sha }} .
      
      - name: Push to local registry
        run: docker push localhost:5001/myapp:${{ github.sha }}
      
      - name: Deploy to test cluster
        run: |
          kubectl config use-context kind-air
          kubectl run test-${{ github.sha }} \
            --image=localhost:5001/myapp:${{ github.sha }} \
            --rm -it --restart=Never
```

---

## Benefits

✅ **One Registry for All Clusters**: No need to push to multiple registries  
✅ **Fast**: Local network speeds, no external pulls  
✅ **Cost-Effective**: No cloud registry fees for testing  
✅ **Flexible**: Customize name and port as needed  
✅ **Compatible**: Works with both Kind and k3s  
✅ **CI/CD Ready**: Perfect for self-hosted runners

---

## References

- [Kind Local Registry Documentation](https://kind.sigs.k8s.io/docs/user/local-registry/)
- [k3s Private Registry Configuration](https://docs.k3s.io/installation/private-registry)
- [Script Source](/scripts/mac/setup-local-registry.sh)
- [Makefile Commands](/Makefile)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

