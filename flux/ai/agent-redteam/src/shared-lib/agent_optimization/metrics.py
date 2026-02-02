"""
Prometheus metrics for agent optimization.

REUTILIZA MÉTRICAS EXISTENTES do knative-lambda-operator e RabbitMQ:

KNATIVE-LAMBDA-OPERATOR METRICS (já disponíveis no Prometheus):
- knative_lambda_operator_workqueue_depth          → Queue Depth
- knative_lambda_operator_workqueue_latency_seconds → Queue Latency  
- knative_lambda_function_invocations_total        → Arrival Rate (λ)
- knative_lambda_function_duration_seconds         → Service Time (μ)
- knative_lambda_function_errors_total             → Error Rate
- knative_lambda_function_cold_starts_total        → Cold Starts

RABBITMQ METRICS (já disponíveis no Prometheus):
- rabbitmq_queue_messages                          → Queue Depth
- rabbitmq_queue_messages_published_total          → Arrival Rate (λ)
- rabbitmq_queue_messages_delivered_total          → Service Rate (μ)

Este módulo CONSULTA essas métricas via PromQL para calcular otimização.
"""

import time
import os
from typing import Optional, Dict, Any, List
from dataclasses import dataclass, field
from prometheus_client import Counter, Gauge, Histogram, CollectorRegistry, REGISTRY
import httpx
import structlog

logger = structlog.get_logger()


# =============================================================================
# EXISTING METRIC NAMES (from knative-lambda-operator and RabbitMQ)
# =============================================================================

class ExistingMetrics:
    """
    Nomes das métricas EXISTENTES no Prometheus.
    
    Estas métricas já são coletadas pelo:
    - knative-lambda-operator (Go)
    - RabbitMQ ServiceMonitor
    """
    
    # Knative Lambda Operator Metrics
    WORKQUEUE_DEPTH = "knative_lambda_operator_workqueue_depth"
    WORKQUEUE_LATENCY = "knative_lambda_operator_workqueue_latency_seconds"
    RECONCILE_TOTAL = "knative_lambda_operator_reconcile_total"
    RECONCILE_DURATION = "knative_lambda_operator_reconcile_duration_seconds"
    
    # Function RED Metrics (Rate, Errors, Duration)
    FUNCTION_INVOCATIONS = "knative_lambda_function_invocations_total"  # Rate (λ)
    FUNCTION_DURATION = "knative_lambda_function_duration_seconds"       # Duration (μ)
    FUNCTION_ERRORS = "knative_lambda_function_errors_total"             # Errors
    FUNCTION_COLD_STARTS = "knative_lambda_function_cold_starts_total"
    
    # RabbitMQ Metrics
    RABBITMQ_QUEUE_MESSAGES = "rabbitmq_queue_messages"
    RABBITMQ_QUEUE_MESSAGES_UNACKED = "rabbitmq_queue_messages_unacked"
    RABBITMQ_QUEUE_PUBLISHED = "rabbitmq_queue_messages_published_total"   # λ
    RABBITMQ_QUEUE_DELIVERED = "rabbitmq_queue_messages_delivered_total"   # μ


# =============================================================================
# PROMETHEUS QUERIES for Queueing Theory
# =============================================================================

