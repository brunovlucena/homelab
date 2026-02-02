"""
Notifi Adapter - Transforms Alertmanager webhooks to notifi-services format.

This adapter receives Alertmanager webhook payloads and forwards them to
notifi-services for multi-channel delivery (Telegram, SMS, Email, Discord, etc.)
"""
import os
import json
from typing import Optional
from dataclasses import dataclass
from datetime import datetime

import httpx
import structlog
from prometheus_client import Counter, Histogram

logger = structlog.get_logger()

# Metrics
ALERTS_RECEIVED = Counter(
    "notifi_adapter_alerts_received_total",
    "Total alerts received from Alertmanager",
    ["severity", "alertname"]
)
ALERTS_FORWARDED = Counter(
    "notifi_adapter_alerts_forwarded_total",
    "Total alerts forwarded to notifi-services",
    ["channel", "status"]
)
FORWARD_LATENCY = Histogram(
    "notifi_adapter_forward_latency_seconds",
    "Time to forward alert to notifi-services",
    ["channel"]
)


@dataclass
class AlertmanagerAlert:
    """Parsed Alertmanager alert."""
    alertname: str
    severity: str
    status: str  # firing, resolved
    starts_at: str
    ends_at: Optional[str]
    labels: dict
    annotations: dict
    fingerprint: str
    generator_url: Optional[str] = None
    
    @property
    def chain(self) -> str:
        return self.labels.get("chain", "unknown")
    
    @property
    def contract_address(self) -> str:
        return self.labels.get("contract_address", "")
    
    @property
    def vuln_type(self) -> str:
        return self.labels.get("vuln_type", "unknown")
    
    @property
    def summary(self) -> str:
        return self.annotations.get("summary", self.alertname)
    
    @property
    def description(self) -> str:
        return self.annotations.get("description", "")


@dataclass 
class NotifiMessage:
    """Message format for notifi-services webhook."""
    subject: str
    body: str
    severity: str
    metadata: dict
    
    def to_webhook_payload(self) -> dict:
        """Convert to notifi-services webhook format."""
        return {
            "eventId": f"agent-contracts-{datetime.utcnow().timestamp()}",
            "ruleType": "DirectTenantMessageRule",
            "version": "1.0",
            "data": {
                "subject": self.subject,
                "body": self.body,
                "severity": self.severity,
                "source": "agent-contracts",
                "metadata": self.metadata,
            }
        }


class NotifiClient:
    """Client for notifi-services webhook API."""
    
    def __init__(self):
        self.webhook_url = os.getenv("NOTIFI_WEBHOOK_URL")
        self.api_key = os.getenv("NOTIFI_API_KEY")
        self.tenant_id = os.getenv("NOTIFI_TENANT_ID", "agent-contracts")
        self.enabled = bool(self.webhook_url)
        
        if not self.enabled:
            logger.warning("notifi_not_configured", 
                          hint="Set NOTIFI_WEBHOOK_URL to enable")
    
    async def send(self, message: NotifiMessage, channels: list[str] = None) -> dict:
        """
        Send message to notifi-services.
        
        Args:
            message: Message to send
            channels: Optional list of channels to use (telegram, discord, email, sms)
        
        Returns:
            dict with delivery status
        """
        if not self.enabled:
            return {"status": "skipped", "reason": "notifi not configured"}
        
        results = {}
        channels = channels or ["webhook"]
        
        for channel in channels:
            with FORWARD_LATENCY.labels(channel=channel).time():
                try:
                    async with httpx.AsyncClient(timeout=10.0) as client:
                        response = await client.post(
                            self.webhook_url,
                            headers={
                                "Content-Type": "application/json",
                                "X-Notifi-Tenant-Id": self.tenant_id,
                                "Authorization": f"Bearer {self.api_key}" if self.api_key else "",
                            },
                            json=message.to_webhook_payload()
                        )
                        response.raise_for_status()
                        results[channel] = "success"
                        ALERTS_FORWARDED.labels(channel=channel, status="success").inc()
                        
                except Exception as e:
                    logger.error("notifi_send_failed", 
                               channel=channel, error=str(e))
                    results[channel] = f"error: {str(e)}"
                    ALERTS_FORWARDED.labels(channel=channel, status="failed").inc()
        
        return results


