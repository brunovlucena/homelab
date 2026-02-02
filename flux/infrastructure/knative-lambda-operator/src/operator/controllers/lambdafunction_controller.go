package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/build"
	"github.com/brunovlucena/knative-lambda-operator/internal/deploy"
	"github.com/brunovlucena/knative-lambda-operator/internal/eventing"
	"github.com/brunovlucena/knative-lambda-operator/internal/events"
	"github.com/brunovlucena/knative-lambda-operator/internal/metrics"
	"github.com/brunovlucena/knative-lambda-operator/internal/observability"
	"github.com/brunovlucena/knative-lambda-operator/internal/validation"
)

const (
	// FinalizerName is the finalizer for LambdaFunction resources
	FinalizerName = "lambdafunction.lambda.knative.io/finalizer"

	// Requeue intervals - tuned for scale
	// Short: for active operations (building)
	// Medium: for monitoring ready state
	// Long: for failed/stable states to reduce API server load
	RequeueShort  = 15 * time.Second
	RequeueMedium = 1 * time.Minute
	RequeueLong   = 5 * time.Minute
)

// ReconcilerOptions configures the reconciler for scale
type ReconcilerOptions struct {
	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles
	// For 10K lambdas, recommended: 50-100
	MaxConcurrentReconciles int

	// RateLimiter limits the rate of reconciliations
	RateLimiter workqueue.TypedRateLimiter[ctrl.Request]
}

// LambdaFunctionReconciler reconciles a LambdaFunction object
type LambdaFunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	BuildManager    *build.Manager
	DeployManager   *deploy.Manager
	EventManager    *events.Manager
	EventingManager *eventing.Manager

	// Metrics for observability at scale
	Metrics *metrics.ReconcilerMetrics

	// OTELProvider for distributed tracing
	OTELProvider *observability.Provider
}

//+kubebuilder:rbac:groups=lambda.knative.io,resources=lambdafunctions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=lambda.knative.io,resources=lambdafunctions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=lambda.knative.io,resources=lambdafunctions/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers;triggers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=rabbitmqbrokerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sources.knative.dev,resources=apiserversources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rabbitmq.com,resources=exchanges;queues;bindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *LambdaFunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	startTime := time.Now()
	log := r.Log.WithValues("lambdafunction", req.NamespacedName)

	// Start tracing span for the reconcile operation
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcileSpan(ctx, req.Name, req.Namespace)
		defer func() {
			span.SetAttributes(
				attribute.Float64("reconcile.duration_ms", float64(time.Since(startTime).Milliseconds())),
			)
			span.End()
		}()
	}

	// Defer metrics recording
	defer func() {
		duration := time.Since(startTime).Seconds()
		if r.Metrics != nil {
			r.Metrics.RecordReconcile("total", "completed", duration)
		}
	}()

	// Fetch the LambdaFunction instance
	lambda := &lambdav1alpha1.LambdaFunction{}
	if err := r.Get(ctx, req.NamespacedName, lambda); err != nil {
		if apierrors.IsNotFound(err) {
			// Object deleted - no requeue needed
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LambdaFunction")
		r.recordError("reconcile", "get_failed")
		observability.RecordError(observability.SpanFromContext(ctx), err, "Failed to get LambdaFunction")
		return ctrl.Result{}, err
	}

	// Add trace attributes for the lambda
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetAttributes(
			attribute.String("lambda.phase", string(lambda.Status.Phase)),
			attribute.Int64("lambda.generation", lambda.Generation),
			attribute.String("lambda.runtime", lambda.Spec.Runtime.Language),
		)
	}

	// Handle deletion
	if !lambda.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, lambda, log)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(lambda, FinalizerName) {
		controllerutil.AddFinalizer(lambda, FinalizerName)
		if err := r.Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Track observed generation for change detection
	if lambda.Status.ObservedGeneration != lambda.Generation {
		lambda.Status.ObservedGeneration = lambda.Generation
	}

	// State machine reconciliation
	var result ctrl.Result
	var err error

	switch lambda.Status.Phase {
	case "", lambdav1alpha1.PhasePending:
		result, err = r.reconcilePending(ctx, lambda, log)
	case lambdav1alpha1.PhaseBuilding:
		result, err = r.reconcileBuilding(ctx, lambda, log)
	case lambdav1alpha1.PhaseDeploying:
		result, err = r.reconcileDeploying(ctx, lambda, log)
	case lambdav1alpha1.PhaseReady:
		result, err = r.reconcileReady(ctx, lambda, log)
	case lambdav1alpha1.PhaseFailed:
		result, err = r.reconcileFailed(ctx, lambda, log)
	default:
		log.Info("Unknown phase, resetting to Pending", "phase", lambda.Status.Phase)
		result, err = r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
	}

	// Record error in span if any
	if err != nil {
		observability.RecordError(observability.SpanFromContext(ctx), err, "Reconcile error")
	} else {
		observability.SetSpanOK(observability.SpanFromContext(ctx))
	}

	// Record phase-specific metrics
	if r.Metrics != nil {
		r.Metrics.RecordReconcile(string(lambda.Status.Phase), resultString(err), time.Since(startTime).Seconds())
	}

	// Update lambda count metrics for this namespace
	r.updateLambdaCounts(ctx, lambda.Namespace)

	return result, err
}

