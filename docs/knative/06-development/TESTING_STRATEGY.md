# 🧪 Testing Strategy

**Comprehensive testing approach for Knative Lambda**

---

## 📋 Testing Pyramid

```
                   ┌─────────────────┐
                   │   E2E Tests     │  ← 10% (Slow, Expensive)
                   │   (pytest)      │
                   ├─────────────────┤
                   │                 │
               ┌───┴──────────────┴───┐
               │  Integration Tests   │  ← 30% (Medium)
               │  (Go test + K8s)     │
           ┌───┴──────────────────┴───┐
           │                           │
       ┌───┴────────────────────────┴───┐
       │       Unit Tests               │  ← 60% (Fast, Cheap)
       │       (Go test)                │
       └────────────────────────────────┘
```

---

## 🎯 Test Coverage Goals

| Layer | Target Coverage | Current | Status |
|-------|----------------|---------|--------|
| **Unit Tests** | 80% | 75% | 🟡 In Progress |
| **Integration Tests** | 70% | 65% | 🟡 In Progress |
| **E2E Tests** | Critical paths | 90% | ✅ Good |
| **Load Tests** | Key scenarios | 100% | ✅ Complete |

---

## 🔬 Unit Tests

### Scope

Test individual functions and methods in isolation.

**What to test**:
- ✅ Event parsing and validation
- ✅ S3 path construction
- ✅ Rate limiting logic
- ✅ Error handling
- ✅ Metrics emission

**What NOT to test**:
- ❌ External dependencies (S3, K8s API)
- ❌ RabbitMQ connection
- ❌ End-to-end workflows

### Example: CloudEvent Validation

```go
// internal/handler/build_handler_test.go
package handler

import (
    "testing"
    cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestValidateBuildEvent(t *testing.T) {
    tests := []struct {
        name    string
        event   cloudevents.Event
        wantErr bool
    }{
        {
            name: "valid build event",
            event: cloudevents.Event{
                Context: &cloudevents.EventContextV1{
                    Type:   "build.start",
                    Source: "test",
                    ID:     "123",
                },
                Data: BuildEventData{
                    ParserID: "parser-123",
                    S3Bucket: "my-bucket",
                },
            },
            wantErr: false,
        },
        {
            name: "missing parser ID",
            event: cloudevents.Event{
                Context: &cloudevents.EventContextV1{
                    Type:   "build.start",
                    Source: "test",
                    ID:     "123",
                },
                Data: BuildEventData{
                    S3Bucket: "my-bucket",
                },
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateBuildEvent(tt.event)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateBuildEvent() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Unit Tests

```bash
# Run all unit tests
make test-unit

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/handler/... -v

# Run with race detector
go test -race ./...
```

---

## 🔗 Integration Tests

### Scope

Test interactions between components (with real dependencies).

**What to test**:
- ✅ S3 file upload/download
- ✅ Kubernetes Job creation
- ✅ Knative Service creation
- ✅ RabbitMQ pub/sub
- ✅ Metrics collection

### Example: Kubernetes Job Creation

```go
// internal/handler/build_handler_integration_test.go
// +build integration

package handler

