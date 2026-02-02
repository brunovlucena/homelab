# WebSocket Load Balancer - Quick Reference

## 🎯 Key Points

### The Problem
- **WebSockets require persistent TCP connections**
- **Layer 7 load balancers** (Cloudflare Tunnel, nginx) can handle WebSocket but have limitations
- **Layer 4 load balancers** (TCP passthrough) are better for WebSocket

### Current Status
- ❌ **No ingress configured** for `messaging-service`
- ❌ **AppRestaurant cannot connect** from outside cluster
- ⚠️ **Cloudflare Tunnel** supports WebSocket but with limitations

## 📊 Impact Summary

### agent-whatsapp-rust (messaging-service)
| Issue | Impact | Severity |
|-------|--------|----------|
| No public ingress | Cannot connect from AppRestaurant | 🔴 Critical |
| Connection duration limits (if using Cloudflare) | Connections drop after 100s | ⚠️ High |
| No session affinity | WebSocket may route to different pods | ⚠️ Medium |

### AppRestaurant
| Issue | Impact | Severity |
|-------|--------|----------|
| Cannot connect | App doesn't work | 🔴 Critical |
| Connection instability | Poor UX, frequent reconnections | ⚠️ High |
| Message loss risk | Data loss during reconnections | ⚠️ Medium |

## 🛠️ Quick Solutions

### Option 1: Cloudflare Tunnel (Quick Fix)
```yaml
# Apply: messaging-service-cloudflare-ingress.yaml
# Pros: Uses existing infrastructure, quick to implement
# Cons: 100s connection limit, connection drops
```

### Option 2: Layer 4 Load Balancer (Recommended)
```yaml
# Use MetalLB or cloud LB with TCP passthrough
# Pros: No connection limits, better performance
# Cons: Requires additional infrastructure
```

## 📋 Immediate Actions

1. **Create Cloudflare Tunnel ingress** (if using Cloudflare)
   ```bash
   kubectl apply -f k8s/base/messaging-service-cloudflare-ingress.yaml
   ```

2. **Test WebSocket connection**
   ```bash
   wscat -c wss://messaging.lucena.cloud/ws
   ```

3. **Update AppRestaurant base URL**
   - Change from `localhost:8080` to `https://messaging.lucena.cloud`

## 🔗 Full Documentation

See [WEBSOCKET_LOAD_BALANCER_IMPACT.md](./WEBSOCKET_LOAD_BALANCER_IMPACT.md) for detailed analysis.








