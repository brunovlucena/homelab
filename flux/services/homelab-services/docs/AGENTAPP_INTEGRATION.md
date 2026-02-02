# 📱 AgentApp Integration - Homelab Services

## 🎯 Visão Geral

A Mobile API do Homelab Services foi adaptada para atender o **AgentApp**, um framework iOS para criar apps de chat com agentes AI usando CloudEvents.

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────┐
│              📱 AgentApp (iOS)                          │
│                                                         │
│  • AgentAppCore Framework                              │
│  • CloudEvents Communication                           │
│  • Multiple Agent Support                              │
└────────────────────┬────────────────────────────────────┘
                     │
                     │ CloudEvents (HTTPS)
                     │
┌────────────────────▼────────────────────────────────────┐
│      🌐 Mobile API (Homelab Services)                    │
│                                                         │
│  • CloudEvents Handler                                  │
│  • Service Router                                       │
│  • Agent Message Processor                              │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────▼──────┐ ┌──▼──────┐ ┌───▼──────┐
│ 🎧 DJ Collab │ │ 🎵 Spotify│ │ 📀 rekordbox│
│   P2P        │ │   P2P     │ │  Cloud    │
└──────────────┘ └──────────┘ └───────────┘
```

## 🔌 CloudEvents Support

### Event Types Suportados

1. **agent.message**
   - Mensagens de agentes para serviços
   - Processamento e roteamento automático

2. **agent.request**
   - Requisições de agentes para ações
   - Roteamento para serviços apropriados

3. **agent.status**
   - Verificação de status de serviços
   - Health checks

4. **service.discovery**
   - Descoberta de serviços disponíveis
   - Lista de serviços e endpoints

### Formato CloudEvent

```json
{
  "specversion": "1.0",
  "type": "agent.message",
  "source": "agentapp/ios",
  "id": "event-123",
  "time": "2025-01-27T10:00:00Z",
  "datacontenttype": "application/json",
  "data": {
    "conversationId": "conv-123",
    "agentId": "dj-collab-agent",
    "userId": "user-123",
    "content": "Criar nova sessão de DJ",
    "timestamp": "2025-01-27T10:00:00Z"
  }
}
```

## 🎧 Agentes Disponíveis

### DJ Collab Agent
- **Agent ID**: `dj-collab-agent`
- **Serviço**: DJ Collab P2P
- **Capabilities**:
  - Criar sessões colaborativas
  - Conectar a sessões existentes
  - Gerenciar playlists
  - Sincronizar BPM e key

### Spotify P2P Agent
- **Agent ID**: `spotify-agent`
- **Serviço**: Spotify P2P
- **Capabilities**:
  - Buscar estações
  - Criar estações pessoais
  - Reproduzir música
  - Gerenciar biblioteca

### rekordbox Agent
- **Agent ID**: `rekordbox-agent`
- **Serviço**: rekordbox Cloud
- **Capabilities**:
  - Sincronizar biblioteca
  - Analisar músicas (BPM, key)
  - Gerenciar playlists
  - Exportar sets

### Library Agent
- **Agent ID**: `library-agent`
- **Serviço**: Library Manager
- **Capabilities**:
  - Buscar músicas
  - Upload de arquivos
  - Organização inteligente
  - Análise automática

## 📡 Endpoints

### CloudEvents Endpoint

```
POST /api/v1/cloudevents
Content-Type: application/json

{
  "specversion": "1.0",
  "type": "agent.message",
  ...
}
```

### Service Discovery

```
GET /api/v1/services

Response:
{
  "services": [
    {
      "id": "dj-collab",
      "name": "DJ Collab P2P",
      "description": "Streaming P2P e colaboração",
      "enabled": true,
      "endpoint": "/api/v1/dj-collab"
    }
  ]
}
```

### Agent Messages

```
POST /api/v1/agents/:agentId/messages

