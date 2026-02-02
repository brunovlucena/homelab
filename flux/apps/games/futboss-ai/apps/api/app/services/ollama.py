# FutBoss AI - Ollama AI Agent Service
# Author: Bruno Lucena (bruno@lucena.cloud)

import httpx
from typing import Optional, Dict, Any
from app.config import get_settings
from app.models.player import Player, PlayerAttributes, PlayerPersonality

settings = get_settings()


class OllamaService:
    """Service to interact with local Ollama for AI agent decisions"""
    
    def __init__(self, base_url: str = None, model: str = None):
        self.base_url = base_url or settings.ollama_base_url
        self.model = model or settings.ollama_model

    def build_agent_prompt(self, player: Player, game_state: Dict[str, Any]) -> str:
        """Build a prompt based on player attributes and personality"""
        attrs = player.attributes
        personality = player.personality
        
        # Personality description
        temp_desc = {
            "calm": "You are calm and composed under pressure, preferring safe plays.",
            "explosive": "You are explosive and unpredictable, taking risks for big rewards.",
            "calculated": "You are calculated and strategic, analyzing every option carefully.",
        }
        
        style_desc = {
            "offensive": "You prefer attacking plays, always looking to score.",
            "defensive": "You focus on defense, prioritizing ball retention and safety.",
            "balanced": "You balance attack and defense based on the situation.",
        }
        
        prompt = f"""You are {player.name}, a {player.position.value} football player.

PERSONALITY:
{temp_desc.get(personality.temperament.value, temp_desc['calm'])}
{style_desc.get(personality.play_style.value, style_desc['balanced'])}

YOUR ATTRIBUTES (1-100 scale):
- Speed: {attrs.speed}
- Strength: {attrs.strength}  
- Finishing: {attrs.finishing}
- Passing: {attrs.passing}
- Dribbling: {attrs.dribbling}
- Defense: {attrs.defense}
- Intelligence: {attrs.intelligence}
- Creativity: {attrs.creativity}
- Aggression: {attrs.aggression}

CURRENT GAME STATE:
- Minute: {game_state.get('minute', 0)}
- Score: {game_state.get('home_score', 0)} - {game_state.get('away_score', 0)}
- Your team possession: {game_state.get('has_ball', False)}
- Ball position: {game_state.get('ball_zone', 'midfield')}

Based on your personality and attributes, what action do you take?
Choose one: PASS, DRIBBLE, SHOOT, TACKLE, HOLD, RUN_FORWARD, FALL_BACK

Respond with ONLY the action name and a brief reason (max 20 words).
Example: "DRIBBLE - My high dribbling skill lets me beat the defender."
"""
        return prompt

    async def get_agent_decision(self, player: Player, game_state: Dict[str, Any]) -> Dict[str, str]:
        """Get AI decision for a player agent"""
        prompt = self.build_agent_prompt(player, game_state)
        
        try:
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    f"{self.base_url}/api/generate",
                    json={
                        "model": self.model,
                        "prompt": prompt,
                        "stream": False,
                        "options": {
                            "temperature": 0.7 + (player.attributes.creativity / 200),  # More creative = more random
                            "num_predict": 50,
                        }
                    }
                )
                response.raise_for_status()
                result = response.json()
                
                text = result.get("response", "").strip()
                parts = text.split(" - ", 1)
                action = parts[0].upper().strip()
                reason = parts[1] if len(parts) > 1 else ""
                
                valid_actions = ["PASS", "DRIBBLE", "SHOOT", "TACKLE", "HOLD", "RUN_FORWARD", "FALL_BACK"]
                if action not in valid_actions:
                    action = "HOLD"
                
                return {"action": action, "reason": reason}
        
        except Exception as e:
            # Fallback to attribute-based decision
            return self._fallback_decision(player, game_state)

    def _fallback_decision(self, player: Player, game_state: Dict[str, Any]) -> Dict[str, str]:
        """Fallback decision when Ollama is unavailable"""
        attrs = player.attributes
        has_ball = game_state.get("has_ball", False)
        ball_zone = game_state.get("ball_zone", "midfield")
        
        if not has_ball:
            if attrs.defense > 60:
                return {"action": "TACKLE", "reason": "Defensive instinct"}
            return {"action": "FALL_BACK", "reason": "Positioning"}
        
        if ball_zone == "attack":
            if attrs.finishing > 70:
                return {"action": "SHOOT", "reason": "In scoring position"}
            if attrs.passing > attrs.dribbling:
                return {"action": "PASS", "reason": "Better passing option"}
            return {"action": "DRIBBLE", "reason": "Create space"}
        
        if attrs.dribbling > 70:
            return {"action": "DRIBBLE", "reason": "Skill advantage"}
        return {"action": "PASS", "reason": "Move ball forward"}

    async def generate_narration(self, event: Dict[str, Any]) -> str:
        """Generate AI narration for match events"""
        event_type = event.get("type", "")
        player_name = event.get("player_name", "Player")
        minute = event.get("minute", 0)
        
        prompt = f"""You are a passionate Brazilian football commentator.
Generate a SHORT, exciting narration (max 30 words) for this event:

Event: {event_type}
Player: {player_name}
Minute: {minute}'

Be dramatic and use Brazilian football expressions!
"""
        
        try:
            async with httpx.AsyncClient(timeout=15.0) as client:
                response = await client.post(
                    f"{self.base_url}/api/generate",
                    json={
                        "model": self.model,
                        "prompt": prompt,
                        "stream": False,
                        "options": {"temperature": 0.9, "num_predict": 60}
                    }
                )
                response.raise_for_status()
                return response.json().get("response", "").strip()
        except:
            narrations = {
                "goal": f"GOOOOL! {player_name} marca aos {minute}!",
                "save": f"Defesa incrível! {player_name} salva o time!",
                "yellow_card": f"Cartão amarelo para {player_name}!",
                "red_card": f"EXPULSO! {player_name} recebe o vermelho!",
            }
            return narrations.get(event_type, f"{player_name} em ação aos {minute}'!")

    async def check_health(self) -> bool:
        """Check if Ollama is available"""
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                response = await client.get(f"{self.base_url}/api/tags")
                return response.status_code == 200
        except:
            return False

