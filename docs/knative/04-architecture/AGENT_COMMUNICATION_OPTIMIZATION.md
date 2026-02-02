# 🎯 Otimização de Comunicação entre Agents - Teorias Matemáticas

**Versão**: 1.0.0  
**Data**: Janeiro 2025  
**Status**: IMPLEMENTADO ✅

---

## 📦 Implementação

A biblioteca de otimização está disponível em:

```
flux/ai/shared-lib/agent_optimization/
├── __init__.py           # Exports principais
├── queueing.py           # FASE 1: Queueing Theory (M/M/c)
├── gametheory.py         # FASE 2: Game Theory (CNP, Shapley, Nash)
├── control.py            # FASE 3: Control Theory (PID, AutoScaler)
├── decision.py           # Decision Engine (integra tudo)
├── metrics.py            # Métricas (reutiliza Prometheus existente)
└── README.md             # Documentação completa
```

### Instalação

```bash
# No requirements.txt do agent:
-e ../../shared-lib
```

### Uso Rápido

```python
from agent_optimization import EventDecisionEngine, AgentState, setup_optimization_metrics

# O agent decide automaticamente se deve processar, encaminhar ou rejeitar
decision = await engine.decide(event_id, event_type, event_data)
```

---

## 📖 Visão Geral

Este documento explora teorias matemáticas aplicáveis para otimizar a comunicação entre agents no homelab via CloudEvents e knative-lambda-operator. O sistema atual utiliza:

- **CloudEvents v1.0** como formato padrão
- **RabbitMQ** como message broker
- **Knative Eventing** para roteamento de eventos
- **Múltiplos agents** (agent-bruno, agent-redteam, agent-pos-edge, etc.)

## 🔧 Métricas Reutilizadas

**NÃO criamos métricas duplicadas!** A biblioteca consulta métricas EXISTENTES:

| Métrica | Fonte | Uso |
|---------|-------|-----|
| `knative_lambda_function_invocations_total` | knative-lambda-operator | Taxa de chegada (λ) |
| `knative_lambda_function_duration_seconds` | knative-lambda-operator | Taxa de serviço (μ) |
| `knative_lambda_operator_workqueue_depth` | knative-lambda-operator | Queue Depth |
| `rabbitmq_queue_messages_published_total` | RabbitMQ | λ (mensagens) |
| `rabbitmq_queue_messages_delivered_total` | RabbitMQ | μ (mensagens) |

---

## 🎮 1. Game Theory (Teoria dos Jogos)

### Aplicação

Game Theory pode otimizar a **alocação de recursos** e **coordenação estratégica** entre agents.

### Casos de Uso

#### 1.1. Alocação de Tarefas (Task Allocation Game)

**Problema**: Múltiplos agents competem para processar eventos do broker.

**Solução**: Modelar como um jogo onde:
- **Players**: Agents (agent-bruno, agent-redteam, etc.)
- **Estratégias**: Escolher quais tipos de eventos processar
- **Payoff**: Eficiência de processamento vs. custo de recursos

**Implementação**:
```python
# Exemplo: Agent decide se deve processar um evento baseado em:
# - Sua capacidade atual (CPU/memória)
# - Prioridade do evento
# - Custo de processamento
# - Recompensa esperada (útil para o sistema)

def should_process_event(agent_state, event):
    # Nash Equilibrium: agent escolhe estratégia ótima
    # considerando ações dos outros agents
    utility = calculate_utility(agent_state, event)
    threshold = calculate_nash_threshold(other_agents_strategies)
    return utility > threshold
```

#### 1.2. Coordenação Cooperativa (Cooperative Game)

**Problema**: Agents precisam coordenar ações sem comunicação centralizada.

**Solução**: **Shapley Value** para distribuir recompensas justamente entre agents cooperativos.

**Exemplo**: Quando múltiplos agents colaboram para resolver um problema:
- Agent A detecta vulnerabilidade
- Agent B valida exploit
- Agent C aplica patch

