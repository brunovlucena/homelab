# 📚 PostgreSQL Runbooks

Comprehensive operational runbooks for troubleshooting and resolving PostgreSQL issues in the homelab Kubernetes cluster.

## Overview

PostgreSQL is our primary relational database running in standalone mode:
- **Architecture**: Standalone (single primary instance)
- **Namespace**: postgres
- **Storage**: 20Gi persistent volume
- **Port**: 5432
- **Version**: PostgreSQL 16
- **Authentication**: Trust method (internal cluster use only)
- **Metrics**: Enabled with Prometheus ServiceMonitor
- **Resource Allocation**:
  - Requests: 200m CPU, 512Mi memory
  - Limits: 1000m CPU, 2Gi memory

## Quick Reference

| Alert | Severity | Impact | Runbook |
|-------|----------|--------|---------|
| PostgreSQLDown | Critical | Complete database outage | [postgres-down.md](./postgres-down.md) |
| PostgreSQLHighMemory | Warning | Memory pressure/OOMKills | [postgres-high-memory.md](./postgres-high-memory.md) |
| PostgreSQLSlowQueries | Warning | Slow database operations | [postgres-slow-queries.md](./postgres-slow-queries.md) |
| PostgreSQLHighConnections | Warning | Connection pool exhaustion | [postgres-high-connections.md](./postgres-high-connections.md) |
| PostgreSQLReplicationLag | Warning | Replication lag (if enabled) | [postgres-replication-lag.md](./postgres-replication-lag.md) |
| PostgreSQLStorageFull | Critical | Disk space exhausted | [postgres-storage-full.md](./postgres-storage-full.md) |
| PostgreSQLDeadlocks | Warning | Transaction deadlocks | [postgres-deadlocks.md](./postgres-deadlocks.md) |

## Runbooks

### 🚨 Critical Issues

#### [PostgreSQL Service Down](./postgres-down.md)
Complete PostgreSQL outage - database unavailable.

**Quick Check:**
```bash
kubectl get pods -n postgres
```

**Quick Fix:**
```bash
# Restart PostgreSQL
kubectl rollout restart statefulset -n postgres postgres-postgresql
```

---

#### [Storage Full](./postgres-storage-full.md)
PostgreSQL storage volume full - cannot write new data.

**Quick Check:**
```bash
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data
```

**Quick Fix:**
```bash
# Increase PVC size or clean old WAL files
kubectl edit pvc -n postgres data-postgres-postgresql-0
```

---

### ⚠️ Warning Issues

#### [High Memory Usage](./postgres-high-memory.md)
PostgreSQL experiencing memory pressure or OOMKills.

**Quick Check:**
```bash
kubectl top pods -n postgres
kubectl get pods -n postgres -o json | jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'
```

**Quick Fix:**
```bash
# Increase memory limits
kubectl edit helmrelease -n postgres postgres
# Update memory limits
```

---

#### [Slow Queries](./postgres-slow-queries.md)
Database queries taking longer than expected.

**Quick Check:**
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '1 second' ORDER BY duration DESC;"
```

**Quick Fix:**
```bash
# Check for missing indexes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT schemaname, tablename, indexname FROM pg_indexes WHERE schemaname NOT IN ('pg_catalog', 'information_schema');"
```

---

#### [High Connections](./postgres-high-connections.md)
PostgreSQL connection count approaching limits.

**Quick Check:**
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT count(*) as connections, max_conn FROM pg_stat_activity, (SELECT setting::int as max_conn FROM pg_settings WHERE name='max_connections') x GROUP BY max_conn;"
```

**Quick Fix:**
```bash
# Increase max connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "ALTER SYSTEM SET max_connections = 300;"
kubectl rollout restart statefulset -n postgres postgres-postgresql
```

---

#### [Deadlocks](./postgres-deadlocks.md)
Transactions experiencing deadlocks.

**Quick Check:**
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT * FROM pg_stat_database WHERE datname = 'bruno_site';"
kubectl logs -n postgres postgres-postgresql-0 --tail=100 | grep -i deadlock
```

**Quick Fix:**
```bash
# Review and optimize transaction logic
# Check blocking queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pid, usename, pg_blocking_pids(pid) as blocked_by, query FROM pg_stat_activity WHERE cardinality(pg_blocking_pids(pid)) > 0;"
```

---

## Common Troubleshooting Commands

### Check Overall Health
```bash
# All pods status
kubectl get pods -n postgres

