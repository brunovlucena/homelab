# 🛡️ Agent-Contracts: Smart Contract Security Agent

## Requirements Analysis Document

**Version**: 0.1.0  
**Author**: Blockchain Specialist  
**Date**: 2025-12-03  
**Status**: Draft  

---

## 1. Executive Summary

### 1.1 Problem Statement

Recent research from MATS/Anthropic Fellows demonstrates that frontier AI models can:
- Generate exploit scripts for smart contracts at **$1.22 per contract**
- Find **zero-day vulnerabilities** autonomously
- Execute full attack chains including transaction sequencing

This creates an asymmetric threat where attackers can automate faster than manual defense.

### 1.2 Solution

Deploy defensive AI agents on homelab infrastructure using **knative-lambda** to:
- Continuously scan smart contracts for vulnerabilities
- Monitor on-chain activity for exploit patterns
- Generate alerts before attackers can exploit
- Race attackers by finding issues first

### 1.3 Deployment Target

| Component | Technology |
|-----------|------------|
| Runtime | Knative Lambda (FaaS) |
| Event Bus | RabbitMQ + CloudEvents |
| LLM Inference | Local (Ollama/vLLM) + Fallback Cloud |
| Storage | S3/MinIO |
| Monitoring | Prometheus + Grafana |
| Tracing | Tempo |

---

## 2. Functional Requirements

### 2.1 Smart Contract Analysis Engine

#### FR-001: Static Analysis Pipeline
- **Priority**: P0 (Critical)
- **Description**: Analyze Solidity/Vyper source code for vulnerability patterns
- **Inputs**: 
  - Source code (`.sol`, `.vy` files)
  - ABI JSON
  - Contract address (for verified contracts)
- **Outputs**:
  - Vulnerability report (JSON)
  - Severity classification (Critical/High/Medium/Low/Info)
  - Exploit feasibility score (0-100)
- **Supported Vulnerability Classes**:
  - Reentrancy (all variants)
  - Access control issues
  - Integer overflow/underflow
  - Flash loan attack vectors
  - Price oracle manipulation
  - MEV vulnerabilities
  - Storage collision
  - Delegatecall injection
  - Missing `view`/`pure` modifiers
  - Arbitrary external calls

#### FR-002: AI-Powered Exploit Generation (Defensive)
- **Priority**: P0 (Critical)
- **Description**: Use LLM to generate potential exploit scripts for identified vulnerabilities
- **Purpose**: Validate vulnerability severity, test defenses
- **Constraints**:
  - Exploits run ONLY against local fork (Anvil/Hardhat)
  - Never execute against mainnet
  - Audit log all generated exploits
- **LLM Integration**:
  - Primary: Local inference (Ollama with CodeLlama/DeepSeek-Coder)
  - Fallback: Claude API (for complex analysis)

#### FR-003: Multi-Chain Contract Fetching
- **Priority**: P1 (High)
- **Description**: Fetch and verify contracts from multiple chains
- **Supported Chains**:
  - Ethereum Mainnet
  - BNB Chain
  - Polygon
  - Arbitrum
  - Base
  - Optimism
- **Data Sources**:
  - Etherscan/BSCScan/PolygonScan APIs
  - Direct RPC bytecode fetch
  - IPFS (for source verification)

#### FR-004: Real-Time Contract Monitoring
- **Priority**: P1 (High)
- **Description**: Monitor newly deployed contracts for vulnerabilities
- **Trigger Sources**:
  - New contract deployment events (via RPC subscription)
  - User-submitted contracts (via API)
  - Scheduled scans of high-TVL protocols
- **SLA**: Scan within 5 minutes of deployment

### 2.2 On-Chain Activity Monitor

#### FR-005: Transaction Pattern Analysis
- **Priority**: P1 (High)
- **Description**: Detect exploit-like transaction patterns
- **Patterns to Detect**:
  - Flash loan sequences
  - Unusual token transfers
  - Contract self-destruct
  - Proxy upgrades
  - Large withdrawals post-interaction
- **Data Sources**:
  - Mempool monitoring (when available)
  - Block transaction analysis

#### FR-006: MEV Detection
- **Priority**: P2 (Medium)
- **Description**: Identify MEV extraction attempts
- **Patterns**:
  - Sandwich attacks
  - Frontrunning
  - Backrunning
  - JIT liquidity

