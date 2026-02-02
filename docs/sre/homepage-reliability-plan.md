# Homepage Reliability Plan

**Author:** Senior SRE Engineer  
**Date:** December 2025  
**Status:** Draft  
**Service:** lucena.cloud (Homepage)

---

## Executive Summary

This document outlines a comprehensive reliability engineering plan for the Homepage service (`lucena.cloud`). The analysis identifies critical gaps in the current architecture and provides actionable improvements across availability, observability, and incident response.

**Current State:** 🔴 Unreliable  
**Target State:** 🟢 Production-Ready (99.9% availability)

---

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         lucena.cloud                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐             │
│  │  Cloudflare  │────▶│   Frontend   │────▶│     API      │             │
│  │    Tunnel    │     │   (nginx)    │     │  (Go/Gin)    │             │
│  └──────────────┘     └──────────────┘     └──────┬───────┘             │
│                                                    │                     │
│                              ┌────────────────────┼────────────────┐    │
│                              │                    │                │    │
│                              ▼                    ▼                ▼    │
│                       ┌──────────┐         ┌──────────┐     ┌────────┐ │
│                       │ PostgreSQL│         │  Redis   │     │ Ollama │ │
│                       │  (Data)   │         │ (Cache)  │     │ (LLM)  │ │
│                       └──────────┘         └──────────┘     └────────┘ │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Current Component Status

| Component | Replicas | HA Ready | PDB | HPA | Probes | Notes |
|-----------|----------|----------|-----|-----|--------|-------|
| Frontend | 1 | ❌ | ❌ | ❌ | ✅ | Single point of failure |
| API | 1 | ❌ | ❌ | ❌ | ✅ | Single point of failure |
| PostgreSQL | 1 | ❌ | ❌ | ❌ | ❌ | **Critical: No HA, no backups** |
| Redis | 1 | ❌ | ❌ | ❌ | ✅ | Cache loss acceptable |

---

## 2. Service Level Objectives (SLOs)

### 2.1 Proposed SLOs

| SLI | Target SLO | Measurement | Alert Threshold |
|-----|------------|-------------|-----------------|
| **Availability** | 99.9% | Successful responses / Total requests | < 99.5% over 5m |
| **Latency (p50)** | < 100ms | Response time percentile | > 150ms over 5m |
| **Latency (p99)** | < 500ms | Response time percentile | > 750ms over 5m |
| **Error Rate** | < 0.1% | 5xx errors / Total requests | > 0.5% over 5m |
| **Database Availability** | 99.95% | Successful queries / Total queries | < 99.9% over 1m |

### 2.2 Error Budget

```
Monthly Error Budget (99.9% SLO):
- Total minutes in month: 43,200
- Allowed downtime: 43.2 minutes
- Current consumption: Unknown (no tracking)
```

---

## 3. Critical Issues Identified

### 🔴 P0 - Critical (Fix Immediately)

#### 3.1 PostgreSQL Single Point of Failure

**Problem:** PostgreSQL runs as a single replica with no replication, backups, or failover capability. The recent WAL corruption incident proves this is a critical risk.

**Impact:** Complete data loss and extended downtime on any PostgreSQL failure.

**Solution:**
```yaml
# Option A: CloudNativePG Operator (Recommended for homelab)
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: homepage-postgres
  namespace: postgres
spec:
  instances: 2  # Primary + 1 standby
  primaryUpdateStrategy: unsupervised
  
  storage:
    size: 10Gi
    storageClass: local-path
    
  backup:
    barmanObjectStore:
      destinationPath: s3://homepage-backups/postgres
      endpointURL: http://minio.minio:9000
      s3Credentials:
        accessKeyId:
          name: minio-credentials
          key: access-key
        secretAccessKey:
          name: minio-credentials
          key: secret-key
      wal:
        compression: gzip
    retentionPolicy: "7d"
    
  monitoring:
    enablePodMonitor: true
```

#### 3.2 No Pod Disruption Budgets

**Problem:** Rolling updates or node drains can take down all replicas simultaneously.

**Solution:**
```yaml
# flux/infrastructure/homepage/k8s/kustomize/base/pdb.yaml
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: homepage-api-pdb
  namespace: homepage
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: homepage
      app.kubernetes.io/component: api
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: homepage-frontend-pdb
  namespace: homepage
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: homepage
      app.kubernetes.io/component: frontend
```

