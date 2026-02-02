# 🎯 Screenshot Agent - Actions

O agente pode executar ações baseadas no conteúdo dos screenshots.

## 🔍 Ações Suportadas

### 1. YouTube Search (`youtube_search`)

Busca vídeos no YouTube baseado no contexto extraído do screenshot.

**Como funciona:**
1. Agente analisa o screenshot
2. Extrai informações relevantes (artista, evento, descrição)
3. Constrói query de busca otimizada
4. Busca no YouTube
5. Retorna resultados

**Exemplos de triggers:**
- Screenshot de Instagram com post sobre um concerto
- Título contém "concert", "live", "performance"
- Descrição menciona artista/evento
- Texto explícito: "find in youtube this concert"

**Resultado:**
```json
{
  "actions_executed": ["youtube_search"],
  "results": {
    "youtube_search": {
      "query": "Artist Name live concert",
      "results": [
        {
          "video_id": "abc123",
          "title": "Artist Name - Live at Venue",
          "url": "https://www.youtube.com/watch?v=abc123",
          "channel": "Channel Name",
          "thumbnail": "..."
        }
      ],
      "count": 5
    }
  }
}
```

## 🔧 Configuração

### YouTube API Key (Opcional)

Para usar a YouTube Data API v3 (recomendado):

```bash
# Obter API key em: https://console.cloud.google.com/apis/credentials
export YOUTUBE_API_KEY="your-api-key-here"
```

**Sem API Key:**
O agente usa fallback que retorna URL de busca do YouTube (funciona, mas menos preciso).

## 📝 Exemplo de Uso

### Screenshot de Instagram Post

**Input:**
- URL: `instagram.com/p/xyz`
- Título: "Simpsons DJ Concert"
- Descrição: "Amazing DJ set with Simpsons characters..."

**Processamento:**
1. Extrai contexto: `{artist: "Simpsons", content_type: "concert", ...}`
2. Detecta ação: `youtube_search`
3. Constrói query: "Simpsons DJ set live"
4. Busca no YouTube
5. Retorna vídeos encontrados

**Output:**
```json
{
  "screenshot_id": "scr_abc123",
  "context": {
    "artist": "Simpsons",
    "content_type": "concert",
    "keywords": ["dj", "set", "live"]
  },
  "actions_executed": ["youtube_search"],
  "results": {
    "youtube_search": {
      "query": "Simpsons DJ set live",
      "results": [...]
    }
  }
}
```

## 🚀 Adicionar Novas Ações

Para adicionar novas ações:

1. **Criar módulo de ação** (ex: `spotify_search.py`)
2. **Adicionar detecção** em `analyzer.py`:
   ```python
   def detect_actions(text: str) -> List[str]:
       # Adicionar padrões para nova ação
       if re.search(r'search.*spotify', text.lower()):
           actions.append("spotify_search")
   ```
3. **Executar ação** em `process_screenshot()`:
   ```python
   if "spotify_search" in actions:
       results["spotify_search"] = await search_spotify(context)
   ```

## 📚 Próximas Ações Planejadas

- [ ] Spotify search
- [ ] SoundCloud search
- [ ] Google search
- [ ] Extrair informações de contato
- [ ] Adicionar a calendário
- [ ] Compartilhar em redes sociais
