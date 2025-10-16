# 🔍 Grafana Sift Guide

## Overview

Grafana Sift is an AI-powered investigation platform integrated into Agent-SRE that provides automated analysis of logs, traces, and metrics to identify issues and anomalies in your infrastructure.

## Features

### 🔬 Error Pattern Detection
Automatically identifies elevated error patterns in logs by:
- Comparing current period against a 24-hour baseline
- Normalizing log patterns to group similar errors
- Calculating elevation factors (how much errors have increased)
- Assigning severity levels (critical, high, medium, low)
- Filtering out noise and focusing on significant changes

### ⏱️ Slow Request Detection
Identifies performance degradations in distributed traces by:
- Analyzing request latencies across services
- Comparing current P95 latencies against baseline
- Detecting slowdown factors (how much slower requests are)
- Grouping by service and operation
- Highlighting critical performance regressions

### 💾 Investigation Management
- Create scoped investigations with labels (cluster, namespace, etc.)
- Store investigation history in SQLite
- Track multiple analyses per investigation
- Query past investigations for trends
- Export results for reporting

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Agent-SRE MCP                        │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │ Prometheus   │  │  Grafana     │  │    Sift      │    │
│  │    Tools     │  │   Tools      │  │   Platform   │    │
│  └──────────────┘  └──────────────┘  └──────┬───────┘    │
│                                               │             │
└───────────────────────────────────────────────┼─────────────┘
                                                │
                    ┌───────────────────────────┴───────────────────────┐
                    │                                                   │
         ┌──────────▼─────────┐                           ┌───────────▼──────────┐
         │   Loki Client      │                           │   Tempo Client       │
         │                    │                           │                      │
         │ - Query logs       │                           │ - Search traces      │
         │ - Label filtering  │                           │ - Tag filtering      │
         │ - Stats retrieval  │                           │ - Latency analysis   │
         └────────┬───────────┘                           └──────────┬───────────┘
                  │                                                  │
         ┌────────▼───────────────────────────────────────────────▼─────────┐
         │                    Investigation Storage                         │
         │                      (SQLite Database)                           │
         │                                                                  │
         │  - Investigations                                                │
         │  - Analyses                                                      │
         │  - Results                                                       │
         └──────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Investigation Models (`investigation.py`)
Defines data structures for:
- **Investigation**: Container for analyses with metadata
- **Analysis**: Individual analysis result (error pattern, slow request)
- **InvestigationStatus**: Status tracking (pending, running, completed, failed)
- **AnalysisType**: Type of analysis (error_pattern, slow_request, etc.)

### 2. Storage Layer (`storage.py`)
SQLite-based persistence for:
- Saving and retrieving investigations
- Listing investigations with filtering
- Deleting investigations
- Transaction management

### 3. Query Clients
**Loki Client** (`loki_client.py`):
- Query logs over time ranges
- Filter by labels
- Get log statistics
- List available labels and values

**Tempo Client** (`tempo_client.py`):
- Search traces by tags
- Filter by duration
- Get trace details
- List available tags

### 4. Analysis Algorithms (`analyzers.py`)

**Error Pattern Analyzer**:
```python
class ErrorPatternAnalyzer:
    def analyze(current_logs, baseline_logs, threshold_multiplier=2.0):
        # 1. Extract and normalize log patterns
        # 2. Compare current vs baseline frequencies
        # 3. Calculate elevation factors
        # 4. Determine severity levels
        # 5. Return top patterns
```

**Slow Request Analyzer**:
```python
class SlowRequestAnalyzer:
    def analyze(current_traces, baseline_traces, percentile=95.0):
        # 1. Group traces by operation
        # 2. Calculate P95 latencies
        # 3. Compare current vs baseline
        # 4. Identify slowdowns
        # 5. Return slow operations
```

### 5. Sift Core (`sift_core.py`)
Orchestration layer that:
- Creates investigations
- Runs analyses (error pattern, slow request)
- Manages workflow and state
- Handles errors and retries
- Stores results

## Usage Examples

### Example 1: Investigate API Errors

```python
# 1. Create investigation
investigation = await sift_core.create_investigation(
    name="API 500 Errors Investigation",
    labels={
        "cluster": "production",
        "namespace": "api",
        "app": "backend"
    },
    start_time=datetime.utcnow() - timedelta(minutes=30),
    end_time=datetime.utcnow()
)

# 2. Run error pattern analysis
analysis = await sift_core.run_error_pattern_analysis(
    investigation_id=investigation.id
)

# 3. Review results
if analysis.status == InvestigationStatus.COMPLETED:
    for pattern in analysis.result["elevated_patterns"]:
        print(f"Pattern: {pattern['pattern']}")
        print(f"Severity: {pattern['severity']}")
        print(f"Elevation: {pattern['elevation_factor']}x")
        print(f"Count: {pattern['current_count']}")
```

### Example 2: Investigate Slow Requests