### 2.3 Alert & Response System

#### FR-007: Multi-Channel Alerting
- **Priority**: P0 (Critical)
- **Description**: Send alerts through multiple channels
- **Channels**:
  - Grafana Alerting (via Prometheus AlertManager)
  - Telegram Bot
  - Discord Webhook
  - Email (for critical)
- **Alert Levels**:
  - 🔴 CRITICAL: Active exploit detected, immediate action required
  - 🟠 HIGH: Zero-day found, exploit feasible
  - 🟡 MEDIUM: Vulnerability found, exploitation requires conditions
  - 🟢 LOW: Code smell, best practice violation

#### FR-008: Incident Creation
- **Priority**: P1 (High)
- **Description**: Auto-create Grafana Incidents for critical findings
- **Integration**: Use Grafana MCP `create_incident` tool
- **Data Included**:
  - Contract address
  - Vulnerability type
  - Exploit PoC (if generated)
  - Recommended remediation

### 2.4 Reporting & Analytics

#### FR-009: Vulnerability Dashboard
- **Priority**: P1 (High)
- **Description**: Grafana dashboard showing:
  - Contracts scanned (total, by chain)
  - Vulnerabilities found (by severity, type)
  - Time-to-detection metrics
  - Cost per scan
  - False positive rate

#### FR-010: Historical Analysis
- **Priority**: P2 (Medium)
- **Description**: Store and query historical scan results
- **Storage**: PostgreSQL or S3 (Parquet format)
- **Retention**: 90 days detailed, 1 year aggregated

---

## 3. Non-Functional Requirements

### 3.1 Performance

| Metric | Target | Notes |
|--------|--------|-------|
| Static analysis latency | < 30s | Per contract |
| LLM exploit generation | < 2min | Using local inference |
| New contract detection | < 5min | From deployment to scan start |
| Concurrent scans | 10+ | Knative auto-scaling |
| Cold start time | < 10s | Knative optimized |

### 3.2 Scalability

- **Horizontal Scaling**: Auto-scale 0→N via Knative
- **Burst Handling**: Handle 100+ contracts/hour during high activity
- **Queue Depth**: RabbitMQ handles backpressure

### 3.3 Reliability

| Metric | Target |
|--------|--------|
| Availability | 99.5% |
| Data durability | 99.99% (S3) |
| Alert delivery | 99.9% |

### 3.4 Security

- **No mainnet execution**: Exploits ONLY on local forks
- **API key rotation**: Every 30 days
- **Audit logging**: All LLM prompts and responses logged
- **Rate limiting**: Prevent abuse of scanning API
- **RBAC**: Kubernetes RBAC for service accounts

### 3.5 Cost Optimization

| Resource | Budget Target |
|----------|---------------|
| LLM inference | $0.50/contract (local), $2/contract (cloud fallback) |
| RPC calls | Use free tier + caching |
| Storage | < 10GB/month |
| Compute | Scale to zero when idle |

---

## 4. Technical Architecture

### 4.1 Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        AGENT-CONTRACTS PLATFORM                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐         │
│  │  CONTRACT       │    │  VULNERABILITY  │    │  EXPLOIT        │         │
│  │  FETCHER        │───▶│  SCANNER        │───▶│  GENERATOR      │         │
│  │  (Knative Fn)   │    │  (Knative Fn)   │    │  (Knative Fn)   │         │
│  └────────┬────────┘    └────────┬────────┘    └────────┬────────┘         │
│           │                      │                      │                   │
│           ▼                      ▼                      ▼                   │
│  ┌────────────────────────────────────────────────────────────────┐        │
│  │                      RABBITMQ (CloudEvents)                     │        │
│  │  Event Types: contract.created | vuln.found | exploit.validated | alert.sent │        │
│  └────────────────────────────────────────────────────────────────┘        │
│           │                      │                      │                   │
│           ▼                      ▼                      ▼                   │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐         │
│  │  CHAIN          │    │  LLM            │    │  ALERT          │         │
│  │  MONITOR        │    │  SERVICE        │    │  DISPATCHER     │         │
│  │  (Knative Fn)   │    │  (Ollama/vLLM)  │    │  (Knative Fn)   │         │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         EXTERNAL INTEGRATIONS                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  • Etherscan/BSCScan APIs (contract source)                                │
│  • RPC Endpoints (Alchemy/Infura/QuickNode)                                │
│  • Grafana (dashboards, alerts, incidents)                                 │
│  • Telegram/Discord (notifications)                                        │
│  • Anvil/Hardhat (local fork for exploit validation)                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Data Flow

