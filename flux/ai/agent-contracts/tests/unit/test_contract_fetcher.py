"""
Unit tests for Contract Fetcher.
"""
import pytest
from unittest.mock import AsyncMock, patch, MagicMock
import json

import sys
sys.path.insert(0, str(__file__).replace("/tests/unit/test_contract_fetcher.py", "/src"))

from contract_fetcher.handler import (
    ContractFetcher,
    ContractData,
    create_contract_created_event,
    CHAIN_EXPLORERS,
)


class TestContractData:
    """Tests for ContractData dataclass."""
    
    def test_source_hash_with_source(self):
        """Test hash generation with source code."""
        contract = ContractData(
            chain="ethereum",
            address="0x1234",
            source_code="contract Test {}",
            abi=None,
            bytecode=None,
            contract_name="Test",
            compiler_version="0.8.20",
            is_verified=True,
            fetched_at="2025-01-01T00:00:00Z",
        )
        
        hash1 = contract.source_hash()
        assert len(hash1) == 16
        
        # Same source = same hash
        contract2 = ContractData(
            chain="ethereum",
            address="0x5678",
            source_code="contract Test {}",
            abi=None,
            bytecode=None,
            contract_name="Test",
            compiler_version="0.8.20",
            is_verified=True,
            fetched_at="2025-01-02T00:00:00Z",
        )
        assert contract.source_hash() == contract2.source_hash()
    
    def test_source_hash_with_bytecode(self):
        """Test hash generation with bytecode when no source."""
        contract = ContractData(
            chain="ethereum",
            address="0x1234",
            source_code=None,
            abi=None,
            bytecode="0x6080604052",
            contract_name=None,
            compiler_version=None,
            is_verified=False,
            fetched_at="2025-01-01T00:00:00Z",
        )
        
        hash1 = contract.source_hash()
        assert len(hash1) == 16
    
    def test_to_dict(self):
        """Test serialization to dict."""
        contract = ContractData(
            chain="ethereum",
            address="0x1234",
            source_code="contract Test {}",
            abi=[{"type": "function", "name": "test"}],
            bytecode="0x6080",
            contract_name="Test",
            compiler_version="0.8.20",
            is_verified=True,
            fetched_at="2025-01-01T00:00:00Z",
        )
        
        d = contract.to_dict()
        assert d["chain"] == "ethereum"
        assert d["address"] == "0x1234"
        assert d["is_verified"] is True


class TestContractFetcher:
    """Tests for ContractFetcher class."""
    
    @pytest.mark.asyncio
    async def test_fetch_from_cache(self, mock_redis):
        """Test fetching contract from Redis cache."""
        cached_data = {
            "chain": "ethereum",
            "address": "0x1234",
            "source_code": "contract Cached {}",
            "abi": None,
            "bytecode": None,
            "contract_name": "Cached",
            "compiler_version": "0.8.20",
            "is_verified": True,
            "fetched_at": "2025-01-01T00:00:00Z",
        }
        mock_redis.get.return_value = json.dumps(cached_data)
        
        fetcher = ContractFetcher(redis_client=mock_redis)
        contract = await fetcher.fetch("ethereum", "0x1234")
        
        assert contract.contract_name == "Cached"
        assert contract.is_verified is True
        mock_redis.get.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_fetch_from_api(self, mock_redis):
        """Test fetching contract from Etherscan API."""
        mock_redis.get.return_value = None
        
        fetcher = ContractFetcher(redis_client=mock_redis)
        
        # Mock the HTTP client
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "status": "1",
            "result": [{
                "SourceCode": "contract FromAPI {}",
                "ABI": "[]",
                "ContractName": "FromAPI",
                "CompilerVersion": "v0.8.20",
            }]
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch.object(fetcher.http, 'get', new_callable=AsyncMock) as mock_get:
            mock_get.return_value = mock_response
            contract = await fetcher.fetch("ethereum", "0x1234")
        
        assert contract.contract_name == "FromAPI"
        assert contract.is_verified is True
    
    def test_supported_chains(self):
        """Test that all expected chains are supported."""
        expected_chains = ["ethereum", "bsc", "polygon", "arbitrum", "base", "optimism"]
        for chain in expected_chains:
            assert chain in CHAIN_EXPLORERS


class TestCloudEventCreation:
    """Tests for CloudEvent creation."""
    
    def test_create_contract_created_event(self):
        """Test creating contract.created CloudEvent."""
        contract = ContractData(
            chain="ethereum",
            address="0x1234567890123456789012345678901234567890",
            source_code="contract Test {}",
            abi=[],
            bytecode=None,
            contract_name="Test",
            compiler_version="0.8.20",
            is_verified=True,
            fetched_at="2025-01-01T00:00:00Z",
        )
        
        event = create_contract_created_event(contract)
        
        assert event["type"] == "io.homelab.contract.created"
        assert event["source"] == "/agent-contracts/contract-fetcher"
        assert "ethereum" in event["subject"]
        assert event.data["chain"] == "ethereum"
        assert event.data["is_verified"] is True

