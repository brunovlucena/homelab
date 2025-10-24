// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔄 JOB MANAGER - Focused Kubernetes job lifecycle management
//
//	🎯 Purpose: Handle Kubernetes job creation, monitoring, and lifecycle operations
//	💡 Features: Job creation, status checking, cleanup, conflict resolution
//
//	🏛️ ARCHITECTURE:
//	🔄 Job Lifecycle - Create, monitor, update, delete Kubernetes jobs
//	📊 Status Management - Check job status, handle transitions
//	🧹 Cleanup Operations - Clean up failed jobs, manage TTL
//	🔍 Conflict Resolution - Handle concurrent job creation
//	🛡️ Kubernetes API Operations - Job management and lifecycle
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"fmt"
	"os"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/handler/helpers"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
	"knative-lambda-new/pkg/builds"
)

// 🔄 JobManagerImpl - "Focused Kubernetes job lifecycle management"
type JobManagerImpl struct {
	k8sClient       kubernetes.Interface
	config          *config.KubernetesConfig
	buildConfig     *config.BuildConfig
	awsConfig       *config.AWSConfig
	storageConfig   *config.StorageConfig
	rateLimitConfig *config.RateLimitingConfig
	obs             *observability.Observability
	// 🛡️ Rate Limiting Protection
	rateLimiter *resilience.MultiLevelRateLimiter
}

// 🔄 JobManagerConfig - "Configuration for creating job manager"
type JobManagerConfig struct {
	K8sClient       kubernetes.Interface
	K8sConfig       *config.KubernetesConfig
	BuildConfig     *config.BuildConfig
	AWSConfig       *config.AWSConfig
	StorageConfig   *config.StorageConfig
	RateLimitConfig *config.RateLimitingConfig
	Observability   *observability.Observability
	RateLimiter     *resilience.MultiLevelRateLimiter
}

// 🏗️ NewJobManager - "Create new job manager with dependencies"
func NewJobManager(config JobManagerConfig) (JobManager, error) {
	if config.K8sClient == nil {
		return nil, errors.NewConfigurationError("job_manager", "k8s_client", "kubernetes client cannot be nil")
	}

	if config.K8sConfig == nil {
		return nil, errors.NewConfigurationError("job_manager", "k8s_config", "kubernetes config cannot be nil")
	}

	if config.BuildConfig == nil {
		return nil, errors.NewConfigurationError("job_manager", "build_config", "build config cannot be nil")
	}

	if config.Observability == nil {
		return nil, errors.NewConfigurationError("job_manager", "observability", "observability cannot be nil")
	}

	return &JobManagerImpl{
		k8sClient:       config.K8sClient,
		config:          config.K8sConfig,
		buildConfig:     config.BuildConfig,
		awsConfig:       config.AWSConfig,
		storageConfig:   config.StorageConfig,
		rateLimitConfig: config.RateLimitConfig,
		obs:             config.Observability,
		rateLimiter:     config.RateLimiter,
	}, nil
}

