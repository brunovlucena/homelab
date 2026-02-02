# 🚶 First Steps with Knative Lambda

**Deploy your first serverless function and understand the platform**

---

## 🎯 What You'll Learn

After completing this guide, you will:

- ✅ Upload code to S3
- ✅ Trigger an automatic build
- ✅ Monitor the build process
- ✅ Test your deployed function
- ✅ Observe auto-scaling in action
- ✅ Understand the complete workflow

**Time Required**: ~15 minutes

---

## Prerequisites

- ✅ Knative Lambda installed ([Installation Guide](INSTALLATION.md))
- ✅ kubectl configured and working
- ✅ AWS CLI configured
- ✅ S3 bucket access

---

## Step 1: Create Your First Function

### Python Example (Recommended)

```python
# parser.py
import json
import os
from datetime import datetime

def handler(event):
    """
    Simple serverless function that processes CloudEvents
    
    Args:
        event: CloudEvent data (dict)
    
    Returns:
        dict: Response data
    """
    # Extract CloudEvent fields
    source = event.get('source', 'unknown')
    event_type = event.get('type', 'unknown')
    data = event.get('data', {})
    
    # Your business logic here
    response = {
        'status': 'success',
        'timestamp': datetime.utcnow().isoformat(),
        'message': 'Function executed successfully! 🚀',
        'received_from': source,
        'event_type': event_type,
        'environment': os.getenv('ENVIRONMENT', 'development'),
        'data': data
    }
    
    print(f"Processed event from {source}: {event_type}")
    
    return response
```

**With Dependencies:**

```python
# parser.py
import requests
from datetime import datetime

def handler(event):
    """Function with external dependencies"""
    # Call external API
    response = requests.get('https://api.github.com/repos/kubernetes/kubernetes')
    
    return {
        'status': 'success',
        'stars': response.json()['stargazers_count'],
        'timestamp': datetime.utcnow().isoformat()
    }
```

```bash
# requirements.txt
requests==2.31.0
```

### Node.js Example

```javascript
// index.js
exports.handler = async (event) => {
    console.log('Received event:', JSON.stringify(event, null, 2));
    
    return {
        statusCode: 200,
        body: JSON.stringify({
            message: 'Hello from Node.js Lambda!',
            timestamp: new Date().toISOString(),
            event: event
        })
    };
};
```

```json
// package.json
{
    "name": "knative-lambda-function",
    "version": "1.0.0",
    "main": "index.js",
    "dependencies": {
        "axios": "^1.6.0"
    }
}
```

### Go Example

```go
// main.go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
)

type Event struct {
    Source string                 `json:"source"`
    Type   string                 `json:"type"`
    Data   map[string]interface{} `json:"data"`
}

type Response struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Message   string    `json:"message"`
}

func handler(event Event) Response {
    log.Printf("Processing event: %+v", event)
    
    return Response{
        Status:    "success",
        Timestamp: time.Now().UTC(),
        Message:   "Hello from Go Lambda! 🚀",
    }
}

func main() {
    // CloudEvent processing logic here
}
```

---

## Step 2: Upload Code to S3

```bash
# Set environment (dev, stg, or prd)
export ENV=dev

# Generate unique parser ID
export PARSER_ID="my-first-function-$(uuidgen | tr '[:upper:]' '[:lower:]')"
echo "Parser ID: ${PARSER_ID}"

# Upload Python function
aws s3 cp parser.py \
  s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}/

# Upload dependencies (if any)
aws s3 cp requirements.txt \
  s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}/

# Verify upload
aws s3 ls s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}/
```

**Expected output:**
```
2025-10-29 10:30:00        512 parser.py
2025-10-29 10:30:01         45 requirements.txt
```

---

## Step 3: Trigger Build

### Option A: Using Test Script (Recommended)

```bash
# Navigate to tests directory
cd /path/to/knative-lambda/tests

# Trigger build
ENV=${ENV} uv run --python 3.9 python create-event-builder.py

# Follow prompts to create CloudEvent
```

### Option B: Manual CloudEvent

```bash
# Create CloudEvent JSON
cat > build-event.json <<EOF
{
  "specversion": "1.0",
  "type": "build.start",
  "source": "manual-trigger",
  "id": "$(uuidgen)",
  "time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "datacontenttype": "application/json",
  "data": {
    "parser_id": "${PARSER_ID}",
    "s3_prefix": "global/parser/${PARSER_ID}/",
    "language": "python",
    "runtime": "python3.9",
    "s3_bucket": "knative-lambda-fusion-modules-tmp",
    "ecr_repository": "339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-functions"
  }
}
EOF

# Send to RabbitMQ (requires port-forward)
make pf-rabbitmq

# In another terminal:
curl -X POST http://localhost:15672/api/exchanges/%2F/knative-broker/publish \
  -u guest:guest \
  -H "Content-Type: application/json" \
  -d @build-event.json
```

