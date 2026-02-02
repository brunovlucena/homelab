# 🌐 BACKEND-006: Knative Service Management

**Priority**: P1 | **Status**: ✅ Implemented K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-226/backend-006-knative-service-management


---

## 📋 User Story

**As a** Backend Developer  
**I want to** dynamically create and manage Knative Services for successful builds  
**So that** parsers are deployed as auto-scaling serverless functions that can process CloudEvents

---

## 🎯 Acceptance Criteria

### ✅ Service Creation
- [ ] Create Knative Service after successful build
- [ ] Use built Docker image from ECR
- [ ] Configure service with proper resource limits
- [ ] Set up auto-scaling configuration
- [ ] Create ServiceAccount for proper RBAC
- [ ] Create ConfigMap for parser configuration
- [ ] Deploy metrics-pusher sidecar (optional)

### ✅ Trigger Management
- [ ] Create Knative Trigger for event routing
- [ ] Configure filter by `third_party_id` and `parser_id`
- [ ] Connect to appropriate broker
- [ ] Set retry policy and backoff
- [ ] Configure RabbitMQ parallelism

### ✅ Service Updates
- [ ] Update service when new image is built
- [ ] Preserve resourceVersion for optimistic concurrency
- [ ] Handle immutable field conflicts
- [ ] Trigger rolling deployment
- [ ] Zero-downtime updates

### ✅ Service Deletion
- [ ] Delete Service when requested
- [ ] Delete associated Trigger
- [ ] Delete ConfigMap
- [ ] Delete ServiceAccount
- [ ] Clean up all resources in order

### ✅ Parallel Resource Creation
- [ ] Create Service and Trigger simultaneously
- [ ] Use goroutines for parallel operations
- [ ] Handle partial failures gracefully
- [ ] Wait for both operations to complete
- [ ] Improved performance vs sequential creation

---

## 🔧 Technical Implementation

### File: `internal/handler/service_manager.go`

```go
// Service Manager Interface
type ServiceManager interface {
    CreateService(ctx context.Context, serviceName string, completionData *builds.BuildCompletionEventData) error
    DeleteService(ctx context.Context, serviceName string) error
    CheckServiceExists(ctx context.Context, serviceName string) (bool, error)
    GenerateServiceName(thirdPartyID, parserID string) string
}

// Create Service with Parallel Resource Creation
func (s *ServiceManagerImpl) CreateService(ctx context.Context, serviceName string, completionData *builds.BuildCompletionEventData) error {
    ctx, span := s.obs.StartSpan(ctx, "create_or_update_knative_service")
    defer span.End()
    
    // 1. Create ServiceAccount
    serviceAccount := s.CreateServiceAccountResource(serviceName, completionData)
    if err := s.ApplyResource(ctx, serviceAccount); err != nil {
        return err
    }
    
    // 2. Create ConfigMap
    configMap := s.CreateConfigMapResource(serviceName, completionData)
    if err := s.ApplyResource(ctx, configMap); err != nil {
        return err
    }
    
    // 3. Parallel creation of Service and Trigger
    resultChan := make(chan resourceResult, 2)
    
    // Create Knative Service
    go func() {
        knativeService := s.CreateKnativeServiceResource(serviceName, completionData)
        err := s.ApplyResource(ctx, knativeService)
        resultChan <- resourceResult{resourceType: "service", err: err}
    }()
    
    // Create Trigger
    go func() {
        trigger := s.CreateTriggerResource(serviceName, completionData)
        err := s.ApplyResource(ctx, trigger)
        resultChan <- resourceResult{resourceType: "trigger", err: err}
    }()
    
    // Wait for both to complete
    var serviceErr, triggerErr error
    for i := 0; i < 2; i++ {
        result := <-resultChan
        if result.resourceType == "service" {
            serviceErr = result.err
        } else {
            triggerErr = result.err
        }
    }
    
    if serviceErr != nil {
        return fmt.Errorf("failed to create service: %w", serviceErr)
    }
    if triggerErr != nil {
        return fmt.Errorf("failed to create trigger: %w", triggerErr)
    }
    
    s.obs.Info(ctx, "Successfully created Knative service and trigger",
        "service_name", serviceName)
    
    return nil
}
```

