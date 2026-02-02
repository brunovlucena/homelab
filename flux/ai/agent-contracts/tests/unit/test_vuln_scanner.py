"""
Unit tests for Vulnerability Scanner.
"""
import pytest
from unittest.mock import AsyncMock, patch, MagicMock
import json

import sys
sys.path.insert(0, str(__file__).replace("/tests/unit/test_vuln_scanner.py", "/src"))

from vuln_scanner.handler import (
    VulnerabilityScanner,
    SlitherAnalyzer,
    LLMAnalyzer,
    Vulnerability,
    ScanResult,
    Severity,
    VulnType,
    create_vuln_found_event,
)


class TestVulnerability:
    """Tests for Vulnerability dataclass."""
    
    def test_to_dict(self):
        """Test serialization to dict."""
        vuln = Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.CRITICAL,
            confidence=0.95,
            location="Contract.sol:15",
            description="Reentrancy vulnerability",
            recommendation="Use reentrancy guard",
            analyzer="slither",
        )
        
        d = vuln.to_dict()
        assert d["type"] == "reentrancy"
        assert d["severity"] == "critical"
        assert d["confidence"] == 0.95


class TestScanResult:
    """Tests for ScanResult dataclass."""
    
    def test_has_critical(self):
        """Test critical vulnerability detection."""
        result = ScanResult(chain="ethereum", address="0x1234")
        assert result.has_critical is False
        
        result.vulnerabilities.append(Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.CRITICAL,
            confidence=0.9,
            location="test",
            description="test",
            recommendation="test",
            analyzer="test",
        ))
        assert result.has_critical is True
    
    def test_max_severity(self):
        """Test max severity calculation."""
        result = ScanResult(chain="ethereum", address="0x1234")
        assert result.max_severity is None
        
        result.vulnerabilities.append(Vulnerability(
            type=VulnType.OTHER,
            severity=Severity.LOW,
            confidence=0.5,
            location="test",
            description="test",
            recommendation="test",
            analyzer="test",
        ))
        assert result.max_severity == Severity.LOW
        
        result.vulnerabilities.append(Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.HIGH,
            confidence=0.9,
            location="test",
            description="test",
            recommendation="test",
            analyzer="test",
        ))
        assert result.max_severity == Severity.HIGH


class TestSlitherAnalyzer:
    """Tests for Slither static analyzer."""
    
    def test_map_detector_to_type(self):
        """Test detector name to VulnType mapping."""
        analyzer = SlitherAnalyzer()
        
        assert analyzer._map_detector_to_type("reentrancy-eth") == VulnType.REENTRANCY
        assert analyzer._map_detector_to_type("reentrancy-no-eth") == VulnType.REENTRANCY
        assert analyzer._map_detector_to_type("arbitrary-send") == VulnType.ARBITRARY_CALL
        assert analyzer._map_detector_to_type("controlled-delegatecall") == VulnType.DELEGATECALL
        assert analyzer._map_detector_to_type("unknown-detector") == VulnType.OTHER
    
    def test_map_confidence(self):
        """Test confidence string to float mapping."""
        analyzer = SlitherAnalyzer()
        
        assert analyzer._map_confidence("High") == 0.9
        assert analyzer._map_confidence("Medium") == 0.7
        assert analyzer._map_confidence("Low") == 0.5
        assert analyzer._map_confidence("Unknown") == 0.5
    
    def test_parse_results(self, mock_slither_output):
        """Test parsing Slither JSON output."""
        analyzer = SlitherAnalyzer()
        vulns = analyzer._parse_results(mock_slither_output)
        
        assert len(vulns) == 1
        assert vulns[0].type == VulnType.REENTRANCY
        assert vulns[0].severity == Severity.HIGH
        assert vulns[0].analyzer == "slither"


