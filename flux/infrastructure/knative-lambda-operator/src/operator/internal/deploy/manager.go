package deploy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

const (
	// DefaultPort is the default port for Lambda functions
	DefaultPort = 8080

	// DefaultConcurrency is the default target concurrency
	DefaultConcurrency = 10

	// DefaultMinReplicas is the default minimum replicas
	DefaultMinReplicas = 0

	// DefaultMaxReplicas is the default maximum replicas
	DefaultMaxReplicas = 50

	// ReceiverServiceAccountName is the dedicated SA for receiver mode lambdas
	// This SA has LIMITED permissions - only what's needed for event processing
	// NOT the operator's full cluster-admin-like permissions
	// See: k8s/base for RBAC manifests (Flux).
	// Security Fix: VULN-013 - Receiver Mode SA Escalation
	ReceiverServiceAccountName = "knative-lambda-receiver"

	// DefaultLambdaServiceAccountName is the SA for regular lambda functions
	// This SA has NO special permissions by default
	DefaultLambdaServiceAccountName = "knative-lambda-function"
)

var (
	// KnativeServiceGVK is the GroupVersionKind for Knative Services
	KnativeServiceGVK = schema.GroupVersionKind{
		Group:   "serving.knative.dev",
		Version: "v1",
		Kind:    "Service",
	}
)

// Manager handles Knative service deployment operations
type Manager struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewManager creates a new deploy manager
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		client: client,
		scheme: scheme,
	}
}

// GetService retrieves the Knative Service for a Lambda function
func (m *Manager) GetService(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) (*unstructured.Unstructured, error) {
	service := &unstructured.Unstructured{}
	service.SetGroupVersionKind(KnativeServiceGVK)

	err := m.client.Get(ctx, types.NamespacedName{
		Name:      lambda.Name,
		Namespace: lambda.Namespace,
	}, service)

	if err != nil {
		return nil, err
	}

	return service, nil
}

// GetServiceImage extracts the container image from a Knative Service
func (m *Manager) GetServiceImage(service *unstructured.Unstructured) string {
	if service == nil {
		return ""
	}
	// Navigate: spec.template.spec.containers[0].image
	spec, ok := service.Object["spec"].(map[string]interface{})
	if !ok {
		return ""
	}
	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		return ""
	}
	templateSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return ""
	}
	containers, ok := templateSpec["containers"].([]interface{})
	if !ok || len(containers) == 0 {
		return ""
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		return ""
	}
	image, _ := container["image"].(string)
	return image
}

// UpdateServiceImage updates the container image in a Knative Service
// CRITICAL: This keeps lambda-command-receiver in sync with operator version
func (m *Manager) UpdateServiceImage(ctx context.Context, service *unstructured.Unstructured, newImage string) error {
	if service == nil {
		return fmt.Errorf("service is nil")
	}
	// Navigate and update: spec.template.spec.containers[0].image
	spec, ok := service.Object["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid service spec")
	}
	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid service template")
	}
	templateSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid service template spec")
	}
	containers, ok := templateSpec["containers"].([]interface{})
	if !ok || len(containers) == 0 {
		return fmt.Errorf("no containers in service")
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid container spec")
	}
	container["image"] = newImage
	return m.client.Update(ctx, service)
}