// 🔄 CreateJob - "Create a new Kubernetes job (KISS approach)"
func (j *JobManagerImpl) CreateJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error) {
	ctx, span := j.obs.StartSpan(ctx, "create_job")
	defer span.End()

	j.obs.Info(ctx, "Starting Kubernetes job creation",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"build_type", buildRequest.BuildType,
		"runtime", buildRequest.Runtime)

	// 🚀 KISS: Delete existing job if it exists, then create new one
	existingJob, err := j.FindExistingJob(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID)
	if err != nil {
		j.obs.Info(ctx, "Failed to check for existing job, continuing with creation",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	} else if existingJob != nil {
		j.obs.Info(ctx, "Deleting existing job before creating new one",
			"existing_job_name", existingJob.Name,
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)

		// Delete the existing job
		if deleteErr := j.DeleteJob(ctx, existingJob.Name); deleteErr != nil {
			j.obs.Error(ctx, deleteErr, "Failed to delete existing job, continuing anyway",
				"existing_job_name", existingJob.Name,
				"job_name", jobName,
				"third_party_id", buildRequest.ThirdPartyID,
				"parser_id", buildRequest.ParserID,
				"correlation_id", buildRequest.CorrelationID)
		}
	}

	// Create Kaniko job from template
	j.obs.Info(ctx, "Creating Kaniko job from template",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	job, err := j.createKanikoJob(ctx, jobName, buildRequest)
	if err != nil {
		j.obs.Error(ctx, err, "Failed to create Kaniko job from template",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return nil, errors.NewSystemError("kubernetes", "create_kaniko_job")
	}

	j.obs.Info(ctx, "Successfully created Kaniko job from template",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"job_uid", string(job.UID))

	// Validate job spec before returning
	if len(job.Spec.Template.Spec.Containers) == 0 {
		j.obs.Error(ctx, fmt.Errorf("job spec validation failed: no containers"), "Job spec validation failed - no containers",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return nil, errors.NewSystemError("kubernetes", "invalid_job_spec")
	}

	j.obs.Info(ctx, "Job spec validation passed",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"container_count", len(job.Spec.Template.Spec.Containers))

	// Check concurrent job limit before creating the job
	activeJobCount, err := j.CountActiveJobs(ctx)
	if err != nil {
		j.obs.Error(ctx, err, "Failed to count active jobs",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return nil, fmt.Errorf("failed to count active jobs: %w", err)
	}

	maxConcurrentJobs := j.rateLimitConfig.MaxConcurrentJobs
	if activeJobCount >= maxConcurrentJobs {
		j.obs.Info(ctx, "Concurrent job limit reached, blocking new job creation",
			"job_name", jobName,
			"active_job_count", activeJobCount,
			"max_concurrent_jobs", maxConcurrentJobs,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return nil, fmt.Errorf("concurrent job limit reached: %d active jobs, max allowed: %d", activeJobCount, maxConcurrentJobs)
	}

	j.obs.Info(ctx, "Concurrent job limit check passed",
		"job_name", jobName,
		"active_job_count", activeJobCount,
		"max_concurrent_jobs", maxConcurrentJobs,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Ensure ECR repository exists before creating the job
	if j.awsConfig != nil {
		j.obs.Info(ctx, "ECR repository should be created manually or by another process",
			"repository_name", "knative-lambda",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	}

	// Apply the job to Kubernetes
	j.obs.Info(ctx, "Applying job to Kubernetes cluster",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"namespace", j.config.Namespace)

	// 🚀 KISS: Simple job creation - if it fails, it fails
	createdJob, err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		// Record K8s job creation failure metric
		if j.obs.GetMetrics() != nil {
			metricsRec := observability.NewMetricsRecorder(j.obs)
			metricsRec.RecordK8sJobCreation(ctx, "kaniko", "failure")
		}

		j.obs.Error(ctx, err, "Failed to apply job to Kubernetes cluster",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"namespace", j.config.Namespace,
			"error_details", err.Error())
		return nil, fmt.Errorf("failed to create job in Kubernetes: %w", err)
	}

	j.obs.Info(ctx, "Successfully created Kubernetes job",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"namespace", j.config.Namespace,
		"job_uid", string(createdJob.UID),
		"job_creation_timestamp", createdJob.CreationTimestamp.String())

	// Record K8s job creation metric
	if j.obs.GetMetrics() != nil {
		metricsRec := observability.NewMetricsRecorder(j.obs)
		metricsRec.RecordK8sJobCreation(ctx, "kaniko", "success")
	}

	return createdJob, nil
}

// 🔍 FindExistingJob - "Find existing job for the same third party ID and parser ID"
func (j *JobManagerImpl) FindExistingJob(ctx context.Context, thirdPartyID, parserID string) (*batchv1.Job, error) {
	ctx, span := j.obs.StartSpan(ctx, "find_existing_job")
	defer span.End()

	// Sanitize IDs for label matching
	sanitizedThirdPartyID := helpers.SanitizeLabelValue(thirdPartyID)
	sanitizedParserID := helpers.SanitizeLabelValue(parserID)

	// Create label selector for jobs with the same parser
	labelSelector := fmt.Sprintf("build.notifi.network/third-party-id=%s,build.notifi.network/parser-id=%s",
		sanitizedThirdPartyID, sanitizedParserID)

	// List jobs directly with Kubernetes client
	jobs, err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		j.obs.Error(ctx, err, "Failed to list jobs",
			"third_party_id", thirdPartyID,
			"parser_id", parserID,
			"label_selector", labelSelector,
			"namespace", j.config.Namespace)
		return nil, fmt.Errorf("job listing failed: %w", err)
	}

	// Return the most recent job if any exist
	if len(jobs.Items) > 0 {
		// Sort by creation timestamp and return the most recent
		var latestJob *batchv1.Job
		for i := range jobs.Items {
			if latestJob == nil || jobs.Items[i].CreationTimestamp.After(latestJob.CreationTimestamp.Time) {
				latestJob = &jobs.Items[i]
			}
		}
		return latestJob, nil
	}

	return nil, nil
}

// 📋 GetJob - "Retrieve a job by name"
func (j *JobManagerImpl) GetJob(ctx context.Context, jobName string) (*batchv1.Job, error) {
	ctx, span := j.obs.StartSpan(ctx, "get_job")
	defer span.End()

	// Get job directly with Kubernetes client
	job, err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		j.obs.Error(ctx, err, "Failed to get job",
			"job_name", jobName,
			"namespace", j.config.Namespace)
		return nil, fmt.Errorf("job retrieval failed: %w", err)
	}

	return job, nil
}

// 🗑️ DeleteJob - "Delete a job by name"
func (j *JobManagerImpl) DeleteJob(ctx context.Context, jobName string) error {
	ctx, span := j.obs.StartSpan(ctx, "delete_job")
	defer span.End()

	// Delete job directly with Kubernetes client
	err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil {
		j.obs.Error(ctx, err, "Failed to delete job",
			"job_name", jobName,
			"namespace", j.config.Namespace)
		return fmt.Errorf("job deletion failed: %w", err)
	}

	j.obs.Info(ctx, "Job deleted successfully",
		"job_name", jobName,
		"namespace", j.config.Namespace)

	return nil
}

// 🔄 IsJobRunning - "Check if a job is currently running"
func (j *JobManagerImpl) IsJobRunning(job *batchv1.Job) bool {
	if job == nil {
		return false
	}
	return job.Status.Active > 0
}

// ❌ IsJobFailed - "Check if a job has failed"
func (j *JobManagerImpl) IsJobFailed(job *batchv1.Job) bool {
	if job == nil {
		return false
	}
	return job.Status.Failed > 0 && job.Status.Succeeded == 0
}

// ✅ IsJobSucceeded - "Check if a job has succeeded"
func (j *JobManagerImpl) IsJobSucceeded(job *batchv1.Job) bool {
	if job == nil {
		return false
	}
	return job.Status.Succeeded > 0
}

// 🔧 GenerateJobName - "Generate unique job name"
func (j *JobManagerImpl) GenerateJobName(thirdPartyID, parserID string) string {
	return helpers.GenerateJobName(thirdPartyID, parserID)
}

// 🧹 CleanupFailedJob - "Clean up failed job"
func (j *JobManagerImpl) CleanupFailedJob(ctx context.Context, jobName string) error {
	ctx, span := j.obs.StartSpan(ctx, "cleanup_failed_job")
	defer span.End()

	j.obs.Info(ctx, "Cleaning up failed job", "job_name", jobName)

	// Delete job directly with Kubernetes client
	err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil {
		// Check if it's a "not found" error (job already deleted)
		if apierrors.IsNotFound(err) {
			j.obs.Info(ctx, "Failed job already deleted", "job_name", jobName)
			return nil // Not an error if already deleted
		}
		j.obs.Error(ctx, err, "Failed to delete failed job", "job_name", jobName)
		return fmt.Errorf("job deletion failed: %w", err)
	}

	j.obs.Info(ctx, "Successfully cleaned up failed job", "job_name", jobName)
	return nil
}

// 🔧 HasFailedJobs - "Check if there are any failed jobs in the namespace"
func (j *JobManagerImpl) HasFailedJobs(ctx context.Context) (bool, error) {
	ctx, span := j.obs.StartSpan(ctx, "check_failed_jobs")
	defer span.End()

	// List all jobs in the namespace
	jobs, err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", constants.AppLabelName, constants.AppLabelValue),
	})
	if err != nil {
		j.obs.Error(ctx, err, "Failed to list jobs")
		return false, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Check for any failed jobs
	for _, job := range jobs.Items {
		if job.Status.Failed > 0 {
			j.obs.Info(ctx, "Found failed job",
				"job_name", job.Name,
				"failed_count", job.Status.Failed,
				"active_count", job.Status.Active,
				"succeeded_count", job.Status.Succeeded)
			return true, nil
		}
	}

	return false, nil
}

// 🔧 CountActiveJobs - "Count the number of active jobs in the namespace"
func (j *JobManagerImpl) CountActiveJobs(ctx context.Context) (int, error) {
	ctx, span := j.obs.StartSpan(ctx, "count_active_jobs")
	defer span.End()

	// List all jobs in the namespace
	jobs, err := j.k8sClient.BatchV1().Jobs(j.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", constants.AppLabelName, constants.AppLabelValue),
	})
	if err != nil {
		j.obs.Error(ctx, err, "Failed to list jobs")
		return 0, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Count active jobs (only truly running jobs)
	activeCount := 0
	for _, job := range jobs.Items {
		if j.IsJobRunning(&job) {
			activeCount++
			j.obs.Info(ctx, "Found active job",
				"job_name", job.Name,
				"active_count", job.Status.Active,
				"succeeded_count", job.Status.Succeeded,
				"failed_count", job.Status.Failed)
		}
	}

	j.obs.Info(ctx, "Counted active jobs",
		"active_job_count", activeCount,
		"total_jobs", len(jobs.Items))

	return activeCount, nil
}

// 🔧 createKanikoJob - "Create a Kubernetes Job from Kaniko template"
func (j *JobManagerImpl) createKanikoJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error) {
	j.obs.Info(ctx, "Creating Kaniko job", "job_name", jobName)

	// Create job metadata
	objectMeta := j.createJobMetadata(jobName, buildRequest)

	// Create job spec
	jobSpec, err := j.createJobSpec(ctx, jobName, buildRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create job spec: %w", err)
	}

	job := &batchv1.Job{
		ObjectMeta: objectMeta,
		Spec:       *jobSpec,
	}

	j.obs.Info(ctx, "Kaniko job spec created successfully", "job_name", job.Name)
	return job, nil
}

// createJobMetadata creates the job metadata with labels
func (j *JobManagerImpl) createJobMetadata(jobName string, buildRequest *builds.BuildRequest) metav1.ObjectMeta {
	sanitizedThirdPartyID := helpers.SanitizeLabelValue(buildRequest.ThirdPartyID)
	sanitizedParserID := helpers.SanitizeLabelValue(buildRequest.ParserID)

	return metav1.ObjectMeta{
		Name:      jobName,
		Namespace: j.config.Namespace,
		Labels: map[string]string{
			constants.AppLabelName:           constants.AppLabelValue,
			constants.BuildThirdPartyIDLabel: sanitizedThirdPartyID,
			constants.BuildParserIDLabel:     sanitizedParserID,
		},
	}
}

// createJobSpec creates the job specification
func (j *JobManagerImpl) createJobSpec(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.JobSpec, error) {
	// Parse resource requirements
	resourceRequirements, err := j.parseResourceRequirements()
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource requirements: %w", err)
	}

	// Create pod template spec
	podTemplateSpec, err := j.createPodTemplateSpec(ctx, jobName, buildRequest, resourceRequirements)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod template spec: %w", err)
	}

	// Helper functions for creating pointers for K8s spec
	int32Ptr := func(i int32) *int32 { return &i }
	int64Ptr := func(i int64) *int64 { return &i }

	return &batchv1.JobSpec{
		TTLSecondsAfterFinished: int32Ptr(int32(j.config.JobTTLSeconds)),
		BackoffLimit:            int32Ptr(constants.JobBackoffLimit),
		ActiveDeadlineSeconds:   int64Ptr(int64(j.buildConfig.BuildTimeout.Seconds())),
		Template:                *podTemplateSpec,
	}, nil
}

// parseResourceRequirements parses CPU and memory resource requirements
func (j *JobManagerImpl) parseResourceRequirements() (corev1.ResourceRequirements, error) {
	cpuLimit, err := resource.ParseQuantity(j.buildConfig.CPULimit)
	if err != nil {
		return corev1.ResourceRequirements{}, errors.NewConfigurationErrorWithValue("kubernetes", "cpu_limit", j.buildConfig.CPULimit, fmt.Sprintf("failed to parse CPU limit: %v", err))
	}

	memoryLimit, err := resource.ParseQuantity(j.buildConfig.MemoryLimit)
	if err != nil {
		return corev1.ResourceRequirements{}, errors.NewConfigurationErrorWithValue("kubernetes", "memory_limit", j.buildConfig.MemoryLimit, fmt.Sprintf("failed to parse Memory limit: %v", err))
	}

	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		},
	}, nil
}

