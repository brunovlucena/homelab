# Pulumi Infrastructure as Code

> **Purpose**: Dynamic cluster provisioning for homelab infrastructure  
> **Language**: Go  
> **Managed Clusters**: All clusters in `flux/clusters/` directory

---

## 🎯 Overview

This Pulumi program **dynamically discovers and reconciles** Kubernetes clusters based on the directory structure in `flux/clusters/`. No hardcoded cluster configurations!

### How It Works

1. **Auto-Discovery**: Reads `flux/clusters/<stack>/` structure
2. **Kind Substrate**: Use `make bootstrap ENV=<cluster>` once to create/prepare the Kind cluster
3. **Flux Bootstrap**: Pulumi deploys Flux bootstrap manifests directly (no local scripts)
4. **Secrets Management**: Pulumi creates Git auth secret, Flux deploys Sealed/External Secrets
5. **GitOps Handoff**: Pulumi applies GitRepository/Kustomizations so Flux manages the rest

### 🔒 Multi-Machine/Client Safety

**IMPORTANT**: This system is designed for multi-machine and multi-client deployments with built-in safety protections:

- **Machine-Specific Stacks**: Each machine/client gets its own Pulumi stack (format: `org/env-machine`)
  - Example: `bvlucena/studio-studio` (studio environment on studio machine)
  - Example: `bvlucena/studio-pro` (studio environment on pro machine)
  - Example: `bvlucena/studio-client1` (studio environment on client1 machine)

- **Destruction Protection**: The `destroy` command includes multiple safety checks:
  1. ✅ Verifies the Kind cluster exists **locally** before allowing destruction
  2. ✅ Prevents accidental deletion of clusters on other machines
  3. ✅ Shows machine hostname and ID in all operations
  4. ✅ Requires explicit confirmation with machine name

- **Why This Matters**: Without these protections, running `make destroy-studio` from Machine B would delete the Pulumi state for Machine A's cluster, even though the actual cluster on Machine A would still be running (orphaned).

---

## 🚀 Usage

### Bootstrap + Deploy a Cluster

```bash
# 1) Prepare Kind substrate (only needed once per cluster/after destroy)
make bootstrap ENV=studio

# 2) Deploy Pulumi-managed infrastructure
make up ENV=studio

# Shortcuts / other clusters
make bootstrap ENV=pro && make up ENV=pro
make bootstrap ENV=<name> && make up ENV=<name>
```

### Destroy a Cluster

```bash
make destroy ENV=studio    # Destroy studio cluster (with safety checks)
make destroy ENV=<name>    # Destroy any cluster
```

**⚠️ Safety Features**:
- The destroy command will **only work if the cluster exists locally**
- This prevents accidentally destroying clusters on other machines
- You must run `destroy` from the machine where the cluster is actually running
- The command shows your machine hostname and requires confirmation

### Check Status

```bash
# View current environment and machine info
make env

# List all Pulumi stacks (across all machines/clients)
make list-stacks

# View Pulumi outputs
cd pulumi && pulumi stack output

# Or use the Makefile
make observe    # Watch Flux resources
```

---

## 📁 Adding a New Cluster

To add a new cluster, simply create:

```
flux/clusters/
  └── mynewcluster/
      └── kind.yaml         # Kind cluster configuration
      └── kustomization.yaml # Flux resources
      └── infrastructure/   # Infrastructure components
      └── apps/            # Application deployments
```

Then deploy:

```bash
make up ENV=mynewcluster
```

**That's it!** No code changes needed. The Pulumi program automatically discovers and provisions it.

---

## 🔧 Architecture

```
Pulumi (main.go)
  ↓
1. Discover cluster config from flux/clusters/<stack>/
  ↓
2. Create Kind cluster using kind.yaml
  ↓
3. Bootstrap Flux GitOps
  ↓
4. Create GitHub token secret
  ↓
5. Create Flux GitRepository + root Kustomization
  ↓
6. Handoff to Flux (GitOps takes over)
  ↓
7. Flux deploys Sealed Secrets + all infrastructure
```

---

## 🔐 Twingate Connector Management

Pulumi automatically manages Twingate connectors for each cluster (studio, pro, air). This replaces the unreliable Docker container approach with declarative infrastructure.

### Setup

