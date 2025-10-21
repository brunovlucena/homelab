# 🚀 How to Trigger Investigations - API & Slack Guide

This guide shows you how to manually trigger investigations using Agent-SRE's API, MCP server, or Slack (via Jamie).

## 🎯 Quick Start

### Via Direct API

```bash
# From inside the cluster
curl -X POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Navigation menu not visible in mobile",
    "description": "Users report the nav menu is hidden on mobile devices",
    "severity": "medium",
    "component": "homepage"
  }'
```

### Via kubectl port-forward

```bash
# Terminal 1: Port-forward the service
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080

# Terminal 2: Trigger investigation
curl -X POST http://localhost:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Database connection timeouts",
    "description": "Seeing increased connection timeouts in production",
    "severity": "high",
    "component": "api"
  }'
```

### Via Slack (Jamie Integration)

```
@jamie investigate "High error rate in homepage API" severity=critical component=homepage
```

## 📋 API Reference

### 1. Create Investigation

**Endpoint:** `POST /investigation/create`

**Request Body:**
```json
{
  "title": "string (required)",
  "description": "string (required)",
  "severity": "critical|high|medium|low (optional, default: medium)",
  "component": "string (optional, default: unknown)"
}
```

**Response:**
```json
{
  "investigation": {
    "investigation_id": "abc-123-xyz",
    "issue_number": 28,
    "issue_url": "https://github.com/brunovlucena/homelab/issues/28",
    "root_cause": "...",
    "recommendations": [...],
    "error_patterns": 3,
    "slow_requests": 1,
    "completed": true
  },
  "service": "sre-agent",
  "timestamp": "2025-10-17T10:30:00Z"
}
```

**Example:**
```bash
curl -X POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Homepage API latency spike",
    "description": "P95 latency increased from 200ms to 2000ms in the last 30 minutes",
    "severity": "high",
    "component": "homepage"
  }'
```

### 2. Get Investigation Details

**Endpoint:** `GET /investigation/{investigation_id}`

**Response:**
```json
{
  "investigation": {
    "id": "abc-123-xyz",
    "name": "...",
    "status": "completed",
    "created_at": "2025-10-17T10:30:00Z",
    "analyses": [...]
  },
  "service": "sre-agent",
  "timestamp": "2025-10-17T10:35:00Z"
}
```

**Example:**
```bash
curl http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/abc-123-xyz
```

### 3. Workflow Failure Investigation (Optional)

**Endpoint:** `POST /investigation/workflow-failure`

This is useful if you want to manually investigate a CI/CD failure:

```bash
curl -X POST http://localhost:8080/investigation/workflow-failure \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "Homepage CI",
    "run_id": "18595106881",
    "job_id": "53019314213",
    "run_url": "https://github.com/brunovlucena/homelab/actions/runs/18595106881",
    "job_url": "https://github.com/brunovlucena/homelab/actions/runs/18595106881/job/53019314213",
    "failure_details": "Build failed with TypeScript errors..."
  }'
```

## 🤖 MCP Server Integration

The investigation functionality is also available via the MCP server protocol.

### MCP Tool: `sre_investigate`

Create this as a new MCP tool in your `mcp-server.py`:

```python
@mcp_server.tool()
async def sre_investigate(
    title: str,
    description: str,
    severity: str = "medium",
    component: str = "unknown"
) -> str:
    """
    Create and run a full investigation with GitHub issue creation.
    
    Args:
        title: Investigation title
        description: Detailed description
        severity: One of: critical, high, medium, low
        component: Affected component/service
    
    Returns:
        Investigation results with GitHub issue URL
    """
    # Call the investigation workflow
    result = await investigation_workflow.investigate(
        title=title,
        description=description,
        severity=severity,
        component=component
    )
    
    return json.dumps(result, indent=2)
```

Then use it via MCP:

```bash
curl -X POST http://sre-agent-mcp-server-service:30120/mcp/tool \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "sre_investigate",
    "arguments": {
      "title": "Memory leak in agent-bruno",
      "description": "Memory usage increasing steadily over 24 hours",
      "severity": "high",
      "component": "agent-bruno"
    }
  }'
```

## 💬 Slack Integration (Jamie)

Integrate investigations with your Slack bot Jamie.

### Option 1: Add Investigation Command to Jamie

Add this to `jamie_slack_bot.py`:

