"""CloudEvents utilities for POS Edge agents."""
import os
import uuid
from datetime import datetime
from typing import Any, Dict, Optional

import httpx
from cloudevents.http import CloudEvent
from cloudevents.conversion import to_structured
import structlog

logger = structlog.get_logger()


class CloudEventEmitter:
    """Emit CloudEvents to the broker."""
    
    def __init__(self, source: str, broker_url: Optional[str] = None):
        self.source = source
        self.broker_url = broker_url or os.getenv(
            "K_SINK",
            os.getenv("BROKER_URL", "http://broker-ingress.knative-eventing.svc.cluster.local/agent-pos-edge/default")
        )
        self._client = httpx.AsyncClient(timeout=10.0)
    
    async def emit(
        self,
        event_type: str,
        data: Dict[str, Any],
        subject: Optional[str] = None,
        extensions: Optional[Dict[str, str]] = None,
    ) -> bool:
        """Emit a CloudEvent."""
        event_id = str(uuid.uuid4())
        
        attributes = {
            "type": event_type,
            "source": self.source,
            "id": event_id,
            "time": datetime.utcnow().isoformat() + "Z",
        }
        
        if subject:
            attributes["subject"] = subject
        
        if extensions:
            attributes.update(extensions)
        
        event = CloudEvent(attributes, data)
        headers, body = to_structured(event)
        
        try:
            response = await self._client.post(
                self.broker_url,
                headers=dict(headers),
                content=body,
            )
            response.raise_for_status()
            
            logger.info(
                "cloudevent_emitted",
                event_type=event_type,
                event_id=event_id,
                status_code=response.status_code,
            )
            return True
            
        except httpx.HTTPError as e:
            logger.error(
                "cloudevent_emit_failed",
                event_type=event_type,
                event_id=event_id,
                error=str(e),
            )
            return False
    
    async def close(self):
        """Close the HTTP client."""
        await self._client.aclose()


# Event type constants
class EventTypes:
    """CloudEvent type constants."""
    
    # Location events
    LOCATION_HEARTBEAT = "pos.location.heartbeat"
    LOCATION_OFFLINE = "pos.location.offline"
    LOCATION_CONFIG_UPDATE = "pos.location.config.update"
    
    # Transaction events
    TRANSACTION_STARTED = "pos.transaction.started"
    TRANSACTION_COMPLETED = "pos.transaction.completed"
    TRANSACTION_FAILED = "pos.transaction.failed"
    
    # Health events
    HEALTH_REPORT = "pos.health.report"
    ALERT_RAISED = "pos.alert.raised"
    
    # Kitchen events
    KITCHEN_ORDER_RECEIVED = "pos.kitchen.order.received"
    KITCHEN_ORDER_STARTED = "pos.kitchen.order.started"
    KITCHEN_ORDER_READY = "pos.kitchen.order.ready"
    KITCHEN_QUEUE_STATUS = "pos.kitchen.queue.status"
    KITCHEN_EQUIPMENT_ALERT = "pos.kitchen.equipment.alert"
    
    # Pump events
    PUMP_TRANSACTION_START = "pos.pump.transaction.start"
    PUMP_TRANSACTION_END = "pos.pump.transaction.end"
    PUMP_STATUS = "pos.pump.status"
    TANK_LEVEL = "pos.tank.level"
    TANK_ALERT_LOW = "pos.tank.alert.low"
    
    # Command center events
    COMMAND_CONFIG_PUSH = "pos.command.config.push"
    COMMAND_ALERT_ACKNOWLEDGE = "pos.command.alert.acknowledge"
    COMMAND_MAINTENANCE_SCHEDULE = "pos.command.maintenance.schedule"
    FLEET_STATUS_REPORT = "pos.fleet.status.report"
