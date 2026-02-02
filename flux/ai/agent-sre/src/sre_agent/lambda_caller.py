"""
LambdaFunction Caller - Calls remediation LambdaFunctions based on runbook instructions.
"""
import os
import uuid
from typing import Dict, Any, Optional
import httpx
import structlog
from cloudevents.http import CloudEvent, to_binary
from opentelemetry import trace

logger = structlog.get_logger()
tracer = trace.get_tracer(__name__)


class LambdaFunctionCaller:
    """Calls LambdaFunction remediation functions."""
    
    def __init__(self, namespace: str = "ai", timeout: int = 60):
        self.namespace = namespace
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)
    
    async def check_lambda_function_availability(
        self,
        function_name: str,
        namespace: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Check if LambdaFunction is available and ready.
        
        Returns:
            Dict with 'available' (bool) and 'reason' (str) if not available
        """
        ns = namespace or self.namespace
        url = f"http://{function_name}.{ns}.svc.cluster.local/"
        
        try:
            # Try to reach the service (health check)
            response = await self.client.get(f"{url}health", timeout=5.0)
            if response.status_code == 200:
                return {"available": True, "reason": "Service is ready"}
            else:
                return {
                    "available": False,
                    "reason": f"Service returned HTTP {response.status_code}"
                }
        except httpx.ConnectError:
            return {
                "available": False,
                "reason": "Service is not reachable (connection error)"
            }
        except httpx.TimeoutException:
            return {
                "available": False,
                "reason": "Service health check timed out"
            }
        except Exception as e:
            return {
                "available": False,
                "reason": f"Service check failed: {str(e)}"
            }
    
    async def call_lambda_function(
        self,
        function_name: str,
        parameters: Dict[str, Any],
        namespace: Optional[str] = None,
        correlation_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Call a LambdaFunction remediation function.
        
        Args:
            function_name: Name of the LambdaFunction
            parameters: Parameters to pass to the function
            namespace: Namespace (default: self.namespace)
        
        Returns:
            Response from the LambdaFunction
        """
        ns = namespace or self.namespace
        url = f"http://{function_name}.{ns}.svc.cluster.local/"
        
        # Generate correlation ID if not provided
        if not correlation_id:
            correlation_id = str(uuid.uuid4())
        
        # Check if LambdaFunction is available before attempting call
        availability = await self.check_lambda_function_availability(function_name, ns)
        if not availability.get("available", False):
            logger.error(
                "lambda_function_unavailable",
                function_name=function_name,
                namespace=ns,
                correlation_id=correlation_id,
                reason=availability.get("reason", "Unknown"),
                cannot_fix=True
            )
            return {
                "status": "error",
                "message": f"LambdaFunction {function_name} is not available",
                "error": availability.get("reason", "Service unavailable"),
                "correlation_id": correlation_id,
                "cannot_fix": True
            }
        
        with tracer.start_as_current_span(
            "lambda_function.call",
            attributes={
                "lambda_function": function_name,
                "namespace": ns,
                "correlation_id": correlation_id,
                "url": url,
            }
        ) as span:
            logger.info(
                "calling_lambda_function",
                function_name=function_name,
                namespace=ns,
                parameters=parameters,
                correlation_id=correlation_id,
                url=url
            )
            
            try:
                # Create CloudEvent for LambdaFunction (runtime expects CloudEvents)
                event = CloudEvent({
                    "type": "io.homelab.agent-sre.remediation.request",
                    "source": "agent-sre",
                    "id": correlation_id,
                    "specversion": "1.0",
                }, parameters)
                
                # Add correlation ID to CloudEvent extensions
                event["correlationid"] = correlation_id
                
                # Convert to HTTP binary format
                headers, body = to_binary(event)
                
                # Add correlation ID to HTTP headers for traceability
                headers["X-Correlation-ID"] = correlation_id
                
                span.set_attributes({
                    "lambda_function.event_id": correlation_id,
                    "lambda_function.event_type": "io.homelab.agent-sre.remediation.request",
                })
                
                # Send CloudEvent to LambdaFunction
                response = await self.client.post(
                    url,
                    content=body,
                    headers=dict(headers)
                )
                response.raise_for_status()
                
                # Parse CloudEvent response
                from cloudevents.http import from_http
                response_event = from_http(response.headers, response.content)
                result = response_event.data if response_event.data else {}
                
                span.set_attributes({
                    "lambda_function.status": result.get("status", "unknown"),
                    "lambda_function.message": result.get("message", ""),
                    "lambda_function.http_status": response.status_code,
                })
                
                logger.info(
                    "lambda_function_completed",
                    function_name=function_name,
                    status=result.get("status"),
                    message=result.get("message"),
                    correlation_id=correlation_id,
                    http_status=response.status_code,
                    result=result
                )
                
                span.set_status(trace.Status(trace.StatusCode.OK))
                return result
            except httpx.HTTPStatusError as e:
                span.record_exception(e)
                span.set_status(trace.Status(trace.StatusCode.ERROR, f"HTTP {e.response.status_code}"))
                span.set_attributes({
                    "lambda_function.http_status": e.response.status_code,
                    "lambda_function.error": str(e),
                })
                
                logger.error(
                    "lambda_function_http_error",
                    function_name=function_name,
                    status_code=e.response.status_code,
                    error=str(e),
                    correlation_id=correlation_id,
                    exc_info=True
                )
                return {
                    "status": "error",
                    "message": f"HTTP error: {e.response.status_code}",
                    "error": str(e),
                    "correlation_id": correlation_id
                }
            except Exception as e:
                span.record_exception(e)
                span.set_status(trace.Status(trace.StatusCode.ERROR, str(e)))
                
                logger.error(
                    "lambda_function_error",
                    function_name=function_name,
                    error=str(e),
                    correlation_id=correlation_id,
                    exc_info=True
                )
                return {
                    "status": "error",
                    "message": f"Unexpected error: {str(e)}",
                    "error": str(e),
                    "correlation_id": correlation_id
                }
    
    async def execute_remediation(
        self,
        alertname: str,
        remediation_config: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Execute remediation based on runbook configuration.
        
        Args:
            alertname: Name of the alert
            remediation_config: Remediation configuration from runbook
                {
                    "lambda_function": "function-name",
                    "parameters": {
                        "name": "...",
                        "namespace": "..."
                    }
                }
        
        Returns:
            Result of remediation execution
        """
        function_name = remediation_config.get("lambda_function")
        parameters = remediation_config.get("parameters", {})
        
        if not function_name:
            logger.warning(
                "no_lambda_function_specified",
                alertname=alertname
            )
            return {
                "status": "error",
                "message": "No lambda_function specified in remediation config"
            }
        
        return await self.call_lambda_function(function_name, parameters)

