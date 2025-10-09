# 🌐 Bruno Site API

Go-based API server for the Bruno Site with MinIO asset proxy support.

## 🎯 Features

- **🖼️ Asset Proxy**: Proxies images and assets from internal MinIO to internet clients
- **📊 Project Management**: CRUD operations for projects
- **🔴 Redis Caching**: Fast data access with Redis
- **🗄️ PostgreSQL**: Reliable data persistence
- **🏥 Health Checks**: Built-in health monitoring

## 🏗️ Architecture

```
Internet Client
      ↓
Frontend (React)
      ↓
API Server (Go)
      ↓
MinIO (Internal) → Assets proxied to client
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker (optional)
- Access to PostgreSQL, Redis, and MinIO

### Environment Variables

```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/dbname?sslmode=disable

# Redis
REDIS_URL=redis://host:6379

# Server
PORT=8080
CORS_ORIGIN=*

# MinIO (for asset proxy)
MINIO_ENDPOINT=minio-service.minio.svc.cluster.local:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=homepage-assets
```

### Run Locally

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

### Run with Docker

```bash
# Build
docker build -t bruno-site-api .

# Run
docker run -p 8080:8080 \
  -e DATABASE_URL=... \
  -e REDIS_URL=... \
  -e MINIO_ENDPOINT=... \
  bruno-site-api
```

## 📁 Project Structure

```
api/
├── main.go              # Application entry point
├── config/              # Configuration management
│   └── config.go
├── database/            # Database connections
│   ├── database.go
│   └── redis.go
├── handlers/            # HTTP handlers
│   ├── health.go
│   ├── projects.go
│   └── assets.go        # Asset proxy handler
├── router/              # Route setup
│   └── router.go
└── storage/             # MinIO client
    └── minio.go
```

## 🔌 API Endpoints

### Health

```
GET /health
```

### Projects

```
GET    /api/v1/projects      # List all projects
GET    /api/v1/projects/:id  # Get project by ID
POST   /api/v1/projects      # Create project
PUT    /api/v1/projects/:id  # Update project
DELETE /api/v1/projects/:id  # Delete project
```

### Assets (Proxy)

```
GET /api/v1/assets/*path     # Proxy asset from MinIO
```

Example:
- Request: `GET /api/v1/assets/eu.webp`
- Proxies: MinIO object `homepage-assets/eu.webp` to client

## 🖼️ Asset Proxy

The asset proxy allows internet clients to access images stored in internal MinIO without exposing MinIO directly:

1. **Client requests**: `https://yourdomain.com/api/v1/assets/eu.webp`
2. **API fetches**: `minio-service.minio.svc.cluster.local:9000/homepage-assets/eu.webp`
3. **API streams**: Image data back to client

### Benefits

- ✅ MinIO stays internal (no internet exposure)
- ✅ Centralized access control
- ✅ Can add caching, authentication, rate limiting
- ✅ Single endpoint for all assets

## 🔧 Development

### Hot Reload

```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Testing

```bash
# Run tests
go test ./...

# Test health endpoint
curl http://localhost:8080/health

# Test asset proxy
curl http://localhost:8080/api/v1/assets/eu.webp
```

## 🐳 Docker Support

### Production Build

```dockerfile
FROM golang:1.21-alpine AS builder
# ... multi-stage build
```

### Development Build

```dockerfile
FROM golang:1.21-alpine
# ... with air hot reload
```

## 📊 Monitoring

The API exposes metrics at `/metrics` (if enabled) for Prometheus scraping.

## 🔒 Security

- CORS configuration
- Input validation
- SQL injection protection (via GORM)
- Environment-based secrets

## 📝 License

MIT License