// CreateService creates a Knative Service for a Lambda function
func (m *Manager) CreateService(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) (*unstructured.Unstructured, error) {
	if lambda.Status.BuildStatus == nil || lambda.Status.BuildStatus.ImageURI == "" {
		return nil, fmt.Errorf("no image URI available")
	}

	// Build service configuration
	serviceName := lambda.Name
	imageURI := lambda.Status.BuildStatus.ImageURI

	// Get scaling configuration
	// containerConcurrency = HARD LIMIT of concurrent requests per pod
	// targetConcurrency = when to scale up (autoscaling target)
	minReplicas := int64(DefaultMinReplicas)
	maxReplicas := int64(DefaultMaxReplicas)
	containerConcurrency := int64(DefaultConcurrency) // Hard limit per pod
	targetConcurrency := int64(DefaultConcurrency)    // Autoscaling target

	if lambda.Spec.Scaling != nil {
		if lambda.Spec.Scaling.MinReplicas != nil {
			minReplicas = int64(*lambda.Spec.Scaling.MinReplicas)
		}
		if lambda.Spec.Scaling.MaxReplicas != nil {
			maxReplicas = int64(*lambda.Spec.Scaling.MaxReplicas)
		}
		if lambda.Spec.Scaling.ContainerConcurrency != nil {
			containerConcurrency = int64(*lambda.Spec.Scaling.ContainerConcurrency)
		}
		if lambda.Spec.Scaling.TargetConcurrency != nil {
			targetConcurrency = int64(*lambda.Spec.Scaling.TargetConcurrency)
		}
		// If containerConcurrency not set but targetConcurrency is, use target as container limit
		if lambda.Spec.Scaling.ContainerConcurrency == nil && lambda.Spec.Scaling.TargetConcurrency != nil {
			containerConcurrency = targetConcurrency
		}
	}

	// Get resource configuration
	memoryRequest := "64Mi"
	cpuRequest := "50m"
	memoryLimit := "128Mi"
	cpuLimit := "100m"

	if lambda.Spec.Resources != nil {
		if lambda.Spec.Resources.Requests != nil {
			if lambda.Spec.Resources.Requests.Memory != "" {
				memoryRequest = lambda.Spec.Resources.Requests.Memory
			}
			if lambda.Spec.Resources.Requests.CPU != "" {
				cpuRequest = lambda.Spec.Resources.Requests.CPU
			}
		}
		if lambda.Spec.Resources.Limits != nil {
			if lambda.Spec.Resources.Limits.Memory != "" {
				memoryLimit = lambda.Spec.Resources.Limits.Memory
			}
			if lambda.Spec.Resources.Limits.CPU != "" {
				cpuLimit = lambda.Spec.Resources.Limits.CPU
			}
		}
	}

	// Build environment variables
	// Note: PORT is reserved by Knative and set automatically
	env := make([]corev1.EnvVar, 0, len(lambda.Spec.Env)+2)

	// Only add HANDLER for non-image sources (image sources have their own entrypoint)
	if lambda.Spec.Source.Type != "image" {
		env = append(env, corev1.EnvVar{
			Name:  "HANDLER",
			Value: lambda.Spec.Runtime.Handler,
		})
	}
	env = append(env, lambda.Spec.Env...)

	// Get container port (default 8080, or from image source config)
	containerPort := int64(DefaultPort)
	if lambda.Spec.Source.Type == "image" && lambda.Spec.Source.Image != nil && lambda.Spec.Source.Image.Port > 0 {
		containerPort = int64(lambda.Spec.Source.Image.Port)
	}

	// Create the Knative Service using unstructured
	service := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      serviceName,
				"namespace": lambda.Namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       lambda.Name,
					"app.kubernetes.io/component":  "lambda",
					"app.kubernetes.io/managed-by": "knative-lambda-operator",
					"lambda.knative.io/name":       lambda.Name,
				},
				"annotations": map[string]interface{}{
					"lambda.knative.io/runtime": lambda.Spec.Runtime.Language,
					"lambda.knative.io/version": lambda.Spec.Runtime.Version,
					// Skip tag resolution to avoid Knative trying to resolve digest from localhost
					"serving.knative.dev/creator":      "knative-lambda-operator",
					"serving.knative.dev/lastModifier": "knative-lambda-operator",
				},
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"autoscaling.knative.dev/min-scale": fmt.Sprintf("%d", minReplicas),
							"autoscaling.knative.dev/max-scale": fmt.Sprintf("%d", maxReplicas),
							"autoscaling.knative.dev/target":    fmt.Sprintf("%d", targetConcurrency),
							"autoscaling.knative.dev/class":     "kpa.autoscaling.knative.dev",
							"autoscaling.knative.dev/metric":    "concurrency",
						},
					},
					"spec": func() map[string]interface{} {
						spec := map[string]interface{}{
							"containerConcurrency": containerConcurrency, // Hard limit per pod
							"containers": []interface{}{
								m.buildContainerSpec(lambda, imageURI, env, memoryRequest, cpuRequest, memoryLimit, cpuLimit, containerPort),
							},
						}

						// Security Fix: VULN-013 - Use dedicated receiver SA instead of operator SA
						// Security Fix: VULN-003 - Disable SA token mounting for regular lambdas
						if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
							// Receiver mode: uses dedicated limited SA (NOT the operator's SA!)
							// The knative-lambda-receiver SA has minimal permissions for event processing
							spec["serviceAccountName"] = ReceiverServiceAccountName
							// Add imagePullSecrets for ghcr.io images
							spec["imagePullSecrets"] = []interface{}{
								map[string]interface{}{"name": "ghcr-secret"},
							}
						} else {
							// Regular lambda: use minimal SA and disable token mounting
							// This prevents SA token theft from inline code execution (VULN-003)
							spec["serviceAccountName"] = DefaultLambdaServiceAccountName
							spec["automountServiceAccountToken"] = false
						}
						return spec
					}(),
				},
			},
		},
	}

	// Set owner reference so the service gets cleaned up when the LambdaFunction is deleted
	if err := controllerutil.SetControllerReference(lambda, service, m.scheme); err != nil {
		return nil, fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create the Knative Service
	if err := m.client.Create(ctx, service); err != nil {
		if errors.IsAlreadyExists(err) {
			// Service already exists, update it
			return m.UpdateService(ctx, lambda)
		}
		return nil, fmt.Errorf("failed to create Knative Service: %w", err)
	}

	return service, nil
}

