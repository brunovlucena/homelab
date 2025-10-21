// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🚀 SERVICE MANAGER - Focused Knative service lifecycle management
//
//	🎯 Purpose: Handle Knative service creation, management, and lifecycle operations
//	💡 Features: Service creation, resource management, Knative integration
//
//	🏛️ ARCHITECTURE:
//	🚀 Service Lifecycle - Create, update, delete Knative services
//	📦 Resource Management - Service accounts, config maps, triggers
//	🔗 Knative Integration - Knative serving and eventing resources
//	📊 Service Monitoring - Service monitors and metrics
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"knative-lambda-new/internal/config"
	internalerrors "knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/handler/helpers"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
	"knative-lambda-new/pkg/builds"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// 🚀 ServiceManagerImpl - "Focused Knative service lifecycle management"
type ServiceManagerImpl struct {
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *config.KubernetesConfig
	obs           *observability.Observability
	// 🛡️ Rate Limiting Protection
	rateLimiter *resilience.MultiLevelRateLimiter
	// 🎯 Knative Configuration (for main builder service)
	knativeConfig *config.KnativeConfig
	// 🚀 Lambda Services Configuration (for dynamically created lambda services)
	lambdaServicesConfig *config.LambdaServicesConfig
	// 🔗 Notifi Configuration
	notifiConfig *config.NotifiConfig
	// 📊 Metrics Pusher Configuration
	metricsPusherConfig *config.MetricsPusherConfig
}

// 🚀 ServiceManagerConfig - "Configuration for creating service manager"
type ServiceManagerConfig struct {
	K8sClient            kubernetes.Interface
	DynamicClient        dynamic.Interface
	K8sConfig            *config.KubernetesConfig
	Observability        *observability.Observability
	RateLimiter          *resilience.MultiLevelRateLimiter
	KnativeConfig        *config.KnativeConfig
	LambdaServicesConfig *config.LambdaServicesConfig
	NotifiConfig         *config.NotifiConfig
	MetricsPusherConfig  *config.MetricsPusherConfig
}

// 🏗️ NewServiceManager - "Create new service manager with dependencies"
func NewServiceManager(config ServiceManagerConfig) (ServiceManager, error) {
	if config.K8sClient == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "k8s_client", "kubernetes client cannot be nil")
	}

	if config.DynamicClient == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "dynamic_client", "dynamic client cannot be nil")
	}

	if config.K8sConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "k8s_config", "kubernetes config cannot be nil")
	}

	if config.Observability == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "observability", "observability cannot be nil")
	}

	if config.KnativeConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "knative_config", "knative config cannot be nil")
	}

	if config.LambdaServicesConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "lambda_services_config", "lambda services config cannot be nil")
	}

	if config.NotifiConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "notifi_config", "notifi config cannot be nil")
	}

	// MetricsPusher config is optional - can be nil

	return &ServiceManagerImpl{
		k8sClient:            config.K8sClient,
		dynamicClient:        config.DynamicClient,
		config:               config.K8sConfig,
		obs:                  config.Observability,
		rateLimiter:          config.RateLimiter,
		knativeConfig:        config.KnativeConfig,
		lambdaServicesConfig: config.LambdaServicesConfig,
		notifiConfig:         config.NotifiConfig,
		metricsPusherConfig:  config.MetricsPusherConfig,
	}, nil
}

