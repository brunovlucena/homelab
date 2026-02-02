# 🌐 BACKEND-003: Kubernetes Job Lifecycle Management

**Priority**: P1 | **Status**: ✅ Implemented K  | **Story Points**: 13  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-216/backend-003-kubernetes-job-lifecycle-management


---

## 📋 User Story

**As a** Backend Developer  
**I want to** manage Kubernetes Job lifecycles for Kaniko builds  
**So that** Docker images are built reliably with proper resource management and error handling

---

## 🎯 Acceptance Criteria

### ✅ Job Creation
- [ ] Create Kubernetes Job with Kaniko container
- [ ] Configure job with proper resource limits (CPU, memory)
- [ ] Set job timeout and backoff limits
- [ ] Include sidecar container for build monitoring
- [ ] Generate unique job names: `build-{third_party_id}-{parser_id}-{short_hash}`
- [ ] Apply proper labels for tracking and cleanup
- [ ] Configure TTL for automatic cleanup (default: 86400s)

### ✅ Concurrent Job Management
- [ ] Enforce maximum concurrent jobs limit (default: 10)
- [ ] Check active job count before creation
- [ ] Queue jobs when limit is reached
- [ ] Return clear error message when limit exceeded
- [ ] Support per-namespace job isolation

### ✅ Job Deduplication
- [ ] Check for existing jobs before creating new one
- [ ] Find jobs by `third_party_id` and `parser_id` labels
- [ ] Skip creation if active job exists for same parser
- [ ] Delete and recreate if previous job completed/failed
- [ ] Implement failure backoff period (default: 5 minutes)

### ✅ Job Status Tracking
- [ ] Monitor job status: Pending, Running, Succeeded, Failed
- [ ] Track job creation timestamp
- [ ] Track job completion timestamp
- [ ] Calculate job duration
- [ ] Emit CloudEvents on job completion
- [ ] Handle job timeouts gracefully

### ✅ Job Cleanup
- [ ] Delete completed jobs after TTL expires
- [ ] Clean up failed jobs immediately
- [ ] Delete associated pods on job deletion
- [ ] Handle orphaned resources
- [ ] Prevent resource accumulation

### ✅ Error Handling
- [ ] Handle job creation conflicts
- [ ] Handle insufficient cluster resources
- [ ] Handle job timeout scenarios
- [ ] Handle pod eviction/preemption
- [ ] Provide detailed error logs for debugging

---

## 🔧 Technical Implementation

### File: `internal/handler/job_manager.go`

```go
// Job Manager Interface
type JobManager interface {
    CreateJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error)
    FindExistingJob(ctx context.Context, thirdPartyID, parserID string) (*batchv1.Job, error)
    GetJob(ctx context.Context, jobName string) (*batchv1.Job, error)
    DeleteJob(ctx context.Context, jobName string) error
    IsJobRunning(job *batchv1.Job) bool
    IsJobFailed(job *batchv1.Job) bool
    IsJobSucceeded(job *batchv1.Job) bool
    CleanupFailedJob(ctx context.Context, jobName string) error
    CountActiveJobs(ctx context.Context) (int, error)
}

// Job Creation with Concurrency Control
func (j *JobManagerImpl) CreateJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error) {
    // 1. Check concurrent job limit
    activeJobCount, err := j.CountActiveJobs(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to count active jobs: %w", err)
    }
    
    if activeJobCount >= j.rateLimitConfig.MaxConcurrentJobs {
        return nil, fmt.Errorf("concurrent job limit reached: %d active jobs, max allowed: %d",
            activeJobCount, j.rateLimitConfig.MaxConcurrentJobs)
    }
    
    // 2. Delete existing job if present (KISS approach)
    existingJob, _ := j.FindExistingJob(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID)
    if existingJob != nil {
        j.DeleteJob(ctx, existingJob.Name)
    }
    
    // 3. Create new Kaniko job
    job, err := j.createKanikoJob(ctx, jobName, buildRequest)
    if err != nil {
        return nil, err
    }
    
    // 4. Apply job to Kubernetes
    createdJob, err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).Create(ctx, job, metav1.CreateOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to create job: %w", err)
    }
    
    j.obs.Info(ctx, "Job created successfully",
        "job_name", jobName,
        "job_uid", string(createdJob.UID))
    
    return createdJob, nil
}
```

### Kaniko Job Specification

