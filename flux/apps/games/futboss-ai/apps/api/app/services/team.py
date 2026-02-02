# FutBoss AI - Team Service
# Author: Bruno Lucena (bruno@lucena.cloud)

from typing import Optional, List
from bson import ObjectId

from app.models.team import Team, TeamCreate, Formation
from app.services.database import teams_collection


class TeamService:
    async def create_team(self, team_data: TeamCreate, owner_id: str) -> Team:
        team = Team(
            name=team_data.name,
            owner_id=owner_id,
            formation=team_data.formation,
        )
        result = await teams_collection.insert_one(team.model_dump(exclude={"id"}))
        team.id = str(result.inserted_id)
        return team

    async def get_team(self, team_id: str) -> Optional[Team]:
        try:
            data = await teams_collection.find_one({"_id": ObjectId(team_id)})
        except:
            return None
        if not data:
            return None
        data["_id"] = str(data["_id"])
        return Team(**data)

    async def get_user_team(self, user_id: str) -> Optional[Team]:
        data = await teams_collection.find_one({"owner_id": user_id})
        if not data:
            return None
        data["_id"] = str(data["_id"])
        return Team(**data)

    async def update_formation(self, team_id: str, formation: Formation) -> Optional[Team]:
        await teams_collection.update_one(
            {"_id": ObjectId(team_id)},
            {"$set": {"formation": formation}}
        )
        return await self.get_team(team_id)

    async def add_player(self, team_id: str, player_id: str) -> bool:
        result = await teams_collection.update_one(
            {"_id": ObjectId(team_id)},
            {"$addToSet": {"player_ids": player_id}}
        )
        return result.modified_count > 0

    async def remove_player(self, team_id: str, player_id: str) -> bool:
        result = await teams_collection.update_one(
            {"_id": ObjectId(team_id)},
            {"$pull": {"player_ids": player_id}}
        )
        return result.modified_count > 0

    async def update_stats(self, team_id: str, goals_for: int, goals_against: int) -> Optional[Team]:
        update = {"$inc": {"goals_for": goals_for, "goals_against": goals_against}}
        if goals_for > goals_against:
            update["$inc"]["wins"] = 1
        elif goals_for < goals_against:
            update["$inc"]["losses"] = 1
        else:
            update["$inc"]["draws"] = 1
        
        await teams_collection.update_one({"_id": ObjectId(team_id)}, update)
        return await self.get_team(team_id)

    async def list_teams(self, skip: int = 0, limit: int = 20) -> List[Team]:
        cursor = teams_collection.find().skip(skip).limit(limit)
        teams = []
        async for data in cursor:
            data["_id"] = str(data["_id"])
            teams.append(Team(**data))
        return teams

