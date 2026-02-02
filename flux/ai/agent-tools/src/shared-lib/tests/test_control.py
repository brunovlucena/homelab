"""Tests for control theory implementation."""
import pytest
import time
from agent_optimization.control import (
    PIDController,
    AutoScaler,
    ScalingDecision,
)


class TestPIDController:
    """Tests for PIDController class."""
    
    @pytest.fixture
    def controller(self):
        """Create a PID controller."""
        return PIDController(
            setpoint=1.0,
            kp=2.0,
            ki=0.1,
            kd=0.5,
            min_output=1,
            max_output=10,
            sample_time=0.0  # Disable sample time for testing
        )
    
    def test_create_controller(self, controller):
        """Test creating a PID controller."""
        assert controller.kp == 2.0
        assert controller.ki == 0.1
        assert controller.kd == 0.5
        assert controller.setpoint == 1.0
    
    def test_update_high_latency(self, controller):
        """Test PID decision for high latency (above setpoint)."""
        # Current latency 2.0 > setpoint 1.0 -> Error positive -> scale up
        decision = controller.update(current_value=2.0, current_replicas=3)
        
        assert decision.error > 0
        assert decision.action == "scale_up"
        assert decision.target_replicas >= 3
    
    def test_update_low_latency(self, controller):
        """Test PID decision for low latency (below setpoint)."""
        # Current latency 0.5 < setpoint 1.0 -> Error negative -> scale down
        decision = controller.update(current_value=0.5, current_replicas=5)
        
        assert decision.error < 0
        assert decision.action == "scale_down"
        assert decision.target_replicas <= 5
    
    def test_update_at_target(self, controller):
        """Test PID decision when at target."""
        controller.reset()
        
        # Current latency == setpoint -> Error ~0 -> no change
        decision = controller.update(current_value=1.0, current_replicas=3)
        
        assert abs(decision.error) < 0.01
        # Might be no_change or small adjustment
        assert decision.target_replicas in [2, 3, 4]
    
    def test_integral_accumulation(self, controller):
        """Test integral term accumulates over time."""
        controller.reset()
        
        # Apply same error multiple times
        decisions = []
        for _ in range(5):
            decision = controller.update(current_value=2.0, current_replicas=3)
            decisions.append(decision)
            time.sleep(0.01)  # Small delay
        
        # Integral should build up (check internal state)
        assert controller._integral > 0
    
    def test_reset(self, controller):
        """Test controller reset."""
        # Build up some integral
        for _ in range(5):
            controller.update(current_value=2.0, current_replicas=3)
            time.sleep(0.01)
        
        controller.reset()
        
        # After reset, integral should be zero
        assert controller._integral == 0.0
        assert controller._last_error == 0.0
    
    def test_tune(self, controller):
        """Test tuning PID parameters."""
        controller.tune(kp=3.0, ki=0.2, setpoint=2.0)
        
        assert controller.kp == 3.0
        assert controller.ki == 0.2
        assert controller.setpoint == 2.0
    
    def test_get_history(self, controller):
        """Test getting history."""
        controller.reset()
        controller.update(current_value=2.0, current_replicas=3)
        controller.update(current_value=1.5, current_replicas=4)
        
        history = controller.get_history()
        
        assert len(history) == 2
        assert "error" in history[0]
        assert "p_term" in history[0]


