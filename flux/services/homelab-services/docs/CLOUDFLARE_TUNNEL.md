# ☁️ Cloudflare Tunnel Integration

## 🎯 Visão Geral

O Homelab Services usa **CloudflareTunnelIngress** (operador customizado) para expor serviços via Cloudflare Tunnel, sem necessidade de abrir portas no firewall ou configurar DNS público.

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────┐
│         🌐 Internet / Cloudflare                        │
│                                                         │
│  • api.music.lucena.cloud                              │
│  • music.lucena.cloud                                  │
│  • dj-collab.music.lucena.cloud                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     │ Cloudflare Tunnel
                     │
┌────────────────────▼────────────────────────────────────┐
│    ☁️ Cloudflare Tunnel Operator                         │
│                                                         │
│  • CloudflareTunnelIngress CR                           │
│  • Auto-sync tunnel config                              │
│  • Health checks                                        │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────▼──────┐ ┌──▼──────┐ ┌───▼──────┐
│ Mobile API   │ │ Gateway │ │ Services │
│ Port: 8080   │ │ Port:8000│ │ Port:8080│
└──────────────┘ └──────────┘ └──────────┘
```

## 📦 CloudflareTunnelIngress CRD

### Estrutura

```yaml
apiVersion: tunnel.cloudflare.io/v1alpha1
kind: CloudflareTunnelIngress
metadata:
  name: homelab-services-mobile-api
  namespace: homelab-services
spec:
  hostname: api.music.lucena.cloud
  service:
    name: mobile-api
    namespace: homelab-services
    port: 8080
    protocol: http
  enabled: true
  syncInterval: "5m"
```

### Campos

- **hostname**: Hostname público no Cloudflare
- **service.name**: Nome do serviço Kubernetes
- **service.namespace**: Namespace do serviço
- **service.port**: Porta do serviço
- **service.protocol**: http ou https
- **enabled**: Ativar/desativar ingress
- **syncInterval**: Intervalo de sincronização

## 🚀 Deploy

### 1. Verificar Operador

```bash
# Verificar se o operador está rodando
kubectl get pods -n cloudflare-tunnel-operator

# Verificar CRD
kubectl get crd cloudflaretunnelingresses.tunnel.cloudflare.io
```

### 2. Aplicar CloudflareTunnelIngress

```bash
# Aplicar todos os ingress
kubectl apply -f cloudflare-tunnel/cloudflaretunnelingress.yaml

# Verificar status
kubectl get cloudflaretunnelingress -n homelab-services
```

### 3. Verificar Status

```bash
# Ver detalhes
kubectl describe cloudflaretunnelingress homelab-services-mobile-api -n homelab-services

# Verificar fase
kubectl get cfti -n homelab-services -o jsonpath='{.items[*].status.phase}'
```

## 📋 Ingress Configurados

### Mobile API
- **Hostname**: `api.music.lucena.cloud`
- **Service**: `mobile-api:8080`
- **Status**: ✅ Enabled

### Kong Gateway
- **Hostname**: `music.lucena.cloud`
- **Service**: `kong-gateway:8000`
- **Status**: ✅ Enabled

### DJ Collab P2P
- **Hostname**: `dj-collab.music.lucena.cloud`
- **Service**: `dj-collab-p2p-server:8080`
- **Status**: ✅ Enabled

### Spotify P2P (quando disponível)
- **Hostname**: `spotify.music.lucena.cloud`
- **Service**: `spotify-p2p-server:8080`
- **Status**: ⏸️ Disabled

### rekordbox Cloud (quando disponível)
- **Hostname**: `rekordbox.music.lucena.cloud`
- **Service**: `rekordbox-cloud-server:8080`
- **Status**: ⏸️ Disabled

## 🔧 Configuração

### Habilitar/Desabilitar Ingress

```bash
# Desabilitar
kubectl patch cloudflaretunnelingress spotify-p2p -n homelab-services \
  --type=merge -p '{"spec":{"enabled":false}}'

# Habilitar
kubectl patch cloudflaretunnelingress spotify-p2p -n homelab-services \
  --type=merge -p '{"spec":{"enabled":true}}'
```

### Alterar Hostname

```bash
kubectl patch cloudflaretunnelingress homelab-services-mobile-api -n homelab-services \
  --type=merge -p '{"spec":{"hostname":"new-api.music.lucena.cloud"}}'
```

### Alterar Porta

```bash
kubectl patch cloudflaretunnelingress homelab-services-mobile-api -n homelab-services \
  --type=merge -p '{"spec":{"service":{"port":9090}}}'
```

## 📊 Status e Monitoramento

### Verificar Status

```bash
# Listar todos os ingress
kubectl get cfti -n homelab-services

# Ver detalhes de um ingress
kubectl get cfti homelab-services-mobile-api -n homelab-services -o yaml

# Verificar condições
kubectl get cfti homelab-services-mobile-api -n homelab-services \
  -o jsonpath='{.status.conditions[*]}'
```

### Fases Possíveis

- **Pending**: Aguardando processamento
- **Syncing**: Sincronizando com Cloudflare
- **Ready**: Configurado e funcionando
- **Failed**: Erro na configuração

### Logs do Operador

```bash
# Ver logs do operador
kubectl logs -n cloudflare-tunnel-operator \
  -l app=cloudflare-tunnel-operator --tail=100
```

## 🐛 Troubleshooting

### Ingress não sincroniza

1. Verificar se o operador está rodando
2. Verificar credenciais do Cloudflare
3. Verificar se o serviço backend existe
4. Verificar logs do operador

### Hostname não resolve

1. Verificar DNS no Cloudflare
2. Verificar se o tunnel está ativo
3. Verificar configuração do Cloudflare Tunnel

### Serviço não acessível

1. Verificar se o serviço está rodando
2. Verificar porta do serviço
3. Verificar conectividade interna
4. Verificar health checks

### Erro 502 Bad Gateway

1. Verificar se o serviço backend está saudável
2. Verificar se a porta está correta
3. Verificar logs do serviço
4. Verificar network policies

## 🔐 Segurança

### TLS/SSL

- Cloudflare fornece TLS automático
- Certificados gerenciados pelo Cloudflare
- HTTPS automático para todos os hostnames

### Autenticação

- JWT tokens para API
- Cloudflare Access (opcional)
- Rate limiting via Cloudflare

### Firewall

- Sem necessidade de abrir portas
- Tráfego via Cloudflare Tunnel
- Proteção DDoS automática

## 📚 Referências

- [Cloudflare Tunnel Operator](../../infrastructure/cloudflare-tunnel-operator/README.md)
- [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [CloudflareTunnelIngress CRD](../../infrastructure/cloudflare-tunnel-operator/k8s/base/crd.yaml)

---

**☁️ Acesso seguro via Cloudflare Tunnel!**
