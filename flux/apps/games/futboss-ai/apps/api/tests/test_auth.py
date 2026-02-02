# FutBoss AI - Authentication Tests
# Author: Bruno Lucena (bruno@lucena.cloud)

import pytest
from unittest.mock import AsyncMock, patch, MagicMock

from app.services.auth import AuthService


class TestAuthService:
    @pytest.fixture
    def auth_service(self):
        return AuthService()

    def test_hash_password(self, auth_service):
        password = "testpassword123"
        hashed = auth_service.hash_password(password)
        assert hashed != password
        assert hashed.startswith("$2b$")

    def test_verify_password_correct(self, auth_service):
        password = "testpassword123"
        hashed = auth_service.hash_password(password)
        assert auth_service.verify_password(password, hashed) is True

    def test_verify_password_wrong(self, auth_service):
        password = "testpassword123"
        hashed = auth_service.hash_password(password)
        assert auth_service.verify_password("wrongpassword", hashed) is False

    def test_create_token(self, auth_service):
        user_id = "507f1f77bcf86cd799439011"
        token = auth_service.create_token(user_id)
        assert isinstance(token, str)
        assert len(token) > 50  # JWT tokens are long

    def test_decode_token_valid(self, auth_service):
        user_id = "507f1f77bcf86cd799439011"
        token = auth_service.create_token(user_id)
        decoded_id = auth_service.decode_token(token)
        assert decoded_id == user_id

    def test_decode_token_invalid(self, auth_service):
        decoded_id = auth_service.decode_token("invalid.token.here")
        assert decoded_id is None

    def test_decode_token_tampered(self, auth_service):
        user_id = "507f1f77bcf86cd799439011"
        token = auth_service.create_token(user_id)
        # Tamper with token
        tampered = token[:-5] + "XXXXX"
        decoded_id = auth_service.decode_token(tampered)
        assert decoded_id is None