// createPodTemplateSpec creates the pod template specification
func (j *JobManagerImpl) createPodTemplateSpec(ctx context.Context, jobName string, buildRequest *builds.BuildRequest, resourceRequirements corev1.ResourceRequirements) (*corev1.PodTemplateSpec, error) {
	sanitizedThirdPartyID := helpers.SanitizeLabelValue(buildRequest.ThirdPartyID)
	sanitizedParserID := helpers.SanitizeLabelValue(buildRequest.ParserID)

	// Create containers
	containers, err := j.createContainers(ctx, jobName, buildRequest, resourceRequirements)
	if err != nil {
		return nil, fmt.Errorf("failed to create containers: %w", err)
	}

	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.AppLabelName:           constants.AppLabelValue,
				constants.BuildThirdPartyIDLabel: sanitizedThirdPartyID,
				constants.BuildParserIDLabel:     sanitizedParserID,
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: j.config.ServiceAccount,
			RestartPolicy:      corev1.RestartPolicyNever,
			Containers:         containers,
		},
	}, nil
}

// createContainers creates the kaniko and sidecar containers
func (j *JobManagerImpl) createContainers(_ context.Context, jobName string, buildRequest *builds.BuildRequest, resourceRequirements corev1.ResourceRequirements) ([]corev1.Container, error) {
	destinationImageURI := j.generateImageURI(buildRequest.ThirdPartyID, buildRequest.ParserID, buildRequest.ContentHash)

	kanikoContainer := j.createKanikoContainer(buildRequest, destinationImageURI, resourceRequirements)
	sidecarContainer := j.createSidecarContainer(jobName, buildRequest, destinationImageURI)

	return []corev1.Container{kanikoContainer, sidecarContainer}, nil
}