// 🚀 CreateService - "Create or update a Knative service"
func (s *ServiceManagerImpl) CreateService(ctx context.Context, serviceName string, completionData *builds.BuildCompletionEventData) error {
	// Validate input parameters
	if serviceName == "" {
		return internalerrors.NewConfigurationError("service_manager", "service_name", "service name cannot be empty")
	}

	if completionData == nil {
		return internalerrors.NewConfigurationError("service_manager", "completion_data", "completion data cannot be nil")
	}

	ctx, span := s.obs.StartSpanWithAttributes(ctx, "create_or_update_knative_service", map[string]string{
		"k8s.operation":      "create_or_update_service",
		"k8s.service_name":   serviceName,
		"k8s.third_party_id": completionData.ThirdPartyID,
		"k8s.parser_id":      completionData.ParserID,
	})
	defer span.End()

	// Check if service already exists
	ctx, checkSpan := s.obs.StartSpan(ctx, "check_service_exists")
	exists, err := s.CheckServiceExists(ctx, serviceName)
	if err != nil {
		checkSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("check_service_exists_%s", serviceName))
	}
	checkSpan.SetAttributes(attribute.Bool("k8s.service_exists", exists))
	checkSpan.End()

	if exists {
		s.obs.Info(ctx, "Service already exists, updating with new configuration", "service_name", serviceName)
	} else {
		s.obs.Info(ctx, "Creating new Knative service", "service_name", serviceName)
	}

	// Create service account
	ctx, saSpan := s.obs.StartSpan(ctx, "create_service_account")
	serviceAccount := s.CreateServiceAccountResource(serviceName, completionData)
	if serviceAccount == nil {
		saSpan.SetStatus(codes.Error, "failed to create service account resource")
		span.SetStatus(codes.Error, "failed to create service account resource")
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_service_account_%s", serviceName))
	}
	err = s.ApplyResource(ctx, serviceAccount)
	if err != nil {
		saSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_service_account_%s", serviceName))
	}
	saSpan.End()

	// Create config map
	ctx, cmSpan := s.obs.StartSpan(ctx, "create_config_map")
	configMap := s.CreateConfigMapResource(serviceName, completionData)
	if configMap == nil {
		cmSpan.SetStatus(codes.Error, "failed to create config map resource")
		span.SetStatus(codes.Error, "failed to create config map resource")
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_config_map_%s", serviceName))
	}
	err = s.ApplyResource(ctx, configMap)
	if err != nil {
		cmSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_config_map_%s", serviceName))
	}
	cmSpan.End()

	// 🚀 PARALLEL CREATION: Create Knative service and trigger simultaneously
	// This improves performance by not waiting for service creation before creating trigger

	// Create resources in parallel using goroutines
	type resourceResult struct {
		resourceType string
		err          error
	}

	resultChan := make(chan resourceResult, 2)

	// Create Knative service in parallel
	go func() {
		ctx, knativeSpan := s.obs.StartSpan(ctx, "create_or_update_knative_service_resource")
		defer knativeSpan.End()

		knativeService := s.CreateKnativeServiceResource(serviceName, completionData)
		if knativeService == nil {
			knativeSpan.SetStatus(codes.Error, "failed to create knative service resource")
			resultChan <- resourceResult{resourceType: "service", err: internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_knative_service_%s", serviceName))}
			return
		}

		s.obs.Info(ctx, "Applying Knative service with image",
			"service_name", serviceName,
			"image_uri", completionData.ImageURI,
			"content_hash", completionData.ContentHash)

		err := s.ApplyResource(ctx, knativeService)
		if err != nil {
			knativeSpan.SetStatus(codes.Error, err.Error())
			resultChan <- resourceResult{resourceType: "service", err: internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_knative_service_%s", serviceName))}
			return
		}

		knativeSpan.SetStatus(codes.Ok, "service created successfully")
		resultChan <- resourceResult{resourceType: "service", err: nil}
	}()

	// Create trigger in parallel
	go func() {
		ctx, triggerSpan := s.obs.StartSpan(ctx, "create_trigger")
		defer triggerSpan.End()

		trigger := s.CreateTriggerResource(serviceName, completionData)
		if trigger == nil {
			triggerSpan.SetStatus(codes.Error, "failed to create trigger resource")
			s.obs.Error(ctx, errors.New("failed to create trigger resource"), "service_name", serviceName, "third_party_id", completionData.ThirdPartyID, "parser_id", completionData.ParserID)
			resultChan <- resourceResult{resourceType: "trigger", err: internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_trigger_%s", serviceName))}
			return
		}

		s.obs.Info(ctx, "Creating trigger for service", "service_name", serviceName, "trigger_name", trigger.GetName(), "broker", s.knativeConfig.GetDefaultBrokerName())

		err := s.ApplyResource(ctx, trigger)
		if err != nil {
			triggerSpan.SetStatus(codes.Error, err.Error())
			s.obs.Error(ctx, errors.New("failed to apply trigger resource"), "service_name", serviceName, "error", err.Error())
			resultChan <- resourceResult{resourceType: "trigger", err: internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_trigger_%s", serviceName))}
			return
		}

		s.obs.Info(ctx, "Successfully created trigger for service", "service_name", serviceName, "trigger_name", trigger.GetName())
		triggerSpan.SetStatus(codes.Ok, "trigger created successfully")
		resultChan <- resourceResult{resourceType: "trigger", err: nil}
	}()

	// Wait for both resources to be created
	serviceCreated := false
	triggerCreated := false
	var serviceErr, triggerErr error

	for i := 0; i < 2; i++ {
		result := <-resultChan
		switch result.resourceType {
		case "service":
			serviceCreated = true
			serviceErr = result.err
		case "trigger":
			triggerCreated = true
			triggerErr = result.err
		}
	}

	// Check for any errors
	if serviceErr != nil {
		s.obs.Error(ctx, serviceErr, "Failed to create Knative service", "service_name", serviceName)
		return serviceErr
	}

	if triggerErr != nil {
		s.obs.Error(ctx, triggerErr, "Failed to create trigger", "service_name", serviceName)
		return triggerErr
	}

	s.obs.Info(ctx, "Successfully created both Knative service and trigger in parallel",
		"service_name", serviceName,
		"service_created", serviceCreated,
		"trigger_created", triggerCreated)

	if exists {
		span.SetAttributes(attribute.Bool("k8s.service_updated", true))
		s.obs.Info(ctx, "Successfully updated Knative service", "service_name", serviceName, "image_uri", completionData.ImageURI)
	} else {
		span.SetAttributes(attribute.Bool("k8s.service_created", true))
		s.obs.Info(ctx, "Successfully created Knative service", "service_name", serviceName, "image_uri", completionData.ImageURI)
	}
	return nil
}

// 🔍 CheckServiceExists - "Check if a service exists"
func (s *ServiceManagerImpl) CheckServiceExists(ctx context.Context, serviceName string) (bool, error) {
	if serviceName == "" {
		return false, internalerrors.NewConfigurationError("service_manager", "service_name", "service name cannot be empty")
	}

	if s.config == nil {
		return false, internalerrors.NewConfigurationError("service_manager", "config", "kubernetes config cannot be nil")
	}

	if s.dynamicClient == nil {
		return false, internalerrors.NewConfigurationError("service_manager", "dynamic_client", "dynamic client cannot be nil")
	}

	ctx, span := s.obs.StartSpanWithAttributes(ctx, "check_knative_service_exists", map[string]string{
		"k8s.operation":    "check_service_exists",
		"k8s.service_name": serviceName,
	})
	defer span.End()

	// Check Knative service
	knativeServiceGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}

	_, err := s.dynamicClient.Resource(knativeServiceGVR).Namespace(s.config.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			span.SetAttributes(attribute.Bool("k8s.service_exists", false))
			return false, nil
		}
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	span.SetAttributes(attribute.Bool("k8s.service_exists", true))
	return true, nil
}

// 🔧 GenerateServiceName - "Generate unique service name"
func (s *ServiceManagerImpl) GenerateServiceName(thirdPartyID, parserID string) string {
	return helpers.GenerateServiceName(thirdPartyID, parserID)
}

// 📦 CreateServiceAccountResource - "Create a service account resource"
func (s *ServiceManagerImpl) CreateServiceAccountResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured {
	if completionData == nil {
		return nil
	}

	if s.config == nil {
		return nil
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":      serviceName,
				"namespace": s.config.Namespace,
				"labels": map[string]interface{}{
					"app":                                 "knative-lambda-service",
					"build.notifi.network/third-party-id": completionData.ThirdPartyID,
					"build.notifi.network/parser-id":      completionData.ParserID,
				},
			},
		},
	}
}

