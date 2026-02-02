"""
AgentChat Messaging Hub Handler

Central message routing and delivery for the AgentChat platform.
Handles CloudEvents for message routing, WebSocket connections, and presence.
"""
import os
import json
import logging
import asyncio
from datetime import datetime
from typing import Any, Dict, Optional
from dataclasses import dataclass, asdict
from cloudevents.http import CloudEvent, to_structured, from_http
from flask import Flask, request, jsonify
import redis.asyncio as redis

# Configure logging
logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Redis connection for presence and pub/sub
REDIS_URL = os.getenv('REDIS_URL', 'redis://redis.agent-chat.svc.cluster.local:6379')

# Event type constants
class EventTypes:
    MESSAGE_TEXT = 'io.agentchat.message.text'
    MESSAGE_VOICE = 'io.agentchat.message.voice'
    MESSAGE_IMAGE = 'io.agentchat.message.image'
    MESSAGE_VIDEO = 'io.agentchat.message.video'
    MESSAGE_DELIVERED = 'io.agentchat.message.delivered'
    MESSAGE_READ = 'io.agentchat.message.read'
    TYPING_START = 'io.agentchat.typing.start'
    TYPING_STOP = 'io.agentchat.typing.stop'
    PRESENCE_UPDATE = 'io.agentchat.presence.update'
    AGENT_INVOKE = 'io.agentchat.agent.invoke'


@dataclass
class Message:
    """Message structure for AgentChat"""
    id: str
    chat_id: str
    sender_id: str
    content: str
    message_type: str
    timestamp: str
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class RoutingResult:
    """Result of message routing"""
    success: bool
    target: str
    event_id: str
    timestamp: str
    error: Optional[str] = None


app = Flask(__name__)


