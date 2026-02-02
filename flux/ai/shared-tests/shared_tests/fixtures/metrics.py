"""
Prometheus metrics fixtures for testing agent metrics.

Provides utilities for testing metrics exposition and validation
across all agents in the homelab infrastructure.
"""

import pytest
from unittest.mock import MagicMock, patch
from typing import Any, Optional
from contextlib import contextmanager


class MetricsTestHelper:
    """Helper class for testing Prometheus metrics."""
    
    def __init__(self):
        self._original_values: dict[str, float] = {}
        self._registry = None
    
    def capture_metric_value(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
    ) -> float:
        """Capture the current value of a metric."""
        try:
            from prometheus_client import REGISTRY
            
            for metric in REGISTRY.collect():
                if metric.name == metric_name or metric.name == metric_name.replace("_total", ""):
                    for sample in metric.samples:
                        if labels is None or all(
                            sample.labels.get(k) == v for k, v in labels.items()
                        ):
                            return sample.value
            return 0.0
        except ImportError:
            return 0.0
    
    def assert_metric_value(
        self,
        metric_name: str,
        expected: float,
        labels: Optional[dict] = None,
        delta: float = 0.001,
    ):
        """Assert a metric has the expected value."""
        actual = self.capture_metric_value(metric_name, labels)
        assert abs(actual - expected) < delta, (
            f"Metric {metric_name} expected {expected}, got {actual}"
        )
    
    def assert_metric_incremented(
        self,
        metric_name: str,
        labels: Optional[dict] = None,
        min_increment: float = 1.0,
    ):
        """Assert a metric was incremented since last capture."""
        key = f"{metric_name}:{labels}"
        current = self.capture_metric_value(metric_name, labels)
        previous = self._original_values.get(key, 0.0)
        increment = current - previous
        
        assert increment >= min_increment, (
            f"Metric {metric_name} expected increment >= {min_increment}, "
            f"got {increment} (previous: {previous}, current: {current})"
        )
    
    def snapshot(self, metric_names: list[str], labels: Optional[dict] = None):
        """Take a snapshot of multiple metrics."""
        for name in metric_names:
            key = f"{name}:{labels}"
            self._original_values[key] = self.capture_metric_value(name, labels)
    
    def get_all_metrics(self) -> dict[str, list[dict]]:
        """Get all metrics from the registry."""
        try:
            from prometheus_client import REGISTRY
            
            result = {}
            for metric in REGISTRY.collect():
                samples = []
                for sample in metric.samples:
                    samples.append({
                        "name": sample.name,
                        "labels": dict(sample.labels),
                        "value": sample.value,
                    })
                if samples:
                    result[metric.name] = samples
            return result
        except ImportError:
            return {}


class MockCounter:
    """Mock Prometheus Counter for testing."""
    
    def __init__(self, name: str, description: str, labelnames: list = None):
        self.name = name
        self.description = description
        self.labelnames = labelnames or []
        self._values: dict[tuple, float] = {}
    
    def labels(self, **kwargs) -> "MockCounter":
        """Return labeled instance."""
        key = tuple(sorted(kwargs.items()))
        if key not in self._values:
            self._values[key] = 0.0
        self._current_labels = key
        return self
    
    def inc(self, amount: float = 1.0):
        """Increment the counter."""
        key = getattr(self, "_current_labels", ())
        self._values[key] = self._values.get(key, 0.0) + amount
    
    def get_value(self, labels: Optional[dict] = None) -> float:
        """Get counter value."""
        key = tuple(sorted((labels or {}).items()))
        return self._values.get(key, 0.0)


class MockGauge:
    """Mock Prometheus Gauge for testing."""
    
    def __init__(self, name: str, description: str, labelnames: list = None):
        self.name = name
        self.description = description
        self.labelnames = labelnames or []
        self._values: dict[tuple, float] = {}
    
    def labels(self, **kwargs) -> "MockGauge":
        key = tuple(sorted(kwargs.items()))
        if key not in self._values:
            self._values[key] = 0.0
        self._current_labels = key
        return self
    
    def set(self, value: float):
        """Set gauge value."""
        key = getattr(self, "_current_labels", ())
        self._values[key] = value
    
    def inc(self, amount: float = 1.0):
        """Increment gauge."""
        key = getattr(self, "_current_labels", ())
        self._values[key] = self._values.get(key, 0.0) + amount
    
    def dec(self, amount: float = 1.0):
        """Decrement gauge."""
        key = getattr(self, "_current_labels", ())
        self._values[key] = self._values.get(key, 0.0) - amount
    
    def get_value(self, labels: Optional[dict] = None) -> float:
        key = tuple(sorted((labels or {}).items()))
        return self._values.get(key, 0.0)


