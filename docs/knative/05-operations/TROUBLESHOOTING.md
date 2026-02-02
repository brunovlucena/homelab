# 🔧 Troubleshooting Guide

**Common issues and solutions for Knative Lambda operations**

---

## 📑 Quick Navigation

- [Build Issues](#build-issues)
- [Deployment Issues](#deployment-issues)
- [Scaling Issues](#scaling-issues)
- [Network Issues](#network-issues)
- [Performance Issues](#performance-issues)
- [RabbitMQ Issues](#rabbitmq-issues)

---

## 🏗️ Build Issues

### Build Job Stuck in Pending

**Symptom**: Kaniko job shows `Pending` status for >2 minutes

**Diagnosis**:
```bash
# Check job status
kubectl get jobs -n knative-lambda

# Describe job to see events
kubectl describe job/kaniko-<parser-id> -n knative-lambda

# Common causes shown in events:
# - Insufficient resources
# - Image pull errors
# - Node selector mismatch
```

**Solutions**:

**1. Insufficient Resources**
```bash
# Check node resources
kubectl top nodes

# If nodes at capacity, scale cluster:
# - AWS: Increase ASG desired count
# - GCP: gcloud container clusters resize
```

**2. Image Pull BackOff**
```bash
# Verify ECR credentials
aws ecr get-login-password --region us-west-2

# Check image exists
aws ecr describe-images --repository-name knative-lambda-functions --region us-west-2

# Recreate imagePullSecret if needed
kubectl create secret docker-registry ecr-credentials \
  --docker-server=339954290315.dkr.ecr.us-west-2.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password --region us-west-2) \
  -n knative-lambda
```

---

### Build Fails with S3 Access Denied

**Symptom**: Init container fails with `403 Forbidden` or `Access Denied`

**Diagnosis**:
```bash
# Check init container logs
kubectl logs job/kaniko-<parser-id> -c fetch-code -n knative-lambda

# Error message:
# fatal error: An error occurred (AccessDenied) when calling the GetObject operation
```

**Solutions**:

**1. Verify IRSA Role**
```bash
# Check service account annotation
kubectl get sa kaniko-builder -n knative-lambda -o yaml | grep eks.amazonaws.com/role-arn

# Should show:
# eks.amazonaws.com/role-arn: arn:aws:iam::123456789:role/knative-lambda-builder
```

**2. Verify IAM Policy**
```bash
# Check IAM role has S3 permissions
aws iam get-role-policy --role-name knative-lambda-builder --policy-name S3Access

# Policy should include:
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:ListBucket"],
      "Resource": [
        "arn:aws:s3:::knative-lambda-*",
        "arn:aws:s3:::knative-lambda-*/*"
      ]
    }
  ]
}
```

**3. Verify S3 Bucket Policy**
```bash
# Check bucket policy
aws s3api get-bucket-policy --bucket knative-lambda-fusion-modules-tmp

# Ensure no DENY rules blocking access
```

---

### Build Fails with ECR Push Denied

**Symptom**: Kaniko container fails during image push

**Diagnosis**:
```bash
# Check Kaniko logs
kubectl logs job/kaniko-<parser-id> -c kaniko -n knative-lambda

# Error message:
# error pushing image: denied: User: arn:aws:sts::123456789:assumed-role/... 
# is not authorized to perform: ecr:PutImage
```

**Solutions**:

**1. Verify ECR Permissions**
```bash
# Add ECR permissions to IAM role
aws iam put-role-policy --role-name knative-lambda-builder \
  --policy-name ECRAccess \
  --policy-document '{
    "Statement": [{
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload",
        "ecr:PutImage"
      ],
      "Resource": "*"
    }]
  }'
```

**2. Verify ECR Repository Exists**
```bash
# Create repository if missing
aws ecr create-repository \
  --repository-name knative-lambda-functions \
  --region us-west-2
```

---

### Build Succeeds but Function Crashes

**Symptom**: Build completes but deployed function crashes on startup

**Diagnosis**:
```bash
# Check function pod logs
kubectl logs -l serving.knative.dev/service=<parser-id> -n knative-lambda

# Common errors:
# - ModuleNotFoundError: Missing dependencies
# - SyntaxError: Code syntax errors
# - ImportError: Import path issues
```

**Solutions**:

**1. Missing Dependencies**
```python
# Verify requirements.txt includes all deps
# Test locally first:
pip install -r requirements.txt
python parser.py
```

**2. Syntax Errors**
```bash
# Lint code before uploading
python -m py_compile parser.py

# Or use flake8
flake8 parser.py
```

**3. Environment Variables**
```bash
# Check if function needs env vars
kubectl get ksvc <parser-id> -n knative-lambda -o yaml | grep -A 10 "env:"

# Add env vars via values.yaml:
lambdaDefaults:
  env:
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: api-secrets
          key: key
```

---

## ☁️ Deployment Issues

### Knative Service Not Ready

**Symptom**: `kubectl get ksvc` shows `Ready: False`

**Diagnosis**:
```bash
# Get service details
kubectl get ksvc <parser-id> -n knative-lambda -o yaml

# Check conditions
kubectl get ksvc <parser-id> -n knative-lambda -o jsonpath='{.status.conditions}'

# Common conditions:
# - ConfigurationsReady: False (image pull issue)
# - RoutesReady: False (networking issue)
```

**Solutions**:

**1. ConfigurationsReady: False**
```bash
# Check revision status
kubectl get revisions -n knative-lambda

# Describe failing revision
kubectl describe revision <revision-name> -n knative-lambda

# Fix image pull issues (see Build Issues above)
```

**2. RoutesReady: False**
```bash
# Check Knative Serving components
kubectl get pods -n knative-serving

# Check service connectivity (internal)
kubectl get service <parser-id> -n knative-lambda

# Verify internal DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup <parser-id>.knative-lambda.svc.cluster.local
```

---

### Function Times Out on First Request (Cold Start)

**Symptom**: First request takes >30 seconds and times out

**Diagnosis**:
```bash
# Check Knative autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler

# Check activator logs
kubectl logs -n knative-serving -l app=activator
```

**Solutions**:

**1. Increase Cold Start Timeout**
```yaml
# In Knative Service spec:
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/scale-up-delay: "60s"  # Increase timeout
```

**2. Keep Warm with Min Scale**
```yaml
# Keep 1 pod always running:
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "1"
```

**3. Optimize Function Startup**
```python
# Reduce import time
# Bad:
import pandas as pd  # Heavy import

# Good:
def handler(event):
    import pandas as pd  # Lazy import only when needed
```

---

## 📈 Scaling Issues

### Function Not Scaling Up Under Load

**Symptom**: High latency but pods not increasing

**Diagnosis**:
```bash
# Check current pod count
kubectl get pods -l serving.knative.dev/service=<parser-id> -n knative-lambda

# Check KPA (Knative Pod Autoscaler)
kubectl get kpa -n knative-lambda

# Check metrics
kubectl get --raw /apis/metrics.k8s.io/v1beta1/namespaces/knative-lambda/pods
```

**Solutions**:

**1. Increase Max Scale**
```yaml
autoscaling.knative.dev/max-scale: "50"  # From default 10
```

**2. Adjust Target Concurrency**
```yaml
autoscaling.knative.dev/target: "50"  # From default 100
# Lower target = more aggressive scaling
```

**3. Enable HPA instead of KPA**
```yaml
autoscaling.knative.dev/class: "hpa.autoscaling.knative.dev"
autoscaling.knative.dev/metric: "cpu"
autoscaling.knative.dev/target: "70"  # 70% CPU
```

---

### Function Not Scaling to Zero

**Symptom**: Pods remain running when idle

**Diagnosis**:
```bash
# Check if min-scale is set
kubectl get ksvc <parser-id> -n knative-lambda -o yaml | grep min-scale

# Check for active connections
kubectl get pods -l serving.knative.dev/service=<parser-id> -n knative-lambda
```

**Solutions**:

**1. Verify Scale-to-Zero Config**
```yaml
autoscaling.knative.dev/min-scale: "0"  # Must be 0
autoscaling.knative.dev/scale-down-delay: "30s"  # Wait time before scaling down
```

**2. Check Knative Global Config**
```bash
kubectl get cm config-autoscaler -n knative-serving -o yaml | grep scale-to-zero

# Should be: enable-scale-to-zero: "true"
```

---

## 🌐 Network Issues

### Cannot Access Function URL

**Symptom**: `curl <function-url>` returns timeout or connection refused

**Diagnosis**:
```bash
# Verify service exists
kubectl get ksvc <parser-id> -n knative-lambda

# Check route
kubectl get routes -n knative-lambda

# Test internal connectivity
kubectl run test-pod --rm -it --image=curlimages/curl -- \
  curl http://<parser-id>.knative-lambda.svc.cluster.local
```

**Solutions**:

**1. Network Policy Blocking**
```bash
# Check network policies
kubectl get networkpolicies -n knative-lambda

# Temporarily delete to test
kubectl delete networkpolicy <policy-name> -n knative-lambda
```

**2. Internal Networking Issue**
```bash
# Check Knative Serving networking
kubectl get pods -n knative-serving | grep activator

# Restart Knative Serving controller (if needed)
kubectl rollout restart deployment/controller -n knative-serving
```

**3. Port-Forward for Debug**
```bash
# Direct access to pod
kubectl port-forward deployment/<parser-id> 8080:8080 -n knative-lambda

# Test locally
curl http://localhost:8080
```

---

## ⚡ Performance Issues

### High Latency

**Symptom**: P95 latency >500ms

**Diagnosis**:
```bash
# Check Prometheus metrics
# Query: histogram_quantile(0.95, rate(serving_revision_request_latencies_bucket[5m]))

# Check if scaling is issue
kubectl top pods -l serving.knative.dev/service=<parser-id> -n knative-lambda

# Check for throttling
kubectl describe pod <pod-name> -n knative-lambda | grep -i throttl
```

**Solutions**:

**1. Increase Resources**
```yaml
resources:
  requests:
    memory: "512Mi"  # From 128Mi
    cpu: "500m"      # From 100m
```

**2. Optimize Code**
```python
# Use connection pooling
import requests
from requests.adapters import HTTPAdapter

session = requests.Session()
adapter = HTTPAdapter(pool_connections=10, pool_maxsize=20)
session.mount('https://', adapter)
```

**3. Add Caching**
```python
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_operation(key):
    # Cached result
    return result
```

---

## 🐰 RabbitMQ Issues

### Events Not Being Processed

**Symptom**: CloudEvents sent but builds not triggering

**Diagnosis**:
```bash
# Check RabbitMQ queues
make rabbitmq-status ENV=dev

# Or use management UI
kubectl port-forward svc/rabbitmq-cluster-dev 15672:15672 -n rabbitmq-dev
# Open: http://localhost:15672 (guest/guest)

# Check builder service logs
kubectl logs -f deployment/knative-lambda-builder -n knative-lambda | grep -i "rabbitmq"
```

**Solutions**:

**1. Reconnect Builder Service**
```bash
# Restart builder
kubectl rollout restart deployment/knative-lambda-builder -n knative-lambda
```

**2. Purge Stuck Messages**
```bash
# Purge queue (use with caution!)
make rabbitmq-purge-lambda-queues-dev
```

**3. Check RabbitMQ Cluster Health**
```bash
# Check cluster status
kubectl exec -it rabbitmq-cluster-dev-0 -n rabbitmq-dev -- rabbitmq-diagnostics cluster_status

# Check for alarms
kubectl exec -it rabbitmq-cluster-dev-0 -n rabbitmq-dev -- rabbitmq-diagnostics alarms
```

---

## 🔍 Debugging Tools

### Essential Commands

```bash
# Get all resources for parser
kubectl get all -l parser-id=<parser-id> -n knative-lambda

# Stream logs from multiple pods
stern <parser-id> -n knative-lambda

# Watch events
kubectl get events -n knative-lambda --sort-by='.lastTimestamp' -w

# Port-forward for debugging
make pf-rabbitmq-admin  # RabbitMQ UI
make pf-prometheus       # Metrics
```

### Enable Debug Logging

```yaml
# values-dev.yaml
builderService:
  env:
    - name: LOG_LEVEL
      value: "debug"  # From "info"
```

---

## 📞 Escalation

### When to Escalate

Escalate to Platform Team if:
- ❌ Multiple functions failing (platform-wide issue)
- ❌ RabbitMQ cluster degraded
- ❌ Knative Serving not responding
- ❌ Security incident detected

### How to Escalate

1. **Collect logs**:
```bash
# Builder logs
kubectl logs deployment/knative-lambda-builder -n knative-lambda --tail=500 > builder.log

# Kaniko logs
kubectl logs job/kaniko-<parser-id> -n knative-lambda > kaniko.log

# Function logs
kubectl logs -l serving.knative.dev/service=<parser-id> -n knative-lambda > function.log
```

2. **Create GitHub Issue**: Include logs, symptoms, timeline

3. **Contact**: `#knative-lambda-incidents` Slack channel

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

