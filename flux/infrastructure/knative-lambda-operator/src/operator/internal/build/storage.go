// Package build provides build context management with generic storage backends.
// This file implements BACKEND-002: Build Context Management with multi-storage support.
//
// Supported storage backends:
//   - ConfigMap: For small contexts (< 768KB), cluster-local, uses owner ref for GC
//   - S3/MinIO: For large contexts or explicit choice, TTL-based cleanup
//   - GCS: For large contexts or explicit choice, TTL-based cleanup
//
// Storage selection is driven by:
//  1. Annotation `lambda.knative.io/build-context-storage` (configmap|s3|gcs)
//  2. Size threshold: contexts > 768KB automatically use object storage
package build

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/minio/minio-go/v7"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

const (
	// StorageAnnotation is the annotation key for explicit storage backend selection
	StorageAnnotation = "lambda.knative.io/build-context-storage"

	// ConfigMapSizeLimit is the maximum size for ConfigMap storage (768KB for safety margin)
	// Kubernetes etcd limit is 1MiB, we use 768KB to leave room for metadata
	ConfigMapSizeLimit = 768 * 1024 // 768KB

	// StorageBackendConfigMap uses ConfigMap for build context storage
	StorageBackendConfigMap = "configmap"
	// StorageBackendS3 uses S3/MinIO for build context storage
	StorageBackendS3 = "s3"
	// StorageBackendGCS uses Google Cloud Storage for build context storage
	StorageBackendGCS = "gcs"
)

// Build context metrics
var (
	buildContextCreationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "creation_total",
			Help:      "Total number of build context creations by storage and source backend",
		},
		[]string{"storage", "source"},
	)

	buildContextCreationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "creation_duration_seconds",
			Help:      "Duration of build context creation in seconds",
			Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"storage"},
	)

	buildContextSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "size_bytes",
			Help:      "Size of build context archives in bytes",
			Buckets:   []float64{1024, 10240, 102400, 524288, 1048576, 5242880, 10485760},
		},
		[]string{"storage"},
	)

	buildContextStorageErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "storage_errors_total",
			Help:      "Total number of storage errors by backend",
		},
		[]string{"storage", "error_type"},
	)

	buildContextSourceErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "source_errors_total",
			Help:      "Total number of source fetch errors by backend",
		},
		[]string{"source", "error_type"},
	)

	buildContextConfigMapSizeLimitTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "configmap_size_limit_total",
			Help:      "Total number of contexts that exceeded ConfigMap size limit",
		},
	)

	buildContextCleanupTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "knative_lambda",
			Subsystem: "build_context",
			Name:      "cleanup_total",
			Help:      "Total number of cleanup operations by storage and result",
		},
		[]string{"storage", "result"},
	)
)

func init() {
	// Register build context metrics
	ctrlmetrics.Registry.MustRegister(
		buildContextCreationTotal,
		buildContextCreationDuration,
		buildContextSizeBytes,
		buildContextStorageErrors,
		buildContextSourceErrors,
		buildContextConfigMapSizeLimitTotal,
		buildContextCleanupTotal,
	)
}

// BuildContextMeta contains metadata for a build context
type BuildContextMeta struct {
	// LambdaName is the name of the Lambda function
	LambdaName string
	// LambdaNamespace is the namespace of the Lambda function
	LambdaNamespace string
	// ContentHash is the SHA-256 hash of the content
	ContentHash string
	// CreatedAt is the creation timestamp
	CreatedAt time.Time
}

// BuildContextLocation describes where the build context is stored
type BuildContextLocation struct {
	// Backend is the storage backend type (configmap, s3, gcs)
	Backend string
	// ConfigMapName is set when Backend is configmap
	ConfigMapName string
	// Bucket is set when Backend is s3 or gcs
	Bucket string
	// Key is the object key for s3/gcs backends
	Key string
	// Endpoint is the S3 endpoint (for MinIO/S3-compatible)
	Endpoint string
	// Region is the AWS region (for S3)
	Region string
	// ContentHash is the SHA-256 hash of the content
	ContentHash string
	// ImageTag is the computed image tag
	ImageTag string
}

// BuildContextStore is the generic interface for build context storage backends.
// Implementations: ConfigMapStore, S3Store, GCSStore
type BuildContextStore interface {
	// Save stores the build context and returns the location
	Save(ctx context.Context, key string, tarGz []byte, meta BuildContextMeta) (*BuildContextLocation, error)

	// Cleanup removes build contexts older than the specified duration
	// Returns the number of items cleaned up
	Cleanup(ctx context.Context, olderThan time.Duration) (int, error)

	// Name returns the backend name for metrics
	Name() string
}

