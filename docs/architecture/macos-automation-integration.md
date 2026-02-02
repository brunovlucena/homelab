# 🍎 macOS Automation Integration

Apple Events and Remote Automation integration for homelab agents.

## Overview

This integration enables Kubernetes agents running in your homelab to control macOS applications (especially Safari) via AppleScript and remote automation. The automation service runs locally on your Mac and exposes an HTTP API that agents can call.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Kubernetes Cluster (Homelab)               │
│                                                         │
│  ┌──────────────┐      ┌──────────────┐                │
│  │ Agent-Bruno  │      │ Agent-*      │                │
│  │ (Knative)    │      │ (Knative)    │                │
│  └──────┬───────┘      └──────┬───────┘                │
│         │                      │                         │
│         └──────────┬───────────┘                         │
│                    │                                      │
│         ┌──────────▼──────────┐                          │
│         │  MacOS Automation   │                          │
│         │  Client Library     │                          │
│         │  (shared-lib)       │                          │
│         └──────────┬──────────┘                          │
└────────────────────┼────────────────────────────────────┘
                     │
                     │ HTTP (host.docker.internal:8080)
                     │
┌────────────────────▼────────────────────────────────────┐
│              macOS Host (Mac Studio)                     │
│                                                         │
│  ┌──────────────────────────────────────────────┐    │
│  │  macOS Automation Service (FastAPI)           │    │
│  │  - HTTP API on localhost:8080                 │    │
│  │  - Executes AppleScript                        │    │
│  │  - Controls Safari/Applications               │    │
│  └──────────────────┬─────────────────────────────┘    │
│                     │                                   │
│                     │ AppleScript                       │
│                     │                                   │
│  ┌──────────────────▼─────────────────────────────┐    │
│  │  Safari (with Remote Automation enabled)       │    │
│  │  - Navigate URLs                               │    │
│  │  - Execute JavaScript                           │    │
│  │  - Control browser                             │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

## Components

### 1. macOS Automation Service

**Location**: `scripts/mac/apple-events/macos-automation-service.py`

FastAPI service that:
- Exposes HTTP API for automation actions
- Executes AppleScript commands
- Controls Safari and other macOS applications
- Handles CloudEvents from Kubernetes agents

**Endpoints**:
- `GET /health` - Health check
- `POST /v1/automation/execute` - Execute automation action
- `POST /v1/events` - Handle CloudEvents

### 2. Client Library

**Location**: `flux/ai/shared-lib/macos_automation/`

Python client library for agents:
- `MacOSAutomationClient` - Async HTTP client
- Supports navigate, execute_js, applescript, info actions
- CloudEvents integration

### 3. Setup Scripts

**Location**: `scripts/mac/apple-events/`

- `setup-apple-events.sh` - Enable Safari remote automation
- `start-automation-service.sh` - Start the service
- `test-apple-events.sh` - Test integration

## Quick Start

### 1. Enable Apple Events

```bash
cd scripts/mac/apple-events
./setup-apple-events.sh
```

This enables:
- Safari remote automation
- JavaScript from Apple Events
- Required Safari developer settings

### 2. Start Service

```bash
./start-automation-service.sh
```

Service runs on `http://localhost:8080`

### 3. Use from Agents

```python
from macos_automation import MacOSAutomationClient

client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")

# Navigate Safari
await client.navigate("https://lucena.cloud")

# Execute JavaScript
result = await client.execute_javascript("document.title")
print(result['result']['output'])
```

## Integration Examples

### agent-bruno Integration

Add automation capability to agent-bruno:

```python
# In agent-bruno/src/chatbot/main.py
from macos_automation import MacOSAutomationClient, AutomationError

async def handle_automation_request(message: str):
    """Handle automation requests from chat"""
    client = MacOSAutomationClient()
    
    try:
        if "open" in message.lower() and "lucena.cloud" in message.lower():
            await client.navigate("https://lucena.cloud")
            return "✅ Opened lucena.cloud in Safari"
        
        elif "get title" in message.lower():
            result = await client.execute_javascript("document.title")
            return f"📄 Page title: {result['result']['output']}"
        
    except AutomationError as e:
        return f"❌ Automation failed: {e}"
    finally:
        await client.close()
```

### CloudEvents Integration

Agents can send CloudEvents to trigger automation:

```python
from macos_automation import MacOSAutomationClient

client = MacOSAutomationClient()

result = await client.send_cloudevent(
    event_type="io.homelab.macos.automation.request",
    event_source="/agent-bruno/automation",
    data={
        "action": "navigate",
        "url": "https://lucena.cloud"
    }
)
```

## Supported Actions

| Action | Description | Required Fields |
|--------|-------------|----------------|
| `navigate` | Navigate browser to URL | `url` |
| `execute_js` | Execute JavaScript in current tab | `javascript` |
| `applescript` | Execute raw AppleScript | `applescript` |
| `info` | Get current browser state | None |

## Security Considerations

1. **Local Network Only**: Service runs on `localhost:8080` by default
2. **Accessibility Permissions**: macOS requires accessibility permissions
3. **Safari Settings**: Remote automation must be enabled
4. **Firewall**: Port 8080 must be accessible from Kubernetes

## Troubleshooting

### Service Not Accessible from Kubernetes

```bash
# Check service is running
curl http://localhost:8080/health

# From Kubernetes pod
curl http://host.docker.internal:8080/health
```

### Safari Not Responding

1. Restart Safari after enabling remote automation
2. Check permissions in System Settings > Privacy & Security > Accessibility
3. Verify settings:
   ```bash
   defaults read com.apple.Safari AllowRemoteAutomation
   ```

## Files

- `scripts/mac/apple-events/` - Setup scripts and service
- `flux/ai/shared-lib/macos_automation/` - Client library
- `docs/architecture/macos-automation-integration.md` - This document

## Related Documentation

- [Apple Events Setup Guide](../../scripts/mac/apple-events/README.md)
- [Shared Library Documentation](../flux/ai/shared-lib/README.md)

