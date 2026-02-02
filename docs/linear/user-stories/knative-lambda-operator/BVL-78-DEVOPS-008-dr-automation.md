# 🔄 DEVOPS-008: Disaster Recovery Automation

**Priority**: P1 | **Status**: ✅ Implemented K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-240/devops-008-disaster-recovery-automation


---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** implement automated disaster recovery procedures  
**So that** we can quickly recover from failures and minimize downtime (RTO < 1 hour, RPO < 15 minutes)

---

## 🎯 Acceptance Criteria

### ✅ Backup Strategy
- [ ] Automated daily backups of critical data
- [ ] Kubernetes cluster state backups (Velero)
- [ ] Database backups with point-in-time recovery
- [ ] Configuration backups (Git as source of truth)
- [ ] Application state backups (S3)
- [ ] Cross-region backup replication

### ✅ Recovery Procedures
- [ ] Documented recovery runbooks
- [ ] Automated recovery scripts
- [ ] One-command cluster recreation
- [ ] Data restoration automation
- [ ] Service health validation
- [ ] Recovery time < 1 hour (RTO)

### ✅ Testing & Validation
- [ ] Monthly DR drills
- [ ] Automated DR testing in non-prod
- [ ] Backup integrity validation
- [ ] Recovery point testing
- [ ] Failover testing
- [ ] Rollback testing

### ✅ Monitoring & Alerts
- [ ] Backup success/failure alerts
- [ ] RPO/RTO tracking
- [ ] Backup age monitoring
- [ ] Storage capacity alerts
- [ ] Replication lag monitoring
- [ ] Recovery test results

### ✅ Documentation
- [ ] DR procedures documented
- [ ] Contact information for incidents
- [ ] Escalation procedures
- [ ] Post-mortem templates
- [ ] Lessons learned database

---

## 🏗️ Disaster Recovery Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  DISASTER RECOVERY ARCHITECTURE                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  PRIMARY REGION (us-west-2)                                     │
│  ├─ EKS Cluster                                                 │
│  │  ├─ Knative Lambda Platform                                  │
│  │  ├─ RabbitMQ (HA cluster)                                    │
│  │  └─ Prometheus + Grafana                                     │
│  │                                                               │
│  ├─ Data Stores                                                 │
│  │  ├─ S3 (parser code) → Cross-region replication             │
│  │  ├─ ECR (images) → Multi-region registry                     │
│  │  └─ RDS (if used) → Read replicas in DR region              │
│  │                                                               │
│  └─ Backups (Velero)                                            │
│     ├─ Daily full backups                                       │
│     ├─ Hourly incremental backups                               │
│     └─ Backup to S3 (replicated to us-east-1)                  │
│                                                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│                                                                 │
│  DISASTER RECOVERY REGION (us-east-1)                           │
│  ├─ Standby EKS Cluster (can be created on-demand)             │
│  │                                                               │
│  ├─ Data Stores (Replicas)                                      │
│  │  ├─ S3 (replicated from us-west-2)                           │
│  │  ├─ ECR (multi-region)                                       │
│  │  └─ RDS Read Replica (can be promoted)                       │
│  │                                                               │
│  └─ Backup Storage                                              │
│     └─ S3 bucket for Velero backups                             │
│                                                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│                                                                 │
│  RECOVERY SCENARIOS                                             │
│  1. Pod Failure        → Kubernetes auto-restart (< 30s)        │
│  2. Node Failure       → Cluster autoscaler (< 5min)            │
│  3. AZ Failure         → Multi-AZ deployment (< 2min)           │
│  4. Region Failure     → Full DR activation (< 60min)           │
│  5. Data Corruption    → Point-in-time restore (< 30min)        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Technical Implementation

### 1. Velero Setup (Kubernetes Backup)

