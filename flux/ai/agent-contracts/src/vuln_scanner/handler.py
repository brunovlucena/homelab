"""
Vulnerability scanner CloudEvent handler.

Emits Prometheus metrics that are consumed by Alertmanager for routing to notifi-services.
"""
import os
import json
import asyncio
import tempfile
from typing import Optional
from pathlib import Path

import structlog
from cloudevents.http import CloudEvent

# Use shared types and metrics
from shared.types import Severity, VulnType, Vulnerability, ScanResult
from shared.metrics import (
    CONTRACTS_SCANNED,
    VULNERABILITIES_FOUND,
    VULNERABILITIES_CRITICAL,
    VULNERABILITIES_HIGH,
    SCAN_DURATION,
    ACTIVE_SCANS,
    LLM_INFERENCE_DURATION,
    API_CALLS,
)

logger = structlog.get_logger()


class SlitherAnalyzer:
    """Slither static analysis integration."""
    
    def __init__(self, timeout: int = 60):
        self.timeout = timeout
    
    async def analyze(self, source_code: str, contract_name: str) -> list[Vulnerability]:
        """Run Slither analysis on source code."""
        vulnerabilities = []
        
        with tempfile.TemporaryDirectory() as tmpdir:
            # Write source to temp file
            source_file = Path(tmpdir) / f"{contract_name}.sol"
            source_file.write_text(source_code)
            
            try:
                # Run slither
                proc = await asyncio.create_subprocess_exec(
                    "slither",
                    str(source_file),
                    "--json", "-",
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE,
                )
                
                stdout, stderr = await asyncio.wait_for(
                    proc.communicate(),
                    timeout=self.timeout
                )
                
                if stdout:
                    results = json.loads(stdout.decode())
                    vulnerabilities = self._parse_results(results)
                    
            except asyncio.TimeoutError:
                logger.warning("slither_timeout", contract=contract_name)
            except Exception as e:
                logger.error("slither_error", error=str(e))
        
        return vulnerabilities
    
    def _parse_results(self, results: dict) -> list[Vulnerability]:
        """Parse Slither JSON output into Vulnerability objects."""
        vulnerabilities = []
        
        severity_map = {
            "High": Severity.HIGH,
            "Medium": Severity.MEDIUM,
            "Low": Severity.LOW,
            "Informational": Severity.INFO,
        }
        
        for detector in results.get("results", {}).get("detectors", []):
            vuln = Vulnerability(
                type=self._map_detector_to_type(detector.get("check", "")),
                severity=severity_map.get(detector.get("impact", ""), Severity.INFO),
                confidence=self._map_confidence(detector.get("confidence", "")),
                location=detector.get("elements", [{}])[0].get("source_mapping", {}).get("filename_relative", "unknown"),
                description=detector.get("description", ""),
                recommendation=detector.get("recommendation", "Review the flagged code."),
                analyzer="slither",
            )
            vulnerabilities.append(vuln)
        
        return vulnerabilities
    
    def _map_detector_to_type(self, check: str) -> VulnType:
        """Map Slither detector name to VulnType."""
        mapping = {
            "reentrancy": VulnType.REENTRANCY,
            "reentrancy-eth": VulnType.REENTRANCY,
            "reentrancy-no-eth": VulnType.REENTRANCY,
            "arbitrary-send": VulnType.ARBITRARY_CALL,
            "controlled-delegatecall": VulnType.DELEGATECALL,
            "suicidal": VulnType.ACCESS_CONTROL,
            "unprotected-upgrade": VulnType.ACCESS_CONTROL,
        }
        
        for key, vuln_type in mapping.items():
            if key in check.lower():
                return vuln_type
        return VulnType.OTHER
    
    def _map_confidence(self, confidence: str) -> float:
        """Map Slither confidence to float."""
        mapping = {"High": 0.9, "Medium": 0.7, "Low": 0.5}
        return mapping.get(confidence, 0.5)


