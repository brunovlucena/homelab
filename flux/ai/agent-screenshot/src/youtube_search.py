"""
YouTube Search Integration

Busca vídeos no YouTube baseado em informações extraídas do screenshot.
"""

import os
import logging
from typing import Optional, List, Dict
from urllib.parse import quote_plus
import httpx

logger = logging.getLogger(__name__)

# YouTube API (Data API v3)
YOUTUBE_API_KEY = os.getenv("YOUTUBE_API_KEY", "")
YOUTUBE_API_URL = "https://www.googleapis.com/youtube/v3/search"


async def search_youtube(query: str, max_results: int = 5) -> List[Dict]:
    """
    Busca vídeos no YouTube.
    
    Args:
        query: Query de busca (ex: "Simpsons DJ concert", "artist name live")
        max_results: Número máximo de resultados
    
    Returns:
        Lista de vídeos encontrados
    """
    if not YOUTUBE_API_KEY:
        logger.warning("YOUTUBE_API_KEY não configurada, usando fallback")
        return _fallback_youtube_search(query, max_results)
    
    try:
        async with httpx.AsyncClient() as client:
            params = {
                "part": "snippet",
                "q": query,
                "type": "video",
                "maxResults": max_results,
                "key": YOUTUBE_API_KEY,
                "order": "relevance",
            }
            
            response = await client.get(YOUTUBE_API_URL, params=params, timeout=10.0)
            response.raise_for_status()
            
            data = response.json()
            videos = []
            
            for item in data.get("items", []):
                video = {
                    "video_id": item["id"]["videoId"],
                    "title": item["snippet"]["title"],
                    "description": item["snippet"]["description"],
                    "channel": item["snippet"]["channelTitle"],
                    "published_at": item["snippet"]["publishedAt"],
                    "thumbnail": item["snippet"]["thumbnails"]["medium"]["url"],
                    "url": f"https://www.youtube.com/watch?v={item['id']['videoId']}",
                }
                videos.append(video)
            
            logger.info(f"YouTube search completed: {len(videos)} results for '{query}'")
            return videos
            
    except Exception as e:
        logger.error(f"Error searching YouTube: {e}")
        return _fallback_youtube_search(query, max_results)


def _fallback_youtube_search(query: str, max_results: int) -> List[Dict]:
    """
    Fallback quando YouTube API não está disponível.
    Retorna resultado simulado com URL de busca do YouTube.
    """
    logger.info(f"Using fallback YouTube search for '{query}'")
    
    # URL de busca do YouTube (funciona sem API)
    search_url = f"https://www.youtube.com/results?search_query={quote_plus(query)}"
    
    return [{
        "title": f"Search results for: {query}",
        "url": search_url,
        "description": "Click to see YouTube search results",
        "note": "YouTube API key not configured. Using search URL fallback.",
    }]


async def build_search_query_from_context(context: Dict) -> str:
    """
    Constrói query de busca do YouTube baseado no contexto extraído do screenshot.
    
    Args:
        context: Dicionário com informações extraídas (artista, evento, descrição, etc.)
    
    Returns:
        Query otimizada para busca no YouTube
    """
    query_parts = []
    
    # Adicionar artista/nome se disponível
    if context.get("artist"):
        query_parts.append(context["artist"])
    
    if context.get("event_name"):
        query_parts.append(context["event_name"])
    
    # Adicionar tipo de conteúdo
    content_type = context.get("content_type", "").lower()
    if "concert" in content_type or "live" in content_type:
        query_parts.append("live")
    elif "dj" in content_type:
        query_parts.append("DJ set")
    
    # Se não tem muito contexto, usar descrição relevante
    if not query_parts and context.get("description"):
        # Extrair palavras-chave da descrição
        desc = context["description"][:100]  # Primeiros 100 chars
        query_parts.append(desc)
    
    query = " ".join(query_parts) if query_parts else "music video"
    
    logger.info(f"Built YouTube query: '{query}' from context: {context}")
    return query
