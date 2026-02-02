# ADR-004: LambdaAgent CRD for AI Agents

## Status
Proposed

## Context

The `LambdaFunction` CRD was designed for serverless functions with:
- Source code in MinIO (zip files)
- Build pipeline (Kaniko)
- Short-lived, stateless executions
- Generic event handling

AI Agents like `agent-bruno` and `agent-contracts` have fundamentally different requirements:
- Pre-built Docker images (no build pipeline needed)
- Long-running with conversation context
- AI/LLM-specific configuration
- Intent-based event routing
- AI-specific observability (tokens, inference latency)

Forcing agents into `LambdaFunction` creates friction:
```
❌ Unnecessary MinIO source upload
❌ Unnecessary build pipeline
❌ No first-class AI configuration
❌ Generic function semantics ≠ agent semantics
```

## Decision

Create a new `LambdaAgent` CRD specifically for AI agents that:
1. **Uses pre-built images** - No MinIO/build required
2. **First-class AI config** - Model, endpoint, temperature, system prompt
3. **Intent routing** - Route events based on detected intent
4. **Agent-specific scaling** - Optimized for conversational workloads
5. **AI observability** - Token usage, inference latency, model metrics
6. **Reuses operator infrastructure** - Broker, triggers, DLQ from LambdaFunction

## Specification

### LambdaAgent CRD

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: agent-bruno
  namespace: agent-bruno
