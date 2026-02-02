# ADR-004 Compliance Gap Analysis

**Generated:** 2025-12-09  
**ADR Reference:** `docs/07-decisions/ADR-004-lambda-agent-crd.md`  
**Current Implementation:** `src/operator/api/v1alpha1/lambdaagent_types.go`  
**Status:** Analysis Complete - 17 Gaps Identified

---

## 📊 Executive Summary

This document provides a detailed comparison between the ADR-004 specification and the current `LambdaAgent` implementation, identifying all features defined in the ADR that are missing or incomplete in the actual CRD and controller.

| Category | ADR Features | Implemented | Gap Count |
|----------|--------------|-------------|-----------|
| **AI Configuration** | 9 | 5 | **4** |
| **Behavior** | 7 | 3 | **4** |
| **Scaling** | 6 | 4 | **2** |
| **Eventing** | 3 | 2 | **1** |
| **Observability** | 8 | 4 | **4** |
| **Status** | 5 | 3 | **2** |
| **TOTAL** | **38** | **21** | **17** |

---

## 🔴 CRITICAL GAPS (Must Have for Production)

### GAP-001: Health Check Configuration Not Configurable

| Field | Detail |
|-------|--------|
| **ADR Spec** | `behavior.healthCheck.path`, `behavior.healthCheck.interval`, `behavior.healthCheck.enabled` |
| **Current** | Hardcoded to `/health` in controller (line 371-388) |
| **Impact** | Agents with custom health endpoints fail probes |
| **Priority** | P0 - Immediate |

**ADR-004 Specification:**
```yaml
behavior:
  healthCheck:
    enabled: true
    path: /health
    interval: 10s
```

**Current Implementation:**
```go
// lambdaagent_controller.go:369-388
ReadinessProbe: &corev1.Probe{
    HTTPGet: &corev1.HTTPGetAction{
        Path: "/health",  // ❌ HARDCODED
    },
},
```

**Remediation:**
```go
// Add to AgentBehaviorSpec
type AgentHealthCheckSpec struct {
    Enabled             bool   `json:"enabled,omitempty"`
    Path                string `json:"path,omitempty"`      // default: /health
    Interval            string `json:"interval,omitempty"`  // default: 10s
    InitialDelaySeconds int32  `json:"initialDelaySeconds,omitempty"`
}
```

---

### GAP-002: AI Status Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `status.aiStatus` with `modelLoaded`, `lastHealthCheck`, `inferenceLatencyP99`, `activeConversations` |
| **Current** | Not present in `LambdaAgentStatus` |
| **Impact** | No visibility into AI backend health |
| **Priority** | P0 - Immediate |

**ADR-004 Specification:**
```yaml
status:
  aiStatus:
    modelLoaded: true
    lastHealthCheck: "2024-01-15T10:35:00Z"
    inferenceLatencyP99: "1.2s"
    activeConversations: 5
```

**Current Implementation:**
```go
// lambdaagent_types.go:315-334
type LambdaAgentStatus struct {
    Phase              LambdaAgentPhase
    Conditions         []metav1.Condition
    ServiceStatus      *AgentServiceStatus
    EventingStatus     *AgentEventingStatus
    // ❌ MISSING: AIStatus
}
```

**Remediation:**
```go
// Add to LambdaAgentStatus
AIStatus *AgentAIStatus `json:"aiStatus,omitempty"`

type AgentAIStatus struct {
    ModelLoaded         bool         `json:"modelLoaded,omitempty"`
    ActiveModel         string       `json:"activeModel,omitempty"`
    LastHealthCheck     *metav1.Time `json:"lastHealthCheck,omitempty"`
    InferenceLatencyP99 string       `json:"inferenceLatencyP99,omitempty"`
    ActiveConversations int32        `json:"activeConversations,omitempty"`
    Error               string       `json:"error,omitempty"`
}
```

---

## 🟠 HIGH PRIORITY GAPS

### GAP-003: Fallback Model Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `ai.fallbackModel` for automatic failover |
| **Current** | Not in CRD |
| **Impact** | No resilience when primary model fails |
| **Priority** | P1 |

**ADR-004 Specification:**
```yaml
ai:
  model: llama3.2:3b
  fallbackModel: llama3.2:1b  # Used if primary fails
```

