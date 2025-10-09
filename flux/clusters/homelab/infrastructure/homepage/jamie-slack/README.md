# 🤖 Jamie Slack Bot

**Your SRE Companion on Slack** - Powered by Ollama AI and Agent-SRE with MCP

Jamie is an intelligent Slack bot that helps you monitor, troubleshoot, and manage your infrastructure directly from Slack. It combines the power of Bruno's fine-tuned Ollama model with Agent-SRE's MCP tools for comprehensive SRE assistance.

## ✨ Features

- 🧠 **AI-Powered Responses** - Bruno's fine-tuned SRE model via Ollama
- 🎯 **Golden Signals Monitoring** - Track latency, traffic, errors, and saturation
- ☸️ **Kubernetes Operations** - Manage pods, deployments, and view logs
- 📈 **Grafana Integration** - Access dashboards, incidents, and alerts
- 🔍 **Log Analysis** - Intelligent log parsing and error detection
- 🚨 **Incident Management** - Create and track incidents
- 💬 **Conversational AI** - Natural language interactions
- 🧠 **Context Awareness** - Remembers conversation history
- 🔄 **Smart Routing** - Automatically routes to best service (Ollama or Agent-SRE)

## 🏗️ Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│             │      │             │      │             │
│   Slack     │─────►│   Jamie     │─────►│   Ollama    │
│   Users     │      │   Bot       │      │   (AI)      │
│             │      │             │      │             │
└─────────────┘      └─────────────┘      └─────────────┘
                            │                     │
                            │                     │
                            ▼                     ▼
                     ┌─────────────┐      ┌─────────────┐
                     │ Conversation│      │ Bruno's     │
                     │ Context     │      │ Fine-tuned  │
                     │ Storage     │      │ SRE Model   │
                     └─────────────┘      └─────────────┘
                            │
                            ▼
                     ┌─────────────┐      ┌─────────────┐
                     │ Smart       │─────►│ Agent-SRE   │
                     │ Routing     │      │ (MCP)       │
                     │ Engine      │      │             │
                     └─────────────┘      └─────────────┘
                                                │
                                                ▼
                                          ┌─────────────┐
                                          │ MCP Tools   │
                                          │ • Grafana   │
                                          │ • K8s       │
                                          │ • Prometheus│
                                          │ • Custom    │
                                          └─────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster
- Slack workspace with admin access
- Ollama server running (Bruno's at 192.168.0.16:11434)
- Agent-SRE service running (optional, for advanced features)
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
  AGENT_SRE_URL: "http://homepage-api:8080"
  OLLAMA_URL: "http://192.168.0.16:11434"
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

### 5. Deploy with Helm

```bash
# Install with Helm
helm install jamie-slack ./chart \
  --namespace homepage \
  --create-namespace \
  --set slack.botToken="xoxb-your-token" \
  --set slack.appToken="xapp-your-token" \
  --set slack.signingSecret="your-secret"

# Or use values file
helm install jamie-slack ./chart \
  --namespace homepage \
  --values my-values.yaml
```

## 💬 Usage

### Mention Jamie in Channels

```
@Jamie hello
@Jamie check the golden signals for bruno site
@Jamie what's the error rate for the API?
@Jamie list all pods in the default namespace
@Jamie how do I troubleshoot a failing deployment?
```

### Direct Messages

Open a DM with Jamie and just type:

```
Hello Jamie!
What are the best practices for monitoring microservices?
Help me understand Kubernetes resource limits
```

### Slash Commands

- `/jamie-help` - Show help and available commands
- `/jamie-status` - Check Ollama and Agent-SRE connection status
- `/jamie-analyze-logs [logs]` - Analyze logs for errors

## 🎯 Example Interactions

**AI-Powered SRE Advice**
```
You: @Jamie how do I improve my application's performance?
Jamie: 🤖 Here are some key strategies to improve application performance:

1. **Monitor Golden Signals**: Track latency, traffic, error rates, and saturation
2. **Optimize Database Queries**: Use indexes, connection pooling, and query optimization
3. **Implement Caching**: Use Redis or Memcached for frequently accessed data
4. **CDN Implementation**: Serve static assets from edge locations
5. **Horizontal Scaling**: Add more instances during high traffic periods

Would you like me to check your current golden signals or analyze specific performance bottlenecks?

Powered by: ollama (bruno-sre)
```

**Golden Signals Monitoring**
```
You: @Jamie check the golden signals for homepage
Jamie: 🤖 The API Response Time is 45ms, Traffic is 120 requests per minute, 
       Error Rate is 0.1%, and CPU Saturation is 35%. Everything looks healthy!

Sources: Agent-SRE
Powered by: agent-sre-mcp
```

**Kubernetes Operations**
```
You: @Jamie show me the pods in the homepage namespace
Jamie: 🤖 I found 3 pods in the homepage namespace:
       • homepage-api-7d6c8f9b4-x8k2p (Running)
       • homepage-frontend-5f8d7c6b-m4n9j (Running)
       • jamie-slack-bot-6g9h8i7j-p5q6r (Running)

Sources: Agent-SRE
Powered by: agent-sre-mcp
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

       🔧 Powered by: agent-sre-mcp
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SLACK_BOT_TOKEN` | Slack bot OAuth token | Required |
| `SLACK_APP_TOKEN` | Slack app-level token | Required |
| `SLACK_SIGNING_SECRET` | Slack signing secret | Required |
| `AGENT_SRE_URL` | Agent-SRE service URL | `http://homepage-api:8080` |
| `OLLAMA_URL` | Ollama server URL | `http://192.168.0.16:11434` |

### Smart Routing

Jamie automatically routes questions to the best service:

**Ollama AI** (for general SRE questions):
- Best practices and advice
- General troubleshooting guidance
- Educational content
- When Agent-SRE is unavailable

**Agent-SRE MCP** (for specific operations):
- Golden signals monitoring
- Kubernetes operations
- Grafana dashboards
- Log analysis
- Incident management

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

3. Check service status:
   ```bash
   /jamie-status
   ```

### Ollama connection issues

1. **Verify Ollama is running**:
   ```bash
   curl http://192.168.0.16:11434/api/tags
   ```

2. **Check network connectivity**:
   ```bash
   kubectl run curl-test --rm -it --image=curlimages/curl -- \
     curl http://192.168.0.16:11434/api/tags
   ```

3. **Verify model is available**:
   ```bash
   curl http://192.168.0.16:11434/api/tags | grep bruno-sre
   ```

### Agent-SRE unavailable

- Ensure Agent-SRE service is running
- Check network connectivity between pods
- Verify `AGENT_SRE_URL` is correct
- Jamie will automatically fallback to Ollama

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
6. **Network policies** - Restrict network access to Ollama and Agent-SRE

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
  - secretKey: OLLAMA_URL
    remoteRef:
      key: infrastructure/ollama/url
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
export AGENT_SRE_URL="http://localhost:8080"
export OLLAMA_URL="http://192.168.0.16:11434"

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

### Modify Routing Logic

Edit the `_route_question()` method to customize which service handles which types of questions.

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
- [Ollama Server](http://192.168.0.16:11434) - Bruno's AI model server

## 💡 Tips

- Use threads for long conversations
- Jamie remembers context within a conversation
- Be specific with service names for better results
- Use slash commands for quick actions
- Check `/jamie-help` for the latest features
- Jamie automatically chooses the best AI service for your question

## 📚 Resources

- [Slack Bolt Documentation](https://slack.dev/bolt-python/)
- [MCP Protocol](https://github.com/modelcontextprotocol)
- [Ollama Documentation](https://ollama.ai/docs)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/)

---

Made with ❤️ by [Bruno Lucena](https://bruno.me)

Powered by **Ollama AI** 🤖 + **Agent-SRE MCP** 🔧