```
1. CONTRACT INGESTION
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ RPC Node │─────▶│ Fetcher  │─────▶│ RabbitMQ │
   │ Event    │      │ Function │      │ contract │
   └──────────┘      └──────────┘      │ .created │
                                       └────┬─────┘
                                            │
2. VULNERABILITY SCANNING                   ▼
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ contract │─────▶│ Scanner  │─────▶│ RabbitMQ │
   │ .created │      │ (Slither │      │ vuln     │
   └──────────┘      │ +Mythril)│      │ .found   │
                     └──────────┘      └────┬─────┘
                                            │
3. EXPLOIT GENERATION                       ▼
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ vuln     │─────▶│ Exploit  │─────▶│ RabbitMQ │
   │ .found   │      │ Gen(LLM) │      │ exploit  │
   └──────────┘      └──────────┘      │.validated│
                                       └────┬─────┘
                                            │
4. VALIDATION & ALERTING                    ▼
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ exploit  │─────▶│ Validator│─────▶│ Alert    │
   │.validated│      │ (Anvil)  │      │Dispatcher│
   └──────────┘      └──────────┘      └──────────┘
```

### 4.3 Knative Function Specifications

| Function | Language | Memory | Timeout | Scale |
|----------|----------|--------|---------|-------|
| contract-fetcher | Python | 256Mi | 30s | 0-5 |
| vulnerability-scanner | Python | 1Gi | 120s | 0-10 |
| exploit-generator | Python | 2Gi | 300s | 0-3 |
| chain-monitor | Python | 512Mi | ∞ (long-running) | 1-3 |
| alert-dispatcher | Python | 256Mi | 30s | 0-5 |

### 4.4 LLM Integration Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│                    LLM INFERENCE HIERARCHY                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  TIER 1: LOCAL INFERENCE (Primary - Cost: ~$0)                 │
│  ├─ Ollama + DeepSeek-Coder-V2 (33B) - Code analysis           │
│  ├─ Ollama + CodeLlama (34B) - Exploit generation              │
│  └─ Latency: 10-30s, Quality: Good for common patterns         │
│                                                                 │
│  TIER 2: CLOUD FALLBACK (Complex cases - Cost: ~$2/contract)   │
│  ├─ Claude Sonnet 4.5 - Complex vulnerability reasoning        │
│  ├─ GPT-4o - Second opinion on critical findings               │
│  └─ Trigger: Low confidence score OR critical severity         │
│                                                                 │
│  ROUTING LOGIC:                                                 │
│  if local_confidence > 0.8 and severity != CRITICAL:           │
│      return local_result                                        │
│  else:                                                          │
│      return cloud_llm_result                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. Integration Points

### 5.1 Knative Lambda Integration

The agent functions will be deployed using the existing knative-lambda infrastructure:

```yaml
# CloudEvent Types
- type: io.homelab.contract.created
  source: chain-monitor
  data:
    chain: "ethereum"
    address: "0x..."
    bytecode: "0x..."
    source_url: "https://etherscan.io/..."

- type: io.homelab.vuln.found
  source: vulnerability-scanner
  data:
    contract_address: "0x..."
    vulnerability_type: "reentrancy"
    severity: "critical"
    confidence: 0.95
    details: {...}

- type: io.homelab.exploit.validated
  source: exploit-generator
  data:
    contract_address: "0x..."
    exploit_script: "..."
    validated: true
    profit_potential: "1.5 ETH"
```

### 5.2 Grafana Integration

Using existing Grafana MCP tools:
- `query_prometheus`: Monitor agent metrics
- `query_loki_logs`: Search agent logs
- `create_incident`: Auto-create incidents for critical findings
- `search_dashboards`: Link to relevant dashboards in alerts

### 5.3 Required External APIs

| Service | Purpose | Auth Method |
|---------|---------|-------------|
| Etherscan | Contract source, ABI | API Key |
| Alchemy/Infura | RPC access | API Key |
| Ollama | Local LLM inference | None (local) |
| Anthropic | Cloud LLM fallback | API Key |

---

## 6. Deployment Strategy