### Knative Service Resource

```go
func (s *ServiceManagerImpl) CreateKnativeServiceResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured {
    return &unstructured.Unstructured{
        Object: map[string]interface{}{
            "apiVersion": "serving.knative.dev/v1",
            "kind":       "Service",
            "metadata": map[string]interface{}{
                "name":      serviceName,
                "namespace": s.config.Namespace,
                "labels": map[string]interface{}{
                    "app":                                 "knative-lambda-service",
                    "build.notifi.network/third-party-id": completionData.ThirdPartyID,
                    "build.notifi.network/parser-id":      completionData.ParserID,
                    "build.notifi.network/content-hash":   completionData.ContentHash,
                },
            },
            "spec": map[string]interface{}{
                "template": map[string]interface{}{
                    "metadata": map[string]interface{}{
                        "annotations": map[string]interface{}{
                            "autoscaling.knative.dev/class":          "kpa.autoscaling.knative.dev",
                            "autoscaling.knative.dev/target":         "100",
                            "autoscaling.knative.dev/minScale":       "0",
                            "autoscaling.knative.dev/maxScale":       "10",
                            "autoscaling.knative.dev/scaleToZero":    "5m",
                        },
                    },
                    "spec": map[string]interface{}{
                        "serviceAccountName": serviceName,
                        "containerConcurrency": 10,
                        "containers": []map[string]interface{}{
                            {
                                "name":  "lambda",
                                "image": completionData.ImageURI,
                                "ports": []map[string]interface{}{
                                    {"containerPort": 8080, "name": "http1"},
                                },
                                "resources": map[string]interface{}{
                                    "limits": map[string]interface{}{
                                        "cpu":    "1000m",
                                        "memory": "512Mi",
                                    },
                                    "requests": map[string]interface{}{
                                        "cpu":    "100m",
                                        "memory": "128Mi",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

### Knative Trigger Resource

```go
func (s *ServiceManagerImpl) CreateTriggerResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured {
    return &unstructured.Unstructured{
        Object: map[string]interface{}{
            "apiVersion": "eventing.knative.dev/v1",
            "kind":       "Trigger",
            "metadata": map[string]interface{}{
                "name":      serviceName,
                "namespace": s.config.Namespace,
                "annotations": map[string]interface{}{
                    "rabbitmq.eventing.knative.dev/parallelism": "50",
                },
            },
            "spec": map[string]interface{}{
                "broker": "knative-lambda-service-broker-dev",
                "filter": map[string]interface{}{
                    "attributes": map[string]interface{}{
                        "type":    "network.notifi.lambda.parser.start",
                        "source":  fmt.Sprintf("network.notifi.%s", completionData.ThirdPartyID),
                        "subject": completionData.ParserID,
                    },
                },
                "subscriber": map[string]interface{}{
                    "ref": map[string]interface{}{
                        "apiVersion": "serving.knative.dev/v1",
                        "kind":       "Service",
                        "name":       serviceName,
                    },
                },
                "delivery": map[string]interface{}{
                    "retry":         5,
                    "backoffPolicy": "exponential",
                    "backoffDelay":  "PT1S",
                },
            },
        },
    }
}
```

---

## 📊 Service Lifecycle

```
┌──────────────────┐
│  Build Complete  │
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Create Service  │ ← ServiceAccount + ConfigMap
└────────┬─────────┘
         │
         ├──────────────────────────┐
         │                          │
         ↓                          ↓
┌─────────────────┐      ┌─────────────────┐
│ Knative Service │      │ Knative Trigger │
│  (parallel)     │      │   (parallel)    │
└────────┬────────┘      └────────┬────────┘
         │                        │
         └────────────┬───────────┘
                      │
                      ↓
          ┌───────────────────────┐
          │  Service Ready        │
          │  (Auto-scaling)       │
          └───────────────────────┘
