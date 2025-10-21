# Redis Runbooks

Comprehensive operational runbooks for managing Redis in Kubernetes.

## 📚 Available Runbooks

### Core Operations

| Runbook | Severity | Description | Est. Resolution Time |
|---------|----------|-------------|---------------------|
| [🚨 Redis Down](redis-down.md) | Critical | Redis service completely unavailable | 5-15 minutes |
| [⚠️ Redis High Memory](redis-high-memory.md) | Warning | Memory usage > 80%, risk of eviction | 10-20 minutes |
| [🐌 Redis Slow Operations](redis-slow-operations.md) | Warning | Commands taking > 100ms | 15-30 minutes |

### Data & Persistence

| Runbook | Severity | Description | Est. Resolution Time |
|---------|----------|-------------|---------------------|
| [💾 Redis Persistence Issues](redis-persistence-issues.md) | High | RDB/AOF persistence failures | 10-30 minutes |
| [📦 Redis Backup & Restore](redis-backup-restore.md) | Info | Backup and recovery procedures | Varies |

### Connectivity & Performance

| Runbook | Severity | Description | Est. Resolution Time |
|---------|----------|-------------|---------------------|
| [🔌 Connection Pool Exhaustion](redis-connection-pool-exhaustion.md) | High | Cannot connect due to pool exhaustion | 10-20 minutes |
| [🔄 Replication Issues](redis-replication-issues.md) | High | Replica sync problems, replication lag | 15-30 minutes |

## 🚀 Quick Start

### Common Commands

```bash
# Check Redis status
kubectl get pods -n redis
kubectl exec -n redis redis-master-0 -- redis-cli ping

# View Redis info
kubectl exec -n redis redis-master-0 -- redis-cli info

# Check memory usage
kubectl exec -n redis redis-master-0 -- redis-cli info memory

# Check replication
kubectl exec -n redis redis-master-0 -- redis-cli info replication

# View connected clients
kubectl exec -n redis redis-master-0 -- redis-cli client list

# Check slow log
kubectl exec -n redis redis-master-0 -- redis-cli slowlog get 10
```

### Health Check Script

```bash
#!/bin/bash
# redis-health-check.sh

echo "=== Redis Health Check ==="
echo ""

# Pod status
echo "Pod Status:"
kubectl get pod -n redis redis-master-0

# Redis connectivity
echo ""
echo "Redis Connectivity:"
kubectl exec -n redis redis-master-0 -- redis-cli ping

# Memory usage
echo ""
echo "Memory Usage:"
kubectl exec -n redis redis-master-0 -- redis-cli info memory | grep -E "used_memory_human|maxmemory_human|mem_fragmentation_ratio"

# Database size
echo ""
echo "Database:"
kubectl exec -n redis redis-master-0 -- redis-cli dbsize

# Replication (if enabled)
echo ""
echo "Replication:"
kubectl exec -n redis redis-master-0 -- redis-cli info replication | grep -E "role|connected_slaves"

# Persistence
echo ""
echo "Last Save:"
kubectl exec -n redis redis-master-0 -- redis-cli lastsave | xargs -I {} date -d @{}

echo ""
echo "✅ Health check complete"
```

## 📊 Monitoring

### Key Metrics to Monitor

#### Availability
- ✅ Pod status and restarts
- ✅ Service endpoints
- ✅ Connection success rate

#### Performance
- ✅ Command latency
- ✅ Operations per second
- ✅ Slow log entries
- ✅ Keyspace hit ratio

#### Resources
- ✅ Memory usage vs limit
- ✅ CPU usage
- ✅ Network throughput
- ✅ Disk I/O (for persistence)

#### Replication (if enabled)
- ✅ Connected replicas count
- ✅ Replication lag
- ✅ Replication link status

### Prometheus Queries

```promql
# Redis availability
up{job="redis"} == 1

# Memory usage percentage
redis_memory_used_bytes / redis_memory_max_bytes * 100

# Operations per second
rate(redis_commands_processed_total[1m])

# Connected clients
redis_connected_clients

# Keyspace hit ratio
rate(redis_keyspace_hits_total[5m]) / 
(rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m])) * 100

# Replication lag (if using replicas)
redis_master_repl_offset - redis_slave_repl_offset
```

### Grafana Dashboards

Recommended dashboard panels:

1. **Overview**
   - Uptime
   - Connected clients
   - Memory usage
   - Operations per second

2. **Performance**
   - Command latency percentiles (p50, p95, p99)
   - Slow log entries
   - Keyspace hit ratio
   - Network throughput

3. **Resources**
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network I/O

4. **Replication**
   - Connected replicas
   - Replication lag
   - Sync operations

## 🎯 Alert Thresholds

### Critical Alerts

```yaml
- alert: RedisDown
  expr: up{job="redis"} == 0
  for: 1m
  severity: critical

- alert: RedisMemoryCritical
  expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.95
  for: 5m
  severity: critical

- alert: RedisPersistenceFailure
  expr: time() - redis_rdb_last_save_timestamp_seconds > 86400
  for: 10m
  severity: critical
```