// StorageSelector selects the appropriate storage backend based on context size and annotations
type StorageSelector struct {
	configMapStore *ConfigMapStore
	s3Store        *S3Store
	gcsStore       *GCSStore
	defaultBackend string
}

// NewStorageSelector creates a new storage selector with the provided backends
func NewStorageSelector(k8sClient client.Client, scheme interface{}, s3Config *S3Config, gcsConfig *GCSConfig) *StorageSelector {
	selector := &StorageSelector{
		configMapStore: NewConfigMapStore(k8sClient, scheme),
		defaultBackend: StorageBackendConfigMap,
	}

	if s3Config != nil && s3Config.Bucket != "" {
		s3Store, err := NewS3Store(s3Config)
		if err == nil {
			selector.s3Store = s3Store
			// If S3 is configured, use it as fallback for large contexts
			selector.defaultBackend = StorageBackendS3
		}
	}

	if gcsConfig != nil && gcsConfig.Bucket != "" {
		gcsStore, err := NewGCSStore(context.Background(), gcsConfig)
		if err == nil {
			selector.gcsStore = gcsStore
		}
	}

	return selector
}

// SelectStore returns the appropriate store based on lambda annotations and context size
func (s *StorageSelector) SelectStore(lambda *lambdav1alpha1.LambdaFunction, contextSize int) (BuildContextStore, error) {
	// Check for explicit annotation
	if lambda.Annotations != nil {
		if backend, ok := lambda.Annotations[StorageAnnotation]; ok {
			switch strings.ToLower(backend) {
			case StorageBackendConfigMap:
				if contextSize > ConfigMapSizeLimit {
					buildContextConfigMapSizeLimitTotal.Inc()
					return nil, fmt.Errorf("context size %d exceeds ConfigMap limit %d; use s3 or gcs storage annotation", contextSize, ConfigMapSizeLimit)
				}
				return s.configMapStore, nil
			case StorageBackendS3:
				if s.s3Store == nil {
					return nil, fmt.Errorf("S3 storage backend not configured")
				}
				return s.s3Store, nil
			case StorageBackendGCS:
				if s.gcsStore == nil {
					return nil, fmt.Errorf("GCS storage backend not configured")
				}
				return s.gcsStore, nil
			default:
				return nil, fmt.Errorf("unsupported storage backend: %s (supported: configmap, s3, gcs)", backend)
			}
		}
	}

	// Auto-select based on size
	if contextSize > ConfigMapSizeLimit {
		buildContextConfigMapSizeLimitTotal.Inc()
		// Try S3 first, then GCS
		if s.s3Store != nil {
			return s.s3Store, nil
		}
		if s.gcsStore != nil {
			return s.gcsStore, nil
		}
		return nil, fmt.Errorf("context size %d exceeds ConfigMap limit %d and no object storage backend is configured", contextSize, ConfigMapSizeLimit)
	}

	// Default to ConfigMap for small contexts
	return s.configMapStore, nil
}

// GetConfigMapStore returns the ConfigMap store for direct access
func (s *StorageSelector) GetConfigMapStore() *ConfigMapStore {
	return s.configMapStore
}

// GetS3Store returns the S3 store for direct access
func (s *StorageSelector) GetS3Store() *S3Store {
	return s.s3Store
}

// GetGCSStore returns the GCS store for direct access
func (s *StorageSelector) GetGCSStore() *GCSStore {
	return s.gcsStore
}

// S3Config holds S3/MinIO configuration
type S3Config struct {
	Endpoint        string
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	// PathPrefix is prepended to all keys
	PathPrefix string
}

// GCSConfig holds GCS configuration
type GCSConfig struct {
	Bucket string
	// Project is the GCP project ID
	Project string
	// CredentialsJSON is the service account JSON key
	CredentialsJSON []byte
	// PathPrefix is prepended to all keys
	PathPrefix string
}

// =============================================================================
// ConfigMapStore Implementation
// =============================================================================

// ConfigMapStore stores build contexts in Kubernetes ConfigMaps
type ConfigMapStore struct {
	client client.Client
	scheme interface{} // runtime.Scheme, stored as interface to avoid import issues
}