**Shapley Value** calcula contribuição justa de cada agent.

#### 1.3. Mechanism Design

**Problema**: Incentivar agents a reportar verdadeiramente sua capacidade/estado.

**Solução**: **Vickrey-Clarke-Groves (VCG) mechanism** para garantir que agents não mintam sobre recursos disponíveis.

---

## 📊 2. Queueing Theory (Teoria de Filas)

### Aplicação

Otimizar o **desempenho do RabbitMQ broker** e **latência de processamento**.

### Modelos Relevantes

#### 2.1. M/M/c Queue (Multiple Servers)

**Modelo**: RabbitMQ broker com múltiplos consumers (agents).

**Parâmetros**:
- **λ (lambda)**: Taxa de chegada de eventos (events/second)
- **μ (mu)**: Taxa de processamento por agent (events/second)
- **c**: Número de agents (servers)

**Métricas Otimizadas**:
```python
# Cálculo de métricas de fila
import numpy as np

def optimize_queue_parameters(arrival_rate, processing_rate, num_agents):
    """
    Otimiza número de agents baseado em teoria de filas.
    
    Objetivo: Minimizar tempo médio de espera (W) e 
              probabilidade de fila vazia (P0)
    """
    rho = arrival_rate / (num_agents * processing_rate)  # Utilização
    
    # Fórmula de Erlang C
    if rho >= 1:
        return "Sistema instável - aumentar agents"
    
    # Tempo médio de espera na fila
    W_q = calculate_waiting_time(arrival_rate, processing_rate, num_agents)
    
    # Número ótimo de agents para minimizar W_q
    optimal_agents = find_optimal_agents(arrival_rate, processing_rate)
    
    return {
        "optimal_agents": optimal_agents,
        "utilization": rho,
        "avg_waiting_time": W_q,
        "throughput": arrival_rate
    }
```

#### 2.2. Priority Queues

**Aplicação**: Priorizar eventos críticos (ex: `io.homelab.alert.critical`).

**Modelo**: M/M/1 com prioridades (preemptive ou non-preemptive).

**Implementação**:
```yaml
# Configuração de prioridades no RabbitMQ
event_priorities:
  critical: 10    # io.homelab.alert.critical
  high: 7         # io.homelab.vuln.found
  medium: 5       # io.homelab.chat.message
  low: 1          # io.homelab.analytics.*
```

#### 2.3. Queue Network Analysis

**Problema**: Eventos passam por múltiplas filas (broker → trigger → agent).

**Solução**: **Jackson Network** para modelar sistema completo e identificar gargalos.

---

## 🤝 3. Multi-Agent Systems (MAS) Optimization

### 3.1. Consensus-Based Optimization

**Aplicação**: Agents convergem para decisões coletivas sem coordenador central.

**Exemplo**: Decidir qual agent deve processar um evento específico.

```python
class ConsensusAgent:
    def __init__(self, agent_id, neighbors):
        self.agent_id = agent_id
        self.neighbors = neighbors  # Outros agents conectados
        self.state = {"capacity": 100, "load": 0}
    
    async def reach_consensus(self, event):
        """
        Algoritmo de consenso para decidir processamento.
        Baseado em: Average Consensus Algorithm
        """
        # Broadcast estado atual
        my_state = self.get_state()
        neighbor_states = await self.get_neighbor_states()
        
        # Atualizar baseado em média ponderada
        consensus_state = weighted_average([my_state] + neighbor_states)
        
        # Decidir ação baseado em consenso
        if self.should_process(consensus_state, event):
            return await self.process_event(event)
```

### 3.2. Contract Net Protocol (CNP)

**Aplicação**: Agents negociam tarefas via "contratos".

**Fluxo**:
1. **Manager** (ex: knative-lambda-operator) anuncia tarefa via CloudEvent
2. **Contractors** (agents) fazem "bids" com suas capacidades
3. **Manager** seleciona melhor bid
4. **Contractor** executa e reporta resultado

