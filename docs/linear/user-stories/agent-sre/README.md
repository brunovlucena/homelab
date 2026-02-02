# 🤖 Agent-SRE User Stories

**Comprehensive user stories for agent-sre integration with knative-lambda-operator and AI/ML tooling**

---

## 📚 Overview

This directory contains user stories for the **agent-sre** project, focusing on:
- AI/ML tool integration (Data Formulator, LLaMA Factory, TinyRecursiveModels, Agent-Lightning, DeepSeek MHC)
- Workflow automation (PrometheusRule → Linear issue creation, remediation selection, issue updates)
- Advanced capabilities (codebase understanding, PR generation, human-in-the-loop escalation)

---

## 🎯 User Story Categories

### 🤖 AI/ML Integration Stories

| Story ID | Title | Priority | Status | Story Points |
|----------|-------|----------|--------|--------------|
| **AI-001** | [Data Formulator Visualization](./BVL-61-AI-001-data-formulator-visualization.md) | 🟡 High | 📋 Backlog | 13 |
| **AI-002** | [LLaMA Factory Fine-Tuning](./BVL-62-AI-002-llama-factory-finetuning.md) | 🟡 High | 📋 Backlog | 13 |
| **AI-003** | [TinyRecursiveModels Integration](./BVL-63-AI-003-tiny-recursive-models.md) | 🟡 High | 📋 Backlog | 8 |
| **AI-004** | [Agent-Lightning RL Training](./BVL-64-AI-004-agent-lightning-rl.md) | 🟡 High | 📋 Backlog | 13 |
| **AI-005** | [DeepSeek MHC Advanced Reasoning](./BVL-68-AI-005-deepseek-mhc-reasoning.md) | 🟢 Normal | 📋 Backlog | 8 |

### 🔄 Workflow Automation Stories

| Story ID | Title | Priority | Status | Story Points |
|----------|-------|----------|--------|--------------|
| **WORKFLOW-001** | [PrometheusRule → Linear Issue with SLM](./BVL-65-WORKFLOW-001-prometheus-to-linear-with-slm.md) | 🔴 Urgent | 📋 Backlog | 13 |
| **WORKFLOW-002** | [Lambda Function Annotation Discovery](./BVL-66-WORKFLOW-002-lambda-annotation-discovery.md) | 🔴 Urgent | 📋 Backlog | 8 |
| **WORKFLOW-003** | [Enriched Issue Updates with Observability](./BVL-67-WORKFLOW-003-enriched-issue-updates.md) | 🟡 High | 📋 Backlog | 13 |
| **WORKFLOW-004** | [Codebase Understanding & Escalation](./BVL-69-WORKFLOW-004-codebase-understanding-escalation.md) | 🟡 High | 📋 Backlog | 13 |
| **WORKFLOW-005** | [PR Generation & Automated Merging](./BVL-70-WORKFLOW-005-pr-generation-merging.md) | 🟢 Normal | 📋 Backlog | 13 |

### 🔥 SRE Operational Stories

