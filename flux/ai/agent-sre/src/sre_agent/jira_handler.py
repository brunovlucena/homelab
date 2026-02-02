"""
Jira Handler for Agent-SRE

Handles creating and updating Jira issues for Prometheus alerts and remediation events.
"""
import os
import sys
from typing import Dict, Any, Optional
import structlog

# Add parent directory (src) to path for imports
# When running as 'python -m src.sre_agent.main', we need to add src to path
parent_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if parent_dir not in sys.path:
    sys.path.insert(0, parent_dir)

try:
    from jira_client import JiraClient, create_agent_issue
except ImportError as e:
    # Fallback if jira_client is not available
    JiraClient = None
    create_agent_issue = None
    # Log warning but don't fail - Jira is optional
    import logging
    logging.warning(f"jira_client not available: {e}")

logger = structlog.get_logger()


class JiraHandler:
    """Handle Jira operations for agent-sre."""
    
    def __init__(self, project_key: Optional[str] = None):
        """
        Initialize Jira handler.
        
        Args:
            project_key: Jira project key (e.g., "PROJ"). If not provided, reads from JIRA_PROJECT_KEY env var.
        """
        self.project_key = project_key or os.getenv("JIRA_PROJECT_KEY", "HOMELAB")
        self.client: Optional[JiraClient] = None
        
        # Initialize client if credentials are available
        try:
            self.client = JiraClient()
            logger.info("jira_handler_initialized", project_key=self.project_key)
        except ValueError as e:
            logger.warning("jira_client_not_available", error=str(e))
            self.client = None
    
    async def create_alert_ticket(
        self,
        alert: Dict[str, Any],
        correlation_id: Optional[str] = None
    ) -> Optional[str]:
        """
        Create a Jira ticket for a Prometheus alert.
        
        Args:
            alert: Alert data dictionary with labels, annotations, etc.
            correlation_id: Correlation ID for tracing
            
        Returns:
            Issue key if created, None otherwise
        """
        if not self.client:
            logger.debug("jira_client_not_available_skipping_ticket_creation")
            return None
        
        try:
            labels = alert.get("labels", {})
            annotations = alert.get("annotations", {})
            common_annotations = alert.get("commonAnnotations", {})
            
            # Merge annotations (alert-specific take precedence)
            all_annotations = {**common_annotations, **annotations}
            
            alertname = labels.get("alertname", "unknown")
            severity = labels.get("severity", "warning")
            status = alert.get("status", "firing")
            
            # Map severity to Jira priority
            priority_map = {
                "critical": "Highest",
                "warning": "High",
                "info": "Medium",
            }
            priority = priority_map.get(severity.lower(), "Medium")
            
            # Map alert status to issue type
            issue_type = "Bug" if status == "firing" else "Task"
            
            # Build summary
            summary = f"[Alert] {alertname}"
            if status == "resolved":
                summary = f"[Resolved] {alertname}"
            
            # Build description
            description_parts = [
                f"*Status*: {status}",
                f"*Severity*: {severity}",
                f"*Alert Name*: `{alertname}`",
            ]
            
            # Add alert description if available
            alert_description = all_annotations.get("description") or all_annotations.get("summary")
            if alert_description:
                description_parts.append(f"\n*Description*:\n{alert_description}")
            
            # Add labels
            if labels:
                label_items = "\n".join([f"* `{k}`: `{v}`" for k, v in labels.items() if k != "alertname"])
                if label_items:
                    description_parts.append(f"\n*Labels*:\n{label_items}")
            
            # Add correlation ID for tracing
            if correlation_id:
                description_parts.append(f"\n*Correlation ID*: `{correlation_id}`")
            
            description_parts.append("\n---\n*Created by agent-sre*")
            
            description = "\n".join(description_parts)
            
            # Create issue
            issue = await self.client.create_issue(
                project_key=self.project_key,
                summary=summary,
                description=description,
                issue_type=issue_type,
                priority=priority
            )
            
            issue_key = issue.get("key")
            
            logger.info(
                "jira_ticket_created",
                alertname=alertname,
                issue_key=issue_key,
                priority=priority,
                correlation_id=correlation_id
            )
            
            return issue_key
            
        except Exception as e:
            logger.error(
                "failed_to_create_jira_ticket",
                alertname=labels.get("alertname", "unknown"),
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            return None
    
    async def create_remediation_failure_ticket(
        self,
        alertname: str,
        lambda_function: str,
        error_message: str,
        parameters: Dict[str, Any],
        correlation_id: Optional[str] = None
    ) -> Optional[str]:
        """
        Create a Jira ticket when remediation fails.
        
        Args:
            alertname: Name of the alert
            lambda_function: Name of the LambdaFunction that failed
            error_message: Error message from remediation
            parameters: Parameters used for remediation
            correlation_id: Correlation ID for tracing
            
        Returns:
            Issue key if created, None otherwise
        """
        if not self.client:
            logger.debug("jira_client_not_available_skipping_failure_ticket")
            return None
        
        try:
            summary = f"[Remediation Failed] {alertname}"
            
            description_parts = [
                f"*Alert*: `{alertname}`",
                f"*Remediation Function*: `{lambda_function}`",
                f"*Status*: ❌ Failed",
                f"\n*Error*:\n{{code}}\n{error_message}\n{{code}}",
            ]
            
            if parameters:
                param_items = "\n".join([f"* `{k}`: `{v}`" for k, v in parameters.items()])
                description_parts.append(f"\n*Parameters*:\n{param_items}")
            
            if correlation_id:
                description_parts.append(f"\n*Correlation ID*: `{correlation_id}`")
            
            description_parts.append("\n---\n*Created by agent-sre*")
            
            description = "\n".join(description_parts)
            
            # Create issue with highest priority (remediation failures are urgent)
            issue = await self.client.create_issue(
                project_key=self.project_key,
                summary=summary,
                description=description,
                issue_type="Bug",
                priority="Highest"
            )
            
            issue_key = issue.get("key")
            
            logger.info(
                "jira_remediation_failure_ticket_created",
                alertname=alertname,
                lambda_function=lambda_function,
                issue_key=issue_key,
                correlation_id=correlation_id
            )
            
            return issue_key
            
        except Exception as e:
            logger.error(
                "failed_to_create_remediation_failure_ticket",
                alertname=alertname,
                lambda_function=lambda_function,
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            return None
    
    async def update_issue_with_remediation_status(
        self,
        issue_key: str,
        remediation_status: str,
        message: Optional[str] = None,
        correlation_id: Optional[str] = None
    ) -> bool:
        """
        Update an existing Jira issue with remediation status.
        
        Args:
            issue_key: Jira issue key (e.g., "PROJ-123")
            remediation_status: Status of remediation (e.g., "success", "failed")
            message: Optional message about the remediation
            correlation_id: Correlation ID for tracing
            
        Returns:
            True if updated, False otherwise
        """
        if not self.client:
            return False
        
        try:
            comment_body = f"*Remediation Status*: {remediation_status}"
            if message:
                comment_body += f"\n\n{message}"
            if correlation_id:
                comment_body += f"\n\n*Correlation ID*: `{correlation_id}`"
            
            await self.client.add_comment(issue_key=issue_key, body=comment_body)
            
            logger.info(
                "jira_issue_updated",
                issue_key=issue_key,
                remediation_status=remediation_status,
                correlation_id=correlation_id
            )
            
            return True
            
        except Exception as e:
            logger.error(
                "failed_to_update_jira_issue",
                issue_key=issue_key,
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            return False

