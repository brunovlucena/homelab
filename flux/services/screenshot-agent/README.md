# 📸 Screenshot Agent - Sistema de Agentes por Screenshot

Sistema que funciona como **Cursor Agents**: cada screenshot enviado pela extensão inicia um agente diferente para análise.

## 🎯 Como Funciona (Como Cursor Agents)

No Cursor Agents:
- Cada task/issue = novo agent run
- Histórico de runs visível
- Cada run é isolado
- Status tracking individual

No nosso sistema:
- **Cada screenshot = novo agent instance**
- Histórico de screenshots processados
- Cada instance é isolado
- Status tracking individual

## 🏗️ Arquitetura

```
Browser Extension
      ↓
[POST /api/v1/screenshots]
      ↓
Screenshot Agent Service
      ↓
[Para cada screenshot - ID único]
      ↓
LambdaAgent (processa em fila) OU LambdaFunction (instância isolada)
      ↓
Análise:
  - Vision (GPT-4V, Claude Vision, ou local)
  - OCR (se necessário)
  - LLM Analysis
      ↓
Resultado armazenado/retornado
```

## 📋 O Que Você Precisa Fazer

### 1. Escolher Arquitetura

**Opção A: LambdaAgent + Events (Recomendado para começar)**
- Mais simples
- Reutiliza infraestrutura existente
- Processa screenshots em fila
- Um agente processa múltiplos screenshots

**Opção B: LambdaFunction por Screenshot**
- Máximo isolamento
- Cada screenshot = função isolada
- Mais complexo de implementar

### 2. Implementar Serviço

Você tem duas opções:

#### Opção 1: Integrar na mobile-api (Mais rápido)

Adicionar handler na `mobile-api/main.go`:

```go
api.POST("/screenshots", handleScreenshot)
```

#### Opção 2: Serviço Dedicado (Mais isolado)

Criar serviço separado em `screenshot-service/` (já criado como exemplo)

### 3. Criar Agente de Processamento

Criar LambdaAgent similar aos outros agentes do homelab:

```bash
cd flux/ai
./scripts/create-agent.sh agent-screenshot
```

O agente deve:
- Receber screenshots (via CloudEvent ou HTTP)
- Processar imagem (vision, OCR)
- Analisar com LLM
- Retornar/salvar resultado

### 4. Configurar Infraestrutura

- **MinIO/S3**: Para armazenar screenshots
- **LambdaAgent**: Para processar screenshots
- **Vision Model**: GPT-4V, Claude Vision, ou modelo local (llava, etc.)
- **OCR**: Tesseract ou serviço cloud (opcional)

## 📂 Estrutura de Arquivos

```
screenshot-agent/
├── browser-extension/          # ✅ Já criado
│   ├── chrome/
│   └── safari/
├── screenshot-service/         # ✅ Exemplo criado
│   └── main.go
├── ARCHITECTURE.md            # ✅ Documentação
├── IMPLEMENTATION.md          # ✅ Guia de implementação
└── README.md                  # Este arquivo
```

## 🚀 Passos para Implementar

### Passo 1: Integrar Handler na mobile-api

1. Adicionar handler `handleScreenshot` na `mobile-api`
2. Receber upload de screenshot
3. Salvar no MinIO
4. Enviar CloudEvent para agente

### Passo 2: Criar Agent-Screenshot

```bash
cd flux/ai
./scripts/create-agent.sh agent-screenshot
```

### Passo 3: Implementar Processamento

No agente:
- Receber CloudEvent com screenshot URL
- Processar imagem (vision)
- Analisar com LLM
- Salvar resultado

### Passo 4: Configurar Vision Model

Escolher um:
- **Cloud**: GPT-4V (OpenAI), Claude Vision (Anthropic)
- **Local**: LLaVA (Ollama), etc.

### Passo 5: Testar End-to-End

1. Instalar extensão no browser
2. Capturar screenshot
3. Verificar agente processar
4. Consultar resultado

## 📚 Referências

- **Outros agentes**: `flux/ai/agent-*`
- **LambdaAgent examples**: `flux/ai/agent-chat/k8s/`
- **CloudEvents**: `mobile-api/agentapp-handler.go`
- **Script de criação**: `flux/ai/scripts/create-agent.sh`

## 🔧 Exemplo de Integração Rápida

Para integrar rapidamente na `mobile-api`:

1. Adicionar handler (veja `screenshot-handler.go` como referência)
2. Criar agente: `./scripts/create-agent.sh agent-screenshot`
3. Implementar processamento no agente
4. Testar!

## ❓ Dúvidas?

Veja `IMPLEMENTATION.md` para mais detalhes sobre as opções de implementação.
