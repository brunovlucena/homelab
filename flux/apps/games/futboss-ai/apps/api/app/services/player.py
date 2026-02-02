# FutBoss AI - Player Service
# Author: Bruno Lucena (bruno@lucena.cloud)

import random
from typing import Optional, List
from bson import ObjectId

from app.models.player import (
    Player, PlayerCreate, PlayerAttributes, PlayerPersonality,
    Position, Temperament, PlayStyle
)
from app.services.database import players_collection
from app.services.token import TokenService

token_service = TokenService()

FIRST_NAMES = ["Lucas", "Gabriel", "Pedro", "Bruno", "Rafael", "Diego", "Carlos", "André", "Marcos", "Felipe",
               "João", "Matheus", "Vinicius", "Neymar", "Ronaldo", "Kaká", "Zico", "Romário", "Rivaldo", "Adriano"]
LAST_NAMES = ["Silva", "Santos", "Oliveira", "Souza", "Lima", "Pereira", "Costa", "Ferreira", "Alves", "Rodrigues",
              "Junior", "Neto", "Filho", "Mendes", "Barbosa", "Ribeiro", "Martins", "Carvalho", "Gomes", "Lopes"]
NATIONALITIES = ["Brasil", "Argentina", "Portugal", "Espanha", "Alemanha", "França", "Itália", "Inglaterra",
                 "Holanda", "Uruguai", "Colômbia", "Chile", "México", "Japão", "Coreia do Sul"]


class PlayerService:
    async def create_player(self, player_data: PlayerCreate) -> Player:
        player = Player(
            name=player_data.name,
            position=player_data.position,
            nationality=player_data.nationality,
            age=player_data.age,
            attributes=player_data.attributes,
            personality=player_data.personality,
        )
        player.price = player.get_market_value()
        result = await players_collection.insert_one(player.model_dump(exclude={"id"}))
        player.id = str(result.inserted_id)
        return player

    async def get_player(self, player_id: str) -> Optional[Player]:
        try:
            data = await players_collection.find_one({"_id": ObjectId(player_id)})
        except:
            return None
        if not data:
            return None
        data["_id"] = str(data["_id"])
        return Player(**data)

    async def get_team_players(self, team_id: str) -> List[Player]:
        cursor = players_collection.find({"team_id": team_id})
        players = []
        async for data in cursor:
            data["_id"] = str(data["_id"])
            players.append(Player(**data))
        return players

    async def get_market_players(
        self,
        position: Optional[Position] = None,
        min_overall: int = 0,
        max_price: Optional[int] = None,
        skip: int = 0,
        limit: int = 20
    ) -> List[Player]:
        query = {"$or": [{"team_id": None}, {"is_listed": True}]}
        if position:
            query["position"] = position
        if max_price:
            query["price"] = {"$lte": max_price}
        
        cursor = players_collection.find(query).skip(skip).limit(limit)
        players = []
        async for data in cursor:
            data["_id"] = str(data["_id"])
            player = Player(**data)
            if player.attributes.overall() >= min_overall:
                players.append(player)
        return players

    async def buy_player(self, player_id: str, team_id: str, user_id: str) -> Optional[Player]:
        player = await self.get_player(player_id)
        if not player:
            return None
        if player.team_id and not player.is_listed:
            return None
        
        # Debit tokens
        tx = await token_service.debit_for_player(user_id, player.price, player_id)
        if not tx:
            return None
        
        # If player had previous owner, credit them
        if player.team_id:
            # TODO: Get previous owner and credit
            pass
        
        # Update player
        await players_collection.update_one(
            {"_id": ObjectId(player_id)},
            {"$set": {"team_id": team_id, "is_listed": False}}
        )
        return await self.get_player(player_id)

    async def list_for_sale(self, player_id: str, team_id: str, price: int) -> Optional[Player]:
        player = await self.get_player(player_id)
        if not player or player.team_id != team_id:
            return None
        
        await players_collection.update_one(
            {"_id": ObjectId(player_id)},
            {"$set": {"is_listed": True, "price": price}}
        )
        return await self.get_player(player_id)

    async def generate_random_player(self, position: Optional[Position] = None) -> Player:
        pos = position or random.choice(list(Position))
        name = f"{random.choice(FIRST_NAMES)} {random.choice(LAST_NAMES)}"
        
        # Generate attributes based on position
        attrs = self._generate_position_attributes(pos)
        personality = PlayerPersonality(
            temperament=random.choice(list(Temperament)),
            play_style=random.choice(list(PlayStyle)),
        )
        
        player_data = PlayerCreate(
            name=name,
            position=pos,
            nationality=random.choice(NATIONALITIES),
            age=random.randint(18, 35),
            attributes=attrs,
            personality=personality,
        )
        return await self.create_player(player_data)

    def _generate_position_attributes(self, position: Position) -> PlayerAttributes:
        base = random.randint(40, 70)
        
        attrs = {
            "speed": base + random.randint(-10, 20),
            "strength": base + random.randint(-10, 20),
            "stamina": base + random.randint(-10, 20),
            "finishing": base + random.randint(-10, 20),
            "passing": base + random.randint(-10, 20),
            "dribbling": base + random.randint(-10, 20),
            "defense": base + random.randint(-10, 20),
            "intelligence": base + random.randint(-10, 20),
            "aggression": base + random.randint(-10, 20),
            "leadership": base + random.randint(-10, 20),
            "creativity": base + random.randint(-10, 20),
        }
        
        # Position-specific boosts
        if position == Position.GK:
            attrs["defense"] += 20
        elif position in [Position.CB, Position.LB, Position.RB]:
            attrs["defense"] += 15
            attrs["strength"] += 10
        elif position in [Position.CDM, Position.CM]:
            attrs["passing"] += 15
            attrs["stamina"] += 10
        elif position == Position.CAM:
            attrs["creativity"] += 15
            attrs["passing"] += 10
        elif position in [Position.LW, Position.RW]:
            attrs["speed"] += 15
            attrs["dribbling"] += 10
        elif position == Position.ST:
            attrs["finishing"] += 20
            attrs["speed"] += 10
        
        # Clamp values
        for key in attrs:
            attrs[key] = max(1, min(100, attrs[key]))
        
        return PlayerAttributes(**attrs)