#### 3.3 Resource Requests Not Set

**Problem:** CPU requests are commented out, preventing proper scheduling and QoS guarantees.

**Current:**
```yaml
resources:
  limits:
    # cpu: 500m  # COMMENTED OUT
    memory: 512Mi
  # requests:    # COMMENTED OUT
  #   cpu: 250m
```

**Solution:**
```yaml
resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### 🟠 P1 - High (Fix This Week)

#### 3.4 No Horizontal Pod Autoscaler

**Problem:** Manual scaling only; cannot respond to traffic spikes.

**Solution:**
```yaml
# flux/infrastructure/homepage/k8s/kustomize/base/hpa.yaml
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: homepage-api-hpa
  namespace: homepage
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: homepage-api
  minReplicas: 2
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: homepage-frontend-hpa
  namespace: homepage
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: homepage-frontend
  minReplicas: 2
  maxReplicas: 4
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

#### 3.5 Health Check Improvements

**Problem:** Current health check returns 200 even if dependencies (DB, Redis) are down.

**Current:**
```go
func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",  // Always returns healthy!
    })
}
```

**Solution:**
```go
func healthCheck(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // Check database
    dbHealthy := db.PingContext(ctx) == nil
    
    // Check Redis
    redisHealthy := redisClient.Ping(ctx).Err() == nil
    
    status := "healthy"
    httpStatus := http.StatusOK
    
    if !dbHealthy || !redisHealthy {
        status = "degraded"
        httpStatus = http.StatusServiceUnavailable
    }
    
    c.JSON(httpStatus, gin.H{
        "status":    status,
        "timestamp": time.Now().UTC(),
        "checks": gin.H{
            "database": dbHealthy,
            "redis":    redisHealthy,
        },
    })
}

// Separate liveness (am I alive?) from readiness (can I serve traffic?)
func livenessCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func readinessCheck(c *gin.Context) {
    // Full dependency check for readiness
    healthCheck(c)
}
```

#### 3.6 No Database Connection Pooling Limits

**Problem:** Connection pool may exhaust under load (25 max connections is low).

**Current:**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
```

**Solution:**
```go
db.SetMaxOpenConns(50)           // Increase for higher concurrency
db.SetMaxIdleConns(25)           // Keep more warm connections
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(2 * time.Minute)  // Add idle timeout
```

### 🟡 P2 - Medium (Fix This Month)

#### 3.7 No Rate Limiting

**Problem:** API vulnerable to abuse and DoS.

**Solution:**
```go
import "github.com/gin-contrib/ratelimit"
import "golang.org/x/time/rate"

// Add rate limiting middleware
router.Use(ratelimit.RateLimiter(
    ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
        Rate:  rate.Every(time.Second),
        Limit: 100,  // 100 requests per second per IP
    }),
    &ratelimit.Options{
        KeyFunc: func(c *gin.Context) string {
            return c.ClientIP()
        },
    },
))
```

#### 3.8 No Circuit Breaker for LLM Service

**Problem:** LLM (Ollama) failures cascade to API timeouts.

**Solution:**
```go
import "github.com/sony/gobreaker"

var llmBreaker *gobreaker.CircuitBreaker

func init() {
    llmBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "llm-service",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    })
}

func callLLM(prompt string) (string, error) {
    result, err := llmBreaker.Execute(func() (interface{}, error) {
        return llmService.Generate(prompt)
    })
    if err != nil {
        return "", err
    }
    return result.(string), nil
}
```

#### 3.9 No Graceful Shutdown

**Problem:** In-flight requests may be dropped during deployments.

**Solution:**
```go
func main() {
    // ... setup code ...
    
    srv := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }

    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // Give outstanding requests 30 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exited")
}
```

---

## 4. Observability Improvements

### 4.1 ServiceMonitor for Prometheus

```yaml
# flux/infrastructure/homepage/k8s/kustomize/base/servicemonitor.yaml
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: homepage-api
  namespace: homepage
  labels:
    app.kubernetes.io/name: homepage
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: homepage
      app.kubernetes.io/component: api
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
  namespaceSelector:
    matchNames:
      - homepage
```

### 4.2 PrometheusRule for Alerts

```yaml
# flux/infrastructure/homepage/k8s/kustomize/base/prometheusrule.yaml
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: homepage-alerts
  namespace: homepage
