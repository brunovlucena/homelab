# 📸 Exemplo de Uso - Screenshot Agent

## 🎯 Cenário: Encontrar Concerto no YouTube

### Situação
Você vê um post no Instagram sobre um concerto/DJ set e quer encontrar o vídeo completo no YouTube.

### Passo a Passo

1. **Abrir o Instagram no browser**
   - Ver post sobre concerto/DJ set
   - Exemplo: Post sobre "Simpsons DJ Concert"

2. **Capturar Screenshot**
   - Clicar no ícone da extensão
   - Clicar em "📸 Capturar Screenshot"
   - Screenshot é enviado para o agente

3. **Agente Processa**
   ```
   Screenshot recebido → Agente analisa
   → Detecta: "concert" ou "DJ set"
   → Ação: youtube_search
   → Extrai contexto: artista, evento
   → Busca no YouTube
   → Retorna resultados
   ```

4. **Resultado**
   ```json
   {
     "screenshot_id": "scr_abc123",
     "context": {
       "artist": "Simpsons",
       "content_type": "concert",
       "url": "instagram.com/p/xyz"
     },
     "actions_executed": ["youtube_search"],
     "results": {
       "youtube_search": {
         "query": "Simpsons DJ set live",
         "results": [
           {
             "title": "Simpsons DJ Set - Live Performance",
             "url": "https://www.youtube.com/watch?v=xyz",
             "channel": "Channel Name",
             "thumbnail": "..."
           }
         ]
       }
     }
   }
   ```

## 🔍 Como o Agente Detecta Ação

### Detecção Automática

O agente detecta automaticamente quando buscar no YouTube se:

1. **Tipo de conteúdo detectado:**
   - Título contém: "concert", "live", "performance", "DJ"
   - URL é do Instagram (geralmente posts de eventos)
   - Descrição menciona evento/artista

2. **Comandos explícitos:**
   - Texto contém: "find in youtube", "search youtube"
   - Comentários como: "find this concert", "where is this"

### Exemplo de Contexto Extraído

**Input:**
- URL: `instagram.com/p/xyz`
- Título: "Simpsons DJ Concert - Amazing Set!"
- Descrição: "Check out this amazing DJ set with Simpsons characters..."

**Contexto Extraído:**
```json
{
  "platform": "instagram",
  "content_type": "concert",
  "artist": "Simpsons",
  "keywords": ["dj", "set", "amazing", "simpsons"],
  "text_extracted": ["Simpsons DJ Concert - Amazing Set!", "Check out this amazing DJ set..."]
}
```

**Query Construída:**
```
"Simpsons DJ set live"
```

## 🚀 Melhorias Futuras

### Com Vision Model
- Analisa imagem do screenshot diretamente
- Extrai texto com OCR
- Entende contexto visual (poster, flyer, etc.)

### Com LLM
- Analisa descrição/comentários com LLM
- Entende intenção do usuário
- Gera queries mais precisas

### Mais Ações
- Spotify search
- SoundCloud search
- Adicionar a playlist
- Compartilhar em redes sociais

## 💡 Dicas de Uso

1. **Screenshots de Instagram Posts:**
   - Melhor para detectar eventos/concertos
   - Descrição geralmente contém informações relevantes

2. **Screenshots de YouTube:**
   - Pode extrair informações do vídeo atual
   - Buscar vídeos relacionados

3. **Screenshots de Artigos/Notícias:**
   - Extrair nome do evento/artista
   - Buscar no YouTube
