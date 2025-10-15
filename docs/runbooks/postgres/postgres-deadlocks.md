# 🚨 Runbook: PostgreSQL Deadlocks

## Alert Information

**Alert Name:** `PostgreSQLDeadlocks`  
**Severity:** Warning  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** Deadlock rate > 1 per minute

## Symptom

Transactions experiencing deadlocks where two or more processes are waiting for each other to release locks, causing transaction failures.

## Impact

- **User Impact:** MEDIUM - Transaction failures, operation retries, degraded performance
- **Business Impact:** MEDIUM - Failed operations, increased latency, user frustration
- **Data Impact:** LOW - No data loss (transactions rolled back), but consistency concerns

## Diagnosis

### 1. Check Deadlock Count

```bash
# Check total deadlocks per database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, deadlocks, 
   deadlocks / NULLIF(EXTRACT(epoch FROM (now() - stats_reset)) / 60, 0) as deadlocks_per_minute
   FROM pg_stat_database
   WHERE datname NOT IN ('template0', 'template1')
   ORDER BY deadlocks DESC;"

# Check when stats were last reset
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT datname, stats_reset FROM pg_stat_database WHERE datname = 'bruno_site';"
```

### 2. Check Recent Deadlocks in Logs

```bash
# Search for deadlock errors in logs
kubectl logs -n postgres postgres-postgresql-0 --tail=500 | grep -i deadlock

# Detailed deadlock information (if log_lock_waits = on)
kubectl logs -n postgres postgres-postgresql-0 --tail=1000 | grep -A 30 "deadlock detected"

# Look for patterns:
# - Which tables are involved
# - Which queries cause deadlocks
# - Transaction isolation levels
```

### 3. Check Current Locks

```bash
# View all current locks
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT locktype, database, relation::regclass, page, tuple, 
   virtualxid, transactionid, classid, objid, objsubid, 
   virtualtransaction, pid, mode, granted, fastpath
   FROM pg_locks
   ORDER BY pid, locktype;"

# Check for locks not granted
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT locktype, relation::regclass, mode, pid, granted
   FROM pg_locks
   WHERE NOT granted
   ORDER BY relation;"
```

### 4. Check Blocking Queries

```bash
# Find blocking queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT 
   blocked_locks.pid AS blocked_pid,
   blocked_activity.usename AS blocked_user,
   blocking_locks.pid AS blocking_pid,
   blocking_activity.usename AS blocking_user,
   blocked_activity.query AS blocked_statement,
   blocking_activity.query AS blocking_statement,
   blocked_activity.application_name AS blocked_application
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

### 5. Check Transaction Isolation Levels

```bash
# Check current transaction settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT name, setting, context FROM pg_settings 
   WHERE name IN ('default_transaction_isolation', 'deadlock_timeout');"

# Check active transactions
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, datname, 
   now() - xact_start AS transaction_duration,
   state, query
   FROM pg_stat_activity
   WHERE xact_start IS NOT NULL
   ORDER BY xact_start;"
```

### 6. Identify Most Locked Tables

```bash
# Tables with most lock activity
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT relation::regclass, mode, count(*) 
   FROM pg_locks 
   WHERE relation IS NOT NULL
   GROUP BY relation, mode
   ORDER BY count(*) DESC
   LIMIT 10;"

# Check table lock conflicts
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "SELECT schemaname, tablename, 
   n_tup_upd, n_tup_del, n_tup_hot_upd,
   n_dead_tup, n_mod_since_analyze
   FROM pg_stat_user_tables
   ORDER BY n_tup_upd + n_tup_del DESC
   LIMIT 10;"
```

## Resolution Steps

### Step 1: Immediate Actions

#### Terminate Blocking Transactions

```bash
# Identify long-running transactions that may be blocking
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, usename, now() - xact_start AS duration, state, query
   FROM pg_stat_activity
   WHERE xact_start < now() - interval '1 minute'
   AND state != 'idle'
   ORDER BY xact_start;"

# Terminate specific blocking transaction (⚠️ Use with caution!)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(<blocking_pid>);"

# Or cancel query without terminating connection
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_cancel_backend(<blocking_pid>);"
```

### Step 2: Enable Deadlock Logging

```bash
# Enable detailed lock logging
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET log_lock_waits = on;"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET deadlock_timeout = '1s';"  # Log locks held > 1s

# Enable verbose deadlock information
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET log_error_verbosity = 'verbose';"

