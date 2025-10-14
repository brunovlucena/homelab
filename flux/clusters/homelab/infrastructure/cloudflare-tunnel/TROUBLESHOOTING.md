# 🔧 Cloudflare Tunnel Troubleshooting Guide

## Common Issues and Solutions

### 1. Certificate Path Errors
**Error**: `Cannot determine default origin certificate path. No file cert.pem`

**Solution**:
- This error is usually harmless when using tunnel tokens
- The tunnel uses token-based authentication instead of certificate files
- Ensure your tunnel token is valid and not expired

### 2. Connection Failures
**Error**: `failed to serve tunnel connection error="control stream encountered a failure while serving"`

**Solutions**:
- Check if the tunnel token is valid: `cloudflared tunnel --token YOUR_TOKEN run`
- Verify network connectivity to Cloudflare edge servers
- Check if there are firewall rules blocking QUIC traffic
- Try using a different protocol: add `--protocol http2` to the deployment args

### 3. Context Cancellation Errors
**Error**: `failed to run the datagram handler error="context canceled"`

**Solutions**:
- Increase resource limits for the cloudflared container
- Check if the pod is being evicted due to resource constraints
- Verify the cluster has sufficient resources

### 4. Buffer Size Issues
**Error**: `failed to sufficiently increase receive buffer size (was: 208 kiB, wanted: 7168 kiB, got: 416 kiB)`

**Root Cause**:
- QUIC protocol requires large UDP receive buffers (7+ MB)
- Container security context prevents increasing system buffer limits
- Kubernetes nodes may have restrictive net.core.rmem_max settings

**Solutions** (in order of preference):

**Option 1: Switch to HTTP/2 Protocol** (Recommended)
```yaml
args:
  - --protocol
  - http2  # Change from 'quic'
```
This is the most stable and secure solution for Kubernetes environments.

**Option 2: Increase Node-Level Buffer Sizes**
On each Kubernetes node, set:
```bash
sudo sysctl -w net.core.rmem_max=7500000
sudo sysctl -w net.core.wmem_max=7500000
```

**Option 3: Add NET_ADMIN Capability** (Less secure)
```yaml
securityContext:
  capabilities:
    add:
      - NET_ADMIN
```

**Option 4: Use InitContainer for Sysctl**
```yaml
initContainers:
- name: sysctl-tuning
  image: busybox
  command: ["sysctl", "-w", "net.core.rmem_max=7500000"]
  securityContext:
    privileged: true
```

## Diagnostic Commands

### Check Tunnel Status
```bash
# Check pod status
kubectl get pods -n cloudflare-tunnel

# Check pod logs
kubectl logs -n cloudflare-tunnel deployment/cloudflared

# Check pod events
kubectl describe pod -n cloudflare-tunnel -l app=cloudflared
```

### Validate Token
```bash
# Test tunnel token locally
cloudflared tunnel --token YOUR_TOKEN run --no-autoupdate

# Check tunnel list via API
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
  "https://api.cloudflare.com/client/v4/accounts/ACCOUNT_ID/cfd_tunnel"
```

### Network Diagnostics
```bash
# Test connectivity to Cloudflare
ping 1.1.1.1
nslookup cloudflare.com

# Test QUIC connectivity
curl -v --http3 https://cloudflare.com
```

## Configuration Best Practices

### 1. Resource Limits
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "100m"
```

### 2. Health Checks
```yaml
livenessProbe:
  httpGet:
    path: /ready
    port: 2000
  failureThreshold: 3
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready
    port: 2000
  failureThreshold: 3
  initialDelaySeconds: 10
  periodSeconds: 5
```

### 3. Tunnel Arguments
```yaml
args:
- tunnel
- --no-autoupdate
- --loglevel
- info
- --metrics
- 0.0.0.0:2000
- --protocol
- quic
- --retries
- "5"
- --heartbeat-count
- "5"
- --heartbeat-interval
- "5s"
- run
- --token
- $(TUNNEL_TOKEN)
```

## Monitoring and Alerting

### Prometheus Metrics
The tunnel exposes metrics on port 2000:
- `cloudflared_tunnel_connector_connection_attempts_total`
- `cloudflared_tunnel_connector_connection_duration_seconds`
- `cloudflared_tunnel_connector_connection_failures_total`

### Grafana Dashboard
Create a dashboard to monitor:
- Connection attempts and failures
- Connection duration
- Tunnel status
- Resource usage

## Recovery Procedures

### 1. Token Regeneration
```bash
# Run the regeneration script
./regenerate-token.sh

# Follow the prompts to create a new tunnel and token
# Update the sealed secret
# Apply changes to the cluster
```

### 2. Complete Tunnel Reset
```bash
# Delete the deployment
kubectl delete deployment cloudflared -n cloudflare-tunnel

# Delete the secret
kubectl delete secret cloudflare-tunnel-credentials -n cloudflare-tunnel

# Regenerate everything
./regenerate-token.sh

# Apply new configuration
kubectl apply -k .
```

### 3. DNS Troubleshooting
```bash
# Check DNS propagation
dig @1.1.1.1 your-domain.com
dig @8.8.8.8 your-domain.com

# Check CNAME records
dig CNAME your-domain.com
```

## Performance Optimization

### 1. Protocol Selection
- **QUIC** (default): Better performance, newer protocol
- **HTTP/2**: More compatible, fallback option

### 2. Connection Pooling
- Use multiple replicas for high availability
- Consider connection pooling for high-traffic scenarios

### 3. Caching
- Enable Cloudflare caching for static content
- Configure cache rules in Cloudflare dashboard

## Security Considerations

### 1. Token Security
- Never commit tunnel tokens to version control
- Use sealed secrets for token storage
- Rotate tokens regularly

### 2. Network Security
- Ensure proper firewall rules
- Use network policies to restrict traffic
- Monitor for suspicious activity

### 3. Access Control
- Use Cloudflare Access for additional authentication
- Configure proper CORS policies
- Implement rate limiting

## Support Resources

- [Cloudflare Tunnel Documentation](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [Cloudflare Community](https://community.cloudflare.com/)
- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug-application-cluster/)

## Emergency Contacts

If all else fails:
1. Check Cloudflare status page
2. Verify account billing status
3. Contact Cloudflare support
4. Consider temporary workarounds (direct IP access, etc.)

