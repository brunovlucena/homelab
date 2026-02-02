"""
SoundCloud Search Integration

Busca músicas no SoundCloud baseado em informações extraídas do screenshot.
"""

import os
import logging
from typing import Optional, List, Dict
from urllib.parse import quote_plus
import httpx

logger = logging.getLogger(__name__)

# SoundCloud API (uses web scraping or unofficial API)
# Note: SoundCloud doesn't have a public search API, so we use web search URL
SOUNDCLOUD_CLIENT_ID = os.getenv("SOUNDCLOUD_CLIENT_ID", "")  # Optional, for future use


async def search_soundcloud(query: str, max_results: int = 10) -> List[Dict]:
    """
    Busca no SoundCloud.
    
    Nota: SoundCloud não tem API pública de busca oficial.
    Usamos URL de busca web como fallback.
    
    Args:
        query: Query de busca
        max_results: Número máximo de resultados (não usado no fallback)
    
    Returns:
        Lista de resultados
    """
    # SoundCloud doesn't have a public search API
    # We use the web search URL as fallback
    logger.info(f"SoundCloud search (web URL) for '{query}'")
    
    search_url = f"https://soundcloud.com/search?q={quote_plus(query)}"
    
    return [{
        "title": f"Search results for: {query}",
        "url": search_url,
        "type": "search",
        "note": "SoundCloud doesn't have a public search API. Using web search URL. For better results, consider using SoundCloud's embed API or web scraping.",
    }]


async def build_soundcloud_query_from_context(context: Dict) -> str:
    """
    Constrói query de busca do SoundCloud baseado no contexto.
    
    Args:
        context: Contexto extraído (artista, evento, etc.)
    
    Returns:
        Query otimizada para busca no SoundCloud
    """
    query_parts = []
    
    # Use LLM suggested query if available
    if context.get("llm_analysis", {}).get("suggested_queries", {}).get("soundcloud"):
        return context["llm_analysis"]["suggested_queries"]["soundcloud"]
    
    # Build query from context
    if context.get("artist"):
        query_parts.append(context["artist"])
    
    if context.get("event_name"):
        query_parts.append(context["event_name"])
    
    # Add content type modifiers
    content_type = context.get("content_type", "").lower()
    if "dj" in content_type or "set" in content_type:
        query_parts.append("DJ set")
    
    query = " ".join(query_parts) if query_parts else context.get("title", "music")
    
    logger.info(f"Built SoundCloud query: '{query}' from context")
    return query
