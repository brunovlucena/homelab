# Linear Integration with Homelab Agents

This guide explains how to integrate Linear with your Kubernetes agents in the homelab using the shared Linear client library.

## Overview

The homelab provides a **shared Linear client library** (`shared-libs/linear_client`) that all agents can use to interact with Linear's GraphQL API. This approach:

- ✅ **Centralized**: One client library maintained in `shared-libs/linear_client`
- ✅ **Simple**: Agents just import and use the client
- ✅ **Secure**: API keys stored in Kubernetes secrets
- ✅ **Tested**: Full unit and integration test coverage

## Current Status

- ✅ **Linear Client Library**: Available at `flux/ai/shared-libs/linear_client/`
- ✅ **Unit Tests**: 12/12 passing (mocked, no API key needed)
- ✅ **Integration Tests**: 8/8 passing (real API calls)
- ⚠️ **Agent Integration**: Add to individual agents as needed

---

## Quick Start

### Step 1: Create Linear API Secret

Create a Kubernetes secret with your Linear API key:

```bash
# Get your API key from Linear Settings → API (do not commit real keys)

kubectl create secret generic linear-api-key \
  -n ai \
  --from-literal=api-key=lin_api_xxxxxxxxxxxxx \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Note**: The secret should be created in the `ai` namespace where your agents run.

### Step 2: Add Linear Client to Your Agent

#### Option A: Copy the Client (Recommended for Standalone Agents)

Copy the `linear_client.py` file to your agent's source directory:

```bash
# Example: Add to agent-sre
cp flux/ai/shared-libs/linear_client/linear_client.py \
   flux/ai/agent-sre/src/linear_client.py
```

#### Option B: Install as Shared Library (For Multiple Agents)

If multiple agents need it, you can package it as a shared library or add it to a common base image.

### Step 3: Add Dependencies

Add `httpx` to your agent's `requirements.txt`:

```txt
httpx>=0.25.0
```

### Step 4: Configure Agent to Use Linear API Key

#### For LambdaAgent (Knative-based)

```yaml
# flux/ai/agent-sre/k8s/kustomize/base/lambdaagent.yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: agent-sre
  namespace: ai
spec:
  # ... existing config ...
  
  # Add Linear API key from secret
  env:
    - name: LINEAR_API_KEY
      valueFrom:
        secretKeyRef:
          name: linear-api-key
          key: api-key
```

#### For Regular Kubernetes Deployment

```yaml
# flux/ai/agent-sre/k8s/kustomize/base/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-sre
  namespace: ai
spec:
  template:
    spec:
      containers:
      - name: agent-sre
        # ... existing config ...
        env:
        - name: LINEAR_API_KEY
          valueFrom:
            secretKeyRef:
              name: linear-api-key
              key: api-key
```

### Step 5: Use in Agent Code

```python
# src/sre_agent/linear_integration.py
import asyncio
from linear_client import LinearClient, create_agent_issue

class SRELinearIntegration:
    """SRE Agent integration with Linear."""
    
    def __init__(self):
        # Client reads LINEAR_API_KEY from environment
        self.linear = LinearClient()
    
    async def create_incident_ticket(self, title: str, description: str, severity: str):
        """Create a Linear issue for an incident."""
        # Get team by key
        team = await self.linear.get_team("BVL")  # Your team key
        
        # Create issue
        issue = await self.linear.create_issue(
            title=f"[Incident] {title}",
            description=f"""
**Severity**: {severity}

{description}

Created by agent-sre
            """.strip(),
            team_id=team["id"],
            priority=1 if severity == "critical" else 2  # Urgent or High
        )
        return issue["url"]
    
    async def list_my_issues(self):
        """List issues assigned to the agent's user."""
        issues = await self.linear.list_issues(assignee="me", limit=10)
        return issues
    
    async def quick_create(self, title: str, description: str):
        """Quick helper to create an issue."""
        return await create_agent_issue(
            title=title,
            description=description,
            team_key="BVL",
            agent_name="agent-sre",
            priority=3  # Normal
        )

