"""
📋 Investigation data models
Defines the structure for Sift investigations and analyses
"""

from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional
from uuid import uuid4


class InvestigationStatus(str, Enum):
    """Investigation status"""

    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"


class AnalysisType(str, Enum):
    """Type of analysis"""

    ERROR_PATTERN = "error_pattern"
    SLOW_REQUEST = "slow_request"
    METRIC_ANOMALY = "metric_anomaly"
    LOG_ANOMALY = "log_anomaly"


@dataclass
class Analysis:
    """Analysis result within an investigation"""

    id: str = field(default_factory=lambda: str(uuid4()))
    type: AnalysisType = AnalysisType.ERROR_PATTERN
    status: InvestigationStatus = InvestigationStatus.PENDING
    start_time: datetime = field(default_factory=datetime.utcnow)
    end_time: Optional[datetime] = None
    result: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        return {
            "id": self.id,
            "type": self.type.value,
            "status": self.status.value,
            "start_time": self.start_time.isoformat(),
            "end_time": self.end_time.isoformat() if self.end_time else None,
            "result": self.result,
            "error": self.error,
            "metadata": self.metadata,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Analysis":
        """Create from dictionary"""
        return cls(
            id=data["id"],
            type=AnalysisType(data["type"]),
            status=InvestigationStatus(data["status"]),
            start_time=datetime.fromisoformat(data["start_time"]),
            end_time=datetime.fromisoformat(data["end_time"]) if data.get("end_time") else None,
            result=data.get("result"),
            error=data.get("error"),
            metadata=data.get("metadata", {}),
        )


@dataclass
class Investigation:
    """Sift investigation"""

    id: str = field(default_factory=lambda: str(uuid4()))
    name: str = "Investigation"
    labels: Dict[str, str] = field(default_factory=dict)
    start_time: datetime = field(default_factory=lambda: datetime.utcnow())
    end_time: Optional[datetime] = None
    status: InvestigationStatus = InvestigationStatus.PENDING
    analyses: List[Analysis] = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        return {
            "id": self.id,
            "name": self.name,
            "labels": self.labels,
            "start_time": self.start_time.isoformat(),
            "end_time": self.end_time.isoformat() if self.end_time else None,
            "status": self.status.value,
            "analyses": [analysis.to_dict() for analysis in self.analyses],
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
            "metadata": self.metadata,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Investigation":
        """Create from dictionary"""
        return cls(
            id=data["id"],
            name=data["name"],
            labels=data.get("labels", {}),
            start_time=datetime.fromisoformat(data["start_time"]),
            end_time=datetime.fromisoformat(data["end_time"]) if data.get("end_time") else None,
            status=InvestigationStatus(data["status"]),
            analyses=[Analysis.from_dict(a) for a in data.get("analyses", [])],
            created_at=datetime.fromisoformat(data["created_at"]),
            updated_at=datetime.fromisoformat(data["updated_at"]),
            metadata=data.get("metadata", {}),
        )

    def add_analysis(self, analysis: Analysis) -> None:
        """Add an analysis to the investigation"""
        self.analyses.append(analysis)
        self.updated_at = datetime.utcnow()

    def update_status(self, status: InvestigationStatus) -> None:
        """Update investigation status"""
        self.status = status
        self.updated_at = datetime.utcnow()
