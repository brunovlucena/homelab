# ⚡ SRE-007: Observability Enhancement

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-224/sre-007-observability-enhancement
**Priority**: P2
**Story Points**: 5  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-173/sre-007-observability-enhancement  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** comprehensive observability (metrics, logs, traces)  
**So that** I can quickly diagnose issues and understand system behavior

---


## 🎯 Acceptance Criteria

- [ ] [ ] Distributed tracing covers 100% of build flows
- [ ] [ ] Custom Grafana dashboards for key workflows
- [ ] [ ] Structured logging with correlation IDs
- [ ] [ ] OpenTelemetry integration enabled
- [ ] [ ] SLO dashboards track 99.9% availability
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] Distributed tracing covers 100% of build flows
- [ ] Custom Grafana dashboards for key workflows
- [ ] Structured logging with correlation IDs
- [ ] OpenTelemetry integration enabled
- [ ] SLO dashboards track 99.9% availability

---

## 📊 Three Pillars of Observability

### 1. Metrics (Prometheus)

```promql
# Golden Signals
# Latency
histogram_quantile(0.95, rate(build_duration_seconds_bucket[5m]))

# Traffic
rate(builds_total[5m])

# Errors
rate(build_failures_total[5m]) / rate(builds_total[5m])

# Saturation
kaniko_jobs_running / kaniko_jobs_limit
```

### 2. Logs (Structured JSON)

```json
{
  "level": "info",
  "ts": "2025-10-29T10:45:32Z",
  "caller": "handler/event_handler.go:123",
  "msg": "Processing build event",
  "correlation_id": "abc123-def456",
  "parser_id": "parser-xyz",
  "third_party_id": "customer-123",
  "duration_ms": 45,
  "trace_id": "a1b2c3d4e5f6"
}
```

### 3. Traces (OpenTelemetry → Tempo)

```
Trace: Build Function abc123 (Total: 65s)
│
├─ receive_cloudevent (1ms)
├─ validate_event (2ms)
├─ fetch_s3_parser (5s)
├─ create_kaniko_job (50ms)
├─ build_image (45s)  ← Slowest span
│  ├─ pull_base_image (10s)
│  ├─ copy_files (2s)
│  ├─ run_pip_install (25s)  ← Bottleneck
│  └─ push_to_ecr (8s)
└─ create_knative_service (15s)
```

---

## 📊 Grafana Dashboards

### Dashboard 1: Platform Overview

**Panels**:
1. Build success rate (last 24h)
2. Build duration (p50, p95, p99)
3. Active Kaniko jobs
4. RabbitMQ queue depth
5. Error rate by type
6. Cost per build

### Dashboard 2: Build Pipeline

**Panels**:
1. Build stages duration (S3 fetch, Kaniko build, ECR push)
2. Cache hit rate
3. Image size distribution
4. Failed jobs by error type
5. Concurrent builds over time

### Dashboard 3: Function Performance

**Panels**:
1. Cold start latency
2. Active functions
3. Scale-up/down events
4. Request rate per function
5. CPU/Memory usage

---

## 💡 Pro Tips

### Debugging with Traces
- Use correlation IDs to link CloudEvents → Logs → Traces
- Tempo integrates with Grafana for seamless navigation
- Exemplars link metrics → traces

### Log Aggregation
- Use `kubectl logs` with `-l` label selector for bulk analysis
- Forward logs to Loki for long-term retention
- Use `jq` for JSON log parsing

### Alerting Best Practices
- Alert on SLO burn rate, not raw metrics
- Use severity levels: P0 (page), P1 (ticket), P2 (email)
- Include runbook links in alert annotations

---

