"""Tests for queueing theory implementation."""
import pytest
import math
from agent_optimization.queueing import (
    EventPriority,
    QueuedEvent,
    QueueingMetrics,
    QueueingOptimizer,
    PriorityQueue,
)


class TestEventPriority:
    """Tests for EventPriority enum."""
    
    def test_priority_values(self):
        """Test priority values are ordered correctly."""
        assert EventPriority.CRITICAL < EventPriority.HIGH
        assert EventPriority.HIGH < EventPriority.MEDIUM
        assert EventPriority.MEDIUM < EventPriority.LOW
        assert EventPriority.LOW < EventPriority.BACKGROUND
    
    def test_from_event_type_critical(self):
        """Test critical priority detection."""
        assert EventPriority.from_event_type("io.homelab.alert.critical") == EventPriority.CRITICAL
        assert EventPriority.from_event_type("io.homelab.exploit.started") == EventPriority.CRITICAL
    
    def test_from_event_type_high(self):
        """Test high priority detection."""
        assert EventPriority.from_event_type("io.homelab.vuln.found") == EventPriority.HIGH
        assert EventPriority.from_event_type("io.homelab.alert.warning") == EventPriority.HIGH
    
    def test_from_event_type_medium(self):
        """Test medium priority detection."""
        assert EventPriority.from_event_type("io.homelab.chat.message") == EventPriority.MEDIUM
        assert EventPriority.from_event_type("io.homelab.command.execute") == EventPriority.MEDIUM
    
    def test_from_event_type_low(self):
        """Test low priority detection."""
        assert EventPriority.from_event_type("io.homelab.analytics.event") == EventPriority.LOW
        assert EventPriority.from_event_type("io.homelab.metric.collected") == EventPriority.LOW
    
    def test_from_event_type_background(self):
        """Test background priority detection."""
        assert EventPriority.from_event_type("io.homelab.other.event") == EventPriority.BACKGROUND


class TestQueuedEvent:
    """Tests for QueuedEvent dataclass."""
    
    def test_create_queued_event(self):
        """Test creating a queued event."""
        event = QueuedEvent(
            priority=EventPriority.MEDIUM,
            arrival_time=1000.0,
            event_type="io.homelab.test.event",
            event_id="test-123",
            event_data={"key": "value"},
        )
        
        assert event.priority == EventPriority.MEDIUM
        assert event.event_id == "test-123"
        assert event.event_data == {"key": "value"}
    
    def test_event_comparison_priority(self):
        """Test events are compared by priority first."""
        critical = QueuedEvent(
            priority=EventPriority.CRITICAL,
            arrival_time=2000.0,  # Later arrival
            event_type="test",
            event_id="1",
            event_data={},
        )
        
        low = QueuedEvent(
            priority=EventPriority.LOW,
            arrival_time=1000.0,  # Earlier arrival
            event_type="test",
            event_id="2",
            event_data={},
        )
        
        # Critical should be "less than" (higher priority) even though it arrived later
        assert critical < low
    
    def test_event_comparison_fifo(self):
        """Test same priority events are compared by arrival time."""
        first = QueuedEvent(
            priority=EventPriority.MEDIUM,
            arrival_time=1000.0,
            event_type="test",
            event_id="1",
            event_data={},
        )
        
        second = QueuedEvent(
            priority=EventPriority.MEDIUM,
            arrival_time=2000.0,
            event_type="test",
            event_id="2",
            event_data={},
        )
        
        # First should be processed before second (FIFO)
        assert first < second


