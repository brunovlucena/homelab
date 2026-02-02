# 🛡️ Garak - LLM Vulnerability Scanner

**NVIDIA's comprehensive LLM security testing tool deployed in your homelab**

---

## 📋 Overview

Garak is an automated LLM vulnerability scanner that detects:
- Prompt injection attacks
- Jailbreak techniques
- Data leakage vulnerabilities
- Hallucinations
- Output manipulation

This deployment includes:
- Garak CLI tool (installed in container)
- FastAPI wrapper for HTTP API access
- Kubernetes Deployment and Service
- Ready to scan your LLM applications

---

## 🚀 Quick Start

### 1. Build and Push Docker Image

```bash
cd flux/clusters/homelab/infrastructure/garak

# Build and push (requires Docker login to GHCR)
make build push

# Or manually:
docker build -t ghcr.io/brunovlucena/garak:latest .
docker push ghcr.io/brunovlucena/garak:latest
```

### 2. Deploy to Cluster

```bash
# Deploy via kubectl (for testing)
kubectl apply -k .

# Or let Flux deploy it automatically (it's in phase4-apps)
# Flux will reconcile and deploy based on GitOps
```

### 3. Access the API

```bash
# Port-forward to access locally
kubectl port-forward -n garak svc/garak-service 8080:8080

# Check health
curl http://localhost:8080/health

# View API docs
open http://localhost:8080/docs
```

### Run a Scan

```bash
# Scan agent-evil (pre-configured endpoint)
curl -X POST http://localhost:8080/scan/agent-evil

# Custom scan
curl -X POST http://localhost:8080/scan \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "openai",
    "model_name": "gpt-3.5-turbo",
    "probes": "prompt_injection,jailbreak",
    "report_format": "json"
  }'
```

### Scan agent-evil in Your Cluster

```bash
# Scan agent-evil via internal service
curl -X POST http://localhost:8080/scan \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "http",
    "url": "http://agent-evil-service.agent-evil.svc.cluster.local:8080/chat",
    "probes": "prompt_injection,jailbreak",
    "report_format": "json"
  }'
```

---

## 🔧 API Endpoints

### `GET /health`
Health check endpoint

### `GET /`
API information and available endpoints

### `GET /probes`
List all available Garak probes

### `POST /scan`
Run a custom vulnerability scan

**Request Body:**
```json
{
  "model_type": "openai",
  "model_name": "gpt-3.5-turbo",
  "probes": "prompt_injection,jailbreak",
  "url": "http://example.com/chat",
  "report_format": "json",
  "output_file": "/tmp/scan.json"
}
```

### `POST /scan/agent-evil`
Quick scan endpoint pre-configured for agent-evil

### `POST /scan/agent-sre`
Quick scan endpoint pre-configured for agent-sre (tests health endpoint for connectivity)

---

## 📊 Example Scan Results

```json
{
  "status": "success",
  "scan_id": "abc12345",
  "command": "garak --model_type openai --model_name gpt-3.5-turbo",
  "output": "{...scan results...}",
  "error": null
}
```

---

## 🎯 Use Cases

### 1. Scan agent-evil
Test your intentionally vulnerable LLM application:
```bash
curl -X POST http://localhost:8080/scan/agent-evil
```

### 2. Scan agent-sre
Monitor the SRE agent for vulnerabilities:
```bash
curl -X POST http://localhost:8080/scan/agent-sre
```

### 2. Scan agent-sre
Test the SRE agent (AI-powered health report generator):
```bash
# Quick scan via pre-configured endpoint
curl -X POST http://localhost:8080/scan/agent-sre

# Custom scan with specific probes
curl -X POST http://localhost:8080/scan \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "http",
    "url": "http://agent-sre.ai.svc.cluster.local/health",
    "probes": "prompt_injection,jailbreak",
    "report_format": "json"
  }'
```

**Note**: Agent-SRE currently doesn't have a chat endpoint for LLM vulnerability testing. The scan endpoint tests connectivity via the health endpoint. For proper LLM vulnerability testing, agent-sre would need a chat endpoint that accepts user messages and returns LLM responses.

### 3. Scan AIGoat
Test the OWASP Top 10 vulnerable application:
```bash
curl -X POST http://localhost:8080/scan \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "http",
    "url": "http://aigoat.aigoat.svc.cluster.local:8000/api/chat",
    "probes": "all",
    "report_format": "json"
  }'
```

### 3. Scan External LLM APIs
Test OpenAI, Anthropic, or other LLM providers:
```bash
curl -X POST http://localhost:8080/scan \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "openai",
    "model_name": "gpt-4",
    "probes": "prompt_injection",
    "report_format": "json"
  }'
```

---

## 🏗️ Architecture

```
┌─────────────────┐
│   FastAPI       │
│   API Wrapper   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Garak CLI     │
│   (NVIDIA)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   LLM Targets   │
│   - agent-evil  │
│   - agent-sre   │
│   - AIGoat      │
│   - Custom APIs │
└─────────────────┘
```

---

## 🔐 Security Notes

- Garak runs in an isolated namespace
- Only accessible via ClusterIP (internal cluster access)
- No external exposure by default
- Use port-forward for local access

---

## 📚 Resources

- [Garak GitHub](https://github.com/leondz/garak)
- [Garak Documentation](https://github.com/leondz/garak#readme)
- [OWASP Top 10 for LLM](https://owasp.org/www-project-top-10-for-large-language-model-applications/)

---

## 🐛 Troubleshooting

### Check Pod Status
```bash
kubectl get pods -n garak
kubectl logs -n garak -l app=garak
```

### Test API Locally
```bash
kubectl port-forward -n garak svc/garak-service 8080:8080
curl http://localhost:8080/health
```

### Run Garak Directly (for debugging)
```bash
kubectl exec -it -n garak deployment/garak -- garak --help
```

---

**Last Updated**: November 2025  
**Deployed via**: Flux GitOps

