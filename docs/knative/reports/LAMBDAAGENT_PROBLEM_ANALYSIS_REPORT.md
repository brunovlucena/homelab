# 🤖 LAMBDAAGENT PROBLEM ANALYSIS REPORT

**Assessment Type:** Principal ML Engineer Technical Review  
**Assessment Date:** December 9, 2025  
**Assessor:** Principal ML Engineer (AI Agent)  
**Target:** LambdaAgent CRD and Controller  
**Repository:** `/Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/knative-lambda-operator`  
**Status:** ANALYSIS ONLY - REMEDIATION PLAN INCLUDED

---

## 📊 EXECUTIVE SUMMARY

This report documents architectural issues, implementation gaps, and operational risks discovered during a comprehensive code review of the `LambdaAgent` CRD and its controller within the knative-lambda-operator project. The LambdaAgent is designed for AI/ML agents with pre-built images, distinct from `LambdaFunction` which builds from source.

### Risk Overview

| Severity | Count | Description |
|----------|-------|-------------|
| 🔴 **CRITICAL** | 3 | Breaking issues preventing reliable operation |
| 🟠 **HIGH** | 5 | Significant issues impacting production readiness |
| 🟡 **MEDIUM** | 7 | Issues affecting maintainability and scale |
| 🔵 **LOW** | 4 | Best practice improvements |
| **TOTAL** | **19** | |

### Problem Categories

| Category | Issues | Impact |
|----------|--------|--------|
| **ADR Implementation Gaps** | 6 | Features defined in ADR-004 not implemented |
| **Validation & Security** | 4 | Missing input validation and security controls |
| **Eventing Architecture** | 3 | Inconsistencies with LambdaFunction eventing |
| **Observability** | 2 | Missing AI-specific metrics and status |
| **Testing Coverage** | 2 | No integration tests for LambdaAgent |
| **Resource Management** | 2 | Potential resource leaks and conflicts |

---

## 🔴 CRITICAL ISSUES

### PROB-001: Missing Integration Test Coverage

| Field | Value |
|-------|-------|
| **ID** | PROB-001 |
| **Severity** | 🔴 CRITICAL |
| **Impact** | Production failures may go undetected |
| **Status** | OPEN |

#### Description
The `LambdaAgent` has **zero integration tests**. While `LambdaFunction` has extensive integration test coverage across SRE, DevOps, Security, and Backend scenarios in `src/tests/integration/`, there are no corresponding tests for `LambdaAgent`.

#### Evidence
```bash
# Search for LambdaAgent tests in integration folder
grep -r "LambdaAgent" src/tests/integration/
# Result: No matches found
```

#### Affected Files
- `src/tests/integration/` - Missing all LambdaAgent test files

#### Risk Assessment
- Cannot validate agent deployment lifecycle in real cluster
- Cannot test eventing infrastructure (Broker, Triggers, Forwards)
- Cannot verify AI configuration injection works correctly
- Cannot test scale-to-zero behavior with minReplicas=1 recommendation

#### Remediation Plan
1. Create `src/tests/integration/agents/` directory
2. Implement test files:
   - `agent_lifecycle_test.go` - Basic CRUD operations
   - `agent_eventing_test.go` - Broker/Trigger creation
   - `agent_ai_config_test.go` - AI env var injection
   - `agent_scaling_test.go` - Knative autoscaling behavior
   - `agent_cross_namespace_test.go` - Event forwarding

---

### PROB-002: ADR-004 Features Not Implemented in CRD

| Field | Value |
|-------|-------|
| **ID** | PROB-002 |
| **Severity** | 🔴 CRITICAL |
| **Impact** | Documented API contract not honored |
| **Status** | OPEN |

#### Description
The Architecture Decision Record (ADR-004) defines a comprehensive `LambdaAgent` API with features that are **not implemented** in the actual CRD or controller.

#### Missing Features from ADR-004

| ADR Feature | CRD Status | Controller Status |
|-------------|------------|-------------------|
| `behavior.intentDetection` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `behavior.conversationTTL` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `behavior.maxConcurrentConversations` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `behavior.healthCheck.path` | ❌ NOT DEFINED | ❌ HARDCODED `/health` |
| `behavior.healthCheck.interval` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `ai.fallbackModel` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `ai.topP` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `ai.contextWindowSize` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `scaling.scaleDownDelay` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `scaling.containerConcurrency` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `scaling.metrics` (Prometheus-based) | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `eventing.routing` (intent-based) | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `observability.aiMetrics` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |
| `status.aiStatus` | ❌ NOT DEFINED | ❌ NOT IMPLEMENTED |

