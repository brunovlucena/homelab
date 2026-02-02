package build

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/validation"
)

const (
	// BuildContextConfigMapSuffix is the suffix for build context ConfigMaps
	BuildContextConfigMapSuffix = "-build-context"

	// JobTTLAfterFinished is the TTL for completed/failed jobs
	JobTTLAfterFinished = int32(300) // 5 minutes

	// Default images - can be overridden via environment variables
	DefaultKanikoImage      = "gcr.io/kaniko-project/executor:v1.19.2"
	DefaultMinioClientImage = "minio/mc:latest"
	DefaultAlpineInitImage  = "alpine:3.19"

	// Default base images for Dockerfiles (using localhost:5001 for Kind cluster)
	DefaultNodeBaseImage   = "localhost:5001/node:20-alpine"
	DefaultPythonBaseImage = "localhost:5001/python:3.11-slim"
	DefaultGoBaseImage     = "localhost:5001/golang:1.21-alpine"
	DefaultAlpineRuntime   = "localhost:5001/alpine:3.19"
)

// Manager handles container image builds using Kaniko
type Manager struct {
	client client.Client
	scheme *runtime.Scheme

	// Registry configuration
	pushRegistry string // Registry for Kaniko to push to (k8s service DNS)
	pullRegistry string // Registry for kubelet to pull from (containerd mirror)

	// Image configuration
	kanikoImage      string
	minioClientImage string
	alpineInitImage  string

	// Base images for Dockerfiles
	nodeBaseImage   string
	pythonBaseImage string
	goBaseImage     string
	alpineRuntime   string
}

// BuildContext contains the context for a build
type BuildContext struct {
	// ConfigMapName is the name of the ConfigMap containing build context
	ConfigMapName string

	// ContentHash is the SHA-256 hash of the source content
	ContentHash string

	// ImageTag is the computed image tag
	ImageTag string
}

// BuildStatus represents the status of a build job
type BuildStatus struct {
	Completed bool
	Success   bool
	ImageURI  string
	Error     string
}

// NewManager creates a new build manager
func NewManager(client client.Client, scheme *runtime.Scheme) (*Manager, error) {
	m := &Manager{
		client: client,
		scheme: scheme,
	}

	// Load registry configuration from environment
	m.pushRegistry = getEnvOrDefault("BUILD_DEFAULT_REGISTRY", "localhost:5001")
	m.pullRegistry = getEnvOrDefault("BUILD_PULL_REGISTRY", "localhost:5001")

	// Load image configuration
	m.kanikoImage = getEnvOrDefault("BUILD_KANIKO_IMAGE", DefaultKanikoImage)
	m.minioClientImage = getEnvOrDefault("BUILD_MINIO_CLIENT_IMAGE", DefaultMinioClientImage)
	m.alpineInitImage = getEnvOrDefault("BUILD_ALPINE_INIT_IMAGE", DefaultAlpineInitImage)

	// Load base images for Dockerfiles
	m.nodeBaseImage = getEnvOrDefault("BUILD_NODE_BASE_IMAGE", DefaultNodeBaseImage)
	m.pythonBaseImage = getEnvOrDefault("BUILD_PYTHON_BASE_IMAGE", DefaultPythonBaseImage)
	m.goBaseImage = getEnvOrDefault("BUILD_GO_BASE_IMAGE", DefaultGoBaseImage)
	m.alpineRuntime = getEnvOrDefault("BUILD_ALPINE_RUNTIME_IMAGE", DefaultAlpineRuntime)

	return m, nil
}

const (
	// BuildServiceAccountName is the name of the service account used for build jobs
	BuildServiceAccountName = "knative-lambda-operator"
	// BuildRoleName is the name of the role for build operations
	BuildRoleName = "lambda-build-role"
	// BuildRoleBindingName is the name of the role binding for build operations
	BuildRoleBindingName = "lambda-build-binding"
)

