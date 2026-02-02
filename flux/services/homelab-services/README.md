# 🏠 Homelab Services - Plataforma Unificada

Plataforma unificada para rodar múltiplos serviços de música/streaming no seu homelab, acessível via mobile e outros dispositivos.

## 🎯 Serviços Implementados

### ✅ Mobile API
- **URL**: `https://api.music.lucena.cloud`
- **Status**: ✅ Implementado
- **Features**: CloudEvents para AgentApp, Service Discovery

### ✅ Kong Gateway
- **URL**: `https://music.lucena.cloud`
- **Status**: ✅ Implementado
- **Features**: API Gateway unificado, JWT auth, Rate limiting

### ✅ DJ Collab P2P
- **URL**: `https://dj-collab.music.lucena.cloud`
- **Status**: ✅ Implementado
- **Features**: Streaming P2P, Colaboração em tempo real, Gamificação

### ✅ Spotify P2P
- **URL**: `https://spotify.music.lucena.cloud`
- **Status**: ✅ Implementado
- **Features**: Streaming de biblioteca pessoal, Estações P2P

### ✅ rekordbox Cloud
- **URL**: `https://rekordbox.music.lucena.cloud`
- **Status**: ✅ Implementado
- **Features**: Sincronização de biblioteca, Análise de música (BPM, key)

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    📱 Mobile/Web Client                      │
│  (AgentApp / React Native / Next.js)                        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ HTTPS/WSS
                       │
┌──────────────────────▼──────────────────────────────────────┐
│              🌐 Cloudflare Tunnel                            │
│  • api.music.lucena.cloud                                    │
│  • music.lucena.cloud                                        │
│  • dj-collab.music.lucena.cloud                             │
│  • spotify.music.lucena.cloud                                │
│  • rekordbox.music.lucena.cloud                             │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼──────┐ ┌────▼─────┐ ┌─────▼──────┐
│ 🌐 Kong      │ │ 📱 Mobile│ │ ☁️ Cloudflare│
│   Gateway    │ │   API    │ │   Tunnel    │
└──────┬───────┘ └──────────┘ └─────────────┘
       │
       │
┌──────▼──────────────────────────────────────┐
│         🎧 Serviços de Música                │
│                                              │
│  • DJ Collab P2P                            │
│  • Spotify P2P                               │
│  • rekordbox Cloud                           │
└──────┬───────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│    🗄️ Shared Infrastructure                  │
│                                              │
│  • MongoDB (Users, Sessions)               │
│  • Redis (Cache, Real-time)                │
│  • IPFS (Content Distribution)              │
│  • MinIO (Object Storage)                   │
│  • PostgreSQL (Metadata)                    │
└──────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Deploy Completo

```bash
# 1. Criar secrets
kubectl apply -f secrets/

# 2. Deploy infraestrutura compartilhada
kubectl apply -f shared-infra/

# 3. Deploy serviços
kubectl apply -k .

# 4. Configurar Cloudflare Tunnel
kubectl apply -f cloudflare-tunnel/cloudflaretunnelingress.yaml

# 5. Verificar status
kubectl get cloudflaretunnelingress -n homelab-services
```

Ver [DEPLOY.md](./DEPLOY.md) para guia completo.

## 📱 AgentApp Integration

A Mobile API suporta CloudEvents para integração com AgentApp:

```swift
// Configurar endpoint
let homelabURL = "https://api.music.lucena.cloud"

// Criar agentes
let djCollabAgent = Agent(
    id: "dj-collab-agent",
    name: "DJ Collab Assistant",
    endpoint: homelabURL + "/api/v1/cloudevents"
)
```

Ver [docs/AGENTAPP_INTEGRATION.md](./docs/AGENTAPP_INTEGRATION.md) para mais detalhes.

## ☁️ Cloudflare Tunnel

Todos os serviços são expostos via Cloudflare Tunnel usando `CloudflareTunnelIngress`:

- ✅ Sem necessidade de abrir portas no firewall
- ✅ TLS/SSL automático
- ✅ Proteção DDoS automática
- ✅ Acesso remoto seguro

Ver [docs/CLOUDFLARE_TUNNEL.md](./docs/CLOUDFLARE_TUNNEL.md) para mais detalhes.

## 📚 Documentação

- [DEPLOY.md](./DEPLOY.md) - Guia de deploy completo
- [QUICK_START.md](./QUICK_START.md) - Início rápido
- [docs/AGENTAPP_INTEGRATION.md](./docs/AGENTAPP_INTEGRATION.md) - Integração AgentApp
- [docs/CLOUDFLARE_TUNNEL.md](./docs/CLOUDFLARE_TUNNEL.md) - Cloudflare Tunnel
- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) - Arquitetura detalhada

## 🔧 Configuração

### Variáveis de Ambiente

```bash
# Gateway
GATEWAY_HOST=music.lucena.cloud
GATEWAY_PORT=8000

# Mobile API
MOBILE_API_HOST=api.music.lucena.cloud
MOBILE_API_PORT=8080

# Serviços
DJ_COLLAB_ENABLED=true
SPOTIFY_P2P_ENABLED=true
REKORDBOX_ENABLED=true
```

## 📊 Status dos Serviços

```bash
# Verificar todos os serviços
kubectl get pods -A | grep -E "homelab-services|dj-collab|spotify|rekordbox"

# Verificar Cloudflare Tunnel Ingress
kubectl get cloudflaretunnelingress -n homelab-services

# Verificar health
curl https://api.music.lucena.cloud/health
```

## 🎯 URLs Finais

- **Mobile API**: `https://api.music.lucena.cloud`
- **Kong Gateway**: `https://music.lucena.cloud`
- **DJ Collab P2P**: `https://dj-collab.music.lucena.cloud`
- **Spotify P2P**: `https://spotify.music.lucena.cloud`
- **rekordbox Cloud**: `https://rekordbox.music.lucena.cloud`

## 📝 Próximos Passos

1. **Desenvolver Backends**: Implementar lógica dos serviços
2. **Mobile Apps**: Criar apps mobile para cada serviço
3. **Monitoramento**: Configurar Prometheus/Grafana
4. **Backup**: Configurar backup automático
5. **Documentação**: Expandir documentação de APIs

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## 📄 Licença

MIT - Bruno Lucena (bruno@lucena.cloud)

---

**🏠 Seu homelab, seus dados, seu controle!**
