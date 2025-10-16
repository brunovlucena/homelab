# 🤖 Agent-SRE

Agent-SRE is an intelligent SRE (Site Reliability Engineering) agent that provides automated monitoring, incident response, observability, and AI-powered investigation capabilities via Grafana Sift.

## Architecture

The Agent-SRE system consists of two main components:

### 1. 🧠 Agent-SRE Core (Port 8080)
- **Purpose**: Main SRE agent with LangGraph state management
- **Features**:
  - Chat interface for SRE tasks
  - Incident response automation
  - Kubernetes query handling
  - Alert webhook processing
- **Technology**: Python, LangGraph, LangChain, Ollama LLM

### 2. 🔌 Agent-SRE MCP Server (Port 3000)
- **Purpose**: Expose observability tools via Model Context Protocol (MCP)
- **Features**:
  - Prometheus query execution (PromQL)
  - Grafana dashboard and datasource queries
  - Time series data retrieval
  - Grafana Sift investigation platform
  - Automated error pattern detection
  - Slow request analysis
  - HTTP API wrapper for Kubernetes deployment
- **Technology**: Python, MCP protocol, aiohttp

### 3. 🔍 Grafana Sift (Integrated)
- **Purpose**: AI-powered investigation platform for automated analysis
- **Features**:
  - Automated investigation management
  - Error pattern detection via Loki logs
  - Slow request detection via Tempo traces
  - Baseline comparison and anomaly detection
  - Investigation storage and history
- **Technology**: Python, SQLite, Loki, Tempo

## MCP Tools

The MCP server exposes the following tools:

### `prometheus_query`
Execute PromQL queries against Prometheus.

**Parameters:**
- `query` (required): PromQL expression
- `time` (optional): Timestamp for query evaluation
- `timeout` (optional): Query timeout

**Example:**
```json
{
  "tool": "prometheus_query",
  "arguments": {
    "query": "rate(http_requests_total[5m])"
  }
}
```

### `prometheus_query_range`
Query Prometheus time series data over a range.

**Parameters:**
- `query` (required): PromQL expression
- `start` (required): Start timestamp
- `end` (required): End timestamp
- `step` (required): Query resolution (e.g., "15s", "1m")
- `timeout` (optional): Query timeout

### `grafana_query`
Query Grafana dashboards or datasources.

**Parameters:**
- `query_type` (required): "dashboard", "datasource", "search", or "panel"
- `query` (required): Query string or identifier
- `dashboard_id` (optional): Dashboard UID for panel queries
- `panel_id` (optional): Panel ID
- `from_time` (optional): Start time
- `to_time` (optional): End time

### `sift_create_investigation`
Create a new Sift investigation for automated analysis.

**Parameters:**
- `name` (required): Investigation name
- `labels` (required): Labels to scope the investigation (e.g., {"cluster": "prod", "namespace": "api"})
- `start_time` (optional): Start time (ISO 8601, defaults to 30 minutes ago)
- `end_time` (optional): End time (ISO 8601, defaults to now)

**Example:**
```json
{
  "tool": "sift_create_investigation",
  "arguments": {
    "name": "API Performance Investigation",
    "labels": {
      "cluster": "production",
      "namespace": "api"
    }
  }
}
```

### `sift_run_error_pattern_analysis`
Run error pattern detection on an investigation.

**Parameters:**
- `investigation_id` (required): Investigation ID
- `log_query` (optional): LogQL query (will be built from labels if not provided)

**Example:**
```json
{
  "tool": "sift_run_error_pattern_analysis",
  "arguments": {
    "investigation_id": "abc-123-xyz"
  }
}
```

### `sift_run_slow_request_analysis`
Run slow request detection on an investigation.

**Parameters:**
- `investigation_id` (required): Investigation ID
- `trace_tags` (optional): Trace tags (will be built from labels if not provided)

**Example:**
```json
{
  "tool": "sift_run_slow_request_analysis",
  "arguments": {
    "investigation_id": "abc-123-xyz"
  }
}
```

### `sift_get_investigation`
Get details of a specific investigation.

**Parameters:**
- `investigation_id` (required): Investigation ID

### `sift_list_investigations`
List recent investigations.

**Parameters:**
- `limit` (optional): Maximum number of investigations to return (default: 10)

## Deployment

### Prerequisites
- Kubernetes cluster
- Prometheus and Grafana deployed
- Docker registry access (ghcr.io)

### Build and Deploy

#### Build MCP Server
```bash
make build-mcp-server
```

#### Push to Registry
```bash
make push-mcp-server
```

#### Deploy to Kubernetes
```bash
make deploy-mcp-server
```

#### View Logs
```bash
make logs-mcp-server
```

### Configuration

#### Environment Variables

**Agent-SRE Core:**
- `OLLAMA_URL`: Ollama LLM server URL
- `MODEL_NAME`: LLM model name
- `AGENT_HOST`: Server bind host (default: 0.0.0.0)
- `AGENT_PORT`: Server port (default: 8080)
- `PROMETHEUS_URL`: Prometheus API URL
- `MCP_SERVER_URL`: MCP server URL