// 📦 CreateConfigMapResource - "Create a config map resource"
func (s *ServiceManagerImpl) CreateConfigMapResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured {
	if completionData == nil {
		return nil
	}

	if s.config == nil {
		return nil
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-config", serviceName),
				"namespace": s.config.Namespace,
				"labels": map[string]interface{}{
					"app":                                 "knative-lambda-service",
					"build.notifi.network/third-party-id": completionData.ThirdPartyID,
					"build.notifi.network/parser-id":      completionData.ParserID,
				},
			},
			"data": map[string]interface{}{
				"THIRD_PARTY_ID":               completionData.ThirdPartyID,
				"PARSER_ID":                    completionData.ParserID,
				"IMAGE_URI":                    completionData.ImageURI,
				"SUBSCRIPTION_MANAGER_ADDRESS": s.notifiConfig.GetSubscriptionManagerAddress(),
				"EPHEMERAL_STORAGE_ADDRESS":    s.notifiConfig.GetEphemeralStorageAddress(),
				"PERSISTENT_STORAGE_ADDRESS":   s.notifiConfig.GetPersistentStorageAddress(),
				"FUSION_FETCH_PROXY_ADDRESS":   s.notifiConfig.GetFusionFetchProxyAddress(),
				"EVM_RPC_ADDRESS":              s.notifiConfig.GetEvmRpcAddress(),
				"SOLANA_RPC_ADDRESS":           s.notifiConfig.GetSolanaRpcAddress(),
				"SUI_RPC_ADDRESS":              s.notifiConfig.GetSuiRpcAddress(),
				"GRPC_INSECURE":                s.notifiConfig.GetGrpcInsecure(),
			},
		},
	}
}

