# ⚡ SRE-001: Build Failure Investigation

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-219/sre-001-build-failure-investigation
**Priority**: P0
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-168/sre-001-build-failure-investigation  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to quickly identify and resolve function build failures  
**So that** developers can deploy their functions without delays

---


## 🎯 Acceptance Criteria

- [ ] [ ] MTTR (Mean Time To Resolution) <30min for common failures
- [ ] [ ] Root cause identified within 10min
- [ ] [ ] Automated alerting for failure rate >5%
- [ ] [ ] Runbook documentation for top 5 failure modes
- [ ] [ ] Post-mortem created for novel failures
- [ ] [ ] Prometheus metrics track failure categories
- [ ] [ ] Failed jobs cleaned up automatically after 24hrs
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] MTTR (Mean Time To Resolution) <30min for common failures
- [ ] Root cause identified within 10min
- [ ] Automated alerting for failure rate >5%
- [ ] Runbook documentation for top 5 failure modes
- [ ] Post-mortem created for novel failures
- [ ] Prometheus metrics track failure categories
- [ ] Failed jobs cleaned up automatically after 24hrs

---

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                    BUILD FAILURE INVESTIGATION                       │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: ALERT FIRES                                               │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Prometheus Alert: BuildFailureRateHigh              │            │
│  │  Severity: critical                                  │            │
│  │  Threshold: >10% failures in 5min window             │            │
│  │  Current: 15% (3/20 builds failed)                   │            │
│  │                                                      │            │
│  │  Alert sent to:                                      │            │
│  │  • PagerDuty → On-call SRE (SMS/push)                │            │
│  │  • Slack #knative-lambda-alerts                      │            │
│  │  • Grafana dashboard (red panel)                     │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=2min: INITIAL TRIAGE                                          │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  SRE acknowledges alert and starts investigation     │            │
│  │                                                      │            │
│  │  Step 1: Check RabbitMQ Queue Health                 │            │
│  │  $ make rabbitmq-status ENV=prd                      │            │
│  │                                                      │            │
│  │  Output:                                             │            │
│  │  Queue Name | Messages | Consumers      │            │
│  │  ------------------------ | ---------- | --------------- │            │
│  │  lambda-build-events-prd | 45 | 2              │            │
│  │  lambda-service-events | 12 | 2              │            │
│  │  parser-results | 0        | 1              │            │
│  │                                                      │            │
│  │  ✅ Queue healthy: No backlog, consumers active      │            │
│  │                                                      │            │
│  │  Step 2: List Recent Failed Jobs                     │            │
│  │  $ kubectl get jobs -n knative-lambda \              │            │
│  │      --field-selector status.successful=0 \          │            │
│  │      --sort-by='.status.startTime'                   │            │
│  │                                                      │            │
│  │  Output:                                             │            │
│  │  NAME                           AGE    STATUS        │            │
│  │  build-parser-abc123-1729xxx    5m     Failed (0/3)  │            │
│  │  build-parser-def456-1729xxx    3m     Failed (0/3)  │            │
│  │  build-parser-ghi789-1729xxx    1m     Failed (0/3)  │            │
│  │                                                      │            │
│  │  🔍 Pattern detected: All failures from same parser? │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=5min: ROOT CAUSE ANALYSIS                                     │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Step 3: Inspect First Failed Job                    │            │
│  │  $ kubectl logs job/build-parser-abc123-1729xxx \    │            │
│  │      -n knative-lambda -c kaniko --tail=50           │            │
│  │                                                      │            │
│  │  Output (Kaniko logs):                               │            │
│  │  ───────────────────────────────────────             │            │
│  │  time="2025-10-29T10:45:32Z" level=info              │            │
│  │    msg="Retrieving image manifest python:3.9-slim"   │            │
│  │  time="2025-10-29T10:45:33Z" level=info              │            │
│  │    msg="Building Dockerfile from S3 context"         │            │
│  │  time="2025-10-29T10:45:34Z" level=error             │            │
│  │    error="failed to download from S3"                │            │
│  │    path="s3://knative-lambda-.../parser/xyz"         │            │
│  │    status=403                                        │            │
│  │  ERROR: build failed: S3 access denied               │            │
│  │  ───────────────────────────────────────             │            │
│  │                                                      │            │
│  │  🎯 ROOT CAUSE IDENTIFIED:                           │            │
│  │  S3 Access Denied (HTTP 403)                         │            │
│  │  • Service account IAM role missing S3 permissions   │            │
│  │  • Or: S3 bucket policy changed                      │            │
│  │  • Or: Parser file deleted from S3                   │            │
│  │                                                      │            │
│  │  Step 4: Verify S3 Access                            │            │
│  │  $ aws s3 ls s3://knative-lambda-fusion-.../ \       │            │
│  │      --profile prd                                   │            │
│  │                                                      │            │
│  │  Output:                                             │            │
│  │  An error occurred (AccessDenied) when calling...    │            │
│  │  Access Denied                                       │            │
│  │                                                      │            │
│  │  ❌ Confirmed: ServiceAccount IAM role broken        │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=10min: RESOLUTION                                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Step 5: Check IAM Policy                            │            │
│  │  $ aws iam get-role-policy \                         │            │
│  │      --role-name knative-lambda-builder-prd \        │            │
│  │      --policy-name S3Access                          │            │
│  │                                                      │            │
│  │  Finding: Policy exists but bucket ARN mismatch      │            │
│  │  Expected: arn:aws:s3:::knative-lambda-*             │            │
│  │  Actual:   arn:aws:s3:::knative-lambda-*             │            │
│  │                                                      │            │
│  │  🔧 FIX: Update Helm values.yaml                     │            │
│  │  builderService:                                     │            │
│  │    serviceAccount:                                   │            │
│  │      annotations:                                    │            │
│  │        eks.amazonaws.com/role-arn: \                 │            │
│  │          arn:aws:iam::339954290315:role/\            │            │
│  │          knative-lambda-builder-prd  # Fixed!        │            │
│  │                                                      │            │
│  │  Step 6: Apply Fix via GitOps                        │            │
│  │  $ git commit -m "fix: correct IAM role for prd"     │            │
│  │  $ git push origin main                              │            │
│  │                                                      │            │
│  │  Step 7: Wait for Flux to apply (30s)                │            │
│  │  $ kubectl rollout status deployment/\               │            │
│  │      knative-lambda-builder -n knative-lambda        │            │
│  │                                                      │            │
│  │  ✅ Deployment rolled out successfully               │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=15min: VALIDATION                                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Step 8: Trigger Test Build                          │            │
│  │  $ make trigger-build-prd                            │            │
│  │                                                      │            │
│  │  Output:                                             │            │
│  │  CloudEvent sent:                                    │            │
│  │    type: network.notifi.lambda.build.start           │            │
│  │    correlation_id: test-fix-1729xxx                  │            │
│  │                                                      │            │
│  │  Step 9: Monitor Job                                 │            │
│  │  $ kubectl get jobs -n knative-lambda -w             │            │
│  │                                                      │            │
│  │  NAME                      STATUS        DURATION    │            │
│  │  build-test-fix-1729xxx    Running       30s         │            │
│  │  build-test-fix-1729xxx    Completed     65s         │            │
│  │                                                      │            │
│  │  ✅ Build succeeded! Resolution confirmed            │            │
│  │                                                      │            │
│  │  Step 10: Clean Up Failed Jobs                       │            │
│  │  $ kubectl delete jobs -n knative-lambda \           │            │
│  │      --field-selector status.successful=0            │            │
│  │                                                      │            │
│  │  ✅ 3 failed jobs deleted                            │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=20min: POST-INCIDENT                                          │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Step 11: Verify Metrics                             │            │
│  │  • build_failures_total: decreasing                  │            │
│  │  • build_success_rate: back to 95%+                  │            │
│  │  • queue_depth: normal (<100)                        │            │
│  │                                                      │            │
│  │  Step 12: Update Runbook                             │            │
│  │  • Add "IAM role ARN mismatch" to common failures    │            │
│  │  • Document verification steps                       │            │
│  │                                                      │            │
│  │  Step 13: Create Post-Mortem (if P0)                 │            │
│  │  • Timeline of events                                │            │
│  │  • Root cause                                        │            │
│  │  • Action items to prevent recurrence                │            │
│  │                                                      │            │
│  │  ✅ INCIDENT RESOLVED                                │            │
│  │  Total Time: 20 minutes                              │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔍 Common Failure Modes