class TestLLMAnalyzer:
    """Tests for LLM-based analyzer."""
    
    def test_build_prompt(self, sample_contract_source, sample_vulnerability):
        """Test prompt construction."""
        analyzer = LLMAnalyzer()
        prompt = analyzer._build_prompt(
            sample_contract_source,
            [Vulnerability(**{
                **sample_vulnerability,
                "type": VulnType.REENTRANCY,
                "severity": Severity.CRITICAL,
                "analyzer": "slither"
            })]
        )
        
        assert "smart contract security auditor" in prompt.lower()
        assert "VulnerableBank" in prompt
        assert "reentrancy" in prompt.lower()
    
    def test_should_fallback(self):
        """Test cloud fallback logic."""
        analyzer = LLMAnalyzer()
        
        # Should fallback for critical
        critical_vuln = Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.CRITICAL,
            confidence=0.9,
            location="test",
            description="test",
            recommendation="test",
            analyzer="test",
        )
        assert analyzer._should_fallback("response", [critical_vuln]) is True
        
        # Should fallback for low confidence
        low_conf_vuln = Vulnerability(
            type=VulnType.OTHER,
            severity=Severity.MEDIUM,
            confidence=0.5,
            location="test",
            description="test",
            recommendation="test",
            analyzer="test",
        )
        assert analyzer._should_fallback("response", [low_conf_vuln]) is True
        
        # No fallback for high confidence non-critical
        normal_vuln = Vulnerability(
            type=VulnType.OTHER,
            severity=Severity.MEDIUM,
            confidence=0.85,
            location="test",
            description="test",
            recommendation="test",
            analyzer="test",
        )
        assert analyzer._should_fallback("response", [normal_vuln]) is False


class TestVulnerabilityScanner:
    """Tests for main VulnerabilityScanner class."""
    
    def test_merge_findings_deduplication(self):
        """Test that duplicate findings are merged."""
        scanner = VulnerabilityScanner()
        
        vuln1 = Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.HIGH,
            confidence=0.7,
            location="test.sol:10",
            description="From slither",
            recommendation="Fix it",
            analyzer="slither",
        )
        
        vuln2 = Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.HIGH,
            confidence=0.9,  # Higher confidence
            location="test.sol:10",
            description="From LLM",
            recommendation="Fix it better",
            analyzer="llm",
        )
        
        merged = scanner._merge_findings([vuln1], [vuln2])
        
        # Should keep only one, with higher confidence
        assert len(merged) == 1
        assert merged[0].confidence == 0.9
        assert merged[0].analyzer == "llm"
    
    def test_merge_findings_sorting(self):
        """Test that findings are sorted by severity."""
        scanner = VulnerabilityScanner()
        
        low = Vulnerability(
            type=VulnType.OTHER,
            severity=Severity.LOW,
            confidence=0.9,
            location="a.sol:1",
            description="Low",
            recommendation="",
            analyzer="test",
        )
        
        critical = Vulnerability(
            type=VulnType.REENTRANCY,
            severity=Severity.CRITICAL,
            confidence=0.9,
            location="b.sol:1",
            description="Critical",
            recommendation="",
            analyzer="test",
        )
        
        merged = scanner._merge_findings([low], [critical])
        
        assert merged[0].severity == Severity.CRITICAL
        assert merged[1].severity == Severity.LOW


class TestCloudEventCreation:
    """Tests for CloudEvent creation."""
    
    def test_create_vuln_found_event(self):
        """Test creating vuln.found CloudEvent."""
        result = ScanResult(
            chain="ethereum",
            address="0x1234",
            vulnerabilities=[
                Vulnerability(
                    type=VulnType.REENTRANCY,
                    severity=Severity.CRITICAL,
                    confidence=0.95,
                    location="test.sol:10",
                    description="Reentrancy",
                    recommendation="Fix",
                    analyzer="slither",
                )
            ],
            scan_duration_seconds=5.5,
            analyzers_used=["slither", "llm"],
        )
        
        event = create_vuln_found_event(result)
        
        assert event["type"] == "io.homelab.vuln.found"
        assert event["source"] == "/agent-contracts/vuln-scanner"
        assert event.data["max_severity"] == "critical"
        assert len(event.data["vulnerabilities"]) == 1

