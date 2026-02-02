# FutBoss AI - Main Application
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import socketio

from app.config import get_settings
from app.routes import auth, teams, players, matches, tokens, payments
from app.websocket.handlers import sio

settings = get_settings()

app = FastAPI(
    title="FutBoss AI",
    description="Multiplayer football management game with AI agents",
    version="1.0.0",
    contact={
        "name": "Bruno Lucena",
        "email": "bruno@lucena.cloud",
    },
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Routes
app.include_router(auth.router, prefix="/api/auth", tags=["Authentication"])
app.include_router(teams.router, prefix="/api/teams", tags=["Teams"])
app.include_router(players.router, prefix="/api/players", tags=["Players"])
app.include_router(matches.router, prefix="/api/matches", tags=["Matches"])
app.include_router(tokens.router, prefix="/api/tokens", tags=["Tokens"])
app.include_router(payments.router, prefix="/api/payments", tags=["Payments"])

# WebSocket
socket_app = socketio.ASGIApp(sio, other_asgi_app=app)


@app.get("/")
async def root():
    return {"message": "FutBoss AI API", "version": "1.0.0"}


@app.get("/health")
async def health():
    return {"status": "healthy"}

