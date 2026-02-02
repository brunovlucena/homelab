# 🚀 Deploy Guide - Homelab Services

Guia completo para fazer deploy de todos os serviços do Homelab Services.

## 📋 Pré-requisitos

- Kubernetes cluster (homelab)
- kubectl configurado
- Cloudflare Tunnel Operator instalado
- Secrets configurados

## 🔧 Configuração Inicial

### 1. Criar Secrets

```bash
# Secrets compartilhados
kubectl create namespace homelab-services

kubectl create secret generic homelab-services-secrets \
  --from-literal=mongodb-uri="mongodb://mongodb.homelab-services.svc.cluster.local:27017" \
  --from-literal=redis-uri="redis://redis.homelab-services.svc.cluster.local:6379" \
  --from-literal=jwt-secret="$(openssl rand -hex 32)" \
  -n homelab-services

# Secrets para Spotify P2P
kubectl create namespace spotify-p2p
kubectl create secret generic spotify-p2p-secrets \
  --from-literal=mongodb-uri="mongodb://mongodb.homelab-services.svc.cluster.local:27017" \
  --from-literal=redis-uri="redis://redis.homelab-services.svc.cluster.local:6379" \
  -n spotify-p2p

# Secrets para rekordbox Cloud
kubectl create namespace rekordbox-cloud
kubectl create secret generic rekordbox-cloud-secrets \
  --from-literal=mongodb-uri="mongodb://mongodb.homelab-services.svc.cluster.local:27017" \
  --from-literal=redis-uri="redis://redis.homelab-services.svc.cluster.local:6379" \
  --from-literal=minio-access-key="minioadmin" \
  --from-literal=minio-secret-key="minioadmin" \
  -n rekordbox-cloud
```

### 2. Deploy Infraestrutura Compartilhada

```bash
# MongoDB, Redis, IPFS, MinIO, PostgreSQL
kubectl apply -f shared-infra/

# Aguardar serviços ficarem prontos
kubectl wait --for=condition=ready pod -l app=mongodb -n homelab-services --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n homelab-services --timeout=300s
```

## 🚀 Deploy dos Serviços

### 1. Kong Gateway

```bash
# Deploy Gateway
kubectl apply -f gateway/

# Verificar status
kubectl get pods -n homelab-services -l app=kong-gateway
kubectl get svc -n homelab-services kong-gateway
```

### 2. Mobile API

```bash
# Deploy Mobile API
kubectl apply -f mobile-api/

# Verificar status
kubectl get pods -n homelab-services -l app=mobile-api
kubectl get svc -n homelab-services mobile-api

# Verificar logs
kubectl logs -n homelab-services -l app=mobile-api --tail=50
```

### 3. DJ Collab P2P

```bash
# Deploy DJ Collab P2P
kubectl apply -k ../dj-collab-p2p/k8s/

# Verificar status
kubectl get pods -n dj-collab-p2p
kubectl get svc -n dj-collab-p2p
```

### 4. Spotify P2P

```bash
# Deploy Spotify P2P
kubectl apply -k spotify-p2p/

# Verificar status
kubectl get pods -n spotify-p2p
kubectl get svc -n spotify-p2p
```

### 5. rekordbox Cloud

```bash
# Deploy rekordbox Cloud
kubectl apply -k rekordbox-cloud/

# Verificar status
kubectl get pods -n rekordbox-cloud
kubectl get svc -n rekordbox-cloud
```

## ☁️ Configurar Cloudflare Tunnel

### 1. Aplicar CloudflareTunnelIngress

```bash
# Aplicar todos os ingress
kubectl apply -f cloudflare-tunnel/cloudflaretunnelingress.yaml

# Verificar status
kubectl get cloudflaretunnelingress -n homelab-services
```

### 2. Verificar Status

```bash
# Listar todos os ingress
kubectl get cfti -A

# Ver detalhes de um ingress
kubectl describe cloudflaretunnelingress homelab-services-mobile-api -n homelab-services

# Verificar fase
kubectl get cfti -n homelab-services -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,HOSTNAME:.spec.hostname
```

### 3. Habilitar Serviços