**Implementação**:
```python
# Event: io.knative.lambda.command.task.announce
{
    "task_id": "build-function-123",
    "requirements": {"cpu": "500m", "memory": "1Gi"},
    "deadline": "2025-01-15T10:00:00Z"
}

# Agent responde com bid
# Event: io.homelab.agent.bid.submitted
{
    "task_id": "build-function-123",
    "agent_id": "agent-bruno",
    "bid": {
        "cost": 0.5,  # Utilidade/custo
        "estimated_time": "5m",
        "confidence": 0.9
    }
}
```

---

## 📡 4. Information Theory

### 4.1. Entropy e Compressão de Eventos

**Aplicação**: Reduzir overhead de comunicação.

**Métricas**:
- **Entropy H(X)**: Quantidade de informação em eventos
- **Mutual Information I(X;Y)**: Informação compartilhada entre agents

**Otimização**:
```python
def optimize_event_payload(event_data):
    """
    Comprimir eventos baseado em entropia.
    Eventos com baixa entropia podem ser comprimidos mais.
    """
    entropy = calculate_entropy(event_data)
    
    if entropy < threshold:
        # Usar compressão (gzip, etc.)
        return compress_event(event_data)
    else:
        # Alta entropia = dados únicos, não comprimir muito
        return event_data
```

### 4.2. Rate-Distortion Theory

**Aplicação**: Balancear qualidade vs. taxa de transmissão.

**Exemplo**: Agents podem receber eventos "resumidos" ou "completos" baseado em bandwidth disponível.

---

## 🎛️ 5. Control Theory

### 5.1. PID Controller para Auto-Scaling

**Aplicação**: Ajustar número de replicas de agents baseado em métricas.

**Modelo**:
```
Error(t) = Target_Latency - Current_Latency(t)
Replicas(t) = Kp * Error(t) + Ki * ∫Error + Kd * dError/dt
```

**Implementação**:
```yaml
# Knative Service com PID-based scaling
apiVersion: serving.knative.dev/v1
kind: Service
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/metric: "concurrency"
        autoscaling.knative.dev/target: "10"
        # PID parameters
        autoscaling.knative.dev/scaleUpRate: "2.0"   # Kp
        autoscaling.knative.dev/scaleDownRate: "0.5" # Ki
```

### 5.2. Model Predictive Control (MPC)

**Aplicação**: Prever carga futura e ajustar recursos proativamente.

**Exemplo**: Prever picos de eventos baseado em padrões históricos e escalar antecipadamente.

---

## 🔄 6. Graph Theory

### 6.1. Topologia de Comunicação

**Aplicação**: Otimizar roteamento de eventos entre agents.

**Modelos**:
- **Star Topology**: Central broker (atual - RabbitMQ)
- **Mesh Topology**: Agents comunicam diretamente
- **Tree Topology**: Hierarquia de agents

**Otimização**: Encontrar **Minimum Spanning Tree** para reduzir latência total.

### 6.2. PageRank para Agents

**Aplicação**: Identificar agents "importantes" na rede (mais conectados, mais críticos).

**Uso**: Priorizar recursos para agents com maior "centralidade".

---

## 🧮 7. Optimization Algorithms

### 7.1. Linear Programming

**Problema**: Alocar recursos (CPU, memória) entre agents para maximizar throughput.

**Modelo**:
```
Maximize: Σ(throughput_i * x_i)
Subject to:
  Σ(cpu_i * x_i) ≤ Total_CPU
  Σ(memory_i * x_i) ≤ Total_Memory
  x_i ≥ 0 (não-negatividade)
```

### 7.2. Genetic Algorithms

**Aplicação**: Evoluir estratégias de roteamento de eventos.

**Exemplo**: Evoluir quais agents devem processar quais tipos de eventos para maximizar eficiência.

