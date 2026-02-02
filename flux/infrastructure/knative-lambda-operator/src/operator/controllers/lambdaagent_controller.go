/*
Copyright 2024 Bruno Lucena.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
*/

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/eventing"
)

const (
	agentFinalizer = "lambdaagent.lambda.knative.io/finalizer"
)

// LambdaAgentReconciler reconciles a LambdaAgent object
type LambdaAgentReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	EventingManager *eventing.Manager
}

// +kubebuilder:rbac:groups=lambda.knative.io,resources=lambdaagents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=lambda.knative.io,resources=lambdaagents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=lambda.knative.io,resources=lambdaagents/finalizers,verbs=update
// +kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers;triggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=messaging.knative.dev,resources=channels;subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

func (r *LambdaAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("lambdaagent", req.NamespacedName)

	// Fetch the LambdaAgent
	agent := &lambdav1alpha1.LambdaAgent{}
	if err := r.Get(ctx, req.NamespacedName, agent); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !agent.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, agent, log)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(agent, agentFinalizer) {
		controllerutil.AddFinalizer(agent, agentFinalizer)
		if err := r.Update(ctx, agent); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize status if needed
	if agent.Status.Phase == "" {
		agent.Status.Phase = lambdav1alpha1.AgentPhasePending
		if err := r.Status().Update(ctx, agent); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Reconcile based on phase
	switch agent.Status.Phase {
	case lambdav1alpha1.AgentPhasePending:
		return r.reconcilePending(ctx, agent, log)
	case lambdav1alpha1.AgentPhaseDeploying:
		return r.reconcileDeploying(ctx, agent, log)
	case lambdav1alpha1.AgentPhaseReady:
		return r.reconcileReady(ctx, agent, log)
	case lambdav1alpha1.AgentPhaseFailed:
		return r.reconcileFailed(ctx, agent, log)
	default:
		agent.Status.Phase = lambdav1alpha1.AgentPhasePending
		if err := r.Status().Update(ctx, agent); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
}

func (r *LambdaAgentReconciler) reconcilePending(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Pending phase")

	// Validate spec
	if agent.Spec.Image.Repository == "" {
		log.Error(nil, "Image repository is required")
		agent.Status.Phase = lambdav1alpha1.AgentPhaseFailed
		r.setCondition(agent, "Ready", metav1.ConditionFalse, "ValidationFailed", "Image repository is required")
		if err := r.Status().Update(ctx, agent); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Move to deploying
	agent.Status.Phase = lambdav1alpha1.AgentPhaseDeploying
	r.setCondition(agent, "Ready", metav1.ConditionFalse, "Deploying", "Creating Knative Service")
	if err := r.Status().Update(ctx, agent); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *LambdaAgentReconciler) reconcileDeploying(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Deploying phase")

	// Reconcile ghcr-secret if needed (for ghcr.io images)
	if err := r.reconcileImagePullSecret(ctx, agent, log); err != nil {
		log.Error(err, "Failed to reconcile image pull secret")
		// Don't fail - the service might still work if secret exists
	}

	// Build image URI
	imageURI := agent.Spec.Image.Repository
	if agent.Spec.Image.Digest != "" {
		imageURI = fmt.Sprintf("%s@%s", agent.Spec.Image.Repository, agent.Spec.Image.Digest)
	} else if agent.Spec.Image.Tag != "" {
		imageURI = fmt.Sprintf("%s:%s", agent.Spec.Image.Repository, agent.Spec.Image.Tag)
	} else {
		imageURI = fmt.Sprintf("%s:latest", agent.Spec.Image.Repository)
	}

	// Create or update Knative Service
	ksvc := r.buildKnativeService(agent, imageURI)

	// Set owner reference
	if err := controllerutil.SetControllerReference(agent, ksvc, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	// Check if service exists
	existingKsvc := &servingv1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: ksvc.Name, Namespace: ksvc.Namespace}, existingKsvc)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create service
			log.Info("Creating Knative Service", "name", ksvc.Name)
			if err := r.Create(ctx, ksvc); err != nil {
				log.Error(err, "Failed to create Knative Service")
				agent.Status.Phase = lambdav1alpha1.AgentPhaseFailed
				r.setCondition(agent, "Ready", metav1.ConditionFalse, "CreateFailed", err.Error())
				if err := r.Status().Update(ctx, agent); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		// Update service if needed
		existingKsvc.Spec = ksvc.Spec
		if err := r.Update(ctx, existingKsvc); err != nil {
			log.Error(err, "Failed to update Knative Service")
			return ctrl.Result{}, err
		}
	}

	// Reconcile eventing infrastructure (Broker, Triggers, Forwards)
	if r.EventingManager != nil && agent.Spec.Eventing != nil && agent.Spec.Eventing.Enabled {
		log.Info("Reconciling eventing infrastructure")
		if err := r.EventingManager.ReconcileAgentEventing(ctx, agent); err != nil {
			log.Error(err, "Failed to reconcile eventing")
			r.setCondition(agent, "Eventing", metav1.ConditionFalse, "EventingFailed", err.Error())
			// Don't fail the whole reconciliation - service might still work
		} else {
			r.setCondition(agent, "Eventing", metav1.ConditionTrue, "EventingReady", "Eventing infrastructure ready")
			// Update eventing status
			eventingStatus, err := r.EventingManager.GetAgentEventingStatus(ctx, agent)
			if err == nil {
				agent.Status.EventingStatus = eventingStatus
			}
		}
	}

	// Reconcile ServiceMonitor for Prometheus scraping
	if err := r.reconcileServiceMonitor(ctx, agent, log); err != nil {
		log.Error(err, "Failed to reconcile ServiceMonitor")
		r.setCondition(agent, "Monitoring", metav1.ConditionFalse, "MonitoringFailed", err.Error())
		// Don't fail the whole reconciliation - service might still work
	} else {
		r.setCondition(agent, "Monitoring", metav1.ConditionTrue, "MonitoringReady", "ServiceMonitor created")
	}

	// Update status
	agent.Status.ServiceStatus = &lambdav1alpha1.AgentServiceStatus{
		ServiceName: agent.Name,
		URL:         fmt.Sprintf("http://%s.%s.svc.cluster.local", agent.Name, agent.Namespace),
	}

	// Update AI status (ADR-004)
	r.updateAIStatus(agent)

	// Update permission status
	r.updatePermissionStatus(agent)

	// Check service status
	return r.checkServiceStatus(ctx, agent, log)
}

func (r *LambdaAgentReconciler) checkServiceStatus(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) (ctrl.Result, error) {
	ksvc := &servingv1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: agent.Name, Namespace: agent.Namespace}, ksvc); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if ready
	for _, cond := range ksvc.Status.Conditions {
		if cond.Type == "Ready" {
			if cond.Status == corev1.ConditionTrue {
				agent.Status.Phase = lambdav1alpha1.AgentPhaseReady
				agent.Status.ServiceStatus.Ready = true
				agent.Status.ServiceStatus.URL = ksvc.Status.URL.String()
				if ksvc.Status.LatestReadyRevisionName != "" {
					agent.Status.ServiceStatus.LatestRevision = ksvc.Status.LatestReadyRevisionName
				}
				r.setCondition(agent, "Ready", metav1.ConditionTrue, "Ready", "Agent is ready")
				if err := r.Status().Update(ctx, agent); err != nil {
					return ctrl.Result{}, err
				}
				log.Info("Agent is ready", "url", agent.Status.ServiceStatus.URL)
				return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
			} else if cond.Status == corev1.ConditionFalse && cond.Reason == "RevisionFailed" {
				agent.Status.Phase = lambdav1alpha1.AgentPhaseFailed
				r.setCondition(agent, "Ready", metav1.ConditionFalse, "RevisionFailed", cond.Message)
				if err := r.Status().Update(ctx, agent); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
			}
		}
	}

	// Still deploying
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *LambdaAgentReconciler) reconcileReady(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) (ctrl.Result, error) {
	// Update AI status on each reconcile (ADR-004)
	r.updateAIStatus(agent)

	// Update permission status on each reconcile
	r.updatePermissionStatus(agent)

	// Ensure ServiceMonitor exists for Prometheus scraping
	if err := r.reconcileServiceMonitor(ctx, agent, log); err != nil {
		log.Error(err, "Failed to reconcile ServiceMonitor")
		r.setCondition(agent, "Monitoring", metav1.ConditionFalse, "MonitoringFailed", err.Error())
		// Don't fail - monitoring is optional
	} else {
		r.setCondition(agent, "Monitoring", metav1.ConditionTrue, "MonitoringReady", "ServiceMonitor created")
	}

	// Verify service still exists and is ready
	return r.checkServiceStatus(ctx, agent, log)
}