---

## Step 4: Monitor Build Progress

### Watch Kaniko Job

```bash
# Watch job creation
kubectl get jobs -n knative-lambda -w

# Should see:
# kaniko-<parser-id>   0/1   0s
# kaniko-<parser-id>   0/1   5s
# kaniko-<parser-id>   1/1   120s
```

### View Build Logs

```bash
# Get Kaniko job name
export JOB_NAME=$(kubectl get jobs -n knative-lambda \
  -l parser-id=${PARSER_ID} -o jsonpath='{.items[0].metadata.name}')

# Stream logs
kubectl logs -f job/${JOB_NAME} -n knative-lambda
```

**Expected log output:**
```
INFO[0000] Retrieving image manifest python:3.9-slim
INFO[0005] Downloading base image python:3.9-slim
INFO[0015] Executing 0 build triggers
INFO[0015] Unpacking rootfs...
INFO[0030] COPY parser.py /app/
INFO[0030] COPY requirements.txt /app/
INFO[0031] RUN pip install --no-cache-dir -r requirements.txt
INFO[0045] Building image...
INFO[0120] Pushing image to 339954290315.dkr.ecr...
INFO[0125] Successfully pushed image
```

### Monitor Builder Service

```bash
# Stream builder logs
kubectl logs -f deployment/knative-lambda-builder -n knative-lambda
```

**Key log messages:**
- ✅ `Received build.start event`
- ✅ `Creating Kaniko build job`
- ✅ `Build job created successfully`
- ✅ `Build completed successfully`
- ✅ `Creating Knative Service`
- ✅ `Knative Service created`

---

## Step 5: Verify Deployment

### Check Knative Service

```bash
# List Knative Services
kubectl get ksvc -n knative-lambda

# Expected output:
# NAME                   URL                                              READY
# parser-abc123-...      http://parser-abc123....svc.cluster.local      True

# Get service details
kubectl get ksvc ${PARSER_ID} -n knative-lambda -o yaml
```

### Get Function URL

```bash
# Extract URL
export FUNCTION_URL=$(kubectl get ksvc ${PARSER_ID} \
  -n knative-lambda \
  -o jsonpath='{.status.url}')

echo "Function URL: ${FUNCTION_URL}"
```

### Check Pods

```bash
# Initially, should scale to zero
kubectl get pods -n knative-lambda -l serving.knative.dev/service=${PARSER_ID}

# Output: No resources found (scaled to zero)
```

---

## Step 6: Test Your Function

### Send Test Request

```bash
# Send CloudEvent
curl -X POST ${FUNCTION_URL} \
  -H "Content-Type: application/json" \
  -H "ce-id: test-$(uuidgen)" \
  -H "ce-source: manual-test" \
  -H "ce-type: test.event" \
  -H "ce-specversion: 1.0" \
  -d '{
    "message": "Hello Knative Lambda!",
    "test_data": {"key": "value"}
  }'
```

**Expected response:**
```json
{
  "status": "success",
  "timestamp": "2025-10-29T10:45:23.123456",
  "message": "Function executed successfully! 🚀",
  "received_from": "manual-test",
  "event_type": "test.event",
  "environment": "development",
  "data": {
    "message": "Hello Knative Lambda!",
    "test_data": {"key": "value"}
  }
}
```

### Watch Pod Spin Up

```bash
# In another terminal, watch pods
kubectl get pods -n knative-lambda \
  -l serving.knative.dev/service=${PARSER_ID} -w

# You'll see:
# 1. Pod created (from 0 pods)
# 2. Pod running (Ready 2/2)
# 3. After 30s idle, pod terminating
# 4. Back to 0 pods
```

---

## Step 7: Observe Auto-Scaling

### Generate Load

```bash
# Send 20 requests rapidly
for i in {1..20}; do
  curl -s -X POST ${FUNCTION_URL} \
    -H "Content-Type: application/json" \
    -H "ce-id: load-test-$i" \
    -H "ce-source: load-test" \
    -H "ce-type: test.event" \
    -H "ce-specversion: 1.0" \
    -d '{"iteration": '$i'}' &
done
wait

echo "Load test complete"
```

### Watch Scaling