// reconcilePending handles the Pending phase - validates and starts build
func (r *LambdaFunctionReconciler) reconcilePending(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.V(1).Info("Reconciling Pending phase")

	// Start phase span
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcilePhaseSpan(ctx, lambda.Name, lambda.Namespace, "Pending")
		defer span.End()
	}

	// Validate spec
	if err := r.validateSpec(lambda); err != nil {
		log.Error(err, "Spec validation failed")
		observability.RecordError(observability.SpanFromContext(ctx), err, "Spec validation failed")
		r.setCondition(lambda, lambdav1alpha1.ConditionSourceReady, metav1.ConditionFalse, "ValidationFailed", err.Error())
		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: RequeueLong}, nil
	}

	// Reconcile eventing infrastructure EARLY - before build starts
	// This ensures Broker/Trigger exist so lambda can receive events as soon as deployed
	eventingEnabled := lambda.Spec.Eventing == nil || lambda.Spec.Eventing.Enabled
	if r.EventingManager != nil && eventingEnabled {
		log.V(1).Info("Reconciling eventing infrastructure (early)")
		if err := r.EventingManager.ReconcileEventing(ctx, lambda); err != nil {
			log.Error(err, "Failed to reconcile eventing infrastructure")
			r.setCondition(lambda, lambdav1alpha1.ConditionEventingReady, metav1.ConditionFalse, "EventingFailed", err.Error())
			r.recordError("eventing", "reconcile_failed")
			if err := r.Status().Update(ctx, lambda); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: RequeueShort}, nil
		}
		r.setCondition(lambda, lambdav1alpha1.ConditionEventingReady, metav1.ConditionTrue, "EventingReady", "Broker and Trigger created")
	}

	// Check if this LambdaFunction should use operator image directly (receiver mode)
	if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/use-operator-image"] == "true" {
		log.Info("Using operator image directly (receiver mode), skipping build")

		// KISS: Construct image from OPERATOR_REGISTRY + OPERATOR_VERSION
		registry := os.Getenv("OPERATOR_REGISTRY")
		if registry == "" {
			registry = "localhost:5001"
		}
		version := os.Getenv("OPERATOR_VERSION")
		if version == "" {
			return ctrl.Result{}, fmt.Errorf("OPERATOR_VERSION not set")
		}
		operatorImage := fmt.Sprintf("%s/knative-lambda-operator:%s", registry, version)

		now := metav1.Now()
		lambda.Status.Phase = lambdav1alpha1.PhaseDeploying
		lambda.Status.BuildStatus = &lambdav1alpha1.BuildStatusInfo{
			ImageURI:    operatorImage,
			StartedAt:   &now,
			CompletedAt: &now,
			Attempt:     1,
		}
		r.setCondition(lambda, lambdav1alpha1.ConditionBuildReady, metav1.ConditionTrue, "BuildSkipped", "Using operator image directly")
		r.setCondition(lambda, lambdav1alpha1.ConditionSourceReady, metav1.ConditionTrue, "SourceReady", "Using operator image directly")

		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Skipped build, using operator image", "image", operatorImage)
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if using pre-built image (source.type: image)
	// This skips the entire build pipeline for FastAPI apps and pre-built containers
	if lambda.Spec.Source.Type == "image" && lambda.Spec.Source.Image != nil {
		log.Info("Using pre-built image (source.type: image), skipping build")

		imageSource := lambda.Spec.Source.Image
		imageURI := imageSource.Repository
		if imageSource.Digest != "" {
			imageURI = fmt.Sprintf("%s@%s", imageSource.Repository, imageSource.Digest)
		} else if imageSource.Tag != "" {
			imageURI = fmt.Sprintf("%s:%s", imageSource.Repository, imageSource.Tag)
		} else {
			imageURI = fmt.Sprintf("%s:latest", imageSource.Repository)
		}

		now := metav1.Now()
		lambda.Status.Phase = lambdav1alpha1.PhaseDeploying
		lambda.Status.BuildStatus = &lambdav1alpha1.BuildStatusInfo{
			ImageURI:    imageURI,
			StartedAt:   &now,
			CompletedAt: &now,
			Attempt:     1,
		}
		r.setCondition(lambda, lambdav1alpha1.ConditionBuildReady, metav1.ConditionTrue, "BuildSkipped", "Using pre-built image")
		r.setCondition(lambda, lambdav1alpha1.ConditionSourceReady, metav1.ConditionTrue, "SourceReady", "Pre-built image configured")

		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Skipped build, using pre-built image", "image", imageURI)
		return ctrl.Result{Requeue: true}, nil
	}

	// Create build context with tracing
	var buildCtxSpan trace.Span
	if r.OTELProvider != nil {
		ctx, buildCtxSpan = r.OTELProvider.StartBuildContextSpan(ctx, lambda.Name, lambda.Namespace, lambda.Spec.Runtime.Language)
	}
	buildCtx, err := r.BuildManager.CreateBuildContext(ctx, lambda)
	if err != nil {
		log.Error(err, "Failed to create build context")
		observability.RecordError(buildCtxSpan, err, "Failed to create build context")
		if buildCtxSpan != nil {
			buildCtxSpan.End()
		}
		r.setCondition(lambda, lambdav1alpha1.ConditionSourceReady, metav1.ConditionFalse, "BuildContextFailed", err.Error())
		r.recordError("build", "context_failed")
		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: RequeueShort}, nil
	}
	if buildCtxSpan != nil {
		buildCtxSpan.SetAttributes(attribute.String("build.context_configmap", buildCtx.ConfigMapName))
		buildCtxSpan.End()
	}

	r.setCondition(lambda, lambdav1alpha1.ConditionSourceReady, metav1.ConditionTrue, "BuildContextCreated", "Build context created")

	// Create Kaniko job with tracing
	var buildJobSpan trace.Span
	if r.OTELProvider != nil {
		ctx, buildJobSpan = r.OTELProvider.StartBuildJobSpan(ctx, lambda.Name, lambda.Namespace, "")
	}
	job, err := r.BuildManager.CreateKanikoJob(ctx, lambda, buildCtx)
	if err != nil {
		log.Error(err, "Failed to create Kaniko job")
		observability.RecordError(buildJobSpan, err, "Failed to create Kaniko job")
		if buildJobSpan != nil {
			buildJobSpan.End()
		}
		r.setCondition(lambda, lambdav1alpha1.ConditionBuildReady, metav1.ConditionFalse, "JobCreationFailed", err.Error())
		r.recordError("build", "job_creation_failed")
		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: RequeueShort}, nil
	}
	if buildJobSpan != nil {
		buildJobSpan.SetAttributes(attribute.String("build.job_name", job.Name))
		observability.SetSpanOK(buildJobSpan)
		buildJobSpan.End()
	}

	// Emit event (non-blocking)
	if r.EventManager != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = r.EventManager.EmitBuildStarted(ctx, lambda, job.Name)
		}()
	}

	// Update status
	now := metav1.Now()
	lambda.Status.Phase = lambdav1alpha1.PhaseBuilding
	lambda.Status.BuildStatus = &lambdav1alpha1.BuildStatusInfo{
		JobName:   job.Name,
		StartedAt: &now,
		Attempt:   1,
	}

	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	// Update active build jobs metric after creating a new build job
	r.updateActiveBuildJobCount(ctx, lambda.Namespace)

	log.Info("Build job created", "job", job.Name)
	return ctrl.Result{RequeueAfter: RequeueShort}, nil
}

