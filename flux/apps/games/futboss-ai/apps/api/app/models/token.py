# FutBoss AI - Token/Wallet Model
# Author: Bruno Lucena (bruno@lucena.cloud)

from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from enum import Enum


class TransactionType(str, Enum):
    PURCHASE = "purchase"  # Bought with real money
    SALE = "sale"  # Sold player
    BUY_PLAYER = "buy_player"  # Bought player
    MATCH_REWARD = "match_reward"  # Won match
    BONUS = "bonus"  # Promotional bonus


class TokenTransaction(BaseModel):
    id: Optional[str] = Field(default=None, alias="_id")
    user_id: str
    amount: int  # Positive = credit, Negative = debit
    transaction_type: TransactionType
    description: str
    reference_id: Optional[str] = None  # Related entity (player_id, match_id, etc.)
    created_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        populate_by_name = True


class TokenWallet(BaseModel):
    user_id: str
    balance: int = 1000
    total_earned: int = 1000
    total_spent: int = 0
    transactions: List[TokenTransaction] = Field(default_factory=list)

    def can_afford(self, amount: int) -> bool:
        return self.balance >= amount

    def credit(self, amount: int, tx_type: TransactionType, description: str, ref_id: str = None) -> TokenTransaction:
        self.balance += amount
        self.total_earned += amount
        tx = TokenTransaction(
            user_id=self.user_id,
            amount=amount,
            transaction_type=tx_type,
            description=description,
            reference_id=ref_id,
        )
        self.transactions.append(tx)
        return tx

    def debit(self, amount: int, tx_type: TransactionType, description: str, ref_id: str = None) -> Optional[TokenTransaction]:
        if not self.can_afford(amount):
            return None
        self.balance -= amount
        self.total_spent += amount
        tx = TokenTransaction(
            user_id=self.user_id,
            amount=-amount,
            transaction_type=tx_type,
            description=description,
            reference_id=ref_id,
        )
        self.transactions.append(tx)
        return tx