```python
# 1. Create investigation
investigation = await sift_core.create_investigation(
    name="Slow API Requests",
    labels={
        "cluster": "production",
        "service": "user-service"
    }
)

# 2. Run slow request analysis
analysis = await sift_core.run_slow_request_analysis(
    investigation_id=investigation.id
)

# 3. Review results
for operation in analysis.result["slow_operations"]:
    print(f"Operation: {operation['operation']}")
    print(f"Current P95: {operation['current_p95_ms']}ms")
    print(f"Baseline P95: {operation['baseline_p95_ms']}ms")
    print(f"Slowdown: {operation['slowdown_factor']}x")
```

### Example 3: List Past Investigations

```python
# List recent investigations
investigations = await sift_core.list_investigations(limit=10)

for inv in investigations:
    print(f"ID: {inv.id}")
    print(f"Name: {inv.name}")
    print(f"Status: {inv.status}")
    print(f"Analyses: {len(inv.analyses)}")
    print(f"Created: {inv.created_at}")
```

## MCP Integration

Sift is exposed via MCP tools that can be called by LLMs and agents:

```json
{
  "tool": "sift_create_investigation",
  "arguments": {
    "name": "Production API Issues",
    "labels": {
      "cluster": "prod",
      "namespace": "api"
    }
  }
}
```

## Configuration

### Environment Variables

```bash
# Loki configuration
LOKI_URL=http://loki-gateway.loki.svc.cluster.local:80

# Tempo configuration
TEMPO_URL=http://tempo.tempo.svc.cluster.local:3100

# Sift storage
SIFT_STORAGE_PATH=/var/lib/sift/investigations.db
```

### Baseline Window
Default: 24 hours

Can be customized when creating analyzers:
```python
error_analyzer = ErrorPatternAnalyzer(baseline_window=48)  # 48 hours
slow_request_analyzer = SlowRequestAnalyzer(baseline_window=12)  # 12 hours
```

### Thresholds

**Error Pattern Detection:**
- Threshold multiplier: 2.0 (2x increase)
- Minimum count: 5 errors
- Severity levels:
  - Critical: 10x elevation or 100+ occurrences
  - High: 5x elevation or 50+ occurrences
  - Medium: 3x elevation or 20+ occurrences
  - Low: 2x elevation

**Slow Request Detection:**
- Slowdown threshold: 1.5x (50% slower)
- Percentile: P95
- Severity levels:
  - Critical: 5x slowdown or 5000ms+
  - High: 3x slowdown or 2000ms+
  - Medium: 2x slowdown or 1000ms+
  - Low: 1.5x slowdown

## Best Practices

### 1. Label Strategy
Use consistent labels across your stack:
```python
labels = {
    "cluster": "production",      # Required
    "namespace": "api",            # Required
    "service": "user-service",     # Recommended
    "environment": "prod",         # Recommended
    "team": "platform"            # Optional
}
```

### 2. Time Windows
- **Short-term issues**: 15-30 minutes
- **Performance trends**: 1-4 hours
- **Incident investigations**: Match incident duration

### 3. Baseline Periods
- **Normal operations**: 24-hour baseline
- **Weekly patterns**: 7-day baseline
- **Seasonal patterns**: 30-day baseline

### 4. Analysis Workflow
1. Create investigation with focused labels
2. Run error pattern analysis first
3. If errors found, investigate related services
4. Run slow request analysis for performance issues
5. Correlate results across analyses

### 5. Storage Management
- Regularly clean old investigations
- Export critical investigations for long-term storage
- Monitor database size
- Consider archiving after 30 days

## Troubleshooting

### Investigation Fails
**Symptom**: Investigation status is "failed"

**Solutions**:
1. Check Loki/Tempo connectivity
2. Verify labels exist in the data sources
3. Check time range has data
4. Review error logs for details

### No Patterns Found
**Symptom**: Analysis completes but finds no elevated patterns

**Solutions**:
1. Verify logs contain error indicators
2. Check if baseline period has sufficient data
3. Adjust threshold multiplier if needed
4. Review label selectors

### Storage Errors
**Symptom**: Cannot save/retrieve investigations

**Solutions**:
1. Check file permissions on database path
2. Verify disk space available
3. Check for database corruption
4. Try creating new database

## Performance Considerations

### Log Volume
- Limit queries to 1000 log lines
- Use specific label selectors
- Narrow time windows for large volumes

### Trace Volume
- Limit trace searches to 100 traces
- Filter by service/operation
- Use min/max duration filters

### Storage
- SQLite handles millions of investigations
- Index on investigation_id and created_at
- Regular VACUUM for optimization

## Future Enhancements

- [ ] Metric anomaly detection (Prometheus)
- [ ] Multi-dimensional analysis
- [ ] Machine learning-based baselines
- [ ] Custom alert integration
- [ ] Dashboard generation
- [ ] Automatic remediation suggestions
- [ ] Investigation templates
- [ ] Collaborative investigations
- [ ] Real-time streaming analysis
- [ ] Investigation playbooks

## Contributing

To add new analysis types:

1. Create analyzer in `analyzers.py`
2. Add analysis type to `AnalysisType` enum
3. Add execution method in `sift_core.py`
4. Add MCP tool definition
5. Add tests
6. Update documentation

## Support

For issues and questions:
- Check logs: `kubectl logs -l app=agent-sre-mcp-server -n agent-sre`
- Review test cases in `tests/test_sift.py`
- Consult README.md for MCP tool documentation

