# MetalLB Setup for agent-whatsapp-rust

## 🎯 Overview

This guide explains how to set up MetalLB for the `messaging-service` WebSocket endpoint in `agent-whatsapp-rust`. MetalLB provides Layer 4 load balancing, which is optimal for long-lived WebSocket connections.

## ✅ Why MetalLB vs Cloudflare Tunnel?

### MetalLB Advantages
- ✅ **No connection duration limits** - WebSocket connections can stay open indefinitely
- ✅ **Lower latency** - TCP passthrough, no HTTP inspection overhead
- ✅ **Better performance** - Designed for persistent TCP connections
- ✅ **No connection drops** - Unaffected by tunnel reconnections
- ✅ **Full control** - On-premise, no external dependencies

### Cloudflare Tunnel Limitations
- ❌ **Connection duration limits** - 100 seconds (free tier), longer but still limited (paid)
- ❌ **Connection drops** - During tunnel pod restarts/reconnections
- ❌ **Layer 7 overhead** - HTTP inspection adds latency
- ❌ **Not optimal** - For long-lived WebSocket connections

## 📋 Prerequisites

1. **Kubernetes cluster** with MetalLB support
2. **Network configuration**:
   - Router/firewall access for port forwarding
   - Reserved IP range: `192.168.1.200-192.168.1.220` (adjust based on your network)
   - Ensure these IPs are **NOT** in your router's DHCP pool
3. **DNS configuration**:
   - Domain: `messaging.lucena.cloud` (or your domain)
   - Point to your public IP address
4. **SSL/TLS certificate**:
   - cert-manager installed
   - Let's Encrypt ClusterIssuer configured

## 🚀 Installation Steps

### Step 1: Deploy MetalLB

MetalLB is deployed via Flux GitOps. The infrastructure is located at:
```
flux/infrastructure/metallb/
```

**Components:**
- `namespace.yaml` - Creates `metallb-system` namespace
- `helmrelease.yaml` - Deploys MetalLB via Helm
- `ipaddresspool.yaml` - Configures IP address pool and L2 advertisement
- `kustomization.yaml` - Kustomize configuration

**Verify installation:**
```bash
# Check MetalLB pods
kubectl get pods -n metallb-system

# Check IP address pool
kubectl get ipaddresspool -n metallb-system

# Check L2 advertisement
kubectl get l2advertisement -n metallb-system
```

### Step 2: Configure Router Port Forwarding

Configure your router to forward traffic to the MetalLB IP:

1. **Get the LoadBalancer IP:**
   ```bash
   kubectl get svc messaging-service-lb -n homelab-services
   ```
   Look for the `EXTERNAL-IP` field (e.g., `192.168.1.200`)

2. **Configure router port forwarding:**
   - External Port: `443` (HTTPS/WSS)
   - Internal IP: `<MetalLB-IP>` (e.g., `192.168.1.200`)
   - Internal Port: `443`
   - Protocol: TCP

### Step 3: Configure DNS

Point your domain to your public IP:

```bash
# Example DNS record (Cloudflare or your DNS provider)
Type: A
Name: messaging
Value: <your-public-ip>
TTL: Auto
```

### Step 4: Deploy LoadBalancer Service

The LoadBalancer service is already configured in:
```
flux/ai/agents-whatsapp-rust/k8s/base/messaging-service-loadbalancer.yaml
```

**Verify deployment:**
```bash
# Check LoadBalancer service
kubectl get svc messaging-service-lb -n homelab-services

# Should show EXTERNAL-IP assigned by MetalLB
# Example output:
# NAME                    TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)        AGE
# messaging-service-lb    LoadBalancer   10.99.1.100   192.168.1.200   443:30001/TCP  5m
```

### Step 5: SSL/TLS Configuration (Optional)

You have two options for SSL/TLS:

#### Option A: Application-Level SSL/TLS (Recommended for Performance)
Handle SSL/TLS termination at the `messaging-service` application level. This provides:
- True Layer 4 passthrough
- Lowest latency
- No ingress overhead

**Configuration:**
- Configure `messaging-service` to listen on HTTPS/WSS
- Use cert-manager to generate certificates
- Mount certificates as secrets in the Knative service

#### Option B: Ingress-Level SSL/TLS Termination
Use Traefik Ingress for SSL/TLS termination. This adds Layer 7 overhead but provides:
- Centralized certificate management
- Easier configuration

**Enable Ingress:**
```yaml
# In k8s/base/kustomization.yaml, uncomment:
- messaging-service-ingress.yaml
```

**Verify:**
```bash
# Check Ingress
kubectl get ingress messaging-service-ingress -n homelab-services

# Check certificate
kubectl get certificate messaging-service-tls -n homelab-services
```

## 🧪 Testing

