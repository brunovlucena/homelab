"""
AI-Powered Remediation Selection - Phase 1: Function Calling with FunctionGemma

Uses FunctionGemma 270M's native function calling capabilities to intelligently
select Lambda functions and parameters for Prometheus alerts.
"""
from typing import Dict, Any, Optional, List
import json
import structlog

from src.report_generator import ReportGenerator

logger = structlog.get_logger()


# Lambda function schema for FunctionGemma function calling
LAMBDA_FUNCTIONS_SCHEMA = {
    "type": "function",
    "function": {
        "name": "select_remediation_action",
        "description": "Select the appropriate Lambda function and parameters for a Prometheus alert",
        "parameters": {
            "type": "object",
            "properties": {
                "lambda_function": {
                    "type": "string",
                    "enum": [
                        "flux-reconcile-kustomization",
                        "flux-reconcile-gitrepository",
                        "flux-reconcile-helmrelease",
                        "pod-restart",
                        "pod-check-status",
                        "scale-deployment",
                        "check-pvc-status"
                    ],
                    "description": "The Lambda function to execute for remediation"
                },
                "parameters": {
                    "type": "object",
                    "properties": {
                        "name": {
                            "type": "string",
                            "description": "Resource name (pod, deployment, kustomization, etc.)"
                        },
                        "namespace": {
                            "type": "string",
                            "description": "Kubernetes namespace"
                        },
                        "type": {
                            "type": "string",
                            "enum": ["pod", "deployment"],
                            "description": "Resource type (for pod-restart)"
                        },
                        "replicas": {
                            "type": "integer",
                            "description": "Target replica count (for scale-deployment)"
                        }
                    },
                    "required": ["name", "namespace"]
                },
                "reasoning": {
                    "type": "string",
                    "description": "Explanation of why this Lambda function was selected"
                }
            },
            "required": ["lambda_function", "parameters", "reasoning"]
        }
    }
}


def create_remediation_prompt(alert_data: Dict[str, Any]) -> str:
    """Create prompt for Lambda function selection."""
    alertname = alert_data.get("labels", {}).get("alertname", "unknown")
    labels = alert_data.get("labels", {})
    annotations = alert_data.get("annotations", {})
    
    prompt = f"""You are an SRE agent analyzing a Prometheus alert to select the appropriate remediation action.

Alert Details:
- Alert Name: {alertname}
- Labels: {json.dumps(labels, indent=2)}
- Annotations: {json.dumps(annotations, indent=2)}

Available Lambda Functions:
1. flux-reconcile-kustomization: Reconcile Flux Kustomization
   Parameters: name (required), namespace (default: flux-system)
   Use when: Kustomization is out of sync or failing

2. flux-reconcile-gitrepository: Reconcile Flux GitRepository
   Parameters: name (required), namespace (default: flux-system)
   Use when: GitRepository sync is failing

3. flux-reconcile-helmrelease: Reconcile Flux HelmRelease
   Parameters: name (required), namespace (default: flux-system)
   Use when: HelmRelease reconciliation is failing

4. pod-restart: Restart a pod or deployment
   Parameters: name (required), namespace (required), type (pod|deployment, default: pod)
   Use when: Pod is crashing, stuck, or needs restart

5. pod-check-status: Check pod status
   Parameters: name (required), namespace (required)
   Use when: Need to verify pod health before remediation

6. scale-deployment: Scale deployment to specific replicas
   Parameters: name (required), namespace (required), replicas (required)
   Use when: Need to scale up/down a deployment

7. check-pvc-status: Check PVC status and usage
   Parameters: name (required), namespace (required)
   Use when: Storage issues suspected

Examples:

Example 1:
Alert: FluxReconciliationFailure
Labels: {{"name": "homepage", "namespace": "flux-system", "kind": "Kustomization"}}
Selection: flux-reconcile-kustomization
Parameters: {{"name": "homepage", "namespace": "flux-system"}}
Reasoning: Alert indicates Kustomization reconciliation failure, so reconcile it.

Example 2:
Alert: PodCrashLooping
Labels: {{"pod": "app-abc123", "namespace": "production"}}
Selection: pod-restart
Parameters: {{"name": "app-abc123", "namespace": "production", "type": "pod"}}
Reasoning: Pod is in crash loop, restarting it may resolve transient issues.

Example 3:
Alert: DeploymentReplicasMismatch
Labels: {{"deployment": "api-server", "namespace": "production", "expected": "3", "actual": "1"}}
Selection: scale-deployment
Parameters: {{"name": "api-server", "namespace": "production", "replicas": 3}}
Reasoning: Deployment has fewer replicas than expected, scale to match expected count.

Now analyze the current alert and select the appropriate Lambda function with parameters.
Return your response as a JSON object with keys: lambda_function, parameters, reasoning.
"""
    return prompt