# Reload configuration
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pg_reload_conf();"
```

### Step 3: Fix Common Deadlock Patterns

#### Issue: UPDATE Order Mismatch
**Cause:** Multiple transactions updating rows in different orders  
**Example:**
```sql
-- Transaction 1
BEGIN;
UPDATE users SET status = 'active' WHERE id = 1;
UPDATE users SET status = 'active' WHERE id = 2;
COMMIT;

-- Transaction 2 (running concurrently)
BEGIN;
UPDATE users SET status = 'active' WHERE id = 2;  -- Waits
UPDATE users SET status = 'active' WHERE id = 1;  -- Deadlock!
COMMIT;
```

**Fix:**
```sql
-- Always update in consistent order (e.g., by ID)
BEGIN;
UPDATE users SET status = 'active' WHERE id IN (1, 2) ORDER BY id;
COMMIT;

-- Or use explicit locking order
BEGIN;
SELECT * FROM users WHERE id IN (1, 2) ORDER BY id FOR UPDATE;
UPDATE users SET status = 'active' WHERE id IN (1, 2);
COMMIT;
```

#### Issue: Foreign Key Lock Conflicts
**Cause:** Parent and child table updates causing lock conflicts  
**Example:**
```sql
-- Transaction 1
BEGIN;
UPDATE orders SET status = 'complete' WHERE id = 1;
UPDATE order_items SET shipped = true WHERE order_id = 1;
COMMIT;

-- Transaction 2
BEGIN;
UPDATE order_items SET shipped = true WHERE order_id = 1;  -- Waits
UPDATE orders SET status = 'complete' WHERE id = 1;  -- Deadlock!
COMMIT;
```

**Fix:**
```sql
-- Lock parent first, then children
BEGIN;
SELECT * FROM orders WHERE id = 1 FOR UPDATE;
UPDATE orders SET status = 'complete' WHERE id = 1;
UPDATE order_items SET shipped = true WHERE order_id = 1;
COMMIT;
```

#### Issue: Read-Then-Write Pattern
**Cause:** Multiple transactions reading then updating same rows  
**Fix:**
```sql
-- Instead of SELECT then UPDATE
BEGIN;
SELECT * FROM inventory WHERE product_id = 1;
-- ... check quantity ...
UPDATE inventory SET quantity = quantity - 1 WHERE product_id = 1;
COMMIT;

-- Use SELECT FOR UPDATE
BEGIN;
SELECT * FROM inventory WHERE product_id = 1 FOR UPDATE;
-- ... check quantity (row is locked) ...
UPDATE inventory SET quantity = quantity - 1 WHERE product_id = 1;
COMMIT;

-- Or use optimistic locking with version column
BEGIN;
SELECT quantity, version FROM inventory WHERE product_id = 1;
UPDATE inventory 
SET quantity = quantity - 1, version = version + 1
WHERE product_id = 1 AND version = <old_version>;
-- Check affected rows, retry if 0
COMMIT;
```

### Step 4: Application-Level Fixes

#### Implement Retry Logic

```python
# Python example with retry logic
import psycopg2
from psycopg2 import OperationalError
import time

def execute_with_retry(conn, query, max_retries=3):
    for attempt in range(max_retries):
        try:
            cur = conn.cursor()
            cur.execute(query)
            conn.commit()
            cur.close()
            return True
        except OperationalError as e:
            if "deadlock detected" in str(e):
                if attempt < max_retries - 1:
                    # Exponential backoff
                    time.sleep(0.1 * (2 ** attempt))
                    conn.rollback()
                    continue
                else:
                    raise
            else:
                raise
    return False
```

```javascript
// Node.js example with retry logic
async function executeWithRetry(client, query, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      await client.query('BEGIN');
      await client.query(query);
      await client.query('COMMIT');
      return;
    } catch (error) {
      await client.query('ROLLBACK');
      
      if (error.code === '40P01' && attempt < maxRetries - 1) {
        // Deadlock detected, retry with exponential backoff
        await new Promise(resolve => 
          setTimeout(resolve, 100 * Math.pow(2, attempt))
        );
        continue;
      }
      throw error;
    }
  }
}
```

#### Reduce Transaction Scope

```python
# ❌ BAD: Long transaction with multiple operations
def process_order(order_id):
    conn = psycopg2.connect(...)
    cur = conn.cursor()
    
    cur.execute("BEGIN")
    cur.execute("UPDATE orders SET status = 'processing' WHERE id = %s", (order_id,))
    
    # ... lots of processing ...
    time.sleep(10)  # Simulating work
    
    cur.execute("UPDATE inventory SET quantity = quantity - 1 WHERE product_id = 123")
    cur.execute("COMMIT")

