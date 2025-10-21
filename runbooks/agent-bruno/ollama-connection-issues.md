# 🚨 Runbook: Agent Bruno Ollama/LLM Connection Issues

## Alert Information

**Alert Name:** `AgentBrunoOllamaConnectionFailure`  
**Severity:** Critical  
**Component:** agent-bruno  
**Service:** ollama-llm

## Symptom

Agent Bruno cannot connect to the Ollama LLM server at 192.168.0.16:11434, causing chat functionality to fail completely.

## Impact

- **User Impact:** SEVERE - No AI responses, chat completely non-functional
- **Business Impact:** HIGH - AI assistant completely unavailable
- **Data Impact:** NONE - Memory systems unaffected

## Diagnosis

### 1. Check Agent Bruno Logs

```bash
kubectl logs -n bruno -l app=agent-bruno --tail=100 | grep -i ollama
```

### 2. Check Ollama Server URL Configuration

```bash
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="OLLAMA_URL")].value}'
# Should output: http://192.168.0.16:11434
```

### 3. Test Ollama Connectivity from Agent Bruno Pod

```bash
# Test network connectivity
kubectl exec -it -n bruno deployment/agent-bruno -- sh -c 'nc -zv 192.168.0.16 11434'

# Test HTTP connectivity
kubectl exec -it -n bruno deployment/agent-bruno -- curl -v http://192.168.0.16:11434/api/version

# Test with Python
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
import requests
try:
    response = requests.get('http://192.168.0.16:11434/api/version', timeout=5)
    print('Status:', response.status_code)
    print('Response:', response.json())
except Exception as e:
    print('ERROR:', str(e))
"
```

### 4. Test Ollama Server Directly (from your machine)

```bash
# Check if Ollama is responding
curl http://192.168.0.16:11434/api/version

# List available models
curl http://192.168.0.16:11434/api/tags

# Test generation
curl http://192.168.0.16:11434/api/generate -d '{
  "model": "llama2",
  "prompt": "Hello",
  "stream": false
}'
```

## Resolution Steps

### Step 1: Verify Ollama Server is Running

```bash
# SSH to Ollama server
ssh user@192.168.0.16

# Check Ollama service status
sudo systemctl status ollama

# Check Ollama is listening on port 11434
sudo netstat -tlnp | grep 11434

# Check Ollama logs
sudo journalctl -u ollama -n 100 --no-pager
```

### Step 2: Common Issues and Fixes

#### Issue: Ollama Service Not Running
**Cause:** Ollama service stopped or crashed  
**Fix:**
```bash
# On Ollama server (192.168.0.16)
sudo systemctl start ollama
sudo systemctl enable ollama
sudo systemctl status ollama

# Check if models are loaded
ollama list
```

#### Issue: Ollama Not Listening on Network Interface
**Cause:** Ollama only listening on localhost  
**Fix:**
```bash
# On Ollama server, edit Ollama service
sudo systemctl edit ollama

# Add or update environment:
[Service]
Environment="OLLAMA_HOST=0.0.0.0:11434"

# Restart Ollama
sudo systemctl restart ollama

# Verify it's listening on all interfaces
sudo netstat -tlnp | grep 11434
```

#### Issue: Firewall Blocking Connection
**Cause:** Firewall rules blocking port 11434  
**Fix:**
```bash
# On Ollama server, check firewall
sudo ufw status
sudo iptables -L -n | grep 11434

# Allow port 11434 if needed
sudo ufw allow 11434/tcp
# Or
sudo iptables -A INPUT -p tcp --dport 11434 -j ACCEPT
```

#### Issue: Network Routing Problem
**Cause:** Kubernetes cluster cannot reach 192.168.0.16  
**Fix:**
```bash
# From any cluster node, test connectivity
ping 192.168.0.16
curl http://192.168.0.16:11434/api/version

# Check cluster network configuration
kubectl get nodes -o wide

# Verify no network policies blocking egress
kubectl get networkpolicies -n bruno
```

#### Issue: Wrong Ollama URL in Agent Bruno
**Cause:** Incorrect OLLAMA_URL environment variable  
**Fix:**
```bash
# Check current URL
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="OLLAMA_URL")].value}'

# Update if incorrect
kubectl edit deployment -n bruno agent-bruno
# Change OLLAMA_URL to: http://192.168.0.16:11434

# Or update via patch
kubectl patch deployment agent-bruno -n bruno --type='json' -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/env",
    "value": [
      {"name": "OLLAMA_URL", "value": "http://192.168.0.16:11434"}
    ]
  }
]'
```