| Story ID | Title | Priority | Status | Story Points |
|----------|-------|----------|--------|--------------|
| **SRE-001** | [Build Failure Investigation](./BVL-45-SRE-001-build-failure-investigation.md) | 🔴 Urgent | ✅ Completed | 13 |
| **SRE-002** | [Performance Tuning](./BVL-46-SRE-002-performance-tuning.md) | 🟡 High | ✅ Completed | 8 |
| **SRE-003** | [Queue Management](./BVL-47-SRE-003-queue-management.md) | 🟡 High | ✅ Completed | 8 |
| **SRE-004** | [Capacity Planning](./BVL-48-SRE-004-capacity-planning.md) | 🟢 Normal | ✅ Completed | 8 |
| **SRE-005** | [Auto-Scaling Optimization](./BVL-49-SRE-005-auto-scaling-optimization.md) | 🟡 High | ✅ Completed | 8 |
| **SRE-006** | [Disaster Recovery](./BVL-50-SRE-006-disaster-recovery.md) | 🔴 Urgent | ✅ Completed | 13 |
| **SRE-007** | [Observability Enhancement](./BVL-51-SRE-007-observability-enhancement.md) | 🟡 High | ✅ Completed | 13 |
| **SRE-009** | [Backup & Restore Operations](./BVL-53-SRE-009-backup-restore-operations.md) | 🟢 Normal | ✅ Completed | 8 |
| **SRE-010** | [Dead Letter Queue Management](./BVL-54-SRE-010-dead-letter-queue-management.md) | 🟡 High | ✅ Completed | 8 |
| **SRE-011** | [Event Ordering & Idempotency](./BVL-55-SRE-011-event-ordering-and-idempotency.md) | 🟡 High | ✅ Completed | 8 |
| **SRE-012** | [Network Partition Resilience](./BVL-56-SRE-012-network-partition-resilience.md) | 🟢 Normal | ✅ Completed | 8 |
| **SRE-013** | [Schema Evolution & Compatibility](./BVL-57-SRE-013-schema-evolution-compatibility.md) | 🟢 Normal | ✅ Completed | 8 |
| **SRE-014** | [Security Incident Response](./BVL-58-SRE-014-security-incident-response.md) | 🔴 Urgent | ✅ Completed | 13 |

### 🔧 Backend Integration Stories

| Story ID | Title | Priority | Status | Story Points |
|----------|-------|----------|--------|--------------|
| **BACKEND-001** | [CloudEvents Processing](./BVL-59-BACKEND-001-cloudevents-processing.md) | 🔴 Urgent | ✅ Completed | 8 |
| **BACKEND-002** | [Build Context Management](./BVL-60-BACKEND-002-build-context-management.md) | 🟡 High | ✅ Completed | 8 |

---

## 🏗️ Architecture Integration

### High-Level Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                    AGENT-SRE WORKFLOW                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. PrometheusRule Fires                                            │
│     ↓                                                               │
│  2. prometheus-events converts to CloudEvent                       │
│     ↓                                                               │
│  3. Agent-SRE receives CloudEvent                                   │
│     ├─→ Extract alert information                                   │
│     ├─→ Query SLM data (SLOs, SLIs, error budgets)                 │
│     └─→ Create Linear issue with SLM context                        │
│     ↓                                                               │
│  4. Select Remediation (Multi-Phase)                                │
│     ├─→ Phase 0: Static annotations (fast path)                    │
│     ├─→ Phase 1: TRM recursive reasoning (7M params)               │
│     ├─→ Phase 2: RAG-based selection                               │
│     ├─→ Phase 3: Few-shot learning                                 │
│     └─→ Phase 4: AI function calling (fallback)                    │
│     ↓                                                               │
│  5. Query Observability Data                                        │
│     ├─→ Prometheus (metrics)                                       │
│     ├─→ Loki (logs)                                                │
│     └─→ Tempo (traces)                                             │
│     ↓                                                               │
│  6. Data Formulator Visualization                                   │
│     ├─→ Generate visualizations from metrics/logs/traces           │
│     ├─→ AI agent analysis and insights                             │
│     └─→ Export charts/images                                       │
│     ↓                                                               │
│  7. Update Linear Issue                                             │
│     ├─→ Add enriched comment with visualizations                   │
│     ├─→ Include AI agent insights                                  │
│     └─→ Link to SLO dashboards                                     │
│     ↓                                                               │
│  8. Execute Remediation                                             │
│     ├─→ Call LambdaFunction via HTTP                               │
│     └─→ Monitor execution                                          │
│     ↓                                                               │
│  9. Verify Remediation                                              │
│     ├─→ Query metrics again                                        │
│     ├─→ Check resource status                                      │
│     └─→ Validate alert resolution                                  │
│     ↓                                                               │
│  10. Update Linear Issue                                            │
│      ├─→ Add verification comment                                  │
│      └─→ Close issue when alert resolves                           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Integration Points

