# 🚀 Quick Start - Homelab Services

## 📋 Pré-requisitos

- Kubernetes cluster (homelab)
- kubectl configurado
- Flux instalado (opcional, para GitOps)
- Domínio configurado (opcional, para acesso externo)

## 🎯 Deploy Rápido

### 1. Clone e Prepare

```bash
cd flux/homelab-services
```

### 2. Configure Secrets

```bash
# Criar secrets
kubectl create namespace homelab-services

kubectl create secret generic homelab-services-secrets \
  --from-literal=mongodb-username=admin \
  --from-literal=mongodb-password=seu-password-seguro \
  --from-literal=redis-password=seu-redis-password \
  --from-literal=jwt-secret=seu-jwt-secret-aleatorio \
  -n homelab-services
```

### 3. Deploy Infraestrutura

```bash
# Deploy compartilhado
kubectl apply -f shared-infra/

# Aguardar serviços ficarem prontos
kubectl wait --for=condition=ready pod -l app=mongodb -n homelab-services --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n homelab-services --timeout=300s
```

### 4. Deploy Gateway

```bash
# Gateway (Kong)
kubectl apply -f gateway/
```

### 5. Deploy Serviços

```bash
# DJ Collab P2P
kubectl apply -k ../dj-collab-p2p/k8s/

# Outros serviços quando disponíveis
# kubectl apply -k ../spotify-p2p/k8s/
# kubectl apply -k ../rekordbox-cloud/k8s/
```

### 6. Deploy Mobile API

```bash
kubectl apply -f mobile-api/
```

### 7. Configurar Cloudflare Tunnel

```bash
# Aplicar CloudflareTunnelIngress
kubectl apply -f cloudflare-tunnel/cloudflaretunnelingress.yaml

# Verificar status
kubectl get cloudflaretunnelingress -n homelab-services

# Serviços estarão disponíveis em:
# - api.music.lucena.cloud (Mobile API)
# - music.lucena.cloud (Gateway)
# - dj-collab.music.lucena.cloud (DJ Collab)
```

## 📱 Configurar Mobile App

### 1. Instalar App

```bash
# iOS
cd apps/mobile
pnpm ios

# Android
pnpm android
```

### 2. Configurar Conexão

```typescript
// No app, configurar URL do homelab
const config = {
  homelabUrl: 'https://music.seu-homelab.com', // ou IP local
  services: ['dj-collab', 'spotify-p2p', 'rekordbox'],
  auth: {
    provider: 'jwt',
    tokenStorage: 'secure-storage'
  }
};

const client = new HomelabClient(config);
```

### 3. Login

```typescript
// Primeiro uso: criar conta
await client.authenticate('seu-usuario', 'sua-senha');
```

## 🔧 Configuração Avançada

### Acesso Externo

#### Opção 1: Cloudflare Tunnel

```bash
# Deploy Cloudflare Tunnel
kubectl apply -f cloudflare-tunnel/
```

#### Opção 2: VPN

```bash
# Configurar VPN (WireGuard, Tailscale, etc)
# Acessar via VPN
```

#### Opção 3: Ingress com TLS

```bash
# Usar cert-manager para TLS
kubectl apply -f ingress/
```

### Monitoramento

```bash
# Deploy Prometheus/Grafana
kubectl apply -f monitoring/
```

### Backup

```bash
# Configurar backup automático
kubectl apply -f backup/
```

## 🧪 Testar

### 1. Health Check

```bash
# Gateway
curl http://localhost:8000/api/v1/health

# Mobile API
curl http://localhost:8080/api/v1/health
```

### 2. Listar Serviços

```bash
# Autenticar primeiro
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.token')

# Listar serviços
curl http://localhost:8080/api/v1/services \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Acessar Serviço

```bash
# DJ Collab
curl http://localhost:8000/api/v1/dj-collab/sessions \
  -H "Authorization: Bearer $TOKEN"
```

## 🐛 Troubleshooting

### Serviços não iniciam

```bash
# Verificar logs
kubectl logs -l app=mongodb -n homelab-services
kubectl logs -l app=redis -n homelab-services
kubectl logs -l app=kong-gateway -n homelab-services
```

### Gateway não roteia

```bash
# Verificar configuração Kong
kubectl exec -it -n homelab-services deployment/kong-gateway -- \
  curl http://localhost:8001/config

# Recarregar configuração
kubectl rollout restart deployment/kong-gateway -n homelab-services
```

### Mobile não conecta

```bash
# Verificar conectividade
curl -v https://music.seu-homelab.com/api/v1/health

# Verificar DNS
nslookup music.seu-homelab.com

# Verificar firewall
# Permitir portas 80, 443, 8080
```

## 📚 Próximos Passos

1. **Explorar Serviços**
   - DJ Collab P2P
   - Spotify P2P (quando disponível)
   - rekordbox Cloud (quando disponível)

2. **Customizar**
   - Adicionar novos serviços
   - Configurar plugins
   - Personalizar UI

3. **Contribuir**
   - Criar issues
   - Fazer pull requests
   - Melhorar documentação

## 💡 Dicas

- Use VPN para acesso seguro
- Configure backup automático
- Monitore recursos (CPU, memória, storage)
- Use storage classes apropriadas
- Configure limites de recursos

## 📞 Suporte

- Issues: GitHub Issues
- Email: bruno@lucena.cloud
- Documentação: `/docs`

---

**🏠 Seu homelab, seus dados, seu controle!**
