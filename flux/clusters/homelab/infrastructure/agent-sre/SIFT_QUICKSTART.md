# 🚀 Sift Quick Start

Get started with Grafana Sift in Agent-SRE in 5 minutes.

## Prerequisites

- Agent-SRE MCP Server running
- Loki deployed and accessible
- Tempo deployed and accessible
- Logs and traces flowing into your observability stack

## Step 1: Verify Setup

Check that Sift is enabled:

```bash
# Check MCP server logs
kubectl logs -l app=agent-sre-mcp-server -n agent-sre | grep Sift

# Expected output:
# 🚀 Starting Agent-SRE MCP Server with Sift
# 📝 Loki URL: http://loki-gateway.loki.svc.cluster.local:80
# 🔍 Tempo URL: http://tempo.tempo.svc.cluster.local:3100
```

## Step 2: Create Your First Investigation

### Via MCP Tool

```python
from mcp_client import MCPClient

client = MCPClient("http://agent-sre-mcp-server.agent-sre:3000")

# Create investigation
result = await client.call_tool(
    "sift_create_investigation",
    {
        "name": "API Performance Investigation",
        "labels": {
            "cluster": "production",
            "namespace": "api"
        }
    }
)

investigation_id = result["investigation"]["id"]
print(f"Created investigation: {investigation_id}")
```

### Via Direct API

```bash
curl -X POST http://agent-sre-mcp-server.agent-sre:3000/mcp/tool \
  -H "Content-Type: application/json" \
  -d '{
    "name": "sift_create_investigation",
    "arguments": {
      "name": "API Performance Investigation",
      "labels": {
        "cluster": "production",
        "namespace": "api"
      }
    }
  }'
```

## Step 3: Run Error Pattern Analysis

This will analyze logs from Loki for elevated error patterns:

```python
# Run error pattern analysis
result = await client.call_tool(
    "sift_run_error_pattern_analysis",
    {
        "investigation_id": investigation_id
    }
)

# Review results
analysis = result["analysis"]
print(f"Status: {analysis['status']}")

if analysis['result']:
    patterns = analysis['result']['elevated_patterns']
    for pattern in patterns:
        print(f"\n🔴 {pattern['severity'].upper()}")
        print(f"   Pattern: {pattern['pattern']}")
        print(f"   Count: {pattern['current_count']}")
        print(f"   Elevation: {pattern['elevation_factor']}x")
```

## Step 4: Run Slow Request Analysis

This will analyze traces from Tempo for slow requests:

```python
# Run slow request analysis
result = await client.call_tool(
    "sift_run_slow_request_analysis",
    {
        "investigation_id": investigation_id
    }
)

# Review results
analysis = result["analysis"]
slow_operations = analysis['result']['slow_operations']

for op in slow_operations:
    print(f"\n⏱️ {op['severity'].upper()}")
    print(f"   Operation: {op['operation']}")
    print(f"   Current P95: {op['current_p95_ms']}ms")
    print(f"   Baseline P95: {op['baseline_p95_ms']}ms")
    print(f"   Slowdown: {op['slowdown_factor']}x")
```

## Step 5: Review Investigation

Get the complete investigation with all analyses:

```python
# Get full investigation
result = await client.call_tool(
    "sift_get_investigation",
    {
        "investigation_id": investigation_id
    }
)

investigation = result["investigation"]
print(f"\nInvestigation: {investigation['name']}")
print(f"Status: {investigation['status']}")
print(f"Analyses: {len(investigation['analyses'])}")

for analysis in investigation['analyses']:
    print(f"\n  - {analysis['type']}: {analysis['status']}")
```

## Example Output

### Error Pattern Analysis

```
🔴 CRITICAL
   Pattern: ERROR Database connection timeout
   Count: 45
   Elevation: 9.0x

⚠️ HIGH
   Pattern: ERROR API rate limit exceeded
   Count: 23
   Elevation: 4.5x
```

### Slow Request Analysis

