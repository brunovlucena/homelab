# WebSocket Load Balancer Impact Analysis

## 🎯 Overview

This document analyzes how WebSocket requirements and Layer 4 load balancing impact:
- **agent-whatsapp-rust** (`messaging-service`)
- **AppRestaurant** (iOS app connecting via WebSocket)

## 🔍 Current Architecture

### Services Using WebSocket

#### 1. `messaging-service` (agents-whatsapp-rust)
- **Endpoint**: `/ws` (WebSocket)
- **Port**: 8080
- **Protocol**: WebSocket over HTTP/HTTPS
- **Deployment**: Knative Service (scales 2-10 replicas)
- **Namespace**: `agents-whatsapp-rust`

#### 2. AppRestaurant (iOS)
- **Connection**: WebSocket to `messaging-service`
- **URL Pattern**: `wss://<hostname>/ws`
- **Protocol**: WebSocket Secure (WSS)
- **Features**: 
  - Real-time messaging
  - Heartbeat (5s interval)
  - Auto-reconnection with exponential backoff

### Current Ingress Setup

**Status**: ⚠️ **No dedicated ingress found for `messaging-service`**

The homelab currently uses:
- **Cloudflare Tunnel** (Layer 7 HTTP/HTTPS proxy)
- **CloudflareTunnelIngress** CRD for service exposure
- **No ingress configured** for `messaging-service` WebSocket endpoint

## ⚠️ The Problem: Layer 7 vs Layer 4 Load Balancing

### WebSocket Requirements

WebSockets require:
1. **Persistent TCP connections** (long-lived)
2. **Connection state preservation** (sticky sessions)
3. **HTTP Upgrade handshake** support
4. **No connection termination** during load balancer operations

### Layer 7 Load Balancing (Current: Cloudflare Tunnel)

**How it works:**
- Inspects HTTP/HTTPS traffic
- Terminates SSL/TLS
- Performs HTTP routing
- Supports WebSocket via HTTP Upgrade

**Limitations for WebSocket:**

#### 1. Connection Duration Limits
```
Cloudflare Free Tier: 100 seconds max connection duration
Cloudflare Paid: Longer but still limited
```
**Impact**: 
- Long-lived WebSocket connections get dropped
- AppRestaurant connections may timeout
- Requires frequent reconnections

#### 2. Connection Drops During Reconnections
```
Cloudflare Tunnel Pod → Reconnection → All WebSocket connections lost
```
**Impact**:
- AppRestaurant loses connection during tunnel restarts
- Messages may be lost during reconnection window
- User experience degradation

#### 3. No Native Load Balancing
```
Cloudflare Tunnel → Single K8s Service → K8s Service Load Balancing
```
**Impact**:
- Relies on Kubernetes Service load balancing
- No session affinity at tunnel level
- WebSocket connections may be routed to different pods

#### 4. HTTP Layer Inspection Overhead
- Additional latency from HTTP parsing
- Memory overhead for connection state
- Not optimized for long-lived connections

### Layer 4 Load Balancing (Recommended for WebSocket)

**How it works:**
- TCP/SSL passthrough (no HTTP inspection)
- Direct connection forwarding
- Lower latency
- Better for long-lived connections

**Benefits:**
- ✅ No connection duration limits
- ✅ Lower latency (no HTTP parsing)
- ✅ Better for persistent connections
- ✅ Supports any TCP-based protocol

## 📊 Impact Analysis

### Impact on `agent-whatsapp-rust` (messaging-service)

#### Current Issues:
1. **No Public Ingress**
   - `messaging-service` is not exposed via Cloudflare Tunnel
   - AppRestaurant cannot connect from outside cluster
   - Only accessible via port-forward or internal services

2. **Scaling Challenges**
   - Knative scales 2-10 replicas
   - Without session affinity, WebSocket connections may be routed to different pods
   - State management becomes complex

3. **Connection Stability**
   - If exposed via Cloudflare Tunnel, connections may drop after 100s (free tier)
   - Tunnel reconnections cause connection loss

#### Required Changes:
```yaml
# Option 1: Cloudflare Tunnel with WebSocket support (Layer 7)
apiVersion: tunnel.cloudflare.io/v1alpha1
kind: CloudflareTunnelIngress
metadata:
  name: messaging-service-ws
  namespace: agents-whatsapp-rust
spec:
  hostname: messaging.lucena.cloud
  service:
    name: messaging-service
    namespace: agents-whatsapp-rust
    port: 8080
    protocol: http  # Supports WebSocket upgrade
  enabled: true
```

**Limitations:**
- Connection duration limits
- Connection drops during tunnel restarts
- Not optimal for long-lived connections

```yaml
# Option 2: Layer 4 Load Balancer (Recommended)
# Use MetalLB + LoadBalancer Service for TCP passthrough
apiVersion: v1
kind: Service
metadata:
  name: messaging-service-lb
  namespace: agents-whatsapp-rust
spec:
  type: LoadBalancer
  ports:
  - port: 443
    targetPort: 8080
    protocol: TCP
    name: wss
  selector:
    serving.knative.dev/service: messaging-service
```

**Benefits:**
- Direct TCP passthrough
- No connection duration limits
- Lower latency
- Better for WebSocket

### Impact on AppRestaurant

#### Current Issues:
1. **Cannot Connect from iOS App**
   - No public endpoint for `messaging-service`
   - AppRestaurant WebSocketService cannot establish connection
   - Users must use port-forward (development only)

2. **Connection Instability** (if exposed via Cloudflare Tunnel)
   - Connections drop after 100 seconds (free tier)
   - Reconnection logic triggered frequently
   - Poor user experience

