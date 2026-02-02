# FutBoss AI - API Endpoint Tests
# Author: Bruno Lucena (bruno@lucena.cloud)

import pytest


class TestHealthEndpoints:
    def test_root_endpoint(self, client):
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "FutBoss AI API"
        assert data["version"] == "1.0.0"

    def test_health_endpoint(self, client):
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestPaymentRates:
    def test_get_rates(self, client):
        response = client.get("/api/payments/rates")
        assert response.status_code == 200
        data = response.json()
        assert "token_price_brl" in data
        assert "min_purchase_brl" in data
        assert data["min_purchase_brl"] == 10.0