// createKanikoContainer creates the kaniko container specification
func (j *JobManagerImpl) createKanikoContainer(buildRequest *builds.BuildRequest, destinationImageURI string, resourceRequirements corev1.ResourceRequirements) corev1.Container {
	baseEnv := []corev1.EnvVar{
		{Name: "AWS_REGION", Value: j.awsConfig.AWSRegion},
		// Registry Configuration
		{Name: "REGISTRY_MIRROR", Value: j.awsConfig.RegistryMirror},
		{Name: "SKIP_TLS_VERIFY_REGISTRY", Value: j.awsConfig.SkipTLSVerifyRegistry},
		// Enhanced network configuration for npm install reliability
		{Name: "NODE_OPTIONS", Value: "--max-old-space-size=4096"},
		{Name: "npm_config_registry", Value: "https://registry.npmjs.org/"},
		{Name: "npm_config_timeout", Value: constants.NpmConfigTimeoutDefault},
		{Name: "npm_config_fetch_retries", Value: constants.NpmConfigFetchRetriesDefault},
		{Name: "npm_config_fetch_retry_mintimeout", Value: constants.NpmConfigFetchRetryMinTimeoutDefault},
		{Name: "npm_config_fetch_retry_maxtimeout", Value: constants.NpmConfigFetchRetryMaxTimeoutDefault},
		{Name: "npm_config_fetch_retry_factor", Value: constants.NpmConfigFetchRetryFactorDefault},
		{Name: "npm_config_cache", Value: "/tmp/.npm"},
		{Name: "npm_config_prefer_offline", Value: "false"},
		{Name: "npm_config_audit", Value: "false"},
		{Name: "npm_config_fund", Value: "false"},
		{Name: "npm_config_update_notifier", Value: "false"},
		{Name: "npm_config_loglevel", Value: "warn"},
		{Name: "npm_config_maxsockets", Value: constants.NpmConfigMaxSocketsDefault},
		{Name: "npm_config_strict_ssl", Value: "true"},
		{Name: "npm_config_user_agent", Value: "npm/kaniko-builder"},
		// Clear potentially problematic SSL settings
		{Name: "npm_config_ca", Value: ""},
		{Name: "npm_config_cafile", Value: ""},
	}

	// 🏠 Add MinIO-specific environment variables if MinIO is configured
	if j.storageConfig.IsMinIO() {
		minioScheme := "http"
		if j.storageConfig.MinIO.UseSSL {
			minioScheme = "https"
		}
		baseEnv = append(baseEnv, []corev1.EnvVar{
			{Name: "AWS_ACCESS_KEY_ID", Value: j.storageConfig.MinIO.AccessKey},
			{Name: "AWS_SECRET_ACCESS_KEY", Value: j.storageConfig.MinIO.SecretKey},
			{Name: "S3_ENDPOINT", Value: fmt.Sprintf("%s://%s", minioScheme, j.storageConfig.MinIO.Endpoint)},
			{Name: "S3_FORCE_PATH_STYLE", Value: "true"},  // MinIO requires path-style addressing
			{Name: "AWS_SDK_LOAD_CONFIG", Value: "false"}, // Disable AWS config loading for MinIO
		}...)
	}

	return corev1.Container{
		Name:      "kaniko",
		Image:     j.buildConfig.KanikoImage,
		Args:      j.createKanikoArgs(buildRequest, destinationImageURI),
		Env:       baseEnv,
		Resources: resourceRequirements,
	}
}