```go
func (j *JobManagerImpl) createKanikoJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error) {
    // Generate image URI with content hash
    imageURI := fmt.Sprintf("%s/knative-lambdas:%s-%s-%s",
        j.awsConfig.ECRRegistry,
        buildRequest.ThirdPartyID,
        buildRequest.ParserID,
        buildRequest.ContentHash[:8])
    
    // Build context S3 location
    buildContextKey := fmt.Sprintf("build-context/%s/context.tar.gz", buildRequest.ParserID)
    contextURL := fmt.Sprintf("s3://%s/%s", j.awsConfig.S3TempBucket, buildContextKey)
    
    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      jobName,
            Namespace: j.config.Namespace,
            Labels: map[string]string{
                "app":                                 "knative-lambda-builder",
                "build.notifi.network/third-party-id": buildRequest.ThirdPartyID,
                "build.notifi.network/parser-id":      buildRequest.ParserID,
            },
        },
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: int32Ptr(86400), // 24 hours
            BackoffLimit:            int32Ptr(0),     // No retries
            ActiveDeadlineSeconds:   int64Ptr(1800),  // 30 minutes timeout
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    ServiceAccountName: "kaniko",
                    RestartPolicy:      corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        // Kaniko builder container
                        {
                            Name:  "kaniko",
                            Image: "gcr.io/kaniko-project/executor:latest",
                            Args: []string{
                                fmt.Sprintf("--context=%s", contextURL),
                                fmt.Sprintf("--destination=%s", imageURI),
                                "--dockerfile=Dockerfile",
                            },
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("2000m"),
                                    corev1.ResourceMemory: resource.MustParse("4Gi"),
                                },
                            },
                        },
                        // Sidecar monitoring container
                        {
                            Name:  "sidecar",
                            Image: "knative-lambda-sidecar:latest",
                            Env: []corev1.EnvVar{
                                {Name: "BUILD_JOB_NAME", Value: jobName},
                                {Name: "THIRD_PARTY_ID", Value: buildRequest.ThirdPartyID},
                                {Name: "PARSER_ID", Value: buildRequest.ParserID},
                                {Name: "IMAGE_URI", Value: imageURI},
                            },
                        },
                    },
                },
            },
        },
    }, nil
}
```

---

## 📊 Job Lifecycle States

```
┌─────────────┐
│   QUEUED    │ ← Job creation queued (async worker pool)
└──────┬──────┘
       │
       ↓
┌─────────────┐
│   PENDING   │ ← Job created in Kubernetes
└──────┬──────┘
       │
       ↓
┌─────────────┐
│   RUNNING   │ ← Pod scheduled and executing
└──────┬──────┘
       │
       ├─────────────┐
       │             │
       ↓             ↓
┌─────────────┐ ┌─────────────┐
│  SUCCEEDED  │ │   FAILED    │
└─────────────┘ └──────┬──────┘
                       │
                       ↓
                ┌─────────────┐
                │  CLEANUP    │ ← Failed job deleted
                └─────────────┘
```

---

## 🧪 Testing Scenarios

### 1. Successful Build Job
```bash
# Trigger build
make trigger-build-dev PARSER_ID=test-parser

# Monitor job
kubectl get jobs -n knative-lambda -w

# Check logs
kubectl logs -n knative-lambda job/build-customer-123-test-parser-abc123 -c kaniko
```

**Expected**:
- Job created with status `Running`
- Kaniko container builds image successfully
- Sidecar emits build.complete event
- Job transitions to `Succeeded`
- Image pushed to ECR

### 2. Concurrent Job Limit
```bash
# Trigger 11 builds simultaneously (limit is 10)
for i in {1..11}; do
  make trigger-build-dev PARSER_ID=parser-$i &
done
```

**Expected**:
- First 10 jobs created successfully
- 11th job returns error: "concurrent job limit reached"
- Jobs complete and allow new builds

### 3. Duplicate Build Prevention
```bash
# Trigger same build twice rapidly
make trigger-build-dev PARSER_ID=test-parser &
make trigger-build-dev PARSER_ID=test-parser &
```

**Expected**:
- First request creates job
- Second request finds existing job and skips creation
- Only one job runs for the parser

### 4. Failed Build Cleanup
```bash
# Trigger build with invalid source
make trigger-build-dev PARSER_ID=invalid-parser
```

**Expected**:
- Job starts but Kaniko fails
- Sidecar detects failure and emits build.failed event
- Failed job cleaned up automatically
- Clear error message in logs

---

## 📈 Performance Requirements

- **Job Creation**: < 2s from request to Kubernetes
- **Job Startup**: < 30s from creation to pod running
- **Build Duration**: < 10 minutes for typical parser
- **Cleanup Latency**: < 5s for failed job deletion
- **Concurrent Jobs**: Support up to 10 simultaneous builds

---

## 🔍 Monitoring & Alerts

### Metrics
- `k8s_job_creation_total{status="success | failure"}` - Job creations
- `k8s_job_duration_seconds{status="succeeded | failed"}` - Build duration
- `k8s_active_jobs` - Current active job count
- `k8s_job_failures_total{reason="timeout | oom | error"}` - Failure reasons

### Alerts
- **High Job Failure Rate**: Alert if > 10% failures over 15 minutes
- **Job Creation Errors**: Alert on any job creation API errors
- **Long Running Jobs**: Alert if job runs > 20 minutes
- **Job Limit Reached**: Alert when hitting concurrent job limit

---

## 🏗️ Code References

**Main Files**:
- `internal/handler/job_manager.go` - Job lifecycle management
- `internal/handler/async_job_creator.go` - Async job creation worker pool
- `internal/config/kubernetes.go` - K8s configuration
- `internal/config/rate_limiting.go` - Concurrency limits

**Sidecar**:
- `sidecar/main.go` - Build monitoring and event emission

---

## 📚 Related Documentation

- [BACKEND-002: Build Context Management](BACKEND-002-build-context-management.md)
- [BACKEND-004: Async Job Processing](BACKEND-004-async-job-processing.md)
- Kubernetes Jobs: https://kubernetes.io/docs/concepts/workloads/controllers/job/
- Kaniko: https://github.com/GoogleContainerTools/kaniko

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: ✅ Implemented K