// NewConfigMapStore creates a new ConfigMap storage backend
func NewConfigMapStore(k8sClient client.Client, scheme interface{}) *ConfigMapStore {
	return &ConfigMapStore{
		client: k8sClient,
		scheme: scheme,
	}
}

// Name returns the backend name
func (s *ConfigMapStore) Name() string {
	return StorageBackendConfigMap
}

// Save stores the build context in a ConfigMap
func (s *ConfigMapStore) Save(ctx context.Context, key string, tarGz []byte, meta BuildContextMeta) (*BuildContextLocation, error) {
	startTime := time.Now()
	defer func() {
		buildContextCreationDuration.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
	}()

	// Check size limit
	if len(tarGz) > ConfigMapSizeLimit {
		buildContextStorageErrors.WithLabelValues(s.Name(), "size_exceeded").Inc()
		return nil, fmt.Errorf("context size %d exceeds ConfigMap limit %d", len(tarGz), ConfigMapSizeLimit)
	}

	configMapName := meta.LambdaName + BuildContextConfigMapSuffix
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: meta.LambdaNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"lambda.knative.io/name":       meta.LambdaName,
				"lambda.knative.io/component":  "build-context",
			},
			Annotations: map[string]string{
				"lambda.knative.io/content-hash": meta.ContentHash,
				"lambda.knative.io/created-at":   meta.CreatedAt.Format(time.RFC3339),
			},
		},
		BinaryData: map[string][]byte{
			"context.tar.gz": tarGz,
		},
	}

	// Try to get existing ConfigMap
	existing := &corev1.ConfigMap{}
	err := s.client.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: meta.LambdaNamespace}, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new ConfigMap
			if err := s.client.Create(ctx, configMap); err != nil {
				buildContextStorageErrors.WithLabelValues(s.Name(), "create_failed").Inc()
				return nil, fmt.Errorf("failed to create build context ConfigMap: %w", err)
			}
		} else {
			buildContextStorageErrors.WithLabelValues(s.Name(), "get_failed").Inc()
			return nil, fmt.Errorf("failed to check existing ConfigMap: %w", err)
		}
	} else {
		// Update existing ConfigMap
		existing.BinaryData = configMap.BinaryData
		existing.Annotations = configMap.Annotations
		if err := s.client.Update(ctx, existing); err != nil {
			buildContextStorageErrors.WithLabelValues(s.Name(), "update_failed").Inc()
			return nil, fmt.Errorf("failed to update build context ConfigMap: %w", err)
		}
	}

	buildContextSizeBytes.WithLabelValues(s.Name()).Observe(float64(len(tarGz)))

	// Compute image tag (first 12 chars of hash or full hash if shorter)
	imageTag := meta.ContentHash
	if len(imageTag) > 12 {
		imageTag = imageTag[:12]
	}

	return &BuildContextLocation{
		Backend:       StorageBackendConfigMap,
		ConfigMapName: configMapName,
		ContentHash:   meta.ContentHash,
		ImageTag:      imageTag,
	}, nil
}

// Cleanup for ConfigMap is a no-op as ConfigMaps use owner references for GC
func (s *ConfigMapStore) Cleanup(ctx context.Context, olderThan time.Duration) (int, error) {
	// ConfigMaps are cleaned up via owner reference when Lambda is deleted
	// No TTL-based cleanup needed
	buildContextCleanupTotal.WithLabelValues(s.Name(), "skipped").Inc()
	return 0, nil
}

// SetOwnerReference sets the owner reference on a ConfigMap for GC
func (s *ConfigMapStore) SetOwnerReference(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, configMapName string, scheme interface{}) error {
	configMap := &corev1.ConfigMap{}
	if err := s.client.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: lambda.Namespace}, configMap); err != nil {
		return fmt.Errorf("failed to get ConfigMap for owner reference: %w", err)
	}

	// Set owner reference manually - scheme parameter kept for backwards compatibility
	// but we don't rely on it since it may be nil or wrong type
	ownerRef := metav1.OwnerReference{
		APIVersion:         lambda.APIVersion,
		Kind:               lambda.Kind,
		Name:               lambda.Name,
		UID:                lambda.UID,
		Controller:         boolPtr(true),
		BlockOwnerDeletion: boolPtr(true),
	}

	// Check if owner reference already exists
	found := false
	for i, ref := range configMap.OwnerReferences {
		if ref.UID == lambda.UID {
			configMap.OwnerReferences[i] = ownerRef
			found = true
			break
		}
	}
	if !found {
		configMap.OwnerReferences = append(configMap.OwnerReferences, ownerRef)
	}

	if err := s.client.Update(ctx, configMap); err != nil {
		return fmt.Errorf("failed to update ConfigMap with owner reference: %w", err)
	}

	return nil
}

