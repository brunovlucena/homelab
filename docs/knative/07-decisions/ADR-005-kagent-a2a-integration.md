# ADR-005: kagent A2A Protocol Integration with LambdaAgent

## Status
Proposed

## Context

The kagent project provides an **Agent-to-Agent (A2A) protocol** that enables:
- Agent discovery via `.well-known/agent.json` endpoint
- Skills-based agent interaction model
- HTTP-based synchronous agent invocation
- MCP (Model Context Protocol) tool integration
- Declarative agent capabilities definition

Our `knative-lambda-operator` `LambdaAgent` currently uses:
- **CloudEvents** (async, event-driven) via RabbitMQ Broker/Triggers
- **Intent-based routing** for event filtering
- **Knative Serving** for auto-scaling
- **Pre-built Docker images** (no build pipeline)

### Problem Statement

From a software engineering perspective, we need to evaluate:
1. **Should LambdaAgent support A2A protocol?**
2. **What's the best integration approach?**
3. **How do A2A and CloudEvents complement each other?**

## Analysis: kagent A2A vs LambdaAgent

### Architecture Comparison

| Aspect | kagent A2A | LambdaAgent (Current) |
|--------|-------------|----------------------|
| **Communication** | HTTP REST (synchronous) | CloudEvents (asynchronous) |
| **Discovery** | `.well-known/agent.json` | Kubernetes Service DNS |
| **Invocation** | Direct HTTP POST | Broker → Trigger → Service |
| **Protocol** | A2A JSON schema | CloudEvents v1.0 |
| **Tools** | MCP servers | Custom tool implementations |
| **Scaling** | kagent controller manages | Knative Serving auto-scales |
| **State** | Managed by kagent | Stateless (conversation in agent) |
| **Deployment** | kagent Agent CRD | LambdaAgent CRD |

### Strengths & Weaknesses

#### kagent A2A Strengths
✅ **Standardized protocol** - Well-defined agent discovery and invocation  
✅ **Skills-based model** - Clear capability declaration  
✅ **MCP integration** - Rich tool ecosystem  
✅ **Synchronous communication** - Direct request/response  
✅ **Agent discovery** - Self-describing agents via `.well-known/agent.json`  

#### kagent A2A Weaknesses
❌ **Requires kagent controller** - Additional infrastructure dependency  
❌ **Synchronous only** - No async event-driven patterns  
❌ **Single protocol** - Less flexible than CloudEvents  
❌ **Tight coupling** - Agents must know A2A endpoints  

#### LambdaAgent Strengths
✅ **CloudEvents standard** - Industry-standard event format  
✅ **Event-driven** - Natural async patterns, decoupling  
✅ **Knative native** - Leverages existing Knative infrastructure  
✅ **Intent routing** - Declarative event filtering  
✅ **No external dependencies** - Self-contained operator  
✅ **Multi-protocol** - Can support HTTP, gRPC, WebSocket via CloudEvents  

#### LambdaAgent Weaknesses
❌ **No agent discovery protocol** - Must know service names  
❌ **Async only** - No direct synchronous invocation  
❌ **No skills model** - Capabilities not self-describing  
❌ **Custom tool integration** - No standard MCP support  

## Decision: Hybrid Approach

### Option 1: Full A2A Integration (❌ NOT RECOMMENDED)

**Approach**: Replace CloudEvents with A2A protocol

**Pros**:
- Standardized agent protocol
- Skills-based model
- MCP tool support

**Cons**:
- ❌ **Breaks existing architecture** - All agents use CloudEvents
- ❌ **Loses event-driven benefits** - Async patterns, decoupling
- ❌ **Adds kagent dependency** - Requires kagent controller
- ❌ **Migration complexity** - Rewrite all agent communication
- ❌ **Less flexible** - CloudEvents supports more patterns

**Verdict**: **REJECT** - Too disruptive, loses architectural benefits

---

### Option 2: A2A as Optional Layer (✅ RECOMMENDED)

