"""
CloudEvent fixtures for testing agent event handlers.

These fixtures provide standardized CloudEvent creation and validation
for all agents in the homelab infrastructure.
"""

import pytest
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Optional
from uuid import uuid4


@dataclass
class CloudEvent:
    """CloudEvent data structure for testing."""
    
    specversion: str = "1.0"
    type: str = ""
    source: str = ""
    id: str = field(default_factory=lambda: str(uuid4()))
    time: str = field(default_factory=lambda: datetime.now(timezone.utc).isoformat())
    datacontenttype: str = "application/json"
    subject: Optional[str] = None
    data: dict = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        """Convert to dictionary format."""
        result = {
            "specversion": self.specversion,
            "type": self.type,
            "source": self.source,
            "id": self.id,
            "time": self.time,
            "datacontenttype": self.datacontenttype,
            "data": self.data,
        }
        if self.subject:
            result["subject"] = self.subject
        return result
    
    def to_headers(self) -> dict:
        """Convert to HTTP headers format for CloudEvents."""
        headers = {
            "ce-specversion": self.specversion,
            "ce-type": self.type,
            "ce-source": self.source,
            "ce-id": self.id,
            "ce-time": self.time,
            "content-type": self.datacontenttype,
        }
        if self.subject:
            headers["ce-subject"] = self.subject
        return headers


class CloudEventFactory:
    """Factory for creating CloudEvents with common defaults."""
    
    # Common event types used across agents
    EVENT_TYPES = {
        # Agent Bruno (Chat)
        "chat.message": "io.homelab.chat.message",
        "chat.response": "io.homelab.chat.response",
        
        # Agent RedTeam (Security Testing)
        "exploit.started": "io.homelab.exploit.started",
        "exploit.success": "io.homelab.exploit.success",
        "exploit.blocked": "io.homelab.exploit.blocked",
        "exploit.failed": "io.homelab.exploit.failed",
        
        # Agent BlueTeam (Defense)
        "defense.activated": "io.homelab.defense.activated",
        "threat.detected": "io.homelab.threat.detected",
        "threat.mitigated": "io.homelab.threat.mitigated",
        
        # Agent Medical
        "medical.query": "io.homelab.medical.query",
        "medical.response": "io.homelab.medical.response",
        
        # Agent Contracts
        "contract.created": "io.homelab.contract.created",
        "contract.scanned": "io.homelab.contract.scanned",
        "vulnerability.found": "io.homelab.vulnerability.found",
        
        # Agent Restaurant
        "order.placed": "io.homelab.restaurant.order.placed",
        "order.prepared": "io.homelab.restaurant.order.prepared",
        "order.served": "io.homelab.restaurant.order.served",
        
        # Agent Store
        "product.listed": "io.homelab.store.product.listed",
        "sale.completed": "io.homelab.store.sale.completed",
        
        # Demo MAG7
        "mag7.attack": "io.homelab.mag7.attack",
        "mag7.defense": "io.homelab.mag7.defense",
        "game.start": "io.homelab.demo.game.start",
        "game.end": "io.homelab.demo.game.end",
        
        # Generic
        "health.check": "io.homelab.health.check",
        "metrics.request": "io.homelab.metrics.request",
    }
    
    def __init__(self, source: str = "/test"):
        self.source = source
    
    def create(
        self,
        type: str,
        data: Optional[dict] = None,
        source: Optional[str] = None,
        id: Optional[str] = None,
        subject: Optional[str] = None,
    ) -> CloudEvent:
        """Create a CloudEvent with the given parameters."""
        # Resolve type from shorthand if available
        event_type = self.EVENT_TYPES.get(type, type)
        
        return CloudEvent(
            type=event_type,
            source=source or self.source,
            id=id or str(uuid4()),
            subject=subject,
            data=data or {},
        )
    
    def create_chat_message(
        self,
        message: str,
        conversation_id: Optional[str] = None,
        user_id: str = "test-user",
    ) -> CloudEvent:
        """Create a chat message event."""
        return self.create(
            type="chat.message",
            data={
                "message": message,
                "conversation_id": conversation_id or str(uuid4()),
                "user_id": user_id,
            },
            source="/agent-bruno/chatbot",
        )
    
    def create_exploit_event(
        self,
        exploit_id: str,
        status: str = "success",
        namespace: str = "test-namespace",
        severity: str = "high",
    ) -> CloudEvent:
        """Create an exploit event from agent-redteam."""
        return self.create(
            type=f"exploit.{status}",
            data={
                "exploit_id": exploit_id,
                "namespace": namespace,
                "severity": severity,
                "status": status,
            },
            source="/agent-redteam/exploit-runner",
        )
    
    def create_defense_event(
        self,
        threat_type: str,
        action: str = "blocked",
        source_agent: str = "agent-blueteam",
    ) -> CloudEvent:
        """Create a defense event from agent-blueteam."""
        return self.create(
            type="defense.activated",
            data={
                "threat_type": threat_type,
                "action": action,
                "severity": "high",
            },
            source=f"/{source_agent}/defense-runner",
        )
    
    def create_contract_event(
        self,
        address: str,
        chain: str = "ethereum",
        vulnerabilities: Optional[list] = None,
    ) -> CloudEvent:
        """Create a contract scanning event."""
        return self.create(
            type="contract.scanned",
            data={
                "address": address,
                "chain": chain,
                "vulnerabilities": vulnerabilities or [],
            },
            source="/agent-contracts/scanner",
        )
    
    def create_mag7_attack(
        self,
        attack_type: str,
        head: str,
        damage: int = 50,
    ) -> CloudEvent:
        """Create a MAG7 battle attack event."""
        return self.create(
            type="mag7.attack",
            data={
                "attack_type": attack_type,
                "head": head,
                "damage": damage,
            },
            source="/demo-mag7-battle",
        )
    
    def create_batch(self, count: int, type: str, **kwargs) -> list[CloudEvent]:
        """Create a batch of CloudEvents."""
        return [self.create(type=type, **kwargs) for _ in range(count)]


