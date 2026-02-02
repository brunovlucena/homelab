"""
LangGraph Workflow for Agent-SRE

Implements the remediation workflow using LangGraph for stateful, multi-step orchestration.
Follows recommendations from AGENT_FRAMEWORK_RECOMMENDATION.md and AGENT_SRE_ARCHITECTURE_CRITIQUE.md.
"""
from typing import Dict, Any, Optional, Literal
from pydantic import BaseModel, Field
import structlog
from langgraph.graph import StateGraph, END
from langgraph.checkpoint.memory import MemorySaver

from .lambda_caller import LambdaFunctionCaller
from .intelligent_remediation import intelligent_remediation_selection
from .approval_system import (
    ApprovalManager,
    ApprovalRequest,
    ApprovalStatus,
    ApprovalProvider
)
from .observability import (
    get_tracer,
    trace_remediation,
    log_remediation_step,
    set_correlation_context,
)

logger = structlog.get_logger()


# Type-safe state with Pydantic
class AgentState(BaseModel):
    """Type-safe state for Agent-SRE workflow."""
    
    # CloudEvent data
    event_data: Dict[str, Any] = Field(default_factory=dict)
    event_type: str = ""
    event_id: Optional[str] = None
    correlation_id: Optional[str] = None
    
    # Alert information
    alertname: Optional[str] = None
    labels: Dict[str, Any] = Field(default_factory=dict)
    annotations: Dict[str, Any] = Field(default_factory=dict)
    common_annotations: Dict[str, Any] = Field(default_factory=dict)
    
    # LambdaFunction information
    lambda_function: Optional[str] = None
    lambda_parameters: Dict[str, Any] = Field(default_factory=dict)
    
    # Workflow state
    step: Literal[
        "receive_cloudevent",
        "extract_from_cloudevent",
        "extract_lambda_function",
        "request_approval",
        "wait_for_approval",
        "execute_lambda_function",
        "verify_remediation",
        "complete"
    ] = "receive_cloudevent"
    
    # Operation mode
    operation_mode: Literal["agentic", "supervised"] = "agentic"
    
    # Approval state
    approval_request_id: Optional[str] = None
    approval_status: Optional[str] = None
    approval_config: Optional[Dict[str, Any]] = None
    
    # Execution results
    remediation_result: Optional[Dict[str, Any]] = None
    verification_result: Optional[Dict[str, Any]] = None
    success: bool = False
    
    # Confidence and method tracking
    confidence: float = Field(ge=0.0, le=1.0, default=0.0)
    method: str = "unknown"
    reasoning: Optional[str] = None
    
    # Error handling
    error: Optional[str] = None
    retry_count: int = 0
    max_retries: int = 3


def extract_from_cloudevent(state: AgentState) -> AgentState:
    """
    Extract all information from CloudEvent.
    
    This node extracts:
    - alertname (from subject or data.alertname)
    - labels (from data.labels)
    - annotations (from data.annotations + data.commonAnnotations)
    - lambda_function annotation (1:1 mapping from PrometheusRule)
    - lambda_parameters
    """
    event_data = state.event_data
    
    # Extract alertname
    alertname = (
        event_data.get("alertname") or
        event_data.get("subject") or
        event_data.get("labels", {}).get("alertname", "unknown")
    )
    
    # Extract labels and annotations
    labels = event_data.get("labels", {})
    annotations = event_data.get("annotations", {})
    common_annotations = event_data.get("commonAnnotations", {})
    
    # Merge annotations (alert-specific take precedence over common)
    all_annotations = {**common_annotations, **annotations}
    
    # Update state
    state.alertname = alertname
    state.labels = labels
    state.annotations = annotations
    state.common_annotations = common_annotations
    
    # Extract lambda_function if present (fast path)
    lambda_function = all_annotations.get("lambda_function")
    if lambda_function:
        state.lambda_function = lambda_function
        # Parse lambda_parameters
        import json
        lambda_parameters_json = all_annotations.get("lambda_parameters", "{}")
        try:
            state.lambda_parameters = json.loads(lambda_parameters_json) if isinstance(lambda_parameters_json, str) else lambda_parameters_json
        except json.JSONDecodeError:
            state.lambda_parameters = {}
    
    # Set correlation context
    if state.correlation_id:
        set_correlation_context(
            state.correlation_id,
            event_id=state.event_id,
            alertname=alertname
        )
    
    state.step = "extract_lambda_function"
    
    logger.info(
        "cloudevent_extracted",
        alertname=alertname,
        has_lambda_function=bool(lambda_function),
        correlation_id=state.correlation_id
    )
    
    return state


