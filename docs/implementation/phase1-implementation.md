# Phase 1 Implementation - Week-by-Week Breakdown

> **Part of**: [Homelab Documentation](../README.md) → Implementation  
> **Last Updated**: November 7, 2025

---

## Overview

Detailed week-by-week breakdown of Phase 1 tasks to achieve 94% production readiness over 12-16 weeks (204 hours).

## Week 1: Kick-off & Secret Management

### Goals

- ✅ Deploy External Secrets Operator
- ✅ Migrate secrets to GitHub repository settings

### Tasks

1. **Deploy ESO via Pulumi/Flux**
   - Create Pulumi stack for ESO
   - Deploy to all clusters (Air, Pro, Studio, Pi, Forge)
   - Verify installation

2. **Store secrets in GitHub repository settings**
   ```bash
   # Add secrets to GitHub repository
   gh secret set AIR_DATABASE_USERNAME --body "airuser"
   gh secret set AIR_DATABASE_PASSWORD --body "airpass"
   gh secret set PRO_DATABASE_USERNAME --body "prouser"
   gh secret set PRO_DATABASE_PASSWORD --body "propass"
   gh secret set STUDIO_DATABASE_USERNAME --body "studiouser"
   gh secret set STUDIO_DATABASE_PASSWORD --body "studiopass"
   ```

3. **Migrate `.zshrc` secrets to GitHub**
   - Identify all secrets in `.zshrc`
   - Add to GitHub repository secrets
   - Test retrieval via ESO

4. **Configure ESO ClusterSecretStore**
   ```yaml
   apiVersion: external-secrets.io/v1beta1
   kind: ClusterSecretStore
   metadata:
     name: github-backend
   spec:
     provider:
       github:
         owner: <org-name>
         repo: <repo-name>
         auth:
           appAuth:
             appId:
               secretRef:
                 name: github-app-credentials
                 key: appId
             installationId:
               secretRef:
                 name: github-app-credentials
                 key: installationId
             privateKey:
               secretRef:
                 name: github-app-credentials
                 key: privateKey
   ```

5. **Test secret sync across Air, Pro, Studio**
   - Deploy test ExternalSecret
   - Verify Kubernetes secret created
   - Test application can read secret

### Deliverables

- ESO operational on all clusters
- All secrets managed in GitHub repository settings
- Documentation: "Secret Management Guide"

### Validation

```bash
# Verify ESO is running
kubectl get pods -n external-secrets

# Test secret sync
kubectl get secretstores -A
kubectl get externalsecrets -A

# Verify secrets are synced
kubectl get secrets -A | grep external-secrets
```

### Time Investment

**8 hours**

---

## Week 2-3: Backup & Disaster Recovery

### Goals

- ✅ Deploy Velero
- ✅ Automated backups operational
- ✅ DR procedures tested

### Tasks

1. **Deploy Velero on all clusters**
   ```bash
   # Install Velero via Pulumi
   # Configure MinIO as backup storage
   ```

2. **Configure backup schedules**
   ```yaml
   # Daily backup
   apiVersion: velero.io/v1
   kind: Schedule
   metadata:
     name: daily-backup
     namespace: velero
   spec:
     schedule: "0 2 * * *"  # 2 AM daily
     template:
       ttl: 168h  # 7 days retention
       includedNamespaces:
       - "*"
   
   # Weekly backup
   apiVersion: velero.io/v1
   kind: Schedule
   metadata:
     name: weekly-backup
     namespace: velero
   spec:
     schedule: "0 3 * * 0"  # 3 AM Sunday
     template:
       ttl: 720h  # 30 days retention
       includedNamespaces:
       - "*"
   ```

3. **Test restore procedures**
   ```bash
   # Create test namespace and resources
   kubectl create namespace test-backup
   kubectl create deployment nginx --image=nginx -n test-backup
   
   # Backup
   velero backup create test-backup --include-namespaces test-backup
   
   # Delete namespace
   kubectl delete namespace test-backup
   
   # Restore
   velero restore create --from-backup test-backup
   
   # Verify
   kubectl get all -n test-backup
   ```

4. **Document DR runbook**
   - RTO: <4 hours
   - RPO: <1 hour
   - Escalation procedures
   - Contact information

### Deliverables

- Velero operational on all clusters
- Automated daily backups
- RTO <4h, RPO <1h
- Runbook: "Disaster Recovery Procedures"

### Validation

```bash
# Verify backups
velero backup get
velero schedule get

# Check backup storage
velero backup-location get

# Test restore
velero restore create --from-backup daily-20251107
```

### Time Investment

**16 hours**

---

## Week 3-4: Alerting & SLO Tracking

### Goals

- ✅ AlertManager deployed
- ✅ PagerDuty integration
- ✅ SLO tracking operational

### Tasks

1. **Deploy AlertManager via Pulumi**
   ```yaml
   global:
     resolve_timeout: 5m
   
   route:
     group_by: ['alertname', 'cluster', 'service']
     group_wait: 10s
     group_interval: 10s
     repeat_interval: 12h
     receiver: 'pagerduty'
   
   receivers:
   - name: 'pagerduty'
     pagerduty_configs:
     - service_key: '<pagerduty-integration-key>'
   ```

2. **Configure PagerDuty integration**
   - Create PagerDuty service
   - Generate integration key
   - Configure AlertManager receiver

