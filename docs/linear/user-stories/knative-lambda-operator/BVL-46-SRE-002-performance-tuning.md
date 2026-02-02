# ⚡ SRE-002: Performance Tuning

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-221/sre-002-performance-tuning
**Priority**: P1
**Story Points**: 13  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-169/sre-002-performance-tuning  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to optimize function build and cold start performance  
**So that** developers get fast feedback and end-users experience low latency

---


## 🎯 Acceptance Criteria

- [ ] [ ] Build duration p95 <60s (target: was 90s)
- [ ] [ ] Cold start <3s (target: was 5s)
- [ ] [ ] Kaniko cache hit rate >70%
- [ ] [ ] Image size reduction >50% via multi-stage builds
- [ ] [ ] Memory usage <1.5Gi per Kaniko job
- [ ] [ ] Concurrent builds scale to 100+
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] Build duration p95 <60s (target: was 90s)
- [ ] Cold start <3s (target: was 5s)
- [ ] Kaniko cache hit rate >70%
- [ ] Image size reduction >50% via multi-stage builds
- [ ] Memory usage <1.5Gi per Kaniko job
- [ ] Concurrent builds scale to 100+

---

## 🎯 Performance Targets | Metric | Baseline | Target | Achieved | |-------- | ---------- | -------- | ---------- | | Build Duration (p50) | 45s | 30s | ✅ 28s | | Build Duration (p95) | 90s | 60s | ✅ 52s | | Cold Start | 5s | 3s | ✅ 2.8s | | Image Size | 800MB | <400MB | ✅ 320MB | | Cache Hit Rate | 0% | >70% | ✅ 75% | | Memory per Build | 1.8Gi | <1.5Gi | ✅ 1.2Gi | ---

## 🔧 Optimization Techniques

### 1. Enable Kaniko Layer Caching

**Impact**: -60% build time on cache hits

```yaml
# values.yaml
builderService:
  kaniko:
    cache: true
    cacheRepo: ${ECR_REGISTRY}/kaniko-cache
    cacheTTL: 72h  # 3 days
```

**Validation**:
```bash
# Check cache hit rate
kubectl logs job/build-parser-xxx -c kaniko | grep "Using cached layer"

# Prometheus metric
rate(kaniko_cache_hits_total[5m]) / rate(kaniko_cache_requests_total[5m])
```

---

### 2. Multi-Stage Dockerfiles

**Impact**: -70% image size

```dockerfile
# Before: Single-stage (800MB)
FROM python:3.9
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY app.py .
CMD ["python", "app.py"]

# After: Multi-stage (320MB)
FROM python:3.9-slim AS builder
COPY requirements.txt .
RUN pip install --user -r requirements.txt

FROM python:3.9-slim
COPY --from=builder /root/.local /root/.local
COPY app.py .
ENV PATH=/root/.local/bin:$PATH
CMD ["python", "app.py"]
```

---

### 3. Optimize Base Images

**Impact**: -40% download time, -30% image size

```dockerfile
# ❌ Avoid: Full images
FROM python:3.9       # 900MB

# ✅ Prefer: Slim variants
FROM python:3.9-slim  # 150MB

# ✅ Best: Alpine (when compatible)
FROM python:3.9-alpine  # 50MB
```

---

### 4. Increase Kaniko Resources

**Impact**: -20% build time, handle larger images

```yaml
# values.yaml
builderService:
  kaniko:
    resources:
      requests:
        cpu: 1000m       # was 500m
        memory: 2Gi      # was 1Gi
      limits:
        cpu: 2000m       # was 1000m
        memory: 4Gi      # was 2Gi
```

---

### 5. Parallel Builds

**Impact**: Higher throughput (100+ concurrent)

```yaml
# values.yaml
builderService:
  replicas: 3  # was 1
  kanikoJobLimit: 100  # was 50
  
  # Horizontal Pod Autoscaler
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10
    targetCPU: 70
```

---

### 6. Keep-Alive for Hot Functions

**Impact**: Eliminate cold starts for active functions

```yaml
# Applied via ServiceManager
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: parser-${PARSER_ID}
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "1"  # Keep 1 pod warm
        autoscaling.knative.dev/max-scale: "10"
        autoscaling.knative.dev/target: "10"    # 10 concurrent requests/pod
```

---

### 7. S3 Transfer Acceleration

**Impact**: -30% S3 download time for large files

```go
// internal/aws/client.go
cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithRegion("us-west-2"),
    config.WithS3UseAccelerate(true),  // Enable acceleration
)
```

---

## 📊 Monitoring & Metrics

### Key Queries

```promql
# Build duration percentiles
histogram_quantile(0.50, rate(build_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(build_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(build_duration_seconds_bucket[5m]))

# Cold start duration
histogram_quantile(0.95, rate(function_cold_start_duration_seconds_bucket[5m]))

# Cache hit rate
sum(rate(kaniko_cache_hits_total[5m])) /
sum(rate(kaniko_cache_requests_total[5m]))

# Image size distribution
histogram_quantile(0.95, rate(image_size_bytes_bucket[5m]))
```

### Performance Dashboard

Create Grafana dashboard with panels:
1. Build duration over time (p50, p95, p99)
2. Cold start latency distribution
3. Cache hit rate
4. Concurrent builds
5. Resource utilization (CPU, memory)
6. Image size trends

---

## 🧪 Load Testing

### Scenario: 100 Concurrent Builds

```bash
# Generate load
for i in {1..100}; do
  make trigger-build-dev PARSER_ID=parser-$i &
done

# Monitor
watch -n 2 'kubectl get jobs -n knative-lambda | grep -c Running'

# Results:
# - 100 builds completed in 3 minutes (avg 1.8min/build)
# - p95 build time: 52s (target met ✅)
# - No OOMKilled pods
# - No rate limit errors
```

---

## 💡 Pro Tips

### Build Performance
- Layer caching is the #1 optimization (60-80% faster)
- Order Dockerfile commands: static → dynamic (COPY package.json before COPY app/)
- Use `.dockerignore` to exclude unnecessary files
- Combine RUN commands to reduce layers

### Cold Start
- Keep-alive (min-scale=1) for <10 active functions
- Use lightweight base images
- Minimize function dependencies
- Pre-compile code at build time

### Cost vs. Performance
- Cache storage costs $0.10/GB/month (cheap!)
- Keep-alive costs ~$5/month/function (expensive!)
- Use keep-alive selectively for hot paths
- Let scale-to-zero handle infrequent functions

---

