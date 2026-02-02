# Agents WhatsApp Rust - Deployment Guide

## Quick Start

### Build and Push Images

```bash
# Build and push to local registry (localhost:5001)
make build-images-local

# Build and push to ghcr.io (multi-arch)
make push-ghcr-multiarch

# Or build individual services
make messaging-build-image-local
make user-build-image-local
make agent-gateway-build-image-local
make message-storage-build-image-local
```

### Deploy to Kubernetes

```bash
# Deploy using kustomize
make deploy

# Or manually
kubectl apply -k k8s/base

# For production (studio cluster)
kubectl apply -k k8s/overlays/studio
```

## Makefile Commands

### Version Management

```bash
make version                    # Show current version
make version-bump NEW_VERSION=1.0.1  # Bump version
make release-patch              # Bump patch version
make release-minor              # Bump minor version
make release-major              # Bump major version
```

### Image Building

```bash
make build-images-local         # Build all images (arm64) to local registry
make build-images               # Build all images (multi-arch) to local registry
make push-ghcr                  # Push all images to ghcr.io (multi-arch)
```

### Deployment

```bash
make deploy                     # Deploy to Kubernetes
make deploy-diff                # Show deployment diff
make kustomize-build            # Build kustomize output
```

### Status

```bash
make status                     # Show full system status
make status-pods                # Show pod status
make status-services            # Show service status
```

## Project Structure

```
agents-whatsapp-rust/
├── Cargo.toml                  # Workspace configuration
├── VERSION                     # Version file
├── Makefile                    # Build and deployment automation
├── README.md                   # Project overview
├── DEPLOYMENT.md              # This file
├── shared/                     # Shared library
│   ├── Cargo.toml
│   └── src/
│       ├── lib.rs
│       ├── models.rs          # Data models
│       ├── errors.rs          # Error types
│       └── utils.rs           # Utility functions
├── messaging-service/          # WebSocket server
│   ├── Cargo.toml
│   ├── Dockerfile
│   └── src/
├── user-service/              # Authentication service
│   ├── Cargo.toml
│   ├── Dockerfile
│   └── src/
├── agent-gateway/             # Message routing
│   ├── Cargo.toml
│   ├── Dockerfile
│   └── src/
├── message-storage-service/   # MongoDB persistence
│   ├── Cargo.toml
│   ├── Dockerfile
│   └── src/
└── k8s/                       # Kubernetes manifests
    ├── base/                  # Base resources
    │   ├── kustomization.yaml
    │   ├── namespace.yaml
    │   ├── messaging-service.yaml
    │   ├── user-service.yaml
    │   ├── agent-gateway.yaml
    │   ├── message-storage-service.yaml
    │   └── broker-trigger.yaml
    └── overlays/
        ├── pro/               # Production overlay
        └── studio/            # Studio cluster overlay
```

## Prerequisites

- Rust 1.70+
- Docker with buildx
- Kubernetes cluster with Knative installed
- MongoDB replica set
- Redis
- Knative Broker configured

## Environment Variables

### Messaging Service
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: Database name (default: messaging_app)
- `REDIS_URI`: Redis connection string
- `BROKER_URL`: Knative Broker URL

### User Service
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: Database name
- `JWT_SECRET`: JWT signing secret (from Kubernetes Secret)

### Agent Gateway
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: Database name
- `REDIS_URI`: Redis connection string
- `BROKER_URL`: Knative Broker URL

### Message Storage Service
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: Database name
- `REDIS_URI`: Redis connection string

## Flux GitOps Integration

To integrate with Flux, add to your cluster kustomization:

```yaml
resources:
  - ../../../../ai/agents-whatsapp-rust/k8s/overlays/studio
```

## Image Registries

### Local Registry (Development)
- Registry: `localhost:5001`
- Images: `localhost:5001/agents-whatsapp-rust-{service}:v{VERSION}`

### GitHub Container Registry (Production)
- Registry: `ghcr.io`
- Images: `ghcr.io/brunovlucena/agents-whatsapp-rust-{service}:v{VERSION}`
- Also tagged as: `latest`

## Troubleshooting

### Build Issues

```bash
# Ensure buildx builder exists
make ensure-buildx-builder

# Check version
make version
```

### Deployment Issues

```bash
# Check status
make status

# View logs
kubectl logs -f deployment/messaging-service -n homelab-services

# Check Knative services
kubectl get ksvc -n homelab-services
```

### Image Pull Issues

```bash
# Verify images exist
docker pull localhost:5001/agents-whatsapp-rust-messaging:v1.0.0

# Check registry connectivity
curl http://localhost:5001/v2/_catalog
```