3. **Deploy Sloth for SLO generation**
   ```yaml
   apiVersion: sloth.slok.dev/v1
   kind: PrometheusServiceLevel
   metadata:
     name: agent-bruno-availability
     namespace: ai-agents
   spec:
     service: agent-bruno
     slos:
       - name: "requests-availability"
         objective: 99.9
         description: "99.9% of requests should succeed"
         sli:
           events:
             errorQuery: sum(rate(http_requests_total{job="agent-bruno",code=~"5.."}[5m]))
             totalQuery: sum(rate(http_requests_total{job="agent-bruno"}[5m]))
         alerting:
           name: AgentBrunoHighErrorRate
           labels:
             severity: page
   ```

4. **Create 15+ alert rules**
   - HighCPUUsage (>80%)
   - HighMemoryUsage (>85%)
   - PodCrashLoopBackOff
   - PVCAlmostFull (>80%)
   - CertificateExpiringSoon (<30 days)
   - BackupFailed
   - LinkerdGatewayDown
   - ExternalSecretsOperatorDown
   - HighErrorRate (>5%)
   - HighLatency (P95 >1s)
   - PodNotReady
   - NodeNotReady
   - DiskPressure
   - MemoryPressure
   - NetworkErrors

5. **Test escalation procedures**
   - Trigger test alert
   - Verify PagerDuty notification
   - Test on-call rotation

### Deliverables

- AlertManager operational
- PagerDuty receiving alerts
- SLO tracking for critical services
- Runbook: "Incident Response Procedures"

### Validation

```bash
# Verify AlertManager
kubectl get pods -n observability -l app=alertmanager

# Test alert
curl -X POST http://alertmanager:9093/api/v1/alerts -d '[{
  "labels": {"alertname": "TestAlert", "severity": "warning"},
  "annotations": {"summary": "Test alert from Week 3-4"}
}]'

# Check PagerDuty incident created
```

### Time Investment

**16 hours**

---

## Week 5-6: GitHub Actions Workflows

### Goals

- ✅ CI/CD pipelines operational
- ✅ Self-hosted runners configured
- ✅ Automated deployments to Air/Pro

### Tasks

1. **Create build-and-test.yml**
   ```yaml
   name: Build and Test
   
   on:
     push:
       branches: ['main', 'develop', 'feature/*']
     pull_request:
       branches: ['main', 'develop']
   
   jobs:
     lint:
       runs-on: [self-hosted]
       steps:
         - uses: actions/checkout@v4
         - name: Lint code
           run: |
             golangci-lint run
             eslint .
     
     test:
       runs-on: [self-hosted]
       steps:
         - uses: actions/checkout@v4
         - name: Run tests
           run: |
             make test
             make test-integration
     
     build:
       runs-on: [self-hosted]
       needs: [lint, test]
       steps:
         - uses: actions/checkout@v4
         - name: Build images
           run: |
             make build-all
             make tag-version VERSION=${{ github.sha }}
   ```

2. **Create security-scan.yml**
   - Trivy container scanning
   - SAST with SonarQube
   - Dependency vulnerability checking

3. **Create deploy-air.yml** (automatic)
4. **Create deploy-pro.yml** (automatic after tests pass)
5. **Create deploy-studio.yml** (manual approval)
6. **Create rollback.yml**

### Deliverables

- 6 GitHub Actions workflows
- Self-hosted runners configured
- Documentation: "CI/CD Pipeline Guide"

### Time Investment

**24 hours**

---

## Week 6-7: Container Security Scanning

### Goals

- ✅ Trivy integrated in CI/CD
- ✅ Security reports generated
- ✅ HIGH/CRITICAL CVEs blocked

### Tasks

1. **Integrate Trivy into all workflows**
2. **Scan all images for vulnerabilities**
3. **Configure policy to block HIGH/CRITICAL**
4. **Generate security reports**

### Time Investment

**16 hours**

---

## Week 7-8: Smoke Tests & Preview Environments

### Goals

- ✅ Automated smoke tests
- ✅ Preview environments for PRs

### Tasks

1. **Create smoke test suite**
2. **Automate preview environment creation**
3. **Integrate with GitHub PR status checks**

### Time Investment

**24 hours**

---

## Week 9-10: Unit & Integration Tests

### Goals

- ✅ 80% unit test coverage
- ✅ Integration tests operational

### Tasks

1. **Write unit tests for all services**
2. **Create integration test suite**
3. **Add test results to CI/CD**

### Time Investment

**40 hours**

---

## Week 10-11: E2E Tests

### Goals

- ✅ E2E test suite operational
- ✅ Browser testing automated

### Tasks

1. **Playwright for web UI testing**
2. **API E2E tests with Postman/Newman**
3. **Cross-cluster connectivity tests**

### Time Investment

**32 hours**

---

## Week 11-12: Chaos Engineering

### Goals

- ✅ Chaos Mesh operational
- ✅ 8+ chaos experiments executed

### Tasks

1. **Deploy Chaos Mesh**
2. **Run chaos experiments**
3. **Document findings**
4. **Implement improvements**

### Chaos Experiments

- Pod failure
- Network latency injection
- Network partition
- CPU stress
- Memory stress
- Disk fill
- Clock skew

### Time Investment

**28 hours**

---

## Summary

**Total Time**: 204 hours over 12-16 weeks

**Production Readiness**: 62% → 94%

**Blockers Resolved**: 6/6

## Related Documentation

- [Operational Maturity Roadmap](operational-maturity-roadmap.md)
- [Phase 2 Preparation](phase2-preparation.md)
- [Production Readiness Analysis](../analysis/production-readiness.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

