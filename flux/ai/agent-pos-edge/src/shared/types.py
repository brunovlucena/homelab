"""Shared types for POS Edge agents."""
from datetime import datetime
from enum import Enum
from typing import Optional, Dict, Any, List
from pydantic import BaseModel, Field


class LocationType(str, Enum):
    GAS_STATION = "gas_station"
    FAST_FOOD = "fast_food"
    RETAIL = "retail"


class AlertSeverity(str, Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class PumpStatus(str, Enum):
    AVAILABLE = "available"
    IN_USE = "in_use"
    RESERVED = "reserved"
    OFFLINE = "offline"
    MAINTENANCE = "maintenance"


class TransactionStatus(str, Enum):
    STARTED = "started"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Base Models
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class Location(BaseModel):
    """Location information."""
    location_id: str
    name: str
    type: LocationType
    address: Optional[str] = None
    timezone: str = "UTC"


class HealthMetrics(BaseModel):
    """System health metrics."""
    cpu_percent: float
    memory_percent: float
    disk_percent: float
    network_up: bool
    timestamp: datetime = Field(default_factory=datetime.utcnow)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Transaction Models
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class TransactionItem(BaseModel):
    """Item in a transaction."""
    name: str
    quantity: float
    unit: str = "unit"
    price: float


class Transaction(BaseModel):
    """POS transaction."""
    transaction_id: str
    pos_id: str
    location_id: str
    status: TransactionStatus
    items: List[TransactionItem] = []
    total: float = 0.0
    payment_type: Optional[str] = None
    started_at: datetime
    completed_at: Optional[datetime] = None
    error: Optional[str] = None


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Kitchen Models (Fast-Food)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class KitchenOrder(BaseModel):
    """Kitchen order."""
    order_id: str
    location_id: str
    items: List[str]
    priority: int = 0
    station: Optional[str] = None
    received_at: datetime
    started_at: Optional[datetime] = None
    ready_at: Optional[datetime] = None
    estimated_time_seconds: int = 180


class KitchenQueueStatus(BaseModel):
    """Kitchen queue status."""
    location_id: str
    queue_depth: int
    avg_wait_seconds: float
    orders_in_progress: int
    stations: Dict[str, int] = {}  # station -> order count
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class EquipmentStatus(BaseModel):
    """Kitchen equipment status."""
    equipment_id: str
    location_id: str
    name: str
    status: str  # online, offline, warning
    temperature: Optional[float] = None
    last_check: datetime = Field(default_factory=datetime.utcnow)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Pump Models (Gas Station)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class PumpTransaction(BaseModel):
    """Fuel pump transaction."""
    pump_id: str
    location_id: str
    fuel_type: str
    liters: float
    total: float
    started_at: datetime
    ended_at: Optional[datetime] = None


class TankLevel(BaseModel):
    """Fuel tank level."""
    tank_id: str
    location_id: str
    fuel_type: str
    current_level: float  # liters
    capacity: float  # liters
    level_percent: float
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class PumpInfo(BaseModel):
    """Pump information."""
    pump_id: str
    location_id: str
    status: PumpStatus
    fuel_types: List[str]
    current_transaction: Optional[PumpTransaction] = None


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Alert Models
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class Alert(BaseModel):
    """Alert from edge agent."""
    alert_id: str
    location_id: str
    source: str  # pos-edge, kitchen, pump
    severity: AlertSeverity
    alert_type: str
    message: str
    data: Dict[str, Any] = {}
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    acknowledged: bool = False
    acknowledged_by: Optional[str] = None


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Event Payloads
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

class HeartbeatPayload(BaseModel):
    """Location heartbeat payload."""
    location_id: str
    location_type: LocationType
    status: str
    pos_count: int
    pump_count: int = 0
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ConfigPushPayload(BaseModel):
    """Configuration push payload."""
    target_locations: List[str]  # location IDs or 'all'
    config: Dict[str, Any]
    version: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
