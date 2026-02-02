# FutBoss AI - Ollama Service Tests
# Author: Bruno Lucena (bruno@lucena.cloud)

import pytest
from unittest.mock import AsyncMock, patch

from app.services.ollama import OllamaService
from app.models.player import Player, PlayerAttributes, PlayerPersonality, Position, Temperament, PlayStyle


class TestOllamaService:
    @pytest.fixture
    def ollama_service(self):
        return OllamaService(base_url="http://localhost:11434", model="llama3.2")

    @pytest.fixture
    def test_player(self):
        return Player(
            name="Test Striker",
            position=Position.ST,
            nationality="Brasil",
            age=25,
            attributes=PlayerAttributes(
                speed=85, strength=70, stamina=80,
                finishing=90, passing=65, dribbling=80, defense=30,
                intelligence=75, aggression=70, leadership=50, creativity=85
            ),
            personality=PlayerPersonality(
                temperament=Temperament.EXPLOSIVE,
                play_style=PlayStyle.OFFENSIVE
            )
        )

    @pytest.fixture
    def game_state(self):
        return {
            "minute": 45,
            "home_score": 1,
            "away_score": 1,
            "has_ball": True,
            "ball_zone": "attack"
        }

    def test_build_agent_prompt(self, ollama_service, test_player, game_state):
        prompt = ollama_service.build_agent_prompt(test_player, game_state)
        
        assert "Test Striker" in prompt
        assert "ST" in prompt
        assert "Speed: 85" in prompt
        assert "explosive" in prompt.lower()
        assert "offensive" in prompt.lower()
        assert "Minute: 45" in prompt

    def test_fallback_decision_with_ball_attack(self, ollama_service, test_player, game_state):
        result = ollama_service._fallback_decision(test_player, game_state)
        
        # High finishing player in attack zone should shoot
        assert result["action"] == "SHOOT"
        assert "reason" in result

    def test_fallback_decision_with_ball_midfield(self, ollama_service, test_player):
        game_state = {"has_ball": True, "ball_zone": "midfield"}
        result = ollama_service._fallback_decision(test_player, game_state)
        
        # High dribbling player should dribble
        assert result["action"] == "DRIBBLE"

    def test_fallback_decision_without_ball_defender(self, ollama_service):
        defender = Player(
            name="Defender",
            position=Position.CB,
            nationality="Brasil",
            age=28,
            attributes=PlayerAttributes(defense=85, speed=60),
            personality=PlayerPersonality()
        )
        game_state = {"has_ball": False, "ball_zone": "defense"}
        
        result = ollama_service._fallback_decision(defender, game_state)
        assert result["action"] == "TACKLE"

    def test_fallback_decision_without_ball_midfielder(self, ollama_service):
        midfielder = Player(
            name="Midfielder",
            position=Position.CM,
            nationality="Brasil",
            age=26,
            attributes=PlayerAttributes(defense=50, passing=75),
            personality=PlayerPersonality()
        )
        game_state = {"has_ball": False, "ball_zone": "midfield"}
        
        result = ollama_service._fallback_decision(midfielder, game_state)
        assert result["action"] == "FALL_BACK"

    @pytest.mark.asyncio
    async def test_get_agent_decision_fallback(self, ollama_service, test_player, game_state):
        # When Ollama is not available, should use fallback
        with patch('httpx.AsyncClient.post', side_effect=Exception("Connection refused")):
            result = await ollama_service.get_agent_decision(test_player, game_state)
            
            assert "action" in result
            assert result["action"] in ["PASS", "DRIBBLE", "SHOOT", "TACKLE", "HOLD", "RUN_FORWARD", "FALL_BACK"]

    @pytest.mark.asyncio
    async def test_check_health_failure(self, ollama_service):
        with patch('httpx.AsyncClient.get', side_effect=Exception("Connection refused")):
            result = await ollama_service.check_health()
            assert result is False

