# 🎯 Homelab Agent Shared Library

Shared libraries for homelab agents: optimization, domain memory, observability, and more.

## Modules

### 🔭 Observability (`observability/`)

Type-safe OpenTelemetry initialization with automatic Tempo tracing support.

```python
from observability import initialize_observability

# Initialize with environment variables
initialize_observability(service_name="agent-bruno")
```

See [observability/README.md](observability/README.md) for details.

### 🧠 Agent Memory (`agent_memory/`)

Domain Memory Factory pattern for stateful AI agents.

### 🔗 Agent Communication (`agent_communication/`)

Inter-agent communication with CloudEvents and observability.

### 🎯 Agent Optimization (`agent_optimization/`)

Mathematical optimization for agent communication via CloudEvents.

### 🍎 macOS Automation (`macos_automation/`)

Client library for controlling macOS applications (Safari, etc.) via AppleScript from Kubernetes agents.

```python
from macos_automation import MacOSAutomationClient

client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")
result = await client.navigate("https://lucena.cloud")
```

See [scripts/mac/apple-events/README.md](../../scripts/mac/apple-events/README.md) for setup and usage.

## Overview

Esta biblioteca implementa teorias matemáticas para otimizar a comunicação entre agents no homelab:

| Fase | Teoria | Implementação |
|------|--------|---------------|
| 1 | Queueing Theory | `queueing.py` - M/M/c model, priority queues |
| 2 | Game Theory | `gametheory.py` - Contract Net Protocol, Shapley Value, Nash Equilibrium |
| 3 | Control Theory | `control.py` - PID Controller, Auto-Scaler |

## Key Feature: Reutiliza Métricas Existentes

**NÃO cria métricas duplicadas!** Consulta métricas já coletadas:

### Knative Lambda Operator Metrics
```promql
# Arrival Rate (λ) - eventos/segundo
rate(knative_lambda_function_invocations_total[5m])

# Service Rate (μ) - eventos processados/segundo
1 / avg(knative_lambda_function_duration_seconds)

# Queue Depth
knative_lambda_operator_workqueue_depth

# Queue Latency P95
histogram_quantile(0.95, knative_lambda_operator_workqueue_latency_seconds_bucket)
```

### RabbitMQ Metrics
```promql
# Arrival Rate (λ) - mensagens publicadas/segundo
rate(rabbitmq_queue_messages_published_total[5m])

# Service Rate (μ) - mensagens entregues/segundo
rate(rabbitmq_queue_messages_delivered_total[5m])

# Queue Depth
rabbitmq_queue_messages
```

## Installation

```bash
# No agent's requirements.txt:
-e ../shared-lib

# Or pip install:
pip install -e /path/to/shared-lib
```

## Quick Start

### 1. Decision Engine (Integração Completa)

```python
from agent_optimization import (
    EventDecisionEngine,
    AgentState,
    PriorityQueue,
    setup_optimization_metrics,
)

# Setup
agent_state = AgentState(
    agent_id="agent-bruno",
    cpu_capacity=100.0,
    memory_capacity=100.0,
    cpu_used=30.0,
    memory_used=40.0,
    avg_processing_time=0.1,
    success_rate=0.95,
    specializations=["chat", "nlp"]
)

queue = PriorityQueue(max_size=10000)
metrics = setup_optimization_metrics(
    agent_id="agent-bruno",
    namespace_filter="agent-bruno",
)

# Create decision engine
engine = EventDecisionEngine(
    agent_id="agent-bruno",
    agent_state=agent_state,
    queue=queue,
    metrics=metrics,
)

# Make decision on incoming event
async def handle_event(event):
    decision = await engine.decide(
        event_id=event["id"],
        event_type=event["type"],
        event_data=event["data"],
        reward=10.0,
    )
    
    if decision.decision == "process":
        await process_event(event)
    elif decision.decision == "forward":
        await forward_to_agent(event, decision.forward_target)
    else:  # reject
        logger.info(f"Rejected: {decision.reason}")
```

### 2. Queueing Theory Only

```python
from agent_optimization import QueueingOptimizer, QueueingMetrics

optimizer = QueueingOptimizer(
    target_latency=1.0,      # 1 second target
    max_utilization=0.8,     # 80% max
)

# Calculate metrics
metrics = optimizer.calculate_metrics(
    arrival_rate=100.0,      # 100 events/sec (λ)
    service_rate=20.0,       # 20 events/sec per agent (μ)
    num_agents=5,            # 5 agents
)

print(f"Utilization: {metrics.utilization:.2%}")
print(f"Avg Wait Time: {metrics.avg_wait_time:.2f}s")
print(f"Recommended Agents: {metrics.recommended_agents}")

# Get scaling recommendation
action, target, reason = optimizer.recommend_scaling(metrics)
print(f"Action: {action}, Target: {target}, Reason: {reason}")
```

### 3. Game Theory (Contract Net Protocol)

