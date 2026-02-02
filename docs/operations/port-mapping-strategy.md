# Port Mapping Strategy

> **Part of**: [Homelab Documentation](../README.md) → Operations  
> **Last Updated**: November 7, 2025

---

## Overview

Port ranges allocated across clusters to avoid conflicts and maintain consistency.

## Port Ranges

| Range | Purpose | Examples |
|-------|---------|----------|
| 80, 443 | Ingress | HTTP, HTTPS |
| 8080-8089 | Applications | APIs, web apps |
| 30000-30099 | Platform Services | Flux UI, Linkerd, cert-manager |
| 30100-30199 | AI/ML Services | VLLM, Ollama, Flyte, JupyterHub |
| 30200-30299 | AI Agents | agent-bruno, agent-auditor, agent-jamie |
| 30300-30399 | Serverless | Knative, lambda functions |
| 30400-30499 | Observability | Prometheus, Grafana, Loki, Tempo |
| 30500-30599 | Data Services | PostgreSQL, MongoDB, Redis, MinIO |
| 30600-30699 | Message Queues | RabbitMQ, Kafka |
| 30700-30799 | Ingress/LB | Cloudflare tunnel, nginx |

## Key Services

### Platform Services (30000-30099)

```yaml
Headlamp (Flux UI): 30001
Linkerd Dashboard: 30002
cert-manager: 30003
```

### AI/ML Services (30100-30199)

```yaml
# Forge Cluster
VLLM: 30100
Ollama: 30101
JupyterHub: 30102
Flyte Console: 30081
MLflow: 30103
```

### AI Agents (30200-30299)

```yaml
# Studio Cluster - AI Agents (User-facing)
agent-bruno: 30120
agent-auditor: 30121  # Legacy - being replaced by agent-auditor service below
agent-jamie: 30122
agent-devops: 30123
agent-mary-kay: 30127
aigoat: 30125

# Studio Cluster - Agent Auditor Service (Event-driven infrastructure response)
agent-auditor-api: 30210      # FastAPI main endpoint
agent-auditor-health: 30211   # Health checks
agent-auditor-events: 30212   # CloudEvents webhook
agent-auditor-mcp: 30213      # MCP protocol (IDE/CLI)
```

### Serverless Platform (30300-30399)

```yaml
# Knative Lambda (Pro/Studio Clusters)
knative-lambda-builder: 30300      # Builder service API
knative-lambda-metrics: 30301      # Metrics endpoint
knative-lambda-health: 30302       # Health checks
knative-lambda-webhook: 30303      # Build webhooks

# Knative Serving (system)
knative-serving-webhook: 30310     # Serving webhook
knative-serving-activator: 30311   # Activator service

# RabbitMQ (event bus for serverless)
rabbitmq-management: 30320         # Management UI
rabbitmq-amqp: 30321              # AMQP protocol
rabbitmq-prometheus: 30322         # Prometheus metrics
```

### Observability (30400-30499)

```yaml
# Studio Cluster
Grafana: 30040
Prometheus: 30041
Loki: 30042
Tempo: 30043
AlertManager: 30044
```

### Data Services (30500-30599)

```yaml
# Pro/Studio Clusters
PostgreSQL: 30060
MongoDB: 30061
Redis: 30062
MinIO: 30063
```

### Message Queues (30600-30699)

```yaml
RabbitMQ: 30080
Kafka: 30081
```

## Port Allocation Rules

1. **Reserve ranges** - Don't use ad-hoc ports
2. **Document immediately** - Update this file when allocating
3. **Cluster consistency** - Use same port across clusters when possible
4. **NodePort vs LoadBalancer** - Prefer LoadBalancer for production (Studio)
5. **Service mesh** - Use Linkerd for cross-cluster communication

## Cross-Cluster Service Access

### From Studio to Forge

```python
# VLLM API on Forge
base_url = "http://vllm.ml-inference.svc.forge.remote:8000/v1"
```

### From Any Cluster to Studio

```bash
# Grafana on Studio
http://grafana.observability.svc.studio.remote:30040
```

## Port Conflicts

If you encounter port conflicts:

1. Check this document first
2. Choose next available port in the range
3. Update service manifest
4. Update this document
5. Commit to git

## Related Documentation

- [Node Labels Reference](node-labels-reference.md)
- [Migration Guide](migration-guide.md)
- [Cluster Documentation](../clusters/README.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

