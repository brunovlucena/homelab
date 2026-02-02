// Package eventing manages Knative Eventing infrastructure for lambda functions.
//
// ARCHITECTURE FOR SCALE:
// - Uses SHARED BROKER per namespace (not per-lambda) to handle 10M+ lambdas
// - Each lambda only creates ONE trigger pointing to the shared broker
// - DLQ resources are shared at namespace level
// - This reduces resources from ~13 per lambda to ~1 per lambda
package eventing

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

const (
	// SharedBrokerName is the name of the shared broker per namespace
	// All lambdas in a namespace share this broker
	SharedBrokerName = "lambda-broker"

	// SharedDLQPrefix is the prefix for shared DLQ resources
	SharedDLQPrefix = "lambda-dlq"
)

// Manager handles eventing infrastructure operations
type Manager struct {
	client   client.Client
	scheme   *runtime.Scheme
	renderer *TemplateRenderer
	config   *Config

	// Mutex for namespace-level operations (broker creation)
	namespaceLocks sync.Map
}

// Config holds default configuration for eventing resources
type Config struct {
	// Default RabbitMQ configuration
	DefaultRabbitMQCluster       string
	DefaultRabbitMQNamespace     string
	DefaultRabbitMQQueueType     string
	DefaultRabbitMQParallelism   int // Concurrent event deliveries per trigger dispatcher
	DefaultRabbitMQPrefetchCount int // Messages prefetched from RabbitMQ before acks

	// Default DLQ configuration
	DefaultDLQEnabled           bool
	DefaultDLQExchangeName      string
	DefaultDLQQueueName         string
	DefaultDLQRoutingKeyPrefix  string
	DefaultDLQRetryMaxAttempts  int
	DefaultDLQRetryBackoffDelay string
	DefaultDLQMessageTTL        int
	DefaultDLQMaxLength         int

	// Default event types
	DefaultEventSource string
}

// DefaultConfig returns default eventing configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultRabbitMQCluster:       "rabbitmq",
		DefaultRabbitMQNamespace:     "rabbitmq-system",
		DefaultRabbitMQQueueType:     "quorum",
		DefaultRabbitMQParallelism:   50,  // High parallelism for better autoscaling
		DefaultRabbitMQPrefetchCount: 100, // High prefetch for throughput
		DefaultDLQEnabled:            true,
		DefaultDLQExchangeName:       SharedDLQPrefix + "-exchange",
		DefaultDLQQueueName:          SharedDLQPrefix + "-queue",
		DefaultDLQRoutingKeyPrefix:   "io.knative.lambda.dlq",
		DefaultDLQRetryMaxAttempts:   5,
		DefaultDLQRetryBackoffDelay:  "PT1S",
		DefaultDLQMessageTTL:         604800000, // 7 days
		DefaultDLQMaxLength:          100000,    // Higher for shared DLQ
		// CloudEvents source format: io.knative.lambda/operator
		DefaultEventSource: "io.knative.lambda/operator",
	}
}

// NewManager creates a new eventing manager
func NewManager(client client.Client, scheme *runtime.Scheme) (*Manager, error) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create template renderer: %w", err)
	}

	return &Manager{
		client:   client,
		scheme:   scheme,
		renderer: renderer,
		config:   DefaultConfig(),
	}, nil
}

// NewManagerWithConfig creates a new eventing manager with custom config
func NewManagerWithConfig(client client.Client, scheme *runtime.Scheme, config *Config) (*Manager, error) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create template renderer: %w", err)
	}

	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		client:   client,
		scheme:   scheme,
		renderer: renderer,
		config:   config,
	}, nil
}

// ReconcileEventing ensures eventing resources exist for a LambdaFunction
//
// SCALE ARCHITECTURE:
// 1. Ensures SHARED namespace broker exists (created once per namespace)
// 2. Creates ONE trigger for this specific lambda
// 3. Shared DLQ at namespace level
func (m *Manager) ReconcileEventing(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	// Skip if eventing disabled
	if lambda.Spec.Eventing != nil && !lambda.Spec.Eventing.Enabled {
		return nil
	}

	// Ensure shared namespace broker exists
	if err := m.ensureSharedBroker(ctx, lambda); err != nil {
		return fmt.Errorf("failed to ensure shared broker: %w", err)
	}

	// Create trigger for this specific lambda (the only per-lambda resource)
	if err := m.reconcileLambdaTrigger(ctx, lambda); err != nil {
		return fmt.Errorf("failed to reconcile lambda trigger: %w", err)
	}

	return nil
}

