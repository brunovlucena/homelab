"""
Prometheus metrics assertions for testing.

Provides semantic assertions for validating Prometheus metrics
in agent tests.
"""

from typing import Optional


# Global snapshot storage for metric tracking
_metric_snapshots: dict[str, float] = {}


def _get_metric_value(
    metric_name: str,
    labels: Optional[dict] = None,
) -> float:
    """Get current value of a Prometheus metric."""
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
    metric_name: str,
    labels: Optional[dict] = None,
):
    """
    Take a snapshot of a metric's current value for later comparison.
    
    Args:
        metric_name: Name of the metric
        labels: Optional labels to filter by
    """
    key = f"{metric_name}:{labels}"
    _metric_snapshots[key] = _get_metric_value(metric_name, labels)


def assert_metric_value(
    metric_name: str,
    expected: float,
    labels: Optional[dict] = None,
    delta: float = 0.001,
):
    """
    Assert a metric has the expected value.
    
    Args:
        metric_name: Name of the metric
        expected: Expected value
        labels: Optional labels to filter by
        delta: Allowed difference for float comparison
    
    Raises:
        AssertionError: If metric value doesn't match expected
    """
    actual = _get_metric_value(metric_name, labels)
    
    assert abs(actual - expected) < delta, (
        f"Metric '{metric_name}' value mismatch: "
        f"expected {expected}, got {actual}"
    )


def assert_metric_incremented(
    metric_name: str,
    labels: Optional[dict] = None,
    min_increment: float = 1.0,
    from_snapshot: bool = True,
):
    """
    Assert a metric was incremented.
    
    Args:
        metric_name: Name of the metric
        labels: Optional labels to filter by
        min_increment: Minimum expected increment
        from_snapshot: If True, compare against snapshot; else assume 0
    
    Raises:
        AssertionError: If metric wasn't incremented enough
    """
    key = f"{metric_name}:{labels}"
    current = _get_metric_value(metric_name, labels)
    previous = _metric_snapshots.get(key, 0.0) if from_snapshot else 0.0
    increment = current - previous
    
    assert increment >= min_increment, (
        f"Metric '{metric_name}' not incremented enough: "
        f"expected >= {min_increment}, got {increment} "
        f"(previous: {previous}, current: {current})"
    )


def assert_metric_decremented(
    metric_name: str,
    labels: Optional[dict] = None,
    min_decrement: float = 1.0,
):
    """
    Assert a metric was decremented.
    
    Args:
        metric_name: Name of the metric
        labels: Optional labels to filter by
        min_decrement: Minimum expected decrement
    
    Raises:
        AssertionError: If metric wasn't decremented enough
    """
    key = f"{metric_name}:{labels}"
    current = _get_metric_value(metric_name, labels)
    previous = _metric_snapshots.get(key, 0.0)
    decrement = previous - current
    
    assert decrement >= min_decrement, (
        f"Metric '{metric_name}' not decremented enough: "
        f"expected >= {min_decrement}, got {decrement}"
    )


def assert_metric_unchanged(
    metric_name: str,
    labels: Optional[dict] = None,
):
    """
    Assert a metric has not changed since snapshot.
    
    Args:
        metric_name: Name of the metric
        labels: Optional labels to filter by
    
    Raises:
        AssertionError: If metric value changed
    """
    key = f"{metric_name}:{labels}"
    current = _get_metric_value(metric_name, labels)
    previous = _metric_snapshots.get(key, 0.0)
    
    assert current == previous, (
        f"Metric '{metric_name}' changed unexpectedly: "
        f"previous {previous}, current {current}"
    )


def assert_metric_exists(
    metric_name: str,
):
    """
    Assert a metric is registered in the Prometheus registry.
    
    Args:
        metric_name: Name of the metric
    
    Raises:
        AssertionError: If metric is not registered
    """
    try:
        from prometheus_client import REGISTRY
        
        found = False
        for metric in REGISTRY.collect():
            if metric.name == metric_name or metric.name.startswith(metric_name):
                found = True
                break
        
        assert found, f"Metric '{metric_name}' not found in registry"
    except ImportError:
        pass  # Skip if prometheus_client not installed


def assert_metric_labels(
    metric_name: str,
    expected_labels: list[str],
):
    """
    Assert a metric has the expected label names.
    
    Args:
        metric_name: Name of the metric
        expected_labels: List of expected label names
    
    Raises:
        AssertionError: If labels don't match
    """
    try:
        from prometheus_client import REGISTRY
        
        for metric in REGISTRY.collect():
            if metric.name == metric_name or metric.name.startswith(metric_name):
                for sample in metric.samples:
                    actual_labels = set(sample.labels.keys())
                    expected_set = set(expected_labels)
                    
                    assert expected_set.issubset(actual_labels), (
                        f"Metric '{metric_name}' missing labels: "
                        f"{expected_set - actual_labels}"
                    )
                    return
        
        raise AssertionError(f"Metric '{metric_name}' not found")
    except ImportError:
        pass


def assert_histogram_observed(
    metric_name: str,
    expected_count: int,
    labels: Optional[dict] = None,
):
    """
    Assert a histogram has the expected number of observations.
    
    Args:
        metric_name: Name of the histogram metric
        expected_count: Expected observation count
        labels: Optional labels to filter by
    
    Raises:
        AssertionError: If observation count doesn't match
    """
    try:
        from prometheus_client import REGISTRY
        
        for metric in REGISTRY.collect():
            if metric.name == metric_name:
                for sample in metric.samples:
                    if "_count" in sample.name:
                        if labels is None or all(
                            sample.labels.get(k) == v 
                            for k, v in labels.items()
                        ):
                            actual = int(sample.value)
                            assert actual == expected_count, (
                                f"Histogram '{metric_name}' observation count: "
                                f"expected {expected_count}, got {actual}"
                            )
                            return
        
        raise AssertionError(f"Histogram '{metric_name}' not found")
    except ImportError:
        pass


def clear_metric_snapshots():
    """Clear all stored metric snapshots."""
    _metric_snapshots.clear()
