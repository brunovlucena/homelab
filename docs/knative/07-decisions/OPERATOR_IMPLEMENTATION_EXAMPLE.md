# Operator Implementation Example

## 📝 Overview

This document provides a concrete implementation example for the Knative Lambda Operator using Kubebuilder/controller-runtime.

## 🏗️ Project Structure

```
operator/
├── api/
│   └── v1alpha1/
│       ├── lambdafunction_types.go
│       └── groupversion_info.go
├── controllers/
│   └── lambdafunction_controller.go
├── internal/
│   ├── build/
│   │   └── manager.go
│   └── deploy/
│       └── manager.go
└── main.go
```

## 📦 CRD Types (api/v1alpha1/lambdafunction_types.go)

```go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LambdaFunctionSpec defines the desired state of LambdaFunction
type LambdaFunctionSpec struct {
	// Source configuration
	Source SourceSpec `json:"source"`
	
	// Runtime configuration
	Runtime RuntimeSpec `json:"runtime"`
	
	// Scaling configuration
	Scaling ScalingSpec `json:"scaling,omitempty"`
	
	// Resource limits
	Resources ResourceSpec `json:"resources,omitempty"`
	
	// Environment variables
	Env []EnvVar `json:"env,omitempty"`
	
	// Event triggers
	Triggers []TriggerSpec `json:"triggers,omitempty"`
	
	// Build configuration
	Build BuildSpec `json:"build,omitempty"`
}

// SourceSpec defines the source code location
type SourceSpec struct {
	Type string `json:"type"` // s3, git, inline
	
	S3 *S3Source `json:"s3,omitempty"`
	Git *GitSource `json:"git,omitempty"`
	Inline *InlineSource `json:"inline,omitempty"`
}

type S3Source struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Region string `json:"region,omitempty"`
}

type GitSource struct {
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"`
	Path string `json:"path,omitempty"`
}

type InlineSource struct {
	Code        string `json:"code"`
	Dependencies string `json:"dependencies,omitempty"`
}

// RuntimeSpec defines the runtime configuration
type RuntimeSpec struct {
	Language string `json:"language"` // nodejs, python, go
	Version  string `json:"version"`
	Handler  string `json:"handler,omitempty"`
}

// ScalingSpec defines autoscaling configuration
type ScalingSpec struct {
	MinReplicas           *int32 `json:"minReplicas,omitempty"`
	MaxReplicas           *int32 `json:"maxReplicas,omitempty"`
	TargetConcurrency     *int32 `json:"targetConcurrency,omitempty"`
	ScaleToZeroGracePeriod string `json:"scaleToZeroGracePeriod,omitempty"`
}

// ResourceSpec defines resource limits
type ResourceSpec struct {
	Requests ResourceRequirements `json:"requests,omitempty"`
	Limits   ResourceRequirements `json:"limits,omitempty"`
}

type ResourceRequirements struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}

// EnvVar defines environment variable
type EnvVar struct {
	Name      string              `json:"name"`
	Value     string              `json:"value,omitempty"`
	ValueFrom *EnvVarSource       `json:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// TriggerSpec defines event trigger
type TriggerSpec struct {
	Broker string            `json:"broker"`
	Filter map[string]string `json:"filter,omitempty"`
}

// BuildSpec defines build configuration
type BuildSpec struct {
	Timeout        string `json:"timeout,omitempty"`
	Registry       string `json:"registry,omitempty"`
	ImagePullSecret string `json:"imagePullSecret,omitempty"`
}

// LambdaFunctionStatus defines the observed state of LambdaFunction
type LambdaFunctionStatus struct {
	// Phase represents the current phase
	Phase string `json:"phase,omitempty"`
	
	// BuildStatus represents build information
	BuildStatus *BuildStatus `json:"buildStatus,omitempty"`
	
	// ServiceStatus represents service information
	ServiceStatus *ServiceStatus `json:"serviceStatus,omitempty"`
	
	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// BuildStatus represents build state
type BuildStatus struct {
	JobName     string `json:"jobName,omitempty"`
	ImageURI    string `json:"imageURI,omitempty"`
	StartedAt   *metav1.Time `json:"startedAt,omitempty"`
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ServiceStatus represents service state
type ServiceStatus struct {
	ServiceName string `json:"serviceName,omitempty"`
	URL         string `json:"url,omitempty"`
	Ready       bool   `json:"ready,omitempty"`
	Replicas    int32  `json:"replicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.buildStatus.imageURI"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.serviceStatus.ready"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// LambdaFunction is the Schema for the lambdafunctions API
type LambdaFunction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LambdaFunctionSpec   `json:"spec,omitempty"`
	Status LambdaFunctionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LambdaFunctionList contains a list of LambdaFunction
type LambdaFunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LambdaFunction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LambdaFunction{}, &LambdaFunctionList{})
}
```

## 🎮 Controller Implementation (controllers/lambdafunction_controller.go)

```go
package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	lambdaapi "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/build"
	"github.com/brunovlucena/knative-lambda-operator/internal/deploy"
)

