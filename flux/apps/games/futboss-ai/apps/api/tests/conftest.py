# FutBoss AI - Test Configuration
# Author: Bruno Lucena (bruno@lucena.cloud)

import pytest
from unittest.mock import AsyncMock, MagicMock
from fastapi.testclient import TestClient

# Mock MongoDB before imports
import sys
sys.modules['motor'] = MagicMock()
sys.modules['motor.motor_asyncio'] = MagicMock()

from app.main import app


@pytest.fixture
def client():
    return TestClient(app)


@pytest.fixture
def mock_user():
    return {
        "_id": "507f1f77bcf86cd799439011",
        "username": "testuser",
        "email": "test@example.com",
        "hashed_password": "$2b$12$test",
        "tokens": 1000,
        "is_active": True,
    }


@pytest.fixture
def mock_team():
    return {
        "_id": "507f1f77bcf86cd799439012",
        "name": "Test FC",
        "owner_id": "507f1f77bcf86cd799439011",
        "formation": "4-4-2",
        "player_ids": [],
        "wins": 5,
        "draws": 3,
        "losses": 2,
        "goals_for": 15,
        "goals_against": 10,
    }


@pytest.fixture
def mock_player():
    return {
        "_id": "507f1f77bcf86cd799439013",
        "name": "Test Player",
        "position": "ST",
        "nationality": "Brasil",
        "age": 25,
        "attributes": {
            "speed": 75,
            "strength": 70,
            "stamina": 80,
            "finishing": 85,
            "passing": 70,
            "dribbling": 78,
            "defense": 40,
            "intelligence": 72,
            "aggression": 60,
            "leadership": 55,
            "creativity": 80,
        },
        "personality": {
            "temperament": "explosive",
            "play_style": "offensive",
        },
        "team_id": None,
        "price": 500,
        "is_listed": True,
    }