---

### GAP-004: Intent Detection Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `behavior.intentDetection` with `enabled`, `intents[]` (name, keywords, confidence) |
| **Current** | Not in CRD |
| **Impact** | Cannot auto-route events based on intent |
| **Priority** | P1 |

**ADR-004 Specification:**
```yaml
behavior:
  intentDetection:
    enabled: true
    intents:
      - name: security
        keywords: ["vulnerability", "exploit", "security"]
        confidence: 0.7
      - name: status
        keywords: ["status", "health", "running"]
        confidence: 0.6
```

---

### GAP-005: Intent-Based Event Routing Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `eventing.routing[]` with `intent`, `eventType`, `target.service`, `target.namespace` |
| **Current** | Not in CRD |
| **Impact** | Cannot route events based on detected intent |
| **Priority** | P1 |

**ADR-004 Specification:**
```yaml
eventing:
  routing:
    - intent: security
      eventType: io.homelab.intent.security
      target:
        service: vuln-scanner
        namespace: agent-contracts
```

---

### GAP-006: Conversation TTL Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `behavior.conversationTTL` |
| **Current** | Not in CRD |
| **Impact** | Cannot configure conversation lifecycle |
| **Priority** | P1 |

---

### GAP-007: Max Concurrent Conversations Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `behavior.maxConcurrentConversations` |
| **Current** | Not in CRD |
| **Impact** | Cannot limit concurrent conversations |
| **Priority** | P1 |

---

## 🟡 MEDIUM PRIORITY GAPS

### GAP-008: AI Advanced Parameters Missing

| Field | Detail |
|-------|--------|
| **ADR Spec** | `ai.topP`, `ai.contextWindowSize` |
| **Current** | Not in CRD |
| **Impact** | Less control over LLM generation |
| **Priority** | P2 |

**ADR-004 Specification:**
```yaml
ai:
  topP: 0.9
  contextWindowSize: 4096
```

---

### GAP-009: System Prompt ConfigMap Reference Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `ai.systemPromptRef` with ConfigMap name/key |
| **Current** | Only `behavior.systemPrompt` string exists |
| **Impact** | Large prompts require inline YAML |
| **Priority** | P2 |

**ADR-004 Specification:**
```yaml
ai:
  systemPromptRef:
    name: agent-bruno-prompts
    key: system-prompt
```

---

### GAP-010: Scaling Container Concurrency Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `scaling.containerConcurrency` |
| **Current** | Not in CRD (only `targetConcurrency`) |
| **Impact** | Cannot configure Knative container concurrency |
| **Priority** | P2 |

---

### GAP-011: Scaling Scale-Down Delay Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `scaling.scaleDownDelay` |
| **Current** | Not in CRD |
| **Impact** | Cannot keep agents warm after traffic drops |
| **Priority** | P2 |

**ADR-004 Specification:**
```yaml
scaling:
  scaleDownDelay: 5m
```

---

### GAP-012: Custom Prometheus Metrics Scaling Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `scaling.metrics[]` with Prometheus custom metrics |
| **Current** | Not in CRD |
| **Impact** | Cannot scale on custom metrics like active conversations |
| **Priority** | P2 |

**ADR-004 Specification:**
```yaml
scaling:
  metrics:
    - type: prometheus
      name: agent_active_conversations
      target: 50
```

---

### GAP-013: AI-Specific Metrics Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `observability.metrics.aiMetrics.enabled` |
| **Current** | Not in CRD or controller |
| **Impact** | No token usage, inference latency metrics |
| **Priority** | P2 |

**ADR-004 Specification:**
```yaml
observability:
  metrics:
    aiMetrics:
      enabled: true
      # Emits: agent_tokens_total, agent_inference_duration_seconds,
      #        agent_model_errors_total, agent_conversations_active
```

---

### GAP-014: Tracing Privacy Controls Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `observability.tracing.samplingRate`, `capturePrompts`, `captureResponses` |
| **Current** | Only `enabled` and `endpoint` exist |
| **Impact** | Cannot control trace sampling or prompt capture |
| **Priority** | P2 |

**ADR-004 Specification:**
```yaml
observability:
  tracing:
    samplingRate: 1.0
    capturePrompts: false
    captureResponses: false
```

---

## 🔵 LOW PRIORITY GAPS

