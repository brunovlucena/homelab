# 🧪 QA-001: End-to-End CloudEvents Testing

**Priority**: P1 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-253/qa-001-end-to-end-cloudevents-testing

**Status:** ✅ Active  
**Priority:** P0  
**Story Points:** 13  
**Sprint:** Sprint 3  
**Team:** QA Engineering  

---

## 📋 Story

**As a** QA Engineer  
**I want** to validate the complete CloudEvents processing pipeline end-to-end  
**So that** I can ensure the system correctly handles real events from broker to service execution

---

## 🎯 Acceptance Criteria

### ✅ AC1: Build Event E2E Testing
- [ ] Send real `network.notifi.lambda.build.start` CloudEvents to broker
- [ ] Verify builder service receives and processes the event
- [ ] Validate Kaniko job is created in Kubernetes
- [ ] Confirm job completes successfully and image is pushed to ECR
- [ ] Verify build completion event is published
- [ ] Assert all observability traces are captured

**Test Cases:**
1. Single build event processing
2. Multiple concurrent build events (3+ parsers)
3. Build event with invalid parser ID (negative test)
4. Build event retry on transient failure

### ✅ AC2: Parser Event E2E Testing
- [ ] Send real `network.notifi.lambda.parser.start` CloudEvents to broker
- [ ] Verify lambda service is created (if first event)
- [ ] Confirm trigger routes events to correct service
- [ ] Validate service scales based on concurrency
- [ ] Assert parser processes the blockchain data correctly
- [ ] Verify metrics are exported to Prometheus

**Test Cases:**
1. First parser event triggers service creation
2. Subsequent events route to existing service
3. Multiple contexts processed concurrently
4. Service autoscaling under load

### ✅ AC3: Service Deletion E2E Testing
- [ ] Send real `network.notifi.lambda.service.delete` CloudEvents to broker
- [ ] Verify service deletion is triggered
- [ ] Confirm all resources are cleaned up (Service, Trigger, ConfigMap, ServiceAccount)
- [ ] Assert namespace is cleaned properly
- [ ] Validate no orphaned resources remain

**Test Cases:**
1. Delete existing service successfully
2. Delete non-existent service (idempotent)
3. Delete service with active connections
4. Multiple simultaneous deletions

### ✅ AC4: Job Lifecycle E2E Testing
- [ ] Send `network.notifi.lambda.job.start` events with different priorities
- [ ] Verify jobs are created asynchronously
- [ ] Confirm priority queue ordering is respected
- [ ] Validate job status updates are tracked
- [ ] Assert job cleanup happens on completion/failure

**Test Cases:**
1. High-priority job processing
2. Low-priority job queueing
3. Job failure and retry logic
4. Job timeout handling

### ✅ AC5: Complete Lifecycle E2E Test
- [ ] Execute full lifecycle: build → deploy → execute → delete
- [ ] Measure end-to-end latency
- [ ] Verify all intermediate states
- [ ] Assert no resource leaks
- [ ] Confirm observability traces connect across all stages

---

## 🔧 Technical Implementation

### Test Framework
```python
# tests/e2e/test_cloudevents_e2e.py

import pytest
import asyncio
import requests
from cloudevents.http import CloudEvent
from kubernetes import client, config

@pytest.mark.e2e
class TestCloudEventsE2E:
    """End-to-end CloudEvents processing tests"""
    
    @pytest.fixture(autouse=True)
    def setup(self):
        """Setup Kubernetes client and broker connection"""
        config.load_kube_config()
        self.k8s_client = client.CoreV1Api()
        self.batch_client = client.BatchV1Api()
        self.broker_url = os.getenv("BROKER_URL", "http://localhost:8081")
        self.env = os.getenv("ENV", "dev")
        
    async def test_build_event_complete_flow(self):
        """AC1: Validate complete build event flow"""
        # Arrange
        event = create_build_event("customer-123", "parser-abc")
        
        # Act
        response = publish_to_broker(event, self.broker_url)
        
        # Assert
        assert response.status_code in [200, 202]
        
        # Wait for job creation
        job_name = await wait_for_job_creation(
            "customer-123", "parser-abc", timeout=60
        )
        assert job_name is not None
        
        # Wait for job completion
        job_status = await wait_for_job_completion(
            job_name, timeout=1800  # 30 minutes
        )
        assert job_status == "Succeeded"
        
        # Verify image in ECR
        image_exists = check_ecr_image("customer-123", "parser-abc")
        assert image_exists
        
    async def test_parser_event_autoscaling(self):
        """AC2: Validate parser event triggers autoscaling"""
        # Arrange
        events = [
            create_parser_event("customer-123", "parser-abc", ctx)
            for ctx in generate_test_contexts(count=20)
        ]
        
        # Act - Send concurrent events
        responses = await asyncio.gather(*[
            publish_to_broker_async(event, self.broker_url)
            for event in events
        ])
        
        # Assert - All events accepted
        assert all(r.status_code in [200, 202] for r in responses)
        
        # Wait for service creation
        service_name = await wait_for_service_creation(
            "customer-123", "parser-abc", timeout=300
        )
        assert service_name is not None
        
        # Verify autoscaling
        await asyncio.sleep(30)  # Allow time for autoscaling
        pod_count = get_pod_count_for_service(service_name)
        assert pod_count > 1, "Service should scale under load"
        
    async def test_service_deletion_cleanup(self):
        """AC3: Validate service deletion cleans up all resources"""
        # Arrange - Create service first
        build_event = create_build_event("customer-456", "parser-xyz")
        await self.test_build_event_complete_flow()  # Reuse
        
        service_name = generate_service_name("customer-456", "parser-xyz")
        
        # Pre-check resources exist
        assert resource_exists("Service", service_name)
        assert resource_exists("Trigger", service_name)
        assert resource_exists("ConfigMap", f"{service_name}-config")
        assert resource_exists("ServiceAccount", service_name)
        
        # Act - Send deletion event
        delete_event = create_delete_event("customer-456", "parser-xyz")
        response = publish_to_broker(delete_event, self.broker_url)
        
        # Assert
        assert response.status_code in [200, 202]
        
        # Wait for cleanup
        await wait_for_resource_deletion(service_name, timeout=120)
        
        # Verify all resources deleted
        assert not resource_exists("Service", service_name)
        assert not resource_exists("Trigger", service_name)
        assert not resource_exists("ConfigMap", f"{service_name}-config")
        assert not resource_exists("ServiceAccount", service_name)
```

