"""
SRE Agent - Main entry point for health report generation and CloudEvents handling.

Uses fine-tuned FunctionGemma 270M model via MLX-LM framework to generate
comprehensive SRE health reports from Prometheus metrics.

Also handles CloudEvents from prometheus-events to trigger Flux reconciliation.

Supports two operation modes:
- Agentic: Autonomous execution (default)
- Supervised: Requires approval before execution
"""
import os
import asyncio
import argparse
from typing import Optional, Dict, Any
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, Response
import uvicorn
import structlog
from opentelemetry import trace

from .agent import SREAgent
from .config import AgentConfig
from .flux_reconciler import FluxReconciler
from .lambda_caller import LambdaFunctionCaller
from .linear_handler import LinearHandler
from .jira_handler import JiraHandler
from .intelligent_remediation import (
    intelligent_remediation_selection,
    record_remediation_success
)
from .langgraph_workflow import run_remediation_workflow
from .approval_system import ApprovalManager, ApprovalProvider
from .observability import (
    initialize_observability,
    get_correlation_id,
    set_correlation_context,
    trace_remediation,
    log_remediation_step,
    record_cloudevent_received,
    get_tracer,
    get_current_trace_context,
)
from src.metrics_collector import MetricsCollector
from src.report_generator import ReportGenerator

logger = structlog.get_logger()

app = FastAPI(title="Agent-SRE", version="0.5.0")


# Global agent instance (initialized on startup)
agent: Optional[SREAgent] = None
flux_reconciler: Optional[FluxReconciler] = None
lambda_caller: Optional[LambdaFunctionCaller] = None
linear_handler: Optional[LinearHandler] = None
jira_handler: Optional[JiraHandler] = None
report_generator: Optional[ReportGenerator] = None
approval_manager: Optional[ApprovalManager] = None


@app.on_event("startup")
async def startup():
    """Initialize agent on startup."""
    global agent, flux_reconciler, lambda_caller, linear_handler, jira_handler, report_generator, approval_manager
    
    # Initialize observability first
    initialize_observability()
    
    config = AgentConfig.from_env()
    
    # Initialize components
    metrics_collector = MetricsCollector(
        prometheus_url=config.prometheus_url,
        timeout=config.prometheus_timeout
    )
    
    report_generator = ReportGenerator(
        model_name=config.model_name,
        model_backend=config.model_backend,
        mlx_enabled=config.mlx_enabled,
        ollama_url=config.ollama_url,
        anthropic_api_key=config.anthropic_api_key
    )
    
    agent = SREAgent(
        metrics_collector=metrics_collector,
        report_generator=report_generator,
        config=config
    )
    
    flux_reconciler = FluxReconciler()
    lambda_caller = LambdaFunctionCaller(namespace="ai", timeout=60)
    linear_handler = LinearHandler()  # Initialize Linear handler (gracefully handles missing API key)
    jira_handler = JiraHandler()  # Initialize Jira handler (gracefully handles missing credentials)
    
    # Initialize approval manager if in supervised mode
    operation_mode = os.getenv("OPERATION_MODE", "agentic")
    if operation_mode == "supervised":
        slack_webhook_url = os.getenv("SLACK_WEBHOOK_URL")
        slack_channel = os.getenv("SLACK_CHANNEL", "#agent-approvals")
        custom_endpoint = os.getenv("CUSTOM_APPROVAL_ENDPOINT")
        
        slack_config = None
        custom_config = None
        
        if slack_webhook_url:
            slack_config = {
                "webhook_url": slack_webhook_url,
                "channel": slack_channel,
                "callback_url": os.getenv("SLACK_CALLBACK_URL")
            }
        
        if custom_endpoint:
            custom_config = {
                "endpoint": custom_endpoint,
                "method": os.getenv("CUSTOM_APPROVAL_METHOD", "POST"),
                "callback_url": os.getenv("CUSTOM_APPROVAL_CALLBACK_URL"),
                "use_webhook": os.getenv("CUSTOM_APPROVAL_USE_WEBHOOK", "true").lower() == "true"
            }
        
        if slack_config or custom_config:
            approval_manager = ApprovalManager(
                slack_config=slack_config,
                custom_config=custom_config
            )
            logger.info("approval_manager_initialized", operation_mode=operation_mode)
    
    logger.info("agent_started", component="agent-sre", operation_mode=operation_mode)


