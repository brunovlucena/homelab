# 🚨 Runbook: PostgreSQL Slow Queries

## Alert Information

**Alert Name:** `PostgreSQLSlowQueries`  
**Severity:** Warning  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** Query duration > 5 seconds for active queries

## Symptom

Database queries taking longer than expected, causing application timeouts and poor user experience.

## Impact

- **User Impact:** MEDIUM - Slow page loads, timeouts, degraded experience
- **Business Impact:** MEDIUM - Reduced throughput, potential revenue impact
- **Data Impact:** LOW - Read/write operations delayed but not lost

## Diagnosis

### 1. Identify Currently Slow Queries

```bash
# Check active slow queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, 
   now() - query_start AS duration, 
   state, 
   substring(query, 1, 100) as query 
   FROM pg_stat_activity 
   WHERE state = 'active' 
   AND now() - query_start > interval '1 second' 
   ORDER BY duration DESC;"

# Find blocking queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT blocked_locks.pid AS blocked_pid,
   blocked_activity.usename AS blocked_user,
   blocking_locks.pid AS blocking_pid,
   blocking_activity.usename AS blocking_user,
   blocked_activity.query AS blocked_statement,
   blocking_activity.query AS blocking_statement
   FROM pg_catalog.pg_locks blocked_locks
   JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
   JOIN pg_catalog.pg_locks blocking_locks 
     ON blocking_locks.locktype = blocked_locks.locktype
     AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
     AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
     AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
     AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
     AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
     AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
     AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
     AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
     AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
     AND blocking_locks.pid != blocked_locks.pid
   JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
   WHERE NOT blocked_locks.granted;"
```

### 2. Check Query Statistics

```bash
# Enable pg_stat_statements if not already enabled
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"

# Get slowest queries by average time
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT 
   substring(query, 1, 80) AS query,
   calls,
   round(total_exec_time::numeric, 2) AS total_time,
   round(mean_exec_time::numeric, 2) AS mean_time,
   round((100 * total_exec_time / sum(total_exec_time) OVER ())::numeric, 2) AS percentage
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;"

# Get most time-consuming queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT 
   substring(query, 1, 80) AS query,
   calls,
   round(total_exec_time::numeric, 2) AS total_time,
   round(mean_exec_time::numeric, 2) AS mean_time
   FROM pg_stat_statements
   ORDER BY total_exec_time DESC
   LIMIT 10;"
```

### 3. Check Table and Index Statistics

```bash
# Check sequential scans (should use indexes instead)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, seq_scan, seq_tup_read,
   idx_scan, idx_tup_fetch,
   seq_scan - idx_scan AS too_much_seq
   FROM pg_stat_user_tables
   WHERE seq_scan - idx_scan > 0
   ORDER BY too_much_seq DESC
   LIMIT 10;"

# Check unused indexes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, indexname, idx_scan
   FROM pg_stat_user_indexes
   WHERE idx_scan = 0
   ORDER BY pg_relation_size(indexrelid) DESC;"

# Check missing indexes on foreign keys
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT c.conrelid::regclass AS table,
   string_agg(a.attname, ', ') AS columns
   FROM pg_constraint c
   JOIN pg_attribute a ON a.attnum = ANY(c.conkey) AND a.attrelid = c.conrelid
   WHERE c.contype = 'f'
   AND NOT EXISTS (
     SELECT 1 FROM pg_index i
     WHERE i.indrelid = c.conrelid
     AND array_to_string(c.conkey, ' ') = array_to_string(i.indkey[0:array_length(c.conkey,1)-1], ' ')
   )
   GROUP BY c.conrelid
   ORDER BY c.conrelid::regclass::text;"
```

### 4. Check Table Bloat

```bash
# Check table sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename,
   pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
   pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
   pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
   FROM pg_tables
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
   ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
   LIMIT 10;"

# Check last vacuum times
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename,
   last_vacuum, last_autovacuum,
   n_tup_ins, n_tup_upd, n_tup_del
   FROM pg_stat_user_tables
   ORDER BY n_tup_upd + n_tup_del DESC
   LIMIT 10;"
```

### 5. Check System Resources

```bash
# Check CPU and memory
kubectl top pod -n postgres postgres-postgresql-0

# Check I/O wait
kubectl exec -n postgres postgres-postgresql-0 -- top -bn1 | grep "Cpu(s)"

# Check disk I/O
kubectl exec -n postgres postgres-postgresql-0 -- iostat -x 1 5
```

### 6. Check Locks and Contention

```bash
# Check lock modes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT locktype, relation::regclass, mode, granted, pid
   FROM pg_locks
   WHERE NOT granted
   ORDER BY relation;"

# Check lock wait events
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, wait_event_type, wait_event, state, query
   FROM pg_stat_activity
   WHERE wait_event IS NOT NULL
   AND state = 'active';"
```

## Resolution Steps

### Step 1: Immediate Actions

#### Terminate Problematic Queries

```bash
# Identify and terminate long-running queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE state = 'active'
   AND now() - query_start > interval '10 minutes'
   AND query NOT LIKE '%pg_stat_activity%';"

# Or terminate specific PID
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(<PID>);"
```