// ensureBuildRBAC ensures the service account, role, and role binding exist for build jobs
// This is called automatically before creating build jobs to ensure RBAC is in place
func (m *Manager) ensureBuildRBAC(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	namespace := lambda.Namespace

	// Ensure ServiceAccount exists
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildServiceAccountName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"lambda.knative.io/component":  "build",
			},
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, m.client, sa, func() error {
		// Add imagePullSecrets if ghcr-secret exists
		secret := &corev1.Secret{}
		if err := m.client.Get(ctx, types.NamespacedName{Name: "ghcr-secret", Namespace: namespace}, secret); err == nil {
			// Check if secret is already in the list
			found := false
			for _, ref := range sa.ImagePullSecrets {
				if ref.Name == "ghcr-secret" {
					found = true
					break
				}
			}
			if !found {
				sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{Name: "ghcr-secret"})
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to ensure service account: %w", err)
	}
	if op != controllerutil.OperationResultNone {
		// Log creation/update if needed
	}

	// Ensure Role exists
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildRoleName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"lambda.knative.io/component":  "build",
			},
		},
	}

	op, err = controllerutil.CreateOrUpdate(ctx, m.client, role, func() error {
		role.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "secrets", "pods", "pods/log"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to ensure role: %w", err)
	}

	// Ensure RoleBinding exists
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildRoleBindingName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"lambda.knative.io/component":  "build",
			},
		},
	}

	op, err = controllerutil.CreateOrUpdate(ctx, m.client, roleBinding, func() error {
		roleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     BuildRoleName,
		}
		roleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      BuildServiceAccountName,
				Namespace: namespace,
			},
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to ensure role binding: %w", err)
	}

	return nil
}

// CreateBuildContext creates the build context for a Lambda function
func (m *Manager) CreateBuildContext(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) (*BuildContext, error) {
	// Get source code
	sourceCode, sourceFilename, err := m.getSourceCode(ctx, lambda)
	if err != nil {
		return nil, fmt.Errorf("failed to get source code: %w", err)
	}

	// Generate Dockerfile
	dockerfile, err := m.generateDockerfile(lambda, sourceFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Generate runtime wrapper
	runtimeWrapper, runtimeFilename, err := m.generateRuntimeWrapper(lambda)
	if err != nil {
		return nil, fmt.Errorf("failed to generate runtime wrapper: %w", err)
	}

	// Compute content hash for image tag
	contentHash := m.computeContentHash(sourceCode, dockerfile, runtimeWrapper)

	// Create tar.gz archive
	archive, err := m.createTarGzArchive(sourceCode, sourceFilename, dockerfile, runtimeWrapper, runtimeFilename, lambda)
	if err != nil {
		return nil, fmt.Errorf("failed to create build archive: %w", err)
	}

	// Create ConfigMap with build context
	configMapName := lambda.Name + BuildContextConfigMapSuffix
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: lambda.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"lambda.knative.io/name":       lambda.Name,
			},
		},
		BinaryData: map[string][]byte{
			"context.tar.gz": archive,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(lambda, configMap, m.scheme); err != nil {
		return nil, fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create or update ConfigMap
	existing := &corev1.ConfigMap{}
	err = m.client.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: lambda.Namespace}, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := m.client.Create(ctx, configMap); err != nil {
				return nil, fmt.Errorf("failed to create build context ConfigMap: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to check existing ConfigMap: %w", err)
		}
	} else {
		existing.BinaryData = configMap.BinaryData
		if err := m.client.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update build context ConfigMap: %w", err)
		}
	}

	// Compute image tag (first 12 chars of hash + timestamp for uniqueness)
	imageTag := contentHash[:12]

	return &BuildContext{
		ConfigMapName: configMapName,
		ContentHash:   contentHash,
		ImageTag:      imageTag,
	}, nil
}