# Usage example
async def handle_alert(alert_data):
    integration = SRELinearIntegration()
    
    # Create Linear ticket for critical alert
    if alert_data["severity"] == "critical":
        ticket_url = await integration.create_incident_ticket(
            title=alert_data["title"],
            description=alert_data["description"],
            severity="critical"
        )
        print(f"Created Linear ticket: {ticket_url}")
```

---

## Linear Client API Reference

The `LinearClient` class provides the following methods:

### List Issues

```python
issues = await client.list_issues(
    team="BVL",           # Team name (optional)
    assignee="me",        # Assignee name or "me" (optional)
    state="In Progress",  # State name (optional)
    limit=50              # Max results (default: 50)
)
```

### Get Issue

```python
issue = await client.get_issue(issue_id="BVL-16")  # Or UUID
```

### Create Issue

```python
issue = await client.create_issue(
    title="Issue title",
    description="Issue description (markdown supported)",
    team_id="team-uuid-here",  # Required
    assignee_id="user-uuid",   # Optional
    priority=2,                # 0=None, 1=Urgent, 2=High, 3=Normal, 4=Low
    state_id="state-uuid",     # Optional initial state
    label_ids=["label-uuid"]   # Optional list of label IDs
)
```

### Update Issue

```python
updated = await client.update_issue(
    issue_id="BVL-16",
    title="Updated title",      # Optional
    description="New desc",    # Optional
    assignee_id="user-uuid",   # Optional
    state_id="state-uuid",     # Optional
    priority=1                 # Optional
)
```

### Create Comment

```python
comment = await client.create_comment(
    issue_id="BVL-16",
    body="Comment text (markdown supported)"
)
```

### List Teams

```python
teams = await client.list_teams()
for team in teams:
    print(f"{team['name']} ({team['key']})")
```

### Get Team by Key

```python
team = await client.get_team("BVL")  # Team key like "BVL", "SRE", etc.
team_id = team["id"]
```

### Quick Helper Function

```python
from linear_client import create_agent_issue

# Creates an issue with agent attribution
url = await create_agent_issue(
    title="System alert",
    description="CPU usage high",
    team_key="BVL",
    agent_name="agent-sre",
    priority=2
)
```

---

## Example: Agent-SRE with Linear Integration

Here's a complete example of how `agent-sre` could use Linear:

```python
# src/sre_agent/linear_handler.py
import asyncio
from linear_client import LinearClient, create_agent_issue
from typing import Dict, List, Optional

class LinearHandler:
    """Handle Linear operations for agent-sre."""
    
    def __init__(self):
        self.client = LinearClient()
        self.team_key = "BVL"  # Your team key
    
    async def create_alert_ticket(self, alert: Dict) -> str:
        """Create a Linear ticket for a Prometheus alert."""
        severity_map = {
            "critical": 1,  # Urgent
            "warning": 2,   # High
            "info": 3       # Normal
        }
        
        priority = severity_map.get(alert.get("severity", "info"), 3)
        
        title = f"[Alert] {alert.get('title', 'Unknown Alert')}"
        description = f"""
**Alert Details:**
- Severity: {alert.get('severity', 'unknown')}
- Status: {alert.get('status', 'unknown')}
- Started: {alert.get('startsAt', 'unknown')}

**Description:**
{alert.get('description', 'No description')}

**Labels:**
{alert.get('labels', {})}

---
*Created by agent-sre*
        """.strip()
        
        return await create_agent_issue(
            title=title,
            description=description,
            team_key=self.team_key,
            agent_name="agent-sre",
            priority=priority
        )
    
    async def track_flux_reconciliation(self, resource: str, status: str) -> Optional[str]:
        """Create or update Linear issue for Flux reconciliation."""
        # Check if issue already exists
        issues = await self.client.list_issues(
            team=self.team_key,
            limit=50
        )
        
        # Look for existing issue about this resource
        existing = None
        for issue in issues:
            if resource in issue.get("title", "") and "Flux" in issue.get("title", ""):
                existing = issue
                break
        
        if existing:
            # Update existing issue
            await self.client.create_comment(
                issue_id=existing["id"],
                body=f"**Status Update**: {status}\n\nResource: `{resource}`"
            )
            return existing["url"]
        else:
            # Create new issue
            issue = await self.client.create_issue(
                title=f"[Flux] Reconciliation issue: {resource}",
                description=f"""
**Resource**: `{resource}`
**Status**: {status}

This issue tracks Flux reconciliation problems for the above resource.

---
*Created by agent-sre*
                """.strip(),
                team_id=(await self.client.get_team(self.team_key))["id"],
                priority=2  # High
            )
            return issue["url"]
    
    async def list_open_issues(self) -> List[Dict]:
        """List all open issues for the team."""
        issues = await self.client.list_issues(
            team=self.team_key,
            state="In Progress",  # Or "Todo", "Backlog", etc.
            limit=50
        )
        return issues