### Step 2: Create Missing Indexes

#### Issue: Sequential Scans on Large Tables
**Cause:** Missing indexes causing full table scans  
**Fix:**
```bash
# Identify columns needing indexes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, attname, n_distinct, correlation
   FROM pg_stats
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
   AND n_distinct > 100
   ORDER BY n_distinct DESC
   LIMIT 20;"

# Create indexes on frequently queried columns
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_users_email ON users(email);"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_projects_user_id ON projects(user_id);"

# Create composite indexes for multi-column queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_posts_author_created ON posts(author_id, created_at DESC);"
```

#### Issue: Missing Indexes on Foreign Keys
**Cause:** Foreign key columns without indexes  
**Fix:**
```bash
# Create indexes on all foreign key columns
# Example:
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_posts_user_id ON posts(user_id);"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_comments_post_id ON comments(post_id);"
```

### Step 3: Optimize Queries

#### Analyze Query Plans

```bash
# Get explain plan for slow query
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "EXPLAIN ANALYZE
   SELECT * FROM users WHERE email = 'test@example.com';"

# Look for:
# - Sequential Scans on large tables (add indexes)
# - High cost operations
# - Nested loops with high row counts
# - Sort operations (consider indexes for ORDER BY)
```

#### Issue: Inefficient JOIN Operations
**Cause:** Poor query structure or missing indexes  
**Fix:**
```sql
-- Before (inefficient)
SELECT * FROM posts p
JOIN users u ON p.user_id = u.id
WHERE u.email = 'test@example.com';

-- After (efficient with index on users.email)
SELECT p.* FROM posts p
JOIN users u ON p.user_id = u.id
WHERE u.email = 'test@example.com'
AND p.created_at > now() - interval '30 days';  -- Add time filter
```

#### Issue: SELECT * Queries
**Cause:** Fetching unnecessary columns  
**Fix:**
```sql
-- Before
SELECT * FROM large_table WHERE id = 1;

-- After (only needed columns)
SELECT id, name, email FROM large_table WHERE id = 1;
```

### Step 4: Vacuum and Analyze

```bash
# Run VACUUM ANALYZE on affected tables
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM ANALYZE users;"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM ANALYZE posts;"

# Full VACUUM if bloat is severe (locks table)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "VACUUM FULL ANALYZE large_bloated_table;"

# Update planner statistics
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "ANALYZE;"
```

### Step 5: Tune PostgreSQL Configuration

#### Update Query Planner Settings

```bash
# Edit HelmRelease
kubectl edit helmrelease -n postgres postgres

# Add configuration tuning:
# primary:
#   configuration: |
#     # Query planning
#     random_page_cost = 1.1              # For SSD (default: 4.0)
#     effective_io_concurrency = 200      # For SSD (default: 1)
#     
#     # Memory for complex queries
#     work_mem = 16MB                     # Per operation (default: 4MB)
#     
#     # Parallel query settings
#     max_parallel_workers_per_gather = 2
#     max_parallel_workers = 4
#     
#     # Query timeout
#     statement_timeout = 30000           # 30s timeout
#     
#     # Logging slow queries
#     log_min_duration_statement = 1000   # Log queries > 1s

# Apply changes
flux reconcile helmrelease postgres -n postgres
```

### Step 6: Implement Query Result Caching

#### Application-Level Caching

```python
# Example with Redis caching
import redis
import psycopg2
import json
import hashlib

redis_client = redis.Redis(host='redis-master.redis.svc.cluster.local')

def cached_query(query, params, ttl=300):
    # Create cache key
    cache_key = f"query:{hashlib.md5(query.encode() + str(params).encode()).hexdigest()}"
    
    # Check cache
    cached = redis_client.get(cache_key)
    if cached:
        return json.loads(cached)
    
    # Execute query
    conn = psycopg2.connect("postgresql://postgres@postgres-postgresql.postgres.svc:5432/bruno_site")
    cur = conn.cursor()
    cur.execute(query, params)
    result = cur.fetchall()
    cur.close()
    conn.close()
    
    # Cache result
    redis_client.setex(cache_key, ttl, json.dumps(result))
    return result
```

### Step 7: Partition Large Tables

```bash
# For very large tables, consider partitioning
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "-- Create partitioned table
   CREATE TABLE logs_new (
     id SERIAL,
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
   
   -- Rename tables
   ALTER TABLE logs RENAME TO logs_old;
   ALTER TABLE logs_new RENAME TO logs;"
```

## Verification

### 1. Check Query Performance Improved

```bash
# Check current slow queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, now() - query_start AS duration, query
   FROM pg_stat_activity
   WHERE state = 'active'
   ORDER BY duration DESC
   LIMIT 5;"

# Should show improved durations
```

### 2. Verify Indexes Are Being Used

```bash
# Check index usage
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
   FROM pg_stat_user_indexes
   WHERE idx_scan > 0
   ORDER BY idx_scan DESC
   LIMIT 10;"
```

### 3. Check Query Statistics