#### 1. **PrometheusRule → Agent-SRE**
- PrometheusRule annotations: `lambda_function`, `lambda_parameters`
- prometheus-events converts alerts to CloudEvents
- Agent-SRE receives CloudEvents via HTTP endpoint

#### 2. **Agent-SRE → Linear**
- Creates Linear issues with SLM context
- Updates issues with enriched comments
- Links issues to SLO dashboards
- Closes issues when alerts resolve

#### 3. **Agent-SRE → LambdaFunctions**
- Selects LambdaFunction via multi-phase selection
- Calls LambdaFunction via HTTP
- Monitors execution and verifies remediation

#### 4. **Agent-SRE → Observability Stack**
- Queries Prometheus for metrics
- Queries Loki for logs
- Queries Tempo for traces
- Uses Data Formulator for visualization

#### 5. **Agent-SRE → AI/ML Tooling**
- **TinyRecursiveModels**: Recursive reasoning for remediation selection
- **LLaMA Factory**: Local LLM fine-tuning for agent-sre
- **Agent-Lightning**: RL training for agent optimization
- **Data Formulator**: Visualization of observability data
- **DeepSeek MHC**: Advanced reasoning models

---

## 🔗 External Tool Integration

### Data Formulator
- **Purpose**: Visualize metrics, logs, traces with AI agent recommendations
- **Integration**: Query Prometheus/Loki/Tempo → Generate visualizations → Embed in Linear issues
- **Reference**: [AI-001: Data Formulator Visualization](./BVL-61-AI-001-data-formulator-visualization.md)

### LLaMA Factory
- **Purpose**: Local LLM fine-tuning at lowest cost and most private possible
- **Integration**: Collect training data → Fine-tune models → Deploy via LambdaAgent
- **Reference**: [AI-002: LLaMA Factory Fine-Tuning](./BVL-62-AI-002-llama-factory-finetuning.md)

### TinyRecursiveModels (TRM)
- **Purpose**: Recursive reasoning with tiny 7M parameter model
- **Integration**: Phase 1 in remediation selection pipeline
- **Reference**: [AI-003: TinyRecursiveModels Integration](./BVL-63-AI-003-tiny-recursive-models.md)

### Agent-Lightning
- **Purpose**: RL training for agent optimization
- **Integration**: Reward function based on remediation success → Optimize agent behavior
- **Reference**: [AI-004: Agent-Lightning RL Training](./BVL-64-AI-004-agent-lightning-rl.md)

### DeepSeek MHC
- **Purpose**: Advanced reasoning models with manifold-constrained hyper-connections
- **Integration**: Enhanced reasoning for complex remediation scenarios
- **Reference**: [AI-005: DeepSeek MHC Advanced Reasoning](./BVL-68-AI-005-deepseek-mhc-reasoning.md)

---

## 📊 Key Metrics & SLAs

### Response Time Targets
- **Alert → Linear Issue Creation**: <5 seconds
- **Alert → Remediation Selection**: <1 second (fast path), <2 seconds (TRM/RAG)
- **Alert → Remediation Execution**: <30 seconds
- **Alert → Verification**: <5 minutes

### Accuracy Targets
- **TRM Remediation Selection**: >85% accuracy
- **Parameter Extraction**: >95% accuracy
- **False Positive Rate**: <5%

### SLM Integration
- **SLO Violation Detection**: Real-time
- **Error Budget Tracking**: Continuous
- **Priority Calculation**: Based on SLM violation severity

---

## 🚀 Quick Start

### Prerequisites
- Prometheus with record rules
- PrometheusRule resources with `lambda_function` annotations
- Linear API access
- Observability stack (Prometheus, Loki, Tempo)
- AI/ML tooling (TRM, LLaMA Factory, Data Formulator, Agent-Lightning)

### Deployment
```bash
# Deploy agent-sre
kubectl apply -f flux/ai/agent-sre/k8s/kustomize/base/

# Deploy prometheus-events (if not already deployed)
kubectl apply -f flux/infrastructure/prometheus-events/k8s/

# Deploy LambdaFunctions for remediation
kubectl apply -f flux/infrastructure/knative-lambda-operator/k8s/lambdafunctions/
```