// DeleteEventing removes eventing resources for a LambdaFunction
// Only deletes the lambda-specific trigger, NOT the shared broker
func (m *Manager) DeleteEventing(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	// Only delete the lambda-specific trigger
	if err := m.deleteLambdaTrigger(ctx, lambda); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete lambda trigger: %w", err)
	}

	// NOTE: Do NOT delete the shared broker or DLQ - other lambdas depend on them
	// Cleanup of shared resources should be done separately (e.g., namespace deletion)

	return nil
}

// ensureSharedBroker creates the shared broker for a namespace if it doesn't exist
func (m *Manager) ensureSharedBroker(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	namespace := lambda.Namespace

	// Get namespace-level lock to prevent race conditions
	lock, _ := m.namespaceLocks.LoadOrStore(namespace, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if broker already exists
	brokerName := m.getSharedBrokerName(lambda)
	broker := &unstructured.Unstructured{}
	broker.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1",
		Kind:    "Broker",
	})

	err := m.client.Get(ctx, types.NamespacedName{Name: brokerName, Namespace: namespace}, broker)
	if err == nil {
		// Broker exists
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Create shared broker and DLQ
	data := m.buildSharedBrokerData(lambda)
	rendered, err := m.renderer.RenderBroker(data)
	if err != nil {
		return fmt.Errorf("failed to render broker template: %w", err)
	}

	// Apply broker manifests (no owner reference - shared resource)
	if err := m.applySharedManifests(ctx, namespace, rendered); err != nil {
		return err
	}

	// Create shared DLQ if enabled
	if m.isDLQEnabled(lambda) {
		dlqData := m.buildSharedDLQData(lambda)
		dlqRendered, err := m.renderer.RenderDLQ(dlqData)
		if err != nil {
			return fmt.Errorf("failed to render DLQ template: %w", err)
		}
		if err := m.applySharedManifests(ctx, namespace, dlqRendered); err != nil {
			return err
		}
	}

	return nil
}

// reconcileLambdaTrigger creates or updates the trigger for a specific lambda
func (m *Manager) reconcileLambdaTrigger(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	triggerName := m.getLambdaTriggerName(lambda)
	brokerName := m.getSharedBrokerName(lambda)

	// For receiver mode, create a catch-all trigger (no filter) - receiver handles routing internally
	// For regular lambdas, filter by subject
	var filterAttrs map[string]interface{}
	if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
		// Catch-all trigger for receiver - no filter, receives all events from broker
		filterAttrs = nil
	} else {
		// Regular lambda - filter by subject
		filterAttrs = map[string]interface{}{
			"subject": lambda.Name,
		}
	}

	// Build trigger spec
	triggerSpec := map[string]interface{}{
		"broker": brokerName,
		"subscriber": map[string]interface{}{
			"ref": map[string]interface{}{
				"apiVersion": "serving.knative.dev/v1",
				"kind":       "Service",
				"name":       lambda.Name,
				"namespace":  lambda.Namespace,
			},
		},
	}

	// Only add filter if not receiver mode (catch-all for receiver)
	if filterAttrs != nil {
		triggerSpec["filter"] = map[string]interface{}{
			"attributes": filterAttrs,
		}
	}

	// Build trigger object
	trigger := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.knative.dev/v1",
			"kind":       "Trigger",
			"metadata": map[string]interface{}{
				"name":      triggerName,
				"namespace": lambda.Namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       lambda.Name,
					"app.kubernetes.io/component":  "trigger",
					"app.kubernetes.io/managed-by": "knative-lambda-operator",
					"lambda.knative.io/name":       lambda.Name,
				},
				"annotations": map[string]interface{}{
					"rabbitmq.eventing.knative.dev/parallelism": fmt.Sprintf("%d", m.getParallelism(lambda)),
				},
			},
			"spec": triggerSpec,
		},
	}

	// Set owner reference so trigger is deleted with lambda
	if err := controllerutil.SetControllerReference(lambda, trigger, m.scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create or update trigger
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(trigger.GetObjectKind().GroupVersionKind())

	err := m.client.Get(ctx, types.NamespacedName{Name: triggerName, Namespace: lambda.Namespace}, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			return m.client.Create(ctx, trigger)
		}
		return err
	}

	// Update existing
	trigger.SetResourceVersion(existing.GetResourceVersion())
	return m.client.Update(ctx, trigger)
}

