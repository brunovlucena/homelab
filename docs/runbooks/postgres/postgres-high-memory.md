# 🚨 Runbook: PostgreSQL High Memory Usage

## Alert Information

**Alert Name:** `PostgreSQLHighMemory`  
**Severity:** Warning  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** Memory usage > 80% of limit or OOMKill events

## Symptom

PostgreSQL experiencing memory pressure, approaching pod memory limits, or being killed by OOMKiller.

## Impact

- **User Impact:** MEDIUM - Slow queries, connection timeouts, service restarts
- **Business Impact:** MEDIUM - Degraded performance, potential data loss on OOMKill
- **Data Impact:** HIGH - Risk of transaction loss during OOM events

## Diagnosis

### 1. Check Current Memory Usage

```bash
# Check pod memory usage
kubectl top pod -n postgres postgres-postgresql-0

# Check memory limits and requests
kubectl get pod -n postgres postgres-postgresql-0 -o jsonpath='{.spec.containers[0].resources}'

# Check OOMKill events
kubectl get pod -n postgres postgres-postgresql-0 -o jsonpath='{.status.containerStatuses[0].lastState}'
kubectl describe pod -n postgres postgres-postgresql-0 | grep -i "oom"
```

### 2. Check PostgreSQL Memory Configuration

```bash
# Check shared_buffers (main cache)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW shared_buffers;"

# Check effective_cache_size (query planner hint)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW effective_cache_size;"

# Check work_mem (per-query memory)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW work_mem;"

# Check maintenance_work_mem (maintenance operations)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW maintenance_work_mem;"

# Get all memory-related settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT name, setting, unit, context FROM pg_settings WHERE name IN 
  ('shared_buffers', 'effective_cache_size', 'work_mem', 'maintenance_work_mem', 'max_connections');"
```

### 3. Check Active Connections

```bash
# Count active connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) as connections, max_conn FROM pg_stat_activity, 
   (SELECT setting::int as max_conn FROM pg_settings WHERE name='max_connections') x 
   GROUP BY max_conn;"

# Check connection details by state
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"

# Check connections by database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, count(*) FROM pg_stat_activity GROUP BY datname ORDER BY count(*) DESC;"
```

### 4. Identify Memory-Intensive Queries

```bash
# Check active queries with state
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, state, 
   now() - query_start AS duration, 
   substring(query, 1, 80) 
   FROM pg_stat_activity 
   WHERE state != 'idle' 
   ORDER BY duration DESC 
   LIMIT 10;"

# Check for long-running transactions
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, 
   now() - xact_start AS duration, 
   state, query 
   FROM pg_stat_activity 
   WHERE xact_start IS NOT NULL 
   ORDER BY duration DESC 
   LIMIT 10;"
```

### 5. Check Database Statistics

```bash
# Check database sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, pg_size_pretty(pg_database_size(datname)) AS size 
   FROM pg_database 
   ORDER BY pg_database_size(datname) DESC;"

# Check table bloat (can cause high memory usage)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, 
   pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
   pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size 
   FROM pg_tables 
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema') 
   ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC 
   LIMIT 10;"
```

### 6. Check Cache Hit Ratio

```bash
# Check buffer cache hit ratio (should be > 99%)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT sum(heap_blks_read) as heap_read, 
   sum(heap_blks_hit) as heap_hit, 
   sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) * 100 as cache_hit_ratio 
   FROM pg_statio_user_tables;"
```

## Resolution Steps

### Step 1: Immediate Actions

#### Terminate Memory-Intensive Queries

```bash
# Identify and terminate problem queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity 
   WHERE state = 'active' 
   AND now() - query_start > interval '5 minutes' 
   AND query NOT LIKE '%pg_stat_activity%';"

# Or terminate specific PID
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(<PID>);"
```

#### Close Idle Connections

```bash
# Find idle connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, state, 
   now() - state_change AS idle_time 
   FROM pg_stat_activity 
   WHERE state = 'idle' 
   ORDER BY idle_time DESC 
   LIMIT 20;"

# Terminate old idle connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity 
   WHERE state = 'idle' 
   AND now() - state_change > interval '1 hour';"
```

### Step 2: Optimize Memory Configuration

#### Issue: shared_buffers Too High
**Cause:** PostgreSQL allocated too much shared memory  
**Fix:**
```bash
# Current setting
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW shared_buffers;"

# Recommended: 25% of system memory
# For 2Gi limit: shared_buffers = 512MB
# For 4Gi limit: shared_buffers = 1GB

# Update in HelmRelease
kubectl edit helmrelease -n postgres postgres

# Add configuration:
# primary:
#   configuration: |
#     shared_buffers = 512MB
#     effective_cache_size = 1536MB  # 75% of memory
#     work_mem = 16MB
#     maintenance_work_mem = 128MB

# Apply changes
flux reconcile helmrelease postgres -n postgres
```

