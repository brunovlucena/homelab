# 🤖 Jamie Slack Bot

**Your SRE Companion on Slack** - Powered by Agent-SRE and MCP (Model Context Protocol)

Jamie is an intelligent Slack bot that helps you monitor, troubleshoot, and manage your infrastructure directly from Slack. It communicates with the Agent-SRE service via MCP to provide real-time insights and actions.

## ✨ Features

- 🎯 **Golden Signals Monitoring** - Track latency, traffic, errors, and saturation
- ☸️ **Kubernetes Operations** - Manage pods, deployments, and view logs
- 📈 **Grafana Integration** - Access dashboards, incidents, and alerts
- 🔍 **Log Analysis** - Intelligent log parsing and error detection
- 🚨 **Incident Management** - Create and track incidents
- 💬 **Conversational AI** - Natural language interactions via MCP
- 🧠 **Context Awareness** - Remembers conversation history

## 🏗️ Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────────────┐
│             │      │             │      │   Agent-SRE MCP     │
│   Slack     │─────►│   Jamie     │─────►│   Server            │
│   Users     │      │   Bot       │ MCP  │   (port 30120)      │
│             │      │             │      │                     │
└─────────────┘      └─────────────┘      └─────────────────────┘
                            │                        │
                            │                        │
                            ▼                        ▼
                     ┌─────────────┐         ┌─────────────┐
                     │ Conversation│         │ MCP Tools   │
                     │ Context     │         │ • Grafana   │
                     │ Storage     │         │ • K8s       │
                     └─────────────┘         │ • Prometheus│
                                             │ • Custom    │
                                             └─────────────┘
```

**Important**: Jamie communicates with the SRE Agent **ONLY via MCP** at:
- Service: `sre-agent-mcp-server-service.agent-sre`
- Port: `30120`
- Protocol: HTTP/MCP

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster
- Slack workspace with admin access
- Agent-SRE service running
- ECR access for Docker images

### 1. Create Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App" → "From a manifest"
3. Paste the manifest from `slack-app-manifest.yaml`
4. Install the app to your workspace

### 2. Get Slack Tokens

From your Slack app settings:

- **Bot Token**: OAuth & Permissions → Bot User OAuth Token (starts with `xoxb-`)
- **App Token**: Basic Information → App-Level Tokens (starts with `xapp-`)
- **Signing Secret**: Basic Information → Signing Secret

### 3. Configure Secrets

Create a secret file with your tokens:

```bash
cat > k8s/secret-production.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: jamie-slack-secrets
  namespace: homepage
type: Opaque
stringData:
  SLACK_BOT_TOKEN: "xoxb-your-token-here"
  SLACK_APP_TOKEN: "xapp-your-token-here"
  SLACK_SIGNING_SECRET: "your-signing-secret-here"
  AGENT_SRE_URL: "http://sre-agent-mcp-server-service.agent-sre:30120"
EOF

# Apply the secret
kubectl apply -f k8s/secret-production.yaml

# Add to .gitignore
echo "k8s/secret-production.yaml" >> .gitignore
```

### 4. Build and Deploy

```bash
# Build the Docker image
make build

# Push to ECR
make push

# Deploy to Kubernetes
make deploy

# Check status
make status

# View logs
make logs
```

### 5. Deploy with Kustomize

```bash
# Deploy using kubectl and kustomize
kubectl apply -k k8s/

# Or if using Flux (GitOps)
# Jamie will be automatically deployed when changes are pushed to Git
# Check deployment status
flux get kustomizations jamie
```

## 💬 Usage

### Mention Jamie in Channels

```
@Jamie check the golden signals for bruno site
@Jamie what's the error rate for the API?
@Jamie list all pods in the default namespace
```

### Direct Messages

Open a DM with Jamie and just type:

```
Check the golden signals
Analyze these logs: [paste logs]
What alerts are firing?
```

### Slash Commands

- `/jamie-help` - Show help and available commands
- `/jamie-status` - Check Agent-SRE connection status
- `/jamie-analyze-logs [logs]` - Analyze logs for errors

## 🎯 Example Interactions

**Golden Signals Monitoring**
```
You: @Jamie check the golden signals for homepage
Jamie: 🤖 The API Response Time is 45ms, Traffic is 120 requests per minute, 
       Error Rate is 0.1%, and CPU Saturation is 35%. Everything looks healthy!
```

**Kubernetes Operations**
```
You: @Jamie show me the pods in the homepage namespace
Jamie: 🤖 I found 3 pods in the homepage namespace:
       • homepage-api-7d6c8f9b4-x8k2p (Running)
       • homepage-frontend-5f8d7c6b-m4n9j (Running)
       • jamie-slack-bot-6g9h8i7j-p5q6r (Running)
