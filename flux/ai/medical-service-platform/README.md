# 🏥 Medical Service Platform

**AI-Powered Medical Consultation Platform with Agent Integration**

A complete medical service platform that enables doctors to interact with their personal AI medical assistant through a WhatsApp-like interface, with support for case summarization, patient record management, and medical correlations.

## 🎯 Overview

This platform integrates:
- **agent-medical**: HIPAA-compliant medical records agent
- **agents-whatsapp-rust**: Production-ready messaging platform
- **Web & Mobile Apps**: Cross-platform doctor interface

### Key Features

- ✅ **Doctor Authentication**: Secure login for medical professionals
- ✅ **Personal AI Assistant**: Each doctor has their own agent-medical instance
- ✅ **Case Summarization**: AI summarizes patient cases automatically
- ✅ **Patient Records**: Access and manage patient exams, lab results, prescriptions
- ✅ **Medical Correlations**: AI-powered analysis and pattern detection
- ✅ **Real-time Messaging**: WhatsApp-like interface for doctor-agent communication
- ✅ **Push Notifications**: Alerts for urgent cases, new exam results, etc.
- ✅ **Cross-Platform**: Works on web browsers and mobile devices

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Doctor Applications                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Web App     │  │  iOS App     │  │ Android App  │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼─────────────────┼─────────────────┼─────────────────┘
          │                 │                 │
          │    WebSocket    │    WebSocket    │    WebSocket
          │                 │                 │
┌─────────▼─────────────────▼─────────────────▼──────────────────┐
│              agents-whatsapp-rust (Messaging Layer)            │
│  • messaging-service (WebSocket)                               │
│  • agent-gateway (Route to agent-medical)                      │
│  • notification-service (Push notifications)                    │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ CloudEvents
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│              medical-service (Integration Layer)                │
│  • Doctor session management                                    │
│  • Agent-medical integration                                    │
│  • Case summarization                                           │
│  • Patient record correlation                                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ CloudEvents
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│              agent-medical (Medical AI Agent)                   │
│  • HIPAA-compliant medical records access                       │
│  • Lab results, prescriptions, medical history                   │
│  • Drug interaction checking                                    │
│  • Medical protocol knowledge                                   │
└─────────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
medical-service-platform/
├── VERSION                    # Version file (single source of truth)
├── Makefile                  # Build, test, deploy commands
├── README.md                 # This file
├── src/
│   └── medical-service/      # Rust service (integration layer)
│       ├── Cargo.toml
│       ├── Dockerfile
│       └── src/
│           ├── main.rs
│           ├── handlers.rs
│           ├── doctor.rs
│           ├── agent.rs
│           └── notifications.rs
├── web/
│   └── doctor-dashboard/     # Next.js web app
├── mobile/
│   └── doctor-app/           # React Native mobile app
├── k8s/
│   └── kustomize/
│       ├── base/             # Base Kubernetes resources
│       ├── pro/              # Production overlay
│       └── studio/           # Studio overlay
└── docs/
    └── integration.md       # Integration guide
```

## 🚀 Quick Start

### Prerequisites

- Rust 1.70+
- Docker & Docker Buildx
- Kubernetes cluster with Knative
- MongoDB (replica set)
- Redis
- agent-medical deployed
- agents-whatsapp-rust deployed

### Build

```bash
# Build and push to local registry
make build-local

# Build for GHCR
make build
make push
```

### Deploy

```bash
# Deploy to studio
make deploy-studio

# Deploy to pro
make deploy-pro
```

### Version Management

```bash
# Show current version
make version

# Bump version (updates VERSION file and all kustomizations)
make version-bump NEW_VERSION=0.2.0

# Auto-bump versions
make release-patch    # 0.1.0 → 0.1.1
make release-minor    # 0.1.0 → 0.2.0
make release-major    # 0.1.0 → 1.0.0

# Full release (bump + build + deploy)
make release NEW_VERSION=0.2.0 ENV=studio
```

## 📋 Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build-local` | Build and push to local registry |
| `make build` | Build Docker image for GHCR |
| `make push` | Push to GHCR |
| `make test` | Run tests |
| `make deploy-studio` | Deploy to studio environment |
| `make deploy-pro` | Deploy to pro environment |
| `make version` | Show current version |
| `make version-bump NEW_VERSION=x.y.z` | Bump version (DRY pattern) |
| `make release-patch` | Auto-bump patch version |
| `make release-minor` | Auto-bump minor version |
| `make release-major` | Auto-bump major version |
| `make status` | Show deployment status |
| `make logs` | Tail logs |
| `make clean` | Clean build artifacts |

## 🔐 Security

- HIPAA-compliant data handling
- JWT-based authentication
- Role-based access control (RBAC)
- End-to-end encryption for messages
- Audit logging for all medical data access

## 📖 Documentation

- [Integration Guide](./docs/integration.md) - How to integrate with agent-medical and agents-whatsapp-rust
- [Architecture Details](./docs/architecture.md) - Detailed architecture documentation

## 📝 License

Part of the homelab project. MIT License.

---

**🏥 HIPAA-Compliant Medical Service Platform with AI! 🏥**
