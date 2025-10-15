# 🤖 Agent Bruno Runbooks

This directory contains operational runbooks for troubleshooting and maintaining Agent Bruno, the AI assistant with IP-based conversation memory.

## 📚 Available Runbooks

### Critical Issues

1. **[🚨 API Down](./api-down.md)**
   - **Severity:** Critical
   - **When:** Agent Bruno API is completely unavailable
   - **Symptoms:** All endpoints returning errors or timing out
   - **Impact:** Complete service outage

2. **[🔄 Pod Crash Loop](./pod-crash-loop.md)**
   - **Severity:** Critical
   - **When:** Pods repeatedly crashing (CrashLoopBackOff)
   - **Symptoms:** Continuous restarts, service unavailable
   - **Impact:** Complete service outage

### High Priority Issues

3. **[🔴 Redis Connection Issues](./redis-connection-issues.md)**
   - **Severity:** High
   - **When:** Cannot connect to Redis session storage
   - **Symptoms:** Session memory failures, no conversation context
   - **Impact:** Loss of recent conversation history

4. **[🍃 MongoDB Connection Issues](./mongodb-connection-issues.md)**
   - **Severity:** High
   - **When:** Cannot connect to MongoDB persistent storage
   - **Symptoms:** Cannot save/retrieve conversation history
   - **Impact:** Loss of long-term conversation persistence

5. **[🤖 Ollama/LLM Connection Issues](./ollama-connection-issues.md)**
   - **Severity:** Critical
   - **When:** Cannot connect to Ollama LLM server
   - **Symptoms:** Chat functionality completely broken
   - **Impact:** No AI responses possible

### Performance Issues

6. **[🧠 High Memory Usage](./high-memory-usage.md)**
   - **Severity:** Warning
   - **When:** Memory usage approaching or exceeding limits
   - **Symptoms:** Slow performance, potential OOMKills
   - **Impact:** Service degradation, possible restarts

7. **[⏱️ High Response Time](./high-response-time.md)**
   - **Severity:** Warning
   - **When:** Responses taking >5 seconds
   - **Symptoms:** Slow API responses, poor user experience
   - **Impact:** Degraded user experience

## 🔍 Quick Diagnostic Guide

### First Steps for Any Issue

```bash
# 1. Check pod status
kubectl get pods -n bruno -l app=agent-bruno

# 2. Check logs
kubectl logs -n bruno -l app=agent-bruno --tail=100

# 3. Check events
kubectl get events -n bruno --sort-by='.lastTimestamp' | head -20

# 4. Check resource usage
kubectl top pods -n bruno -l app=agent-bruno
```

### Quick Health Checks

```bash
# Port forward to service
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Test chat
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'
```

## 🏗️ Agent Bruno Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Bruno                             │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  FastAPI     │    │    Memory    │    │  Knowledge   │   │
│  │  Server      │◄──►│   Manager    │◄──►│    Base      │   │
│  │              │    │              │    │              │   │
│  └──────┬───────┘    └──────┬───────┘    └──────────────┘   │
│         │                   │                               │
│         │                   ▼                               │
│         │            ┌──────────────┐                       │
│         │            │    Redis     │                       │
│         │            │   (Session)  │                       │
│         │            └──────┬───────┘                       │
│         │                   │                               │
│         │                   ▼                               │
│         │            ┌──────────────┐                       │
│         │            │   MongoDB    │                       │
│         │            │  (Persistent)│                       │
│         │            └──────────────┘                       │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────┐                                           │
│  │ Ollama LLM   │                                           │
│  │192.168.0.16  │                                           │
│  └──────────────┘                                           │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Component Dependencies

### Critical Dependencies
- **Redis** (`redis-master.redis.svc.cluster.local:6379`)
  - Session storage with 24h TTL
  - Recent conversation context
  
- **MongoDB** (`mongodb.mongodb.svc.cluster.local:27017`)
  - Persistent conversation history
  - Long-term memory storage

- **Ollama** (`http://192.168.0.16:11434`)
  - LLM for AI responses
  - Required for chat functionality

### Configuration
- **Namespace:** `bruno`
- **Port:** `8080`
- **Replicas:** `1-3` (HPA enabled)
- **Memory Limit:** `1Gi`
- **CPU Limit:** `1000m`

## 📊 Monitoring

### Key Metrics

```promql
# Request rate
rate(bruno_requests_total[5m])

# Response time (95th percentile)
histogram_quantile(0.95, bruno_request_duration_seconds_bucket)

# Error rate
rate(bruno_requests_total{status=~"5.."}[5m])

# Active sessions
bruno_active_sessions

# Memory operations
rate(bruno_memory_operations_total[5m])
```

### Health Endpoints

- `GET /health` - Basic health check
- `GET /ready` - Readiness (checks Redis/MongoDB)
- `GET /metrics` - Prometheus metrics
- `GET /stats` - System statistics

## 🚀 Common Operations

### Scale Up/Down

```bash
# Manual scaling
kubectl scale deployment agent-bruno -n bruno --replicas=3

# Check HPA status
kubectl get hpa -n bruno agent-bruno-hpa
```

### Restart Service

```bash
# Rolling restart
kubectl rollout restart deployment/agent-bruno -n bruno

# Check rollout status
kubectl rollout status deployment/agent-bruno -n bruno
```

### View Logs

```bash
# Recent logs
kubectl logs -n bruno -l app=agent-bruno --tail=100

# Follow logs
kubectl logs -n bruno -l app=agent-bruno -f

# Previous container logs (if crashed)
kubectl logs -n bruno -l app=agent-bruno --previous
```

### Check Dependencies

```bash
# Redis
kubectl get pods -n redis
kubectl exec -it -n redis statefulset/redis-master -- redis-cli ping

# MongoDB
kubectl get pods -n mongodb
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.adminCommand('ping')"

# Ollama
curl http://192.168.0.16:11434/api/version
```

## 📞 Escalation Path

1. **Level 1:** Check runbook for specific issue
2. **Level 2:** Review logs and metrics
3. **Level 3:** Check dependencies (Redis, MongoDB, Ollama)
4. **Level 4:** Contact infrastructure team
5. **Level 5:** Contact development team

## 🔗 Related Documentation

- [Agent Bruno README](../../../flux/clusters/homelab/infrastructure/agent-bruno/README.md)
- [Agent Bruno Implementation](../../../flux/clusters/homelab/infrastructure/agent-bruno/IMPLEMENTATION.md)
- [Agent Bruno Quick Start](../../../flux/clusters/homelab/infrastructure/agent-bruno/QUICKSTART.md)
- [Homepage Runbooks](../homepage/)
- [Homelab Architecture](../../../ARCHITECTURE.md)

## 📝 Runbook Maintenance

These runbooks should be updated when:
- New failure modes are discovered
- Configuration changes are made
- Dependencies are updated
- New monitoring is added
- Incident post-mortems reveal gaps

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Maintained by:** SRE Team

---

## 🆘 Emergency Contacts

For critical issues outside of runbook scope:
- **Infrastructure Team:** Check on-call schedule
- **Development Team:** Check on-call schedule
- **Escalation:** Follow standard incident response procedures

## 📖 Additional Resources

- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Redis Operations](https://redis.io/docs/management/)
- [MongoDB Operations](https://www.mongodb.com/docs/manual/administration/)
- [Ollama Documentation](https://ollama.ai/docs)

