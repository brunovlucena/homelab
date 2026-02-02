# 🛡️ Agent-Contracts

**AI-Powered Smart Contract Security Agent for Homelab**

Automated vulnerability detection and exploit validation for DeFi smart contracts, deployed as serverless functions on Knative Lambda.

## 🎯 Overview

Following recent research showing AI agents can identify and exploit smart contract vulnerabilities at **$1.22/contract**, this project deploys defensive AI agents to:

- **Scan contracts** for vulnerabilities using Slither + LLM analysis
- **Generate exploits** (defensively) to validate severity
- **Monitor chains** for newly deployed vulnerable contracts
- **Alert** via Grafana, Telegram, and Discord

## 📋 Quick Start

```bash
# Install dependencies
make install

# Run locally
make run-scanner

# Scan a contract
make scan-contract CHAIN=ethereum ADDR=0x1234...
```

## 🏗️ Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Contract   │───▶│    Vuln      │───▶│   Exploit    │───▶│    Alert     │
│   Fetcher    │    │   Scanner    │    │  Generator   │    │  Dispatcher  │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
       │                   │                   │                   │
       └───────────────────┴───────────────────┴───────────────────┘
                                    │
                              RabbitMQ
                            (CloudEvents)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## 📁 Project Structure

```
agent-contracts/
├── docs/
│   ├── ARCHITECTURE.md    # System architecture
│   └── REQUIREMENTS.md    # Full requirements
├── src/
│   ├── contract_fetcher/  # Fetch contracts from explorers
│   ├── vuln_scanner/      # Static + LLM analysis
│   ├── exploit_generator/ # Generate exploit PoCs
│   └── alert_dispatcher/  # Multi-channel alerts
├── k8s/
│   └── kustomize/         # Kubernetes manifests
├── tests/
│   ├── unit/
│   └── integration/
├── Makefile
└── README.md
```

## 🔧 Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `OLLAMA_URL` | Local LLM endpoint | `http://ollama:11434` |
| `ANTHROPIC_API_KEY` | Claude API (fallback) | - |
| `ETHERSCAN_API_KEY` | Etherscan API | - |
| `ETHEREUM_RPC_URL` | Ethereum RPC | - |
| `REDIS_URL` | Redis cache | `redis://redis:6379` |
| `S3_BUCKET` | MinIO bucket | `agent-contracts` |

## 🚀 Deployment

### Prerequisites

- Knative Lambda infrastructure deployed
- Ollama with `deepseek-coder-v2:33b` model
- RabbitMQ cluster
- Redis (optional, for caching)

### Deploy to Homelab

```bash
# Build images
make build

# Push to registry
make push

# Deploy to Kubernetes
make deploy-pro
```

## 📊 Monitoring

Metrics exposed at `/metrics`:

- `contracts_fetched_total{chain, status}`
- `vulnerabilities_found_total{chain, severity, type}`
- `scan_duration_seconds{chain, analyzer}`
- `exploits_validated_total{chain, success}`

## ⚠️ Safety

- **Exploits run ONLY on local Anvil forks** - never mainnet
- All LLM prompts and responses are audit logged
- Rate limiting prevents abuse

## 📚 Documentation

- [Requirements](REQUIREMENTS.md) - Full requirements analysis
- [Architecture](docs/ARCHITECTURE.md) - System design

## 📄 License

MIT

