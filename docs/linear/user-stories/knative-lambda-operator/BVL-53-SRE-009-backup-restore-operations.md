# SRE-009: Backup and Restore Operations

**Status**: Backlog
**Priority**: P0
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-181/sre-009-backup-and-restore-operations  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** backup and restore operations  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] Verify backup integrity
- [ ] Confirm backup contains required data
- [ ] Schedule maintenance window
- [ ] Notify stakeholders
- [ ] Document current state
- [ ] Prepare rollback plan
- [ ] Verify backup availability and integrity
- [ ] Create isolated test environment
- [ ] Execute restore procedure
- [ ] Validate RTO (Recovery Time Objective) < 30 minutes

---


## Overview

This runbook provides comprehensive procedures for backup and restore operations across all critical components of the Knative Lambda infrastructure, including databases, Kubernetes resources, configurations, and persistent volumes.

## Backup Strategy

### Backup Schedule | Component | Frequency | Retention | Backup Method | Storage Location | |----------- | ----------- | ----------- | --------------- | ------------------ | | PostgreSQL Database | Every 6 hours | 30 days | pg_dump | S3 bucket | | Kubernetes Resources | Every 2 hours | 14 days | Velero | S3 bucket | | ConfigMaps/Secrets | Every 1 hour | 7 days | etcd snapshot | S3 bucket | | Persistent Volumes | Daily | 30 days | Velero PV backup | S3 bucket | | Application State | Every 4 hours | 14 days | Custom scripts | S3 bucket | ### Backup Locations

**Primary Backup Storage:**
- **S3 Bucket:** `s3://knative-lambda-backups-primary/`
- **Region:** us-west-2
- **Encryption:** AES-256 at rest
- **Versioning:** Enabled

**Disaster Recovery Storage:**
- **S3 Bucket:** `s3://knative-lambda-backups-dr/`
- **Region:** us-east-1 (cross-region replication)
- **Encryption:** AES-256 at rest

## Component-Specific Backup Procedures

### 1. PostgreSQL Database Backup

```bash
# Manual database backup
kubectl exec -n database postgres-0 -- pg_dumpall -U postgres | \
  gzip > postgres-backup-$(date +%Y%m%d-%H%M%S).sql.gz

# Upload to S3
aws s3 cp postgres-backup-*.sql.gz \
  s3://knative-lambda-backups-primary/database/$(date +%Y-%m-%d)/

# Verify backup
aws s3 ls s3://knative-lambda-backups-primary/database/$(date +%Y-%m-%d)/

# Automated backup (CronJob)
kubectl get cronjob -n database postgres-backup
kubectl logs -n database $(kubectl get pods -n database -l job-name=postgres-backup -o name | tail -1)
```

### 2. Kubernetes Resources Backup (Velero)

```bash
# Create on-demand backup of all resources
velero backup create manual-backup-$(date +%Y%m%d-%H%M%S) \
  --include-namespaces knative-serving,lambda-system,database \
  --snapshot-volumes

# Create namespace-specific backup
velero backup create lambda-backup-$(date +%Y%m%d) \
  --include-namespaces lambda-system \
  --include-resources deployments,services,configmaps,secrets

# Check backup status
velero backup describe <backup-name>

# List all backups
velero backup get

# Verify backup completeness
velero backup logs <backup-name>
```

### 3. etcd Snapshot (ConfigMaps/Secrets)

```bash
# Create etcd snapshot
ETCDCTL_API=3 etcdctl snapshot save /backup/etcd-snapshot-$(date +%Y%m%d-%H%M%S).db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key

# Verify snapshot
ETCDCTL_API=3 etcdctl snapshot status /backup/etcd-snapshot-*.db -w table

# Upload to S3
aws s3 cp /backup/etcd-snapshot-*.db \
  s3://knative-lambda-backups-primary/etcd/$(date +%Y-%m-%d)/
```

### 4. Persistent Volume Backup

```bash
# List all PVCs
kubectl get pvc --all-namespaces

# Create Velero backup with volume snapshots
velero backup create pv-backup-$(date +%Y%m%d) \
  --snapshot-volumes \
  --include-namespaces lambda-system,database

# Check volume snapshot status
velero backup describe pv-backup-$(date +%Y%m%d) --details

# List volume snapshots
kubectl get volumesnapshots --all-namespaces
```