class LLMAnalyzer:
    """LLM-based vulnerability analysis."""
    
    def __init__(self, ollama_url: str = None, anthropic_key: str = None):
        self.ollama_url = ollama_url or os.getenv("OLLAMA_URL", "http://ollama-native.ollama.svc.cluster.local:11434")
        self.anthropic_key = anthropic_key or os.getenv("ANTHROPIC_API_KEY")
        self.confidence_threshold = 0.8
    
    async def analyze(
        self, 
        source_code: str, 
        existing_findings: list[Vulnerability],
        use_cloud_fallback: bool = False
    ) -> list[Vulnerability]:
        """
        Use LLM to analyze contract and validate/enhance findings.
        """
        prompt = self._build_prompt(source_code, existing_findings)
        
        # Try local inference first
        with LLM_INFERENCE_DURATION.labels(model="ollama", operation="vuln_analysis").time():
            response = await self._query_ollama(prompt)
        
        # Fallback to Claude for complex cases
        if use_cloud_fallback and self._should_fallback(response, existing_findings):
            API_CALLS.labels(service="anthropic", status="called").inc()
            with LLM_INFERENCE_DURATION.labels(model="claude", operation="vuln_analysis").time():
                response = await self._query_claude(prompt)
        
        return self._parse_llm_response(response)
    
    def _build_prompt(self, source_code: str, findings: list[Vulnerability]) -> str:
        """Build analysis prompt for LLM."""
        findings_text = "\n".join([
            f"- {f.type.value}: {f.description} (confidence: {f.confidence})"
            for f in findings
        ]) if findings else "No prior findings."
        
        return f"""You are a smart contract security auditor. Analyze the following Solidity code for vulnerabilities.

EXISTING FINDINGS:
{findings_text}

CONTRACT SOURCE:
```solidity
{source_code[:8000]}  # Truncate for context window
```

For each vulnerability found:
1. Classify the vulnerability type
2. Assess severity (critical/high/medium/low)
3. Explain the attack vector
4. Provide confidence score (0-1)
5. Suggest remediation

Focus on: reentrancy, access control, flash loan vectors, price manipulation, integer issues.

Respond in JSON format:
{{"vulnerabilities": [{{"type": "...", "severity": "...", "description": "...", "confidence": 0.X, "recommendation": "..."}}]}}
"""
    
    async def _query_ollama(self, prompt: str) -> str:
        """Query local Ollama instance."""
        import httpx
        
        try:
            async with httpx.AsyncClient(timeout=120.0) as client:
                response = await client.post(
                    f"{self.ollama_url}/api/generate",
                    json={
                        "model": "deepseek-coder-v2:33b",
                        "prompt": prompt,
                        "stream": False,
                    }
                )
                response.raise_for_status()
                API_CALLS.labels(service="ollama", status="success").inc()
                return response.json().get("response", "")
        except Exception as e:
            API_CALLS.labels(service="ollama", status="error").inc()
            logger.error("ollama_query_failed", error=str(e))
            return ""
    
    async def _query_claude(self, prompt: str) -> str:
        """Query Claude API as fallback."""
        if not self.anthropic_key:
            return ""
        
        try:
            import anthropic
            
            client = anthropic.Anthropic(api_key=self.anthropic_key)
            message = client.messages.create(
                model="claude-sonnet-4-20250514",
                max_tokens=4096,
                messages=[{"role": "user", "content": prompt}]
            )
            API_CALLS.labels(service="anthropic", status="success").inc()
            return message.content[0].text
        except Exception as e:
            API_CALLS.labels(service="anthropic", status="error").inc()
            logger.error("claude_query_failed", error=str(e))
            return ""
    
    def _should_fallback(self, response: str, findings: list[Vulnerability]) -> bool:
        """Determine if cloud fallback is needed."""
        has_critical = any(f.severity == Severity.CRITICAL for f in findings)
        low_confidence = any(f.confidence < 0.7 for f in findings)
        return has_critical or low_confidence
    
    def _parse_llm_response(self, response: str) -> list[Vulnerability]:
        """Parse LLM JSON response into Vulnerability objects."""
        vulnerabilities = []
        
        try:
            import re
            json_match = re.search(r'\{.*\}', response, re.DOTALL)
            if json_match:
                data = json.loads(json_match.group())
                for v in data.get("vulnerabilities", []):
                    try:
                        vuln = Vulnerability(
                            type=VulnType(v.get("type", "other")),
                            severity=Severity(v.get("severity", "info")),
                            confidence=float(v.get("confidence", 0.5)),
                            location=v.get("location", "unknown"),
                            description=v.get("description", ""),
                            recommendation=v.get("recommendation", ""),
                            analyzer="llm",
                        )
                        vulnerabilities.append(vuln)
                    except ValueError:
                        # Skip invalid vulnerability entries
                        continue
        except (json.JSONDecodeError, ValueError) as e:
            logger.error("llm_parse_error", error=str(e))
        
        return vulnerabilities