### Test WebSocket Connection

```bash
# Using wscat
wscat -c wss://messaging.lucena.cloud/ws

# Using curl (test HTTP upgrade)
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: test" \
  https://messaging.lucena.cloud/ws
```

### Test from AppRestaurant

1. Update AppRestaurant base URL:
   ```swift
   // Change from localhost:8080 to:
   let baseURL = "https://messaging.lucena.cloud"
   ```

2. Test connection from iOS app
3. Monitor connection stability (should not drop after 100 seconds)

## 📊 Monitoring

### Check LoadBalancer Status

```bash
# Get LoadBalancer service details
kubectl describe svc messaging-service-lb -n homelab-services

# Check MetalLB speaker logs
kubectl logs -n metallb-system -l app=metallb-speaker

# Check MetalLB controller logs
kubectl logs -n metallb-system -l app=metallb-controller
```

### Monitor WebSocket Connections

```bash
# Check messaging-service pods
kubectl get pods -n homelab-services -l serving.knative.dev/service=messaging-service

# Check connection metrics (if Prometheus is configured)
# Query: websocket_connections_total
```

## 🔧 Troubleshooting

### Issue: LoadBalancer IP not assigned

**Symptoms:**
- `EXTERNAL-IP` shows `<pending>`

**Solutions:**
1. Check MetalLB pods are running:
   ```bash
   kubectl get pods -n metallb-system
   ```

2. Check IP address pool:
   ```bash
   kubectl get ipaddresspool -n metallb-system -o yaml
   ```

3. Verify IP range doesn't conflict with DHCP:
   - Ensure `192.168.1.200-192.168.1.220` is not in router's DHCP pool

### Issue: Cannot connect from outside

**Symptoms:**
- Connection timeout from external clients

**Solutions:**
1. Verify router port forwarding:
   - External Port: `443` → Internal IP: `<MetalLB-IP>`:443

2. Check firewall rules:
   ```bash
   # Allow traffic on port 443
   sudo ufw allow 443/tcp
   ```

3. Test from inside network:
   ```bash
   # Should work from inside network
   curl https://192.168.1.200/health
   ```

### Issue: SSL/TLS certificate errors

**Symptoms:**
- Certificate validation failures

**Solutions:**
1. Check cert-manager:
   ```bash
   kubectl get certificate -n homelab-services
   kubectl describe certificate messaging-service-tls -n homelab-services
   ```

2. Check ClusterIssuer:
   ```bash
   kubectl get clusterissuer letsencrypt-prod
   ```

3. Verify DNS is pointing to correct IP:
   ```bash
   dig messaging.lucena.cloud
   ```

## 🔄 Migration from Cloudflare Tunnel

If you're currently using Cloudflare Tunnel:

1. **Deploy MetalLB** (Step 1)
2. **Deploy LoadBalancer service** (Step 4)
3. **Test connectivity** (Step 5)
4. **Disable Cloudflare Tunnel ingress:**
   ```bash
   kubectl patch cloudflaretunnelingress messaging-service-ws \
     -n agents-whatsapp-rust \
     --type merge \
     -p '{"spec":{"enabled":false}}'
   ```
5. **Update AppRestaurant** to use new endpoint
6. **Monitor** for connection stability improvements

## 📝 Configuration Files

### MetalLB IP Address Pool
**Location:** `flux/infrastructure/metallb/ipaddresspool.yaml`

**Customization:**
- Adjust IP range based on your network: `192.168.1.200-192.168.1.220`
- Ensure range doesn't conflict with DHCP

### LoadBalancer Service
**Location:** `flux/ai/agents-whatsapp-rust/k8s/base/messaging-service-loadbalancer.yaml`

**Key settings:**
- Port: `443` (WSS)
- Target Port: `8080` (messaging-service)
- Session Affinity: `ClientIP` (3 hours timeout)

### Ingress (Optional)
**Location:** `flux/ai/agents-whatsapp-rust/k8s/base/messaging-service-ingress.yaml`

**Key settings:**
- WebSocket timeouts: `3600` seconds
- Session affinity: Enabled
- SSL/TLS: cert-manager managed

## 🔗 Related Documentation

- [WebSocket Load Balancer Impact Analysis](./WEBSOCKET_LOAD_BALANCER_IMPACT.md)
- [WebSocket Load Balancer Quick Reference](./WEBSOCKET_LOAD_BALANCER_QUICK_REFERENCE.md)
- [MetalLB Official Documentation](https://metallb.universe.tf/)

## 📅 Last Updated

**Date:** 2025-01-19  
**Status:** ✅ Ready for deployment

---

**Next Steps:**
1. Deploy MetalLB infrastructure
2. Configure router port forwarding
3. Test WebSocket connectivity
4. Monitor connection stability
