"""
Alert Dispatcher - Multi-channel notification delivery.
"""
import os
import json
from typing import Optional
from dataclasses import dataclass
from enum import Enum
from datetime import datetime, timezone

import httpx
import structlog
from cloudevents.http import CloudEvent
from prometheus_client import Counter, Histogram

logger = structlog.get_logger()

# Metrics
ALERTS_SENT = Counter(
    "alerts_sent_total",
    "Total alerts sent",
    ["channel", "severity", "status"]
)
ALERT_LATENCY = Histogram(
    "alert_delivery_latency_seconds",
    "Time to deliver alert",
    ["channel"]
)


class Severity(str, Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"


class AlertChannel(str, Enum):
    TELEGRAM = "telegram"
    DISCORD = "discord"
    GRAFANA = "grafana"
    GRAFANA_INCIDENT = "grafana_incident"
    EMAIL = "email"


# Channel routing by severity
SEVERITY_CHANNELS = {
    Severity.CRITICAL: [
        AlertChannel.TELEGRAM,
        AlertChannel.DISCORD,
        AlertChannel.GRAFANA_INCIDENT,
        AlertChannel.EMAIL,
    ],
    Severity.HIGH: [
        AlertChannel.TELEGRAM,
        AlertChannel.DISCORD,
        AlertChannel.GRAFANA,
    ],
    Severity.MEDIUM: [
        AlertChannel.DISCORD,
        AlertChannel.GRAFANA,
    ],
    Severity.LOW: [
        AlertChannel.GRAFANA,
    ],
    Severity.INFO: [
        AlertChannel.GRAFANA,
    ],
}


@dataclass
class Alert:
    """Alert to be dispatched."""
    severity: Severity
    title: str
    chain: str
    contract_address: str
    vulnerability_type: str
    description: str = ""
    profit_potential: Optional[str] = None
    exploit_validated: bool = False
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "severity": self.severity.value,
            "title": self.title,
            "chain": self.chain,
            "contract_address": self.contract_address,
            "vulnerability_type": self.vulnerability_type,
            "description": self.description,
            "profit_potential": self.profit_potential,
            "exploit_validated": self.exploit_validated,
            "timestamp": self.timestamp,
        }


class TelegramNotifier:
    """Send alerts to Telegram."""
    
    def __init__(self):
        self.bot_token = os.getenv("TELEGRAM_BOT_TOKEN")
        self.chat_id = os.getenv("TELEGRAM_CHAT_ID")
        self.enabled = bool(self.bot_token and self.chat_id)
    
    async def send(self, alert: Alert) -> bool:
        if not self.enabled:
            logger.warning("telegram_not_configured")
            return False
        
        # Format message with emoji based on severity
        emoji = {
            Severity.CRITICAL: "🚨",
            Severity.HIGH: "⚠️",
            Severity.MEDIUM: "🟡",
            Severity.LOW: "🟢",
            Severity.INFO: "ℹ️",
        }.get(alert.severity, "📢")
        
        message = f"""
{emoji} *{alert.severity.value.upper()}: {alert.title}*

*Chain:* `{alert.chain}`
*Contract:* `{alert.contract_address}`
*Vulnerability:* {alert.vulnerability_type}
{f"*Profit Potential:* {alert.profit_potential}" if alert.profit_potential else ""}
{f"*Exploit Validated:* ✅" if alert.exploit_validated else ""}

{alert.description[:500] if alert.description else ""}

_Detected by Agent-Contracts_
"""
        
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.post(
                    f"https://api.telegram.org/bot{self.bot_token}/sendMessage",
                    json={
                        "chat_id": self.chat_id,
                        "text": message.strip(),
                        "parse_mode": "Markdown",
                        "disable_web_page_preview": True,
                    }
                )
                response.raise_for_status()
                return True
        except Exception as e:
            logger.error("telegram_send_failed", error=str(e))
            return False


class DiscordNotifier:
    """Send alerts to Discord webhook."""
    
    def __init__(self):
        self.webhook_url = os.getenv("DISCORD_WEBHOOK_URL")
        self.enabled = bool(self.webhook_url)
    
    async def send(self, alert: Alert) -> bool:
        if not self.enabled:
            logger.warning("discord_not_configured")
            return False
        
        # Color based on severity
        color = {
            Severity.CRITICAL: 0xFF0000,  # Red
            Severity.HIGH: 0xFF8C00,      # Orange
            Severity.MEDIUM: 0xFFD700,    # Gold
            Severity.LOW: 0x32CD32,       # Green
            Severity.INFO: 0x4169E1,      # Blue
        }.get(alert.severity, 0x808080)
        
        embed = {
            "title": f"🛡️ {alert.title}",
            "color": color,
            "fields": [
                {"name": "Severity", "value": alert.severity.value.upper(), "inline": True},
                {"name": "Chain", "value": alert.chain, "inline": True},
                {"name": "Type", "value": alert.vulnerability_type, "inline": True},
                {"name": "Contract", "value": f"`{alert.contract_address}`", "inline": False},
            ],
            "footer": {"text": "Agent-Contracts Security Scanner"},
            "timestamp": alert.timestamp,
        }
        
        if alert.profit_potential:
            embed["fields"].append({
                "name": "Profit Potential",
                "value": alert.profit_potential,
                "inline": True
            })
        
        if alert.exploit_validated:
            embed["fields"].append({
                "name": "Exploit Status",
                "value": "✅ Validated",
                "inline": True
            })
        
        if alert.description:
            embed["description"] = alert.description[:1000]
        
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.post(
                    self.webhook_url,
                    json={"embeds": [embed]}
                )
                response.raise_for_status()
                return True
        except Exception as e:
            logger.error("discord_send_failed", error=str(e))
            return False


