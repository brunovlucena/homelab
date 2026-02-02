# FutBoss AI - WebSocket Handlers
# Author: Bruno Lucena (bruno@lucena.cloud)

import socketio
from typing import Dict, Set

sio = socketio.AsyncServer(async_mode="asgi", cors_allowed_origins="*")

# Active connections: {user_id: sid}
active_users: Dict[str, str] = {}
# Match rooms: {match_id: set(user_ids)}
match_rooms: Dict[str, Set[str]] = {}


@sio.event
async def connect(sid, environ, auth):
    """Handle new connection"""
    user_id = auth.get("user_id") if auth else None
    if user_id:
        active_users[user_id] = sid
        await sio.emit("connected", {"status": "ok", "user_id": user_id}, to=sid)
    print(f"Client connected: {sid}")


@sio.event
async def disconnect(sid):
    """Handle disconnection"""
    # Remove from active users
    user_to_remove = None
    for user_id, user_sid in active_users.items():
        if user_sid == sid:
            user_to_remove = user_id
            break
    if user_to_remove:
        del active_users[user_to_remove]
    
    # Remove from match rooms
    for match_id, users in match_rooms.items():
        if user_to_remove in users:
            users.remove(user_to_remove)
    
    print(f"Client disconnected: {sid}")


@sio.event
async def join_match(sid, data):
    """Join a match room"""
    match_id = data.get("match_id")
    user_id = data.get("user_id")
    
    if not match_id or not user_id:
        return {"error": "Missing match_id or user_id"}
    
    if match_id not in match_rooms:
        match_rooms[match_id] = set()
    
    match_rooms[match_id].add(user_id)
    await sio.enter_room(sid, f"match:{match_id}")
    
    await sio.emit("joined_match", {
        "match_id": match_id,
        "players": list(match_rooms[match_id])
    }, room=f"match:{match_id}")


@sio.event
async def leave_match(sid, data):
    """Leave a match room"""
    match_id = data.get("match_id")
    user_id = data.get("user_id")
    
    if match_id in match_rooms and user_id in match_rooms[match_id]:
        match_rooms[match_id].remove(user_id)
    
    await sio.leave_room(sid, f"match:{match_id}")
    await sio.emit("left_match", {"match_id": match_id, "user_id": user_id}, room=f"match:{match_id}")


@sio.event
async def match_action(sid, data):
    """Handle match action from player"""
    match_id = data.get("match_id")
    action = data.get("action")
    player_id = data.get("player_id")
    
    # Broadcast action to match room
    await sio.emit("match_update", {
        "match_id": match_id,
        "action": action,
        "player_id": player_id,
        "data": data.get("data", {})
    }, room=f"match:{match_id}")


@sio.event
async def chat_message(sid, data):
    """Handle chat message in match"""
    match_id = data.get("match_id")
    user_id = data.get("user_id")
    message = data.get("message")
    
    await sio.emit("chat", {
        "match_id": match_id,
        "user_id": user_id,
        "message": message
    }, room=f"match:{match_id}")


# Helper functions for server-side events
async def broadcast_match_event(match_id: str, event_type: str, event_data: dict):
    """Broadcast match event to all players in room"""
    await sio.emit("match_event", {
        "match_id": match_id,
        "type": event_type,
        "data": event_data
    }, room=f"match:{match_id}")


async def send_to_user(user_id: str, event: str, data: dict):
    """Send event to specific user"""
    if user_id in active_users:
        await sio.emit(event, data, to=active_users[user_id])