### Warning Alerts

```yaml
- alert: RedisHighMemory
  expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.80
  for: 10m
  severity: warning

- alert: RedisSlowOperations
  expr: redis_slowlog_length > 10
  for: 10m
  severity: warning

- alert: RedisHighConnections
  expr: redis_connected_clients / redis_config_maxclients > 0.80
  for: 10m
  severity: warning

- alert: RedisReplicationLag
  expr: redis_master_repl_offset - redis_slave_repl_offset > 1000000
  for: 10m
  severity: warning
```

## 🔧 Troubleshooting Decision Tree

```
Redis Issue
│
├─ Cannot connect
│  ├─ Pod not running → See: redis-down.md
│  ├─ Connection timeout → See: redis-connection-pool-exhaustion.md
│  └─ DNS resolution failed → Check Kubernetes DNS
│
├─ Slow performance
│  ├─ High latency → See: redis-slow-operations.md
│  ├─ High memory usage → See: redis-high-memory.md
│  └─ CPU throttling → Scale up resources
│
├─ Data issues
│  ├─ Data loss → See: redis-backup-restore.md
│  ├─ Persistence failing → See: redis-persistence-issues.md
│  └─ Replication lag → See: redis-replication-issues.md
│
└─ Resource issues
   ├─ Out of memory → See: redis-high-memory.md
   ├─ Disk full → See: redis-persistence-issues.md
   └─ Connection limit → See: redis-connection-pool-exhaustion.md
```

## 🛠️ Common Maintenance Tasks

### Daily
- [ ] Check Redis pod status
- [ ] Monitor memory usage trends
- [ ] Review slow log entries
- [ ] Check backup status

### Weekly
- [ ] Review resource utilization
- [ ] Analyze key patterns and sizes
- [ ] Check for keys without TTL
- [ ] Verify replication status (if enabled)
- [ ] Test backup restore procedure

### Monthly
- [ ] Full disaster recovery test
- [ ] Review and optimize configuration
- [ ] Analyze performance metrics
- [ ] Update runbooks with lessons learned
- [ ] Clean up old backups

## 📖 Configuration Examples

### Minimal Setup (Development)

```yaml
master:
  resources:
    limits:
      memory: "256Mi"
      cpu: "200m"
  persistence:
    enabled: false
  configuration: |
    maxmemory 200mb
    maxmemory-policy allkeys-lru
```

### Standard Setup (Production)

```yaml
master:
  resources:
    limits:
      memory: "2Gi"
      cpu: "1000m"
    requests:
      memory: "1Gi"
      cpu: "200m"
  persistence:
    enabled: true
    size: 16Gi
  configuration: |
    maxmemory 1600mb
    maxmemory-policy allkeys-lfu
    save 900 1
    save 300 10
    appendonly yes
    appendfsync everysec
```

### High Availability Setup

```yaml
master:
  resources:
    limits:
      memory: "4Gi"
      cpu: "2000m"
  persistence:
    enabled: true
    size: 32Gi
  configuration: |
    maxmemory 3200mb
    maxmemory-policy allkeys-lfu
    save 300 1
    appendonly yes
    appendfsync everysec
    min-replicas-to-write 1
    min-replicas-max-lag 10

replica:
  replicaCount: 2
  resources:
    limits:
      memory: "4Gi"
      cpu: "2000m"

sentinel:
  enabled: true
  replicas: 3
  quorum: 2
```

## 🔗 Related Documentation

- [Redis Official Documentation](https://redis.io/docs/)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
- [Redis Security](https://redis.io/docs/management/security/)
- [Redis Cluster Tutorial](https://redis.io/docs/manual/scaling/)

## 📞 Escalation Contacts

| Issue Type | Contact | Response Time |
|------------|---------|---------------|
| Redis Down (Critical) | On-call SRE | 15 minutes |
| Performance Issues | Infrastructure Team | 1 hour |
| Data Loss | DBA Team | 30 minutes |
| Configuration Changes | Platform Team | 4 hours |

## 📝 Contributing

When creating or updating runbooks:

1. ✅ Use consistent formatting
2. ✅ Include actual commands that work
3. ✅ Add verification steps
4. ✅ Document rollback procedures
5. ✅ Test all commands before committing
6. ✅ Update the table of contents
7. ✅ Add estimated resolution times
8. ✅ Include related alerts

## 📜 Changelog

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-15 | 1.0 | Initial Redis runbooks created |

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

## Quick Links

- 🏠 [Back to Main Runbooks](../)
- 🔍 [Search Runbooks](https://github.com/search)
- 📊 [Grafana Dashboards](https://grafana.homelab)
- 📈 [Prometheus](https://prometheus.homelab)
- 🚨 [Alerts](https://alertmanager.homelab)