// 🚀 CreateKnativeServiceResource - "Create a Knative service resource"
func (s *ServiceManagerImpl) CreateKnativeServiceResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured {
	if completionData == nil {
		return nil
	}

	if s.config == nil {
		return nil
	}

	// Create base containers slice
	containers := []map[string]interface{}{
		{
			"name":  "lambda",
			"image": completionData.ImageURI,
			"env": []map[string]interface{}{
				{
					"name":  "SERVICE_NAME",
					"value": serviceName,
				},
				{
					"name":  "THIRD_PARTY_ID",
					"value": completionData.ThirdPartyID,
				},
				{
					"name":  "PARSER_ID",
					"value": completionData.ParserID,
				},
				{
					"name":  "HTTP_PORT",
					"value": "8080",
				},
				{
					"name":  "SUBSCRIPTION_MANAGER_ADDRESS",
					"value": s.notifiConfig.GetSubscriptionManagerAddress(),
				},
				{
					"name":  "EPHEMERAL_STORAGE_ADDRESS",
					"value": s.notifiConfig.GetEphemeralStorageAddress(),
				},
				{
					"name":  "PERSISTENT_STORAGE_ADDRESS",
					"value": s.notifiConfig.GetPersistentStorageAddress(),
				},
				{
					"name":  "FUSION_FETCH_PROXY_ADDRESS",
					"value": s.notifiConfig.GetFusionFetchProxyAddress(),
				},
				{
					"name":  "EVM_RPC_ADDRESS",
					"value": s.notifiConfig.GetEvmRpcAddress(),
				},
				{
					"name":  "SOLANA_RPC_ADDRESS",
					"value": s.notifiConfig.GetSolanaRpcAddress(),
				},
				{
					"name":  "SUI_RPC_ADDRESS",
					"value": s.notifiConfig.GetSuiRpcAddress(),
				},
				{
					"name":  "GRPC_INSECURE",
					"value": s.notifiConfig.GetGrpcInsecure(),
				},
				// 🚀 Lambda Services Autoscaling Configuration
				{
					"name":  "LAMBDA_SERVICES_TARGET_CONCURRENCY",
					"value": s.lambdaServicesConfig.TargetConcurrency,
				},
				{
					"name":  "LAMBDA_SERVICES_TARGET_UTILIZATION",
					"value": s.lambdaServicesConfig.TargetUtilization,
				},
				{
					"name":  "LAMBDA_SERVICES_TARGET",
					"value": s.lambdaServicesConfig.Target,
				},
				{
					"name":  "LAMBDA_SERVICES_CONTAINER_CONCURRENCY",
					"value": s.lambdaServicesConfig.ContainerConcurrency,
				},
				{
					"name":  "LAMBDA_SERVICES_MIN_SCALE",
					"value": s.lambdaServicesConfig.MinScale,
				},
				{
					"name":  "LAMBDA_SERVICES_MAX_SCALE",
					"value": s.lambdaServicesConfig.MaxScale,
				},
				{
					"name":  "LAMBDA_SERVICES_SCALE_TO_ZERO_GRACE_PERIOD",
					"value": s.lambdaServicesConfig.ScaleToZeroGracePeriod,
				},
				{
					"name":  "LAMBDA_SERVICES_SCALE_DOWN_DELAY",
					"value": s.lambdaServicesConfig.ScaleDownDelay,
				},
				{
					"name":  "LAMBDA_SERVICES_STABLE_WINDOW",
					"value": s.lambdaServicesConfig.StableWindow,
				},
			},
			"ports": []map[string]interface{}{
				{
					"containerPort": 8080, // Use port 8081 for consistency
					"name":          "http1",
				},
			},
			"resources": map[string]interface{}{
				"limits": map[string]interface{}{
					"cpu":    s.lambdaServicesConfig.ResourceCPULimit,
					"memory": s.lambdaServicesConfig.ResourceMemoryLimit,
				},
				"requests": map[string]interface{}{
					"cpu":    s.lambdaServicesConfig.ResourceCPURequest,
					"memory": s.lambdaServicesConfig.ResourceMemoryRequest,
				},
			},
			"readinessProbe": map[string]interface{}{
				"tcpSocket": map[string]interface{}{
					"port": 8080,
				},
				"successThreshold": 5,
			},
		},
	}

	// Add metrics-pusher sidecar if enabled
	if s.metricsPusherConfig != nil && s.metricsPusherConfig.Enabled {
		// Add metrics-pusher sidecar container
		metricsPusherContainer := map[string]interface{}{
			"name":  "metrics-pusher",
			"image": s.metricsPusherConfig.GetImageURL(),
			"env": []map[string]interface{}{
				{
					"name":  "PROMETHEUS_REMOTE_WRITE_URL",
					"value": s.metricsPusherConfig.RemoteWriteURL,
				},
				{
					"name":  "PUSH_INTERVAL",
					"value": s.metricsPusherConfig.GetPushIntervalString(),
				},
				{
					"name":  "TIMEOUT",
					"value": s.metricsPusherConfig.GetTimeoutString(),
				},
				{
					"name":  "LOG_LEVEL",
					"value": s.metricsPusherConfig.LogLevel,
				},
				{
					"name":  "NAMESPACE",
					"value": s.config.Namespace,
				},
				{
					"name":  "SERVICE_NAME",
					"value": serviceName,
				},
				{
					"name":  "THIRD_PARTY_ID",
					"value": completionData.ThirdPartyID,
				},
				{
					"name":  "PARSER_ID",
					"value": completionData.ParserID,
				},
				{
					"name":  "QUEUE_PROXY_METRICS_PORT",
					"value": s.metricsPusherConfig.QueueProxyMetricsPort,
				},
				{
					"name":  "QUEUE_PROXY_METRICS_PATH",
					"value": s.metricsPusherConfig.QueueProxyMetricsPath,
				},
				{
					"name":  "METRICS_PORT",
					"value": "8080",
				},
				{
					"name":  "METRICS_PATH",
					"value": "/metrics",
				},
				{
					"name":  "METRICS_PUSHER_ENABLED",
					"value": "true",
				},
				{
					"name":  "METRICS_PUSHER_FAILURE_TOLERANCE",
					"value": s.metricsPusherConfig.GetFailureToleranceString(),
				},
			},
			"resources": map[string]interface{}{
				"limits": map[string]interface{}{
					"cpu":    s.metricsPusherConfig.ResourceCPULimit,
					"memory": s.metricsPusherConfig.ResourceMemoryLimit,
				},
				"requests": map[string]interface{}{
					"cpu":    s.metricsPusherConfig.ResourceCPURequest,
					"memory": s.metricsPusherConfig.ResourceMemoryRequest,
				},
			},
		}

		containers = append(containers, metricsPusherContainer)
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      serviceName,
				"namespace": s.config.Namespace,
				"labels": map[string]interface{}{
					"app":                                 "knative-lambda-service",
					"build.notifi.network/third-party-id": completionData.ThirdPartyID,
					"build.notifi.network/parser-id":      completionData.ParserID,
					"build.notifi.network/content-hash":   completionData.ContentHash, // New: track content hash
				},
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app":                                 "knative-lambda-service",
							"build.notifi.network/third-party-id": completionData.ThirdPartyID,
							"build.notifi.network/parser-id":      completionData.ParserID,
							"build.notifi.network/content-hash":   completionData.ContentHash, // New: track content hash
						},
						"annotations": map[string]interface{}{
							"autoscaling.knative.dev/class":                    "kpa.autoscaling.knative.dev",
							"autoscaling.knative.dev/target":                   s.lambdaServicesConfig.Target,
							"autoscaling.knative.dev/targetUtilization":        s.lambdaServicesConfig.TargetUtilization,
							"autoscaling.knative.dev/minScale":                 s.lambdaServicesConfig.MinScale,
							"autoscaling.knative.dev/maxScale":                 s.lambdaServicesConfig.MaxScale,
							"autoscaling.knative.dev/scaleToZeroGracePeriod":   s.lambdaServicesConfig.ScaleToZeroGracePeriod,
							"autoscaling.knative.dev/scaleDownDelay":           s.lambdaServicesConfig.ScaleDownDelay,
							"autoscaling.knative.dev/stableWindow":             s.lambdaServicesConfig.StableWindow,
							"autoscaling.knative.dev/panicWindowPercentage":    s.lambdaServicesConfig.PanicWindowPercentage,
							"autoscaling.knative.dev/panicThresholdPercentage": s.lambdaServicesConfig.PanicThresholdPercentage,
						},
					},
					"spec": map[string]interface{}{
						"serviceAccountName": serviceName,
						// ⚠️  TODO: Before going to prd fine-tune this.
						"containerConcurrency": s.lambdaServicesConfig.GetContainerConcurrencyInt(),
						// 🔒 Best practice: Disable automatic service discovery to prevent environment variable pollution
						"enableServiceLinks": s.knativeConfig.GetEnableServiceLinks(),
						"containers":         containers,
					},
				},
			},
		},
	}
}