# ✅ GOOD: Shorter transactions
def process_order(order_id):
    conn = psycopg2.connect(...)
    
    # Transaction 1: Quick update
    cur = conn.cursor()
    cur.execute("BEGIN")
    cur.execute("UPDATE orders SET status = 'processing' WHERE id = %s", (order_id,))
    cur.execute("COMMIT")
    
    # Do processing outside transaction
    # ... lots of processing ...
    time.sleep(10)
    
    # Transaction 2: Quick inventory update
    cur = conn.cursor()
    cur.execute("BEGIN")
    cur.execute("UPDATE inventory SET quantity = quantity - 1 WHERE product_id = 123")
    cur.execute("COMMIT")
```

#### Use Consistent Lock Ordering

```python
# ✅ Always lock resources in same order (e.g., by ID)
def update_users(user_ids):
    conn = psycopg2.connect(...)
    cur = conn.cursor()
    
    # Sort IDs to ensure consistent locking order
    sorted_ids = sorted(user_ids)
    
    cur.execute("BEGIN")
    # Lock all rows in consistent order
    cur.execute(
        "SELECT * FROM users WHERE id = ANY(%s) ORDER BY id FOR UPDATE",
        (sorted_ids,)
    )
    # Now safe to update
    cur.execute("UPDATE users SET status = 'active' WHERE id = ANY(%s)", (sorted_ids,))
    cur.execute("COMMIT")
```

### Step 5: Database Configuration Tuning

```bash
# Adjust deadlock timeout (time before checking for deadlock)
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET deadlock_timeout = '1s';"  # Default: 1s

# Consider READ COMMITTED isolation level (default)
# vs REPEATABLE READ or SERIALIZABLE
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SHOW default_transaction_isolation;"

# Reload configuration
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pg_reload_conf();"
```

### Step 6: Add Indexes to Reduce Lock Contention

```bash
# Add indexes on frequently locked columns
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_orders_status ON orders(status);"

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE INDEX CONCURRENTLY idx_order_items_order_id ON order_items(order_id);"

# Indexes can reduce lock contention by:
# - Reducing number of rows scanned
# - Enabling index-only scans
# - Improving query performance (shorter lock duration)
```

## Verification

### 1. Check Deadlock Rate Decreased

```bash
# Monitor deadlock count over time
watch -n 10 'kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -t -c "SELECT datname, deadlocks FROM pg_stat_database WHERE datname = '\''bruno_site'\'';"'

# Should show stable or decreasing count
```

### 2. Monitor Lock Waits

```bash
# Check for lock waits
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) FROM pg_locks WHERE NOT granted;"

# Should be 0 or very low
```

### 3. Check Application Logs

```bash
# Check application logs for deadlock errors
kubectl logs -n homepage deployment/homepage-api --tail=100 | grep -i deadlock

# Should see reduced deadlock errors
```

### 4. Review Transaction Durations

```bash
# Check transaction durations
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, now() - xact_start AS duration, state, query
   FROM pg_stat_activity
   WHERE xact_start IS NOT NULL
   ORDER BY duration DESC
   LIMIT 10;"

# Should show shorter transaction times
```

### 5. Verify Lock Patterns

```bash
# Check lock distribution
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT locktype, mode, count(*) 
   FROM pg_locks 
   GROUP BY locktype, mode
   ORDER BY count(*) DESC;"
```

## Prevention

### 1. Application Design Best Practices

- ✅ Keep transactions short and focused
- ✅ Access tables in consistent order
- ✅ Use explicit locking (SELECT FOR UPDATE) when needed
- ✅ Implement retry logic for deadlock errors
- ✅ Use optimistic locking with version columns
- ✅ Avoid mixing read and write operations in transactions
- ✅ Use appropriate isolation levels

### 2. Code Review Checklist

```python
# Transaction code review checklist:
# [ ] Transaction duration < 1 second
# [ ] Resources locked in consistent order (e.g., sorted by ID)
# [ ] Explicit locking used where appropriate (FOR UPDATE)
# [ ] Retry logic implemented for deadlock errors
# [ ] No external API calls within transaction
# [ ] No file I/O within transaction
# [ ] Proper exception handling and rollback
```

### 3. Enable Comprehensive Logging

```yaml
# In HelmRelease configuration
primary:
  configuration: |
    # Lock and deadlock logging
    log_lock_waits = on
    deadlock_timeout = 1s
    log_error_verbosity = verbose
    
    # Statement logging
    log_statement = 'mod'  # Log all modifications
    log_duration = on
    log_min_duration_statement = 1000  # Log queries > 1s