```
⏱️ CRITICAL
   Operation: GET /api/users
   Current P95: 2500ms
   Baseline P95: 500ms
   Slowdown: 5.0x

⚠️ HIGH
   Operation: POST /api/orders
   Current P95: 1800ms
   Baseline P95: 600ms
   Slowdown: 3.0x
```

## Common Use Cases

### 1. Incident Investigation

When an incident occurs:

```python
# Create investigation for incident time window
investigation = await client.call_tool(
    "sift_create_investigation",
    {
        "name": f"Incident {incident_id} Investigation",
        "labels": {
            "cluster": incident.cluster,
            "namespace": incident.namespace
        },
        "start_time": incident.start_time.isoformat(),
        "end_time": incident.end_time.isoformat()
    }
)

# Run both analyses
await client.call_tool("sift_run_error_pattern_analysis", 
                       {"investigation_id": investigation["id"]})
await client.call_tool("sift_run_slow_request_analysis", 
                       {"investigation_id": investigation["id"]})
```

### 2. Proactive Monitoring

Periodic health checks:

```python
# Create investigation for last 30 minutes
investigation = await client.call_tool(
    "sift_create_investigation",
    {
        "name": "Health Check - API",
        "labels": {
            "cluster": "production",
            "namespace": "api"
        }
    }
)

# Analyze
results = await run_all_analyses(investigation["id"])

# Alert if issues found
if has_critical_issues(results):
    send_alert("Critical issues detected", results)
```

### 3. Performance Baseline

Regular performance checks:

```python
# Weekly performance review
investigation = await client.call_tool(
    "sift_create_investigation",
    {
        "name": f"Weekly Performance Review - {date}",
        "labels": {
            "cluster": "production",
            "service": "user-service"
        }
    }
)

# Focus on slow requests
await client.call_tool("sift_run_slow_request_analysis",
                       {"investigation_id": investigation["id"]})
```

## List Past Investigations

```python
# List recent investigations
result = await client.call_tool(
    "sift_list_investigations",
    {
        "limit": 10
    }
)

investigations = result["investigations"]
for inv in investigations:
    print(f"{inv['created_at']}: {inv['name']} - {inv['status']}")
```

## Tips & Tricks

### 1. Label Strategy
Use specific labels for better results:
```python
# ✅ Good - specific labels
labels = {
    "cluster": "production",
    "namespace": "api",
    "service": "user-service"
}

# ❌ Bad - too broad
labels = {"cluster": "production"}
```

### 2. Time Windows
Match your investigation window to the incident:
```python
# Short-term spike
start_time = now - timedelta(minutes=15)

# Gradual degradation
start_time = now - timedelta(hours=2)

# Post-deployment issues
start_time = deployment_time
```

### 3. Custom Queries
Override default queries for specific cases:
```python
# Custom LogQL query
await client.call_tool(
    "sift_run_error_pattern_analysis",
    {
        "investigation_id": inv_id,
        "log_query": '{namespace="api"} |= "database" |= "error"'
    }
)
```

### 4. Baseline Periods
Adjust baseline for different patterns:
```python
# For weekly patterns, query 7 days back as baseline
# (Requires analyzer configuration adjustment)
```

## Troubleshooting

### No Data Found

**Problem**: Analysis completes but finds nothing

**Solutions**:
1. Verify logs/traces exist for the time range:
   ```bash
   # Check Loki
   logcli query '{namespace="api"}' --since=30m --limit=10
   
   # Check Tempo
   tempo-cli search --service=api --since=30m
   ```

2. Check label selectors match your data
3. Verify time window covers incident period

### Analysis Fails

**Problem**: Investigation status is "failed"

**Solutions**:
1. Check MCP server logs:
   ```bash
   kubectl logs -l app=agent-sre-mcp-server -n agent-sre --tail=100
   ```

