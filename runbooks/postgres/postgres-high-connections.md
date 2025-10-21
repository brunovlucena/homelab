# 🚨 Runbook: PostgreSQL High Connections

## Alert Information

**Alert Name:** `PostgreSQLHighConnections`  
**Severity:** Warning  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** Active connections > 80% of max_connections

## Symptom

PostgreSQL connection count approaching or exceeding maximum allowed connections, causing new connection attempts to fail.

## Impact

- **User Impact:** HIGH - Connection refused errors, application failures
- **Business Impact:** HIGH - Service unavailability for new requests
- **Data Impact:** LOW - Existing connections work, but new ones fail

## Diagnosis

### 1. Check Current Connection Count

```bash
# Get connection count vs max
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) as current_connections, 
   (SELECT setting::int FROM pg_settings WHERE name='max_connections') as max_connections,
   round(100.0 * count(*) / (SELECT setting::int FROM pg_settings WHERE name='max_connections'), 2) as percentage
   FROM pg_stat_activity;"

# Connection breakdown by state
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT state, count(*) 
   FROM pg_stat_activity 
   GROUP BY state 
   ORDER BY count(*) DESC;"
```

### 2. Identify Connection Sources

```bash
# Connections by database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, count(*) as connections
   FROM pg_stat_activity
   GROUP BY datname
   ORDER BY count(*) DESC;"

# Connections by user
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT usename, count(*) as connections
   FROM pg_stat_activity
   GROUP BY usename
   ORDER BY count(*) DESC;"

# Connections by application
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT application_name, count(*) as connections
   FROM pg_stat_activity
   WHERE application_name != ''
   GROUP BY application_name
   ORDER BY count(*) DESC;"

# Connections by client IP
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT client_addr, count(*) as connections
   FROM pg_stat_activity
   WHERE client_addr IS NOT NULL
   GROUP BY client_addr
   ORDER BY count(*) DESC;"
```

### 3. Check Idle Connections

```bash
# Count idle connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT state, count(*) 
   FROM pg_stat_activity 
   WHERE state IN ('idle', 'idle in transaction', 'idle in transaction (aborted)')
   GROUP BY state;"

# Idle connections with duration
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, application_name, state,
   now() - state_change AS idle_duration,
   now() - query_start AS query_duration
   FROM pg_stat_activity
   WHERE state = 'idle'
   ORDER BY state_change
   LIMIT 20;"

# Long-running idle transactions (dangerous!)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, state,
   now() - xact_start AS transaction_duration,
   now() - state_change AS idle_duration
   FROM pg_stat_activity
   WHERE state = 'idle in transaction'
   AND now() - state_change > interval '5 minutes'
   ORDER BY state_change;"
```

### 4. Check Reserved Connections

```bash
# Check superuser reserved connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT name, setting, context 
   FROM pg_settings 
   WHERE name IN ('max_connections', 'superuser_reserved_connections', 'reserved_connections');"
```

### 5. Check Application Connection Pools

```bash
# Check homepage API connections
kubectl exec -n homepage deployment/homepage-api -- env | grep -i postgres

# Check other applications
kubectl get pods -A -l app.kubernetes.io/name=postgres-client -o wide
```

### 6. Monitor Connection Rate

```bash
# Check connection creation rate from logs
kubectl logs -n postgres postgres-postgresql-0 --tail=100 | grep "connection"

# Watch connections in real-time
watch -n 2 'kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -t -c "SELECT count(*) FROM pg_stat_activity;"'
```

## Resolution Steps

### Step 1: Immediate Actions

#### Terminate Idle Connections

```bash
# Terminate long-idle connections (> 1 hour)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE state = 'idle'
   AND now() - state_change > interval '1 hour'
   AND pid != pg_backend_pid();"

# Terminate idle in transaction (dangerous - be careful!)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE state = 'idle in transaction'
   AND now() - state_change > interval '10 minutes'
   AND pid != pg_backend_pid();"
```

#### Kill Specific Application Connections