```bash
# Habilitar Spotify P2P (quando serviço estiver pronto)
kubectl patch cloudflaretunnelingress spotify-p2p -n homelab-services \
  --type=merge -p '{"spec":{"enabled":true}}'

# Habilitar rekordbox Cloud (quando serviço estiver pronto)
kubectl patch cloudflaretunnelingress rekordbox-cloud -n homelab-services \
  --type=merge -p '{"spec":{"enabled":true}}'
```

## ✅ Verificação

### 1. Health Checks

```bash
# Mobile API
curl https://api.music.lucena.cloud/health

# Kong Gateway
curl https://music.lucena.cloud/api/v1/health

# DJ Collab P2P
curl https://dj-collab.music.lucena.cloud/api/v1/health
```

### 2. Service Discovery

```bash
# Listar serviços via Mobile API
curl https://api.music.lucena.cloud/api/v1/services \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Testar CloudEvents (AgentApp)

```bash
# Enviar CloudEvent
curl -X POST https://api.music.lucena.cloud/api/v1/cloudevents \
  -H "Content-Type: application/json" \
  -H "ce-specversion: 1.0" \
  -H "ce-type: agent.message" \
  -H "ce-source: test/client" \
  -H "ce-id: test-123" \
  -H "ce-time: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -d '{
    "data": {
      "agentId": "dj-collab-agent",
      "content": "Test message"
    }
  }'
```

## 📊 Monitoramento

### Verificar Pods

```bash
# Todos os serviços
kubectl get pods -A | grep -E "homelab-services|dj-collab|spotify|rekordbox"

# Por namespace
kubectl get pods -n homelab-services
kubectl get pods -n dj-collab-p2p
kubectl get pods -n spotify-p2p
kubectl get pods -n rekordbox-cloud
```

### Verificar Services

```bash
# Todos os services
kubectl get svc -A | grep -E "homelab-services|dj-collab|spotify|rekordbox"
```

### Verificar Logs

```bash
# Mobile API
kubectl logs -n homelab-services -l app=mobile-api --tail=100 -f

# Kong Gateway
kubectl logs -n homelab-services -l app=kong-gateway --tail=100 -f

# DJ Collab
kubectl logs -n dj-collab-p2p -l app=dj-collab-p2p-server --tail=100 -f
```

## 🔧 Troubleshooting

### Serviço não inicia

```bash
# Verificar eventos
kubectl describe pod <pod-name> -n <namespace>

# Verificar logs
kubectl logs <pod-name> -n <namespace>

# Verificar recursos
kubectl top pod <pod-name> -n <namespace>
```

### Cloudflare Tunnel não conecta

```bash
# Verificar operador
kubectl get pods -n cloudflare-tunnel-operator

# Verificar logs do operador
kubectl logs -n cloudflare-tunnel-operator -l app=cloudflare-tunnel-operator --tail=100

# Verificar ingress
kubectl get cloudflaretunnelingress -n homelab-services -o yaml
```

### Gateway não roteia

```bash
# Verificar configuração Kong
kubectl exec -it -n homelab-services deployment/kong-gateway -- \
  curl http://localhost:8001/config

# Recarregar configuração
kubectl rollout restart deployment/kong-gateway -n homelab-services
```

## 📝 Próximos Passos

1. **Configurar DNS**: Garantir que os hostnames estão configurados no Cloudflare
2. **Configurar TLS**: Cloudflare fornece TLS automático
3. **Monitoramento**: Configurar Prometheus/Grafana
4. **Backup**: Configurar backup automático
5. **Scaling**: Configurar HPA se necessário

## 🎯 URLs Finais

Após o deploy, os serviços estarão disponíveis em:

- **Mobile API**: `https://api.music.lucena.cloud`
- **Kong Gateway**: `https://music.lucena.cloud`
- **DJ Collab P2P**: `https://dj-collab.music.lucena.cloud`
- **Spotify P2P**: `https://spotify.music.lucena.cloud` (quando habilitado)
- **rekordbox Cloud**: `https://rekordbox.music.lucena.cloud` (quando habilitado)

---

**🚀 Todos os serviços deployados e funcionando!**
