# 🚨 Linkerd Certificate Expired

## Alert Details

**Alert Name:** `LinkerdCertificateExpired`  
**Severity:** Critical  
**Component:** Linkerd Service Mesh - Certificate Management  
**Category:** Security

## Problem Description

Linkerd certificates have expired, causing mTLS (mutual TLS) connection failures between services in the mesh. This prevents secure communication and breaks service-to-service connectivity.

### Error Symptoms

```
linkerd_reconnect: Failed to connect 
error=invalid peer certificate: certificate expired: 
verification time 1761231823 (UNIX), 
but certificate is not valid after 1761177843 (53980 seconds ago)
```

or

```
WARN identity:identity{server.addr=linkerd-identity-headless.linkerd.svc.cluster.local:8080}:
controller{addr=linkerd-identity-headless.linkerd.svc.cluster.local:8080}:
endpoint{addr=10.244.1.111:8080}: linkerd_reconnect: 
Failed to connect error=endpoint 10.244.1.111:8080: 
invalid peer certificate: BadSignature 
error.sources=[invalid peer certificate: BadSignature]
```

## Root Cause

This issue occurs when:

1. **Trust Anchor Rotation**: Linkerd's trust anchor certificate was automatically rotated/renewed
2. **Stale Pod Certificates**: Some pods were running **before** the trust anchor rotation and still have certificates issued by the old (now expired) trust anchor
3. **Certificate Validation Failure**: When these pods try to communicate via Linkerd's mTLS, certificates fail validation because:
   - Their certificates were signed by the old (expired) trust anchor
   - The new trust anchor doesn't recognize these old certificates

## Impact

- ✗ Service-to-service mTLS communication fails
- ✗ Affected pods cannot communicate securely within the mesh
- ✗ Application functionality may be degraded or broken
- ✗ Service mesh security is compromised

## Diagnosis Steps

### 1. Verify the Alert

```bash
# Check for certificate-related errors in proxy logs
kubectl logs -n linkerd -l linkerd.io/control-plane-component=destination --tail=50 | grep -i "certificate\|expired\|tls"
```

### 2. Run Linkerd Health Check

```bash
# Run comprehensive Linkerd check
linkerd check --proxy

# Look for this specific warning:
# ‼ data plane proxies certificate match CA
#     Some pods do not have the current trust bundle and must be restarted
```

### 3. Identify Affected Pods

```bash
# Get list of pods with certificate mismatches
linkerd check --proxy 2>&1 | grep -A 20 "data plane proxies certificate match CA"

# Example output:
# Some pods do not have the current trust bundle and must be restarted:
#     * notifi-test/mock-grpc-4000-6b499dc758-bkf7t
#     * notifi-test/mock-health-6000-69549bd7b6-rf55k
#     * notifi-test/mock-http-5000-7b6474c96c-tlxkj
```

### 4. Check Certificate Expiry Dates

```bash
# Check identity issuer certificate
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.data.crt\.pem}' | base64 -d | openssl x509 -noout -dates

# Check trust anchor
kubectl -n linkerd get cm linkerd-identity-trust-roots -o yaml | grep -A 20 "ca-bundle.crt"

# Check webhook certificates
for cert in linkerd-policy-validator-k8s-tls linkerd-proxy-injector-k8s-tls linkerd-sp-validator-k8s-tls; do 
  echo "=== $cert ==="
  kubectl -n linkerd get secret $cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
done
```

### 5. Check Identity Service Logs

```bash
# Check identity controller for certificate issuance
kubectl logs -n linkerd deploy/linkerd-identity --tail=100 | grep -i "certificate\|expired\|error"
```

## Resolution Steps

### Immediate Fix (Tested and Verified ✅)

**These are the exact steps performed during the October 23, 2025 incident:**

#### Step 1: Identify All Affected Deployments

```bash
# Run linkerd check to get the list
linkerd check --proxy | grep -A 50 "data plane proxies certificate match CA"

# In our case, affected deployments in notifi-test namespace were:
# - mock-grpc-4000
# - mock-health-6000
# - mock-http-5000
# - mock-http-80
# - mock-metrics-7000
# - mock-rabbitmq-5672
# - mock-redis-6379
# - mock-websocket-8080
```

#### Step 2: Restart Affected Deployments

```bash
# Restart all affected deployments (replace with your namespace and deployment names)
kubectl rollout restart deployment -n notifi-test \
  mock-health-6000 \
  mock-http-5000 \
  mock-http-80 \
  mock-metrics-7000 \
  mock-rabbitmq-5672 \
  mock-redis-6379 \
  mock-websocket-8080

# Note: mock-grpc-4000 deployment didn't exist in this case
```

#### Step 3: Wait for Rollout to Complete

```bash
# Monitor rollout status
kubectl rollout status deployment -n notifi-test mock-http-80 --timeout=60s

# Check all pods are running
kubectl get pods -n notifi-test
```

#### Step 4: Verify the Fix

```bash
# Wait for pods to fully restart (10-30 seconds)
sleep 15

# Run linkerd check again
linkerd check --proxy | grep -A 10 "linkerd-identity-data-plane"

# Should now show:
# √ data plane proxies certificate match CA
```

#### Step 5: Verify mTLS Communication

```bash
# Check for certificate errors in logs
kubectl logs -n linkerd -l linkerd.io/control-plane-component=destination --tail=20 | grep -i "certificate\|error"

# Should see no certificate errors

# Verify service communication
linkerd viz stat deployment -n notifi-test
```

### Alternative: Restart Specific Pods

If you have specific pods (not deployments):