// reconcileBuilding handles the Building phase - monitors build job
func (r *LambdaFunctionReconciler) reconcileBuilding(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.V(1).Info("Reconciling Building phase")

	// Start phase span
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcilePhaseSpan(ctx, lambda.Name, lambda.Namespace, "Building")
		defer span.End()
	}

	if lambda.Status.BuildStatus == nil || lambda.Status.BuildStatus.JobName == "" {
		log.Info("No build job name, resetting to Pending")
		return r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
	}

	// Check build job status with tracing
	var buildStatusSpan trace.Span
	if r.OTELProvider != nil {
		ctx, buildStatusSpan = r.OTELProvider.StartBuildStatusSpan(ctx, lambda.Namespace, lambda.Status.BuildStatus.JobName)
	}
	status, err := r.BuildManager.GetBuildStatus(ctx, lambda.Namespace, lambda.Status.BuildStatus.JobName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Build job not found, resetting to Pending")
			if buildStatusSpan != nil {
				buildStatusSpan.End()
			}
			return r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
		}
		observability.RecordError(buildStatusSpan, err, "Failed to get build status")
		if buildStatusSpan != nil {
			buildStatusSpan.End()
		}
		return ctrl.Result{}, err
	}
	if buildStatusSpan != nil {
		buildStatusSpan.SetAttributes(
			attribute.Bool("build.completed", status.Completed),
			attribute.Bool("build.success", status.Success),
		)
		if status.ImageURI != "" {
			buildStatusSpan.SetAttributes(attribute.String("build.image_uri", status.ImageURI))
		}
		buildStatusSpan.End()
	}

	if !status.Completed {
		// Still building - requeue
		return ctrl.Result{RequeueAfter: RequeueShort}, nil
	}

	// Build completed
	now := metav1.Now()
	lambda.Status.BuildStatus.CompletedAt = &now

	// Calculate build duration for metrics
	var buildDuration float64
	if lambda.Status.BuildStatus.StartedAt != nil {
		buildDuration = now.Sub(lambda.Status.BuildStatus.StartedAt.Time).Seconds()
	}

	if status.Success {
		log.Info("Build succeeded", "imageURI", status.ImageURI, "duration", buildDuration)
		lambda.Status.Phase = lambdav1alpha1.PhaseDeploying
		lambda.Status.BuildStatus.ImageURI = status.ImageURI
		r.setCondition(lambda, lambdav1alpha1.ConditionBuildReady, metav1.ConditionTrue, "BuildSucceeded", "Image built successfully")

		if r.Metrics != nil {
			r.Metrics.RecordBuild(lambda.Spec.Runtime.Language, "success", buildDuration)
		}

		// Emit event (non-blocking)
		if r.EventManager != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = r.EventManager.EmitBuildCompleted(ctx, lambda, status.ImageURI)
			}()
		}
	} else {
		log.Info("Build failed", "error", status.Error, "duration", buildDuration)
		lambda.Status.Phase = lambdav1alpha1.PhaseFailed
		lambda.Status.BuildStatus.Error = status.Error
		r.setCondition(lambda, lambdav1alpha1.ConditionBuildReady, metav1.ConditionFalse, "BuildFailed", status.Error)

		if r.Metrics != nil {
			r.Metrics.RecordBuild(lambda.Spec.Runtime.Language, "failed", buildDuration)
		}
		r.recordError("build", "build_failed")

		// Emit event (non-blocking)
		if r.EventManager != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = r.EventManager.EmitBuildFailed(ctx, lambda, status.Error)
			}()
		}
	}

	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	// Update active build jobs metric after build completes/fails
	r.updateActiveBuildJobCount(ctx, lambda.Namespace)

	if status.Success {
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

// reconcileDeploying handles the Deploying phase - creates eventing and service
func (r *LambdaFunctionReconciler) reconcileDeploying(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.V(1).Info("Reconciling Deploying phase")

	// Start phase span
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcilePhaseSpan(ctx, lambda.Name, lambda.Namespace, "Deploying")
		defer span.End()
	}

	if lambda.Status.BuildStatus == nil || lambda.Status.BuildStatus.ImageURI == "" {
		log.Info("No image URI, resetting to Pending")
		return r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
	}

	// Ensure eventing infrastructure exists (safety check - should already exist from Pending phase)
	eventingEnabled := lambda.Spec.Eventing == nil || lambda.Spec.Eventing.Enabled
	if r.EventingManager != nil && eventingEnabled {
		log.V(1).Info("Ensuring eventing infrastructure exists")
		// Add eventing reconcile span
		var eventingSpan trace.Span
		if r.OTELProvider != nil {
			ctx, eventingSpan = r.OTELProvider.StartEventingReconcileSpan(ctx, lambda.Name, lambda.Namespace)
		}
		if err := r.EventingManager.ReconcileEventing(ctx, lambda); err != nil {
			log.Error(err, "Failed to reconcile eventing infrastructure")
			observability.RecordError(eventingSpan, err, "Failed to reconcile eventing")
			if eventingSpan != nil {
				eventingSpan.End()
			}
			r.setCondition(lambda, lambdav1alpha1.ConditionEventingReady, metav1.ConditionFalse, "EventingFailed", err.Error())
			r.recordError("eventing", "reconcile_failed")
			if err := r.Status().Update(ctx, lambda); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: RequeueShort}, nil
		}
		if eventingSpan != nil {
			observability.SetSpanOK(eventingSpan)
			eventingSpan.End()
		}
	}

	// Get or create Knative service with tracing
	var deploySpan trace.Span
	if r.OTELProvider != nil {
		ctx, deploySpan = r.OTELProvider.StartDeployServiceSpan(ctx, lambda.Name, lambda.Namespace)
	}
	service, err := r.DeployManager.GetService(ctx, lambda)
	if err != nil && !apierrors.IsNotFound(err) {
		observability.RecordError(deploySpan, err, "Failed to get service")
		if deploySpan != nil {
			deploySpan.End()
		}
		return ctrl.Result{}, err
	}

	if service == nil {
		log.Info("Creating Knative service")
		observability.AddSpanEvent(deploySpan, "creating_service")
		service, err = r.DeployManager.CreateService(ctx, lambda)
		if err != nil {
			log.Error(err, "Failed to create Knative service")
			observability.RecordError(deploySpan, err, "Failed to create service")
			if deploySpan != nil {
				deploySpan.End()
			}
			r.setCondition(lambda, lambdav1alpha1.ConditionDeployReady, metav1.ConditionFalse, "ServiceCreationFailed", err.Error())
			r.recordError("deploy", "service_creation_failed")
			if err := r.Status().Update(ctx, lambda); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: RequeueShort}, nil
		}
		observability.AddSpanEvent(deploySpan, "service_created", attribute.String("service.name", service.GetName()))

		// Emit event (non-blocking)
		if r.EventManager != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = r.EventManager.EmitServiceCreated(ctx, lambda, service.GetName())
			}()
		}
	} else {
		// Service exists - ensure version matches operator for use-operator-image lambdas
		if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/use-operator-image"] == "true" {
			registry := os.Getenv("OPERATOR_REGISTRY")
			if registry == "" {
				registry = "localhost:5001"
			}
			version := os.Getenv("OPERATOR_VERSION")
			if version == "" {
				log.Error(nil, "OPERATOR_VERSION not set")
				return ctrl.Result{RequeueAfter: RequeueShort}, nil
			}
			operatorImage := fmt.Sprintf("%s/knative-lambda-operator:%s", registry, version)

			currentImage := r.DeployManager.GetServiceImage(service)
			if currentImage != operatorImage {
				log.Info("VERSION MISMATCH: Syncing lambda-command-receiver with operator",
					"current", currentImage,
					"expected", operatorImage)

				lambda.Status.BuildStatus.ImageURI = operatorImage
				if err := r.Status().Update(ctx, lambda); err != nil {
					return ctrl.Result{}, err
				}

				if err := r.DeployManager.UpdateServiceImage(ctx, service, operatorImage); err != nil {
					log.Error(err, "Failed to update service image")
					return ctrl.Result{RequeueAfter: RequeueShort}, nil
				}
				log.Info("Service synced", "image", operatorImage)
			}
		}
	}

	// Check service status with tracing
	var statusSpan trace.Span
	if r.OTELProvider != nil {
		ctx, statusSpan = r.OTELProvider.StartDeployStatusSpan(ctx, lambda.Name, lambda.Namespace)
	}
	ready, url, replicas, err := r.DeployManager.GetServiceStatus(ctx, service)
	if err != nil {
		observability.RecordError(statusSpan, err, "Failed to get service status")
		if statusSpan != nil {
			statusSpan.End()
		}
		if deploySpan != nil {
			deploySpan.End()
		}
		return ctrl.Result{}, err
	}
	if statusSpan != nil {
		statusSpan.SetAttributes(
			attribute.Bool("service.ready", ready),
			attribute.Int("service.replicas", int(replicas)),
		)
		if url != "" {
			statusSpan.SetAttributes(attribute.String("service.url", url))
		}
		statusSpan.End()
	}

	// End deploy span
	if deploySpan != nil {
		deploySpan.SetAttributes(attribute.Bool("service.ready", ready))
		observability.SetSpanOK(deploySpan)
		deploySpan.End()
	}

	// Update service status
	lambda.Status.ServiceStatus = &lambdav1alpha1.ServiceStatusInfo{
		ServiceName: service.GetName(),
		URL:         url,
		Ready:       ready,
		Replicas:    replicas,
	}

	if ready {
		log.Info("Service is ready", "url", url)
		lambda.Status.Phase = lambdav1alpha1.PhaseReady
		r.setCondition(lambda, lambdav1alpha1.ConditionDeployReady, metav1.ConditionTrue, "ServiceDeployed", "Service deployed successfully")
		r.setCondition(lambda, lambdav1alpha1.ConditionServiceReady, metav1.ConditionTrue, "ServiceReady", "Service is ready to receive CloudEvents")

		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: RequeueMedium}, nil
	}

	// Still deploying
	r.setCondition(lambda, lambdav1alpha1.ConditionDeployReady, metav1.ConditionFalse, "ServiceDeploying", "Service is being deployed")
	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: RequeueShort}, nil
}