```python
async def handle_investigate_command(self, command_text: str, channel_id: str, user_id: str):
    """Handle /investigate command or @jamie investigate"""
    
    # Parse command: "investigate <title> [severity=high] [component=homepage]"
    import re
    
    # Extract title (everything before first key=value)
    title_match = re.match(r'^investigate\s+"([^"]+)"', command_text)
    if not title_match:
        title_match = re.match(r'^investigate\s+([^\s]+(?:\s+[^\s=]+)*)', command_text)
    
    if not title_match:
        await self.send_message(
            channel_id,
            "❌ Usage: `@jamie investigate \"Issue title\" [severity=high] [component=homepage]`"
        )
        return
    
    title = title_match.group(1)
    
    # Extract optional parameters
    severity_match = re.search(r'severity=(\w+)', command_text)
    component_match = re.search(r'component=(\w+)', command_text)
    
    severity = severity_match.group(1) if severity_match else "medium"
    component = component_match.group(1) if component_match else "unknown"
    
    # Build description with context
    description = f"""## Investigation Requested via Slack

**Requested by**: <@{user_id}>
**Channel**: <#{channel_id}>
**Timestamp**: {datetime.now().isoformat()}

### User Input
{command_text}

---

_This investigation was triggered from Slack._
"""
    
    # Send initial message
    await self.send_message(
        channel_id,
        f"🔍 Starting investigation: *{title}*\n\n" +
        f"_Severity: {severity} | Component: {component}_\n\n" +
        "This may take 1-2 minutes..."
    )
    
    try:
        # Call Agent-SRE investigation API
        async with httpx.AsyncClient() as client:
            response = await client.post(
                "http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create",
                json={
                    "title": title,
                    "description": description,
                    "severity": severity,
                    "component": component,
                },
                timeout=180.0  # 3 minutes
            )
            
            if response.status_code == 200:
                result = response.json()
                investigation = result.get("investigation", {})
                
                # Format results for Slack
                blocks = [
                    {
                        "type": "header",
                        "text": {
                            "type": "plain_text",
                            "text": "🔍 Investigation Complete"
                        }
                    },
                    {
                        "type": "section",
                        "fields": [
                            {
                                "type": "mrkdwn",
                                "text": f"*Title:*\n{title}"
                            },
                            {
                                "type": "mrkdwn",
                                "text": f"*Severity:*\n{severity}"
                            },
                            {
                                "type": "mrkdwn",
                                "text": f"*Component:*\n{component}"
                            },
                            {
                                "type": "mrkdwn",
                                "text": f"*Status:*\n{'✅ Completed' if investigation.get('completed') else '⚠️ Partial'}"
                            }
                        ]
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"*📊 Findings:*\n• Error patterns: {investigation.get('error_patterns', 0)}\n• Slow requests: {investigation.get('slow_requests', 0)}"
                        }
                    }
                ]
                
                # Add GitHub issue link if available
                if investigation.get("issue_url"):
                    blocks.append({
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"*🔗 GitHub Issue:*\n<{investigation['issue_url']}|Issue #{investigation.get('issue_number')}>"
                        }
                    })
                
                # Add root cause preview
                if investigation.get("root_cause"):
                    root_cause = investigation["root_cause"][:200] + "..." if len(investigation.get("root_cause", "")) > 200 else investigation.get("root_cause", "")
                    blocks.append({
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"*🧠 Root Cause (preview):*\n```{root_cause}```"
                        }
                    })
                
                blocks.append({
                    "type": "context",
                    "elements": [
                        {
                            "type": "mrkdwn",
                            "text": f"Investigation ID: `{investigation.get('investigation_id', 'N/A')}`"
                        }
                    ]
                })
                
                await self.send_message(channel_id, blocks=blocks)
                
            else:
                await self.send_message(
                    channel_id,
                    f"❌ Investigation failed: HTTP {response.status_code}\n```{response.text[:500]}```"
                )
    
    except Exception as e:
        logger.error(f"Error running investigation: {e}", exc_info=True)
        await self.send_message(
            channel_id,
            f"❌ Investigation failed: {str(e)}"
        )
```

Then register the command in Jamie's message handler:

```python
async def handle_message(self, event: Dict[str, Any]):
    text = event.get("text", "")
    
    # ... existing code ...
    
    # Check for investigate command
    if "investigate" in text.lower() and (
        f"<@{self.bot_user_id}>" in text or 
        text.startswith("/investigate")
    ):
        await self.handle_investigate_command(
            text, 
            event.get("channel"),
            event.get("user")
        )
        return
```

### Option 2: Interactive Slack Button

Add investigation button to error notifications:

```python
# When Jamie posts an error/alert in Slack, add action buttons:
blocks = [
    # ... error details ...
    {
        "type": "actions",
        "elements": [
            {
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "🔍 Investigate"
                },
                "style": "primary",
                "action_id": "trigger_investigation",
                "value": json.dumps({
                    "title": f"Alert: {alert_name}",
                    "severity": severity,
                    "component": component
                })
            },
            {
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "📊 View Grafana"
                },
                "url": grafana_url
            }
        ]
    }
]
```

Then handle the button click:

```python
async def handle_button_click(self, action: Dict[str, Any], response_url: str):
    action_id = action.get("action_id")
    
    if action_id == "trigger_investigation":
        value = json.loads(action.get("value", "{}"))
        
        # Trigger investigation
        await self.handle_investigate_command(
            f"investigate \"{value['title']}\" severity={value['severity']} component={value['component']}",
            action.get("channel_id"),
            action.get("user_id")
        )
```

## 📝 Common Use Cases

### Use Case 1: Investigating an Alert

When you receive an alert in Slack or Prometheus:

```bash
curl -X POST http://localhost:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "High CPU usage on homepage pods",
    "description": "CPU usage spiked to 90% on homepage-api pods. Alert triggered at 10:45 AM.",
    "severity": "high",
    "component": "homepage"
  }'