```bash
# Delete pods directly - they will be recreated by their controller
kubectl delete pod -n <namespace> <pod-name>

# Or for multiple pods
kubectl delete pod -n notifi-test \
  mock-grpc-4000-6b499dc758-bkf7t \
  mock-health-6000-69549bd7b6-rf55k
```

### Full Namespace Restart (Use with Caution)

For critical situations affecting an entire namespace:

```bash
# Restart all deployments in a namespace
kubectl get deployments -n <namespace> -o name | xargs -I {} kubectl rollout restart {} -n <namespace>

# Or restart all pods with Linkerd proxy
kubectl get pods -n <namespace> -l linkerd.io/proxy-deployment --no-headers | \
  awk '{print $1}' | xargs -I {} kubectl delete pod -n <namespace> {}
```

## Prevention and Monitoring

### 1. Set Up Proactive Monitoring

The Prometheus alert rules have been configured in:
```
/repos/homelab/flux/clusters/homelab/infrastructure/linkerd/prometheus-rules.yaml
```

Key alerts:
- **LinkerdCertificateExpired** - Fires immediately on certificate expiry
- **LinkerdCertificateBadSignature** - Fires on signature validation failures
- **LinkerdCertificateExpiringSoon** - Fires 7 days before expiry (preventive)

### 2. Regular Certificate Audits

```bash
# Weekly certificate health check (add to cron)
linkerd check --proxy | tee /var/log/linkerd-health-$(date +%Y%m%d).log

# Monthly detailed certificate inspection
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.data.crt\.pem}' | \
  base64 -d | openssl x509 -text | grep -A 2 "Validity"
```

### 3. Automated Pod Rotation

Consider implementing a post-rotation automation:

```bash
# Script to restart pods after trust anchor rotation
#!/bin/bash
# File: /scripts/linkerd-cert-rotation-handler.sh

echo "Checking for certificate mismatches..."
AFFECTED_PODS=$(linkerd check --proxy 2>&1 | grep -A 100 "data plane proxies certificate match CA" | grep "^\s*\*" | awk '{print $2}')

if [ -n "$AFFECTED_PODS" ]; then
  echo "Found affected pods, restarting..."
  echo "$AFFECTED_PODS" | while read pod_ref; do
    namespace=$(echo $pod_ref | cut -d'/' -f1)
    pod=$(echo $pod_ref | cut -d'/' -f2)
    echo "Deleting pod: $namespace/$pod"
    kubectl delete pod -n $namespace $pod
  done
else
  echo "No affected pods found."
fi
```

### 4. Documentation Updates

Keep this runbook updated with:
- Latest incident timestamps
- New affected namespaces
- Certificate rotation patterns observed

## Post-Incident Actions

### 1. Verify Service Health

```bash
# Check all services are communicating
linkerd viz stat deploy --all-namespaces

# Check for any remaining errors
kubectl get pods -A -l linkerd.io/proxy-deployment | grep -v Running
```

### 2. Update Monitoring

- Verify alerts are firing correctly
- Update alert thresholds if needed
- Add new namespaces to monitoring

### 3. Document the Incident

Record in incident log:
- Date and time of certificate expiry
- Number of affected pods
- Time to resolution
- Any service impact

### 4. Review Certificate Lifecycle

```bash
# Document current certificate timeline
echo "=== Certificate Timeline ===" > /tmp/linkerd-cert-timeline.txt
echo "Trust Anchor:" >> /tmp/linkerd-cert-timeline.txt
kubectl -n linkerd get cm linkerd-identity-trust-roots -o yaml | \
  grep -A 50 "BEGIN CERTIFICATE" | base64 -d | openssl x509 -noout -dates >> /tmp/linkerd-cert-timeline.txt

echo -e "\nIdentity Issuer:" >> /tmp/linkerd-cert-timeline.txt
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.data.crt\.pem}' | \
  base64 -d | openssl x509 -noout -dates >> /tmp/linkerd-cert-timeline.txt
```

## Related Runbooks

- [Certificate Bad Signature](certificate-bad-signature.md)
- [Certificate Invalid](certificate-invalid.md)
- [Identity Service Down](identity-service-down.md)
- [Proxy Connection Failures](proxy-connection-failures.md)

## Additional Resources

- [Linkerd Certificate Management](https://linkerd.io/2.11/tasks/automatically-rotating-control-plane-tls-credentials/)
- [Linkerd Identity Service](https://linkerd.io/2.11/reference/architecture/#identity)
- [mTLS Deep Dive](https://linkerd.io/2.11/features/automatic-mtls/)

## Incident History

### October 23, 2025 - Certificate Expiry Incident

**Timeline:**
- Trust anchor rotated: Oct 22, 2025 00:47:24 UTC
- Certificate expired: ~15 hours after rotation
- Incident detected: Oct 23, 2025 (via logs)
- Resolution started: Oct 23, 2025
- Resolution completed: Oct 23, 2025 (< 5 minutes)

**Affected Services:**
- Namespace: `notifi-test`
- 7 deployments affected
- All pods successfully restarted

**Lessons Learned:**
1. ✅ Certificate rotation is automatic but requires pod restarts
2. ✅ Monitoring alerts now configured to detect early
3. ✅ Documented exact remediation steps
4. ✅ Recovery time: < 5 minutes once diagnosed

**Future Improvements:**
- [ ] Implement automated pod restart on certificate rotation
- [ ] Add certificate rotation notifications to Slack
- [ ] Create Grafana dashboard for certificate lifecycle
- [ ] Set up weekly automated certificate audits
