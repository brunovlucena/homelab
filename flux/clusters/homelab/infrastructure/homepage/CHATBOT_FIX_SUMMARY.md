# 🤖 Chatbot Frontend Fix Summary

## Problem
The frontend chatbot was not properly connecting to the Agent-SRE service through the API proxy. The implementation had:
1. Incorrect environment variable reference (`process.env` instead of `import.meta.env`)
2. Bypass logic that tried to connect directly to NodePort in development instead of using the API proxy
3. Tests using legacy mocking patterns

## Solution

### 1. Fixed Chatbot Service (`src/services/chatbot.ts`)

**Before:**
```typescript
const isProduction = process.env.NODE_ENV === 'production'
const apiUrl = process.env.VITE_API_URL || '/api/v1'

// Use API proxy for agent-sre in production, direct NodePort in dev
this.agentBaseUrl = isProduction 
  ? `${apiUrl}/agent-sre`  // Proxy through homepage API
  : 'http://localhost:31081' // Direct NodePort access
```

**After:**
```typescript
// Always use the API proxy path for agent-sre
// In development: Vite proxy forwards /api/* to the API service
// In production: Nginx forwards /api/* to the API service
// API service then proxies to agent-sre service
const apiUrl = import.meta.env.VITE_API_URL || '/api/v1'
this.agentBaseUrl = `${apiUrl}/agent-sre`
```

**Key Changes:**
- ✅ Changed `process.env.VITE_API_URL` to `import.meta.env.VITE_API_URL` (Vite's way)
- ✅ Removed environment-specific logic
- ✅ Now always uses API proxy path: `/api/v1/agent-sre/*`
- ✅ Updated comments to document the proxy flow

### 2. Updated Tests (`src/services/chatbot.test.ts`)

**Changes:**
- ✅ Improved mock setup with shared `mockAxiosInstance`
- ✅ Replaced all `(service as any).client.post` with proper `mockAxiosInstance.post`
- ✅ Replaced all `(service as any).client.get` with proper `mockAxiosInstance.get`
- ✅ Made tests more maintainable and consistent

## Architecture Flow

### Development Mode
```
Frontend (React) 
  → /api/v1/agent-sre/chat
    → Vite Dev Proxy (vite.config.ts)
      → API Service (localhost:8080)
        → Agent-SRE Service Proxy Handler
          → Agent-SRE Service (sre-agent-service)
```

### Production Mode
```
Frontend (React) 
  → /api/v1/agent-sre/chat
    → Nginx (frontend container)
      → API Service (homepage-bruno-site-api:8080)
        → Agent-SRE Service Proxy Handler
          → Agent-SRE Service (sre-agent-service.agent-sre.svc.cluster.local:8080)
```

## Configuration Verification

### Frontend Environment Variables
- **Development**: `VITE_API_URL=/api` (docker-compose.yml)
- **Production**: `VITE_API_URL=/api` (Helm chart)

### API Service Configuration
- **Development**: `AGENT_SRE_URL=http://host.docker.internal:31081` (docker-compose.yml)
- **Production**: `url: "http://sre-agent-service.agent-sre.svc.cluster.local:8080"` (Helm values)

### API Routes (router/router.go)
```go
// Agent-SRE proxy routes
agentSRE := api.Group("/agent-sre")
{
    // Health and status endpoints
    agentSRE.GET("/health", agentSREHandler.Health)
    agentSRE.GET("/ready", agentSREHandler.Ready)
    agentSRE.GET("/status", agentSREHandler.Status)

    // Chat endpoints
    agentSRE.POST("/chat", agentSREHandler.Chat)
    agentSRE.POST("/mcp/chat", agentSREHandler.MCPChat)

    // Log analysis endpoints
    agentSRE.POST("/analyze-logs", agentSREHandler.AnalyzeLogs)
    agentSRE.POST("/mcp/analyze-logs", agentSREHandler.MCPAnalyzeLogs)
}
```

## Available Endpoints

The chatbot now correctly calls:
- ✅ `POST /api/v1/agent-sre/chat` - Direct chat
- ✅ `POST /api/v1/agent-sre/mcp/chat` - MCP chat
- ✅ `POST /api/v1/agent-sre/analyze-logs` - Direct log analysis
- ✅ `POST /api/v1/agent-sre/mcp/analyze-logs` - MCP log analysis
- ✅ `GET /api/v1/agent-sre/health` - Health check
- ✅ `GET /api/v1/agent-sre/ready` - Readiness check
- ✅ `GET /api/v1/agent-sre/status` - Status info

## Testing

Run the tests:
```bash
cd frontend
npm test -- chatbot.test.ts
```

## Benefits

1. **Simplified Architecture**: No more environment-specific connection logic
2. **Consistent Behavior**: Same code path in dev and production
3. **Better Proxy Handling**: All requests go through the API service proxy
4. **Correct Vite Usage**: Using `import.meta.env` instead of `process.env`
5. **Improved Tests**: More maintainable and consistent mocking

## No Legacy Code Found

✅ No legacy chatbot implementations were found
✅ All files are using the latest patterns
✅ No outdated API references

## Next Steps

1. **Test in Development**:
   ```bash
   make start
   ```
   Then open http://localhost:3000 and test the chatbot

2. **Test in Production**: Deploy to cluster and verify chatbot works through the full proxy chain

3. **Monitor Logs**: Check API service logs to see proxy requests flowing through

## Files Changed

- `frontend/src/services/chatbot.ts` - Main service fix
- `frontend/src/services/chatbot.test.ts` - Test improvements
- `CHATBOT_FIX_SUMMARY.md` - This summary document (new)