### 1. S3 Access Denied (HTTP 403)

**Symptoms**:
- Kaniko logs show `failed to download from S3: status=403`
- All builds fail for specific parser

**Root Causes**:
- ServiceAccount IAM role missing S3:GetObject permission
- S3 bucket policy changed
- Parser file deleted/moved

**Resolution**:
```bash
# Verify IAM role
aws iam get-role-policy \
  --role-name knative-lambda-builder-${ENV} \
  --policy-name S3Access

# Verify parser file exists
aws s3 ls s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}

# Fix: Update IAM policy or restore parser file
```

**Prevention**:
- Use IAM policy condition to require specific tags
- Enable S3 versioning for parser files
- Add pre-flight check before build

---

### 2. ECR Push Rate Limit (HTTP 429)

**Symptoms**:
- Kaniko logs show `error pushing image: Too Many Requests`
- Failures during high-traffic periods

**Root Causes**:
- ECR has rate limits: 10 pushes/sec per repository
- Multiple builds pushing concurrently

**Resolution**:
```bash
# Check ECR metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECR \
  --metric-name ThrottledRequestCount \
  --dimensions Name=RepositoryName,Value=knative-lambdas/builder

# Temporary fix: Space out builds
kubectl scale deployment/knative-lambda-builder --replicas=2 -n knative-lambda

# Permanent fix: Implement retry with exponential backoff
# (already in code via resilience package)
```