**MCP Server:**
- `PROMETHEUS_URL`: Prometheus API URL (default: http://prometheus-k8s.prometheus.svc.cluster.local:9090)
- `GRAFANA_URL`: Grafana API URL (default: http://grafana.grafana.svc.cluster.local:3000)
- `GRAFANA_API_KEY`: Grafana API key (optional)
- `LOKI_URL`: Loki API URL (default: http://loki-gateway.loki.svc.cluster.local:80)
- `TEMPO_URL`: Tempo API URL (default: http://tempo.tempo.svc.cluster.local:3100)
- `SIFT_STORAGE_PATH`: Path for Sift investigation database (default: /tmp/sift_investigations.db)
- `HTTP_HOST`: Server bind host (default: 0.0.0.0)
- `HTTP_PORT`: Server port (default: 3000)

## Integration with Jamie

Jamie (the Slack bot) uses the Agent-SRE MCP server to query observability data. The integration works as follows:

1. Jamie receives a message in Slack requesting Prometheus or Grafana data
2. Jamie calls the Agent-SRE MCP server via HTTP using the `AgentSREClient`
3. MCP server executes the query against Prometheus/Grafana
4. Results are returned to Jamie and formatted for Slack

**Example Integration:**
```python
async with AgentSREClient() as agent:
    result = await agent.prometheus_query("up")
    # Process and display result in Slack
```

## API Endpoints

### Agent-SRE Core (Port 8080)
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /status` - Agent status
- `POST /chat` - Chat with agent
- `POST /mcp/chat` - Chat via MCP
- `POST /prometheus/query` - Prometheus query (deprecated, use MCP)
- `POST /grafana/query` - Grafana query (deprecated, use MCP)
- `POST /k8s/query` - Kubernetes query
- `POST /webhook/alert` - Alertmanager webhook

### MCP Server (Port 3000)
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /tools` - List available MCP tools
- `POST /mcp/tool` - Execute MCP tool
- `POST /tools/prometheus_query` - Direct Prometheus query
- `POST /tools/prometheus_query_range` - Direct range query
- `POST /tools/grafana_query` - Direct Grafana query

## Development

### Install Dependencies
```bash
uv pip install -e .
```

### Run Tests
```bash
pytest tests/
```

### Run Locally

**Agent-SRE Core:**
```bash
cd deployments/agent
python agent.py
```

**MCP Server:**
```bash
cd deployments/mcp-server
python mcp_http_wrapper.py
```

## Grafana Sift Usage

### Quick Start

1. **Create an Investigation**
   ```python
   # Via MCP tool
   investigation = await sift_create_investigation(
       name="API Performance Issues",
       labels={"cluster": "prod", "namespace": "api"}
   )
   ```

2. **Run Error Pattern Analysis**
   ```python
   # Analyzes logs for elevated error patterns
   analysis = await sift_run_error_pattern_analysis(
       investigation_id=investigation["id"]
   )
   ```

3. **Run Slow Request Analysis**
   ```python
   # Analyzes traces for slow requests
   analysis = await sift_run_slow_request_analysis(
       investigation_id=investigation["id"]
   )
   ```

4. **Review Results**
   ```python
   # Get full investigation with all analyses
   results = await sift_get_investigation(
       investigation_id=investigation["id"]
   )
   ```

### Investigation Workflow

```
1. Create Investigation → 2. Run Analyses → 3. Review Results → 4. Take Action
                             ↓
                     Error Pattern Analysis
                     Slow Request Analysis
```

### What Sift Analyzes

**Error Pattern Detection:**
- Compares current period logs against 24-hour baseline
- Identifies elevated error patterns (2x or more increase)
- Normalizes log patterns to group similar errors
- Reports severity: critical, high, medium, low
- Shows elevation factors and occurrence counts

**Slow Request Detection:**
- Compares current period traces against 24-hour baseline
- Identifies operations with degraded performance (1.5x or slower)
- Calculates P95 latencies for comparison
- Reports slowdown factors and request counts
- Groups by service/operation

### Example Results

**Error Pattern Analysis:**
```json
{
  "elevated_patterns": [
    {
      "pattern": "ERROR: Database connection timeout",
      "current_count": 45,
      "baseline_count": 5,
      "elevation_factor": 9.0,
      "severity": "critical"
    }
  ]
}
```

**Slow Request Analysis:**
```json
{
  "slow_operations": [
    {
      "operation": "GET /api/users",
      "current_p95_ms": 2500,
      "baseline_p95_ms": 500,
      "slowdown_factor": 5.0,
      "severity": "critical"
    }
  ]
}
```

## Monitoring

The system integrates with:
- **Logfire**: For observability and tracing
- **LangSmith**: For LLM call tracking
- **Prometheus**: For metrics collection
- **Grafana**: For visualization
- **Loki**: For log aggregation (Sift)
- **Tempo**: For distributed tracing (Sift)

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]