```

### 4. Set Up Monitoring

```yaml
# Prometheus alert rules
- alert: PostgreSQLDeadlocksHigh
  expr: |
    rate(pg_stat_database_deadlocks_total[5m]) > 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL deadlock rate high"

- alert: PostgreSQLLockWaits
  expr: |
    pg_locks_count{mode="ExclusiveLock",granted="false"} > 5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL has waiting locks"
```

### 5. Regular Lock Analysis

```bash
# Create CronJob for lock analysis
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-lock-analysis
  namespace: postgres
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: analysis
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              psql -h postgres-postgresql -U postgres -d bruno_site <<EOF
              -- Check deadlock count
              SELECT datname, deadlocks FROM pg_stat_database WHERE datname = 'bruno_site';
              
              -- Check most locked tables
              SELECT relation::regclass, mode, count(*) 
              FROM pg_locks 
              WHERE relation IS NOT NULL
              GROUP BY relation, mode
              ORDER BY count(*) DESC
              LIMIT 10;
              EOF
          restartPolicy: OnFailure
```

## Common Deadlock Scenarios

### Scenario 1: Multiple Row Updates

```sql
-- Problem: Two transactions updating different rows in different orders
-- Transaction 1: UPDATE row 1, then row 2
-- Transaction 2: UPDATE row 2, then row 1

-- Solution: Always update in consistent order
UPDATE users SET status = 'active' WHERE id IN (1, 2) ORDER BY id;
```

### Scenario 2: Parent-Child Updates

```sql
-- Problem: One transaction updates parent then child, 
-- another updates child then parent

-- Solution: Always lock parent first
BEGIN;
SELECT * FROM orders WHERE id = 1 FOR UPDATE;
UPDATE orders SET total = 100 WHERE id = 1;
UPDATE order_items SET price = 50 WHERE order_id = 1;
COMMIT;
```

### Scenario 3: Read-Modify-Write

```sql
-- Problem: SELECT then UPDATE without locking

-- Solution: Use SELECT FOR UPDATE
BEGIN;
SELECT quantity FROM inventory WHERE product_id = 1 FOR UPDATE;
-- Now row is locked, safe to update
UPDATE inventory SET quantity = quantity - 1 WHERE product_id = 1;
COMMIT;
```

## Deadlock Analysis Script

```bash
#!/bin/bash
# deadlock-analysis.sh

NAMESPACE="postgres"
POD="postgres-postgresql-0"

echo "=== PostgreSQL Deadlock Analysis ==="
echo ""

echo "1. Deadlock Count:"
kubectl exec -n $NAMESPACE $POD -- psql -U postgres -c \
  "SELECT datname, deadlocks FROM pg_stat_database WHERE datname NOT IN ('template0', 'template1');"

echo ""
echo "2. Recent Deadlocks in Logs:"
kubectl logs -n $NAMESPACE $POD --tail=1000 | grep -i "deadlock" | tail -20

echo ""
echo "3. Current Lock Waits:"
kubectl exec -n $NAMESPACE $POD -- psql -U postgres -c \
  "SELECT count(*) as waiting_locks FROM pg_locks WHERE NOT granted;"

echo ""
echo "4. Most Locked Tables:"
kubectl exec -n $NAMESPACE $POD -- psql -U postgres -c \
  "SELECT relation::regclass, mode, count(*) 
   FROM pg_locks 
   WHERE relation IS NOT NULL
   GROUP BY relation, mode
   ORDER BY count(*) DESC
   LIMIT 10;"

echo ""
echo "5. Long Running Transactions:"
kubectl exec -n $NAMESPACE $POD -- psql -U postgres -c \
  "SELECT pid, usename, now() - xact_start AS duration, state
   FROM pg_stat_activity
   WHERE xact_start IS NOT NULL
   ORDER BY xact_start
   LIMIT 10;"
```

## Related Alerts

- `PostgreSQLDown`
- `PostgreSQLSlowQueries`
- `PostgreSQLHighConnections`

## Escalation

If deadlocks persist after applying fixes:

1. ✅ Review application transaction logic with development team
2. 📊 Analyze deadlock patterns from logs
3. 🔍 Profile application code for lock ordering issues
4. 💾 Consider application architecture changes
5. 🔄 Evaluate queue-based processing for conflicting operations
6. 📞 Consult database team for advanced locking strategies
7. 🆘 Consider database sharding if hot-spot contention

## Additional Resources

- [PostgreSQL Deadlocks](https://www.postgresql.org/docs/current/explicit-locking.html)
- [Lock Management](https://www.postgresql.org/docs/current/monitoring-locks.html)
- [Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [MVCC and Locking](https://www.postgresql.org/docs/current/mvcc.html)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

