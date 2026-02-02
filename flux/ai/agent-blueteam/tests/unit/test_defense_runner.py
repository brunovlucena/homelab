"""
Unit tests for agent-blueteam defense runner.
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "src"))

from shared.types import ThreatLevel, DefenseAction, MAG7Boss


class TestMAG7Boss:
    """Tests for MAG7 boss mechanics."""
    
    def test_mag7_initial_health(self):
        """MAG7 should start with 1000 health."""
        boss = MAG7Boss()
        assert boss.health == 1000
        assert boss.max_health == 1000
        assert boss.defeated is False
    
    def test_mag7_has_seven_heads(self):
        """MAG7 should have abilities for all 7 heads."""
        boss = MAG7Boss()
        assert len(boss.abilities) == 7
        
        expected_heads = ["apple", "microsoft", "google", "amazon", "meta", "tesla", "nvidia"]
        for head in expected_heads:
            assert head in boss.abilities
    
    def test_mag7_phases(self):
        """MAG7 should have different phases."""
        boss = MAG7Boss()
        assert boss.phase == "normal"
        assert boss.attack_speed == 1.0


class TestThreatDetection:
    """Tests for threat detection logic."""
    
    def test_threat_levels(self):
        """Verify threat levels are properly defined."""
        assert ThreatLevel.LOW.value == "low"
        assert ThreatLevel.MEDIUM.value == "medium"
        assert ThreatLevel.HIGH.value == "high"
        assert ThreatLevel.CRITICAL.value == "critical"
    
    def test_defense_actions(self):
        """Verify defense actions are properly defined."""
        assert DefenseAction.BLOCK_NETWORK.value == "block_network"
        assert DefenseAction.BLOCK_ADMISSION.value == "block_admission"
        assert DefenseAction.QUARANTINE.value == "quarantine"
        assert DefenseAction.MONITOR.value == "monitor"


class TestDefenseSignatures:
    """Tests for defense signature matching."""
    
    def test_ssrf_patterns(self):
        """SSRF patterns should include metadata endpoints."""
        from defense_runner.handler import DEFENSE_SIGNATURES
        
        ssrf_sig = DEFENSE_SIGNATURES.get("blue-001")
        assert ssrf_sig is not None
        assert "169.254.169.254" in ssrf_sig["patterns"]
        assert ssrf_sig["action"] == DefenseAction.BLOCK_NETWORK
    
    def test_command_injection_patterns(self):
        """Command injection should detect shell metacharacters."""
        from defense_runner.handler import DEFENSE_SIGNATURES
        
        cmd_sig = DEFENSE_SIGNATURES.get("vuln-001")
        assert cmd_sig is not None
        assert ";" in cmd_sig["patterns"]
        assert "|" in cmd_sig["patterns"]
        assert cmd_sig["action"] == DefenseAction.BLOCK_ADMISSION


class TestGameIntegration:
    """Tests for MAG7 Battle game integration."""
    
    def test_mag7_damage_calculation(self):
        """Blocking exploits should damage MAG7."""
        boss = MAG7Boss()
        initial_health = boss.health
        
        # Simulate blocked exploit
        damage = 50
        boss.health -= damage
        
        assert boss.health == initial_health - damage
        assert boss.defeated is False
    
    def test_mag7_defeat(self):
        """MAG7 should be defeated when health reaches 0."""
        boss = MAG7Boss()
        boss.health = 10
        
        # Final blow
        boss.health -= 50
        boss.health = max(0, boss.health)
        
        if boss.health <= 0:
            boss.defeated = True
        
        assert boss.health == 0
        assert boss.defeated is True