#### Evidence

**ADR-004 defines:**
```yaml
behavior:
  conversationTTL: 1h
  maxConcurrentConversations: 100
  intentDetection:
    enabled: true
    intents:
      - name: security
        keywords: ["vulnerability", "exploit"]
        confidence: 0.7
```

**Actual CRD has:**
```go
// AgentBehaviorSpec defines agent behavior
type AgentBehaviorSpec struct {
    MaxContextMessages int32  `json:"maxContextMessages,omitempty"`
    EmitEvents         bool   `json:"emitEvents,omitempty"`
    SystemPrompt       string `json:"systemPrompt,omitempty"`
}
// Missing: intentDetection, conversationTTL, maxConcurrentConversations, healthCheck
```

#### Affected Files
- `src/operator/api/v1alpha1/lambdaagent_types.go` - CRD types incomplete
- `k8s/base/crd-lambdaagent.yaml` - CRD manifest missing fields
- `docs/07-decisions/ADR-004-lambda-agent-crd.md` - Defines features not implemented

#### Remediation Plan
**Phase 1: Essential Features**
1. Add `behavior.healthCheck.path` and `behavior.healthCheck.interval`
2. Add `scaling.containerConcurrency` and `scaling.scaleDownDelay`
3. Add `status.aiStatus` for AI-specific health reporting

**Phase 2: AI Features**
1. Implement `ai.fallbackModel` with automatic failover
2. Add `ai.topP` and `ai.contextWindowSize` configuration

**Phase 3: Advanced Features**
1. Implement intent detection and routing
2. Add conversation TTL management
3. Implement AI metrics (token usage, inference latency)

---

### PROB-003: Health Check Path Hardcoded

| Field | Value |
|-------|-------|
| **ID** | PROB-003 |
| **Severity** | 🔴 CRITICAL |
| **Impact** | Agents with different health endpoints fail health checks |
| **Status** | OPEN |

#### Description
The controller hardcodes `/health` as the readiness and liveness probe path, but agents may expose health at different endpoints (`/healthz`, `/status`, `/ready`, etc.).

#### Evidence

```go
// File: src/operator/controllers/lambdaagent_controller.go:369-388
ReadinessProbe: &corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        HTTPGet: &corev1.HTTPGetAction{
            Path: "/health",  // ❌ HARDCODED - should be configurable
            Port: intstr.FromInt32(containerPort),
        },
    },
    InitialDelaySeconds: 5,
    PeriodSeconds:       10,
},
LivenessProbe: &corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        HTTPGet: &corev1.HTTPGetAction{
            Path: "/health",  // ❌ HARDCODED - should be configurable
            Port: intstr.FromInt32(containerPort),
        },
    },
    InitialDelaySeconds: 15,
    PeriodSeconds:       20,
},
```

#### Affected Files
- `src/operator/controllers/lambdaagent_controller.go` (lines 369-388)
- `src/operator/api/v1alpha1/lambdaagent_types.go` - Missing `healthCheck` spec

#### Risk Assessment
- FastAPI agents with `/healthz` endpoint will fail health checks
- Custom agents with `/status` endpoints won't deploy correctly
- No way to disable probes for agents without health endpoints

#### Remediation Plan

**1. Add HealthCheck to CRD:**
```go
type AgentBehaviorSpec struct {
    // ... existing fields ...
    
    // HealthCheck configuration
    // +optional
    HealthCheck *AgentHealthCheckSpec `json:"healthCheck,omitempty"`
}

type AgentHealthCheckSpec struct {
    // Path for health check endpoint
    // +kubebuilder:default="/health"
    Path string `json:"path,omitempty"`
    
    // Port for health check (defaults to image.port)
    // +optional
    Port int32 `json:"port,omitempty"`
    
    // Initial delay before probes start
    // +kubebuilder:default=5
    InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
    
    // Period between probes
    // +kubebuilder:default=10
    PeriodSeconds int32 `json:"periodSeconds,omitempty"`
    
    // Disable health checks entirely
    // +kubebuilder:default=false
    Disabled bool `json:"disabled,omitempty"`
}
```

