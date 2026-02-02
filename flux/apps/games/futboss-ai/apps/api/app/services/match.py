# FutBoss AI - Match Service
# Author: Bruno Lucena (bruno@lucena.cloud)

from datetime import datetime
from typing import Optional, List
from bson import ObjectId

from app.models.match import Match, MatchCreate, MatchStatus, MatchEvent, MatchState
from app.services.database import matches_collection


class MatchService:
    async def create_match(self, match_data: MatchCreate) -> Match:
        match = Match(
            home_team_id=match_data.home_team_id,
            away_team_id=match_data.away_team_id,
        )
        result = await matches_collection.insert_one(match.model_dump(exclude={"id"}))
        match.id = str(result.inserted_id)
        return match

    async def get_match(self, match_id: str) -> Optional[Match]:
        try:
            data = await matches_collection.find_one({"_id": ObjectId(match_id)})
        except:
            return None
        if not data:
            return None
        data["_id"] = str(data["_id"])
        return Match(**data)

    async def start_match(self, match_id: str) -> Optional[Match]:
        match = await self.get_match(match_id)
        if not match or match.status != MatchStatus.PENDING:
            return None
        
        await matches_collection.update_one(
            {"_id": ObjectId(match_id)},
            {"$set": {
                "status": MatchStatus.IN_PROGRESS,
                "started_at": datetime.utcnow(),
            }}
        )
        return await self.get_match(match_id)

    async def finish_match(self, match_id: str, home_score: int, away_score: int) -> Optional[Match]:
        match = await self.get_match(match_id)
        if not match or match.status != MatchStatus.IN_PROGRESS:
            return None
        
        await matches_collection.update_one(
            {"_id": ObjectId(match_id)},
            {"$set": {
                "status": MatchStatus.FINISHED,
                "home_score": home_score,
                "away_score": away_score,
                "finished_at": datetime.utcnow(),
            }}
        )
        return await self.get_match(match_id)

    async def add_event(self, match_id: str, event: MatchEvent) -> Optional[Match]:
        await matches_collection.update_one(
            {"_id": ObjectId(match_id)},
            {"$push": {"events": event.model_dump()}}
        )
        return await self.get_match(match_id)

    async def update_state(self, match_id: str, state: MatchState) -> Optional[Match]:
        await matches_collection.update_one(
            {"_id": ObjectId(match_id)},
            {"$set": {"state": state.model_dump()}}
        )
        return await self.get_match(match_id)

    async def get_team_matches(
        self,
        team_id: str,
        status: Optional[MatchStatus] = None,
        skip: int = 0,
        limit: int = 20
    ) -> List[Match]:
        query = {"$or": [{"home_team_id": team_id}, {"away_team_id": team_id}]}
        if status:
            query["status"] = status
        
        cursor = matches_collection.find(query).sort("created_at", -1).skip(skip).limit(limit)
        matches = []
        async for data in cursor:
            data["_id"] = str(data["_id"])
            matches.append(Match(**data))
        return matches

    async def list_matches(
        self,
        status: Optional[MatchStatus] = None,
        skip: int = 0,
        limit: int = 20
    ) -> List[Match]:
        query = {}
        if status:
            query["status"] = status
        
        cursor = matches_collection.find(query).sort("created_at", -1).skip(skip).limit(limit)
        matches = []
        async for data in cursor:
            data["_id"] = str(data["_id"])
            matches.append(Match(**data))
        return matches

