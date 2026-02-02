"""
Client for Agent-Reasoning service.

Allows agents to call the TRM reasoning service via HTTP or CloudEvents.
"""
import os
from typing import Optional, Dict, Any
import httpx
import structlog
from cloudevents.http import CloudEvent, to_structured

from .types import ReasoningRequest, ReasoningResponse, TaskType

logger = structlog.get_logger()


class ReasoningClient:
    """
    Client for calling Agent-Reasoning service.
    
    Supports both HTTP and CloudEvents communication patterns.
    """
    
    def __init__(
        self,
        base_url: Optional[str] = None,
        use_events: bool = False,
        broker_url: Optional[str] = None,
    ):
        self.base_url = base_url or os.getenv(
            "REASONING_SERVICE_URL",
            "http://agent-reasoning.ai-agents.svc.cluster.local:8080"
        )
        self.use_events = use_events or os.getenv("REASONING_USE_EVENTS", "false").lower() == "true"
        self.broker_url = broker_url or os.getenv("KNATIVE_BROKER_URL")
        
        logger.info(
            "reasoning_client_initialized",
            base_url=self.base_url,
            use_events=self.use_events,
        )
    
    async def reason(
        self,
        question: str,
        context: Optional[Dict[str, Any]] = None,
        max_steps: int = 6,
        task_type: TaskType = TaskType.GENERAL,
        conversation_id: Optional[str] = None,
    ) -> ReasoningResponse:
        """
        Perform recursive reasoning on a question.
        
        Args:
            question: The question to reason about
            context: Optional context dictionary
            max_steps: Maximum reasoning steps (1-20)
            task_type: Type of reasoning task
            conversation_id: Optional conversation ID
            
        Returns:
            ReasoningResponse with answer and reasoning trace
        """
        request = ReasoningRequest(
            question=question,
            context=context or {},
            max_steps=max_steps,
            task_type=task_type,
            conversation_id=conversation_id,
        )
        
        if self.use_events and self.broker_url:
            return await self._reason_via_events(request)
        else:
            return await self._reason_via_http(request)
    
    async def _reason_via_http(self, request: ReasoningRequest) -> ReasoningResponse:
        """Call reasoning service via HTTP."""
        try:
            async with httpx.AsyncClient(timeout=120.0) as client:
                response = await client.post(
                    f"{self.base_url}/reason",
                    json=request.dict(),
                )
                response.raise_for_status()
                
                result = response.json()
                return ReasoningResponse(**result)
                
        except httpx.TimeoutException:
            logger.error("reasoning_timeout", question=request.question[:100])
            raise
        except httpx.HTTPStatusError as e:
            logger.error(
                "reasoning_http_error",
                status=e.response.status_code,
                question=request.question[:100],
            )
            raise
        except Exception as e:
            logger.error("reasoning_request_failed", error=str(e))
            raise
    
    async def _reason_via_events(self, request: ReasoningRequest) -> ReasoningResponse:
        """Call reasoning service via CloudEvents."""
        # Create CloudEvent
        event = CloudEvent(
            type="io.homelab.reasoning.requested",
            source="agent-client",
            data=request.dict(),
        )
        
        # Send to broker
        try:
            async with httpx.AsyncClient(timeout=120.0) as client:
                headers, body = to_structured(event)
                response = await client.post(
                    self.broker_url,
                    headers=dict(headers),
                    content=body,
                )
                response.raise_for_status()
                
                # In a real implementation, you'd wait for the response event
                # For now, we'll fall back to HTTP
                logger.warning("event_based_reasoning_not_fully_implemented")
                return await self._reason_via_http(request)
                
        except Exception as e:
            logger.error("reasoning_event_failed", error=str(e))
            # Fallback to HTTP
            return await self._reason_via_http(request)
    
    async def health_check(self) -> bool:
        """Check if reasoning service is available."""
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.get(f"{self.base_url}/health")
                return response.status_code == 200
        except Exception:
            return False