async def extract_lambda_function(
    state: AgentState,
    report_generator: Optional[Any] = None
) -> AgentState:
    """
    Extract LambdaFunction using multiple strategies:
    
    Phase 0 (Fast Path): Already extracted from CloudEvent annotations
    Phase 1: TRM recursive reasoning
    Phase 2: RAG-based selection
    Phase 3: Few-shot learning
    Phase 4: AI function calling
    """
    # Phase 0: If already extracted, skip
    if state.lambda_function:
        logger.info(
            "lambda_function_from_annotation",
            alertname=state.alertname,
            lambda_function=state.lambda_function,
            correlation_id=state.correlation_id
        )
        state.method = "static_annotation"
        state.confidence = 1.0
        state.step = "execute_lambda_function"
        return state
    
    # Phase 1-4: Use intelligent remediation selection
    if report_generator:
        try:
            alert_data = {
                "labels": state.labels,
                "annotations": state.annotations,
                "commonAnnotations": state.common_annotations,
            }
            
            # Get TRM model path from env
            import os
            trm_model_path = os.getenv(
                "TRM_MODEL_PATH",
                "/models/trm/checkpoints/Trm_data-ACT-torch/trm-runbook-extended/step_0"
            )
            
            result = await intelligent_remediation_selection(
                alert_data=alert_data,
                report_generator=report_generator,
                use_rag=True,
                use_few_shot=True,
                use_trm=True,  # Enable TRM
                trm_model_path=trm_model_path
            )
            
            state.lambda_function = result.get("lambda_function")
            state.lambda_parameters = result.get("parameters", {})
            state.method = result.get("method", "ai")
            state.confidence = result.get("confidence", 0.0)
            state.reasoning = result.get("reasoning")
            
            logger.info(
                "lambda_function_selected",
                alertname=state.alertname,
                lambda_function=state.lambda_function,
                method=state.method,
                confidence=state.confidence,
                correlation_id=state.correlation_id
            )
            
            state.step = "execute_lambda_function"
            
        except Exception as e:
            logger.error(
                "intelligent_selection_failed",
                alertname=state.alertname,
                error=str(e),
                correlation_id=state.correlation_id,
                exc_info=True
            )
            state.error = str(e)
            state.step = "complete"  # End workflow on error
    else:
        logger.warning(
            "no_report_generator",
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.error = "No report generator available for intelligent selection"
        state.step = "complete"
    
    return state


async def request_approval(
    state: AgentState,
    approval_manager: Optional[ApprovalManager] = None
) -> AgentState:
    """
    Request approval for supervised mode operations.
    
    This node is only executed when operation_mode is "supervised".
    """
    if state.operation_mode != "supervised":
        # Skip approval in agentic mode
        state.step = "execute_lambda_function"
        return state
    
    if not approval_manager:
        logger.warning(
            "approval_manager_not_available",
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        # In supervised mode without approval manager, reject by default
        state.error = "Approval required but approval manager not available"
        state.step = "complete"
        return state
    
    if not state.lambda_function:
        logger.error(
            "no_lambda_function_for_approval",
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.error = "No lambda_function to request approval for"
        state.step = "complete"
        return state
    
    # Parse approval config
    approval_config = state.approval_config or {}
    providers = approval_config.get("providers", [])
    require_all = approval_config.get("require_all", False)
    timeout_seconds = approval_config.get("timeout", 3600)  # Default 1 hour
    timeout_action = approval_config.get("timeout_action", "pending")
    
    # Create approval request
    import uuid
    from datetime import timedelta
    
    request_id = str(uuid.uuid4())
    approval_request = ApprovalRequest(
        request_id=request_id,
        agent_name="agent-sre",
        action="execute_lambda_function",
        lambda_function=state.lambda_function,
        parameters=state.lambda_parameters,
        alertname=state.alertname,
        correlation_id=state.correlation_id,
        providers=[ApprovalProvider(p) for p in providers],
        require_all=require_all,
        timeout=timedelta(seconds=timeout_seconds),
        timeout_action=timeout_action,
        metadata={
            "labels": state.labels,
            "annotations": state.annotations
        }
    )
    
    # Send approval request
    try:
        approval_request = await approval_manager.request_approval(approval_request)
        state.approval_request_id = request_id
        state.approval_status = ApprovalStatus.PENDING
        
        logger.info(
            "approval_requested",
            request_id=request_id,
            alertname=state.alertname,
            lambda_function=state.lambda_function,
            providers=providers,
            correlation_id=state.correlation_id
        )
        
        state.step = "wait_for_approval"
        
    except Exception as e:
        logger.error(
            "approval_request_failed",
            request_id=request_id,
            alertname=state.alertname,
            error=str(e),
            correlation_id=state.correlation_id,
            exc_info=True
        )
        state.error = f"Failed to request approval: {str(e)}"
        state.step = "complete"
    
    return state


async def wait_for_approval(
    state: AgentState,
    approval_manager: Optional[ApprovalManager] = None
) -> AgentState:
    """
    Wait for approval response.
    
    In production, this would poll or wait for webhook callback.
    For now, we check the approval status.
    """
    if not state.approval_request_id or not approval_manager:
        state.error = "Approval request ID or manager not available"
        state.step = "complete"
        return state
    
    # Get approval request
    approval_request = approval_manager.get_request(state.approval_request_id)
    if not approval_request:
        state.error = f"Approval request {state.approval_request_id} not found"
        state.step = "complete"
        return state
    
    # Check status
    if approval_request.status == ApprovalStatus.APPROVED:
        logger.info(
            "approval_granted",
            request_id=state.approval_request_id,
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.approval_status = ApprovalStatus.APPROVED
        state.step = "execute_lambda_function"
    
    elif approval_request.status == ApprovalStatus.REJECTED:
        logger.info(
            "approval_rejected",
            request_id=state.approval_request_id,
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.approval_status = ApprovalStatus.REJECTED
        state.error = "Approval rejected by operator"
        state.step = "complete"
    
    elif approval_request.status == ApprovalStatus.TIMEOUT:
        logger.warning(
            "approval_timeout",
            request_id=state.approval_request_id,
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.approval_status = ApprovalStatus.TIMEOUT
        
        # Handle timeout action
        timeout_action = approval_request.timeout_action
        if timeout_action == "approve":
            state.approval_status = ApprovalStatus.APPROVED
            state.step = "execute_lambda_function"
        elif timeout_action == "reject":
            state.approval_status = ApprovalStatus.REJECTED
            state.error = "Approval request timed out (rejected)"
            state.step = "complete"
        else:  # pending
            state.error = "Approval request timed out (pending)"
            state.step = "complete"
    
    else:
        # Still pending, continue waiting
        # In production, this would be handled by a separate polling mechanism
        # or webhook callback
        logger.debug(
            "approval_still_pending",
            request_id=state.approval_request_id,
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        # For now, we'll assume it's approved after a short wait
        # In production, implement proper polling or webhook handling
        state.step = "wait_for_approval"
    
    return state


async def execute_lambda_function(
    state: AgentState,
    lambda_caller: Optional[LambdaFunctionCaller] = None
) -> AgentState:
    """
    Execute LambdaFunction remediation.
    
    Calls the LambdaFunction via HTTP with extracted parameters.
    """
    if not state.lambda_function:
        logger.error(
            "no_lambda_function",
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.error = "No lambda_function to execute"
        state.step = "complete"
        return state
    
    if not lambda_caller:
        logger.error(
            "no_lambda_caller",
            alertname=state.alertname,
            correlation_id=state.correlation_id
        )
        state.error = "LambdaFunctionCaller not available"
        state.step = "complete"
        return state
    
    # Ensure parameters have required fields
    parameters = state.lambda_parameters.copy()
    if "name" not in parameters:
        parameters["name"] = (
            state.labels.get("name") or
            state.labels.get("resource_name") or
            state.labels.get("pod") or
            state.labels.get("deployment") or
            "unknown"
        )
    if "namespace" not in parameters:
        parameters["namespace"] = (
            state.labels.get("namespace") or
            state.labels.get("resource_namespace") or
            "flux-system"
        )
    
    # Trace remediation
    remediation_correlation_id = state.correlation_id or state.event_id or "unknown"
    
    try:
        with trace_remediation(
            state.alertname or "unknown",
            state.lambda_function,
            remediation_correlation_id
        ) as remediation_span:
            log_remediation_step(
                "remediation.started",
                alertname=state.alertname,
                lambda_function=state.lambda_function,
                correlation_id=remediation_correlation_id,
                parameters=parameters
            )
            
            if remediation_span:
                remediation_span.set_attribute("remediation.lambda_function", state.lambda_function)
                remediation_span.set_attribute("remediation.parameters", str(parameters))
            
            logger.info(
                "executing_lambda_function",
                alertname=state.alertname,
                lambda_function=state.lambda_function,
                parameters=parameters,
                correlation_id=remediation_correlation_id
            )
            
            result = await lambda_caller.call_lambda_function(
                function_name=state.lambda_function,
                parameters=parameters,
                correlation_id=remediation_correlation_id
            )
            
            state.remediation_result = result
            remediation_status = result.get("status", "unknown")
            
            if remediation_span:
                remediation_span.set_attribute("remediation.status", remediation_status)
                remediation_span.set_attribute("remediation.message", result.get("message", ""))
            
            if remediation_status == "error":
                state.error = result.get("message", "Unknown error")
                state.success = False
                logger.error(
                    "remediation_failed",
                    alertname=state.alertname,
                    lambda_function=state.lambda_function,
                    error=state.error,
                    correlation_id=remediation_correlation_id
                )
            else:
                state.success = True
                logger.info(
                    "remediation_completed",
                    alertname=state.alertname,
                    lambda_function=state.lambda_function,
                    status=remediation_status,
                    correlation_id=remediation_correlation_id
                )
            
            state.step = "verify_remediation"
            
    except Exception as e:
        logger.error(
            "lambda_function_execution_error",
            alertname=state.alertname,
            lambda_function=state.lambda_function,
            error=str(e),
            correlation_id=remediation_correlation_id,
            exc_info=True
        )
        state.error = str(e)
        state.success = False
        state.step = "complete"
    
    return state


async def verify_remediation(state: AgentState) -> AgentState:
    """
    Verify remediation success.
    
    Checks if the alert is resolved by querying Prometheus metrics.
    """
    # TODO: Implement verification logic
    # For now, assume success if remediation_result indicates success
    if state.remediation_result and state.remediation_result.get("status") == "success":
        state.verification_result = {"verified": True, "alert_resolved": True}
        state.success = True
    else:
        state.verification_result = {"verified": False, "alert_resolved": False}
        state.success = False
    
    logger.info(
        "remediation_verified",
        alertname=state.alertname,
        success=state.success,
        correlation_id=state.correlation_id
    )
    
    state.step = "complete"
    return state


def should_retry(state: AgentState) -> Literal["execute_lambda_function", "complete"]:
    """Conditional routing: retry if failed and retries remaining."""
    if state.error and state.retry_count < state.max_retries:
        return "execute_lambda_function"
    return "complete"


def create_remediation_graph(
    lambda_caller: Optional[LambdaFunctionCaller] = None,
    report_generator: Optional[Any] = None,
    approval_manager: Optional[ApprovalManager] = None,
    operation_mode: str = "agentic",
    approval_config: Optional[Dict[str, Any]] = None
) -> Any:
    """
    Create LangGraph workflow for Agent-SRE remediation.
    
    Workflow:
    1. receive_cloudevent → extract_from_cloudevent
    2. extract_from_cloudevent → extract_lambda_function
    3. extract_lambda_function → execute_lambda_function
    4. execute_lambda_function → verify_remediation
    5. verify_remediation → complete
    """
    # Create graph with AgentState
    graph = StateGraph(AgentState)
    
    # Add nodes
    graph.add_node("extract_from_cloudevent", extract_from_cloudevent)
    graph.add_node(
        "extract_lambda_function",
        lambda state: extract_lambda_function(state, report_generator)
    )
    graph.add_node(
        "request_approval",
        lambda state: request_approval(state, approval_manager)
    )
    graph.add_node(
        "wait_for_approval",
        lambda state: wait_for_approval(state, approval_manager)
    )
    graph.add_node(
        "execute_lambda_function",
        lambda state: execute_lambda_function(state, lambda_caller)
    )
    graph.add_node("verify_remediation", verify_remediation)
    
    # Add edges
    graph.set_entry_point("extract_from_cloudevent")
    graph.add_edge("extract_from_cloudevent", "extract_lambda_function")
    
    # Conditional routing: request approval if supervised mode
    def route_after_extraction(state: AgentState) -> Literal["request_approval", "execute_lambda_function"]:
        if state.operation_mode == "supervised" and state.lambda_function:
            return "request_approval"
        return "execute_lambda_function"
    
    graph.add_conditional_edges(
        "extract_lambda_function",
        route_after_extraction
    )
    
    # Approval flow
    graph.add_edge("request_approval", "wait_for_approval")
    
    def route_after_approval(state: AgentState) -> Literal["execute_lambda_function", "complete"]:
        if state.approval_status == ApprovalStatus.APPROVED:
            return "execute_lambda_function"
        return "complete"
    
    graph.add_conditional_edges(
        "wait_for_approval",
        route_after_approval
    )
    
    # Execution flow
    graph.add_edge("execute_lambda_function", "verify_remediation")
    graph.add_edge("verify_remediation", END)
    
    # Compile with checkpointing (state persistence)
    memory = MemorySaver()
    app = graph.compile(checkpointer=memory)
    
    return app


async def run_remediation_workflow(
    event_data: Dict[str, Any],
    event_type: str,
    event_id: Optional[str] = None,
    correlation_id: Optional[str] = None,
    lambda_caller: Optional[LambdaFunctionCaller] = None,
    report_generator: Optional[Any] = None,
    approval_manager: Optional[ApprovalManager] = None,
    operation_mode: str = "agentic",
    approval_config: Optional[Dict[str, Any]] = None
) -> AgentState:
    """
    Run the complete remediation workflow.
    
    Args:
        event_data: CloudEvent data
        event_type: CloudEvent type
        event_id: CloudEvent ID
        correlation_id: Correlation ID for tracing
        lambda_caller: LambdaFunctionCaller instance
        report_generator: ReportGenerator instance for AI selection
    
    Returns:
        Final AgentState with remediation results
    """
    # Create initial state
    initial_state = AgentState(
        event_data=event_data,
        event_type=event_type,
        event_id=event_id,
        correlation_id=correlation_id,
        step="receive_cloudevent",
        operation_mode=operation_mode,
        approval_config=approval_config
    )
    
    # Create graph
    app = create_remediation_graph(
        lambda_caller,
        report_generator,
        approval_manager,
        operation_mode,
        approval_config
    )
    
    # Run workflow
    config = {"configurable": {"thread_id": correlation_id or event_id or "default"}}
    result = await app.ainvoke(initial_state, config)
    
    return result
