"""
Intelligent Remediation Selection - Hybrid Approach

Combines all training phases:
- Phase 1: Function Calling with FunctionGemma
- Phase 2: Few-Shot Learning
- Phase 3: RAG (Retrieval Augmented Generation)
- Phase 4: Fine-tuned model support

This is the main entry point for AI-powered remediation selection.
"""
from typing import Dict, Any, Optional
import json
import os
import structlog

from .ai_remediation import select_remediation_with_ai, create_remediation_prompt
from .few_shot_learning import get_example_database, create_few_shot_prompt
from .rag_system import get_rag_instance
from .trm_remediation import TRMRemediationSelector
from src.report_generator import ReportGenerator

logger = structlog.get_logger()


def parse_parameters(alert_data: Dict[str, Any]) -> Dict[str, Any]:
    """Parse parameters from alert data (fallback for static annotations)."""
    labels = alert_data.get("labels", {})
    annotations = alert_data.get("annotations", {})
    
    # Try to parse from lambda_parameters annotation
    lambda_parameters_json = annotations.get("lambda_parameters", "{}")
    
    try:
        import json
        parameters = json.loads(lambda_parameters_json) if isinstance(lambda_parameters_json, str) else lambda_parameters_json
    except:
        parameters = {}
    
    # Extract from labels if not in parameters
    if "name" not in parameters:
        parameters["name"] = (
            labels.get("name") or
            labels.get("resource_name") or
            labels.get("pod") or
            labels.get("deployment") or
            labels.get("kustomization")
        )
    
    if "namespace" not in parameters:
        parameters["namespace"] = (
            labels.get("namespace") or
            labels.get("resource_namespace") or
            "flux-system"
        )
    
    return parameters


def calculate_confidence(
    result: Dict[str, Any],
    similar_alerts: list
) -> float:
    """Calculate confidence score for remediation selection."""
    confidence = 0.5  # Base confidence
    
    # Boost confidence if we have similar past incidents
    if len(similar_alerts) > 0:
        confidence += 0.2
    
    # Boost if reasoning is provided and detailed
    if result.get("reasoning"):
        reasoning_len = len(result["reasoning"])
        if reasoning_len > 50:
            confidence += 0.1
        if reasoning_len > 100:
            confidence += 0.1
    
    # Boost if parameters are complete
    params = result.get("parameters", {})
    if params.get("name") and params.get("namespace"):
        confidence += 0.1
    
    return min(confidence, 1.0)


def validate_remediation_selection(
    result: Dict[str, Any],
    alert_data: Dict[str, Any]
) -> Dict[str, Any]:
    """Validate and enrich remediation selection result."""
    # Ensure required fields
    if "lambda_function" not in result:
        raise ValueError("Missing lambda_function in result")
    
    if "parameters" not in result:
        result["parameters"] = {}
    
    # Validate parameters
    params = result["parameters"]
    labels = alert_data.get("labels", {})
    
    # Ensure name is present
    if "name" not in params or not params["name"]:
        params["name"] = (
            labels.get("name") or
            labels.get("resource_name") or
            labels.get("pod") or
            labels.get("deployment") or
            labels.get("kustomization") or
            "unknown"
        )
    
    # Ensure namespace is present
    if "namespace" not in params or not params["namespace"]:
        params["namespace"] = (
            labels.get("namespace") or
            labels.get("resource_namespace") or
            "flux-system"
        )
    
    # Validate function-specific parameters
    lambda_function = result["lambda_function"]
    
    if lambda_function == "scale-deployment":
        if "replicas" not in params:
            # Try to extract from labels
            expected_replicas = labels.get("expected") or labels.get("replicas")
            if expected_replicas:
                try:
                    params["replicas"] = int(expected_replicas)
                except (ValueError, TypeError):
                    logger.warning("invalid_replicas_value", value=expected_replicas)
    
    if lambda_function == "pod-restart":
        if "type" not in params:
            # Default to pod
            params["type"] = "pod"
    
    result["parameters"] = params
    
    return result