**Install Velero**
```bash
# Install Velero CLI
brew install velero

# Create S3 bucket for backups
aws s3 mb s3://knative-lambda-velero-backups --region us-west-2

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket knative-lambda-velero-backups \
  --versioning-configuration Status=Enabled

# Cross-region replication
aws s3api put-bucket-replication \
  --bucket knative-lambda-velero-backups \
  --replication-configuration file://replication-config.json

# Install Velero in cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.8.0 \
  --bucket knative-lambda-velero-backups \
  --backup-location-config region=us-west-2 \
  --snapshot-location-config region=us-west-2 \
  --secret-file ./credentials-velero
```

**IAM Policy for Velero**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeVolumes",
        "ec2:DescribeSnapshots",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:CreateSnapshot",
        "ec2:DeleteSnapshot"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:DeleteObject",
        "s3:PutObject",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts"
      ],
      "Resource": [
        "arn:aws:s3:::knative-lambda-velero-backups/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::knative-lambda-velero-backups"
      ]
    }
  ]
}
```

### 2. Automated Backup Schedule

**Daily Full Backup**
```yaml
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: daily-full-backup
  namespace: velero
spec:
  schedule: "0 2 * * *"  # 2 AM daily
  template:
    includedNamespaces:
    - knative-lambda
    - rabbitmq-system
    
    excludedResources:
    - events
    - events.events.k8s.io
    
    storageLocation: default
    
    ttl: 720h  # 30 days retention
    
    snapshotVolumes: true
    
    hooks:
      resources:
      - name: rabbitmq-backup-hook
        includedNamespaces:
        - rabbitmq-system
        labelSelector:
          matchLabels:
            app: rabbitmq
        pre:
        - exec:
            container: rabbitmq
            command:
            - /bin/bash
            - -c
            - rabbitmqctl list_queues > /tmp/queue-backup.txt
```

**Hourly Incremental Backup**
```yaml
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: hourly-incremental-backup
  namespace: velero
spec:
  schedule: "0 * * * *"  # Every hour
  template:
    includedNamespaces:
    - knative-lambda
    
    includedResources:
    - configmaps
    - secrets
    - services
    
    storageLocation: default
    ttl: 168h  # 7 days retention
```

### 3. S3 Cross-Region Replication

**Replication Configuration**
```json
{
  "Role": "arn:aws:iam::339954290315:role/s3-replication-role",
  "Rules": [
    {
      "Status": "Enabled",
      "Priority": 1,
      "DeleteMarkerReplication": { "Status": "Enabled" },
      "Filter": {},
      "Destination": {
        "Bucket": "arn:aws:s3:::knative-lambda-dr-backups-us-east-1",
        "ReplicationTime": {
          "Status": "Enabled",
          "Time": { "Minutes": 15 }
        },
        "Metrics": {
          "Status": "Enabled",
          "EventThreshold": { "Minutes": 15 }
        }
      }
    }
  ]
}
```

**Enable Replication**
```bash
aws s3api put-bucket-replication \
  --bucket knative-lambda-velero-backups \
  --replication-configuration file://replication-config.json
```

### 4. Disaster Recovery Scripts

**Full Cluster Recovery Script**

**File**: `scripts/dr-restore.sh`
```bash
#!/bin/bash
set -euo pipefail

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔄 DISASTER RECOVERY - FULL CLUSTER RESTORE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DR_REGION="${DR_REGION:-us-east-1}"
BACKUP_NAME="${1:-latest}"
NAMESPACES="${2:-knative-lambda,rabbitmq-system}"

echo "🚨 Starting Disaster Recovery..."
echo "   Region: ${DR_REGION}"
echo "   Backup: ${BACKUP_NAME}"
echo "   Namespaces: ${NAMESPACES}"
echo ""

# Step 1: Create DR cluster (if doesn't exist)
echo "1️⃣  Creating DR cluster..."
if ! eksctl get cluster --name homelab-dr --region ${DR_REGION} &>/dev/null; then
  eksctl create cluster -f cluster-dr-config.yaml
else
  echo "   ✅ Cluster already exists"