// deleteLambdaTrigger deletes the trigger for a specific lambda
func (m *Manager) deleteLambdaTrigger(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	trigger := &unstructured.Unstructured{}
	trigger.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1",
		Kind:    "Trigger",
	})
	trigger.SetName(m.getLambdaTriggerName(lambda))
	trigger.SetNamespace(lambda.Namespace)

	return m.client.Delete(ctx, trigger)
}

// applySharedManifests applies YAML manifests without owner references (shared resources)
func (m *Manager) applySharedManifests(ctx context.Context, namespace string, manifests string) error {
	// Split on YAML document separator (--- on its own line)
	// Template may start with "---" or have "\n---\n" between documents
	manifests = strings.TrimSpace(manifests)

	// Split on document separator - handle both "\n---\n" and standalone "---\n" at start
	// First, normalize: replace all "---\n" patterns with a marker
	normalized := strings.ReplaceAll(manifests, "\n---\n", "\n__DOC_SEP__\n")
	normalized = strings.ReplaceAll(normalized, "---\n", "__DOC_SEP__\n")
	normalized = strings.ReplaceAll(normalized, "\n---", "\n__DOC_SEP__")

	// Split on the marker
	docs := strings.Split(normalized, "__DOC_SEP__")

	// Remove the first empty doc if template starts with "---"
	if len(docs) > 0 && strings.TrimSpace(docs[0]) == "" {
		docs = docs[1:]
	}

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		// Skip empty docs or docs that are only comments
		if doc == "" {
			continue
		}
		// Skip if entire doc is just comments (but allow docs that start with comments)
		lines := strings.Split(doc, "\n")
		nonCommentLines := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				nonCommentLines++
			}
		}
		if nonCommentLines == 0 {
			continue
		}

		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			return fmt.Errorf("failed to unmarshal YAML document %d: %w\nDocument:\n%s", i, err, doc)
		}

		if len(obj.Object) == 0 {
			continue
		}

		// Ensure namespace is set
		if obj.GetNamespace() == "" {
			obj.SetNamespace(namespace)
		}

		// Set GVK from the unmarshaled object
		apiVersion, _, _ := unstructured.NestedString(obj.Object, "apiVersion")
		kind, _, _ := unstructured.NestedString(obj.Object, "kind")
		var gvk schema.GroupVersionKind
		if apiVersion != "" && kind != "" {
			parts := strings.Split(apiVersion, "/")
			if len(parts) == 2 {
				gvk.Group = parts[0]
				gvk.Version = parts[1]
			} else {
				gvk.Version = apiVersion
			}
			gvk.Kind = kind
			obj.SetGroupVersionKind(gvk)
		} else {
			// Fallback to GetObjectKind if apiVersion/kind not found
			gvk = obj.GetObjectKind().GroupVersionKind()
		}

		// NO owner reference - shared resources are not owned by any single lambda

		// Create or update
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(gvk)

		key := types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		}

		err := m.client.Get(ctx, key, existing)
		if err != nil {
			if errors.IsNotFound(err) {
				if err := m.client.Create(ctx, obj); err != nil {
					return fmt.Errorf("failed to create %s/%s: %w", obj.GetKind(), obj.GetName(), err)
				}
			} else {
				return fmt.Errorf("failed to get %s/%s: %w", obj.GetKind(), obj.GetName(), err)
			}
		}
		// Don't update existing shared resources - they're managed at namespace level
	}

	return nil
}

// Naming helpers

func (m *Manager) getSharedBrokerName(lambda *lambdav1alpha1.LambdaFunction) string {
	// Use custom broker name if specified, otherwise use shared default
	if lambda.Spec.Eventing != nil && lambda.Spec.Eventing.BrokerName != "" {
		return lambda.Spec.Eventing.BrokerName
	}
	return SharedBrokerName
}

