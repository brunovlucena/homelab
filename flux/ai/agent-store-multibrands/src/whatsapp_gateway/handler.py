"""
WhatsApp Business API handler.

Processes incoming webhooks and sends outgoing messages via Meta's API.
"""
import os
import hashlib
import hmac
from typing import Optional
from datetime import datetime, timezone

import httpx
import structlog
from cloudevents.http import CloudEvent

from shared.types import (
    Brand,
    WhatsAppMessage,
    WhatsAppMessageType,
    Customer,
)
from shared.events import (
    EventType,
    EventPublisher,
    EventSubscriber,
    WhatsAppMessageEvent,
    ChatMessageEvent,
    ChatResponseEvent,
)
from shared.metrics import (
    WHATSAPP_MESSAGES_RECEIVED,
    WHATSAPP_MESSAGES_SENT,
    WHATSAPP_API_LATENCY,
    WHATSAPP_API_ERRORS,
)

logger = structlog.get_logger()


class WhatsAppGateway:
    """
    WhatsApp Business API Gateway.
    
    Handles:
    - Webhook verification
    - Incoming message processing
    - Outgoing message sending
    - Media upload/download
    - Interactive messages (buttons, lists)
    """
    
    def __init__(
        self,
        phone_number_id: str = None,
        access_token: str = None,
        verify_token: str = None,
        app_secret: str = None,
        event_publisher: Optional[EventPublisher] = None,
        event_subscriber: Optional[EventSubscriber] = None,
    ):
        self.phone_number_id = phone_number_id or os.getenv("WHATSAPP_PHONE_NUMBER_ID")
        self.access_token = access_token or os.getenv("WHATSAPP_ACCESS_TOKEN")
        self.verify_token = verify_token or os.getenv("WHATSAPP_VERIFY_TOKEN", "homelab-store")
        self.app_secret = app_secret or os.getenv("WHATSAPP_APP_SECRET")
        
        self.api_version = os.getenv("WHATSAPP_API_VERSION", "v18.0")
        self.api_base_url = f"https://graph.facebook.com/{self.api_version}"
        
        self.event_publisher = event_publisher
        self.event_subscriber = event_subscriber
        
        # Customer cache (simple in-memory, use Redis in production)
        self._customers: dict[str, Customer] = {}
        
        # Brand routing based on phone prefix or customer preference
        self._default_brand = Brand.FASHION
        
        logger.info(
            "whatsapp_gateway_initialized",
            phone_number_id=self.phone_number_id[:4] + "..." if self.phone_number_id else None,
            api_version=self.api_version,
        )
    
    # =========================================================================
    # Webhook Handling
    # =========================================================================
    
    def verify_webhook(self, mode: str, token: str, challenge: str) -> Optional[str]:
        """
        Verify webhook subscription.
        
        Returns challenge if verification passes, None otherwise.
        """
        if mode == "subscribe" and token == self.verify_token:
            logger.info("webhook_verified")
            return challenge
        logger.warning("webhook_verification_failed", mode=mode)
        return None
    
    def verify_signature(self, payload: bytes, signature: str) -> bool:
        """Verify webhook payload signature."""
        if not self.app_secret:
            logger.warning("app_secret_not_configured")
            return True  # Skip verification if not configured
        
        expected = hmac.new(
            self.app_secret.encode(),
            payload,
            hashlib.sha256
        ).hexdigest()
        
        return hmac.compare_digest(f"sha256={expected}", signature)
    
    async def process_webhook(self, payload: dict) -> list[WhatsAppMessage]:
        """
        Process incoming webhook payload.
        
        Returns list of messages extracted from the webhook.
        """
        messages = []
        
        try:
            # Extract entries from webhook
            entries = payload.get("entry", [])
            
            for entry in entries:
                changes = entry.get("changes", [])
                
                for change in changes:
                    if change.get("field") != "messages":
                        continue
                    
                    value = change.get("value", {})
                    
                    # Process messages
                    for msg_data in value.get("messages", []):
                        message = await self._parse_message(msg_data, value)
                        if message:
                            messages.append(message)
                            await self._emit_message_received(message)
                    
                    # Process status updates
                    for status in value.get("statuses", []):
                        await self._process_status_update(status)
            
            return messages
            
        except Exception as e:
            logger.error("webhook_processing_failed", error=str(e))
            WHATSAPP_API_ERRORS.labels(
                operation="webhook",
                error_type=type(e).__name__
            ).inc()
            return []
    
    async def _parse_message(self, msg_data: dict, value: dict) -> Optional[WhatsAppMessage]:
        """Parse a message from webhook data."""
        try:
            msg_id = msg_data.get("id")
            phone_from = msg_data.get("from")
            timestamp = msg_data.get("timestamp")
            msg_type = msg_data.get("type", "text")
            
            # Get our phone number from contacts
            contacts = value.get("contacts", [{}])
            phone_to = value.get("metadata", {}).get("display_phone_number", "")
            
            # Extract content based on type
            content = ""
            media_url = None
            media_mime = None
            interactive_data = None
            
            if msg_type == "text":
                content = msg_data.get("text", {}).get("body", "")
            elif msg_type == "image":
                img_data = msg_data.get("image", {})
                content = img_data.get("caption", "[Image]")
                media_url = img_data.get("id")  # Media ID, need to download
                media_mime = img_data.get("mime_type")
            elif msg_type == "audio":
                audio_data = msg_data.get("audio", {})
                content = "[Audio message]"
                media_url = audio_data.get("id")
                media_mime = audio_data.get("mime_type")
            elif msg_type == "document":
                doc_data = msg_data.get("document", {})
                content = doc_data.get("filename", "[Document]")
                media_url = doc_data.get("id")
                media_mime = doc_data.get("mime_type")
            elif msg_type == "location":
                loc_data = msg_data.get("location", {})
                content = f"[Location: {loc_data.get('latitude')}, {loc_data.get('longitude')}]"
            elif msg_type == "interactive":
                interactive = msg_data.get("interactive", {})
                interactive_type = interactive.get("type")
                if interactive_type == "button_reply":
                    content = interactive.get("button_reply", {}).get("title", "")
                    interactive_data = interactive.get("button_reply")
                elif interactive_type == "list_reply":
                    content = interactive.get("list_reply", {}).get("title", "")
                    interactive_data = interactive.get("list_reply")
            
            # Map to our enum
            wa_type = {
                "text": WhatsAppMessageType.TEXT,
                "image": WhatsAppMessageType.IMAGE,
                "audio": WhatsAppMessageType.AUDIO,
                "video": WhatsAppMessageType.VIDEO,
                "document": WhatsAppMessageType.DOCUMENT,
                "location": WhatsAppMessageType.LOCATION,
                "sticker": WhatsAppMessageType.STICKER,
                "interactive": WhatsAppMessageType.INTERACTIVE,
            }.get(msg_type, WhatsAppMessageType.TEXT)
            
            message = WhatsAppMessage(
                id=msg_id,
                phone_from=phone_from,
                phone_to=phone_to,
                type=wa_type,
                content=content,
                media_url=media_url,
                media_mime_type=media_mime,
                timestamp=datetime.fromtimestamp(int(timestamp), timezone.utc).isoformat() if timestamp else "",
                context_message_id=msg_data.get("context", {}).get("id"),
                interactive_data=interactive_data,
            )
            
            WHATSAPP_MESSAGES_RECEIVED.labels(message_type=wa_type.value).inc()
            
            logger.info(
                "message_received",
                message_id=msg_id,
                phone_from=phone_from,
                type=wa_type.value,
            )
            
            return message
            
        except Exception as e:
            logger.error("message_parse_failed", error=str(e))
            return None
    
    async def _process_status_update(self, status: dict):
        """Process message status update."""
        msg_id = status.get("id")
        recipient = status.get("recipient_id")
        status_type = status.get("status")  # sent, delivered, read, failed
        
        logger.debug(
            "status_update",
            message_id=msg_id,
            recipient=recipient,
            status=status_type,
        )
        
        # Could emit status event here if needed
    
    # =========================================================================
    # Message Sending
    # =========================================================================
    
    async def send_text_message(
        self,
        to: str,
        text: str,
        reply_to: Optional[str] = None,
    ) -> Optional[str]:
        """
        Send a text message.
        
        Returns message ID if successful.
        """
        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "text",
            "text": {"body": text},
        }
        
        if reply_to:
            payload["context"] = {"message_id": reply_to}
        
        return await self._send_message(payload)
    
    async def send_image_message(
        self,
        to: str,
        image_url: str,
        caption: Optional[str] = None,
    ) -> Optional[str]:
        """Send an image message."""
        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "image",
            "image": {
                "link": image_url,
            },
        }
        
        if caption:
            payload["image"]["caption"] = caption
        
        return await self._send_message(payload)
    
    async def send_interactive_buttons(
        self,
        to: str,
        body_text: str,
        buttons: list[dict],
        header: Optional[str] = None,
        footer: Optional[str] = None,
    ) -> Optional[str]:
        """
        Send interactive button message.
        
        Args:
            buttons: List of {"id": "btn_id", "title": "Button Text"}
        """
        interactive = {
            "type": "button",
            "body": {"text": body_text},
            "action": {
                "buttons": [
                    {"type": "reply", "reply": btn}
                    for btn in buttons[:3]  # Max 3 buttons
                ]
            },
        }
        
        if header:
            interactive["header"] = {"type": "text", "text": header}
        if footer:
            interactive["footer"] = {"text": footer}
        
        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "interactive",
            "interactive": interactive,
        }
        
        return await self._send_message(payload)
    
    async def send_interactive_list(
        self,
        to: str,
        body_text: str,
        button_text: str,
        sections: list[dict],
        header: Optional[str] = None,
        footer: Optional[str] = None,
    ) -> Optional[str]:
        """
        Send interactive list message.
        
        Args:
            sections: List of {"title": "Section", "rows": [{"id": "1", "title": "Item", "description": "..."}]}
        """
        interactive = {
            "type": "list",
            "body": {"text": body_text},
            "action": {
                "button": button_text,
                "sections": sections,
            },
        }
        
        if header:
            interactive["header"] = {"type": "text", "text": header}
        if footer:
            interactive["footer"] = {"text": footer}
        
        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "interactive",
            "interactive": interactive,
        }
        
        return await self._send_message(payload)
    
    async def _send_message(self, payload: dict) -> Optional[str]:
        """Send message via WhatsApp API."""
        import time
        start_time = time.time()
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.api_base_url}/{self.phone_number_id}/messages",
                    headers={
                        "Authorization": f"Bearer {self.access_token}",
                        "Content-Type": "application/json",
                    },
                    json=payload,
                    timeout=30.0,
                )
                
                duration = time.time() - start_time
                WHATSAPP_API_LATENCY.labels(operation="send").observe(duration)
                
                if response.status_code == 200:
                    result = response.json()
                    msg_id = result.get("messages", [{}])[0].get("id")
                    
                    WHATSAPP_MESSAGES_SENT.labels(
                        message_type=payload.get("type", "text")
                    ).inc()
                    
                    logger.info(
                        "message_sent",
                        message_id=msg_id,
                        to=payload.get("to"),
                        type=payload.get("type"),
                    )
                    
                    return msg_id
                else:
                    error = response.json()
                    WHATSAPP_API_ERRORS.labels(
                        operation="send",
                        error_type=str(response.status_code)
                    ).inc()
                    logger.error(
                        "message_send_failed",
                        status=response.status_code,
                        error=error,
                    )
                    return None
                    
        except Exception as e:
            WHATSAPP_API_ERRORS.labels(
                operation="send",
                error_type=type(e).__name__
            ).inc()
            logger.error("message_send_error", error=str(e))
            return None
    
    # =========================================================================
    # Event Integration
    # =========================================================================
    
    async def _emit_message_received(self, message: WhatsAppMessage):
        """Emit message received event."""
        if not self.event_publisher:
            return
        
        # Determine brand based on customer preference or default
        customer = self._get_or_create_customer(message.phone_from)
        brand = customer.preferred_brand or self._default_brand
        
        # Emit WhatsApp event
        await self.event_publisher.emit_whatsapp_received(
            WhatsAppMessageEvent(
                message_id=message.id,
                phone_from=message.phone_from,
                phone_to=message.phone_to,
                message_type=message.type.value,
                content=message.content,
                media_url=message.media_url,
                context_message_id=message.context_message_id,
            )
        )
        
        # Also emit chat message event for AI sellers
        await self.event_publisher.emit_chat_message(
            ChatMessageEvent(
                conversation_id=f"wa-{message.phone_from}",
                customer_id=customer.id,
                customer_phone=message.phone_from,
                brand=brand.value,
                message=message.content,
                message_role="customer",
            )
        )
    
    async def handle_chat_response(self, event: CloudEvent):
        """
        Handle AI seller response event.
        
        Sends the response back to the customer via WhatsApp.
        """
        data = event.data or {}
        
        customer_phone = data.get("customer_phone")
        response = data.get("response")
        recommendations = data.get("recommendations", [])
        
        if not customer_phone or not response:
            logger.warning("invalid_chat_response_event")
            return
        
        # Send text response
        await self.send_text_message(customer_phone, response)
        
        # If there are product recommendations, send as interactive list
        if recommendations:
            await self._send_product_recommendations(customer_phone, recommendations)
    
    async def _send_product_recommendations(
        self,
        phone: str,
        recommendations: list[dict],
    ):
        """Send product recommendations as interactive list."""
        if not recommendations:
            return
        
        rows = []
        for rec in recommendations[:10]:  # Max 10 items
            rows.append({
                "id": rec.get("product_id", ""),
                "title": rec.get("product_name", "")[:24],  # Max 24 chars
                "description": rec.get("price_info", "")[:72],  # Max 72 chars
            })
        
        await self.send_interactive_list(
            to=phone,
            body_text="Aqui estão algumas sugestões para você! 🛍️",
            button_text="Ver Produtos",
            sections=[{
                "title": "Produtos Recomendados",
                "rows": rows,
            }],
            footer="Toque para ver mais detalhes",
        )
    
    def _get_or_create_customer(self, phone: str) -> Customer:
        """Get or create customer record."""
        if phone not in self._customers:
            from uuid import uuid4
            self._customers[phone] = Customer(
                id=str(uuid4()),
                phone=phone,
            )
        return self._customers[phone]
    
    def setup_event_handlers(self):
        """Register event handlers."""
        if self.event_subscriber:
            self.event_subscriber.register(
                EventType.CHAT_RESPONSE,
                self.handle_chat_response
            )
            logger.info("whatsapp_event_handlers_registered")