**2. Update Controller:**
```go
func (r *LambdaAgentReconciler) buildHealthProbe(agent *lambdav1alpha1.LambdaAgent, containerPort int32) (*corev1.Probe, *corev1.Probe) {
    healthPath := "/health"
    initialDelay := int32(5)
    period := int32(10)
    
    if agent.Spec.Behavior != nil && agent.Spec.Behavior.HealthCheck != nil {
        hc := agent.Spec.Behavior.HealthCheck
        if hc.Disabled {
            return nil, nil  // No probes
        }
        if hc.Path != "" {
            healthPath = hc.Path
        }
        if hc.InitialDelaySeconds > 0 {
            initialDelay = hc.InitialDelaySeconds
        }
        if hc.PeriodSeconds > 0 {
            period = hc.PeriodSeconds
        }
    }
    
    // ... build probes with configurable values
}
```

---

## 🟠 HIGH SEVERITY ISSUES

### PROB-004: Environment Variable Collision Risk

| Field | Value |
|-------|-------|
| **ID** | PROB-004 |
| **Severity** | 🟠 HIGH |
| **Impact** | User-defined env vars may be overwritten silently |
| **Status** | OPEN |

#### Description
The controller unconditionally appends `OLLAMA_URL`, `OLLAMA_MODEL`, and other environment variables to the container spec. If the user defines the same variables in `spec.env`, they may be overwritten without warning.

#### Evidence

```go
// File: src/operator/controllers/lambdaagent_controller.go:297-319
// Build environment variables
env := agent.Spec.Env  // User's env vars

// Add AI configuration as env vars if specified
if agent.Spec.AI != nil {
    if agent.Spec.AI.Endpoint != "" {
        env = append(env, corev1.EnvVar{Name: "OLLAMA_URL", Value: agent.Spec.AI.Endpoint})
        // ❌ This APPENDS, not overwrites, but order matters - user vars first!
    }
    if agent.Spec.AI.Model != "" {
        env = append(env, corev1.EnvVar{Name: "OLLAMA_MODEL", Value: agent.Spec.AI.Model})
    }
}
// ❌ If user sets OLLAMA_URL in spec.env, they now have DUPLICATE env vars
```

#### Risk Assessment
- Kubernetes uses LAST occurrence of duplicate env vars
- Operator's values override user's custom values
- No warning or error when collision detected
- Debugging is difficult (env list shows duplicates)

#### Remediation Plan

```go
// Add collision detection and precedence logic
func (r *LambdaAgentReconciler) buildEnvVars(agent *lambdav1alpha1.LambdaAgent) []corev1.EnvVar {
    // Start with operator-managed env vars (lower precedence)
    operatorEnv := make(map[string]corev1.EnvVar)
    
    if agent.Spec.AI != nil {
        if agent.Spec.AI.Endpoint != "" {
            operatorEnv["OLLAMA_URL"] = corev1.EnvVar{Name: "OLLAMA_URL", Value: agent.Spec.AI.Endpoint}
        }
        if agent.Spec.AI.Model != "" {
            operatorEnv["OLLAMA_MODEL"] = corev1.EnvVar{Name: "OLLAMA_MODEL", Value: agent.Spec.AI.Model}
        }
    }
    
    // User-defined env vars have HIGHER precedence
    for _, userEnv := range agent.Spec.Env {
        if _, exists := operatorEnv[userEnv.Name]; exists {
            // Log warning but allow user override
            log.Info("User env var overrides operator default", "name", userEnv.Name)
        }
        operatorEnv[userEnv.Name] = userEnv
    }
    
    // Convert map to slice
    result := make([]corev1.EnvVar, 0, len(operatorEnv))
    for _, env := range operatorEnv {
        result = append(result, env)
    }
    return result
}
```

---

### PROB-005: Eventing Architecture Inconsistency

| Field | Value |
|-------|-------|
| **ID** | PROB-005 |
| **Severity** | 🟠 HIGH |
| **Impact** | Resource bloat and operational confusion |
| **Status** | OPEN |

#### Description
`LambdaFunction` and `LambdaAgent` use **fundamentally different** eventing architectures:

- **LambdaFunction**: Uses a **shared broker** per namespace (`lambda-broker`)
- **LambdaAgent**: Creates a **per-agent broker** (`<agent-name>-broker`)

This inconsistency leads to:
1. Resource proliferation (1000 agents = 1000 brokers)
2. Different operational procedures for functions vs agents
3. Cross-type event routing complexity

#### Evidence

