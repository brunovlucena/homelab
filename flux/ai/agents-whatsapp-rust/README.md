# Agents WhatsApp Rust

Production-ready messaging platform built in Rust with Tokio and Axum.

## Project Structure

```
agents-whatsapp-rust/
├── Cargo.toml              # Workspace configuration
├── shared/                  # Shared library (models, errors, utils)
├── messaging-service/       # WebSocket server
├── user-service/            # Authentication & user management
├── agent-gateway/           # Message routing to agents
├── media-service/           # File upload/download
├── presence-service/        # Online/offline status
├── notification-service/    # Push notifications
├── message-storage-service/ # MongoDB persistence
└── k8s/                     # Kubernetes manifests
```

## Services

### Messaging Service
- WebSocket server for real-time communication
- Redis Pub/Sub for cross-instance routing
- MongoDB for idempotency and sequence numbers
- Connection management and migration

### User Service
- User registration and authentication
- JWT token generation/validation
- Profile management

### Agent Gateway
- Routes messages to AI agents
- Idempotency checking
- Sequence number generation
- CloudEvents integration

### Message Storage Service
- Consumes from Knative Broker
- Stores messages in MongoDB
- Updates conversation metadata

## Quick Start

### Prerequisites
- Rust 1.70+
- MongoDB (replica set)
- Redis
- Kubernetes cluster with Knative

### Build

```bash
cargo build --release
```

### Run Locally

```bash
# Set environment variables
export MONGODB_URI="mongodb://localhost:27017"
export REDIS_URI="redis://localhost:6379"
export MONGODB_DATABASE="messaging_app"

# Run messaging service
cd messaging-service
cargo run
```

### Deploy to Kubernetes

```bash
# Build Docker images
docker build -t messaging-service:latest ./messaging-service

# Apply Kubernetes manifests
kubectl apply -f k8s/
```

## Architecture

See [docs/agents-whatsapp-rust/ARCHITECTURE.md](../../docs/agents-whatsapp-rust/ARCHITECTURE.md) for complete architecture documentation.

## Development

### Running Tests

```bash
cargo test
```

### Code Formatting

```bash
cargo fmt
```

### Linting

```bash
cargo clippy
```

## Status

🟡 **In Development** - Core services being implemented
