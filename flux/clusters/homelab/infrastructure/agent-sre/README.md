# 🤖 Agent-SRE

Agent-SRE is an intelligent SRE (Site Reliability Engineering) agent that provides automated monitoring, incident response, and observability capabilities.

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
  - HTTP API wrapper for Kubernetes deployment
- **Technology**: Python, MCP protocol, aiohttp

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

## Monitoring

The system integrates with:
- **Logfire**: For observability and tracing
- **LangSmith**: For LLM call tracking
- **Prometheus**: For metrics collection
- **Grafana**: For visualization

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]
