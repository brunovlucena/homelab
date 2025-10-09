# 🚀 Quick Start Guide - Jamie Slack Bot

Get Jamie up and running in your Slack workspace in 10 minutes!

## 📋 Prerequisites

- ✅ Slack workspace with admin access
- ✅ Kubernetes cluster running
- ✅ Agent-SRE service deployed
- ✅ kubectl configured
- ✅ Docker installed
- ✅ ECR access (or other container registry)

## 🎯 Step 1: Create Slack App (5 minutes)

### Option A: Using Manifest (Recommended)

1. Go to https://api.slack.com/apps
2. Click **"Create New App"**
3. Select **"From a manifest"**
4. Choose your workspace
5. Copy and paste the contents from `slack-app-manifest.yaml`
6. Click **"Create"**
7. Review permissions and click **"Install to Workspace"**

### Option B: Manual Setup

1. Go to https://api.slack.com/apps
2. Click **"Create New App"** → **"From scratch"**
3. Name it **"Jamie"** and select your workspace
4. Configure the following:

**OAuth & Permissions** → Add these Bot Token Scopes:
```
app_mentions:read
chat:write
chat:write.public
commands
im:history
im:read
im:write
channels:history
channels:read
groups:history
groups:read
mpim:history
mpim:read
```

**Event Subscriptions** → Enable and subscribe to:
```
app_mention
message.channels
message.groups
message.im
message.mpim
```

**Socket Mode** → Enable Socket Mode

**App-Level Tokens** → Create token with `connections:write` scope

**Slash Commands** → Create:
- `/jamie-help`
- `/jamie-status`
- `/jamie-analyze-logs`

5. Install to workspace

## 🔑 Step 2: Get Your Tokens (2 minutes)

### Bot Token
1. Go to **OAuth & Permissions**
2. Copy **Bot User OAuth Token** (starts with `xoxb-`)

### App Token
1. Go to **Basic Information**
2. Scroll to **App-Level Tokens**
3. Copy your token (starts with `xapp-`)

### Signing Secret
1. Go to **Basic Information**
2. Copy **Signing Secret**

## 🔐 Step 3: Configure Secrets (1 minute)

Create your production secrets file:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/jamie-slack

cat > k8s/secret-production.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: jamie-slack-secrets
  namespace: homepage
type: Opaque
stringData:
  SLACK_BOT_TOKEN: "xoxb-YOUR-TOKEN-HERE"
  SLACK_APP_TOKEN: "xapp-YOUR-TOKEN-HERE"
  SLACK_SIGNING_SECRET: "YOUR-SIGNING-SECRET-HERE"
  AGENT_SRE_URL: "http://sre-agent-mcp-server-service.agent-sre:30120"
EOF

# Apply the secret
kubectl apply -f k8s/secret-production.yaml

# Make sure it's in .gitignore
echo "k8s/secret-production.yaml" >> .gitignore
```

## 🏗️ Step 4: Build & Deploy (2 minutes)

```bash
# Build the Docker image
make build

# Push to ECR (login first if needed)
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin 565265565115.dkr.ecr.us-west-2.amazonaws.com

make push

# Deploy to Kubernetes
make deploy

# Check status
make status
```

### Alternative: Deploy with Kustomize Only

```bash
# Deploy directly using kubectl
kubectl apply -k k8s/

# Or wait for Flux to sync (if using GitOps)
flux reconcile kustomization jamie --with-source
```

## ✅ Step 5: Test Jamie (1 minute)

1. **In Slack**, go to any channel where Jamie is added
2. Mention Jamie:
   ```
   @Jamie hello
   ```

3. You should see:
   ```
   🤖 Hey! I'm Jamie, your SRE assistant. How can I help you today?
   ```

4. Try a command:
   ```
   @Jamie check the golden signals for homepage
   ```

5. Or send a DM:
   - Open a direct message with Jamie
   - Type: `help`

## 🎉 You're Done!

Jamie is now running in your Slack workspace!

## 🔍 Verify Everything Works

```bash
# Check pod is running
kubectl get pods -n homepage -l app=jamie-slack-bot

# View logs
kubectl logs -n homepage -l app=jamie-slack-bot -f

# Check Agent-SRE connection
kubectl exec -it -n homepage deployment/jamie-slack-bot -- \
  python -c "import requests; print(requests.get('http://sre-agent-mcp-server-service.agent-sre:30120/health').json())"
```

## 🐛 Troubleshooting

### Jamie not responding?

1. **Check pod status:**
   ```bash
   kubectl get pods -n homepage -l app=jamie-slack-bot
   ```

2. **View logs:**
   ```bash
   kubectl logs -n homepage -l app=jamie-slack-bot --tail=50
   ```

3. **Verify secrets:**
   ```bash
   kubectl get secret jamie-slack-secrets -n homepage -o yaml
   ```

### "Agent-SRE unavailable" error?

1. **Check Agent-SRE is running:**
   ```bash
   kubectl get pods -n homepage -l app=agent-sre
   ```

2. **Test Agent-SRE directly:**
   ```bash
   kubectl run curl-test --rm -it --image=curlimages/curl -- \
     curl http://sre-agent-mcp-server-service.agent-sre:30120/health
   ```

3. **Check network policies:**
   ```bash
   kubectl get networkpolicies -n homepage
   ```

### Slack connection issues?

1. **Verify Socket Mode is enabled** in Slack app settings
2. **Check tokens are correct** and not expired
3. **Ensure app is installed** to workspace
4. **Restart the pod:**
   ```bash
   make restart
   ```

## 📱 Next Steps

### Try These Commands

**Golden Signals:**
```
@Jamie check the golden signals for bruno site
@Jamie what's the error rate?
```

**Kubernetes:**
```
@Jamie list pods in homepage namespace
@Jamie show logs for homepage-api
```

**Log Analysis:**
```
/jamie-analyze-logs ERROR: Database connection timeout
```

**Help:**
```
/jamie-help
```

### Customize Jamie

1. **Change bot emoji** - Edit `jamie_slack_bot.py`
2. **Add custom commands** - See README.md
3. **Adjust resources** - Edit `k8s/deployment.yaml`
4. **Configure monitoring** - Add ServiceMonitor

### Join the Community

- 📖 Read the full README.md
- 🐛 Report issues on GitHub
- 💡 Share your use cases
- 🤝 Contribute improvements

---

**Need help?** Check the main [README.md](README.md) or logs:
```bash
make logs
```

Enjoy your new SRE companion! 🎉

