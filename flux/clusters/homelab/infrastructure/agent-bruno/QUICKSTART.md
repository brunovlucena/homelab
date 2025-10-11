# 🚀 Agent Bruno - Quick Start Guide

## 🎯 What is Agent Bruno?

Agent Bruno is an AI assistant that:
- 📚 Has deep knowledge of your homepage application
- 🧠 Remembers conversations per IP address
- 💾 Uses Redis for recent context (24h) and MongoDB for long-term memory
- 🤖 Powered by Ollama/LLM at 192.168.0.16:11434

## 🏃 Quick Deploy to Kubernetes

### Prerequisites

1. Kubernetes cluster running
2. Flux CD installed
3. Bitnami Helm repository configured

### Deploy Everything

```bash
# Navigate to homelab directory
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab

# Commit and push (Flux will auto-deploy)
git add infrastructure/
git commit -m "feat: add agent-bruno with redis and mongodb"
git push

# Monitor deployment
watch kubectl get pods -n redis
watch kubectl get pods -n mongodb
watch kubectl get pods -n agent-bruno
```

### Verify Deployment

```bash
# Check Redis
kubectl get pods -n redis
kubectl logs -n redis -l app.kubernetes.io/name=redis

# Check MongoDB
kubectl get pods -n mongodb
kubectl logs -n mongodb -l app.kubernetes.io/name=mongodb

# Check Agent Bruno
kubectl get pods -n agent-bruno
kubectl logs -n agent-bruno -l app=agent-bruno

# Check all services
kubectl get svc -n redis
kubectl get svc -n mongodb
kubectl get svc -n agent-bruno
```

## 🧪 Test Agent Bruno

### Port Forward to Local Machine

```bash
# Forward agent-bruno service
kubectl port-forward -n agent-bruno svc/agent-bruno-service 8080:8080
```

### Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Ready check
curl http://localhost:8080/ready

# Get knowledge summary
curl http://localhost:8080/knowledge/summary

# Chat with agent
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How do I deploy the homepage application?"
  }'

# Get memory stats (replace with your IP)
curl http://localhost:8080/memory/192.168.1.100

# Search knowledge
curl "http://localhost:8080/knowledge/search?q=deployment"

# Get system stats
curl http://localhost:8080/stats
```

## 🔌 Use via Homepage API

Once the homepage API is updated and redeployed:

```bash
# Port forward homepage API
kubectl port-forward -n homepage svc/homepage-api-service 8080:8080

# Chat through homepage API
curl -X POST http://localhost:8080/api/v1/agent-bruno/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What API endpoints are available?"
  }'

# Get memory through homepage API
curl http://localhost:8080/api/v1/agent-bruno/memory/192.168.1.100

# Clear memory through homepage API
curl -X DELETE http://localhost:8080/api/v1/agent-bruno/memory/192.168.1.100
```

## 🐳 Local Development with Docker Compose

```bash
# Navigate to agent-bruno directory
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f agent-bruno

# Test
curl http://localhost:8080/health

# Stop
docker-compose down
```

## 💻 Local Development without Docker

```bash
# Navigate to agent-bruno directory
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno

# Install dependencies
make install

# Start Redis and MongoDB locally
make redis-local
make mongodb-local

# Run agent in development mode
make dev

# In another terminal, test
make test-api
make test-chat

# Stop local services when done
make stop-local
```

## 📊 Monitor with Prometheus/Grafana

### Metrics Available

- `bruno_requests_total` - Total requests
- `bruno_request_duration_seconds` - Request duration
- `bruno_memory_operations_total` - Memory operations
- `bruno_active_sessions` - Active sessions

### Access Metrics

```bash
# Port forward agent-bruno
kubectl port-forward -n agent-bruno svc/agent-bruno-service 8080:8080

# View metrics
curl http://localhost:8080/metrics
```

ServiceMonitor is automatically created for Prometheus scraping.

## 🛠️ Troubleshooting

### Agent Bruno not starting

```bash
# Check logs
kubectl logs -n agent-bruno -l app=agent-bruno

# Check events
kubectl get events -n agent-bruno --sort-by='.lastTimestamp'

# Check if Redis is ready
kubectl get pods -n redis

# Check if MongoDB is ready
kubectl get pods -n mongodb
```

### Connection errors

```bash
# Test Redis connection from agent-bruno pod
kubectl exec -it -n agent-bruno deployment/agent-bruno -- sh
# Inside pod:
# python -c "import redis; r=redis.Redis(host='redis-master.redis.svc.cluster.local'); print(r.ping())"

# Test MongoDB connection
kubectl exec -it -n agent-bruno deployment/agent-bruno -- sh
# Inside pod:
# python -c "from pymongo import MongoClient; c=MongoClient('mongodb://mongodb.mongodb.svc.cluster.local:27017'); print(c.server_info())"
```

### Memory not persisting

```bash
# Check Redis persistence
kubectl exec -it -n redis statefulset/redis-master -- redis-cli
# Inside redis-cli:
# CONFIG GET dir
# CONFIG GET appendonly

# Check MongoDB data
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh
# Inside mongosh:
# use agent_bruno
# db.conversations.countDocuments()
```

## 🎨 Example Conversations

### Deployment Questions

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How do I deploy the homepage to Kubernetes?"
  }'
```

### API Questions

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What API endpoints are available for projects?"
  }'
```

### Architecture Questions

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Explain the homepage architecture"
  }'
```

## 📚 Next Steps

1. **Update Homepage API Image**
   - Rebuild homepage API with agent-bruno integration
   - Deploy updated API

2. **Frontend Integration**
   - Add chatbot UI component
   - Connect to `/api/v1/agent-bruno/chat`

3. **Custom Configuration**
   - Adjust session TTL
   - Configure memory limits
   - Add authentication (optional)

4. **Monitoring Setup**
   - Create Grafana dashboard
   - Set up alerts
   - Monitor memory usage

## 🆘 Getting Help

- **Logs:** `kubectl logs -n agent-bruno -l app=agent-bruno -f`
- **Status:** `kubectl get all -n agent-bruno`
- **Describe:** `kubectl describe pod -n agent-bruno <pod-name>`
- **Documentation:** See README.md and IMPLEMENTATION.md

## 🎉 You're Ready!

Agent Bruno is now deployed and ready to answer questions about your homepage application with full memory of past conversations per IP!

---

**Quick Commands Reference:**

```bash
# Deploy
git push  # Flux auto-deploys

# Test
kubectl port-forward -n agent-bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/health

# Monitor
kubectl logs -n agent-bruno -l app=agent-bruno -f

# Clean up (if needed)
kubectl delete namespace agent-bruno
kubectl delete namespace redis
kubectl delete namespace mongodb
```

