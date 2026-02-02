"""POS Edge event handler - runs on each POS terminal."""
import os
import asyncio
from datetime import datetime
from typing import Any, Dict, Optional
from collections import deque

import structlog
from cloudevents.http import CloudEvent

from shared.events import CloudEventEmitter, EventTypes
from shared.types import (
    AlertSeverity,
    HealthMetrics,
    Transaction,
    TransactionStatus,
)
from shared import metrics

logger = structlog.get_logger()


class POSEdgeHandler:
    """
    POS Edge Agent - Lightweight agent running on POS terminals.
    
    Responsibilities:
    - Send heartbeats to command center
    - Monitor system health
    - Track transactions
    - Buffer events when offline
    - Forward events to central broker
    """
    
    def __init__(self):
        self.location_id = os.getenv("LOCATION_ID", "unknown")
        self.pos_id = os.getenv("POS_ID", "pos-01")
        
        self.emitter = CloudEventEmitter(
            source=f"/agent-pos-edge/pos-edge/{self.location_id}/{self.pos_id}"
        )
        
        # Offline buffer
        self.offline_buffer: deque = deque(
            maxlen=int(os.getenv("OFFLINE_BUFFER_MAX_EVENTS", "1000"))
        )
        self.is_online = True
        
        # Configuration
        self.heartbeat_interval = int(
            os.getenv("HEARTBEAT_INTERVAL_SECONDS", "30")
        )
        self.health_check_interval = int(
            os.getenv("HEALTH_CHECK_INTERVAL_SECONDS", "60")
        )
        
        # Current transaction
        self.current_transaction: Optional[Transaction] = None
    
    async def handle(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle incoming CloudEvent (from command center)."""
        event_type = event["type"]
        data = event.data or {}
        
        logger.info(
            "pos_edge_event_received",
            event_type=event_type,
            event_id=event["id"],
        )
        
        metrics.EVENTS_RECEIVED.labels(
            event_type=event_type,
            location_id=self.location_id,
            agent_role="pos-edge",
        ).inc()
        
        handlers = {
            EventTypes.COMMAND_CONFIG_PUSH: self._handle_config_push,
            EventTypes.COMMAND_MAINTENANCE_SCHEDULE: self._handle_maintenance,
        }
        
        handler = handlers.get(event_type)
        if handler:
            return await handler(data)
        
        return {"status": "ignored"}
    
    async def _handle_config_push(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle configuration update from command center."""
        target_locations = data.get("target_locations", [])
        
        if "all" in target_locations or self.location_id in target_locations:
            config = data.get("config", {})
            
            # Apply configuration (simplified)
            if "heartbeat_interval" in config:
                self.heartbeat_interval = config["heartbeat_interval"]
            
            logger.info(
                "config_updated",
                location_id=self.location_id,
                config_version=data.get("version"),
            )
            
            return {"status": "config_applied"}
        
        return {"status": "ignored", "reason": "not_target"}
    
    async def _handle_maintenance(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle maintenance schedule notification."""
        logger.info(
            "maintenance_scheduled",
            location_id=self.location_id,
            datetime=data.get("datetime"),
            type=data.get("type"),
        )
        return {"status": "acknowledged"}
    
    # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    # Outbound Events (to Command Center)
    # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    async def send_heartbeat(self) -> bool:
        """Send heartbeat to command center."""
        return await self._emit_event(
            event_type=EventTypes.LOCATION_HEARTBEAT,
            data={
                "location_id": self.location_id,
                "pos_id": self.pos_id,
                "status": "healthy",
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
    
    async def send_health_report(self, health: HealthMetrics) -> bool:
        """Send health metrics."""
        return await self._emit_event(
            event_type=EventTypes.HEALTH_REPORT,
            data={
                "location_id": self.location_id,
                "pos_id": self.pos_id,
                "cpu_percent": health.cpu_percent,
                "memory_percent": health.memory_percent,
                "disk_percent": health.disk_percent,
                "network_up": health.network_up,
                "timestamp": health.timestamp.isoformat(),
            },
        )
    
    async def start_transaction(
        self,
        transaction_id: str,
        items: list,
    ) -> bool:
        """Start a new transaction."""
        now = datetime.utcnow()
        
        self.current_transaction = Transaction(
            transaction_id=transaction_id,
            pos_id=self.pos_id,
            location_id=self.location_id,
            status=TransactionStatus.STARTED,
            items=items,
            started_at=now,
        )
        
        return await self._emit_event(
            event_type=EventTypes.TRANSACTION_STARTED,
            data={
                "transaction_id": transaction_id,
                "pos_id": self.pos_id,
                "location_id": self.location_id,
                "items": [item.dict() for item in items],
                "started_at": now.isoformat(),
            },
        )
    
    async def complete_transaction(
        self,
        total: float,
        payment_type: str,
    ) -> bool:
        """Complete current transaction."""
        if not self.current_transaction:
            return False
        
        now = datetime.utcnow()
        self.current_transaction.status = TransactionStatus.COMPLETED
        self.current_transaction.total = total
        self.current_transaction.payment_type = payment_type
        self.current_transaction.completed_at = now
        
        # Calculate duration
        duration = (now - self.current_transaction.started_at).total_seconds()
        
        metrics.TRANSACTIONS_TOTAL.labels(
            location_id=self.location_id,
            status="completed",
            payment_type=payment_type,
        ).inc()
        
        metrics.TRANSACTION_DURATION.labels(
            location_id=self.location_id
        ).observe(duration)
        
        success = await self._emit_event(
            event_type=EventTypes.TRANSACTION_COMPLETED,
            data={
                "transaction_id": self.current_transaction.transaction_id,
                "pos_id": self.pos_id,
                "location_id": self.location_id,
                "total": total,
                "payment_type": payment_type,
                "items": [item.dict() for item in self.current_transaction.items],
                "started_at": self.current_transaction.started_at.isoformat(),
                "completed_at": now.isoformat(),
                "duration_seconds": duration,
            },
        )
        
        self.current_transaction = None
        return success
    
    async def fail_transaction(self, error: str) -> bool:
        """Mark current transaction as failed."""
        if not self.current_transaction:
            return False
        
        self.current_transaction.status = TransactionStatus.FAILED
        self.current_transaction.error = error
        
        metrics.TRANSACTIONS_TOTAL.labels(
            location_id=self.location_id,
            status="failed",
            payment_type="unknown",
        ).inc()
        
        success = await self._emit_event(
            event_type=EventTypes.TRANSACTION_FAILED,
            data={
                "transaction_id": self.current_transaction.transaction_id,
                "pos_id": self.pos_id,
                "location_id": self.location_id,
                "error": error,
                "started_at": self.current_transaction.started_at.isoformat(),
            },
        )
        
        self.current_transaction = None
        return success
    
    async def raise_alert(
        self,
        severity: AlertSeverity,
        alert_type: str,
        message: str,
    ) -> bool:
        """Raise an alert."""
        import uuid
        
        return await self._emit_event(
            event_type=EventTypes.ALERT_RAISED,
            data={
                "alert_id": str(uuid.uuid4()),
                "location_id": self.location_id,
                "pos_id": self.pos_id,
                "source": "pos-edge",
                "severity": severity.value,
                "alert_type": alert_type,
                "message": message,
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
    
    async def _emit_event(
        self,
        event_type: str,
        data: Dict[str, Any],
    ) -> bool:
        """Emit event, buffer if offline."""
        if self.is_online:
            success = await self.emitter.emit(event_type, data)
            
            if success:
                metrics.EVENTS_EMITTED.labels(
                    event_type=event_type,
                    location_id=self.location_id,
                    agent_role="pos-edge",
                ).inc()
                return True
            
            # Failed to send - mark offline and buffer
            self.is_online = False
        
        # Buffer event for later
        self.offline_buffer.append({
            "event_type": event_type,
            "data": data,
            "buffered_at": datetime.utcnow().isoformat(),
        })
        
        logger.warning(
            "event_buffered",
            event_type=event_type,
            buffer_size=len(self.offline_buffer),
        )
        
        return False
    
    async def flush_buffer(self) -> int:
        """Try to send buffered events."""
        sent = 0
        
        while self.offline_buffer:
            event = self.offline_buffer[0]
            success = await self.emitter.emit(
                event["event_type"],
                event["data"],
            )
            
            if success:
                self.offline_buffer.popleft()
                sent += 1
            else:
                break
        
        if sent > 0:
            logger.info("buffer_flushed", events_sent=sent)
            self.is_online = True
        
        return sent
