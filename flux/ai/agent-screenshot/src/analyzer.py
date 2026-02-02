"""
Screenshot Analyzer

Analisa screenshots e extrai informações relevantes para ações (como buscar no YouTube).
"""

import logging
import re
from typing import Dict, List, Optional

logger = logging.getLogger(__name__)


def extract_text_from_metadata(metadata: Dict, url: str, title: str) -> Dict:
    """
    Extrai informações relevantes de metadados, URL e título.
    
    Por enquanto, faz análise básica. Futuramente pode usar:
    - OCR para extrair texto da imagem
    - Vision model para entender contexto visual
    - LLM para análise semântica
    """
    context = {
        "url": url,
        "title": title,
        "text_extracted": [],
        "artist": None,
        "event_name": None,
        "content_type": None,
        "keywords": [],
    }
    
    # Analisar URL
    if "instagram.com" in url.lower():
        context["platform"] = "instagram"
        # Instagram posts geralmente têm texto na descrição/comentários
        # (que viria via OCR ou metadados)
    
    if "youtube.com" in url.lower():
        context["platform"] = "youtube"
    
    # Analisar título da página
    if title:
        context["text_extracted"].append(title)
        
        # Tentar extrair informações comuns
        # Ex: "Artist Name - Concert Name" ou "Event Name"
        title_lower = title.lower()
        
        # Detectar tipos de conteúdo
        if any(word in title_lower for word in ["concert", "live", "performance"]):
            context["content_type"] = "concert"
        elif "dj" in title_lower:
            context["content_type"] = "dj_set"
        elif "album" in title_lower:
            context["content_type"] = "album"
    
    # Analisar metadados se disponíveis
    if metadata:
        # Se há descrição nos metadados
        if "description" in metadata:
            desc = metadata["description"]
            context["text_extracted"].append(desc)
            
            # Extrair palavras-chave
            keywords = _extract_keywords(desc)
            context["keywords"].extend(keywords)
            
            # Tentar identificar artista/evento
            artist = _extract_artist(desc)
            if artist:
                context["artist"] = artist
            
            event = _extract_event(desc)
            if event:
                context["event_name"] = event
        
        # Se há comentários ou texto adicional
        if "comments" in metadata:
            comments_text = " ".join(metadata.get("comments", []))
            context["text_extracted"].append(comments_text)
    
    logger.info(f"Extracted context: {context}")
    return context


def _extract_keywords(text: str) -> List[str]:
    """Extrai palavras-chave relevantes do texto."""
    # Palavras comuns a ignorar
    stop_words = {"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
    
    # Extrair palavras (simples, pode melhorar)
    words = re.findall(r'\b[a-zA-Z]{3,}\b', text.lower())
    keywords = [w for w in words if w not in stop_words][:10]  # Top 10
    
    return keywords


def _extract_artist(text: str) -> Optional[str]:
    """Tenta extrair nome do artista do texto."""
    # Padrões comuns
    patterns = [
        r'@(\w+)',  # @artistname (Instagram/Twitter)
        r'by\s+([A-Z][a-zA-Z\s]+)',  # by Artist Name
        r'artist[:\s]+([A-Z][a-zA-Z\s]+)',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, text, re.IGNORECASE)
        if match:
            return match.group(1).strip()
    
    return None


def _extract_event(text: str) -> Optional[str]:
    """Tenta extrair nome do evento/concerte do texto."""
    patterns = [
        r'concert[:\s]+([A-Z][a-zA-Z\s]+)',
        r'event[:\s]+([A-Z][a-zA-Z\s]+)',
        r'live at\s+([A-Z][a-zA-Z\s]+)',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, text, re.IGNORECASE)
        if match:
            return match.group(1).strip()
    
    return None


def detect_actions(text: str) -> List[str]:
    """
    Detecta ações solicitadas no texto.
    
    Exemplos:
    - "find in youtube this concert" → ["youtube_search"]
    - "find this on spotify" → ["spotify_search"]
    - "search soundcloud" → ["soundcloud_search"]
    """
    text_lower = text.lower()
    actions = []
    
    # Detectar busca no YouTube
    youtube_patterns = [
        r'find.*youtube',
        r'search.*youtube',
        r'youtube.*find',
        r'youtube.*search',
        r'find.*this.*concert',
        r'find.*this.*video',
    ]
    
    for pattern in youtube_patterns:
        if re.search(pattern, text_lower):
            actions.append("youtube_search")
            break
    
    # Detectar busca no Spotify
    spotify_patterns = [
        r'find.*spotify',
        r'search.*spotify',
        r'spotify.*find',
        r'find.*this.*song',
        r'find.*this.*track',
    ]
    
    for pattern in spotify_patterns:
        if re.search(pattern, text_lower):
            actions.append("spotify_search")
            break
    
    # Detectar busca no SoundCloud
    soundcloud_patterns = [
        r'find.*soundcloud',
        r'search.*soundcloud',
        r'soundcloud.*find',
    ]
    
    for pattern in soundcloud_patterns:
        if re.search(pattern, text_lower):
            actions.append("soundcloud_search")
            break
    
    return actions
