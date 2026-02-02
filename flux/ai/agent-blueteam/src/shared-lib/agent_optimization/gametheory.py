"""
FASE 2: Game Theory Implementation.

Contract Net Protocol para alocação de tarefas:
- Agents fazem bids para processar eventos
- Shapley Value para distribuição de recompensas
- Nash Equilibrium para decisões estratégicas
"""

import time
import asyncio
import math
from dataclasses import dataclass, field
from typing import Any, Optional, List, Dict, Callable, Awaitable, Tuple
from enum import Enum
import structlog

logger = structlog.get_logger()


@dataclass
class AgentState:
    """
    Estado atual do agent.
    
    Usado para cálculo de bids e decisões Nash.
    """
    agent_id: str
    
    # Recursos
    cpu_capacity: float = 100.0      # Capacidade total CPU (%)
    memory_capacity: float = 100.0   # Capacidade total memória (%)
    cpu_used: float = 0.0           # CPU em uso (%)
    memory_used: float = 0.0        # Memória em uso (%)
    
    # Performance
    avg_processing_time: float = 0.1  # Tempo médio de processamento (s)
    success_rate: float = 0.95        # Taxa de sucesso histórica
    
    # Custos
    processing_cost: float = 1.0      # Custo relativo de processamento
    
    # Especialização (tipos de eventos que o agent processa bem)
    specializations: List[str] = field(default_factory=list)
    
    @property
    def cpu_available(self) -> float:
        """CPU disponível (%)."""
        return max(0, self.cpu_capacity - self.cpu_used)
    
    @property
    def memory_available(self) -> float:
        """Memória disponível (%)."""
        return max(0, self.memory_capacity - self.memory_used)
    
    @property
    def utilization(self) -> float:
        """Utilização média (0-1)."""
        cpu_util = self.cpu_used / self.cpu_capacity if self.cpu_capacity > 0 else 1.0
        mem_util = self.memory_used / self.memory_capacity if self.memory_capacity > 0 else 1.0
        return (cpu_util + mem_util) / 2
    
    def can_process(self, cpu_required: float = 10, memory_required: float = 10) -> bool:
        """Verifica se agent tem recursos para processar."""
        return (self.cpu_available >= cpu_required and 
                self.memory_available >= memory_required)


@dataclass
class TaskBid:
    """
    Bid de um agent para processar um evento/tarefa.
    
    Contract Net Protocol bid.
    """
    agent_id: str
    task_id: str
    
    # Proposta
    estimated_time: float      # Tempo estimado (segundos)
    estimated_cost: float      # Custo estimado
    confidence: float          # Confiança (0-1)
    
    # Utility calculada
    utility: float = 0.0
    
    # Metadata
    timestamp: float = field(default_factory=time.time)
    specialization_match: bool = False
    
    def __lt__(self, other: "TaskBid") -> bool:
        """Compare by utility (higher is better)."""
        return self.utility > other.utility


@dataclass 
class Task:
    """Tarefa a ser alocada via Contract Net Protocol."""
    task_id: str
    event_type: str
    event_data: Any
    
    # Requisitos
    priority: int = 5             # 1-10 (10 = highest)
    deadline: Optional[float] = None  # Absolute deadline
    cpu_required: float = 10.0    # CPU necessário (%)
    memory_required: float = 10.0 # Memória necessária (%)
    
    # Recompensa para agent que processar
    reward: float = 10.0