// updatePermissionStatus updates the permission status in the agent status
func (r *LambdaAgentReconciler) updatePermissionStatus(agent *lambdav1alpha1.LambdaAgent) {
	if agent.Spec.Permissions == nil {
		// No permissions configured, clear status
		agent.Status.PermissionStatus = nil
		return
	}

	now := metav1.Now()
	permStatus := &lambdav1alpha1.AgentPermissionStatus{
		BrokerCreationDisabled:   agent.Spec.Permissions.DisableBrokerCreation,
		TriggerCreationDisabled:  agent.Spec.Permissions.DisableTriggerCreation,
		FunctionCreationDisabled: agent.Spec.Permissions.DisableFunctionCreation,
		LastEvaluated:            &now,
	}

	// Preserve any event-disabled capabilities from previous status
	if agent.Status.PermissionStatus != nil && len(agent.Status.PermissionStatus.DisabledByEvent) > 0 {
		// Filter out expired entries
		var activeDisables []lambdav1alpha1.DisabledCapability
		for _, dc := range agent.Status.PermissionStatus.DisabledByEvent {
			if dc.ExpiresAt == nil || dc.ExpiresAt.After(now.Time) {
				activeDisables = append(activeDisables, dc)
			}
		}
		permStatus.DisabledByEvent = activeDisables

		// Update status based on event disables
		for _, dc := range activeDisables {
			switch dc.Capability {
			case "broker":
				permStatus.BrokerCreationDisabled = true
			case "trigger":
				permStatus.TriggerCreationDisabled = true
			case "function":
				permStatus.FunctionCreationDisabled = true
			}
		}
	}

	agent.Status.PermissionStatus = permStatus
}