**LambdaFunction eventing (shared broker):**
```go
// File: src/operator/internal/eventing/manager.go:31-36
const (
    // SharedBrokerName is the name of the shared broker per namespace
    // All lambdas in a namespace share this broker
    SharedBrokerName = "lambda-broker"
)
```

**LambdaAgent eventing (per-agent broker):**
```go
// File: src/operator/internal/eventing/manager.go:955-957
func (m *Manager) getAgentBrokerName(agent *lambdav1alpha1.LambdaAgent) string {
    return agent.Name + "-broker"  // ❌ Creates a broker per agent!
}
```

#### Risk Assessment
- At scale (1000+ agents), creates 1000+ brokers + 1000+ RabbitmqBrokerConfigs
- Each broker consumes RabbitMQ resources (exchanges, queues)
- Inconsistent operational model between functions and agents
- Event routing between agents and functions requires explicit forwarding

#### Remediation Options

**Option A: Align with LambdaFunction (Recommended)**
```go
func (m *Manager) getAgentBrokerName(agent *lambdav1alpha1.LambdaAgent) string {
    // Use shared broker like LambdaFunction
    return SharedBrokerName
}
```

**Option B: Make Configurable**
```go
// In AgentEventingSpec
type AgentEventingSpec struct {
    // UseSharedBroker determines if agent uses namespace-shared broker
    // +kubebuilder:default=true
    UseSharedBroker bool `json:"useSharedBroker,omitempty"`
    
    // CustomBrokerName for dedicated broker (only if UseSharedBroker=false)
    // +optional
    CustomBrokerName string `json:"customBrokerName,omitempty"`
}
```

---

### PROB-006: Missing AI Status in Status Subresource

| Field | Value |
|-------|-------|
| **ID** | PROB-006 |
| **Severity** | 🟠 HIGH |
| **Impact** | No visibility into AI backend health |
| **Status** | OPEN |

#### Description
The ADR-004 defines `status.aiStatus` for AI-specific status reporting, but this is not implemented. Operators have no way to know if the AI backend (Ollama, OpenAI, etc.) is healthy.

#### ADR-004 Specification (Not Implemented)
```yaml
status:
  aiStatus:
    modelLoaded: true
    lastHealthCheck: "2024-01-15T10:35:00Z"
    inferenceLatencyP99: "1.2s"
    activeConversations: 5
```

#### Current Implementation
```go
// File: src/operator/api/v1alpha1/lambdaagent_types.go:315-334
type LambdaAgentStatus struct {
    Phase              LambdaAgentPhase       `json:"phase,omitempty"`
    Conditions         []metav1.Condition     `json:"conditions,omitempty"`
    ServiceStatus      *AgentServiceStatus    `json:"serviceStatus,omitempty"`
    EventingStatus     *AgentEventingStatus   `json:"eventingStatus,omitempty"`
    ObservedGeneration int64                  `json:"observedGeneration,omitempty"`
    // ❌ MISSING: AIStatus field
}
```

#### Remediation Plan

**1. Add AIStatus to Types:**
```go
type LambdaAgentStatus struct {
    // ... existing fields ...
    
    // AI backend status
    // +optional
    AIStatus *AgentAIStatus `json:"aiStatus,omitempty"`
}

type AgentAIStatus struct {
    // Whether the AI model is loaded and available
    ModelAvailable bool `json:"modelAvailable,omitempty"`
    
    // Model currently in use
    ActiveModel string `json:"activeModel,omitempty"`
    
    // Last health check timestamp
    LastHealthCheck *metav1.Time `json:"lastHealthCheck,omitempty"`
    
    // P99 inference latency (from metrics if available)
    InferenceLatencyP99 string `json:"inferenceLatencyP99,omitempty"`
    
    // Error message if AI backend is unhealthy
    Error string `json:"error,omitempty"`
}
```

**2. Implement Health Check in Controller:**
```go
func (r *LambdaAgentReconciler) checkAIHealth(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
    if agent.Spec.AI == nil || agent.Spec.AI.Endpoint == "" {
        return nil
    }
    
    // Call AI endpoint health check
    resp, err := http.Get(agent.Spec.AI.Endpoint + "/api/tags")
    if err != nil {
        agent.Status.AIStatus = &lambdav1alpha1.AgentAIStatus{
            ModelAvailable: false,
            Error:          err.Error(),
        }
        return err
    }
    
    // Parse response and update status
    // ...
}
```

---

### PROB-007: Forward Trigger URL Pattern Incorrect

