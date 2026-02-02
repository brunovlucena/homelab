"""
Vision Model - Image Analysis

Analisa screenshots usando modelos de visão (GPT-4V, Claude Vision, ou LLaVA local).
"""

import os
import logging
from typing import Optional, Dict, List
import base64
import httpx

logger = logging.getLogger(__name__)

# Configuration
OLLAMA_URL = os.getenv("OLLAMA_URL", "http://ollama-native.ollama.svc.cluster.local:11434")
VISION_MODEL = os.getenv("VISION_MODEL", "llava:7b")  # Local vision model
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY", "")


async def analyze_image_with_vision(image_bytes: bytes, prompt: str = None) -> Dict[str, any]:
    """
    Analisa imagem usando modelo de visão.
    
    Tenta na ordem:
    1. LLaVA (Ollama local) - padrão
    2. GPT-4V (OpenAI) - se API key disponível
    3. Claude Vision (Anthropic) - se API key disponível
    
    Args:
        image_bytes: Bytes da imagem
        prompt: Prompt opcional para análise específica
    
    Returns:
        Dicionário com análise da imagem
    """
    if not image_bytes:
        return {"description": "", "method": "none"}
    
    # Default prompt
    if prompt is None:
        prompt = """Analyze this screenshot/image and describe:
- What type of content is shown (social media post, video, article, etc.)
- Key text visible in the image
- Any artists, events, or music-related content
- Context and relevant information for searching content (YouTube, Spotify, etc.)
- Any explicit requests or actions mentioned (like "find this on YouTube")
"""
    
    # Try LLaVA (local) first
    if OLLAMA_URL:
        try:
            result = await _analyze_with_llava(image_bytes, prompt)
            if result.get("description"):
                return result
        except Exception as e:
            logger.warning(f"LLaVA analysis failed: {e}")
    
    # Try GPT-4V
    if OPENAI_API_KEY:
        try:
            result = await _analyze_with_gpt4v(image_bytes, prompt)
            if result.get("description"):
                return result
        except Exception as e:
            logger.warning(f"GPT-4V analysis failed: {e}")
    
    # Try Claude Vision
    if ANTHROPIC_API_KEY:
        try:
            result = await _analyze_with_claude_vision(image_bytes, prompt)
            if result.get("description"):
                return result
        except Exception as e:
            logger.warning(f"Claude Vision analysis failed: {e}")
    
    # No vision model available
    logger.warning("No vision model available for analysis")
    return {"description": "", "method": "none"}


async def _analyze_with_llava(image_bytes: bytes, prompt: str) -> Dict[str, any]:
    """Analisa imagem usando LLaVA via Ollama."""
    try:
        import base64
        
        # Convert image to base64
        image_b64 = base64.b64encode(image_bytes).decode('utf-8')
        
        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.post(
                f"{OLLAMA_URL}/api/generate",
                json={
                    "model": VISION_MODEL,
                    "prompt": prompt,
                    "images": [image_b64],
                    "stream": False,
                }
            )
            response.raise_for_status()
            result = response.json()
            
            description = result.get("response", "").strip()
            
            logger.info(f"LLaVA analysis completed, length: {len(description)}")
            
            return {
                "description": description,
                "method": "llava",
                "model": VISION_MODEL,
            }
            
    except Exception as e:
        logger.error(f"LLaVA analysis error: {e}")
        raise


async def _analyze_with_gpt4v(image_bytes: bytes, prompt: str) -> Dict[str, any]:
    """Analisa imagem usando GPT-4V (OpenAI)."""
    try:
        import base64
        
        image_b64 = base64.b64encode(image_bytes).decode('utf-8')
        
        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.post(
                "https://api.openai.com/v1/chat/completions",
                headers={
                    "Authorization": f"Bearer {OPENAI_API_KEY}",
                    "Content-Type": "application/json",
                },
                json={
                    "model": "gpt-4-vision-preview",
                    "messages": [
                        {
                            "role": "user",
                            "content": [
                                {"type": "text", "text": prompt},
                                {
                                    "type": "image_url",
                                    "image_url": {
                                        "url": f"data:image/jpeg;base64,{image_b64}"
                                    }
                                }
                            ]
                        }
                    ],
                    "max_tokens": 1000,
                }
            )
            response.raise_for_status()
            result = response.json()
            
            description = result["choices"][0]["message"]["content"]
            
            logger.info(f"GPT-4V analysis completed")
            
            return {
                "description": description,
                "method": "gpt4v",
            }
            
    except Exception as e:
        logger.error(f"GPT-4V analysis error: {e}")
        raise


async def _analyze_with_claude_vision(image_bytes: bytes, prompt: str) -> Dict[str, any]:
    """Analisa imagem usando Claude Vision (Anthropic)."""
    try:
        import base64
        
        image_b64 = base64.b64encode(image_bytes).decode('utf-8')
        
        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.post(
                "https://api.anthropic.com/v1/messages",
                headers={
                    "x-api-key": ANTHROPIC_API_KEY,
                    "anthropic-version": "2023-06-01",
                    "Content-Type": "application/json",
                },
                json={
                    "model": "claude-3-5-sonnet-20241022",
                    "max_tokens": 1024,
                    "messages": [
                        {
                            "role": "user",
                            "content": [
                                {
                                    "type": "image",
                                    "source": {
                                        "type": "base64",
                                        "media_type": "image/jpeg",
                                        "data": image_b64
                                    }
                                },
                                {
                                    "type": "text",
                                    "text": prompt
                                }
                            ]
                        }
                    ]
                }
            )
            response.raise_for_status()
            result = response.json()
            
            description = result["content"][0]["text"]
            
            logger.info(f"Claude Vision analysis completed")
            
            return {
                "description": description,
                "method": "claude_vision",
            }
            
    except Exception as e:
        logger.error(f"Claude Vision analysis error: {e}")
        raise