import (
    "context"
    "testing"
    batchv1 "k8s.io/api/batch/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func TestCreateKanikoJob_Integration(t *testing.T) {
    // Skip if not in integration test mode
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup
    ctx := context.Background()
    clientset := getK8sClient(t)  // Uses kubeconfig
    namespace := "knative-lambda"

    handler := &BuildHandler{
        K8sClient: clientset,
        Namespace: namespace,
    }

    // Test
    job, err := handler.CreateKanikoJob(ctx, BuildJobConfig{
        ParserID: "test-parser-123",
        S3Bucket: "test-bucket",
        S3Prefix: "global/parser/test-parser-123/",
    })

    if err != nil {
        t.Fatalf("CreateKanikoJob() failed: %v", err)
    }

    // Verify job created in Kubernetes
    _, err = clientset.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
    if err != nil {
        t.Errorf("Job not found in Kubernetes: %v", err)
    }

    // Cleanup
    defer clientset.BatchV1().Jobs(namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
}
```

### Running Integration Tests

```bash
# Run integration tests (requires kubeconfig)
make test-integration

# Run with specific namespace
TEST_NAMESPACE=knative-lambda make test-integration

# Run against dev cluster
export KUBECONFIG=~/.kube/config-dev
make test-integration
```

---

## 🌐 End-to-End (E2E) Tests

### Scope

Test complete user workflows from start to finish.

**Scenarios**:
- ✅ BACKEND-001: Full build pipeline (S3 → Kaniko → ECR → Knative)
- ✅ QA-001: CloudEvent processing with multiple event types
- ✅ QA-002: Load testing and auto-scaling validation

### Example: Full Build Pipeline

```python
# tests/e2e/test_build_pipeline.py
import pytest
import boto3
import requests
from kubernetes import client, config
import time
import uuid

@pytest.fixture(scope="module")
def setup():
    """Setup test environment"""
    config.load_kube_config()
    return {
        "s3": boto3.client("s3"),
        "k8s_batch": client.BatchV1Api(),
        "k8s_serving": client.CustomObjectsApi(),
        "namespace": "knative-lambda",
        "bucket": "knative-lambda-fusion-modules-tmp",
    }

def test_full_build_pipeline(setup):
    """
    Test: Upload code → Trigger build → Verify deployment → Test function
    """
    parser_id = f"e2e-test-{uuid.uuid4().hex[:8]}"
    
    # Step 1: Upload code to S3
    code = """
def handler(event):
    return {'status': 'success', 'test': 'e2e'}
"""
    setup["s3"].put_object(
        Bucket=setup["bucket"],
        Key=f"global/parser/{parser_id}/parser.py",
        Body=code
    )
    
    # Step 2: Trigger build via CloudEvent
    response = requests.post(
        "http://localhost:8080/build",  # Port-forwarded builder service
        headers={
            "ce-type": "build.start",
            "ce-source": "e2e-test",
            "ce-id": str(uuid.uuid4()),
        },
        json={
            "parser_id": parser_id,
            "s3_bucket": setup["bucket"],
            "s3_prefix": f"global/parser/{parser_id}/",
            "language": "python",
        }
    )
    assert response.status_code == 200
    
    # Step 3: Wait for Kaniko job to complete
    timeout = 300  # 5 minutes
    start_time = time.time()
    
    while time.time() - start_time < timeout:
        jobs = setup["k8s_batch"].list_namespaced_job(
            namespace=setup["namespace"],
            label_selector=f"parser-id={parser_id}"
        )
        
        if jobs.items and jobs.items[0].status.succeeded:
            break
        
        time.sleep(10)
    else:
        pytest.fail("Build job did not complete in time")
    
    # Step 4: Verify Knative Service created
    time.sleep(30)  # Wait for service deployment
    
    service = setup["k8s_serving"].get_namespaced_custom_object(
        group="serving.knative.dev",
        version="v1",
        namespace=setup["namespace"],
        plural="services",
        name=parser_id
    )
    
    assert service["status"]["conditions"][0]["status"] == "True"
    service_url = service["status"]["url"]
    
    # Step 5: Test function execution
    response = requests.post(
        service_url,
        headers={
            "ce-type": "test.event",
            "ce-source": "e2e-test",
            "ce-id": str(uuid.uuid4()),
        },
        json={"test": "data"}
    )
    
    assert response.status_code == 200
    assert response.json()["status"] == "success"
    
    # Cleanup
    setup["k8s_serving"].delete_namespaced_custom_object(
        group="serving.knative.dev",
        version="v1",
        namespace=setup["namespace"],
        plural="services",
        name=parser_id
    )
```

### Running E2E Tests

```bash
# Run E2E tests
make test-e2e

# Run specific test
cd tests/e2e
pytest test_build_pipeline.py::test_full_build_pipeline -v

# Run with retries (flaky tests)
pytest test_build_pipeline.py --reruns 3 --reruns-delay 10
```

---

## 📊 Load Tests

### Scope

Validate performance and scaling under load.

**Scenarios**:
- ✅ Concurrent builds (10, 50, 100)
- ✅ Function invocations (1k, 10k, 100k req/s)
- ✅ Scale-to-zero behavior
- ✅ Auto-scaling (0→10→0 pods)

### Example: k6 Load Test

```javascript
// tests/load/build-load.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '1m', target: 10 },   // Ramp up to 10 builds/s
    { duration: '3m', target: 50 },   // Increase to 50 builds/s
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],  // 95% under 5s
    http_req_failed: ['rate<0.1'],       // <10% failure rate
  },
};

export default function () {
  const parserId = `load-test-${randomString(8)}`;
  
  const payload = JSON.stringify({
    parser_id: parserId,
    s3_bucket: 'knative-lambda-fusion-modules-tmp',
    s3_prefix: `global/parser/${parserId}/`,
    language: 'python',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'ce-type': 'build.start',
      'ce-source': 'k6-load-test',
      'ce-id': randomString(32),
    },
  };

  const res = http.post('http://builder-service.knative-lambda/build', payload, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 5s': (r) => r.timings.duration < 5000,
  });

  sleep(1);
}
```

### Running Load Tests

```bash
# Run k6 load test
make test-load-k6 EVENT_TYPE=build

# Run Python async load test
make test-load-python CONCURRENT=50 DURATION=60

# Monitor during load test
watch -n 1 'kubectl get pods -n knative-lambda'
```

---

## 🔧 Test Environments

### Local Testing

```bash
# Run tests locally (no K8s)
make test-unit

# Run with local kind cluster
kind create cluster --name test
make test-integration
```

### CI/CD Testing

```yaml
# .github/workflows/test.yaml
name: Test
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - run: make test-unit
      - run: make test-coverage

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: helm/kind-action@v1
      - run: make install-deps-test
      - run: make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: helm/kind-action@v1
      - run: make deploy-test
      - run: make test-e2e
```

---

## 📈 Test Metrics & Reporting

### Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View HTML report
go tool cover -html=coverage.out

# Enforce minimum coverage
make test-coverage-check MIN_COVERAGE=80
```

### Test Results

```bash
# Generate JUnit XML (for CI)
go test ./... -v 2>&1 | go-junit-report > junit.xml

# Generate test summary
make test-summary
```

---

## 🎯 Testing Best Practices

### 1. **Test Isolation**

✅ Each test should be independent  
✅ Use unique IDs (UUID) for test resources  
✅ Clean up resources after test

### 2. **Fast Feedback**

✅ Unit tests run in <5s  
✅ Integration tests run in <2min  
✅ E2E tests run in <10min

### 3. **Deterministic Tests**

❌ Avoid flaky tests (race conditions, timing)  
✅ Use retries only when necessary  
✅ Mock external dependencies in unit tests

### 4. **Clear Assertions**

```go
// Bad: Unclear failure
assert.True(t, len(result) > 0)

// Good: Clear failure message
assert.Greater(t, len(result), 0, "Expected non-empty result")
```

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