// buildContainerSpec builds the container spec for a Lambda function
// If receiver mode is enabled, adds --mode=receiver args
// For image sources, supports custom command, args, and port
func (m *Manager) buildContainerSpec(lambda *lambdav1alpha1.LambdaFunction, imageURI string, env []corev1.EnvVar, memoryRequest, cpuRequest, memoryLimit, cpuLimit string, containerPort int64) map[string]interface{} {
	// Determine image pull policy:
	// - If explicitly set in spec, use that
	// - For built images (minio, s3, git, inline): default to "Always" to avoid stale cache
	// - For pre-built images (source.type=image): default to "IfNotPresent"
	imagePullPolicy := lambda.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		if lambda.Spec.Source.Type == "image" {
			imagePullPolicy = "IfNotPresent"
		} else {
			imagePullPolicy = "Always"
		}
	}

	container := map[string]interface{}{
		"image":           imageURI,
		"imagePullPolicy": imagePullPolicy,
		"ports": []interface{}{
			map[string]interface{}{
				"containerPort": containerPort,
				"protocol":      "TCP",
			},
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": memoryRequest,
				"cpu":    cpuRequest,
			},
			"limits": map[string]interface{}{
				"memory": memoryLimit,
				"cpu":    cpuLimit,
			},
		},
		"env": m.convertEnvVars(env),
		"readinessProbe": map[string]interface{}{
			"httpGet": map[string]interface{}{
				"path": "/health",
				"port": containerPort,
			},
			"initialDelaySeconds": int64(5),
			"periodSeconds":       int64(10),
		},
	}

	// Handle image source: custom command, args, and health endpoints
	if lambda.Spec.Source.Type == "image" && lambda.Spec.Source.Image != nil {
		imageSource := lambda.Spec.Source.Image

		// Set custom command if specified
		if len(imageSource.Command) > 0 {
			container["command"] = imageSource.Command
		}

		// Set custom args if specified
		if len(imageSource.Args) > 0 {
			container["args"] = imageSource.Args
		}
	}

	// Check if receiver mode is enabled
	if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
		// Add receiver mode args
		args := []string{
			"--mode=receiver",
		}

		// Add default namespace if specified in env
		for _, e := range env {
			if e.Name == "DEFAULT_NAMESPACE" && e.Value != "" {
				args = append(args, fmt.Sprintf("--default-namespace=%s", e.Value))
				break
			}
		}

		// Add rate limit if specified in env
		for _, e := range env {
			if e.Name == "RATE_LIMIT" && e.Value != "" {
				args = append(args, fmt.Sprintf("--rate-limit=%s", e.Value))
				break
			}
		}

		// Add burst size if specified in env
		for _, e := range env {
			if e.Name == "BURST_SIZE" && e.Value != "" {
				args = append(args, fmt.Sprintf("--burst-size=%s", e.Value))
				break
			}
		}

		container["args"] = args

		// Update readiness probe for receiver mode (uses /ready endpoint)
		container["readinessProbe"] = map[string]interface{}{
			"httpGet": map[string]interface{}{
				"path": "/ready",
				"port": containerPort,
			},
			"initialDelaySeconds": int64(2),
			"periodSeconds":       int64(5),
		}

		// Update liveness probe for receiver mode (uses /health endpoint)
		container["livenessProbe"] = map[string]interface{}{
			"httpGet": map[string]interface{}{
				"path": "/health",
				"port": containerPort,
			},
			"initialDelaySeconds": int64(5),
			"periodSeconds":       int64(10),
		}
	}

	return container
}

