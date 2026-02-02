"""Command Center event handler."""
import os
from datetime import datetime, timedelta
from typing import Any, Dict, Optional
from collections import defaultdict

import structlog
from cloudevents.http import CloudEvent

from shared.events import CloudEventEmitter, EventTypes
from shared.types import AlertSeverity, HeartbeatPayload
from shared import metrics

logger = structlog.get_logger()


class CommandCenterHandler:
    """
    Command Center - Central monitoring and control hub.
    
    Responsibilities:
    - Aggregate data from all edge locations
    - Track location health via heartbeats
    - Dispatch alerts based on thresholds
    - Push configuration to edge agents
    """
    
    def __init__(self):
        self.emitter = CloudEventEmitter(
            source="/agent-pos-edge/command-center"
        )
        
        # In-memory state (would be Redis/DB in production)
        self.locations: Dict[str, Dict[str, Any]] = {}
        self.alerts: Dict[str, Dict[str, Any]] = {}
        self.metrics_buffer: Dict[str, list] = defaultdict(list)
        
        # Configuration
        self.heartbeat_timeout_seconds = int(
            os.getenv("HEARTBEAT_TIMEOUT_SECONDS", "120")
        )
    
    async def handle(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle incoming CloudEvent."""
        event_type = event["type"]
        data = event.data or {}
        
        logger.info(
            "command_center_event_received",
            event_type=event_type,
            event_id=event["id"],
        )
        
        metrics.EVENTS_RECEIVED.labels(
            event_type=event_type,
            location_id=data.get("location_id", "unknown"),
            agent_role="command-center",
        ).inc()
        
        # Route to appropriate handler
        handlers = {
            EventTypes.LOCATION_HEARTBEAT: self._handle_heartbeat,
            EventTypes.TRANSACTION_COMPLETED: self._handle_transaction,
            EventTypes.TRANSACTION_FAILED: self._handle_transaction_failed,
            EventTypes.HEALTH_REPORT: self._handle_health_report,
            EventTypes.ALERT_RAISED: self._handle_alert,
            EventTypes.KITCHEN_QUEUE_STATUS: self._handle_kitchen_status,
            EventTypes.TANK_ALERT_LOW: self._handle_tank_alert,
            EventTypes.PUMP_STATUS: self._handle_pump_status,
        }
        
        handler = handlers.get(event_type)
        if handler:
            return await handler(data)
        
        logger.warning("unhandled_event_type", event_type=event_type)
        return {"status": "ignored", "reason": "unknown_event_type"}
    
    async def _handle_heartbeat(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process location heartbeat."""
        location_id = data.get("location_id")
        if not location_id:
            return {"status": "error", "reason": "missing_location_id"}
        
        now = datetime.utcnow()
        
        # Update location state
        self.locations[location_id] = {
            "last_heartbeat": now,
            "status": data.get("status", "unknown"),
            "type": data.get("location_type"),
            "pos_count": data.get("pos_count", 0),
            "pump_count": data.get("pump_count", 0),
        }
        
        # Update metrics
        metrics.LOCATION_HEARTBEAT_TIMESTAMP.labels(
            location_id=location_id
        ).set(now.timestamp())
        
        metrics.LOCATION_STATUS.labels(
            location_id=location_id,
            location_type=data.get("location_type", "unknown"),
        ).set(1 if data.get("status") == "healthy" else 0)
        
        logger.info(
            "heartbeat_processed",
            location_id=location_id,
            status=data.get("status"),
        )
        
        return {"status": "acknowledged", "location_id": location_id}
    
    async def _handle_transaction(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process completed transaction."""
        location_id = data.get("location_id")
        total = data.get("total", 0)
        payment_type = data.get("payment_type", "unknown")
        
        # Update metrics
        metrics.TRANSACTIONS_TOTAL.labels(
            location_id=location_id,
            status="completed",
            payment_type=payment_type,
        ).inc()
        
        metrics.TRANSACTION_VALUE.labels(
            location_id=location_id,
            payment_type=payment_type,
        ).inc(total)
        
        logger.info(
            "transaction_processed",
            location_id=location_id,
            transaction_id=data.get("transaction_id"),
            total=total,
        )
        
        return {"status": "processed"}
    
    async def _handle_transaction_failed(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process failed transaction - may trigger alert."""
        location_id = data.get("location_id")
        error = data.get("error", "unknown")
        
        metrics.TRANSACTIONS_TOTAL.labels(
            location_id=location_id,
            status="failed",
            payment_type=data.get("payment_type", "unknown"),
        ).inc()
        
        # Track failure rate - alert if too high
        # (simplified - would use sliding window in production)
        await self._raise_alert(
            location_id=location_id,
            severity=AlertSeverity.HIGH,
            alert_type="transaction_failure",
            message=f"Transaction failed: {error}",
            data=data,
        )
        
        return {"status": "processed", "alert_raised": True}
    
    async def _handle_health_report(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process health metrics from edge."""
        location_id = data.get("location_id") or data.get("pos_id", "").split("-")[0]
        
        # Check for concerning metrics
        cpu = data.get("cpu_percent", 0)
        memory = data.get("memory_percent", 0)
        disk = data.get("disk_percent", 0)
        
        if cpu > 90 or memory > 90:
            await self._raise_alert(
                location_id=location_id,
                severity=AlertSeverity.MEDIUM,
                alert_type="resource_high",
                message=f"High resource usage: CPU={cpu}%, Memory={memory}%",
                data=data,
            )
        
        if disk > 95:
            await self._raise_alert(
                location_id=location_id,
                severity=AlertSeverity.HIGH,
                alert_type="disk_full",
                message=f"Disk almost full: {disk}%",
                data=data,
            )
        
        return {"status": "processed"}
    
    async def _handle_alert(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process alert from edge agent."""
        alert_id = data.get("alert_id")
        location_id = data.get("location_id")
        severity = data.get("severity", "medium")
        
        # Store alert
        self.alerts[alert_id] = {
            **data,
            "received_at": datetime.utcnow(),
        }
        
        # Update metrics
        metrics.ALERTS_TOTAL.labels(
            location_id=location_id,
            severity=severity,
            alert_type=data.get("alert_type", "unknown"),
        ).inc()
        
        metrics.ACTIVE_ALERTS.labels(
            location_id=location_id,
            severity=severity,
        ).inc()
        
        # Dispatch alert (webhook, notification, etc.)
        await self.emitter.emit(
            event_type="pos.command.alert.dispatch",
            data={
                "alert_id": alert_id,
                "location_id": location_id,
                "severity": severity,
                "message": data.get("message"),
                "dispatched_at": datetime.utcnow().isoformat(),
            },
        )
        
        logger.warning(
            "alert_received",
            alert_id=alert_id,
            location_id=location_id,
            severity=severity,
            message=data.get("message"),
        )
        
        return {"status": "alert_dispatched", "alert_id": alert_id}
    
    async def _handle_kitchen_status(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process kitchen queue status."""
        location_id = data.get("location_id")
        queue_depth = data.get("queue_depth", 0)
        avg_wait = data.get("avg_wait_seconds", 0)
        
        # Update metrics
        metrics.KITCHEN_QUEUE_DEPTH.labels(
            location_id=location_id
        ).set(queue_depth)
        
        metrics.KITCHEN_AVG_WAIT.labels(
            location_id=location_id
        ).set(avg_wait)
        
        # Alert on high queue
        if queue_depth > 10:
            await self._raise_alert(
                location_id=location_id,
                severity=AlertSeverity.HIGH,
                alert_type="kitchen_queue_high",
                message=f"Kitchen queue depth: {queue_depth} orders",
                data=data,
            )
        
        return {"status": "processed"}
    
    async def _handle_tank_alert(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process low tank alert."""
        location_id = data.get("location_id")
        tank_id = data.get("tank_id")
        level = data.get("current_level", 0)
        
        await self._raise_alert(
            location_id=location_id,
            severity=AlertSeverity.HIGH if level < 10 else AlertSeverity.MEDIUM,
            alert_type="tank_low",
            message=f"Tank {tank_id} low: {level}%",
            data=data,
        )
        
        return {"status": "alert_raised"}
    
    async def _handle_pump_status(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process pump status change."""
        location_id = data.get("location_id")
        pump_id = data.get("pump_id")
        status = data.get("status")
        
        status_map = {
            "available": 1,
            "in_use": 2,
            "reserved": 3,
            "offline": 0,
            "maintenance": 0,
        }
        
        metrics.PUMP_STATUS_INFO.labels(
            location_id=location_id,
            pump_id=pump_id,
        ).set(status_map.get(status, 0))
        
        # Alert on unexpected offline
        if status == "offline":
            await self._raise_alert(
                location_id=location_id,
                severity=AlertSeverity.HIGH,
                alert_type="pump_offline",
                message=f"Pump {pump_id} went offline",
                data=data,
            )
        
        return {"status": "processed"}
    
    async def _raise_alert(
        self,
        location_id: str,
        severity: AlertSeverity,
        alert_type: str,
        message: str,
        data: Dict[str, Any],
    ) -> None:
        """Raise an alert."""
        import uuid
        alert_id = str(uuid.uuid4())
        
        await self.emitter.emit(
            event_type=EventTypes.ALERT_RAISED,
            data={
                "alert_id": alert_id,
                "location_id": location_id,
                "source": "command-center",
                "severity": severity.value,
                "alert_type": alert_type,
                "message": message,
                "data": data,
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
    
    async def check_stale_locations(self) -> list:
        """Check for locations that haven't sent heartbeat."""
        stale = []
        now = datetime.utcnow()
        timeout = timedelta(seconds=self.heartbeat_timeout_seconds)
        
        for location_id, info in self.locations.items():
            last_heartbeat = info.get("last_heartbeat")
            if last_heartbeat and (now - last_heartbeat) > timeout:
                stale.append(location_id)
                
                await self.emitter.emit(
                    event_type=EventTypes.LOCATION_OFFLINE,
                    data={
                        "location_id": location_id,
                        "last_seen": last_heartbeat.isoformat(),
                        "timeout_seconds": self.heartbeat_timeout_seconds,
                    },
                )
        
        return stale