{
  "conversationId": "conv-123",
  "content": "Criar sessão",
  "metadata": {}
}
```

## 🔧 Configuração no AgentApp

### 1. Configurar Endpoint

```swift
import AgentAppCore

// Configurar endpoint do homelab
let homelabURL = "https://api.music.lucena.cloud"

// Criar AgentService com endpoint customizado
let agentService = AgentService(
    endpoint: homelabURL + "/api/v1/cloudevents",
    userId: currentUserId
)
```

### 2. Criar Agentes

```swift
// DJ Collab Agent
let djCollabAgent = Agent(
    id: "dj-collab-agent",
    name: "DJ Collab Assistant",
    description: "Ajuda com sessões colaborativas de DJ",
    endpoint: homelabURL + "/api/v1/agents/dj-collab-agent"
)

// Spotify P2P Agent
let spotifyAgent = Agent(
    id: "spotify-agent",
    name: "Spotify P2P Assistant",
    description: "Ajuda com streaming P2P",
    endpoint: homelabURL + "/api/v1/agents/spotify-agent"
)
```

### 3. Enviar Mensagens

```swift
// Enviar mensagem via CloudEvent
let event = CloudEvent(
    specVersion: "1.0",
    type: "agent.message",
    source: "agentapp/ios",
    id: UUID().uuidString,
    time: Date(),
    data: [
        "conversationId": conversationId,
        "agentId": agent.id,
        "userId": userId,
        "content": messageText
    ]
)

try await agentService.sendEvent(event)
```

## 🚀 Exemplo de Uso

### Criar Sessão DJ Collab via Agent

```swift
// No AgentApp
let message = "Criar uma nova sessão de DJ colaborativa"

// AgentApp envia CloudEvent
let event = CloudEvent(
    type: "agent.message",
    source: "agentapp/ios",
    data: [
        "agentId": "dj-collab-agent",
        "content": message
    ]
)

// Mobile API recebe e processa
// Roteia para DJ Collab service
// Retorna resposta via CloudEvent
```

### Resposta

```json
{
  "specversion": "1.0",
  "type": "agent.response",
  "source": "homelab-services/mobile-api",
  "id": "response-123",
  "time": "2025-01-27T10:00:01Z",
  "data": {
    "conversationId": "conv-123",
    "agentId": "dj-collab-agent",
    "response": {
      "sessionId": "session-456",
      "status": "created",
      "message": "Sessão criada com sucesso"
    }
  }
}
```

## 🔐 Autenticação

### JWT Token

```swift
// Adicionar token no header
let headers = [
    "Authorization": "Bearer \(jwtToken)",
    "Content-Type": "application/json",
    "ce-specversion": "1.0"
]
```

### No Mobile API

```go
// Validar JWT token
func validateToken(c *gin.Context) {
    token := c.GetHeader("Authorization")
    // Validar e extrair user ID
}
```

## 📊 Monitoramento

### Métricas

- Eventos CloudEvents recebidos
- Tempo de processamento
- Taxa de sucesso
- Erros por tipo de evento

### Logs

```go
log.Printf("CloudEvent received: type=%s, id=%s", event.Type, event.ID)
log.Printf("Routed to service: %s", service)
log.Printf("Response sent: id=%s", responseID)
```

## 🐛 Troubleshooting

### Evento não processado

1. Verificar formato CloudEvent
2. Verificar tipo de evento suportado
3. Verificar logs do Mobile API
4. Verificar conectividade com serviços

### Agente não encontrado

1. Verificar mapeamento agentId -> service
2. Verificar se serviço está ativo
3. Verificar endpoint do serviço

### Timeout

1. Verificar latência de rede
2. Verificar se serviços estão respondendo
3. Aumentar timeout se necessário

## 📚 Referências

- [CloudEvents Specification](https://cloudevents.io/)
- [AgentApp Documentation](../../AgentApp/README.md)
- [Mobile API Documentation](./MOBILE_API.md)

---

**🎧 Integração completa entre AgentApp e Homelab Services!**
