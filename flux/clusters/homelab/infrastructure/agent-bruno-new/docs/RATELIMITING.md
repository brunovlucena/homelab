# Rate Limiting - MCP Server & Client

**[← Back to README](../README.md)** | **[RBAC](RBAC.md)** | **[Multi-Tenancy](MULTI-TENANCY.md)** | **[Architecture](ARCHITECTURE.md)**

---

## Overview

This document covers rate limiting strategies for **both inbound and outbound MCP traffic**:

1. **Inbound**: Rate limiting for Agent Bruno's MCP Server (when other agents/services call us)
2. **Outbound**: Rate limiting for Agent Bruno's MCP Client (when we call remote MCP servers)

Rate limiting protects services from abuse, ensures fair resource allocation, maintains system stability under load, and prevents violating external API quotas.

## Table of Contents

### Inbound Rate Limiting (Agent Bruno MCP Server)
1. [Why Rate Limiting?](#why-rate-limiting)
2. [Rate Limiting Layers](#rate-limiting-layers)
3. [Pattern 1: Local Access](#pattern-1-local-access-minimal-rate-limiting)
4. [Pattern 2: Remote Access](#pattern-2-remote-access-full-rate-limiting)
5. [Pattern 3: Multi-Tenancy](#pattern-3-multi-tenancy-advanced-rate-limiting)

### Outbound Rate Limiting (MCP Client to Remote Servers)
6. [Remote MCP Server Rate Limiting](#remote-mcp-server-rate-limiting)
7. [Client-Side Rate Limiting Strategy](#client-side-rate-limiting-strategy)
8. [Quota Management](#quota-management)

### Implementation & Operations
9. [Implementation Details](#implementation-details)
10. [Monitoring & Alerting](#monitoring--alerting)
11. [Testing](#testing)
12. [Troubleshooting](#troubleshooting)

---

## Why Rate Limiting?

### Protection Against
- **Abuse**: Malicious clients overwhelming the service
- **Bugs**: Runaway clients in infinite loops
- **Cost**: Excessive LLM token usage (Ollama calls)
- **Resource Exhaustion**: CPU, memory, and database overload
- **Noisy Neighbors**: One client affecting others (multi-tenant)

### Benefits
- ✅ Predictable service performance
- ✅ Fair resource allocation
- ✅ Cost control (especially important for LLM calls)
- ✅ Graceful degradation under load
- ✅ Security posture improvement

---

## Rate Limiting Layers

Agent Bruno implements defense-in-depth with multiple rate limiting layers:

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: Cloudflare (Global, Remote only)                 │
│  - Rate: 1000 req/min globally                             │
│  - Purpose: DDoS protection, malicious traffic             │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Ingress/Service Mesh (Remote only)               │
│  - Rate: 500 req/min per IP                                │
│  - Purpose: Application-level protection                    │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: Application (Per-Client)                         │
│  - Rate: 100 req/min per API key                           │
│  - Purpose: Fair allocation, client quotas                  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: Tool-Specific (Per-Operation)                    │
│  - Rate: Varies by tool (e.g., 10 search/min)              │
│  - Purpose: Protect expensive operations                    │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 5: Resource-Based (LLM Tokens)                      │
│  - Rate: 100k tokens/hour per client                       │
│  - Purpose: Cost control for LLM inference                  │
└─────────────────────────────────────────────────────────────┘
```

---

## Pattern 1: Local Access (Minimal Rate Limiting)

### Strategy
For local access via `kubectl port-forward`, rate limiting is **optional** since:
- Access is already controlled by Kubernetes RBAC
- Limited to authenticated kubectl users
- Typically used for development/testing
- No cost concerns (internal LLM server)

### Recommended Configuration

```yaml
# config/rate-limits-local.yaml
rate_limiting:
  enabled: false  # Optional for local deployments
  
  # Optional: Basic protection against runaway scripts
  basic_protection:
    enabled: true
    max_requests_per_minute: 1000  # Very high threshold
    max_concurrent_requests: 50
    
  # Tool-specific limits (always recommended)
  tool_limits:
    search_runbooks:
      max_per_minute: 60
      max_concurrent: 10
    
    query_metrics:
      max_per_minute: 30
      max_concurrent: 5
    
    llm_generate:
      max_per_minute: 20
      max_concurrent: 3
      max_tokens_per_hour: 500000  # 500k tokens/hour
```

### Implementation (Optional Protection)

```python
# agent_bruno/middleware/rate_limit_local.py
from fastapi import Request, HTTPException
from collections import defaultdict
from datetime import datetime, timedelta
import asyncio

class BasicRateLimiter:
    """Simple in-memory rate limiter for local deployments."""
    
    def __init__(self, max_requests: int = 1000, window_seconds: int = 60):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self.requests = defaultdict(list)
        self._lock = asyncio.Lock()
    
    async def check_limit(self, client_id: str = "local") -> bool:
        """Check if request is within rate limit."""
        async with self._lock:
            now = datetime.utcnow()
            window_start = now - timedelta(seconds=self.window_seconds)
            
            # Clean old requests
            self.requests[client_id] = [
                req_time for req_time in self.requests[client_id]
                if req_time > window_start
            ]
            
            # Check limit
            if len(self.requests[client_id]) >= self.max_requests:
                return False
            
            # Record request
            self.requests[client_id].append(now)
            return True
    
    async def get_remaining(self, client_id: str = "local") -> int:
        """Get remaining requests in current window."""
        async with self._lock:
            now = datetime.utcnow()
            window_start = now - timedelta(seconds=self.window_seconds)
            
            self.requests[client_id] = [
                req_time for req_time in self.requests[client_id]
                if req_time > window_start
            ]
            
            return max(0, self.max_requests - len(self.requests[client_id]))


# FastAPI middleware
from fastapi import FastAPI

app = FastAPI()
rate_limiter = BasicRateLimiter(max_requests=1000, window_seconds=60)

@app.middleware("http")
async def rate_limit_middleware(request: Request, call_next):
    """Apply basic rate limiting to all requests."""
    if not await rate_limiter.check_limit():
        raise HTTPException(
            status_code=429,
            detail="Rate limit exceeded (local protection)",
            headers={"Retry-After": "60"}
        )
    
    response = await call_next(request)
    
    # Add rate limit headers
    remaining = await rate_limiter.get_remaining()
    response.headers["X-RateLimit-Limit"] = "1000"
    response.headers["X-RateLimit-Remaining"] = str(remaining)
    response.headers["X-RateLimit-Reset"] = str(60)
    
    return response
```

---

## Pattern 2: Remote Access (Full Rate Limiting)

### Strategy
For remote access via internet, implement **comprehensive multi-layer rate limiting**:
- Cloudflare for DDoS and global limits
- Application-level per-client quotas
- Tool-specific limits for expensive operations
- Token-based limits for LLM cost control

### Architecture

```
External Client (API Key: client-a)
    │
    ▼
┌─────────────────────────────────────────────┐
│  Cloudflare WAF                             │
│  - 1000 req/min globally                    │
│  - 500 req/min per IP                       │
│  - Challenge on suspicious behavior         │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│  Knative Service (agent-mcp-server)         │
│  ┌───────────────────────────────────────┐  │
│  │  Redis-based Rate Limiter             │  │
│  │  - 100 req/min per API key            │  │
│  │  - 10 concurrent requests per client  │  │
│  │  - Token bucket algorithm             │  │
│  └───────────────────────────────────────┘  │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│  Tool Router                                │
│  - Tool-specific limits                     │
│  - Cost tracking (LLM tokens)               │
└─────────────────────────────────────────────┘
```

### Configuration

```yaml
# config/rate-limits-remote.yaml
rate_limiting:
  enabled: true
  
  # Redis for distributed rate limiting (multiple replicas)
  redis:
    enabled: true
    host: "redis.agent-bruno.svc.cluster.local"
    port: 6379
    db: 0
    password_secret: "redis-password"
  
  # Global limits per client (API key)
  client_limits:
    default:
      requests_per_minute: 100
      requests_per_hour: 5000
      requests_per_day: 50000
      concurrent_requests: 10
      
    # Premium tier (future)
    premium:
      requests_per_minute: 500
      requests_per_hour: 25000
      requests_per_day: 250000
      concurrent_requests: 50
  
  # Tool-specific limits (applied after client limits)
  tool_limits:
    search_runbooks:
      requests_per_minute: 30
      max_concurrent: 5
      cost_weight: 1  # Low cost
    
    query_metrics:
      requests_per_minute: 20
      max_concurrent: 3
      cost_weight: 2  # Medium cost (Prometheus query)
    
    llm_generate:
      requests_per_minute: 10
      max_concurrent: 2
      cost_weight: 10  # High cost (LLM inference)
      max_tokens_per_request: 4096
      max_tokens_per_hour: 100000
  
  # Burst allowance (token bucket)
  burst:
    enabled: true
    multiplier: 1.5  # Allow 150% of rate for short bursts
    
  # Response headers
  include_headers: true  # X-RateLimit-* headers
  
  # Penalties for abuse
  penalties:
    enabled: true
    threshold_multiplier: 3  # 3x over limit triggers penalty
    penalty_duration_seconds: 300  # 5 minute timeout
```

### Implementation (Redis-backed)

```python
# agent_bruno/middleware/rate_limit_remote.py
from fastapi import Request, HTTPException, Response
from redis import asyncio as aioredis
from datetime import datetime, timedelta
import json
from typing import Optional

class RedisRateLimiter:
    """Distributed rate limiter using Redis."""
    
    def __init__(
        self,
        redis_url: str,
        limits: dict,
        tool_limits: dict
    ):
        self.redis = aioredis.from_url(redis_url)
        self.limits = limits
        self.tool_limits = tool_limits
    
    async def check_limit(
        self,
        client_id: str,
        tool_name: Optional[str] = None
    ) -> tuple[bool, dict]:
        """
        Check rate limits for client and optionally tool.
        
        Returns:
            (allowed, metadata) where metadata contains rate limit info
        """
        now = datetime.utcnow()
        
        # Get client tier (default or premium)
        tier = await self._get_client_tier(client_id)
        client_limits = self.limits.get(tier, self.limits["default"])
        
        # Check minute limit (most common)
        minute_key = f"ratelimit:{client_id}:minute:{now.strftime('%Y%m%d%H%M')}"
        minute_count = await self.redis.incr(minute_key)
        
        if minute_count == 1:
            await self.redis.expire(minute_key, 60)
        
        if minute_count > client_limits["requests_per_minute"]:
            return False, {
                "limit": client_limits["requests_per_minute"],
                "remaining": 0,
                "reset": 60 - now.second,
                "retry_after": 60 - now.second
            }
        
        # Check hourly limit
        hour_key = f"ratelimit:{client_id}:hour:{now.strftime('%Y%m%d%H')}"
        hour_count = await self.redis.incr(hour_key)
        
        if hour_count == 1:
            await self.redis.expire(hour_key, 3600)
        
        if hour_count > client_limits["requests_per_hour"]:
            return False, {
                "limit": client_limits["requests_per_hour"],
                "remaining": 0,
                "reset": 3600 - (now.minute * 60 + now.second),
                "retry_after": 3600 - (now.minute * 60 + now.second)
            }
        
        # Check concurrent requests
        concurrent_key = f"ratelimit:{client_id}:concurrent"
        concurrent = await self.redis.get(concurrent_key)
        concurrent = int(concurrent) if concurrent else 0
        
        if concurrent >= client_limits["concurrent_requests"]:
            return False, {
                "limit": client_limits["concurrent_requests"],
                "remaining": 0,
                "reset": 0,
                "retry_after": 5,
                "error": "Too many concurrent requests"
            }
        
        # Check tool-specific limits if applicable
        if tool_name and tool_name in self.tool_limits:
            tool_limit = self.tool_limits[tool_name]
            tool_key = f"ratelimit:{client_id}:tool:{tool_name}:{now.strftime('%Y%m%d%H%M')}"
            tool_count = await self.redis.incr(tool_key)
            
            if tool_count == 1:
                await self.redis.expire(tool_key, 60)
            
            if tool_count > tool_limit["requests_per_minute"]:
                return False, {
                    "limit": tool_limit["requests_per_minute"],
                    "remaining": 0,
                    "reset": 60 - now.second,
                    "retry_after": 60 - now.second,
                    "error": f"Tool rate limit exceeded: {tool_name}"
                }
        
        # All checks passed
        metadata = {
            "limit": client_limits["requests_per_minute"],
            "remaining": client_limits["requests_per_minute"] - minute_count,
            "reset": 60 - now.second
        }
        
        return True, metadata
    
    async def increment_concurrent(self, client_id: str):
        """Increment concurrent request counter."""
        key = f"ratelimit:{client_id}:concurrent"
        await self.redis.incr(key)
        await self.redis.expire(key, 300)  # Auto-cleanup after 5 min
    
    async def decrement_concurrent(self, client_id: str):
        """Decrement concurrent request counter."""
        key = f"ratelimit:{client_id}:concurrent"
        await self.redis.decr(key)
    
    async def track_tokens(self, client_id: str, tokens: int) -> bool:
        """
        Track LLM token usage.
        
        Returns:
            True if under limit, False if over limit
        """
        now = datetime.utcnow()
        hour_key = f"tokens:{client_id}:hour:{now.strftime('%Y%m%d%H')}"
        
        total_tokens = await self.redis.incrby(hour_key, tokens)
        
        if total_tokens == tokens:
            await self.redis.expire(hour_key, 3600)
        
        # Check against limit (default: 100k tokens/hour)
        tier = await self._get_client_tier(client_id)
        tool_limits = self.tool_limits.get("llm_generate", {})
        max_tokens = tool_limits.get("max_tokens_per_hour", 100000)
        
        return total_tokens <= max_tokens
    
    async def _get_client_tier(self, client_id: str) -> str:
        """Get client tier (default or premium)."""
        tier_key = f"client:{client_id}:tier"
        tier = await self.redis.get(tier_key)
        return tier.decode() if tier else "default"


# FastAPI integration
from fastapi import FastAPI, Depends

app = FastAPI()

rate_limiter = RedisRateLimiter(
    redis_url="redis://redis.agent-bruno.svc.cluster.local:6379/0",
    limits={
        "default": {
            "requests_per_minute": 100,
            "requests_per_hour": 5000,
            "concurrent_requests": 10
        }
    },
    tool_limits={
        "llm_generate": {
            "requests_per_minute": 10,
            "max_concurrent": 2
        }
    }
)

async def get_client_id(request: Request) -> str:
    """Extract client ID from API key."""
    auth_header = request.headers.get("Authorization", "")
    if auth_header.startswith("Bearer "):
        api_key = auth_header[7:]
        # Validate and map API key to client ID
        # (implementation depends on your auth system)
        return api_key[:16]  # Simplified
    return "anonymous"

@app.middleware("http")
async def rate_limit_middleware(request: Request, call_next):
    """Apply rate limiting to all requests."""
    client_id = await get_client_id(request)
    
    # Extract tool name from path if present
    tool_name = None
    if "/mcp/tools/" in request.url.path:
        tool_name = request.url.path.split("/")[-1]
    
    # Check rate limit
    allowed, metadata = await rate_limiter.check_limit(client_id, tool_name)
    
    if not allowed:
        raise HTTPException(
            status_code=429,
            detail=metadata.get("error", "Rate limit exceeded"),
            headers={
                "Retry-After": str(metadata.get("retry_after", 60)),
                "X-RateLimit-Limit": str(metadata["limit"]),
                "X-RateLimit-Remaining": "0",
                "X-RateLimit-Reset": str(metadata["reset"])
            }
        )
    
    # Track concurrent request
    await rate_limiter.increment_concurrent(client_id)
    
    try:
        response = await call_next(request)
    finally:
        await rate_limiter.decrement_concurrent(client_id)
    
    # Add rate limit headers
    response.headers["X-RateLimit-Limit"] = str(metadata["limit"])
    response.headers["X-RateLimit-Remaining"] = str(metadata["remaining"])
    response.headers["X-RateLimit-Reset"] = str(metadata["reset"])
    
    return response
```

### Cloudflare Configuration

```javascript
// Cloudflare Worker for rate limiting
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
  const ip = request.headers.get('CF-Connecting-IP')
  
  // Global rate limit: 1000 req/min
  const globalKey = `rate_limit:global:${Math.floor(Date.now() / 60000)}`
  const globalCount = await incrementCounter(globalKey, 60)
  
  if (globalCount > 1000) {
    return new Response('Global rate limit exceeded', {
      status: 429,
      headers: {
        'Retry-After': '60',
        'Content-Type': 'text/plain'
      }
    })
  }
  
  // Per-IP rate limit: 500 req/min
  const ipKey = `rate_limit:ip:${ip}:${Math.floor(Date.now() / 60000)}`
  const ipCount = await incrementCounter(ipKey, 60)
  
  if (ipCount > 500) {
    return new Response('IP rate limit exceeded', {
      status: 429,
      headers: {
        'Retry-After': '60',
        'Content-Type': 'text/plain'
      }
    })
  }
  
  // Forward to backend
  return fetch(request)
}

async function incrementCounter(key, ttl) {
  // Use Cloudflare KV or Durable Objects for distributed counter
  // Implementation details depend on your Cloudflare setup
}
```

---

## Pattern 3: Multi-Tenancy (Advanced Rate Limiting)

### Strategy
For Kamaji-based multi-tenancy, implement **per-tenant quotas** with hierarchy:
- Management cluster enforces global limits
- Each tenant control plane enforces tenant-specific limits
- Resource quotas ensure tenant isolation

### Kubernetes Resource Quotas

```yaml
# Per-tenant resource quota (enforced by Kamaji)
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-a-quota
  namespace: tenant-a
spec:
  hard:
    # Compute resources
    requests.cpu: "16"
    requests.memory: "32Gi"
    limits.cpu: "32"
    limits.memory: "64Gi"
    
    # Storage
    requests.storage: "100Gi"
    persistentvolumeclaims: "10"
    
    # Services
    services.loadbalancers: "2"
    services.nodeports: "0"
    
    # Pods and deployments
    pods: "50"
    count/deployments.apps: "20"
    count/jobs.batch: "10"
```

### Tenant-Level Rate Limiting

```python
# agent_bruno/middleware/rate_limit_tenant.py
class TenantRateLimiter:
    """Multi-tenant rate limiter with hierarchical quotas."""
    
    def __init__(self, redis_url: str, tenant_quotas: dict):
        self.redis = aioredis.from_url(redis_url)
        self.tenant_quotas = tenant_quotas
    
    async def check_limit(
        self,
        tenant_id: str,
        client_id: str,
        tool_name: Optional[str] = None
    ) -> tuple[bool, dict]:
        """
        Check rate limits at both tenant and client level.
        
        Hierarchy:
        1. Tenant-level quota (aggregate of all clients in tenant)
        2. Client-level quota (individual client within tenant)
        3. Tool-level quota (if applicable)
        """
        now = datetime.utcnow()
        
        # 1. Check tenant-level quota
        tenant_quota = self.tenant_quotas.get(tenant_id, self.tenant_quotas["default"])
        tenant_key = f"ratelimit:tenant:{tenant_id}:minute:{now.strftime('%Y%m%d%H%M')}"
        tenant_count = await self.redis.incr(tenant_key)
        
        if tenant_count == 1:
            await self.redis.expire(tenant_key, 60)
        
        if tenant_count > tenant_quota["requests_per_minute"]:
            return False, {
                "error": "Tenant rate limit exceeded",
                "limit": tenant_quota["requests_per_minute"],
                "remaining": 0,
                "reset": 60 - now.second,
                "retry_after": 60 - now.second,
                "level": "tenant"
            }
        
        # 2. Check client-level quota within tenant
        client_quota = tenant_quota["per_client"]
        client_key = f"ratelimit:tenant:{tenant_id}:client:{client_id}:minute:{now.strftime('%Y%m%d%H%M')}"
        client_count = await self.redis.incr(client_key)
        
        if client_count == 1:
            await self.redis.expire(client_key, 60)
        
        if client_count > client_quota["requests_per_minute"]:
            return False, {
                "error": "Client rate limit exceeded",
                "limit": client_quota["requests_per_minute"],
                "remaining": 0,
                "reset": 60 - now.second,
                "retry_after": 60 - now.second,
                "level": "client"
            }
        
        # 3. Check tool-specific limits (if applicable)
        # ... similar to remote pattern
        
        return True, {
            "tenant_limit": tenant_quota["requests_per_minute"],
            "tenant_remaining": tenant_quota["requests_per_minute"] - tenant_count,
            "client_limit": client_quota["requests_per_minute"],
            "client_remaining": client_quota["requests_per_minute"] - client_count,
            "reset": 60 - now.second
        }
```

### Tenant Quota Configuration

```yaml
# config/tenant-quotas.yaml
tenant_quotas:
  # Default quota for new tenants
  default:
    requests_per_minute: 500
    requests_per_hour: 25000
    requests_per_day: 250000
    per_client:
      requests_per_minute: 100
      requests_per_hour: 5000
  
  # Tenant A: Premium tier
  tenant-a:
    requests_per_minute: 2000
    requests_per_hour: 100000
    requests_per_day: 1000000
    per_client:
      requests_per_minute: 500
      requests_per_hour: 25000
  
  # Tenant B: Standard tier
  tenant-b:
    requests_per_minute: 1000
    requests_per_hour: 50000
    requests_per_day: 500000
    per_client:
      requests_per_minute: 200
      requests_per_hour: 10000
```

---

## Remote MCP Server Rate Limiting

When Agent Bruno acts as an **MCP client** calling external MCP servers (GitHub, Grafana, etc.), we need to respect **their** rate limits and implement client-side throttling.

### MCP Servers Agent Bruno Calls

```
┌────────────────────────────────────────────────────────────────┐
│              Agent Bruno (MCP Client)                          │
└────────────────────────────────────────────────────────────────┘
         │                │                │
         ├────────────────┼────────────────┤
         │                │                │
         ▼                ▼                ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│ GitHub MCP     │ │ Grafana MCP    │ │ Custom MCP     │
│ (External)     │ │ (External)     │ │ (Internal)     │
│                │ │                │ │                │
│ Rate Limits:   │ │ Rate Limits:   │ │ Rate Limits:   │
│ 5000 req/hour  │ │ 100 req/min    │ │ No limits      │
│ (per API key)  │ │ (per org)      │ │                │
└────────────────┘ └────────────────┘ └────────────────┘
```

### Common Remote MCP Server Limits

| MCP Server | Provider | Rate Limit | Quota Reset | Notes |
|------------|----------|------------|-------------|-------|
| **GitHub MCP** | GitHub API | 5,000 req/hour | Hourly | Per OAuth token |
| **Grafana MCP** | Grafana Cloud | 100 req/min | Minute | Per org/API key |
| **Slack MCP** | Slack API | 50 req/min | Minute | Per workspace |
| **Custom Internal MCP** | Self-hosted | Configurable | N/A | Set by us |

### Consequences of Exceeding Remote Limits

**GitHub API**:
```http
HTTP/1.1 403 Forbidden
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1635724800

{
  "message": "API rate limit exceeded",
  "documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting"
}
```

**Impact**:
- ❌ Request fails immediately
- ❌ Subsequent requests blocked until reset
- ❌ Agent Bruno workflows disrupted
- ❌ User experience degraded
- ⚠️ Potential API key suspension (if repeated)

---

## Client-Side Rate Limiting Strategy

Implement **client-side rate limiting** to stay well under remote MCP server limits.

### Architecture: MCP Client Rate Limiter

```python
# agent-bruno/mcp/rate_limiter.py
from datetime import datetime, timedelta
from typing import Dict, Optional
import asyncio
import logging

logger = logging.getLogger(__name__)

class MCPClientRateLimiter:
    """Client-side rate limiter for remote MCP servers."""
    
    def __init__(self):
        # Per-server rate limit tracking
        self.limits: Dict[str, ServerLimit] = {
            "github": ServerLimit(
                requests_per_hour=4500,  # 90% of GitHub's 5000 limit
                requests_per_minute=75,  # Safety margin
                burst_size=20
            ),
            "grafana": ServerLimit(
                requests_per_minute=90,  # 90% of Grafana's 100 limit
                burst_size=10
            ),
            "slack": ServerLimit(
                requests_per_minute=45,  # 90% of Slack's 50 limit
                burst_size=5
            ),
        }
    
    async def acquire(self, server: str, endpoint: str) -> bool:
        """
        Acquire permission to make request to remote MCP server.
        Returns True if allowed, False if rate limited.
        """
        limit = self.limits.get(server)
        if not limit:
            # No rate limit configured - allow
            return True
        
        # Check if we'd exceed limits
        if not limit.can_make_request():
            # Log rate limit hit
            logger.warning(
                "Client-side rate limit hit",
                extra={
                    "server": server,
                    "endpoint": endpoint,
                    "remaining_tokens": limit.remaining_tokens(),
                    "reset_time": limit.reset_time().isoformat()
                }
            )
            return False
        
        # Consume a token
        limit.consume()
        return True
    
    async def wait_if_needed(self, server: str, endpoint: str):
        """Wait until rate limit allows request (blocking)."""
        limit = self.limits.get(server)
        if not limit:
            return
        
        while not limit.can_make_request():
            wait_time = limit.time_until_reset()
            logger.info(
                f"Waiting {wait_time}s for {server} rate limit reset",
                extra={"server": server, "endpoint": endpoint}
            )
            await asyncio.sleep(min(wait_time, 1.0))  # Check every second
        
        limit.consume()

class ServerLimit:
    """Rate limit tracker for a specific server."""
    
    def __init__(
        self,
        requests_per_hour: Optional[int] = None,
        requests_per_minute: Optional[int] = None,
        burst_size: int = 10
    ):
        self.requests_per_hour = requests_per_hour
        self.requests_per_minute = requests_per_minute
        self.burst_size = burst_size
        
        # Token bucket algorithm
        self.tokens = burst_size
        self.max_tokens = burst_size
        self.last_refill = datetime.utcnow()
    
    def can_make_request(self) -> bool:
        """Check if we can make a request without exceeding limits."""
        self._refill()
        return self.tokens >= 1
    
    def consume(self):
        """Consume a token for making a request."""
        self._refill()
        if self.tokens >= 1:
            self.tokens -= 1
    
    def remaining_tokens(self) -> int:
        """Get number of remaining tokens."""
        self._refill()
        return int(self.tokens)
    
    def time_until_reset(self) -> float:
        """Time in seconds until next token refill."""
        if self.requests_per_minute:
            # Tokens refill per minute
            refill_rate = self.requests_per_minute / 60.0
        elif self.requests_per_hour:
            # Tokens refill per hour
            refill_rate = self.requests_per_hour / 3600.0
        else:
            return 0.0
        
        time_for_one_token = 1.0 / refill_rate
        return time_for_one_token
    
    def reset_time(self) -> datetime:
        """When tokens will fully replenish."""
        wait_time = (self.max_tokens - self.tokens) * self.time_until_reset()
        return datetime.utcnow() + timedelta(seconds=wait_time)
    
    def _refill(self):
        """Refill tokens based on elapsed time (token bucket)."""
        now = datetime.utcnow()
        elapsed = (now - self.last_refill).total_seconds()
        
        if self.requests_per_minute:
            # Refill based on requests per minute
            refill_rate = self.requests_per_minute / 60.0
        elif self.requests_per_hour:
            # Refill based on requests per hour
            refill_rate = self.requests_per_hour / 3600.0
        else:
            return
        
        # Add tokens based on time elapsed
        new_tokens = elapsed * refill_rate
        self.tokens = min(self.max_tokens, self.tokens + new_tokens)
        self.last_refill = now
```

### Usage in MCP Client

```python
# agent-bruno/mcp/client.py
from mcp import ClientSession
from .rate_limiter import MCPClientRateLimiter

class MCPClient:
    """MCP client with built-in rate limiting."""
    
    def __init__(self):
        self.rate_limiter = MCPClientRateLimiter()
        self.sessions: Dict[str, ClientSession] = {}
    
    async def call_tool(
        self,
        server: str,
        tool_name: str,
        arguments: dict
    ):
        """Call MCP tool with rate limiting."""
        
        # Check rate limit BEFORE making request
        if not await self.rate_limiter.acquire(server, tool_name):
            raise RateLimitError(
                f"Client-side rate limit exceeded for {server}",
                server=server,
                tool=tool_name
            )
        
        # Make the actual MCP call
        session = self.sessions[server]
        
        try:
            result = await session.call_tool(
                name=tool_name,
                arguments=arguments
            )
            return result
            
        except Exception as e:
            # Check if remote server rate limited us
            if self._is_rate_limit_error(e):
                logger.error(
                    "Remote MCP server rate limited us!",
                    extra={
                        "server": server,
                        "tool": tool_name,
                        "error": str(e)
                    }
                )
                # TODO: Adjust client-side limits dynamically
            raise
    
    async def call_tool_with_retry(
        self,
        server: str,
        tool_name: str,
        arguments: dict,
        max_retries: int = 3
    ):
        """Call MCP tool with automatic retry on rate limit."""
        
        for attempt in range(max_retries):
            # Wait if rate limited
            await self.rate_limiter.wait_if_needed(server, tool_name)
            
            try:
                return await self.call_tool(server, tool_name, arguments)
            except RateLimitError:
                if attempt == max_retries - 1:
                    raise
                # Exponential backoff
                wait_time = 2 ** attempt
                logger.info(f"Rate limited, retrying in {wait_time}s")
                await asyncio.sleep(wait_time)
    
    def _is_rate_limit_error(self, error: Exception) -> bool:
        """Check if error is due to remote rate limiting."""
        error_str = str(error).lower()
        return any([
            "rate limit" in error_str,
            "429" in error_str,
            "too many requests" in error_str,
            "quota exceeded" in error_str,
        ])
```

### Adaptive Rate Limiting

Dynamically adjust client-side limits based on remote server responses:

```python
# agent-bruno/mcp/adaptive_limiter.py
class AdaptiveMCPRateLimiter(MCPClientRateLimiter):
    """Rate limiter that adapts to remote server responses."""
    
    def handle_response_headers(self, server: str, headers: dict):
        """Adjust limits based on remote server rate limit headers."""
        
        # GitHub-style headers
        if "X-RateLimit-Remaining" in headers:
            remaining = int(headers["X-RateLimit-Remaining"])
            limit = int(headers.get("X-RateLimit-Limit", 5000))
            reset_time = int(headers.get("X-RateLimit-Reset", 0))
            
            # If we're getting close to limit, reduce our rate
            usage_ratio = 1 - (remaining / limit)
            
            if usage_ratio > 0.9:
                # Reduce to 50% of normal rate
                self.limits[server].reduce_rate(0.5)
                logger.warning(
                    f"Reducing {server} rate limit to 50%",
                    extra={
                        "remaining": remaining,
                        "limit": limit,
                        "usage_ratio": usage_ratio
                    }
                )
            elif usage_ratio > 0.7:
                # Reduce to 75% of normal rate
                self.limits[server].reduce_rate(0.75)
        
        # Retry-After header (429 response)
        if "Retry-After" in headers:
            retry_after = int(headers["Retry-After"])
            logger.warning(
                f"Server requested retry after {retry_after}s",
                extra={"server": server}
            )
            # Pause requests for this duration
            self.limits[server].pause_until(
                datetime.utcnow() + timedelta(seconds=retry_after)
            )
```

---

## Quota Management

Track and manage API quota usage across different time windows.

### Quota Tracking

```python
# agent-bruno/mcp/quota_tracker.py
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Dict
import logging

logger = logging.getLogger(__name__)

@dataclass
class QuotaWindow:
    """Track quota usage in a time window."""
    window_size: timedelta
    max_requests: int
    requests: list[datetime]  # Sliding window of request times
    
    def can_make_request(self) -> bool:
        """Check if quota allows another request."""
        self._cleanup_old_requests()
        return len(self.requests) < self.max_requests
    
    def record_request(self):
        """Record a new request."""
        self._cleanup_old_requests()
        self.requests.append(datetime.utcnow())
    
    def remaining_quota(self) -> int:
        """Remaining requests in quota."""
        self._cleanup_old_requests()
        return max(0, self.max_requests - len(self.requests))
    
    def _cleanup_old_requests(self):
        """Remove requests outside the current window."""
        cutoff = datetime.utcnow() - self.window_size
        self.requests = [r for r in self.requests if r > cutoff]

class QuotaManager:
    """Manage API quotas for remote MCP servers."""
    
    def __init__(self):
        self.quotas: Dict[str, Dict[str, QuotaWindow]] = {
            "github": {
                "hourly": QuotaWindow(
                    window_size=timedelta(hours=1),
                    max_requests=4500,  # 90% of GitHub's 5000
                    requests=[]
                ),
                "daily": QuotaWindow(
                    window_size=timedelta(days=1),
                    max_requests=100000,  # Daily soft limit
                    requests=[]
                ),
            },
            "grafana": {
                "minute": QuotaWindow(
                    window_size=timedelta(minutes=1),
                    max_requests=90,  # 90% of Grafana's 100
                    requests=[]
                ),
                "hourly": QuotaWindow(
                    window_size=timedelta(hours=1),
                    max_requests=5000,
                    requests=[]
                ),
            },
        }
    
    def check_quota(self, server: str) -> bool:
        """Check if all quotas allow request."""
        if server not in self.quotas:
            return True  # No quota defined
        
        for window_name, window in self.quotas[server].items():
            if not window.can_make_request():
                logger.warning(
                    f"Quota exceeded for {server}",
                    extra={
                        "server": server,
                        "window": window_name,
                        "remaining": window.remaining_quota()
                    }
                )
                return False
        return True
    
    def record_request(self, server: str):
        """Record request in all quota windows."""
        if server in self.quotas:
            for window in self.quotas[server].values():
                window.record_request()
    
    def get_quota_status(self, server: str) -> dict:
        """Get current quota status for server."""
        if server not in self.quotas:
            return {}
        
        return {
            window_name: {
                "remaining": window.remaining_quota(),
                "max": window.max_requests,
                "window": str(window.window_size)
            }
            for window_name, window in self.quotas[server].items()
        }
```

### Prometheus Metrics for Quota Tracking

```python
# agent-bruno/mcp/metrics.py
from prometheus_client import Gauge, Counter

# Quota remaining for each MCP server
mcp_quota_remaining = Gauge(
    'mcp_client_quota_remaining',
    'Remaining quota for remote MCP server',
    ['server', 'window']
)

# Quota usage counter
mcp_quota_used = Counter(
    'mcp_client_quota_used_total',
    'Total quota consumed for remote MCP server',
    ['server', 'endpoint']
)

# Quota exceeded events
mcp_quota_exceeded = Counter(
    'mcp_client_quota_exceeded_total',
    'Times client-side quota was exceeded',
    ['server', 'window']
)

# Update metrics periodically
async def update_quota_metrics(quota_manager: QuotaManager):
    """Update Prometheus metrics for quota status."""
    while True:
        for server in ["github", "grafana", "slack"]:
            status = quota_manager.get_quota_status(server)
            for window, stats in status.items():
                mcp_quota_remaining.labels(
                    server=server,
                    window=window
                ).set(stats["remaining"])
        
        await asyncio.sleep(10)  # Update every 10 seconds
```

### Grafana Dashboard for MCP Client Quotas

```json
{
  "dashboard": {
    "title": "MCP Client - Remote Server Quotas",
    "panels": [
      {
        "title": "GitHub API Quota Remaining",
        "type": "gauge",
        "targets": [{
          "expr": "mcp_client_quota_remaining{server=\"github\", window=\"hourly\"}",
          "legendFormat": "Hourly Quota"
        }],
        "thresholds": [
          {"value": 0, "color": "red"},
          {"value": 1000, "color": "yellow"},
          {"value": 2000, "color": "green"}
        ]
      },
      {
        "title": "MCP Client Request Rate by Server",
        "type": "graph",
        "targets": [{
          "expr": "rate(mcp_quota_used_total[5m])",
          "legendFormat": "{{server}} - {{endpoint}}"
        }]
      },
      {
        "title": "Quota Exceeded Events",
        "type": "graph",
        "targets": [{
          "expr": "increase(mcp_quota_exceeded_total[1h])",
          "legendFormat": "{{server}} - {{window}}"
        }]
      }
    ]
  }
}
```

---

## Implementation Details

### Metrics

```promql
# Rate limit hit rate (should be < 1%)
sum(rate(rate_limit_exceeded_total[5m])) by (client_id)
/
sum(rate(http_requests_total[5m])) by (client_id)

# Top rate-limited clients
topk(10, sum(rate(rate_limit_exceeded_total[5m])) by (client_id))

# Rate limit violations by tier
sum(rate(rate_limit_exceeded_total[5m])) by (tier, limit_type)

# Token usage per client (cost tracking)
sum(llm_tokens_used_total) by (client_id, hour)

# Concurrent requests per client
max(rate_limit_concurrent_requests) by (client_id)

# Redis rate limiter latency
histogram_quantile(0.95, 
  rate(rate_limit_check_duration_seconds_bucket[5m])
)
```

### Grafana Dashboard

```yaml
# dashboards/rate-limiting.json
{
  "dashboard": {
    "title": "Agent Bruno - Rate Limiting",
    "panels": [
      {
        "title": "Rate Limit Hit Rate",
        "targets": [{
          "expr": "sum(rate(rate_limit_exceeded_total[5m])) / sum(rate(http_requests_total[5m]))"
        }],
        "alert": {
          "conditions": [{
            "evaluator": { "gt": 0.05 },
            "message": "Rate limit hit rate > 5%"
          }]
        }
      },
      {
        "title": "Top Rate-Limited Clients",
        "targets": [{
          "expr": "topk(10, sum(rate(rate_limit_exceeded_total[5m])) by (client_id))"
        }]
      },
      {
        "title": "Token Usage by Client",
        "targets": [{
          "expr": "sum(llm_tokens_used_total) by (client_id)"
        }]
      },
      {
        "title": "Concurrent Requests",
        "targets": [{
          "expr": "max(rate_limit_concurrent_requests) by (client_id)"
        }]
      }
    ]
  }
}
```

### Alerts

```yaml
# prometheus/alerts/rate-limiting.yaml
groups:
- name: rate_limiting
  interval: 30s
  rules:
  
  # High rate limit violation rate
  - alert: HighRateLimitViolationRate
    expr: |
      sum(rate(rate_limit_exceeded_total[5m]))
      /
      sum(rate(http_requests_total[5m]))
      > 0.05
    for: 5m
    labels:
      severity: warning
      component: rate-limiter
    annotations:
      summary: "High rate limit violation rate"
      description: "{{ $value | humanizePercentage }} of requests are being rate limited"
      runbook: "runbooks/agent-bruno/high-rate-limit-violations.md"
  
  # Client exceeding quota consistently
  - alert: ClientConsistentlyRateLimited
    expr: |
      sum(rate(rate_limit_exceeded_total[10m])) by (client_id)
      > 10
    for: 30m
    labels:
      severity: info
      component: rate-limiter
    annotations:
      summary: "Client {{ $labels.client_id }} consistently rate limited"
      description: "Client may need quota increase or has runaway behavior"
      runbook: "runbooks/agent-bruno/client-rate-limited.md"
  
  # Redis rate limiter unavailable
  - alert: RateLimiterRedisDown
    expr: |
      up{job="redis",namespace="agent-bruno"} == 0
    for: 2m
    labels:
      severity: critical
      component: rate-limiter
    annotations:
      summary: "Redis for rate limiting is down"
      description: "Rate limiting may fall back to in-memory (not distributed)"
      runbook: "runbooks/agent-bruno/redis-down.md"
  
  # Token usage approaching limit
  - alert: TokenUsageHigh
    expr: |
      sum(llm_tokens_used_total) by (client_id, hour)
      > 80000
    labels:
      severity: warning
      component: rate-limiter
    annotations:
      summary: "Client {{ $labels.client_id }} approaching token limit"
      description: "80k/100k tokens used in current hour"
      runbook: "runbooks/agent-bruno/high-token-usage.md"
```

---

## Testing

### Unit Tests

```python
# tests/test_rate_limiter.py
import pytest
from datetime import datetime
from agent_bruno.middleware.rate_limit_remote import RedisRateLimiter

@pytest.mark.asyncio
async def test_basic_rate_limit(redis_client):
    """Test basic rate limiting functionality."""
    limiter = RedisRateLimiter(
        redis_url="redis://localhost:6379/15",  # Test DB
        limits={
            "default": {
                "requests_per_minute": 10,
                "requests_per_hour": 100,
                "concurrent_requests": 3
            }
        },
        tool_limits={}
    )
    
    client_id = "test-client"
    
    # First 10 requests should succeed
    for i in range(10):
        allowed, metadata = await limiter.check_limit(client_id)
        assert allowed is True
        assert metadata["remaining"] == 10 - i - 1
    
    # 11th request should be rate limited
    allowed, metadata = await limiter.check_limit(client_id)
    assert allowed is False
    assert metadata["remaining"] == 0
    assert "retry_after" in metadata

@pytest.mark.asyncio
async def test_tool_specific_limit(redis_client):
    """Test tool-specific rate limits."""
    limiter = RedisRateLimiter(
        redis_url="redis://localhost:6379/15",
        limits={
            "default": {
                "requests_per_minute": 100,
                "requests_per_hour": 1000,
                "concurrent_requests": 10
            }
        },
        tool_limits={
            "llm_generate": {
                "requests_per_minute": 5,
                "max_concurrent": 2
            }
        }
    )
    
    client_id = "test-client"
    
    # First 5 LLM requests should succeed
    for i in range(5):
        allowed, metadata = await limiter.check_limit(client_id, "llm_generate")
        assert allowed is True
    
    # 6th LLM request should be rate limited
    allowed, metadata = await limiter.check_limit(client_id, "llm_generate")
    assert allowed is False
    assert "llm_generate" in metadata.get("error", "")

@pytest.mark.asyncio
async def test_concurrent_limit(redis_client):
    """Test concurrent request limiting."""
    limiter = RedisRateLimiter(
        redis_url="redis://localhost:6379/15",
        limits={
            "default": {
                "requests_per_minute": 100,
                "requests_per_hour": 1000,
                "concurrent_requests": 2
            }
        },
        tool_limits={}
    )
    
    client_id = "test-client"
    
    # Start 2 concurrent requests
    await limiter.increment_concurrent(client_id)
    await limiter.increment_concurrent(client_id)
    
    # 3rd concurrent request should be blocked
    allowed, metadata = await limiter.check_limit(client_id)
    assert allowed is False
    assert "concurrent" in metadata.get("error", "")
    
    # Release one
    await limiter.decrement_concurrent(client_id)
    
    # Now should be allowed
    allowed, metadata = await limiter.check_limit(client_id)
    assert allowed is True
```

### Load Tests (k6)

```javascript
// load-tests/rate-limiting.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const rateLimitErrorRate = new Rate('rate_limit_errors');

export let options = {
  stages: [
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 150 },  // Ramp up to 150 users (should hit limits)
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    'rate_limit_errors': ['rate<0.1'],  // Less than 10% rate limited
    'http_req_duration': ['p(95)<2000'], // 95% under 2s
  },
};

export default function() {
  const url = 'https://mcp.bruno.dev/mcp/tools/search_runbooks';
  const params = {
    headers: {
      'Authorization': 'Bearer test-api-key',
      'Content-Type': 'application/json',
    },
  };
  
  const payload = JSON.stringify({
    query: 'loki crashes',
  });
  
  const res = http.post(url, payload, params);
  
  // Check for rate limiting
  const rateLimited = res.status === 429;
  rateLimitErrorRate.add(rateLimited);
  
  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'has rate limit headers': (r) => r.headers['X-Ratelimit-Limit'] !== undefined,
  });
  
  if (rateLimited) {
    const retryAfter = parseInt(res.headers['Retry-After'] || '60');
    console.log(`Rate limited! Retry after ${retryAfter}s`);
    sleep(retryAfter);
  } else {
    sleep(1);
  }
}
```

Run load test:
```bash
k6 run load-tests/rate-limiting.js
```

---

## Troubleshooting

### Issue: Rate limiter not working

```bash
# Check Redis connectivity
kubectl exec -it deployment/agent-mcp-server -n agent-bruno -- \
  redis-cli -h redis.agent-bruno.svc.cluster.local ping

# Check rate limiter logs
kubectl logs -n agent-bruno -l app=agent-mcp-server --tail=100 | grep -i "rate"

# Verify configuration
kubectl get configmap agent-bruno-config -n agent-bruno -o yaml | grep -A 20 rate_limiting
```

### Issue: Client incorrectly rate limited

```bash
# Check Redis keys for client
kubectl exec -it deployment/redis -n agent-bruno -- \
  redis-cli KEYS "ratelimit:client-a:*"

# Check current counts
kubectl exec -it deployment/redis -n agent-bruno -- \
  redis-cli GET "ratelimit:client-a:minute:$(date +%Y%m%d%H%M)"

# Reset client rate limit (emergency)
kubectl exec -it deployment/redis -n agent-bruno -- \
  redis-cli DEL $(redis-cli KEYS "ratelimit:client-a:*")
```

### Issue: Rate limits too aggressive

```bash
# Temporarily increase limits (ConfigMap)
kubectl edit configmap agent-bruno-config -n agent-bruno
# Update rate_limiting.client_limits.default.requests_per_minute

# Restart pods to pick up new config
kubectl rollout restart deployment/agent-mcp-server -n agent-bruno
```

### Issue: Redis memory issues

```bash
# Check Redis memory usage
kubectl exec -it deployment/redis -n agent-bruno -- \
  redis-cli INFO memory

# Check key count
kubectl exec -it deployment/redis -n agent-bruno -- \
  redis-cli DBSIZE

# Evict old keys (if needed)
kubectl exec -it deployment/redis -n agent-bruno -- \
  redis-cli --scan --pattern "ratelimit:*" | \
  xargs -L 1 redis-cli DEL
```

---

## Best Practices

### ✅ DO
- Implement rate limiting at multiple layers (defense in depth)
- Use Redis for distributed rate limiting (multiple replicas)
- Include clear error messages with retry-after headers
- Monitor rate limit violations as a metric
- Alert on unusual patterns (client abuse, system issues)
- Provide different tiers (default, premium) for flexibility
- Track token usage for cost control
- Test rate limits under load (k6, locust)

### ❌ DON'T
- Rely solely on application-layer rate limiting (use Cloudflare too)
- Use in-memory rate limiting for multi-replica deployments
- Return 500 errors for rate limit violations (use 429)
- Set limits too low (causes user frustration)
- Set limits too high (allows abuse)
- Forget to handle Redis failures gracefully
- Hardcode limits (use configuration)
- Ignore rate limit metrics

---

## Summary

| Pattern | Rate Limiting | Complexity | Use Case |
|---------|--------------|------------|----------|
| **Local** | Minimal/Optional | ⭐ Low | Development, testing |
| **Remote** | Full Multi-Layer | ⭐⭐⭐ Medium | Production, multi-agent |
| **Multi-Tenant** | Hierarchical | ⭐⭐⭐⭐ High | SaaS, enterprise |

**Default Recommendation**: 
- **Pattern 1** (Local): Basic protection against runaway scripts (optional)
- **Pattern 2** (Remote): Comprehensive multi-layer rate limiting (required)
- **Pattern 3** (Kamaji): Tenant + client quotas (when needed)

---

**Last Updated**: October 22, 2025  
**Next Review**: January 22, 2026  
**Owner**: SRE/Platform Team

---

## 📋 Document Review

**Review Completed By**: 
- ✅ **AI Senior Pentester (COMPLETE)** - October 22, 2025 - Identified V11 (DoS) and emphasized need for rate limiting
- [AI Senior SRE (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review (1/7 complete)  
**Next Review**: After rate limiting implementation

---

