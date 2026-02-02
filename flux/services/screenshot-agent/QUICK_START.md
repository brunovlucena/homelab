# 🚀 Quick Start - Screenshot Agent

## ✅ O Que Foi Implementado

### 1. ✅ Handler na mobile-api
- **Arquivo**: `flux/services/homelab-services/mobile-api/screenshot-handler.go`
- **Endpoint**: `POST /api/v1/screenshots`
- **Endpoint Status**: `GET /api/v1/screenshots/:id`

### 2. ✅ Estrutura do LambdaAgent
- **Diretório**: `flux/ai/agent-screenshot/`
- **LambdaAgent YAML**: `k8s/kustomize/base/lambdaagent.yaml`
- Configurado para receber eventos `screenshot.upload`

### 3. ✅ Processamento Básico
- **Handler**: `src/main.py` (FastAPI)
- Recebe CloudEvents do tipo `screenshot.upload`
- Processa cada screenshot isoladamente
- Retorna análise básica

## 🔧 Próximos Passos para Completar

### Passo 1: Build e Deploy do Agente

```bash
cd flux/ai/agent-screenshot

# Build local
make build-local

# Push para registry local
make push-local

# Deploy
make deploy-pro  # ou deploy-studio
```

### Passo 2: Configurar mobile-api

A mobile-api precisa enviar CloudEvents para o broker. Por enquanto, o handler envia diretamente para o agente. Para produção, você pode:

1. **Opção A**: Enviar para Knative Broker
   ```go
   brokerURL := "http://broker-knative-lambda.knative-lambda.svc.cluster.local"
   // Enviar CloudEvent para broker
   ```

2. **Opção B**: Enviar diretamente para o agente (atual)
   ```go
   agentURL := "http://agent-screenshot.ai.svc.cluster.local"
   ```

### Passo 3: Testar End-to-End

1. **Instalar extensão no browser**:
   ```bash
   cd flux/services/screenshot-agent/browser-extension/chrome
   # Abrir Chrome → Extensions → Developer mode → Load unpacked
   ```

2. **Configurar URL do agente**:
   - Clique no ícone da extensão
   - Configurações
   - URL: `http://localhost:8080/api/v1/screenshots` (local)
   - OU: `https://api.lucena.cloud/api/v1/screenshots` (produção)

3. **Capturar screenshot**:
   - Navegar para uma página
   - Clicar no ícone da extensão
   - Clicar em "Capturar Screenshot"

4. **Verificar processamento**:
   ```bash
   # Logs do agente
   kubectl logs -n ai -l app.kubernetes.io/name=agent-screenshot -f
   
   # Logs da mobile-api
   kubectl logs -n homelab-services -l app.kubernetes.io/name=mobile-api -f
   ```

## 🎯 Como Funciona (Cursor Agents Style)

1. **Browser Extension** → Envia screenshot para `/api/v1/screenshots`
2. **mobile-api** → Recebe, gera ID único, envia CloudEvent
3. **LambdaAgent** → Recebe evento, processa screenshot isoladamente
4. **Resultado** → Cada screenshot = agent run separado

## 📝 Notas

- **Processamento atual**: Básico (apenas estrutura)
- **Próximo**: Adicionar vision model (GPT-4V, Claude Vision, ou local)
- **Próximo**: Adicionar OCR (opcional)
- **Próximo**: Adicionar análise com LLM

## 🔍 Verificar Status

```bash
# Verificar se agente está rodando
kubectl get lambdaagent -n ai agent-screenshot

# Verificar pods
kubectl get pods -n ai -l app.kubernetes.io/name=agent-screenshot

# Ver logs
kubectl logs -n ai -l app.kubernetes.io/name=agent-screenshot
```

## 🐛 Troubleshooting

### Agente não recebe eventos
- Verificar se LambdaAgent está deployed
- Verificar eventing configuration no YAML
- Verificar broker/triggers no namespace `ai`

### Handler retorna erro
- Verificar logs da mobile-api
- Verificar se CloudEvent está sendo enviado corretamente
- Verificar formato do evento

### Screenshot não processa
- Verificar logs do agente
- Verificar se evento está chegando
- Verificar formato dos dados no evento
