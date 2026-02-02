#!/usr/bin/env python3
"""
agent-screenshot - Screenshot Analysis Agent

Recebe screenshots via CloudEvents e processa cada screenshot em uma instância isolada.
Cada screenshot = novo agent run (estilo Cursor Agents).

Capabilities:
- Analisa screenshots
- Extrai contexto (artista, evento, etc.)
- Busca no YouTube baseado no contexto
- Executa ações solicitadas (find in youtube, etc.)
"""

import os
import logging
from typing import Optional, Dict, List
from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import JSONResponse
from cloudevents.http import from_http
import structlog

from analyzer import extract_text_from_metadata, detect_actions
from youtube_search import search_youtube, build_search_query_from_context
from spotify_search import search_spotify, build_spotify_query_from_context
from soundcloud_search import search_soundcloud, build_soundcloud_query_from_context
from ocr import get_ocr_processor
from vision import analyze_image_with_vision
from llm_analysis import analyze_context_with_llm

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = structlog.get_logger()

app = FastAPI(
    title="Agent-Screenshot",
    description="Screenshot analysis agent - cada screenshot inicia um agente",
    version="0.1.0",
)

# =============================================================================
# Request/Response Models
# =============================================================================

class ScreenshotAnalysis:
    """Análise de screenshot"""
    screenshot_id: str
    url: str
    title: str
    status: str = "processing"
    analysis: Optional[dict] = None
    error: Optional[str] = None


# =============================================================================
# Helper Functions
# =============================================================================

async def process_screenshot(screenshot_id: str, url: str, title: str, metadata: dict, image_bytes: bytes = None) -> dict:
    """
    Processa um screenshot e executa ações solicitadas.
    
    Cada screenshot é processado de forma isolada (como um agent run separado).
    
    Pipeline:
    1. Extrair contexto básico (URL, título, metadados)
    2. OCR: extrair texto da imagem (se disponível)
    3. Vision Model: analisar imagem (se disponível)
    4. LLM: entender contexto e sugerir ações
    5. Executar ações (YouTube, Spotify, SoundCloud)
    
    Actions suportadas:
    - youtube_search: Busca no YouTube
    - spotify_search: Busca no Spotify
    - soundcloud_search: Busca no SoundCloud
    """
    logger.info(
        "screenshot_processing_started",
        screenshot_id=screenshot_id,
        url=url,
        title=title,
        has_image=image_bytes is not None,
    )
    
    try:
        # 1. Extrair contexto básico do screenshot
        context = extract_text_from_metadata(metadata, url, title)
        
        # 2. OCR: extrair texto da imagem (se disponível)
        ocr_result = {}
        ocr_text = ""
        if image_bytes:
            try:
                ocr_processor = get_ocr_processor()
                ocr_result = ocr_processor.extract_text_from_image(image_bytes)
                ocr_text = ocr_result.get("text", "")
                if ocr_text:
                    context["text_extracted"].append(ocr_text)
                    context["ocr"] = ocr_result
                    logger.info(f"OCR extracted {len(ocr_text)} characters")
            except Exception as e:
                logger.warning(f"OCR failed: {e}")
        
        # 3. Vision Model: analisar imagem (se disponível)
        vision_result = {}
        vision_description = ""
        if image_bytes:
            try:
                vision_result = await analyze_image_with_vision(image_bytes)
                vision_description = vision_result.get("description", "")
                if vision_description:
                    context["vision_analysis"] = vision_result
                    logger.info(f"Vision analysis completed: {vision_result.get('method', 'unknown')}")
            except Exception as e:
                logger.warning(f"Vision analysis failed: {e}")
        
        # 4. LLM: entender contexto e sugerir ações/queries
        try:
            context = await analyze_context_with_llm(context, ocr_text, vision_description)
            logger.info("LLM analysis completed")
        except Exception as e:
            logger.warning(f"LLM analysis failed: {e}")
        
        # 5. Detectar ações (do contexto, LLM, ou padrões)
        actions = []
        
        # Ações sugeridas pelo LLM
        if context.get("llm_analysis", {}).get("actions"):
            actions.extend(context["llm_analysis"]["actions"])
        
        # Detecção automática baseada em tipo de conteúdo
        if context.get("content_type") in ["concert", "dj_set", "album"]:
            actions.append("youtube_search")
            actions.append("spotify_search")
        
        # Detecção de ações explícitas no texto
        all_text = " ".join(context.get("text_extracted", []))
        detected_actions = detect_actions(all_text + " " + title)
        actions.extend(detected_actions)
        
        # Remove duplicatas
        actions = list(set(actions))
        
        # 6. Executar ações
        results = {}
        
        if "youtube_search" in actions:
            logger.info(f"Executando YouTube search para screenshot {screenshot_id}")
            query = await build_search_query_from_context(context)
            youtube_results = await search_youtube(query, max_results=5)
            results["youtube_search"] = {
                "query": query,
                "results": youtube_results,
                "count": len(youtube_results),
            }
        
        if "spotify_search" in actions:
            logger.info(f"Executando Spotify search para screenshot {screenshot_id}")
            query = await build_spotify_query_from_context(context)
            spotify_results = await search_spotify(query, max_results=5)
            results["spotify_search"] = {
                "query": query,
                "results": spotify_results,
                "count": len(spotify_results),
            }
        
        if "soundcloud_search" in actions:
            logger.info(f"Executando SoundCloud search para screenshot {screenshot_id}")
            query = await build_soundcloud_query_from_context(context)
            soundcloud_results = await search_soundcloud(query, max_results=5)
            results["soundcloud_search"] = {
                "query": query,
                "results": soundcloud_results,
                "count": len(soundcloud_results),
            }
        
        # 7. Construir resultado
        analysis_result = {
            "screenshot_id": screenshot_id,
            "url": url,
            "title": title,
            "status": "completed",
            "context": {
                "basic": {k: v for k, v in context.items() if k not in ["llm_analysis", "vision_analysis", "ocr"]},
                "ocr": ocr_result if ocr_result else None,
                "vision": vision_result if vision_result else None,
            },
            "actions_executed": actions,
            "results": results,
            "summary": context.get("llm_analysis", {}).get("summary") or f"Screenshot analisado: {title}",
        }
        
        logger.info(
            "screenshot_processing_completed",
            screenshot_id=screenshot_id,
            status="completed",
            actions=actions,
            results_count=len(results),
        )
        
        return analysis_result
        
    except Exception as e:
        logger.error(
            "screenshot_processing_failed",
            screenshot_id=screenshot_id,
            error=str(e),
        )
        raise