class VulnerabilityScanner:
    """Main vulnerability scanner orchestrating multiple analyzers."""
    
    def __init__(self):
        self.slither = SlitherAnalyzer(timeout=60)
        self.llm = LLMAnalyzer()
    
    async def scan(self, chain: str, address: str, source_code: str, contract_name: str = "Contract") -> ScanResult:
        """
        Run full vulnerability scan on contract.
        
        Emits metrics that trigger Alertmanager alerts:
        - VULNERABILITIES_CRITICAL: triggers immediate notification
        - VULNERABILITIES_HIGH: triggers high-priority notification
        - VULNERABILITIES_FOUND: general tracking
        """
        log = logger.bind(chain=chain, address=address)
        log.info("scan_started")
        
        ACTIVE_SCANS.labels(chain=chain).inc()
        result = ScanResult(chain=chain, address=address)
        
        try:
            import time
            start = time.monotonic()
            
            # Stage 1: Slither
            with SCAN_DURATION.labels(chain=chain, analyzer="slither").time():
                slither_vulns = await self.slither.analyze(source_code, contract_name)
                result.vulnerabilities.extend(slither_vulns)
                result.analyzers_used.append("slither")
            
            # Stage 2: LLM analysis (validate and enhance)
            if source_code:
                with SCAN_DURATION.labels(chain=chain, analyzer="llm").time():
                    use_fallback = result.has_critical or not slither_vulns
                    llm_vulns = await self.llm.analyze(
                        source_code, 
                        slither_vulns,
                        use_cloud_fallback=use_fallback
                    )
                    
                    # Deduplicate and merge
                    result.vulnerabilities = self._merge_findings(
                        result.vulnerabilities, 
                        llm_vulns
                    )
                    result.analyzers_used.append("llm")
            
            result.scan_duration_seconds = time.monotonic() - start
            
            # ============================================================
            # EMIT ALERTABLE METRICS
            # These are consumed by Alertmanager to trigger notifications
            # ============================================================
            for vuln in result.vulnerabilities:
                # General vulnerability counter
                VULNERABILITIES_FOUND.labels(
                    chain=chain, 
                    severity=vuln.severity.value,
                    vuln_type=vuln.type.value
                ).inc()
                
                # CRITICAL: Triggers immediate alert via Alertmanager
                if vuln.severity == Severity.CRITICAL:
                    VULNERABILITIES_CRITICAL.labels(
                        chain=chain,
                        vuln_type=vuln.type.value,
                        contract_address=address
                    ).inc()
                    log.warning("critical_vulnerability_found",
                              vuln_type=vuln.type.value,
                              description=vuln.description[:200])
                
                # HIGH: Triggers high-priority alert via Alertmanager
                elif vuln.severity == Severity.HIGH:
                    VULNERABILITIES_HIGH.labels(
                        chain=chain,
                        vuln_type=vuln.type.value,
                        contract_address=address
                    ).inc()
            
            # Mark scan as success
            CONTRACTS_SCANNED.labels(chain=chain, status="success").inc()
            
            log.info(
                "scan_completed",
                vulnerabilities=len(result.vulnerabilities),
                max_severity=result.max_severity.value if result.max_severity else None,
                duration=result.scan_duration_seconds
            )
            
        except Exception as e:
            result.error = str(e)
            CONTRACTS_SCANNED.labels(chain=chain, status="error").inc()
            log.error("scan_failed", error=str(e))
        finally:
            ACTIVE_SCANS.labels(chain=chain).dec()
        
        return result
    
    def _merge_findings(self, existing: list[Vulnerability], new: list[Vulnerability]) -> list[Vulnerability]:
        """Merge and deduplicate findings from multiple analyzers."""
        seen = set()
        merged = []
        
        for vuln in existing + new:
            key = (vuln.type, vuln.location)
            if key not in seen:
                seen.add(key)
                merged.append(vuln)
            else:
                # Keep highest confidence version
                for i, existing_vuln in enumerate(merged):
                    if (existing_vuln.type, existing_vuln.location) == key:
                        if vuln.confidence > existing_vuln.confidence:
                            merged[i] = vuln
                        break
        
        # Sort by severity
        severity_order = {Severity.CRITICAL: 0, Severity.HIGH: 1, Severity.MEDIUM: 2, Severity.LOW: 3, Severity.INFO: 4}
        merged.sort(key=lambda v: severity_order.get(v.severity, 5))
        
        return merged


def create_vuln_found_event(result: ScanResult) -> CloudEvent:
    """Create CloudEvent for vulnerability findings."""
    attributes = {
        "type": "io.homelab.vuln.found",
        "source": "/agent-contracts/vuln-scanner",
        "subject": f"{result.chain}/{result.address}",
    }
    
    data = {
        "chain": result.chain,
        "address": result.address,
        "vulnerabilities": [v.to_dict() for v in result.vulnerabilities],
        "max_severity": result.max_severity.value if result.max_severity else None,
        "scan_duration_seconds": result.scan_duration_seconds,
        "analyzers_used": result.analyzers_used,
    }
    
    return CloudEvent(attributes, data)
