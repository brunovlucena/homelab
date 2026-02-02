"""
CloudEvent handler for contract fetching.
"""
import os
import json
import hashlib
from typing import Optional
from dataclasses import dataclass, asdict
from datetime import datetime, timezone

import httpx
import structlog
from cloudevents.http import CloudEvent, to_structured

from shared.metrics import CONTRACTS_FETCHED

logger = structlog.get_logger()

# Supported chains and their explorer APIs
CHAIN_EXPLORERS = {
    "ethereum": {
        "api_url": "https://api.etherscan.io/api",
        "api_key_env": "ETHERSCAN_API_KEY",
    },
    "bsc": {
        "api_url": "https://api.bscscan.com/api",
        "api_key_env": "BSCSCAN_API_KEY",
    },
    "polygon": {
        "api_url": "https://api.polygonscan.com/api",
        "api_key_env": "POLYGONSCAN_API_KEY",
    },
    "arbitrum": {
        "api_url": "https://api.arbiscan.io/api",
        "api_key_env": "ARBISCAN_API_KEY",
    },
    "base": {
        "api_url": "https://api.basescan.org/api",
        "api_key_env": "BASESCAN_API_KEY",
    },
    "optimism": {
        "api_url": "https://api-optimistic.etherscan.io/api",
        "api_key_env": "OPTIMISM_API_KEY",
    },
}


@dataclass
class ContractData:
    """Fetched contract data."""
    chain: str
    address: str
    source_code: Optional[str]
    abi: Optional[list]
    bytecode: Optional[str]
    contract_name: Optional[str]
    compiler_version: Optional[str]
    is_verified: bool
    fetched_at: str
    
    def to_dict(self) -> dict:
        return asdict(self)
    
    def source_hash(self) -> str:
        """Generate hash of source code for deduplication."""
        if self.source_code:
            return hashlib.sha256(self.source_code.encode()).hexdigest()[:16]
        return hashlib.sha256(self.bytecode.encode()).hexdigest()[:16] if self.bytecode else "unknown"


class ContractFetcher:
    """Fetches smart contract data from block explorers."""
    
    def __init__(self, redis_client=None, s3_client=None):
        self.redis = redis_client
        self.s3 = s3_client
        self.http = httpx.AsyncClient(timeout=30.0)
    
    async def fetch(self, chain: str, address: str) -> ContractData:
        """
        Fetch contract from explorer API.
        
        Args:
            chain: Blockchain identifier (ethereum, bsc, etc.)
            address: Contract address (checksummed)
            
        Returns:
            ContractData with source code and metadata
        """
        log = logger.bind(chain=chain, address=address)
        
        if chain not in CHAIN_EXPLORERS:
            raise ValueError(f"Unsupported chain: {chain}")
        
        # Check cache first
        cache_key = f"contract:{chain}:{address}:metadata"
        if self.redis:
            cached = await self.redis.get(cache_key)
            if cached:
                log.info("contract_cache_hit")
                CONTRACTS_FETCHED.labels(chain=chain, status="cache_hit", source="redis").inc()
                return ContractData(**json.loads(cached))
        
        explorer = CHAIN_EXPLORERS[chain]
        api_key = os.getenv(explorer["api_key_env"], "")
        
        # Fetch source code
        source_data = await self._fetch_source(explorer["api_url"], address, api_key)
        
        # Fetch bytecode if source not verified
        bytecode = None
        if not source_data.get("SourceCode"):
            bytecode = await self._fetch_bytecode(chain, address)
        
        contract = ContractData(
            chain=chain,
            address=address,
            source_code=source_data.get("SourceCode"),
            abi=json.loads(source_data.get("ABI", "[]")) if source_data.get("ABI") != "Contract source code not verified" else None,
            bytecode=bytecode,
            contract_name=source_data.get("ContractName"),
            compiler_version=source_data.get("CompilerVersion"),
            is_verified=bool(source_data.get("SourceCode")),
            fetched_at=datetime.now(timezone.utc).isoformat(),
        )
        
        # Cache result
        if self.redis:
            await self.redis.setex(cache_key, 86400, json.dumps(contract.to_dict()))  # 24h TTL
        
        status = "verified" if contract.is_verified else "unverified"
        CONTRACTS_FETCHED.labels(chain=chain, status=status, source="api").inc()
        log.info("contract_fetched", verified=contract.is_verified)
        
        return contract
    
    async def _fetch_source(self, api_url: str, address: str, api_key: str) -> dict:
        """Fetch source code from explorer API."""
        params = {
            "module": "contract",
            "action": "getsourcecode",
            "address": address,
            "apikey": api_key,
        }
        
        response = await self.http.get(api_url, params=params)
        response.raise_for_status()
        
        data = response.json()
        if data.get("status") == "1" and data.get("result"):
            return data["result"][0]
        return {}
    
    async def _fetch_bytecode(self, chain: str, address: str) -> Optional[str]:
        """Fetch bytecode directly from RPC."""
        rpc_url = os.getenv(f"{chain.upper()}_RPC_URL")
        if not rpc_url:
            return None
        
        payload = {
            "jsonrpc": "2.0",
            "method": "eth_getCode",
            "params": [address, "latest"],
            "id": 1,
        }
        
        response = await self.http.post(rpc_url, json=payload)
        response.raise_for_status()
        
        result = response.json().get("result")
        return result if result and result != "0x" else None
    
    async def close(self):
        await self.http.aclose()


def create_contract_created_event(contract: ContractData) -> CloudEvent:
    """Create CloudEvent for new contract."""
    attributes = {
        "type": "io.homelab.contract.created",
        "source": "/agent-contracts/contract-fetcher",
        "subject": f"{contract.chain}/{contract.address}",
    }
    
    data = {
        "chain": contract.chain,
        "address": contract.address,
        "source_code": contract.source_code,
        "abi": contract.abi,
        "bytecode": contract.bytecode,
        "contract_name": contract.contract_name,
        "is_verified": contract.is_verified,
        "source_hash": contract.source_hash(),
    }
    
    return CloudEvent(attributes, data)

