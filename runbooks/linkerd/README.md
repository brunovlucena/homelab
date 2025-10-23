# 🔒 Linkerd Service Mesh Runbooks

This directory contains comprehensive runbooks for troubleshooting and managing Linkerd service mesh issues in the homelab environment.

## 📋 Available Runbooks

### 🚨 Critical Issues
- **[Certificate Expired](certificate-expired.md)** - Linkerd certificate has expired causing mTLS failures
- **[Certificate Bad Signature](certificate-bad-signature.md)** - Certificate signature validation failures
- **[Certificate Invalid](certificate-invalid.md)** - Invalid certificate causing mTLS failures
- **[Identity Service Down](identity-service-down.md)** - Linkerd identity service unavailable

### ⚠️ Warning Issues
- **[Certificate Expiring Soon](certificate-expiring-soon.md)** - Certificate will expire within 7 days
- **[Proxy Connection Failures](proxy-connection-failures.md)** - High rate of proxy connection failures
- **[Proxy High Latency](proxy-high-latency.md)** - High latency in Linkerd proxy
- **[Proxy High Error Rate](proxy-high-error-rate.md)** - High error rate in Linkerd proxy

## 🔧 Quick Diagnostic Commands

### Check Linkerd Status
```bash
# Overall Linkerd health
linkerd check

# Check proxy status
linkerd check --proxy

# Check specific namespace
linkerd check --proxy --namespace notifi-test
```

### Certificate Management
```bash
# Check certificate status
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.data.crt\.pem}' | base64 -d | openssl x509 -noout -dates

# Check trust anchor
kubectl -n linkerd get cm linkerd-identity-trust-roots -o yaml

# List all certificates
kubectl -n linkerd get secrets --field-selector type=kubernetes.io/tls
```

### Pod and Service Status
```bash
# Check Linkerd control plane
kubectl get pods -n linkerd

# Check proxy injection status
kubectl get pods -A -l linkerd.io/proxy-deployment

# Check service mesh traffic
linkerd viz stat deployment -n notifi-test
```

## 🚨 Emergency Procedures

### 1. Certificate Expiry Emergency
```bash
# 1. Identify affected pods
linkerd check --proxy | grep -A 10 "linkerd-identity-data-plane"

# 2. Restart affected deployments
kubectl rollout restart deployment -n <namespace> <deployment-name>

# 3. Verify fix
linkerd check --proxy
```

### 2. Complete Service Mesh Recovery
```bash
# 1. Restart Linkerd control plane
kubectl rollout restart deployment -n linkerd

# 2. Wait for control plane to be ready
kubectl wait --for=condition=available deployment -n linkerd --timeout=300s

# 3. Restart all injected pods
kubectl get pods -A -l linkerd.io/proxy-deployment --no-headers | while read ns pod _; do
  kubectl delete pod -n $ns $pod
done
```

## 📊 Monitoring and Alerting

### Prometheus Alerts
- **LinkerdCertificateExpired** - Critical certificate expiry
- **LinkerdCertificateBadSignature** - Critical signature validation failure
- **LinkerdCertificateInvalid** - Critical invalid certificate
- **LinkerdCertificateExpiringSoon** - Warning for upcoming expiry
- **LinkerdProxyConnectionFailures** - Critical connection failures
- **LinkerdIdentityServiceIssues** - Critical identity service down
- **LinkerdProxyHighLatency** - Warning for high latency
- **LinkerdProxyHighErrorRate** - Warning for high error rate

### Grafana Dashboards
- **Linkerd Overview** - Service mesh health overview
- **Linkerd Traffic** - Traffic patterns and metrics
- **Linkerd Security** - Certificate and mTLS status

## 🔄 Maintenance Procedures

### Certificate Rotation
1. **Automatic Rotation**: Linkerd handles trust anchor rotation automatically
2. **Pod Restart**: Affected pods need restart after rotation
3. **Monitoring**: Watch for certificate-related alerts during rotation

### Regular Health Checks
```bash
# Daily health check
linkerd check --proxy

# Weekly certificate check
kubectl -n linkerd get secret linkerd-identity-issuer -o jsonpath='{.data.crt\.pem}' | base64 -d | openssl x509 -noout -dates

# Monthly security audit
linkerd check --proxy | grep -i certificate
```

## 📚 Additional Resources

- [Linkerd Documentation](https://linkerd.io/2.11/)
- [Linkerd Certificate Management](https://linkerd.io/2.11/tasks/automatically-rotating-control-plane-tls-credentials/)
- [Linkerd Troubleshooting](https://linkerd.io/2.11/tasks/troubleshooting/)
- [Service Mesh Security](https://linkerd.io/2.11/features/automatic-mtls/)

## 🆘 Emergency Contacts

- **Primary SRE**: Bruno Lucena
- **Escalation**: Check #jamie-sre-chatbot Slack channel
- **Documentation**: This runbook directory
