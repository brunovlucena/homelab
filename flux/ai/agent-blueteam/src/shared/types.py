"""
Type definitions for agent-blueteam defense runner.

🛡️ Blue Team Types - Defense is the best offense!
"""
from dataclasses import dataclass, field
from enum import Enum
from typing import Optional, Any
from datetime import datetime


class ThreatLevel(Enum):
    """Threat severity levels."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class DefenseAction(Enum):
    """Actions that can be taken in response to threats."""
    NONE = "none"
    MONITOR = "monitor"
    ALERT = "alert"
    BLOCK_NETWORK = "block_network"
    BLOCK_ADMISSION = "block_admission"
    BLOCK_RBAC = "block_rbac"
    QUARANTINE = "quarantine"
    SANDBOX = "sandbox"
    SANITIZE_INPUT = "sanitize_input"
    REVOKE_TOKEN = "revoke_token"
    TERMINATE = "terminate"


@dataclass
class ThreatReport:
    """Report generated after analyzing a potential threat."""
    id: str
    source_event: dict
    threat_level: ThreatLevel
    confidence: float = 0.0
    matched_signature: Optional[str] = None
    matched_pattern: Optional[str] = None
    recommended_action: Optional[DefenseAction] = None
    countermeasure: Optional[str] = None
    analyzed_at: str = field(default_factory=lambda: datetime.utcnow().isoformat())


@dataclass
class DefenseResult:
    """Result of executing a defensive action."""
    threat_report_id: str
    action_taken: DefenseAction
    success: bool
    message: str = ""
    artifacts: list[str] = field(default_factory=list)
    executed_at: str = field(default_factory=lambda: datetime.utcnow().isoformat())


@dataclass
class MAG7Boss:
    """
    MAG7 Dragon Boss - The final boss!
    
    A seven-headed dragon with CEO heads representing the Magnificent 7
    tech companies: Apple, Microsoft, Google, Amazon, Meta, Tesla, Nvidia.
    
    Each head has special powers and must be defeated by blocking exploits!
    """
    health: int = 1000
    max_health: int = 1000
    phase: str = "normal"  # normal, enraged, desperate
    attack_speed: float = 1.0
    defeated: bool = False
    
    # Special abilities by head
    abilities: dict = field(default_factory=lambda: {
        "apple": {"name": "Walled Garden", "effect": "blocks_network", "damage": 30},
        "microsoft": {"name": "Blue Screen", "effect": "crashes_pods", "damage": 25},
        "google": {"name": "Data Harvest", "effect": "exfiltrates_data", "damage": 35},
        "amazon": {"name": "Cloud Lock", "effect": "scales_infinitely", "damage": 20},
        "meta": {"name": "Privacy Void", "effect": "leaks_secrets", "damage": 40},
        "tesla": {"name": "Self-Drive", "effect": "autonomous_attack", "damage": 45},
        "nvidia": {"name": "GPU Meltdown", "effect": "resource_exhaustion", "damage": 50},
    })


@dataclass
class GameState:
    """State of the MAG7 Battle game."""
    game_id: str
    wave: int = 1
    score: int = 0
    player_health: int = 100
    mag7: MAG7Boss = field(default_factory=MAG7Boss)
    exploits_blocked: int = 0
    exploits_missed: int = 0
    defenses_activated: list[str] = field(default_factory=list)
    started_at: str = field(default_factory=lambda: datetime.utcnow().isoformat())
    ended_at: Optional[str] = None
    victory: Optional[bool] = None