```

**Log Analysis**
```
You: /jamie-analyze-logs ERROR: Database connection timeout after 30s
Jamie: 🤖 Log Analysis Results

       📝 Analysis: Database connection timeout indicates network or resource issues

       ⚠️ Severity: high

       🔧 Recommendations:
       1. Check database pod health and resource allocation
       2. Verify network connectivity between services
       3. Review database connection pool settings
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SLACK_BOT_TOKEN` | Slack bot OAuth token | Required |
| `SLACK_APP_TOKEN` | Slack app-level token | Required |
| `SLACK_SIGNING_SECRET` | Slack signing secret | Required |
| `AGENT_SRE_URL` | Agent-SRE MCP service URL | `http://sre-agent-mcp-server-service.agent-sre:30120` |

### Resource Limits

Default resource allocation:

```yaml
resources:
  requests:
    memory: 128Mi
    cpu: 100m
  limits:
    memory: 256Mi
    cpu: 200m
```

## 📊 Monitoring

Jamie doesn't expose metrics by default but logs all interactions:

```bash
# View logs
kubectl logs -n homepage -l app=jamie-slack-bot -f

# Follow specific pod
kubectl logs -n homepage jamie-slack-bot-xxx-yyy -f
```

## 🐛 Troubleshooting

### Bot not responding

1. Check pod status:
   ```bash
   kubectl get pods -n homepage -l app=jamie-slack-bot
   ```

2. View logs for errors:
   ```bash
   kubectl logs -n homepage -l app=jamie-slack-bot --tail=100
   ```

3. Verify Agent-SRE is running:
   ```bash
   curl http://sre-agent-mcp-server-service.agent-sre:30120/health
   ```

### "Agent-SRE unavailable" errors

- Ensure Agent-SRE service is running
- Check network connectivity between pods
- Verify `AGENT_SRE_URL` is correct

### Slack connection issues

- Verify tokens are correct and not expired
- Check Socket Mode is enabled in Slack app settings
- Ensure app is installed to the workspace

## 🔐 Security

### Best Practices

1. **Never commit secrets** - Use Kubernetes secrets or external secret management
2. **Use RBAC** - Limit pod permissions via ServiceAccount
3. **Enable Socket Mode** - More secure than webhooks
4. **Rotate tokens regularly** - Update Slack tokens periodically
5. **Monitor logs** - Watch for suspicious activity

### Production Secrets

For production, use one of these secret management solutions:

- **Sealed Secrets**: Encrypt secrets in Git
- **External Secrets Operator**: Sync from AWS Secrets Manager, Vault, etc.
- **SOPS**: Encrypt secrets with age or GPG

Example with External Secrets:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: jamie-slack-secrets
  namespace: homepage
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secretsmanager
    kind: SecretStore
  target:
    name: jamie-slack-secrets
  data:
  - secretKey: SLACK_BOT_TOKEN
    remoteRef:
      key: slack/jamie/bot-token
```

## 🚀 Development

### Local Development

```bash
# Install dependencies
pip install -r requirements.txt

# Set environment variables
export SLACK_BOT_TOKEN="xoxb-..."
export SLACK_APP_TOKEN="xapp-..."
export SLACK_SIGNING_SECRET="..."
export AGENT_SRE_URL="http://sre-agent-mcp-server-service.agent-sre:30120"

# Run locally
python jamie_slack_bot.py
```

### Testing

```bash
# Run tests
make test

# Lint code
black jamie_slack_bot.py
flake8 jamie_slack_bot.py
```

## 📝 Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build` | Build Docker image |
| `make push` | Push image to ECR |
| `make deploy` | Deploy to Kubernetes |
| `make delete` | Delete from Kubernetes |
| `make logs` | Show pod logs |
| `make status` | Check deployment status |
| `make restart` | Restart deployment |
| `make clean` | Clean up local resources |
| `make all` | Build, push, and deploy |

## 🎨 Customization

### Change Bot Emoji

Edit `jamie_slack_bot.py`:

```python
self.bot_emoji = "🚀"  # Change to your preferred emoji
```

### Add Custom Commands

Add new slash commands in `_setup_handlers()`:

```python
@self.app.command("/jamie-custom")
async def handle_custom_command(ack, respond, command):
    await ack()
    # Your custom logic here
    await respond("Custom response")
```

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

MIT License - See LICENSE file for details

## 🔗 Related Projects

- [Agent-SRE](../agent-sre) - The MCP-powered SRE agent
- [Homepage API](../api) - Backend API service
- [Homepage Frontend](../frontend) - Web interface

## 💡 Tips

- Use threads for long conversations
- Jamie remembers context within a conversation
- Be specific with service names for better results
- Use slash commands for quick actions
- Check `/jamie-help` for the latest features

## 📚 Resources

- [Slack Bolt Documentation](https://slack.dev/bolt-python/)
- [MCP Protocol](https://github.com/modelcontextprotocol)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/)

---

Made with ❤️ by [Bruno Lucena](https://bruno.me)