// updateAIStatus updates the AI status in the agent status (ADR-004 compliance)
func (r *LambdaAgentReconciler) updateAIStatus(agent *lambdav1alpha1.LambdaAgent) {
	if agent.Spec.AI == nil {
		// No AI configuration, clear status
		agent.Status.AIStatus = nil
		return
	}

	now := metav1.Now()
	aiStatus := &lambdav1alpha1.AgentAIStatus{
		Provider:        agent.Spec.AI.Provider,
		Endpoint:        agent.Spec.AI.Endpoint,
		ActiveModel:     agent.Spec.AI.Model,
		LastHealthCheck: &now,
	}

	// Determine model availability based on provider and configuration
	switch agent.Spec.AI.Provider {
	case "ollama":
		// Ollama is generally available if endpoint is set
		aiStatus.ModelAvailable = agent.Spec.AI.Endpoint != "" && agent.Spec.AI.Model != ""
	case "openai", "anthropic":
		// Cloud providers require API key
		aiStatus.ModelAvailable = agent.Spec.AI.APIKeySecretRef != nil && agent.Spec.AI.Model != ""
	case "none":
		aiStatus.ModelAvailable = true // No AI needed
	default:
		aiStatus.ModelAvailable = false
		aiStatus.Error = fmt.Sprintf("unknown AI provider: %s", agent.Spec.AI.Provider)
	}

	// Set error if model not available
	if !aiStatus.ModelAvailable && aiStatus.Error == "" {
		if agent.Spec.AI.Endpoint == "" {
			aiStatus.Error = "AI endpoint not configured"
		} else if agent.Spec.AI.Model == "" {
			aiStatus.Error = "AI model not configured"
		} else if agent.Spec.AI.APIKeySecretRef == nil && (agent.Spec.AI.Provider == "openai" || agent.Spec.AI.Provider == "anthropic") {
			aiStatus.Error = fmt.Sprintf("API key required for provider %s", agent.Spec.AI.Provider)
		}
	}

	agent.Status.AIStatus = aiStatus
}