// reconcileReady handles the Ready phase - monitors health
func (r *LambdaFunctionReconciler) reconcileReady(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.V(2).Info("Reconciling Ready phase")

	// Start phase span
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcilePhaseSpan(ctx, lambda.Name, lambda.Namespace, "Ready")
		defer span.End()
	}

	// Check if spec was updated (requires rebuild)
	if lambda.Generation != lambda.Status.ObservedGeneration {
		log.Info("Spec updated, triggering rebuild")
		observability.AddSpanEvent(observability.SpanFromContext(ctx), "spec_updated_rebuild_triggered")
		return r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
	}

	// Monitor service health
	service, err := r.DeployManager.GetService(ctx, lambda)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Service deleted, resetting to Pending")
			return r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
		}
		return ctrl.Result{}, err
	}

	ready, url, replicas, err := r.DeployManager.GetServiceStatus(ctx, service)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status only if changed (reduces API server load)
	statusChanged := false
	if lambda.Status.ServiceStatus.Ready != ready {
		lambda.Status.ServiceStatus.Ready = ready
		statusChanged = true
	}
	if lambda.Status.ServiceStatus.URL != url {
		lambda.Status.ServiceStatus.URL = url
		statusChanged = true
	}
	if lambda.Status.ServiceStatus.Replicas != replicas {
		lambda.Status.ServiceStatus.Replicas = replicas
		statusChanged = true
	}

	if !ready {
		log.Info("Service no longer ready")
		lambda.Status.Phase = lambdav1alpha1.PhaseFailed
		r.setCondition(lambda, lambdav1alpha1.ConditionServiceReady, metav1.ConditionFalse, "ServiceUnhealthy", "Service is not ready")
		statusChanged = true
	}

	if statusChanged {
		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: RequeueMedium}, nil
}

