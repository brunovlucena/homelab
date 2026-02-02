"""
Linear Handler for Agent-SRE

Handles creating and updating Linear issues for Prometheus alerts and remediation events.
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
    from linear_client import LinearClient, create_agent_issue
except ImportError as e:
    # Fallback if linear_client is not available
    LinearClient = None
    create_agent_issue = None
    # Log warning but don't fail - Linear is optional
    import logging
    logging.warning(f"linear_client not available: {e}")

logger = structlog.get_logger()


class LinearHandler:
    """Handle Linear operations for agent-sre."""
    
    def __init__(self, team_key: Optional[str] = None):
        """
        Initialize Linear handler.
        
        Args:
            team_key: Linear team key (e.g., "BVL"). If not provided, reads from LINEAR_TEAM_KEY env var.
        """
        self.team_key = team_key or os.getenv("LINEAR_TEAM_KEY", "BVL")
        self.client: Optional[LinearClient] = None
        self._team_id: Optional[str] = None
        
        # Initialize client if API key is available
        try:
            self.client = LinearClient()
            logger.info("linear_handler_initialized", team_key=self.team_key)
        except ValueError as e:
            logger.warning("linear_client_not_available", error=str(e))
            self.client = None
    
    async def _ensure_team_id(self) -> Optional[str]:
        """Get team ID, caching it for subsequent calls."""
        if not self.client:
            return None
        
        if self._team_id:
            return self._team_id
        
        try:
            team = await self.client.get_team(self.team_key)
            self._team_id = team["id"]
            return self._team_id
        except Exception as e:
            logger.error("failed_to_get_team", team_key=self.team_key, error=str(e))
            return None
    
    async def create_alert_ticket(
        self,
        alert: Dict[str, Any],
        correlation_id: Optional[str] = None
    ) -> Optional[str]:
        """
        Create a Linear ticket for a Prometheus alert.
        
        Args:
            alert: Alert data dictionary with labels, annotations, etc.
            correlation_id: Correlation ID for tracing
            
        Returns:
            Issue URL if created, None otherwise
        """
        if not self.client:
            logger.debug("linear_client_not_available_skipping_ticket_creation")
            return None
        
        try:
            team_id = await self._ensure_team_id()
            if not team_id:
                logger.warning("cannot_create_ticket_no_team_id", team_key=self.team_key)
                return None
            
            labels = alert.get("labels", {})
            annotations = alert.get("annotations", {})
            common_annotations = alert.get("commonAnnotations", {})
            
            # Merge annotations (alert-specific take precedence)
            all_annotations = {**common_annotations, **annotations}
            
            alertname = labels.get("alertname", "unknown")
            severity = labels.get("severity", "warning")
            status = alert.get("status", "firing")
            
            # Map severity to priority
            # 0=None, 1=Urgent, 2=High, 3=Normal, 4=Low
            priority_map = {
                "critical": 1,  # Urgent
                "warning": 2,   # High
                "info": 3,      # Normal
            }
            priority = priority_map.get(severity.lower(), 3)
            
            # Build title
            title = f"[Alert] {alertname}"
            if status == "resolved":
                title = f"[Resolved] {alertname}"
            
            # Build description
            description_parts = [
                f"**Status**: {status}",
                f"**Severity**: {severity}",
                f"**Alert Name**: `{alertname}`",
            ]
            
            # Add alert description if available
            alert_description = all_annotations.get("description") or all_annotations.get("summary")
            if alert_description:
                description_parts.append(f"\n**Description**:\n{alert_description}")
            
            # Add labels
            if labels:
                label_items = "\n".join([f"- `{k}`: `{v}`" for k, v in labels.items() if k != "alertname"])
                if label_items:
                    description_parts.append(f"\n**Labels**:\n{label_items}")
            
            # Add correlation ID for tracing
            if correlation_id:
                description_parts.append(f"\n**Correlation ID**: `{correlation_id}`")
            
            description_parts.append("\n---\n*Created by agent-sre*")
            
            description = "\n".join(description_parts)
            
            # Create issue
            issue = await self.client.create_issue(
                title=title,
                description=description,
                team_id=team_id,
                priority=priority
            )
            
            logger.info(
                "linear_ticket_created",
                alertname=alertname,
                issue_id=issue.get("identifier"),
                issue_url=issue.get("url"),
                priority=priority,
                correlation_id=correlation_id
            )
            
            return issue.get("url")
            
        except Exception as e:
            logger.error(
                "failed_to_create_linear_ticket",
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
        Create a Linear ticket when remediation fails.
        
        Args:
            alertname: Name of the alert
            lambda_function: Name of the LambdaFunction that failed
            error_message: Error message from remediation
            parameters: Parameters used for remediation
            correlation_id: Correlation ID for tracing
            
        Returns:
            Issue URL if created, None otherwise
        """
        if not self.client:
            logger.debug("linear_client_not_available_skipping_failure_ticket")
            return None
        
        try:
            team_id = await self._ensure_team_id()
            if not team_id:
                return None
            
            title = f"[Remediation Failed] {alertname}"
            
            description_parts = [
                f"**Alert**: `{alertname}`",
                f"**Remediation Function**: `{lambda_function}`",
                f"**Status**: ❌ Failed",
                f"\n**Error**:\n```\n{error_message}\n```",
            ]
            
            if parameters:
                param_items = "\n".join([f"- `{k}`: `{v}`" for k, v in parameters.items()])
                description_parts.append(f"\n**Parameters**:\n{param_items}")
            
            if correlation_id:
                description_parts.append(f"\n**Correlation ID**: `{correlation_id}`")
            
            description_parts.append("\n---\n*Created by agent-sre*")
            
            description = "\n".join(description_parts)
            
            # Create issue with high priority (remediation failures are urgent)
            issue = await self.client.create_issue(
                title=title,
                description=description,
                team_id=team_id,
                priority=1  # Urgent
            )
            
            logger.info(
                "linear_remediation_failure_ticket_created",
                alertname=alertname,
                lambda_function=lambda_function,
                issue_id=issue.get("identifier"),
                issue_url=issue.get("url"),
                correlation_id=correlation_id
            )
            
            return issue.get("url")
            
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
        issue_url: str,
        remediation_status: str,
        message: Optional[str] = None,
        correlation_id: Optional[str] = None
    ) -> bool:
        """
        Update an existing Linear issue with remediation status.
        
        Args:
            issue_url: URL of the Linear issue
            remediation_status: Status of remediation (e.g., "success", "failed")
            message: Optional message about the remediation
            correlation_id: Correlation ID for tracing
            
        Returns:
            True if updated, False otherwise
        """
        if not self.client:
            return False
        
        try:
            # Extract issue ID from URL (format: https://linear.app/workspace/issue/ISSUE-ID)
            # Or use identifier like "BVL-16"
            issue_id = issue_url.split("/")[-1]
            
            comment_body = f"**Remediation Status**: {remediation_status}"
            if message:
                comment_body += f"\n\n{message}"
            if correlation_id:
                comment_body += f"\n\n**Correlation ID**: `{correlation_id}`"
            
            await self.client.create_comment(issue_id=issue_id, body=comment_body)
            
            logger.info(
                "linear_issue_updated",
                issue_url=issue_url,
                remediation_status=remediation_status,
                correlation_id=correlation_id
            )
            
            return True
            
        except Exception as e:
            logger.error(
                "failed_to_update_linear_issue",
                issue_url=issue_url,
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            return False

