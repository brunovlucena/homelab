"""Tests for game theory implementation."""
import pytest
import time
from agent_optimization.gametheory import (
    AgentState,
    TaskBid,
    Task,
    ContractNetProtocol,
    ShapleyCalculator,
    NashEquilibrium,
)


class TestAgentState:
    """Tests for AgentState dataclass."""
    
    def test_create_agent_state(self):
        """Test creating agent state."""
        state = AgentState(
            agent_id="agent-1",
            cpu_capacity=100.0,
            memory_capacity=100.0,
            cpu_used=30.0,
            memory_used=40.0,
        )
        
        assert state.agent_id == "agent-1"
        assert state.cpu_available == 70.0
        assert state.memory_available == 60.0
    
    def test_utilization(self):
        """Test utilization calculation."""
        state = AgentState(
            agent_id="agent-1",
            cpu_capacity=100.0,
            memory_capacity=100.0,
            cpu_used=50.0,
            memory_used=50.0,
        )
        
        assert state.utilization == 0.5
    
    def test_can_process_sufficient_resources(self):
        """Test can_process with sufficient resources."""
        state = AgentState(
            agent_id="agent-1",
            cpu_capacity=100.0,
            memory_capacity=100.0,
            cpu_used=50.0,
            memory_used=50.0,
        )
        
        assert state.can_process(cpu_required=20, memory_required=20) is True
    
    def test_can_process_insufficient_resources(self):
        """Test can_process with insufficient resources."""
        state = AgentState(
            agent_id="agent-1",
            cpu_capacity=100.0,
            memory_capacity=100.0,
            cpu_used=90.0,
            memory_used=90.0,
        )
        
        assert state.can_process(cpu_required=20, memory_required=20) is False
    
    def test_specializations(self):
        """Test agent specializations."""
        state = AgentState(
            agent_id="agent-1",
            specializations=["chat", "nlp", "security"],
        )
        
        assert "chat" in state.specializations
        assert "nlp" in state.specializations


class TestTaskBid:
    """Tests for TaskBid dataclass."""
    
    def test_create_task_bid(self):
        """Test creating a task bid."""
        bid = TaskBid(
            agent_id="agent-1",
            task_id="task-123",
            estimated_time=0.5,
            estimated_cost=1.0,
            confidence=0.9,
            utility=0.8,
        )
        
        assert bid.agent_id == "agent-1"
        assert bid.task_id == "task-123"
        assert bid.utility == 0.8
    
    def test_bid_comparison(self):
        """Test bid comparison (higher utility wins)."""
        bid1 = TaskBid(
            agent_id="agent-1",
            task_id="task-123",
            estimated_time=0.5,
            estimated_cost=1.0,
            confidence=0.9,
            utility=0.8,
        )
        
        bid2 = TaskBid(
            agent_id="agent-2",
            task_id="task-123",
            estimated_time=0.3,
            estimated_cost=0.5,
            confidence=0.95,
            utility=0.9,
        )
        
        # Higher utility should be "less than" (sorted first)
        assert bid2 < bid1


class TestTask:
    """Tests for Task dataclass."""
    
    def test_create_task(self):
        """Test creating a task."""
        task = Task(
            task_id="task-123",
            event_type="io.homelab.chat.message",
            event_data={"message": "Hello"},
            priority=8,
            reward=10.0,
        )
        
        assert task.task_id == "task-123"
        assert task.priority == 8
        assert task.reward == 10.0