3. **Message Loss Risk**
   - During connection drops, messages may be lost
   - Retry queue may fill up
   - Sequence number gaps

#### AppRestaurant WebSocketService Behavior:
```swift
// Current reconnection logic
- Exponential backoff: 2s, 4s, 8s, 16s, 32s, max 60s
- Max reconnection attempts: 10
- Heartbeat interval: 5s
- Auto-retry on connection loss
```

**Impact of Layer 7 Load Balancer:**
- Frequent reconnections due to 100s timeout
- Exponential backoff may exhaust quickly
- User sees connection errors frequently

## 🛠️ Recommended Solutions

### Solution 1: Cloudflare Tunnel with Optimizations (Quick Fix)

**Pros:**
- Uses existing infrastructure
- Quick to implement
- No additional costs

**Cons:**
- Still has connection duration limits
- Connection drops during tunnel restarts

**Implementation:**
```yaml
apiVersion: tunnel.cloudflare.io/v1alpha1
kind: CloudflareTunnelIngress
metadata:
  name: messaging-service-ws
  namespace: agents-whatsapp-rust
spec:
  hostname: messaging.lucena.cloud
  service:
    name: messaging-service
    namespace: agents-whatsapp-rust
    port: 8080
    protocol: http
  enabled: true
  syncInterval: "5m"
```

**Additional Configurations:**
1. **Session Affinity** (if using nginx ingress):
```yaml
annotations:
  nginx.ingress.kubernetes.io/affinity: "cookie"
  nginx.ingress.kubernetes.io/affinity-mode: "persistent"
  nginx.ingress.kubernetes.io/session-cookie-name: "messaging-session"
  nginx.ingress.kubernetes.io/session-cookie-expires: "3600"
```

2. **WebSocket-specific headers**:
```yaml
annotations:
  nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
  nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
  nginx.ingress.kubernetes.io/websocket-services: "messaging-service"
```

### Solution 2: Layer 4 Load Balancer (Recommended)

**Pros:**
- Optimal for WebSocket
- No connection duration limits
- Lower latency
- Better performance

**Cons:**
- Requires additional infrastructure (MetalLB or cloud LB)
- May have additional costs
- More complex setup

**Implementation Options:**

#### Option A: MetalLB (On-Premise)
```yaml
apiVersion: v1
kind: Service
metadata:
  name: messaging-service-lb
  namespace: agents-whatsapp-rust
spec:
  type: LoadBalancer
  ports:
  - port: 443
    targetPort: 8080
    protocol: TCP
    name: wss
  selector:
    serving.knative.dev/service: messaging-service
```

#### Option B: Cloud Load Balancer (AWS/GCP/Azure)
- Use cloud provider's Layer 4 load balancer
- TCP passthrough mode
- SSL termination at load balancer (optional)

#### Option C: HAProxy/Nginx TCP Mode
```nginx
# nginx.conf (TCP mode)
stream {
    upstream messaging_backend {
        least_conn;
        server messaging-service-1:8080;
        server messaging-service-2:8080;
        server messaging-service-3:8080;
    }
    
    server {
        listen 443;
        proxy_pass messaging_backend;
        proxy_timeout 1s;
        proxy_responses 1;
        health_check;
    }
}
```

### Solution 3: Hybrid Approach

**Architecture:**
```
Internet
  │
  ├─→ Cloudflare Tunnel (Layer 7)
  │   └─→ HTTP/HTTPS endpoints
  │
  └─→ Layer 4 Load Balancer
      └─→ WebSocket endpoints (/ws)
```

**Benefits:**
- Use Cloudflare Tunnel for HTTP endpoints
- Use Layer 4 LB for WebSocket
- Best of both worlds

## 📋 Action Items

### Immediate (Required for AppRestaurant to work)

1. **Create CloudflareTunnelIngress for messaging-service**
   ```bash
   kubectl apply -f messaging-service-cloudflare-ingress.yaml
   ```

2. **Verify WebSocket connectivity**
   ```bash
   # Test WebSocket connection
   wscat -c wss://messaging.lucena.cloud/ws
   ```

3. **Update AppRestaurant configuration**
   - Set base URL to `https://messaging.lucena.cloud`
   - Test connection from iOS app

### Short-term (Improve stability)

1. **Implement session affinity**
   - Configure nginx ingress with sticky sessions
   - Or use Knative session affinity annotations

2. **Monitor connection stability**
   - Track connection duration
   - Monitor reconnection frequency
   - Alert on connection drops

3. **Optimize reconnection logic**
   - Adjust heartbeat interval
   - Fine-tune exponential backoff

### Long-term (Optimal solution)

1. **Deploy Layer 4 load balancer**
   - Evaluate MetalLB vs cloud LB
   - Implement TCP passthrough
   - Migrate WebSocket traffic

2. **Implement connection migration**
   - Support server-initiated migration
   - Graceful connection handoff
   - Zero-downtime updates

## 🔗 Related Documentation

- [Cloudflare Tunnel Limitations](../../notifi/repos/infra/20-platform/services/agent-auditor/docs/08-assessments/CLOUDFLAE_TUNNELS.md)
- [WebSocket Deployment Guide](./DEPLOYMENT.md#ingress-configuration)
- [Global Deployment Readiness](./GLOBAL_DEPLOYMENT_READINESS.md#load-balancer-strategy)

## 📝 Notes

- **Cloudflare Tunnel** supports WebSocket but with limitations
- **Layer 4 load balancing** is recommended for production WebSocket deployments
- **Session affinity** is critical for stateful WebSocket connections
- **Connection migration** can help with zero-downtime deployments

---

**Last Updated**: 2025-01-19
**Status**: ⚠️ Action Required - No ingress configured for messaging-service