func (m *Manager) getLambdaTriggerName(lambda *lambdav1alpha1.LambdaFunction) string {
	return lambda.Name + "-trigger"
}

func (m *Manager) getParallelism(lambda *lambdav1alpha1.LambdaFunction) int {
	if lambda.Spec.Eventing != nil && lambda.Spec.Eventing.RabbitMQ != nil && lambda.Spec.Eventing.RabbitMQ.Parallelism > 0 {
		return lambda.Spec.Eventing.RabbitMQ.Parallelism
	}
	return m.config.DefaultRabbitMQParallelism
}

func (m *Manager) isDLQEnabled(lambda *lambdav1alpha1.LambdaFunction) bool {
	if lambda.Spec.Eventing == nil || lambda.Spec.Eventing.DLQ == nil {
		return m.config.DefaultDLQEnabled
	}
	return lambda.Spec.Eventing.DLQ.Enabled
}

// Data builders for templates

func (m *Manager) buildSharedBrokerData(lambda *lambdav1alpha1.LambdaFunction) BrokerData {
	return BrokerData{
		Name:       SharedBrokerName,
		Namespace:  lambda.Namespace,
		LambdaName: "shared", // Shared broker
		RabbitMQ:   m.buildRabbitMQConfig(lambda),
		DLQ:        m.buildDLQConfig(lambda),
	}
}

func (m *Manager) buildSharedDLQData(lambda *lambdav1alpha1.LambdaFunction) DLQData {
	monitoringEnabled := false
	if lambda.Spec.Eventing != nil && lambda.Spec.Eventing.Monitoring != nil {
		monitoringEnabled = lambda.Spec.Eventing.Monitoring.Enabled
	}

	return DLQData{
		Name:       SharedDLQPrefix,
		Namespace:  lambda.Namespace,
		LambdaName: "shared",
		RabbitMQ:   m.buildRabbitMQConfig(lambda),
		DLQ:        m.buildDLQConfig(lambda),
		Monitoring: MonitoringConfig{Enabled: monitoringEnabled},
	}
}

func (m *Manager) buildRabbitMQConfig(lambda *lambdav1alpha1.LambdaFunction) RabbitMQConfig {
	config := RabbitMQConfig{
		ClusterName:   m.config.DefaultRabbitMQCluster,
		Namespace:     m.config.DefaultRabbitMQNamespace,
		QueueType:     m.config.DefaultRabbitMQQueueType,
		Parallelism:   m.config.DefaultRabbitMQParallelism,
		PrefetchCount: m.config.DefaultRabbitMQPrefetchCount,
	}

	if lambda.Spec.Eventing != nil && lambda.Spec.Eventing.RabbitMQ != nil {
		rmq := lambda.Spec.Eventing.RabbitMQ
		if rmq.ClusterName != "" {
			config.ClusterName = rmq.ClusterName
		}
		if rmq.Namespace != "" {
			config.Namespace = rmq.Namespace
		}
		if rmq.QueueType != "" {
			config.QueueType = rmq.QueueType
		}
		if rmq.Parallelism > 0 {
			config.Parallelism = rmq.Parallelism
		}
		if rmq.PrefetchCount > 0 {
			config.PrefetchCount = rmq.PrefetchCount
		}
	}

	return config
}

