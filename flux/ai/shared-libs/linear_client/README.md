# Linear Client for Homelab Agents

Simple Linear API client for use in homelab agents.

## Installation

```bash
pip install httpx
```

Or add to your agent's `requirements.txt`:
```
httpx>=0.25.0
```

## Testing

### Unit Tests (Mocked - No API Key Required)

```bash
# Run all unit tests
pytest test_linear_client.py -v

# Or use make
make test
```

### Integration Tests (Requires Real API Key)

```bash
# Set your Linear API key
export LINEAR_API_KEY=lin_api_xxxxxxxxxxxxx

# Run full integration test suite
python test_integration.py

# Or quick connection test
python test_integration.py --quick

# Or use make
make test-integration
make test-quick
```

### Test Coverage

```bash
pytest test_linear_client.py --cov=linear_client --cov-report=html
# Open htmlcov/index.html in browser

# Or use make
make test-coverage
```

## Usage

### Basic Example

```python
from linear_client import LinearClient

# Initialize client (reads LINEAR_API_KEY from env)
client = LinearClient()

# List issues
issues = await client.list_issues(team="SRE", assignee="me")
for issue in issues:
    print(f"{issue['identifier']}: {issue['title']}")

# Create issue
issue = await client.create_issue(
    title="Agent-generated issue",
    description="This was created by an agent",
    team_id="team-uuid-here",
    priority=2  # High priority
)
print(f"Created: {issue['url']}")
```

### Quick Helper

```python
from linear_client import create_agent_issue

# Quick issue creation
url = await create_agent_issue(
    title="System alert: High CPU",
    description="CPU usage exceeded 90% for 5 minutes",
    team_key="SRE",
    agent_name="agent-sre",
    priority=1  # Urgent
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
    - name: LINEAR_API_KEY
      valueFrom:
        secretKeyRef:
          name: linear-api-key
          key: api-key
```

## API Reference

See `linear_client.py` for full API documentation.