### GAP-015: Behavior Event Types Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `behavior.eventTypes[]` - List of event types agent emits |
| **Current** | Not in CRD |
| **Impact** | Documentation/discovery only |
| **Priority** | P3 |

---

### GAP-016: Condition Type AIConfigReady Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `status.conditions[].type: AIConfigReady` |
| **Current** | Only `Ready` and `Eventing` conditions |
| **Impact** | Less granular status reporting |
| **Priority** | P3 |

---

### GAP-017: Condition Type ScalingReady Not Implemented

| Field | Detail |
|-------|--------|
| **ADR Spec** | `status.conditions[].type: ScalingReady` |
| **Current** | Not implemented |
| **Impact** | Less granular status reporting |
| **Priority** | P3 |

---

## 📋 Implementation Roadmap

### Phase 1: Production Blockers (Sprint 1)
| Gap ID | Feature | Effort |
|--------|---------|--------|
| GAP-001 | Health check configuration | 2 days |
| GAP-002 | AI status in status subresource | 2 days |
| N/A | Admission webhook validation | 3 days |

### Phase 2: AI Resilience (Sprint 2)
| Gap ID | Feature | Effort |
|--------|---------|--------|
| GAP-003 | Fallback model | 2 days |
| GAP-008 | AI advanced parameters (topP, contextWindowSize) | 1 day |
| GAP-009 | System prompt ConfigMap reference | 1 day |

### Phase 3: Conversation Management (Sprint 3)
| Gap ID | Feature | Effort |
|--------|---------|--------|
| GAP-006 | Conversation TTL | 2 days |
| GAP-007 | Max concurrent conversations | 1 day |
| GAP-010 | Container concurrency | 1 day |
| GAP-011 | Scale-down delay | 1 day |

### Phase 4: Intent Routing (Sprint 4-5)
| Gap ID | Feature | Effort |
|--------|---------|--------|
| GAP-004 | Intent detection | 5 days |
| GAP-005 | Intent-based event routing | 5 days |

### Phase 5: Advanced Observability (Sprint 6)
| Gap ID | Feature | Effort |
|--------|---------|--------|
| GAP-012 | Custom Prometheus metrics scaling | 3 days |
| GAP-013 | AI-specific metrics | 3 days |
| GAP-014 | Tracing privacy controls | 1 day |
| GAP-015/16/17 | Additional conditions | 1 day |

---

## 📊 Summary Table

| Gap ID | Feature | CRD Change | Controller Change | Priority |
|--------|---------|------------|-------------------|----------|
| GAP-001 | Health check config | ✅ Yes | ✅ Yes | P0 |
| GAP-002 | AI status | ✅ Yes | ✅ Yes | P0 |
| GAP-003 | Fallback model | ✅ Yes | ✅ Yes | P1 |
| GAP-004 | Intent detection | ✅ Yes | ✅ Yes | P1 |
| GAP-005 | Intent routing | ✅ Yes | ✅ Yes | P1 |
| GAP-006 | Conversation TTL | ✅ Yes | ❌ No (app logic) | P1 |
| GAP-007 | Max conversations | ✅ Yes | ❌ No (app logic) | P1 |
| GAP-008 | topP, contextWindowSize | ✅ Yes | ✅ Yes (env vars) | P2 |
| GAP-009 | systemPromptRef | ✅ Yes | ✅ Yes | P2 |
| GAP-010 | containerConcurrency | ✅ Yes | ✅ Yes | P2 |
| GAP-011 | scaleDownDelay | ✅ Yes | ✅ Yes | P2 |
| GAP-012 | Custom metrics scaling | ✅ Yes | ✅ Yes | P2 |
| GAP-013 | AI metrics | ✅ Yes | ✅ Yes | P2 |
| GAP-014 | Tracing privacy | ✅ Yes | ✅ Yes (env vars) | P2 |
| GAP-015 | eventTypes | ✅ Yes | ❌ No | P3 |
| GAP-016 | AIConfigReady condition | ❌ No | ✅ Yes | P3 |
| GAP-017 | ScalingReady condition | ❌ No | ✅ Yes | P3 |

---

**Document Version:** 1.0  
**Last Updated:** 2025-12-09  
**Author:** Principal ML Engineer / Senior Cloud Architect
