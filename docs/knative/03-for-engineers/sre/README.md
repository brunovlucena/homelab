# 🔥 SRE Engineer - Knative Lambda

**Operational excellence for serverless functions on Kubernetes**

---

## 🎯 Overview

As an SRE engineer working with Knative Lambda, you're responsible for ensuring the platform runs reliably, scales efficiently, and recovers gracefully from failures. This guide covers alert response, debugging, capacity planning, and operational best practices.

---

## 🚀 Quick Start

### 1. Access the Platform

```bash
# Set environment (dev/prd)
export ENV=dev

# Access builder service logs
kubectl logs -f deployment/knative-lambda-builder -n knative-lambda

# Check Knative services
kubectl get ksvc -n knative-lambda

# Monitor RabbitMQ queues
make rabbitmq-status ENV=${ENV}
```

### 2. Common Commands

```bash
# Port forward RabbitMQ admin UI
make pf-rabbitmq-admin ENV=${ENV}
# Access at http://localhost:15672 (guest/guest)

# Port forward Prometheus
make pf-prometheus
# Access at http://localhost:9090

# Trigger test build
make trigger-build-${ENV}

# Clean up failed jobs
kubectl delete jobs --field-selector status.successful=0 -n knative-lambda
```

### 3. Key Dashboards

| Dashboard | URL | Purpose |
|-----------|-----|---------|
| **Grafana** | `http://grafana.homelab/d/knative-lambda` | Platform metrics |
| **Prometheus** | `http://prometheus.homelab` | Raw metrics + alerts |
| **RabbitMQ** | `http://rabbitmq.homelab:15672` | Queue monitoring |
| **ArgoCD** | `http://argocd.homelab` | GitOps deployments |

---

## 📊 Key Metrics

### Golden Signals

| Signal | Metric | Threshold | Alert |
|--------|--------|-----------|-------|
| **Latency** | `build_duration_seconds` | p95 <90s | Build Slow |
| **Traffic** | `cloudevents_received_total` | >1000/min | High Load |
| **Errors** | `build_failures_total` | >5% | Build Failures |
| **Saturation** | `kaniko_jobs_running` | >50 concurrent | Resource Pressure |

### Business Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| `builds_success_rate` | % of successful builds | >95% |
| `cold_start_duration` | Time to first request | <5s |
| `scale_to_zero_time` | Time to idle → 0 pods | <5min |
| `scale_up_time` | Time 0 → active pod | <30s |

---

## 🔥 Alert Response

### Critical Alerts

#### 1. **Build Failure Rate High**
```yaml
Alert: BuildFailureRateHigh
Severity: critical
Threshold: >10% failures in 5min
```

**Response Steps**:
1. Check RabbitMQ queue depth
   ```bash
   make rabbitmq-status ENV=${ENV}
   ```
2. Inspect failed job logs
   ```bash
   kubectl get jobs -n knative-lambda | grep -v Completed
   kubectl logs job/<job-name> -n knative-lambda
   ```
3. Common causes:
   - S3 parser file not found
   - ECR push failures (rate limiting)
   - Invalid Dockerfile generation
   - Kaniko pod resource limits

**Runbook**: [Build Failure Investigation](user-stories/SRE-001-build-failure-investigation.md)

---

#### 2. **Builder Service Down**
```yaml
Alert: BuilderServiceDown
Severity: critical
Threshold: No healthy pods for 2min
```

**Response Steps**:
1. Check pod status
   ```bash
   kubectl get pods -n knative-lambda -l app=knative-lambda-builder
   kubectl describe pod <pod-name> -n knative-lambda
   ```
2. Check recent events
   ```bash
   kubectl get events -n knative-lambda --sort-by='.lastTimestamp'
   ```
3. Common causes:
   - OOMKilled (memory limit too low)
   - CrashLoopBackOff (config error)
   - ImagePullBackOff (ECR auth issue)
   - Node resource pressure

**Runbook**: [Service Recovery](user-stories/SRE-006-disaster-recovery.md)

---

#### 3. **RabbitMQ Queue Backlog**
```yaml
Alert: RabbitMQQueueBacklog
Severity: warning
Threshold: >100 messages pending for 10min
```

**Response Steps**:
1. Check queue depth
   ```bash
   make rabbitmq-status ENV=${ENV}
   ```
2. Increase builder replicas (temporary)
   ```bash
   kubectl scale deployment/knative-lambda-builder --replicas=5 -n knative-lambda
   ```
3. Identify stuck jobs
   ```bash
   kubectl get jobs -n knative-lambda --sort-by='.status.startTime'
   ```
4. Purge dead letters if needed
   ```bash
   make rabbitmq-purge-lambda-queues-${ENV}
   ```

**Runbook**: [Queue Management](user-stories/SRE-003-queue-management.md)

---

## 🔍 Debugging Techniques

### 1. Build Failures

```bash
# Find recent failed builds
kubectl get jobs -n knative-lambda \
  --field-selector status.successful=0 \
  --sort-by='.status.startTime'

# Get logs from failed Kaniko pod
kubectl logs job/<job-name> -n knative-lambda -c kaniko

# Check S3 file exists
aws s3 ls s3://knative-lambda-fusion-modules-tmp/global/parser/
```

### 2. Function Cold Start Issues

```bash
# Check Knative service configuration
kubectl get ksvc <service-name> -n knative-lambda -o yaml

# View revision status
kubectl get revision -n knative-lambda

# Check autoscaler metrics
kubectl get kpa -n knative-lambda

# Inspect pod startup logs
kubectl logs -n knative-lambda \
  -l serving.knative.dev/service=<service-name> \
  --tail=100
```

