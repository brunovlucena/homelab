"""Pump Agent handler - monitors gas station fuel operations."""
import os
from datetime import datetime, timedelta
from typing import Any, Dict, Optional

import structlog
from cloudevents.http import CloudEvent

from shared.events import CloudEventEmitter, EventTypes
from shared.types import (
    AlertSeverity,
    PumpStatus,
    PumpTransaction,
    TankLevel,
    PumpInfo,
)
from shared import metrics

logger = structlog.get_logger()


class PumpAgentHandler:
    """
    Pump Agent - Monitors gas station fuel operations.
    
    Responsibilities:
    - Monitor pump status
    - Track fuel transactions
    - Monitor tank levels
    - Predict refill needs
    - Safety alerts
    """
    
    def __init__(self):
        self.location_id = os.getenv("LOCATION_ID", "unknown")
        
        self.emitter = CloudEventEmitter(
            source=f"/agent-pos-edge/pump/{self.location_id}"
        )
        
        # Pumps
        self.pumps: Dict[str, PumpInfo] = {}
        
        # Tanks
        self.tanks: Dict[str, TankLevel] = {}
        
        # Active transactions
        self.active_transactions: Dict[str, PumpTransaction] = {}
        
        # Consumption history for predictions
        self.consumption_history: Dict[str, list] = {}
        
        # Config
        self.tank_low_threshold = int(
            os.getenv("TANK_LOW_THRESHOLD_PERCENT", "20")
        )
        self.tank_critical_threshold = int(
            os.getenv("TANK_CRITICAL_THRESHOLD_PERCENT", "10")
        )
    
    async def handle(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle incoming CloudEvent."""
        event_type = event["type"]
        data = event.data or {}
        
        logger.info("pump_event_received", event_type=event_type)
        
        metrics.EVENTS_RECEIVED.labels(
            event_type=event_type,
            location_id=self.location_id,
            agent_role="pump",
        ).inc()
        
        handlers = {
            EventTypes.TRANSACTION_STARTED: self._handle_prepay,
            "pos.pump.sensor.reading": self._handle_pump_sensor,
            "pos.tank.sensor.reading": self._handle_tank_sensor,
            EventTypes.COMMAND_CONFIG_PUSH: self._handle_config,
            EventTypes.COMMAND_MAINTENANCE_SCHEDULE: self._handle_maintenance,
        }
        
        handler = handlers.get(event_type)
        if handler:
            return await handler(data)
        
        return {"status": "ignored"}
    
    async def _handle_prepay(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle pre-pay transaction to activate pump."""
        # Check if any items are fuel
        items = data.get("items", [])
        fuel_items = [i for i in items if "fuel" in i.get("name", "").lower()]
        
        if not fuel_items:
            return {"status": "ignored", "reason": "not_fuel"}
        
        pump_id = data.get("pump_id")
        if pump_id:
            await self.reserve_pump(pump_id)
        
        return {"status": "pump_reserved", "pump_id": pump_id}
    
    async def _handle_pump_sensor(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle pump sensor data."""
        pump_id = data.get("pump_id")
        status = data.get("status")
        flow_rate = data.get("flow_rate", 0)
        
        # Update pump info
        if pump_id not in self.pumps:
            self.pumps[pump_id] = PumpInfo(
                pump_id=pump_id,
                location_id=self.location_id,
                status=PumpStatus(status),
                fuel_types=data.get("fuel_types", ["regular"]),
            )
        else:
            old_status = self.pumps[pump_id].status
            self.pumps[pump_id].status = PumpStatus(status)
            
            # Detect status transitions
            if old_status != PumpStatus(status):
                await self._handle_pump_status_change(pump_id, old_status, PumpStatus(status))
        
        # Update flow during transaction
        if pump_id in self.active_transactions and flow_rate > 0:
            txn = self.active_transactions[pump_id]
            # Update liters (simplified)
            txn.liters += flow_rate * 1  # Assume 1 second interval
        
        return {"status": "processed"}
    
    async def _handle_tank_sensor(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle tank level sensor data."""
        tank_id = data.get("tank_id")
        level = data.get("level")
        capacity = data.get("capacity", 20000)
        fuel_type = data.get("fuel_type", "regular")
        
        level_percent = (level / capacity) * 100 if capacity > 0 else 0
        
        self.tanks[tank_id] = TankLevel(
            tank_id=tank_id,
            location_id=self.location_id,
            fuel_type=fuel_type,
            current_level=level,
            capacity=capacity,
            level_percent=level_percent,
        )
        
        # Update metrics
        metrics.TANK_LEVEL_PERCENT.labels(
            location_id=self.location_id,
            tank_id=tank_id,
            fuel_type=fuel_type,
        ).set(level_percent)
        
        # Emit tank level event
        await self.emitter.emit(
            event_type=EventTypes.TANK_LEVEL,
            data={
                "tank_id": tank_id,
                "location_id": self.location_id,
                "fuel_type": fuel_type,
                "current_level": level,
                "capacity": capacity,
                "level_percent": level_percent,
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
        
        # Check thresholds
        if level_percent < self.tank_critical_threshold:
            await self._raise_tank_alert(tank_id, fuel_type, level_percent, "critical")
        elif level_percent < self.tank_low_threshold:
            await self._raise_tank_alert(tank_id, fuel_type, level_percent, "low")
        
        return {"status": "processed"}
    
    async def _handle_config(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle configuration update."""
        config = data.get("config", {})
        
        if "tank_low_threshold" in config:
            self.tank_low_threshold = config["tank_low_threshold"]
        if "tank_critical_threshold" in config:
            self.tank_critical_threshold = config["tank_critical_threshold"]
        
        return {"status": "config_applied"}
    
    async def _handle_maintenance(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle scheduled maintenance for pump."""
        pump_id = data.get("pump_id")
        
        if pump_id and pump_id in self.pumps:
            self.pumps[pump_id].status = PumpStatus.MAINTENANCE
            
            await self.emitter.emit(
                event_type=EventTypes.PUMP_STATUS,
                data={
                    "pump_id": pump_id,
                    "location_id": self.location_id,
                    "status": "maintenance",
                    "reason": data.get("type", "scheduled"),
                    "timestamp": datetime.utcnow().isoformat(),
                },
            )
        
        return {"status": "maintenance_set"}
    
    # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    # Pump Operations
    # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    async def reserve_pump(self, pump_id: str) -> bool:
        """Reserve pump for pre-pay."""
        if pump_id not in self.pumps:
            self.pumps[pump_id] = PumpInfo(
                pump_id=pump_id,
                location_id=self.location_id,
                status=PumpStatus.RESERVED,
                fuel_types=["regular"],
            )
        
        self.pumps[pump_id].status = PumpStatus.RESERVED
        
        await self._emit_pump_status(pump_id, "reserved")
        return True
    
    async def start_pumping(
        self,
        pump_id: str,
        fuel_type: str,
    ) -> bool:
        """Start fuel dispensing."""
        if pump_id not in self.pumps:
            return False
        
        now = datetime.utcnow()
        
        self.pumps[pump_id].status = PumpStatus.IN_USE
        
        txn = PumpTransaction(
            pump_id=pump_id,
            location_id=self.location_id,
            fuel_type=fuel_type,
            liters=0,
            total=0,
            started_at=now,
        )
        
        self.active_transactions[pump_id] = txn
        self.pumps[pump_id].current_transaction = txn
        
        await self.emitter.emit(
            event_type=EventTypes.PUMP_TRANSACTION_START,
            data={
                "pump_id": pump_id,
                "location_id": self.location_id,
                "fuel_type": fuel_type,
                "started_at": now.isoformat(),
            },
        )
        
        await self._emit_pump_status(pump_id, "in_use")
        return True
    
    async def end_pumping(
        self,
        pump_id: str,
        price_per_liter: float = 5.0,
    ) -> Optional[PumpTransaction]:
        """End fuel dispensing."""
        if pump_id not in self.active_transactions:
            return None
        
        txn = self.active_transactions[pump_id]
        txn.ended_at = datetime.utcnow()
        txn.total = txn.liters * price_per_liter
        
        # Update metrics
        metrics.PUMP_TRANSACTIONS_TOTAL.labels(
            location_id=self.location_id,
            pump_id=pump_id,
            fuel_type=txn.fuel_type,
        ).inc()
        
        metrics.PUMP_LITERS_TOTAL.labels(
            location_id=self.location_id,
            fuel_type=txn.fuel_type,
        ).inc(txn.liters)
        
        # Track consumption for predictions
        if txn.fuel_type not in self.consumption_history:
            self.consumption_history[txn.fuel_type] = []
        self.consumption_history[txn.fuel_type].append({
            "liters": txn.liters,
            "timestamp": txn.ended_at.isoformat(),
        })
        
        await self.emitter.emit(
            event_type=EventTypes.PUMP_TRANSACTION_END,
            data={
                "pump_id": pump_id,
                "location_id": self.location_id,
                "fuel_type": txn.fuel_type,
                "liters": txn.liters,
                "total": txn.total,
                "started_at": txn.started_at.isoformat(),
                "ended_at": txn.ended_at.isoformat(),
            },
        )
        
        # Reset pump
        del self.active_transactions[pump_id]
        self.pumps[pump_id].status = PumpStatus.AVAILABLE
        self.pumps[pump_id].current_transaction = None
        
        await self._emit_pump_status(pump_id, "available")
        
        return txn
    
    async def _handle_pump_status_change(
        self,
        pump_id: str,
        old_status: PumpStatus,
        new_status: PumpStatus,
    ) -> None:
        """Handle pump status transitions."""
        if new_status == PumpStatus.OFFLINE and old_status != PumpStatus.MAINTENANCE:
            # Unexpected offline
            await self.emitter.emit(
                event_type=EventTypes.ALERT_RAISED,
                data={
                    "alert_id": f"pump-offline-{pump_id}-{datetime.utcnow().timestamp()}",
                    "location_id": self.location_id,
                    "source": "pump",
                    "severity": AlertSeverity.HIGH.value,
                    "alert_type": "pump_offline",
                    "message": f"Pump {pump_id} went offline unexpectedly",
                    "data": {"pump_id": pump_id},
                    "timestamp": datetime.utcnow().isoformat(),
                },
            )
    
    async def _emit_pump_status(self, pump_id: str, status: str) -> None:
        """Emit pump status event."""
        await self.emitter.emit(
            event_type=EventTypes.PUMP_STATUS,
            data={
                "pump_id": pump_id,
                "location_id": self.location_id,
                "status": status,
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
        
        status_map = {
            "available": 1,
            "in_use": 2,
            "reserved": 3,
            "offline": 0,
            "maintenance": 0,
        }
        
        metrics.PUMP_STATUS_INFO.labels(
            location_id=self.location_id,
            pump_id=pump_id,
        ).set(status_map.get(status, 0))
    
    async def _raise_tank_alert(
        self,
        tank_id: str,
        fuel_type: str,
        level_percent: float,
        severity_type: str,
    ) -> None:
        """Raise tank level alert."""
        severity = AlertSeverity.CRITICAL if severity_type == "critical" else AlertSeverity.MEDIUM
        
        await self.emitter.emit(
            event_type=EventTypes.TANK_ALERT_LOW,
            data={
                "tank_id": tank_id,
                "location_id": self.location_id,
                "fuel_type": fuel_type,
                "current_level": level_percent,
                "severity": severity.value,
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
    
    def predict_refill(self, tank_id: str) -> Optional[Dict[str, Any]]:
        """Predict when tank will need refill."""
        if tank_id not in self.tanks:
            return None
        
        tank = self.tanks[tank_id]
        fuel_type = tank.fuel_type
        
        if fuel_type not in self.consumption_history:
            return None
        
        history = self.consumption_history[fuel_type]
        if len(history) < 10:
            return None
        
        # Calculate average daily consumption (simplified)
        total_liters = sum(h["liters"] for h in history[-50:])
        avg_daily = total_liters / 7  # Assume last 50 txns over 7 days
        
        liters_until_low = tank.current_level - (tank.capacity * self.tank_low_threshold / 100)
        
        if avg_daily > 0:
            days_until_low = liters_until_low / avg_daily
            refill_date = datetime.utcnow() + timedelta(days=days_until_low)
            
            return {
                "tank_id": tank_id,
                "fuel_type": fuel_type,
                "days_until_low": days_until_low,
                "recommended_refill_date": refill_date.isoformat(),
                "avg_daily_consumption": avg_daily,
            }
        
        return None
