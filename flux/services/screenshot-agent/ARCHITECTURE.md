# 📸 Screenshot Agent - Arquitetura Cursor Agents Style

## 🎯 Objetivo

Criar um sistema onde **cada screenshot inicia um agente diferente**, similar ao Cursor Agents. Cada screenshot é processado por uma instância isolada de agente.

## 🏗️ Arquitetura

```
Browser Extension
      ↓
[POST /api/v1/screenshots]
      ↓
Screenshot Agent Service
      ↓
[Para cada screenshot]
      ↓
LambdaFunction (instância única por screenshot)
      ↓
Agent Processa Screenshot
  - Análise de imagem (vision)
  - OCR (se necessário)
  - Extração de informações
  - Análise com LLM
      ↓
Resultado armazenado/retornado
```

## 📋 Componentes

### 1. Screenshot Handler Service
- Recebe uploads de screenshots
- Para cada screenshot recebido:
  - Gera ID único
  - Cria/invoca LambdaFunction com ID único
  - Passa screenshot como payload
- Retorna ID do agente/sessão

### 2. LambdaFunction (Template)
- Função serverless que processa screenshots
- Cada invocação = agente isolado
- Processa:
  - Análise de imagem (vision model)
  - OCR (texto na imagem)
  - Análise contextual com LLM
  - Extração de informações estruturadas

### 3. Storage
- MinIO/S3 para screenshots
- Redis/Postgres para metadados das sessões
- Resultados das análises

## 🔄 Fluxo Detalhado

1. **Upload Screenshot**
   ```
   POST /api/v1/screenshots
   FormData:
     - screenshot: file
     - url: string
     - title: string
   ```

2. **Criar Agente Instance**
   ```
   screenshot_id = generate_unique_id()
   agent_instance_id = f"screenshot-agent-{screenshot_id}"
   
   Invoke LambdaFunction:
     - Function: screenshot-processor
     - Payload: {
         screenshot_id,
         screenshot_url,
         metadata: {url, title, timestamp}
       }
   ```

3. **Processamento (dentro do LambdaFunction)**
   ```
   agent_instance = ScreenshotAgent(screenshot_id)
   
   # 1. Upload screenshot para MinIO
   screenshot_url = upload_to_minio(screenshot)
   
   # 2. Análise de imagem (vision)
   analysis = await vision_model.analyze(screenshot_url)
   
   # 3. OCR (se necessário)
   text = await ocr.extract(screenshot_url)
   
   # 4. Análise contextual com LLM
   context = await llm.analyze({
     image_analysis: analysis,
     text: text,
     metadata: metadata
   })
   
   # 5. Salvar resultado
   await save_result(screenshot_id, {
     analysis,
     text,
     context
   })
   ```

4. **Retornar Resultado**
   ```
   GET /api/v1/screenshots/{screenshot_id}/result
   ```

## 🛠️ Implementação

### Opção 1: LambdaFunction (Recomendado)
- Usa Knative Lambda Operator
- Cada invocação = instância isolada
- Scale-to-zero automático
- Payload único por screenshot

### Opção 2: LambdaAgent + Events
- LambdaAgent persiste
- Cada screenshot envia CloudEvent
- Agente processa eventos em fila
- Menos isolamento, mais eficiente

### Opção 3: Kubernetes Jobs
- Job por screenshot
- Máximo isolamento
- Mais overhead, menos escalável

## 🎨 Como Cursor Agents

No Cursor Agents:
- Cada task/issue = novo agent run
- Histórico de runs
- Cada run é isolado
- Status tracking

No nosso sistema:
- Cada screenshot = novo agent instance
- Histórico de screenshots processados
- Cada instance é isolado (LambdaFunction)
- Status tracking via API

## 📊 Métricas e Observabilidade

- Screenshots recebidos/processados
- Tempo de processamento por screenshot
- Taxa de sucesso/erro
- Uso de recursos (vision, LLM)
- Custo por screenshot

## 🔐 Segurança

- Autenticação no endpoint
- Validação de tamanho/formato
- Rate limiting
- Isolamento entre screenshots
- PII/data sanitization

## 🚀 Próximos Passos

1. Criar screenshot-agent-service
2. Criar LambdaFunction template
3. Integrar vision model (GPT-4V, Claude Vision, ou local)
4. Implementar OCR
5. Criar API de resultados
6. Dashboard para visualizar histórico