class TestContractNetProtocol:
    """Tests for ContractNetProtocol class."""
    
    @pytest.fixture
    def cnp(self):
        """Create a Contract Net Protocol instance."""
        return ContractNetProtocol(bid_timeout=5.0, min_bids=1)
    
    @pytest.fixture
    def agents(self):
        """Create sample agent states."""
        return [
            AgentState(
                agent_id="agent-chat",
                cpu_capacity=100.0,
                cpu_used=30.0,
                memory_capacity=100.0,
                memory_used=40.0,
                specializations=["chat", "nlp"],
                success_rate=0.95,
            ),
            AgentState(
                agent_id="agent-security",
                cpu_capacity=100.0,
                cpu_used=20.0,
                memory_capacity=100.0,
                memory_used=30.0,
                specializations=["security", "exploit"],
                success_rate=0.9,
            ),
            AgentState(
                agent_id="agent-overloaded",
                cpu_capacity=100.0,
                cpu_used=95.0,
                memory_capacity=100.0,
                memory_used=95.0,
                specializations=["chat"],
                success_rate=0.8,
            ),
        ]
    
    def test_register_agent(self, cnp, agents):
        """Test registering agents."""
        for agent in agents:
            cnp.register_agent(agent)
        
        assert len(cnp._agents) == 3
        assert "agent-chat" in cnp._agents
    
    def test_calculate_bid(self, cnp, agents):
        """Test bid calculation."""
        agent = agents[0]  # agent-chat
        task = Task(
            task_id="task-123",
            event_type="io.homelab.chat.message",
            event_data={},
            reward=10.0,
        )
        
        bid = cnp.calculate_bid(agent, task)
        
        assert bid is not None
        assert bid.agent_id == "agent-chat"
        assert bid.specialization_match is True  # "chat" matches
        assert bid.utility > 0
    
    def test_calculate_bid_overloaded_agent(self, cnp, agents):
        """Test bid calculation for overloaded agent."""
        agent = agents[2]  # agent-overloaded
        task = Task(
            task_id="task-123",
            event_type="io.homelab.chat.message",
            event_data={},
            cpu_required=10.0,
            memory_required=10.0,
        )
        
        bid = cnp.calculate_bid(agent, task)
        
        # Overloaded agent can't process (only 5% available)
        assert bid is None
    
    @pytest.mark.asyncio
    async def test_announce_task(self, cnp, agents):
        """Test task announcement and bid collection."""
        for agent in agents:
            cnp.register_agent(agent)
        
        task = Task(
            task_id="task-123",
            event_type="io.homelab.chat.message",
            event_data={},
        )
        
        bids = await cnp.announce_task(task)
        
        # Should have 2 bids (overloaded agent can't bid)
        assert len(bids) == 2
        # Bids should be sorted by utility
        assert bids[0].utility >= bids[1].utility
    
    @pytest.mark.asyncio
    async def test_select_winner(self, cnp, agents):
        """Test winner selection."""
        for agent in agents:
            cnp.register_agent(agent)
        
        task = Task(
            task_id="task-123",
            event_type="io.homelab.chat.message",
            event_data={},
        )
        
        await cnp.announce_task(task)
        winner = cnp.select_winner("task-123")
        
        assert winner is not None
        # Chat agent should win for chat message (specialization match)
        assert winner.agent_id == "agent-chat"
    
    def test_update_agent_state(self, cnp, agents):
        """Test updating agent state."""
        cnp.register_agent(agents[0])
        
        cnp.update_agent_state("agent-chat", cpu_used=50.0)
        
        assert cnp._agents["agent-chat"].cpu_used == 50.0


class TestShapleyCalculator:
    """Tests for ShapleyCalculator class."""
    
    @pytest.fixture
    def calculator(self):
        """Create a Shapley calculator."""
        return ShapleyCalculator()
    
    @pytest.fixture
    def agents(self):
        """Create sample agents for coalition."""
        return [
            AgentState(
                agent_id="agent-1",
                cpu_capacity=100.0,
                cpu_used=20.0,
                memory_capacity=100.0,
                memory_used=20.0,
                success_rate=0.95,
            ),
            AgentState(
                agent_id="agent-2",
                cpu_capacity=100.0,
                cpu_used=40.0,
                memory_capacity=100.0,
                memory_used=40.0,
                success_rate=0.9,
            ),
        ]
    
    def test_single_agent(self, calculator):
        """Test Shapley value for single agent."""
        agents = [
            AgentState(agent_id="agent-1", success_rate=1.0),
        ]
        
        values = calculator.calculate(agents, total_reward=100.0)
        
        assert values["agent-1"] == 100.0
    
    def test_two_agents(self, calculator, agents):
        """Test Shapley value for two agents."""
        values = calculator.calculate(agents, total_reward=100.0)
        
        # Total should equal reward
        total = sum(values.values())
        assert total == pytest.approx(100.0, rel=0.01)
        
        # More capable agent should get more
        # agent-1 has more available capacity
        assert values["agent-1"] > values["agent-2"]
    
    def test_empty_coalition(self, calculator):
        """Test Shapley value for empty coalition."""
        values = calculator.calculate([], total_reward=100.0)
        
        assert values == {}
    
    def test_custom_contribution_function(self, calculator, agents):
        """Test Shapley with custom contribution function."""
        # Custom function: equal contribution
        def equal_contribution(coalition):
            return len(coalition) * 10.0
        
        values = calculator.calculate(
            agents,
            total_reward=100.0,
            contribution_fn=equal_contribution,
        )
        
        # With equal contribution, values should be equal
        assert values["agent-1"] == pytest.approx(50.0, rel=0.01)
        assert values["agent-2"] == pytest.approx(50.0, rel=0.01)


