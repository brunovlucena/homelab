"""
Agent Communication Optimization Library.

Implementa teorias matemáticas para otimizar comunicação entre agents:
- Queueing Theory (Fase 1)
- Game Theory (Fase 2)
- Control Theory + RL (Fase 3)
"""

from .queueing import (
    QueueingMetrics,
    QueueingOptimizer,
    PriorityQueue,
    EventPriority,
)
from .gametheory import (
    AgentState,
    TaskBid,
    ContractNetProtocol,
    ShapleyCalculator,
    NashEquilibrium,
)
from .control import (
    PIDController,
    AutoScaler,
    ScalingDecision,
)
from .decision import (
    EventDecisionEngine,
    ProcessingDecision,
    DecisionReason,
)
from .metrics import (
    OptimizationMetrics,
    setup_optimization_metrics,
)

__version__ = "1.0.0"
__all__ = [
    # Queueing Theory
    "QueueingMetrics",
    "QueueingOptimizer", 
    "PriorityQueue",
    "EventPriority",
    # Game Theory
    "AgentState",
    "TaskBid",
    "ContractNetProtocol",
    "ShapleyCalculator",
    "NashEquilibrium",
    # Control Theory
    "PIDController",
    "AutoScaler",
    "ScalingDecision",
    # Decision Engine
    "EventDecisionEngine",
    "ProcessingDecision",
    "DecisionReason",
    # Metrics
    "OptimizationMetrics",
    "setup_optimization_metrics",
]