// =============================================================================
// S3Store Implementation
// =============================================================================

// S3Store stores build contexts in S3-compatible storage (AWS S3, MinIO)
type S3Store struct {
	client     *minio.Client
	bucket     string
	pathPrefix string
	endpoint   string
	region     string
}

// NewS3Store creates a new S3 storage backend
func NewS3Store(config *S3Config) (*S3Store, error) {
	if config.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", config.Region)
	}

	// Remove protocol prefix if present
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return &S3Store{
		client:     minioClient,
		bucket:     config.Bucket,
		pathPrefix: config.PathPrefix,
		endpoint:   endpoint,
		region:     config.Region,
	}, nil
}

// Name returns the backend name
func (s *S3Store) Name() string {
	return StorageBackendS3
}

// Save stores the build context in S3
func (s *S3Store) Save(ctx context.Context, key string, tarGz []byte, meta BuildContextMeta) (*BuildContextLocation, error) {
	startTime := time.Now()
	defer func() {
		buildContextCreationDuration.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
	}()

	// Build the full key with prefix
	fullKey := key
	if s.pathPrefix != "" {
		fullKey = strings.TrimSuffix(s.pathPrefix, "/") + "/" + key
	}

	// Upload to S3
	reader := bytes.NewReader(tarGz)
	_, err := s.client.PutObject(ctx, s.bucket, fullKey, reader, int64(len(tarGz)), minio.PutObjectOptions{
		ContentType: "application/gzip",
		UserMetadata: map[string]string{
			"lambda-name":      meta.LambdaName,
			"lambda-namespace": meta.LambdaNamespace,
			"content-hash":     meta.ContentHash,
			"created-at":       meta.CreatedAt.Format(time.RFC3339),
		},
	})
	if err != nil {
		buildContextStorageErrors.WithLabelValues(s.Name(), "upload_failed").Inc()
		return nil, fmt.Errorf("failed to upload build context to S3: %w", err)
	}

	buildContextSizeBytes.WithLabelValues(s.Name()).Observe(float64(len(tarGz)))

	// Compute image tag (first 12 chars of hash or full hash if shorter)
	imageTag := meta.ContentHash
	if len(imageTag) > 12 {
		imageTag = imageTag[:12]
	}

	return &BuildContextLocation{
		Backend:     StorageBackendS3,
		Bucket:      s.bucket,
		Key:         fullKey,
		Endpoint:    s.endpoint,
		Region:      s.region,
		ContentHash: meta.ContentHash,
		ImageTag:    imageTag,
	}, nil
}

// Cleanup removes build contexts older than the specified duration
func (s *S3Store) Cleanup(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	cleaned := 0

	// List objects with the path prefix
	prefix := s.pathPrefix
	if prefix == "" {
		prefix = "build-context/"
	}

	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			buildContextCleanupTotal.WithLabelValues(s.Name(), "list_error").Inc()
			return cleaned, fmt.Errorf("error listing objects: %w", object.Err)
		}

		// Check if object is older than cutoff
		if object.LastModified.Before(cutoff) {
			err := s.client.RemoveObject(ctx, s.bucket, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				buildContextCleanupTotal.WithLabelValues(s.Name(), "delete_error").Inc()
				// Continue cleaning other objects even if one fails
				continue
			}
			cleaned++
		}
	}

	buildContextCleanupTotal.WithLabelValues(s.Name(), "success").Add(float64(cleaned))
	return cleaned, nil
}

// GetClient returns the underlying MinIO client for direct operations
func (s *S3Store) GetClient() *minio.Client {
	return s.client
}

// GetBucket returns the bucket name
func (s *S3Store) GetBucket() string {
	return s.bucket
}

// =============================================================================
// GCSStore Implementation
// =============================================================================

// GCSStore stores build contexts in Google Cloud Storage
type GCSStore struct {
	client     *storage.Client
	bucket     string
	pathPrefix string
	project    string
}