# Resource usage
kubectl top pods -n postgres

# Recent events
kubectl get events -n postgres --sort-by='.lastTimestamp' | head -20
```

### Check PostgreSQL Status
```bash
# Connect to psql
kubectl exec -it -n postgres postgres-postgresql-0 -- psql -U postgres

# Check version
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT version();"

# Check server status
kubectl exec -n postgres postgres-postgresql-0 -- pg_isready -U postgres

# Check database list
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "\l"

# Check active connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT datname, count(*) FROM pg_stat_activity GROUP BY datname;"
```

### Check Storage
```bash
# Check disk usage
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data

# Check PVC status
kubectl get pvc -n postgres

# Check database sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pg_database.datname, pg_size_pretty(pg_database_size(pg_database.datname)) AS size FROM pg_database ORDER BY pg_database_size(pg_database.datname) DESC;"

# Check table sizes
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

### Check Performance
```bash
# Check current activity
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pid, usename, datname, state, wait_event_type, query FROM pg_stat_activity WHERE state != 'idle';"

# Check slow queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '1 second' ORDER BY duration DESC LIMIT 10;"

# Check locks
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT locktype, database, relation::regclass, mode, granted FROM pg_locks WHERE NOT granted;"

# Check blocking queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT pid, usename, pg_blocking_pids(pid) as blocked_by, query FROM pg_stat_activity WHERE cardinality(pg_blocking_pids(pid)) > 0;"

# Check cache hit ratio
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT sum(heap_blks_read) as heap_read, sum(heap_blks_hit) as heap_hit, sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as cache_hit_ratio FROM pg_statio_user_tables;"
```

### Check Logs
```bash
# View PostgreSQL logs
kubectl logs -n postgres postgres-postgresql-0 --tail=100

# Follow logs
kubectl logs -n postgres postgres-postgresql-0 -f

# Check for errors
kubectl logs -n postgres postgres-postgresql-0 --tail=500 | grep -i "error\|fatal\|panic"

# Check for slow queries
kubectl logs -n postgres postgres-postgresql-0 --tail=500 | grep "duration:"
```

### Test Connectivity
```bash
# Test from homepage API
kubectl exec -n homepage deployment/homepage-api -- nc -zv postgres-postgresql.postgres.svc.cluster.local 5432

# Test from another pod
kubectl run postgres-test --image=postgres:16 --rm -it --restart=Never -- psql postgresql://postgres@postgres-postgresql.postgres.svc.cluster.local:5432/postgres -c "SELECT 1;"
```

## Architecture

```
┌─────────────────┐
│  Applications   │
│  - homepage     │
│  - other apps   │
└────────┬────────┘
         │
         │ postgresql://postgres-postgresql.postgres.svc:5432
         ▼
┌─────────────────┐
│ PostgreSQL Svc  │
│   (ClusterIP)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ PostgreSQL Pod  │
│ postgres-0      │
│ (StatefulSet)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Persistent Vol │
│     (20Gi)      │
│ /var/lib/postgre│
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Prometheus     │◄── Metrics Exporter
└─────────────────┘
```

## Configuration

**Location**: `flux/clusters/homelab/infrastructure/postgres/helmrelease.yaml`

**Key Settings**:
- Chart: postgresql (Bitnami)
- Version: 16.4.4
- Architecture: Standalone
- Auth: Trust method (no password for internal use)
- Persistence: 20Gi
- Metrics: Enabled with ServiceMonitor
- Max connections: 200
- Shared buffers: 256MB
- Effective cache size: 1GB

## Database Management

### Create Database
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "CREATE DATABASE myapp;"
```

### List Databases
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "\l"
```

### Create User
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "CREATE USER myuser WITH PASSWORD 'mypassword';"
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE myapp TO myuser;"
```

### Create Index
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d myapp -c "CREATE INDEX idx_users_email ON users(email);"
```

### Check Indexes
```bash
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d myapp -c "\di"
```

### Backup Database
```bash
# Dump database
kubectl exec -n postgres postgres-postgresql-0 -- pg_dump -U postgres myapp > postgres-backup-$(date +%Y%m%d).sql

# Or use kubectl cp
kubectl exec -n postgres postgres-postgresql-0 -- pg_dump -U postgres myapp > /tmp/backup.sql
kubectl cp postgres/postgres-postgresql-0:/tmp/backup.sql ./postgres-backup-$(date +%Y%m%d).sql
```