#### Issue: work_mem Too High
**Cause:** Per-query memory too large (multiplied by connections)  
**Fix:**
```bash
# Check current work_mem
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW work_mem;"

# Calculate safe work_mem:
# work_mem = (Total RAM - shared_buffers) / (max_connections * 3)
# Example for 2GB with 200 connections: (2GB - 512MB) / (200 * 3) ≈ 2.5MB
# Conservative setting: 4-16MB

# Update configuration
kubectl edit helmrelease -n postgres postgres
# Set: work_mem = 8MB
```

#### Issue: Too Many Connections
**Cause:** Each connection consumes memory  
**Fix:**
```bash
# Check max_connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW max_connections;"

# Reduce max_connections if too high
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_connections = 100;"

# Or update in HelmRelease for permanent change
kubectl edit helmrelease -n postgres postgres
# Set: max_connections = 100

# Restart required
kubectl delete pod -n postgres postgres-postgresql-0
```

### Step 3: Increase Pod Resources

```bash
# Edit HelmRelease to increase memory limits
kubectl edit helmrelease -n postgres postgres

# Update resources:
# primary:
#   resources:
#     limits:
#       memory: "4Gi"  # Increased from 2Gi
#       cpu: "2000m"
#     requests:
#       memory: "1Gi"  # Increased from 512Mi
#       cpu: "500m"

# Apply changes
flux reconcile helmrelease postgres -n postgres

# Monitor restart
kubectl get pod -n postgres postgres-postgresql-0 -w
```

### Step 4: Implement Connection Pooling

#### Using PgBouncer

```yaml
# Deploy PgBouncer for connection pooling
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
  namespace: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pgbouncer
  template:
    metadata:
      labels:
        app: pgbouncer
    spec:
      containers:
      - name: pgbouncer
        image: edoburu/pgbouncer:1.21.0
        ports:
        - containerPort: 5432
        env:
        - name: DATABASE_URL
          value: "postgres://postgres@postgres-postgresql.postgres.svc.cluster.local:5432/bruno_site"
        - name: POOL_MODE
          value: "transaction"
        - name: MAX_CLIENT_CONN
          value: "1000"
        - name: DEFAULT_POOL_SIZE
          value: "25"
---
apiVersion: v1
kind: Service
metadata:
  name: pgbouncer
  namespace: postgres
spec:
  ports:
  - port: 5432
    targetPort: 5432
  selector:
    app: pgbouncer
```

#### Application Configuration

```bash
# Update applications to use PgBouncer
# Instead of: postgres-postgresql.postgres.svc.cluster.local
# Use: pgbouncer.postgres.svc.cluster.local
```

### Step 5: Database Optimization

#### Run VACUUM ANALYZE

```bash
# Vacuum and analyze to reclaim memory from dead tuples
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM ANALYZE;"

# Check last vacuum times
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, 
   last_vacuum, last_autovacuum, 
   last_analyze, last_autoanalyze 
   FROM pg_stat_user_tables 
   ORDER BY last_vacuum DESC NULLS LAST;"
```

#### Optimize Queries

```bash
# Find tables missing indexes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, attname 
   FROM pg_stats 
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema') 
   AND n_distinct < 0 
   ORDER BY n_distinct 
   LIMIT 10;"

# Check slow queries (if slow log enabled)
kubectl logs -n postgres postgres-postgresql-0 --tail=500 | grep "duration:" | sort -rn | head -20

# Enable slow query logging
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET log_min_duration_statement = 1000;"  # Log queries > 1s
kubectl delete pod -n postgres postgres-postgresql-0
```

### Step 6: Application-Level Fixes

#### Configure Application Connection Pool

```yaml
# Example: Python application with psycopg2
POSTGRES_POOL_MIN: "5"
POSTGRES_POOL_MAX: "20"
POSTGRES_POOL_TIMEOUT: "30"
POSTGRES_STATEMENT_TIMEOUT: "30000"  # 30s

# Example: Node.js with pg
PG_POOL_MIN: "2"
PG_POOL_MAX: "10"
PG_IDLE_TIMEOUT: "30000"
PG_CONNECTION_TIMEOUT: "5000"
```

#### Close Connections Properly

```python
# Python example
import psycopg2
from psycopg2 import pool

# Use connection pooling
connection_pool = psycopg2.pool.SimpleConnectionPool(
    5,  # min connections
    20,  # max connections
    host='postgres-postgresql.postgres.svc.cluster.local',
    database='bruno_site',
    user='postgres'
)

# Always close connections
conn = connection_pool.getconn()
try:
    # Execute queries
    pass
finally:
    connection_pool.putconn(conn)
```

## Verification

### 1. Check Memory Usage Decreased