func (m *Manager) buildDLQConfig(lambda *lambdav1alpha1.LambdaFunction) DLQConfig {
	config := DLQConfig{
		Enabled:                m.config.DefaultDLQEnabled,
		ExchangeName:           m.config.DefaultDLQExchangeName,
		QueueName:              m.config.DefaultDLQQueueName,
		RoutingKeyPrefix:       m.config.DefaultDLQRoutingKeyPrefix,
		RetryMaxAttempts:       m.config.DefaultDLQRetryMaxAttempts,
		RetryBackoffDelay:      m.config.DefaultDLQRetryBackoffDelay,
		RetryBackoffMultiplier: 2,
		BackoffPolicy:          "exponential",
		RetryPolicy:            "exponential",
		MessageTTL:             m.config.DefaultDLQMessageTTL,
		MaxLength:              m.config.DefaultDLQMaxLength,
		OverflowPolicy:         "reject-publish",
		AlertThreshold:         100,
		DepthThreshold:         1000,
		AgeThreshold:           "24h",
		Cleanup: CleanupConfig{
			Enabled:   false,
			Interval:  "1h",
			Retention: "168h",
		},
	}

	if lambda.Spec.Eventing != nil && lambda.Spec.Eventing.DLQ != nil {
		dlq := lambda.Spec.Eventing.DLQ
		config.Enabled = dlq.Enabled
		if dlq.ExchangeName != "" {
			config.ExchangeName = dlq.ExchangeName
		}
		if dlq.QueueName != "" {
			config.QueueName = dlq.QueueName
		}
		if dlq.RoutingKeyPrefix != "" {
			config.RoutingKeyPrefix = dlq.RoutingKeyPrefix
		}
		if dlq.RetryMaxAttempts > 0 {
			config.RetryMaxAttempts = dlq.RetryMaxAttempts
		}
		if dlq.RetryBackoffDelay != "" {
			config.RetryBackoffDelay = dlq.RetryBackoffDelay
		}
		if dlq.MessageTTL > 0 {
			config.MessageTTL = dlq.MessageTTL
		}
		if dlq.MaxLength > 0 {
			config.MaxLength = dlq.MaxLength
		}
		if dlq.OverflowPolicy != "" {
			config.OverflowPolicy = dlq.OverflowPolicy
		}
		if dlq.Cleanup != nil {
			config.Cleanup.Enabled = dlq.Cleanup.Enabled
			if dlq.Cleanup.Interval != "" {
				config.Cleanup.Interval = dlq.Cleanup.Interval
			}
			if dlq.Cleanup.Retention != "" {
				config.Cleanup.Retention = dlq.Cleanup.Retention
			}
		}
	}

	return config
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  🤖 LAMBDA AGENT EVENTING
//
//  LambdaAgents are AI agents with pre-built images. They have:
//  - A dedicated Broker per namespace (agent-<name>-broker)
//  - Triggers for each subscription (inbound events)
//  - Cross-namespace forwarding via Channels/Subscriptions
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ReconcileAgentEventing ensures eventing resources exist for a LambdaAgent
func (m *Manager) ReconcileAgentEventing(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	// Skip if eventing disabled
	if agent.Spec.Eventing == nil || !agent.Spec.Eventing.Enabled {
		return nil
	}

	// 1. Ensure broker exists for this agent
	if err := m.ensureAgentBroker(ctx, agent); err != nil {
		return fmt.Errorf("failed to ensure agent broker: %w", err)
	}

	// 2. Create triggers for each subscription (inbound events)
	if err := m.reconcileAgentSubscriptions(ctx, agent); err != nil {
		return fmt.Errorf("failed to reconcile agent subscriptions: %w", err)
	}

	// 3. Create forwarding infrastructure (outbound to other agents)
	if err := m.reconcileAgentForwards(ctx, agent); err != nil {
		return fmt.Errorf("failed to reconcile agent forwards: %w", err)
	}

	return nil
}

// DeleteAgentEventing removes eventing resources for a LambdaAgent
func (m *Manager) DeleteAgentEventing(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	// Delete all triggers
	if err := m.deleteAgentTriggers(ctx, agent); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete agent triggers: %w", err)
	}

	// Delete forwarding channels
	if err := m.deleteAgentChannels(ctx, agent); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete agent channels: %w", err)
	}

	// NOTE: Broker is NOT deleted - it may be shared or recreated on next reconcile
	// Cleanup of broker should be done on namespace deletion

	return nil
}