class QueueingTheoryQueries:
    """
    PromQL queries para calcular métricas de Queueing Theory.
    
    Baseado no modelo M/M/c:
    - λ (lambda) = arrival rate = taxa de chegada de eventos
    - μ (mu) = service rate = taxa de processamento por agent
    - c = número de agents
    - ρ (rho) = utilization = λ / (c * μ)
    """
    
    @staticmethod
    def arrival_rate(namespace: str = ".*", function: str = ".*", window: str = "5m") -> str:
        """
        Taxa de chegada de eventos (λ) - events/second.
        
        Usa: knative_lambda_function_invocations_total
        """
        return f'sum(rate({ExistingMetrics.FUNCTION_INVOCATIONS}{{namespace=~"{namespace}", function=~"{function}"}}[{window}]))'
    
    @staticmethod
    def service_rate(namespace: str = ".*", function: str = ".*", window: str = "5m") -> str:
        """
        Taxa de processamento (μ) - events/second.
        
        Usa: 1 / avg_duration (from histogram)
        """
        return f'''
            1 / (
                sum(rate({ExistingMetrics.FUNCTION_DURATION}_sum{{namespace=~"{namespace}", function=~"{function}"}}[{window}])) 
                / 
                sum(rate({ExistingMetrics.FUNCTION_DURATION}_count{{namespace=~"{namespace}", function=~"{function}"}}[{window}]))
            )
        '''
    
    @staticmethod
    def utilization(namespace: str = ".*", window: str = "5m") -> str:
        """
        Utilização do sistema (ρ = λ / μ).
        """
        return f'''
            sum(rate({ExistingMetrics.FUNCTION_INVOCATIONS}{{namespace=~"{namespace}"}}[{window}]))
            /
            (
                1 / (
                    sum(rate({ExistingMetrics.FUNCTION_DURATION}_sum{{namespace=~"{namespace}"}}[{window}])) 
                    / 
                    sum(rate({ExistingMetrics.FUNCTION_DURATION}_count{{namespace=~"{namespace}"}}[{window}]))
                )
            )
        '''
    
    @staticmethod
    def queue_depth() -> str:
        """
        Profundidade da fila (operator work queue).
        """
        return f"sum({ExistingMetrics.WORKQUEUE_DEPTH})"
    
    @staticmethod
    def queue_latency_p95(window: str = "5m") -> str:
        """
        P95 latência na fila (segundos).
        """
        return f"histogram_quantile(0.95, sum by (le) (rate({ExistingMetrics.WORKQUEUE_LATENCY}_bucket[{window}])))"
    
    @staticmethod
    def error_rate(namespace: str = ".*", function: str = ".*", window: str = "5m") -> str:
        """
        Taxa de erros (para SLO).
        """
        return f'''
            sum(rate({ExistingMetrics.FUNCTION_ERRORS}{{namespace=~"{namespace}", function=~"{function}"}}[{window}]))
            /
            sum(rate({ExistingMetrics.FUNCTION_INVOCATIONS}{{namespace=~"{namespace}", function=~"{function}"}}[{window}]))
        '''
    
    @staticmethod
    def rabbitmq_arrival_rate(queue: str = ".*", window: str = "5m") -> str:
        """
        Taxa de chegada no RabbitMQ (λ).
        """
        return f'sum(rate({ExistingMetrics.RABBITMQ_QUEUE_PUBLISHED}{{queue=~"{queue}"}}[{window}]))'
    
    @staticmethod
    def rabbitmq_service_rate(queue: str = ".*", window: str = "5m") -> str:
        """
        Taxa de processamento do RabbitMQ (μ).
        """
        return f'sum(rate({ExistingMetrics.RABBITMQ_QUEUE_DELIVERED}{{queue=~"{queue}"}}[{window}]))'
    
    @staticmethod
    def rabbitmq_queue_depth(queue: str = ".*") -> str:
        """
        Profundidade da fila RabbitMQ.
        """
        return f'sum({ExistingMetrics.RABBITMQ_QUEUE_MESSAGES}{{queue=~"{queue}"}})'


# =============================================================================
# AGENT-SPECIFIC METRICS (new metrics for optimization)
# =============================================================================

# Decision Engine Metrics (NEW - specific to optimization)
DECISIONS_MADE = Counter(
    "agent_optimization_decisions_total",
    "Total decisions made by decision engine",
    ["agent_id", "decision"],  # process, forward, reject
    registry=REGISTRY
)

DECISION_LATENCY = Histogram(
    "agent_optimization_decision_latency_seconds",
    "Time to make processing decision",
    ["agent_id"],
    buckets=[0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1],
    registry=REGISTRY
)

