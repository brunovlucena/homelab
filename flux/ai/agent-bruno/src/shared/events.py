"""
CloudEvents integration for agent-bruno.

Enables bidirectional communication with other agents via Knative Eventing.

Event Types Emitted:
- io.homelab.chat.message: User sent a message (for analytics)
- io.homelab.chat.intent.security: User asked about security/vulnerabilities
- io.homelab.chat.intent.status: User asked about service status

Event Types Received:
- io.homelab.vuln.found: Vulnerability found by agent-contracts
- io.homelab.exploit.validated: Exploit validated by agent-contracts
- io.homelab.alert.fired: System alert fired
"""
import os
import json
import asyncio
from typing import Optional, Callable, Awaitable
from dataclasses import dataclass, asdict
from datetime import datetime, timezone
from enum import Enum

import httpx
import structlog
from cloudevents.http import CloudEvent, to_structured, from_http
from cloudevents.conversion import to_json

logger = structlog.get_logger()

# =============================================================================
# Metrics - Import from shared metrics module to avoid duplicates
# =============================================================================

from .metrics import EVENTS_PUBLISHED as EVENTS_EMITTED, EVENTS_RECEIVED


# =============================================================================
# Event Type Definitions
# =============================================================================

class EventType(str, Enum):
    """CloudEvent types for agent-bruno."""
    # Emitted events
    CHAT_MESSAGE = "io.homelab.chat.message"
    CHAT_INTENT_SECURITY = "io.homelab.chat.intent.security"
    CHAT_INTENT_STATUS = "io.homelab.chat.intent.status"
    CHAT_INTENT_PROJECTS = "io.homelab.chat.intent.projects"
    CHAT_INTENT_SKILLS = "io.homelab.chat.intent.skills"
    CHAT_INTENT_CONTACT = "io.homelab.chat.intent.contact"
    
    # Request/Response events for agent-to-agent communication
    AGENT_QUERY = "io.homelab.agent.query"
    AGENT_RESPONSE = "io.homelab.agent.response"
    
    # Received events
    VULN_FOUND = "io.homelab.vuln.found"
    EXPLOIT_VALIDATED = "io.homelab.exploit.validated"
    ALERT_FIRED = "io.homelab.alert.fired"
    
    # Received from agent-contracts
    CONTRACTS_STATUS = "io.homelab.contracts.status"


# Security-related keywords for intent detection
SECURITY_KEYWORDS = [
    "vulnerability", "vuln", "exploit", "security", "hack", "attack",
    "reentrancy", "overflow", "contract", "audit", "scan", "critical",
    "threat", "risk", "compromised", "breach", "malicious",
]

STATUS_KEYWORDS = [
    "status", "health", "running", "down", "up", "available",
    "service", "pod", "deployment", "error", "failing", "working",
]

# Project-related keywords - triggers agent-contracts query
PROJECT_KEYWORDS = [
    "project", "projects", "portfolio", "work", "built", "created",
    "developed", "github", "repository", "repo", "code", "application",
    "app", "homelab", "infrastructure", "kubernetes", "k8s",
    "agent-contracts", "agent-bruno", "knative", "lambda",
]

# Skills-related keywords
SKILLS_KEYWORDS = [
    "skill", "skills", "technology", "technologies", "tech", "stack",
    "experience", "expertise", "know", "language", "framework",
    "tool", "tools", "proficient", "certified", "certification",
]

# Contact-related keywords
CONTACT_KEYWORDS = [
    "contact", "email", "reach", "hire", "hiring", "available",
    "linkedin", "github", "social", "message", "call", "meet",
]


@dataclass
class ChatMessageEvent:
    """Data for chat message events."""
    conversation_id: str
    message_length: int
    response_length: int
    tokens_used: int
    model: str
    duration_ms: float
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class ChatIntentEvent:
    """Data for detected intent events."""
    conversation_id: str
    intent: str
    query: str  # Truncated user query
    keywords_matched: list[str]
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class AgentQueryEvent:
    """Data for agent-to-agent query events."""
    conversation_id: str
    source_agent: str
    target_agent: str
    query_type: str  # e.g., "projects", "skills", "status"
    query: str
    context: Optional[str] = None
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class AgentResponseEvent:
    """Data for agent-to-agent response events."""
    conversation_id: str
    source_agent: str
    target_agent: str
    query_type: str
    response: str
    data: Optional[dict] = None
    success: bool = True
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class ContractsStatusResponse:
    """Response from agent-contracts about its status."""
    active_scans: int
    total_vulnerabilities: int
    recent_contracts: list[dict]
    chains_monitored: list[str]
    
    @classmethod
    def from_cloudevent(cls, event: CloudEvent) -> "ContractsStatusResponse":
        """Parse from CloudEvent data."""
        data = event.data or {}
        return cls(
            active_scans=data.get("active_scans", 0),
            total_vulnerabilities=data.get("total_vulnerabilities", 0),
            recent_contracts=data.get("recent_contracts", []),
            chains_monitored=data.get("chains_monitored", []),
        )


