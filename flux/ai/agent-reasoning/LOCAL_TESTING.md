# Local Testing Guide

## Quick Start

### Option 1: Automated Test Script

```bash
# Run the automated test script
./test_local.sh
```

This will:
- Start the container
- Test all endpoints
- Show results
- Keep container running for further testing

### Option 2: Manual Testing

```bash
# Start the container
docker run -d \
  --name agent-reasoning-local \
  -p 8080:8080 \
  -e DEVICE=cpu \
  agent-reasoning:latest

# Wait a few seconds for startup
sleep 5

# Test health
curl http://localhost:8080/health | jq

# Test reasoning
curl -X POST http://localhost:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How should I optimize my Kubernetes cluster?",
    "max_steps": 6,
    "task_type": "optimization"
  }' | jq
```

### Option 3: Python Test Script

```bash
# Install requests if needed
pip install requests

# Run comprehensive tests
python3 test_requests.py
```

## Available Endpoints

### 1. Health Check
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "model_loaded": true,
  "device": "cpu",
  "gpu_available": false
}
```

### 2. Root Endpoint
```bash
curl http://localhost:8080/
```

Shows service information and available endpoints.

### 3. Reasoning Endpoint
```bash
curl -X POST http://localhost:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Your question here",
    "context": {"key": "value"},
    "max_steps": 6,
    "task_type": "optimization"
  }'
```

**Task Types:**
- `planning` - For planning tasks
- `optimization` - For optimization problems
- `troubleshooting` - For debugging/analysis
- `logic` - For logic puzzles
- `general` - General reasoning

### 4. Metrics Endpoint
```bash
curl http://localhost:8080/metrics
```

Returns Prometheus metrics.

### 5. Ready Check
```bash
curl http://localhost:8080/ready
```

## Example Test Scenarios

### Scenario 1: Infrastructure Planning
```bash
curl -X POST http://localhost:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How should I deploy this application across my clusters?",
    "context": {
      "clusters": ["studio", "pro"],
      "app_requirements": {"cpu": "2", "memory": "4Gi"}
    },
    "max_steps": 6,
    "task_type": "planning"
  }' | jq
```

### Scenario 2: Troubleshooting
```bash
curl -X POST http://localhost:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Why is my service slow?",
    "context": {
      "metrics": {"cpu": "80%", "memory": "90%"},
      "logs": ["error connecting to database"]
    },
    "max_steps": 4,
    "task_type": "troubleshooting"
  }' | jq
```

### Scenario 3: Optimization
```bash
curl -X POST http://localhost:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Optimize my Kubernetes resource allocation",
    "context": {
      "current_allocation": {"cpu": "100", "memory": "200Gi"},
      "workloads": 50
    },
    "max_steps": 8,
    "task_type": "optimization"
  }' | jq
```

## Container Management

### View Logs
```bash
docker logs -f agent-reasoning-local
```

### Stop Container
```bash
docker stop agent-reasoning-local
docker rm agent-reasoning-local
```

### Restart Container
```bash
docker restart agent-reasoning-local
```

### Run in Interactive Mode
```bash
docker run -it --rm \
  -p 8080:8080 \
  -e DEVICE=cpu \
  agent-reasoning:latest \
  /bin/bash
```

## Testing with Different Configurations

### CPU Mode (Default)
```bash
docker run -d \
  --name agent-reasoning-local \
  -p 8080:8080 \
  -e DEVICE=cpu \
  agent-reasoning:latest
```

### GPU Mode (if available)
```bash
docker run -d \
  --name agent-reasoning-local \
  -p 8080:8080 \
  --gpus all \
  -e DEVICE=cuda \
  agent-reasoning:latest
```

### Custom Model Path
```bash
docker run -d \
  --name agent-reasoning-local \
  -p 8080:8080 \
  -v /path/to/models:/models \
  -e MODEL_PATH=/models/trm-checkpoint.pth \
  agent-reasoning:latest
```

## Expected Behavior

### Successful Response
- Status code: 200
- Response includes:
  - `answer`: The reasoned answer
  - `steps`: Number of reasoning steps used
  - `confidence`: Confidence score (0.0-1.0)
  - `reasoning_trace`: Step-by-step reasoning process
  - `duration_ms`: Processing time

### Error Responses
- 400: Invalid request (check JSON format)
- 500: Server error (check logs)
- 503: Service not ready (wait a few seconds)

## Debugging

### Check Service Status
```bash
# Health check
curl http://localhost:8080/health

# Ready check
curl http://localhost:8080/ready

# View logs
docker logs agent-reasoning-local
```

### Common Issues

1. **Port already in use**
   ```bash
   # Use different port
   docker run -p 8081:8080 agent-reasoning:latest
   ```

2. **Service not responding**
   ```bash
   # Check if container is running
   docker ps | grep agent-reasoning
   
   # Check logs
   docker logs agent-reasoning-local
   ```

3. **Model not loading**
   - This is expected if no checkpoint is provided
   - Service will still respond with mock reasoning
   - For real TRM, mount model checkpoint

## Next Steps

After local testing:
1. ✅ Verify all endpoints work
2. ✅ Test with different question types
3. ✅ Check metrics are being collected
4. ⬜ Integrate with actual TRM model
5. ⬜ Deploy to Kubernetes

## Tips

- Use `jq` for pretty JSON output: `curl ... | jq`
- Use `-f` flag with docker logs for real-time: `docker logs -f`
- Test with different `max_steps` values (1-20)
- Try different `task_type` values for varied responses