```

**What happens:**
1. ✅ GitHub issue created with alert details
2. 🔬 Sift runs error pattern and slow request analysis
3. 🧠 LLM analyzes logs and metrics
4. 💡 Recommendations generated
5. 🔄 Issue updated with findings

### Use Case 2: Investigating User Reports

When users report issues:

```bash
curl -X POST http://localhost:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Navigation menu not visible on mobile",
    "description": "Multiple users report the navigation menu is not visible on mobile devices (iOS Safari and Chrome Android). Started around 9 AM today.",
    "severity": "medium",
    "component": "homepage"
  }'
```

### Use Case 3: Proactive Investigation

Scheduled investigations via cronjob:

```bash
# Create a simple script
cat > /usr/local/bin/investigate-platform-health.sh <<'EOF'
#!/bin/bash
curl -X POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Daily Platform Health Check",
    "description": "Scheduled daily investigation to detect anomalies",
    "severity": "low",
    "component": "platform"
  }'
EOF

chmod +x /usr/local/bin/investigate-platform-health.sh

# Add to crontab
crontab -e
# Add: 0 6 * * * /usr/local/bin/investigate-platform-health.sh
```

## 🛠️ Advanced Usage

### Investigation from Python

```python
import httpx
import asyncio

async def trigger_investigation(title: str, description: str):
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create",
            json={
                "title": title,
                "description": description,
                "severity": "high",
                "component": "api"
            },
            timeout=180.0
        )
        
        if response.status_code == 200:
            result = response.json()
            print(f"✅ Investigation created: {result['investigation']['issue_url']}")
            return result
        else:
            print(f"❌ Failed: {response.status_code}")
            return None

# Use it
asyncio.run(trigger_investigation(
    "Database query performance degradation",
    "Queries taking 5x longer than baseline"
))
```

### Investigation from Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func triggerInvestigation(title, description, severity, component string) error {
    payload := map[string]string{
        "title":       title,
        "description": description,
        "severity":    severity,
        "component":   component,
    }
    
    jsonData, _ := json.Marshal(payload)
    
    resp, err := http.Post(
        "http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 200 {
        var result map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&result)
        fmt.Printf("✅ Investigation created: %v\n", result)
    }
    
    return nil
}
```

## 🔒 Security Considerations

### Authentication (Optional)

Add authentication to the investigation endpoint:

```python
# In agent.py
async def handle_create_investigation(self, request: Request) -> Response:
    # Check for API key
    api_key = request.headers.get("X-API-Key")
    expected_key = os.getenv("AGENT_SRE_API_KEY")
    
    if expected_key and api_key != expected_key:
        return web.json_response({"error": "Unauthorized"}, status=401)
    
    # ... rest of the handler ...
```

Then use it:

```bash
curl -X POST http://localhost:8080/investigation/create \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-key" \
  -d '{...}'
```

### Rate Limiting (Optional)

Prevent abuse with rate limiting:

```python
from collections import defaultdict
import time

class RateLimiter:
    def __init__(self, max_requests=10, window=60):
        self.max_requests = max_requests
        self.window = window
        self.requests = defaultdict(list)
    
    def is_allowed(self, client_id: str) -> bool:
        now = time.time()
        # Clean old requests
        self.requests[client_id] = [
            req_time for req_time in self.requests[client_id]
            if now - req_time < self.window
        ]
        
        if len(self.requests[client_id]) < self.max_requests:
            self.requests[client_id].append(now)
            return True
        return False

# Use in handler
rate_limiter = RateLimiter(max_requests=5, window=300)  # 5 per 5 minutes

async def handle_create_investigation(self, request: Request) -> Response:
    client_ip = request.remote
    
    if not rate_limiter.is_allowed(client_ip):
        return web.json_response(
            {"error": "Rate limit exceeded. Try again later."},
            status=429
        )
    
    # ... rest of handler ...
```

## 📊 Monitoring

Track investigation metrics:

```bash
# Get investigation stats
kubectl exec -n agent-sre deployment/sre-agent -- \
  curl -s http://localhost:8080/status | jq '.'

# Watch logs
kubectl logs -n agent-sre -l app=sre-agent -f --tail=50

# Check Sift database
kubectl exec -n agent-sre deployment/sre-agent-mcp-server -- \
  sqlite3 /tmp/sift_investigations.db "SELECT COUNT(*) FROM investigations;"
```

## 🎉 Next Steps

1. **Test the API**: Try creating a test investigation
2. **Integrate with Jamie**: Add the Slack command to your bot
3. **Create shortcuts**: Add aliases or scripts for common investigations
4. **Monitor results**: Check GitHub issues and Grafana dashboards
5. **Customize**: Extend the LangGraph workflow for your specific needs

## 📚 Related Documentation

- [Automated Investigation Guide](AUTOMATED_INVESTIGATION_GUIDE.md) - Full architecture details
- [Sift Implementation](SIFT_IMPLEMENTATION.md) - How Sift works
- [Agent-SRE README](README.md) - Overall system documentation

---

**Questions?** Open an issue or check the logs!