class GrafanaNotifier:
    """Send metrics and create incidents in Grafana."""
    
    def __init__(self):
        self.grafana_url = os.getenv("GRAFANA_URL", "http://grafana.monitoring:3000")
        self.api_key = os.getenv("GRAFANA_API_KEY")
        self.enabled = bool(self.api_key)
    
    async def send(self, alert: Alert) -> bool:
        """Push alert as annotation to Grafana."""
        if not self.enabled:
            logger.warning("grafana_not_configured")
            return False
        
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                # Create annotation
                response = await client.post(
                    f"{self.grafana_url}/api/annotations",
                    headers={"Authorization": f"Bearer {self.api_key}"},
                    json={
                        "text": f"{alert.severity.value.upper()}: {alert.title}\n\nChain: {alert.chain}\nContract: {alert.contract_address}\nType: {alert.vulnerability_type}",
                        "tags": [
                            "agent-contracts",
                            f"severity:{alert.severity.value}",
                            f"chain:{alert.chain}",
                            alert.vulnerability_type,
                        ],
                    }
                )
                response.raise_for_status()
                return True
        except Exception as e:
            logger.error("grafana_annotation_failed", error=str(e))
            return False
    
    async def create_incident(self, alert: Alert) -> bool:
        """Create Grafana Incident for critical findings."""
        if not self.enabled:
            return False
        
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.post(
                    f"{self.grafana_url}/api/plugins/grafana-incident-app/resources/api/v1/incidents",
                    headers={"Authorization": f"Bearer {self.api_key}"},
                    json={
                        "title": alert.title,
                        "severity": alert.severity.value,
                        "status": "active",
                        "labels": [
                            {"key": "chain", "value": alert.chain},
                            {"key": "contract", "value": alert.contract_address},
                            {"key": "type", "value": alert.vulnerability_type},
                        ],
                        "attachCaption": "Contract Details",
                        "attachUrl": f"https://etherscan.io/address/{alert.contract_address}",
                    }
                )
                response.raise_for_status()
                logger.info("grafana_incident_created", title=alert.title)
                return True
        except Exception as e:
            logger.error("grafana_incident_failed", error=str(e))
            return False


class AlertDispatcher:
    """Main dispatcher orchestrating multi-channel delivery."""
    
    def __init__(self):
        self.telegram = TelegramNotifier()
        self.discord = DiscordNotifier()
        self.grafana = GrafanaNotifier()
    
    async def dispatch(self, alert: Alert) -> dict:
        """
        Dispatch alert to appropriate channels based on severity.
        
        Returns dict with delivery status per channel.
        """
        log = logger.bind(
            severity=alert.severity.value,
            chain=alert.chain,
            contract=alert.contract_address,
        )
        log.info("dispatching_alert")
        
        channels = SEVERITY_CHANNELS.get(alert.severity, [AlertChannel.GRAFANA])
        results = {}
        
        for channel in channels:
            success = False
            
            with ALERT_LATENCY.labels(channel=channel.value).time():
                if channel == AlertChannel.TELEGRAM:
                    success = await self.telegram.send(alert)
                elif channel == AlertChannel.DISCORD:
                    success = await self.discord.send(alert)
                elif channel == AlertChannel.GRAFANA:
                    success = await self.grafana.send(alert)
                elif channel == AlertChannel.GRAFANA_INCIDENT:
                    success = await self.grafana.create_incident(alert)
                elif channel == AlertChannel.EMAIL:
                    # TODO: Implement email notifications
                    success = False
            
            results[channel.value] = success
            ALERTS_SENT.labels(
                channel=channel.value,
                severity=alert.severity.value,
                status="success" if success else "failed"
            ).inc()
        
        log.info("alert_dispatched", results=results)
        return results


async def handle_alert_event(event: CloudEvent) -> dict:
    """Handle incoming alert CloudEvent."""
    data = event.data
    
    alert = Alert(
        severity=Severity(data.get("severity", "info")),
        title=data.get("title", "Security Alert"),
        chain=data.get("chain", "unknown"),
        contract_address=data.get("contract_address", ""),
        vulnerability_type=data.get("vulnerability_type", "unknown"),
        description=data.get("description", ""),
        profit_potential=data.get("profit_potential"),
        exploit_validated=data.get("exploit_validated", False),
    )
    
    dispatcher = AlertDispatcher()
    return await dispatcher.dispatch(alert)