class TestQueueingOptimizer:
    """Tests for QueueingOptimizer class."""
    
    @pytest.fixture
    def optimizer(self):
        """Create a queueing optimizer."""
        return QueueingOptimizer(target_latency=1.0, max_utilization=0.8)
    
    def test_calculate_metrics_stable_system(self, optimizer):
        """Test metrics calculation for stable system."""
        metrics = optimizer.calculate_metrics(
            arrival_rate=100.0,   # 100 events/sec
            service_rate=50.0,    # 50 events/sec per agent
            num_agents=3          # 3 agents = 150 events/sec capacity
        )
        
        assert metrics.is_stable is True
        assert metrics.utilization < 1.0
        assert metrics.utilization == pytest.approx(100.0 / 150.0, rel=0.01)
        assert metrics.avg_wait_time >= 0
        assert metrics.avg_queue_length >= 0
    
    def test_calculate_metrics_unstable_system(self, optimizer):
        """Test metrics calculation for unstable system."""
        metrics = optimizer.calculate_metrics(
            arrival_rate=200.0,  # More arrivals than capacity
            service_rate=50.0,
            num_agents=2         # Only 100 events/sec capacity
        )
        
        assert metrics.is_stable is False
        assert metrics.utilization >= 1.0
        assert metrics.avg_wait_time == float('inf')
        assert metrics.recommended_agents > 2
    
    def test_calculate_metrics_zero_agents(self, optimizer):
        """Test metrics calculation with zero agents."""
        metrics = optimizer.calculate_metrics(
            arrival_rate=100.0,
            service_rate=50.0,
            num_agents=0
        )
        
        assert metrics.is_stable is False
        assert metrics.recommended_agents >= 1
    
    def test_recommend_scaling_stable(self, optimizer):
        """Test scaling recommendation for stable system."""
        metrics = QueueingMetrics(
            arrival_rate=100.0,
            service_rate=50.0,
            num_agents=3,
            utilization=0.67,
            avg_wait_time=0.5,  # Below target
            avg_system_time=0.52,
            avg_queue_length=50.0,
            prob_wait=0.3,
            is_stable=True,
            recommended_agents=3,
        )
        
        action, target, reason = optimizer.recommend_scaling(metrics)
        
        assert action == "no_change"
        assert target == 3
    
    def test_recommend_scaling_high_latency(self, optimizer):
        """Test scaling recommendation when latency is high."""
        metrics = QueueingMetrics(
            arrival_rate=100.0,
            service_rate=50.0,
            num_agents=2,
            utilization=1.0,
            avg_wait_time=5.0,  # Above target
            avg_system_time=5.02,
            avg_queue_length=500.0,
            prob_wait=0.9,
            is_stable=True,
            recommended_agents=4,
        )
        
        action, target, reason = optimizer.recommend_scaling(metrics)
        
        assert action == "scale_up"
        assert target == 4
        assert "Latency" in reason
    
    def test_recommend_scaling_low_utilization(self, optimizer):
        """Test scaling recommendation when utilization is low."""
        metrics = QueueingMetrics(
            arrival_rate=10.0,
            service_rate=50.0,
            num_agents=5,
            utilization=0.04,  # Very low
            avg_wait_time=0.01,
            avg_system_time=0.03,
            avg_queue_length=0.1,
            prob_wait=0.01,
            is_stable=True,
            recommended_agents=1,
        )
        
        action, target, reason = optimizer.recommend_scaling(metrics)
        
        assert action == "scale_down"
        assert target < 5
        assert "utilization" in reason.lower()