# =============================================================================
# Endpoints
# =============================================================================

@app.get("/health")
async def health():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "service": "agent-screenshot",
        "version": "0.1.0",
    }


@app.post("/")
async def handle_cloudevent(request: Request):
    """
    Handle CloudEvents from mobile-api.
    
    Event type esperado: screenshot.upload
    """
    try:
        # Parse CloudEvent
        headers = dict(request.headers)
        body = await request.body()
        event = from_http(headers, body)
        
        event_type = event.get("type")
        event_source = event.get("source")
        event_data = event.get("data", {})
        
        logger.info(
            "cloudevent_received",
            event_type=event_type,
            source=event_source,
            event_id=event.get("id"),
        )
        
        # Processar apenas eventos do tipo screenshot.upload
        if event_type != "screenshot.upload":
            logger.warning(
                "unexpected_event_type",
                event_type=event_type,
                expected="screenshot.upload",
            )
            return JSONResponse(
                status_code=200,
                content={
                    "status": "ignored",
                    "message": f"Event type {event_type} not handled",
                },
            )
        
        # Extrair dados do screenshot
        screenshot_id = event_data.get("screenshot_id")
        url = event_data.get("url", "")
        title = event_data.get("title", "")
        timestamp = event_data.get("timestamp", "")
        metadata = event_data.get("metadata", {})
        screenshot_url = event_data.get("screenshot_url")  # URL do screenshot no MinIO (opcional)
        
        if not screenshot_id:
            raise HTTPException(
                status_code=400,
                detail="screenshot_id missing in event data",
            )
        
        # Download imagem se URL fornecida (TODO: implementar download do MinIO)
        image_bytes = None
        if screenshot_url:
            # TODO: Download da imagem do MinIO/S3
            # Por enquanto, processa sem imagem
            logger.info(f"Screenshot URL provided but download not implemented: {screenshot_url}")
        
        # Processar screenshot (cada screenshot = novo agent run)
        # image_bytes pode ser None - OCR e Vision só rodam se disponível
        result = await process_screenshot(screenshot_id, url, title, metadata, image_bytes)
        
        # Retornar resultado
        return JSONResponse(
            status_code=200,
            content={
                "status": "success",
                "screenshot_id": screenshot_id,
                "result": result,
            },
        )
        
    except Exception as e:
        logger.error(
            "cloudevent_processing_error",
            error=str(e),
            error_type=type(e).__name__,
        )
        raise HTTPException(status_code=500, detail=f"Error processing event: {str(e)}")


@app.get("/screenshot/{screenshot_id}")
async def get_screenshot_status(screenshot_id: str):
    """
    Consulta status de um screenshot processado.
    
    TODO: Implementar storage (Redis/Postgres) para rastrear status.
    """
    return {
        "screenshot_id": screenshot_id,
        "status": "processing",
        "message": "Status tracking será implementado com storage",
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(app, host="0.0.0.0", port=port)