// reconcileFailed handles the Failed phase - allows retry on spec change
func (r *LambdaFunctionReconciler) reconcileFailed(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.V(1).Info("Reconciling Failed phase")

	// Start phase span
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcilePhaseSpan(ctx, lambda.Name, lambda.Namespace, "Failed")
		defer span.End()
	}

	// Check if spec was updated (user might have fixed the issue)
	if lambda.Generation != lambda.Status.ObservedGeneration {
		log.Info("Spec updated after failure, retrying")
		observability.AddSpanEvent(observability.SpanFromContext(ctx), "spec_updated_retry_triggered")
		return r.setPhase(ctx, lambda, lambdav1alpha1.PhasePending, log)
	}

	// Don't requeue too frequently for failed state
	return ctrl.Result{RequeueAfter: RequeueLong}, nil
}

// reconcileDelete handles deletion with cleanup
func (r *LambdaFunctionReconciler) reconcileDelete(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling deletion")

	// Start phase span
	if r.OTELProvider != nil {
		var span trace.Span
		ctx, span = r.OTELProvider.StartReconcilePhaseSpan(ctx, lambda.Name, lambda.Namespace, "Deleting")
		defer span.End()
	}

	// Update phase
	lambda.Status.Phase = lambdav1alpha1.PhaseDeleting
	if err := r.Status().Update(ctx, lambda); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	// Delete eventing infrastructure
	eventingWasEnabled := lambda.Spec.Eventing == nil || lambda.Spec.Eventing.Enabled
	if r.EventingManager != nil && eventingWasEnabled {
		log.Info("Deleting eventing infrastructure")
		if err := r.EventingManager.DeleteEventing(ctx, lambda); err != nil {
			log.Error(err, "Failed to delete eventing infrastructure")
			// Continue - don't block deletion
		}
	}

	// Delete Knative service
	if err := r.DeployManager.DeleteService(ctx, lambda); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to delete Knative service")
			return ctrl.Result{}, err
		}
	}

	// Emit event (non-blocking)
	if r.EventManager != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = r.EventManager.EmitServiceDeleted(ctx, lambda)
		}()
	}

	// Clean up build jobs
	if lambda.Status.BuildStatus != nil && lambda.Status.BuildStatus.JobName != "" {
		if err := r.BuildManager.DeleteJob(ctx, lambda.Namespace, lambda.Status.BuildStatus.JobName); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "Failed to delete build job")
			}
		}
		// Update active build jobs metric after deleting a build job
		r.updateActiveBuildJobCount(ctx, lambda.Namespace)
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(lambda, FinalizerName)
	if err := r.Update(ctx, lambda); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	log.Info("Successfully deleted LambdaFunction")
	return ctrl.Result{}, nil
}

