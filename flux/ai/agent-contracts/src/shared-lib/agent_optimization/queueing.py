"""
FASE 1: Queueing Theory Implementation.

Modelo M/M/c para otimização de agents:
- Cálculo de número ótimo de agents
- Priority queues para eventos
- Previsão de latência e throughput
"""

import math
import heapq
import time
import asyncio
from dataclasses import dataclass, field
from enum import IntEnum
from typing import Any, Optional, List, Tuple, Dict, Callable, Awaitable
import structlog

logger = structlog.get_logger()


class EventPriority(IntEnum):
    """Event priorities for queue ordering."""
    CRITICAL = 1   # io.homelab.alert.critical
    HIGH = 2       # io.homelab.vuln.found
    MEDIUM = 3     # io.homelab.chat.message
    LOW = 4        # io.homelab.analytics.*
    BACKGROUND = 5 # Non-urgent tasks
    
    @classmethod
    def from_event_type(cls, event_type: str) -> "EventPriority":
        """Determine priority from CloudEvent type."""
        if "critical" in event_type or "exploit" in event_type:
            return cls.CRITICAL
        elif "vuln" in event_type or "alert" in event_type:
            return cls.HIGH
        elif "chat" in event_type or "command" in event_type:
            return cls.MEDIUM
        elif "analytics" in event_type or "metric" in event_type:
            return cls.LOW
        else:
            return cls.BACKGROUND


@dataclass
class QueuedEvent:
    """Event in the priority queue."""
    priority: EventPriority
    arrival_time: float
    event_type: str
    event_id: str
    event_data: Any
    deadline: Optional[float] = None  # Absolute time deadline
    
    def __lt__(self, other: "QueuedEvent") -> bool:
        """Compare by priority, then by arrival time (FIFO within priority)."""
        if self.priority != other.priority:
            return self.priority < other.priority
        return self.arrival_time < other.arrival_time


@dataclass
class QueueingMetrics:
    """M/M/c queue metrics."""
    arrival_rate: float        # λ - events/second
    service_rate: float        # μ - events/second per agent
    num_agents: int           # c - number of agents
    utilization: float        # ρ = λ / (c * μ)
    avg_wait_time: float      # Wq - average time in queue
    avg_system_time: float    # W - average time in system
    avg_queue_length: float   # Lq - average queue length
    prob_wait: float          # P(wait > 0) - probability of waiting
    is_stable: bool           # ρ < 1
    recommended_agents: int   # Optimal number of agents


