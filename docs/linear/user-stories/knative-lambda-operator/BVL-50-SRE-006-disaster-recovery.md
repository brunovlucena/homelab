# ⚡ SRE-006: Disaster Recovery

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-225/sre-006-disaster-recovery
**Priority**: P0
**Story Points**: 13  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-174/sre-006-disaster-recovery  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to have tested disaster recovery procedures  
**So that** we can recover from catastrophic failures with minimal data loss

---


## 🎯 Acceptance Criteria

- [ ] [ ] RTO (Recovery Time Objective): <15min
- [ ] [ ] RPO (Recovery Point Objective): <5min
- [ ] [ ] Automated backup/restore procedures
- [ ] [ ] DR drill tested quarterly
- [ ] [ ] Runbook documented and accessible
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] RTO (Recovery Time Objective): <15min
- [ ] RPO (Recovery Point Objective): <5min
- [ ] Automated backup/restore procedures
- [ ] DR drill tested quarterly
- [ ] Runbook documented and accessible

---

## 🔥 Failure Scenarios

### Scenario 1: Builder Service Complete Failure

**Trigger**: All builder pods crash (bad deployment, node failure)

**Impact**: 
- No new builds processed
- Queue backs up
- P0 severity

**Recovery**:
```bash
# 1. Rollback to previous version (GitOps)
flux suspend kustomization knative-lambda
kubectl rollout undo deployment/knative-lambda-builder -n knative-lambda

# 2. Wait for pods ready (30s)
kubectl wait --for=condition=ready pod \
  -l app=knative-lambda-builder \
  -n knative-lambda \
  --timeout=60s

# 3. Resume Flux
flux resume kustomization knative-lambda

# 4. Validate queue processing
make rabbitmq-status ENV=prd
```

**RTO**: 5 minutes  
**RPO**: 0 (no data loss, messages in RabbitMQ)

---

### Scenario 2: RabbitMQ Cluster Failure

**Trigger**: RabbitMQ cluster down, data corruption

**Impact**:
- Build events cannot be queued
- In-flight messages lost
- P0 severity

**Recovery**:
```bash
# 1. Check RabbitMQ cluster status
kubectl get rabbitmqcluster -n rabbitmq-prd

# 2. If corrupted, restore from backup
kubectl exec rabbitmq-cluster-prd-0 -n rabbitmq-prd -- \
  rabbitmqctl restore /backup/rabbitmq-2025-10-29.backup

# 3. Or recreate cluster (if no backup)
kubectl delete rabbitmqcluster rabbitmq-cluster-prd -n rabbitmq-prd
flux reconcile kustomization rabbitmq-prd

# 4. Recreate queues via builder service restart
kubectl rollout restart deployment/knative-lambda-builder -n knative-lambda
```

**RTO**: 15 minutes  
**RPO**: 5 minutes (messages not yet acked)

---

### Scenario 3: ECR Registry Unavailable

**Trigger**: AWS ECR outage, rate limiting

**Impact**:
- Cannot push built images
- Knative cannot pull images
- P1 severity

**Recovery**:
```bash
# 1. Switch to backup registry (if configured)
# Update values.yaml
builderService:
  ecrRegistry: backup-registry.io/knative-lambdas

# 2. Or wait for ECR recovery + retry failed jobs
kubectl get jobs -n knative-lambda \
  --field-selector status.successful=0 | \
  xargs -I {} kubectl delete job {} -n knative-lambda

# Requeue events via RabbitMQ replay
```

**RTO**: 30 minutes (waiting for AWS)  
**RPO**: 0 (rebuild from S3 parsers)

---

## 💾 Backup Strategy

### What to Backup | Component | Frequency | Retention | Method | |----------- | ----------- | ----------- | -------- | | RabbitMQ Definitions | Daily | 7 days | `rabbitmqctl export_definitions` | | RabbitMQ Messages | Continuous | 24 hours | Persistent volumes | | S3 Parser Files | N/A | Indefinite | S3 versioning enabled | | ECR Images | N/A | 90 days | Lifecycle policy | | Knative Services | N/A | GitOps | Git repository | | Prometheus Metrics | Hourly | 30 days | Thanos/Cortex | ### Backup Commands

```bash
# Backup RabbitMQ definitions
kubectl exec rabbitmq-cluster-prd-0 -n rabbitmq-prd -- \
  rabbitmqctl export_definitions /tmp/definitions.json

kubectl cp rabbitmq-prd/rabbitmq-cluster-prd-0:/tmp/definitions.json \
  ./backup/rabbitmq-definitions-$(date +%Y%m%d).json
```

---

## 🔄 Restore Procedures

### Restoring from Backup

**When to Restore**:
- Data corruption detected
- Accidental deletion
- Disaster recovery scenario

**Restore Steps**:

```bash
# 1. Restore RabbitMQ from backup
kubectl exec rabbitmq-cluster-prd-0 -n rabbitmq-prd -- \
  rabbitmqctl import_definitions /backup/rabbitmq-definitions-latest.json

# 2. Restore persistent volumes (if needed)
velero restore create --from-backup rabbitmq-backup-20251029

# 3. Verify restoration
kubectl exec rabbitmq-cluster-prd-0 -n rabbitmq-prd -- \
  rabbitmqctl list_queues name messages consumers

# 4. Restart builder service to reconnect
kubectl rollout restart deployment/knative-lambda-builder -n knative-lambda
```

**Validation**:
- All queues restored with correct bindings
- Consumer connections re-established  
- No message loss (compare counts)
- Builder service processing events normally

---

## 🧪 Testing DR Procedures

### DR Drill (Quarterly)

### Drill Checklist

1. [ ] Notify team of planned DR drill
2. [ ] Simulate failure (e.g., delete builder deployment)
3. [ ] Execute recovery procedure
4. [ ] Measure RTO/RPO
5. [ ] Document lessons learned
6. [ ] Update runbook

**Last Drill**: 2025-10-15  
**RTO Achieved**: 12 minutes (target: <15min ✅)  
**RPO Achieved**: 0 minutes (target: <5min ✅)

---