#### Issue: Ollama Model Not Loaded
**Cause:** Required model not available  
**Fix:**
```bash
# On Ollama server, check available models
ollama list

# Pull required model (e.g., llama2)
ollama pull llama2

# Or pull specific model used by Agent Bruno
ollama pull mistral
ollama pull codellama
ollama pull llama2:13b

# Verify model is available
ollama list
```

#### Issue: Ollama Out of Memory
**Cause:** Ollama server running out of GPU/CPU memory  
**Fix:**
```bash
# On Ollama server, check memory usage
free -h
nvidia-smi  # If using GPU

# Check Ollama resource usage
top -p $(pgrep ollama)

# Restart Ollama to free memory
sudo systemctl restart ollama

# Consider limiting concurrent requests or using smaller models
```

#### Issue: Timeout Connecting to Ollama
**Cause:** Ollama taking too long to respond  
**Fix:**
```bash
# Check Ollama server load
ssh user@192.168.0.16
top
nvidia-smi  # For GPU

# Increase timeout in Agent Bruno (if configurable)
# Or switch to faster/smaller model

# Check if Ollama is processing other requests
ollama ps  # Shows running models
```

### Step 3: Restart Agent Bruno

```bash
kubectl rollout restart deployment/agent-bruno -n bruno
kubectl rollout status deployment/agent-bruno -n bruno
```

## Verification

1. Test Ollama connectivity from Agent Bruno:
```bash
kubectl exec -it -n bruno deployment/agent-bruno -- curl http://192.168.0.16:11434/api/version
```

2. Test chat functionality via API:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Send a test chat message
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, tell me about the homepage application"}'

# Should receive AI-generated response
```

3. Check Agent Bruno logs for successful Ollama connections:
```bash
kubectl logs -n bruno -l app=agent-bruno --tail=50 | grep -i ollama
```

4. Test knowledge search (uses Ollama):
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl "http://localhost:8080/knowledge/search?q=deployment"
```

## Prevention

1. **Monitor Ollama Server Health**
   - Set up health checks for Ollama service
   - Monitor memory usage (especially GPU memory)
   - Alert on service down or high latency

2. **Ollama Server Redundancy**
   - Consider setting up multiple Ollama servers
   - Implement load balancing or failover
   - Use Kubernetes-native Ollama deployment

3. **Resource Management**
   - Ensure adequate GPU/CPU resources
   - Monitor model loading times
   - Limit concurrent requests if needed

4. **Network Reliability**
   - Ensure stable network between cluster and Ollama server
   - Monitor network latency
   - Consider co-locating Ollama in cluster

5. **Configuration Management**
   - Version control Ollama configuration
   - Test Ollama connectivity in CI/CD
   - Document model requirements

## Performance Tips

1. **Model Selection**
   - Use smaller models for faster responses (llama2:7b vs llama2:13b)
   - Consider quantized models (Q4, Q5) for better performance
   - Match model size to available hardware

2. **Caching**
   - Ollama caches model context
   - Keep models loaded for better response times
   - Monitor cache hit rates

3. **Request Optimization**
   - Limit prompt size
   - Use streaming for long responses
   - Implement request queuing in high load

4. **Hardware Optimization**
   - Use GPU acceleration when available
   - Ensure adequate VRAM for models
   - Consider CPU-only models for smaller deployments

## Alternative Solutions

If Ollama server is unavailable and cannot be quickly restored:

1. **Temporary Fallback**
   - Deploy Ollama in Kubernetes cluster
   - Use external LLM API (OpenAI, Anthropic)
   - Implement graceful degradation (return error messages)

2. **Ollama in Kubernetes** (Long-term solution)
   ```bash
   # Deploy Ollama in cluster
   kubectl create namespace ollama
   kubectl apply -f ollama-deployment.yaml
   
   # Update Agent Bruno to use cluster Ollama
   kubectl edit deployment -n bruno agent-bruno
   # Change OLLAMA_URL to: http://ollama.ollama.svc.cluster.local:11434
   ```

## Related Alerts

- `AgentBrunoAPIDown`
- `AgentBrunoHighResponseTime`
- `AgentBrunoChatFailures`
- `OllamaServerDown` (if configured)

## Escalation

If unable to resolve within 15 minutes:
1. Check Ollama server infrastructure
2. Verify network connectivity from cluster
3. Check if Ollama server needs reboot
4. Consider failover to backup Ollama server
5. Contact infrastructure team
6. Escalate to Ollama server administrator

## Additional Resources

- [Ollama Documentation](https://ollama.ai/docs)
- [Ollama API Reference](https://github.com/ollama/ollama/blob/main/docs/api.md)
- [Agent Bruno Configuration](../../../flux/clusters/homelab/infrastructure/agent-bruno/README.md#-configuration)
- [Ollama Troubleshooting](https://github.com/ollama/ollama/blob/main/docs/troubleshooting.md)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