// ensureAgentBroker creates the broker for a LambdaAgent
// Uses RabbitmqBrokerConfig for proper RabbitMQ integration
func (m *Manager) ensureAgentBroker(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	namespace := agent.Namespace
	brokerName := m.getAgentBrokerName(agent)

	// Get namespace-level lock to prevent race conditions
	lock, _ := m.namespaceLocks.LoadOrStore(namespace, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if broker already exists
	broker := &unstructured.Unstructured{}
	broker.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1",
		Kind:    "Broker",
	})

	err := m.client.Get(ctx, types.NamespacedName{Name: brokerName, Namespace: namespace}, broker)
	if err == nil {
		// Broker exists
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Build RabbitMQ config from agent spec
	rmqConfig := m.buildAgentRabbitMQConfig(agent)

	// Create RabbitmqBrokerConfig (required by RabbitMQ eventing webhook)
	brokerConfigName := brokerName + "-config"
	brokerConfig := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.knative.dev/v1alpha1",
			"kind":       "RabbitmqBrokerConfig",
			"metadata": map[string]interface{}{
				"name":      brokerConfigName,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       agent.Name,
					"app.kubernetes.io/managed-by": "knative-lambda-operator",
					"app.kubernetes.io/component":  "broker-config",
				},
			},
			"spec": map[string]interface{}{
				"rabbitmqClusterReference": map[string]interface{}{
					"name":      rmqConfig.ClusterName,
					"namespace": rmqConfig.Namespace,
				},
				"queueType": rmqConfig.QueueType,
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(agent, brokerConfig, m.scheme); err != nil {
		return fmt.Errorf("failed to set owner reference for broker config: %w", err)
	}

	// Create or update RabbitmqBrokerConfig
	existingConfig := &unstructured.Unstructured{}
	existingConfig.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1alpha1",
		Kind:    "RabbitmqBrokerConfig",
	})
	err = m.client.Get(ctx, types.NamespacedName{Name: brokerConfigName, Namespace: namespace}, existingConfig)
	if errors.IsNotFound(err) {
		if err := m.client.Create(ctx, brokerConfig); err != nil {
			return fmt.Errorf("failed to create RabbitmqBrokerConfig: %w", err)
		}
	}

	// Create broker referencing the RabbitmqBrokerConfig
	brokerObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.knative.dev/v1",
			"kind":       "Broker",
			"metadata": map[string]interface{}{
				"name":      brokerName,
				"namespace": namespace,
				"annotations": map[string]interface{}{
					"eventing.knative.dev/broker.class": "RabbitMQBroker",
				},
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       agent.Name,
					"app.kubernetes.io/managed-by": "knative-lambda-operator",
					"app.kubernetes.io/component":  "broker",
				},
			},
			"spec": map[string]interface{}{
				"config": map[string]interface{}{
					"apiVersion": "eventing.knative.dev/v1alpha1",
					"kind":       "RabbitmqBrokerConfig",
					"name":       brokerConfigName,
				},
			},
		},
	}

	// Set owner reference so broker is deleted with agent
	if err := controllerutil.SetControllerReference(agent, brokerObj, m.scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	if err := m.client.Create(ctx, brokerObj); err != nil {
		return fmt.Errorf("failed to create broker: %w", err)
	}

	return nil
}

// reconcileAgentSubscriptions creates Triggers for each subscription
func (m *Manager) reconcileAgentSubscriptions(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	if agent.Spec.Eventing == nil || len(agent.Spec.Eventing.Subscriptions) == 0 {
		return nil
	}

	brokerName := m.getAgentBrokerName(agent)

	for _, sub := range agent.Spec.Eventing.Subscriptions {
		triggerName := fmt.Sprintf("%s-%s", agent.Name, sanitizeEventType(sub.EventType))

		// Build filter attributes
		filterAttrs := map[string]interface{}{
			"type": sub.EventType,
		}
		if sub.Source != "" {
			filterAttrs["source"] = sub.Source
		}

		// Build trigger
		trigger := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "eventing.knative.dev/v1",
				"kind":       "Trigger",
				"metadata": map[string]interface{}{
					"name":      triggerName,
					"namespace": agent.Namespace,
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       agent.Name,
						"app.kubernetes.io/managed-by": "knative-lambda-operator",
						"app.kubernetes.io/component":  "trigger",
						"lambda.knative.io/event-type": sanitizeEventType(sub.EventType),
					},
				},
				"spec": map[string]interface{}{
					"broker": brokerName,
					"filter": map[string]interface{}{
						"attributes": filterAttrs,
					},
					"subscriber": map[string]interface{}{
						"ref": map[string]interface{}{
							"apiVersion": "serving.knative.dev/v1",
							"kind":       "Service",
							"name":       agent.Name,
						},
					},
				},
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(agent, trigger, m.scheme); err != nil {
			return fmt.Errorf("failed to set owner reference for trigger %s: %w", triggerName, err)
		}

		// Create or update
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "eventing.knative.dev",
			Version: "v1",
			Kind:    "Trigger",
		})

		err := m.client.Get(ctx, types.NamespacedName{Name: triggerName, Namespace: agent.Namespace}, existing)
		if err != nil {
			if errors.IsNotFound(err) {
				if err := m.client.Create(ctx, trigger); err != nil {
					return fmt.Errorf("failed to create trigger %s: %w", triggerName, err)
				}
			} else {
				return err
			}
		} else {
			trigger.SetResourceVersion(existing.GetResourceVersion())
			if err := m.client.Update(ctx, trigger); err != nil {
				return fmt.Errorf("failed to update trigger %s: %w", triggerName, err)
			}
		}
	}

	return nil
}