**Approach**: Add A2A protocol support as **optional feature** alongside CloudEvents

**Architecture**:
```
┌─────────────────────────────────────────────────────────────┐
│                    LambdaAgent CRD                          │
│  ┌──────────────────────┐  ┌──────────────────────────┐  │
│  │  CloudEvents (Primary)│  │  A2A Protocol (Optional) │  │
│  │  - RabbitMQ Broker    │  │  - .well-known/agent.json │  │
│  │  - Triggers           │  │  - Skills endpoint       │  │
│  │  - Intent routing     │  │  - Direct HTTP invoke    │  │
│  └──────────────────────┘  └──────────────────────────┘  │
│                          ↓                                  │
│              ┌─────────────────────────┐                  │
│              │   Knative Service        │                  │
│              │   (Agent Container)     │                  │
│              └─────────────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

**Implementation**:
1. **Add A2A spec to LambdaAgent CRD**:
   ```yaml
   spec:
     a2a:
       enabled: true  # Optional, defaults to false
       skills:
         - id: answer-questions
           name: Answer Questions
           description: Answer questions about Kubernetes
           inputModes: ["text"]
           outputModes: ["text"]
           tags: ["kubernetes"]
   ```

2. **Operator creates A2A endpoints**:
   - `.well-known/agent.json` - Agent discovery
   - `/api/a2a/skills/{skill-id}` - Skill invocation
   - Both endpoints proxy to underlying Knative Service

3. **Dual protocol support**:
   - **CloudEvents** (async) - Primary, existing architecture
   - **A2A** (sync) - Optional, for direct agent-to-agent calls

**Pros**:
- ✅ **Best of both worlds** - Async (CloudEvents) + Sync (A2A)
- ✅ **Backward compatible** - Existing agents unchanged
- ✅ **Optional feature** - Opt-in, no breaking changes
- ✅ **No kagent dependency** - Implement A2A in operator
- ✅ **Skills model** - Self-describing capabilities
- ✅ **MCP integration path** - Can add MCP server support later

**Cons**:
- ⚠️ **Additional complexity** - Two protocols to maintain
- ⚠️ **More endpoints** - Operator must expose A2A routes

**Verdict**: **ACCEPT** - Provides flexibility without breaking changes

---

### Option 3: A2A Gateway Service (⚠️ ALTERNATIVE)

**Approach**: Create separate A2A gateway service that translates A2A → CloudEvents

**Architecture**:
```
A2A Client → A2A Gateway → CloudEvents → Broker → LambdaAgent
```

**Pros**:
- ✅ **Separation of concerns** - Gateway handles protocol translation
- ✅ **No operator changes** - LambdaAgent unchanged

**Cons**:
- ❌ **Additional service** - More infrastructure to manage
- ❌ **Translation overhead** - Protocol conversion layer
- ❌ **Less efficient** - Extra hop in request path

**Verdict**: **CONSIDER** - Good for external A2A clients, but less integrated

---

## Recommended Implementation: Option 2

### Phase 1: A2A Spec Extension

Add to `LambdaAgentSpec`:
```go
// A2AConfig defines Agent-to-Agent protocol configuration
type A2AConfig struct {
    // Enable A2A protocol support
    Enabled bool `json:"enabled,omitempty"`
    
    // Skills this agent exposes via A2A
    Skills []A2ASkill `json:"skills,omitempty"`
    
    // A2A endpoint configuration
    Endpoint *A2AEndpointConfig `json:"endpoint,omitempty"`
}

type A2ASkill struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    InputModes  []string `json:"inputModes"`
    OutputModes []string `json:"outputModes"`
    Tags        []string `json:"tags,omitempty"`
}