@app.get("/health")
async def health():
    """
    Health check endpoint for Kubernetes probes.
    
    This endpoint is required by the LambdaAgent operator which hardcodes
    /health for readiness and liveness probes. The endpoint only responds
    when the pod is active (not scaled to zero), so it doesn't prevent
    scale-to-zero behavior. The queue-proxy handles health checks for
    scale-to-zero decisions.
    """
    return {"status": "ok"}


@app.get("/ready")
async def ready():
    """
    Readiness check endpoint for Knative queue-proxy.
    
    This endpoint is used by Knative's queue-proxy to determine if the
    user-container is ready to receive traffic. It checks that all required
    components are initialized.
    """
    global agent, flux_reconciler
    if agent is None or flux_reconciler is None:
        return JSONResponse(status_code=503, content={"status": "not ready"})
    return {"status": "ready"}


@app.post("/approval/callback")
async def approval_callback(request: Request):
    """
    Callback endpoint for approval responses from Slack or custom app.
    
    Expected payload:
    {
        "request_id": "abc-123",
        "provider": "slack" | "custom",
        "decision": "approve" | "reject",
        "user_id": "...",
        "user_name": "...",
        "timestamp": "..."
    }
    """
    global approval_manager
    
    if not approval_manager:
        return JSONResponse(
            status_code=503,
            content={"error": "Approval manager not available"}
        )
    
    try:
        payload = await request.json()
        provider_str = payload.get("provider")
        
        if provider_str == "slack":
            from .approval_system import ApprovalProvider
            # Handle Slack interactive message payload
            approval_request = await approval_manager.handle_approval_response(
                ApprovalProvider.SLACK,
                payload
            )
        elif provider_str == "custom":
            from .approval_system import ApprovalProvider
            approval_request = await approval_manager.handle_approval_response(
                ApprovalProvider.CUSTOM,
                payload
            )
        else:
            return JSONResponse(
                status_code=400,
                content={"error": f"Unknown provider: {provider_str}"}
            )
        
        if approval_request:
            logger.info(
                "approval_callback_processed",
                request_id=approval_request.request_id,
                provider=provider_str,
                decision=payload.get("decision"),
                status=approval_request.status
            )
            
            return JSONResponse(
                status_code=200,
                content={
                    "status": "processed",
                    "request_id": approval_request.request_id,
                    "approval_status": approval_request.status
                }
            )
        else:
            return JSONResponse(
                status_code=404,
                content={"error": "Approval request not found"}
            )
            
    except Exception as e:
        logger.error(
            "approval_callback_error",
            error=str(e),
            exc_info=True
        )
        return JSONResponse(
            status_code=500,
            content={"error": str(e)}
        )


@app.get("/metrics")
async def metrics():
    """
    Prometheus metrics endpoint.
    
    NOTE: OpenTelemetry metrics are exported via OTLP to Alloy, which converts
    them to Prometheus format. This endpoint is for backward compatibility.
    
    Metrics are only available when the pod is active (not scaled to zero).
    """
    from .observability import get_prometheus_metrics
    try:
        metrics_content = get_prometheus_metrics()
        from prometheus_client import CONTENT_TYPE_LATEST
        return Response(
            content=metrics_content,
            media_type=CONTENT_TYPE_LATEST
        )
    except Exception as e:
        logger.error("metrics_generation_failed", error=str(e))
        return JSONResponse(
            status_code=500,
            content={"error": "Failed to generate metrics"}
        )