// 🔗 CreateTriggerResource - "Create a trigger resource"
func (s *ServiceManagerImpl) CreateTriggerResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured {
	if completionData == nil {
		return nil
	}

	if s.config == nil {
		return nil
	}

	if s.knativeConfig == nil {
		return nil
	}

	// Create Knative Trigger that subscribes to the service broker (not builder broker)
	// This matches the logic from fix-triggers.yaml
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.knative.dev/v1",
			"kind":       "Trigger",
			"metadata": map[string]interface{}{
				"name":      serviceName,
				"namespace": s.config.Namespace,
				"annotations": map[string]interface{}{
					"rabbitmq.eventing.knative.dev/parallelism": s.getRabbitMQParallelism(),
				},
				"labels": map[string]interface{}{
					"app":                                 "knative-lambda-service",
					"build.notifi.network/third-party-id": completionData.ThirdPartyID,
					"build.notifi.network/parser-id":      completionData.ParserID,
					"build.notifi.network/content-hash":   completionData.ContentHash, // New: track content hash
				},
			},
			"spec": map[string]interface{}{
				"broker": s.knativeConfig.GetDefaultBrokerName(), // This should be knative-lambda-service-broker-dev
				"filter": map[string]interface{}{
					"attributes": map[string]interface{}{
						"type":    s.knativeConfig.GetDefaultEventType(), // network.notifi.lambda.parser.start
						"source":  fmt.Sprintf("network.notifi.%s", completionData.ThirdPartyID),
						"subject": fmt.Sprintf("%s", completionData.ParserID),
					},
				},
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"apiVersion": "serving.knative.dev/v1",
						"kind":       "Service",
						"name":       serviceName,
						"namespace":  s.config.Namespace,
					},
				},
				"delivery": map[string]interface{}{
					"retry":         5,
					"backoffPolicy": "exponential",
					"backoffDelay":  "PT1S",
				},
			},
		},
	}
}

