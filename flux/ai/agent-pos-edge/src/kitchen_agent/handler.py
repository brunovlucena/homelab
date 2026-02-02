"""Kitchen Agent handler - monitors kitchen operations."""
import os
from datetime import datetime
from typing import Any, Dict, List, Optional
from collections import deque

import structlog
from cloudevents.http import CloudEvent

from shared.events import CloudEventEmitter, EventTypes
from shared.types import (
    AlertSeverity,
    KitchenOrder,
    KitchenQueueStatus,
    EquipmentStatus,
)
from shared import metrics

logger = structlog.get_logger()


class KitchenAgentHandler:
    """
    Kitchen Agent - Monitors fast-food kitchen operations.
    
    Responsibilities:
    - Track order queue
    - Monitor Kitchen Display System (KDS)
    - Track equipment status
    - Calculate wait times
    - Detect bottlenecks
    """
    
    def __init__(self):
        self.location_id = os.getenv("LOCATION_ID", "unknown")
        
        self.emitter = CloudEventEmitter(
            source=f"/agent-pos-edge/kitchen/{self.location_id}"
        )
        
        # Order queue
        self.orders: Dict[str, KitchenOrder] = {}
        self.order_queue: deque = deque()
        
        # Equipment
        self.equipment: Dict[str, EquipmentStatus] = {}
        
        # Config
        self.queue_alert_threshold = int(
            os.getenv("QUEUE_ALERT_THRESHOLD", "10")
        )
        self.wait_time_alert_seconds = int(
            os.getenv("WAIT_TIME_ALERT_SECONDS", "300")
        )
    
    async def handle(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle incoming CloudEvent."""
        event_type = event["type"]
        data = event.data or {}
        
        logger.info("kitchen_event_received", event_type=event_type)
        
        metrics.EVENTS_RECEIVED.labels(
            event_type=event_type,
            location_id=self.location_id,
            agent_role="kitchen",
        ).inc()
        
        handlers = {
            EventTypes.TRANSACTION_COMPLETED: self._handle_new_order,
            "pos.equipment.temperature": self._handle_equipment_temp,
            "pos.equipment.status": self._handle_equipment_status,
            EventTypes.COMMAND_CONFIG_PUSH: self._handle_config_push,
        }
        
        handler = handlers.get(event_type)
        if handler:
            return await handler(data)
        
        return {"status": "ignored"}
    
    async def _handle_new_order(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle new order from POS."""
        order_id = data.get("transaction_id")
        items = data.get("items", [])
        
        order = KitchenOrder(
            order_id=order_id,
            location_id=self.location_id,
            items=[item.get("name", "") for item in items],
            received_at=datetime.utcnow(),
        )
        
        self.orders[order_id] = order
        self.order_queue.append(order_id)
        
        # Emit order received event
        await self.emitter.emit(
            event_type=EventTypes.KITCHEN_ORDER_RECEIVED,
            data={
                "order_id": order_id,
                "location_id": self.location_id,
                "items": order.items,
                "queue_position": len(self.order_queue),
                "received_at": order.received_at.isoformat(),
            },
        )
        
        # Update queue metrics
        await self._update_queue_status()
        
        metrics.KITCHEN_ORDERS_TOTAL.labels(
            location_id=self.location_id,
            status="received",
        ).inc()
        
        return {"status": "order_queued", "order_id": order_id}
    
    async def _handle_equipment_temp(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle equipment temperature reading."""
        equipment_id = data.get("equipment_id")
        temperature = data.get("temperature")
        
        if equipment_id in self.equipment:
            self.equipment[equipment_id].temperature = temperature
            
            # Check temperature thresholds
            min_temp = data.get("min_threshold", 0)
            max_temp = data.get("max_threshold", 200)
            
            if temperature < min_temp or temperature > max_temp:
                await self._raise_equipment_alert(
                    equipment_id=equipment_id,
                    alert_type="temperature_out_of_range",
                    message=f"Temperature {temperature}°C out of range [{min_temp}-{max_temp}]",
                )
        
        return {"status": "processed"}
    
    async def _handle_equipment_status(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle equipment status change."""
        equipment_id = data.get("equipment_id")
        status = data.get("status")
        
        self.equipment[equipment_id] = EquipmentStatus(
            equipment_id=equipment_id,
            location_id=self.location_id,
            name=data.get("name", equipment_id),
            status=status,
        )
        
        if status == "offline":
            await self._raise_equipment_alert(
                equipment_id=equipment_id,
                alert_type="equipment_offline",
                message=f"Equipment {data.get('name', equipment_id)} went offline",
            )
        
        return {"status": "updated"}
    
    async def _handle_config_push(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Handle configuration update."""
        config = data.get("config", {})
        
        if "queue_alert_threshold" in config:
            self.queue_alert_threshold = config["queue_alert_threshold"]
        if "wait_time_alert_seconds" in config:
            self.wait_time_alert_seconds = config["wait_time_alert_seconds"]
        
        return {"status": "config_applied"}
    
    # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    # Kitchen Operations
    # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    async def start_order(self, order_id: str, station: str) -> bool:
        """Mark order as being prepared."""
        if order_id not in self.orders:
            return False
        
        order = self.orders[order_id]
        order.started_at = datetime.utcnow()
        order.station = station
        
        await self.emitter.emit(
            event_type=EventTypes.KITCHEN_ORDER_STARTED,
            data={
                "order_id": order_id,
                "location_id": self.location_id,
                "station": station,
                "started_at": order.started_at.isoformat(),
            },
        )
        
        return True
    
    async def complete_order(self, order_id: str) -> bool:
        """Mark order as ready."""
        if order_id not in self.orders:
            return False
        
        order = self.orders[order_id]
        order.ready_at = datetime.utcnow()
        
        # Calculate prep time
        prep_time = 0
        if order.started_at:
            prep_time = (order.ready_at - order.started_at).total_seconds()
        
        metrics.KITCHEN_ORDER_DURATION.labels(
            location_id=self.location_id,
            station=order.station or "unknown",
        ).observe(prep_time)
        
        metrics.KITCHEN_ORDERS_TOTAL.labels(
            location_id=self.location_id,
            status="completed",
        ).inc()
        
        await self.emitter.emit(
            event_type=EventTypes.KITCHEN_ORDER_READY,
            data={
                "order_id": order_id,
                "location_id": self.location_id,
                "prep_time_seconds": prep_time,
                "ready_at": order.ready_at.isoformat(),
            },
        )
        
        # Remove from queue
        if order_id in self.order_queue:
            self.order_queue.remove(order_id)
        del self.orders[order_id]
        
        await self._update_queue_status()
        
        return True
    
    async def _update_queue_status(self) -> None:
        """Update and emit queue status."""
        queue_depth = len(self.order_queue)
        
        # Calculate average wait time
        avg_wait = 0
        if self.order_queue:
            now = datetime.utcnow()
            wait_times = []
            for order_id in self.order_queue:
                if order_id in self.orders:
                    wait = (now - self.orders[order_id].received_at).total_seconds()
                    wait_times.append(wait)
            avg_wait = sum(wait_times) / len(wait_times) if wait_times else 0
        
        # Update metrics
        metrics.KITCHEN_QUEUE_DEPTH.labels(
            location_id=self.location_id
        ).set(queue_depth)
        
        metrics.KITCHEN_AVG_WAIT.labels(
            location_id=self.location_id
        ).set(avg_wait)
        
        # Emit status
        await self.emitter.emit(
            event_type=EventTypes.KITCHEN_QUEUE_STATUS,
            data={
                "location_id": self.location_id,
                "queue_depth": queue_depth,
                "avg_wait_seconds": avg_wait,
                "orders_in_progress": sum(
                    1 for o in self.orders.values() if o.started_at and not o.ready_at
                ),
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
        
        # Check thresholds
        if queue_depth > self.queue_alert_threshold:
            await self.emitter.emit(
                event_type=EventTypes.ALERT_RAISED,
                data={
                    "alert_id": f"queue-{self.location_id}-{datetime.utcnow().timestamp()}",
                    "location_id": self.location_id,
                    "source": "kitchen",
                    "severity": AlertSeverity.HIGH.value,
                    "alert_type": "kitchen_queue_high",
                    "message": f"Kitchen queue depth: {queue_depth} orders",
                    "timestamp": datetime.utcnow().isoformat(),
                },
            )
    
    async def _raise_equipment_alert(
        self,
        equipment_id: str,
        alert_type: str,
        message: str,
    ) -> None:
        """Raise equipment alert."""
        await self.emitter.emit(
            event_type=EventTypes.KITCHEN_EQUIPMENT_ALERT,
            data={
                "equipment_id": equipment_id,
                "location_id": self.location_id,
                "alert_type": alert_type,
                "message": message,
                "timestamp": datetime.utcnow().isoformat(),
            },
        )
