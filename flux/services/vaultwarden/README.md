# 🔐 Self-Hosted Password Manager (Vault-based)

A self-hosted password manager compatible with Bitwarden clients, using HashiCorp Vault as the backend storage engine.

**Now with AppAgentVault iOS app** - An AgentApp-style iOS application for natural language password management.

## Overview

This project provides:
- **Backend API Server**: Go-based server implementing Bitwarden API compatibility
- **Vault Integration**: Stores encrypted passwords in HashiCorp Vault
- **AppAgentVault iOS App**: Native iOS client using AgentApp pattern with CloudEvents
- **Browser Extensions**: Chrome, Firefox, and Safari extensions
- **Kubernetes Deployment**: Full GitOps deployment using Flux

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                         │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │AppAgentVault│  │   Browser    │  │   Browser    │   │
│  │   (iOS)     │  │  Extension   │  │  Extension   │   │
│  │ CloudEvents │  │  (Chrome)    │  │  (Firefox)   │   │
│  └─────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────┬───────────────────────────┘
                              │ HTTPS/TLS
┌─────────────────────────────▼───────────────────────────┐
│                  API SERVER (Go)                        │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Bitwarden-compatible REST API                   │   │
│  │  CloudEvents endpoints (for AgentApp)           │   │
│  │  - Authentication (JWT)                          │   │
│  │  - Password CRUD operations                      │   │
│  │  - Encryption/Decryption (client-side)           │   │
│  └────────────────────┬─────────────────────────────┘   │
└────────────────────────┼─────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│              HASHICORP VAULT                             │
│  ┌──────────────────────────────────────────────────┐   │
│  │  KV Secrets Engine                               │   │
│  │  - Encrypted password storage                    │   │
│  │  - User metadata                                 │   │
│  │  - Audit logging                                 │   │
│  └──────────────────────────────────────────────────┘   │
└───────────────────────────────────────────────────────────┘
```

## Features

- ✅ Bitwarden API compatibility
- ✅ CloudEvents support (AgentApp pattern)
- ✅ Client-side encryption (server never sees plaintext passwords)
- ✅ HashiCorp Vault backend for secure storage
- ✅ Multi-device sync
- ✅ iOS native app (AppAgentVault)
- ✅ Browser extensions (Chrome, Firefox, Safari)
- ✅ Kubernetes deployment
- ✅ GitOps with Flux
- ✅ TLS/SSL encryption
- ✅ JWT authentication
- ✅ Natural language chat interface (iOS)

## Project Structure

```
vaultwarden/
├── backend/              # Go API server
│   ├── cmd/server/
│   ├── internal/
│   │   ├── api/         # HTTP handlers
│   │   ├── vault/       # Vault integration
│   │   ├── auth/        # Authentication logic
│   │   └── models/      # Data models
│   ├── go.mod
│   └── Dockerfile
├── ios/
│   └── AppAgentVault/   # iOS app (AgentApp pattern)
│       ├── Models/
│       ├── Services/
│       ├── ViewModels/
│       ├── Views/
│       └── AppAgentVaultApp.swift
├── browser-extension/    # Browser extensions
│   ├── chrome/
│   ├── firefox/
│   └── safari/
├── k8s/                  # Kubernetes manifests
│   └── base/
└── README.md
```

## Quick Start

See [QUICK_START.md](./QUICK_START.md) for detailed setup instructions.

### Deploy Backend

```bash
# Deploy Vault
kubectl apply -k flux/infrastructure/vault

# Deploy API Server
kubectl apply -k k8s/base
```

### Build iOS App

```bash
cd ios/AppAgentVault
# Open in Xcode and build
```

## AppAgentVault iOS App

The iOS app follows the AgentApp pattern used throughout the homelab:

- **CloudEvents Communication**: Uses CloudEvents protocol for natural language queries
- **Chat Interface**: Talk to your password manager in natural language
- **Traditional Vault View**: Direct password management interface
- **SwiftUI**: Modern, native iOS interface

See [ios/README.md](./ios/README.md) for iOS-specific documentation.

## API Endpoints

### REST API (Bitwarden-compatible)

- `POST /api/identity/connect/token` - User authentication
- `GET /api/ciphers` - List password entries
- `POST /api/ciphers` - Create password entry
- `PUT /api/ciphers/:id` - Update password entry
- `DELETE /api/ciphers/:id` - Delete password entry
- `GET /api/profile` - User profile

### CloudEvents API (AgentApp)

- `POST /api/vault/chat` - Natural language queries
- `POST /api/vault/save` - Save password via CloudEvent
- `GET /api/vault/list` - List passwords via CloudEvent

## Security

- All passwords encrypted client-side before sending to server
- Server only stores encrypted blobs
- HashiCorp Vault provides additional encryption at rest
- TLS/SSL for all communications
- JWT tokens for authentication
- Audit logging in Vault

## Development

See individual README files in each subdirectory for development instructions.

## License

MIT
