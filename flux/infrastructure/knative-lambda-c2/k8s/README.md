# Command & Control Center for Knative Lambda Operator

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

## Architecture

### Backend (FastAPI)
- FastAPI server providing REST API
- MinIO integration with presigned URLs
- Supports both `lambda-functions` and `agent-files` buckets

### Frontend (React + TypeScript)
- React 18 with TypeScript
- Vite for fast development and builds
- react-dropzone for file uploads
- Modern UI with gradient design

## Deployment

### Prerequisites

1. MinIO must be running and accessible in the `minio` namespace
2. MinIO credentials secret (`minio-credentials`) must exist in the `minio` namespace
3. Images must be built and pushed to your registry
4. The secret will be automatically copied to `knative-lambda-c2` namespace by the init job

### Build Images

```bash
# Build backend
cd src/c2-backend
docker build -t localhost:5001/knative-lambda-c2-backend:latest .

# Build frontend
cd src/c2-frontend
docker build -t localhost:5001/knative-lambda-c2-frontend:latest .
```

### Deploy to Kubernetes

```bash
kubectl apply -k k8s/c2
```

### Access the UI

The UI is accessible via the Ingress at `c2.lambda.local` (or configure your own domain).

## Configuration

### Environment Variables (Backend)

- `MINIO_ENDPOINT`: MinIO endpoint (default: `minio.minio.svc.cluster.local:9000`)
- `MINIO_ACCESS_KEY`: MinIO access key (from secret)
- `MINIO_SECRET_KEY`: MinIO secret key (from secret)
- `LAMBDA_BUCKET`: Bucket for lambda functions (default: `lambda-functions`)
- `AGENT_BUCKET`: Bucket for agent files (default: `agent-files`)

### Frontend Configuration

Set `VITE_API_BASE` environment variable to override the API base URL (default: `/api/v1`).

## API Endpoints

### `POST /api/v1/files/presigned-url`
Generate a presigned URL for file upload.

**Request:**
```json
{
  "filename": "my-function.zip",
  "mimeType": "application/zip",
  "size": 1024000,
  "target": "lambda",
  "path": "my-function/"  // Optional
}
```

**Response:**
```json
{
  "uploadUrl": "http://minio...",
  "fileId": "uuid",
  "expiresIn": 300,
  "objectPath": "lambda/uuid/my-function.zip"
}
```

### `POST /api/v1/files/{fileId}/complete`
Notify backend that upload is complete.

**Request:**
```json
{
  "fileId": "uuid",
  "objectPath": "lambda/uuid/my-function.zip"
}
```

### `GET /api/v1/files/list?target=lambda&prefix=my-function/`
List files in MinIO bucket.

### `DELETE /api/v1/files/{target}/{path}`
Delete a file from MinIO.

## Usage

### For Chatbot Agents (Learning Materials)

1. **Select "Chatbot Agents"** tab
2. **Choose Agent**: Select the agent from the dropdown (e.g., `agent-bruno`, `agent-medical`)
3. **Upload Learning Materials**: Drag & drop runbooks, documentation, or training files
4. **Monitor Progress**: Watch upload progress in real-time
5. **View Library**: Browse all learning materials for the selected agent
6. **Manage Files**: Delete files as needed

### For Lambda Functions

1. **Select "Lambda Functions"** tab
2. **Optional Path**: Enter a path prefix (e.g., `my-function/`) or leave empty
3. **Upload Files**: Drag & drop files or click to select
4. **Monitor Progress**: Watch upload progress in real-time
5. **Manage Files**: View and delete uploaded files

## File Structure

Files are stored in MinIO with the following structure:

### Chatbot Agents (Learning Materials)
- **Agent Files**: `agent-files/{agent-name}/{fileId}/{filename}`
- Example: `agent-files/agent-bruno/{uuid}/runbook.md`

### Lambda Functions
- **Lambda Files**: `lambda-functions/{path}/{fileId}/{filename}`
- If no path provided: `lambda-functions/lambda/{fileId}/{filename}`

## Supported Agents (Hardcoded)

The following agents are currently supported:
- `agent-bruno` - AI Chatbot for Homepage
- `agent-assistant` - Personal AI Assistant
- `messaging-hub` - Message routing agent
- `agent-voice` - Voice processing agent
- `agent-media` - Media generation agent
- `agent-location` - Location services agent
- `agent-command-center` - Command center agent
- `agent-devsecops` - DevSecOps agent
- `agent-blueteam` - Blue team security agent
- `agent-medical` - Medical assistant agent
- `agent-contracts` - Contract analysis agent
- `agent-store-multibrands` - Store sales assistant
- `agent-rpg-lucca` - RPG character agent
- `agent-rpg-marle` - RPG character agent
- `agent-rpg-robo` - RPG character agent

> **Note**: Agent list is currently hardcoded. Automation for dynamic agent discovery will be added later.

## Development

### Backend

```bash
cd src/c2-backend
pip install -r requirements.txt
uvicorn main:app --reload --host 0.0.0.0 --port 8080
```

### Frontend

```bash
cd src/c2-frontend
npm install
npm run dev
```

## Technologies

- **Backend**: FastAPI, MinIO Python SDK
- **Frontend**: React 18, TypeScript, Vite, react-dropzone
- **Deployment**: Kubernetes, Docker, Nginx