@dataclass
class VulnerabilityNotification:
    """Parsed vulnerability event from agent-contracts."""
    chain: str
    address: str
    max_severity: Optional[str]
    vulnerability_count: int
    vuln_types: list[str]
    
    @classmethod
    def from_cloudevent(cls, event: CloudEvent) -> "VulnerabilityNotification":
        """Parse from CloudEvent data."""
        data = event.data or {}
        vulns = data.get("vulnerabilities", [])
        return cls(
            chain=data.get("chain", "unknown"),
            address=data.get("address", "unknown"),
            max_severity=data.get("max_severity"),
            vulnerability_count=len(vulns),
            vuln_types=list(set(v.get("type", "unknown") for v in vulns)),
        )


@dataclass
class ExploitNotification:
    """Parsed exploit validation event from agent-contracts."""
    chain: str
    contract_address: str
    vulnerability_type: str
    validated: bool
    profit_potential: Optional[str]
    
    @classmethod
    def from_cloudevent(cls, event: CloudEvent) -> "ExploitNotification":
        """Parse from CloudEvent data."""
        data = event.data or {}
        return cls(
            chain=data.get("chain", "unknown"),
            contract_address=data.get("contract_address", "unknown"),
            vulnerability_type=data.get("vulnerability_type", "unknown"),
            validated=data.get("validated", False),
            profit_potential=data.get("profit_potential"),
        )


# =============================================================================
# Event Publisher
# =============================================================================