// CreateKanikoJob creates a Kaniko build job for the Lambda function
func (m *Manager) CreateKanikoJob(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, buildCtx *BuildContext) (*batchv1.Job, error) {
	// Ensure RBAC exists before creating the job
	if err := m.ensureBuildRBAC(ctx, lambda); err != nil {
		return nil, fmt.Errorf("failed to ensure build RBAC: %w", err)
	}

	jobName := fmt.Sprintf("%s-build-%d", lambda.Name, time.Now().Unix())

	// Compute full image URI
	// Use pushRegistry for Kaniko (k8s service DNS), but store pullRegistry version in status
	imageURI := fmt.Sprintf("%s/%s/%s:%s", m.pushRegistry, lambda.Namespace, lambda.Name, buildCtx.ImageTag)

	backoffLimit := int32(0)
	ttlSeconds := JobTTLAfterFinished

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: lambda.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"lambda.knative.io/name":       lambda.Name,
				"lambda.knative.io/build":      "true",
			},
			Annotations: map[string]string{
				"lambda.knative.io/image-uri":     imageURI,
				"lambda.knative.io/content-hash":  buildCtx.ContentHash,
				"lambda.knative.io/pull-registry": m.pullRegistry,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSeconds,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"lambda.knative.io/name":  lambda.Name,
						"lambda.knative.io/build": "true",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					// Security Fix: BLUE-006 - Disable SA token auto-mount in build jobs
					// Build jobs don't need K8s API access, so we prevent token exposure
					AutomountServiceAccountToken: ptr.To(false),
					// Use the build service account (created automatically by ensureBuildRBAC)
					ServiceAccountName: BuildServiceAccountName,
					InitContainers: []corev1.Container{
						{
							Name:    "extract-context",
							Image:   m.alpineInitImage,
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								"cp /build-context/context.tar.gz /workspace/ && cd /workspace && tar -xzf context.tar.gz && rm context.tar.gz && ls -la",
							},
							// Security Fix: BLUE-007 - Add resource limits to init container
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "build-context",
									MountPath: "/build-context",
								},
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: m.kanikoImage,
							Args: []string{
								"--dockerfile=/workspace/Dockerfile",
								"--context=dir:///workspace",
								fmt.Sprintf("--destination=%s", imageURI),
								"--insecure",
								"--insecure-pull",
								"--skip-tls-verify",
								"--cache=false",
							},
							// Security Fix: BLUE-007 - Add resource limits to Kaniko container
							// Prevents resource exhaustion DoS attacks via malicious builds
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2"),
									corev1.ResourceMemory: resource.MustParse("4Gi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "build-context",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: buildCtx.ConfigMapName,
									},
								},
							},
						},
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(lambda, job, m.scheme); err != nil {
		return nil, fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create the job
	if err := m.client.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create Kaniko job: %w", err)
	}

	return job, nil
}

// GetBuildStatus returns the status of a build job
func (m *Manager) GetBuildStatus(ctx context.Context, namespace, jobName string) (*BuildStatus, error) {
	job := &batchv1.Job{}
	if err := m.client.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, job); err != nil {
		return nil, err
	}

	status := &BuildStatus{}

	// Check job conditions
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			status.Completed = true
			status.Success = true

			// Get image URI from annotation, converting to pull registry
			if imageURI, ok := job.Annotations["lambda.knative.io/image-uri"]; ok {
				pullRegistry := job.Annotations["lambda.knative.io/pull-registry"]
				if pullRegistry == "" {
					pullRegistry = m.pullRegistry
				}
				// Replace push registry with pull registry
				status.ImageURI = strings.Replace(imageURI, m.pushRegistry, pullRegistry, 1)
			}
			break
		}
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			status.Completed = true
			status.Success = false
			status.Error = condition.Message
			if status.Error == "" {
				status.Error = "Build job failed"
			}
			break
		}
	}

	return status, nil
}

