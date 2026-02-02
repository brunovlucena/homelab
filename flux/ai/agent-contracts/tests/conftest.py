"""
Pytest configuration and fixtures for agent-contracts tests.
"""
import os
import pytest
import asyncio
from unittest.mock import AsyncMock, MagicMock

# Set test environment
os.environ["TESTING"] = "true"
os.environ["OLLAMA_URL"] = "http://localhost:11434"
os.environ["ANVIL_URL"] = "http://localhost:8545"


@pytest.fixture(scope="session")
def event_loop():
    """Create event loop for async tests."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def mock_redis():
    """Mock Redis client."""
    redis = AsyncMock()
    redis.get.return_value = None
    redis.setex.return_value = True
    redis.ping.return_value = True
    return redis


@pytest.fixture
def mock_httpx_client():
    """Mock httpx client for API calls."""
    client = AsyncMock()
    return client


@pytest.fixture
def sample_contract_source():
    """Sample vulnerable Solidity contract for testing."""
    return '''
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract VulnerableBank {
    mapping(address => uint256) public balances;
    
    function deposit() external payable {
        balances[msg.sender] += msg.value;
    }
    
    // Vulnerable to reentrancy
    function withdraw(uint256 amount) external {
        require(balances[msg.sender] >= amount, "Insufficient balance");
        
        // Vulnerable: state change after external call
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success, "Transfer failed");
        
        balances[msg.sender] -= amount;
    }
    
    function getBalance() external view returns (uint256) {
        return address(this).balance;
    }
}
'''


@pytest.fixture
def sample_vulnerability():
    """Sample vulnerability data."""
    return {
        "type": "reentrancy",
        "severity": "critical",
        "confidence": 0.95,
        "location": "VulnerableBank.sol:15",
        "description": "The withdraw function is vulnerable to reentrancy attacks. The external call is made before updating the balance.",
        "recommendation": "Use the checks-effects-interactions pattern or a reentrancy guard.",
    }


@pytest.fixture
def sample_cloudevent_data():
    """Sample CloudEvent data for contract scanning."""
    return {
        "specversion": "1.0",
        "type": "io.homelab.contract.created",
        "source": "/test",
        "id": "test-event-123",
        "data": {
            "chain": "ethereum",
            "address": "0x1234567890123456789012345678901234567890",
            "source_code": "contract Test {}",
            "is_verified": True,
        }
    }


@pytest.fixture
def mock_ollama_response():
    """Mock Ollama LLM response."""
    return {
        "response": '''```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";

contract ExploitTest is Test {
    function testExploit() public {
        // Exploit logic here
        assertTrue(true);
    }
}
```''',
        "eval_count": 150,
    }


@pytest.fixture
def mock_slither_output():
    """Mock Slither analysis output."""
    return {
        "success": True,
        "results": {
            "detectors": [
                {
                    "check": "reentrancy-eth",
                    "impact": "High",
                    "confidence": "High",
                    "description": "Reentrancy in VulnerableBank.withdraw()",
                    "elements": [
                        {
                            "source_mapping": {
                                "filename_relative": "VulnerableBank.sol",
                                "lines": [15, 16, 17, 18, 19, 20],
                            }
                        }
                    ],
                    "recommendation": "Use checks-effects-interactions pattern",
                }
            ]
        }
    }