```bash
# If one application is hogging connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE application_name = 'problematic_app'
   AND state = 'idle';"

# Or by database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE datname = 'old_database'
   AND state = 'idle';"
```

### Step 2: Increase max_connections (Temporary)

#### Issue: Too Many Legitimate Connections
**Cause:** max_connections set too low for workload  
**Fix:**
```bash
# Check current max_connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SHOW max_connections;"

# Increase max_connections (requires restart)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_connections = 300;"

# Restart PostgreSQL to apply
kubectl delete pod -n postgres postgres-postgresql-0

# Wait for pod to be ready
kubectl wait --for=condition=ready pod -n postgres postgres-postgresql-0 --timeout=300s

# Verify new setting
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SHOW max_connections;"
```

**⚠️ Warning:** Increasing max_connections increases memory usage. Each connection uses ~10MB of memory.

### Step 3: Configure Connection Timeouts

```bash
# Set idle connection timeout
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET idle_in_transaction_session_timeout = '10min';"

# Set statement timeout
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET statement_timeout = '30s';"

# Set connection timeout
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET tcp_user_timeout = '30000';"  # 30 seconds

# Apply changes
kubectl delete pod -n postgres postgres-postgresql-0
```

### Step 4: Deploy PgBouncer for Connection Pooling

```yaml
# Create PgBouncer deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
  namespace: postgres
spec:
  replicas: 2
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
          name: postgres
        env:
        - name: DATABASE_URL
          value: "postgres://postgres@postgres-postgresql.postgres.svc.cluster.local:5432/bruno_site"
        - name: POOL_MODE
          value: "transaction"  # or "session" or "statement"
        - name: MAX_CLIENT_CONN
          value: "1000"
        - name: DEFAULT_POOL_SIZE
          value: "25"
        - name: MIN_POOL_SIZE
          value: "5"
        - name: RESERVE_POOL_SIZE
          value: "5"
        - name: MAX_DB_CONNECTIONS
          value: "50"
        - name: SERVER_IDLE_TIMEOUT
          value: "600"  # 10 minutes
        - name: SERVER_LIFETIME
          value: "3600"  # 1 hour
        resources:
          limits:
            memory: "256Mi"
            cpu: "500m"
          requests:
            memory: "128Mi"
            cpu: "100m"
---
apiVersion: v1
kind: Service
metadata:
  name: pgbouncer
  namespace: postgres
spec:
  type: ClusterIP
  ports:
  - port: 5432
    targetPort: 5432
    protocol: TCP
  selector:
    app: pgbouncer
```

```bash
# Apply PgBouncer
kubectl apply -f pgbouncer.yaml

# Verify PgBouncer is running
kubectl get pods -n postgres -l app=pgbouncer

# Test connection through PgBouncer
kubectl run postgres-test --image=postgres:16 --rm -it --restart=Never -- \
  psql postgresql://postgres@pgbouncer.postgres.svc.cluster.local:5432/bruno_site -c "SELECT 1;"
```

### Step 5: Update Application Configuration

#### Configure Application Connection Pools

```yaml
# Homepage API example
apiVersion: v1
kind: ConfigMap
metadata:
  name: homepage-api-config
  namespace: homepage
data:
  # Use PgBouncer instead of direct connection
  POSTGRES_HOST: "pgbouncer.postgres.svc.cluster.local"
  POSTGRES_PORT: "5432"
  POSTGRES_DB: "bruno_site"
  POSTGRES_USER: "postgres"
  
  # Connection pool settings
  POSTGRES_POOL_MIN: "2"
  POSTGRES_POOL_MAX: "10"
  POSTGRES_POOL_IDLE_TIMEOUT: "30000"  # 30s
  POSTGRES_CONNECTION_TIMEOUT: "5000"  # 5s
  POSTGRES_STATEMENT_TIMEOUT: "30000"  # 30s
```