// NewGCSStore creates a new GCS storage backend
func NewGCSStore(ctx context.Context, config *GCSConfig) (*GCSStore, error) {
	if config.Bucket == "" {
		return nil, fmt.Errorf("GCS bucket is required")
	}

	// Create client with credentials if provided
	var client *storage.Client
	var err error

	if len(config.CredentialsJSON) > 0 {
		// Use provided credentials directly (thread-safe, no temp files)
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON(config.CredentialsJSON))
	} else {
		// Use default credentials (workload identity, etc.)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStore{
		client:     client,
		bucket:     config.Bucket,
		pathPrefix: config.PathPrefix,
		project:    config.Project,
	}, nil
}

// Close closes the GCS client and releases resources
func (s *GCSStore) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Name returns the backend name
func (s *GCSStore) Name() string {
	return StorageBackendGCS
}

// Save stores the build context in GCS
func (s *GCSStore) Save(ctx context.Context, key string, tarGz []byte, meta BuildContextMeta) (*BuildContextLocation, error) {
	startTime := time.Now()
	defer func() {
		buildContextCreationDuration.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
	}()

	// Build the full key with prefix
	fullKey := key
	if s.pathPrefix != "" {
		fullKey = strings.TrimSuffix(s.pathPrefix, "/") + "/" + key
	}

	// Create object writer
	obj := s.client.Bucket(s.bucket).Object(fullKey)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/gzip"
	writer.Metadata = map[string]string{
		"lambda-name":      meta.LambdaName,
		"lambda-namespace": meta.LambdaNamespace,
		"content-hash":     meta.ContentHash,
		"created-at":       meta.CreatedAt.Format(time.RFC3339),
	}

	// Write data
	if _, err := writer.Write(tarGz); err != nil {
		writer.Close()
		buildContextStorageErrors.WithLabelValues(s.Name(), "write_failed").Inc()
		return nil, fmt.Errorf("failed to write build context to GCS: %w", err)
	}

	// Close writer to complete upload
	if err := writer.Close(); err != nil {
		buildContextStorageErrors.WithLabelValues(s.Name(), "close_failed").Inc()
		return nil, fmt.Errorf("failed to complete GCS upload: %w", err)
	}

	buildContextSizeBytes.WithLabelValues(s.Name()).Observe(float64(len(tarGz)))

	// Compute image tag (first 12 chars of hash or full hash if shorter)
	imageTag := meta.ContentHash
	if len(imageTag) > 12 {
		imageTag = imageTag[:12]
	}

	return &BuildContextLocation{
		Backend:     StorageBackendGCS,
		Bucket:      s.bucket,
		Key:         fullKey,
		ContentHash: meta.ContentHash,
		ImageTag:    imageTag,
	}, nil
}

// Cleanup removes build contexts older than the specified duration
func (s *GCSStore) Cleanup(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	cleaned := 0

	// List objects with the path prefix
	prefix := s.pathPrefix
	if prefix == "" {
		prefix = "build-context/"
	}

	bucket := s.client.Bucket(s.bucket)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			buildContextCleanupTotal.WithLabelValues(s.Name(), "list_error").Inc()
			return cleaned, fmt.Errorf("error listing GCS objects: %w", err)
		}

		// Check if object is older than cutoff
		if attrs.Updated.Before(cutoff) {
			if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
				buildContextCleanupTotal.WithLabelValues(s.Name(), "delete_error").Inc()
				// Continue cleaning other objects even if one fails
				continue
			}
			cleaned++
		}
	}

	buildContextCleanupTotal.WithLabelValues(s.Name(), "success").Add(float64(cleaned))
	return cleaned, nil
}

// GetClient returns the underlying GCS client for direct operations
func (s *GCSStore) GetClient() *storage.Client {
	return s.client
}

// GetBucket returns the bucket name
func (s *GCSStore) GetBucket() string {
	return s.bucket
}

// Download retrieves a build context from GCS
func (s *GCSStore) Download(ctx context.Context, key string) ([]byte, error) {
	obj := s.client.Bucket(s.bucket).Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read GCS object: %w", err)
	}

	return data, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// RecordSourceFetch records a source fetch operation for metrics
func RecordSourceFetch(sourceType, result string) {
	if result == "success" {
		buildContextCreationTotal.WithLabelValues("", sourceType).Inc()
	} else {
		buildContextSourceErrors.WithLabelValues(sourceType, result).Inc()
	}
}

// RecordStorageSave records a storage save operation for metrics
func RecordStorageSave(storageType, sourceType string) {
	buildContextCreationTotal.WithLabelValues(storageType, sourceType).Inc()
}