// LambdaFunctionReconciler reconciles a LambdaFunction object
type LambdaFunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
	
	BuildManager  *build.Manager
	DeployManager *deploy.Manager
}

//+kubebuilder:rbac:groups=lambda.knative.io,resources=lambdafunctions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=lambda.knative.io,resources=lambdafunctions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=lambda.knative.io,resources=lambdafunctions/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *LambdaFunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("lambdafunction", req.NamespacedName)

	// Fetch the LambdaFunction instance
	lambda := &lambdaapi.LambdaFunction{}
	if err := r.Get(ctx, req.NamespacedName, lambda); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !lambda.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, lambda, log)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(lambda, "lambdafunction.lambda.knative.io/finalizer") {
		controllerutil.AddFinalizer(lambda, "lambdafunction.lambda.knative.io/finalizer")
		if err := r.Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile based on current phase
	switch lambda.Status.Phase {
	case "", "Pending":
		return r.reconcilePending(ctx, lambda, log)
	case "Building":
		return r.reconcileBuilding(ctx, lambda, log)
	case "Deploying":
		return r.reconcileDeploying(ctx, lambda, log)
	case "Ready":
		return r.reconcileReady(ctx, lambda, log)
	case "Failed":
		return r.reconcileFailed(ctx, lambda, log)
	default:
		log.Info("Unknown phase, resetting to Pending", "phase", lambda.Status.Phase)
		return r.setPhase(ctx, lambda, "Pending", log)
	}
}