spec:
  # ==========================================================================
  # IMAGE (Required) - Pre-built Docker image, NO MinIO/build
  # ==========================================================================
  image:
    repository: localhost:5001/agent-bruno/chatbot
    tag: "0.1.0"
    pullPolicy: IfNotPresent
    # Optional: private registry auth
    imagePullSecrets:
      - name: regcred

  # ==========================================================================
  # AI CONFIGURATION (Required) - First-class LLM settings
  # ==========================================================================
  ai:
    # LLM Provider
    provider: ollama  # ollama | openai | anthropic | local
    
    # Endpoint configuration
    endpoint: http://ollama.ollama.svc.cluster.local:11434
    # For cloud providers:
    # endpoint: https://api.openai.com/v1
    # apiKeySecret:
    #   name: openai-credentials
    #   key: api-key
    
    # Model configuration
    model: llama3.2:3b
    fallbackModel: llama3.2:1b  # Used if primary fails
    
    # Generation parameters
    temperature: 0.7
    maxTokens: 2048
    topP: 0.9
    
    # Context management
    maxContextMessages: 10
    contextWindowSize: 4096
    
    # System prompt (can also be ConfigMap ref)
    systemPrompt: |
      You are Bruno's AI assistant on his homepage.
      You help visitors learn about Bruno's projects and experience.
      Be helpful, concise, and friendly.
    # Or reference a ConfigMap:
    # systemPromptRef:
    #   name: agent-bruno-prompts
    #   key: system-prompt

  # ==========================================================================
  # AGENT BEHAVIOR - Conversation and state management
  # ==========================================================================
  behavior:
    # Conversation lifecycle
    conversationTTL: 1h
    maxConcurrentConversations: 100
    
    # Event emission
    emitEvents: true
    eventTypes:
      - io.homelab.chat.message
      - io.homelab.chat.intent
    
    # Intent detection and routing
    intentDetection:
      enabled: true
      # Built-in intents
      intents:
        - name: security
          keywords: ["vulnerability", "exploit", "security", "audit", "contract"]
          confidence: 0.7
        - name: status
          keywords: ["status", "health", "running", "deploy"]
          confidence: 0.6
    
    # Health checks
    healthCheck:
      enabled: true
      path: /health
      interval: 10s

  # ==========================================================================
  # SCALING - Optimized for conversational agents
  # ==========================================================================
  scaling:
    # Minimum replicas (agents should stay warm)
    minReplicas: 1
    maxReplicas: 5
    
    # Concurrency per instance
    targetConcurrency: 10
    containerConcurrency: 10
    
    # Scale-down delay (keep conversations alive)
    scaleDownDelay: 5m
    scaleToZeroGracePeriod: 0s  # Never scale to zero for agents
    
    # Metrics-based scaling
    metrics:
      - type: concurrency
        target: 10
      # Custom metrics (requires Prometheus)
      - type: prometheus
        name: agent_active_conversations
        target: 50

  # ==========================================================================
  # RESOURCES
  # ==========================================================================
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"

  # ==========================================================================
  # EVENTING - CloudEvents integration (reuses operator infrastructure)
  # ==========================================================================
  eventing:
    enabled: true
    
    # Event source identifier
    eventSource: /agent-bruno/chatbot
    
    # Broker configuration (operator creates lambda-broker by default)
    # brokerName: lambda-broker  # Optional override
    
    # RabbitMQ configuration
    rabbitmq:
      clusterName: rabbitmq-cluster-knative-lambda
      namespace: knative-lambda
    
    # Intent-based routing (agent-specific feature)
    routing:
      - intent: security
        eventType: io.homelab.intent.security
        target:
          service: vuln-scanner
          namespace: agent-contracts
      - intent: status
        eventType: io.homelab.intent.status
        target:
          service: homepage-api
          namespace: homepage
    
    # Subscriptions (events this agent listens to)
    subscriptions:
      - eventType: io.homelab.vuln.found
        source: /agent-contracts/*
      - eventType: io.homelab.exploit.validated
        source: /agent-contracts/*
      - eventType: io.homelab.alert.fired
        source: /alertmanager/*
    
    # Dead Letter Queue
    dlq:
      enabled: true
      retryMaxAttempts: 3
      retryBackoffDelay: PT1S

  # ==========================================================================
  # OBSERVABILITY - AI-specific metrics and tracing
  # ==========================================================================
  observability:
    # Metrics
    metrics:
      enabled: true
      port: 9090
      path: /metrics
      
      # AI-specific metrics (agent feature)
      aiMetrics:
        enabled: true
        # Emits: agent_tokens_total, agent_inference_duration_seconds,
        #        agent_model_errors_total, agent_conversations_active
    
    # Distributed tracing
    tracing:
      enabled: true
      endpoint: alloy.observability.svc:4317
      samplingRate: 1.0
      # Privacy: don't capture prompts/responses in traces
      capturePrompts: false
      captureResponses: false
    
    # Structured logging
    logging:
      level: info
      format: json
      # Include trace context in logs
      traceContext: true

  # ==========================================================================
  # ENVIRONMENT - Additional env vars
  # ==========================================================================
  env:
    - name: EMIT_EVENTS
      value: "true"
    - name: LOG_LEVEL
      value: "info"
  
  envFrom:
    - configMapRef:
        name: agent-bruno-config
    - secretRef:
        name: agent-bruno-secrets
        optional: true
```

### Status Subresource

```yaml
status:
  # Overall readiness
  phase: Ready  # Pending | Building | Deploying | Ready | Failed
  ready: true
  
  # Service URL
  url: http://agent-bruno.agent-bruno.svc.cluster.local
  
  # Image deployed
  image: localhost:5001/agent-bruno/chatbot:0.1.0
  
  # Conditions
  conditions:
    - type: Ready
      status: "True"
      reason: ServiceReady
      message: Agent is ready to receive requests
      lastTransitionTime: "2024-01-15T10:30:00Z"
    
    - type: AIConfigReady
      status: "True"
      reason: ModelAvailable
      message: LLM endpoint is healthy and model is loaded
      lastTransitionTime: "2024-01-15T10:29:50Z"
    
    - type: EventingReady
      status: "True"
      reason: BrokerAndTriggersCreated
      message: Broker, triggers, and subscriptions are ready
      lastTransitionTime: "2024-01-15T10:29:55Z"
    
    - type: ScalingReady
      status: "True"
      reason: HPAConfigured
      message: Autoscaling is configured
      lastTransitionTime: "2024-01-15T10:29:58Z"
  
  # AI-specific status
  aiStatus:
    modelLoaded: true
    lastHealthCheck: "2024-01-15T10:35:00Z"
    inferenceLatencyP99: "1.2s"
    activeConversations: 5
  
  # Observed generation
  observedGeneration: 1
```

## Comparison: LambdaFunction vs LambdaAgent

| Feature | LambdaFunction | LambdaAgent |
|---------|---------------|-------------|
| **Source** | MinIO zip → build | Pre-built image |
| **Build** | Kaniko pipeline | None needed |
| **AI Config** | ❌ Generic env | ✅ First-class |
| **Intent Routing** | ❌ Manual | ✅ Declarative |
| **Scaling** | Scale to zero | Min replicas warm |
| **Metrics** | Generic | AI-specific |
| **State** | Stateless | Conversation context |
| **Use Case** | Functions | AI Agents |

## Implementation Plan

### Phase 1: CRD Definition
1. Add `LambdaAgent` types to `api/v1alpha1/`
2. Generate CRD with controller-gen
3. Add validation webhooks

### Phase 2: Controller
1. Create `lambdaagent_controller.go`
2. Reuse eventing manager from LambdaFunction
3. Skip build pipeline - deploy image directly
4. Add AI health checks

### Phase 3: AI Features
1. Implement intent detection trigger creation
2. Add subscription management
3. Emit AI-specific metrics

### Phase 4: Observability
1. Add AI metrics to Prometheus
2. Create Grafana dashboard for agents
3. Add AI-specific alerts

## Consequences

### Positive
- Clean separation of concerns (functions vs agents)
- No unnecessary build pipeline for agents
- First-class AI configuration
- Intent-based routing simplifies agent communication
- AI-specific observability out of the box

### Negative
- Two CRDs to maintain
- Some code duplication with LambdaFunction
- Migration needed for existing agents

### Mitigations
- Share eventing, observability code between controllers
- Provide migration guide and tooling
- Keep API similar where possible

## References
- [LambdaFunction CRD](../04-architecture/SYSTEM_DESIGN.md)
- [CloudEvents Specification](../04-architecture/CLOUDEVENTS_SPECIFICATION.md)
- [Knative Serving](https://knative.dev/docs/serving/)
