# 🚀 Quick Start Guide - DJ Collab P2P Game

## 📋 Pré-requisitos

- Node.js 18+ e pnpm 8+
- Go 1.21+
- Docker e Docker Compose
- Kubernetes cluster (para deploy no homelab) - opcional

## 🎯 Visão Geral

Este projeto permite que DJs façam streaming P2P das suas músicas e colaborem em tempo real para criar sets juntos, sem necessidade de servidores centralizados para o áudio.

## 🏃 Início Rápido

### 1. Clone e Instale

```bash
cd flux/dj-collab-p2p
pnpm install
cd apps/server && go mod download
```

### 2. Inicie a Infraestrutura

```bash
# Inicia MongoDB, Redis, IPFS e Coturn (STUN/TURN)
make docker-up

# Ou manualmente:
docker-compose up -d
```

### 3. Inicie o Servidor de Coordenação

```bash
# Terminal 1
cd apps/server
go run main.go

# O servidor estará disponível em http://localhost:8080
```

### 4. Inicie o App Web (quando disponível)

```bash
# Terminal 2
cd apps/web
pnpm dev

# O app estará disponível em http://localhost:3000
```

## 🎧 Como Funciona

### Fluxo Básico

1. **DJ A cria uma sessão**
   - Seleciona músicas da biblioteca local
   - Sistema analisa BPM e key automaticamente
   - Sessão é indexada no DHT/IPFS

2. **DJ B encontra e entra na sessão**
   - Busca sessões disponíveis
   - Conecta via WebRTC (P2P)
   - Recebe stream de áudio diretamente do DJ A

3. **Colaboração em Tempo Real**
   - Ambos podem controlar BPM, key, efeitos
   - Estado sincronizado via WebSocket
   - Mixagem colaborativa

4. **Gamificação**
   - Pontos por transições suaves
   - Conquistas desbloqueadas
   - Leaderboards

## 🔧 Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz:

```bash
# Servidor de Coordenação
COORDINATION_HOST=localhost
COORDINATION_PORT=8080

# MongoDB
MONGODB_URI=mongodb://localhost:27017

# Redis
REDIS_URI=redis://localhost:6379

# IPFS
IPFS_API_URL=http://localhost:5001
IPFS_GATEWAY_URL=http://localhost:8080

# Signaling (STUN/TURN)
SIGNALING_HOST=localhost
SIGNALING_PORT=3478
STUN_SERVERS=stun:stun.l.google.com:19302

# P2P
P2P_ENABLED=true
```

## 🎮 Uso Básico

### Criar Sessão

```typescript
import { P2PEngine } from '@dj-collab-p2p/p2p-engine';

const engine = new P2PEngine({
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' }
  ],
  signalingServer: 'ws://localhost:8080'
});

await engine.initialize();
```

### Conectar a Sessão

```typescript
// DJ A (Host)
const offer = await engine.createOffer();
// Enviar offer para servidor de coordenação

// DJ B (Participant)
await engine.handleOffer(offer);
const answer = await engine.createAnswer();
// Enviar answer para servidor de coordenação
```

### Sincronizar Estado

```typescript
// Via WebSocket
const ws = new WebSocket('ws://localhost:8080/api/v1/sessions/123/ws');

ws.send(JSON.stringify({
  type: 'state',
  sessionId: '123',
  data: {
    bpm: 128,
    key: 'C Major',
    position: 45.5,
    isPlaying: true
  }
}));
```

## 🐳 Deploy no Homelab

### Via Flux (GitOps)

```bash
# Aplicar configuração Kubernetes
kubectl apply -k k8s/

# Ou usar Flux CLI
flux reconcile source git dj-collab-p2p
```

### Manual

```bash
# Criar namespace
kubectl create namespace dj-collab-p2p

# Aplicar manifests
kubectl apply -f k8s/
```

## 🧪 Testes

```bash
# Todos os testes
make test

# Testes específicos
pnpm test:server    # Backend Go
pnpm test:desktop   # Desktop app
pnpm test:mobile    # Mobile app
pnpm test:p2p       # P2P integration
```

## 📚 Próximos Passos

1. **Explorar o código**
   - `apps/server/` - Backend Go
   - `packages/p2p-engine/` - Engine P2P
   - `packages/shared/` - Tipos compartilhados

2. **Ler a documentação**
   - `README.md` - Visão geral
   - `docs/business/dj-collab-p2p-game.md` - Análise de negócio

3. **Contribuir**
   - Criar issues
   - Fazer pull requests
   - Melhorar documentação

## 🐛 Troubleshooting

### Problemas com P2P

- **Conexão não estabelece**: Verifique STUN/TURN servers
- **Áudio não funciona**: Verifique permissões de microfone
- **Latência alta**: Use servidores TURN mais próximos

### Problemas com Infraestrutura

- **MongoDB não conecta**: Verifique `docker-compose.yml`
- **Redis não conecta**: Verifique porta 6379
- **IPFS não funciona**: Verifique portas 4001, 5001, 8080

## 💡 Dicas

- Use servidores STUN/TURN públicos para testes
- Para produção, configure seus próprios servidores TURN
- Use IPFS para distribuir metadados e sets gravados
- Cache análises musicais localmente para performance

## 📞 Suporte

- Issues: GitHub Issues
- Email: bruno@lucena.cloud

---

**Happy DJing! 🎧**
