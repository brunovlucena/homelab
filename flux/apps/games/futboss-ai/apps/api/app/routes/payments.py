# FutBoss AI - Payment Routes (PIX & Bitcoin)
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import APIRouter, HTTPException, Depends, Request
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel
from typing import Optional

from app.services.auth import AuthService
from app.services.payment import PaymentService, PaymentMethod

router = APIRouter()
security = HTTPBearer()
auth_service = AuthService()
payment_service = PaymentService()


class PaymentRequest(BaseModel):
    amount_brl: float  # Amount in BRL
    method: PaymentMethod


class PaymentResponse(BaseModel):
    payment_id: str
    method: PaymentMethod
    amount_brl: float
    tokens_to_receive: int
    status: str
    pix_code: Optional[str] = None  # PIX copy-paste code
    pix_qr_code: Optional[str] = None  # PIX QR code base64
    bitcoin_address: Optional[str] = None
    bitcoin_amount: Optional[float] = None


async def get_current_user_id(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    user = await auth_service.get_current_user(credentials.credentials)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user.id


@router.post("/create", response_model=PaymentResponse)
async def create_payment(payment: PaymentRequest, user_id: str = Depends(get_current_user_id)):
    """Create a new payment (PIX or Bitcoin)"""
    if payment.amount_brl < 10:
        raise HTTPException(status_code=400, detail="Minimum payment is R$10")
    
    result = await payment_service.create_payment(user_id, payment.amount_brl, payment.method)
    if not result:
        raise HTTPException(status_code=500, detail="Failed to create payment")
    
    return result


@router.get("/{payment_id}", response_model=PaymentResponse)
async def get_payment_status(payment_id: str, user_id: str = Depends(get_current_user_id)):
    """Get payment status"""
    result = await payment_service.get_payment(payment_id, user_id)
    if not result:
        raise HTTPException(status_code=404, detail="Payment not found")
    return result


@router.post("/webhook/pix")
async def pix_webhook(request: Request):
    """Webhook for PIX payment notifications"""
    body = await request.json()
    await payment_service.process_pix_webhook(body)
    return {"status": "ok"}


@router.post("/webhook/bitcoin")
async def bitcoin_webhook(request: Request):
    """Webhook for Bitcoin payment notifications (BTCPay)"""
    body = await request.json()
    await payment_service.process_bitcoin_webhook(body)
    return {"status": "ok"}


@router.get("/rates", response_model=dict)
async def get_rates():
    """Get current token rates"""
    return {
        "token_price_brl": 0.01,  # 1 FutCoin = R$0.01
        "min_purchase_brl": 10.0,
        "tokens_per_10_brl": 1000,
    }