type A2AEndpointConfig struct {
    // Base path for A2A endpoints (default: /api/a2a)
    BasePath string `json:"basePath,omitempty"`
    
    // Enable agent discovery endpoint
    DiscoveryEnabled bool `json:"discoveryEnabled,omitempty"`
}
```

### Phase 2: Operator A2A Endpoint Handler

Create A2A handler in operator that:
1. **Serves `.well-known/agent.json`**:
   ```json
   {
     "name": "agent-bruno",
     "description": "AI Chatbot for Homepage",
     "url": "http://agent-bruno.agent-bruno.svc.cluster.local/api/a2a",
     "version": "1",
     "capabilities": {
       "streaming": false,
       "pushNotifications": false
     },
     "skills": [
       {
         "id": "answer-questions",
         "name": "Answer Questions",
         "description": "Answer questions about Kubernetes",
         "inputModes": ["text"],
         "outputModes": ["text"],
         "tags": ["kubernetes"]
       }
     ]
   }
   ```

2. **Proxies skill invocations** to Knative Service:
   ```
   POST /api/a2a/skills/answer-questions
   → Forward to agent service
   → Return response
   ```

### Phase 3: MCP Tool Integration (Future)

Add MCP server support to LambdaAgent:
```yaml
spec:
  tools:
    - type: mcp
      mcpServer:
        name: k8s-tools
        endpoint: http://mcp-k8s.mcp.svc:8080
        toolNames:
          - k8s_get_resources
          - k8s_get_pod_logs
```

## Benefits of Hybrid Approach

### 1. **Protocol Flexibility**
- **CloudEvents** for async, event-driven workflows
- **A2A** for synchronous, direct agent-to-agent calls
- Choose based on use case

### 2. **Backward Compatibility**
- Existing agents continue working
- A2A is opt-in feature
- No migration required

### 3. **Skills Model**
- Self-describing agent capabilities
- Better agent discovery
- Clearer agent contracts

### 4. **Future MCP Support**
- Path to MCP tool integration
- Rich tool ecosystem
- Standardized tool protocol

### 5. **Best Practices**
- **Separation of concerns** - Protocols serve different purposes
- **Progressive enhancement** - Add features without breaking
- **Standards compliance** - Support industry standards

## Comparison Matrix

| Feature | CloudEvents Only | A2A Only | Hybrid (Recommended) |
|---------|-----------------|----------|---------------------|
| **Async patterns** | ✅ | ❌ | ✅ |
| **Sync patterns** | ❌ | ✅ | ✅ |
| **Agent discovery** | ❌ | ✅ | ✅ |
| **Skills model** | ❌ | ✅ | ✅ |
| **Backward compat** | ✅ | ❌ | ✅ |
| **MCP support** | ❌ | ✅ | ✅ (future) |
| **Complexity** | Low | Medium | Medium |
| **Flexibility** | High | Low | **Highest** |

## Implementation Considerations

### 1. **Operator Changes**
- Add A2A endpoint handler to operator
- Serve `.well-known/agent.json` from operator
- Proxy skill invocations to agent services

### 2. **Agent Container Changes**
- **None required** - Agents continue using CloudEvents
- Optional: Add A2A skill handlers if needed

### 3. **Service Discovery**
- A2A endpoints exposed via operator service
- Or: Expose via Knative Service directly (simpler)

### 4. **Authentication**
- Reuse existing RBAC
- Add A2A-specific auth if needed

## Conclusion

**Recommendation**: Implement **Option 2 (Hybrid Approach)**

**Rationale**:
1. ✅ **Preserves existing architecture** - CloudEvents remains primary
2. ✅ **Adds value** - Skills model, agent discovery, sync communication
3. ✅ **No breaking changes** - Backward compatible
4. ✅ **Future-proof** - Path to MCP integration
5. ✅ **Best practices** - Multiple protocols for different use cases

**Next Steps**:
1. Design A2A spec extension for LambdaAgent CRD
2. Implement A2A endpoint handler in operator
3. Add agent discovery endpoint
4. Test with sample agent
5. Document A2A usage patterns

## References

- [kagent A2A Documentation](https://kagent.dev/docs/kagent/examples/slack-a2a)
- [CloudEvents Specification](https://cloudevents.io/)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)
- [ADR-004: LambdaAgent CRD](./ADR-004-lambda-agent-crd.md)