class QueueingOptimizer:
    """
    M/M/c Queue Optimizer.
    
    Calcula métricas e recomendações baseadas em teoria de filas.
    """
    
    def __init__(self, target_latency: float = 1.0, max_utilization: float = 0.8):
        """
        Args:
            target_latency: Target average wait time in seconds
            max_utilization: Maximum acceptable utilization (0-1)
        """
        self.target_latency = target_latency
        self.max_utilization = max_utilization
    
    def calculate_metrics(
        self,
        arrival_rate: float,
        service_rate: float,
        num_agents: int
    ) -> QueueingMetrics:
        """
        Calculate M/M/c queue metrics.
        
        Args:
            arrival_rate: λ - events arriving per second
            service_rate: μ - events processed per second (per agent)
            num_agents: c - number of agents
        """
        if num_agents <= 0 or service_rate <= 0:
            return QueueingMetrics(
                arrival_rate=arrival_rate,
                service_rate=service_rate,
                num_agents=num_agents,
                utilization=float('inf'),
                avg_wait_time=float('inf'),
                avg_system_time=float('inf'),
                avg_queue_length=float('inf'),
                prob_wait=1.0,
                is_stable=False,
                recommended_agents=max(1, int(math.ceil(arrival_rate / service_rate)) + 1)
            )
        
        # Total service rate
        total_service_rate = num_agents * service_rate
        
        # Utilization (ρ = λ / (c * μ))
        utilization = arrival_rate / total_service_rate
        is_stable = utilization < 1.0
        
        if not is_stable:
            # System unstable - calculate what we can
            recommended = int(math.ceil(arrival_rate / (service_rate * self.max_utilization)))
            return QueueingMetrics(
                arrival_rate=arrival_rate,
                service_rate=service_rate,
                num_agents=num_agents,
                utilization=utilization,
                avg_wait_time=float('inf'),
                avg_system_time=float('inf'),
                avg_queue_length=float('inf'),
                prob_wait=1.0,
                is_stable=False,
                recommended_agents=recommended
            )
        
        # Erlang C formula components
        rho = arrival_rate / service_rate  # Offered load
        
        # P0 - Probability of empty system
        p0 = self._calculate_p0(rho, num_agents)
        
        # Erlang C - Probability of waiting
        prob_wait = self._erlang_c(rho, num_agents, p0)
        
        # Average wait time in queue (Wq)
        avg_wait_time = prob_wait / (num_agents * service_rate - arrival_rate)
        
        # Average time in system (W = Wq + 1/μ)
        avg_system_time = avg_wait_time + (1.0 / service_rate)
        
        # Average queue length (Lq = λ * Wq) - Little's Law
        avg_queue_length = arrival_rate * avg_wait_time
        
        # Calculate recommended agents for target latency
        recommended_agents = self._optimize_agents(
            arrival_rate, service_rate, num_agents
        )
        
        return QueueingMetrics(
            arrival_rate=arrival_rate,
            service_rate=service_rate,
            num_agents=num_agents,
            utilization=utilization,
            avg_wait_time=avg_wait_time,
            avg_system_time=avg_system_time,
            avg_queue_length=avg_queue_length,
            prob_wait=prob_wait,
            is_stable=is_stable,
            recommended_agents=recommended_agents
        )
    
    def _calculate_p0(self, rho: float, c: int) -> float:
        """Calculate P0 (probability of empty system)."""
        sum_term = sum(
            (rho ** k) / math.factorial(k)
            for k in range(c)
        )
        
        last_term = ((rho ** c) / math.factorial(c)) * (c / (c - rho))
        
        p0 = 1.0 / (sum_term + last_term)
        return p0
    
    def _erlang_c(self, rho: float, c: int, p0: float) -> float:
        """Calculate Erlang C (probability of waiting)."""
        numerator = ((rho ** c) / math.factorial(c)) * (c / (c - rho)) * p0
        return numerator
    
    def _optimize_agents(
        self,
        arrival_rate: float,
        service_rate: float,
        current_agents: int
    ) -> int:
        """Find optimal number of agents for target latency."""
        # Start with minimum stable configuration
        min_agents = max(1, int(math.ceil(arrival_rate / service_rate)))
        
        for agents in range(min_agents, min_agents + 20):
            # Calculate metrics directly without recursion
            utilization = arrival_rate / (agents * service_rate)
            
            if utilization >= 1.0:
                continue  # Not stable
            
            rho = arrival_rate / service_rate
            if rho >= agents:
                continue  # Not stable
            
            try:
                p0 = self._calculate_p0(rho, agents)
                prob_wait = self._erlang_c(rho, agents, p0)
                avg_wait_time = prob_wait / (agents * service_rate - arrival_rate)
                
                if (avg_wait_time <= self.target_latency and
                    utilization <= self.max_utilization):
                    return agents
            except (ZeroDivisionError, ValueError):
                continue
        
        # Fallback
        return min_agents + 5
    
    def recommend_scaling(
        self,
        current_metrics: QueueingMetrics
    ) -> Tuple[str, int, str]:
        """
        Recommend scaling action.
        
        Returns:
            Tuple of (action, target_agents, reason)
            action: "scale_up", "scale_down", "no_change"
        """
        if not current_metrics.is_stable:
            return (
                "scale_up",
                current_metrics.recommended_agents,
                f"System unstable (ρ={current_metrics.utilization:.2f} >= 1)"
            )
        
        if current_metrics.avg_wait_time > self.target_latency:
            return (
                "scale_up",
                current_metrics.recommended_agents,
                f"Latency {current_metrics.avg_wait_time:.2f}s > target {self.target_latency}s"
            )
        
        if (current_metrics.utilization < self.max_utilization * 0.5 and 
            current_metrics.num_agents > 1):
            return (
                "scale_down",
                max(1, current_metrics.num_agents - 1),
                f"Low utilization ({current_metrics.utilization:.2%})"
            )
        
        return (
            "no_change",
            current_metrics.num_agents,
            f"System stable (ρ={current_metrics.utilization:.2%}, Wq={current_metrics.avg_wait_time:.2f}s)"
        )