### 6.1 Phase 1: MVP (Week 1-2)
- [ ] Contract fetcher function
- [ ] Basic vulnerability scanner (Slither integration)
- [ ] Alert dispatcher (Telegram only)
- [ ] Single chain support (Ethereum)

### 6.2 Phase 2: LLM Integration (Week 3-4)
- [ ] Ollama deployment on homelab
- [ ] LLM-powered exploit generator
- [ ] Exploit validation on Anvil fork
- [ ] Grafana dashboard

### 6.3 Phase 3: Multi-Chain & Production (Week 5-6)
- [ ] Multi-chain support
- [ ] Real-time monitoring
- [ ] Grafana Incident integration
- [ ] Performance optimization

---

## 7. Risk Analysis

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| False positives overwhelming alerts | High | Medium | Confidence thresholding, human review queue |
| LLM generates dangerous code | Medium | High | Sandbox execution, audit logging, no mainnet |
| High compute costs | Medium | Medium | Local inference first, aggressive caching |
| API rate limits | Medium | Low | Request batching, multiple API keys |
| Legal concerns | Low | High | Only analyze public contracts, no exploit deployment |

---

## 8. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Vulnerabilities detected before exploit | 80%+ | Compare with known exploits database |
| False positive rate | < 20% | Manual review of flagged contracts |
| Time-to-detection | < 10 min | From deployment to alert |
| Cost per contract | < $1 | Including compute + API costs |
| System availability | 99.5%+ | Prometheus uptime metrics |

---

## 9. Open Questions

1. **Local LLM Hardware**: What GPU resources are available on homelab for Ollama?
2. **Chain Priority**: Which chains should be prioritized first?
3. **Alert Recipients**: Who should receive critical alerts?
4. **Budget**: What's the monthly budget for cloud LLM fallback?
5. **Legal Review**: Any concerns about automated vulnerability disclosure?

---

## 10. Next Steps

1. ✅ Create project structure
2. ⬜ Review and approve requirements
3. ⬜ Set up Ollama with code-focused models
4. ⬜ Implement contract fetcher function
5. ⬜ Integrate Slither for static analysis
6. ⬜ Deploy MVP to homelab

---

## Appendix A: Vulnerability Detection Techniques

### A.1 Static Analysis Tools (Integrated)
- **Slither**: Fast, low false-positive rate
- **Mythril**: Symbolic execution, deeper analysis
- **Semgrep**: Custom rule patterns

### A.2 LLM Prompting Strategy

```
SYSTEM: You are a smart contract security auditor. Analyze the following 
Solidity code for vulnerabilities. For each finding:
1. Identify the vulnerability type
2. Explain the attack vector
3. Provide a proof-of-concept exploit
4. Suggest remediation

Focus on: reentrancy, access control, flash loan vectors, price manipulation,
integer issues, and arbitrary external calls.

USER: [Contract source code]

OUTPUT FORMAT: JSON with fields: vulnerabilities[], each containing:
{type, severity, location, description, exploit_poc, remediation}
```

### A.3 Exploit Validation Template

```python
# Anvil fork validation
from eth_abi import encode
from web3 import Web3

def validate_exploit(contract_address: str, exploit_calldata: str) -> bool:
    """
    Run exploit against Anvil fork, return True if successful.
    NEVER runs against mainnet.
    """
    w3 = Web3(Web3.HTTPProvider("http://anvil:8545"))  # Local fork only
    
    # Fork mainnet state
    w3.provider.make_request("anvil_reset", [{
        "forking": {"jsonRpcUrl": MAINNET_RPC, "blockNumber": "latest"}
    }])
    
    # Execute exploit
    tx = w3.eth.send_transaction({...})
    
    # Check if exploit succeeded (e.g., balance change)
    return check_exploit_success(tx)
```

---

## Appendix B: CloudEvent Schema

```json
{
  "specversion": "1.0",
  "type": "io.homelab.contract.created",
  "source": "/agent-contracts/chain-monitor",
  "id": "uuid-v4",
  "time": "2025-12-03T10:00:00Z",
  "datacontenttype": "application/json",
  "data": {
    "chain": "ethereum",
    "address": "0x1234...",
    "deployer": "0xabcd...",
    "block_number": 12345678,
    "bytecode": "0x6080...",
    "source_verified": true,
    "source_url": "https://etherscan.io/address/0x1234#code"
  }
}
```

