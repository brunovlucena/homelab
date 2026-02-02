"""
Approval System for Agent-SRE

Handles approval requests for supervised mode operations.
Supports Slack and custom app approval providers.
"""
from typing import Dict, Any, Optional, List, Literal
from enum import Enum
import os
import json
import httpx
import structlog
from datetime import datetime, timedelta
from pydantic import BaseModel, Field

logger = structlog.get_logger()


class ApprovalStatus(str, Enum):
    """Approval request status."""
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    TIMEOUT = "timeout"
    CANCELLED = "cancelled"


class ApprovalProvider(str, Enum):
    """Approval provider types."""
    SLACK = "slack"
    CUSTOM = "custom"


class ApprovalRequest(BaseModel):
    """Approval request model."""
    request_id: str
    agent_name: str
    action: str  # e.g., "execute_lambda_function"
    lambda_function: Optional[str] = None
    parameters: Dict[str, Any] = Field(default_factory=dict)
    alertname: Optional[str] = None
    correlation_id: Optional[str] = None
    providers: List[ApprovalProvider] = Field(default_factory=list)
    require_all: bool = False
    timeout: timedelta = Field(default=timedelta(hours=1))
    timeout_action: Literal["approve", "reject", "pending"] = "pending"
    created_at: datetime = Field(default_factory=datetime.utcnow)
    status: ApprovalStatus = ApprovalStatus.PENDING
    approvals: Dict[str, ApprovalStatus] = Field(default_factory=dict)  # provider -> status
    metadata: Dict[str, Any] = Field(default_factory=dict)


