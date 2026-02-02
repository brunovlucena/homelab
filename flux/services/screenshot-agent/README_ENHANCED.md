# 📸 Screenshot Agent - Enhanced Version

Sistema avançado de análise de screenshots com OCR, Vision Model, LLM, e múltiplas integrações.

## 🎯 Funcionalidades

### ✅ Implementado

1. **OCR (Optical Character Recognition)**
   - EasyOCR (multilíngue: inglês, português)
   - Tesseract (fallback)
   - Extração de texto de screenshots

2. **Vision Model**
   - LLaVA (Ollama local) - padrão
   - GPT-4V (OpenAI) - opcional
   - Claude Vision (Anthropic) - opcional
   - Análise visual de screenshots

3. **LLM Analysis**
   - Ollama (local) para análise de contexto
   - Entendimento semântico
   - Sugestão de queries otimizadas
   - Detecção de ações

4. **Busca Multi-plataforma**
   - YouTube Search
   - Spotify Search
   - SoundCloud Search

## 🔄 Pipeline de Processamento

```
Screenshot Recebido
   ↓
1. Extrair Contexto Básico (URL, título, metadados)
   ↓
2. OCR: Extrair texto da imagem (se disponível)
   ↓
3. Vision Model: Analisar imagem (se disponível)
   ↓
4. LLM: Entender contexto e sugerir ações/queries
   ↓
5. Detectar Ações (LLM + padrões)
   ↓
6. Executar Ações:
   - YouTube Search
   - Spotify Search
   - SoundCloud Search
   ↓
7. Retornar Resultados Consolidados
```

## 📦 Estrutura de Código

```
src/
├── main.py              # FastAPI app, CloudEvents handler
├── analyzer.py          # Análise básica de contexto
├── ocr.py               # OCR (EasyOCR/Tesseract)
├── vision.py            # Vision models (LLaVA/GPT-4V/Claude)
├── llm_analysis.py      # LLM para análise de contexto
├── youtube_search.py    # Busca no YouTube
├── spotify_search.py    # Busca no Spotify
└── soundcloud_search.py # Busca no SoundCloud
```

## ⚙️ Configuração

### Variáveis de Ambiente

#### Obrigatórias (com defaults)
- `OLLAMA_URL`: URL do Ollama (default: `http://ollama-native.ollama.svc.cluster.local:11434`)
- `OLLAMA_MODEL`: Modelo Ollama (default: `llama3.2:3b`)

#### Opcionais (com fallbacks)
- `YOUTUBE_API_KEY`: YouTube Data API v3 key
- `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`: Spotify API credentials
- `OPENAI_API_KEY`: Para GPT-4V
- `ANTHROPIC_API_KEY`: Para Claude Vision
- `VISION_MODEL`: Modelo vision (default: `llava:7b`)

### Kubernetes

Ver:
- `YOUTUBE_SETUP.md` - Configurar YouTube API
- `SPOTIFY_SETUP.md` - Configurar Spotify API
- LambdaAgent YAML - Configurar Ollama e outros

## 🚀 Uso

### Exemplo: Screenshot de Instagram Post

1. **Usuário captura screenshot** de post sobre concerto
2. **Agente processa:**
   - OCR extrai texto (descrição, comentários)
   - Vision analisa imagem (identifica tipo de conteúdo)
   - LLM entende contexto (artista, evento)
   - Detecta ações: YouTube, Spotify
3. **Busca e retorna:**
   - Vídeos do YouTube
   - Tracks do Spotify
   - Resultados do SoundCloud

### Resultado

```json
{
  "screenshot_id": "scr_abc123",
  "context": {
    "basic": {
      "url": "instagram.com/p/xyz",
      "title": "Simpsons DJ Concert",
      "artist": "Simpsons",
      "content_type": "concert"
    },
    "ocr": {
      "text": "Amazing DJ set...",
      "confidence": 0.95
    },
    "vision": {
      "description": "Instagram post showing DJ set...",
      "method": "llava"
    }
  },
  "actions_executed": ["youtube_search", "spotify_search"],
  "results": {
    "youtube_search": {
      "query": "Simpsons DJ set live",
      "results": [...]
    },
    "spotify_search": {
      "query": "Simpsons",
      "results": [...]
    }
  }
}
```

## 📚 Documentação

- `ACTIONS.md` - Ações suportadas
- `IMPROVEMENTS.md` - Melhorias implementadas
- `YOUTUBE_SETUP.md` - Setup YouTube API
- `SPOTIFY_SETUP.md` - Setup Spotify API
- `EXAMPLE_USAGE.md` - Exemplos de uso

## 🔮 Próximos Passos

- [ ] Download automático de imagem do MinIO
- [ ] Cache de resultados OCR/Vision
- [ ] Mais plataformas (Bandcamp, Apple Music, etc.)
- [ ] NER (Named Entity Recognition)
- [ ] Análise de sentimento
- [ ] Histórico de screenshots processados