// createKanikoArgs creates the kaniko command arguments
func (j *JobManagerImpl) createKanikoArgs(buildRequest *builds.BuildRequest, destinationImageURI string) []string {
	// Generate build context key using the same pattern as BuildContextManager
	buildContextKey := fmt.Sprintf("build-context/%s/context.tar.gz", buildRequest.ParserID)

	// Get temp bucket from storage configuration (supports both S3 and MinIO)
	tempBucket := j.storageConfig.GetTempBucket()

	return []string{
		fmt.Sprintf("--context=s3://%s/%s", tempBucket, buildContextKey),
		fmt.Sprintf("--destination=%s", destinationImageURI),
		fmt.Sprintf("--dockerfile=%s", constants.DefaultDockerfilePath),
		fmt.Sprintf("--registry-mirror=%s", j.awsConfig.RegistryMirror),
		fmt.Sprintf("--skip-tls-verify-registry=%s", j.awsConfig.SkipTLSVerifyRegistry),
		fmt.Sprintf("--build-arg=NODE_BASE_IMAGE=%s", j.awsConfig.NodeBaseImage),
		fmt.Sprintf("--build-arg=PYTHON_BASE_IMAGE=%s", j.awsConfig.PythonBaseImage),
		fmt.Sprintf("--build-arg=GO_BASE_IMAGE=%s", j.awsConfig.GoBaseImage),
	}
}