```python
from agent_optimization import ContractNetProtocol, AgentState, Task

# Create CNP
cnp = ContractNetProtocol(bid_timeout=5.0)

# Register agents
cnp.register_agent(AgentState(
    agent_id="agent-bruno",
    cpu_capacity=100.0,
    cpu_used=30.0,
    efficiency=0.9,
    specializations=["chat"]
))

cnp.register_agent(AgentState(
    agent_id="agent-redteam",
    cpu_capacity=80.0,
    cpu_used=20.0,
    efficiency=0.85,
    specializations=["security"]
))

# Announce task
task = Task(
    task_id="process-event-123",
    event_type="io.homelab.chat.message",
    event_data={"message": "hello"},
    priority=8,
    reward=10.0,
)

# Collect bids
bids = await cnp.announce_task(task)

# Select winner
winner, winning_bid = cnp.select_winner(task.task_id)
print(f"Winner: {winner.agent_id}, Utility: {winning_bid.utility:.2f}")
```

### 4. Control Theory (Auto-Scaler)

```python
from agent_optimization import AutoScaler

scaler = AutoScaler(
    target_latency=1.0,
    target_utilization=0.7,
    min_replicas=1,
    max_replicas=20,
    scale_up_cooldown=60.0,
    scale_down_cooldown=300.0,
)

# Get recommendation
decision = scaler.recommend(
    current_latency=1.5,      # 1.5s (above target)
    current_utilization=0.85, # 85% (above target)
    current_replicas=3,
)

print(f"Action: {decision.action}")
print(f"Target Replicas: {decision.target_replicas}")
print(f"Reason: {decision.reason}")
```

## Integration with Existing Agents

### agent-bruno

```python
# src/shared/events.py
from agent_optimization import EventDecisionEngine, setup_optimization_metrics

class EventSubscriber:
    def __init__(self):
        self.metrics = setup_optimization_metrics(
            agent_id="agent-bruno",
            namespace_filter="agent-bruno",
        )
        # ... rest of init
    
    async def handle(self, event: CloudEvent) -> bool:
        # Use decision engine
        decision = await self.decision_engine.decide(
            event_id=event["id"],
            event_type=event["type"],
            event_data=event.data,
        )
        
        if decision.decision != "process":
            return False
        
        # Continue with normal processing
        # ...
```

## Metrics Exposed

The library exposes NEW metrics for optimization decisions:

```promql
# Decision counts
agent_optimization_decisions_total{agent_id="agent-bruno", decision="process"}
agent_optimization_decisions_total{agent_id="agent-bruno", decision="forward"}
agent_optimization_decisions_total{agent_id="agent-bruno", decision="reject"}

# Decision latency
agent_optimization_decision_latency_seconds{agent_id="agent-bruno"}

# Game Theory metrics
agent_optimization_bids_submitted_total{agent_id="agent-bruno", task_type="chat"}
agent_optimization_bids_won_total{agent_id="agent-bruno", task_type="chat"}
agent_optimization_bid_utility{agent_id="agent-bruno", task_type="chat"}
agent_optimization_shapley_value{agent_id="agent-bruno", coalition_id="task-123"}

# Control Theory metrics
agent_optimization_scaling_decisions_total{agent_id="agent-bruno", direction="up"}
agent_optimization_pid_error{agent_id="agent-bruno", metric_type="latency"}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     AGENT OPTIMIZATION LIBRARY                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────┐    ┌──────────────────┐    ┌───────────────────┐  │
│  │ EventDecision   │    │ QueueingOptimizer│    │ ContractNet       │  │
│  │ Engine          │───▶│ (M/M/c model)    │    │ Protocol          │  │
│  │                 │    └──────────────────┘    │ (Game Theory)     │  │
│  │ Decides:        │    ┌──────────────────┐    └───────────────────┘  │
│  │ - process       │───▶│ NashEquilibrium  │    ┌───────────────────┐  │
│  │ - forward       │    │ (Best Response)  │    │ ShapleyCalculator │  │
│  │ - reject        │    └──────────────────┘    │ (Fair Rewards)    │  │
│  └─────────────────┘    ┌──────────────────┐    └───────────────────┘  │
│          │              │ PIDController    │                           │
│          │              │ AutoScaler       │                           │
│          ▼              │ (Control Theory) │                           │
│  ┌─────────────────┐    └──────────────────┘                           │
│  │ OptimizationMetrics ─────────────────────────────────────────────┐  │
│  │ (Prometheus Client)                                              │  │
│  │                                                                  │  │
│  │  QUERIES EXISTING METRICS:                                       │  │
│  │  ├── knative_lambda_function_invocations_total → λ               │  │
│  │  ├── knative_lambda_function_duration_seconds  → μ               │  │
│  │  ├── knative_lambda_operator_workqueue_depth   → Queue Depth     │  │
│  │  ├── rabbitmq_queue_messages_published_total   → RabbitMQ λ      │  │
│  │  └── rabbitmq_queue_messages_delivered_total   → RabbitMQ μ      │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │         PROMETHEUS            │
                    │  (kube-prometheus-stack)      │
                    │                               │
                    │  Already collecting:          │
                    │  - knative-lambda-operator    │
                    │  - RabbitMQ                   │
                    │  - Knative Serving            │
                    └───────────────────────────────┘
```

## References

1. **Queueing Theory**: "Fundamentals of Queueing Theory" - Gross & Harris
2. **Game Theory**: "Algorithmic Game Theory" - Nisan et al.
3. **Control Theory**: "Feedback Control of Dynamic Systems" - Franklin et al.
4. **Multi-Agent Systems**: "An Introduction to MultiAgent Systems" - Wooldridge
