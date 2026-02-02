"""
Base test class for Prometheus metrics testing.

Provides utilities for testing metric exposition, values,
and labels across all agents.
"""

import pytest
from typing import Any, Optional
from contextlib import contextmanager


class BaseMetricsTest:
    """
    Base class for testing Prometheus metrics.
    
    Provides utilities for:
    - Asserting metric values
    - Testing metric labels
    - Verifying metric increments/decrements
    - Testing metric cardinality
    
    Usage:
        class TestMyMetrics(BaseMetricsTest):
            def test_counter_increments(self):
                from myagent.metrics import REQUESTS_TOTAL
                
                initial = self.get_metric_value("requests_total")
                REQUESTS_TOTAL.labels(status="success").inc()
                
                self.assert_metric_incremented("requests_total", labels={"status": "success"})
    """
    
    def __init__(self):
        self._snapshots: dict[str, float] = {}
    
    def get_metric_value(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
    ) -> float:
        """Get current value of a metric."""
        try:
            from prometheus_client import REGISTRY
            
            for metric in REGISTRY.collect():
                # Handle both with and without _total suffix
                if metric.name == metric_name or metric.name == metric_name.replace("_total", ""):
                    for sample in metric.samples:
                        if labels is None:
                            return sample.value
                        if all(sample.labels.get(k) == v for k, v in labels.items()):
                            return sample.value
            return 0.0
        except ImportError:
            return 0.0
    
    def snapshot_metric(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
    ):
        """Take a snapshot of a metric's current value."""
        key = f"{metric_name}:{labels}"
        self._snapshots[key] = self.get_metric_value(metric_name, labels)
    
    def snapshot_metrics(
        self,
        metric_names: list[str],
        labels: Optional[dict] = None,
    ):
        """Take snapshots of multiple metrics."""
        for name in metric_names:
            self.snapshot_metric(name, labels)
    
    def assert_metric_value(
        self,
        metric_name: str,
        expected: float,
        labels: Optional[dict] = None,
        delta: float = 0.001,
    ):
        """Assert metric has expected value."""
        actual = self.get_metric_value(metric_name, labels)
        assert abs(actual - expected) < delta, (
            f"Metric {metric_name} expected {expected}, got {actual}"
        )
    
    def assert_metric_incremented(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
        min_increment: float = 1.0,
    ):
        """Assert metric was incremented since last snapshot."""
        key = f"{metric_name}:{labels}"
        current = self.get_metric_value(metric_name, labels)
        previous = self._snapshots.get(key, 0.0)
        increment = current - previous
        
        assert increment >= min_increment, (
            f"Metric {metric_name} expected increment >= {min_increment}, "
            f"got {increment}"
        )
    
    def assert_metric_decremented(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
        min_decrement: float = 1.0,
    ):
        """Assert metric was decremented since last snapshot."""
        key = f"{metric_name}:{labels}"
        current = self.get_metric_value(metric_name, labels)
        previous = self._snapshots.get(key, 0.0)
        decrement = previous - current
        
        assert decrement >= min_decrement, (
            f"Metric {metric_name} expected decrement >= {min_decrement}, "
            f"got {decrement}"
        )
    
    def assert_metric_unchanged(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
    ):
        """Assert metric has not changed since last snapshot."""
        key = f"{metric_name}:{labels}"
        current = self.get_metric_value(metric_name, labels)
        previous = self._snapshots.get(key, 0.0)
        
        assert current == previous, (
            f"Metric {metric_name} changed from {previous} to {current}"
        )
    
    def assert_metric_exists(
        self,
        metric_name: str,
    ):
        """Assert metric is registered."""
        try:
            from prometheus_client import REGISTRY
            
            found = False
            for metric in REGISTRY.collect():
                if metric.name == metric_name or metric.name.startswith(metric_name):
                    found = True
                    break
            
            assert found, f"Metric {metric_name} not found in registry"
        except ImportError:
            pytest.skip("prometheus_client not installed")
    
    def assert_metric_labels(
        self,
        metric_name: str,
        expected_labels: list[str],
    ):
        """Assert metric has expected label names."""
        try:
            from prometheus_client import REGISTRY
            
            for metric in REGISTRY.collect():
                if metric.name == metric_name or metric.name.startswith(metric_name):
                    for sample in metric.samples:
                        actual_labels = set(sample.labels.keys())
                        expected_set = set(expected_labels)
                        assert expected_set.issubset(actual_labels), (
                            f"Metric {metric_name} missing labels: "
                            f"{expected_set - actual_labels}"
                        )
                        return
            
            pytest.fail(f"Metric {metric_name} not found")
        except ImportError:
            pytest.skip("prometheus_client not installed")
    
    def get_histogram_observations(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
    ) -> dict:
        """Get histogram bucket values and sum/count."""
        try:
            from prometheus_client import REGISTRY
            
            result = {"buckets": {}, "sum": 0.0, "count": 0}
            
            for metric in REGISTRY.collect():
                if metric.name == metric_name:
                    for sample in metric.samples:
                        if labels and not all(
                            sample.labels.get(k) == v 
                            for k, v in labels.items() 
                            if k != "le"
                        ):
                            continue
                        
                        if "_bucket" in sample.name:
                            le = sample.labels.get("le")
                            result["buckets"][le] = sample.value
                        elif "_sum" in sample.name:
                            result["sum"] = sample.value
                        elif "_count" in sample.name:
                            result["count"] = int(sample.value)
            
            return result
        except ImportError:
            return {"buckets": {}, "sum": 0.0, "count": 0}
    
    def assert_histogram_observation(
        self,
        metric_name: str,
        expected_count: int,
        labels: Optional[dict] = None,
    ):
        """Assert histogram has expected number of observations."""
        observations = self.get_histogram_observations(metric_name, labels)
        assert observations["count"] == expected_count, (
            f"Histogram {metric_name} expected {expected_count} observations, "
            f"got {observations['count']}"
        )
    
    @contextmanager
    def track_metrics(self, metric_names: list[str], labels: Optional[dict] = None):
        """Context manager to track metric changes."""
        self.snapshot_metrics(metric_names, labels)
        yield
        # After yield, metrics can be asserted using assert_metric_incremented etc.