@app.post("/")
async def handle_cloudevent(request: Request):
    """Handle CloudEvents from prometheus-events."""
    global agent, flux_reconciler, lambda_caller
    
    if agent is None or flux_reconciler is None or lambda_caller is None:
        return JSONResponse(
            status_code=503,
            content={"error": "Agent not initialized"}
        )
    
    # Extract correlation ID and set context
    headers = dict(request.headers)
    correlation_id = get_correlation_id(headers=headers)
    
    tracer = get_tracer()
    
    # Use context manager for span if tracer is available
    if tracer:
        with tracer.start_as_current_span("cloudevent.handle") as span:
            try:
                return await _process_cloudevent(request, headers, correlation_id, span)
            except Exception as e:
                from opentelemetry import trace as otel_trace
                span.record_exception(e)
                span.set_status(otel_trace.Status(otel_trace.StatusCode.ERROR, str(e)))
                logger.error(
                    "cloudevent_processing_failed",
                    error=str(e),
                    correlation_id=correlation_id,
                    exc_info=True
                )
                return JSONResponse(
                    status_code=500,
                    content={"error": str(e)}
                )
    else:
        # Fallback if OpenTelemetry not available
        try:
            return await _process_cloudevent(request, headers, correlation_id, None)
        except Exception as e:
            logger.error(
                "cloudevent_processing_failed",
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            return JSONResponse(
                status_code=500,
                content={"error": str(e)}
            )


async def _process_cloudevent(
    request: Request,
    headers: Dict[str, str],
    correlation_id: str,
    span: Optional[Any]
) -> JSONResponse:
    """Process CloudEvent with optional OpenTelemetry span."""
    global flux_reconciler
    
    # Parse CloudEvent headers
    body = await request.body()
    
    # Extract CloudEvent information
    from cloudevents.http import from_http
    import json
    
    # Handle application/cloudevents+json format (structured content mode)
    # When Content-Type is application/cloudevents+json, the entire event is in the body
    content_type = headers.get("content-type", "").lower()
    if "application/cloudevents+json" in content_type:
        # Parse JSON body directly - it contains the full CloudEvent
        try:
            event_dict = json.loads(body)
            event_id = event_dict.get("id")
            event_type = event_dict.get("type")
            event_source = event_dict.get("source")
            event_data = event_dict.get("data", {})
        except (json.JSONDecodeError, AttributeError) as e:
            logger.error("failed_to_parse_cloudevent_json", error=str(e), body_preview=str(body)[:200])
            # Fallback to from_http
            event = from_http(headers, body)
            event_id = event.get("id")
            event_type = event.get("type")
            event_source = event.get("source")
            event_data = event.get("data", {})
    else:
        # Use standard from_http for binary/structured content mode
        event = from_http(headers, body)
        event_id = event.get("id")
        event_type = event.get("type")
        event_source = event.get("source")
        event_data = event.get("data", {})
    
    # Debug: Log event_data structure for troubleshooting
    if isinstance(event_data, dict):
        logger.info(
            "cloudevent_data_structure",
            event_id=event_id,
            has_labels="labels" in event_data,
            has_annotations="annotations" in event_data,
            labels_keys=list(event_data.get("labels", {}).keys()) if isinstance(event_data.get("labels"), dict) else None,
            annotations_keys=list(event_data.get("annotations", {}).keys()) if isinstance(event_data.get("annotations"), dict) else None,
            event_data_keys=list(event_data.keys())[:10]  # First 10 keys for debugging
        )
    
    # Update correlation ID with event ID if available
    if event_id:
        correlation_id = event_id
        set_correlation_context(correlation_id, event_id=event_id)
    
    # Update span with event details
    if span:
        span.set_attribute("event_id", event_id)
        span.set_attribute("event_type", event_type)
        span.set_attribute("event_source", event_source)
    
    # Record metrics using OpenTelemetry
    record_cloudevent_received(event_type, event_source)
    
    logger.info(
        "cloudevent_received",
        event_id=event_id,
        event_source=event_source,
        event_type=event_type,
        correlation_id=correlation_id
    )
    
    # Handle Prometheus alert events
    if event_type.startswith("io.homelab.prometheus.alert"):
        # Use LangGraph workflow for remediation (new approach)
        global report_generator, approval_manager
        try:
            # Only use LangGraph for firing alerts
            if event_type == "io.homelab.prometheus.alert.fired":
                # Get operation mode and approval config from environment
                operation_mode = os.getenv("OPERATION_MODE", "agentic")
                approval_config = None
                
                if operation_mode == "supervised":
                    # Build approval config from environment variables
                    providers = []
                    if os.getenv("SLACK_WEBHOOK_URL"):
                        providers.append("slack")
                    if os.getenv("CUSTOM_APPROVAL_ENDPOINT"):
                        providers.append("custom")
                    
                    if providers:
                        approval_config = {
                            "providers": providers,
                            "require_all": os.getenv("APPROVAL_REQUIRE_ALL", "false").lower() == "true",
                            "timeout": int(os.getenv("APPROVAL_TIMEOUT_SECONDS", "3600")),
                            "timeout_action": os.getenv("APPROVAL_TIMEOUT_ACTION", "pending")
                        }
                
                result_state = await run_remediation_workflow(
                    event_data=event_data,
                    event_type=event_type,
                    event_id=event_id,
                    correlation_id=correlation_id,
                    lambda_caller=lambda_caller,
                    report_generator=report_generator,
                    approval_manager=approval_manager,
                    operation_mode=operation_mode,
                    approval_config=approval_config
                )
                
                # Log workflow completion
                logger.info(
                    "remediation_workflow_completed",
                    alertname=result_state.alertname,
                    success=result_state.success,
                    lambda_function=result_state.lambda_function,
                    method=result_state.method,
                    correlation_id=correlation_id
                )
                
                # Record success for training if successful
                if result_state.success and result_state.lambda_function:
                    try:
                        record_remediation_success(
                            alert_data=event_data,
                            lambda_function=result_state.lambda_function,
                            parameters=result_state.lambda_parameters,
                            success=True
                        )
                    except Exception as e:
                        logger.debug("failed_to_record_success", error=str(e))
            else:
                # For resolved alerts, use legacy handler
                await handle_prometheus_alert(
                    event_data,
                    event_type,
                    flux_reconciler,
                    correlation_id=correlation_id,
                    event_id=event_id
                )
            
        except Exception as e:
            logger.error(
                "langgraph_workflow_failed",
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            # Fallback to legacy handler
            await handle_prometheus_alert(
                event_data,
                event_type,
                flux_reconciler,
                correlation_id=correlation_id,
                event_id=event_id
            )
    
    if span:
        from opentelemetry import trace as otel_trace
        span.set_status(otel_trace.Status(otel_trace.StatusCode.OK))
    
    return JSONResponse(
        status_code=200,
        content={"status": "processed", "event_id": event_id, "correlation_id": correlation_id}
    )


async def handle_prometheus_alert(
    event_data: Dict[str, Any],
    event_type: str,
    reconciler: FluxReconciler,
    correlation_id: Optional[str] = None,
    event_id: Optional[str] = None
):
    """Handle Prometheus alert CloudEvent."""
    global lambda_caller, linear_handler, jira_handler
    import uuid
    
    labels = event_data.get("labels", {})
    annotations = event_data.get("annotations", {})
    common_annotations = event_data.get("commonAnnotations", {})
    prometheus_rule = event_data.get("prometheusRule") or labels.get("prometheus_rule")
    alertname = labels.get("alertname", "unknown")
    
    # Merge annotations (alert-specific take precedence over common)
    all_annotations = {**common_annotations, **annotations}
    
    # Set correlation context for this alert
    if correlation_id:
        set_correlation_context(correlation_id, event_id=event_id, alertname=alertname)
    
    tracer = get_tracer()
    
    # Use context manager for span if tracer is available
    if tracer:
        with tracer.start_as_current_span(
            "alert.process",
            attributes={
                "alertname": alertname,
                "event_type": event_type,
                "correlation_id": correlation_id or "",
                "event_id": event_id or "",
                "has_flux_reconcile": bool(all_annotations.get("flux_reconcile")),
                "has_lambda_function": bool(all_annotations.get("lambda_function")),
            }
        ) as span:
            await _process_prometheus_alert(
                event_data, event_type, reconciler, correlation_id, event_id, span
            )
    else:
        await _process_prometheus_alert(
            event_data, event_type, reconciler, correlation_id, event_id, None
        )


async def _process_prometheus_alert(
    event_data: Dict[str, Any],
    event_type: str,
    reconciler: FluxReconciler,
    correlation_id: Optional[str],
    event_id: Optional[str],
    span: Optional[Any]
):
    """Process Prometheus alert with optional OpenTelemetry span."""
    global lambda_caller, linear_handler, jira_handler, report_generator
    import uuid
    
    # Extract alert data from event_data
    labels = event_data.get("labels", {})
    annotations = event_data.get("annotations", {})
    common_annotations = event_data.get("commonAnnotations", {})
    prometheus_rule = event_data.get("prometheusRule") or labels.get("prometheus_rule")
    alertname = labels.get("alertname", "unknown")
    
    # Merge annotations (alert-specific take precedence over common)
    all_annotations = {**common_annotations, **annotations}
    
    # Set correlation context for this alert
    if correlation_id:
        set_correlation_context(correlation_id, event_id=event_id, alertname=alertname)
    
    try:
        logger.info(
            "processing_prometheus_alert",
            alertname=alertname,
            event_type=event_type,
            prometheus_rule=prometheus_rule,
            has_flux_reconcile=bool(all_annotations.get("flux_reconcile")),
            has_lambda_function=bool(all_annotations.get("lambda_function")),
            correlation_id=correlation_id,
            event_id=event_id
        )
        
        # Only process firing alerts
        if event_type == "io.homelab.prometheus.alert.fired":
            # Create Linear ticket for the alert (if Linear is available)
            if linear_handler:
                try:
                    ticket_url = await linear_handler.create_alert_ticket(
                        alert=event_data,
                        correlation_id=correlation_id
                    )
                    if ticket_url:
                        logger.info(
                            "linear_ticket_created_for_alert",
                            alertname=alertname,
                            ticket_url=ticket_url,
                            correlation_id=correlation_id
                        )
                        if span:
                            span.set_attribute("linear.ticket_created", True)
                            span.set_attribute("linear.ticket_url", ticket_url)
                except Exception as e:
                    logger.warning(
                        "failed_to_create_linear_ticket_for_alert",
                        alertname=alertname,
                        error=str(e),
                        correlation_id=correlation_id
                    )
            
            # Create Jira ticket for the alert (if Jira is available)
            if jira_handler:
                try:
                    jira_issue_key = await jira_handler.create_alert_ticket(
                        alert=event_data,
                        correlation_id=correlation_id
                    )
                    if jira_issue_key:
                        logger.info(
                            "jira_ticket_created_for_alert",
                            alertname=alertname,
                            issue_key=jira_issue_key,
                            correlation_id=correlation_id
                        )
                        if span:
                            span.set_attribute("jira.ticket_created", True)
                            span.set_attribute("jira.issue_key", jira_issue_key)
                except Exception as e:
                    logger.warning(
                        "failed_to_create_jira_ticket_for_alert",
                        alertname=alertname,
                        error=str(e),
                        correlation_id=correlation_id
                    )
            
            # Priority 1: Check for LambdaFunction remediation (new approach)
            lambda_function = all_annotations.get("lambda_function")
            lambda_parameters_json = all_annotations.get("lambda_parameters", "{}")
            
            # If no static annotation, try AI-powered selection
            if not lambda_function and report_generator:
                try:
                    logger.info(
                        "attempting_ai_remediation_selection",
                        alertname=alertname,
                        correlation_id=correlation_id
                    )
                    
                    # Get TRM model path from env or use default
                    trm_model_path = os.getenv(
                        "TRM_MODEL_PATH",
                        "/workspace/bruno/repos/homelab/flux/ai/trm/checkpoints/Trm_data-ACT-torch/trm-runbook-extended/step_0"
                    )
                    
                    ai_result = await intelligent_remediation_selection(
                        alert_data=event_data,
                        report_generator=report_generator,
                        use_rag=True,
                        use_few_shot=True,
                        use_trm=True,  # Enable TRM model
                        trm_model_path=trm_model_path
                    )
                    
                    lambda_function = ai_result["lambda_function"]
                    parameters = ai_result["parameters"]
                    
                    logger.info(
                        "ai_remediation_selected",
                        alertname=alertname,
                        lambda_function=lambda_function,
                        method=ai_result.get("method", "ai"),
                        confidence=ai_result.get("confidence", 0.0),
                        correlation_id=correlation_id
                    )
                    
                    # Use AI-selected function and parameters
                    if lambda_function and lambda_caller:
                        # Skip to execution below
                        pass
                    else:
                        logger.warning(
                            "ai_selection_failed_no_function",
                            alertname=alertname,
                            correlation_id=correlation_id
                        )
                        # Fall through to legacy Flux reconciliation
                        lambda_function = None
                        
                except Exception as e:
                    logger.warning(
                        "ai_remediation_selection_failed",
                        alertname=alertname,
                        error=str(e),
                        correlation_id=correlation_id,
                        exc_info=True
                    )
                # Fall through to static annotation or legacy Flux reconciliation
                lambda_function = None
            
            if lambda_function and lambda_caller:
                import json
                try:
                    logger.debug(
                        "lambda_function_check_passed",
                        alertname=alertname,
                        lambda_function=lambda_function,
                        lambda_caller_initialized=lambda_caller is not None,
                        lambda_parameters_json=lambda_parameters_json,
                        correlation_id=correlation_id
                    )
                    
                    # Parse parameters from annotation (JSON string) if not already set by AI
                    if "parameters" not in locals():
                        parameters = json.loads(lambda_parameters_json) if isinstance(lambda_parameters_json, str) else lambda_parameters_json
                        
                        # Extract dynamic parameters from alert labels
                        # Common parameters: name, namespace, pod_name, deployment_name, etc.
                        if "name" not in parameters:
                            parameters["name"] = labels.get("name") or labels.get("resource_name")
                        if "namespace" not in parameters:
                            parameters["namespace"] = labels.get("namespace") or labels.get("resource_namespace")
                    
                    # Trace remediation with full context
                    remediation_correlation_id = correlation_id or event_id or str(uuid.uuid4())
                    with trace_remediation(alertname, lambda_function, remediation_correlation_id) as remediation_span:
                        log_remediation_step(
                        "remediation.started",
                        alertname=alertname,
                        lambda_function=lambda_function,
                        correlation_id=remediation_correlation_id,
                        parameters=parameters
                    )
                    
                    if remediation_span:
                        remediation_span.set_attribute("remediation.lambda_function", lambda_function)
                        remediation_span.set_attribute("remediation.parameters", str(parameters))
                    
                    logger.info(
                        "calling_remediation_lambda_function",
                        alertname=alertname,
                        lambda_function=lambda_function,
                        parameters=parameters,
                        correlation_id=remediation_correlation_id
                    )
                    
                    result = await lambda_caller.call_lambda_function(
                        function_name=lambda_function,
                        parameters=parameters,
                        correlation_id=remediation_correlation_id
                    )
                    
                    # Metrics are recorded in trace_remediation context manager
                    remediation_status = result.get("status", "unknown")
                    
                    # Record successful remediation for training (Phase 2 & 3)
                    if remediation_status == "success":
                        try:
                            record_remediation_success(
                                alert_data=event_data,
                                lambda_function=lambda_function,
                                parameters=parameters,
                                success=True
                            )
                        except Exception as e:
                            logger.debug("failed_to_record_success", error=str(e))
                    
                    # Check if remediation failed
                    if remediation_status == "error":
                        # Log detailed failure information
                        error_message = result.get("message", "Unknown error")
                        error_details = result.get("error", "")
                        
                        logger.error(
                            "remediation_failed",
                            alertname=alertname,
                            lambda_function=lambda_function,
                            status=remediation_status,
                            message=error_message,
                            error=error_details,
                            correlation_id=remediation_correlation_id,
                            pod_name=parameters.get("name"),
                            namespace=parameters.get("namespace"),
                            result=result,
                            cannot_fix=True
                        )
                        
                        # Create Linear ticket for remediation failure
                        if linear_handler:
                            try:
                                failure_ticket_url = await linear_handler.create_remediation_failure_ticket(
                                    alertname=alertname,
                                    lambda_function=lambda_function,
                                    error_message=error_message,
                                    parameters=parameters,
                                    correlation_id=remediation_correlation_id
                                )
                                if failure_ticket_url:
                                    logger.info(
                                        "linear_ticket_created_for_remediation_failure",
                                        alertname=alertname,
                                        lambda_function=lambda_function,
                                        ticket_url=failure_ticket_url,
                                        correlation_id=remediation_correlation_id
                                    )
                                    if remediation_span:
                                        remediation_span.set_attribute("linear.failure_ticket_created", True)
                                        remediation_span.set_attribute("linear.failure_ticket_url", failure_ticket_url)
                            except Exception as e:
                                logger.warning(
                                    "failed_to_create_linear_ticket_for_remediation_failure",
                                    alertname=alertname,
                                    lambda_function=lambda_function,
                                    error=str(e),
                                    correlation_id=remediation_correlation_id
                                )
                        
                        # Create Jira ticket for remediation failure
                        if jira_handler:
                            try:
                                jira_failure_issue_key = await jira_handler.create_remediation_failure_ticket(
                                    alertname=alertname,
                                    lambda_function=lambda_function,
                                    error_message=error_message,
                                    parameters=parameters,
                                    correlation_id=remediation_correlation_id
                                )
                                if jira_failure_issue_key:
                                    logger.info(
                                        "jira_ticket_created_for_remediation_failure",
                                        alertname=alertname,
                                        lambda_function=lambda_function,
                                        issue_key=jira_failure_issue_key,
                                        correlation_id=remediation_correlation_id
                                    )
                                    if remediation_span:
                                        remediation_span.set_attribute("jira.failure_ticket_created", True)
                                        remediation_span.set_attribute("jira.failure_issue_key", jira_failure_issue_key)
                            except Exception as e:
                                logger.warning(
                                    "failed_to_create_jira_ticket_for_remediation_failure",
                                    alertname=alertname,
                                    lambda_function=lambda_function,
                                    error=str(e),
                                    correlation_id=remediation_correlation_id
                                )
                        
                        # Set span attributes for failed remediation
                        if remediation_span:
                            remediation_span.set_attribute("remediation.status", "error")
                            remediation_span.set_attribute("remediation.message", error_message)
                            remediation_span.set_attribute("remediation.error", error_details)
                            remediation_span.set_attribute("remediation.cannot_fix", True)
                        
                        if span:
                            span.set_attribute("remediation.status", "error")
                            span.set_attribute("remediation.message", error_message)
                            span.set_attribute("remediation.cannot_fix", True)
                    else:
                        log_remediation_step(
                            "remediation.completed",
                            alertname=alertname,
                            lambda_function=lambda_function,
                            correlation_id=remediation_correlation_id,
                            status=remediation_status,
                            message=result.get("message")
                        )
                        
                        logger.info(
                            "remediation_completed",
                            alertname=alertname,
                            lambda_function=lambda_function,
                            status=remediation_status,
                            message=result.get("message"),
                            correlation_id=remediation_correlation_id,
                            result=result
                        )
                        
                        if remediation_span:
                            remediation_span.set_attribute("remediation.status", remediation_status)
                            remediation_span.set_attribute("remediation.message", result.get("message", ""))
                        
                        if span:
                            span.set_attribute("remediation.status", remediation_status)
                            span.set_attribute("remediation.message", result.get("message", ""))
                
                    return
                except json.JSONDecodeError as e:
                    logger.error(
                        "lambda_parameters_parse_error",
                        alertname=alertname,
                        lambda_function=lambda_function,
                        parameters_json=lambda_parameters_json,
                        error=str(e)
                    )
                except Exception as e:
                    logger.error(
                        "lambda_function_call_error",
                        alertname=alertname,
                        lambda_function=lambda_function,
                        error=str(e),
                        error_type=type(e).__name__
                    )
            else:
                logger.warning(
                    "lambda_function_condition_failed",
                    alertname=alertname,
                    lambda_function=lambda_function,
                    lambda_caller_initialized=lambda_caller is not None,
                    has_lambda_function=bool(lambda_function),
                    has_lambda_caller=bool(lambda_caller)
                )
            
            # Priority 2: Fallback to Flux reconciliation (legacy approach)
            flux_reconcile_target = all_annotations.get("flux_reconcile")
            flux_reconcile_kind = all_annotations.get("flux_reconcile_kind", "Kustomization")
            flux_reconcile_namespace = all_annotations.get("flux_reconcile_namespace", "flux-system")
            
            if flux_reconcile_target:
                logger.info(
                    "triggering_flux_reconciliation",
                    target=flux_reconcile_target,
                    kind=flux_reconcile_kind,
                    namespace=flux_reconcile_namespace
                )
                
                # Reconcile based on kind
                success = False
                if flux_reconcile_kind.lower() == "kustomization":
                    success = reconciler.reconcile_kustomization(flux_reconcile_target, flux_reconcile_namespace)
                elif flux_reconcile_kind.lower() == "gitrepository":
                    success = reconciler.reconcile_gitrepository(flux_reconcile_target, flux_reconcile_namespace)
                elif flux_reconcile_kind.lower() == "helmrelease":
                    success = reconciler.reconcile_helmrelease(flux_reconcile_target, flux_reconcile_namespace)
                else:
                    logger.warning(
                        "unknown_flux_kind",
                        kind=flux_reconcile_kind,
                        target=flux_reconcile_target
                    )
                
                logger.info(
                    "flux_reconciliation_completed",
                    target=flux_reconcile_target,
                    kind=flux_reconcile_kind,
                    namespace=flux_reconcile_namespace,
                    success=success
                )
            else:
                logger.info(
                    "no_remediation_annotation",
                    alertname=alertname,
                    annotations=all_annotations
                )
    finally:
        # Set final span status if span is available
        if span:
            from opentelemetry import trace as otel_trace
            span.set_status(otel_trace.Status(otel_trace.StatusCode.OK))


async def main(component: Optional[str] = None):
    """Main entry point for SRE agent (CLI mode)."""
    config = AgentConfig.from_env()
    
    # Initialize components
    metrics_collector = MetricsCollector(
        prometheus_url=config.prometheus_url,
        timeout=config.prometheus_timeout
    )
    
    report_generator = ReportGenerator(
        model_name=config.model_name,
        model_backend=config.model_backend,
        mlx_enabled=config.mlx_enabled,
        ollama_url=config.ollama_url,
        anthropic_api_key=config.anthropic_api_key
    )
    
    agent = SREAgent(
        metrics_collector=metrics_collector,
        report_generator=report_generator,
        config=config
    )
    
    # Generate report
    if component:
        report = await agent.generate_component_report(component)
    else:
        report = await agent.generate_full_report()
    
    # Output report
    print(report.to_markdown())
    
    return report


def cli():
    """CLI entry point."""
    parser = argparse.ArgumentParser(description="SRE Health Report Agent")
    parser.add_argument(
        "--component",
        choices=["loki", "prometheus", "infrastructure", "observability"],
        help="Generate report for specific component"
    )
    parser.add_argument(
        "--output",
        choices=["markdown", "json", "html"],
        default="markdown",
        help="Output format"
    )
    parser.add_argument(
        "--prometheus-url",
        default=os.getenv("PROMETHEUS_URL", "http://prometheus:9090"),
        help="Prometheus URL"
    )
    parser.add_argument(
        "--server",
        action="store_true",
        help="Run as web server (CloudEvents handler)"
    )
    parser.add_argument(
        "--port",
        type=int,
        default=int(os.getenv("PORT", "8000")),
        help="Server port"
    )
    
    args = parser.parse_args()
    
    # Run as server if requested
    if args.server:
        uvicorn.run(app, host="0.0.0.0", port=args.port)
    else:
        # Run async main (CLI mode)
        report = asyncio.run(main(component=args.component))
        
        # Output in requested format
        if args.output == "json":
            print(report.to_json())
        elif args.output == "html":
            print(report.to_html())
        else:
            print(report.to_markdown())


if __name__ == "__main__":
    cli()