class TestNashEquilibrium:
    """Tests for NashEquilibrium class."""
    
    @pytest.fixture
    def nash(self):
        """Create a Nash equilibrium calculator."""
        return NashEquilibrium()
    
    @pytest.fixture
    def agent(self):
        """Create a sample agent."""
        return AgentState(
            agent_id="agent-1",
            cpu_capacity=100.0,
            cpu_used=30.0,
            memory_capacity=100.0,
            memory_used=30.0,
            success_rate=0.95,
            processing_cost=1.0,
            avg_processing_time=0.1,
        )
    
    @pytest.fixture
    def other_agents(self):
        """Create other agents for forwarding."""
        return [
            AgentState(
                agent_id="agent-2",
                cpu_capacity=100.0,
                cpu_used=50.0,
                memory_capacity=100.0,
                memory_used=50.0,
                success_rate=0.9,
            ),
        ]
    
    def test_best_response_process(self, nash, agent, other_agents):
        """Test best response when processing is optimal."""
        task = Task(
            task_id="task-123",
            event_type="io.homelab.test.event",
            event_data={},
            priority=8,
            reward=10.0,
        )
        
        strategy, utility = nash.calculate_best_response(
            agent, task, other_agents
        )
        
        # Should choose to process (high reward, low cost)
        assert strategy == NashEquilibrium.Strategy.PROCESS
        assert utility > 0
    
    def test_best_response_overloaded(self, nash, other_agents):
        """Test best response when agent is overloaded."""
        overloaded = AgentState(
            agent_id="agent-overloaded",
            cpu_capacity=100.0,
            cpu_used=95.0,
            memory_capacity=100.0,
            memory_used=95.0,
        )
        
        task = Task(
            task_id="task-123",
            event_type="io.homelab.test.event",
            event_data={},
            cpu_required=10.0,
            memory_required=10.0,
            reward=10.0,
        )
        
        strategy, utility = nash.calculate_best_response(
            overloaded, task, other_agents
        )
        
        # Should choose to forward or reject (can't process)
        assert strategy in (
            NashEquilibrium.Strategy.FORWARD,
            NashEquilibrium.Strategy.REJECT,
        )
    
    def test_best_response_no_other_agents(self, nash, agent):
        """Test best response with no other agents to forward to."""
        task = Task(
            task_id="task-123",
            event_type="io.homelab.test.event",
            event_data={},
            reward=10.0,
        )
        
        strategy, utility = nash.calculate_best_response(
            agent, task, []  # No other agents
        )
        
        # Should process or reject (can't forward)
        assert strategy in (
            NashEquilibrium.Strategy.PROCESS,
            NashEquilibrium.Strategy.REJECT,
        )
    
    def test_utility_process(self, nash, agent):
        """Test utility calculation for processing."""
        task = Task(
            task_id="task-123",
            event_type="io.homelab.test.event",
            event_data={},
            priority=5,
            reward=10.0,
        )
        
        utility = nash._utility_process(agent, task)
        
        assert utility > 0  # Should be positive for capable agent
    
    def test_utility_forward(self, nash, agent, other_agents):
        """Test utility calculation for forwarding."""
        task = Task(
            task_id="task-123",
            event_type="io.homelab.test.event",
            event_data={},
            reward=10.0,
        )
        
        utility, target = nash._utility_forward(agent, task, other_agents)
        
        assert target == "agent-2"
        assert utility < nash._utility_process(agent, task)  # Processing should be better