## Restore Procedures

### Pre-Restore Checklist

- [ ] Verify backup integrity
- [ ] Confirm backup contains required data
- [ ] Schedule maintenance window
- [ ] Notify stakeholders
- [ ] Document current state
- [ ] Prepare rollback plan

### 1. PostgreSQL Database Restore

```bash
# Step 1: Scale down applications using the database
kubectl scale deployment -n lambda-system --replicas=0 --all

# Step 2: Download backup from S3
aws s3 cp s3://knative-lambda-backups-primary/database/<date>/postgres-backup.sql.gz .

# Step 3: Extract backup
gunzip postgres-backup.sql.gz

# Step 4: Restore database
kubectl exec -n database postgres-0 -- psql -U postgres < postgres-backup.sql

# Step 5: Verify restore
kubectl exec -n database postgres-0 -- psql -U postgres -c "\l"
kubectl exec -n database postgres-0 -- psql -U postgres -d <dbname> -c "SELECT COUNT(*) FROM <table>;"

# Step 6: Scale applications back up
kubectl scale deployment -n lambda-system --replicas=3 --all

# Step 7: Verify applications
kubectl get pods -n lambda-system
kubectl logs -n lambda-system <pod-name>
```

### 2. Kubernetes Resources Restore (Velero)

```bash
# List available backups
velero backup get

# Restore entire backup
velero restore create restore-$(date +%Y%m%d) \
  --from-backup <backup-name>

# Restore specific namespace
velero restore create restore-lambda-$(date +%Y%m%d) \
  --from-backup <backup-name> \
  --include-namespaces lambda-system

# Restore specific resources
velero restore create restore-configmaps-$(date +%Y%m%d) \
  --from-backup <backup-name> \
  --include-resources configmaps,secrets \
  --include-namespaces lambda-system

# Monitor restore progress
velero restore describe <restore-name>

# Check restore logs
velero restore logs <restore-name>

# Verify restored resources
kubectl get all -n lambda-system
```

### 3. etcd Restore

```bash
# Stop kube-apiserver
systemctl stop kube-apiserver

# Restore etcd from snapshot
ETCDCTL_API=3 etcdctl snapshot restore /backup/etcd-snapshot.db \
  --data-dir=/var/lib/etcd-restored \
  --initial-cluster=etcd-0=https://10.0.0.1:2380 \
  --initial-advertise-peer-urls=https://10.0.0.1:2380

# Update etcd data directory
mv /var/lib/etcd /var/lib/etcd-old
mv /var/lib/etcd-restored /var/lib/etcd

# Restart etcd
systemctl restart etcd

# Start kube-apiserver
systemctl start kube-apiserver

# Verify cluster health
kubectl cluster-info
kubectl get nodes
```

### 4. Persistent Volume Restore

```bash
# Create restore from backup with PV
velero restore create pv-restore-$(date +%Y%m%d) \
  --from-backup pv-backup-<date> \
  --restore-volumes

# Monitor PV restore
kubectl get pvc -n lambda-system -w

# Verify PV data
kubectl exec -n lambda-system <pod-name> -- ls -la /data
```

## Point-in-Time Restore

### Database Point-in-Time Recovery (PITR)

```bash
# Step 1: Identify target timestamp
TARGET_TIME="2024-01-15 14:30:00"

# Step 2: Find backup before target time
aws s3 ls s3://knative-lambda-backups-primary/database/ --recursive | \
  grep "$(date -d "$TARGET_TIME" +%Y-%m-%d)"

# Step 3: Restore base backup
kubectl exec -n database postgres-0 -- psql -U postgres < base-backup.sql

# Step 4: Apply WAL logs up to target time
kubectl exec -n database postgres-0 -- \
  pg_waldump <wal-file> --rmgr=Transaction --start=$TARGET_TIME

# Step 5: Verify data at target time
kubectl exec -n database postgres-0 -- \
  psql -U postgres -d <dbname> -c "SELECT * FROM <table> WHERE timestamp <= '$TARGET_TIME';"
```

### Kubernetes Resources Point-in-Time

