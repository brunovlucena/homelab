"""
FASE 3: Control Theory Implementation.

PID Controller para auto-scaling:
- Ajusta número de replicas baseado em métricas
- Suaviza variações (evita oscilação)
- Predição de carga futura
"""

import time
import asyncio
from dataclasses import dataclass, field
from typing import Optional, List, Dict, Tuple
from enum import Enum
import structlog

logger = structlog.get_logger()


@dataclass
class ScalingDecision:
    """Decisão de scaling do PID controller."""
    action: str           # "scale_up", "scale_down", "no_change"
    current_replicas: int
    target_replicas: int
    reason: str
    
    # PID components
    error: float          # Current error
    p_term: float         # Proportional term
    i_term: float         # Integral term
    d_term: float         # Derivative term
    
    # Confidence
    confidence: float     # 0-1


class PIDController:
    """
    PID Controller para auto-scaling.
    
    Controla número de replicas baseado em:
    - Target latency (setpoint)
    - Current latency (process variable)
    
    PID equation:
    u(t) = Kp * e(t) + Ki * ∫e(τ)dτ + Kd * de/dt
    
    where e(t) = setpoint - process_variable
    """
    
    def __init__(
        self,
        setpoint: float,
        kp: float = 2.0,      # Proportional gain
        ki: float = 0.1,      # Integral gain
        kd: float = 0.5,      # Derivative gain
        min_output: float = 1,
        max_output: float = 100,
        sample_time: float = 10.0,  # Seconds between updates
        anti_windup: float = 50.0   # Anti-windup limit for integral
    ):
        """
        Args:
            setpoint: Target value (e.g., target latency in seconds)
            kp: Proportional gain (reação ao erro atual)
            ki: Integral gain (corrige erro persistente)
            kd: Derivative gain (suaviza resposta)
            min_output: Minimum controller output
            max_output: Maximum controller output
            sample_time: Minimum time between updates
            anti_windup: Limit for integral term (prevents windup)
        """
        self.setpoint = setpoint
        self.kp = kp
        self.ki = ki
        self.kd = kd
        self.min_output = min_output
        self.max_output = max_output
        self.sample_time = sample_time
        self.anti_windup = anti_windup
        
        # Internal state
        self._last_time: Optional[float] = None
        self._last_error: float = 0.0
        self._integral: float = 0.0
        self._last_output: float = min_output
        
        # History for analysis
        self._history: List[Dict] = []
    
    def reset(self):
        """Reset controller state."""
        self._last_time = None
        self._last_error = 0.0
        self._integral = 0.0
        self._history.clear()
    
    def update(self, current_value: float, current_replicas: int) -> ScalingDecision:
        """
        Update PID controller and get scaling decision.
        
        Args:
            current_value: Current process variable (e.g., current latency)
            current_replicas: Current number of replicas
        
        Returns:
            ScalingDecision with recommended action
        """
        now = time.time()
        
        # Check sample time
        if self._last_time is not None:
            dt = now - self._last_time
            if dt < self.sample_time:
                return ScalingDecision(
                    action="no_change",
                    current_replicas=current_replicas,
                    target_replicas=current_replicas,
                    reason=f"Sample time not elapsed ({dt:.1f}s < {self.sample_time}s)",
                    error=self._last_error,
                    p_term=0,
                    i_term=self._integral,
                    d_term=0,
                    confidence=0.0
                )
        else:
            dt = self.sample_time
        
        # Calculate error (positive = need more capacity)
        error = current_value - self.setpoint
        
        # Proportional term
        p_term = self.kp * error
        
        # Integral term (with anti-windup)
        self._integral += error * dt
        self._integral = max(-self.anti_windup, min(self.anti_windup, self._integral))
        i_term = self.ki * self._integral
        
        # Derivative term
        if dt > 0:
            d_term = self.kd * (error - self._last_error) / dt
        else:
            d_term = 0.0
        
        # Calculate output
        output = p_term + i_term + d_term
        
        # Calculate target replicas
        # Positive output = need more replicas
        # Negative output = can reduce replicas
        target_replicas = current_replicas + int(round(output))
        target_replicas = max(
            int(self.min_output), 
            min(int(self.max_output), target_replicas)
        )
        
        # Determine action
        if target_replicas > current_replicas:
            action = "scale_up"
        elif target_replicas < current_replicas:
            action = "scale_down"
        else:
            action = "no_change"
        
        # Calculate confidence
        # Higher when error is significant and consistent
        confidence = min(1.0, abs(error) / self.setpoint) if self.setpoint > 0 else 0.5
        
        # Build reason
        reason = (
            f"PID: error={error:.3f}, P={p_term:.2f}, I={i_term:.2f}, D={d_term:.2f}, "
            f"output={output:.2f}, target={target_replicas}"
        )
        
        # Update state
        self._last_time = now
        self._last_error = error
        self._last_output = output
        
        # Record history
        self._history.append({
            "timestamp": now,
            "current_value": current_value,
            "setpoint": self.setpoint,
            "error": error,
            "p_term": p_term,
            "i_term": i_term,
            "d_term": d_term,
            "output": output,
            "current_replicas": current_replicas,
            "target_replicas": target_replicas,
        })
        
        # Keep only last 100 entries
        if len(self._history) > 100:
            self._history = self._history[-100:]
        
        decision = ScalingDecision(
            action=action,
            current_replicas=current_replicas,
            target_replicas=target_replicas,
            reason=reason,
            error=error,
            p_term=p_term,
            i_term=i_term,
            d_term=d_term,
            confidence=confidence
        )
        
        logger.debug(
            "pid_update",
            error=error,
            output=output,
            action=action,
            current=current_replicas,
            target=target_replicas
        )
        
        return decision
    
    def tune(
        self,
        kp: Optional[float] = None,
        ki: Optional[float] = None,
        kd: Optional[float] = None,
        setpoint: Optional[float] = None
    ):
        """Adjust PID parameters at runtime."""
        if kp is not None:
            self.kp = kp
        if ki is not None:
            self.ki = ki
        if kd is not None:
            self.kd = kd
        if setpoint is not None:
            self.setpoint = setpoint
            self._integral = 0.0  # Reset integral on setpoint change
    
    def get_history(self) -> List[Dict]:
        """Get controller history for analysis."""
        return self._history.copy()