class MessagingHub:
    """Central messaging hub for AgentChat"""
    
    def __init__(self):
        self.redis_client = None
        self.message_handlers = {
            EventTypes.MESSAGE_TEXT: self.handle_text_message,
            EventTypes.MESSAGE_VOICE: self.handle_voice_message,
            EventTypes.MESSAGE_IMAGE: self.handle_image_message,
            EventTypes.MESSAGE_VIDEO: self.handle_video_message,
            EventTypes.TYPING_START: self.handle_typing,
            EventTypes.TYPING_STOP: self.handle_typing,
            EventTypes.PRESENCE_UPDATE: self.handle_presence,
        }
    
    async def get_redis(self) -> redis.Redis:
        """Get or create Redis connection"""
        if self.redis_client is None:
            self.redis_client = redis.from_url(REDIS_URL, decode_responses=True)
        return self.redis_client
    
    async def route_message(self, event: CloudEvent) -> RoutingResult:
        """Route incoming CloudEvent to appropriate handler"""
        event_type = event.get('type')
        event_id = event.get('id')
        
        logger.info(f"Routing event {event_id} of type {event_type}")
        
        handler = self.message_handlers.get(event_type)
        if handler:
            try:
                result = await handler(event)
                return RoutingResult(
                    success=True,
                    target=result.get('target', 'processed'),
                    event_id=event_id,
                    timestamp=datetime.utcnow().isoformat()
                )
            except Exception as e:
                logger.error(f"Error handling event {event_id}: {e}")
                return RoutingResult(
                    success=False,
                    target='error',
                    event_id=event_id,
                    timestamp=datetime.utcnow().isoformat(),
                    error=str(e)
                )
        else:
            logger.warning(f"No handler for event type: {event_type}")
            return RoutingResult(
                success=False,
                target='unknown',
                event_id=event_id,
                timestamp=datetime.utcnow().isoformat(),
                error=f"Unknown event type: {event_type}"
            )
    
    async def handle_text_message(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle text message event"""
        data = event.data
        chat_id = data.get('chatId')
        sender_id = data.get('userId')
        content = data.get('content')
        
        logger.info(f"Processing text message from {sender_id} in chat {chat_id}")
        
        # Store message in Redis for history
        r = await self.get_redis()
        message = Message(
            id=event.get('id'),
            chat_id=chat_id,
            sender_id=sender_id,
            content=content,
            message_type='text',
            timestamp=datetime.utcnow().isoformat()
        )
        
        # Store in chat history (last 100 messages)
        await r.lpush(f"chat:{chat_id}:messages", json.dumps(asdict(message)))
        await r.ltrim(f"chat:{chat_id}:messages", 0, 99)
        
        # Publish to chat channel for WebSocket delivery
        await r.publish(f"chat:{chat_id}", json.dumps({
            'type': 'message',
            'data': asdict(message)
        }))
        
        # Check if this is a message to an agent
        if sender_id.startswith('user-') and chat_id.startswith('agent-'):
            # Route to agent assistant
            return {'target': 'agent-assistant', 'forwarded': True}
        
        return {'target': 'delivered', 'stored': True}
    
    async def handle_voice_message(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle voice message event - forward to voice agent"""
        data = event.data
        logger.info(f"Forwarding voice message to voice-agent")
        
        # Forward to voice agent for processing
        return {'target': 'voice-agent', 'forwarded': True}
    
    async def handle_image_message(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle image message event"""
        data = event.data
        logger.info(f"Processing image message")
        
        # Store reference and forward if analysis needed
        if data.get('requestAnalysis'):
            return {'target': 'media-agent', 'forwarded': True}
        
        return {'target': 'delivered', 'stored': True}
    
    async def handle_video_message(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle video message event"""
        data = event.data
        logger.info(f"Processing video message")
        
        return {'target': 'delivered', 'stored': True}
    
    async def handle_typing(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle typing indicators"""
        data = event.data
        chat_id = data.get('chatId')
        user_id = data.get('userId')
        is_typing = event.get('type') == EventTypes.TYPING_START
        
        r = await self.get_redis()
        await r.publish(f"chat:{chat_id}:typing", json.dumps({
            'userId': user_id,
            'isTyping': is_typing
        }))
        
        return {'target': 'broadcast', 'chat_id': chat_id}
    
    async def handle_presence(self, event: CloudEvent) -> Dict[str, Any]:
        """Handle user presence updates"""
        data = event.data
        user_id = data.get('userId')
        status = data.get('status', 'online')
        
        r = await self.get_redis()
        await r.hset(f"presence:{user_id}", mapping={
            'status': status,
            'lastSeen': datetime.utcnow().isoformat()
        })
        
        # Expire presence after 5 minutes of no updates
        await r.expire(f"presence:{user_id}", 300)
        
        return {'target': 'presence-updated', 'user_id': user_id}


# Global hub instance
hub = MessagingHub()


@app.route('/', methods=['POST'])
def handle_cloudevent():
    """Handle incoming CloudEvent"""
    try:
        # Parse CloudEvent
        event = from_http(request.headers, request.get_data())
        
        # Route the event
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        result = loop.run_until_complete(hub.route_message(event))
        loop.close()
        
        if result.success:
            return jsonify(asdict(result)), 200
        else:
            return jsonify(asdict(result)), 400
            
    except Exception as e:
        logger.error(f"Error processing request: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({'status': 'healthy', 'service': 'messaging-hub'}), 200


@app.route('/ready', methods=['GET'])
def ready():
    """Readiness check endpoint"""
    return jsonify({'status': 'ready', 'service': 'messaging-hub'}), 200


@app.route('/metrics', methods=['GET'])
def metrics():
    """Prometheus metrics endpoint"""
    # Basic metrics - in production use prometheus_client
    return """
# HELP agentchat_messaging_hub_requests_total Total requests processed
# TYPE agentchat_messaging_hub_requests_total counter
agentchat_messaging_hub_requests_total 0
# HELP agentchat_messaging_hub_up Service is up
# TYPE agentchat_messaging_hub_up gauge
agentchat_messaging_hub_up 1
""", 200, {'Content-Type': 'text/plain'}


if __name__ == '__main__':
    port = int(os.getenv('PORT', 8080))
    logger.info(f"Starting AgentChat Messaging Hub on port {port}")
    app.run(host='0.0.0.0', port=port)