@pytest.fixture
def cloudevent_factory():
    """Factory fixture for creating CloudEvents."""
    return CloudEventFactory()


@pytest.fixture
def sample_cloudevent(cloudevent_factory):
    """A basic sample CloudEvent for testing."""
    return cloudevent_factory.create(
        type="health.check",
        data={"status": "ok"},
    )


@pytest.fixture
def sample_cloudevent_batch(cloudevent_factory):
    """A batch of sample CloudEvents."""
    return cloudevent_factory.create_batch(5, type="health.check")


# Common event data fixtures
@pytest.fixture
def sample_chat_event(cloudevent_factory):
    """Sample chat message event."""
    return cloudevent_factory.create_chat_message(
        message="Hello, how can you help me?",
        user_id="test-user-123",
    )


@pytest.fixture
def sample_exploit_event(cloudevent_factory):
    """Sample exploit success event."""
    return cloudevent_factory.create_exploit_event(
        exploit_id="vuln-001",
        status="success",
        namespace="redteam-test",
        severity="critical",
    )


@pytest.fixture
def sample_defense_event(cloudevent_factory):
    """Sample defense activation event."""
    return cloudevent_factory.create_defense_event(
        threat_type="ssrf",
        action="blocked",
    )


@pytest.fixture
def sample_contract_event(cloudevent_factory):
    """Sample contract scan event."""
    return cloudevent_factory.create_contract_event(
        address="0x1234567890123456789012345678901234567890",
        chain="ethereum",
        vulnerabilities=[
            {"type": "reentrancy", "severity": "critical"},
        ],
    )


@pytest.fixture
def sample_mag7_event(cloudevent_factory):
    """Sample MAG7 battle event."""
    return cloudevent_factory.create_mag7_attack(
        attack_type="data_breach",
        head="google",
        damage=75,
    )