// 🔧 ApplyResource - "Apply a Kubernetes resource"
func (s *ServiceManagerImpl) ApplyResource(ctx context.Context, obj *unstructured.Unstructured) error {
	if obj == nil {
		return internalerrors.NewConfigurationError("service_manager", "resource", "kubernetes resource cannot be nil")
	}

	if s.config == nil {
		return internalerrors.NewConfigurationError("service_manager", "config", "kubernetes config cannot be nil")
	}

	if s.dynamicClient == nil {
		return internalerrors.NewConfigurationError("service_manager", "dynamic_client", "dynamic client cannot be nil")
	}

	ctx, span := s.obs.StartSpanWithAttributes(ctx, "apply_k8s_resource", map[string]string{
		"k8s.operation":          "apply_resource",
		"k8s.resource_kind":      obj.GetKind(),
		"k8s.resource_name":      obj.GetName(),
		"k8s.resource_namespace": obj.GetNamespace(),
	})
	defer span.End()

	// Parse API version to get group and version
	apiVersion := obj.GetAPIVersion()
	parts := strings.Split(apiVersion, "/")

	var group, version string
	if len(parts) == 2 {
		// Group/Version format (e.g., "apps/v1")
		group = parts[0]
		version = parts[1]
	} else if len(parts) == 1 {
		// Version only format (e.g., "v1")
		group = ""
		version = parts[0]
	} else {
		span.SetStatus(codes.Error, fmt.Sprintf("invalid_api_version_%s", apiVersion))
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("invalid_api_version_%s", apiVersion))
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: strings.ToLower(obj.GetKind()) + "s",
	}

	s.obs.Info(ctx, "Applying Kubernetes resource",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
		"api_version", apiVersion,
		"gvr_group", gvr.Group,
		"gvr_version", gvr.Version,
		"gvr_resource", gvr.Resource)

	// Check if resource exists and get it for proper update handling
	existingObj, err := s.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		// Resource doesn't exist, create it
		ctx, createSpan := s.obs.StartSpan(ctx, "create_k8s_resource")
		_, err = s.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			createSpan.SetStatus(codes.Error, err.Error())
			span.SetStatus(codes.Error, err.Error())
			s.obs.Error(ctx, err, "Failed to create Kubernetes resource", "kind", obj.GetKind(), "name", obj.GetName())
			return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("create_%s_%s: %v", strings.ToLower(obj.GetKind()), obj.GetName(), err))
		}
		createSpan.SetAttributes(attribute.Bool("k8s.resource_created", true))
		createSpan.End()
	} else {
		// Resource exists, update it with proper resourceVersion handling
		ctx, updateSpan := s.obs.StartSpan(ctx, "update_k8s_resource")
		s.obs.Info(ctx, "Updating existing Kubernetes resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())

		// For Knative services, we need to be more careful with updates
		if obj.GetKind() == "Service" && obj.GetAPIVersion() == "serving.knative.dev/v1" {
			s.obs.Info(ctx, "Updating Knative service - ensuring image update",
				"service_name", obj.GetName())
		}

		// Preserve the resourceVersion from the existing resource for optimistic concurrency control
		resourceVersion := existingObj.GetResourceVersion()
		if resourceVersion == "" {
			updateSpan.SetStatus(codes.Error, "resource version is required for update")
			span.SetStatus(codes.Error, "resource version is required for update")
			s.obs.Error(ctx, errors.New("resource version is required for update"), "Failed to update Kubernetes resource - missing resourceVersion", "kind", obj.GetKind(), "name", obj.GetName())
			return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("update_%s_%s: %v", strings.ToLower(obj.GetKind()), obj.GetName(), "resource version is required for update"))
		}

		// Set the resourceVersion on the object to be updated
		obj.SetResourceVersion(resourceVersion)

		// For Knative services, preserve existing annotations to avoid immutable annotation conflicts
		if obj.GetKind() == "Service" && obj.GetAPIVersion() == "serving.knative.dev/v1" {
			existingAnnotations := existingObj.GetAnnotations()
			if existingAnnotations != nil {
				// Get current annotations from the object to be updated
				currentAnnotations := obj.GetAnnotations()
				if currentAnnotations == nil {
					currentAnnotations = make(map[string]string)
				}

				// Preserve immutable annotations from the existing service
				immutableAnnotations := []string{
					"serving.knative.dev/creator",
					"serving.knative.dev/lastModifier",
					"serving.knative.dev/creatorTimestamp",
				}

				for _, annotationKey := range immutableAnnotations {
					if value, exists := existingAnnotations[annotationKey]; exists {
						currentAnnotations[annotationKey] = value
					}
				}

				// Set the merged annotations back to the object
				obj.SetAnnotations(currentAnnotations)
			}
		}

		_, err = s.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Update(ctx, obj, metav1.UpdateOptions{})
		if err != nil {
			updateSpan.SetStatus(codes.Error, err.Error())
			span.SetStatus(codes.Error, err.Error())
			s.obs.Error(ctx, err, "Failed to update Kubernetes resource", "kind", obj.GetKind(), "name", obj.GetName())
			return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("update_%s_%s: %v", strings.ToLower(obj.GetKind()), obj.GetName(), err))
		}
		updateSpan.SetAttributes(attribute.Bool("k8s.resource_updated", true))
		s.obs.Info(ctx, "Successfully updated Kubernetes resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())
		updateSpan.End()
	}

	return nil
}

