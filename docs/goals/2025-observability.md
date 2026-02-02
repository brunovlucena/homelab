# 📊 Observability Goals 2025

## Overview

Monitoring, alerting, and observability targets for complete visibility into homelab operations.

---

## 📊 Current State

### Dashboard Coverage

| Component | Dashboard | Metrics | Alerts | Traces |
|-----------|-----------|---------|--------|--------|
| knative-lambda-operator | ✅ | ✅ | ⚠️ | ⚠️ |
| homepage | ⚠️ | ⚠️ | ❌ | ❌ |
| agent-bruno | ✅ | ✅ | ⚠️ | ⚠️ |
| agent-redteam | ✅ | ✅ | ⚠️ | ⚠️ |
| agent-blueteam | ✅ | ✅ | ⚠️ | ⚠️ |
| agent-contracts | ✅ | ✅ | ⚠️ | ⚠️ |
| agent-medical | ✅ | ✅ | ⚠️ | ⚠️ |
| agent-restaurant | ✅ | ⚠️ | ❌ | ❌ |
| agent-tools | ✅ | ⚠️ | ❌ | ❌ |
| agent-pos-edge | ✅ | ⚠️ | ❌ | ❌ |
| agent-store-multibrands | ✅ | ⚠️ | ❌ | ❌ |
| agent-chat | ❌ | ⚠️ | ❌ | ❌ |
| agent-rpg | ❌ | ⚠️ | ❌ | ❌ |
| agent-devsecops | ❌ | ⚠️ | ❌ | ❌ |
| agent-versions | ✅ | ✅ | ❌ | N/A |
| lambdafunctions-versions | ✅ | ✅ | ❌ | N/A |

### Metrics Coverage

| Type | Current | Target |
|------|---------|--------|
| Build info metrics | 12/14 | 14/14 |
| Request metrics | 8/14 | 14/14 |
| Business metrics | 4/14 | 14/14 |
| Error metrics | 6/14 | 14/14 |

---

## 🎯 2025 Targets

### Dashboards

| Target | Q1 | Q2 | Q3 | Q4 |
|--------|----|----|----|----|
| Component dashboards | 14/16 | 16/16 | 16/16 | 16/16 |
| SLO dashboards | 2 | 5 | 8 | 12 |
| Business dashboards | 1 | 3 | 5 | 8 |
| Infrastructure dashboards | 3 | 5 | 6 | 8 |

### Metrics

| Target | Q1 | Q2 | Q3 | Q4 |
|--------|----|----|----|----|
| Build info coverage | 100% | 100% | 100% | 100% |
| RED metrics | 60% | 80% | 90% | 100% |
| USE metrics | 40% | 60% | 80% | 100% |
| Business metrics | 30% | 50% | 70% | 90% |

### Alerting

| Target | Q1 | Q2 | Q3 | Q4 |
|--------|----|----|----|----|
| SLO alerts | 4 | 8 | 12 | 16 |
| Error rate alerts | 4 | 8 | 12 | 16 |
| Latency alerts | 2 | 6 | 10 | 14 |
| Business alerts | 0 | 2 | 4 | 6 |

### Tracing

| Target | Q1 | Q2 | Q3 | Q4 |
|--------|----|----|----|----|
| Traced services | 4 | 8 | 12 | 16 |
| Span coverage | 30% | 50% | 70% | 90% |
| Cross-service traces | 20% | 40% | 60% | 80% |

---

## 📈 Required Metrics

### RED Metrics (Request, Error, Duration)

Every service should expose:

```prometheus
# Request rate
<service>_requests_total{method, endpoint, status}

# Error rate  
<service>_errors_total{method, endpoint, error_type}

# Duration
<service>_request_duration_seconds{method, endpoint}
```

### USE Metrics (Utilization, Saturation, Errors)

For resources:

```prometheus
# Utilization
<resource>_utilization_ratio

# Saturation
<resource>_queue_length

# Errors
<resource>_errors_total
```

