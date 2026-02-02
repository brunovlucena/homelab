# 🏗️ Arquitetura - Homelab Services

## Visão Geral

O Homelab Services é uma plataforma unificada que permite rodar múltiplos serviços de música/streaming no seu homelab Kubernetes, acessível via mobile e web.

## Princípios de Design

1. **Homelab como Servidor**: Tudo roda no seu homelab, você tem controle total
2. **Acesso Remoto**: Mobile/web se conectam ao homelab via internet
3. **Serviços Modulares**: Cada serviço é independente mas compartilha infraestrutura
4. **API Unificada**: Gateway único para todos os serviços
5. **Autenticação Centralizada**: Um login para todos os serviços

## Componentes Principais

### 1. API Gateway (Kong)

**Responsabilidades:**
- Roteamento de requisições para serviços
- Autenticação e autorização (JWT)
- Rate limiting
- CORS
- Load balancing

**Configuração:**
```yaml
services:
  - name: dj-collab-p2p
    url: http://dj-collab-p2p-server:8080
    routes:
      - paths: ["/api/v1/dj-collab"]
```

### 2. Mobile API

**Responsabilidades:**
- Endpoint unificado para mobile
- Service discovery
- Agregação de dados de múltiplos serviços
- Cache e otimização

**Endpoints:**
```
GET  /api/v1/services          # Lista serviços disponíveis
GET  /api/v1/dashboard         # Dashboard agregado
POST /api/v1/auth/login        # Autenticação
```

### 3. Serviços

#### DJ Collab P2P
- Streaming P2P entre DJs
- Colaboração em tempo real
- Gamificação

#### Spotify P2P
- Streaming de biblioteca pessoal
- Estações P2P
- Descoberta descentralizada

#### rekordbox Cloud
- Sincronização de biblioteca
- Análise de música
- Cloud sync P2P

#### Library Manager
- Gerenciamento de biblioteca
- Análise automática
- Organização inteligente

### 4. Infraestrutura Compartilhada

#### MongoDB
- Usuários
- Sessões
- Configurações
- Metadados

#### Redis
- Cache de sessões
- Real-time pub/sub
- Rate limiting
- Session storage

#### IPFS
- Distribuição de conteúdo
- Metadados de música
- Sets gravados
- Compartilhamento P2P

#### MinIO (S3-compatible)
- Arquivos de música
- Artworks
- Backups
- Cache

#### PostgreSQL
- Metadados estruturados
- Análises de música
- Estatísticas
- Relatórios

## Fluxo de Dados

### 1. Autenticação

```
Mobile App
    ↓
POST /api/v1/auth/login
    ↓
Auth Service
    ↓
JWT Token
    ↓
Mobile App (armazena token)
```

### 2. Acesso a Serviço

```
Mobile App
    ↓
GET /api/v1/dj-collab/sessions
    ↓
Kong Gateway (valida JWT)
    ↓
DJ Collab Service
    ↓
MongoDB/Redis
    ↓
Response
```

### 3. Streaming P2P

```
DJ A (Mobile)
    ↓
Cria sessão via API
    ↓
WebRTC signaling via Gateway
    ↓
Conexão P2P direta (bypass Gateway)
    ↓
Streaming de áudio
```

## Segurança

### Autenticação
- JWT tokens com expiração
- Refresh tokens
- OAuth2 para serviços externos (opcional)

### Autorização
- RBAC por serviço
- Permissões granulares
- API keys para serviços

### Comunicação
- TLS/SSL obrigatório
- Criptografia em trânsito
- Criptografia em repouso (secrets)

### Acesso
- VPN recomendado
- Cloudflare Tunnel (opcional)
- Firewall rules
- Rate limiting

## Escalabilidade

### Horizontal
- Múltiplas réplicas de serviços
- Load balancing automático
- Auto-scaling baseado em métricas

### Vertical
- Ajuste de recursos por serviço
- Resource quotas
- Priority classes

### Storage
- Persistent volumes
- Storage classes
- Backup automático

## Monitoramento

### Métricas
- Prometheus para métricas
- Grafana para visualização
- Alertas via Alertmanager

### Logs
- Loki para agregação
- Grafana para visualização
- Retenção configurável

### Tracing
- Jaeger para distributed tracing
- OpenTelemetry para instrumentação

## Backup e Recuperação

### Dados
- Backup automático de bancos
- Snapshots de volumes
- Replicação geográfica (opcional)

### Configuração
- GitOps com Flux
- Versionamento de configs
- Rollback automático

## Desenvolvimento

### Local
```bash
# Usar docker-compose para desenvolvimento local
docker-compose up -d

# Serviços disponíveis em localhost
```

### Homelab
```bash
# Deploy via Flux
kubectl apply -k flux/homelab-services/

# Ou manual
kubectl apply -f k8s/
```

## Próximos Passos

1. Implementar serviços individuais
2. Criar mobile app completo
3. Adicionar monitoramento
4. Configurar backup
5. Documentar APIs

---

**🏠 Seu homelab, seus dados, seu controle!**