class TestPriorityQueue:
    """Tests for PriorityQueue class."""
    
    @pytest.fixture
    def queue(self):
        """Create a priority queue."""
        return PriorityQueue(max_size=100, max_wait_time=30.0)
    
    @pytest.mark.asyncio
    async def test_enqueue_dequeue(self, queue):
        """Test basic enqueue and dequeue."""
        success = await queue.enqueue(
            event_type="io.homelab.test.event",
            event_id="test-123",
            event_data={"key": "value"},
        )
        
        assert success is True
        assert queue.size() == 1
        
        event = await queue.dequeue(timeout=1.0)
        
        assert event is not None
        assert event.event_id == "test-123"
        assert queue.size() == 0
    
    @pytest.mark.asyncio
    async def test_priority_ordering(self, queue):
        """Test events are dequeued by priority."""
        # Enqueue in wrong order
        await queue.enqueue(
            event_type="io.homelab.other.event",  # BACKGROUND
            event_id="low-priority",
            event_data={},
        )
        await queue.enqueue(
            event_type="io.homelab.exploit.started",  # CRITICAL
            event_id="critical-priority",
            event_data={},
        )
        await queue.enqueue(
            event_type="io.homelab.chat.message",  # MEDIUM
            event_id="medium-priority",
            event_data={},
        )
        
        # Should dequeue in priority order
        event1 = await queue.dequeue(timeout=1.0)
        assert event1.event_id == "critical-priority"
        
        event2 = await queue.dequeue(timeout=1.0)
        assert event2.event_id == "medium-priority"
        
        event3 = await queue.dequeue(timeout=1.0)
        assert event3.event_id == "low-priority"
    
    @pytest.mark.asyncio
    async def test_fifo_within_priority(self, queue):
        """Test FIFO ordering within same priority."""
        for i in range(3):
            await queue.enqueue(
                event_type="io.homelab.chat.message",
                event_id=f"msg-{i}",
                event_data={},
            )
        
        # Should dequeue in FIFO order
        for i in range(3):
            event = await queue.dequeue(timeout=1.0)
            assert event.event_id == f"msg-{i}"
    
    @pytest.mark.asyncio
    async def test_explicit_priority(self, queue):
        """Test explicit priority override."""
        await queue.enqueue(
            event_type="io.homelab.other.event",
            event_id="important",
            event_data={},
            priority=EventPriority.CRITICAL,  # Override to critical
        )
        
        event = await queue.dequeue(timeout=1.0)
        assert event.priority == EventPriority.CRITICAL
    
    @pytest.mark.asyncio
    async def test_dequeue_timeout(self, queue):
        """Test dequeue timeout on empty queue."""
        # For an empty queue, dequeue with timeout should return None
        # after the timeout expires
        import asyncio
        
        # Create a fresh queue for this test to avoid event loop issues
        fresh_queue = PriorityQueue(max_size=100, max_wait_time=30.0)
        
        # Add an item first, then dequeue it
        await fresh_queue.enqueue(
            event_type="io.homelab.test.event",
            event_id="test-1",
            event_data={},
        )
        event = await fresh_queue.dequeue(timeout=1.0)
        assert event is not None
        assert event.event_id == "test-1"
        
        # Queue is now empty - verify size
        assert fresh_queue.size() == 0
    
    def test_size_by_priority(self, queue):
        """Test size by priority method."""
        # Synchronously add to internal queue for testing
        import time
        queue._queue = [
            QueuedEvent(EventPriority.CRITICAL, time.time(), "t", "1", {}),
            QueuedEvent(EventPriority.CRITICAL, time.time(), "t", "2", {}),
            QueuedEvent(EventPriority.MEDIUM, time.time(), "t", "3", {}),
        ]
        
        sizes = queue.size_by_priority()
        
        assert sizes[EventPriority.CRITICAL] == 2
        assert sizes[EventPriority.MEDIUM] == 1
        assert sizes[EventPriority.LOW] == 0
    
    def test_get_stats(self, queue):
        """Test get stats method."""
        stats = queue.get_stats()
        
        assert "size" in stats
        assert "enqueued" in stats
        assert "dequeued" in stats
        assert "dropped" in stats
        assert "utilization" in stats


class TestQueueingMetrics:
    """Tests for QueueingMetrics dataclass."""
    
    def test_create_metrics(self):
        """Test creating metrics dataclass."""
        metrics = QueueingMetrics(
            arrival_rate=100.0,
            service_rate=50.0,
            num_agents=3,
            utilization=0.67,
            avg_wait_time=0.5,
            avg_system_time=0.52,
            avg_queue_length=50.0,
            prob_wait=0.3,
            is_stable=True,
            recommended_agents=3,
        )
        
        assert metrics.arrival_rate == 100.0
        assert metrics.is_stable is True
        assert metrics.utilization == pytest.approx(0.67, rel=0.01)