### 7.3. Reinforcement Learning

**Aplicação**: Agents aprendem políticas ótimas de processamento via tentativa e erro.

**Modelo**: **Multi-Agent Reinforcement Learning (MARL)**

```python
class AgentRL:
    def __init__(self):
        self.q_table = {}  # Q-learning table
    
    def choose_action(self, state, event):
        """
        Estado: (agent_load, event_type, queue_depth)
        Ações: (process, forward, reject)
        Recompensa: -latency + throughput - cost
        """
        action = self.epsilon_greedy(state, event)
        return action
    
    def update_policy(self, state, action, reward, next_state):
        # Q-learning update
        self.q_table[state][action] += alpha * (
            reward + gamma * max(self.q_table[next_state]) - 
            self.q_table[state][action]
        )
```

---

## 🎯 Recomendações de Implementação

### Fase 1: Queueing Theory (Mais Imediato)

1. **Instrumentar métricas**:
   - Taxa de chegada de eventos (λ)
   - Taxa de processamento por agent (μ)
   - Tempo médio na fila (W_q)

2. **Aplicar modelo M/M/c**:
   - Calcular número ótimo de agents
   - Ajustar auto-scaling baseado em teoria

3. **Implementar prioridades**:
   - Configurar RabbitMQ com priority queues
   - Modelar como M/M/1 com prioridades

### Fase 2: Game Theory (Médio Prazo)

1. **Implementar Contract Net Protocol**:
   - Agents fazem bids para tarefas
   - Operator seleciona melhor bid

2. **Aplicar Shapley Value**:
   - Distribuir recompensas em tarefas colaborativas

### Fase 3: Control Theory + RL (Longo Prazo)

1. **PID Controller para scaling**:
   - Ajustar replicas baseado em latência

2. **Reinforcement Learning**:
   - Agents aprendem políticas ótimas

---

## 📚 Referências

1. **Game Theory**:
   - "Algorithmic Game Theory" - Nisan et al.
   - "Multi-Agent Systems" - Wooldridge

2. **Queueing Theory**:
   - "Fundamentals of Queueing Theory" - Gross & Harris
   - "Performance Modeling and Design of Computer Systems" - Harchol-Balter

3. **Multi-Agent Systems**:
   - "An Introduction to MultiAgent Systems" - Wooldridge
   - "Distributed Algorithms" - Lynch

4. **Control Theory**:
   - "Feedback Control of Dynamic Systems" - Franklin et al.

5. **Information Theory**:
   - "Elements of Information Theory" - Cover & Thomas

---

## 🔗 Integração com CloudEvents

Todas as otimizações devem manter compatibilidade com CloudEvents v1.0:

```python
# Exemplo: Event otimizado com metadata de teoria
{
    "specversion": "1.0",
    "type": "io.knative.lambda.command.task.announce",
    "source": "knative-lambda-operator",
    "id": "task-123",
    "time": "2025-01-15T10:00:00Z",
    "data": {
        "task_id": "build-function-123",
        "requirements": {...}
    },
    # Extensions para otimização
    "priority": 10,              # Queueing Theory
    "shapley_contribution": 0.3,  # Game Theory
    "expected_latency": "5m",     # Control Theory
    "entropy": 2.5                # Information Theory
}
```

---

## ✅ Conclusão

A combinação de **Game Theory**, **Queueing Theory**, **Control Theory** e **Multi-Agent Systems Optimization** oferece um framework matemático robusto para otimizar a comunicação entre agents no homelab.

**Prioridade de Implementação**:
1. 🥇 **Queueing Theory** - Impacto imediato no desempenho
2. 🥈 **Game Theory (CNP)** - Melhora coordenação
3. 🥉 **Control Theory** - Otimiza auto-scaling
4. 🏅 **Reinforcement Learning** - Aprendizado adaptativo

---

**Autor**: Documento técnico para otimização de agents  
**Última Atualização**: Janeiro 2025
