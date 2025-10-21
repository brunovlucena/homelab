# 🚨 Runbook: PostgreSQL Storage Full

## Alert Information

**Alert Name:** `PostgreSQLStorageFull`  
**Severity:** Critical  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** Disk usage > 85% of PVC size

## Symptom

PostgreSQL persistent volume approaching or at capacity, preventing new data writes and potentially causing database corruption.

## Impact

- **User Impact:** CRITICAL - Cannot write new data, application failures
- **Business Impact:** CRITICAL - Data loss risk, service outage
- **Data Impact:** CRITICAL - Risk of database corruption, transaction loss

## Diagnosis

### 1. Check Current Disk Usage

```bash
# Check disk usage in pod
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data

# Expected output format:
# Filesystem      Size  Used Avail Use% Mounted on
# /dev/sda1        20G   18G    2G  90% /var/lib/postgresql/data
```

### 2. Check PVC Status

```bash
# Check PVC details
kubectl get pvc -n postgres

# Get PVC size and status
kubectl describe pvc -n postgres data-postgres-postgresql-0

# Check storage class
kubectl get pvc -n postgres data-postgres-postgresql-0 -o jsonpath='{.spec.storageClassName}'
```

### 3. Identify What's Using Space

```bash
# Check database sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, 
   pg_size_pretty(pg_database_size(datname)) AS size,
   pg_database_size(datname) AS size_bytes
   FROM pg_database
   ORDER BY pg_database_size(datname) DESC;"

# Check table sizes in bruno_site database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename,
   pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
   pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
   pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
   FROM pg_tables
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
   ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
   LIMIT 20;"

# Check WAL (Write-Ahead Log) size
kubectl exec -n postgres postgres-postgresql-0 -- du -sh /var/lib/postgresql/data/pg_wal

# Check individual WAL files
kubectl exec -n postgres postgres-postgresql-0 -- ls -lh /var/lib/postgresql/data/pg_wal | head -20
```

### 4. Check WAL Configuration

```bash
# Check WAL settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT name, setting, unit, context 
   FROM pg_settings 
   WHERE name IN ('wal_level', 'max_wal_size', 'min_wal_size', 'wal_keep_size', 'archive_mode');"

# Check archive status (if enabled)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT archived_count, failed_count, last_archived_time, last_failed_time 
   FROM pg_stat_archiver;"
```

### 5. Check Table Bloat

```bash
# Estimate table bloat
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename,
   pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
   n_dead_tup, n_live_tup,
   round(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_tuple_percent
   FROM pg_stat_user_tables
   WHERE n_dead_tup > 1000
   ORDER BY n_dead_tup DESC
   LIMIT 10;"

# Check last vacuum times
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename,
   last_vacuum, last_autovacuum,
   last_analyze, last_autoanalyze
   FROM pg_stat_user_tables
   ORDER BY last_autovacuum NULLS FIRST
   LIMIT 10;"
```

### 6. Check for Temporary Files

```bash
# Check temporary file usage
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, temp_files, 
   pg_size_pretty(temp_bytes) AS temp_size
   FROM pg_stat_database
   WHERE temp_files > 0
   ORDER BY temp_bytes DESC;"

# Check for large queries creating temp files
kubectl exec -n postgres postgres-postgresql-0 -- du -sh /var/lib/postgresql/data/base/pgsql_tmp*
```

## Resolution Steps

### Step 1: Immediate Actions

#### Clean WAL Files (If Safe)

```bash
# Check current WAL usage
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_current_wal_lsn(), pg_walfile_name(pg_current_wal_lsn());"

# If archive_mode is off and no replication, old WAL can be removed
# PostgreSQL should do this automatically, but if not:

# Force checkpoint to flush WAL
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "CHECKPOINT;"

# Check WAL size after checkpoint
kubectl exec -n postgres postgres-postgresql-0 -- du -sh /var/lib/postgresql/data/pg_wal
```

**⚠️ WARNING:** Only manually delete WAL files if you understand the risks!

#### Remove Old Logs