```python
# Python application example
import psycopg2.pool

# Create connection pool
connection_pool = psycopg2.pool.ThreadedConnectionPool(
    minconn=2,
    maxconn=10,
    host='pgbouncer.postgres.svc.cluster.local',
    port=5432,
    database='bruno_site',
    user='postgres',
    connect_timeout=5
)

# Always return connections to pool
def execute_query(query):
    conn = connection_pool.getconn()
    try:
        cur = conn.cursor()
        cur.execute(query)
        result = cur.fetchall()
        cur.close()
        conn.commit()
        return result
    finally:
        connection_pool.putconn(conn)
```

```javascript
// Node.js application example
const { Pool } = require('pg');

const pool = new Pool({
  host: 'pgbouncer.postgres.svc.cluster.local',
  port: 5432,
  database: 'bruno_site',
  user: 'postgres',
  min: 2,
  max: 10,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 5000,
  statement_timeout: 30000
});

// Use pool for queries
async function executeQuery(query) {
  const client = await pool.connect();
  try {
    const result = await client.query(query);
    return result.rows;
  } finally {
    client.release();
  }
}
```

### Step 6: Set Connection Limits Per Database/User

```bash
# Set connection limit per database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER DATABASE bruno_site CONNECTION LIMIT 100;"

# Set connection limit per user
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER USER app_user CONNECTION LIMIT 20;"

# Check current limits
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, datconnlimit FROM pg_database WHERE datconnlimit != -1;"
```

### Step 7: Implement Connection Monitoring

```bash
# Create monitoring script
cat <<'EOF' > /tmp/connection-monitor.sh
#!/bin/bash
while true; do
  echo "=== $(date) ==="
  kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -t -c \
    "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"
  echo ""
  sleep 30
done
EOF

chmod +x /tmp/connection-monitor.sh
/tmp/connection-monitor.sh
```

## Verification

### 1. Check Connection Count Decreased

```bash
# Current connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) as current_connections,
   (SELECT setting::int FROM pg_settings WHERE name='max_connections') as max_connections
   FROM pg_stat_activity;"

# Should be < 80% of max_connections
```

### 2. Verify PgBouncer Working

```bash
# Check PgBouncer pods
kubectl get pods -n postgres -l app=pgbouncer

# Check PgBouncer stats (if accessible)
kubectl exec -n postgres deployment/pgbouncer -- psql -U postgres -p 5432 pgbouncer -c "SHOW POOLS;"
kubectl exec -n postgres deployment/pgbouncer -- psql -U postgres -p 5432 pgbouncer -c "SHOW DATABASES;"
```

### 3. Test Application Connectivity

```bash
# Test from homepage
kubectl exec -n homepage deployment/homepage-api -- nc -zv pgbouncer.postgres.svc.cluster.local 5432

# Check application logs for connection errors
kubectl logs -n homepage deployment/homepage-api --tail=50 | grep -i "connection\|pool"
```

### 4. Monitor Connection Stability

```bash
# Watch connections over time
watch -n 5 'kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -t -c "SELECT count(*) FROM pg_stat_activity;"'

# Should remain stable and below threshold
```

### 5. Check Idle Connection Timeout

```bash
# Verify timeouts are working
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT name, setting, unit FROM pg_settings 
   WHERE name IN ('idle_in_transaction_session_timeout', 'statement_timeout', 'tcp_user_timeout');"
```

## Prevention

### 1. Permanent PostgreSQL Configuration

```yaml
# In HelmRelease values
primary:
  configuration: |
    # Connection settings
    max_connections = 200
    superuser_reserved_connections = 3
    
    # Timeout settings
    idle_in_transaction_session_timeout = 600000  # 10 minutes
    statement_timeout = 30000                      # 30 seconds
    tcp_user_timeout = 30000                       # 30 seconds
    
    # Logging
    log_connections = on
    log_disconnections = on
```

### 2. Set Up Monitoring Alerts

```yaml
# Prometheus alert rules
- alert: PostgreSQLHighConnections
  expr: |
    (pg_stat_activity_count / pg_settings_max_connections) > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL connection count high"
    description: "{{ $value | humanizePercentage }} of max connections in use"

- alert: PostgreSQLMaxConnectionsReached
  expr: |
    pg_stat_activity_count >= pg_settings_max_connections
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL max connections reached"

- alert: PostgreSQLTooManyIdleConnections
  expr: |
    pg_stat_activity_count{state="idle"} > 50
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Too many idle PostgreSQL connections"
```

