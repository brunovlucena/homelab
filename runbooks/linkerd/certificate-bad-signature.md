# 🚨 Linkerd Certificate Bad Signature

## Alert Details

**Alert Name:** `LinkerdCertificateBadSignature`  
**Severity:** Critical  
**Component:** Linkerd Service Mesh - Certificate Validation  
**Category:** Security

## Problem Description

Linkerd proxy is rejecting certificates due to invalid signatures. This typically occurs during or after a trust anchor rotation when pods have certificates signed by a different trust anchor than what's currently active.

### Error Symptoms

```
WARN identity:identity{server.addr=linkerd-identity-headless.linkerd.svc.cluster.local:8080}:
controller{addr=linkerd-identity-headless.linkerd.svc.cluster.local:8080}:
endpoint{addr=10.244.1.111:8080}: linkerd_reconnect: 
Failed to connect error=endpoint 10.244.1.111:8080: 
invalid peer certificate: BadSignature 
error.sources=[invalid peer certificate: BadSignature]
```

## Root Cause

Certificate signature validation failures occur when:

1. **Trust Anchor Mismatch**: Pod's certificate was signed by a trust anchor that differs from the current active trust anchor
2. **Certificate Rotation in Progress**: Trust anchor has been rotated but some pods haven't received new certificates yet
3. **Clock Skew**: Significant time differences between nodes can cause validation issues
4. **Certificate Corruption**: Rare cases of corrupted certificate data

## Impact

- ✗ mTLS connections fail between affected pods
- ✗ Service mesh security degraded
- ✗ Application communication errors
- ✗ Service discovery issues

## Diagnosis Steps

### 1. Check for BadSignature Errors

```bash
# Check proxy logs for signature errors
kubectl logs -n linkerd -l linkerd.io/control-plane-component=destination --tail=100 | grep -i "badsignature"

# Check identity service logs
kubectl logs -n linkerd deploy/linkerd-identity --tail=100 | grep -i "signature"
```

### 2. Verify Trust Anchor Status

```bash
# Get current trust anchor
kubectl -n linkerd get cm linkerd-identity-trust-roots -o yaml

# Check trust anchor certificate details
kubectl -n linkerd get cm linkerd-identity-trust-roots -o jsonpath='{.data.ca-bundle\.crt}' | \
  openssl x509 -text -noout | grep -A 5 "Validity\|Subject"
```

### 3. Check Certificate Issuance Times

```bash
# Check when identity issuer cert was created
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.metadata.creationTimestamp}'

# Check cert validity
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.data.crt\.pem}' | \
  base64 -d | openssl x509 -noout -dates
```

### 4. Identify Affected Pods

```bash
# Run linkerd check to find mismatched certificates
linkerd check --proxy | grep -A 20 "data plane proxies certificate match CA"

# Check for pods with connection issues
kubectl get pods -A -l linkerd.io/proxy-deployment -o wide
```

### 5. Verify System Time Sync

```bash
# Check time on nodes (if you have node access)
kubectl get nodes -o wide

# Check if chronyd or ntpd is running
kubectl get pods -n kube-system | grep -i "time\|chrony\|ntp"
```

## Resolution Steps

### Immediate Fix

#### Step 1: Restart Affected Pods

```bash
# Get list of affected pods from linkerd check
linkerd check --proxy 2>&1 | grep -A 50 "data plane proxies certificate match CA"

# Restart deployments in affected namespace
kubectl rollout restart deployment -n <namespace> <deployment-name>

# Example from Oct 23, 2025 incident:
kubectl rollout restart deployment -n notifi-test \
  mock-health-6000 \
  mock-http-5000 \
  mock-http-80 \
  mock-metrics-7000
```

#### Step 2: Verify New Certificates

```bash
# Wait for pods to restart
sleep 15

# Verify no more BadSignature errors
kubectl logs -n linkerd -l linkerd.io/control-plane-component=destination --tail=50 | grep -i "badsignature"

# Should return nothing if fixed
```

#### Step 3: Run Health Check

```bash
# Verify all certificates match
linkerd check --proxy

# Look for:
# √ data plane proxies certificate match CA
```

### If Issue Persists

#### Option 1: Restart Linkerd Identity Service

```bash
# Restart identity service to reissue certificates
kubectl rollout restart deployment -n linkerd linkerd-identity

# Wait for restart
kubectl rollout status deployment -n linkerd linkerd-identity

# Then restart affected workloads
```

#### Option 2: Verify Trust Anchor Configuration

```bash
# Check if trust anchor is properly configured
linkerd check | grep -A 10 "trust anchors"

# Verify trust anchor validity
kubectl -n linkerd get cm linkerd-identity-trust-roots -o yaml | \
  grep -A 50 "ca-bundle.crt" | base64 -d | openssl x509 -text
```

#### Option 3: Check for Clock Skew

```bash
# Check current time on pods
kubectl get pods -A -l linkerd.io/proxy-deployment -o name | head -5 | \
  xargs -I {} kubectl exec {} -c linkerd-proxy -- date

# All times should be within a few seconds of each other
```

## Prevention

### 1. Automated Certificate Monitoring

Prometheus alert configured in `prometheus-rules.yaml`:

```yaml
- alert: LinkerdCertificateBadSignature
  expr: |
    increase(linkerd_proxy_connect_errors_total{error=~".*BadSignature.*"}[5m]) > 0
  for: 0m
  labels:
    severity: critical
```

### 2. Trust Anchor Rotation Procedure

When manually rotating trust anchor:

```bash
# 1. Generate new trust anchor
# 2. Update trust anchor in cluster
# 3. Wait for propagation
# 4. Restart all meshed pods systematically
# 5. Verify connectivity

# Automated restart script
kubectl get pods -A -l linkerd.io/proxy-deployment --no-headers | \
  awk '{print $1,$2}' | while read ns pod; do
    echo "Restarting $ns/$pod"
    kubectl delete pod -n $ns $pod
    sleep 5  # Graceful restart
  done
```

### 3. Time Synchronization

Ensure NTP/chrony is configured on all nodes:

```bash
# Verify time sync on nodes
kubectl get nodes -o wide

# Check for time sync daemon
kubectl get pods -n kube-system | grep -E "chrony|ntp"
```

## Post-Resolution Verification

### 1. Verify All Certificates Match

```bash
# Full certificate check
linkerd check --proxy --output short

# Should show all checks passing
```

### 2. Test Service Communication

```bash
# Check service-to-service communication
linkerd viz stat deploy --all-namespaces

# Verify no errors in SUCCESS column
```

### 3. Monitor for Recurring Issues

```bash
# Watch for new BadSignature errors (run for 5 minutes)
kubectl logs -n linkerd -l linkerd.io/control-plane-component=destination -f | grep -i "badsignature"
```

## Related Issues

- Often occurs alongside [Certificate Expired](certificate-expired.md)
- May indicate [Identity Service Down](identity-service-down.md)
- Can cause [Proxy Connection Failures](proxy-connection-failures.md)

## Additional Resources

- [Linkerd mTLS Documentation](https://linkerd.io/2.11/features/automatic-mtls/)
- [Certificate Rotation Guide](https://linkerd.io/2.11/tasks/automatically-rotating-control-plane-tls-credentials/)
- [Troubleshooting Identity](https://linkerd.io/2.11/tasks/troubleshooting/#linkerd-identity)

## Notes

- BadSignature errors are often transient during certificate rotation
- Always verify system time synchronization across cluster
- Restart pods in small batches to avoid service disruption
- Monitor alerts during and after trust anchor rotations