| Field | Value |
|-------|-------|
| **ID** | PROB-007 |
| **Severity** | 🟠 HIGH |
| **Impact** | Cross-namespace event forwarding fails |
| **Status** | OPEN |

#### Description
The forward trigger constructs a hardcoded broker URL that doesn't match the actual Knative broker ingress pattern.

#### Evidence

```go
// File: src/operator/internal/eventing/manager.go:834
// Target broker URL in another namespace
targetBrokerURL := fmt.Sprintf("http://%s-broker-ingress.%s.svc.cluster.local", fwd.TargetAgent, fwd.TargetNamespace)
// ❌ This assumes broker ingress follows pattern "<agent>-broker-ingress"
// ❌ Actual Knative broker ingress is typically "broker-ingress.<namespace>"
```

#### Correct Knative Broker URL Pattern
The actual Knative broker ingress URL pattern is:
```
http://broker-ingress.knative-eventing.svc.cluster.local/<namespace>/<broker-name>
```

#### Remediation Plan

```go
// Correct broker URL construction
func (m *Manager) getTargetBrokerURL(targetNamespace, targetAgent string) string {
    // Knative broker ingress URL pattern
    brokerName := targetAgent + "-broker"
    return fmt.Sprintf("http://broker-ingress.knative-eventing.svc.cluster.local/%s/%s",
        targetNamespace, brokerName)
}
```

---

### PROB-008: No Admission Webhook Validation

| Field | Value |
|-------|-------|
| **ID** | PROB-008 |
| **Severity** | 🟠 HIGH |
| **Impact** | Invalid resources accepted, runtime failures |
| **Status** | OPEN |

#### Description
Unlike `LambdaFunction` which has comprehensive schema validation, `LambdaAgent` has no admission webhook to validate resources before creation. Invalid configurations are accepted and fail at runtime.

#### Validation Gaps

| Field | Required Validation | Status |
|-------|---------------------|--------|
| `image.repository` | Non-empty, valid registry format | ❌ MISSING |
| `ai.temperature` | Float between 0.0 and 2.0 | ❌ MISSING (string type!) |
| `ai.maxTokens` | Positive integer, reasonable limit | ❌ MISSING |
| `ai.endpoint` | Valid URL format | ❌ MISSING |
| `eventing.subscriptions[].eventType` | CloudEvent type format | ❌ MISSING |
| `resources.requests/limits` | Valid Kubernetes quantities | ❌ MISSING |

#### Evidence

```go
// File: src/operator/api/v1alpha1/lambdaagent_types.go:101-107
type AgentAISpec struct {
    // ...
    // Temperature for generation
    // +kubebuilder:default="0.7"
    Temperature string `json:"temperature,omitempty"`
    // ❌ String type - no numeric validation!
    // ❌ User can set "invalid", "-5", "99999"
}
```

#### Remediation Plan

**1. Create Validating Webhook:**
```go
// File: src/operator/webhooks/lambdaagent_webhook.go
func (v *LambdaAgentValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
    agent := obj.(*lambdav1alpha1.LambdaAgent)
    
    // Validate image
    if agent.Spec.Image.Repository == "" {
        return field.Required(field.NewPath("spec", "image", "repository"), "repository is required")
    }
    
    // Validate AI config
    if agent.Spec.AI != nil {
        if err := validateTemperature(agent.Spec.AI.Temperature); err != nil {
            return err
        }
        if agent.Spec.AI.Endpoint != "" {
            if _, err := url.Parse(agent.Spec.AI.Endpoint); err != nil {
                return field.Invalid(field.NewPath("spec", "ai", "endpoint"), agent.Spec.AI.Endpoint, "invalid URL")
            }
        }
    }
    
    return nil
}
```

---

## 🟡 MEDIUM SEVERITY ISSUES

### PROB-009: Temperature Type Should Be Numeric

| Field | Value |
|-------|-------|
| **ID** | PROB-009 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
`ai.temperature` is defined as `string` type but should be a numeric type with validation.

**Current:**
```go
Temperature string `json:"temperature,omitempty"`
```

**Should Be:**
```go
// Temperature for generation (0.0 - 2.0)
// +kubebuilder:validation:Pattern=`^(0(\.\d+)?|1(\.\d+)?|2(\.0+)?)$`
Temperature string `json:"temperature,omitempty"`
// OR better: use resource.Quantity or custom float validation
```

---