```bash
# List backups around target time
velero backup get | grep "2024-01-15"

# Restore from closest backup before target
velero restore create pitr-restore-$(date +%Y%m%d) \
  --from-backup backup-20240115-143000

# Verify resource states
kubectl get all -n lambda-system -o yaml | grep -A5 "creationTimestamp"
```

### RabbitMQ Message Queue Restore

```bash
# Step 1: Export RabbitMQ definitions (vhosts, exchanges, queues, bindings)
kubectl exec -n messaging rabbitmq-0 -- rabbitmqadmin export /tmp/rabbitmq-backup.json

# Step 2: Copy backup to local
kubectl cp messaging/rabbitmq-0:/tmp/rabbitmq-backup.json ./rabbitmq-backup-$(date +%Y%m%d).json

# Step 3: Upload to S3
aws s3 cp rabbitmq-backup-*.json s3://knative-lambda-backups-primary/rabbitmq/$(date +%Y-%m-%d)/

# Step 4: Restore RabbitMQ definitions
kubectl cp ./rabbitmq-backup.json messaging/rabbitmq-0:/tmp/rabbitmq-restore.json
kubectl exec -n messaging rabbitmq-0 -- rabbitmqadmin import /tmp/rabbitmq-restore.json

# Step 5: Verify queues and exchanges
kubectl exec -n messaging rabbitmq-0 -- rabbitmqadmin list queues
kubectl exec -n messaging rabbitmq-0 -- rabbitmqadmin list exchanges
```

### Full Cluster Restore

Complete disaster recovery procedure (see Scenario 1: Complete Cluster Loss above).

**RTO:** 2 hours  
**RPO:** 2 hours

```bash
# 1. Provision new cluster (30 min)
# 2. Install Velero (10 min)
# 3. Restore all namespaces (45 min)
# 4. Restore database (30 min)
# 5. Verify and test (5 min)
```

### Partial Restore

Restore specific components without affecting the rest of the system.

```bash
# Restore specific namespace only
velero restore create partial-restore-lambda \
  --from-backup latest-backup \
  --include-namespaces lambda-system

# Restore specific resource types
velero restore create partial-restore-configmaps \
  --from-backup latest-backup \
  --include-resources configmaps,secrets

# Restore individual resources
kubectl apply -f backup-configmap.yaml
kubectl apply -f backup-secret.yaml
```

## Backup Validation and Testing

### Monthly Restore Drill Procedure

**Schedule:** First Monday of each month, 2:00 AM UTC

**Drill Checklist:**
- [ ] Verify backup availability and integrity
- [ ] Create isolated test environment
- [ ] Execute restore procedure
- [ ] Validate RTO (Recovery Time Objective) < 30 minutes
- [ ] Validate RPO (Recovery Point Objective) < 6 hours
- [ ] Test application functionality post-restore
- [ ] Document any issues or improvements
- [ ] Clean up test resources

```bash
# Step 1: Create isolated test namespace
kubectl create namespace restore-drill-$(date +%Y%m%d)

# Step 2: Restore backup to test namespace
velero restore create drill-restore-$(date +%Y%m%d) \
  --from-backup latest-backup \
  --namespace-mappings lambda-system:restore-drill-$(date +%Y%m%d)

# Step 3: Verify all resources restored correctly
kubectl get all -n restore-drill-$(date +%Y%m%d)

# Step 4: Run application health checks
kubectl exec -n restore-drill-$(date +%Y%m%d) <pod> -- /health-check.sh

# Step 5: Document results
echo "Restore drill completed: $(date)" >> /var/log/restore-drills.log
echo "Resources restored: $(kubectl get all -n restore-drill-$(date +%Y%m%d) | wc -l)" >> /var/log/restore-drills.log

# Step 6: Cleanup test namespace
kubectl delete namespace restore-drill-$(date +%Y%m%d)
```

### Backup Integrity Verification

```bash
# Verify database backup
gunzip -t postgres-backup.sql.gz && echo "Database backup integrity: OK"

# Verify Velero backup
velero backup describe <backup-name> | grep "Phase: Completed"

# Verify etcd snapshot
ETCDCTL_API=3 etcdctl snapshot status etcd-snapshot.db -w table

# Verify S3 backup checksums
aws s3api head-object \
  --bucket knative-lambda-backups-primary \
  --key database/2024-01-15/postgres-backup.sql.gz \
  --query 'ETag' --output text
```

