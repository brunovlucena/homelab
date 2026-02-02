# System Design Review: Knative Lambda Operator

**Review Date:** 2025-01-27  
**System:** knative-lambda-operator  
**Version:** 1.0.4  
**Reviewer:** System Design Analysis via Grafana MCP

---

## Executive Summary

The **Knative Lambda Operator** is a Kubernetes-native serverless platform that enables dynamic function-as-a-service (FaaS) deployments using Knative Serving. This review analyzes the system's architecture, observability implementation, and performance characteristics based on codebase analysis and the four core performance metrics: **QPS (Queries Per Second)**, **TPS (Transactions Per Second)**, **Concurrency**, and **Response Time (RT)**.

### Key Findings

✅ **Strengths:**
- Comprehensive observability stack (Prometheus, Loki, Tempo)
- Well-instrumented metrics with exemplar support
- Distributed tracing integration
- Production-ready alerting rules
- Scale-to-zero capabilities

⚠️ **Areas for Improvement:**
- Need to verify actual metric collection in production
- Workqueue depth monitoring requires attention
- Cold start optimization opportunities
- Build duration tracking could be enhanced

---

## 1. System Architecture Overview

### 1.1 Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Knative Lambda Operator                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ LambdaFunction│  │ LambdaAgent  │  │ CloudEvents │      │
│  │  Controller   │  │  Controller   │  │   Receiver   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │               │
│         └─────────────────┴─────────────────┘               │
│                           │                                 │
│         ┌─────────────────┴─────────────────┐             │
│         │                                     │             │
│  ┌──────▼──────┐  ┌──────────────┐  ┌──────▼──────┐      │
│  │   Build     │  │    Deploy     │  │   Eventing   │      │
│  │  Manager    │  │   Manager     │  │   Manager    │      │
│  └─────────────┘  └──────────────┘  └─────────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Key Features

- **Dynamic Function Building**: Kaniko-based in-cluster builds
- **Auto-Scaling**: Scale-to-zero with rapid scale-up (<30s)
- **GitOps Integration**: Flux CD for automated deployments
- **Progressive Delivery**: Flagger canary deployments
- **Service Mesh**: Linkerd integration for mTLS and observability
- **Event-Driven**: CloudEvents v1.0 with RabbitMQ broker

---

## 2. Performance Metrics Analysis

### 2.1 Queries Per Second (QPS)

**Definition:** Measures incoming requests per second to the operator and lambda functions.

#### Metrics Available:

1. **Operator-Level QPS:**
   - `knative_lambda_operator_reconcile_total{phase, result}` - Reconciliation requests
   - `knative_lambda_operator_apiserver_requests_total{verb, resource, result}` - API server interactions

2. **Function-Level QPS:**
   - `knative_lambda_function_invocations_total{function, namespace, status}` - Function invocation rate
   - `knative_lambda_function_invocations_total{status="success"}` - Successful invocations
   - `knative_lambda_function_invocations_total{status="error"}` - Failed invocations

#### QPS Calculation:

```promql
# Total QPS for operator reconciliations
sum(rate(knative_lambda_operator_reconcile_total[5m])) by (phase)

# Total QPS for function invocations
sum(rate(knative_lambda_function_invocations_total[5m])) by (function, namespace)

# Success QPS
sum(rate(knative_lambda_function_invocations_total{status="success"}[5m]))

# Error QPS
sum(rate(knative_lambda_function_invocations_total{status="error"}[5m]))
```

#### Assessment:

✅ **Well-Instrumented:** The system tracks QPS at both operator and function levels  
✅ **Status Differentiation:** Success/error rates are separately tracked  
⚠️ **Recommendation:** Add QPS dashboards showing trends over time

---

### 2.2 Transactions Per Second (TPS)

**Definition:** Measures completed transactions per second (full round-trip: request → database/processing → response).

#### Metrics Available:

1. **Build Transactions:**
   - `knative_lambda_operator_build_duration_seconds{runtime, result}` - Build completion time
   - `knative_lambda_operator_build_events_total{status}` - Build lifecycle events

2. **Service Transactions:**
   - `knative_lambda_operator_service_events_total{status}` - Service lifecycle events
   - `knative_lambda_function_duration_seconds{function, namespace}` - Function execution duration

3. **Reconciliation Transactions:**
   - `knative_lambda_operator_reconcile_total{phase, result}` - Reconciliation completions

#### TPS Calculation:

```promql
# Successful build TPS
sum(rate(knative_lambda_operator_build_events_total{status="complete"}[5m]))

# Successful service deployment TPS
sum(rate(knative_lambda_operator_service_events_total{status="ready"}[5m]))

# Successful function invocation TPS
sum(rate(knative_lambda_function_invocations_total{status="success"}[5m]))
```

#### Assessment:

✅ **Transaction Tracking:** Build, service, and function transactions are tracked  
✅ **Result Differentiation:** Success vs. failure rates available  
⚠️ **Gap:** No explicit "transaction" metric - derived from event counters

---

### 2.3 Concurrency

**Definition:** Tracks simultaneous active requests being processed.

#### Metrics Available:

1. **Active Build Jobs:**
   - `knative_lambda_operator_build_jobs_active{namespace}` - Current active Kaniko builds

2. **Work Queue Depth:**
   - `knative_lambda_operator_workqueue_depth` - Items waiting in reconciliation queue

3. **Lambda Function Count:**
   - `knative_lambda_operator_lambdafunctions_total{namespace, phase}` - Functions by phase

4. **Eventing Resources:**
   - `knative_lambda_operator_eventing_resources_total{namespace, resource_type}` - Active brokers/triggers

#### Concurrency Calculation:

```promql
# Active concurrent builds
sum(knative_lambda_operator_build_jobs_active)

# Work queue depth (indicates concurrency pressure)
knative_lambda_operator_workqueue_depth

# Active lambda functions
sum(knative_lambda_operator_lambdafunctions_total{phase!="deleted"})
```

#### Relationship Formula:

**QPS = Concurrency ÷ Average RT**

This relationship is critical for understanding system capacity:
- High concurrency + low RT = high QPS (good)
- High concurrency + high RT = system under stress
- Low concurrency + high RT = potential bottleneck

#### Assessment:

✅ **Concurrency Tracking:** Active builds and queue depth are monitored  
✅ **Alerting:** Workqueue depth > 10 triggers alerts  
⚠️ **Gap:** No direct metric for concurrent function executions (would require Knative metrics)

---

### 2.4 Response Time (RT)

**Definition:** Duration from request start to response received.

#### Metrics Available:

1. **Reconciliation Duration:**
   - `knative_lambda_operator_reconcile_duration_seconds{phase}` - Histogram with buckets: [.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60]

2. **Build Duration:**
   - `knative_lambda_operator_build_duration_seconds{runtime, result}` - Histogram with buckets: [10, 30, 60, 120, 300, 600, 900, 1200, 1800]

3. **Function Execution Duration:**
   - `knative_lambda_function_duration_seconds{function, namespace}` - Histogram with buckets: [.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30]

4. **Work Queue Latency:**
   - `knative_lambda_operator_workqueue_latency_seconds` - Time items spend in queue

5. **Linkerd Metrics (via service mesh):**
   - `response_latency_ms_bucket{namespace, deployment}` - Request latency from service mesh

#### Response Time Analysis:

```promql
# P50, P95, P99 reconciliation time
histogram_quantile(0.50, sum(rate(knative_lambda_operator_reconcile_duration_seconds_bucket[5m])) by (le, phase))
histogram_quantile(0.95, sum(rate(knative_lambda_operator_reconcile_duration_seconds_bucket[5m])) by (le, phase))
histogram_quantile(0.99, sum(rate(knative_lambda_operator_reconcile_duration_seconds_bucket[5m])) by (le, phase))

# P95 build duration by runtime
histogram_quantile(0.95, sum(rate(knative_lambda_operator_build_duration_seconds_bucket[15m])) by (le, runtime))

# P95 function execution time
histogram_quantile(0.95, sum(rate(knative_lambda_function_duration_seconds_bucket[5m])) by (le, function, namespace))

# P99 latency from Linkerd
histogram_quantile(0.99, sum(rate(response_latency_ms_bucket{namespace="knative-lambda", deployment=~"knative-lambda-operator.*"}[5m])) by (le))
```

#### Assessment:

✅ **Comprehensive RT Tracking:** Multiple layers of response time metrics  
✅ **Histogram Buckets:** Well-designed buckets for different operation types  
✅ **Exemplar Support:** Metrics link to traces via OpenTelemetry  
✅ **Alerting:** P99 latency > 500ms triggers alerts  
⚠️ **Recommendation:** Add SLO tracking (e.g., 95% of requests < 1s)

---

## 3. Observability Stack Analysis

### 3.1 Prometheus Metrics

