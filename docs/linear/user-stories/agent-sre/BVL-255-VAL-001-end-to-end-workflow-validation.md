# ✅ BVL-255 VAL-001: End-to-End Workflow Validation

**Linear URL**: https://linear.app/bvlucena/issue/bvl-255/end-to-end-workflow-validation

---

## 📋 User Story

## 📋 User Story

**As a** Principal Manager Engineer  
**I want to** validate the complete end-to-end workflow from Prometheus alert to remediation execution  
**So that** I can ensure the agent-sre system works correctly in production


---


## 📊 Validation Scope

This ticket validates the complete workflow:
1. PrometheusRule fires alert
2. prometheus-events converts to CloudEvent
3. Agent-SRE receives CloudEvent
4. Linear issue created with SLM context
5. Remediation selected (static annotation or AI-powered)
6. LambdaFunction executed
7. Verification performed
8. Linear issue updated/closed

---

## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


### AC1: Alert Reception & Processing
- [ ] PrometheusRule alert fires correctly
- [ ] prometheus-events converts alert to CloudEvent
- [ ] Agent-SRE receives CloudEvent within 5 seconds
- [ ] CloudEvent parsed correctly (structured and binary modes)
- [ ] Correlation ID propagated throughout workflow
- [ ] OpenTelemetry traces created for entire flow

### AC2: Linear Issue Creation
- [ ] Linear issue created within 10 seconds of alert
- [ ] Issue contains all alert details (labels, annotations, description)
- [ ] SLM context included in issue (SLO, SLI, error budget)
- [ ] Issue priority calculated correctly based on SLM violation
- [ ] Issue assigned to on-call engineer
- [ ] Issue linked to SLO dashboard
- [ ] Issue labels applied correctly

### AC3: Remediation Selection
- [ ] Static annotation remediation selected (if present)
- [ ] AI-powered remediation selected (if no annotation)
- [ ] TRM model used for Phase 1 selection
- [ ] RAG system used for Phase 2 selection
- [ ] Few-shot learning used for Phase 3 selection
- [ ] AI function calling used as fallback (Phase 4)
- [ ] Remediation confidence score calculated
- [ ] Selection method logged

### AC4: LambdaFunction Execution
- [ ] LambdaFunction called via HTTP
- [ ] Parameters extracted correctly from annotations/labels
- [ ] Execution monitored with timeout
- [ ] Success/failure status captured
- [ ] Error messages logged if execution fails
- [ ] Remediation result recorded for training

### AC5: Verification & Issue Updates
- [ ] Metrics queried after remediation
- [ ] Alert resolution detected
- [ ] Linear issue updated with verification results
- [ ] Issue closed when alert resolves
- [ ] Resolution time tracked
- [ ] SLO impact documented

### AC6: Error Handling
- [ ] Malformed CloudEvents handled gracefully
- [ ] Linear API failures don't crash agent
- [ ] LambdaFunction failures create failure tickets
- [ ] Timeout handling works correctly
- [ ] Retry logic functions properly
- [ ] Dead letter queue used for failed events

---

## 🧪 Testing Scenarios

### Scenario 1: Happy Path (Static Annotation)
1. Create PrometheusRule with `lambda_function` annotation
2. Trigger alert
3. Verify Linear issue created
4. Verify LambdaFunction executed
5. Verify issue closed after resolution

### Scenario 2: AI-Powered Selection
1. Create PrometheusRule without annotation
2. Trigger alert
3. Verify AI remediation selection
4. Verify LambdaFunction executed
5. Verify issue updated with selection method

### Scenario 3: Remediation Failure
1. Create PrometheusRule with invalid LambdaFunction
2. Trigger alert
3. Verify failure ticket created
4. Verify error logged correctly
5. Verify issue updated with failure details

### Scenario 4: Alert Resolution
1. Trigger alert
2. Execute remediation
3. Wait for alert to resolve
4. Verify issue closed automatically
5. Verify resolution metrics recorded

### Scenario 5: Concurrent Alerts
1. Trigger 10 alerts simultaneously
2. Verify all processed correctly
3. Verify no race conditions
4. Verify all Linear issues created
5. Verify performance metrics acceptable

---

## 📈 Performance Requirements

- **Alert → Issue Creation**: < 10 seconds (P95)
- **Issue → Remediation**: < 30 seconds (P95)
- **Remediation → Verification**: < 5 minutes (P95)
- **Remediation Success Rate**: > 90%
- **False Positive Rate**: < 5%
- **System Availability**: > 99.5%

## 📊 Success Metrics

- **Zero** workflow failures in production
- **100%** of test scenarios pass
- **All** acceptance criteria met
- **All** performance targets achieved

---

## 🔍 Monitoring & Alerts

### Metrics
- `agent_sre_workflow_duration_seconds` - End-to-end workflow latency
- `agent_sre_workflow_success_total` - Successful workflow completions
- `agent_sre_workflow_failure_total` - Failed workflow attempts
- `agent_sre_alert_to_issue_seconds` - Alert to issue creation time
- `agent_sre_remediation_success_rate` - Remediation success percentage

### Alerts
- **Workflow Failure Rate**: Alert if > 5% over 5 minutes
- **Workflow Latency**: Alert if p95 > 1 minute
- **Remediation Failure Rate**: Alert if > 10% of attempts

---

## 🏗️ Code References

**Main Files**:
- `src/sre_agent/main.py` - Main CloudEvent handler
- `src/sre_agent/langgraph_workflow.py` - Remediation workflow
- `src/sre_agent/linear_handler.py` - Linear issue creation
- `src/sre_agent/lambda_caller.py` - LambdaFunction execution
- `src/sre_agent/intelligent_remediation.py` - Remediation selection

**Configuration**:
- `src/sre_agent/config.py` - Agent configuration
- `k8s/kustomize/base/` - Kubernetes manifests

## 📚 Related Stories

- [WORKFLOW-001: PrometheusRule → Linear Issue Creation](./BVL-65-WORKFLOW-001-prometheus-to-linear-with-slm.md)
- [WORKFLOW-002: Lambda Function Annotation Discovery](./BVL-66-WORKFLOW-002-lambda-annotation-discovery.md)
- [WORKFLOW-003: Enriched Issue Updates](./BVL-67-WORKFLOW-003-enriched-issue-updates.md)
- [BACKEND-001: CloudEvents Processing](./BVL-59-BACKEND-001-cloudevents-processing.md)

## 🔗 References

- [CloudEvents Specification](https://cloudevents.io/)
- [Linear API Documentation](https://developers.linear.app/docs)
- [Knative Lambda Operator](../knative-lambda-operator/README.md)

---

**Test File**: `tests/test_end_to_end_workflow.py`  
**Owner**: Principal Manager Engineer  
**Last Updated**: 2026-01-08