class EventPublisher:
    """
    Publishes CloudEvents to Knative broker.
    
    Uses HTTP POST to the broker ingress endpoint.
    """
    
    def __init__(self, broker_url: str = None):
        self.broker_url = broker_url or os.getenv(
            "K_SINK",  # Knative standard env var
            os.getenv(
                "KNATIVE_BROKER_URL",
                "http://agent-bruno-broker-broker-ingress.ai.svc.cluster.local"  # Fixed: broker is in 'ai' namespace
            )
        )
        self.source = "/agent-bruno/chatbot"
        self._client: Optional[httpx.AsyncClient] = None
        self._enabled = bool(self.broker_url)  # Only enable if broker URL is set
    
    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client."""
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(timeout=10.0)
        return self._client
    
    async def publish(self, event: CloudEvent) -> bool:
        """
        Publish a CloudEvent to the broker.
        
        Returns True if successful, False otherwise.
        """
        if not self._enabled or not self.broker_url:
            logger.debug("event_publishing_disabled", event_type=event.get("type"))
            return False
            
        try:
            client = await self._get_client()
            headers, body = to_structured(event)
            
            response = await client.post(
                self.broker_url,
                headers=headers,
                content=body,
                timeout=5.0,  # Add timeout to prevent hanging
            )
            
            if response.status_code in (200, 202, 204):
                EVENTS_EMITTED.labels(
                    event_type=event["type"],
                    status="success"
                ).inc()
                logger.debug("event_published", event_type=event["type"])
                return True
            else:
                EVENTS_EMITTED.labels(
                    event_type=event["type"],
                    status="error"
                ).inc()
                logger.warning(
                    "event_publish_failed",
                    event_type=event["type"],
                    status_code=response.status_code,
                    broker_url=self.broker_url,
                )
                return False
                
        except (httpx.ConnectError, httpx.TimeoutException, OSError) as e:
            # DNS resolution errors, connection errors - log as warning, not error
            EVENTS_EMITTED.labels(
                event_type=event.get("type", "unknown"),
                status="error"
            ).inc()
            logger.warning(
                "event_publish_error",
                error=str(e),
                error_type=type(e).__name__,
                broker_url=self.broker_url,
                event_type=event.get("type")
            )
            return False
        except Exception as e:
            EVENTS_EMITTED.labels(
                event_type=event.get("type", "unknown"),
                status="error"
            ).inc()
            logger.error(
                "event_publish_error",
                error=str(e),
                error_type=type(e).__name__,
                broker_url=self.broker_url,
                event_type=event.get("type")
            )
            return False
    
    async def emit_chat_message(self, data: ChatMessageEvent) -> bool:
        """Emit a chat message event."""
        event = CloudEvent({
            "type": EventType.CHAT_MESSAGE.value,
            "source": self.source,
            "subject": data.conversation_id,
        }, asdict(data))
        return await self.publish(event)
    
    async def emit_intent(self, intent_type: EventType, data: ChatIntentEvent) -> bool:
        """Emit an intent detection event."""
        event = CloudEvent({
            "type": intent_type.value,
            "source": self.source,
            "subject": data.conversation_id,
        }, asdict(data))
        return await self.publish(event)
    
    async def query_agent(self, data: AgentQueryEvent) -> bool:
        """
        Send a query to another agent via CloudEvents.
        
        The target agent should respond with an AGENT_RESPONSE event.
        """
        event = CloudEvent({
            "type": EventType.AGENT_QUERY.value,
            "source": self.source,
            "subject": f"{data.target_agent}/{data.query_type}",
        }, asdict(data))
        
        logger.info(
            "agent_query_sent",
            target=data.target_agent,
            query_type=data.query_type,
            conversation_id=data.conversation_id,
        )
        return await self.publish(event)
    
    async def respond_to_agent(self, data: AgentResponseEvent) -> bool:
        """
        Send a response to an agent query.
        """
        event = CloudEvent({
            "type": EventType.AGENT_RESPONSE.value,
            "source": self.source,
            "subject": f"{data.target_agent}/{data.query_type}",
        }, asdict(data))
        return await self.publish(event)
    
    async def close(self):
        """Close HTTP client."""
        if self._client and not self._client.is_closed:
            await self._client.aclose()


# =============================================================================
# Event Subscriber / Handler
# =============================================================================

# Type alias for event handlers
EventHandler = Callable[[CloudEvent], Awaitable[None]]


class EventSubscriber:
    """
    Handles incoming CloudEvents from Knative triggers.
    
    Registered handlers are called when matching events arrive.
    """
    
    def __init__(self):
        self._handlers: dict[str, list[EventHandler]] = {}
        self._notifications: list[dict] = []  # Store recent notifications
        self._max_notifications = 50
    
    def register(self, event_type: str, handler: EventHandler):
        """Register a handler for an event type."""
        if event_type not in self._handlers:
            self._handlers[event_type] = []
        self._handlers[event_type].append(handler)
        logger.info("event_handler_registered", event_type=event_type)
    
    async def handle(self, event: CloudEvent) -> bool:
        """
        Handle an incoming CloudEvent.
        
        Returns True if handled successfully.
        """
        event_type = event["type"]
        handlers = self._handlers.get(event_type, [])
        
        if not handlers:
            EVENTS_RECEIVED.labels(event_type=event_type, status="no_handler").inc()
            logger.debug("event_no_handler", event_type=event_type)
            return False
        
        try:
            for handler in handlers:
                await handler(event)
            
            EVENTS_RECEIVED.labels(event_type=event_type, status="success").inc()
            logger.info("event_handled", event_type=event_type)
            return True
            
        except Exception as e:
            EVENTS_RECEIVED.labels(event_type=event_type, status="error").inc()
            logger.error("event_handler_error", event_type=event_type, error=str(e))
            return False
    
    def add_notification(self, notification: dict):
        """Store a notification for the chatbot to reference."""
        self._notifications.insert(0, notification)
        if len(self._notifications) > self._max_notifications:
            self._notifications = self._notifications[:self._max_notifications]
    
    def get_recent_notifications(self, limit: int = 5) -> list[dict]:
        """Get recent notifications."""
        return self._notifications[:limit]
    
    def clear_notifications(self):
        """Clear all stored notifications."""
        self._notifications.clear()


# =============================================================================
# Intent Detection
# =============================================================================

def detect_intent(message: str) -> tuple[Optional[EventType], list[str]]:
    """
    Detect user intent from message.
    
    Returns (intent_type, matched_keywords) or (None, []) if no intent detected.
    
    Priority order:
    1. Security (highest - could be urgent)
    2. Projects (agent-contracts integration)
    3. Skills
    4. Contact
    5. Status (lowest)
    """
    message_lower = message.lower()
    
    # Check for security intent (highest priority)
    security_matches = [kw for kw in SECURITY_KEYWORDS if kw in message_lower]
    if security_matches:
        return EventType.CHAT_INTENT_SECURITY, security_matches
    
    # Check for projects intent (triggers agent-contracts query)
    project_matches = [kw for kw in PROJECT_KEYWORDS if kw in message_lower]
    if project_matches:
        return EventType.CHAT_INTENT_PROJECTS, project_matches
    
    # Check for skills intent
    skills_matches = [kw for kw in SKILLS_KEYWORDS if kw in message_lower]
    if skills_matches:
        return EventType.CHAT_INTENT_SKILLS, skills_matches
    
    # Check for contact intent
    contact_matches = [kw for kw in CONTACT_KEYWORDS if kw in message_lower]
    if contact_matches:
        return EventType.CHAT_INTENT_CONTACT, contact_matches
    
    # Check for status intent
    status_matches = [kw for kw in STATUS_KEYWORDS if kw in message_lower]
    if status_matches:
        return EventType.CHAT_INTENT_STATUS, status_matches
    
    return None, []


# =============================================================================
# Default Event Handlers
# =============================================================================

def create_default_handlers(subscriber: EventSubscriber):
    """
    Create and register default event handlers.
    
    These handlers process events from agent-contracts and store
    notifications that the chatbot can reference.
    """
    
    async def handle_vuln_found(event: CloudEvent):
        """Handle vulnerability found events from agent-contracts."""
        notification = VulnerabilityNotification.from_cloudevent(event)
        
        if notification.max_severity in ("critical", "high"):
            subscriber.add_notification({
                "type": "vulnerability",
                "severity": notification.max_severity,
                "chain": notification.chain,
                "address": notification.address[:10] + "...",
                "count": notification.vulnerability_count,
                "vuln_types": notification.vuln_types,
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "message": f"🔴 {notification.max_severity.upper()} vulnerability found on {notification.chain}: {notification.vulnerability_count} issue(s) detected",
            })
            logger.warning(
                "high_severity_vuln_received",
                chain=notification.chain,
                severity=notification.max_severity,
            )
    
    async def handle_exploit_validated(event: CloudEvent):
        """Handle exploit validated events from agent-contracts."""
        notification = ExploitNotification.from_cloudevent(event)
        
        if notification.validated:
            subscriber.add_notification({
                "type": "exploit",
                "chain": notification.chain,
                "address": notification.contract_address[:10] + "...",
                "vuln_type": notification.vulnerability_type,
                "profit_potential": notification.profit_potential,
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "message": f"🚨 CRITICAL: Validated {notification.vulnerability_type} exploit on {notification.chain}",
            })
            logger.error(
                "validated_exploit_received",
                chain=notification.chain,
                vuln_type=notification.vulnerability_type,
            )
    
    async def handle_alert_fired(event: CloudEvent):
        """Handle system alert events."""
        data = event.data or {}
        subscriber.add_notification({
            "type": "alert",
            "alertname": data.get("alertname", "unknown"),
            "severity": data.get("severity", "warning"),
            "message": data.get("message", "System alert fired"),
            "timestamp": datetime.now(timezone.utc).isoformat(),
        })
    
    async def handle_agent_response(event: CloudEvent):
        """Handle responses from other agents (e.g., agent-contracts)."""
        data = event.data or {}
        source_agent = data.get("source_agent", "unknown")
        query_type = data.get("query_type", "unknown")
        response = data.get("response", "")
        extra_data = data.get("data", {})
        
        subscriber.add_notification({
            "type": "agent_response",
            "source_agent": source_agent,
            "query_type": query_type,
            "response": response,
            "data": extra_data,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "message": f"📬 Response from {source_agent}: {response[:100]}..." if len(response) > 100 else f"📬 Response from {source_agent}: {response}",
        })
        
        logger.info(
            "agent_response_received",
            source_agent=source_agent,
            query_type=query_type,
        )
    
    async def handle_contracts_status(event: CloudEvent):
        """Handle status updates from agent-contracts."""
        status = ContractsStatusResponse.from_cloudevent(event)
        
        subscriber.add_notification({
            "type": "contracts_status",
            "active_scans": status.active_scans,
            "total_vulnerabilities": status.total_vulnerabilities,
            "chains_monitored": status.chains_monitored,
            "recent_contracts": status.recent_contracts[:3],  # Limit to 3
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "message": f"📊 Agent-Contracts: {status.active_scans} active scans, {status.total_vulnerabilities} vulnerabilities found, monitoring {', '.join(status.chains_monitored)}",
        })
        
        logger.info(
            "contracts_status_received",
            active_scans=status.active_scans,
            total_vulns=status.total_vulnerabilities,
        )
    
    # Register handlers
    subscriber.register(EventType.VULN_FOUND.value, handle_vuln_found)
    subscriber.register(EventType.EXPLOIT_VALIDATED.value, handle_exploit_validated)
    subscriber.register(EventType.ALERT_FIRED.value, handle_alert_fired)
    subscriber.register(EventType.AGENT_RESPONSE.value, handle_agent_response)
    subscriber.register(EventType.CONTRACTS_STATUS.value, handle_contracts_status)


# =============================================================================
# Global Instances
# =============================================================================

# These are initialized by the FastAPI lifespan
publisher: Optional[EventPublisher] = None
subscriber: Optional[EventSubscriber] = None


def init_events() -> tuple[EventPublisher, EventSubscriber]:
    """Initialize event publisher and subscriber."""
    global publisher, subscriber
    
    publisher = EventPublisher()
    subscriber = EventSubscriber()
    create_default_handlers(subscriber)
    
    logger.info(
        "events_initialized",
        broker_url=publisher.broker_url,
    )
    
    return publisher, subscriber


async def shutdown_events():
    """Shutdown event resources."""
    global publisher
    if publisher:
        await publisher.close()
        publisher = None