func (r *LambdaFunctionReconciler) reconcilePending(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Pending phase")

	// Validate spec
	if err := r.validateSpec(lambda); err != nil {
		return r.setCondition(ctx, lambda, "SourceReady", "False", "ValidationFailed", err.Error(), log)
	}

	// Create build context
	buildCtx, err := r.BuildManager.CreateBuildContext(ctx, lambda)
	if err != nil {
		return r.setCondition(ctx, lambda, "SourceReady", "False", "BuildContextFailed", err.Error(), log)
	}

	// Update condition
	if err := r.setCondition(ctx, lambda, "SourceReady", "True", "BuildContextCreated", "Build context created successfully", log); err != nil {
		return ctrl.Result{}, err
	}

	// Create Kaniko job
	job, err := r.BuildManager.CreateKanikoJob(ctx, lambda, buildCtx)
	if err != nil {
		return r.setCondition(ctx, lambda, "BuildReady", "False", "JobCreationFailed", err.Error(), log)
	}

	// Update status
	lambda.Status.Phase = "Building"
	lambda.Status.BuildStatus = &lambdaapi.BuildStatus{
		JobName:   job.Name,
		StartedAt: &metav1.Time{Time: time.Now()},
	}

	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *LambdaFunctionReconciler) reconcileBuilding(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Building phase")

	// Check build job status
	status, err := r.BuildManager.GetBuildStatus(ctx, lambda)
	if err != nil {
		return ctrl.Result{}, err
	}

	if status.Completed {
		if status.Success {
			// Build succeeded, move to Deploying
			lambda.Status.Phase = "Deploying"
			lambda.Status.BuildStatus.ImageURI = status.ImageURI
			lambda.Status.BuildStatus.CompletedAt = &metav1.Time{Time: time.Now()}
			
			if err := r.setCondition(ctx, lambda, "BuildReady", "True", "BuildSucceeded", "Image built successfully", log); err != nil {
				return ctrl.Result{}, err
			}
			
			if err := r.Status().Update(ctx, lambda); err != nil {
				return ctrl.Result{}, err
			}
			
			return ctrl.Result{Requeue: true}, nil
		} else {
			// Build failed
			lambda.Status.Phase = "Failed"
			lambda.Status.BuildStatus.Error = status.Error
			lambda.Status.BuildStatus.CompletedAt = &metav1.Time{Time: time.Now()}
			
			if err := r.setCondition(ctx, lambda, "BuildReady", "False", "BuildFailed", status.Error, log); err != nil {
				return ctrl.Result{}, err
			}
			
			if err := r.Status().Update(ctx, lambda); err != nil {
				return ctrl.Result{}, err
			}
			
			return ctrl.Result{}, nil
		}
	}

	// Still building, requeue
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *LambdaFunctionReconciler) reconcileDeploying(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Deploying phase")

	// Check if service exists
	service, err := r.DeployManager.GetService(ctx, lambda)
	if err != nil && !client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	if service == nil {
		// Create Knative service
		service, err = r.DeployManager.CreateService(ctx, lambda)
		if err != nil {
			return r.setCondition(ctx, lambda, "DeployReady", "False", "ServiceCreationFailed", err.Error(), log)
		}
	}

	// Check service status
	ready, url, replicas, err := r.DeployManager.GetServiceStatus(ctx, service)
	if err != nil {
		return ctrl.Result{}, err
	}

	if ready {
		// Service is ready
		lambda.Status.Phase = "Ready"
		lambda.Status.ServiceStatus = &lambdaapi.ServiceStatus{
			ServiceName: service.Name,
			URL:         url,
			Ready:       true,
			Replicas:    replicas,
		}

		if err := r.setCondition(ctx, lambda, "DeployReady", "True", "ServiceDeployed", "Service deployed successfully", log); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.setCondition(ctx, lambda, "ServiceReady", "True", "ServiceReady", "Service is ready", log); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Status().Update(ctx, lambda); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Still deploying, update status and requeue
	lambda.Status.ServiceStatus = &lambdaapi.ServiceStatus{
		ServiceName: service.Name,
		Ready:       false,
		Replicas:    replicas,
	}

	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *LambdaFunctionReconciler) reconcileReady(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Ready phase")

	// Monitor service health
	service, err := r.DeployManager.GetService(ctx, lambda)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Service deleted, reset to Pending
			return r.setPhase(ctx, lambda, "Pending", log)
		}
		return ctrl.Result{}, err
	}

	ready, url, replicas, err := r.DeployManager.GetServiceStatus(ctx, service)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	lambda.Status.ServiceStatus.Ready = ready
	lambda.Status.ServiceStatus.URL = url
	lambda.Status.ServiceStatus.Replicas = replicas

	if !ready {
		lambda.Status.Phase = "Failed"
		if err := r.setCondition(ctx, lambda, "ServiceReady", "False", "ServiceUnhealthy", "Service is not ready", log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *LambdaFunctionReconciler) reconcileFailed(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling Failed phase")
	// Could implement retry logic here
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

func (r *LambdaFunctionReconciler) reconcileDelete(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
	log.Info("Reconciling deletion")

	// Delete Knative service
	if err := r.DeployManager.DeleteService(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(lambda, "lambdafunction.lambda.knative.io/finalizer")
	if err := r.Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Helper methods
func (r *LambdaFunctionReconciler) validateSpec(lambda *lambdaapi.LambdaFunction) error {
	// Validation logic
	return nil
}

func (r *LambdaFunctionReconciler) setPhase(ctx context.Context, lambda *lambdaapi.LambdaFunction, phase string, log logr.Logger) (ctrl.Result, error) {
	lambda.Status.Phase = phase
	if err := r.Status().Update(ctx, lambda); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *LambdaFunctionReconciler) setCondition(ctx context.Context, lambda *lambdaapi.LambdaFunction, conditionType, status, reason, message string, log logr.Logger) error {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             metav1.ConditionStatus(status),
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}

	meta.SetStatusCondition(&lambda.Status.Conditions, condition)
	return r.Status().Update(ctx, lambda)
}

// SetupWithManager sets up the controller with the Manager
func (r *LambdaFunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lambdaapi.LambdaFunction{}).
		Owns(&batchv1.Job{}).
		Owns(&servingv1.Service{}).
		Complete(r)
}
```

## 🚀 Main Entry Point (main.go)

```go
package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	lambdaapi "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/controllers"
	"github.com/brunovlucena/knative-lambda-operator/internal/build"
	"github.com/brunovlucena/knative-lambda-operator/internal/deploy"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(lambdaapi.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "lambdafunction.lambda.knative.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize managers
	buildManager := build.NewManager(mgr.GetClient(), mgr.GetScheme())
	deployManager := deploy.NewManager(mgr.GetClient(), mgr.GetScheme())

	// Setup controller
	if err = (&controllers.LambdaFunctionReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Log:           ctrl.Log.WithName("controllers").WithName("LambdaFunction"),
		BuildManager:  buildManager,
		DeployManager: deployManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LambdaFunction")
		os.Exit(1)
	}

	// Health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
```

## 📋 Next Steps

1. **Initialize Kubebuilder project**: `kubebuilder init --domain lambda.knative.io`
2. **Create API**: `kubebuilder create api --group lambda --version v1alpha1 --kind LambdaFunction`
3. **Implement managers**: Build and Deploy managers
4. **Add tests**: Unit and integration tests
5. **Build and deploy**: Create Docker image and deploy operator