### Test Environment Setup
```yaml
# tests/e2e/pytest.ini
[pytest]
markers =
    e2e: End-to-end tests (slow, requires cluster)
    build: Build event tests
    parser: Parser event tests
    delete: Deletion tests
    lifecycle: Full lifecycle tests

# tests/e2e/conftest.py
import pytest
import os
from kubernetes import client, config

@pytest.fixture(scope="session")
def kubernetes_client():
    """Provide Kubernetes client for tests"""
    config.load_kube_config()
    return {
        'core': client.CoreV1Api(),
        'batch': client.BatchV1Api(),
        'apps': client.AppsV1Api(),
        'dynamic': client.DynamicClient()
    }

@pytest.fixture(scope="session")
def broker_url():
    """Get broker URL from environment"""
    env = os.getenv("ENV", "dev")
    # Assume port-forward is active
    return "http://0.0.0.0:8081"

@pytest.fixture(scope="session")
def test_environment():
    """Get test environment configuration"""
    return {
        'env': os.getenv("ENV", "dev"),
        'namespace': f"knative-lambda-{os.getenv('ENV', 'dev')}",
        'rabbitmq_namespace': f"rabbitmq-{os.getenv('ENV', 'dev')}",
    }
```

---

## 📊 Test Execution

### Prerequisites
```bash
# 1. Ensure cluster is accessible
kubectl get nodes

# 2. Port-forward broker
make pf-broker ENV=dev &

# 3. Port-forward RabbitMQ (optional, for debugging)
make pf-rabbitmq-admin ENV=dev &
```

### Run Tests
```bash
# Run all E2E tests
make test-e2e ENV=dev

# Run specific test category
pytest tests/e2e -m build
pytest tests/e2e -m parser
pytest tests/e2e -m delete
pytest tests/e2e -m lifecycle

# Run with verbose output
pytest tests/e2e -v --log-cli-level=INFO

# Run specific test
pytest tests/e2e/test_cloudevents_e2e.py::TestCloudEventsE2E::test_build_event_complete_flow
```

---

## 🎯 Success Metrics | Metric | Target | Measurement | |-------- | -------- | ------------- | | E2E Test Pass Rate | > 95% | Pytest report | | Build Event Latency | < 5 minutes | Time from event to image | | Parser Event Latency | < 30 seconds | Time from event to execution | | Resource Cleanup Time | < 2 minutes | Deletion to all resources gone | | Test Execution Time | < 20 minutes | Full E2E suite runtime | ---

## 🔍 Observability

### Test Metrics
- E2E test duration per scenario
- Event-to-execution latency
- Resource creation/deletion times
- Test failure rates and reasons

### Monitoring Integration
- Tests export metrics to Prometheus
- Failed tests create alerts
- Test traces viewable in Tempo
- Test results tracked in dashboard

---

## 🚀 CI/CD Integration

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on:
  pull_request:
    branches: [develop, main]
  schedule:
    - cron: '0 2 * * *'  # Nightly

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup kubeconfig
        run: | mkdir -p ~/.kube
          echo "${{ secrets.KUBECONFIG }}" > ~/.kube/config
          
      - name: Port-forward broker
        run: | make pf-broker ENV=dev &
          sleep 10
          
      - name: Run E2E tests
        run: make test-e2e ENV=dev
        
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: e2e-test-results
          path: tests/e2e/results/
```

---

## 📚 Related Stories

- **BACKEND-001:** CloudEvents HTTP Processing (provides HTTP endpoint)
- **BACKEND-002:** Build Context Management (provides build logic)
- **BACKEND-003:** Job Lifecycle Management (provides job creation)
- **BACKEND-006:** Knative Service Management (provides service CRUD)
- **QA-002:** Load Testing (validates performance under load)

---

## 🔗 Dependencies

- Kubernetes cluster access (dev/staging)
- RabbitMQ broker running
- ECR repository access
- Knative eventing installed
- Port-forwarding capability

---

## 📝 Notes

- E2E tests are **slow** (5-20 minutes) - run selectively in CI
- Tests create **real resources** - ensure proper cleanup
- Requires **cluster permissions** for resource management
- Port-forwarding must be active during test execution
- Tests are **environment-aware** - use ENV variable