// createSidecarContainer creates the sidecar container specification
func (j *JobManagerImpl) createSidecarContainer(jobName string, buildRequest *builds.BuildRequest, destinationImageURI string) corev1.Container {
	return corev1.Container{
		Name:  "sidecar",
		Image: j.buildConfig.SidecarImage,
		Env:   j.createSidecarEnvVars(jobName, buildRequest, destinationImageURI),
	}
}

// createSidecarEnvVars creates the sidecar environment variables
func (j *JobManagerImpl) createSidecarEnvVars(jobName string, buildRequest *builds.BuildRequest, destinationImageURI string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "KANIKO_NAMESPACE", Value: j.config.Namespace},
		{
			Name: "KANIKO_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
		{Name: "KANIKO_CONTAINER_NAME", Value: "kaniko"},
		{Name: "BUILD_JOB_NAME", Value: jobName},
		{Name: "IMAGE_URI", Value: destinationImageURI},
		{Name: "THIRD_PARTY_ID", Value: buildRequest.ThirdPartyID},
		{Name: "PARSER_ID", Value: buildRequest.ParserID},
		{Name: "CONTENT_HASH", Value: buildRequest.ContentHash}, // New: pass content hash to sidecar
		{Name: "CORRELATION_ID", Value: buildRequest.CorrelationID},
		{Name: "BUILD_TIMEOUT", Value: j.buildConfig.BuildTimeout.String()},
		{Name: "MONITOR_INTERVAL", Value: constants.MonitorIntervalDefault},
		{Name: "KNATIVE_BROKER_URL", Value: j.getBrokerURL()},
	}
}