2. Verify connectivity:
   ```bash
   # Test Loki
   kubectl run curl --rm -it --image=curlimages/curl -- \
     curl http://loki-gateway.loki.svc.cluster.local:80/ready
   
   # Test Tempo
   kubectl run curl --rm -it --image=curlimages/curl -- \
     curl http://tempo.tempo.svc.cluster.local:3100/ready
   ```

### Storage Issues

**Problem**: Cannot save investigations

**Solutions**:
1. Check storage path permissions
2. Verify disk space available
3. Check SQLite database health

## Next Steps

- 📖 Read [SIFT_GUIDE.md](SIFT_GUIDE.md) for detailed documentation
- 🧪 Run tests: `pytest tests/test_sift.py`
- 🔧 Customize thresholds and baseline windows
- 📊 Integrate with your alerting system
- 🤖 Use with LLM agents via MCP protocol

## Getting Help

- Check logs: `kubectl logs -l app=agent-sre-mcp-server -n agent-sre`
- Review documentation: `SIFT_GUIDE.md`
- Run tests: `pytest tests/test_sift.py -v`
- Check implementation: `SIFT_IMPLEMENTATION.md`

## Example: Complete Investigation Flow

Here's a complete example from start to finish:

```python
import asyncio
from datetime import datetime, timedelta
from mcp_client import MCPClient

async def investigate_incident():
    client = MCPClient("http://agent-sre-mcp-server.agent-sre:3000")
    
    # 1. Create investigation
    print("🔍 Creating investigation...")
    inv_result = await client.call_tool(
        "sift_create_investigation",
        {
            "name": "Production API Latency Spike",
            "labels": {
                "cluster": "production",
                "namespace": "api",
                "service": "user-service"
            },
            "start_time": (datetime.utcnow() - timedelta(minutes=30)).isoformat(),
            "end_time": datetime.utcnow().isoformat()
        }
    )
    
    inv_id = inv_result["investigation"]["id"]
    print(f"✅ Investigation created: {inv_id}")
    
    # 2. Run error pattern analysis
    print("\n🔬 Analyzing error patterns...")
    error_result = await client.call_tool(
        "sift_run_error_pattern_analysis",
        {"investigation_id": inv_id}
    )
    
    patterns = error_result["analysis"]["result"]["elevated_patterns"]
    print(f"✅ Found {len(patterns)} elevated error patterns")
    
    for pattern in patterns[:3]:  # Top 3
        print(f"\n  🔴 {pattern['severity'].upper()}")
        print(f"     {pattern['pattern']}")
        print(f"     Count: {pattern['current_count']} "
              f"(was {pattern['baseline_count']})")
        print(f"     Elevation: {pattern['elevation_factor']}x")
    
    # 3. Run slow request analysis
    print("\n⏱️ Analyzing request latencies...")
    slow_result = await client.call_tool(
        "sift_run_slow_request_analysis",
        {"investigation_id": inv_id}
    )
    
    slow_ops = slow_result["analysis"]["result"]["slow_operations"]
    print(f"✅ Found {len(slow_ops)} slow operations")
    
    for op in slow_ops[:3]:  # Top 3
        print(f"\n  ⏱️ {op['severity'].upper()}")
        print(f"     {op['operation']}")
        print(f"     Current P95: {op['current_p95_ms']}ms "
              f"(was {op['baseline_p95_ms']}ms)")
        print(f"     Slowdown: {op['slowdown_factor']}x")
    
    # 4. Get complete investigation
    print("\n📋 Fetching complete investigation...")
    final_result = await client.call_tool(
        "sift_get_investigation",
        {"investigation_id": inv_id}
    )
    
    investigation = final_result["investigation"]
    print(f"\n✅ Investigation Complete!")
    print(f"   Name: {investigation['name']}")
    print(f"   Status: {investigation['status']}")
    print(f"   Analyses: {len(investigation['analyses'])}")
    print(f"   ID: {investigation['id']}")
    
    return investigation

# Run the investigation
if __name__ == "__main__":
    asyncio.run(investigate_incident())
```

Happy investigating! 🔍