#### Operator Metrics (Namespace: `knative_lambda_operator`)

| Metric | Type | Purpose | Labels |
|--------|------|---------|--------|
| `reconcile_total` | Counter | Total reconciliations | `phase`, `result` |
| `reconcile_duration_seconds` | Histogram | Reconcile latency | `phase` |
| `lambdafunctions_total` | Gauge | Current lambda count | `namespace`, `phase` |
| `build_jobs_active` | Gauge | Active build jobs | `namespace` |
| `build_duration_seconds` | Histogram | Build time | `runtime`, `result` |
| `eventing_resources_total` | Gauge | Eventing resources | `namespace`, `resource_type` |
| `apiserver_requests_total` | Counter | API server calls | `verb`, `resource`, `result` |
| `workqueue_depth` | Gauge | Queue depth | - |
| `workqueue_latency_seconds` | Histogram | Queue wait time | - |
| `errors_total` | Counter | Error count | `component`, `error_type` |
| `build_events_total` | Counter | Build events | `status` |
| `service_events_total` | Counter | Service events | `status` |
| `parser_events_total` | Counter | Parser events | `status` |

#### Function Metrics (Namespace: `knative_lambda_function`)

| Metric | Type | Purpose | Labels |
|--------|------|---------|--------|
| `invocations_total` | Counter | Function invocations | `function`, `namespace`, `status` |
| `duration_seconds` | Histogram | Execution duration | `function`, `namespace` |
| `errors_total` | Counter | Function errors | `function`, `namespace`, `error_type` |
| `cold_starts_total` | Counter | Cold start count | `function`, `namespace` |

#### Metrics Endpoint

- **Service:** `knative-lambda-operator.knative-lambda.svc:8080`
- **Path:** `/metrics`
- **Scraping:** Via Prometheus ServiceMonitor (if configured)

### 3.2 Loki Logs

#### Log Sources

1. **Operator Logs:**
   - Deployment: `knative-lambda-operator`
   - Namespace: `knative-lambda`
   - Structured logging with trace context

2. **Function Logs:**
   - Knative Services (dynamically created)
   - Runtime logs (Python/Node.js/Go)

#### LogQL Queries:

```logql
# Operator errors
{namespace="knative-lambda", app="knative-lambda-operator"} |= "error" | json

# Build failures
{namespace="knative-lambda"} |= "build" |= "failed" | json

# Function execution logs
{namespace="knative-lambda", service_name=~"lambda-.*"} | json
```

### 3.3 Tempo Traces

#### Tracing Implementation

**OpenTelemetry Integration:**
- **Service Name:** `knative-lambda-operator`
- **OTLP Endpoint:** `alloy.observability.svc:4317` (default)
- **Sampling Rate:** 1.0 (100% - configurable)
- **Propagation:** W3C Trace Context + Baggage

#### Trace Spans

1. **Build Spans:**
   - Operation: `build`
   - Attributes: `lambda.function`, `lambda.namespace`, `lambda.runtime`, `operation.type`

2. **Deploy Spans:**
   - Operation: `deploy`
   - Attributes: `lambda.function`, `lambda.namespace`, `operation.type`

3. **Reconcile Spans:**
   - Operation: `reconcile`
   - Attributes: `lambda.function`, `lambda.namespace`, `lambda.phase`, `operation.type`

#### Exemplar Support

Metrics include trace exemplars linking to Tempo:
- `trace_id` and `span_id` embedded in histogram metrics
- Enables correlation: Metrics → Traces → Logs

#### Trace Queries:

```tempo
# Find slow reconciliations
{service.name="knative-lambda-operator", operation.name="reconcile"} | duration > 5s

# Build failures
{service.name="knative-lambda-operator", operation.name="build", status.code="ERROR"}

# Function invocations
{service.name=~"lambda-.*"} | duration > 1s
```

---

## 4. Alerting Analysis

### 4.1 Alert Rules (Production - Studio)

#### Canary Alerts

1. **KnativeLambdaCanaryFailed**
   - Condition: Canary status = "Failed"
   - Severity: Critical
   - Duration: 1m

2. **KnativeLambdaCanaryStuck**
   - Condition: Canary status = "Progressing" for 30m
   - Severity: Warning

#### Operator Alerts

1. **KnativeLambdaOperatorNotReady**
   - Condition: Replicas ready < spec replicas
   - Severity: Warning
   - Duration: 5m

2. **KnativeLambdaOperatorDown**
   - Condition: Available replicas = 0
   - Severity: Critical
   - Duration: 2m