### PROB-010: Missing APIKeySecretRef Validation

| Field | Value |
|-------|-------|
| **ID** | PROB-010 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
When `ai.apiKeySecretRef` is specified, the controller doesn't validate that the referenced secret exists before deploying.

#### Remediation Plan
Add secret existence check in `reconcilePending` or `reconcileDeploying`:
```go
if agent.Spec.AI != nil && agent.Spec.AI.APIKeySecretRef != nil {
    secret := &corev1.Secret{}
    err := r.Get(ctx, types.NamespacedName{
        Name:      agent.Spec.AI.APIKeySecretRef.Name,
        Namespace: agent.Namespace,
    }, secret)
    if err != nil {
        return ctrl.Result{}, fmt.Errorf("AI API key secret not found: %w", err)
    }
}
```

---

### PROB-011: Broker Deletion on Agent Deletion

| Field | Value |
|-------|-------|
| **ID** | PROB-011 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
When a `LambdaAgent` is deleted, its broker is deleted via owner reference. However, if multiple agents share the same broker (future feature), this could delete a broker still in use.

#### Evidence
```go
// File: src/operator/internal/eventing/manager.go:726-728
// Set owner reference so broker is deleted with agent
if err := controllerutil.SetControllerReference(agent, brokerObj, m.scheme); err != nil {
    return fmt.Errorf("failed to set owner reference: %w", err)
}
```

---

### PROB-012: Scale-to-Zero Configuration Mismatch

| Field | Value |
|-------|-------|
| **ID** | PROB-012 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
ADR-004 recommends `minReplicas: 1` for agents (to keep them warm), but the CRD defaults to `minReplicas: 0`. This creates cold start issues for conversational agents.

#### ADR-004 Recommendation
```yaml
scaling:
  minReplicas: 1  # agents should stay warm
  scaleToZeroGracePeriod: 0s  # Never scale to zero for agents
```

#### Current CRD Default
```yaml
minReplicas:
  default: 0  # ❌ Allows scale to zero
```

#### Remediation
Consider changing default to `1` for agents, or document the cold-start implications clearly.

---

### PROB-013: Observability Configuration Not Applied to Container

| Field | Value |
|-------|-------|
| **ID** | PROB-013 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
While `observability.metrics` and `observability.logging` are defined in the CRD, only `observability.tracing` is applied to the container environment.

#### Evidence
```go
// File: src/operator/controllers/lambdaagent_controller.go:322-326
// Add observability configuration
if agent.Spec.Observability != nil && agent.Spec.Observability.Tracing != nil {
    if agent.Spec.Observability.Tracing.Enabled {
        env = append(env, corev1.EnvVar{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: agent.Spec.Observability.Tracing.Endpoint})
    }
}
// ❌ MISSING: Metrics configuration (exemplars, port, path)
// ❌ MISSING: Logging configuration (level, format, traceContext)
```

---

### PROB-014: No Rate Limiting for Reconciliation

| Field | Value |
|-------|-------|
| **ID** | PROB-014 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
`LambdaFunctionReconciler` has `ReconcilerOptions` with `MaxConcurrentReconciles` and `RateLimiter` for scale. `LambdaAgentReconciler` lacks these options.

#### LambdaFunction (has options):
```go
type ReconcilerOptions struct {
    MaxConcurrentReconciles int
    RateLimiter workqueue.TypedRateLimiter[ctrl.Request]
}
```

#### LambdaAgent (missing):
```go
type LambdaAgentReconciler struct {
    client.Client
    Scheme          *runtime.Scheme
    EventingManager *eventing.Manager
    // ❌ No ReconcilerOptions
}
```

---

### PROB-015: Provider-Specific Env Var Names

| Field | Value |
|-------|-------|
| **ID** | PROB-015 |
| **Severity** | 🟡 MEDIUM |
| **Status** | OPEN |

#### Description
The controller uses Ollama-specific environment variable names (`OLLAMA_URL`, `OLLAMA_MODEL`) regardless of the AI provider setting.

#### Evidence
```go
// File: src/operator/controllers/lambdaagent_controller.go:303-308
if agent.Spec.AI != nil {
    if agent.Spec.AI.Endpoint != "" {
        env = append(env, corev1.EnvVar{Name: "OLLAMA_URL", Value: agent.Spec.AI.Endpoint})
        // ❌ Always "OLLAMA_URL" even for OpenAI/Anthropic
    }
}
```