// DeleteJob deletes a build job
func (m *Manager) DeleteJob(ctx context.Context, namespace, jobName string) error {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
	}

	propagation := metav1.DeletePropagationBackground
	if err := m.client.Delete(ctx, job, &client.DeleteOptions{
		PropagationPolicy: &propagation,
	}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

// getSourceCode retrieves the source code for a Lambda function
func (m *Manager) getSourceCode(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) ([]byte, string, error) {
	switch lambda.Spec.Source.Type {
	case "inline":
		if lambda.Spec.Source.Inline == nil {
			return nil, "", fmt.Errorf("inline source configuration is required")
		}
		filename := m.getSourceFilename(lambda.Spec.Runtime.Language)
		return []byte(lambda.Spec.Source.Inline.Code), filename, nil

	case "minio":
		return m.getMinioSourceCode(ctx, lambda)

	case "s3":
		return m.getS3SourceCode(ctx, lambda)

	case "git":
		return m.getGitSourceCode(ctx, lambda)

	default:
		return nil, "", fmt.Errorf("unsupported source type: %s", lambda.Spec.Source.Type)
	}
}

// getMinioSourceCode downloads source code from MinIO
// Security Fix: Validates endpoint, bucket, and key to prevent SSRF and injection
func (m *Manager) getMinioSourceCode(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) ([]byte, string, error) {
	if lambda.Spec.Source.MinIO == nil {
		return nil, "", fmt.Errorf("minio source configuration is required")
	}

	minioSpec := lambda.Spec.Source.MinIO

	// Security Fix: Validate MinIO source configuration
	if err := validation.ValidateMinIOSource(minioSpec.Endpoint, minioSpec.Bucket, minioSpec.Key); err != nil {
		return nil, "", fmt.Errorf("security validation failed for minio source: %w", err)
	}

	// Get credentials from secret
	accessKey := ""
	secretKey := ""
	if minioSpec.SecretRef != nil {
		secret := &corev1.Secret{}
		secretName := types.NamespacedName{
			Name:      minioSpec.SecretRef.Name,
			Namespace: lambda.Namespace,
		}
		if err := m.client.Get(ctx, secretName, secret); err != nil {
			return nil, "", fmt.Errorf("failed to get minio credentials secret '%s' in namespace '%s': %w (ensure the secret exists and contains 'access-key' and 'secret-key' keys)", secretName.Name, secretName.Namespace, err)
		}
		accessKey = string(secret.Data["accesskey"])
		if accessKey == "" {
			accessKey = string(secret.Data["access-key"])
		}
		if accessKey == "" {
			accessKey = string(secret.Data["AWS_ACCESS_KEY_ID"])
		}
		secretKey = string(secret.Data["secretkey"])
		if secretKey == "" {
			secretKey = string(secret.Data["secret-key"])
		}
		if secretKey == "" {
			secretKey = string(secret.Data["AWS_SECRET_ACCESS_KEY"])
		}

		if accessKey == "" || secretKey == "" {
			return nil, "", fmt.Errorf("minio credentials secret '%s' in namespace '%s' is missing required keys. Found keys: %v. Expected one of: access-key/accesskey/AWS_ACCESS_KEY_ID and secret-key/secretkey/AWS_SECRET_ACCESS_KEY", secretName.Name, secretName.Namespace, getSecretKeys(secret))
		}
	}

	// Default endpoint
	endpoint := minioSpec.Endpoint
	if endpoint == "" {
		endpoint = "minio.minio.svc.cluster.local:9000"
	}

	// Create MinIO client
	useSSL := strings.HasPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create minio client for endpoint '%s': %w", endpoint, err)
	}

	// Check if the key is a directory (ends with /)
	key := minioSpec.Key
	isDirectory := strings.HasSuffix(key, "/")

	if isDirectory {
		// Download all files from the directory and find the main source file
		return m.downloadMinioDirectory(ctx, minioClient, minioSpec.Bucket, key, lambda.Spec.Runtime.Language)
	}

	// Download single file
	return m.downloadMinioFile(ctx, minioClient, minioSpec.Bucket, key, lambda.Spec.Runtime.Language)
}

// getSecretKeys returns a list of keys in the secret (for error messages)
func getSecretKeys(secret *corev1.Secret) []string {
	keys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		keys = append(keys, k)
	}
	return keys
}

// downloadMinioFile downloads a single file from MinIO
func (m *Manager) downloadMinioFile(ctx context.Context, client *minio.Client, bucket, key, language string) ([]byte, string, error) {
	object, err := client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from minio: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read object from minio: %w", err)
	}

	filename := m.getSourceFilename(language)
	return data, filename, nil
}

// downloadMinioDirectory downloads files from a MinIO directory prefix
func (m *Manager) downloadMinioDirectory(ctx context.Context, minioClient *minio.Client, bucket, prefix, language string) ([]byte, string, error) {
	expectedFilename := m.getSourceFilename(language)

	// List objects in the directory
	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var sourceCode []byte
	var foundFiles []string
	var listErrors []error

	for object := range objectCh {
		if object.Err != nil {
			listErrors = append(listErrors, object.Err)
			continue
		}

		// Skip directories
		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		// Track all files found for better error messages
		foundFiles = append(foundFiles, object.Key)

		// Get the filename relative to the prefix
		filename := strings.TrimPrefix(object.Key, prefix)

		// Look for the main source file
		if filename == expectedFilename || filepath.Base(object.Key) == expectedFilename {
			obj, err := minioClient.GetObject(ctx, bucket, object.Key, minio.GetObjectOptions{})
			if err != nil {
				return nil, "", fmt.Errorf("failed to get object %s from minio bucket %s: %w", object.Key, bucket, err)
			}

			data, err := io.ReadAll(obj)
			obj.Close()
			if err != nil {
				return nil, "", fmt.Errorf("failed to read object %s from minio bucket %s: %w", object.Key, bucket, err)
			}

			sourceCode = data
			break
		}
	}

	// Check for listing errors
	if len(listErrors) > 0 {
		return nil, "", fmt.Errorf("failed to list objects in minio bucket %s with prefix %s: %v (this may indicate the bucket doesn't exist, credentials are wrong, or network connectivity issues)", bucket, prefix, listErrors)
	}

	if sourceCode == nil {
		var filesList string
		if len(foundFiles) > 0 {
			filesList = fmt.Sprintf(" Found %d file(s): %s", len(foundFiles), strings.Join(foundFiles, ", "))
		} else {
			filesList = " No files found in this path."
		}
		return nil, "", fmt.Errorf("main source file '%s' not found in minio bucket '%s' with prefix '%s'.%s Expected path: s3://%s/%s%s", expectedFilename, bucket, prefix, filesList, bucket, prefix, expectedFilename)
	}

	return sourceCode, expectedFilename, nil
}

