"""
Shared types for agent-contracts.
"""
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional
from datetime import datetime, timezone


class Severity(str, Enum):
    """Vulnerability severity levels."""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"


class VulnType(str, Enum):
    """Vulnerability type classification."""
    REENTRANCY = "reentrancy"
    ACCESS_CONTROL = "access_control"
    INTEGER_OVERFLOW = "integer_overflow"
    FLASH_LOAN = "flash_loan"
    ORACLE_MANIPULATION = "oracle_manipulation"
    MEV = "mev"
    STORAGE_COLLISION = "storage_collision"
    DELEGATECALL = "delegatecall"
    ARBITRARY_CALL = "arbitrary_call"
    MISSING_MODIFIER = "missing_modifier"
    UNCHECKED_RETURN = "unchecked_return"
    FRONT_RUNNING = "front_running"
    OTHER = "other"


@dataclass
class Vulnerability:
    """Detected vulnerability."""
    type: VulnType
    severity: Severity
    confidence: float  # 0.0 - 1.0
    location: str  # file:line or function name
    description: str
    recommendation: str
    analyzer: str  # slither, mythril, llm
    exploit_feasibility: Optional[str] = None
    
    def to_dict(self) -> dict:
        return {
            "type": self.type.value,
            "severity": self.severity.value,
            "confidence": self.confidence,
            "location": self.location,
            "description": self.description,
            "recommendation": self.recommendation,
            "analyzer": self.analyzer,
            "exploit_feasibility": self.exploit_feasibility,
        }


@dataclass
class ScanResult:
    """Result of vulnerability scan."""
    chain: str
    address: str
    vulnerabilities: list[Vulnerability] = field(default_factory=list)
    scan_duration_seconds: float = 0.0
    analyzers_used: list[str] = field(default_factory=list)
    error: Optional[str] = None
    
    @property
    def has_critical(self) -> bool:
        return any(v.severity == Severity.CRITICAL for v in self.vulnerabilities)
    
    @property
    def max_severity(self) -> Optional[Severity]:
        if not self.vulnerabilities:
            return None
        severity_order = [Severity.CRITICAL, Severity.HIGH, Severity.MEDIUM, Severity.LOW, Severity.INFO]
        for severity in severity_order:
            if any(v.severity == severity for v in self.vulnerabilities):
                return severity
        return None


@dataclass
class ExploitResult:
    """Result of exploit validation."""
    chain: str
    contract_address: str
    vulnerability_type: VulnType
    validated: bool
    profit_potential: Optional[str] = None
    gas_used: int = 0
    error: Optional[str] = None
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "chain": self.chain,
            "contract_address": self.contract_address,
            "vulnerability_type": self.vulnerability_type.value,
            "validated": self.validated,
            "profit_potential": self.profit_potential,
            "gas_used": self.gas_used,
            "error": self.error,
            "timestamp": self.timestamp,
        }