```bash
# Watch pods scale up
kubectl get pods -n knative-lambda \
  -l serving.knative.dev/service=${PARSER_ID} -w

# Expected behavior:
# 0 pods → 1 pod → 2 pods → 3 pods (under load)
# After idle (30s): 3 pods → 2 pods → 1 pod → 0 pods
```

### Check Metrics

```bash
# Port-forward Prometheus
kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n prometheus

# Open http://localhost:9090
# Query: rate(serving_revision_request_count{revision=~"${PARSER_ID}.*"}[1m])
```

---

## Step 8: View Logs

### Function Logs

```bash
# Get function logs
kubectl logs -l serving.knative.dev/service=${PARSER_ID} \
  -n knative-lambda \
  --tail=50

# Expected:
# Processed event from manual-test: test.event
# Processed event from load-test: test.event
```

### Structured Logging Example

```python
# parser.py with logging
import logging
import json

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def handler(event):
    logger.info("Processing event", extra={
        'source': event.get('source'),
        'type': event.get('type'),
        'parser_id': os.getenv('PARSER_ID')
    })
    
    # ... function logic
```

---

## Step 9: Update Your Function

### Modify Code

```python
# parser.py - updated version
def handler(event):
    return {
        'status': 'success',
        'message': 'Updated function! Version 2.0 🚀',
        'version': '2.0',
        'data': event.get('data', {})
    }
```

### Trigger Rebuild

```bash
# Upload updated code
aws s3 cp parser.py \
  s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}/

# Trigger new build
cd /path/to/knative-lambda/tests
ENV=${ENV} uv run --python 3.9 python create-event-builder.py

# Monitor build
kubectl get jobs -n knative-lambda -w
```

### Test Updated Function

```bash
# Send test request
curl -X POST ${FUNCTION_URL} \
  -H "Content-Type: application/json" \
  -H "ce-id: test-v2-$(uuidgen)" \
  -H "ce-source: manual-test" \
  -H "ce-type: test.event" \
  -H "ce-specversion: 1.0" \
  -d '{"test": "version 2"}'

# Expected: "message": "Updated function! Version 2.0 🚀"
```

---

## Step 10: Clean Up (Optional)

```bash
# Delete Knative Service
kubectl delete ksvc ${PARSER_ID} -n knative-lambda

# Delete S3 files
aws s3 rm s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}/ --recursive

# Delete ECR image (optional)
aws ecr batch-delete-image \
  --repository-name knative-lambda-functions \
  --image-ids imageTag=${PARSER_ID} \
  --region us-west-2

# Clean up build jobs (optional)
kubectl delete job -l parser-id=${PARSER_ID} -n knative-lambda
```

---

## 🎓 What You Learned

✅ **Build Process**: Code → S3 → CloudEvent → Kaniko → ECR → Knative Service  
✅ **Auto-Scaling**: Scale from 0→N→0 based on traffic  
✅ **CloudEvents**: Standards-based event processing  
✅ **Monitoring**: Logs, metrics, and observability  
✅ **Updates**: Modify and redeploy functions

---

## 📚 Next Steps

### Learn More

| Topic | Link |
|-------|------|
| **CloudEvents format** | [Backend Guide](../03-for-engineers/backend/README.md) |
| **Multi-language support** | [Build Pipeline](../04-architecture/BUILD_PIPELINE.md) |
| **Production deployment** | [DevOps Guide](../03-for-engineers/devops/README.md) |
| **Monitoring & alerts** | [Observability](../04-architecture/OBSERVABILITY.md) |

### Advanced Topics

- **Event-driven workflows**: [CloudEvents Integration](../03-for-engineers/backend/user-stories/BACKEND-001-cloudevents-processing.md)
- **Rate limiting**: [Resilience Guide](../03-for-engineers/sre/user-stories/SRE-005-autoscaling-optimization.md)
- **Multi-environment**: [DevOps Multi-Env](../03-for-engineers/devops/user-stories/DEVOPS-003-multi-environment.md)

---

## ❓ Common Questions

**Q: How long does a build take?**  
A: Typically 60-180 seconds depending on dependencies and base image size.

**Q: Can I use private PyPI packages?**  
A: Yes! Configure pip credentials in your build context or use a private PyPI mirror.

**Q: What languages are supported?**  
A: Python, Node.js, and Go are supported out-of-the-box. See [Multi-Language Strategy](../07-decisions/MULTI_LANGUAGE_STRATEGY.md).

**Q: How do I debug build failures?**  
A: Check Kaniko job logs: `kubectl logs job/kaniko-<job-name>`

**More questions**: [FAQ](FAQ.md)

---

**Congratulations!** 🎉 You've successfully deployed and tested your first Knative Lambda function!

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