### 3. Performance Degradation

```bash
# Check resource usage
kubectl top pods -n knative-lambda

# View Prometheus metrics
# build_duration_seconds{quantile="0.95"}
# cloudevents_processing_duration_seconds
# kaniko_build_duration_seconds

# Trace specific build
# Query Tempo for trace ID from CloudEvent correlation_id
```

---

## 📈 Capacity Planning

### Current Limits

| Resource | Limit | Usage (avg) | Headroom |
|----------|-------|-------------|----------|
| **Kaniko Jobs** | 50 concurrent | 15 | 70% |
| **Builder CPU** | 2 cores | 0.5 cores | 75% |
| **Builder Memory** | 4Gi | 1.2Gi | 70% |
| **RabbitMQ Messages** | 10k queue depth | 200 | 98% |

### Scaling Recommendations

**When to scale UP**:
- Kaniko jobs >70% capacity for >10min
- RabbitMQ queue depth >1000 for >5min
- Build duration p95 >120s (Kaniko resource starved)
- Function cold start >10s consistently

**How to scale**:
```bash
# Increase Kaniko job limit (edit values.yaml)
builderService:
  kanikoJobLimit: 100  # was 50

# Increase builder replicas
kubectl scale deployment/knative-lambda-builder --replicas=3 -n knative-lambda

# Increase RabbitMQ resources (edit RabbitMQCluster CR)
```

**Runbook**: [Capacity Planning](user-stories/SRE-004-capacity-planning.md)

---

## 🛠️ Operational Tasks

### Daily Tasks
- [ ] Check alert dashboard (5min)
- [ ] Review failed builds (10min)
- [ ] Monitor RabbitMQ queue health (5min)
- [ ] Verify backup jobs completed (2min)

### Weekly Tasks
- [ ] Review capacity metrics (30min)
- [ ] Analyze build duration trends (20min)
- [ ] Update runbooks with new issues (30min)
- [ ] Test disaster recovery procedure (1hr)

### Monthly Tasks
- [ ] Capacity planning review (2hrs)
- [ ] Security patch evaluation (1hr)
- [ ] Performance optimization (4hrs)
- [ ] Documentation updates (2hrs)

---

## 📚 User Stories

| Story ID | Title | Priority | Status |
|----------|-------|----------|--------|
| **SRE-001** | [Build Failure Investigation](user-stories/SRE-001-build-failure-investigation.md) | P0 | ✅ |
| **SRE-002** | [Performance Tuning](user-stories/SRE-002-performance-tuning.md) | P1 | ✅ |
| **SRE-003** | [Queue Management](user-stories/SRE-003-queue-management.md) | P0 | ✅ |
| **SRE-004** | [Capacity Planning](user-stories/SRE-004-capacity-planning.md) | P1 | ✅ |
| **SRE-005** | [Auto-Scaling Optimization](user-stories/SRE-005-autoscaling-optimization.md) | P1 | ✅ |
| **SRE-006** | [Disaster Recovery](user-stories/SRE-006-disaster-recovery.md) | P0 | ✅ |
| **SRE-007** | [Observability Enhancement](user-stories/SRE-007-observability-enhancement.md) | P2 | ✅ |
| **SRE-008** | [Certificate Lifecycle Management](user-stories/SRE-008-certificate-lifecycle-management.md) | P0 | 🆕 |
| **SRE-009** | [Backup and Restore Operations](user-stories/SRE-009-backup-restore-operations.md) | P0 | 🆕 |
| **SRE-014** | [Security Incident Response](user-stories/SRE-014-security-incident-response.md) | P0 | 🆕 |

→ **[View All User Stories](user-stories/README.md)**

---

## 🎓 Learning Resources

### Internal Docs
- [Architecture Overview](../../04-architecture/README.md)
- [Makefile Commands](../../../Makefile) - All operational commands
- [Alert Definitions](../../../deploy/templates/alerts.yaml)
- [Prometheus Rules](../../../deploy/templates/prometheus-rules.yaml)

### External Resources
- [Knative Serving Docs](https://knative.dev/docs/serving/)
- [Kaniko Documentation](https://github.com/GoogleContainerTools/kaniko)
- [CloudEvents Spec](https://cloudevents.io/)
- [RabbitMQ Monitoring](https://www.rabbitmq.com/monitoring.html)

---

## 💡 Pro Tips

### Performance
- Enable Kaniko cache for faster rebuilds (`--cache=true`)
- Use multi-stage Dockerfiles to reduce image size
- Keep warm functions with `min-scale: 1` annotation
- Monitor ECR rate limits (push/pull)

### Reliability
- Always test rollback procedures
- Use canary deployments for platform updates
- Set aggressive resource limits to prevent noisy neighbors
- Monitor queue depth trends, not just current value

### Cost Optimization
- Scale to zero idle functions (default behavior)
- Clean up old revisions weekly
- Use spot instances for Kaniko build nodes
- Implement build caching aggressively

---

## 🚨 Escalation Path

| Severity | Contact | Response Time | Method |
|----------|---------|---------------|--------|
| **P0 (Critical)** | On-call SRE | <15min | PagerDuty |
| **P1 (High)** | SRE team lead | <1hr | Slack `#sre-oncall` |
| **P2 (Medium)** | Platform team | <4hrs | Slack `#platform` |
| **P3 (Low)** | Best effort | <24hrs | GitHub issue |

---

**Need help?** Join `#knative-lambda` on Slack or file a GitHub issue.

