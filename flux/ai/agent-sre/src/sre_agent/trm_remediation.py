"""
🎯 TRM-Based Remediation Selection for Agent-SRE

Integrates TRM model for recursive reasoning about Lambda function selection.
TRM does NOT support tool calling, so we parse its structured output.
"""

import os
from typing import Dict, Any, Optional
import json
import structlog
import httpx
import time
from cloudevents.http import CloudEvent, to_binary

try:
    from .observability import (
        record_trm_inference,
        record_trm_confidence,
        record_trm_fallback,
        set_trm_model_loaded,
    )
except ImportError:
    # Metrics not available (fallback for testing)
    def record_trm_inference(method: str, status: str, duration: Optional[float] = None):
        pass
    def record_trm_confidence(alertname: str, confidence: float):
        pass
    def record_trm_fallback(reason: str):
        pass
    def set_trm_model_loaded(loaded: bool):
        pass

logger = structlog.get_logger()


class TRMRemediationSelector:
    """
    Uses TRM model for recursive reasoning to select Lambda functions.
    
    Since TRM doesn't support tool calling, we:
    1. Use TRM to reason about the alert
    2. Parse structured output (JSON) from TRM
    3. Send CloudEvent to trigger selected Lambda function
    """
    
    def __init__(
        self,
        trm_model_path: Optional[str] = None,
        trm_api_url: Optional[str] = None,
        broker_url: Optional[str] = None
    ):
        """
        Args:
            trm_model_path: Path to TRM model (for local inference)
            trm_api_url: URL to TRM inference service (for remote inference)
            broker_url: RabbitMQ broker URL for sending CloudEvents
        """
        self.trm_model_path = trm_model_path
        self.trm_api_url = trm_api_url or os.getenv("TRM_API_URL", "http://trm-reasoning.ml-platform.svc:8080")
        self.broker_url = broker_url or os.getenv("BROKER_URL", "http://lambda-broker.knative-lambda.svc.cluster.local")
        self._model = None
    
    async def select_and_trigger(
        self,
        alert_data: Dict[str, Any],
        correlation_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Use TRM to select Lambda function and trigger it via CloudEvent.
        
        Args:
            alert_data: Prometheus alert data
            correlation_id: Correlation ID for tracing
            
        Returns:
            Dict with lambda_function, parameters, event_id, status
        """
        logger.info(
            "trm_remediation_start",
            alertname=alert_data.get("labels", {}).get("alertname"),
            correlation_id=correlation_id
        )
        
        # Step 1: Use TRM to reason about remediation
        trm_result = await self._reason_with_trm(alert_data)
        
        if not trm_result.get("lambda_function"):
            logger.warning(
                "trm_no_function_selected",
                alertname=alert_data.get("labels", {}).get("alertname"),
                correlation_id=correlation_id
            )
            return {
                "lambda_function": None,
                "status": "no_function_selected",
                "reasoning": trm_result.get("reasoning", "")
            }
        
        # Step 2: Send CloudEvent to trigger Lambda function
        event_result = await self._send_lambda_cloudevent(
            lambda_function=trm_result["lambda_function"],
            parameters=trm_result["parameters"],
            alert_data=alert_data,
            correlation_id=correlation_id
        )
        
        return {
            "lambda_function": trm_result["lambda_function"],
            "parameters": trm_result["parameters"],
            "reasoning": trm_result.get("reasoning", ""),
            "confidence": trm_result.get("confidence", 0.0),
            "event_id": event_result.get("event_id"),
            "status": "triggered" if event_result.get("success") else "failed"
        }
    
    async def _reason_with_trm(self, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Use TRM model for recursive reasoning."""
        if self.trm_api_url:
            # Remote inference via API
            return await self._reason_via_api(alert_data)
        else:
            # Local inference (if model loaded)
            return await self._reason_local(alert_data)
    
    async def _reason_via_api(self, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Call TRM inference API."""
        try:
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    f"{self.trm_api_url}/reason",
                    json={
                        "problem": self._create_problem_prompt(alert_data),
                        "max_iterations": 10
                    }
                )
                response.raise_for_status()
                result = response.json()
                
                # Parse TRM output
                return self._parse_trm_output(result.get("result", ""))
        except Exception as e:
            logger.error("trm_api_error", error=str(e))
            return {"lambda_function": None, "reasoning": f"TRM API error: {e}"}
    
    async def _reason_local(self, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Local TRM inference (if model available)."""
        import time
        
        if not self.trm_model_path:
            logger.warning("trm_model_path_not_set", fallback="rule_based")
            record_trm_fallback(reason="model_path_not_set")
            return self._rule_based_selection(alert_data)
        
        start_time = time.time()
        try:
            # Import TRM selector from trm-finetune
            import sys
            from pathlib import Path
            
            # Add trm-finetune to path
            trm_finetune_path = Path(self.trm_model_path).parent.parent.parent / "trm-finetune" / "src"
            if trm_finetune_path.exists():
                sys.path.insert(0, str(trm_finetune_path))
            
            from trm_remediation_selector import TRMRemediationSelector
            
            # Initialize selector with trained model
            selector = TRMRemediationSelector(
                model_path=self.trm_model_path,
                trm_repo_path=os.getenv("TRM_REPO_PATH", "../trm")
            )
            
            # Check if model loaded
            if selector.model is None:
                set_trm_model_loaded(loaded=False)
                record_trm_fallback(reason="model_not_loaded")
                logger.warning("trm_model_not_loaded", fallback="rule_based")
                return self._rule_based_selection(alert_data)
            
            set_trm_model_loaded(loaded=True)
            
            # Run selection (synchronous, but we're in async context)
            result = selector.select_remediation(alert_data, max_iterations=10)
            
            duration = time.time() - start_time
            method = result.get("method", "trm_local")
            confidence = result.get("confidence", 0.0)
            alertname = alert_data.get("labels", {}).get("alertname", "unknown")
            
            # Record metrics using OpenTelemetry
            record_trm_inference(method=method, status="success", duration=duration)
            record_trm_confidence(alertname=alertname, confidence=confidence)
            
            logger.info(
                "trm_inference_complete",
                alertname=alertname,
                lambda_function=result.get("lambda_function"),
                confidence=confidence,
                method=method,
                duration_ms=duration * 1000
            )
            
            return {
                "lambda_function": result.get("lambda_function"),
                "parameters": result.get("parameters", {}),
                "reasoning": result.get("reasoning", ""),
                "confidence": confidence,
                "method": method
            }
        except Exception as e:
            duration = time.time() - start_time
            record_trm_inference(method="trm_local", status="error", duration=duration)
            record_trm_fallback(reason=f"error: {str(e)[:50]}")
            logger.error("trm_local_inference_error", error=str(e), exc_info=True)
            return self._rule_based_selection(alert_data)
    
    def _create_problem_prompt(self, alert_data: Dict[str, Any]) -> str:
        """Create problem prompt for TRM."""
        labels = alert_data.get("labels", {})
        annotations = alert_data.get("annotations", {})
        alertname = labels.get("alertname", "unknown")
        
        prompt = f"""Analyze this Prometheus alert and select the appropriate remediation Lambda function.

Alert Name: {alertname}
Labels: {json.dumps(labels, indent=2)}
Annotations: {json.dumps(annotations, indent=2)}

Available Lambda Functions:
- flux-reconcile-kustomization: Reconcile Flux Kustomization
- flux-reconcile-gitrepository: Reconcile Flux GitRepository
- flux-reconcile-helmrelease: Reconcile Flux HelmRelease
- pod-restart: Restart a pod or deployment
- pod-check-status: Check pod status
- scale-deployment: Scale deployment
- check-pvc-status: Check PVC status

Output JSON: {{"lambda_function": "...", "parameters": {{...}}, "reasoning": "..."}}
"""
        return prompt
    
    def _parse_trm_output(self, output_text: str) -> Dict[str, Any]:
        """Parse TRM output to extract structured decision."""
        import re
        
        # Try to extract JSON
        json_match = re.search(r'\{[^{}]*"lambda_function"[^{}]*\}', output_text, re.DOTALL)
        if json_match:
            try:
                return json.loads(json_match.group(0))
            except:
                pass
        
        # Fallback: rule-based
        return self._rule_based_selection({"labels": {"alertname": "unknown"}})
    
    def _rule_based_selection(self, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Fallback rule-based selection."""
        labels = alert_data.get("labels", {})
        alertname = labels.get("alertname", "")
        
        if "FluxReconciliationFailure" in alertname:
            return {
                "lambda_function": "flux-reconcile-kustomization",
                "parameters": {
                    "name": labels.get("name", ""),
                    "namespace": labels.get("namespace", "flux-system")
                },
                "reasoning": "Rule-based: Flux Kustomization failure",
                "confidence": 0.6
            }
        
        return {"lambda_function": None, "reasoning": "No match found"}
    
    async def _send_lambda_cloudevent(
        self,
        lambda_function: str,
        parameters: Dict[str, Any],
        alert_data: Dict[str, Any],
        correlation_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Send CloudEvent to trigger Lambda function.
        
        Event Type: io.homelab.agent-sre.lambda.trigger
        """
        import uuid
        
        event_id = str(uuid.uuid4())
        
        # Create CloudEvent
        event = CloudEvent(
            {
                "type": "io.homelab.agent-sre.lambda.trigger",
                "source": "/agent-sre/remediation",
                "id": event_id,
                "correlationid": correlation_id or str(uuid.uuid4())
            },
            {
                "lambda_function": lambda_function,
                "parameters": parameters,
                "alert": alert_data,
                "triggered_by": "trm-reasoning"
            }
        )
        
        try:
            # Send to broker
            headers, body = to_binary(event)
            
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.post(
                    self.broker_url,
                    headers=dict(headers),
                    content=body
                )
                response.raise_for_status()
                
                logger.info(
                    "lambda_cloudevent_sent",
                    lambda_function=lambda_function,
                    event_id=event_id,
                    correlation_id=correlation_id
                )
                
                return {
                    "success": True,
                    "event_id": event_id
                }
        except Exception as e:
            logger.error(
                "lambda_cloudevent_failed",
                lambda_function=lambda_function,
                error=str(e),
                correlation_id=correlation_id
            )
            return {
                "success": False,
                "error": str(e)
            }


async def select_remediation_with_trm(
    alert_data: Dict[str, Any],
    trm_api_url: Optional[str] = None,
    broker_url: Optional[str] = None,
    correlation_id: Optional[str] = None
) -> Dict[str, Any]:
    """
    Main entry point for TRM-based remediation selection.
    
    This function:
    1. Uses TRM to reason about the alert
    2. Selects appropriate Lambda function
    3. Sends CloudEvent to trigger Lambda function
    
    Returns:
        Dict with lambda_function, parameters, event_id, status
    """
    selector = TRMRemediationSelector(
        trm_api_url=trm_api_url,
        broker_url=broker_url
    )
    
    return await selector.select_and_trigger(alert_data, correlation_id=correlation_id)

