# 🚀 Screenshot Agent - Guia de Implementação

## 📝 O Que Você Precisa Fazer

Para implementar o sistema de agentes por screenshot (estilo Cursor Agents), você precisa:

### 1. Criar o Serviço de Recebimento

**Arquivo:** `screenshot-agent-service/main.go`

Este serviço:
- Recebe uploads de screenshots da extensão
- Para cada screenshot, cria/invoca uma LambdaFunction
- Retorna ID da sessão do agente

### 2. Criar LambdaFunction para Processamento

**Arquivo:** `lambda-function-screenshot-processor/`

Esta função:
- É invocada para cada screenshot
- Processa a imagem (vision, OCR)
- Analisa com LLM
- Salva resultados

### 3. Configurar Infraestrutura

- MinIO/S3 para armazenar screenshots
- Redis/Postgres para metadados (opcional)
- LambdaFunction template no Kubernetes

## 🔧 Opções de Implementação

### Opção A: LambdaFunction (Recomendado - Mais Isolado)

Cada screenshot invoca uma LambdaFunction isolada.

**Vantagens:**
- ✅ Isolamento total por screenshot
- ✅ Scale-to-zero automático
- ✅ Payload único por invocação

**Como fazer:**
1. Criar LambdaFunction template
2. Serviço invoca função com payload único
3. Função processa e retorna resultado

### Opção B: LambdaAgent + Event Queue (Mais Eficiente)

Um LambdaAgent persiste e processa screenshots em fila.

**Vantagens:**
- ✅ Mais eficiente (reutiliza instância)
- ✅ Melhor para alto volume
- ✅ Fila garante processamento

**Como fazer:**
1. Criar LambdaAgent para screenshots
2. Serviço envia CloudEvent por screenshot
3. Agente processa eventos da fila

### Opção C: Kubernetes Jobs (Máximo Isolamento)

Job do Kubernetes por screenshot.

**Vantagens:**
- ✅ Máximo isolamento
- ✅ Recursos dedicados

**Desvantagens:**
- ❌ Mais overhead
- ❌ Menos escalável

## 🎯 Implementação Recomendada (Opção B)

Para começar rápido, recomendo **Opção B** (LambdaAgent + Events):

1. **Criar LambdaAgent** (similar aos outros agentes do homelab)
2. **Serviço envia CloudEvent** por screenshot
3. **Agente processa** screenshots da fila

Isso é mais simples e reutiliza a infraestrutura existente.

## 📂 Estrutura de Arquivos

```
screenshot-agent/
├── screenshot-service/          # Serviço que recebe screenshots
│   ├── main.go
│   ├── handler.go
│   └── Dockerfile
├── lambda-agent-screenshot/     # Agente que processa screenshots
│   ├── src/
│   │   ├── handler.py
│   │   ├── vision_processor.py
│   │   └── main.py
│   └── k8s/
│       └── lambdaagent.yaml
└── browser-extension/           # Já criado
```

## 🔄 Fluxo Simplificado

```
1. Browser Extension → POST /api/v1/screenshots
2. Service recebe screenshot
3. Service salva screenshot no MinIO
4. Service envia CloudEvent para agente
5. LambdaAgent processa screenshot:
   - Vision analysis
   - OCR
   - LLM analysis
6. Resultado salvo/retornado
```

## 📋 Checklist de Implementação

- [ ] Criar serviço screenshot-service
- [ ] Integrar com mobile-api OU criar serviço dedicado
- [ ] Criar LambdaAgent para processar screenshots
- [ ] Configurar MinIO para armazenar screenshots
- [ ] Integrar vision model (GPT-4V, Claude, ou local)
- [ ] Implementar OCR (Tesseract, ou cloud)
- [ ] Criar API para consultar resultados
- [ ] Adicionar métricas e observabilidade
- [ ] Testar end-to-end

## 🎨 Integração com Mobile-API

Você pode:
1. **Adicionar handler na mobile-api** (mais simples)
2. **Criar serviço dedicado** (mais isolado)

Recomendo começar integrando na mobile-api, depois separar se necessário.

## 📚 Referências

- Ver outros agentes: `flux/ai/agent-*`
- LambdaAgent examples: `flux/ai/agent-chat/k8s/`
- CloudEvents: Ver `mobile-api/agentapp-handler.go`
