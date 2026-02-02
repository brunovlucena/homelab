# 🎵 Spotify API Setup

Para usar a busca no Spotify com a API oficial (opcional, mas recomendado).

## 🔑 Obter Credenciais

1. **Acessar Spotify Developer Dashboard:**
   - https://developer.spotify.com/dashboard

2. **Criar aplicação:**
   - Log in com sua conta Spotify
   - Create app
   - Preencher nome e descrição
   - Accept terms

3. **Obter credenciais:**
   - Client ID
   - Client Secret

## ⚙️ Configurar no Agente

### Kubernetes Secret

```bash
kubectl create secret generic agent-screenshot-secrets \
  -n ai \
  --from-literal=spotify-client-id='YOUR_CLIENT_ID' \
  --from-literal=spotify-client-secret='YOUR_CLIENT_SECRET'
```

### LambdaAgent YAML

```yaml
env:
  - name: SPOTIFY_CLIENT_ID
    valueFrom:
      secretKeyRef:
        name: agent-screenshot-secrets
        key: spotify-client-id
  - name: SPOTIFY_CLIENT_SECRET
    valueFrom:
      secretKeyRef:
        name: agent-screenshot-secrets
        key: spotify-client-secret
```

## 🚫 Sem Credenciais (Fallback)

Se não configurar as credenciais, o agente usa fallback:
- Constrói URL de busca do Spotify
- Retorna link para resultados
- Funciona, mas não retorna metadados estruturados

## 📊 Quotas da API

- **Rate Limits**: 
  - 300 requests per 30 seconds
  - 1800 requests per hour
- **Free tier**: Sem custo

## 🔐 Segurança

- **Nunca commitar credenciais no código**
- **Usar Secrets do Kubernetes**
- **Restringir redirect URIs** no Spotify Dashboard (se usando OAuth)