// getS3SourceCode downloads source code from S3
// Security Fix: Validates bucket and key to prevent injection attacks
func (m *Manager) getS3SourceCode(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) ([]byte, string, error) {
	if lambda.Spec.Source.S3 == nil {
		return nil, "", fmt.Errorf("s3 source configuration is required")
	}

	s3Spec := lambda.Spec.Source.S3

	// Security Fix: Validate S3 source configuration
	if err := validation.ValidateS3Source(s3Spec.Bucket, s3Spec.Key, s3Spec.Region); err != nil {
		return nil, "", fmt.Errorf("security validation failed for s3 source: %w", err)
	}

	// Get credentials from secret
	accessKey := ""
	secretKey := ""
	if s3Spec.SecretRef != nil {
		secret := &corev1.Secret{}
		secretName := types.NamespacedName{
			Name:      s3Spec.SecretRef.Name,
			Namespace: lambda.Namespace,
		}
		if err := m.client.Get(ctx, secretName, secret); err != nil {
			return nil, "", fmt.Errorf("failed to get s3 credentials secret: %w", err)
		}
		accessKey = string(secret.Data["AWS_ACCESS_KEY_ID"])
		if accessKey == "" {
			accessKey = string(secret.Data["accesskey"])
		}
		secretKey = string(secret.Data["AWS_SECRET_ACCESS_KEY"])
		if secretKey == "" {
			secretKey = string(secret.Data["secretkey"])
		}
	}

	// AWS S3 endpoint
	region := s3Spec.Region
	if region == "" {
		region = "us-east-1"
	}
	endpoint := fmt.Sprintf("s3.%s.amazonaws.com", region)

	// Create S3-compatible client using MinIO SDK
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create s3 client: %w", err)
	}

	// Check if the key is a directory (ends with /)
	key := s3Spec.Key
	isDirectory := strings.HasSuffix(key, "/")

	if isDirectory {
		return m.downloadMinioDirectory(ctx, s3Client, s3Spec.Bucket, key, lambda.Spec.Runtime.Language)
	}

	return m.downloadMinioFile(ctx, s3Client, s3Spec.Bucket, key, lambda.Spec.Runtime.Language)
}

// getGitSourceCode clones a Git repository and retrieves the source code
// Security Fixes:
// - BLUE-001: SSRF via Go-Git Library - validates URL against blocked hosts/IPs
// - BLUE-005: Path Traversal in Git Source Path - validates path doesn't escape repo
func (m *Manager) getGitSourceCode(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) ([]byte, string, error) {
	if lambda.Spec.Source.Git == nil {
		return nil, "", fmt.Errorf("git source configuration is required")
	}

	gitSpec := lambda.Spec.Source.Git

	// Security Fix: BLUE-001 - Validate Git URL to prevent SSRF
	if err := validation.ValidateGitURL(gitSpec.URL); err != nil {
		return nil, "", fmt.Errorf("security validation failed for git URL: %w", err)
	}

	// Security Fix: Validate Git ref to prevent command injection
	if err := validation.ValidateGitRef(gitSpec.Ref); err != nil {
		return nil, "", fmt.Errorf("security validation failed for git ref: %w", err)
	}

	// Security Fix: BLUE-005 - Validate Git path to prevent path traversal
	if err := validation.ValidateGitPath(gitSpec.Path); err != nil {
		return nil, "", fmt.Errorf("security validation failed for git path: %w", err)
	}

	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "lambda-git-clone-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Prepare clone options
	cloneOpts := &git.CloneOptions{
		URL:   gitSpec.URL,
		Depth: 1, // Shallow clone for efficiency
	}

	// Set up authentication if credentials are provided
	if gitSpec.SecretRef != nil {
		auth, err := m.getGitAuth(ctx, lambda.Namespace, gitSpec.SecretRef.Name, gitSpec.URL)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get git credentials: %w", err)
		}
		if auth != nil {
			cloneOpts.Auth = auth
		}
	}

	// Clone the repository
	repo, err := git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to clone git repository %s: %w", gitSpec.URL, err)
	}

	// Checkout the specified ref (branch, tag, or commit)
	ref := gitSpec.Ref
	if ref == "" {
		ref = "main"
	}

	if err := m.checkoutGitRef(repo, ref); err != nil {
		return nil, "", fmt.Errorf("failed to checkout ref %s: %w", ref, err)
	}

	// Determine the source file path
	basePath := tmpDir
	if gitSpec.Path != "" {
		// Security Fix: BLUE-005 - Validate the resolved path doesn't escape tmpDir
		if err := validation.ValidateSecurePath(tmpDir, gitSpec.Path); err != nil {
			return nil, "", fmt.Errorf("security validation failed: %w", err)
		}
		basePath = filepath.Join(tmpDir, gitSpec.Path)
	}

	// Find and read the source file
	expectedFilename := m.getSourceFilename(lambda.Spec.Runtime.Language)
	sourceFilePath := filepath.Join(basePath, expectedFilename)

	// Check if the path is a directory or file
	info, err := os.Stat(basePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to access path %s: %w", gitSpec.Path, err)
	}

	if info.IsDir() {
		// Look for the main source file in the directory
		sourceCode, err := os.ReadFile(sourceFilePath)
		if err != nil {
			// Try to find any matching source file
			sourceCode, expectedFilename, err = m.findSourceFileInDir(basePath, lambda.Spec.Runtime.Language)
			if err != nil {
				return nil, "", fmt.Errorf("failed to find source file %s in git repository path %s: %w", expectedFilename, gitSpec.Path, err)
			}
		}
		return sourceCode, expectedFilename, nil
	}

	// Path points to a specific file
	sourceCode, err := os.ReadFile(basePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read source file from git: %w", err)
	}

	// Use the actual filename from the path
	actualFilename := filepath.Base(basePath)
	return sourceCode, actualFilename, nil
}