func (r *LambdaAgentReconciler) reconcileFailed(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) (ctrl.Result, error) {
	log.Info("Agent in failed state, will retry")
	// Reset to pending for retry
	agent.Status.Phase = lambdav1alpha1.AgentPhasePending
	if err := r.Status().Update(ctx, agent); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

func (r *LambdaAgentReconciler) reconcileDelete(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling deletion")

	agent.Status.Phase = lambdav1alpha1.AgentPhaseDeleting
	if err := r.Status().Update(ctx, agent); err != nil {
		return ctrl.Result{}, err
	}

	// Delete eventing resources (Triggers, Channels)
	if r.EventingManager != nil {
		log.Info("Deleting eventing resources")
		if err := r.EventingManager.DeleteAgentEventing(ctx, agent); err != nil {
			log.Error(err, "Failed to delete eventing resources (continuing)")
		}
	}

	// Delete ServiceMonitor and metrics Service
	log.Info("Deleting monitoring resources")
	if err := r.deleteServiceMonitor(ctx, agent, log); err != nil {
		log.Error(err, "Failed to delete monitoring resources (continuing)")
	}

	// Delete Knative Service
	ksvc := &servingv1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: agent.Name, Namespace: agent.Namespace}, ksvc); err == nil {
		if err := r.Delete(ctx, ksvc); err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		log.Info("Deleted Knative Service")
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(agent, agentFinalizer)
	if err := r.Update(ctx, agent); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Successfully deleted LambdaAgent")
	return ctrl.Result{}, nil
}

