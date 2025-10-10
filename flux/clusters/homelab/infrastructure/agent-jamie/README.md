# 🤖 Jamie - Your SRE Assistant

Jamie is a sophisticated SRE assistant that combines AI-powered intelligence with real-time infrastructure monitoring and troubleshooting capabilities. It consists of two main components:

1. **Jamie Slack Bot** - Interactive Slack assistant with LLM brain
2. **Jamie MCP Server** - Model Context Protocol server for IDE integration (Cursor)

## 🏗️ Architecture

Jamie follows a modular architecture similar to agent-sre:

```
jamie/
├── deployments/
│   ├── slack-bot/              # Slack Bot deployment
│   │   ├── core.py             # Shared core with logfire
│   │   ├── jamie_slack_bot.py  # Slack bot implementation
│   │   ├── Dockerfile          # Bot container image
│   │   ├── k8s-slack-bot.yaml  # K8s deployment manifest
│   │   └── kustomization.yaml  # Kustomize configuration
│   └── mcp-server/             # MCP Server deployment
│       ├── core.py             # Shared core with logfire
│       ├── mcp_server.py       # MCP server implementation
│       ├── Dockerfile          # MCP container image
│       ├── k8s-mcp-server.yaml # K8s deployment manifest
│       └── kustomization.yaml  # Kustomize configuration
├── k8s/                        # Shared K8s resources
│   ├── namespace.yaml
│   ├── serviceaccount.yaml
│   └── secret.yaml
├── pyproject.toml              # Python project configuration
├── requirements.txt            # Python dependencies
├── Makefile                    # Build and deployment automation
├── kustomization.yaml          # Root kustomization
└── README.md                   # This file
```

## ✨ Features

### Jamie Slack Bot
- 🧠 **AI-Powered** - Uses Ollama for intelligent responses
- 🔧 **Tool Integration** - Connects to agent-sre for infrastructure operations
- 📊 **Golden Signals** - Monitor latency, traffic, errors, saturation
- ☸️ **Kubernetes** - Pod logs, deployments, resource management
- 📈 **Prometheus** - Custom PromQL queries and metrics
- 🔍 **Log Analysis** - AI-powered log analysis and insights
- 💬 **Conversational** - Natural language interaction via Slack

### Jamie MCP Server
- 🎯 **Cursor Integration** - Use Jamie directly in your IDE via MCP protocol
- 🌐 **Homepage Integration** - REST API for Homepage chatbot
- 🔌 **Dual Interface** - Supports both REST API and MCP protocol
- 🛠️ **Tool Exposure** - All SRE tools accessible via REST/MCP
- 🤖 **AI Assistance** - Ask Jamie questions about SRE practices
- 📡 **Real-time** - Direct connection to infrastructure

## 🔧 Logfire Integration

