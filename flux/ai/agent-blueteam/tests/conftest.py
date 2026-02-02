"""
Test fixtures for agent-blueteam tests.
"""
import pytest
import sys
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))


@pytest.fixture
def sample_exploit_event():
    """Sample exploit event from agent-redteam."""
    return {
        "type": "io.homelab.exploit.success",
        "source": "/agent-redteam/exploit-runner",
        "exploit_id": "vuln-001",
        "namespace": "redteam-test",
        "payload": {
            "name": "Command Injection via Git URL",
            "severity": "critical",
            "status": "success",
        }
    }


@pytest.fixture
def sample_mag7_event():
    """Sample MAG7 game event."""
    return {
        "type": "io.homelab.mag7.attack",
        "source": "/demo-mag7-battle",
        "data": {
            "attack_type": "gpu_meltdown",
            "head": "nvidia",
            "damage": 50,
        }
    }


@pytest.fixture
def sample_game_start_event():
    """Sample game start event."""
    return {
        "type": "io.homelab.demo.game.start",
        "source": "/demo-mag7-battle",
        "data": {}
    }