// reconcileAgentForwards creates cross-namespace forwarding infrastructure
func (m *Manager) reconcileAgentForwards(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	if agent.Spec.Eventing == nil || len(agent.Spec.Eventing.Forwards) == 0 {
		return nil
	}

	brokerName := m.getAgentBrokerName(agent)

	for _, fwd := range agent.Spec.Eventing.Forwards {
		// Create a trigger for each event type that forwards to the target namespace
		for _, eventType := range fwd.EventTypes {
			triggerName := fmt.Sprintf("%s-fwd-%s-%s", agent.Name, fwd.TargetAgent, sanitizeEventType(eventType))

			// Target broker URL in another namespace
			targetBrokerURL := fmt.Sprintf("http://%s-broker-ingress.%s.svc.cluster.local", fwd.TargetAgent, fwd.TargetNamespace)

			// Build trigger that forwards to external URL
			trigger := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "eventing.knative.dev/v1",
					"kind":       "Trigger",
					"metadata": map[string]interface{}{
						"name":      triggerName,
						"namespace": agent.Namespace,
						"labels": map[string]interface{}{
							"app.kubernetes.io/name":         agent.Name,
							"app.kubernetes.io/managed-by":   "knative-lambda-operator",
							"app.kubernetes.io/component":    "trigger-forward",
							"lambda.knative.io/target-agent": fwd.TargetAgent,
							"lambda.knative.io/target-ns":    fwd.TargetNamespace,
						},
					},
					"spec": map[string]interface{}{
						"broker": brokerName,
						"filter": map[string]interface{}{
							"attributes": map[string]interface{}{
								"type": eventType,
							},
						},
						"subscriber": map[string]interface{}{
							"uri": targetBrokerURL,
						},
					},
				},
			}

			// Set owner reference
			if err := controllerutil.SetControllerReference(agent, trigger, m.scheme); err != nil {
				return fmt.Errorf("failed to set owner reference for forward trigger %s: %w", triggerName, err)
			}

			// Create or update
			existing := &unstructured.Unstructured{}
			existing.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "eventing.knative.dev",
				Version: "v1",
				Kind:    "Trigger",
			})

			err := m.client.Get(ctx, types.NamespacedName{Name: triggerName, Namespace: agent.Namespace}, existing)
			if err != nil {
				if errors.IsNotFound(err) {
					if err := m.client.Create(ctx, trigger); err != nil {
						return fmt.Errorf("failed to create forward trigger %s: %w", triggerName, err)
					}
				} else {
					return err
				}
			} else {
				trigger.SetResourceVersion(existing.GetResourceVersion())
				if err := m.client.Update(ctx, trigger); err != nil {
					return fmt.Errorf("failed to update forward trigger %s: %w", triggerName, err)
				}
			}
		}
	}

	return nil
}