### Build Info Metrics

Every component should expose:

```prometheus
<service>_build_info{version, commit}
```

---

## 📊 Dashboard Requirements

### Per-Component Dashboard

- [ ] Overview panel (health, version)
- [ ] Request rate graph
- [ ] Error rate graph
- [ ] Latency percentiles (P50, P95, P99)
- [ ] Resource utilization
- [ ] Active instances

### SLO Dashboard

- [ ] SLO compliance percentage
- [ ] Error budget remaining
- [ ] Burn rate alerts
- [ ] Historical trends

### Agent Versions Dashboard (✅ Completed)

- [x] Total agents count
- [x] Outdated agents count
- [x] Up-to-date agents count
- [x] Version details table
- [x] Red highlighting for outdated

### LambdaFunctions Dashboard (✅ Completed)

- [x] Total functions count
- [x] Outdated images count
- [x] Function versions table
- [x] Invocation rate
- [x] Duration P95

---

## 🔔 Alerting Strategy

### Alert Severity Levels

| Severity | Response | Examples |
|----------|----------|----------|
| Critical | Page immediately | Service down, data loss |
| Warning | Respond in 4h | SLO breach risk |
| Info | Review daily | Performance degradation |

### Required Alerts per Service

| Alert | Type | Threshold |
|-------|------|-----------|
| High error rate | Critical | > 5% for 5min |
| High latency | Warning | P95 > 2x baseline |
| Low availability | Critical | < 95% for 5min |
| Error budget burn | Warning | > 2x normal |

### Alert Routing

```yaml
routes:
  critical:
    - slack: #alerts-critical
    - pagerduty: on-call
  warning:
    - slack: #alerts-warning
  info:
    - slack: #alerts-info
```

---

## 🔍 Distributed Tracing

### OpenTelemetry Integration

| Component | Instrumented | Exporter |
|-----------|--------------|----------|
| agent-bruno | ⚠️ Partial | Tempo |
| agent-redteam | ⚠️ Partial | Tempo |
| agent-contracts | ⚠️ Partial | Tempo |
| knative-lambda | ⚠️ Partial | Tempo |

### Tracing Goals

- [ ] 100% request tracing
- [ ] Cross-service context propagation
- [ ] Error span correlation
- [ ] Latency breakdown analysis

---

## 📝 Logging Strategy

### Log Levels

| Level | Use Case | Retention |
|-------|----------|-----------|
| ERROR | Failures, exceptions | 30 days |
| WARN | Potential issues | 14 days |
| INFO | Normal operations | 7 days |
| DEBUG | Troubleshooting | 1 day |

### Structured Logging

All services should use:

```json
{
  "timestamp": "2025-01-01T00:00:00Z",
  "level": "INFO",
  "service": "agent-bruno",
  "trace_id": "abc123",
  "message": "Request processed",
  "duration_ms": 123,
  "status": "success"
}
```

---

## 🗓️ Milestones

| Milestone | Date | Description |
|-----------|------|-------------|
| Build info 100% | Jan 31 | All services expose build_info |
| Dashboards 100% | Mar 31 | All components have dashboards |
| Alerting foundation | Jun 30 | SLO alerts for Tier 1 |
| Full tracing | Sep 30 | 90% span coverage |
| Observability maturity | Dec 31 | Full observability stack |

---

## 📋 Tools

### Current Stack

| Tool | Purpose | Status |
|------|---------|--------|
| Prometheus | Metrics | ✅ Active |
| Grafana | Dashboards | ✅ Active |
| Loki | Logs | ✅ Active |
| Tempo | Traces | ⚠️ Partial |
| Alertmanager | Alerts | ⚠️ Partial |

### Planned Additions

- [ ] Grafana Incident for incident management
- [ ] SLO tracking with Pyrra or Sloth
- [ ] Grafana OnCall for paging
