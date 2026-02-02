"""
Event Decision Engine - Integrates All Optimization Theories.

O agent decide se deve processar um evento baseado em:
- Sua capacidade atual (CPU/memória) - Queueing Theory
- Prioridade do evento - Priority Queue
- Custo de processamento - Game Theory
- Recompensa esperada (útil para o sistema) - Nash Equilibrium
"""

import time
import asyncio
from dataclasses import dataclass
from typing import Any, Optional, List, Dict, Tuple
from enum import Enum
import structlog

from .queueing import (
    QueueingOptimizer, 
    QueueingMetrics,
    PriorityQueue,
    EventPriority,
    QueuedEvent
)
from .gametheory import (
    AgentState,
    TaskBid,
    Task,
    ContractNetProtocol,
    NashEquilibrium
)
from .metrics import (
    OptimizationMetrics,
    DECISIONS_MADE,
    DECISION_LATENCY,
    BIDS_SUBMITTED,
    BIDS_WON,
    BID_UTILITY,
)

logger = structlog.get_logger()


class DecisionReason(Enum):
    """Reasons for processing decisions."""
    # Process reasons
    HIGH_PRIORITY = "high_priority"
    GOOD_CAPACITY = "good_capacity"
    SPECIALIZATION = "specialization"
    BEST_BID = "best_bid"
    NASH_OPTIMAL = "nash_optimal"
    
    # Forward reasons
    LOW_CAPACITY = "low_capacity"
    BETTER_TARGET = "better_target"
    OVERLOADED = "overloaded"
    
    # Reject reasons
    NO_CAPACITY = "no_capacity"
    DEADLINE_MISSED = "deadline_missed"
    QUEUE_FULL = "queue_full"
    SYSTEM_UNSTABLE = "system_unstable"


@dataclass
class ProcessingDecision:
    """
    Decision on whether to process an event.
    
    Resultado do Decision Engine.
    """
    decision: str           # "process", "forward", "reject"
    reason: DecisionReason
    confidence: float       # 0-1
    
    # Event info
    event_id: str
    event_type: str
    priority: EventPriority
    
    # Agent state at decision time
    agent_utilization: float
    agent_capacity_available: float
    
    # For forwarding
    forward_target: Optional[str] = None
    
    # Game theory metrics
    bid_utility: Optional[float] = None
    nash_payoff: Optional[float] = None
    
    # Timing
    decision_time_ms: float = 0.0