// 🗑️ DeleteService - "Delete a Knative service and all associated resources"
func (s *ServiceManagerImpl) DeleteService(ctx context.Context, serviceName string) error {
	// Validate input parameters
	if serviceName == "" {
		return internalerrors.NewConfigurationError("service_manager", "service_name", "service name cannot be empty")
	}

	ctx, span := s.obs.StartSpanWithAttributes(ctx, "delete_knative_service", map[string]string{
		"k8s.operation":    "delete_service",
		"k8s.service_name": serviceName,
	})
	defer span.End()

	// Check if service exists before attempting deletion
	ctx, checkSpan := s.obs.StartSpan(ctx, "check_service_exists_before_delete")
	exists, err := s.CheckServiceExists(ctx, serviceName)
	if err != nil {
		checkSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("check_service_exists_%s", serviceName))
	}
	checkSpan.SetAttributes(attribute.Bool("k8s.service_exists", exists))
	checkSpan.End()

	if !exists {
		s.obs.Info(ctx, "Service does not exist, nothing to delete", "service_name", serviceName)
		span.SetAttributes(attribute.Bool("k8s.service_deleted", false))
		return nil
	}

	s.obs.Info(ctx, "Deleting Knative service and associated resources", "service_name", serviceName)

	// Delete resources in reverse order of creation to handle dependencies properly
	// 1. Delete Trigger first (depends on Service)
	ctx, triggerSpan := s.obs.StartSpan(ctx, "delete_trigger")
	if err := s.deleteTrigger(ctx, serviceName); err != nil {
		triggerSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		s.obs.Error(ctx, err, "Failed to delete trigger", "service_name", serviceName)
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("delete_trigger_%s", serviceName))
	}
	triggerSpan.End()

	// 2. Delete Knative Service
	ctx, serviceSpan := s.obs.StartSpan(ctx, "delete_knative_service_resource")
	if err := s.deleteKnativeService(ctx, serviceName); err != nil {
		serviceSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		s.obs.Error(ctx, err, "Failed to delete Knative service", "service_name", serviceName)
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("delete_knative_service_%s", serviceName))
	}
	serviceSpan.End()

	// 3. Delete ConfigMap
	ctx, cmSpan := s.obs.StartSpan(ctx, "delete_config_map")
	if err := s.deleteConfigMap(ctx, serviceName); err != nil {
		cmSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		s.obs.Error(ctx, err, "Failed to delete config map", "service_name", serviceName)
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("delete_config_map_%s", serviceName))
	}
	cmSpan.End()

	// 4. Delete ServiceAccount
	ctx, saSpan := s.obs.StartSpan(ctx, "delete_service_account")
	if err := s.deleteServiceAccount(ctx, serviceName); err != nil {
		saSpan.SetStatus(codes.Error, err.Error())
		span.SetStatus(codes.Error, err.Error())
		s.obs.Error(ctx, err, "Failed to delete service account", "service_name", serviceName)
		return internalerrors.NewSystemError("kubernetes", fmt.Sprintf("delete_service_account_%s", serviceName))
	}
	saSpan.End()

	span.SetAttributes(attribute.Bool("k8s.service_deleted", true))
	s.obs.Info(ctx, "Successfully deleted Knative service and all associated resources", "service_name", serviceName)
	return nil
}

