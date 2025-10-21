# LanceDB High CPU Usage

## 🚨 Alert

**Alert Name:** LanceDBHighCPUUsage  
**Severity:** Warning  
**Component:** LanceDB

## 📊 Description

LanceDB CPU usage is above 85% of the configured limit. This indicates heavy query load or inefficient operations.

## 🔍 Investigation Steps

### 1. Check Current CPU Usage

\`\`\`bash
kubectl top pods -n lancedb
kubectl describe pod -n lancedb -l app=lancedb | grep -A5 "Limits\\|Requests"
\`\`\`

### 2. Check CPU Metrics in Grafana

Navigate to the LanceDB dashboard in Grafana and review:
- CPU usage trends
- Query rate
- Query latency

### 3. Review Application Logs

\`\`\`bash
kubectl logs -n lancedb deployment/lancedb --tail=200
\`\`\`

### 4. Check for Heavy Queries

\`\`\`bash
# Check recent queries if LanceDB provides query logs
kubectl logs -n lancedb deployment/lancedb | grep -i "query\\|search"
\`\`\`

## 🔧 Solutions

### Solution 1: Increase CPU Limits

If CPU usage is legitimate:

\`\`\`bash
kubectl edit deployment lancedb -n lancedb

# Update CPU limits:
# resources:
#   limits:
#     cpu: 2000m  # Increase from 1000m
#   requests:
#     cpu: 500m   # Increase from 200m
\`\`\`

### Solution 2: Optimize Queries

Review and optimize application queries:

\`\`\`python
# Example: Use indexes, limit results, optimize filters
import lancedb

db = lancedb.connect("http://lancedb.lancedb.svc.cluster.local:8000")

# Instead of scanning all data
# results = table.search("query").to_pandas()

# Use filters and limits
results = (table.search("query")
    .where("date > '2025-10-01'")
    .limit(100)
    .to_pandas())
\`\`\`

### Solution 3: Implement Caching

Add caching layer for frequently accessed data:
- Use Redis for query result caching
- Implement application-level caching

### Solution 4: Scale Horizontally

Consider read replicas or sharding for high query loads.

## 📊 Monitoring

Monitor:
- Query patterns
- Query response times
- CPU usage trends
- Number of concurrent requests

## ✅ Resolution Verification

1. Check CPU usage has decreased:
   \`\`\`bash
   kubectl top pods -n lancedb
   \`\`\`

2. Monitor query performance

3. Check application logs for errors

## 📝 Prevention

- Optimize queries before deploying
- Implement query result caching
- Set up query timeouts
- Monitor and alert on query performance
- Consider implementing rate limiting for clients

## 🔗 Related Runbooks

- [LanceDB Down](./lancedb-down.md)
- [High Memory Usage](./high-memory-usage.md)