// GetServiceStatus returns the status of a Knative Service
func (m *Manager) GetServiceStatus(ctx context.Context, service *unstructured.Unstructured) (ready bool, url string, replicas int32, err error) {
	// Re-fetch the service to get the latest status
	err = m.client.Get(ctx, types.NamespacedName{
		Name:      service.GetName(),
		Namespace: service.GetNamespace(),
	}, service)

	if err != nil {
		return false, "", 0, err
	}

	// Extract status from unstructured
	status, found, err := unstructured.NestedMap(service.Object, "status")
	if err != nil || !found {
		return false, "", 0, nil
	}

	// Get URL
	if urlVal, ok := status["url"].(string); ok {
		url = urlVal
	}

	// Check conditions for Ready status
	conditions, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil || !found {
		return false, url, 0, nil
	}

	for _, c := range conditions {
		condition, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _ := condition["type"].(string)
		condStatus, _ := condition["status"].(string)

		if condType == "Ready" && condStatus == "True" {
			ready = true
			break
		}
	}

	// Try to get replicas from traffic (Knative doesn't expose replicas directly in status)
	// Default to 1 if ready, 0 if not
	if ready {
		replicas = 1
	}

	return ready, url, replicas, nil
}

// DeleteService deletes the Knative Service for a Lambda function
func (m *Manager) DeleteService(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	service := &unstructured.Unstructured{}
	service.SetGroupVersionKind(KnativeServiceGVK)
	service.SetName(lambda.Name)
	service.SetNamespace(lambda.Namespace)

	if err := m.client.Delete(ctx, service); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

// UpdateService updates the Knative Service for a Lambda function
func (m *Manager) UpdateService(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) (*unstructured.Unstructured, error) {
	// Get the existing service first
	existing, err := m.GetService(ctx, lambda)
	if err != nil {
		if errors.IsNotFound(err) {
			// Service doesn't exist, create it
			return m.CreateService(ctx, lambda)
		}
		return nil, err
	}

	if lambda.Status.BuildStatus == nil || lambda.Status.BuildStatus.ImageURI == "" {
		return nil, fmt.Errorf("no image URI available")
	}

	imageURI := lambda.Status.BuildStatus.ImageURI

	// Get scaling configuration
	minReplicas := int64(DefaultMinReplicas)
	maxReplicas := int64(DefaultMaxReplicas)
	targetConcurrency := int64(DefaultConcurrency)

	if lambda.Spec.Scaling != nil {
		if lambda.Spec.Scaling.MinReplicas != nil {
			minReplicas = int64(*lambda.Spec.Scaling.MinReplicas)
		}
		if lambda.Spec.Scaling.MaxReplicas != nil {
			maxReplicas = int64(*lambda.Spec.Scaling.MaxReplicas)
		}
		if lambda.Spec.Scaling.TargetConcurrency != nil {
			targetConcurrency = int64(*lambda.Spec.Scaling.TargetConcurrency)
		}
	}

	// Get resource configuration
	memoryRequest := "64Mi"
	cpuRequest := "50m"
	memoryLimit := "128Mi"
	cpuLimit := "100m"

	if lambda.Spec.Resources != nil {
		if lambda.Spec.Resources.Requests != nil {
			if lambda.Spec.Resources.Requests.Memory != "" {
				memoryRequest = lambda.Spec.Resources.Requests.Memory
			}
			if lambda.Spec.Resources.Requests.CPU != "" {
				cpuRequest = lambda.Spec.Resources.Requests.CPU
			}
		}
		if lambda.Spec.Resources.Limits != nil {
			if lambda.Spec.Resources.Limits.Memory != "" {
				memoryLimit = lambda.Spec.Resources.Limits.Memory
			}
			if lambda.Spec.Resources.Limits.CPU != "" {
				cpuLimit = lambda.Spec.Resources.Limits.CPU
			}
		}
	}

	// Build environment variables
	// Note: PORT is reserved by Knative and set automatically
	env := make([]corev1.EnvVar, 0, len(lambda.Spec.Env)+2)

	// Only add HANDLER for non-image sources (image sources have their own entrypoint)
	if lambda.Spec.Source.Type != "image" {
		env = append(env, corev1.EnvVar{
			Name:  "HANDLER",
			Value: lambda.Spec.Runtime.Handler,
		})
	}
	env = append(env, lambda.Spec.Env...)

	// Get container port (default 8080, or from image source config)
	containerPort := int64(DefaultPort)
	if lambda.Spec.Source.Type == "image" && lambda.Spec.Source.Image != nil && lambda.Spec.Source.Image.Port > 0 {
		containerPort = int64(lambda.Spec.Source.Image.Port)
	}

	// Update the spec
	spec := map[string]interface{}{
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"autoscaling.knative.dev/min-scale": fmt.Sprintf("%d", minReplicas),
					"autoscaling.knative.dev/max-scale": fmt.Sprintf("%d", maxReplicas),
					"autoscaling.knative.dev/target":    fmt.Sprintf("%d", targetConcurrency),
					"autoscaling.knative.dev/class":     "kpa.autoscaling.knative.dev",
					"autoscaling.knative.dev/metric":    "concurrency",
				},
			},
			"spec": func() map[string]interface{} {
				spec := map[string]interface{}{
					"containerConcurrency": targetConcurrency,
					"containers": []interface{}{
						m.buildContainerSpec(lambda, imageURI, env, memoryRequest, cpuRequest, memoryLimit, cpuLimit, containerPort),
					},
				}

				// Security Fix: VULN-013 - Use dedicated receiver SA instead of operator SA
				// Security Fix: VULN-003 - Disable SA token mounting for regular lambdas
				if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
					// Receiver mode: uses dedicated limited SA (NOT the operator's SA!)
					spec["serviceAccountName"] = ReceiverServiceAccountName
				} else {
					// Regular lambda: use minimal SA and disable token mounting
					spec["serviceAccountName"] = DefaultLambdaServiceAccountName
					spec["automountServiceAccountToken"] = false
				}
				return spec
			}(),
		},
	}

	if err := unstructured.SetNestedField(existing.Object, spec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to set spec: %w", err)
	}

	// Update the service
	if err := m.client.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update Knative Service: %w", err)
	}

	return existing, nil
}