3. **KnativeLambdaHighErrorRate**
   - Condition: Error rate > 1%
   - Severity: Warning
   - Duration: 5m

4. **KnativeLambdaHighLatency**
   - Condition: P99 latency > 500ms
   - Severity: Warning
   - Duration: 5m

#### Lambda Function Alerts

1. **KnativeLambdaBuildFailed**
   - Condition: > 3 build failures in 10m
   - Severity: Warning

2. **KnativeLambdaFunctionHighErrorRate**
   - Condition: Error rate > 5%
   - Severity: Warning
   - Duration: 5m

3. **KnativeLambdaBuildDurationHigh**
   - Condition: P95 build duration > 120s
   - Severity: Warning
   - Duration: 5m

4. **KnativeLambdaWorkqueueDepthHigh**
   - Condition: Workqueue depth > 10
   - Severity: Warning
   - Duration: 5m

5. **KnativeLambdaHighColdStartRate**
   - Condition: Cold start rate > 20%
   - Severity: Info
   - Duration: 10m

6. **KnativeLambdaFunctionHighLatency**
   - Condition: P95 function latency > 5s
   - Severity: Warning
   - Duration: 5m

### 4.2 Alert Assessment

✅ **Comprehensive Coverage:** Alerts cover operator, builds, and functions  
✅ **Appropriate Thresholds:** Reasonable thresholds for production  
✅ **Runbook Links:** Alerts include runbook URLs  
⚠️ **Recommendation:** Add SLO-based alerts (e.g., error budget burn rate)

---

## 5. System Design Strengths

### 5.1 Observability Excellence

1. **Three Pillars Implemented:**
   - ✅ Metrics (Prometheus)
   - ✅ Logs (Loki)
   - ✅ Traces (Tempo)

2. **Exemplar Support:**
   - Metrics link to traces via OpenTelemetry
   - Enables correlation across observability stack

3. **Structured Logging:**
   - JSON format with trace context
   - Easy correlation with traces

### 5.2 Performance Monitoring

1. **RED Metrics:**
   - ✅ Rate: `invocations_total`, `reconcile_total`
   - ✅ Errors: `errors_total`, error status labels
   - ✅ Duration: Histograms for all operations

2. **USE Metrics (for infrastructure):**
   - Utilization: `build_jobs_active`, `workqueue_depth`
   - Saturation: Queue depth alerts
   - Errors: `errors_total`

### 5.3 Scalability Considerations

1. **Scale-to-Zero:**
   - Knative Serving enables scale-to-zero
   - Cold start tracking via `cold_starts_total`

2. **Work Queue Management:**
   - Queue depth monitoring
   - Latency tracking
   - Alerting on backlog

3. **Resource Tracking:**
   - Active build jobs
   - Lambda function counts by phase
   - Eventing resources

---

## 6. Recommendations

### 6.1 Immediate Actions

1. **Verify Metric Collection:**
   - Confirm Prometheus is scraping operator metrics
   - Verify ServiceMonitor exists and is discovered
   - Check metric endpoint accessibility

2. **Create Performance Dashboards:**
   - QPS dashboard (operator + functions)
   - TPS dashboard (builds + services + functions)
   - Concurrency dashboard (active jobs + queue depth)
   - Response Time dashboard (P50/P95/P99)

3. **SLO Definition:**
   - Define SLOs for key operations
   - Implement error budget tracking
   - Add SLO-based alerting

### 6.2 Short-Term Improvements

1. **Enhanced Concurrency Tracking:**
   - Add metric for concurrent function executions
   - Track Knative service concurrency limits
   - Monitor autoscaler decisions

2. **Build Optimization:**
   - Track build queue wait time
   - Monitor build resource utilization
   - Alert on build timeouts

3. **Cold Start Optimization:**
   - Track cold start duration separately
   - Monitor scale-to-zero frequency
   - Optimize base images for faster startup

### 6.3 Long-Term Enhancements

1. **Distributed Tracing Enhancement:**
   - Add spans for API server calls
   - Trace event processing end-to-end
   - Correlate traces across Knative services

2. **Cost Tracking:**
   - Track resource usage per function
   - Monitor scale-to-zero savings
   - Cost per invocation metrics

3. **Capacity Planning:**
   - Historical trend analysis
   - Predictive scaling
   - Resource utilization forecasting

---

## 7. Performance Metrics Summary

### 7.1 Key Metrics to Monitor