# Game Theory Metrics (NEW)
BIDS_SUBMITTED = Counter(
    "agent_optimization_bids_submitted_total",
    "Contract Net Protocol bids submitted",
    ["agent_id", "task_type"],
    registry=REGISTRY
)

BIDS_WON = Counter(
    "agent_optimization_bids_won_total",
    "Contract Net Protocol bids won",
    ["agent_id", "task_type"],
    registry=REGISTRY
)

BID_UTILITY = Histogram(
    "agent_optimization_bid_utility",
    "Bid utility values (Nash equilibrium)",
    ["agent_id", "task_type"],
    buckets=[0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
    registry=REGISTRY
)

SHAPLEY_VALUE = Gauge(
    "agent_optimization_shapley_value",
    "Shapley value contribution in cooperative tasks",
    ["agent_id", "coalition_id"],
    registry=REGISTRY
)

# Control Theory Metrics (NEW)
SCALING_DECISIONS = Counter(
    "agent_optimization_scaling_decisions_total",
    "Auto-scaling decisions made by PID controller",
    ["agent_id", "direction"],  # up, down, none
    registry=REGISTRY
)

PID_ERROR = Gauge(
    "agent_optimization_pid_error",
    "PID controller error signal",
    ["agent_id", "metric_type"],  # latency, utilization
    registry=REGISTRY
)


# =============================================================================
# PROMETHEUS CLIENT - Query existing metrics
# =============================================================================

class PrometheusClient:
    """
    Cliente para consultar métricas EXISTENTES do Prometheus.
    
    Usa PromQL para calcular λ, μ, ρ a partir das métricas já coletadas.
    """
    
    def __init__(self, prometheus_url: Optional[str] = None):
        """
        Args:
            prometheus_url: URL do Prometheus (default: from env or cluster DNS)
        """
        self.prometheus_url = prometheus_url or os.getenv(
            "PROMETHEUS_URL",
            "http://kube-prometheus-stack-prometheus.prometheus.svc.cluster.local:9090"
        )
        self._client: Optional[httpx.AsyncClient] = None
    
    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client."""
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(timeout=30.0)
        return self._client
    
    async def query(self, promql: str) -> Optional[float]:
        """
        Execute PromQL query and return scalar result.
        
        Returns None if query fails or returns no data.
        """
        try:
            client = await self._get_client()
            response = await client.get(
                f"{self.prometheus_url}/api/v1/query",
                params={"query": promql.strip()}
            )
            response.raise_for_status()
            
            data = response.json()
            if data["status"] != "success":
                logger.warning("prometheus_query_failed", query=promql, status=data["status"])
                return None
            
            result = data.get("data", {}).get("result", [])
            if not result:
                return None
            
            # Get first result value
            value = result[0].get("value", [None, None])[1]
            return float(value) if value is not None else None
            
        except Exception as e:
            logger.error("prometheus_query_error", query=promql[:100], error=str(e))
            return None
    
    async def get_arrival_rate(
        self, 
        namespace: str = ".*", 
        function: str = ".*",
        window: str = "5m"
    ) -> float:
        """
        Get arrival rate (λ) from Prometheus.
        
        Uses: knative_lambda_function_invocations_total
        """
        query = QueueingTheoryQueries.arrival_rate(namespace, function, window)
        result = await self.query(query)
        return result if result is not None else 0.0
    
    async def get_service_rate(
        self,
        namespace: str = ".*",
        function: str = ".*", 
        window: str = "5m"
    ) -> float:
        """
        Get service rate (μ) from Prometheus.
        
        Calculates: 1 / avg_duration
        """
        query = QueueingTheoryQueries.service_rate(namespace, function, window)
        result = await self.query(query)
        return result if result is not None else 1.0
    
    async def get_utilization(self, namespace: str = ".*", window: str = "5m") -> float:
        """
        Get system utilization (ρ = λ / μ).
        """
        arrival_rate = await self.get_arrival_rate(namespace, ".*", window)
        service_rate = await self.get_service_rate(namespace, ".*", window)
        
        if service_rate <= 0:
            return float('inf') if arrival_rate > 0 else 0.0
        
        return arrival_rate / service_rate
    
    async def get_queue_depth(self) -> float:
        """Get operator work queue depth."""
        query = QueueingTheoryQueries.queue_depth()
        result = await self.query(query)
        return result if result is not None else 0.0
    
    async def get_queue_latency_p95(self, window: str = "5m") -> float:
        """Get P95 queue latency (seconds)."""
        query = QueueingTheoryQueries.queue_latency_p95(window)
        result = await self.query(query)
        return result if result is not None else 0.0
    
    async def get_error_rate(
        self,
        namespace: str = ".*",
        function: str = ".*",
        window: str = "5m"
    ) -> float:
        """Get error rate (for SLO)."""
        query = QueueingTheoryQueries.error_rate(namespace, function, window)
        result = await self.query(query)
        return result if result is not None else 0.0
    
    async def get_rabbitmq_metrics(self, queue: str = ".*", window: str = "5m") -> Dict[str, float]:
        """Get RabbitMQ queue metrics."""
        return {
            "arrival_rate": await self.query(
                QueueingTheoryQueries.rabbitmq_arrival_rate(queue, window)
            ) or 0.0,
            "service_rate": await self.query(
                QueueingTheoryQueries.rabbitmq_service_rate(queue, window)
            ) or 0.0,
            "queue_depth": await self.query(
                QueueingTheoryQueries.rabbitmq_queue_depth(queue)
            ) or 0.0,
        }
    
    async def get_all_queueing_metrics(
        self,
        namespace: str = ".*",
        function: str = ".*",
        window: str = "5m"
    ) -> Dict[str, float]:
        """
        Get all queueing theory metrics in one call.
        
        Returns dict with: arrival_rate, service_rate, utilization, 
        queue_depth, queue_latency_p95, error_rate
        """
        return {
            "arrival_rate": await self.get_arrival_rate(namespace, function, window),
            "service_rate": await self.get_service_rate(namespace, function, window),
            "utilization": await self.get_utilization(namespace, window),
            "queue_depth": await self.get_queue_depth(),
            "queue_latency_p95": await self.get_queue_latency_p95(window),
            "error_rate": await self.get_error_rate(namespace, function, window),
        }
    
    async def close(self):
        """Close HTTP client."""
        if self._client and not self._client.is_closed:
            await self._client.aclose()


# =============================================================================
# Metrics Collector (hybrid: local + Prometheus)
# =============================================================================

@dataclass
class OptimizationMetrics:
    """
    Collector for optimization metrics.
    
    HYBRID approach:
    - Uses EXISTING metrics from Prometheus (knative-lambda, RabbitMQ)
    - Tracks LOCAL metrics for this agent's decisions
    """
    agent_id: str
    
    # Prometheus client for querying existing metrics
    prometheus: PrometheusClient = field(default_factory=PrometheusClient)
    
    # Sliding window for local rate calculation (last N seconds)
    window_size: float = 60.0
    
    # Namespace/function filters for Prometheus queries
    namespace_filter: str = ".*"
    function_filter: str = ".*"
    
    # Internal state for LOCAL tracking (fallback when Prometheus unavailable)
    _arrival_times: list = field(default_factory=list)
    _processing_times: list = field(default_factory=list)
    _current_processing: int = 0
    
    def record_arrival(self, event_type: str, priority: str = "medium"):
        """Record event arrival for LOCAL λ calculation (fallback)."""
        now = time.time()
        self._arrival_times.append(now)
        self._cleanup_old_times()
        
    def record_processing_start(self):
        """Record start of event processing."""
        self._current_processing += 1
        
    def record_processing_end(
        self, 
        event_type: str, 
        status: str,
        processing_time: float,
        wait_time: float = 0.0,
        priority: str = "medium"
    ):
        """Record end of event processing for LOCAL μ calculation."""
        now = time.time()
        self._processing_times.append((now, processing_time))
        self._current_processing = max(0, self._current_processing - 1)
        self._cleanup_old_times()
    
    def record_rejection(self, event_type: str, reason: str):
        """Record rejected event."""
        DECISIONS_MADE.labels(agent_id=self.agent_id, decision="reject").inc()
        
    def record_forward(self, event_type: str, target_agent: str):
        """Record forwarded event."""
        DECISIONS_MADE.labels(agent_id=self.agent_id, decision="forward").inc()
    
    def record_process(self, event_type: str):
        """Record processed event."""
        DECISIONS_MADE.labels(agent_id=self.agent_id, decision="process").inc()
    
    async def get_arrival_rate(self) -> float:
        """
        Get arrival rate (λ) - PREFER Prometheus, fallback to local.
        """
        # Try Prometheus first
        try:
            rate = await self.prometheus.get_arrival_rate(
                self.namespace_filter, 
                self.function_filter
            )
            if rate > 0:
                return rate
        except Exception:
            pass
        
        # Fallback to local calculation
        return self._get_local_arrival_rate()
    
    async def get_service_rate(self) -> float:
        """
        Get service rate (μ) - PREFER Prometheus, fallback to local.
        """
        # Try Prometheus first
        try:
            rate = await self.prometheus.get_service_rate(
                self.namespace_filter,
                self.function_filter
            )
            if rate > 0:
                return rate
        except Exception:
            pass
        
        # Fallback to local calculation
        return self._get_local_service_rate()
    
    async def get_utilization(self) -> float:
        """
        Get utilization (ρ = λ / μ).
        """
        arrival_rate = await self.get_arrival_rate()
        service_rate = await self.get_service_rate()
        
        if service_rate <= 0:
            return float('inf') if arrival_rate > 0 else 0.0
            
        return arrival_rate / service_rate
    
    async def get_all_metrics(self) -> Dict[str, float]:
        """Get all queueing theory metrics."""
        return await self.prometheus.get_all_queueing_metrics(
            self.namespace_filter,
            self.function_filter
        )
    
    def _get_local_arrival_rate(self) -> float:
        """Calculate arrival rate (λ) from LOCAL data."""
        self._cleanup_old_times()
        if len(self._arrival_times) < 2:
            return 0.0
        
        window = min(self.window_size, self._arrival_times[-1] - self._arrival_times[0])
        if window <= 0:
            return 0.0
            
        return len(self._arrival_times) / window
    
    def _get_local_service_rate(self) -> float:
        """Calculate service rate (μ) from LOCAL data."""
        self._cleanup_old_times()
        if not self._processing_times:
            return 1.0  # Default to 1 event/second
        
        total_time = sum(pt for _, pt in self._processing_times)
        avg_processing_time = total_time / len(self._processing_times)
        
        if avg_processing_time <= 0:
            return 1.0
            
        return 1.0 / avg_processing_time
    
    def _cleanup_old_times(self):
        """Remove times outside the sliding window."""
        cutoff = time.time() - self.window_size
        
        self._arrival_times = [t for t in self._arrival_times if t > cutoff]
        self._processing_times = [(t, pt) for t, pt in self._processing_times if t > cutoff]


def setup_optimization_metrics(
    agent_id: str, 
    window_size: float = 60.0,
    prometheus_url: Optional[str] = None,
    namespace_filter: str = ".*",
    function_filter: str = ".*"
) -> OptimizationMetrics:
    """
    Factory function to create metrics collector.
    
    Args:
        agent_id: Unique agent identifier
        window_size: Sliding window for local rate calculation
        prometheus_url: Prometheus URL (default: cluster DNS)
        namespace_filter: Filter for Prometheus queries
        function_filter: Filter for Prometheus queries
    """
    prometheus = PrometheusClient(prometheus_url)
    
    return OptimizationMetrics(
        agent_id=agent_id,
        prometheus=prometheus,
        window_size=window_size,
        namespace_filter=namespace_filter,
        function_filter=function_filter
    )