// getGitAuth returns the appropriate authentication method based on the secret contents
func (m *Manager) getGitAuth(ctx context.Context, namespace, secretName, repoURL string) (transport.AuthMethod, error) {
	secret := &corev1.Secret{}
	secretKey := types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}
	if err := m.client.Get(ctx, secretKey, secret); err != nil {
		return nil, fmt.Errorf("failed to get git credentials secret: %w", err)
	}

	// Check for SSH private key (for git@github.com:... URLs)
	if sshKey, ok := secret.Data["ssh-privatekey"]; ok {
		// Parse SSH private key
		publicKeys, err := ssh.NewPublicKeys("git", sshKey, "")
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH private key: %w", err)
		}
		return publicKeys, nil
	}

	// Check for username/password or token (for HTTPS URLs)
	username := string(secret.Data["username"])
	password := string(secret.Data["password"])

	// Support token-only auth (common for GitHub/GitLab)
	if token, ok := secret.Data["token"]; ok && password == "" {
		password = string(token)
		if username == "" {
			// GitHub and GitLab accept any username with token auth
			username = "git"
		}
	}

	// Support GitHub-style personal access tokens
	if pat, ok := secret.Data["github-token"]; ok && password == "" {
		password = string(pat)
		username = "x-access-token" // GitHub's expected username for PATs
	}

	// Support GitLab-style access tokens
	if pat, ok := secret.Data["gitlab-token"]; ok && password == "" {
		password = string(pat)
		username = "oauth2" // GitLab's expected username for tokens
	}

	if username != "" && password != "" {
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil
	}

	// No auth found in secret
	return nil, nil
}

// checkoutGitRef checks out a specific ref (branch, tag, or commit hash)
func (m *Manager) checkoutGitRef(repo *git.Repository, ref string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to resolve as a branch first
	branchRef := plumbing.NewBranchReferenceName(ref)
	_, err = repo.Reference(branchRef, true)
	if err == nil {
		return worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		})
	}

	// Try as a tag
	tagRef := plumbing.NewTagReferenceName(ref)
	_, err = repo.Reference(tagRef, true)
	if err == nil {
		return worktree.Checkout(&git.CheckoutOptions{
			Branch: tagRef,
		})
	}

	// Try as a remote branch
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", ref)
	remoteRef, err := repo.Reference(remoteBranchRef, true)
	if err == nil {
		return worktree.Checkout(&git.CheckoutOptions{
			Hash: remoteRef.Hash(),
		})
	}

	// Try as a commit hash
	if len(ref) >= 7 {
		hash := plumbing.NewHash(ref)
		if !hash.IsZero() {
			return worktree.Checkout(&git.CheckoutOptions{
				Hash: hash,
			})
		}
	}

	// If nothing worked, the ref might already be checked out (e.g., default branch)
	// Just verify HEAD is valid
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("unable to resolve ref %s and HEAD is invalid: %w", ref, err)
	}

	// Log that we're using the default branch
	_ = headRef // Ref resolved to default

	return nil
}