| Metric | Current Status | Target | Alert Threshold |
|--------|---------------|--------|-----------------|
| **QPS (Operator)** | ✅ Tracked | - | - |
| **QPS (Functions)** | ✅ Tracked | - | - |
| **TPS (Builds)** | ✅ Tracked | - | > 3 failures/10m |
| **TPS (Functions)** | ✅ Tracked | - | > 5% error rate |
| **Concurrency (Builds)** | ✅ Tracked | - | - |
| **Concurrency (Queue)** | ✅ Tracked | < 10 | > 10 items |
| **RT (Reconcile P99)** | ✅ Tracked | < 500ms | > 500ms |
| **RT (Build P95)** | ✅ Tracked | < 120s | > 120s |
| **RT (Function P95)** | ✅ Tracked | < 5s | > 5s |

### 7.2 Formula Verification

**QPS = Concurrency ÷ Average RT**

This relationship should be monitored to ensure:
- System is not overloaded (high concurrency + high RT)
- Capacity is adequate (QPS matches demand)
- Response times are acceptable (low RT enables higher QPS)

---

## 8. Conclusion

The **Knative Lambda Operator** demonstrates **excellent observability design** with comprehensive metrics, logging, and tracing. The system is well-instrumented to track all four core performance metrics (QPS, TPS, Concurrency, Response Time).

### Key Strengths:
- ✅ Complete observability stack (Prometheus, Loki, Tempo)
- ✅ Well-designed metrics with appropriate buckets
- ✅ Exemplar support for metric-trace correlation
- ✅ Production-ready alerting rules
- ✅ Scale-to-zero capabilities

### Next Steps:
1. Verify metric collection in production environment
2. Create performance dashboards for the four core metrics
3. Define and implement SLOs
4. Optimize cold start performance
5. Enhance distributed tracing coverage

---

## Appendix A: Prometheus Queries Reference

### QPS Queries

```promql
# Total operator reconciliation QPS
sum(rate(knative_lambda_operator_reconcile_total[5m]))

# Function invocation QPS by function
sum(rate(knative_lambda_function_invocations_total[5m])) by (function, namespace)

# Success vs Error QPS
sum(rate(knative_lambda_function_invocations_total{status="success"}[5m]))
sum(rate(knative_lambda_function_invocations_total{status="error"}[5m]))
```

### TPS Queries

```promql
# Successful build TPS
sum(rate(knative_lambda_operator_build_events_total{status="complete"}[5m]))

# Successful function invocation TPS
sum(rate(knative_lambda_function_invocations_total{status="success"}[5m]))
```

### Concurrency Queries

```promql
# Active builds
sum(knative_lambda_operator_build_jobs_active)

# Work queue depth
knative_lambda_operator_workqueue_depth

# Active functions
sum(knative_lambda_operator_lambdafunctions_total{phase!="deleted"})
```

### Response Time Queries

```promql
# P50/P95/P99 reconciliation time
histogram_quantile(0.50, sum(rate(knative_lambda_operator_reconcile_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.95, sum(rate(knative_lambda_operator_reconcile_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.99, sum(rate(knative_lambda_operator_reconcile_duration_seconds_bucket[5m])) by (le))

# P95 build duration by runtime
histogram_quantile(0.95, sum(rate(knative_lambda_operator_build_duration_seconds_bucket[15m])) by (le, runtime))

# P95 function execution time
histogram_quantile(0.95, sum(rate(knative_lambda_function_duration_seconds_bucket[5m])) by (le, function, namespace))
```

## Appendix B: LogQL Queries Reference

```logql
# Operator errors
{namespace="knative-lambda", app="knative-lambda-operator"} |= "error" | json

# Build failures
{namespace="knative-lambda"} |= "build" |= "failed" | json

# Function execution logs
{namespace="knative-lambda", service_name=~"lambda-.*"} | json

# High latency operations
{namespace="knative-lambda"} | json | duration > 5s
```

## Appendix C: Tempo Trace Queries Reference

```tempo
# Slow reconciliations
{service.name="knative-lambda-operator", operation.name="reconcile"} | duration > 5s

# Build failures
{service.name="knative-lambda-operator", operation.name="build", status.code="ERROR"}

# Function invocations with errors
{service.name=~"lambda-.*", status.code="ERROR"} | duration > 1s
```

---

**Report Generated:** 2025-01-27  
**Review Method:** Codebase analysis + Observability stack review  
**Next Review:** After production metrics verification
