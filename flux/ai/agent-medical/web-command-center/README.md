# 🏥 Medical Agent Command Center

**HIPAA-Compliant Web Dashboard for Medical Records Management**

A modern, secure web interface for managing and monitoring the Medical Agent system. Built with Next.js 14, React 18, TypeScript, and Tailwind CSS.

## 🎯 Features

- **📊 Real-time Dashboard**: Monitor patient records, queries, and system health
- **👥 Patient Management**: Browse and manage patient data (HIPAA-compliant)
- **📄 Medical Records**: Access and search medical records
- **🛡️ HIPAA Compliance**: Built-in compliance monitoring and audit logging
- **🔔 Alerts & Notifications**: System alerts and compliance warnings
- **🎨 Modern UI**: Beautiful, responsive interface with smooth animations

## 🏗️ Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Data Fetching**: TanStack Query
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Charts**: Recharts

## 🚀 Quick Start

### Prerequisites

- Node.js 20+
- npm or yarn
- Docker (for containerization)
- Kubernetes cluster (for deployment)

### Local Development

```bash
# Install dependencies
npm ci

# Run development server
npm run dev

# Open http://localhost:3002
```

### Build

```bash
# Build Next.js app
npm run build

# Start production server
npm start
```

### Docker

```bash
# Build Docker image
make build

# Test locally
make test-local

# Push to local registry
make push-local
```

### Kubernetes Deployment

```bash
# Deploy to cluster
make deploy

# Check status
make status

# View logs
make logs

# Delete deployment
make delete
```

## 🌐 Access

Once deployed, the command center is accessible at:

- **Local (NodePort)**: http://localhost:30129 (via agent-medical-nodeport)
- **Cloudflare Tunnel**: https://medical.lucena.cloud

## 📋 Available Views

### Dashboard
- Real-time metrics and statistics
- Patient count, medical records, queries
- HIPAA compliance score
- Agent status monitoring

### Patients
- Patient list and search
- Patient management (when connected to backend)

### Medical Records
- Browse medical records
- Search functionality
- Record creation (when connected to backend)

### HIPAA Compliance
- Compliance dashboard
- Security features overview
- Audit log access
- Compliance score tracking

### Alerts
- System alerts and warnings
- Backend connection status
- Compliance notifications

## 🔧 Configuration

### Environment Variables

```bash
# Optional: Custom agent URL
AGENT_MEDICAL_URL=http://agent-medical.agent-medical.svc.cluster.local:8080
```

### Kubernetes Configuration

Edit `k8s/kustomize/base/deployment.yaml` to customize:
- Resource limits
- Replicas
- Environment variables
- Service configuration

## 🛡️ Security Features

- **HIPAA Compliant**: Built with HIPAA requirements in mind
- **Encrypted Communication**: TLS for all external connections
- **Role-Based Access**: Integration with backend RBAC
- **Audit Logging**: All actions logged for compliance
- **Secure by Default**: No sensitive data in frontend

## 📦 Project Structure

```
web-command-center/
├── src/
│   ├── app/                    # Next.js app directory
│   │   ├── api/               # API routes
│   │   │   ├── agents/        # Agent status endpoint
│   │   │   └── metrics/       # Metrics endpoint
│   │   ├── layout.tsx         # Root layout
│   │   ├── page.tsx           # Main page
│   │   └── globals.css        # Global styles
│   ├── components/            # React components
│   │   ├── DashboardView.tsx
│   │   ├── PatientsView.tsx
│   │   ├── RecordsView.tsx
│   │   ├── ComplianceView.tsx
│   │   ├── AlertsView.tsx
│   │   ├── Sidebar.tsx
│   │   └── Header.tsx
│   ├── lib/                   # Utilities
│   └── types/                 # TypeScript types
├── k8s/                       # Kubernetes manifests
│   └── kustomize/base/
│       ├── deployment.yaml
│       ├── cloudflare-tunnel-ingress.yaml
│       └── kustomization.yaml
├── Dockerfile                 # Container image
├── Makefile                   # Build automation
├── package.json              # Dependencies
├── tsconfig.json             # TypeScript config
├── tailwind.config.ts        # Tailwind config
└── next.config.js            # Next.js config
```

## 🔗 Integration

The command center connects to:
- **Medical Agent Backend**: `agent-medical.agent-medical.svc.cluster.local:8080`
- **Kubernetes API**: For agent status (optional)
- **Prometheus**: For metrics (optional)

## 🎨 Customization

### Colors

Edit `tailwind.config.ts` to customize the color scheme:

```typescript
colors: {
  'medical-blue': '#1e40af',
  'medical-green': '#059669',
  'medical-red': '#dc2626',
}
```

### Components

All components are in `src/components/` and use Tailwind CSS for styling.

## 📝 License

Part of the Homelab project by Bruno Lucena.

---

**🏥 HIPAA-Compliant Medical Records Management**
