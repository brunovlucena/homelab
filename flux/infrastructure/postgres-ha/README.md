# PostgreSQL High Availability for Homepage

> **100% LOCAL** - Everything runs on YOUR Kind cluster, nothing in the cloud!

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    YOUR LOCAL KIND CLUSTER                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────┐     ┌─────────────────┐                        │
│  │   PgBouncer     │     │   PgBouncer     │  ← Connection Pooling  │
│  │   (pooler-rw)   │     │   (pooler-ro)   │                        │
│  └────────┬────────┘     └────────┬────────┘                        │
│           │                       │                                  │
│           ▼                       ▼                                  │
│  ┌─────────────────┐     ┌─────────────────┐                        │
│  │   PostgreSQL    │────▶│   PostgreSQL    │  ← Streaming           │
│  │   PRIMARY       │     │   STANDBY       │    Replication         │
│  └────────┬────────┘     └─────────────────┘                        │
│           │                                                          │
│           ▼                                                          │
│  ┌─────────────────────────────────────────┐                        │
│  │              YOUR LOCAL MINIO           │  ← WAL Archiving       │
│  │         s3://postgres-backups/          │    & Backups           │
│  └─────────────────────────────────────────┘                        │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Features

| Feature | Description |
|---------|-------------|
| **High Availability** | 2 instances with automatic failover |
| **Streaming Replication** | Real-time data sync to standby |
| **WAL Archiving** | Continuous backup to local MinIO |
| **Point-in-Time Recovery** | Restore to any second within 7 days |
| **Connection Pooling** | PgBouncer for efficient connections |
| **Automatic Backups** | Every 6 hours to MinIO |
| **Monitoring** | Prometheus metrics + Grafana dashboards |
| **Alerting** | Critical alerts for failures |

## Quick Start

```bash
# 1. Ensure MinIO bucket exists
kubectl exec -n minio deploy/minio -- mc mb local/postgres-backups --ignore-existing

# 2. Apply the PostgreSQL HA stack
kubectl apply -k flux/infrastructure/postgres-ha/

# 3. Wait for cluster to be ready
kubectl wait --for=condition=Ready cluster/homepage-postgres -n postgres-ha --timeout=300s

# 4. Check cluster status
kubectl get cluster -n postgres-ha
```

## Connection Strings

### For Homepage API (via PgBouncer)

```yaml
# Update homepage deployment to use:
env:
  - name: DATABASE_HOST
    value: "homepage-postgres-pooler-rw.postgres-ha.svc"
  - name: DATABASE_PORT
    value: "5432"
  - name: DATABASE_NAME
    value: "homepage"
  - name: DATABASE_USER
    value: "postgres"
  - name: PGPASSWORD
    valueFrom:
      secretKeyRef:
        name: homepage-postgres-credentials
        key: password
```

### Direct Connection (for admin/migrations)

```bash
kubectl exec -it -n postgres-ha homepage-postgres-1 -- psql -U postgres -d homepage
```

## Backup & Recovery

### Manual Backup

```bash
# Create on-demand backup before risky operations
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Backup
metadata:
  name: homepage-postgres-manual-$(date +%Y%m%d-%H%M%S)
  namespace: postgres-ha
spec:
  cluster:
    name: homepage-postgres
  method: barmanObjectStore
EOF
```

### List Backups

```bash
kubectl get backups -n postgres-ha
```

### Point-in-Time Recovery

```bash
# Recover to specific time
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: homepage-postgres-recovery
  namespace: postgres-ha
spec:
  instances: 1
  storage:
    size: 10Gi
  bootstrap:
    recovery:
      source: homepage-postgres
      recoveryTarget:
        targetTime: "2024-12-05T12:00:00Z"  # Recover to this point
  externalClusters:
    - name: homepage-postgres
      barmanObjectStore:
        destinationPath: s3://postgres-backups/homepage
        endpointURL: http://minio.minio.svc.cluster.local:9000
        s3Credentials:
          accessKeyId:
            name: minio-backup-credentials
            key: ACCESS_KEY_ID
          secretAccessKey:
            name: minio-backup-credentials
            key: SECRET_ACCESS_KEY
EOF
```

## Failover

### Automatic Failover

CloudNativePG automatically promotes a standby to primary if the primary fails.

### Manual Switchover

```bash
# Planned switchover (zero downtime)
kubectl cnpg promote homepage-postgres homepage-postgres-2 -n postgres-ha
```

## Monitoring

### Check Cluster Health

```bash
# Cluster status
kubectl get cluster homepage-postgres -n postgres-ha -o yaml | grep -A 20 status:

# Pod status
kubectl get pods -n postgres-ha -l cnpg.io/cluster=homepage-postgres

# Replication status
kubectl exec -n postgres-ha homepage-postgres-1 -- psql -U postgres -c "SELECT * FROM pg_stat_replication;"
```

### Prometheus Metrics

Key metrics exposed:
- `cnpg_pg_replication_lag` - Replication lag in seconds
- `cnpg_pg_database_size_bytes` - Database size
- `cnpg_collector_up` - Cluster health
- `cnpg_pg_last_backup_timestamp` - Last backup time

## Troubleshooting

### Cluster Won't Start

```bash
# Check operator logs
kubectl logs -n postgres-ha -l app.kubernetes.io/name=cloudnative-pg

# Check cluster events
kubectl describe cluster homepage-postgres -n postgres-ha
```

### Replication Lag

```bash
# Check lag on standby
kubectl exec -n postgres-ha homepage-postgres-2 -- psql -U postgres -c \
  "SELECT now() - pg_last_xact_replay_timestamp() AS lag;"
```

### Backup Failing

```bash
# Check backup status
kubectl get backups -n postgres-ha

# Check MinIO connectivity
kubectl exec -n postgres-ha homepage-postgres-1 -- \
  curl -s http://minio.minio.svc.cluster.local:9000/minio/health/live
```

## Migration from Old PostgreSQL

```bash
# 1. Dump from old PostgreSQL
kubectl exec -n postgres deploy/postgres -- pg_dump -U postgres homepage > homepage_backup.sql

# 2. Wait for new cluster to be ready
kubectl wait --for=condition=Ready cluster/homepage-postgres -n postgres-ha --timeout=300s

# 3. Restore to new cluster
kubectl exec -i -n postgres-ha homepage-postgres-1 -- psql -U postgres -d homepage < homepage_backup.sql

# 4. Update homepage to use new connection string
# (see Connection Strings section above)
```