// deleteAgentTriggers deletes all triggers for a LambdaAgent
func (m *Manager) deleteAgentTriggers(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	// List and delete all triggers with label matching this agent
	triggerList := &unstructured.UnstructuredList{}
	triggerList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1",
		Kind:    "TriggerList",
	})

	if err := m.client.List(ctx, triggerList,
		client.InNamespace(agent.Namespace),
		client.MatchingLabels{"app.kubernetes.io/name": agent.Name}); err != nil {
		return err
	}

	for _, trigger := range triggerList.Items {
		if err := m.client.Delete(ctx, &trigger); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// deleteAgentChannels deletes all channels for a LambdaAgent
func (m *Manager) deleteAgentChannels(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) error {
	// List and delete all channels with label matching this agent
	channelList := &unstructured.UnstructuredList{}
	channelList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "messaging.knative.dev",
		Version: "v1",
		Kind:    "ChannelList",
	})

	if err := m.client.List(ctx, channelList,
		client.InNamespace(agent.Namespace),
		client.MatchingLabels{"app.kubernetes.io/name": agent.Name}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	for _, channel := range channelList.Items {
		if err := m.client.Delete(ctx, &channel); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// Helper functions for LambdaAgent

func (m *Manager) getAgentBrokerName(agent *lambdav1alpha1.LambdaAgent) string {
	return agent.Name + "-broker"
}

func (m *Manager) buildAgentRabbitMQConfig(agent *lambdav1alpha1.LambdaAgent) RabbitMQConfig {
	config := RabbitMQConfig{
		ClusterName: m.config.DefaultRabbitMQCluster,
		Namespace:   m.config.DefaultRabbitMQNamespace,
		QueueType:   m.config.DefaultRabbitMQQueueType,
		Parallelism: m.config.DefaultRabbitMQParallelism,
	}

	if agent.Spec.Eventing != nil && agent.Spec.Eventing.RabbitMQ != nil {
		rmq := agent.Spec.Eventing.RabbitMQ
		if rmq.ClusterName != "" {
			config.ClusterName = rmq.ClusterName
		}
		if rmq.Namespace != "" {
			config.Namespace = rmq.Namespace
		}
	}

	return config
}

// sanitizeEventType converts event type to a valid K8s name suffix
func sanitizeEventType(eventType string) string {
	// Replace dots and slashes with dashes, lowercase
	result := strings.ToLower(eventType)
	result = strings.ReplaceAll(result, ".", "-")
	result = strings.ReplaceAll(result, "/", "-")
	result = strings.ReplaceAll(result, "_", "-")
	// Replace wildcards with 'all' for valid K8s resource names
	result = strings.ReplaceAll(result, "*", "all")
	// Truncate to reasonable length
	if len(result) > 40 {
		result = result[:40]
	}
	return result
}

// GetAgentEventingStatus returns the eventing status for a LambdaAgent
func (m *Manager) GetAgentEventingStatus(ctx context.Context, agent *lambdav1alpha1.LambdaAgent) (*lambdav1alpha1.AgentEventingStatus, error) {
	status := &lambdav1alpha1.AgentEventingStatus{
		BrokerName: m.getAgentBrokerName(agent),
	}

	// Check broker status
	broker := &unstructured.Unstructured{}
	broker.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1",
		Kind:    "Broker",
	})

	err := m.client.Get(ctx, types.NamespacedName{
		Name:      status.BrokerName,
		Namespace: agent.Namespace,
	}, broker)

	if err == nil {
		// Check if ready
		conditions, found, _ := unstructured.NestedSlice(broker.Object, "status", "conditions")
		if found {
			for _, c := range conditions {
				cond, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				if cond["type"] == "Ready" && cond["status"] == "True" {
					status.BrokerReady = true
					break
				}
			}
		}
		// Get broker URL
		url, found, _ := unstructured.NestedString(broker.Object, "status", "address", "url")
		if found {
			status.BrokerURL = url
		}
	}

	// Count triggers
	triggerList := &unstructured.UnstructuredList{}
	triggerList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.knative.dev",
		Version: "v1",
		Kind:    "TriggerList",
	})

	if err := m.client.List(ctx, triggerList,
		client.InNamespace(agent.Namespace),
		client.MatchingLabels{"app.kubernetes.io/name": agent.Name}); err == nil {
		status.TriggerCount = len(triggerList.Items)
	}

	// Count forwards
	forwardCount := 0
	for _, trigger := range triggerList.Items {
		labels := trigger.GetLabels()
		if labels != nil && labels["app.kubernetes.io/component"] == "trigger-forward" {
			forwardCount++
		}
	}
	status.ForwardCount = forwardCount

	return status, nil
}

// Unused import guard
var _ metav1.Object = (*unstructured.Unstructured)(nil)