func (r *LambdaAgentReconciler) buildKnativeService(agent *lambdav1alpha1.LambdaAgent, imageURI string) *servingv1.Service {
	// Build environment variables
	env := agent.Spec.Env

	// Add AI configuration as env vars if specified
	if agent.Spec.AI != nil {
		if agent.Spec.AI.Endpoint != "" {
			env = append(env, corev1.EnvVar{Name: "OLLAMA_URL", Value: agent.Spec.AI.Endpoint})
		}
		if agent.Spec.AI.Model != "" {
			env = append(env, corev1.EnvVar{Name: "OLLAMA_MODEL", Value: agent.Spec.AI.Model})
		}
	}

	// Add behavior configuration
	if agent.Spec.Behavior != nil {
		if agent.Spec.Behavior.EmitEvents {
			env = append(env, corev1.EnvVar{Name: "EMIT_EVENTS", Value: "true"})
		}
		if agent.Spec.Behavior.MaxContextMessages > 0 {
			env = append(env, corev1.EnvVar{Name: "MAX_CONTEXT_MESSAGES", Value: fmt.Sprintf("%d", agent.Spec.Behavior.MaxContextMessages)})
		}
	}

	// Add observability configuration
	if agent.Spec.Observability != nil && agent.Spec.Observability.Tracing != nil {
		if agent.Spec.Observability.Tracing.Enabled {
			env = append(env, corev1.EnvVar{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: agent.Spec.Observability.Tracing.Endpoint})
		}
	}

	// Build container port
	containerPort := int32(8080)
	if agent.Spec.Image.Port > 0 {
		containerPort = agent.Spec.Image.Port
	}

	// Build resource requirements
	resources := corev1.ResourceRequirements{}
	if agent.Spec.Resources != nil {
		if agent.Spec.Resources.Requests != nil {
			resources.Requests = corev1.ResourceList{}
			if agent.Spec.Resources.Requests.Memory != "" {
				resources.Requests[corev1.ResourceMemory] = resource.MustParse(agent.Spec.Resources.Requests.Memory)
			}
			if agent.Spec.Resources.Requests.CPU != "" {
				resources.Requests[corev1.ResourceCPU] = resource.MustParse(agent.Spec.Resources.Requests.CPU)
			}
		}
		if agent.Spec.Resources.Limits != nil {
			resources.Limits = corev1.ResourceList{}
			if agent.Spec.Resources.Limits.Memory != "" {
				resources.Limits[corev1.ResourceMemory] = resource.MustParse(agent.Spec.Resources.Limits.Memory)
			}
			if agent.Spec.Resources.Limits.CPU != "" {
				resources.Limits[corev1.ResourceCPU] = resource.MustParse(agent.Spec.Resources.Limits.CPU)
			}
		}
	}

	// Build container
	container := corev1.Container{
		Name:      "user-container",
		Image:     imageURI,
		Env:       env,
		Resources: resources,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: containerPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health",
					Port: intstr.FromInt32(containerPort),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       10,
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health",
					Port: intstr.FromInt32(containerPort),
				},
			},
			InitialDelaySeconds: 15,
			PeriodSeconds:       20,
		},
	}

	// Add command and args if specified
	if len(agent.Spec.Image.Command) > 0 {
		container.Command = agent.Spec.Image.Command
	}
	if len(agent.Spec.Image.Args) > 0 {
		container.Args = agent.Spec.Image.Args
	}

	// Build annotations for scaling
	annotations := map[string]string{
		"autoscaling.knative.dev/class": "kpa.autoscaling.knative.dev",
		// Fix: Disable Knative activator warm-pod preference to prevent traffic concentration
		"serving.knative.dev/rollout-duration": "0s",
	}
	if agent.Spec.Scaling != nil {
		annotations["autoscaling.knative.dev/min-scale"] = fmt.Sprintf("%d", agent.Spec.Scaling.MinReplicas)
		annotations["autoscaling.knative.dev/max-scale"] = fmt.Sprintf("%d", agent.Spec.Scaling.MaxReplicas)
		annotations["autoscaling.knative.dev/target"] = fmt.Sprintf("%d", agent.Spec.Scaling.TargetConcurrency)
		if agent.Spec.Scaling.ScaleToZeroGracePeriod != "" {
			annotations["autoscaling.knative.dev/scale-to-zero-pod-retention-period"] = agent.Spec.Scaling.ScaleToZeroGracePeriod
		}
	}

	// Build PodSpec
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
	}

	// Set ServiceAccountName if specified
	if agent.Spec.ServiceAccountName != "" {
		podSpec.ServiceAccountName = agent.Spec.ServiceAccountName
	}

	// Add imagePullSecrets for ghcr.io images or if explicitly specified
	if len(agent.Spec.Image.ImagePullSecrets) > 0 {
		// Use explicitly specified secrets (convert []string to []LocalObjectReference)
		for _, secretName := range agent.Spec.Image.ImagePullSecrets {
			podSpec.ImagePullSecrets = append(podSpec.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName})
		}
	} else if strings.HasPrefix(agent.Spec.Image.Repository, "ghcr.io/") {
		// Auto-add ghcr-secret for ghcr.io images
		podSpec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: "ghcr-secret"},
		}
	}

	// NOTE: TopologySpreadConstraints and Affinity are not allowed by Knative Serving's admission webhook
	// These fields cannot be set directly on PodSpec within a Knative Service.
	// If pod distribution is needed, it should be handled at the cluster/node level or via other mechanisms.
	// Removed to fix: "validation failed: must not set the field(s): spec.template.spec.affinity, spec.template.spec.topologySpreadConstraints"
	//
	// podSpec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{...}
	// podSpec.Affinity = &corev1.Affinity{...}

	// Build Knative Service
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       agent.Name,
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"app.kubernetes.io/component":  "agent",
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: podSpec,
					},
				},
			},
		},
	}

	return ksvc
}

func (r *LambdaAgentReconciler) setCondition(agent *lambdav1alpha1.LambdaAgent, condType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
	agent.Status.SetCondition(condition)
}