```

---

## 🧪 Testing Scenarios

### 1. Successful Service Creation
```bash
# Upload parser
aws s3 cp parser.js s3://knative-lambda-fusion-modules-tmp/global/parser/test-parser

# Trigger build
make trigger-build-dev PARSER_ID=test-parser

# Wait for build completion
kubectl get jobs -n knative-lambda -w

# Check service created
kubectl get ksvc -n knative-lambda
```

**Expected**:
- Knative Service created with name `lambda-customer-123-test-parser`
- Trigger created with same name
- Service scales from 0 → 1 on first request
- CloudEvents routed correctly

### 2. Service Update (New Build)
```bash
# Update parser code
aws s3 cp parser-v2.js s3://knative-lambda-fusion-modules-tmp/global/parser/test-parser

# Trigger rebuild
make trigger-build-dev PARSER_ID=test-parser

# Monitor update
kubectl describe ksvc lambda-customer-123-test-parser -n knative-lambda
```

**Expected**:
- New image built with different tag
- Service updated with new image
- Rolling deployment triggered
- Zero downtime during update
- Old revision scaled down

### 3. Service Deletion
```bash
# Send service delete event
cd tests && ENV=dev uv run python create-event-delete.py
```

**Expected**:
- Service deleted successfully
- Trigger removed
- ConfigMap deleted
- ServiceAccount deleted
- No orphaned resources

### 4. Auto-Scaling Behavior
```bash
# Send 100 events to trigger scaling
for i in {1..100}; do
  curl -X POST http://broker.knative-lambda \
    -H "Ce-Type: network.notifi.lambda.parser.start" \
    -H "Ce-Source: network.notifi.customer-123" \
    -H "Ce-Subject: test-parser" \
    -d "{\"data\":\"test-$i\"}" &
done

# Watch auto-scaling
kubectl get pods -n knative-lambda -w -l serving.knative.dev/service=lambda-customer-123-test-parser
```

**Expected**:
- Service scales from 0 to N pods (max: 10)
- Scale up based on concurrency target
- Scale down after idle period
- Scale to zero after 5 minutes

---

## 📈 Performance Requirements

- **Service Creation**: < 10s from build complete to service ready
- **Service Update**: < 15s for rolling deployment
- **Cold Start**: < 5s from 0 → 1 pod
- **Scale Up**: < 10s to add additional pods
- **Scale Down**: 5 minutes idle before scale to zero

---

## 🔍 Monitoring & Alerts

### Metrics
- `knative_service_creation_total{status="success | failure"}` - Service creations
- `knative_service_creation_duration_seconds` - Creation latency
- `knative_service_active_pods` - Current pod count
- `knative_service_requests_per_second` - Request rate
- `knative_service_cold_starts_total` - Cold start count

### Alerts
- **Service Creation Failures**: Alert if > 5% failure rate
- **High Cold Start Rate**: Alert if > 50 cold starts/hour
- **Service Update Failures**: Alert on any update failures
- **Scale Limit Reached**: Alert when hitting maxScale frequently

---

## 🏗️ Code References

**Main Files**:
- `internal/handler/service_manager.go` - Service lifecycle management
- `internal/handler/event_handler.go` - Build completion handling
- `internal/config/knative.go` - Knative configuration
- `internal/config/lambda_services.go` - Lambda service config

**K8s Resources**:
- Knative Service (serving.knative.dev/v1)
- Knative Trigger (eventing.knative.dev/v1)
- ServiceAccount (v1)
- ConfigMap (v1)

---

## 📚 Related Documentation

- [BACKEND-003: Kubernetes Job Lifecycle](BACKEND-003-kubernetes-job-lifecycle.md)
- [BACKEND-007: Observability and Tracing](BACKEND-007-observability-tracing.md)
- Knative Serving: https://knative.dev/docs/serving/
- Knative Eventing: https://knative.dev/docs/eventing/

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: ✅ Implemented K