### Restore Database
```bash
# Copy backup to pod
kubectl cp ./postgres-backup.sql postgres/postgres-postgresql-0:/tmp/restore.sql

# Restore
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres myapp < /tmp/restore.sql
```

### Vacuum and Analyze
```bash
# Vacuum database
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d myapp -c "VACUUM ANALYZE;"

# Check last vacuum time
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d myapp -c "SELECT schemaname, tablename, last_vacuum, last_autovacuum FROM pg_stat_user_tables;"
```

## Performance Tuning

### Connection Pool Settings
For application clients (example for homepage):
```env
POSTGRES_HOST=postgres-postgresql.postgres.svc.cluster.local
POSTGRES_PORT=5432
POSTGRES_DB=bruno_site
POSTGRES_USER=postgres
POSTGRES_MAX_CONNECTIONS=20
POSTGRES_MIN_CONNECTIONS=5
POSTGRES_IDLE_TIMEOUT=30000
```

### Query Optimization
```bash
# Enable query logging for slow queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "ALTER SYSTEM SET log_min_duration_statement = 1000;" # Log queries > 1s
kubectl rollout restart statefulset -n postgres postgres-postgresql

# Explain query plan
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c "EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';"

# Check index usage
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c "SELECT schemaname, tablename, indexname, idx_scan FROM pg_stat_user_indexes ORDER BY idx_scan;"
```

### Memory Management
```bash
# Check current settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW shared_buffers;"
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW effective_cache_size;"
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SHOW work_mem;"
```

## Monitoring

### Key Metrics to Monitor
- Connection count and utilization
- Query execution time
- Cache hit ratio
- Disk I/O and storage usage
- Replication lag (if applicable)
- Lock contention and deadlocks
- WAL generation rate

### Prometheus Queries
```promql
# PostgreSQL up status
up{job="postgres"}

# Connection count
pg_stat_activity_count

# Transaction rate
rate(pg_stat_database_xact_commit_total[5m])

# Cache hit ratio
pg_stat_database_blks_hit / (pg_stat_database_blks_hit + pg_stat_database_blks_read)

# Deadlocks
rate(pg_stat_database_deadlocks_total[5m])
```

## Escalation Matrix

| Issue | First Response | Escalation Time | Escalate To |
|-------|---------------|-----------------|-------------|
| Complete outage | Restart pod | 15 minutes | Database team |
| Storage full | Expand PVC/cleanup | 30 minutes | Storage admin |
| High memory | Increase limits | 30 minutes | Capacity planning |
| Slow queries | Check indexes, analyze | 1 hour | Database team |
| Connection issues | Check app pools | 30 minutes | Application team |
| Deadlocks | Analyze query patterns | 1 hour | Application team |

## Related Documentation

- [PostgreSQL Configuration](../../../flux/clusters/homelab/infrastructure/postgres/helmrelease.yaml)
- [Homepage Database Configuration](../homepage/database-down.md)
- [Architecture Overview](../../../ARCHITECTURE.md)
- [PostgreSQL Official Docs](https://www.postgresql.org/docs/)

## Best Practices

1. **Indexes**: Create indexes for frequently queried columns
2. **Connection Pooling**: Use connection pools (e.g., PgBouncer) in applications
3. **Query Optimization**: Use EXPLAIN ANALYZE to optimize queries
4. **Vacuuming**: Regular VACUUM ANALYZE for table maintenance
5. **Monitoring**: Set up alerts for connection count, slow queries, and locks
6. **Backups**: Schedule regular pg_dump backups
7. **Resource Limits**: Monitor and adjust shared_buffers and work_mem

## Support

For issues not covered by these runbooks:
1. Check PostgreSQL logs: `kubectl logs -n postgres postgres-postgresql-0`
2. Review HelmRelease: `flux get helmreleases -n postgres`
3. Consult [PostgreSQL documentation](https://www.postgresql.org/docs/)
4. Check [PostgreSQL Community](https://www.postgresql.org/community/)

---

**Last Updated**: 2025-10-15  
**PostgreSQL Version**: 16.x (managed by Helm chart 16.4.4)  
**Maintainer**: Homelab Platform Team