```bash
# Check log file sizes
kubectl exec -n postgres postgres-postgresql-0 -- du -sh /var/lib/postgresql/data/log

# If logs are large, consider truncating (or configure log rotation)
kubectl exec -n postgres postgres-postgresql-0 -- find /var/lib/postgresql/data/log -type f -name "*.log" -mtime +7 -delete
```

### Step 2: Clean Up Databases

#### Drop Unused Databases

```bash
# List all databases
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "\l"

# Check database sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, pg_size_pretty(pg_database_size(datname)) AS size
   FROM pg_database
   ORDER BY pg_database_size(datname) DESC;"

# Drop unused databases (⚠️ CAREFUL!)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "DROP DATABASE old_unused_db;"
```

#### Truncate Old Data

```bash
# For tables with time-series data, delete old records
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "DELETE FROM logs WHERE created_at < now() - interval '90 days';"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "DELETE FROM events WHERE created_at < now() - interval '30 days';"

# After deletion, reclaim space with VACUUM FULL
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM FULL logs;"
```

### Step 3: Run VACUUM to Reclaim Space

#### Regular VACUUM

```bash
# Run VACUUM ANALYZE (doesn't lock table, doesn't immediately free space)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM ANALYZE;"

# Vacuum specific bloated tables
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM VERBOSE large_table;"
```

#### VACUUM FULL (Last Resort)

```bash
# VACUUM FULL reclaims all space but locks table (⚠️ DOWNTIME!)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM FULL VERBOSE large_table;"

# Check space reclaimed
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data
```

### Step 4: Increase PVC Size

#### Check if Storage Class Supports Expansion

```bash
# Check storage class allows volume expansion
kubectl get sc -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.allowVolumeExpansion}{"\n"}{end}'

# Check current PVC
kubectl get pvc -n postgres data-postgres-postgresql-0 -o yaml | grep -A 3 "resources:"
```

#### Expand PVC

```bash
# Edit PVC to increase size (if allowVolumeExpansion: true)
kubectl edit pvc -n postgres data-postgres-postgresql-0

# Change spec.resources.requests.storage from 20Gi to 40Gi
# spec:
#   resources:
#     requests:
#       storage: 40Gi

# Wait for expansion to complete
kubectl get pvc -n postgres data-postgres-postgresql-0 -w

# May need to delete and recreate pod for filesystem to resize
kubectl delete pod -n postgres postgres-postgresql-0

# Wait for pod to be ready
kubectl wait --for=condition=ready pod -n postgres postgres-postgresql-0 --timeout=300s

# Verify new size
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data
```

#### If Storage Class Doesn't Support Expansion

```bash
# Create new larger PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-postgres-postgresql-0-new
  namespace: postgres
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: local-path  # your storage class
  resources:
    requests:
      storage: 40Gi
EOF

# Backup database
kubectl exec -n postgres postgres-postgresql-0 -- pg_dumpall -U postgres > /tmp/postgres-backup-$(date +%Y%m%d).sql

# Scale down StatefulSet
kubectl scale statefulset -n postgres postgres-postgresql --replicas=0

# Update StatefulSet to use new PVC
kubectl edit statefulset -n postgres postgres-postgresql
# Update volumeClaimTemplates or manually edit pod spec

# Scale up with new PVC
kubectl scale statefulset -n postgres postgres-postgresql --replicas=1

# Restore data
kubectl cp /tmp/postgres-backup-*.sql postgres/postgres-postgresql-0:/tmp/restore.sql
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres < /tmp/restore.sql
```

### Step 5: Archive Old Data

#### Export Old Data

```bash
# Export old data to file
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "COPY (SELECT * FROM logs WHERE created_at < '2024-01-01') TO '/tmp/old_logs.csv' WITH CSV HEADER;"

# Copy file out of pod
kubectl cp postgres/postgres-postgresql-0:/tmp/old_logs.csv ./archive/old_logs-$(date +%Y%m%d).csv

# Delete old data after verification
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "DELETE FROM logs WHERE created_at < '2024-01-01';"

# Reclaim space
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM FULL logs;"
```

#### Implement Table Partitioning