**Prevention**:
- Use ECR repositories per tenant (sharding)
- Implement build queue rate limiting
- Enable Kaniko push retry

---

### 3. OOMKilled (Out of Memory)

**Symptoms**:
- Pod exit code 137
- Kaniko container OOMKilled
- Large parser files or dependencies

**Root Causes**:
- Kaniko memory limit too low (default 1Gi)
- Large base images (>1GB)
- Many layers in Dockerfile

**Resolution**:
```bash
# Check pod resource usage
kubectl top pods -n knative-lambda \
  -l job-name=build-parser-${PARSER_ID}

# Increase memory limit in values.yaml
builderService:
  kaniko:
    resources:
      limits:
        memory: 2Gi  # was 1Gi
```

**Prevention**:
- Use slim base images (alpine, slim)
- Implement multi-stage builds
- Add memory limit alerts

---

### 4. Invalid Dockerfile Generation

**Symptoms**:
- Kaniko logs show `error building image: failed to parse Dockerfile`
- Syntax errors in generated Dockerfile

**Root Causes**:
- Template rendering bug
- Invalid parser metadata
- Special characters in environment variables

**Resolution**:
```bash
# Extract generated Dockerfile
kubectl logs job/build-parser-${PARSER_ID} \
  -n knative-lambda \
  -c build-context-manager \ | grep "Generated Dockerfile"

# Validate Dockerfile locally
docker build -f /tmp/Dockerfile .

# Fix template in internal/templates/templates.go
```

**Prevention**:
- Add Dockerfile validation before Kaniko
- Unit test template rendering
- Sanitize inputs

---

### 5. Timeout (Build >5min)

**Symptoms**:
- Job status `DeadlineExceeded`
- No error in logs, just timeout

**Root Causes**:
- Large image build
- Slow S3 download
- Kaniko resource starvation

**Resolution**:
```bash
# Increase job timeout
builderService:
  buildTimeout: 600s  # was 300s

# Check S3 download speed
kubectl logs job/build-parser-${PARSER_ID} \
  -c build-context-manager \ | grep "Download speed"

# Enable Kaniko cache
builderService:
  kaniko:
    cache: true
    cacheRepo: ${ECR_REGISTRY}/kaniko-cache
```

**Prevention**:
- Enable layer caching
- Use faster S3 endpoints
- Optimize Dockerfiles

---

### 6. Image Pull Failures

**Symptoms**:
- Kaniko logs show `error pulling image`
- `ImagePullBackOff` or `ErrImagePull` status
- Build fails at base image pull step

**Root Causes**:
- Base image doesn't exist or wrong tag
- ECR rate limiting on pulls
- Network connectivity issues
- Authentication failure for private registries

