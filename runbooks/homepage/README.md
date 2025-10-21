# 📚 Homepage Runbooks

This directory contains operational runbooks for the Bruno Site (Homepage) application running in the `homepage` namespace.

## 📋 Overview

Each runbook provides step-by-step guidance for responding to specific alerts from Prometheus. They include:
- Alert information and severity
- Symptoms and impact analysis
- Diagnostic commands
- Resolution steps
- Prevention strategies
- Related alerts
- Escalation procedures

## 🚨 Critical Alerts

### Service Availability
- [api-down.md](./api-down.md) - API service is completely down
- [database-down.md](./database-down.md) - PostgreSQL database is unavailable
- [redis-down.md](./redis-down.md) - Redis cache is unavailable
- [high-error-rate.md](./high-error-rate.md) - API error rate exceeds 5%
- [low-availability.md](./low-availability.md) - Overall availability below 99%
- [health-check-failures.md](./health-check-failures.md) - Health endpoint failures

### Database & Connections
- [database-connection-issues.md](./database-connection-issues.md) - Database deadlocks
- [database-slow-queries.md](./database-slow-queries.md) - Inefficient query patterns
- [high-connection-pool-usage.md](./high-connection-pool-usage.md) - Connection pool at 80%+
- [slow-database-queries.md](./slow-database-queries.md) - Queries fetching excessive tuples

### Data Operations
- [experience-data-load-failure.md](./experience-data-load-failure.md) - Experience data loading failures
- [experience-data-high-error-rate.md](./experience-data-high-error-rate.md) - High error rate for experience data
- [experience-data-slow-load.md](./experience-data-slow-load.md) - Slow experience data loading
- [experience-database-unavailable.md](./experience-database-unavailable.md) - Database unavailable for experience data
- [projects-load-failed.md](./projects-load-failed.md) - Projects data loading failures
- [projects-slow-load.md](./projects-slow-load.md) - Slow projects data loading

## ⚠️ Warning Alerts

### Performance
- [high-response-time.md](./high-response-time.md) - P95 response time > 1 second
- [high-cpu-usage.md](./high-cpu-usage.md) - CPU usage > 80%
- [high-memory-usage.md](./high-memory-usage.md) - Memory usage > 90%
- [redis-high-memory.md](./redis-high-memory.md) - Redis memory > 80%
- [redis-slow-operations.md](./redis-slow-operations.md) - Redis operations > 0.1s

### Application Features
- [project-errors.md](./project-errors.md) - High error rate for project views
- [chat-errors.md](./chat-errors.md) - AI chat API errors
- [high-analytics-load.md](./high-analytics-load.md) - High analytics tracking load
- [llm-service-down.md](./llm-service-down.md) - Ollama LLM service unavailable

### Infrastructure
- [pod-crash-loop.md](./pod-crash-loop.md) - Pod continuously restarting
- [pod-not-ready.md](./pod-not-ready.md) - Pod not in Running state
- [hpa-high-replicas.md](./hpa-high-replicas.md) - HPA scaling to high replica count
- [storage-space-low.md](./storage-space-low.md) - Storage space < 10%

## 🔒 Security & Abuse Alerts

### Rate Limiting & DDoS
- [abuse-high-rate.md](./abuse-high-rate.md) - Extremely high request rate (>10 req/sec)
- [abuse-repeated-errors.md](./abuse-repeated-errors.md) - High rate of 4xx errors
- [abuse-api-scraping.md](./abuse-api-scraping.md) - API scraping pattern detected
- [abuse-bandwidth.md](./abuse-bandwidth.md) - Excessive bandwidth consumption

### Attack Detection
- [rate-limit-hits.md](./rate-limit-hits.md) - High rate limit hits
- [suspicious-activity.md](./suspicious-activity.md) - Suspicious traffic patterns
- [sql-injection-attempts.md](./sql-injection-attempts.md) - Potential SQL injection attempts

## 🎯 Usage

### 1. When an Alert Fires

1. Check the alert in Grafana or Prometheus
2. Find the corresponding runbook using the alert name
3. Follow the diagnostic steps
4. Apply the resolution
5. Verify the issue is resolved
6. Document any additional findings

### 2. Runbook Structure

Each runbook follows this structure:

```markdown
# Alert Name

## Alert Information
- Alert Name
- Severity
- Component
- Service

## Symptom
What you're seeing

## Impact
User, Business, and Data impact

## Diagnosis
Commands to identify the problem

## Resolution Steps
Step-by-step fix instructions

## Verification
How to confirm it's fixed

## Prevention
How to avoid this in the future

## Related Alerts
Other alerts that may be related

## Escalation
When to escalate and to whom
```

### 3. Quick Commands

```bash
# View all runbooks
ls -1 /Users/brunolucena/workspace/bruno/repos/homelab/docs/runbooks/homepage/

# Search for a specific topic
grep -r "database" /Users/brunolucena/workspace/bruno/repos/homelab/docs/runbooks/homepage/

# Count total runbooks
ls -1 /Users/brunolucena/workspace/bruno/repos/homelab/docs/runbooks/homepage/*.md | wc -l
```

## 🔗 Related Documentation

- [Homepage Architecture](../../../flux/clusters/homelab/infrastructure/homepage/ARCHITECTURE.md)
- [Homepage README](../../../flux/clusters/homelab/infrastructure/homepage/README.md)
- [Homepage Security](../../../flux/clusters/homelab/infrastructure/homepage/SECURITY.md)
- [Frontend Metrics](../../../flux/clusters/homelab/infrastructure/homepage/FRONTEND_METRICS.md)
- [Prometheus Rules](../../../flux/clusters/homelab/infrastructure/homepage/chart/templates/prometheus-rules.yaml)

## 📊 Alert Categories

| Category | Count | Severity |
|----------|-------|----------|
| Service Availability | 6 | Critical |
| Database & Connections | 4 | Critical/Warning |
| Data Operations | 6 | Critical/Warning |
| Performance | 5 | Warning |
| Application Features | 4 | Warning |
| Infrastructure | 4 | Warning/Critical |
| Security & Abuse | 7 | Critical |

**Total Runbooks: 36**

## 🆘 Emergency Contacts

For critical issues that cannot be resolved using these runbooks:
1. Check related alerts for cascading failures
2. Review recent deployments and changes
3. Escalate to on-call engineer
4. Document the incident for post-mortem

## 📝 Contributing

When adding new alerts:
1. Create a corresponding runbook
2. Follow the standard structure
3. Include specific commands for your environment
4. Test the runbook procedures
5. Update this README

## 🔄 Maintenance

These runbooks should be reviewed and updated:
- After each incident (lessons learned)
- When infrastructure changes
- Quarterly for accuracy
- When new alerts are added

---

Last Updated: 2025-10-15  
Maintained by: SRE Team