```sql
-- For large tables, consider partitioning
-- Example: Partition logs table by month
CREATE TABLE logs_new (
  id BIGSERIAL,
  created_at TIMESTAMP NOT NULL,
  message TEXT
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE logs_2025_01 PARTITION OF logs_new
  FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE logs_2025_02 PARTITION OF logs_new
  FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Migrate data
INSERT INTO logs_new SELECT * FROM logs;

-- Swap tables
ALTER TABLE logs RENAME TO logs_old;
ALTER TABLE logs_new RENAME TO logs;

-- Later, drop old partitions easily
DROP TABLE logs_2024_01;  -- Much faster than DELETE
```

### Step 6: Configure WAL Management

```bash
# Adjust WAL settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_wal_size = '2GB';"  # Default: 1GB

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET min_wal_size = '80MB';"  # Default: 80MB

# Restart to apply
kubectl delete pod -n postgres postgres-postgresql-0
```

## Verification

### 1. Check Disk Usage Decreased

```bash
# Check current usage
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data

# Should show reduced usage or larger capacity
```

### 2. Verify Database Sizes

```bash
# Check database sizes after cleanup
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, pg_size_pretty(pg_database_size(datname)) AS size
   FROM pg_database
   ORDER BY pg_database_size(datname) DESC;"
```

### 3. Check WAL Size

```bash
# WAL directory should be reasonable size
kubectl exec -n postgres postgres-postgresql-0 -- du -sh /var/lib/postgresql/data/pg_wal

# Should be < max_wal_size setting
```

### 4. Verify PVC Expansion

```bash
# If PVC was expanded, verify new size
kubectl get pvc -n postgres data-postgres-postgresql-0

# Check filesystem size
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data
```

### 5. Test Database Operations

```bash
# Test write operations
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE TABLE test_write (id serial, data text, created_at timestamp DEFAULT now());"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "INSERT INTO test_write (data) VALUES ('test');"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "DROP TABLE test_write;"
```

## Prevention

### 1. Set Up Disk Usage Monitoring

```yaml
# Prometheus alert rules
- alert: PostgreSQLDiskUsageHigh
  expr: |
    (kubelet_volume_stats_used_bytes{persistentvolumeclaim="data-postgres-postgresql-0",namespace="postgres"} 
    / kubelet_volume_stats_capacity_bytes{persistentvolumeclaim="data-postgres-postgresql-0",namespace="postgres"}) 
    > 0.75
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL disk usage > 75%"

- alert: PostgreSQLDiskUsageCritical
  expr: |
    (kubelet_volume_stats_used_bytes{persistentvolumeclaim="data-postgres-postgresql-0",namespace="postgres"} 
    / kubelet_volume_stats_capacity_bytes{persistentvolumeclaim="data-postgres-postgresql-0",namespace="postgres"}) 
    > 0.85
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL disk usage > 85%"

- alert: PostgreSQLDiskFull
  expr: |
    (kubelet_volume_stats_used_bytes{persistentvolumeclaim="data-postgres-postgresql-0",namespace="postgres"} 
    / kubelet_volume_stats_capacity_bytes{persistentvolumeclaim="data-postgres-postgresql-0",namespace="postgres"}) 
    > 0.95
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL disk nearly full"
```

### 2. Configure Auto-Vacuum

```yaml
# In HelmRelease configuration
primary:
  configuration: |
    # Auto-vacuum settings
    autovacuum = on
    autovacuum_max_workers = 3
    autovacuum_naptime = 60s
    autovacuum_vacuum_threshold = 50
    autovacuum_analyze_threshold = 50
    autovacuum_vacuum_scale_factor = 0.1
    autovacuum_analyze_scale_factor = 0.05
    autovacuum_vacuum_cost_delay = 20ms
    autovacuum_vacuum_cost_limit = 200
```

### 3. Implement Data Retention Policy

```sql
-- Create retention policy function
CREATE OR REPLACE FUNCTION cleanup_old_data() RETURNS void AS $$
BEGIN
  -- Delete logs older than 90 days
  DELETE FROM logs WHERE created_at < now() - interval '90 days';
  
  -- Delete events older than 30 days
  DELETE FROM events WHERE created_at < now() - interval '30 days';
  
  -- Delete temporary data older than 7 days
  DELETE FROM temp_data WHERE created_at < now() - interval '7 days';
  
  -- Vacuum tables after deletion
  VACUUM ANALYZE logs;
  VACUUM ANALYZE events;
  VACUUM ANALYZE temp_data;
END;
$$ LANGUAGE plpgsql;
```