class PriorityQueue:
    """
    Async priority queue for CloudEvents.
    
    Uses heap-based priority queue with configurable limits.
    Implements backpressure when queue is full.
    """
    
    def __init__(
        self,
        max_size: int = 10000,
        max_wait_time: float = 30.0,
        on_deadline_miss: Optional[Callable[[QueuedEvent], Awaitable[None]]] = None
    ):
        """
        Args:
            max_size: Maximum queue size (backpressure threshold)
            max_wait_time: Maximum time event can wait in queue
            on_deadline_miss: Callback when event misses deadline
        """
        self.max_size = max_size
        self.max_wait_time = max_wait_time
        self.on_deadline_miss = on_deadline_miss
        
        self._queue: List[QueuedEvent] = []
        self._lock = asyncio.Lock()
        self._not_empty = asyncio.Condition()
        self._not_full = asyncio.Condition()
        
        # Metrics
        self._enqueued = 0
        self._dequeued = 0
        self._dropped = 0
        self._deadline_misses = 0
    
    async def enqueue(
        self,
        event_type: str,
        event_id: str,
        event_data: Any,
        priority: Optional[EventPriority] = None,
        deadline: Optional[float] = None,
        timeout: float = 5.0
    ) -> bool:
        """
        Enqueue event with priority.
        
        Returns:
            True if enqueued, False if dropped (queue full)
        """
        if priority is None:
            priority = EventPriority.from_event_type(event_type)
        
        arrival_time = time.time()
        
        if deadline is None:
            deadline = arrival_time + self.max_wait_time
        
        event = QueuedEvent(
            priority=priority,
            arrival_time=arrival_time,
            event_type=event_type,
            event_id=event_id,
            event_data=event_data,
            deadline=deadline
        )
        
        async with self._not_full:
            # Wait for space with timeout
            try:
                await asyncio.wait_for(
                    self._wait_for_space(),
                    timeout=timeout
                )
            except asyncio.TimeoutError:
                self._dropped += 1
                logger.warning(
                    "queue_event_dropped",
                    event_id=event_id,
                    event_type=event_type,
                    reason="queue_full_timeout"
                )
                return False
        
        async with self._lock:
            heapq.heappush(self._queue, event)
            self._enqueued += 1
        
        async with self._not_empty:
            self._not_empty.notify()
        
        return True
    
    async def _wait_for_space(self):
        """Wait until queue has space."""
        while len(self._queue) >= self.max_size:
            await self._not_full.wait()
    
    async def dequeue(self, timeout: Optional[float] = None) -> Optional[QueuedEvent]:
        """
        Dequeue highest priority event.
        
        Checks deadline and calls callback if missed.
        """
        async with self._not_empty:
            if timeout is not None:
                try:
                    await asyncio.wait_for(
                        self._wait_for_event(),
                        timeout=timeout
                    )
                except asyncio.TimeoutError:
                    return None
            else:
                await self._wait_for_event()
        
        async with self._lock:
            if not self._queue:
                return None
            
            event = heapq.heappop(self._queue)
            self._dequeued += 1
        
        async with self._not_full:
            self._not_full.notify()
        
        # Check deadline
        now = time.time()
        if event.deadline and now > event.deadline:
            self._deadline_misses += 1
            logger.warning(
                "queue_deadline_missed",
                event_id=event.event_id,
                event_type=event.event_type,
                wait_time=now - event.arrival_time,
                deadline=event.deadline
            )
            if self.on_deadline_miss:
                await self.on_deadline_miss(event)
        
        return event
    
    async def _wait_for_event(self):
        """Wait until queue has events."""
        while not self._queue:
            await self._not_empty.wait()
    
    def size(self) -> int:
        """Current queue size."""
        return len(self._queue)
    
    def size_by_priority(self) -> Dict[EventPriority, int]:
        """Queue size by priority level."""
        counts = {p: 0 for p in EventPriority}
        for event in self._queue:
            counts[event.priority] += 1
        return counts
    
    def get_stats(self) -> Dict[str, Any]:
        """Get queue statistics."""
        return {
            "size": len(self._queue),
            "enqueued": self._enqueued,
            "dequeued": self._dequeued,
            "dropped": self._dropped,
            "deadline_misses": self._deadline_misses,
            "utilization": len(self._queue) / self.max_size if self.max_size > 0 else 0,
        }
    
    def get_wait_time(self) -> float:
        """Estimate current wait time for new events."""
        if not self._queue:
            return 0.0
        
        # Average of all waiting events
        now = time.time()
        wait_times = [now - e.arrival_time for e in self._queue]
        return sum(wait_times) / len(wait_times) if wait_times else 0.0
