# Jira Client for Homelab Agents

Simple Jira API client for use in homelab agents.

## Installation

```bash
pip install httpx
```

Or add to your agent's `requirements.txt`:
```
httpx>=0.25.0
```

## Configuration

The client requires the following environment variables:

- `JIRA_URL`: Your Jira instance URL (e.g., "https://your-domain.atlassian.net")
- `JIRA_EMAIL`: Your Jira user email
- `JIRA_API_TOKEN`: Your Jira API token

## Usage

### Basic Example

```python
from jira_client import JiraClient

# Initialize client (reads from environment)
client = JiraClient()

# Create issue
issue = await client.create_issue(
    project_key="PROJ",
    summary="Agent-generated issue",
    description="This was created by an agent",
    issue_type="Task",
    priority="High"
)
print(f"Created: {issue['key']}")

# Get issue
issue = await client.get_issue("PROJ-123")

# Search issues
issues = await client.search_issues("project = PROJ AND status = Open")
```

### Quick Helper

```python
from jira_client import create_agent_issue

# Quick issue creation
issue_key = await create_agent_issue(
    project_key="PROJ",
    summary="System alert: High CPU",
    description="CPU usage exceeded 90% for 5 minutes",
    agent_name="agent-sre",
    priority="High"
)
```

### In Kubernetes Agent

```yaml
# lambdaagent.yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: agent-sre
spec:
  env:
    - name: JIRA_URL
      value: "https://your-domain.atlassian.net"
    - name: JIRA_EMAIL
      valueFrom:
        secretKeyRef:
          name: jira-credentials
          key: email
    - name: JIRA_API_TOKEN
      valueFrom:
        secretKeyRef:
          name: jira-credentials
          key: api-token
```

## API Reference

See `jira_client.py` for full API documentation.