async def select_remediation_with_ai(
    alert_data: Dict[str, Any],
    report_generator: ReportGenerator
) -> Dict[str, Any]:
    """
    Use AI to select Lambda function and parameters.
    
    Args:
        alert_data: Prometheus alert data (labels, annotations, etc.)
        report_generator: ReportGenerator instance for AI inference
        
    Returns:
        Dict with lambda_function, parameters, reasoning, and confidence
    """
    try:
        prompt = create_remediation_prompt(alert_data)
        
        # Use FunctionGemma's function calling via structured output
        # For now, we'll use structured prompt and parse JSON response
        # TODO: Integrate native function calling when Ollama/MLX supports it
        
        model = await report_generator._get_model()
        
        # Generate response with structured output instruction
        enhanced_prompt = f"""{prompt}

IMPORTANT: Respond with ONLY a valid JSON object in this exact format:
{{
  "lambda_function": "<function-name>",
  "parameters": {{
    "name": "<resource-name>",
    "namespace": "<namespace>",
    ...
  }},
  "reasoning": "<explanation>"
}}
"""
        
        response_data = await report_generator._generate_with_model(model, enhanced_prompt)
        
        # Parse response - handle both direct JSON and text with JSON
        if isinstance(response_data, dict):
            result = response_data
        else:
            # Try to extract JSON from text response
            try:
                result = report_generator._parse_json_response(str(response_data))
            except Exception as parse_error:
                logger.warning(
                    "failed_to_parse_ai_response",
                    error=str(parse_error),
                    response_preview=str(response_data)[:200]
                )
                # Try to extract lambda_function from text if JSON parsing fails
                response_text = str(response_data).lower()
                result = {}
                # Look for common patterns
                for func_name in ["flux-reconcile-kustomization", "flux-reconcile-gitrepository", 
                                 "flux-reconcile-helmrelease", "pod-restart", "pod-check-status",
                                 "scale-deployment", "check-pvc-status"]:
                    if func_name.lower() in response_text:
                        result["lambda_function"] = func_name
                        break
        
        # Validate and enrich result
        if "lambda_function" not in result or not result.get("lambda_function"):
            logger.error(
                "ai_response_missing_lambda_function",
                response_data=str(response_data)[:500],
                parsed_result=result
            )
            raise ValueError("AI response missing lambda_function")
        if "parameters" not in result:
            result["parameters"] = {}
        
        # Ensure required parameters
        params = result.get("parameters", {})
        labels = alert_data.get("labels", {})
        
        # Extract name from labels if not provided
        if "name" not in params:
            params["name"] = (
                labels.get("name") or
                labels.get("resource_name") or
                labels.get("pod") or
                labels.get("deployment") or
                labels.get("kustomization")
            )
        
        # Extract namespace from labels if not provided
        if "namespace" not in params:
            params["namespace"] = (
                labels.get("namespace") or
                labels.get("resource_namespace") or
                "flux-system"  # Default for Flux resources
            )
        
        result["parameters"] = params
        result["method"] = "ai_function_calling"
        result["confidence"] = 0.8  # Default confidence, can be improved with validation
        
        logger.info(
            "ai_remediation_selected",
            lambda_function=result["lambda_function"],
            parameters=params,
            reasoning=result.get("reasoning", ""),
            confidence=result["confidence"]
        )
        
        return result
        
    except Exception as e:
        logger.error(
            "ai_remediation_selection_failed",
            error=str(e),
            alertname=alert_data.get("labels", {}).get("alertname", "unknown"),
            exc_info=True
        )
        raise