// Helper methods

// validateSpec performs comprehensive validation of the LambdaFunction spec
// Security Fixes Applied:
// - BLUE-001: SSRF prevention via Git URL validation
// - BLUE-002: Template injection prevention via handler validation
// - BLUE-005: Path traversal prevention via Git path validation
// - VULN-009: Input validation for all source types
func (r *LambdaFunctionReconciler) validateSpec(lambda *lambdav1alpha1.LambdaFunction) error {
	if lambda.Spec.Source.Type == "" {
		return fmt.Errorf("source type is required")
	}

	switch lambda.Spec.Source.Type {
	case "minio":
		if lambda.Spec.Source.MinIO == nil {
			return fmt.Errorf("minio configuration is required for source type 'minio'")
		}
		if lambda.Spec.Source.MinIO.Bucket == "" || lambda.Spec.Source.MinIO.Key == "" {
			return fmt.Errorf("minio bucket and key are required")
		}
		// Security Fix: Validate MinIO source to prevent SSRF and injection
		if err := validation.ValidateMinIOSource(
			lambda.Spec.Source.MinIO.Endpoint,
			lambda.Spec.Source.MinIO.Bucket,
			lambda.Spec.Source.MinIO.Key,
		); err != nil {
			return fmt.Errorf("minio source validation failed: %w", err)
		}

	case "s3":
		if lambda.Spec.Source.S3 == nil {
			return fmt.Errorf("s3 configuration is required for source type 's3'")
		}
		if lambda.Spec.Source.S3.Bucket == "" || lambda.Spec.Source.S3.Key == "" {
			return fmt.Errorf("s3 bucket and key are required")
		}
		// Security Fix: Validate S3 source to prevent injection
		if err := validation.ValidateS3Source(
			lambda.Spec.Source.S3.Bucket,
			lambda.Spec.Source.S3.Key,
			lambda.Spec.Source.S3.Region,
		); err != nil {
			return fmt.Errorf("s3 source validation failed: %w", err)
		}

	case "gcs":
		if lambda.Spec.Source.GCS == nil {
			return fmt.Errorf("gcs configuration is required for source type 'gcs'")
		}
		if lambda.Spec.Source.GCS.Bucket == "" || lambda.Spec.Source.GCS.Key == "" {
			return fmt.Errorf("gcs bucket and key are required")
		}
		// Security Fix: Validate GCS source (similar to S3)
		if err := validation.ValidateBucketName(lambda.Spec.Source.GCS.Bucket); err != nil {
			return fmt.Errorf("gcs bucket validation failed: %w", err)
		}
		if err := validation.ValidateObjectKey(lambda.Spec.Source.GCS.Key); err != nil {
			return fmt.Errorf("gcs key validation failed: %w", err)
		}

	case "git":
		if lambda.Spec.Source.Git == nil {
			return fmt.Errorf("git configuration is required for source type 'git'")
		}
		if lambda.Spec.Source.Git.URL == "" {
			return fmt.Errorf("git url is required")
		}
		// Security Fix: BLUE-001, BLUE-005 - Validate Git source to prevent SSRF and path traversal
		if err := validation.ValidateGitSource(
			lambda.Spec.Source.Git.URL,
			lambda.Spec.Source.Git.Ref,
			lambda.Spec.Source.Git.Path,
		); err != nil {
			return fmt.Errorf("git source validation failed: %w", err)
		}

	case "inline":
		if lambda.Spec.Source.Inline == nil {
			return fmt.Errorf("inline configuration is required for source type 'inline'")
		}
		if lambda.Spec.Source.Inline.Code == "" {
			return fmt.Errorf("inline code is required")
		}

	case "image":
		if lambda.Spec.Source.Image == nil {
			return fmt.Errorf("image configuration is required for source type 'image'")
		}
		if lambda.Spec.Source.Image.Repository == "" {
			return fmt.Errorf("image repository is required")
		}

	default:
		return fmt.Errorf("unsupported source type: %s", lambda.Spec.Source.Type)
	}

	if lambda.Spec.Runtime.Language == "" {
		return fmt.Errorf("runtime language is required")
	}
	if lambda.Spec.Runtime.Version == "" {
		return fmt.Errorf("runtime version is required")
	}

	// Security Fix: BLUE-002 - Validate handler to prevent template injection
	if err := validation.ValidateHandler(lambda.Spec.Runtime.Handler); err != nil {
		return fmt.Errorf("handler validation failed: %w", err)
	}

	return nil
}