Both components integrate with [Logfire](https://logfire.pydantic.dev/) for observability:

- **Request Tracing** - Track all MCP requests and Slack interactions
- **Performance Monitoring** - Monitor response times and errors
- **AI Instrumentation** - Track LLM calls and agent decisions
- **Infrastructure Insights** - Correlate with infrastructure metrics

### Configuration

Set these environment variables for logfire:

```bash
# For Slack Bot
export LOGFIRE_TOKEN_JAMIE="your-jamie-token"

# For MCP Server
export LOGFIRE_TOKEN_JAMIE_MCP="your-jamie-mcp-token"
```

If not set, Jamie will continue to work without logfire (graceful degradation).

## 🚀 Quick Start

### 1. Build and Deploy

```bash
# Build both images
make build

# Push to registry (requires GITHUB_TOKEN)
make push

# Deploy to Kubernetes
make deploy

# Or do everything at once
make all
```

### 2. Check Status

```bash
# Check both components
make status

# Check individual components
make status-bot
make status-mcp
```

### 3. View Logs

```bash
# View Slack Bot logs
make logs-bot

# View MCP Server logs
make logs-mcp
```

## 📱 Using Jamie in Slack

### Setup

1. Create a Slack App with these permissions:
   - `app_mentions:read`
   - `chat:write`
   - `im:history`
   - `im:write`

2. Create secrets in Kubernetes:
   ```bash
   kubectl create secret generic jamie-secrets \
     --from-literal=SLACK_BOT_TOKEN="xoxb-..." \
     --from-literal=SLACK_APP_TOKEN="xapp-..." \
     --from-literal=SLACK_SIGNING_SECRET="..." \
     -n jamie
   ```

3. Deploy Jamie:
   ```bash
   make push-bot deploy
   ```

### Usage

```
# Ask Jamie anything
@Jamie how do I check service health?

# Check golden signals
@Jamie check golden signals for homepage

# Query Prometheus
@Jamie query prometheus: up{job="homepage"}

# Get pod logs
@Jamie show me logs from pod homepage-xyz

# Analyze logs
@Jamie analyze these logs: [paste logs]

# Get SRE advice
@Jamie what are best practices for alerting?
```

## 🌐 Using Jamie from Homepage

Jamie MCP Server now exposes a REST API specifically for Homepage integration!

### REST API Endpoints

```bash
# Main chatbot endpoint
POST /api/chat
{
  "message": "Your question here"
}

# Check golden signals
POST /api/golden-signals
{
  "service_name": "homepage",
  "namespace": "default"
}

# Execute PromQL query
POST /api/prometheus/query
{
  "query": "up{job=\"homepage\"}"
}

# Get pod logs
POST /api/pod-logs
{
  "pod_name": "homepage-xyz",
  "namespace": "default",
  "tail_lines": 100
}

# Analyze logs
POST /api/analyze-logs
{
  "logs": "your logs here",
  "context": "optional context"
}
```

### Example from Homepage API (Go)

```go
// Homepage Go API automatically proxies to Jamie via:
// POST /api/v1/jamie/chat
// POST /api/v1/jamie/golden-signals
// POST /api/v1/jamie/prometheus/query
// POST /api/v1/jamie/pod-logs
// POST /api/v1/jamie/analyze-logs
```

The Homepage API handler automatically forwards requests to Jamie MCP Server at:
`http://jamie-mcp-server-service.jamie.svc.cluster.local:30121`

## 💻 Using Jamie in Cursor

### Setup

1. Build and deploy MCP server:
   ```bash
   make setup-mcp
   ```

2. Configure Cursor (`~/.cursor/mcp.json`):
   ```json
   {
     "mcpServers": {
       "jamie": {
         "url": "http://192.168.0.16:30121/mcp",
         "name": "Jamie - SRE Assistant"
       }
     }
   }
   ```

3. Restart Cursor

### Usage

```
@jamie ask how do I check if my homepage service is healthy?
@jamie check golden signals for homepage in default namespace
@jamie query prometheus: up{job="homepage"}
@jamie get logs from pod homepage-xyz in default namespace
@jamie analyze these logs: [paste logs here]
```

See [MCP_README.md](MCP_README.md) for detailed MCP setup instructions.

## 🛠️ Development

### Local Development

```bash
# Run Slack Bot locally
make dev-bot

# Run MCP Server locally
make dev-mcp

# Run both with docker-compose
make dev

# Stop local development
make dev-down
```

### Testing

```bash
# Run tests
make test

# Test MCP endpoints
make test-mcp
```

### Code Structure

Each deployment has:
- `core.py` - Shared configuration and logfire setup
- Main implementation file (bot or server)
- `Dockerfile` - Container image definition
- `k8s-*.yaml` - Kubernetes manifests
- `kustomization.yaml` - Kustomize configuration

## 📊 Configuration

### Environment Variables

#### Slack Bot
- `SLACK_BOT_TOKEN` - Slack bot token (from secret)
- `SLACK_APP_TOKEN` - Slack app token (from secret)
- `SLACK_SIGNING_SECRET` - Slack signing secret (from secret)
- `AGENT_SRE_URL` - Agent-SRE Service URL (default: http://sre-agent-service.agent-sre:8080)
- `OLLAMA_URL` - Ollama server URL (default: http://192.168.0.16:11434)
- `MODEL_NAME` - Ollama model name (default: llama3.2:3b)
- `LOGFIRE_TOKEN_JAMIE` - Logfire token for observability (optional)

#### MCP Server
- `AGENT_SRE_URL` - Agent-SRE MCP URL
- `OLLAMA_URL` - Ollama server URL
- `MODEL_NAME` - Ollama model name
- `MCP_HOST` - Server host (default: 0.0.0.0)
- `MCP_PORT` - Server port (default: 30121)
- `LOGFIRE_TOKEN_JAMIE_MCP` - Logfire token for observability (optional)

## 🔍 Troubleshooting

### Slack Bot Issues

```bash
# Check logs
make logs-bot

# Check status
make status-bot

# Restart
make restart-bot

# Common issues:
# - Check if secrets exist: kubectl get secret -n jamie
# - Verify Ollama is accessible: curl http://192.168.0.16:11434/api/tags
# - Verify agent-sre is running: kubectl get pods -n agent-sre
```

### MCP Server Issues

```bash
# Check logs
make logs-mcp

# Check status
make status-mcp

# Test endpoints
make test-mcp

# Common issues:
# - Check if service is exposed: kubectl get svc -n jamie
# - Test health: curl http://192.168.0.16:30121/health
# - Verify agent-sre connectivity: kubectl get svc -n agent-sre
```

## 🎯 Available Tools

Both Jamie components have access to these tools via agent-sre:

- **check_golden_signals** - Monitor service health metrics
- **query_prometheus** - Execute PromQL queries
- **get_pod_logs** - Retrieve Kubernetes pod logs
- **analyze_logs** - AI-powered log analysis
- **sre_chat** - General SRE consultation
- **health_check** - Check agent status

## 🏥 Health Checks

### Slack Bot
- **Liveness**: Python execution check
- **Readiness**: Agent-SRE connectivity check

### MCP Server
- **Liveness**: HTTP `/health` endpoint
- **Readiness**: HTTP `/ready` endpoint + Agent-SRE check

## 📚 Additional Resources

- [MCP_README.md](MCP_README.md) - Detailed MCP setup and usage
- [agent-sre](../agent-sre) - Backend SRE agent
- [Logfire Documentation](https://logfire.pydantic.dev/)
- [LangChain Documentation](https://python.langchain.com/)
- [Slack Bolt Framework](https://slack.dev/bolt-python/)

## 🤝 Contributing

Jamie follows the same patterns as agent-sre:
1. Modular deployment structure
2. Shared core modules
3. Logfire instrumentation
4. Kustomize for K8s management
5. Makefile for automation

When adding features:
1. Add logfire instrumentation with `@logfire.instrument()`
2. Update both deployments if needed
3. Update tests
4. Update documentation

## 📝 License

Same as homelab repository.

---

Happy SRE-ing with Jamie! 🤖✨
