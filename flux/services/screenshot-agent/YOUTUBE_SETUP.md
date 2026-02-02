# 📺 YouTube API Setup

Para usar a busca no YouTube com a API oficial (opcional, mas recomendado).

## 🔑 Obter API Key

1. **Acessar Google Cloud Console:**
   - https://console.cloud.google.com/

2. **Criar projeto** (ou usar existente)

3. **Habilitar YouTube Data API v3:**
   - APIs & Services → Library
   - Buscar "YouTube Data API v3"
   - Clicar em "Enable"

4. **Criar credenciais:**
   - APIs & Services → Credentials
   - Create Credentials → API Key
   - Copiar a chave

## ⚙️ Configurar no Agente

### Opção 1: Variável de Ambiente

No Kubernetes (LambdaAgent YAML):

```yaml
env:
  - name: YOUTUBE_API_KEY
    valueFrom:
      secretKeyRef:
        name: agent-screenshot-secrets
        key: youtube-api-key
```

Criar secret:

```bash
kubectl create secret generic agent-screenshot-secrets \
  -n ai \
  --from-literal=youtube-api-key='YOUR_API_KEY_HERE'
```

### Opção 2: ConfigMap (menos seguro)

```yaml
env:
  - name: YOUTUBE_API_KEY
    valueFrom:
      configMapKeyRef:
        name: agent-screenshot-config
        key: youtube-api-key
```

## 🚫 Sem API Key (Fallback)

Se não configurar a API key, o agente usa um fallback:
- Constrói URL de busca do YouTube
- Retorna link para resultados
- Funciona, mas não retorna metadados estruturados

## 📊 Quotas da API

- **Free tier**: 10,000 units/day
- Cada search = 100 units
- ~100 buscas/dia (grátis)

Para produção, considerar upgrade do quota.

## 🔐 Segurança

- **Nunca commitar API key no código**
- **Usar Secrets do Kubernetes**
- **Restringir API key** no Google Cloud Console:
  - Application restrictions
  - API restrictions (só YouTube Data API v3)