// findSourceFileInDir searches for a source file in a directory based on language
func (m *Manager) findSourceFileInDir(dir, language string) ([]byte, string, error) {
	expectedFilename := m.getSourceFilename(language)

	// First, try the expected filename directly
	directPath := filepath.Join(dir, expectedFilename)
	if data, err := os.ReadFile(directPath); err == nil {
		return data, expectedFilename, nil
	}

	// Get file extensions for the language
	var extensions []string
	switch strings.ToLower(language) {
	case "python", "python3":
		extensions = []string{".py"}
	case "nodejs", "node", "javascript":
		extensions = []string{".js", ".mjs"}
	case "go", "golang":
		extensions = []string{".go"}
	default:
		extensions = []string{".py"} // Default to Python
	}

	// Search for any matching file
	var foundFile string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				// Prefer main/index files
				base := filepath.Base(path)
				if base == expectedFilename || strings.HasPrefix(base, "main") || strings.HasPrefix(base, "index") {
					foundFile = path
					return filepath.SkipAll
				}
				// Keep the first match as fallback
				if foundFile == "" {
					foundFile = path
				}
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, "", fmt.Errorf("error searching for source files: %w", err)
	}

	if foundFile == "" {
		return nil, "", fmt.Errorf("no source file with extensions %v found in directory", extensions)
	}

	data, err := os.ReadFile(foundFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read source file %s: %w", foundFile, err)
	}

	return data, filepath.Base(foundFile), nil
}

// getSourceFilename returns the appropriate source filename for a language
func (m *Manager) getSourceFilename(language string) string {
	switch strings.ToLower(language) {
	case "python", "python3":
		return "main.py"
	case "nodejs", "node", "javascript":
		return "index.js"
	case "go", "golang":
		return "main.go"
	default:
		return "main.py"
	}
}

// generateDockerfile generates a Dockerfile for the Lambda function using embedded templates
func (m *Manager) generateDockerfile(lambda *lambdav1alpha1.LambdaFunction, sourceFilename string) ([]byte, error) {
	language := strings.ToLower(lambda.Spec.Runtime.Language)
	handler := lambda.Spec.Runtime.Handler
	if handler == "" {
		handler = "handler"
	}

	// Get the Dockerfile template from embedded files
	tmplContent, err := GetDockerfileTemplate(language)
	if err != nil {
		return nil, fmt.Errorf("failed to get Dockerfile template: %w", err)
	}

	// Determine base image and version using configured images
	var baseImage, version string
	switch language {
	case "python", "python3":
		// Use configured python base image, extract registry/repo without tag
		baseImage = extractImageWithoutTag(m.pythonBaseImage, "localhost:5001/python")
		version = lambda.Spec.Runtime.Version
		if version == "" {
			version = "3.11"
		}
	case "nodejs", "node", "javascript":
		// Use configured node base image
		baseImage = extractImageWithoutTag(m.nodeBaseImage, "localhost:5001/node")
		version = lambda.Spec.Runtime.Version
		if version == "" {
			version = "20"
		}
	case "go", "golang":
		// Use configured go base image
		baseImage = extractImageWithoutTag(m.goBaseImage, "localhost:5001/golang")
		version = lambda.Spec.Runtime.Version
		if version == "" {
			version = "1.22"
		}
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Parse and execute the template
	tmpl, err := template.New("dockerfile").Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Dockerfile template: %w", err)
	}

	// Template data - includes AlpineImage for Go multi-stage builds
	data := struct {
		BaseImage         string
		AlpineImage       string
		Version           string
		Handler           string
		FunctionName      string
		FunctionNamespace string
		TimeoutSeconds    int
	}{
		BaseImage:         baseImage,
		AlpineImage:       m.alpineRuntime,
		Version:           version,
		Handler:           handler,
		FunctionName:      lambda.Name,
		FunctionNamespace: lambda.Namespace,
		TimeoutSeconds:    300,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute Dockerfile template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateRuntimeWrapper generates the runtime wrapper for the Lambda function
// Security Fix: BLUE-002 - Sanitizes handler field before Go template rendering
func (m *Manager) generateRuntimeWrapper(lambda *lambdav1alpha1.LambdaFunction) ([]byte, string, error) {
	language := strings.ToLower(lambda.Spec.Runtime.Language)

	var filename string
	switch language {
	case "python", "python3":
		filename = "runtime.py"
	case "nodejs", "node", "javascript":
		filename = "runtime.js"
	case "go", "golang":
		// Go doesn't need a runtime wrapper - it's compiled
		return nil, "", nil
	default:
		return nil, "", fmt.Errorf("unsupported language: %s", language)
	}

	// Security Fix: BLUE-002 - Validate and sanitize handler to prevent template injection
	// The handler field is interpolated into the runtime template via Go's text/template
	// Malicious handlers could inject code that escapes the template context
	handler := lambda.Spec.Runtime.Handler
	if err := validation.ValidateHandler(handler); err != nil {
		return nil, "", fmt.Errorf("security validation failed for handler: %w", err)
	}
	// Use sanitized handler (returns safe default if invalid)
	handler = validation.SanitizeHandler(handler)

	// Get the template from embedded files
	tmplContent, err := GetRuntimeTemplate(language)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get runtime template: %w", err)
	}

	// Parse and execute the template with lambda-specific values
	tmpl, err := template.New("runtime").Parse(tmplContent)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse runtime template: %w", err)
	}

	// Template data - using validated/sanitized handler
	data := struct {
		FunctionName string
		Handler      string
	}{
		FunctionName: lambda.Name,
		Handler:      handler, // Security: validated and sanitized
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, "", fmt.Errorf("failed to execute runtime template: %w", err)
	}

	return buf.Bytes(), filename, nil
}

