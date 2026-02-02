# Homepage Architecture & Workflows

## 🏗️ System Architecture Overview

This document explains how the homepage application works, including production flows, development workflows, and how Vite fits into the picture.

---

## 📊 Production Architecture Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         PRODUCTION FLOW                                  │
└─────────────────────────────────────────────────────────────────────────┘

    Internet User
         │
         ▼
    ┌─────────┐
    │Cloudflare│  (CDN, DDoS protection, SSL termination)
    │  Tunnel  │
    └────┬─────┘
         │
         ▼
    ┌──────────────────────────────────────────────────────────────┐
    │                    Kubernetes Cluster                          │
    │                                                                │
    │  ┌────────────────────────────────────────────────────────┐  │
    │  │         Frontend Pod (homepage-frontend)                │  │
    │  │  ┌──────────────────────────────────────────────────┐  │  │
    │  │  │  nginx (port 8080)                               │  │  │
    │  │  │  - Serves static files from /usr/share/nginx/html │  │  │
    │  │  │  - Built by Vite (React app bundled)              │  │  │
    │  │  │  - Proxies /api/* to API service                  │  │  │
    │  │  └──────────────────────────────────────────────────┘  │  │
    │  └────────────────────────────────────────────────────────┘  │
    │                        │                                       │
    │                        │ /api/* requests                       │
    │                        ▼                                       │
    │  ┌────────────────────────────────────────────────────────┐  │
    │  │         API Pod (homepage-api)                          │  │
    │  │  ┌──────────────────────────────────────────────────┐  │  │
    │  │  │  Go API Server (Gin framework, port 8080)       │  │  │
    │  │  │  - REST endpoints: /api/projects, /api/skills    │  │  │
    │  │  │  - Chat endpoint: /api/chat                       │  │  │
    │  │  │  - Health: /health, Metrics: /metrics            │  │  │
    │  │  └──────────────────────────────────────────────────┘  │  │
    │  └────────────────────────────────────────────────────────┘  │
    │         │              │              │                       │
    │         │              │              │                       │
    │         ▼              ▼              ▼                       │
    │  ┌──────────┐   ┌──────────┐   ┌──────────┐                │
    │  │PostgreSQL│   │  Redis   │   │Agent-Bruno│                │
    │  │(postgres │   │(redis    │   │(LLM Chat) │                │
    │  │ namespace)│   │ namespace)│   │          │                │
    │  └──────────┘   └──────────┘   └──────────┘                │
    │                                                                │
    └────────────────────────────────────────────────────────────────┘
```

### 🔄 Request Flow in Production

1. **User Request** → Cloudflare Tunnel receives HTTPS request
2. **Cloudflare** → Routes to Kubernetes Service (homepage-frontend)
3. **Frontend Pod (nginx)**:
   - If request is `/api/*` → Proxies to `homepage-api.homepage.svc.cluster.local:8080`
   - If request is static file → Serves from `/usr/share/nginx/html` (Vite-built assets)
   - If request is `/` or route → Serves `index.html` (React Router handles routing client-side)
4. **API Pod**:
   - Processes request (e.g., `/api/projects`)
   - Queries PostgreSQL for data
   - Uses Redis for caching
   - Returns JSON response
5. **Response** → Frontend → Cloudflare → User

---

## 🛠️ Development Workflows

### Local Development (No Kubernetes)

```
┌─────────────────────────────────────────────────────────────────┐
│                    LOCAL DEV WORKFLOW                            │
└─────────────────────────────────────────────────────────────────┘

Developer Machine
     │
     ├─► Frontend Dev Server (Vite)
     │   └─► npm run dev
     │       ├─► Runs on http://localhost:5173
     │       ├─► Hot Module Replacement (HMR) enabled
     │       ├─► Watches file changes
     │       └─► Proxies /api/* to http://localhost:8080
     │
     └─► API Dev Server (Go)
         └─► go run main.go
             ├─► Runs on http://localhost:8080
             ├─► Connects to PostgreSQL (via port-forward or local)
             └─► Connects to Redis (via port-forward or local)

Makefile: make dev
    - Starts API in background
    - Starts Vite dev server in foreground
    - Both run locally, no containers
```

**Vite Dev Server Configuration** (`vite.config.ts`):
- Port: `8080` (but Makefile runs on `5173` for local dev)
- Proxy: `/api/*` → `http://homepage-api.homepage.svc.cluster.local:8080` (for K8s)
- For local: Vite proxies to `http://localhost:8080` (API running locally)

---

### Telepresence Development Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│              TELEPRESENCE DEV WORKFLOW (Hybrid Local + K8s)             │
└─────────────────────────────────────────────────────────────────────────┘

Developer Machine                    Kubernetes Cluster
     │                                      │
     │  1. Setup Port-Forwards              │
     ├─► kubectl port-forward               │
     │   - agent-injector:8443              │
     │   - traffic-manager:8081             │
     │   - postgres:5432                    │
     │                                      │
     │  2. Connect Telepresence             │
     ├─► telepresence connect               │
     │   - Creates VPN tunnel to cluster   │
     │   - Maps cluster DNS to local        │
     │                                      │
     │  3. Intercept API Service            │
     ├─► make tp-api                        │
     │   ┌──────────────────────────────┐  │
     │   │  Local Docker Container      │  │  ┌──────────────────────┐
     │   │  - Runs API image            │  │  │  K8s Service         │
     │   │  - Port 8080                 │◄─┼──│  homepage-api        │
     │   │  - Connects to localhost:5432 │  │  │  (intercepted)      │
     │   │    (via port-forward)        │  │  └──────────────────────┘
     │   └──────────────────────────────┘  │
     │                                      │
     │  4. Intercept Frontend Service       │
     ├─► make tp-frontend                   │
     │   ┌──────────────────────────────┐  │  ┌──────────────────────┐
     │   │  Local Docker Container     │  │  │  K8s Service         │
     │   │  - Runs frontend dev image  │  │  │  homepage-frontend    │
     │   │  - Port 80:8080             │◄─┼──│  (intercepted)       │
     │   │  - Volume mount:            │  │  └──────────────────────┘
     │   │    ./src/frontend → /app    │  │
     │   │  - Vite dev server          │  │
     │   │  - Hot reload enabled        │  │
     │   └──────────────────────────────┘  │
     │                                      │
     │  Traffic Flow:                       │
     │  User → Cloudflare → K8s Service     │
     │         → Telepresence Intercept     │
     │         → Local Container            │
     │         → (API: localhost:5432)      │
     │         → (Frontend: Vite HMR)       │
     │                                      │
     └──────────────────────────────────────┘
```

**Telepresence Commands** (from Makefile):

1. **Setup**: `make tp-port-forward`
   - Port-forwards Telepresence services and PostgreSQL

2. **Connect**: `make tp-api` or `make tp-frontend`
   - Builds local Docker image
   - Creates Telepresence intercept
   - Routes cluster traffic to local container
   - For API: Uses local PostgreSQL via port-forward
   - For Frontend: Mounts local code, runs Vite dev server

3. **Cleanup**: `make tp-clean`
   - Stops intercepts and containers

---

## 🏭 Build & Deployment Process

### Frontend Build (Vite)

```
┌─────────────────────────────────────────────────────────────────┐
│                    FRONTEND BUILD PROCESS                       │
└─────────────────────────────────────────────────────────────────┘

Source Code (React + TypeScript)
     │
     ▼
┌─────────────────────────────────┐
│  npm run build                  │
│  (runs: vite build)             │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Vite Build Steps:               │
│  1. TypeScript compilation       │
│  2. React component bundling     │
│  3. Code splitting               │
│     - vendor.js (React, ReactDOM)│
│     - router.js (React Router)   │
│     - app.js (your code)         │
│  4. Asset optimization           │
│  5. Output to /dist              │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Docker Build (Dockerfile)       │
│  Stage 1: Builder                │
│  - node:22-alpine                │
│  - npm ci                        │
│  - vite build                    │
│                                  │
│  Stage 2: Production             │
│  - nginx:alpine                  │
│  - Copy /dist → /usr/share/     │
│    nginx/html                    │
│  - Copy nginx.conf               │
└────────────┬────────────────────┘
             │
             ▼
    Docker Image: homepage-frontend:v0.1.20
             │
             ▼
    Pushed to: localhost:5001 (or GHCR)
             │
             ▼
    Kubernetes Deployment
    - Pulls image
    - Runs nginx serving static files
```

### API Build

```
┌─────────────────────────────────────────────────────────────────┐
│                      API BUILD PROCESS                          │
└─────────────────────────────────────────────────────────────────┘

Source Code (Go)
     │
     ▼
┌─────────────────────────────────┐
│  Docker Build (Dockerfile)       │
│  Stage 1: Builder                │
│  - golang:1.25-alpine            │
│  - go mod download               │
│  - go build (static binary)      │
│                                  │
│  Stage 2: Production             │
│  - scratch (minimal image)       │
│  - Copy binary + CA certs        │
└────────────┬────────────────────┘
             │
             ▼
    Docker Image: homepage-api:v0.1.19
             │
             ▼
    Pushed to: localhost:5001 (or GHCR)
             │
             ▼
    Kubernetes Deployment
    - Pulls image
    - Runs Go binary
    - Connects to PostgreSQL & Redis
```

---

## 🔧 How Vite Works

### What is Vite?

Vite is a **build tool and dev server** for modern web applications. It's the replacement for Webpack/CRA.

### Vite in Development Mode

```
┌─────────────────────────────────────────────────────────────────┐
│              VITE DEV SERVER (npm run dev)                      │
└─────────────────────────────────────────────────────────────────┘

Browser Request: http://localhost:5173/
     │
     ▼
┌─────────────────────────────────┐
│  Vite Dev Server                │
│  - Listens on port 5173         │
│  - Serves index.html            │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  index.html loads:               │
│  <script type="module"          │
│    src="/src/main.tsx">         │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Vite transforms on-the-fly:    │
│  - TypeScript → JavaScript       │
│  - JSX → React.createElement()  │
│  - Imports → ES modules         │
│  - CSS → Injected <style> tags  │
└────────────┬────────────────────┘
             │
             ▼
    Browser executes transformed code
             │
             ▼
    File Change Detected
             │
             ▼
    Hot Module Replacement (HMR)
    - Updates changed component
    - Preserves React state
    - No full page reload
```

**Key Vite Features**:
- **Fast HMR**: Only updates changed modules
- **ESM-based**: Uses native ES modules in dev
- **On-demand compilation**: Only compiles what's requested
- **Proxy support**: Forwards `/api/*` to backend

### Vite in Production Build

```
┌─────────────────────────────────────────────────────────────────┐
│              VITE BUILD (npm run build)                        │
└─────────────────────────────────────────────────────────────────┘

Source Files
     │
     ▼
┌─────────────────────────────────┐
│  Vite Build Process             │
│  1. TypeScript → JavaScript     │
│  2. JSX → React code            │
│  3. Tree-shaking (remove unused)│
│  4. Code splitting              │
│  5. Minification                │
│  6. Asset optimization          │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Output: /dist                   │
│  ├── index.html                 │
│  ├── assets/                    │
│  │   ├── vendor-abc123.js       │
│  │   ├── router-def456.js      │
│  │   ├── app-ghi789.js          │
│  │   └── styles-jkl012.css     │
│  └── ...                        │
└────────────┬────────────────────┘
             │
             ▼
    Static files served by nginx
```

---

## 📡 API Communication Flow

### Frontend → API Communication

```
┌─────────────────────────────────────────────────────────────────┐
│              FRONTEND API CALLS                                  │
└─────────────────────────────────────────────────────────────────┘

React Component (e.g., Home.tsx)
     │
     ▼
┌─────────────────────────────────┐
│  apiClient.getProjects()        │
│  (from src/services/api.ts)      │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Axios Request                  │
│  baseURL: '/api'                │
│  (or VITE_API_URL env var)      │
└────────────┬────────────────────┘
             │
     ┌───────┴───────┐
     │               │
     ▼               ▼
┌─────────┐    ┌─────────┐
│  Dev    │    │  Prod   │
│  Mode   │    │  Mode   │
└────┬────┘    └────┬────┘
     │              │
     │              │
     ▼              ▼
┌─────────────────────────────────┐
│  Vite Proxy (dev) or            │
│  nginx Proxy (prod)             │
│  /api/* → API service           │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  Go API Server                   │
│  - Handles /api/projects         │
│  - Queries PostgreSQL            │
│  - Returns JSON                  │
└────────────┬────────────────────┘
             │
             ▼
    React Component receives data
    Updates UI via React Query
```

**Configuration**:
- **Dev**: Vite proxy in `vite.config.ts` forwards `/api/*` to API
- **Prod**: nginx in `nginx.conf` proxies `/api/*` to `homepage-api.homepage.svc.cluster.local:8080`
- **Frontend code**: Uses relative URLs (`/api/projects`) so it works in both modes

---

## 🗄️ Database & Cache Flow

```
┌─────────────────────────────────────────────────────────────────┐
│              DATA LAYER                                          │
└─────────────────────────────────────────────────────────────────┘

API Request (e.g., GET /api/projects)
     │
     ▼
┌─────────────────────────────────┐
│  Go Handler                     │
│  - Checks Redis cache first     │
└────────────┬────────────────────┘
     │       │
     │       ▼ Cache Miss
     │  ┌──────────┐
     │  │  Redis   │
     │  │  (miss)  │
     │  └──────────┘
     │
     ▼ Cache Miss or Write
┌─────────────────────────────────┐
│  PostgreSQL Query                │
│  - SELECT * FROM projects       │
└────────────┬────────────────────┘
     │
     ▼
┌─────────────────────────────────┐
│  Store in Redis (for reads)     │
│  - Key: projects:all            │
│  - TTL: 5 minutes               │
└────────────┬────────────────────┘
     │
     ▼
    Return JSON to frontend
```

---

## 🚀 Deployment Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│              DEPLOYMENT PIPELINE                                 │
└─────────────────────────────────────────────────────────────────┘

Developer
     │
     ▼
┌─────────────────────────────────┐
│  make deploy                    │
│  (or make deploy-frontend/api)  │
└────────────┬────────────────────┘
     │
     ├─► 1. Git commit & push
     │
     ├─► 2. Build Docker images
     │   make build-images-local
     │   - Builds API & Frontend
     │   - Tags with version from VERSION file
     │   - Pushes to localhost:5001
     │
     ├─► 3. Rollout restart
     │   make rollout
     │   - kubectl rollout restart
     │   - Waits for new pods to be ready
     │
     └─► 4. Verify
         make verify
         - Checks pod status
         - Tests /health endpoints
```

**Alternative**: GitHub Actions builds multi-arch images to GHCR, then `make sync-images` pulls them to local registry.

---

## 🔑 Key Configuration Files

### Frontend
- **`vite.config.ts`**: Vite dev server config, proxy settings, build options
- **`nginx.conf`**: Production nginx config (proxies `/api/*` to API)
- **`src/services/api.ts`**: Axios client for API calls
- **`Dockerfile`**: Multi-stage build (Node builder → nginx production)
- **`Dockerfile.dev`**: Dev image with Vite dev server

### API
- **`main.go`**: Go server setup, routes, middleware
- **`Dockerfile`**: Multi-stage build (Go builder → scratch production)
- **`Dockerfile.dev`**: Dev image for local development

### Kubernetes
- **`k8s/kustomize/base/*.yaml`**: Base deployments, services
- **`k8s/kustomize/studio/*.yaml`**: Environment-specific overrides

### Makefile
- **`make dev`**: Run both locally (no containers)
- **`make tp-api`**: Telepresence intercept for API
- **`make tp-frontend`**: Telepresence intercept for frontend
- **`make deploy`**: Full deployment pipeline

---

## 💡 Quick Reference

### Development Commands

```bash
# Local development (no K8s)
make dev                    # Run API + Frontend locally

# Telepresence (hybrid local + K8s)
make tp-port-forward        # Setup port-forwards
make tp-api                 # Intercept API with local container
make tp-frontend           # Intercept Frontend with local container
make tp-clean              # Clean up intercepts

# Build & Deploy
make build-images-local     # Build Docker images
make deploy-frontend        # Deploy frontend only
make deploy-api            # Deploy API only
make deploy                # Full deployment
make rollout               # Restart deployments
```

### Understanding the Flow

1. **Production**: User → Cloudflare → nginx (static files) → API (Go) → PostgreSQL/Redis
2. **Local Dev**: Vite dev server (port 5173) → API (port 8080) → Local DB
3. **Telepresence**: Cloudflare → K8s Service → Telepresence → Local Container → Local DB
4. **Vite**: Dev server with HMR, or build tool for production bundles

---

## 🎯 Summary

- **Vite**: Dev server (HMR) + build tool (bundles React app)
- **nginx**: Serves static files in production, proxies API requests
- **API**: Go server handling REST endpoints, connects to PostgreSQL/Redis
- **Telepresence**: Routes K8s traffic to local containers for development
- **Flow**: Frontend makes `/api/*` calls → proxied to API → queries DB → returns JSON

The beauty of this setup: Frontend code uses relative URLs (`/api/*`), so it works the same in dev (Vite proxy) and prod (nginx proxy) without changes!