```bash
# Check pg_stat_statements for improvements
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT substring(query, 1, 80) AS query,
   calls,
   round(mean_exec_time::numeric, 2) AS mean_time
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;"
```

### 4. Monitor Application Performance

```bash
# Check application response times
kubectl logs -n homepage deployment/homepage-api --tail=100 | grep "response_time"

# Test from application
kubectl exec -n homepage deployment/homepage-api -- curl -w "@curl-format.txt" http://localhost:8080/api/projects
```

### 5. Check Cache Hit Ratio

```bash
# Should be > 99%
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT 
   sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) * 100 as cache_hit_ratio
   FROM pg_statio_user_tables;"
```

## Prevention

### 1. Enable Query Logging

```yaml
# In HelmRelease configuration
primary:
  configuration: |
    # Logging
    log_min_duration_statement = 1000    # Log queries > 1s
    log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
    log_checkpoints = on
    log_connections = on
    log_disconnections = on
    log_lock_waits = on
```

### 2. Set Up pg_stat_statements

```yaml
# In HelmRelease configuration
primary:
  configuration: |
    shared_preload_libraries = 'pg_stat_statements'
    pg_stat_statements.track = all
    pg_stat_statements.max = 10000
```

### 3. Regular Index Maintenance

```bash
# Create CronJob for index analysis
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-index-check
  namespace: postgres
spec:
  schedule: "0 4 * * 0"  # Weekly on Sunday at 4 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: index-check
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              # Check for missing indexes
              psql -h postgres-postgresql -U postgres -d bruno_site <<EOF
              SELECT schemaname, tablename, seq_scan, seq_tup_read
              FROM pg_stat_user_tables
              WHERE seq_scan > idx_scan
              AND seq_tup_read > 10000
              ORDER BY seq_tup_read DESC;
              EOF
          restartPolicy: OnFailure
```

### 4. Implement Query Timeouts

```yaml
# Application configuration
DATABASE_STATEMENT_TIMEOUT: "30000"     # 30s per query
DATABASE_LOCK_TIMEOUT: "10000"          # 10s for locks
DATABASE_IDLE_IN_TRANSACTION_TIMEOUT: "60000"  # 1 min
```

### 5. Set Up Monitoring

```yaml
# Prometheus alert rules
- alert: PostgreSQLSlowQueries
  expr: |
    rate(pg_stat_statements_mean_exec_time_seconds[5m]) > 5
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL has slow queries"

- alert: PostgreSQLHighSeqScans
  expr: |
    rate(pg_stat_user_tables_seq_scan[5m]) > 100
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL performing many sequential scans"
```

### 6. Database Design Best Practices

- ✅ Index all foreign keys
- ✅ Index frequently queried columns
- ✅ Use appropriate data types (INT vs VARCHAR)
- ✅ Normalize appropriately (avoid over-normalization)
- ✅ Use composite indexes for multi-column queries
- ✅ Add indexes for ORDER BY columns
- ✅ Consider partial indexes for filtered queries

### 7. Query Best Practices

- ✅ Avoid SELECT * - fetch only needed columns
- ✅ Use LIMIT for pagination
- ✅ Add WHERE clauses to filter data early
- ✅ Use JOINs efficiently
- ✅ Avoid N+1 queries (use JOINs or batching)
- ✅ Use connection pooling
- ✅ Implement application-level caching

## Query Optimization Checklist

```bash
# For each slow query:

# 1. Get EXPLAIN ANALYZE
EXPLAIN ANALYZE <your_query>;

# 2. Check for:
- [ ] Sequential Scans → Add indexes
- [ ] High cost Nested Loops → Check join conditions
- [ ] Sort operations → Add indexes for ORDER BY
- [ ] Hash operations → Consider increasing work_mem
- [ ] Bitmap Heap Scans with high rows → Check index selectivity

# 3. Verify indexes exist
SELECT * FROM pg_indexes WHERE tablename = '<table_name>';

# 4. Check index usage
SELECT * FROM pg_stat_user_indexes WHERE tablename = '<table_name>';

# 5. Analyze table statistics
ANALYZE <table_name>;

# 6. Check table bloat
SELECT pg_size_pretty(pg_total_relation_size('<table_name>'));
```

## Related Alerts

- `PostgreSQLDown`
- `PostgreSQLHighMemory`
- `PostgreSQLHighConnections`
- `PostgreSQLDeadlocks`

## Escalation

If slow queries persist after optimization:

1. ✅ Review application code for inefficient queries
2. 📊 Analyze query patterns over time
3. 🔍 Consider read replicas for read-heavy workloads
4. 💾 Evaluate horizontal partitioning
5. 🔄 Consider connection pooling with PgBouncer
6. 📞 Contact development team for query refactoring
7. 🆘 Consider upgrading to larger instance size

## Additional Resources

- [PostgreSQL Query Optimization](https://www.postgresql.org/docs/current/performance-tips.html)
- [Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html)
- [pg_stat_statements](https://www.postgresql.org/docs/current/pgstatstatements.html)
- [Index Types](https://www.postgresql.org/docs/current/indexes-types.html)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