// computeContentHash computes a SHA-256 hash of the build content
func (m *Manager) computeContentHash(sourceCode, dockerfile, runtimeWrapper []byte) string {
	h := sha256.New()
	h.Write(sourceCode)
	h.Write(dockerfile)
	if runtimeWrapper != nil {
		h.Write(runtimeWrapper)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// createTarGzArchive creates a tar.gz archive of the build context
func (m *Manager) createTarGzArchive(sourceCode []byte, sourceFilename string, dockerfile, runtimeWrapper []byte, runtimeFilename string, lambda *lambdav1alpha1.LambdaFunction) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	// Add Dockerfile
	if err := m.addFileToTar(tarWriter, "Dockerfile", dockerfile); err != nil {
		return nil, err
	}

	// Add source code
	if err := m.addFileToTar(tarWriter, sourceFilename, sourceCode); err != nil {
		return nil, err
	}

	// Add runtime wrapper if present
	if runtimeWrapper != nil && runtimeFilename != "" {
		if err := m.addFileToTar(tarWriter, runtimeFilename, runtimeWrapper); err != nil {
			return nil, err
		}
	}

	// Add empty requirements.txt for Python if not inline with requirements
	if strings.ToLower(lambda.Spec.Runtime.Language) == "python" || strings.ToLower(lambda.Spec.Runtime.Language) == "python3" {
		if err := m.addFileToTar(tarWriter, "requirements.txt", []byte("# Auto-generated\n")); err != nil {
			return nil, err
		}
	}

	// Add package.json with runtime dependencies for Node.js
	lang := strings.ToLower(lambda.Spec.Runtime.Language)
	if lang == "nodejs" || lang == "node" || lang == "javascript" {
		// The runtime.js wrapper requires uuid for CloudEvent ID generation
		emptyPackageJSON := []byte(`{
  "name": "lambda-function",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "uuid": "^9.0.0"
  }
}
`)
		if err := m.addFileToTar(tarWriter, "package.json", emptyPackageJSON); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// addFileToTar adds a file to the tar archive
func (m *Manager) addFileToTar(tarWriter *tar.Writer, filename string, content []byte) error {
	header := &tar.Header{
		Name:    filename,
		Size:    int64(len(content)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", filename, err)
	}

	if _, err := tarWriter.Write(content); err != nil {
		return fmt.Errorf("failed to write tar content for %s: %w", filename, err)
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// extractImageWithoutTag extracts the image name without the tag from a full image reference.
// For example: "localhost:5001/python:3.11-slim" -> "localhost:5001/python"
// Handles port numbers in registry names correctly.
func extractImageWithoutTag(image, defaultImage string) string {
	if image == "" {
		return defaultImage
	}
	// Find the last colon that's after the last slash (that's the tag separator)
	lastSlash := strings.LastIndex(image, "/")
	if lastSlash == -1 {
		// No slash, so the colon (if any) is the tag separator
		if idx := strings.LastIndex(image, ":"); idx != -1 {
			return image[:idx]
		}
		return image
	}
	// Find the colon after the last slash (that's the tag separator)
	afterSlash := image[lastSlash:]
	if idx := strings.Index(afterSlash, ":"); idx != -1 {
		return image[:lastSlash+idx]
	}
	return image
}