class BaseAgentMetricsTest(BaseMetricsTest):
    """
    Specialized metrics test class for agent-specific metrics.
    
    Provides common assertions for standard agent metrics patterns.
    """
    
    # Standard metric names used across agents
    COMMON_METRICS = {
        "requests_total": "{agent}_requests_total",
        "request_duration": "{agent}_request_duration_seconds",
        "errors_total": "{agent}_errors_total",
        "active_connections": "{agent}_active_connections",
        "events_processed": "{agent}_events_processed_total",
    }
    
    agent_prefix: str = "agent"  # Override in subclass
    
    def get_agent_metric_name(self, metric_type: str) -> str:
        """Get full metric name for this agent."""
        pattern = self.COMMON_METRICS.get(metric_type, metric_type)
        return pattern.format(agent=self.agent_prefix)
    
    def assert_request_recorded(
        self,
        status: str = "success",
    ):
        """Assert a request was recorded in metrics."""
        metric_name = self.get_agent_metric_name("requests_total")
        self.assert_metric_incremented(metric_name, labels={"status": status})
    
    def assert_error_recorded(
        self,
        error_type: Optional[str] = None,
    ):
        """Assert an error was recorded in metrics."""
        metric_name = self.get_agent_metric_name("errors_total")
        labels = {"type": error_type} if error_type else None
        self.assert_metric_incremented(metric_name, labels=labels)
    
    def assert_event_processed(
        self,
        event_type: Optional[str] = None,
    ):
        """Assert an event was processed."""
        metric_name = self.get_agent_metric_name("events_processed")
        labels = {"type": event_type} if event_type else None
        self.assert_metric_incremented(metric_name, labels=labels)
