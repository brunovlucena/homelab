# FutBoss AI - Token Service
# Author: Bruno Lucena (bruno@lucena.cloud)

from typing import Optional, List
from bson import ObjectId

from app.models.token import TokenWallet, TokenTransaction, TransactionType
from app.services.database import users_collection, transactions_collection


class TokenService:
    async def get_wallet(self, user_id: str) -> TokenWallet:
        user = await users_collection.find_one({"_id": ObjectId(user_id)})
        if not user:
            return TokenWallet(user_id=user_id, balance=0)
        
        return TokenWallet(
            user_id=user_id,
            balance=user.get("tokens", 0),
            total_earned=user.get("total_earned", user.get("tokens", 0)),
            total_spent=user.get("total_spent", 0),
        )

    async def get_transactions(self, user_id: str, skip: int = 0, limit: int = 50) -> List[TokenTransaction]:
        cursor = transactions_collection.find({"user_id": user_id}).sort("created_at", -1).skip(skip).limit(limit)
        transactions = []
        async for data in cursor:
            data["_id"] = str(data["_id"])
            transactions.append(TokenTransaction(**data))
        return transactions

    async def credit(
        self,
        user_id: str,
        amount: int,
        tx_type: TransactionType,
        description: str,
        reference_id: str = None
    ) -> TokenTransaction:
        tx = TokenTransaction(
            user_id=user_id,
            amount=amount,
            transaction_type=tx_type,
            description=description,
            reference_id=reference_id,
        )
        await transactions_collection.insert_one(tx.model_dump(exclude={"id"}))
        await users_collection.update_one(
            {"_id": ObjectId(user_id)},
            {"$inc": {"tokens": amount, "total_earned": amount}}
        )
        return tx

    async def debit(
        self,
        user_id: str,
        amount: int,
        tx_type: TransactionType,
        description: str,
        reference_id: str = None
    ) -> Optional[TokenTransaction]:
        wallet = await self.get_wallet(user_id)
        if not wallet.can_afford(amount):
            return None
        
        tx = TokenTransaction(
            user_id=user_id,
            amount=-amount,
            transaction_type=tx_type,
            description=description,
            reference_id=reference_id,
        )
        await transactions_collection.insert_one(tx.model_dump(exclude={"id"}))
        await users_collection.update_one(
            {"_id": ObjectId(user_id)},
            {"$inc": {"tokens": -amount, "total_spent": amount}}
        )
        return tx

    async def debit_for_player(self, user_id: str, amount: int, player_id: str) -> Optional[TokenTransaction]:
        return await self.debit(
            user_id,
            amount,
            TransactionType.BUY_PLAYER,
            f"Purchased player",
            player_id
        )

    async def credit_for_sale(self, user_id: str, amount: int, player_id: str) -> TokenTransaction:
        return await self.credit(
            user_id,
            amount,
            TransactionType.SALE,
            f"Sold player",
            player_id
        )

    async def credit_match_reward(self, user_id: str, amount: int, match_id: str) -> TokenTransaction:
        return await self.credit(
            user_id,
            amount,
            TransactionType.MATCH_REWARD,
            f"Match reward",
            match_id
        )

    async def transfer(self, from_user_id: str, to_user_id: str, amount: int) -> Optional[TokenTransaction]:
        debit_tx = await self.debit(
            from_user_id,
            amount,
            TransactionType.SALE,
            f"Transfer to user",
            to_user_id
        )
        if not debit_tx:
            return None
        
        await self.credit(
            to_user_id,
            amount,
            TransactionType.PURCHASE,
            f"Transfer from user",
            from_user_id
        )
        return debit_tx