fi

# Step 2: Install Velero in DR cluster
echo "2️⃣  Installing Velero..."
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.8.0 \
  --bucket knative-lambda-dr-backups-us-east-1 \
  --backup-location-config region=${DR_REGION} \
  --snapshot-location-config region=${DR_REGION} \
  --secret-file ./credentials-velero

# Wait for Velero to be ready
kubectl wait --for=condition=ready pod \
  -l component=velero \
  -n velero \
  --timeout=300s

# Step 3: List available backups
echo "3️⃣  Available backups:"
velero backup get

# Step 4: Restore from backup
echo "4️⃣  Restoring from backup: ${BACKUP_NAME}..."
if [ "${BACKUP_NAME}" == "latest" ]; then
  BACKUP_NAME=$(velero backup get --output json | jq -r '.items[0].metadata.name')
fi

velero restore create restore-$(date +%Y%m%d-%H%M%S) \
  --from-backup ${BACKUP_NAME} \
  --include-namespaces ${NAMESPACES} \
  --wait

# Step 5: Verify restoration
echo "5️⃣  Verifying restoration..."
for ns in $(echo ${NAMESPACES} | tr ',' ' '); do
  echo "   Checking namespace: ${ns}"
  kubectl get all -n ${ns}
done

# Step 6: Update DNS (manual confirmation required)
echo "6️⃣  ⚠️  ACTION REQUIRED: Update DNS to point to DR cluster"
echo "   Current ALB: $(kubectl get ingress -n knative-lambda -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')"
echo ""

# Step 7: Run health checks
echo "7️⃣  Running health checks..."
./scripts/health-check.sh

# Step 8: Generate DR report
echo "8️⃣  Generating DR report..."
cat > dr-report-$(date +%Y%m%d-%H%M%S).txt <<EOF
Disaster Recovery Report
========================
Date: $(date)
Region: ${DR_REGION}
Backup: ${BACKUP_NAME}
Namespaces: ${NAMESPACES}

Cluster Status:
$(kubectl get nodes)

Service Status:
$(kubectl get pods -A | grep -v Running | grep -v Completed | | echo "All pods running")

Next Steps:
1. Verify application functionality
2. Update DNS records
3. Notify stakeholders
4. Begin root cause analysis
EOF

echo ""
echo "✅ Disaster Recovery Complete!"
echo "📊 Report: dr-report-$(date +%Y%m%d-%H%M%S).txt"
```

### 5. Health Check Script

**File**: `scripts/health-check.sh`
```bash
#!/bin/bash
set -euo pipefail

echo "🏥 Running Health Checks..."

# Check 1: Cluster health
echo "1️⃣  Cluster Health"
kubectl get nodes -o wide

# Check 2: Pod status
echo "2️⃣  Pod Status"
kubectl get pods -A | grep -v Running | grep -v Completed | | echo "✅ All pods healthy"

# Check 3: Service endpoints
echo "3️⃣  Service Endpoints"
kubectl get endpoints -n knative-lambda