```yaml
# Create CronJob for data cleanup
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-cleanup
  namespace: postgres
spec:
  schedule: "0 3 * * 0"  # Weekly on Sunday at 3 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              psql -h postgres-postgresql -U postgres -d bruno_site <<EOF
              SELECT cleanup_old_data();
              EOF
          restartPolicy: OnFailure
```

### 4. Configure Log Rotation

```yaml
# In HelmRelease configuration
primary:
  configuration: |
    # Logging
    logging_collector = on
    log_directory = 'log'
    log_filename = 'postgresql-%a.log'  # One file per day of week (rotates weekly)
    log_truncate_on_rotation = on
    log_rotation_age = 1d
    log_rotation_size = 100MB
```

### 5. Set Up Automated Backups

```yaml
# Backup CronJob (backs up and optionally archives to S3)
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: postgres
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              BACKUP_FILE="/backup/bruno_site-$(date +%Y%m%d-%H%M%S).sql.gz"
              pg_dump -h postgres-postgresql -U postgres bruno_site | gzip > $BACKUP_FILE
              
              # Keep only last 7 days of backups
              find /backup -name "bruno_site-*.sql.gz" -mtime +7 -delete
              
              echo "Backup completed: $BACKUP_FILE"
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: postgres-backup-pvc
          restartPolicy: OnFailure
```

### 6. Use Appropriate PVC Size

```yaml
# In HelmRelease values - size appropriately from start
primary:
  persistence:
    enabled: true
    size: 40Gi  # Start with larger size if expecting growth
    storageClass: local-path
```

### 7. Table Partitioning for Large Tables

```sql
-- For logs or time-series data
CREATE TABLE logs (
  id BIGSERIAL,
  created_at TIMESTAMP NOT NULL,
  level TEXT,
  message TEXT
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE logs_2025_01 PARTITION OF logs
  FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE logs_2025_02 PARTITION OF logs
  FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Automate partition creation with function
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
  start_date DATE;
  end_date DATE;
  table_name TEXT;
BEGIN
  start_date := date_trunc('month', CURRENT_DATE + interval '1 month');
  end_date := start_date + interval '1 month';
  table_name := 'logs_' || to_char(start_date, 'YYYY_MM');
  
  EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF logs FOR VALUES FROM (%L) TO (%L)',
                 table_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;
```

## Storage Capacity Planning

```
Current usage: 18GB / 20GB (90%)
Daily growth rate: ~500MB
Days until full: (20GB - 18GB) / 500MB = 4 days

Recommendations:
1. Immediate: Clean up + expand to 40GB
2. Short term: Implement retention policy
3. Long term: Monitor growth, plan for 60-80GB
```

## Quick Reference Commands

```bash
# Check disk usage
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data

# Check database sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT datname, pg_size_pretty(pg_database_size(datname)) FROM pg_database;"

# Check largest tables
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::regclass)) FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(tablename::regclass) DESC LIMIT 10;"

# Run VACUUM
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c "VACUUM VERBOSE;"

# Expand PVC
kubectl edit pvc -n postgres data-postgres-postgresql-0
```

## Related Alerts

- `PostgreSQLDown`
- `PostgreSQLHighMemory`
- `PostgreSQLSlowQueries`

## Escalation

If storage issues persist:

1. ✅ Review data retention requirements
2. 📊 Analyze data growth patterns
3. 🔍 Identify largest growing tables
4. 💾 Implement archival strategy
5. 🔄 Consider object storage for old data (S3, MinIO)
6. 📞 Contact database team for capacity planning
7. 🆘 Evaluate managed PostgreSQL with auto-scaling storage

## Additional Resources

- [PostgreSQL VACUUM](https://www.postgresql.org/docs/current/sql-vacuum.html)
- [PostgreSQL WAL Management](https://www.postgresql.org/docs/current/wal-configuration.html)
- [Table Partitioning](https://www.postgresql.org/docs/current/ddl-partitioning.html)
- [Kubernetes PVC Expansion](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

