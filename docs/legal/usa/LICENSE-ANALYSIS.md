# 📜 Homelab License Analysis Report

> **Document Version**: 1.0  
> **Last Updated**: December 11, 2025  
> **Author**: Bruno Lucena  
> **Purpose**: Comprehensive analysis of all third-party licenses in the Homelab project

---

## Executive Summary

This document provides a detailed analysis of all software licenses used in the Homelab project. The goal is to identify license obligations, compatibility issues, and commercial use implications before transforming this project into a commercial product.

### ⚠️ Critical Findings

| Risk Level | Component | License | Issue |
|------------|-----------|---------|-------|
| 🔴 **HIGH** | slither-analyzer | AGPLv3 | Copyleft - requires source disclosure for SaaS |
| 🟡 **MEDIUM** | Anthropic Claude API | Commercial | Cannot build competing AI products |
| 🟢 **LOW** | Most dependencies | MIT/Apache 2.0 | Permissive - commercial use allowed |

---

## Table of Contents

1. [License Categories](#license-categories)
2. [Go Dependencies](#go-dependencies)
3. [Python Dependencies](#python-dependencies)
4. [Node.js Dependencies](#nodejs-dependencies)
5. [Infrastructure Components](#infrastructure-components)
6. [API Services](#api-services)
7. [LLM Models](#llm-models)
8. [License Compatibility Matrix](#license-compatibility-matrix)
9. [Commercial Use Analysis](#commercial-use-analysis)
10. [Remediation Plan](#remediation-plan)

---

## License Categories

### Permissive Licenses (✅ Commercial-Friendly)

| License | Obligations | Commercial Use |
|---------|-------------|----------------|
| **MIT** | Include copyright notice | ✅ Allowed |
| **Apache 2.0** | Include license, state changes, patent grant | ✅ Allowed |
| **BSD 2/3-Clause** | Include copyright notice | ✅ Allowed |
| **ISC** | Include copyright notice | ✅ Allowed |

### Copyleft Licenses (⚠️ Requires Attention)

| License | Obligations | Commercial Use |
|---------|-------------|----------------|
| **GPLv2** | Source disclosure if distributed | ⚠️ Conditional |
| **GPLv3** | Source disclosure if distributed | ⚠️ Conditional |
| **AGPLv3** | Source disclosure even for SaaS | ⚠️ Conditional |
| **LGPL** | Source disclosure for modifications to library | ⚠️ Conditional |

### Proprietary/Commercial Licenses (💰 May Require Payment)

| License | Obligations | Commercial Use |
|---------|-------------|----------------|
| **Commercial API Terms** | Follow terms of service | 💰 Fee-based |
| **Dual License** | Choose OSS or commercial | 💰 Optional fee |

---

## Go Dependencies

### Knative Lambda Operator (`flux/infrastructure/knative-lambda-operator/src/operator/go.mod`)

| Dependency | License | Commercial Use | Notes |
|------------|---------|----------------|-------|
| `k8s.io/api` | Apache 2.0 | ✅ Allowed | Core Kubernetes API |
| `k8s.io/apimachinery` | Apache 2.0 | ✅ Allowed | Kubernetes types |
| `k8s.io/client-go` | Apache 2.0 | ✅ Allowed | Kubernetes client |
| `sigs.k8s.io/controller-runtime` | Apache 2.0 | ✅ Allowed | Operator framework |
| `knative.dev/serving` | Apache 2.0 | ✅ Allowed | Knative Serving |
| `github.com/prometheus/client_golang` | Apache 2.0 | ✅ Allowed | Prometheus metrics |
| `github.com/cloudevents/sdk-go` | Apache 2.0 | ✅ Allowed | CloudEvents SDK |
| `github.com/minio/minio-go` | Apache 2.0 | ✅ Allowed | S3 client |
| `github.com/go-git/go-git` | Apache 2.0 | ✅ Allowed | Git operations |
| `go.opentelemetry.io/*` | Apache 2.0 | ✅ Allowed | Observability |
| `google.golang.org/grpc` | Apache 2.0 | ✅ Allowed | gRPC framework |

### Cloudflare Tunnel Operator (`flux/infrastructure/cloudflare-tunnel-operator/src/go.mod`)

| Dependency | License | Commercial Use | Notes |
|------------|---------|----------------|-------|
| `k8s.io/*` | Apache 2.0 | ✅ Allowed | Kubernetes libs |
| `sigs.k8s.io/controller-runtime` | Apache 2.0 | ✅ Allowed | Operator SDK |

### Pulumi (`pulumi/go.mod`)

| Dependency | License | Commercial Use | Notes |
|------------|---------|----------------|-------|
| `github.com/pulumi/pulumi/sdk` | Apache 2.0 | ✅ Allowed | Pulumi SDK |
| `github.com/pulumi/pulumi-kubernetes/sdk` | Apache 2.0 | ✅ Allowed | K8s provider |

**Go Summary**: ✅ All Go dependencies use Apache 2.0 - fully commercial compatible.

---

## Python Dependencies

### Agent-Contracts (`flux/ai/agent-contracts/src/requirements.txt`)

| Dependency | License | Commercial Use | Risk |
|------------|---------|----------------|------|
| `slither-analyzer` | **AGPLv3** | ⚠️ CONDITIONAL | 🔴 **HIGH RISK** |
| `py-solc-x` | MIT | ✅ Allowed | |
| `web3` | MIT | ✅ Allowed | |
| `anthropic` | Commercial API | ⚠️ See API terms | |
| `fastapi` | MIT | ✅ Allowed | |
| `uvicorn` | BSD | ✅ Allowed | |
| `pydantic` | MIT | ✅ Allowed | |
| `cloudevents` | Apache 2.0 | ✅ Allowed | |
| `pika` | BSD | ✅ Allowed | |
| `boto3` | Apache 2.0 | ✅ Allowed | |
| `redis` | MIT | ✅ Allowed | |
| `opentelemetry-*` | Apache 2.0 | ✅ Allowed | |
| `prometheus-client` | Apache 2.0 | ✅ Allowed | |
| `structlog` | MIT/Apache 2.0 | ✅ Allowed | |

#### 🔴 CRITICAL: Slither-Analyzer (AGPLv3)

**Issue**: AGPLv3 requires you to provide source code to users who interact with the software over a network (SaaS loophole closure).

**Implications for Commercial Use**:
1. If you offer Agent-Contracts as a SaaS, you MUST provide your modified source code
2. Your proprietary improvements become open source
3. Cannot keep business logic confidential

**Remediation Options**:
1. **Purchase commercial license from Trail of Bits** (recommended)
2. Replace with alternative tools (Mythril - MIT, Securify2 - Apache 2.0)
3. Run Slither in isolated subprocess and communicate via API (may not satisfy AGPL)
4. Keep Agent-Contracts as open source and monetize differently

### Agent-Bruno (`flux/ai/agent-bruno/src/requirements.txt`)

| Dependency | License | Commercial Use |
|------------|---------|----------------|
| `httpx` | BSD | ✅ Allowed |
| `fastapi` | MIT | ✅ Allowed |
| `cloudevents` | Apache 2.0 | ✅ Allowed |
| `aio-pika` | Apache 2.0 | ✅ Allowed |
| `redis` | MIT | ✅ Allowed |
| `asyncpg` | Apache 2.0 | ✅ Allowed |

**Agent-Bruno Summary**: ✅ All permissive licenses - fully commercial compatible.

### Agent-RedTeam (`flux/ai/agent-redteam/src/requirements.txt`)

| Dependency | License | Commercial Use |
|------------|---------|----------------|
| `fastapi` | MIT | ✅ Allowed |
| `kubernetes` | Apache 2.0 | ✅ Allowed |
| All others | MIT/Apache 2.0/BSD | ✅ Allowed |

### Agent-DevSecOps (`flux/ai/agent-devsecops/src/requirements.txt`)

| Dependency | License | Commercial Use |
|------------|---------|----------------|
| `flask` | BSD | ✅ Allowed |
| `cloudevents` | Apache 2.0 | ✅ Allowed |
| `kubernetes` | Apache 2.0 | ✅ Allowed |

### Shared Library (`flux/ai/shared-lib/pyproject.toml`)

| Dependency | License | Commercial Use |
|------------|---------|----------------|
| `prometheus-client` | Apache 2.0 | ✅ Allowed |
| `structlog` | MIT/Apache 2.0 | ✅ Allowed |
| `pydantic` | MIT | ✅ Allowed |

---

## Node.js Dependencies

### Restaurant Command Center (`flux/ai/agent-restaurant/web/package.json`)

| Dependency | License | Commercial Use |
|------------|---------|----------------|
| `next` | MIT | ✅ Allowed |
| `react` | MIT | ✅ Allowed |
| `react-dom` | MIT | ✅ Allowed |
| `cloudevents` | Apache 2.0 | ✅ Allowed |
| `framer-motion` | MIT | ✅ Allowed |
| `lucide-react` | ISC | ✅ Allowed |
| `tailwindcss` | MIT | ✅ Allowed |
| `zustand` | MIT | ✅ Allowed |

### Agent-Chat Command Center (`flux/ai/agent-chat/web-command-center/package.json`)

| Dependency | License | Commercial Use |
|------------|---------|----------------|
| `next` | MIT | ✅ Allowed |
| `react` | MIT | ✅ Allowed |
| `@kubernetes/client-node` | Apache 2.0 | ✅ Allowed |
| `@tanstack/react-query` | MIT | ✅ Allowed |
| `socket.io-client` | MIT | ✅ Allowed |
| `recharts` | MIT | ✅ Allowed |

**Node.js Summary**: ✅ All permissive licenses - fully commercial compatible.

---

## Infrastructure Components

### Kubernetes Ecosystem

| Component | License | Commercial Use | Notes |
|-----------|---------|----------------|-------|
| Kubernetes | Apache 2.0 | ✅ Allowed | Core orchestration |
| Knative | Apache 2.0 | ✅ Allowed | Serverless runtime |
| Flux CD | Apache 2.0 | ✅ Allowed | GitOps |
| Linkerd | Apache 2.0 | ✅ Allowed | Service mesh |
| cert-manager | Apache 2.0 | ✅ Allowed | TLS management |

### Observability Stack

| Component | License | Commercial Use | Notes |
|-----------|---------|----------------|-------|
| Prometheus | Apache 2.0 | ✅ Allowed | Metrics |
| Grafana | AGPLv3 | ⚠️ Conditional | Dashboards (see note) |
| Loki | AGPLv3 | ⚠️ Conditional | Logs |
| Tempo | AGPLv3 | ⚠️ Conditional | Traces |

**Note on Grafana/Loki/Tempo**: These are AGPLv3 but you're using them as deployed services, not modifying/redistributing them. If you:
- Deploy as-is: ✅ No source disclosure needed
- Modify and offer as SaaS: ⚠️ Must release modifications
- Bundle modifications with your product: ⚠️ Must release modifications

### Storage & Messaging

| Component | License | Commercial Use |
|-----------|---------|----------------|
| MinIO | AGPLv3 | ⚠️ Conditional (similar to Grafana) |
| RabbitMQ | MPL 2.0 | ✅ Allowed |
| Redis | BSD | ✅ Allowed |

---

## API Services

### Anthropic Claude API

**License Type**: Commercial API Terms of Service

**Key Restrictions**:
1. ✅ You CAN integrate Claude outputs into your products
2. ✅ You CAN sell products that use Claude API
3. ❌ You CANNOT build competing AI products/models
4. ❌ You CANNOT reverse engineer the service
5. ❌ You CANNOT use outputs to train competing models

**Recommendation**: Claude API usage is compatible with commercial products as long as you're not building a competing AI service.

### Ollama (Local LLM Inference)

**License**: MIT (Ollama platform)

**Model Licenses** (Varies by model):
| Model | License | Commercial Use |
|-------|---------|----------------|
| Llama 3.1 | Llama 3.1 Community License | ✅ Allowed (with conditions) |
| DeepSeek-Coder | DeepSeek License | ⚠️ Check specific terms |
| Mistral | Apache 2.0 | ✅ Allowed |
| CodeLlama | Llama 2 Community License | ✅ Allowed |

**Recommendation**: Verify the license of each model you deploy for production use.

---

## LLM Models

### Model License Summary

| Model | License | Commercial Use | Revenue Threshold |
|-------|---------|----------------|-------------------|
| Llama 3.1 | Meta License | ✅ Allowed | Notify Meta if >700M MAU |
| Llama 2 | Meta License | ✅ Allowed | Notify Meta if >700M MAU |
| Mistral | Apache 2.0 | ✅ Allowed | None |
| DeepSeek-Coder | DeepSeek | ⚠️ Review terms | Varies |
| OpenAI GPT-4 | API Terms | ✅ Allowed | Per-token pricing |

---

## License Compatibility Matrix

When combining code under different licenses:

| License A | License B | Compatible? | Result |
|-----------|-----------|-------------|--------|
| MIT | Apache 2.0 | ✅ Yes | Either |
| MIT | GPLv3 | ✅ Yes | GPLv3 |
| MIT | AGPLv3 | ✅ Yes | AGPLv3 |
| Apache 2.0 | GPLv2 | ❌ No | Incompatible |
| Apache 2.0 | GPLv3 | ✅ Yes | GPLv3 |
| Apache 2.0 | AGPLv3 | ✅ Yes | AGPLv3 |
| GPLv3 | AGPLv3 | ✅ Yes | AGPLv3 |

**Your Project Impact**: 
- Most of your code can be MIT/Apache 2.0
- Agent-Contracts (using Slither) must address AGPL compliance
- Infrastructure components (Grafana/Loki) are used as services, not distributed

---

## Commercial Use Analysis

### Components Safe for Commercial Use (No Restrictions)

| Component | Why Safe |
|-----------|----------|
| Knative Lambda Operator | All Apache 2.0 dependencies |
| Agent-Bruno | All MIT/Apache 2.0 |
| Agent-RedTeam | All MIT/Apache 2.0 |
| Agent-DevSecOps | All MIT/Apache 2.0/BSD |
| All Web UIs | All MIT |
| Cloudflare Tunnel Operator | All Apache 2.0 |
| Pulumi Infrastructure | All Apache 2.0 |

### Components Requiring Attention

| Component | Issue | Action Required |
|-----------|-------|-----------------|
| Agent-Contracts | Uses AGPLv3 Slither | Purchase commercial license OR replace tool |
| Grafana/Loki/Tempo | AGPLv3 if modified | Don't modify OR release modifications |
| MinIO | AGPLv3 | Use as service OR purchase commercial |
| Claude API | No competing products | Ensure product doesn't compete with Anthropic |

---

## Remediation Plan

### Priority 1: Agent-Contracts (Slither)

**Options** (ranked by recommendation):

1. **Purchase Commercial License** (~$5,000-50,000/year estimated)
   - Contact: Trail of Bits
   - Pro: Clean commercial use, support included
   - Con: Ongoing cost

2. **Replace with MIT/Apache 2.0 Alternatives**
   - Mythril (MIT License)
   - Securify2 (Apache 2.0)
   - Custom analysis with AST parsing
   - Pro: No licensing cost
   - Con: Development effort, may miss vulnerabilities

3. **Keep Open Source**
   - Release Agent-Contracts as open source
   - Monetize via support, managed service, or premium features
   - Pro: Community contributions
   - Con: Competitors can use your code

### Priority 2: Grafana Stack

**Recommendation**: Use as deployed services without modification
- Deploy official images
- Configure via external configs
- Don't patch or modify source

### Priority 3: API Terms Compliance

**Anthropic**:
- Document that your product uses Claude API
- Ensure marketing doesn't position product as AI competitor
- Consider backup LLM provider for redundancy

---

## License Compliance Checklist

### For Release

- [ ] Add LICENSE file (MIT recommended for your code)
- [ ] Include NOTICE file with all third-party attributions
- [ ] Document Apache 2.0 modifications (if any)
- [ ] Resolve Slither/AGPL issue
- [ ] Add license headers to all source files
- [ ] Create third-party-licenses directory

### For SaaS Deployment

- [ ] Verify no modified AGPL components exposed to users
- [ ] Ensure API terms compliance (Anthropic, etc.)
- [ ] Document model licenses for deployed LLMs
- [ ] Implement license key validation for commercial features

---

## Appendix: Full Dependency Tree

See individual `go.mod`, `requirements.txt`, and `package.json` files for complete dependency lists.

---

**Document prepared for**: Bruno Lucena / Homelab Project  
**Disclaimer**: This analysis is for informational purposes only and does not constitute legal advice. Consult with a qualified intellectual property attorney before making commercial decisions.
