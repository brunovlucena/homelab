# 🔄 DEVOPS-020: Integrate agent-sre with Notifi observability stack

**Status**: Backlog  | **Priority**: P3**Linear URL**: https://linear.app/bvlucena/issue/BVL-14/integration-integrate-agent-sre-with-notifi-observability-stack | **Status**: Backlog  | **Priority**: P3**Linear URL**: https://linear.app/bvlucena/issue/BVL-14/integration-integrate-agent-sre-with-notifi-observability-stack | **Story Points**: 13

**Created**: 2025-12-26T14:37:39.033Z  
**Updated**: 2025-12-26T14:37:39.033Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Integrate agent-sre with Notifi observability stack (Prometheus, Grafana, Loki, Tempo) for comprehensive monitoring and dashboards.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** integrate agent-sre with notifi observability stack  
**So that** I can improve system reliability, security, and performance

---


## 📋 Current State

* agent-sre exposes basic metrics
* Prometheus/Loki queries work
* Need Grafana dashboards and alerting

## 🔧 Tasks

### 1\. Metrics Enhancement

- [ ] Review existing metrics exposed by agent-sre
- [ ] Add additional metrics if needed (model latency, token usage, etc.)
- [ ] Ensure metrics follow Prometheus naming conventions
- [ ] Verify ServiceMonitor configuration (if operator doesn't create it)
- [ ] Test metrics scraping

### 2\. Grafana Dashboard Creation

- [ ] Review existing Grafana dashboard structure in Notifi
- [ ] Create agent-sre dashboard following Notifi patterns
- [ ] Add panels for agent health and status
- [ ] Add panels for model inference metrics (latency, tokens, errors)
- [ ] Add panels for report generation metrics
- [ ] Add panels for Prometheus/Loki query performance
- [ ] Add panels for resource utilization
- [ ] Add appropriate time ranges and refresh intervals
- [ ] Test dashboard functionality

### 3\. Alerting Rules

- [ ] Create PrometheusRule for agent-sre alerts
- [ ] Configure alerts for agent unavailable/pod crashes
- [ ] Configure alerts for high model inference latency
- [ ] Configure alerts for high error rates
- [ ] Configure alerts for report generation failures
- [ ] Configure alerts for Ollama connectivity issues
- [ ] Configure alert severity levels
- [ ] Configure alert routing (Slack/PagerDuty/etc.)
- [ ] Test alert triggers

### 4\. Distributed Tracing

- [ ] Configure OpenTelemetry tracing (if not already done)
- [ ] Add trace instrumentation to agent-sre code
- [ ] Configure Tempo datasource in Grafana
- [ ] Verify traces are collected
- [ ] Add trace visualization to dashboard

### 5\. Logging Enhancement

- [ ] Ensure structured logging format
- [ ] Add correlation IDs for request tracing
- [ ] Verify Loki labels are correct
- [ ] Create log queries/alerts if needed
- [ ] Integrate logs with Grafana dashboard

### 6\. Documentation

- [ ] Document metrics exposed
- [ ] Document dashboard usage
- [ ] Document alerting runbooks
- [ ] Update troubleshooting guides

## ✅ Acceptance Criteria

- [ ] Metrics properly exposed and scraped
- [ ] Grafana dashboard created and functional
- [ ] Alerting rules configured and tested
- [ ] Distributed tracing working (if applicable)
- [ ] Logs properly collected and queryable
- [ ] Documentation complete
- [ ] Team trained on monitoring

## 📚 References

* Notifi dashboard patterns: `20-platform/services/dashboards/deploy/dashboards/`
* Alert patterns: `20-platform/services/prometheus/deploy/`

## 🔗 Dependencies

INFRA-MIG-009: Deploy and validate agent-sre in Notifi environment