class EventDecisionEngine:
    """
    Engine that decides how to handle incoming CloudEvents.
    
    Integra todas as teorias de otimização:
    - Queueing Theory: verifica capacidade e estabilidade do sistema
    - Game Theory: calcula utilidade e Nash equilibrium
    - Control Theory: considera estado do auto-scaler
    
    O agent decide se deve:
    1. PROCESS - Processar o evento localmente
    2. FORWARD - Encaminhar para outro agent
    3. REJECT - Rejeitar (backpressure)
    """
    
    def __init__(
        self,
        agent_id: str,
        agent_state: AgentState,
        queue: PriorityQueue,
        metrics: OptimizationMetrics,
        
        # Thresholds
        max_utilization: float = 0.85,
        min_utility_threshold: float = 0.3,
        
        # Other agents for forwarding
        other_agents: Optional[List[AgentState]] = None,
        
        # Optional components
        contract_net: Optional[ContractNetProtocol] = None,
        queueing_optimizer: Optional[QueueingOptimizer] = None
    ):
        """
        Args:
            agent_id: This agent's ID
            agent_state: Current agent state
            queue: Priority queue for events
            metrics: Metrics collector
            max_utilization: Maximum utilization before rejecting
            min_utility_threshold: Minimum Nash utility to process
            other_agents: Other agents available for forwarding
            contract_net: Contract Net Protocol instance
            queueing_optimizer: Queueing theory optimizer
        """
        self.agent_id = agent_id
        self.agent_state = agent_state
        self.queue = queue
        self.metrics = metrics
        self.max_utilization = max_utilization
        self.min_utility_threshold = min_utility_threshold
        self.other_agents = other_agents or []
        
        # Game theory components
        self.contract_net = contract_net or ContractNetProtocol()
        self.nash = NashEquilibrium()
        
        # Queueing theory
        self.queueing = queueing_optimizer or QueueingOptimizer(target_latency=1.0)
        
        # Register self in contract net
        self.contract_net.register_agent(agent_state)
        for agent in self.other_agents:
            self.contract_net.register_agent(agent)
    
    def update_state(
        self,
        cpu_used: Optional[float] = None,
        memory_used: Optional[float] = None,
        avg_processing_time: Optional[float] = None,
        success_rate: Optional[float] = None
    ):
        """Update agent state for decision making."""
        if cpu_used is not None:
            self.agent_state.cpu_used = cpu_used
        if memory_used is not None:
            self.agent_state.memory_used = memory_used
        if avg_processing_time is not None:
            self.agent_state.avg_processing_time = avg_processing_time
        if success_rate is not None:
            self.agent_state.success_rate = success_rate
        
        # Update metrics
        self.metrics.update_load(
            cpu=self.agent_state.cpu_used,
            memory=self.agent_state.memory_used,
            connections=0
        )
    
    async def decide(
        self,
        event_id: str,
        event_type: str,
        event_data: Any,
        priority: Optional[EventPriority] = None,
        deadline: Optional[float] = None,
        reward: float = 10.0
    ) -> ProcessingDecision:
        """
        Decide how to handle an incoming event.
        
        Uses all optimization theories to make the best decision.
        
        Args:
            event_id: CloudEvent ID
            event_type: CloudEvent type
            event_data: Event payload
            priority: Event priority (auto-detected if None)
            deadline: Absolute deadline (Unix timestamp)
            reward: Reward for processing this event
        
        Returns:
            ProcessingDecision with action and reasoning
        """
        start_time = time.time()
        
        # Record arrival
        self.metrics.record_arrival(event_type, priority.name if priority else "auto")
        
        # Auto-detect priority
        if priority is None:
            priority = EventPriority.from_event_type(event_type)
        
        # Create task for Game Theory calculations
        task = Task(
            task_id=event_id,
            event_type=event_type,
            event_data=event_data,
            priority=priority.value,
            deadline=deadline,
            reward=reward
        )
        
        # =====================================================================
        # FASE 1: Queueing Theory - Check system stability
        # =====================================================================
        arrival_rate = self.metrics.get_arrival_rate()
        service_rate = self.metrics.get_service_rate()
        utilization = self.metrics.get_utilization()
        
        queue_metrics = self.queueing.calculate_metrics(
            arrival_rate=arrival_rate,
            service_rate=service_rate,
            num_agents=1  # This agent
        )
        
        # System unstable - reject
        if not queue_metrics.is_stable:
            decision = self._make_decision(
                decision="reject",
                reason=DecisionReason.SYSTEM_UNSTABLE,
                confidence=0.9,
                event_id=event_id,
                event_type=event_type,
                priority=priority,
                utilization=utilization,
                start_time=start_time
            )
            self.metrics.record_rejection(event_type, "system_unstable")
            DECISIONS_MADE.labels(agent_id=self.agent_id, decision="reject").inc()
            return decision
        
        # =====================================================================
        # FASE 2: Game Theory - Calculate utility and Nash equilibrium
        # =====================================================================
        
        # Calculate our bid
        bid = self.contract_net.calculate_bid(self.agent_state, task)
        bid_utility = bid.utility if bid else 0.0
        
        # Calculate Nash equilibrium best response
        nash_strategy, nash_payoff = self.nash.calculate_best_response(
            self.agent_state,
            task,
            self.other_agents
        )
        
        # =====================================================================
        # Decision Logic
        # =====================================================================
        
        # CRITICAL priority - always try to process
        if priority == EventPriority.CRITICAL:
            if self.agent_state.can_process(task.cpu_required, task.memory_required):
                decision = self._make_decision(
                    decision="process",
                    reason=DecisionReason.HIGH_PRIORITY,
                    confidence=0.95,
                    event_id=event_id,
                    event_type=event_type,
                    priority=priority,
                    utilization=utilization,
                    start_time=start_time,
                    bid_utility=bid_utility,
                    nash_payoff=nash_payoff
                )
                DECISIONS_MADE.labels(agent_id=self.agent_id, decision="process").inc()
                return decision
        
        # Check capacity
        if not self.agent_state.can_process(task.cpu_required, task.memory_required):
            # Try to forward
            best_target = self._find_best_forward_target(task)
            if best_target:
                decision = self._make_decision(
                    decision="forward",
                    reason=DecisionReason.NO_CAPACITY,
                    confidence=0.8,
                    event_id=event_id,
                    event_type=event_type,
                    priority=priority,
                    utilization=utilization,
                    start_time=start_time,
                    forward_target=best_target.agent_id,
                    bid_utility=bid_utility,
                    nash_payoff=nash_payoff
                )
                self.metrics.record_forward(event_type, best_target.agent_id)
                DECISIONS_MADE.labels(agent_id=self.agent_id, decision="forward").inc()
                return decision
            else:
                decision = self._make_decision(
                    decision="reject",
                    reason=DecisionReason.NO_CAPACITY,
                    confidence=0.9,
                    event_id=event_id,
                    event_type=event_type,
                    priority=priority,
                    utilization=utilization,
                    start_time=start_time,
                    bid_utility=bid_utility,
                    nash_payoff=nash_payoff
                )
                self.metrics.record_rejection(event_type, "no_capacity")
                DECISIONS_MADE.labels(agent_id=self.agent_id, decision="reject").inc()
                return decision
        
        # Check utilization threshold
        if utilization > self.max_utilization:
            best_target = self._find_best_forward_target(task)
            if best_target:
                decision = self._make_decision(
                    decision="forward",
                    reason=DecisionReason.OVERLOADED,
                    confidence=0.75,
                    event_id=event_id,
                    event_type=event_type,
                    priority=priority,
                    utilization=utilization,
                    start_time=start_time,
                    forward_target=best_target.agent_id,
                    bid_utility=bid_utility,
                    nash_payoff=nash_payoff
                )
                self.metrics.record_forward(event_type, best_target.agent_id)
                DECISIONS_MADE.labels(agent_id=self.agent_id, decision="forward").inc()
                return decision
        
        # Use Nash equilibrium strategy
        if nash_strategy == NashEquilibrium.Strategy.PROCESS:
            decision = self._make_decision(
                decision="process",
                reason=DecisionReason.NASH_OPTIMAL,
                confidence=min(0.9, 0.5 + nash_payoff),
                event_id=event_id,
                event_type=event_type,
                priority=priority,
                utilization=utilization,
                start_time=start_time,
                bid_utility=bid_utility,
                nash_payoff=nash_payoff
            )
            DECISIONS_MADE.labels(agent_id=self.agent_id, decision="process").inc()
            return decision
        
        elif nash_strategy == NashEquilibrium.Strategy.FORWARD:
            best_target = self._find_best_forward_target(task)
            if best_target:
                decision = self._make_decision(
                    decision="forward",
                    reason=DecisionReason.BETTER_TARGET,
                    confidence=0.7,
                    event_id=event_id,
                    event_type=event_type,
                    priority=priority,
                    utilization=utilization,
                    start_time=start_time,
                    forward_target=best_target.agent_id,
                    bid_utility=bid_utility,
                    nash_payoff=nash_payoff
                )
                self.metrics.record_forward(event_type, best_target.agent_id)
                DECISIONS_MADE.labels(agent_id=self.agent_id, decision="forward").inc()
                return decision
        
        # Default: process if utility is above threshold
        if bid_utility >= self.min_utility_threshold:
            decision = self._make_decision(
                decision="process",
                reason=DecisionReason.GOOD_CAPACITY,
                confidence=0.6 + bid_utility * 0.3,
                event_id=event_id,
                event_type=event_type,
                priority=priority,
                utilization=utilization,
                start_time=start_time,
                bid_utility=bid_utility,
                nash_payoff=nash_payoff
            )
            DECISIONS_MADE.labels(agent_id=self.agent_id, decision="process").inc()
            return decision
        
        # Low utility - reject
        decision = self._make_decision(
            decision="reject",
            reason=DecisionReason.LOW_CAPACITY,
            confidence=0.5,
            event_id=event_id,
            event_type=event_type,
            priority=priority,
            utilization=utilization,
            start_time=start_time,
            bid_utility=bid_utility,
            nash_payoff=nash_payoff
        )
        self.metrics.record_rejection(event_type, "low_utility")
        DECISIONS_MADE.labels(agent_id=self.agent_id, decision="reject").inc()
        return decision
    
    def _find_best_forward_target(self, task: Task) -> Optional[AgentState]:
        """Find best agent to forward task to."""
        best_agent = None
        best_utility = -float('inf')
        
        for agent in self.other_agents:
            if not agent.can_process(task.cpu_required, task.memory_required):
                continue
            
            bid = self.contract_net.calculate_bid(agent, task)
            if bid and bid.utility > best_utility:
                best_utility = bid.utility
                best_agent = agent
        
        return best_agent
    
    def _make_decision(
        self,
        decision: str,
        reason: DecisionReason,
        confidence: float,
        event_id: str,
        event_type: str,
        priority: EventPriority,
        utilization: float,
        start_time: float,
        forward_target: Optional[str] = None,
        bid_utility: Optional[float] = None,
        nash_payoff: Optional[float] = None
    ) -> ProcessingDecision:
        """Create ProcessingDecision with timing."""
        decision_time_ms = (time.time() - start_time) * 1000
        
        DECISION_LATENCY.labels(agent_id=self.agent_id).observe(decision_time_ms / 1000)
        
        logger.info(
            "event_decision",
            agent_id=self.agent_id,
            event_id=event_id,
            event_type=event_type,
            decision=decision,
            reason=reason.value,
            confidence=confidence,
            utilization=utilization,
            bid_utility=bid_utility,
            nash_payoff=nash_payoff,
            decision_time_ms=decision_time_ms
        )
        
        return ProcessingDecision(
            decision=decision,
            reason=reason,
            confidence=confidence,
            event_id=event_id,
            event_type=event_type,
            priority=priority,
            agent_utilization=utilization,
            agent_capacity_available=(
                self.agent_state.cpu_available + self.agent_state.memory_available
            ) / 2,
            forward_target=forward_target,
            bid_utility=bid_utility,
            nash_payoff=nash_payoff,
            decision_time_ms=decision_time_ms
        )