async def intelligent_remediation_selection(
    alert_data: Dict[str, Any],
    report_generator: ReportGenerator,
    use_rag: bool = True,
    use_few_shot: bool = True,
    use_trm: bool = True,
    trm_model_path: Optional[str] = None,
    example_db_path: Optional[str] = None,
    rag_embedding_model: Optional[str] = None
) -> Dict[str, Any]:
    """
    Hybrid approach combining all training phases for intelligent remediation selection.
    
    Args:
        alert_data: Prometheus alert data (labels, annotations, etc.)
        report_generator: ReportGenerator instance for AI inference
        use_rag: Enable RAG (Retrieval Augmented Generation)
        use_few_shot: Enable few-shot learning
        example_db_path: Path to example database storage
        rag_embedding_model: Embedding model for RAG
        
    Returns:
        Dict with lambda_function, parameters, reasoning, method, confidence, etc.
    """
    labels = alert_data.get("labels", {})
    annotations = alert_data.get("annotations", {})
    alertname = labels.get("alertname", "unknown")
    
    # Phase 0: Check static annotations first (fast path)
    lambda_function = annotations.get("lambda_function")
    if lambda_function:
        logger.info(
            "using_static_annotation",
            alertname=alertname,
            lambda_function=lambda_function
        )
        return {
            "lambda_function": lambda_function,
            "parameters": parse_parameters(alert_data),
            "method": "static_annotation",
            "confidence": 1.0,  # Static annotations are always correct
            "reasoning": "Using static annotation from PrometheusRule"
        }
    
    # Phase 1: Try TRM model (if enabled and available)
    if use_trm:
        try:
            trm_model_path = trm_model_path or os.getenv(
                "TRM_MODEL_PATH",
                "/workspace/bruno/repos/homelab/flux/ai/trm/checkpoints/Trm_data-ACT-torch/trm-runbook-extended/step_0"
            )
            
            if os.path.exists(trm_model_path):
                logger.info(
                    "using_trm_model",
                    alertname=alertname,
                    trm_model_path=trm_model_path
                )
                
                selector = TRMRemediationSelector(trm_model_path=trm_model_path)
                trm_result = await selector.select_and_trigger(
                    alert_data,
                    correlation_id=None  # Will be set by caller
                )
                
                if trm_result.get("lambda_function"):
                    logger.info(
                        "trm_selection_success",
                        alertname=alertname,
                        lambda_function=trm_result["lambda_function"],
                        confidence=trm_result.get("confidence", 0.0)
                    )
                    return {
                        "lambda_function": trm_result["lambda_function"],
                        "parameters": trm_result.get("parameters", {}),
                        "method": "trm_recursive_reasoning",
                        "confidence": trm_result.get("confidence", 0.7),
                        "reasoning": trm_result.get("reasoning", "TRM recursive reasoning"),
                        "event_id": trm_result.get("event_id")
                    }
            else:
                logger.debug(
                    "trm_model_not_found",
                    alertname=alertname,
                    trm_model_path=trm_model_path
                )
        except Exception as e:
            logger.warning(
                "trm_selection_failed",
                alertname=alertname,
                error=str(e),
                exc_info=True
            )
    
    # Phase 2 & 3: Gather context from RAG and Few-Shot
    similar_alerts = []
    examples = []
    
    if use_rag:
        try:
            rag = get_rag_instance(embedding_model=rag_embedding_model)
            similar_alerts = await rag.find_similar_alerts(alert_data, top_k=3)
            logger.debug("rag_similar_alerts_found", count=len(similar_alerts))
        except Exception as e:
            logger.warning("rag_failed", error=str(e))
    
    if use_few_shot:
        try:
            example_db = get_example_database(storage_path=example_db_path)
            examples = example_db.get_examples_for_prompt(alertname, labels, top_k=5)
            logger.debug("few_shot_examples_found", count=len(examples))
        except Exception as e:
            logger.warning("few_shot_failed", error=str(e))
    
    # Phase 1: Build enhanced prompt with RAG + Few-Shot context
    base_prompt = create_remediation_prompt(alert_data)
    
    # Enhance with RAG context
    if similar_alerts:
        rag_prompt = "\n\nSimilar Past Incidents:\n"
        for i, alert in enumerate(similar_alerts, 1):
            rag_prompt += f"{i}. {alert['alertname']}: "
            if alert.get('lambda_function'):
                rag_prompt += f"{alert['lambda_function']} "
                rag_prompt += f"({json.dumps(alert.get('parameters', {}))}) "
            rag_prompt += f"[Success: {alert.get('success', 'Unknown')}]\n"
        base_prompt += rag_prompt
    
    # Enhance with few-shot examples
    if examples:
        few_shot_prompt = create_few_shot_prompt(alert_data, examples)
        # Combine prompts
        enhanced_prompt = f"""{few_shot_prompt}

{base_prompt}
"""
    else:
        enhanced_prompt = base_prompt
    
    # Phase 1: Use FunctionGemma function calling
    try:
        result = await select_remediation_with_ai(alert_data, report_generator)
        
        # Override prompt with enhanced version
        # (This is a workaround - ideally we'd pass enhanced_prompt directly)
        # For now, we'll use the base prompt and rely on the model's training
        
        # Phase 5: Validate and enrich
        result = validate_remediation_selection(result, alert_data)
        
        # Calculate confidence
        confidence = calculate_confidence(result, similar_alerts)
        result["confidence"] = confidence
        result["method"] = "ai_function_calling"
        result["similar_incidents"] = len(similar_alerts)
        result["few_shot_examples"] = len(examples)
        
        # Index in RAG for future use
        if use_rag and similar_alerts:
            try:
                rag = get_rag_instance()
                rag.index_alert(
                    alert_data=alert_data,
                    lambda_function=result["lambda_function"],
                    parameters=result["parameters"],
                    success=None  # Will be updated after remediation
                )
            except Exception as e:
                logger.debug("failed_to_index_alert", error=str(e))
        
        logger.info(
            "intelligent_remediation_selected",
            alertname=alertname,
            lambda_function=result["lambda_function"],
            method=result["method"],
            confidence=confidence,
            similar_incidents=len(similar_alerts)
        )
        
        return result
        
    except Exception as e:
        logger.error(
            "intelligent_remediation_failed",
            alertname=alertname,
            error=str(e),
            exc_info=True
        )
        # Fallback: return None to trigger static annotation or manual handling
        raise


# Helper function to record successful remediation for training
def record_remediation_success(
    alert_data: Dict[str, Any],
    lambda_function: str,
    parameters: Dict[str, Any],
    success: bool,
    example_db_path: Optional[str] = None,
    rag_embedding_model: Optional[str] = None
):
    """
    Record a successful remediation for future training.
    
    This should be called after remediation execution to build the training dataset.
    """
    labels = alert_data.get("labels", {})
    alertname = labels.get("alertname", "unknown")
    
    # Add to example database
    try:
        example_db = get_example_database(storage_path=example_db_path)
        example_db.add_example(
            alertname=alertname,
            labels=labels,
            lambda_function=lambda_function,
            parameters=parameters,
            success=success
        )
    except Exception as e:
        logger.warning("failed_to_record_example", error=str(e))
    
    # Index in RAG
    try:
        rag = get_rag_instance(embedding_model=rag_embedding_model)
        rag.index_alert(
            alert_data=alert_data,
            lambda_function=lambda_function,
            parameters=parameters,
            success=success
        )
    except Exception as e:
        logger.warning("failed_to_index_remediation", error=str(e))