class AutoScaler:
    """
    Auto-scaler combining multiple signals.
    
    Uses:
    - PID controller for latency
    - Queueing theory for utilization
    - Hysteresis to prevent oscillation
    """
    
    def __init__(
        self,
        target_latency: float = 1.0,
        target_utilization: float = 0.7,
        min_replicas: int = 1,
        max_replicas: int = 20,
        scale_up_cooldown: float = 60.0,    # Seconds
        scale_down_cooldown: float = 300.0,  # Seconds
        hysteresis: float = 0.1              # Prevent oscillation
    ):
        """
        Args:
            target_latency: Target average latency (seconds)
            target_utilization: Target utilization (0-1)
            min_replicas: Minimum number of replicas
            max_replicas: Maximum number of replicas
            scale_up_cooldown: Minimum time between scale ups
            scale_down_cooldown: Minimum time between scale downs
            hysteresis: Tolerance band to prevent oscillation
        """
        self.target_latency = target_latency
        self.target_utilization = target_utilization
        self.min_replicas = min_replicas
        self.max_replicas = max_replicas
        self.scale_up_cooldown = scale_up_cooldown
        self.scale_down_cooldown = scale_down_cooldown
        self.hysteresis = hysteresis
        
        # PID controller for latency
        self.latency_pid = PIDController(
            setpoint=target_latency,
            kp=2.0,
            ki=0.1,
            kd=0.5,
            min_output=min_replicas,
            max_output=max_replicas
        )
        
        # PID controller for utilization
        self.utilization_pid = PIDController(
            setpoint=target_utilization,
            kp=1.0,
            ki=0.05,
            kd=0.3,
            min_output=min_replicas,
            max_output=max_replicas
        )
        
        # Cooldown tracking
        self._last_scale_up: Optional[float] = None
        self._last_scale_down: Optional[float] = None
        self._current_replicas: int = min_replicas
    
    def recommend(
        self,
        current_latency: float,
        current_utilization: float,
        current_replicas: int
    ) -> ScalingDecision:
        """
        Get scaling recommendation based on current metrics.
        
        Combines latency and utilization signals with hysteresis.
        """
        now = time.time()
        self._current_replicas = current_replicas
        
        # Get PID recommendations
        latency_decision = self.latency_pid.update(current_latency, current_replicas)
        utilization_decision = self.utilization_pid.update(
            current_utilization, current_replicas
        )
        
        # Combine recommendations (max for scale up, min for scale down)
        if latency_decision.action == "scale_up" or utilization_decision.action == "scale_up":
            # Check cooldown
            if self._last_scale_up and (now - self._last_scale_up) < self.scale_up_cooldown:
                return ScalingDecision(
                    action="no_change",
                    current_replicas=current_replicas,
                    target_replicas=current_replicas,
                    reason=f"Scale up cooldown ({now - self._last_scale_up:.0f}s < {self.scale_up_cooldown}s)",
                    error=latency_decision.error,
                    p_term=latency_decision.p_term,
                    i_term=latency_decision.i_term,
                    d_term=latency_decision.d_term,
                    confidence=0.0
                )
            
            # Check hysteresis
            latency_over = current_latency > self.target_latency * (1 + self.hysteresis)
            util_over = current_utilization > self.target_utilization * (1 + self.hysteresis)
            
            if not (latency_over or util_over):
                return ScalingDecision(
                    action="no_change",
                    current_replicas=current_replicas,
                    target_replicas=current_replicas,
                    reason=f"Within hysteresis band (lat={current_latency:.2f}, util={current_utilization:.2f})",
                    error=latency_decision.error,
                    p_term=latency_decision.p_term,
                    i_term=latency_decision.i_term,
                    d_term=latency_decision.d_term,
                    confidence=0.5
                )
            
            target = max(
                latency_decision.target_replicas,
                utilization_decision.target_replicas
            )
            target = min(target, self.max_replicas)
            
            self._last_scale_up = now
            
            return ScalingDecision(
                action="scale_up",
                current_replicas=current_replicas,
                target_replicas=target,
                reason=f"Scale up: latency={current_latency:.2f}s (target={self.target_latency}s), "
                       f"util={current_utilization:.2%} (target={self.target_utilization:.0%})",
                error=latency_decision.error,
                p_term=latency_decision.p_term,
                i_term=latency_decision.i_term,
                d_term=latency_decision.d_term,
                confidence=max(latency_decision.confidence, utilization_decision.confidence)
            )
        
        elif (latency_decision.action == "scale_down" and 
              utilization_decision.action == "scale_down"):
            # Check cooldown
            if self._last_scale_down and (now - self._last_scale_down) < self.scale_down_cooldown:
                return ScalingDecision(
                    action="no_change",
                    current_replicas=current_replicas,
                    target_replicas=current_replicas,
                    reason=f"Scale down cooldown ({now - self._last_scale_down:.0f}s < {self.scale_down_cooldown}s)",
                    error=latency_decision.error,
                    p_term=latency_decision.p_term,
                    i_term=latency_decision.i_term,
                    d_term=latency_decision.d_term,
                    confidence=0.0
                )
            
            # Check hysteresis
            latency_under = current_latency < self.target_latency * (1 - self.hysteresis)
            util_under = current_utilization < self.target_utilization * (1 - self.hysteresis)
            
            if not (latency_under and util_under):
                return ScalingDecision(
                    action="no_change",
                    current_replicas=current_replicas,
                    target_replicas=current_replicas,
                    reason=f"Within hysteresis band (lat={current_latency:.2f}, util={current_utilization:.2f})",
                    error=latency_decision.error,
                    p_term=latency_decision.p_term,
                    i_term=latency_decision.i_term,
                    d_term=latency_decision.d_term,
                    confidence=0.5
                )
            
            target = max(
                min(latency_decision.target_replicas, utilization_decision.target_replicas),
                self.min_replicas
            )
            
            self._last_scale_down = now
            
            return ScalingDecision(
                action="scale_down",
                current_replicas=current_replicas,
                target_replicas=target,
                reason=f"Scale down: latency={current_latency:.2f}s, util={current_utilization:.2%}",
                error=latency_decision.error,
                p_term=latency_decision.p_term,
                i_term=latency_decision.i_term,
                d_term=latency_decision.d_term,
                confidence=min(latency_decision.confidence, utilization_decision.confidence)
            )
        
        # No change
        return ScalingDecision(
            action="no_change",
            current_replicas=current_replicas,
            target_replicas=current_replicas,
            reason=f"Stable: latency={current_latency:.2f}s, util={current_utilization:.2%}",
            error=latency_decision.error,
            p_term=latency_decision.p_term,
            i_term=latency_decision.i_term,
            d_term=latency_decision.d_term,
            confidence=0.8
        )