### 3. Application Best Practices

```python
# ✅ DO: Use connection pooling
pool = psycopg2.pool.SimpleConnectionPool(2, 10, **db_config)

# ✅ DO: Always close connections
try:
    conn = pool.getconn()
    # use connection
finally:
    pool.putconn(conn)

# ✅ DO: Use context managers
with pool.getconn() as conn:
    with conn.cursor() as cur:
        cur.execute(query)

# ❌ DON'T: Create new connection for each request
# conn = psycopg2.connect(**db_config)  # Bad!

# ❌ DON'T: Leave connections open
# conn = psycopg2.connect(**db_config)
# # ... never closed

# ✅ DO: Set reasonable pool limits
# min_pool_size = 2-5 per pod
# max_pool_size = 10-20 per pod
```

### 4. PgBouncer Pool Mode Selection

```yaml
# transaction mode (recommended for most apps)
POOL_MODE: "transaction"
# - Connections returned after each transaction
# - Most efficient
# - Cannot use: prepared statements, LISTEN/NOTIFY, cursors

# session mode (for apps needing session features)
POOL_MODE: "session"
# - Connection held for entire session
# - Can use all PostgreSQL features
# - Less efficient pooling

# statement mode (most aggressive pooling)
POOL_MODE: "statement"
# - Connection returned after each statement
# - Cannot use transactions
# - Rarely used
```

### 5. Regular Connection Audit

```bash
# Create CronJob for connection audit
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-connection-audit
  namespace: postgres
spec:
  schedule: "*/30 * * * *"  # Every 30 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: audit
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              psql -h postgres-postgresql -U postgres <<EOF
              SELECT 
                application_name,
                state,
                count(*) as connections,
                max(now() - state_change) as max_idle_time
              FROM pg_stat_activity
              WHERE state = 'idle'
              GROUP BY application_name, state
              HAVING count(*) > 10;
              EOF
          restartPolicy: OnFailure
```

## Connection Pool Sizing Guide

```
Total connections needed = (Number of app instances × connections per instance) + overhead

Example:
- 3 homepage-api pods × 10 connections = 30
- 2 worker pods × 5 connections = 10
- 1 admin app × 5 connections = 5
- Monitoring tools = 5
- Overhead (admin, maintenance) = 10
Total = 60 connections

Recommended max_connections = 100 (60 × 1.5 buffer)

With PgBouncer:
- max_client_conn = 1000 (all apps)
- default_pool_size = 25 (to PostgreSQL)
- max_db_connections = 50 per database
- PostgreSQL max_connections = 100
```

## Troubleshooting Connection Leaks

```bash
# Find applications with growing connection count
while true; do
  kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
    "SELECT application_name, count(*), 
     max(now() - backend_start) as oldest_connection
     FROM pg_stat_activity
     WHERE application_name != ''
     GROUP BY application_name
     ORDER BY count(*) DESC;"
  sleep 60
done

# If connections grow indefinitely, application has connection leak
# Fix: Review application code and ensure connections are always closed
```

## Related Alerts

- `PostgreSQLDown`
- `PostgreSQLHighMemory`
- `PostgreSQLSlowQueries`

## Escalation

If connection issues persist:

1. ✅ Review all application connection pool configurations
2. 📊 Analyze connection patterns over time
3. 🔍 Check for connection leaks in application code
4. 💾 Consider increasing server resources for more connections
5. 🔄 Evaluate splitting workload across multiple databases
6. 📞 Contact application team to optimize connection usage
7. 🆘 Consider managed PostgreSQL service with auto-scaling

## Additional Resources

- [PostgreSQL Connection Pooling](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [PgBouncer Documentation](https://www.pgbouncer.org/)
- [Connection Pool Sizing](https://wiki.postgresql.org/wiki/Number_Of_Database_Connections)
- [Bitnami PostgreSQL HA](https://github.com/bitnami/charts/tree/main/bitnami/postgresql-ha)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

