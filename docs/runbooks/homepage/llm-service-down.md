# ⚠️ Runbook: LLM Service Down

## Alert Information
**Alert Name:** `BrunoSiteLLMServiceDown`  
**Severity:** Warning  

## Symptom
Ollama LLM service has been down for more than 5 minutes.

## Impact
AI chat functionality is affected but site remains operational.

## Diagnosis
```bash
# Check if Ollama is accessible
curl -v http://192.168.0.16:11434/api/version

# Check from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- curl http://192.168.0.16:11434/api/version
```

## Resolution
1. Verify Ollama server is running on 192.168.0.16
2. Check network connectivity
3. Restart Ollama service if needed
4. Verify firewall rules allow port 11434

## Related Alerts
- `BrunoSiteChatAPIErrors`