// 🗑️ deleteTrigger - "Delete a Knative trigger"
func (s *ServiceManagerImpl) deleteTrigger(ctx context.Context, serviceName string) error {
	triggerGVR := schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1",
		Resource: "triggers",
	}

	s.obs.Info(ctx, "Deleting Knative trigger", "service_name", serviceName, "trigger_name", serviceName)

	err := s.dynamicClient.Resource(triggerGVR).Namespace(s.config.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.obs.Info(ctx, "Trigger not found, skipping deletion", "service_name", serviceName)
			return nil
		}
		return err
	}

	s.obs.Info(ctx, "Successfully deleted Knative trigger", "service_name", serviceName)
	return nil
}

// 🗑️ deleteKnativeService - "Delete a Knative service"
func (s *ServiceManagerImpl) deleteKnativeService(ctx context.Context, serviceName string) error {
	knativeServiceGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}

	s.obs.Info(ctx, "Deleting Knative service", "service_name", serviceName)

	err := s.dynamicClient.Resource(knativeServiceGVR).Namespace(s.config.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.obs.Info(ctx, "Knative service not found, skipping deletion", "service_name", serviceName)
			return nil
		}
		return err
	}

	s.obs.Info(ctx, "Successfully deleted Knative service", "service_name", serviceName)
	return nil
}

// 🗑️ deleteConfigMap - "Delete a config map"
func (s *ServiceManagerImpl) deleteConfigMap(ctx context.Context, serviceName string) error {
	configMapName := fmt.Sprintf("%s-config", serviceName)
	configMapGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}

	s.obs.Info(ctx, "Deleting config map", "service_name", serviceName, "config_map_name", configMapName)

	err := s.dynamicClient.Resource(configMapGVR).Namespace(s.config.Namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.obs.Info(ctx, "Config map not found, skipping deletion", "service_name", serviceName)
			return nil
		}
		return err
	}

	s.obs.Info(ctx, "Successfully deleted config map", "service_name", serviceName)
	return nil
}

// 🗑️ deleteServiceAccount - "Delete a service account"
func (s *ServiceManagerImpl) deleteServiceAccount(ctx context.Context, serviceName string) error {
	serviceAccountGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "serviceaccounts",
	}

	s.obs.Info(ctx, "Deleting service account", "service_name", serviceName)

	err := s.dynamicClient.Resource(serviceAccountGVR).Namespace(s.config.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.obs.Info(ctx, "Service account not found, skipping deletion", "service_name", serviceName)
			return nil
		}
		return err
	}

	s.obs.Info(ctx, "Successfully deleted service account", "service_name", serviceName)
	return nil
}

// 🔧 getRabbitMQParallelism - "Get RabbitMQ parallelism from config or fallback to constant"
// TODO: REFACTOR THIS! DRY!
func (s *ServiceManagerImpl) getRabbitMQParallelism() string {
	// Try to get from Knative config first
	if s.knativeConfig != nil {
		parallelism := s.knativeConfig.GetRabbitMQEventingParallelism()
		if parallelism > 0 {
			return strconv.Itoa(parallelism)
		}
	}

	// Fallback to constant from constants.go
	return "50" // RabbitMQEventingParallelismDefault from constants.go
}