class SlackApprovalClient:
    """Client for sending approval requests to Slack."""
    
    def __init__(
        self,
        webhook_url: Optional[str] = None,
        webhook_url_secret_ref: Optional[Dict[str, str]] = None,
        bot_token: Optional[str] = None,
        channel: Optional[str] = None,
        callback_url: Optional[str] = None
    ):
        self.webhook_url = webhook_url
        self.webhook_url_secret_ref = webhook_url_secret_ref
        self.bot_token = bot_token
        self.channel = channel or "#agent-approvals"
        self.callback_url = callback_url
        
        # Load webhook URL from secret if provided
        if webhook_url_secret_ref and not webhook_url:
            self.webhook_url = self._load_secret(
                webhook_url_secret_ref.get("name"),
                webhook_url_secret_ref.get("key", "webhook_url")
            )
    
    def _load_secret(self, secret_name: str, key: str) -> Optional[str]:
        """Load secret from Kubernetes secret."""
        # TODO: Implement Kubernetes secret loading
        # For now, try environment variable
        env_key = f"{secret_name.upper().replace('-', '_')}_{key.upper()}"
        return os.getenv(env_key)
    
    async def send_approval_request(self, request: ApprovalRequest) -> Dict[str, Any]:
        """
        Send approval request to Slack.
        
        Returns:
            Dict with message_ts, channel, etc.
        """
        if not self.webhook_url:
            raise ValueError("Slack webhook URL not configured")
        
        # Build Slack message with interactive buttons
        blocks = [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": f"🔐 Approval Required: {request.action}"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Agent:* {request.agent_name}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Request ID:* `{request.request_id}`"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Action:* {request.action}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Alert:* {request.alertname or 'N/A'}"
                    }
                ]
            }
        ]
        
        # Add LambdaFunction details if present
        if request.lambda_function:
            blocks.append({
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*LambdaFunction:* `{request.lambda_function}`\n*Parameters:* ```{json.dumps(request.parameters, indent=2)}```"
                }
            })
        
        # Add approval buttons
        blocks.append({
            "type": "actions",
            "elements": [
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "✅ Approve"
                    },
                    "style": "primary",
                    "action_id": f"approve_{request.request_id}",
                    "value": json.dumps({
                        "request_id": request.request_id,
                        "action": "approve"
                    })
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "❌ Reject"
                    },
                    "style": "danger",
                    "action_id": f"reject_{request.request_id}",
                    "value": json.dumps({
                        "request_id": request.request_id,
                        "action": "reject"
                    })
                }
            ]
        })
        
        # Add footer with timeout info
        timeout_str = f"{request.timeout.total_seconds() / 60:.0f} minutes"
        blocks.append({
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": f"⏱️ Request will timeout in {timeout_str}. Default action: {request.timeout_action}"
                }
            ]
        })
        
        payload = {
            "channel": self.channel,
            "blocks": blocks,
            "text": f"Approval required for {request.agent_name}: {request.action}"
        }
        
        async with httpx.AsyncClient() as client:
            response = await client.post(
                self.webhook_url,
                json=payload,
                timeout=10.0
            )
            response.raise_for_status()
            result = response.json()
            
            logger.info(
                "slack_approval_request_sent",
                request_id=request.request_id,
                channel=self.channel,
                message_ts=result.get("ts")
            )
            
            return {
                "provider": ApprovalProvider.SLACK,
                "message_ts": result.get("ts"),
                "channel": self.channel,
                "status": "sent"
            }
    
    async def handle_approval_response(
        self,
        payload: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Handle approval response from Slack.
        
        Expected payload format (from Slack interactive message):
        {
            "type": "block_actions",
            "actions": [{"action_id": "approve_<request_id>", "value": "..."}],
            "user": {"id": "...", "name": "..."},
            "response_url": "..."
        }
        """
        if payload.get("type") != "block_actions":
            return {"status": "ignored", "reason": "not_block_actions"}
        
        actions = payload.get("actions", [])
        if not actions:
            return {"status": "error", "reason": "no_actions"}
        
        action = actions[0]
        action_id = action.get("action_id", "")
        
        # Parse request_id from action_id
        if action_id.startswith("approve_"):
            request_id = action_id.replace("approve_", "")
            decision = "approve"
        elif action_id.startswith("reject_"):
            request_id = action_id.replace("reject_", "")
            decision = "reject"
        else:
            return {"status": "error", "reason": "unknown_action"}
        
        user = payload.get("user", {})
        
        return {
            "request_id": request_id,
            "decision": decision,
            "provider": ApprovalProvider.SLACK,
            "user_id": user.get("id"),
            "user_name": user.get("name"),
            "timestamp": datetime.utcnow().isoformat()
        }


class CustomApprovalClient:
    """Client for sending approval requests to custom app."""
    
    def __init__(
        self,
        endpoint: str,
        method: str = "POST",
        headers: Optional[Dict[str, str]] = None,
        auth_secret_ref: Optional[Dict[str, str]] = None,
        callback_url: Optional[str] = None,
        use_webhook: bool = True,
        poll_interval: timedelta = timedelta(seconds=10)
    ):
        self.endpoint = endpoint
        self.method = method
        self.headers = headers or {}
        self.auth_secret_ref = auth_secret_ref
        self.callback_url = callback_url
        self.use_webhook = use_webhook
        self.poll_interval = poll_interval
        
        # Load auth from secret if provided
        if auth_secret_ref:
            auth_value = self._load_secret(
                auth_secret_ref.get("name"),
                auth_secret_ref.get("key", "api_key")
            )
            if auth_value:
                self.headers["Authorization"] = f"Bearer {auth_value}"
    
    def _load_secret(self, secret_name: str, key: str) -> Optional[str]:
        """Load secret from Kubernetes secret."""
        # TODO: Implement Kubernetes secret loading
        env_key = f"{secret_name.upper().replace('-', '_')}_{key.upper()}"
        return os.getenv(env_key)
    
    async def send_approval_request(self, request: ApprovalRequest) -> Dict[str, Any]:
        """
        Send approval request to custom app.
        
        Returns:
            Dict with approval_id, status_url, etc.
        """
        payload = {
            "request_id": request.request_id,
            "agent_name": request.agent_name,
            "action": request.action,
            "lambda_function": request.lambda_function,
            "parameters": request.parameters,
            "alertname": request.alertname,
            "correlation_id": request.correlation_id,
            "timeout": request.timeout.total_seconds(),
            "callback_url": self.callback_url,
            "metadata": request.metadata
        }
        
        async with httpx.AsyncClient() as client:
            response = await client.request(
                method=self.method,
                url=self.endpoint,
                json=payload,
                headers=self.headers,
                timeout=10.0
            )
            response.raise_for_status()
            result = response.json()
            
            logger.info(
                "custom_approval_request_sent",
                request_id=request.request_id,
                endpoint=self.endpoint,
                approval_id=result.get("approval_id")
            )
            
            return {
                "provider": ApprovalProvider.CUSTOM,
                "approval_id": result.get("approval_id"),
                "status_url": result.get("status_url"),
                "status": "sent"
            }
    
    async def check_approval_status(
        self,
        approval_id: str
    ) -> Dict[str, Any]:
        """
        Check approval status (if using polling).
        
        Returns:
            Dict with status, decision, etc.
        """
        status_url = f"{self.endpoint}/status/{approval_id}"
        
        async with httpx.AsyncClient() as client:
            response = await client.get(
                status_url,
                headers=self.headers,
                timeout=5.0
            )
            response.raise_for_status()
            return response.json()


class ApprovalManager:
    """Manages approval requests across multiple providers."""
    
    def __init__(
        self,
        slack_config: Optional[Dict[str, Any]] = None,
        custom_config: Optional[Dict[str, Any]] = None
    ):
        self.slack_client = None
        self.custom_client = None
        
        if slack_config:
            self.slack_client = SlackApprovalClient(**slack_config)
        
        if custom_config:
            self.custom_client = CustomApprovalClient(**custom_config)
        
        # In-memory store for approval requests (in production, use Redis/K8s CRD)
        self._requests: Dict[str, ApprovalRequest] = {}
    
    async def request_approval(
        self,
        request: ApprovalRequest
    ) -> ApprovalRequest:
        """
        Request approval from configured providers.
        
        Returns:
            ApprovalRequest with updated status
        """
        # Store request
        self._requests[request.request_id] = request
        
        # Send to providers
        for provider in request.providers:
            try:
                if provider == ApprovalProvider.SLACK and self.slack_client:
                    result = await self.slack_client.send_approval_request(request)
                    request.approvals[ApprovalProvider.SLACK] = ApprovalStatus.PENDING
                    request.metadata["slack"] = result
                
                elif provider == ApprovalProvider.CUSTOM and self.custom_client:
                    result = await self.custom_client.send_approval_request(request)
                    request.approvals[ApprovalProvider.CUSTOM] = ApprovalStatus.PENDING
                    request.metadata["custom"] = result
                
            except Exception as e:
                logger.error(
                    "approval_request_failed",
                    request_id=request.request_id,
                    provider=provider,
                    error=str(e),
                    exc_info=True
                )
                request.approvals[provider] = ApprovalStatus.REJECTED
        
        return request
    
    async def handle_approval_response(
        self,
        provider: ApprovalProvider,
        payload: Dict[str, Any]
    ) -> Optional[ApprovalRequest]:
        """
        Handle approval response from a provider.
        
        Returns:
            Updated ApprovalRequest if found
        """
        if provider == ApprovalProvider.SLACK and self.slack_client:
            response = await self.slack_client.handle_approval_response(payload)
        elif provider == ApprovalProvider.CUSTOM:
            # Custom app responses come via webhook or polling
            request_id = payload.get("request_id")
            decision = payload.get("decision")
            response = {
                "request_id": request_id,
                "decision": decision,
                "provider": ApprovalProvider.CUSTOM
            }
        else:
            return None
        
        request_id = response.get("request_id")
        if not request_id or request_id not in self._requests:
            logger.warning("approval_request_not_found", request_id=request_id)
            return None
        
        request = self._requests[request_id]
        decision = response.get("decision")
        
        # Update approval status
        if decision == "approve":
            request.approvals[provider] = ApprovalStatus.APPROVED
        elif decision == "reject":
            request.approvals[provider] = ApprovalStatus.REJECTED
        
        # Check if all required approvals are received
        if request.require_all:
            # All providers must approve
            all_approved = all(
                status == ApprovalStatus.APPROVED
                for status in request.approvals.values()
            )
            any_rejected = any(
                status == ApprovalStatus.REJECTED
                for status in request.approvals.values()
            )
            
            if all_approved:
                request.status = ApprovalStatus.APPROVED
            elif any_rejected:
                request.status = ApprovalStatus.REJECTED
        else:
            # Any provider can approve
            if any(
                status == ApprovalStatus.APPROVED
                for status in request.approvals.values()
            ):
                request.status = ApprovalStatus.APPROVED
            elif all(
                status == ApprovalStatus.REJECTED
                for status in request.approvals.values()
            ):
                request.status = ApprovalStatus.REJECTED
        
        return request
    
    def get_request(self, request_id: str) -> Optional[ApprovalRequest]:
        """Get approval request by ID."""
        return self._requests.get(request_id)
    
    async def check_timeouts(self) -> List[ApprovalRequest]:
        """
        Check for timed-out approval requests.
        
        Returns:
            List of timed-out requests
        """
        now = datetime.utcnow()
        timed_out = []
        
        for request in self._requests.values():
            if request.status != ApprovalStatus.PENDING:
                continue
            
            elapsed = now - request.created_at
            if elapsed >= request.timeout:
                request.status = ApprovalStatus.TIMEOUT
                
                if request.timeout_action == "approve":
                    request.status = ApprovalStatus.APPROVED
                elif request.timeout_action == "reject":
                    request.status = ApprovalStatus.REJECTED
                
                timed_out.append(request)
        
        return timed_out