func (r *LambdaFunctionReconciler) setPhase(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, phase lambdav1alpha1.LambdaPhase, log logr.Logger) (ctrl.Result, error) {
	lambda.Status.Phase = phase
	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *LambdaFunctionReconciler) setCondition(lambda *lambdav1alpha1.LambdaFunction, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
	lambda.Status.SetCondition(condition)
}

func (r *LambdaFunctionReconciler) recordError(component, errorType string) {
	if r.Metrics != nil {
		r.Metrics.RecordError(component, errorType)
	}
}

// updateLambdaCounts updates the lambda count metrics for a namespace
func (r *LambdaFunctionReconciler) updateLambdaCounts(ctx context.Context, namespace string) {
	if r.Metrics == nil {
		return
	}

	// List all lambdas in the namespace
	lambdaList := &lambdav1alpha1.LambdaFunctionList{}
	if err := r.List(ctx, lambdaList, client.InNamespace(namespace)); err != nil {
		r.Log.V(1).Info("Failed to list lambdas for metrics", "namespace", namespace, "error", err)
		return
	}

	// Count by phase
	counts := make(map[string]float64)
	for _, lambda := range lambdaList.Items {
		phase := string(lambda.Status.Phase)
		if phase == "" {
			phase = "Pending"
		}
		counts[phase]++
	}

	// Update gauge for each phase
	for phase, count := range counts {
		r.Metrics.SetLambdaCount(namespace, phase, count)
	}
}