```

---

## Testing

### Unit Tests (No API Key Required)

The Linear client includes comprehensive unit tests with mocked responses:

```bash
cd flux/ai/shared-libs/linear_client
pytest test_linear_client.py -v

# Or use make
make test
```

### Integration Tests (Requires API Key)

Test against the real Linear API:

```bash
cd flux/ai/shared-libs/linear_client

# Source your .zshrc to get LINEAR_API_KEY
source ~/.zshrc

# Run full integration test suite
python3 test_integration.py

# Or quick connection test
python3 test_integration.py --quick

# Or use make
make test-integration
make test-quick
```

---

## Security Best Practices

1. **Use Kubernetes Secrets**: Never hardcode API keys in code or config files
2. **Namespace Isolation**: Create secrets in the `ai` namespace where agents run
3. **RBAC**: Limit which service accounts can access the Linear secret
4. **Key Rotation**: Rotate API keys regularly in Linear settings
5. **Monitor Usage**: Check Linear dashboard for API usage and anomalies

### Example RBAC for Secret Access

```yaml
# Allow only specific service accounts to read Linear secret
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: linear-api-reader
  namespace: ai
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["linear-api-key"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: linear-api-reader-binding
  namespace: ai
subjects:
- kind: ServiceAccount
  name: agent-sre-sa
  namespace: ai
roleRef:
  kind: Role
  name: linear-api-reader
  apiGroup: rbac.authorization.k8s.io
```

---

## Troubleshooting

### Agent can't connect to Linear

1. **Check secret exists**:
   ```bash
   kubectl get secret linear-api-key -n ai
   ```

2. **Verify API key is set**:
   ```bash
   kubectl get secret linear-api-key -n ai -o jsonpath='{.data.api-key}' | base64 -d
   ```

3. **Test API key locally**:
   ```bash
   export LINEAR_API_KEY=$(kubectl get secret linear-api-key -n ai -o jsonpath='{.data.api-key}' | base64 -d)
   cd flux/ai/shared-libs/linear_client
   python3 test_integration.py --quick
   ```

4. **Check network policies** allow egress to `api.linear.app:443`

5. **Check agent logs**:
   ```bash
   kubectl logs -n ai -l app=agent-sre --tail=100
   ```

### Common Errors

**"Linear API key required"**
- Ensure `LINEAR_API_KEY` env var is set in agent config
- Verify secret exists and is mounted correctly

**"Team with key 'XXX' not found"**
- Verify team key is correct (case-sensitive)
- List teams: `await client.list_teams()` to see available keys

**"HTTP 400 Bad Request"**
- Check GraphQL query syntax
- Verify team IDs and other UUIDs are correct
- See error details in exception message

---

## Next Steps

1. ✅ **Create Linear API secret** in Kubernetes
2. ✅ **Add Linear client** to your agent's codebase
3. ✅ **Add `httpx` dependency** to `requirements.txt`
4. ✅ **Configure agent** to use Linear API key from secret
5. ✅ **Test integration** with a simple issue creation
6. ✅ **Integrate into agent workflows** (alerts, incidents, etc.)

---

## References

- **Linear Client Library**: `flux/ai/shared-libs/linear_client/`
- **Linear API Documentation**: https://developers.linear.app/docs
- **Linear GraphQL API**: https://developers.linear.app/docs/graphql/working-with-the-graphql-api
- **Linear API Explorer**: https://linear.app/settings/api (create/manage API keys)
