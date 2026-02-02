"""
Shared types for agent-redteam.
"""
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional
from datetime import datetime, timezone


class ExploitSeverity(str, Enum):
    """Exploit severity levels."""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"


class ExploitCategory(str, Enum):
    """Exploit category classification."""
    SSRF = "ssrf"
    COMMAND_INJECTION = "command_injection"
    CODE_INJECTION = "code_injection"
    TEMPLATE_INJECTION = "template_injection"
    PATH_TRAVERSAL = "path_traversal"
    RBAC_ESCALATION = "rbac_escalation"
    TOKEN_EXPOSURE = "token_exposure"
    RECEIVER_ESCALATION = "receiver_escalation"
    OTHER = "other"


class ExploitStatus(str, Enum):
    """Exploit execution status."""
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    BLOCKED = "blocked"  # Exploit was mitigated
    TIMEOUT = "timeout"
    ERROR = "error"


class TargetComponent(str, Enum):
    """Target component for exploit."""
    LAMBDA_OPERATOR = "lambda_operator"
    BUILD_POD = "build_pod"
    GIT_CLONE = "git_clone"
    MINIO = "minio"
    KUBERNETES_API = "kubernetes_api"
    SERVICE_ACCOUNT = "service_account"
    OTHER = "other"


@dataclass
class ExploitDefinition:
    """Definition of an exploit to run."""
    id: str  # e.g., "vuln-001"
    name: str
    description: str
    severity: ExploitSeverity
    category: ExploitCategory
    target_component: TargetComponent
    manifest_path: str  # Path to the exploit YAML
    expected_outcome: str  # What success looks like
    detection_signature: Optional[str] = None
    prerequisites: list[str] = field(default_factory=list)
    tags: list[str] = field(default_factory=list)
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "severity": self.severity.value,
            "category": self.category.value,
            "target_component": self.target_component.value,
            "manifest_path": self.manifest_path,
            "expected_outcome": self.expected_outcome,
            "detection_signature": self.detection_signature,
            "prerequisites": self.prerequisites,
            "tags": self.tags,
        }


@dataclass
class ExploitResult:
    """Result of an exploit execution."""
    exploit_id: str
    status: ExploitStatus
    started_at: str = ""
    completed_at: str = ""
    duration_seconds: float = 0.0
    output: str = ""
    error: Optional[str] = None
    artifacts: list[str] = field(default_factory=list)  # Captured data/tokens
    mitigated_by: Optional[str] = None  # What blocked the exploit
    
    def __post_init__(self):
        if not self.started_at:
            self.started_at = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "exploit_id": self.exploit_id,
            "status": self.status.value,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
            "duration_seconds": self.duration_seconds,
            "output": self.output,
            "error": self.error,
            "artifacts": self.artifacts,
            "mitigated_by": self.mitigated_by,
        }


@dataclass
class TestRun:
    """A complete test run of multiple exploits."""
    id: str
    name: str
    target_cluster: str
    target_namespace: str
    started_at: str = ""
    completed_at: str = ""
    results: list[ExploitResult] = field(default_factory=list)
    status: str = "pending"  # pending, running, completed, failed
    
    def __post_init__(self):
        if not self.started_at:
            self.started_at = datetime.now(timezone.utc).isoformat()
    
    @property
    def total_exploits(self) -> int:
        return len(self.results)
    
    @property
    def successful_exploits(self) -> int:
        return len([r for r in self.results if r.status == ExploitStatus.SUCCESS])
    
    @property
    def blocked_exploits(self) -> int:
        return len([r for r in self.results if r.status == ExploitStatus.BLOCKED])
    
    @property
    def failed_exploits(self) -> int:
        return len([r for r in self.results if r.status in (ExploitStatus.FAILED, ExploitStatus.ERROR)])
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "name": self.name,
            "target_cluster": self.target_cluster,
            "target_namespace": self.target_namespace,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
            "status": self.status,
            "total_exploits": self.total_exploits,
            "successful_exploits": self.successful_exploits,
            "blocked_exploits": self.blocked_exploits,
            "failed_exploits": self.failed_exploits,
            "results": [r.to_dict() for r in self.results],
        }


@dataclass
class ExploitCatalog:
    """Catalog of all available exploits."""
    exploits: list[ExploitDefinition] = field(default_factory=list)
    version: str = "1.0.0"
    
    def get_by_id(self, exploit_id: str) -> Optional[ExploitDefinition]:
        for exploit in self.exploits:
            if exploit.id == exploit_id:
                return exploit
        return None
    
    def get_by_category(self, category: ExploitCategory) -> list[ExploitDefinition]:
        return [e for e in self.exploits if e.category == category]
    
    def get_by_severity(self, severity: ExploitSeverity) -> list[ExploitDefinition]:
        return [e for e in self.exploits if e.severity == severity]
    
    def to_dict(self) -> dict:
        return {
            "version": self.version,
            "exploits": [e.to_dict() for e in self.exploits],
        }