// updateActiveBuildJobCount counts active build jobs and updates the metric
// This provides accurate tracking of Kaniko build jobs across the namespace
func (r *LambdaFunctionReconciler) updateActiveBuildJobCount(ctx context.Context, namespace string) {
	if r.Metrics == nil {
		return
	}

	// List all build jobs in the namespace with the build label
	jobList := &batchv1.JobList{}
	if err := r.List(ctx, jobList, client.InNamespace(namespace), client.MatchingLabels{
		"lambda.knative.io/build": "true",
	}); err != nil {
		r.Log.V(1).Info("Failed to list build jobs for metrics", "namespace", namespace, "error", err)
		return
	}

	// Count active (not completed/failed) jobs
	var activeCount float64
	for _, job := range jobList.Items {
		// A job is active if it has no completion time and hasn't failed
		isCompleted := false
		for _, condition := range job.Status.Conditions {
			if (condition.Type == batchv1.JobComplete || condition.Type == batchv1.JobFailed) &&
				condition.Status == corev1.ConditionTrue {
				isCompleted = true
				break
			}
		}
		if !isCompleted {
			activeCount++
		}
	}

	// Update the gauge metric
	r.Metrics.SetActiveBuildJobs(namespace, activeCount)
	r.Log.V(2).Info("Updated active build jobs metric", "namespace", namespace, "count", activeCount)
}

func resultString(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

// SetupWithManager sets up the controller with the Manager
func (r *LambdaFunctionReconciler) SetupWithManager(mgr ctrl.Manager, opts ReconcilerOptions) error {
	// Create an unstructured object for Knative Service to watch
	knativeService := &unstructured.Unstructured{}
	knativeService.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "serving.knative.dev",
		Version: "v1",
		Kind:    "Service",
	})

	// Build controller options for scale
	ctrlOpts := controller.Options{
		MaxConcurrentReconciles: opts.MaxConcurrentReconciles,
	}
	if opts.RateLimiter != nil {
		ctrlOpts.RateLimiter = opts.RateLimiter
	}

	// Use predicates to filter unnecessary reconciles
	generationChangedPredicate := predicate.GenerationChangedPredicate{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&lambdav1alpha1.LambdaFunction{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.ConfigMap{}).
		Owns(knativeService).
		WithEventFilter(generationChangedPredicate).
		WithOptions(ctrlOpts).
		Complete(r)
}