// reconcileServiceMonitor creates or updates the metrics Service and ServiceMonitor for Prometheus scraping
func (r *LambdaAgentReconciler) reconcileServiceMonitor(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) error {
	// Check if ServiceMonitor is enabled (default: true when metrics are enabled)
	// ServiceMonitor creation is ON by default if metrics are enabled
	metricsEnabled := true
	serviceMonitorEnabled := true // Default to true
	metricsPath := "/metrics"
	interval := "30s"
	timeout := "10s"

	if agent.Spec.Observability != nil && agent.Spec.Observability.Metrics != nil {
		metricsEnabled = agent.Spec.Observability.Metrics.Enabled
		// Only check ServiceMonitor field if metrics section exists and field was explicitly set
		// We use Path field presence as indicator that config was intentionally set
		if agent.Spec.Observability.Metrics.Path != "" {
			metricsPath = agent.Spec.Observability.Metrics.Path
		}
		if agent.Spec.Observability.Metrics.Interval != "" {
			interval = agent.Spec.Observability.Metrics.Interval
		}
		if agent.Spec.Observability.Metrics.Timeout != "" {
			timeout = agent.Spec.Observability.Metrics.Timeout
		}
		// ServiceMonitor field: only disable if explicitly set to false AND the field was provided
		// Since Go bool defaults to false, we need to check if metrics section has any custom config
		// For now, default to creating ServiceMonitor for all agents with metrics enabled
	}

	if !metricsEnabled {
		log.Info("Metrics disabled for agent, skipping ServiceMonitor")
		return nil
	}

	if !serviceMonitorEnabled {
		log.Info("ServiceMonitor explicitly disabled for agent")
		return nil
	}

	// Get container port
	containerPort := int32(8080)
	if agent.Spec.Image.Port > 0 {
		containerPort = agent.Spec.Image.Port
	}

	// Create metrics Service
	metricsServiceName := fmt.Sprintf("%s-metrics", agent.Name)
	metricsSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsServiceName,
			Namespace: agent.Namespace,
			Labels: map[string]string{
				"app":                          agent.Name,
				"app.kubernetes.io/name":       agent.Name,
				"app.kubernetes.io/component":  "metrics",
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"mirror.linkerd.io/exported":   "true", // Linkerd multicluster support
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				// Select all pods from this Knative service (any revision)
				"serving.knative.dev/service": agent.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       containerPort,
					TargetPort: intstr.FromInt32(containerPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(agent, metricsSvc, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on metrics service: %w", err)
	}

	// Create or update metrics Service
	existingSvc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: metricsServiceName, Namespace: agent.Namespace}, existingSvc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating metrics Service", "name", metricsServiceName)
			if err := r.Create(ctx, metricsSvc); err != nil {
				return fmt.Errorf("failed to create metrics service: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get metrics service: %w", err)
		}
	} else {
		// Update existing service
		existingSvc.Spec.Selector = metricsSvc.Spec.Selector
		existingSvc.Spec.Ports = metricsSvc.Spec.Ports
		existingSvc.Labels = metricsSvc.Labels
		if err := r.Update(ctx, existingSvc); err != nil {
			return fmt.Errorf("failed to update metrics service: %w", err)
		}
	}

	// Create ServiceMonitor
	intervalDuration := monitoringv1.Duration(interval)
	timeoutDuration := monitoringv1.Duration(timeout)

	serviceMonitor := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels: map[string]string{
				"app":                          agent.Name,
				"app.kubernetes.io/name":       agent.Name,
				"app.kubernetes.io/managed-by": "knative-lambda-operator",
				"release":                      "kube-prometheus-stack", // Required for Prometheus Operator discovery
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                         agent.Name,
					"app.kubernetes.io/component": "metrics",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{agent.Namespace},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:          "metrics",
					Path:          metricsPath,
					Interval:      intervalDuration,
					ScrapeTimeout: timeoutDuration,
					HonorLabels:   true,
				},
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(agent, serviceMonitor, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on servicemonitor: %w", err)
	}

	// Create or update ServiceMonitor
	existingSM := &monitoringv1.ServiceMonitor{}
	err = r.Get(ctx, types.NamespacedName{Name: agent.Name, Namespace: agent.Namespace}, existingSM)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ServiceMonitor", "name", agent.Name)
			if err := r.Create(ctx, serviceMonitor); err != nil {
				return fmt.Errorf("failed to create servicemonitor: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get servicemonitor: %w", err)
		}
	} else {
		// Update existing ServiceMonitor
		existingSM.Spec = serviceMonitor.Spec
		existingSM.Labels = serviceMonitor.Labels
		if err := r.Update(ctx, existingSM); err != nil {
			return fmt.Errorf("failed to update servicemonitor: %w", err)
		}
	}

	log.Info("ServiceMonitor reconciled successfully", "name", agent.Name, "metricsService", metricsServiceName)
	return nil
}