#### Remediation
Use provider-agnostic names or provider-specific names based on `ai.provider`:

```go
switch agent.Spec.AI.Provider {
case "openai":
    env = append(env, corev1.EnvVar{Name: "OPENAI_API_BASE", Value: agent.Spec.AI.Endpoint})
    env = append(env, corev1.EnvVar{Name: "OPENAI_MODEL", Value: agent.Spec.AI.Model})
case "anthropic":
    env = append(env, corev1.EnvVar{Name: "ANTHROPIC_BASE_URL", Value: agent.Spec.AI.Endpoint})
    env = append(env, corev1.EnvVar{Name: "ANTHROPIC_MODEL", Value: agent.Spec.AI.Model})
case "ollama":
default:
    env = append(env, corev1.EnvVar{Name: "OLLAMA_URL", Value: agent.Spec.AI.Endpoint})
    env = append(env, corev1.EnvVar{Name: "OLLAMA_MODEL", Value: agent.Spec.AI.Model})
}
```

---

## 🔵 LOW SEVERITY ISSUES

### PROB-016: Missing ImagePullSecrets Support in Knative Service

| Field | Value |
|-------|-------|
| **ID** | PROB-016 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Description
The CRD defines `image.imagePullSecrets` but the controller doesn't apply them to the Knative Service.

---

### PROB-017: No Metrics for LambdaAgent Operations

| Field | Value |
|-------|-------|
| **ID** | PROB-017 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Description
`LambdaFunctionReconciler` has `Metrics *metrics.ReconcilerMetrics` but `LambdaAgentReconciler` doesn't emit any Prometheus metrics.

---

### PROB-018: Missing Event Emission on Phase Changes

| Field | Value |
|-------|-------|
| **ID** | PROB-018 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Description
`LambdaFunctionReconciler` has `EventManager *events.Manager` for emitting Kubernetes events on phase changes. `LambdaAgentReconciler` doesn't emit events.

---

### PROB-019: Unit Test Coverage Incomplete

| Field | Value |
|-------|-------|
| **ID** | PROB-019 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Description
Unit tests exist for `LambdaAgentReconciler` but don't cover:
- AI secret injection
- Provider-specific env var generation
- Forward trigger creation
- DLQ configuration

---

## 📋 REMEDIATION PRIORITY

### Immediate (Sprint 1)
| ID | Issue | Effort |
|----|-------|--------|
| PROB-003 | Health check hardcoding | 2 days |
| PROB-004 | Env var collision | 1 day |
| PROB-007 | Forward trigger URL | 1 day |

### Short-term (Sprint 2)
| ID | Issue | Effort |
|----|-------|--------|
| PROB-001 | Integration tests | 5 days |
| PROB-008 | Admission webhook | 3 days |
| PROB-006 | AI status in status | 2 days |

### Medium-term (Sprint 3-4)
| ID | Issue | Effort |
|----|-------|--------|
| PROB-002 | ADR-004 features | 10 days |
| PROB-005 | Eventing architecture | 5 days |
| PROB-015 | Provider-specific env | 2 days |

### Long-term (Backlog)
| ID | Issue | Effort |
|----|-------|--------|
| PROB-002 | Intent detection | 10 days |
| PROB-002 | AI metrics | 5 days |
| PROB-002 | Conversation management | 8 days |

---

## 📊 APPENDIX: FILE REFERENCES

### Key Files for LambdaAgent
| File | Purpose |
|------|---------|
| `src/operator/api/v1alpha1/lambdaagent_types.go` | CRD Go types |
| `k8s/base/crd-lambdaagent.yaml` | CRD YAML manifest |
| `src/operator/controllers/lambdaagent_controller.go` | Reconciliation logic |
| `src/operator/controllers/lambdaagent_controller_test.go` | Unit tests |
| `src/operator/internal/eventing/manager.go` | Eventing (ReconcileAgentEventing) |
| `docs/07-decisions/ADR-004-lambda-agent-crd.md` | Architecture decision |

### Comparison Files (LambdaFunction)
| File | Purpose |
|------|---------|
| `src/operator/api/v1alpha1/lambdafunction_types.go` | Reference for patterns |
| `src/operator/controllers/lambdafunction_controller.go` | Reference for features |
| `src/tests/integration/` | Integration test patterns |

---

**Last Updated:** December 9, 2025  
**Version:** 1.0.0  
**Status:** FINDINGS DOCUMENTED - REMEDIATION PENDING
