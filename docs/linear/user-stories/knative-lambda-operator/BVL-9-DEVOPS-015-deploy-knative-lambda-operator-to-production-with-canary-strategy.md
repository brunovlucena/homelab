# 🔄 DEVOPS-015: Deploy knative-lambda-operator to production with canary strategy

**Status**: Backlog  | **Priority**: P0**Linear URL**: https://linear.app/bvlucena/issue/BVL-9/migration-deploy-knative-lambda-operator-to-production-with-canary-strategy | **Status**: Backlog  | **Priority**: P0**Linear URL**: https://linear.app/bvlucena/issue/BVL-9/migration-deploy-knative-lambda-operator-to-production-with-canary-strategy | **Story Points**: 13

**Created**: 2025-12-26T14:36:52.267Z  
**Updated**: 2025-12-26T14:36:52.267Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Deploy knative-lambda-operator to Notifi production environment using canary deployment strategy (Flagger).


## 📋 User Story

**As a** DevOps Engineer  
**I want to** deploy knative-lambda-operator to production with canary strategy  
**So that** I can improve system reliability, security, and performance

---


## 📋 Prerequisites

* Operator validated in dev environment
* Production Helm values configured
* Flagger available in production cluster

## 🔧 Tasks

### 1\. Production Configuration

- [ ] Review and finalize production values.yaml
- [ ] Configure production replica counts (HA: 2+ replicas)
- [ ] Configure production resource limits
- [ ] Configure production image tags (no :latest)
- [ ] Configure canary deployment settings
- [ ] Configure A/B testing rules (if applicable)

### 2\. Flagger Canary Setup

- [ ] Create Flagger Canary resource for operator
- [ ] Configure canary strategy (step weight, interval, threshold)
- [ ] Configure Prometheus metrics for canary analysis
- [ ] Configure success/failure thresholds
- [ ] Test canary process in dev/staging first

### 3\. Production Deployment

- [ ] Deploy via ArgoCD or manual Helm install
- [ ] Monitor initial deployment
- [ ] Verify operator pods are running
- [ ] Verify CRDs are installed
- [ ] Verify operator functionality

### 4\. Canary Validation

- [ ] Trigger canary deployment process
- [ ] Monitor canary metrics
- [ ] Verify canary analysis succeeds
- [ ] Verify canary promotion process
- [ ] Test rollback procedure if needed

### 5\. Production Validation

- [ ] Test LambdaFunction creation in production
- [ ] Test LambdaAgent creation in production
- [ ] Verify production workload isolation
- [ ] Verify production resource constraints
- [ ] Verify production monitoring and alerting

### 6\. Alerting Configuration

- [ ] Configure PrometheusRule alerts for production
- [ ] Test alert delivery (Slack/PagerDuty/etc.)
- [ ] Verify alert thresholds are appropriate
- [ ] Document alert runbooks

### 7\. Documentation

- [ ] Update production deployment documentation
- [ ] Document canary deployment process
- [ ] Document rollback procedures
- [ ] Update runbooks
- [ ] Update incident response procedures

## ✅ Acceptance Criteria

- [ ] Operator deployed successfully in production
- [ ] Canary deployment process works correctly
- [ ] All production tests passing
- [ ] Monitoring and alerting configured
- [ ] Documentation complete
- [ ] Team trained on production deployment process

## 🔗 Dependencies

INFRA-MIG-004: Deploy and validate knative-lambda-operator in dev environment
Flagger must be available in production
