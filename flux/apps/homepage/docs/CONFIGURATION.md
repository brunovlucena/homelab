# Frontend Configuration Guide

## Environment Variables

Create a `.env` file in the `src/frontend/` directory with the following variables:

### Required Variables

```bash
# MinIO CDN Configuration
# Update this URL to match your MinIO endpoint
VITE_MINIO_CDN_URL=http://minio.minio.svc.cluster.local:9000/homepage-assets/agent-screenshots

# API Configuration
VITE_API_BASE_URL=http://localhost:8080
```

### Optional Variables

```bash
# Feature Flags
VITE_ENABLE_CHATBOT=true
VITE_ENABLE_ANALYTICS=false
```

## Environment-Specific Configuration

### Development (Local)

```bash
VITE_MINIO_CDN_URL=http://localhost:9000/homepage-assets/agent-screenshots
VITE_API_BASE_URL=http://localhost:8080
```

### Kubernetes (Internal)

```bash
VITE_MINIO_CDN_URL=http://minio.minio.svc.cluster.local:9000/homepage-assets/agent-screenshots
VITE_API_BASE_URL=http://backend.homepage.svc.cluster.local:8080
```

### Production (Public CDN)

```bash
VITE_MINIO_CDN_URL=https://cdn.lucena.cloud/homepage-assets/agent-screenshots
VITE_API_BASE_URL=https://api.lucena.cloud
```

## Services Page Screenshot Configuration

The Services page (`src/pages/Services.tsx`) uses the `VITE_MINIO_CDN_URL` to load agent screenshots.

### Screenshot URL Pattern

Screenshots are expected to follow this pattern:
```
${VITE_MINIO_CDN_URL}/<agent-name>.png
```

Example:
```
http://minio.minio.svc.cluster.local:9000/homepage-assets/agent-screenshots/healthcare-agent.png
```

### Updating Screenshot Paths

Edit the `services` array in `Services.tsx`:

```typescript
const services: Service[] = [
  {
    id: 'healthcare',
    screenshot: `${CDN_BASE_URL}/healthcare-agent.png`,
    // ... other properties
  }
]
```

## Building for Production

The environment variables are embedded at build time:

```bash
npm run build
```

Make sure your `.env` file is configured correctly before building.

## Docker Configuration

When using Docker, pass environment variables via the Dockerfile or docker-compose:

```dockerfile
ARG VITE_MINIO_CDN_URL
ARG VITE_API_BASE_URL
ENV VITE_MINIO_CDN_URL=$VITE_MINIO_CDN_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
```

## Troubleshooting

### Screenshots not loading

1. Check the browser console for network errors
2. Verify the MinIO URL is accessible
3. Check CORS configuration on MinIO
4. Ensure bucket permissions allow public read access

### Environment variables not updating

1. Restart the dev server after changing `.env`
2. Clear browser cache
3. Rebuild the project: `npm run build`
