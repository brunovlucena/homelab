# FutBoss AI - Payment Service (PIX & Bitcoin)
# Author: Bruno Lucena (bruno@lucena.cloud)

import uuid
from typing import Optional
from enum import Enum
from datetime import datetime
from bson import ObjectId

from app.config import get_settings
from app.services.database import payments_collection
from app.services.token import TokenService, TransactionType

settings = get_settings()
token_service = TokenService()


class PaymentMethod(str, Enum):
    PIX = "pix"
    BITCOIN = "bitcoin"


class PaymentStatus(str, Enum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    EXPIRED = "expired"
    FAILED = "failed"


class PaymentService:
    def calculate_tokens(self, amount_brl: float) -> int:
        """Calculate tokens for BRL amount"""
        return int(amount_brl / settings.token_to_brl_rate)

    async def create_payment(self, user_id: str, amount_brl: float, method: PaymentMethod) -> Optional[dict]:
        payment_id = str(uuid.uuid4())
        tokens = self.calculate_tokens(amount_brl)
        
        payment_data = {
            "payment_id": payment_id,
            "user_id": user_id,
            "method": method,
            "amount_brl": amount_brl,
            "tokens_to_receive": tokens,
            "status": PaymentStatus.PENDING,
            "created_at": datetime.utcnow(),
        }
        
        if method == PaymentMethod.PIX:
            # Generate PIX code (mock - in production integrate with EfiBank/MercadoPago)
            pix_code = f"00020126580014br.gov.bcb.pix0136{uuid.uuid4()}5204000053039865802BR5925FUTBOSS AI6009SAO PAULO62070503***6304"
            payment_data["pix_code"] = pix_code
            payment_data["pix_qr_code"] = None  # Would be base64 QR in production
        
        elif method == PaymentMethod.BITCOIN:
            # Generate Bitcoin invoice (mock - in production integrate with BTCPay)
            btc_rate = 350000  # Mock BTC/BRL rate
            btc_amount = amount_brl / btc_rate
            payment_data["bitcoin_address"] = f"bc1q{uuid.uuid4().hex[:38]}"
            payment_data["bitcoin_amount"] = round(btc_amount, 8)
        
        await payments_collection.insert_one(payment_data)
        
        return {
            "payment_id": payment_id,
            "method": method,
            "amount_brl": amount_brl,
            "tokens_to_receive": tokens,
            "status": PaymentStatus.PENDING,
            "pix_code": payment_data.get("pix_code"),
            "pix_qr_code": payment_data.get("pix_qr_code"),
            "bitcoin_address": payment_data.get("bitcoin_address"),
            "bitcoin_amount": payment_data.get("bitcoin_amount"),
        }

    async def get_payment(self, payment_id: str, user_id: str) -> Optional[dict]:
        data = await payments_collection.find_one({"payment_id": payment_id, "user_id": user_id})
        if not data:
            return None
        return {
            "payment_id": data["payment_id"],
            "method": data["method"],
            "amount_brl": data["amount_brl"],
            "tokens_to_receive": data["tokens_to_receive"],
            "status": data["status"],
            "pix_code": data.get("pix_code"),
            "pix_qr_code": data.get("pix_qr_code"),
            "bitcoin_address": data.get("bitcoin_address"),
            "bitcoin_amount": data.get("bitcoin_amount"),
        }

    async def confirm_payment(self, payment_id: str) -> bool:
        data = await payments_collection.find_one({"payment_id": payment_id})
        if not data or data["status"] != PaymentStatus.PENDING:
            return False
        
        # Credit tokens to user
        await token_service.credit(
            data["user_id"],
            data["tokens_to_receive"],
            TransactionType.PURCHASE,
            f"Token purchase via {data['method']}",
            payment_id
        )
        
        # Update payment status
        await payments_collection.update_one(
            {"payment_id": payment_id},
            {"$set": {"status": PaymentStatus.CONFIRMED, "confirmed_at": datetime.utcnow()}}
        )
        return True

    async def process_pix_webhook(self, payload: dict) -> bool:
        """Process PIX webhook from payment provider"""
        # In production, validate webhook signature
        payment_id = payload.get("txid") or payload.get("payment_id")
        if not payment_id:
            return False
        return await self.confirm_payment(payment_id)

    async def process_bitcoin_webhook(self, payload: dict) -> bool:
        """Process BTCPay webhook"""
        # In production, validate webhook signature
        payment_id = payload.get("invoiceId") or payload.get("payment_id")
        status = payload.get("type") or payload.get("status")
        
        if status in ["InvoiceSettled", "confirmed", "complete"]:
            return await self.confirm_payment(payment_id)
        return False

