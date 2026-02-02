# 🔧 Backend Developer - Knative Lambda

**Function development and integration**

---

## 🎯 Overview

As a backend developer working with Knative Lambda, you'll create serverless functions that process CloudEvents, integrate with external APIs, and scale automatically. This guide covers function development, testing, debugging, and best practices.

---

## 🚀 Quick Start

### 1. Create a Function

```python
# parser.py - Basic function template
import json
import os
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process CloudEvent and return result
    
    Args:
        event: CloudEvent data
        
    Returns:
        Processing result
    """
    # Extract event data
    parser_id = event.get('parser_id')
    third_party_id = event.get('third_party_id')
    data = event.get('data', {})
    
    # Business logic
    result = process_data(data)
    
    # Return result
    return {
        'status': 'success',
        'parser_id': parser_id,
        'result': result
    }

def process_data(data: Dict[str, Any]) -> Dict[str, Any]:
    """Your business logic here"""
    return {'processed': True}

if __name__ == "__main__":
    # For local testing
    test_event = {
        'parser_id': 'test-parser',
        'third_party_id': 'customer-123',
        'data': {'key': 'value'}
    }
    print(json.dumps(handler(test_event), indent=2))
```

### 2. Test Locally

```bash
# Run function locally
python parser.py

# Test with mock CloudEvent
cat <<EOF | python parser.py
{
  "parser_id": "test-parser",
  "third_party_id": "customer-123",
  "data": {"key": "value"}
}
EOF
```

### 3. Deploy Function

```bash
# 1. Upload parser to S3
export PARSER_ID="my-parser-$(uuidgen)"
aws s3 cp parser.py s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}

# 2. Trigger build
make trigger-build-dev PARSER_ID=${PARSER_ID}

# 3. Monitor build
kubectl get jobs -n knative-lambda -w

# 4. Test deployed function
curl -X POST https://parser-${PARSER_ID}.knative-lambda.homelab \
  -H "Content-Type: application/json" \
  -d '{"parser_id":"'${PARSER_ID}'","data":{"key":"value"}}'
```

---

## 📚 User Stories

| Story ID | Title | Priority | Status |
|----------|-------|----------|--------|
| **Backend-001** | [Function Development](user-stories/BACKEND-001-function-development.md) | P0 | ✅ |
| **Backend-002** | [Local Testing](user-stories/BACKEND-002-local-testing.md) | P0 | ✅ |
| **Backend-003** | [CloudEvents Integration](user-stories/BACKEND-003-cloudevents-integration.md) | P0 | ✅ |
| **Backend-004** | [Error Handling](user-stories/BACKEND-004-error-handling.md) | P1 | ✅ |
| **Backend-005** | [Performance Optimization](user-stories/BACKEND-005-performance-optimization.md) | P1 | ✅ |
| **Backend-006** | [Debugging Techniques](user-stories/BACKEND-006-debugging.md) | P1 | ✅ |

→ **[View All User Stories](user-stories/README.md)**

---

## 💡 Best Practices

### Function Structure
- Keep functions small (<500 lines)
- Use environment variables for configuration
- Return structured JSON responses
- Handle errors gracefully

### Performance
- Minimize cold start time (<3s)
- Cache expensive operations
- Use connection pooling for databases
- Optimize image size (<400MB)

### Security
- Validate all inputs
- Use least-privilege IAM roles
- Never log sensitive data
- Implement rate limiting

---

**Need help?** Join `#knative-lambda` on Slack or file a GitHub issue.