spec:
  groups:
    - name: homepage.availability
      rules:
        - alert: HomepageAPIDown
          expr: |
            up{job="homepage-api"} == 0
          for: 1m
          labels:
            severity: critical
            service: homepage
          annotations:
            summary: "Homepage API is down"
            description: "Homepage API has been unreachable for more than 1 minute."

        - alert: HomepageHighErrorRate
          expr: |
            sum(rate(http_requests_total{job="homepage-api", status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total{job="homepage-api"}[5m])) > 0.01
          for: 5m
          labels:
            severity: warning
            service: homepage
          annotations:
            summary: "Homepage API high error rate"
            description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes."

        - alert: HomepageHighLatency
          expr: |
            histogram_quantile(0.99, 
              sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le)
            ) > 0.5
          for: 5m
          labels:
            severity: warning
            service: homepage
          annotations:
            summary: "Homepage API high latency"
            description: "P99 latency is {{ $value | humanizeDuration }}."

        - alert: HomepagePostgresDown
          expr: |
            pg_up{job="postgres"} == 0
          for: 1m
          labels:
            severity: critical
            service: homepage
          annotations:
            summary: "PostgreSQL is down"
            description: "PostgreSQL database is unreachable."

        - alert: HomepagePodCrashLooping
          expr: |
            increase(kube_pod_container_status_restarts_total{namespace="homepage"}[1h]) > 3
          labels:
            severity: warning
            service: homepage
          annotations:
            summary: "Homepage pod is crash looping"
            description: "Pod {{ $labels.pod }} has restarted {{ $value }} times in the last hour."
```

### 4.3 Structured Logging

```go
import "go.uber.org/zap"

var logger *zap.Logger

func initLogger() {
    var err error
    if getEnv("GIN_MODE", "release") == "debug" {
        logger, err = zap.NewDevelopment()
    } else {
        logger, err = zap.NewProduction()
    }
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
}

// Use structured logging
logger.Info("request completed",
    zap.String("method", c.Request.Method),
    zap.String("path", c.Request.URL.Path),
    zap.Int("status", c.Writer.Status()),
    zap.Duration("latency", latency),
    zap.String("client_ip", c.ClientIP()),
)
```

---

## 5. Disaster Recovery

### 5.1 Backup Strategy

| Component | Backup Method | Frequency | Retention | RTO | RPO |
|-----------|---------------|-----------|-----------|-----|-----|
| PostgreSQL | pg_dump to MinIO | Hourly | 7 days | 30 min | 1 hour |
| PostgreSQL WAL | Streaming to MinIO | Continuous | 24 hours | 15 min | 5 min |
| Redis | RDB snapshots | 6 hours | 3 days | N/A | N/A |

### 5.2 Backup CronJob

```yaml
# flux/infrastructure/homepage/k8s/kustomize/base/backup-cronjob.yaml
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: homepage-postgres-backup
  namespace: homepage
spec:
  schedule: "0 * * * *"  # Every hour
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: backup
              image: localhost:5001/postgres:16-alpine
              command:
                - /bin/sh
                - -c
                - |
                  set -e
                  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
                  BACKUP_FILE="/tmp/homepage_${TIMESTAMP}.sql.gz"
                  
                  echo "Starting backup at ${TIMESTAMP}..."
                  pg_dump -h postgres.postgres -U postgres homepage | gzip > "${BACKUP_FILE}"
                  
                  echo "Uploading to MinIO..."
                  mc alias set minio http://minio.minio:9000 ${MINIO_ACCESS_KEY} ${MINIO_SECRET_KEY}
                  mc cp "${BACKUP_FILE}" minio/homepage-backups/postgres/
                  
                  echo "Cleaning up old backups (keeping last 168 = 7 days)..."
                  mc ls minio/homepage-backups/postgres/ | sort | head -n -168 | awk '{print $NF}' | \
                    xargs -I {} mc rm minio/homepage-backups/postgres/{}
                  
                  echo "Backup completed successfully"
              env:
                - name: PGPASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: homepage
                      key: homepage-postgres-password
                - name: MINIO_ACCESS_KEY
                  valueFrom:
                    secretKeyRef:
                      name: minio-credentials
                      key: access-key
                - name: MINIO_SECRET_KEY
                  valueFrom:
                    secretKeyRef:
                      name: minio-credentials
                      key: secret-key