// deleteServiceMonitor deletes the metrics Service and ServiceMonitor
func (r *LambdaAgentReconciler) deleteServiceMonitor(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) error {
	// Delete ServiceMonitor
	sm := &monitoringv1.ServiceMonitor{}
	if err := r.Get(ctx, types.NamespacedName{Name: agent.Name, Namespace: agent.Namespace}, sm); err == nil {
		if err := r.Delete(ctx, sm); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete servicemonitor: %w", err)
		}
		log.Info("Deleted ServiceMonitor", "name", agent.Name)
	}

	// Delete metrics Service
	metricsServiceName := fmt.Sprintf("%s-metrics", agent.Name)
	svc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: metricsServiceName, Namespace: agent.Namespace}, svc); err == nil {
		if err := r.Delete(ctx, svc); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete metrics service: %w", err)
		}
		log.Info("Deleted metrics Service", "name", metricsServiceName)
	}

	return nil
}

// reconcileImagePullSecret creates ghcr-secret in the agent's namespace if the image is from ghcr.io
// It copies the github-token from flux-system namespace
func (r *LambdaAgentReconciler) reconcileImagePullSecret(ctx context.Context, agent *lambdav1alpha1.LambdaAgent, log logr.Logger) error {
	// Check if image is from ghcr.io
	if agent.Spec.Image.Repository == "" {
		return nil
	}

	// Only create secret for ghcr.io images
	isGHCR := len(agent.Spec.Image.Repository) >= 8 && agent.Spec.Image.Repository[:8] == "ghcr.io/"
	if !isGHCR {
		log.Info("Image not from ghcr.io, skipping secret creation", "repository", agent.Spec.Image.Repository)
		return nil
	}

	// Check if ghcr-secret already exists in agent namespace
	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: "ghcr-secret", Namespace: agent.Namespace}, existingSecret)
	if err == nil {
		// Secret already exists
		log.Info("ghcr-secret already exists", "namespace", agent.Namespace)
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check for existing ghcr-secret: %w", err)
	}

	// Get github-token from flux-system namespace
	sourceSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: "github-token", Namespace: "flux-system"}, sourceSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("github-token secret not found in flux-system, skipping ghcr-secret creation")
			return nil
		}
		return fmt.Errorf("failed to get github-token secret: %w", err)
	}

	// Extract token from github-token secret
	var token string
	for _, key := range []string{"password", "token", "value"} {
		if val, ok := sourceSecret.Data[key]; ok && len(val) > 0 {
			token = string(val)
			break
		}
	}
	if token == "" {
		log.Info("No token found in github-token secret, skipping ghcr-secret creation")
		return nil
	}

	// Get username or use default
	username := "git"
	if val, ok := sourceSecret.Data["username"]; ok && len(val) > 0 {
		username = string(val)
	}

	// Create docker config JSON
	dockerConfig := fmt.Sprintf(`{"auths":{"ghcr.io":{"username":"%s","password":"%s","auth":"%s"}}}`,
		username, token, base64Encode(fmt.Sprintf("%s:%s", username, token)))

	// Create ghcr-secret
	ghcrSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ghcr-secret",
			Namespace: agent.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":  "knative-lambda-operator",
				"app.kubernetes.io/created-for": agent.Name,
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte(dockerConfig),
		},
	}

	log.Info("Creating ghcr-secret", "namespace", agent.Namespace)
	if err := r.Create(ctx, ghcrSecret); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("failed to create ghcr-secret: %w", err)
	}

	log.Info("Successfully created ghcr-secret", "namespace", agent.Namespace)
	return nil
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// SetupWithManager sets up the controller with the Manager
func (r *LambdaAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lambdav1alpha1.LambdaAgent{}).
		Owns(&servingv1.Service{}).
		Complete(r)
}