### Configuration
```yaml
# agent-sre ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: agent-sre-config
  namespace: ai
data:
  LINEAR_API_KEY: "${LINEAR_API_KEY}"
  PROMETHEUS_URL: "http://prometheus:9090"
  LOKI_URL: "http://loki:3100"
  TEMPO_URL: "http://tempo:3200"
  TRM_MODEL_PATH: "/models/trm-sre-remediation.pt"
  DATA_FORMULATOR_URL: "http://data-formulator:5000"
  SLM_ENABLED: "true"
```

---

## 📚 Related Documentation

- [Agent-SRE Architecture](../../docs/architecture/agent-sre-architecture.md)
- [TRM Integration Guide](../../docs/TRM_AGENT_SRE_INTEGRATION.md)
- [Linear Integration Guide](../../docs/integrations/linear-agent-integration.md)
- [SLM Best Practices](../../docs/slm-best-practices.md)
- [Knative Lambda Operator User Stories](../knative-lambda-operator/README.md)

---

### ✅ Validation Stories

| Story ID | Title | Priority | Status | Story Points |
|----------|-------|----------|--------|--------------|
| **BVL-255 VAL-001** | [End-to-End Workflow Validation](./BVL-255-VAL-001-end-to-end-workflow-validation.md) | 🔴 Urgent | 📋 Backlog | 13 |
| **BVL-256 VAL-002** | [Integration Testing Validation](./BVL-256-VAL-002-integration-testing-validation.md) | 🔴 Urgent | 📋 Backlog | 8 |
| **BVL-257 VAL-003** | [Remediation Selection Accuracy Validation](./BVL-257-VAL-003-remediation-selection-accuracy-validation.md) | 🟡 High | 📋 Backlog | 13 |
| **BVL-258 VAL-004** | [LambdaFunction Execution Validation](./BVL-258-VAL-004-lambdafunction-execution-validation.md) | 🔴 Urgent | 📋 Backlog | 8 |
| **BVL-259 VAL-005** | [Observability & Tracing Validation](./BVL-259-VAL-005-observability-tracing-validation.md) | 🟡 High | 📋 Backlog | 8 |
| **BVL-260 VAL-006** | [Approval System Validation](./BVL-260-VAL-006-approval-system-validation.md) | 🟡 High | 📋 Backlog | 8 |
| **BVL-261 VAL-007** | [Error Handling & Resilience Validation](./BVL-261-VAL-007-error-handling-resilience-validation.md) | 🔴 Urgent | 📋 Backlog | 13 |
| **BVL-262 VAL-008** | [Performance & Scalability Validation](./BVL-262-VAL-008-performance-scalability-validation.md) | 🟡 High | 📋 Backlog | 8 |
| **BVL-263 VAL-009** | [Security Validation](./BVL-263-VAL-009-security-validation.md) | 🔴 Urgent | 📋 Backlog | 13 |
| **BVL-264 VAL-010** | [SLM Integration Validation](./BVL-264-VAL-010-slm-integration-validation.md) | 🟡 High | 📋 Backlog | 8 |

## ✅ Completion Status

**Total Stories**: 34  
**Completed**: 14 (41%)  
**In Progress**: 0 (0%)  
**Backlog**: 20 (59%)

**Total Story Points**: 293  
**Completed Points**: 120 (41%)  
**Remaining Points**: 173 (59%)

---

## 🔄 Next Steps

1. **Complete AI/ML Integration Stories** (AI-001 through AI-005)
2. **Complete Workflow Automation Stories** (WORKFLOW-001 through WORKFLOW-005)
3. **Refactor Existing Stories** to incorporate new integrations
4. **Update Documentation** with new capabilities
5. **Integration Testing** for all new features

---

**Last Updated**: 2026-01-15  
**Owner**: SRE Team  
**Status**: Active Development