```

---

## 6. Implementation Roadmap

### Phase 1: Stabilization (Week 1)

| Task | Priority | Effort | Owner |
|------|----------|--------|-------|
| Fix PostgreSQL WAL corruption | P0 | 1h | SRE |
| Enable CPU resource requests | P0 | 30m | SRE |
| Add PodDisruptionBudgets | P0 | 30m | SRE |
| Increase API replicas to 2 | P1 | 15m | SRE |
| Fix health check to verify dependencies | P1 | 2h | Backend |

### Phase 2: Observability (Week 2)

| Task | Priority | Effort | Owner |
|------|----------|--------|-------|
| Deploy ServiceMonitor | P1 | 1h | SRE |
| Configure PrometheusRules | P1 | 2h | SRE |
| Add structured logging (zap) | P2 | 4h | Backend |
| Create Grafana dashboard | P2 | 3h | SRE |

### Phase 3: Resilience (Week 3-4)

| Task | Priority | Effort | Owner |
|------|----------|--------|-------|
| Implement HPA | P1 | 2h | SRE |
| Add circuit breaker for LLM | P2 | 3h | Backend |
| Implement rate limiting | P2 | 2h | Backend |
| Add graceful shutdown | P2 | 2h | Backend |
| Configure PostgreSQL backups | P1 | 4h | SRE |

### Phase 4: High Availability (Month 2)

| Task | Priority | Effort | Owner |
|------|----------|--------|-------|
| Deploy CloudNativePG | P1 | 8h | SRE |
| Configure WAL archiving | P1 | 4h | SRE |
| Add Redis Sentinel/Cluster | P3 | 8h | SRE |
| Document runbooks | P2 | 8h | SRE |

---

## 7. Runbooks

### 7.1 PostgreSQL WAL Corruption Recovery

```bash
# 1. Check current state
kubectl get pods -n postgres
kubectl logs -n postgres -l app=postgres

# 2. If PANIC on checkpoint, reset WAL (DATA LOSS WARNING)
kubectl delete pvc postgres-pvc -n postgres
kubectl delete pod -n postgres -l app=postgres

# 3. Wait for new pod to start
kubectl wait --for=condition=Ready pod -l app=postgres -n postgres --timeout=120s

# 4. Trigger db-init job to recreate schema
kubectl create job --from=cronjob/homepage-db-init homepage-db-init-recovery -n homepage

# 5. Verify
kubectl logs -n homepage -l component=db-init -f
```

### 7.2 API Not Responding

```bash
# 1. Check pod status
kubectl get pods -n homepage -l app.kubernetes.io/component=api

# 2. Check logs for errors
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=100

# 3. Check dependencies
kubectl exec -n homepage deploy/homepage-api -- sh -c "
  psql -h postgres.postgres -U postgres -d homepage -c 'SELECT 1' && echo 'DB OK'
  redis-cli -h redis.redis ping && echo 'Redis OK'
"

# 4. If dependency failure, check dependency pods
kubectl get pods -n postgres
kubectl get pods -n redis

# 5. If API is stuck, rolling restart
kubectl rollout restart deployment/homepage-api -n homepage
```

---

## 8. Success Metrics

Track these metrics weekly to measure reliability improvement:

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Availability | Unknown | 99.9% | Prometheus uptime |
| MTTR | Unknown | < 15 min | Incident tracking |
| MTBF | Unknown | > 7 days | Incident tracking |
| P99 Latency | Unknown | < 500ms | Prometheus histogram |
| Error Budget Consumed | Unknown | < 50% | SLO tracking |
| Backup Success Rate | 0% | 100% | CronJob monitoring |

---

## 9. Appendix

### A. Files to Create/Modify

```
flux/infrastructure/homepage/k8s/kustomize/base/
├── api-deployment.yaml          # UPDATE: resources, probes
├── frontend-deployment.yaml     # UPDATE: resources
├── pdb.yaml                     # NEW
├── hpa.yaml                     # NEW
├── servicemonitor.yaml          # NEW
├── prometheusrule.yaml          # NEW
├── backup-cronjob.yaml          # NEW
└── kustomization.yaml           # UPDATE: add new resources
```

### B. Go Code Changes

```
src/api/
├── main.go                      # UPDATE: health checks, graceful shutdown
├── middleware.go                # UPDATE: rate limiting
├── services/
│   └── llm_service.go          # UPDATE: circuit breaker
```

---

**Document Version:** 1.0  
**Last Updated:** December 5, 2025  
**Next Review:** January 5, 2026