class ContractNetProtocol:
    """
    Contract Net Protocol para alocação de tarefas.
    
    Flow:
    1. Manager (operador) anuncia tarefa
    2. Contractors (agents) submetem bids
    3. Manager seleciona melhor bid
    4. Contractor executa e reporta
    """
    
    def __init__(
        self,
        bid_timeout: float = 5.0,
        min_bids: int = 1
    ):
        """
        Args:
            bid_timeout: Tempo máximo para aguardar bids (segundos)
            min_bids: Número mínimo de bids antes de decidir
        """
        self.bid_timeout = bid_timeout
        self.min_bids = min_bids
        
        # Registered agents
        self._agents: Dict[str, AgentState] = {}
        
        # Pending tasks waiting for bids
        self._pending_tasks: Dict[str, Task] = {}
        self._task_bids: Dict[str, List[TaskBid]] = {}
        
        # History for learning
        self._bid_history: List[Tuple[TaskBid, bool]] = []  # (bid, won)
    
    def register_agent(self, state: AgentState):
        """Register an agent that can participate in bidding."""
        self._agents[state.agent_id] = state
        logger.info("cnp_agent_registered", agent_id=state.agent_id)
    
    def update_agent_state(self, agent_id: str, **updates):
        """Update agent state."""
        if agent_id in self._agents:
            for key, value in updates.items():
                if hasattr(self._agents[agent_id], key):
                    setattr(self._agents[agent_id], key, value)
    
    def calculate_bid(self, agent_state: AgentState, task: Task) -> Optional[TaskBid]:
        """
        Calculate bid for a task based on agent state.
        
        Bid considera:
        - Capacidade disponível
        - Especialização
        - Custo de processamento
        - Deadline
        """
        # Check if agent can process
        if not agent_state.can_process(task.cpu_required, task.memory_required):
            return None
        
        # Specialization bonus
        specialization_match = any(
            spec in task.event_type 
            for spec in agent_state.specializations
        )
        specialization_bonus = 1.3 if specialization_match else 1.0
        
        # Estimate processing time
        base_time = agent_state.avg_processing_time
        load_factor = 1 + (agent_state.utilization * 0.5)  # More load = slower
        estimated_time = base_time * load_factor / specialization_bonus
        
        # Estimate cost
        estimated_cost = estimated_time * agent_state.processing_cost
        
        # Confidence based on success rate and load
        confidence = agent_state.success_rate * (1 - agent_state.utilization * 0.3)
        
        # Calculate utility (agent's expected value)
        # Utility = Reward * Confidence - Cost - Deadline Penalty
        utility = task.reward * confidence - estimated_cost
        
        if task.deadline:
            time_remaining = task.deadline - time.time()
            if time_remaining < estimated_time:
                utility -= (estimated_time - time_remaining) * 5  # Penalty
        
        # Normalize utility to 0-1
        utility = max(0, min(1, utility / (task.reward + 1)))
        
        return TaskBid(
            agent_id=agent_state.agent_id,
            task_id=task.task_id,
            estimated_time=estimated_time,
            estimated_cost=estimated_cost,
            confidence=confidence,
            utility=utility,
            specialization_match=specialization_match
        )
    
    async def announce_task(self, task: Task) -> List[TaskBid]:
        """
        Announce task and collect bids from all registered agents.
        
        Returns list of bids sorted by utility.
        """
        self._pending_tasks[task.task_id] = task
        self._task_bids[task.task_id] = []
        
        # Collect bids from all agents
        bids = []
        for agent_id, agent_state in self._agents.items():
            bid = self.calculate_bid(agent_state, task)
            if bid:
                bids.append(bid)
                self._task_bids[task.task_id].append(bid)
        
        # Sort by utility (best first)
        bids.sort()
        
        logger.info(
            "cnp_task_announced",
            task_id=task.task_id,
            event_type=task.event_type,
            bids_received=len(bids)
        )
        
        return bids
    
    def select_winner(self, task_id: str) -> Optional[TaskBid]:
        """
        Select winning bid for a task.
        
        Uses highest utility bid.
        """
        bids = self._task_bids.get(task_id, [])
        if not bids:
            return None
        
        # Already sorted by utility
        winner = bids[0]
        
        # Record history
        for bid in bids:
            self._bid_history.append((bid, bid.agent_id == winner.agent_id))
        
        # Cleanup
        del self._pending_tasks[task_id]
        del self._task_bids[task_id]
        
        logger.info(
            "cnp_winner_selected",
            task_id=task_id,
            winner=winner.agent_id,
            utility=winner.utility,
            estimated_time=winner.estimated_time
        )
        
        return winner


