# SRE-008: Certificate Lifecycle Management

**Status**: Backlog
**Priority**: P0
**Story Points**: 5  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-175/sre-008-certificate-lifecycle-management  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** certificate lifecycle management  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] All certificates must be monitored for expiry with alerts at 30-day and 7-day thresholds
- [ ] Automated renewal must be enabled for all managed certificates
- [ ] Zero-downtime rotation procedures must be documented and tested
- [ ] Emergency procedures must enable certificate replacement within 15 minutes

---


## Overview

This runbook provides comprehensive procedures for managing TLS certificates in the Knative Lambda infrastructure, including monitoring, renewal, rotation, and incident response.

### Acceptance Criteria
- All certificates must be monitored for expiry with alerts at 30-day and 7-day thresholds
- Automated renewal must be enabled for all managed certificates
- Zero-downtime rotation procedures must be documented and tested
- Emergency procedures must enable certificate replacement within 15 minutes

## Certificate Inventory

### Cluster Certificates | Certificate | Namespace | Issuer | Expiry Monitoring | Auto-Renewal | |------------ | ----------- | -------- | ------------------- | -------------- | | Ingress TLS | knative-serving | cert-manager | ✅ Enabled | ✅ Automated | | Webhook Certificates | knative-serving | cert-manager | ✅ Enabled | ✅ Automated | | Service Mesh Certificates | istio-system | cert-manager | ✅ Enabled | ✅ Automated | | Lambda Service TLS | lambda-system | cert-manager | ✅ Enabled | ✅ Automated | ## Certificate Expiry Monitoring

### Alert Thresholds

- **30-day warning**: Certificate expiring in 30 days - Review and plan renewal
- **7-day critical**: Certificate expiring in 7 days - Immediate action required
- **Expired**: Certificate has expired - Emergency rotation required

### Monitoring Commands

```bash
# List all TLS secrets with expiry dates
kubectl get secrets --all-namespaces \
  -o json | jq -r '.items[] | select(.type=="kubernetes.io/tls") | "\(.metadata.namespace)/\(.metadata.name)"'

# Check specific certificate expiry
kubectl get secret <secret-name> -n <namespace> -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -noout -enddate

# Check cert-manager certificate status
kubectl get certificates --all-namespaces
```

## Automated Certificate Renewal

### cert-manager Configuration

All certificates are managed by cert-manager with automatic renewal enabled.

**Default Renewal Policy:**
- Renewal triggers 30 days before expiry
- Maximum 5 renewal attempts
- Exponential backoff on failures

### Verification

```bash
# Check cert-manager is running
kubectl get pods -n cert-manager

# Verify certificate renewal configuration
kubectl describe certificate <cert-name> -n <namespace>

# Check for renewal events
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | grep -i certificate
```

## Manual Certificate Renewal Procedures

### Emergency Manual Renewal

When automated renewal fails or for emergency rotation:

```bash
# Step 1: Backup existing certificate
kubectl get secret <cert-secret> -n <namespace> -o yaml > cert-backup.yaml

# Step 2: Delete certificate to trigger renewal
kubectl delete certificate <cert-name> -n <namespace>

# Step 3: cert-manager will automatically recreate it
# Monitor the renewal process
kubectl get certificate <cert-name> -n <namespace> -w

# Step 4: Verify new certificate
kubectl get secret <cert-secret> -n <namespace> -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -noout -text
```

### Manual Certificate Replacement

For certificates not managed by cert-manager:

```bash
# Step 1: Generate new certificate (using your CA)
# ...

# Step 2: Create new secret
kubectl create secret tls <new-secret-name> \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  -n <namespace>

# Step 3: Update services to use new secret
kubectl patch <resource> <name> -n <namespace> \
  -p '{"spec":{"tls":[{"secretName":"<new-secret-name>"}]}}'

# Step 4: Verify services are using new certificate
kubectl describe <resource> <name> -n <namespace>

# Step 5: Delete old secret after verification
kubectl delete secret <old-secret-name> -n <namespace>
```

## Zero-Downtime Certificate Rotation

### Pre-Rotation Checklist

- [ ] Verify new certificate is valid
- [ ] Check certificate chain is complete
- [ ] Confirm proper SAN (Subject Alternative Names)
- [ ] Test certificate with openssl s_client
- [ ] Backup existing certificate
- [ ] Schedule during maintenance window (if possible)

### Rotation Procedure

```bash
# Step 1: Create new secret with different name
kubectl create secret tls <cert-new> \
  --cert=new-cert.pem \
  --key=new-key.pem \
  -n <namespace>

# Step 2: Update service to use new secret (zero-downtime)
kubectl patch service <service-name> -n <namespace> \
  --type='json' \
  -p='[{"op": "replace", "path": "/spec/tls/0/secretName", "value":"<cert-new>"}]'

# Step 3: Monitor for connection errors
kubectl logs -n <namespace> <pod-name> --tail=100 -f

# Step 4: Verify traffic is using new certificate
curl -v https://<service-url> 2>&1 | grep "subject:"

# Step 5: After 24h of successful operation, remove old secret
kubectl delete secret <cert-old> -n <namespace>
```

### Rollback Procedure

If issues occur after rotation:

```bash
# Immediate rollback to old certificate
kubectl patch service <service-name> -n <namespace> \
  --type='json' \
  -p='[{"op": "replace", "path": "/spec/tls/0/secretName", "value":"<cert-old>"}]'

# Verify rollback
kubectl get service <service-name> -n <namespace> -o yaml | grep secretName
```

## Certificate Incident Response

### Investigation Steps

When a certificate issue is reported:

1. **Identify affected certificate:**
   ```bash
   kubectl get certificates --all-namespaces | grep -v Ready
   ```

2. **Check certificate details:**
   ```bash
   kubectl describe certificate <cert-name> -n <namespace>
   ```

3. **Review cert-manager logs:**
   ```bash
   kubectl logs -n cert-manager deployment/cert-manager --tail=100
   ```

4. **Check certificate request status:**
   ```bash
   kubectl get certificaterequest -n <namespace>
   kubectl describe certificaterequest <cr-name> -n <namespace>
   ```

5. **Verify issuer health:**
   ```bash
   kubectl describe issuer <issuer-name> -n <namespace>
   ```

### Incident Types

#### 1. Certificate Expired
**Severity:** P0 - Service disruption

**Immediate Actions:**
1. Check if automated renewal failed: `kubectl describe certificate <cert-name>`
2. Review cert-manager logs: `kubectl logs -n cert-manager deployment/cert-manager`
3. If renewal is stuck, force manual renewal (see Manual Renewal section)
4. Update monitoring alerts

#### 2. Certificate Renewal Failure
**Severity:** P1 - Potential future disruption

**Actions:**
1. Check ACME challenge status: `kubectl describe challenges --all-namespaces`
2. Verify DNS/HTTP-01 challenge accessibility
3. Check issuer configuration: `kubectl describe issuer -n <namespace>`
4. Review cert-manager logs for errors
5. If persistent, escalate to Platform team

#### 3. Invalid Certificate Detected
**Severity:** P1 - Security issue

**Actions:**
1. Identify affected services
2. Generate new certificate immediately
3. Rotate certificate using zero-downtime procedure
4. Investigate root cause (compromised CA, misconfiguration)
5. Update all affected certificates
6. File incident report

## Troubleshooting

### Common Issues

#### cert-manager not issuing certificates

```bash
# Check cert-manager status
kubectl get pods -n cert-manager
kubectl logs -n cert-manager deployment/cert-manager

# Check issuer status
kubectl describe issuer <issuer-name> -n <namespace>

# Check certificate request
kubectl get certificaterequest -n <namespace>
kubectl describe certificaterequest <cr-name> -n <namespace>
```

#### ACME challenge failures

```bash
# List challenges
kubectl get challenges --all-namespaces

# Check specific challenge
kubectl describe challenge <challenge-name> -n <namespace>

# Verify ingress for HTTP-01 challenge
kubectl get ingress -n <namespace>

# Test challenge URL accessibility
curl -v http://<domain>/.well-known/acme-challenge/<token>
```

### Zero-Downtime Rotation

All certificate rotations in this infrastructure support zero-downtime operations through:

1. **Dual-Certificate Staging**: New certificates are created alongside existing ones
2. **Gradual Traffic Migration**: Services gradually shift to new certificates
3. **Automatic Fallback**: Failed rotations automatically revert to stable certificates
4. **Health Checks**: Continuous monitoring during rotation prevents service disruption

**Rotation Testing:**
```bash
# Before rotation - record baseline
kubectl get endpoints <service> -n <namespace>
curl -o /dev/null -s -w "%{http_code}\n" https://<service-url>

# During rotation - monitor continuously
watch -n 1 'curl -o /dev/null -s -w "%{http_code}\n" https://<service-url>'

# After rotation - verify zero errors
kubectl logs -n <namespace> <pod> --since=1h | grep -i error
```

## Prometheus Monitoring

### Certificate Expiry Metrics

```prometheus
# Certificate expiration time (seconds until expiry)
cert_manager_certificate_expiration_timestamp_seconds

# Alert on certificates expiring soon
(cert_manager_certificate_expiration_timestamp_seconds - time()) < (30 * 24 * 60 * 60)
```

### Alert Definitions

```yaml
# 30-day warning
- alert: CertificateExpiringSoon
  expr: (cert_manager_certificate_expiration_timestamp_seconds - time()) < (30 * 24 * 60 * 60)
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Certificate {{ $labels.name }} expires in less than 30 days"

# 7-day critical
- alert: CertificateExpiringCritical
  expr: (cert_manager_certificate_expiration_timestamp_seconds - time()) < (7 * 24 * 60 * 60)
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Certificate {{ $labels.name }} expires in less than 7 days - URGENT"

# Certificate expired
- alert: CertificateExpired
  expr: (cert_manager_certificate_expiration_timestamp_seconds - time()) < 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Certificate {{ $labels.name }} has EXPIRED"
```

## Escalation

- **P0 (Service Down):** Immediately notify Platform team + On-call engineer
- **P1 (Degraded Service):** Create ticket + Notify Platform team
- **P2 (Warning):** Create ticket + Review in next standup

## Related Documentation

- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Kubernetes TLS Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets)
- [SRE Runbook Index](../README.md)

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2024-01-15 | SRE Team | Initial runbook creation |