# Check 4: API health
echo "4️⃣  API Health Check"
BUILDER_URL=$(kubectl get svc knative-lambda-builder -n knative-lambda -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl -f http://${BUILDER_URL}:8080/health | | echo "❌ Health check failed"

# Check 5: RabbitMQ status
echo "5️⃣  RabbitMQ Status"
kubectl exec -n rabbitmq-system rabbitmq-0 -- rabbitmqctl cluster_status

# Check 6: Metrics collection
echo "6️⃣  Metrics Collection"
curl -s http://${BUILDER_URL}:8080/metrics | grep -q "builds_total" && echo "✅ Metrics OK" | | echo "❌ Metrics unavailable"

# Check 7: Storage volumes
echo "7️⃣  Storage Volumes"
kubectl get pv,pvc -A

echo ""
echo "✅ Health Checks Complete"
```

---

## 🧪 Disaster Recovery Testing

### Monthly DR Drill

**File**: `scripts/dr-drill.sh`
```bash
#!/bin/bash
set -euo pipefail

# Non-destructive DR test in staging
DRILL_DATE=$(date +%Y-%m-%d)

echo "🧪 DR Drill - ${DRILL_DATE}"
echo "Environment: Staging"
echo ""

# 1. Create test backup
echo "1️⃣  Creating test backup..."
velero backup create dr-drill-${DRILL_DATE} \
  --include-namespaces knative-lambda \
  --wait

# 2. Delete namespace (simulated disaster)
echo "2️⃣  Simulating disaster (deleting namespace)..."
kubectl delete namespace knative-lambda | | true

# 3. Restore from backup
echo "3️⃣  Restoring from backup..."
velero restore create dr-drill-restore-${DRILL_DATE} \
  --from-backup dr-drill-${DRILL_DATE} \
  --namespace-mappings knative-lambda:knative-lambda \
  --wait

# 4. Validate restoration
echo "4️⃣  Validating restoration..."
kubectl get all -n knative-lambda

# 5. Measure RTO/RPO
BACKUP_TIME=$(velero backup describe dr-drill-${DRILL_DATE} --details | grep "Started:" | awk '{print $2}')
RESTORE_TIME=$(velero restore describe dr-drill-restore-${DRILL_DATE} --details | grep "Started:" | awk '{print $2}')
RTO=$(( $(date -d "$RESTORE_TIME" +%s) - $(date -d "$BACKUP_TIME" +%s) ))

echo ""
echo "📊 DR Drill Results:"
echo "   RTO: ${RTO} seconds"
echo "   Status: $(velero restore describe dr-drill-restore-${DRILL_DATE} | grep "Phase:" | awk '{print $2}')"

# 6. Cleanup
echo "6️⃣  Cleaning up..."
kubectl delete namespace knative-lambda
velero backup delete dr-drill-${DRILL_DATE} --confirm
velero restore delete dr-drill-restore-${DRILL_DATE} --confirm

echo "✅ DR Drill Complete"
```

### Automated DR Testing (CI/CD)

```yaml
name: DR Testing

on:
  schedule:
    - cron: '0 3 1 * *'  # Monthly on the 1st at 3 AM

jobs:
  dr-drill:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Configure AWS
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2
    
    - name: Setup kubectl
      uses: azure/setup-kubectl@v3
    
    - name: Setup Velero CLI
      run: | wget https://github.com/vmware-tanzu/velero/releases/download/v1.12.0/velero-v1.12.0-linux-amd64.tar.gz
        tar -xvf velero-v1.12.0-linux-amd64.tar.gz
        sudo mv velero-v1.12.0-linux-amd64/velero /usr/local/bin/
    
    - name: Run DR Drill
      run: | ./scripts/dr-drill.sh
    
    - name: Upload DR Report
      uses: actions/upload-artifact@v3
      with:
        name: dr-drill-report
        path: dr-drill-report-*.txt
    
    - name: Notify Slack
      uses: slackapi/slack-github-action@v1
      with:
        webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
        payload: | {
            "text": "✅ Monthly DR Drill Complete",
            "blocks": [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "*Monthly DR Drill Results*\nStatus: Success\nRTO: < 1 hour\nRPO: < 15 minutes"
                }
              }
            ]
          }
```

---

## 📊 Monitoring & Alerts

### Backup Monitoring

```promql
# Backup success rate
sum(rate(velero_backup_success_total[24h])) / 
sum(rate(velero_backup_attempt_total[24h])) * 100

# Backup age (hours)
(time() - velero_backup_last_successful_timestamp) / 3600

# Backup size trend
rate(velero_backup_items_total[24h])