1. **Get Twingate API Token**:
   - Go to [Twingate Admin Console](https://console.twingate.com)
   - Navigate to Settings → API
   - Generate a new API token (different from connector tokens)
   - Copy the token

2. **Configure Pulumi**:
   ```bash
   cd pulumi
   pulumi stack select <stack-name>  # e.g., studio-studio
   pulumi config set twingate:apiToken <YOUR_API_TOKEN> --secret
   pulumi config set twingate:network bvlucena  # Optional, defaults to bvlucena
   ```

3. **Deploy**:
   ```bash
   make up ENV=studio  # Creates connector for studio
   make up ENV=pro     # Creates connector for pro
   make up ENV=air     # Creates connector for air
   ```

### How It Works

- **Single Remote Network**: All clusters share the "home" remote network
- **Per-Cluster Connectors**: Each cluster (studio, pro, air) gets its own connector
- **Automatic Creation**: Connectors are created automatically when you run `pulumi up`
- **State Management**: Connector state is tracked in Pulumi, making it reliable and reproducible

### Connector Installation

After Pulumi creates the connector resource in Twingate, you need to install the connector software on each machine:

**For each machine (studio, pro, air)**:

```bash
# Install Twingate connector
curl -s https://binaries.twingate.com/connector/setup.sh | sudo bash

# The connector will use the tokens from Pulumi-managed resources
# Check connector status
sudo twingate-connector status
```

**Note**: The connector tokens (access/refresh) are different from the API token. The API token is used by Pulumi to manage resources, while connector tokens are used to run the connector software.

---

## 🔐 Secret Management with Sealed Secrets

Secrets are managed via **Sealed Secrets** (Bitnami Sealed Secrets Controller):

1. **Deployment**: Flux deploys Sealed Secrets Controller via GitOps (01-core phase)
2. **Encryption**: Secrets are encrypted client-side using `kubeseal` CLI
3. **GitOps**: Encrypted SealedSecret resources are committed to Git
4. **Decryption**: Controller automatically decrypts and creates Secret resources in cluster

### Creating Sealed Secrets

```bash
# Create a regular secret (don't commit this!)
kubectl create secret generic my-secret \
  --from-literal=key=value \
  --dry-run=client -o yaml > /tmp/secret.yaml

# Encrypt it with kubeseal
kubeseal --format yaml \
  --controller-namespace sealed-secrets \
  --controller-name sealed-secrets \
  < /tmp/secret.yaml > sealed-secret.yaml

# Commit the sealed secret to Git
git add sealed-secret.yaml
git commit -m "Add sealed secret"
git push

# Flux will deploy it and Sealed Secrets will decrypt it
```

**→ Sealed Secrets Documentation**: [Bitnami Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)

---

## 📝 Files

| File | Purpose |
|------|---------|
| `main.go` | Main Pulumi program (cluster provisioning) |
| `go.mod` | Go module dependencies |
| `Pulumi.yaml` | Pulumi project configuration |
| `.gitignore` | Excludes Pulumi state and build artifacts |
| `README.md` | This file |

---

## 🔄 Workflow

### First Deployment

```bash
# Export required environment variables
export GITHUB_TOKEN="ghp_..."
export CLOUDFLARE_API_KEY="..."  # Optional for some clusters

# Prepare Kind substrate (registry + Kind cluster)
make bootstrap ENV=studio

# Deploy Pulumi-managed components
make up ENV=studio

# Pulumi will:
# 1. Auto-restore sealed-secrets keys (if backups exist)
# 2. Ensure Flux bootstrap manifests are applied declaratively
# 3. Create GitHub token secret
# 4. Configure Flux Git/Helm repositories
# 5. Deploy sealed/external secrets bootstrap
# 6. Apply cluster root kustomization so Flux manages everything else
```

### Subsequent Updates

```bash
# Make changes to cluster config in flux/clusters/studio/
git commit -m "Update studio cluster config"
git push

# Option 1: Let Flux auto-sync (default: 1 minute)
# Option 2: Force reconcile
make reconcile-all

# Option 3: Re-run Pulumi (if infrastructure changed)
make up ENV=studio
```

---

## 🛠️ Development

### Prerequisites

- Go 1.25+
- Pulumi CLI
- kubectl
- kind
- flux CLI

### Local Development

```bash
cd pulumi

# Install dependencies
go mod download

# Format code
go fmt ./...

# Run locally (not recommended, use Makefile instead)
PULUMI_CONFIG_PASSPHRASE="" pulumi stack select studio
PULUMI_CONFIG_PASSPHRASE="" pulumi up
```

### Adding New Features

1. **Cluster-level features**: Add to `flux/clusters/<name>/`
2. **Infrastructure changes**: Modify `main.go` if needed
3. **Script changes**: Update `scripts/` and reference in `main.go`

---

## 🐛 Troubleshooting

### Pulumi stack is locked

```bash
make cancel ENV=studio
```

### Cluster exists but Pulumi doesn't know

```bash
cd pulumi
PULUMI_CONFIG_PASSPHRASE="" pulumi refresh --yes
```

### Sealed Secrets not working

```bash
# Check if Sealed Secrets controller is running
kubectl get pods -n sealed-secrets

# Check controller logs
kubectl logs -n sealed-secrets -l app.kubernetes.io/name=sealed-secrets

# Verify a sealed secret
kubectl get sealedsecret -A
kubectl get secret <secret-name> -n <namespace>
```

### Clean slate

```bash
# Destroy cluster (must be run from the machine where cluster exists)
make destroy ENV=studio

# Recreate cluster
make up ENV=studio
```

### Multi-Machine Deployment

When deploying to multiple machines or clients:

```bash
# On Machine A (e.g., studio machine)
make up ENV=studio
# Creates stack: bvlucena/studio-studio

# On Machine B (e.g., pro machine)  
make up ENV=studio
# Creates stack: bvlucena/studio-pro

# Each machine has its own stack, preventing conflicts
# List all stacks:
make list-stacks
```

**Key Points**:
- Each machine automatically gets a unique stack name based on hostname
- Stacks are isolated per machine, so multiple machines can deploy the same `ENV`
- Destruction is protected - you can only destroy clusters that exist locally
- Perfect for client deployments where each client has their own machine

---

## 📚 References

- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)
- [Kind Documentation](https://kind.sigs.k8s.io/)
- [Flux Documentation](https://fluxcd.io/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)

---

**Last Updated**: 2025-11-07  
**Maintained by**: SRE Team (Bruno Lucena)

