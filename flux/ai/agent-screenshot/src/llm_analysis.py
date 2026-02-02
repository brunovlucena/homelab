"""
LLM Analysis - Context Understanding

Usa LLM para entender contexto e melhorar queries de busca.
"""

import os
import logging
from typing import Dict, Optional
import httpx
import json

logger = logging.getLogger(__name__)

# Configuration
OLLAMA_URL = os.getenv("OLLAMA_URL", "http://ollama-native.ollama.svc.cluster.local:11434")
OLLAMA_MODEL = os.getenv("OLLAMA_MODEL", "llama3.2:3b")


async def analyze_context_with_llm(context: Dict, ocr_text: str = "", vision_description: str = "") -> Dict:
    """
    Usa LLM para analisar contexto e extrair informações estruturadas.
    
    Args:
        context: Contexto básico extraído (URL, título, etc.)
        ocr_text: Texto extraído via OCR
        vision_description: Descrição da imagem via Vision Model
    
    Returns:
        Contexto enriquecido com análise do LLM
    """
    # Build prompt
    prompt = f"""Analyze this screenshot context and extract structured information for content search.

Context:
- URL: {context.get('url', 'N/A')}
- Title: {context.get('title', 'N/A')}
- Platform: {context.get('platform', 'N/A')}

OCR Text (extracted from image):
{ocr_text[:500] if ocr_text else 'No text extracted'}

Vision Analysis:
{vision_description[:500] if vision_description else 'No vision analysis'}

Extract and return a JSON object with:
{{
  "artist": "artist name if found",
  "event_name": "event/concert name if found",
  "content_type": "concert|dj_set|album|video|other",
  "keywords": ["relevant", "keywords"],
  "search_intent": "what user wants to find (youtube, spotify, soundcloud, etc.)",
  "suggested_queries": {{
    "youtube": "optimized query for YouTube search",
    "spotify": "optimized query for Spotify search",
    "soundcloud": "optimized query for SoundCloud search"
  }},
  "actions": ["youtube_search", "spotify_search", etc.],
  "summary": "brief summary of the content"
}}

Return ONLY valid JSON, no additional text.
"""
    
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(
                f"{OLLAMA_URL}/api/generate",
                json={
                    "model": OLLAMA_MODEL,
                    "prompt": prompt,
                    "stream": False,
                    "options": {
                        "temperature": 0.3,  # Lower temperature for structured output
                        "num_predict": 1000,
                    }
                }
            )
            response.raise_for_status()
            result = response.json()
            
            response_text = result.get("response", "").strip()
            
            # Try to extract JSON from response
            try:
                # Remove markdown code blocks if present
                if "```json" in response_text:
                    response_text = response_text.split("```json")[1].split("```")[0].strip()
                elif "```" in response_text:
                    response_text = response_text.split("```")[1].split("```")[0].strip()
                
                llm_data = json.loads(response_text)
                
                # Merge with original context
                enriched_context = context.copy()
                enriched_context.update({
                    "llm_analysis": llm_data,
                    "artist": llm_data.get("artist") or context.get("artist"),
                    "event_name": llm_data.get("event_name") or context.get("event_name"),
                    "content_type": llm_data.get("content_type") or context.get("content_type"),
                    "keywords": llm_data.get("keywords", []) + context.get("keywords", []),
                })
                
                logger.info(f"LLM analysis completed: {llm_data.get('summary', '')[:50]}")
                
                return enriched_context
                
            except json.JSONDecodeError as e:
                logger.warning(f"Failed to parse LLM JSON response: {e}")
                logger.debug(f"Response text: {response_text[:200]}")
                # Return original context if parsing fails
                return context
            
    except Exception as e:
        logger.error(f"LLM analysis error: {e}")
        # Return original context on error
        return context