class AlertmanagerAdapter:
    """Transforms Alertmanager webhooks to notifi-services format."""
    
    # Severity to emoji mapping
    SEVERITY_EMOJI = {
        "critical": "🚨",
        "high": "⚠️",
        "warning": "🟡",
        "medium": "🟡", 
        "low": "🟢",
        "info": "ℹ️",
    }
    
    # Severity to channels mapping
    SEVERITY_CHANNELS = {
        "critical": ["telegram", "discord", "email", "webhook"],
        "high": ["telegram", "discord", "webhook"],
        "warning": ["discord", "webhook"],
        "medium": ["discord", "webhook"],
        "low": ["webhook"],
        "info": ["webhook"],
    }
    
    def __init__(self):
        self.notifi = NotifiClient()
    
    def parse_alertmanager_payload(self, payload: dict) -> list[AlertmanagerAlert]:
        """Parse Alertmanager webhook payload into alerts."""
        alerts = []
        
        for alert_data in payload.get("alerts", []):
            alert = AlertmanagerAlert(
                alertname=alert_data.get("labels", {}).get("alertname", "unknown"),
                severity=alert_data.get("labels", {}).get("severity", "info"),
                status=alert_data.get("status", "firing"),
                starts_at=alert_data.get("startsAt", ""),
                ends_at=alert_data.get("endsAt"),
                labels=alert_data.get("labels", {}),
                annotations=alert_data.get("annotations", {}),
                fingerprint=alert_data.get("fingerprint", ""),
                generator_url=alert_data.get("generatorURL"),
            )
            alerts.append(alert)
            
            ALERTS_RECEIVED.labels(
                severity=alert.severity,
                alertname=alert.alertname
            ).inc()
        
        return alerts
    
    def transform_to_notifi(self, alert: AlertmanagerAlert) -> NotifiMessage:
        """Transform Alertmanager alert to notifi-services message."""
        emoji = self.SEVERITY_EMOJI.get(alert.severity.lower(), "📢")
        status_emoji = "🔥" if alert.status == "firing" else "✅"
        
        # Build subject
        subject = f"{emoji} [{alert.severity.upper()}] {alert.alertname}"
        if alert.status == "resolved":
            subject = f"✅ [RESOLVED] {alert.alertname}"
        
        # Build body
        body_parts = [
            f"**Status:** {status_emoji} {alert.status.upper()}",
            f"**Severity:** {alert.severity}",
        ]
        
        if alert.chain:
            body_parts.append(f"**Chain:** {alert.chain}")
        if alert.contract_address:
            body_parts.append(f"**Contract:** `{alert.contract_address}`")
        if alert.vuln_type and alert.vuln_type != "unknown":
            body_parts.append(f"**Type:** {alert.vuln_type}")
        
        if alert.summary:
            body_parts.append(f"\n**Summary:** {alert.summary}")
        if alert.description:
            body_parts.append(f"\n{alert.description}")
        
        body_parts.append(f"\n_Started: {alert.starts_at}_")
        if alert.generator_url:
            body_parts.append(f"\n[View in Grafana]({alert.generator_url})")
        
        body = "\n".join(body_parts)
        
        return NotifiMessage(
            subject=subject,
            body=body,
            severity=alert.severity,
            metadata={
                "alertname": alert.alertname,
                "fingerprint": alert.fingerprint,
                "chain": alert.chain,
                "contract_address": alert.contract_address,
                "vuln_type": alert.vuln_type,
                "status": alert.status,
                "labels": alert.labels,
            }
        )
    
    async def process_webhook(self, payload: dict) -> dict:
        """
        Process Alertmanager webhook and forward to notifi-services.
        
        Returns dict with processing results.
        """
        log = logger.bind(
            receiver=payload.get("receiver"),
            group_key=payload.get("groupKey"),
        )
        log.info("alertmanager_webhook_received")
        
        alerts = self.parse_alertmanager_payload(payload)
        results = []
        
        for alert in alerts:
            log = log.bind(
                alertname=alert.alertname,
                severity=alert.severity,
                status=alert.status,
            )
            
            # Skip resolved alerts for low severity
            if alert.status == "resolved" and alert.severity.lower() in ["low", "info"]:
                log.info("skipping_resolved_low_severity")
                continue
            
            message = self.transform_to_notifi(alert)
            channels = self.SEVERITY_CHANNELS.get(alert.severity.lower(), ["webhook"])
            
            result = await self.notifi.send(message, channels)
            results.append({
                "alertname": alert.alertname,
                "fingerprint": alert.fingerprint,
                "status": alert.status,
                "delivery": result,
            })
            
            log.info("alert_forwarded", channels=channels, result=result)
        
        return {
            "processed": len(results),
            "alerts": results,
        }