## Disaster Recovery Scenarios

### Scenario 1: Complete Cluster Loss

```bash
# Step 1: Provision new cluster
# (Follow cluster provisioning runbook)

# Step 2: Install Velero in new cluster
velero install \
  --provider aws \
  --bucket knative-lambda-backups-primary \
  --secret-file ./credentials-velero

# Step 3: Restore all namespaces
velero restore create dr-full-restore \
  --from-backup latest-full-backup

# Step 4: Restore database
# (Follow PostgreSQL restore procedure above)

# Step 5: Update DNS to point to new cluster

# Step 6: Verify all services
kubectl get pods --all-namespaces
curl https://api.knative-lambda.com/health
```

### Scenario 2: Namespace Corruption

```bash
# Step 1: Delete corrupted namespace
kubectl delete namespace lambda-system --grace-period=0 --force

# Step 2: Restore namespace from backup
velero restore create namespace-restore \
  --from-backup latest-backup \
  --include-namespaces lambda-system

# Step 3: Verify resources
kubectl get all -n lambda-system
```

### Scenario 3: Data Corruption

```bash
# Step 1: Identify corruption scope
kubectl logs -n lambda-system <pod> | grep -i "corruption\ | error"

# Step 2: Restore from last known good backup
velero restore create data-restore \
  --from-backup backup-<timestamp> \
  --include-resources persistentvolumeclaims

# Step 3: Verify data integrity
kubectl exec -n lambda-system <pod> -- /verify-data.sh
```

## Monitoring and Alerts

### Backup Health Metrics

```prometheus
# Backup success rate
rate(backup_success_total[1h]) / rate(backup_attempts_total[1h])

# Backup duration
histogram_quantile(0.95, backup_duration_seconds_bucket)

# Backup size
backup_size_bytes{component="database"}

# Last successful backup age
time() - backup_last_success_timestamp_seconds
```

### Critical Alerts

**Alert: BackupFailed**
```yaml
alert: BackupFailed
expr: backup_success_total == 0
for: 1h
severity: critical
description: Backup has failed for component {{ $labels.component }}
```

**Alert: BackupTooOld**
```yaml
alert: BackupTooOld
expr: (time() - backup_last_success_timestamp_seconds) > 86400
for: 1h
severity: warning
description: Last successful backup is older than 24 hours for {{ $labels.component }}
```

## Troubleshooting

### Common Backup Issues

#### Velero backup stuck in "InProgress"

```bash
# Check Velero logs
kubectl logs -n velero deployment/velero

# Check backup details
velero backup describe <backup-name> --details

# Cancel stuck backup
velero backup delete <backup-name> --confirm

# Restart Velero
kubectl rollout restart deployment/velero -n velero
```

#### Database backup fails with "disk full"

```bash
# Check available space
kubectl exec -n database postgres-0 -- df -h

# Clean old backups
kubectl exec -n database postgres-0 -- find /backups -mtime +30 -delete

# Increase PVC size
kubectl patch pvc postgres-backup -n database \
  -p '{"spec":{"resources":{"requests":{"storage":"200Gi"}}}}'
```

#### etcd snapshot fails

```bash
# Check etcd health
ETCDCTL_API=3 etcdctl endpoint health

# Verify certificates
ls -la /etc/kubernetes/pki/etcd/

# Check disk space on etcd node
df -h /var/lib/etcd
```

## Escalation

- **P0 (Backup System Down):** Immediately notify Platform team + On-call SRE
- **P1 (Backup Failure):** Create incident ticket + Notify Platform team within 1 hour
- **P2 (Backup Warning):** Create ticket + Review in next standup

## Related Documentation

- [Velero Documentation](https://velero.io/docs/)
- [PostgreSQL Backup Best Practices](https://www.postgresql.org/docs/current/backup.html)
- [etcd Disaster Recovery](https://etcd.io/docs/v3.5/op-guide/recovery/)
- [SRE Runbook Index](../README.md)

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2024-01-15 | SRE Team | Initial runbook creation |