// 🔧 getBrokerURL - "Get broker URL from environment or generate dynamically"
func (j *JobManagerImpl) getBrokerURL() string {
	// Try to get from environment variable first (set by Helm template)
	if brokerURL := os.Getenv("SIDECAR_BROKER_URL"); brokerURL != "" {
		return brokerURL
	}

	// Fallback to dynamic generation with environment-specific broker name
	brokerName := fmt.Sprintf("knative-lambda-builder-broker-%s", j.getEnvironment())
	return fmt.Sprintf("http://%s-broker-ingress.%s.svc.cluster.local", brokerName, j.config.Namespace)
}

// 🔧 getEnvironment - "Get environment from namespace or default"
func (j *JobManagerImpl) getEnvironment() string {
	// Extract environment from namespace (e.g., knative-lambda-prd -> prd)
	if strings.HasSuffix(j.config.Namespace, "-prd") {
		return "prd"
	}
	if strings.HasSuffix(j.config.Namespace, "-dev") {
		return "dev"
	}
	if strings.HasSuffix(j.config.Namespace, "-local") {
		return "local"
	}
	return "dev" // default fallback
}

// 🔧 generateImageURI - "Generate image URI for the build"
func (j *JobManagerImpl) generateImageURI(thirdPartyID, parserID, contentHash string) string {
	if j.awsConfig == nil {
		return fmt.Sprintf("unknown-registry/%s/%s:latest", thirdPartyID, parserID)
	}

	// Always use content hash for unique tagging
	imageTag := fmt.Sprintf("%s-%s-%s", thirdPartyID, parserID, contentHash[:8])
	return fmt.Sprintf("%s/knative-lambda:%s", j.awsConfig.ECRRegistry, imageTag)
}
