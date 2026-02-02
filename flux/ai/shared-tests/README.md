# 🧪 Shared Testing Library for Homelab Agents

## Overview

This library provides reusable test fixtures, base classes, and utilities for testing all agents in the homelab infrastructure. By centralizing common testing patterns, we ensure consistency across agents and reduce code duplication.

## Installation

Add to your agent's `tests/requirements.txt`:

```
-e ../../shared-tests
```

Or in your test configuration:

```python
# conftest.py
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "shared-tests"))

from shared_tests.fixtures import *  # Import all fixtures
from shared_tests.base import *      # Import base test classes
```

## Available Fixtures

### CloudEvent Fixtures

```python
@pytest.fixture
def cloudevent_factory():
    """Factory for creating CloudEvents."""
    pass

@pytest.fixture
def sample_cloudevent():
    """A basic CloudEvent for testing."""
    pass
```

### Kubernetes Fixtures

```python
@pytest.fixture
def mock_k8s_client():
    """Mocked Kubernetes client."""
    pass

@pytest.fixture
def mock_k8s_pod():
    """Sample Pod resource."""
    pass
```

### HTTP Client Fixtures

```python
@pytest.fixture
def mock_httpx_client():
    """Mocked async HTTP client."""
    pass

@pytest.fixture
def mock_ollama_response():
    """Mocked Ollama LLM response."""
    pass
```

### Metrics Fixtures

```python
@pytest.fixture
def metrics_registry():
    """Fresh Prometheus registry for testing."""
    pass

@pytest.fixture
def assert_metric_value():
    """Helper to assert metric values."""
    pass
```

## Base Test Classes

### `BaseAgentTest`
Base class for testing any agent handler.

```python
from shared_tests.base import BaseAgentTest

class TestMyAgent(BaseAgentTest):
    def test_handler_processes_event(self):
        event = self.create_cloudevent("io.homelab.test", {"key": "value"})
        result = await self.handler.process(event)
        self.assert_event_processed(result)
```

### `BaseCloudEventTest`
Specialized tests for CloudEvent processing.

### `BaseMetricsTest`
Tests for Prometheus metrics exposition.

### `BaseHealthCheckTest`
Tests for agent health endpoints.

## Test Utilities

### Assertions

```python
from shared_tests.assertions import (
    assert_cloudevent_valid,
    assert_metric_incremented,
    assert_health_check_passes,
    assert_k8s_resource_created,
)
```

### Factories

```python
from shared_tests.factories import (
    CloudEventFactory,
    K8sResourceFactory,
    MetricsFactory,
)
```

## Usage Examples

### Testing a CloudEvent Handler

```python
import pytest
from shared_tests.fixtures import *
from shared_tests.base import BaseCloudEventTest

class TestMyHandler(BaseCloudEventTest):
    
    @pytest.fixture(autouse=True)
    def setup(self, mock_k8s_client, mock_httpx_client):
        from myagent.handler import MyHandler
        self.handler = MyHandler(k8s_client=mock_k8s_client)
    
    @pytest.mark.asyncio
    async def test_handles_valid_event(self, cloudevent_factory):
        event = cloudevent_factory.create(
            type="io.homelab.myagent.action",
            data={"action": "test"}
        )
        
        result = await self.handler.process(event)
        
        assert result.success is True
        self.assert_metrics_incremented("myagent_events_processed_total")
```

### Testing Metrics

```python
from shared_tests.base import BaseMetricsTest

class TestMyMetrics(BaseMetricsTest):
    
    def test_counter_increments(self, metrics_registry):
        from myagent.metrics import EVENTS_PROCESSED
        
        EVENTS_PROCESSED.labels(status="success").inc()
        
        self.assert_metric_value(
            "myagent_events_processed_total",
            expected=1,
            labels={"status": "success"}
        )
```

## Directory Structure

```
shared-tests/
├── README.md
├── pyproject.toml
├── shared_tests/
│   ├── __init__.py
│   ├── fixtures/
│   │   ├── __init__.py
│   │   ├── cloudevents.py
│   │   ├── kubernetes.py
│   │   ├── http.py
│   │   ├── metrics.py
│   │   └── ollama.py
│   ├── base/
│   │   ├── __init__.py
│   │   ├── agent.py
│   │   ├── cloudevent.py
│   │   ├── metrics.py
│   │   └── health.py
│   ├── assertions/
│   │   ├── __init__.py
│   │   ├── cloudevent.py
│   │   ├── metrics.py
│   │   └── kubernetes.py
│   ├── factories/
│   │   ├── __init__.py
│   │   ├── cloudevent.py
│   │   └── kubernetes.py
│   └── utils/
│       ├── __init__.py
│       ├── async_helpers.py
│       └── test_data.py
└── tests/
    └── test_shared_tests.py
```

## Contributing

When adding new fixtures or utilities:

1. Add tests for the new functionality
2. Update this README with usage examples
3. Ensure backward compatibility with existing agent tests