// convertEnvVars converts corev1.EnvVar to map format for unstructured
func (m *Manager) convertEnvVars(envVars []corev1.EnvVar) []interface{} {
	result := make([]interface{}, 0, len(envVars))
	for _, ev := range envVars {
		envMap := map[string]interface{}{
			"name": ev.Name,
		}
		if ev.Value != "" {
			envMap["value"] = ev.Value
		}
		if ev.ValueFrom != nil {
			valueFrom := map[string]interface{}{}
			if ev.ValueFrom.SecretKeyRef != nil {
				valueFrom["secretKeyRef"] = map[string]interface{}{
					"name": ev.ValueFrom.SecretKeyRef.Name,
					"key":  ev.ValueFrom.SecretKeyRef.Key,
				}
			}
			if ev.ValueFrom.ConfigMapKeyRef != nil {
				valueFrom["configMapKeyRef"] = map[string]interface{}{
					"name": ev.ValueFrom.ConfigMapKeyRef.Name,
					"key":  ev.ValueFrom.ConfigMapKeyRef.Key,
				}
			}
			if len(valueFrom) > 0 {
				envMap["valueFrom"] = valueFrom
			}
		}
		result = append(result, envMap)
	}
	return result
}

// parseResourceQuantity parses a resource quantity string
func parseResourceQuantity(value string) resource.Quantity {
	q, err := resource.ParseQuantity(value)
	if err != nil {
		return resource.Quantity{}
	}
	return q
}
