"""
Spotify Search Integration

Busca músicas/artistas no Spotify baseado em informações extraídas do screenshot.
"""

import os
import logging
from typing import Optional, List, Dict
from urllib.parse import quote_plus
import httpx
import base64

logger = logging.getLogger(__name__)

# Spotify API Configuration
SPOTIFY_CLIENT_ID = os.getenv("SPOTIFY_CLIENT_ID", "")
SPOTIFY_CLIENT_SECRET = os.getenv("SPOTIFY_CLIENT_SECRET", "")
SPOTIFY_API_URL = "https://api.spotify.com/v1"


async def get_spotify_token() -> Optional[str]:
    """Obtém token de acesso do Spotify."""
    if not SPOTIFY_CLIENT_ID or not SPOTIFY_CLIENT_SECRET:
        return None
    
    try:
        credentials = base64.b64encode(
            f"{SPOTIFY_CLIENT_ID}:{SPOTIFY_CLIENT_SECRET}".encode()
        ).decode()
        
        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.post(
                "https://accounts.spotify.com/api/token",
                headers={
                    "Authorization": f"Basic {credentials}",
                    "Content-Type": "application/x-www-form-urlencoded",
                },
                data={"grant_type": "client_credentials"},
            )
            response.raise_for_status()
            data = response.json()
            return data.get("access_token")
            
    except Exception as e:
        logger.error(f"Failed to get Spotify token: {e}")
        return None


async def search_spotify(query: str, max_results: int = 10, search_type: str = "track") -> List[Dict]:
    """
    Busca no Spotify.
    
    Args:
        query: Query de busca
        max_results: Número máximo de resultados
        search_type: Tipo de busca (track, album, artist, playlist)
    
    Returns:
        Lista de resultados
    """
    token = await get_spotify_token()
    if not token:
        logger.warning("Spotify API credentials not configured, using fallback")
        return _fallback_spotify_search(query)
    
    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.get(
                f"{SPOTIFY_API_URL}/search",
                headers={"Authorization": f"Bearer {token}"},
                params={
                    "q": query,
                    "type": search_type,
                    "limit": min(max_results, 50),
                }
            )
            response.raise_for_status()
            data = response.json()
            
            results = []
            items = data.get(f"{search_type}s", {}).get("items", [])
            
            for item in items:
                result = {
                    "name": item["name"],
                    "type": search_type,
                    "spotify_url": item["external_urls"]["spotify"],
                    "spotify_id": item["id"],
                }
                
                if search_type == "track":
                    result["artists"] = [a["name"] for a in item.get("artists", [])]
                    result["album"] = item.get("album", {}).get("name", "")
                    result["duration_ms"] = item.get("duration_ms", 0)
                    result["preview_url"] = item.get("preview_url")
                
                elif search_type == "artist":
                    result["genres"] = item.get("genres", [])
                    result["followers"] = item.get("followers", {}).get("total", 0)
                
                elif search_type == "album":
                    result["artists"] = [a["name"] for a in item.get("artists", [])]
                    result["release_date"] = item.get("release_date", "")
                
                # Add image if available
                images = item.get("images", [])
                if images:
                    result["image"] = images[0]["url"]
                
                results.append(result)
            
            logger.info(f"Spotify search completed: {len(results)} results for '{query}'")
            return results
            
    except Exception as e:
        logger.error(f"Error searching Spotify: {e}")
        return _fallback_spotify_search(query)


def _fallback_spotify_search(query: str) -> List[Dict]:
    """Fallback quando Spotify API não está disponível."""
    logger.info(f"Using fallback Spotify search for '{query}'")
    
    search_url = f"https://open.spotify.com/search/{quote_plus(query)}"
    
    return [{
        "name": f"Search results for: {query}",
        "spotify_url": search_url,
        "type": "search",
        "note": "Spotify API credentials not configured. Using search URL fallback.",
    }]


async def build_spotify_query_from_context(context: Dict) -> str:
    """
    Constrói query de busca do Spotify baseado no contexto.
    
    Args:
        context: Contexto extraído (artista, evento, etc.)
    
    Returns:
        Query otimizada para busca no Spotify
    """
    query_parts = []
    
    # Use LLM suggested query if available
    if context.get("llm_analysis", {}).get("suggested_queries", {}).get("spotify"):
        return context["llm_analysis"]["suggested_queries"]["spotify"]
    
    # Build query from context
    if context.get("artist"):
        query_parts.append(context["artist"])
    
    if context.get("event_name"):
        query_parts.append(context["event_name"])
    
    # Add content type modifiers
    content_type = context.get("content_type", "").lower()
    if "album" in content_type:
        query_parts.append("album")
    
    query = " ".join(query_parts) if query_parts else context.get("title", "music")
    
    logger.info(f"Built Spotify query: '{query}' from context")
    return query