```bash
# Check current usage
kubectl top pod -n postgres postgres-postgresql-0

# Should be < 80% of limit
```

### 2. Verify Configuration Applied

```bash
# Check all memory settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT name, setting, unit 
   FROM pg_settings 
   WHERE name IN ('shared_buffers', 'work_mem', 'maintenance_work_mem', 
                  'effective_cache_size', 'max_connections');"
```

### 3. Check Connection Count

```bash
# Should be reduced
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity;"
```

### 4. Monitor Cache Hit Ratio

```bash
# Should be > 99%
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) * 100 as cache_hit_ratio 
   FROM pg_statio_user_tables;"
```

### 5. Verify No OOMKills

```bash
# Check recent restarts
kubectl get pod -n postgres postgres-postgresql-0 -o jsonpath='{.status.containerStatuses[0].restartCount}'

# Check for OOMKill events
kubectl get events -n postgres --field-selector involvedObject.name=postgres-postgresql-0 | grep OOMKilled
```

## Prevention

### 1. Proper Memory Configuration

```yaml
# In HelmRelease values.yaml
primary:
  resources:
    limits:
      memory: "4Gi"
      cpu: "2000m"
    requests:
      memory: "1Gi"
      cpu: "500m"
  
  configuration: |
    # Memory settings (for 4Gi pod)
    shared_buffers = 1GB                # 25% of memory
    effective_cache_size = 3GB          # 75% of memory
    work_mem = 8MB                      # Conservative
    maintenance_work_mem = 256MB        # For VACUUM, CREATE INDEX
    
    # Connection settings
    max_connections = 100
    
    # Query tuning
    random_page_cost = 1.1              # For SSD
    effective_io_concurrency = 200      # For SSD
```

### 2. Set Up Monitoring

```yaml
# Prometheus alert rules
- alert: PostgreSQLHighMemory
  expr: |
    container_memory_usage_bytes{pod=~"postgres-postgresql-.*", namespace="postgres"} 
    / container_spec_memory_limit_bytes{pod=~"postgres-postgresql-.*", namespace="postgres"} 
    > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL high memory usage"

- alert: PostgreSQLOOMKilled
  expr: |
    kube_pod_container_status_last_terminated_reason{namespace="postgres", 
    pod=~"postgres-postgresql-.*", reason="OOMKilled"} > 0
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL was OOMKilled"
```

### 3. Connection Timeout Configuration

```yaml
# In application environment variables
DB_STATEMENT_TIMEOUT: "30000"      # 30s
DB_IDLE_IN_TRANSACTION_TIMEOUT: "60000"  # 1 minute
DB_CONNECTION_TIMEOUT: "5000"      # 5s
```

### 4. Regular Maintenance

```bash
# Schedule regular VACUUM
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-vacuum
  namespace: postgres
spec:
  schedule: "0 3 * * 0"  # Weekly on Sunday at 3 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: vacuum
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              psql -h postgres-postgresql -U postgres -d bruno_site -c "VACUUM ANALYZE VERBOSE;"
          restartPolicy: OnFailure
```

### 5. Implement Auto-Vacuum Tuning

```yaml
# In PostgreSQL configuration
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 60s
autovacuum_vacuum_threshold = 50
autovacuum_analyze_threshold = 50
autovacuum_vacuum_scale_factor = 0.1
autovacuum_analyze_scale_factor = 0.05
```

## Performance Tips

1. **Right-Size shared_buffers:** 25% of system memory is a good starting point
2. **Monitor work_mem:** Too high = OOM, too low = disk sorts
3. **Use Connection Pooling:** PgBouncer reduces memory per connection
4. **Close Idle Connections:** Set idle_in_transaction_session_timeout
5. **Regular VACUUM:** Prevents bloat and memory waste
6. **Index Optimization:** Proper indexes reduce memory-intensive sorts

## Related Alerts

- `PostgreSQLDown`
- `PostgreSQLSlowQueries`
- `PostgreSQLHighConnections`
- `PostgreSQLOOMKilled`

## Escalation

If memory issues persist after applying fixes:

1. ✅ Review application query patterns for optimization
2. 📊 Analyze long-term memory trends
3. 🔍 Check for connection leaks in application code
4. 💾 Consider vertical scaling (more memory)
5. 🔄 Evaluate read replicas for horizontal scaling
6. 📞 Contact database team for advanced tuning
7. 🆘 Consider managed PostgreSQL service

## Additional Resources

- [PostgreSQL Memory Configuration](https://www.postgresql.org/docs/current/runtime-config-resource.html)
- [PgBouncer Documentation](https://www.pgbouncer.org/)
- [PostgreSQL Tuning Guide](https://wiki.postgresql.org/wiki/Tuning_Your_PostgreSQL_Server)
- [Bitnami PostgreSQL Configuration](https://github.com/bitnami/charts/tree/main/bitnami/postgresql)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