**Resolution**:
```bash
# Check if image exists
docker pull python:3.9-slim

# Check ECR authentication
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin ${ECR_REGISTRY}

# Verify network connectivity
kubectl exec -it kaniko-pod -n knative-lambda -- \
  curl -I https://gcr.io/v2/

# Check image pull secrets
kubectl get secret -n knative-lambda | grep docker
```

**Prevention**:
- Pin base image versions (avoid `latest`)
- Mirror frequently-used base images to ECR
- Implement image pull retry logic
- Pre-warm node image cache

---

### 7. Dependency Installation Failures

**Symptoms**:
- Kaniko logs show `npm install failed` or `pip install failed`
- Build fails during RUN step
- Dependency resolution errors

**Root Causes**:
- Package repository unreachable (npmjs.com, pypi.org)
- Invalid package versions in requirements
- Network timeouts during installation
- Disk space exhausted

**Resolution**:
```bash
# Check Kaniko logs for specific dependency error
kubectl logs job/build-parser-${PARSER_ID} \
  -n knative-lambda -c kaniko | grep -A10 "RUN"

# Verify parser requirements file
aws s3 cp s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID} - | \
  tar xzOf - package.json

# Test dependency installation locally
docker run --rm python:3.9-slim pip install -r requirements.txt
```

**Prevention**:
- Use dependency lock files (package-lock.json, requirements.txt with hashes)
- Mirror critical dependencies to private registry
- Add retry logic for transient network failures
- Validate dependencies before triggering build

---

## 📈 Metrics & Alerting

### Key Metrics

```promql
# Build failure rate (5min window)
rate(build_failures_total[5m]) / rate(builds_total[5m]) > 0.10

# Build duration p95
histogram_quantile(0.95, 
  rate(build_duration_seconds_bucket[5m])
) > 90

# Failed jobs count
count(kube_job_status_failed{namespace="knative-lambda"} == 1)
```

### Alert Thresholds | Alert | Threshold | Severity | Action | |------- | ----------- | ---------- | -------- | | BuildFailureRateHigh | >10% in 5min | critical | Immediate investigation | | BuildSlow | p95 >90s | warning | Check Kaniko resources | | FailedJobsAccumulating | >10 failed jobs | warning | Clean up + investigate | | KanikoOOMKilled | exit code 137 | critical | Increase memory limits | ---

## 🧪 Test Scenarios

### Scenario 1: S3 Access Denied

**Given**: IAM role missing S3 permissions  
**When**: Build triggered  
**Then**:
- Kaniko logs show "Access Denied"
- Job fails within 30s
- Alert fires
- SRE identifies IAM issue within 5min
- Fix applied via GitOps
- Build succeeds after retry

### Scenario 2: ECR Rate Limit

**Given**: 15 concurrent builds  
**When**: All push to ECR simultaneously  
**Then**:
- Some builds fail with HTTP 429
- Retry logic kicks in
- Builds complete with 30-60s delay
- No manual intervention needed

### Scenario 3: Invalid Parser File

**Given**: Parser file has syntax error  
**When**: Build triggered  
**Then**:
- Kaniko fails at RUN step
- Logs show Python syntax error
- SRE contacts developer
- Developer fixes parser
- Rebuild succeeds

---

## 📚 Runbook Quick Reference

```bash
# 1. Check failure rate
make pf-prometheus
# Query: rate(build_failures_total[5m])

# 2. List failed jobs
kubectl get jobs -n knative-lambda \
  --field-selector status.successful=0

# 3. Get logs
kubectl logs job/<job-name> -n knative-lambda -c kaniko --tail=100

# 4. Check S3 access
aws s3 ls s3://knative-lambda-fusion-modules-tmp/global/parser/

# 5. Verify IAM role
kubectl describe serviceaccount knative-lambda-builder -n knative-lambda

# 6. Clean up failed jobs
kubectl delete jobs -n knative-lambda \
  --field-selector status.successful=0

# 7. Trigger test build
make trigger-build-${ENV} PARSER_ID=test
```

---

## 💡 Pro Tips

- **Always check RabbitMQ first**: Queue backlog causes cascading failures
- **Correlation IDs are key**: Trace events end-to-end with `correlation_id`
- **Enable verbose logging temporarily**: `LOG_LEVEL=debug` for debugging
- **Keep runbook updated**: Add new failure modes as you encounter them
- **Automate common fixes**: Script IAM checks, S3 validation

---

