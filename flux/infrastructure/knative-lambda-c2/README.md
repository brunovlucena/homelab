# Knative Lambda Command & Control Center

A modern and simple web interface for uploading and managing learning materials (runbooks, documentation) for chatbot fine-tuning. Files are stored in MinIO with agent-specific folder organization.

## Features

- 🤖 **Chatbot Learning Materials**: Upload runbooks and training materials for fine-tuning
- 📚 **Agent-Specific Folders**: Each agent has its own folder for organized storage
- 🚀 **Modern UI**: Clean, responsive React/TypeScript interface
- 📤 **File Upload**: Drag & drop or click to upload files
- 🔐 **Secure**: Presigned URL uploads directly to MinIO (no backend proxy)
- 📦 **Dual Target**: Upload to both Lambda Functions and Chatbot Agents
- 📋 **File Management**: List and delete uploaded files
- ⚡ **Fast**: Direct uploads to MinIO using presigned URLs
- 🔄 **Dynamic**: Automatically lists LambdaFunctions and LambdaAgents from Kubernetes

## Architecture

### Backend (FastAPI)
- FastAPI server providing REST API
- MinIO integration with presigned URLs
- Kubernetes client for dynamic resource listing
- Supports both `lambda-functions` and `agent-files` buckets

### Frontend (React + TypeScript)
- React 18 with TypeScript
- Vite for fast development and builds
- react-dropzone for file uploads
- Modern UI with gradient design
- Dynamic resource loading from Kubernetes

## Project Structure

```
knative-lambda-c2/
├── src/
│   ├── backend/          # FastAPI backend
│   └── frontend/         # React frontend
├── k8s/                  # Kubernetes manifests
└── README.md
```

## Quick Start

### Development

**Backend:**
```bash
cd src/backend
pip install -r requirements.txt
uvicorn main:app --reload --host 0.0.0.0 --port 8080
```

**Frontend:**
```bash
cd src/frontend
npm install
npm run dev
```

### Deployment

```bash
# Build images
cd src/backend
docker build -t localhost:5001/knative-lambda-c2-backend:latest .
docker push localhost:5001/knative-lambda-c2-backend:latest

cd ../frontend
docker build -t localhost:5001/knative-lambda-c2-frontend:latest .
docker push localhost:5001/knative-lambda-c2-frontend:latest

# Deploy to Kubernetes
kubectl apply -k k8s/
```

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/lambdas` - List LambdaFunctions from Kubernetes
- `GET /api/v1/agents` - List LambdaAgents from Kubernetes
- `POST /api/v1/files/presigned-url` - Generate presigned URL for upload
- `POST /api/v1/files/{fileId}/complete` - Complete upload
- `GET /api/v1/files/list` - List files in MinIO
- `DELETE /api/v1/files/{target}/{path}` - Delete file

## Configuration

### Environment Variables (Backend)

- `MINIO_ENDPOINT`: MinIO endpoint (default: `minio.minio.svc.cluster.local:9000`)
- `MINIO_ACCESS_KEY`: MinIO access key (from secret)
- `MINIO_SECRET_KEY`: MinIO secret key (from secret)
- `LAMBDA_BUCKET`: Bucket for lambda functions (default: `lambda-functions`)
- `AGENT_BUCKET`: Bucket for agent files (default: `agent-files`)

## File Structure

Files are stored in MinIO with the following structure:

- **Lambda Functions**: `lambda-functions/{function-name}/{fileId}/{filename}`
- **Agents**: `agent-files/{agent-name}/{fileId}/{filename}`

## Technologies

- **Backend**: FastAPI, MinIO Python SDK, Kubernetes Python Client
- **Frontend**: React 18, TypeScript, Vite, react-dropzone
- **Deployment**: Kubernetes, Docker, Nginx
