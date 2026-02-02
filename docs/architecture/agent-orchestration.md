# 🎯 Agent Orchestration

> **Part of**: [AI Agent Architecture](ai-agent-architecture.md)  
> **Related**: [AI Components](ai-components.md) | [AI Connectivity](ai-connectivity.md) | [Studio Cluster](../clusters/studio-cluster.md)  
> **Last Updated**: November 7, 2025

---

## Overview

This document describes how AI agents orchestrate tasks across SLMs, LLMs, and the Knowledge Graph. It includes:

- [Agent Architecture](#agent-architecture)
- [Agent Logic](#agent-logic)
- [Workflow Examples](#workflow-examples)
- [Decision Making](#decision-making)

---

## Agent Architecture

### Deployment

**Location**: [Studio Cluster](../clusters/studio-cluster.md) (ai-agents nodes)

**Technology Stack**:
```yaml
Framework: Python + FastAPI + LangChain
Runtime: Python 3.11
Container: Docker
Orchestration: Kubernetes Deployment
Service Mesh: Linkerd (mTLS)
```

### Deployed Agents

| Agent | Port | Purpose | Status |
|-------|------|---------|--------|
| agent-bruno | 30120 | General purpose assistant | ✅ Active |
| agent-auditor | 30121 | SRE/DevOps automation | ✅ Active |
| agent-jamie | 30122 | Data science workflows | ✅ Active |
| agent-mary-kay | 30127 | Customer interaction | ✅ Active |

**Service Pattern**: `agent-*.ai-agents.svc.cluster.local`

### Agent Components

```
┌─────────────────────────────────────────────────────┐
│                    Agent Structure                  │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌───────────────────────────────────────────────┐ │
│  │         Agent Core (FastAPI)                  │ │
│  │  • HTTP API endpoints                         │ │
│  │  • Request routing                            │ │
│  │  • Authentication                             │ │
│  └───────────────────────────────────────────────┘ │
│                      │                             │
│  ┌───────────────────────────────────────────────┐ │
│  │         Intent Classifier                     │ │
│  │  • SLM-based classification                   │ │
│  │  • Query complexity analysis                  │ │
│  │  • Tool selection                             │ │
│  └───────────────────────────────────────────────┘ │
│                      │                             │
│  ┌───────────────────────────────────────────────┐ │
│  │         Knowledge Graph Client                │ │
│  │  • LanceDB connection                         │ │
│  │  • RAG pipeline                               │ │
│  │  • Context retrieval                          │ │
│  └───────────────────────────────────────────────┘ │
│                      │                             │
│  ┌───────────────────────────────────────────────┐ │
│  │         Model Router                          │ │
│  │  • SLM client (Ollama)                        │ │
│  │  • LLM client (VLLM)                          │ │
│  │  • Dynamic model selection                    │ │
│  └───────────────────────────────────────────────┘ │
│                      │                             │
│  ┌───────────────────────────────────────────────┐ │
│  │         Tool Executor                         │ │
│  │  • kubectl wrapper                            │ │
│  │  • flux reconcile                             │ │
│  │  • MCP server calls                           │ │
│  └───────────────────────────────────────────────┘ │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

## Agent Logic

### Base Agent Implementation

```python
# agent-base.py
from langchain.agents import AgentExecutor, create_react_agent
from langchain.tools import Tool

class BaseAgent:
    def __init__(self, name: str):
        self.name = name
        self.slm = OllamaClient("ollama.ml-inference.svc.forge.remote:11434")
        self.knowledge_graph = LanceDBClient("lancedb.ml-storage.svc.cluster.local:8000")
        self.llm = VLLMClient("vllm.ml-inference.svc.forge.remote:8000")
        self.mcp = MCPClient("mcp-observability.observability.svc.cluster.local:8080")
    
    async def handle_request(self, query: str) -> dict:
        """
        Main request handler following the SLM + KG + LLM pattern
        """
        # 1. Quick classification with SLM (fast, <100ms)
        intent = await self.classify_intent(query)
        
        # 2. Retrieve context from Knowledge Graph
        context = await self.retrieve_context(query, intent)
        
        # 3. Decide: SLM or LLM?
        if intent.complexity == "low":
            # Use fast SLM
            response = await self.generate_with_slm(query, context)
        else:
            # Use powerful LLM
            response = await self.generate_with_llm(query, context)
        
        # 4. Update Knowledge Graph with interaction
        await self.update_knowledge_graph(query, response)
        
        return {
            "response": response,
            "model": "slm" if intent.complexity == "low" else "llm",
            "context_used": len(context),
            "latency_ms": self.get_latency()
        }
    
    async def classify_intent(self, query: str) -> Intent:
        """Use SLM for quick intent classification"""
        prompt = f"""Classify this query:
        
Query: {query}

Categories:
- deploy: Deployment-related tasks
- query: Information retrieval
- troubleshoot: Debugging and issue resolution
- analyze: Data analysis

Complexity:
- low: Simple, single-step tasks
- medium: Multi-step tasks
- high: Complex reasoning required

Respond in JSON: {{"category": "...", "complexity": "..."}}
"""
        
        result = await self.slm.generate(
            model="llama3:8b",
            prompt=prompt,
            temperature=0.1
        )
        
        return Intent.from_json(result)
    
    async def retrieve_context(self, query: str, intent: Intent) -> list[str]:
        """Retrieve relevant context from Knowledge Graph"""
        collection = self.select_collection(intent.category)
        
        results = await self.knowledge_graph.search(
            collection=collection,
            query=query,
            limit=5
        )
        
        return [r["text"] for r in results]
    
    async def generate_with_slm(self, query: str, context: list[str]) -> str:
        """Generate response using SLM (fast, specialized)"""
        prompt = self.build_prompt(query, context)
        
        response = await self.slm.generate(
            model="codellama:13b",
            prompt=prompt,
            temperature=0.3
        )
        
        return response
    
    async def generate_with_llm(self, query: str, context: list[str]) -> str:
        """Generate response using LLM (powerful, complex)"""
        prompt = self.build_prompt(query, context)
        
        response = await self.llm.chat.completions.create(
            model="meta-llama/Meta-Llama-3.1-70B-Instruct",
            messages=[
                {"role": "system", "content": f"You are {self.name}, an expert AI assistant."},
                {"role": "user", "content": prompt}
            ],
            temperature=0.7,
            max_tokens=2000
        )
        
        return response.choices[0].message.content
    
    async def update_knowledge_graph(self, query: str, response: str):
        """Store interaction in Knowledge Graph for learning"""
        await self.knowledge_graph.insert(
            collection="agent-interactions",
            data={
                "agent": self.name,
                "query": query,
                "response": response,
                "timestamp": datetime.now().isoformat()
            }
        )
```

### Agent-Specific Implementation

```python
# agent-auditor.py (SRE Agent)
from agent_base import BaseAgent

class SREAgent(BaseAgent):
    def __init__(self):
        super().__init__(name="agent-auditor")
        
        # SRE-specific tools
        self.tools = [
            Tool(name="kubectl", func=self.kubectl_wrapper),
            Tool(name="flux_reconcile", func=self.flux_reconcile),
            Tool(name="query_prometheus", func=self.query_prometheus),
            Tool(name="query_loki", func=self.query_loki),
            Tool(name="get_traces", func=self.get_traces),
        ]
    
    async def kubectl_wrapper(self, command: str) -> str:
        """Execute kubectl commands safely"""
        # Validate command
        if not self.is_safe_command(command):
            return "Error: Unsafe command rejected"
        
        # Execute via Kubernetes API
        result = await self.k8s_client.execute(command)
        return result
    
    async def flux_reconcile(self, resource: str) -> str:
        """Trigger Flux reconciliation"""
        result = await self.mcp.call_tool(
            tool="flux_reconcile",
            params={"resource": resource}
        )
        return result
    
    async def query_prometheus(self, query: str) -> dict:
        """Query Prometheus metrics via MCP"""
        result = await self.mcp.call_tool(
            tool="query_prometheus",
            params={
                "query": query,
                "start": "now-1h",
                "end": "now"
            }
        )
        return result
    
    async def query_loki(self, logql: str) -> dict:
        """Query Loki logs via MCP"""
        result = await self.mcp.call_tool(
            tool="query_loki",
            params={
                "logql": logql,
                "start": "now-1h",
                "end": "now",
                "limit": 100
            }
        )
        return result
    
    async def get_traces(self, service: str) -> dict:
        """Get distributed traces via MCP"""
        result = await self.mcp.call_tool(
            tool="get_traces",
            params={"service": service}
        )
        return result

# Create agent executor
agent = SREAgent()
executor = AgentExecutor(
    agent=create_react_agent(llm=agent.llm, tools=agent.tools),
    tools=agent.tools,
    verbose=True
)
```

---

## Workflow Examples

### Example 1: Cross-Cluster Deployment

**Scenario**: Developer asks agent to deploy application to the best cluster

```python
# User (Dev Team) → Agent Bruno (Studio Cluster)
user_query = "Deploy my API to the cluster with best availability"

# Agent Bruno orchestration
async def handle_deployment(query: str):
    # 1. Extract intent with SLM (fast, <100ms)
    intent = await ollama_client.classify(
        prompt=query,
        categories=["deploy", "query", "troubleshoot"]
    )
    # Result: intent = "deploy"
    
    # 2. Query Knowledge Graph for context
    kg_results = await lancedb.search(
        query="cluster deployment best practices",
        collection="homelab-docs",
        top_k=3
    )
    # Result: Deployment patterns, cluster preferences
    
    # 3. Get real-time cluster metrics (cross-cluster via Linkerd)
    cluster_status = await prometheus.query_range(
        query="""
            avg by (cluster) (
                1 - (node_cpu_seconds_total{mode="idle"} / node_cpu_seconds_total)
            )
        """,
        clusters=["air", "pro", "studio"]
    )
    # Result: {"air": 0.75, "pro": 0.35, "studio": 0.88}
    
    # 4. Decision: Use SLM or LLM?
    if intent.complexity == "low":
        # Use SLM for simple decision
        decision = await ollama_client.generate(
            model="codellama",
            prompt=f"Context: {kg_results}\nMetrics: {cluster_status}\nChoose cluster.",
            temperature=0.1
        )
    else:
        # Use LLM for complex reasoning
        decision = await vllm_client.chat.completions.create(
            model="llama-3.1-70b",
            messages=[
                {"role": "system", "content": "You are an expert SRE."},
                {"role": "user", "content": f"Context: {kg_results}\nMetrics: {cluster_status}\nChoose best cluster and explain."}
            ]
        )
    # Result: "Deploy to Pro cluster (35% CPU, has free capacity)"
    
    # 5. Execute deployment via Flux GitOps
    await flux.reconcile(
        cluster="pro",
        app="user-api",
        manifest="manifests/api/deployment.yaml"
    )
    
    # 6. Update Knowledge Graph
    await lancedb.insert(
        collection="deployment-history",
        data={
            "timestamp": now(),
            "cluster": "pro",
            "app": "user-api",
            "reason": "best availability",
            "cpu_before": 0.35,
            "decision_maker": "agent-bruno"
        }
    )
    
    return {
        "status": "deployed",
        "cluster": "pro",
        "reason": decision,
        "monitoring": "https://grafana.studio/d/deployments"
    }
```

**Flow Summary**:
1. ⚡ **SLM** classifies intent (100ms)
2. 🔍 **Knowledge Graph** provides context (20ms)
3. 📊 **Prometheus** fetches metrics (50ms)
4. 🤖 **SLM/LLM** makes decision (100ms-2s)
5. 🚀 **Flux** deploys to cluster (30s)
6. 💾 **Knowledge Graph** stores interaction (10ms)

**Total Time**: ~30-32 seconds (mostly deployment)

### Example 2: Incident Analysis

**Scenario**: SRE asks agent to analyze production errors

```python
# User (SRE Team) → Agent Auditor (Studio Cluster)
user_query = "Analyze api-service errors in the last hour"

# Agent Auditor orchestration
async def analyze_incident(query: str):
    # 1. Classify with SLM
    intent = await slm.classify(query)
    # Result: category="troubleshoot", complexity="medium"
    
    # 2. Retrieve similar incidents from Knowledge Graph
    past_incidents = await kg.search(
        collection="incident-history",
        query="api-service errors timeouts",
        top_k=5
    )
    
    # 3. Query logs via MCP
    logs = await mcp.query_loki(
        logql='{app="api-service", level="error"}',
        start="now-1h",
        limit=100
    )
    
    # 4. Query metrics via MCP
    metrics = await mcp.query_prometheus(
        query='rate(http_requests_total{app="api-service", status=~"5.."}[5m])'
    )
    
    # 5. Get distributed traces via MCP
    traces = await mcp.get_traces(
        service="api-service",
        operation="POST /api/orders"
    )
    
    # 6. Analyze with LLM (complex reasoning)
    analysis = await llm.chat.completions.create(
        model="llama-3.1-70b",
        messages=[
            {"role": "system", "content": "You are an expert SRE debugging production issues."},
            {"role": "user", "content": f"""
Analyze this incident:

Past Incidents: {past_incidents}
Logs: {logs}
Metrics: {metrics}
Traces: {traces}

Provide:
1. Root cause analysis
2. Impact assessment
3. Remediation steps
4. Prevention recommendations
"""}
        ]
    )
    
    # 7. Update Knowledge Graph
    await kg.insert(
        collection="incident-history",
        data={
            "service": "api-service",
            "timestamp": now(),
            "logs_count": len(logs),
            "root_cause": analysis.root_cause,
            "resolution": analysis.remediation
        }
    )
    
    return analysis
```

**Flow Summary**:
1. ⚡ **SLM** classifies (100ms)
2. 🔍 **Knowledge Graph** retrieves past incidents (30ms)
3. 📊 **MCP** queries logs, metrics, traces (200ms)
4. 🤖 **LLM** analyzes and reasons (3-5s)
5. 💾 **Knowledge Graph** stores incident (10ms)

**Total Time**: ~3-5 seconds

### Example 3: Data Science Workflow

**Scenario**: Data scientist requests model training

```python
# User (Data Science Team) → Agent Jamie (Studio Cluster)
user_query = "Train sentiment classifier on customer feedback data"

# Agent Jamie orchestration
async def train_model(query: str):
    # 1. Classify with SLM
    intent = await slm.classify(query)
    # Result: category="train", complexity="high"
    
    # 2. Get training best practices from KG
    best_practices = await kg.search(
        collection="ml-workflows",
        query="model training sentiment classification",
        top_k=3
    )
    
    # 3. Generate training code with LLM
    training_code = await llm.generate(
        prompt=f"""
Context: {best_practices}

Generate a Flyte workflow for training a sentiment classifier:
- Data: customer_feedback table in PostgreSQL
- Model: DistilBERT fine-tuned
- Infrastructure: Forge cluster (GPU)
- Metrics: accuracy, F1, confusion matrix
"""
    )
    
    # 4. Submit to Flyte (Forge cluster)
    workflow_id = await flyte.submit_workflow(
        cluster="forge",
        code=training_code,
        resources={"gpu": 1, "memory": "16Gi"}
    )
    
    # 5. Monitor workflow
    status = await flyte.watch_workflow(workflow_id)
    
    # 6. Update Knowledge Graph
    await kg.insert(
        collection="ml-experiments",
        data={
            "workflow_id": workflow_id,
            "model": "sentiment-classifier",
            "accuracy": status.metrics.accuracy,
            "training_time": status.duration
        }
    )
    
    return {
        "status": "training_complete",
        "workflow_id": workflow_id,
        "metrics": status.metrics,
        "model_url": status.model_artifact
    }
```

---

## Decision Making

### Model Selection Algorithm

```python
def select_model(query: str, intent: Intent, context: list[str]) -> str:
    """
    Intelligent model selection based on task characteristics
    """
    # Calculate complexity score
    complexity_score = (
        len(query.split()) * 0.1 +           # Query length
        intent.reasoning_depth * 2 +          # Reasoning requirements
        len(context) * 0.5 +                  # Context size
        intent.accuracy_requirement * 3       # Accuracy needs
    )
    
    if complexity_score < 5:
        # Simple task: Use small, fast SLM
        return "ollama/llama3:8b"
    
    elif complexity_score < 10:
        # Medium task: Use larger SLM
        return "ollama/codellama:13b"
    
    else:
        # Complex task: Use powerful LLM
        return "vllm/llama-3.1-70b"
```

### Tool Selection Logic

```python
def select_tools(intent: Intent) -> list[Tool]:
    """
    Select appropriate tools based on intent
    """
    tools = []
    
    if intent.category == "deploy":
        tools.extend([kubectl, flux_reconcile, git_commit])
    
    elif intent.category == "troubleshoot":
        tools.extend([query_loki, query_prometheus, get_traces])
    
    elif intent.category == "query":
        tools.extend([knowledge_search, query_prometheus])
    
    elif intent.category == "analyze":
        tools.extend([query_prometheus, query_loki, pandas_analyze])
    
    return tools
```

---

## Related Documentation

- [🤖 AI Architecture Overview](ai-agent-architecture.md)
- [🔧 AI Components](ai-components.md)
- [🌐 AI Connectivity](ai-connectivity.md)
- [📊 MCP Observability](mcp-observability.md)
- [🎯 Studio Cluster](../clusters/studio-cluster.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