class MockHistogram:
    """Mock Prometheus Histogram for testing."""
    
    def __init__(
        self,
        name: str,
        description: str,
        labelnames: list = None,
        buckets: list = None,
    ):
        self.name = name
        self.description = description
        self.labelnames = labelnames or []
        self.buckets = buckets or [0.1, 0.5, 1.0, 5.0, 10.0]
        self._observations: dict[tuple, list] = {}
    
    def labels(self, **kwargs) -> "MockHistogram":
        key = tuple(sorted(kwargs.items()))
        if key not in self._observations:
            self._observations[key] = []
        self._current_labels = key
        return self
    
    def observe(self, value: float):
        """Record an observation."""
        key = getattr(self, "_current_labels", ())
        if key not in self._observations:
            self._observations[key] = []
        self._observations[key].append(value)
    
    def get_observations(self, labels: Optional[dict] = None) -> list[float]:
        key = tuple(sorted((labels or {}).items()))
        return self._observations.get(key, [])


class MockMetricsRegistry:
    """Mock Prometheus registry for isolated testing."""
    
    def __init__(self):
        self._metrics: dict[str, Any] = {}
    
    def counter(
        self,
        name: str,
        description: str,
        labelnames: list = None,
    ) -> MockCounter:
        """Create a mock counter."""
        counter = MockCounter(name, description, labelnames)
        self._metrics[name] = counter
        return counter
    
    def gauge(
        self,
        name: str,
        description: str,
        labelnames: list = None,
    ) -> MockGauge:
        """Create a mock gauge."""
        gauge = MockGauge(name, description, labelnames)
        self._metrics[name] = gauge
        return gauge
    
    def histogram(
        self,
        name: str,
        description: str,
        labelnames: list = None,
        buckets: list = None,
    ) -> MockHistogram:
        """Create a mock histogram."""
        histogram = MockHistogram(name, description, labelnames, buckets)
        self._metrics[name] = histogram
        return histogram
    
    def get_metric(self, name: str) -> Optional[Any]:
        """Get a metric by name."""
        return self._metrics.get(name)
    
    def reset_all(self):
        """Reset all metrics."""
        self._metrics.clear()


@pytest.fixture
def metrics_registry():
    """Fresh mock metrics registry for testing."""
    return MockMetricsRegistry()


@pytest.fixture
def metrics_helper():
    """Helper for testing Prometheus metrics."""
    return MetricsTestHelper()


@pytest.fixture
def assert_metric_value(metrics_helper):
    """Fixture function for asserting metric values."""
    return metrics_helper.assert_metric_value


@pytest.fixture
def reset_metrics():
    """Reset Prometheus metrics between tests."""
    @contextmanager
    def _reset():
        try:
            from prometheus_client import REGISTRY
            # Note: In real tests, you might need to unregister collectors
            yield
        except ImportError:
            yield
    return _reset


# Sample metric fixtures
@pytest.fixture
def sample_counter(metrics_registry):
    """Sample counter metric."""
    return metrics_registry.counter(
        "test_requests_total",
        "Total test requests",
        labelnames=["status", "method"],
    )


@pytest.fixture
def sample_gauge(metrics_registry):
    """Sample gauge metric."""
    return metrics_registry.gauge(
        "test_active_connections",
        "Active test connections",
        labelnames=["server"],
    )


@pytest.fixture
def sample_histogram(metrics_registry):
    """Sample histogram metric."""
    return metrics_registry.histogram(
        "test_request_duration_seconds",
        "Test request duration",
        labelnames=["endpoint"],
        buckets=[0.01, 0.05, 0.1, 0.5, 1.0, 5.0],
    )