class ShapleyCalculator:
    """
    Calcula Shapley Value para distribuição justa de recompensas.
    
    Usado quando múltiplos agents colaboram em uma tarefa.
    """
    
    def __init__(self):
        self._coalition_cache: Dict[str, Dict[str, float]] = {}
    
    def calculate(
        self,
        agents: List[AgentState],
        total_reward: float,
        contribution_fn: Optional[Callable[[List[str]], float]] = None
    ) -> Dict[str, float]:
        """
        Calculate Shapley Value for each agent.
        
        Args:
            agents: List of participating agents
            total_reward: Total reward to distribute
            contribution_fn: Optional custom function to calculate coalition value
                            If None, uses weighted capacity contribution
        
        Returns:
            Dict mapping agent_id to their Shapley value (share of reward)
        """
        n = len(agents)
        if n == 0:
            return {}
        
        if n == 1:
            return {agents[0].agent_id: total_reward}
        
        agent_ids = [a.agent_id for a in agents]
        
        if contribution_fn is None:
            contribution_fn = self._default_contribution(agents)
        
        # Calculate Shapley Value for each agent
        shapley_values = {}
        
        for agent in agents:
            shapley_value = 0.0
            
            # Iterate over all possible coalitions excluding this agent
            for coalition_size in range(n):
                coalitions = self._get_coalitions(
                    [a for a in agent_ids if a != agent.agent_id],
                    coalition_size
                )
                
                for coalition in coalitions:
                    # Value with agent
                    with_agent = contribution_fn(list(coalition) + [agent.agent_id])
                    # Value without agent
                    without_agent = contribution_fn(list(coalition))
                    
                    # Marginal contribution
                    marginal = with_agent - without_agent
                    
                    # Shapley weight
                    weight = (
                        math.factorial(len(coalition)) * 
                        math.factorial(n - len(coalition) - 1) /
                        math.factorial(n)
                    )
                    
                    shapley_value += weight * marginal
            
            shapley_values[agent.agent_id] = shapley_value
        
        # Normalize to total reward
        total_shapley = sum(shapley_values.values())
        if total_shapley > 0:
            for agent_id in shapley_values:
                shapley_values[agent_id] = (
                    shapley_values[agent_id] / total_shapley * total_reward
                )
        
        return shapley_values
    
    def _default_contribution(
        self, 
        agents: List[AgentState]
    ) -> Callable[[List[str]], float]:
        """Create default contribution function based on agent capacity."""
        agent_map = {a.agent_id: a for a in agents}
        
        def contribution(coalition: List[str]) -> float:
            if not coalition:
                return 0.0
            
            total = 0.0
            for agent_id in coalition:
                if agent_id in agent_map:
                    agent = agent_map[agent_id]
                    # Contribution = available capacity * success rate
                    total += (
                        (agent.cpu_available + agent.memory_available) / 2 *
                        agent.success_rate
                    )
            return total
        
        return contribution
    
    def _get_coalitions(self, items: List[str], size: int) -> List[Tuple[str, ...]]:
        """Generate all coalitions of given size."""
        if size == 0:
            return [()]
        if size > len(items):
            return []
        
        coalitions = []
        for i, item in enumerate(items):
            for rest in self._get_coalitions(items[i+1:], size - 1):
                coalitions.append((item,) + rest)
        
        return coalitions


class NashEquilibrium:
    """
    Nash Equilibrium calculator for agent decisions.
    
    Determina a estratégia ótima para um agent considerando
    as estratégias dos outros agents.
    """
    
    class Strategy(Enum):
        PROCESS = "process"    # Processar o evento
        FORWARD = "forward"    # Encaminhar para outro agent
        REJECT = "reject"      # Rejeitar o evento
    
    def __init__(self):
        # Payoff matrix components
        self._payoff_cache: Dict[str, float] = {}
    
    def calculate_best_response(
        self,
        agent: AgentState,
        task: Task,
        other_agents: List[AgentState]
    ) -> Tuple[Strategy, float]:
        """
        Calculate best response strategy using Nash Equilibrium.
        
        Returns:
            Tuple of (best_strategy, expected_utility)
        """
        # Calculate utility for each strategy
        utilities = {}
        
        # PROCESS utility
        utilities[self.Strategy.PROCESS] = self._utility_process(agent, task)
        
        # FORWARD utility (best other agent to forward to)
        forward_utility, best_target = self._utility_forward(
            agent, task, other_agents
        )
        utilities[self.Strategy.FORWARD] = forward_utility
        
        # REJECT utility (always 0, but avoids negative utility)
        utilities[self.Strategy.REJECT] = 0.0
        
        # Find Nash equilibrium (best response)
        best_strategy = max(utilities, key=utilities.get)
        
        return best_strategy, utilities[best_strategy]
    
    def _utility_process(self, agent: AgentState, task: Task) -> float:
        """Calculate utility of processing the task."""
        if not agent.can_process(task.cpu_required, task.memory_required):
            return float('-inf')
        
        # Reward
        reward = task.reward * task.priority / 10  # Normalize priority
        
        # Cost
        cost = agent.processing_cost * agent.avg_processing_time
        
        # Success probability
        success_prob = agent.success_rate * (1 - agent.utilization * 0.2)
        
        # Deadline risk
        deadline_penalty = 0.0
        if task.deadline:
            time_remaining = task.deadline - time.time()
            if time_remaining < agent.avg_processing_time:
                deadline_penalty = (agent.avg_processing_time - time_remaining) * 2
        
        # Expected utility
        utility = (reward * success_prob) - cost - deadline_penalty
        
        return utility
    
    def _utility_forward(
        self,
        agent: AgentState,
        task: Task,
        other_agents: List[AgentState]
    ) -> Tuple[float, Optional[str]]:
        """Calculate utility of forwarding to another agent."""
        if not other_agents:
            return float('-inf'), None
        
        best_utility = float('-inf')
        best_target = None
        
        # Forwarding cost (small penalty)
        forward_cost = 0.5
        
        for other in other_agents:
            if not other.can_process(task.cpu_required, task.memory_required):
                continue
            
            # Estimate their success
            other_success = other.success_rate * (1 - other.utilization * 0.2)
            
            # Our share of reward for forwarding (smaller)
            our_share = task.reward * 0.1  # 10% finder's fee
            
            utility = (our_share * other_success) - forward_cost
            
            if utility > best_utility:
                best_utility = utility
                best_target = other.agent_id
        
        return best_utility, best_target