class TestAutoScaler:
    """Tests for AutoScaler class."""
    
    @pytest.fixture
    def scaler(self):
        """Create an auto scaler."""
        return AutoScaler(
            target_latency=1.0,
            target_utilization=0.7,
            min_replicas=1,
            max_replicas=10,
            scale_up_cooldown=60.0,
            scale_down_cooldown=300.0,
        )
    
    def test_create_scaler(self, scaler):
        """Test creating an auto scaler."""
        assert scaler.target_latency == 1.0
        assert scaler.target_utilization == 0.7
        assert scaler.min_replicas == 1
        assert scaler.max_replicas == 10
    
    def test_recommend_no_change(self, scaler):
        """Test recommendation when metrics are within targets."""
        decision = scaler.recommend(
            current_latency=0.5,     # Below target
            current_utilization=0.6,  # Below target
            current_replicas=3,
        )
        
        assert decision.action == "no_change"
        assert decision.target_replicas == 3
    
    def test_recommend_scale_up_high_latency(self, scaler):
        """Test recommendation to scale up when latency is high."""
        decision = scaler.recommend(
            current_latency=2.0,     # Above target
            current_utilization=0.8,
            current_replicas=3,
        )
        
        assert decision.action == "scale_up"
        assert decision.target_replicas > 3
    
    def test_recommend_scale_up_high_utilization(self, scaler):
        """Test recommendation to scale up when utilization is high."""
        # Both latency AND utilization need to trigger scale up
        # due to hysteresis band checking
        decision = scaler.recommend(
            current_latency=1.5,       # Above target (triggers scale up)
            current_utilization=0.95,  # Above target
            current_replicas=3,
        )
        
        assert decision.action == "scale_up"
        assert decision.target_replicas > 3
    
    def test_recommend_scale_down_low_utilization(self, scaler):
        """Test recommendation to scale down when utilization is low."""
        # Force bypass cooldown for testing
        scaler._last_scale_down = 0
        
        decision = scaler.recommend(
            current_latency=0.1,
            current_utilization=0.2,  # Very low
            current_replicas=5,
        )
        
        assert decision.action == "scale_down"
        assert decision.target_replicas < 5
    
    def test_respect_min_replicas(self, scaler):
        """Test that scaling respects minimum replicas."""
        scaler._last_scale_down = 0
        
        decision = scaler.recommend(
            current_latency=0.1,
            current_utilization=0.1,
            current_replicas=1,  # Already at minimum
        )
        
        assert decision.target_replicas >= 1
    
    def test_respect_max_replicas(self, scaler):
        """Test that scaling respects maximum replicas."""
        decision = scaler.recommend(
            current_latency=10.0,
            current_utilization=0.99,
            current_replicas=10,  # Already at maximum
        )
        
        assert decision.target_replicas <= 10
    
    def test_cooldown_prevents_scaling(self, scaler):
        """Test that cooldown prevents rapid scaling."""
        import time
        
        # Simulate recent scale up
        scaler._last_scale_up = time.time()
        
        decision = scaler.recommend(
            current_latency=2.0,     # Would normally trigger scale up
            current_utilization=0.9,
            current_replicas=3,
        )
        
        # Should not scale due to cooldown
        assert decision.action in ("no_change", "cooldown")
    
    def test_get_recommendation_reason(self, scaler):
        """Test that recommendations include a reason."""
        decision = scaler.recommend(
            current_latency=2.0,
            current_utilization=0.9,
            current_replicas=3,
        )
        
        assert decision.reason is not None
        assert len(decision.reason) > 0


class TestScalingDecision:
    """Tests for ScalingDecision dataclass."""
    
    def test_create_decision(self):
        """Test creating a scaling decision."""
        decision = ScalingDecision(
            action="scale_up",
            current_replicas=3,
            target_replicas=5,
            reason="High latency detected",
            error=0.5,
            p_term=1.0,
            i_term=0.1,
            d_term=0.05,
            confidence=0.9,
        )
        
        assert decision.action == "scale_up"
        assert decision.target_replicas == 5
        assert decision.reason == "High latency detected"
        assert decision.current_replicas == 3
        assert decision.confidence == 0.9
    
    def test_decision_actions(self):
        """Test valid decision actions."""
        valid_actions = ["scale_up", "scale_down", "no_change"]
        
        for action in valid_actions:
            decision = ScalingDecision(
                action=action,
                current_replicas=3,
                target_replicas=3,
                reason="Test",
                error=0.0,
                p_term=0.0,
                i_term=0.0,
                d_term=0.0,
                confidence=0.5,
            )
            assert decision.action == action