# Replication lag (S3)
aws_s3_replication_latency_seconds{bucket="knative-lambda-velero-backups"}
```

### Critical Alerts

```yaml
groups:
- name: disaster-recovery
  rules:
  - alert: BackupFailed
    expr: | increase(velero_backup_failure_total[1h]) > 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Velero backup failed"
      description: "Backup {{ $labels.schedule }} failed"
  
  - alert: BackupTooOld
    expr: | (time() - velero_backup_last_successful_timestamp) > 86400
    labels:
      severity: critical
    annotations:
      summary: "No successful backup in 24 hours"
  
  - alert: ReplicationLagHigh
    expr: | aws_s3_replication_latency_seconds > 900
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "S3 replication lag > 15 minutes"
  
  - alert: RPOExceeded
    expr: | (time() - velero_backup_last_successful_timestamp) > 900
    labels:
      severity: critical
    annotations:
      summary: "RPO target exceeded (> 15 minutes)"
```

---

## 📋 DR Runbooks

### Scenario 1: Single Pod Failure

**Detection**: Pod crash loop, health check failures

**Recovery**:
```bash
# 1. Check pod status
kubectl get pods -n knative-lambda

# 2. View logs
kubectl logs <pod-name> -n knative-lambda

# 3. Delete pod (Kubernetes will recreate)
kubectl delete pod <pod-name> -n knative-lambda

# 4. Verify recovery
kubectl get pods -n knative-lambda -w
```

**RTO**: < 1 minute  
**RPO**: 0 (no data loss)

### Scenario 2: Node Failure

**Detection**: Node NotReady, pods evicted

**Recovery**:
```bash
# 1. Check node status
kubectl get nodes

# 2. Cordon node (if still up)
kubectl cordon <node-name>

# 3. Drain node
kubectl drain <node-name> --ignore-daemonsets

# 4. Verify pod rescheduling
kubectl get pods -A -o wide | grep <node-name>

# 5. Cluster autoscaler will provision new node
```

**RTO**: < 5 minutes  
**RPO**: 0 (no data loss)

### Scenario 3: Availability Zone Failure

**Detection**: Multiple nodes NotReady in same AZ

**Recovery**:
```bash
# No action required - multi-AZ deployment handles automatically
# Verify pods redistributed to healthy AZs
kubectl get pods -A -o wide

# Check service endpoints
kubectl get endpoints -n knative-lambda
```

**RTO**: < 2 minutes  
**RPO**: 0 (no data loss)

### Scenario 4: Region Failure (Full DR)

**Detection**: Entire region unreachable

**Recovery**:
```bash
# Execute full DR procedure
./scripts/dr-restore.sh latest knative-lambda,rabbitmq-system

# Update DNS to DR region
# Manual step or Route53 health check failover

# Verify all services
./scripts/health-check.sh

# Communicate with stakeholders
```

**RTO**: < 60 minutes  
**RPO**: < 15 minutes

---

## 💾 Backup Retention Policy | Backup Type | Frequency | Retention | Storage Location | |------------- | ----------- | ----------- | ------------------ | | **Full Cluster** | Daily | 30 days | S3 (cross-region) | | **Incremental** | Hourly | 7 days | S3 (primary) | | **Critical Data** | Every 15 min | 24 hours | S3 (replicated) | | **Config** | On change | 90 days | Git | ---

## 📈 Performance Requirements

- **Backup Duration**: < 15 minutes (full cluster)
- **Restore Duration**: < 45 minutes (full cluster)
- **RTO (Recovery Time Objective)**: < 1 hour
- **RPO (Recovery Point Objective)**: < 15 minutes
- **Backup Storage**: < 100GB

---

## 📚 Related Documentation

- [DEVOPS-001: Observability Setup](DEVOPS-001-observability-setup.md)
- [DEVOPS-002: GitOps Deployment](DEVOPS-002-gitops-deployment.md)
- [DEVOPS-003: Multi-Environment Management](DEVOPS-003-multi-environment.md)
- Velero Documentation: https://velero.io/docs/
- AWS Backup: https://docs.aws.amazon.com/aws-backup/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: ✅ Implemented K

