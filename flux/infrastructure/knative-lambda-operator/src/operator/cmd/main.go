package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/controllers"
	"github.com/brunovlucena/knative-lambda-operator/internal/build"
	"github.com/brunovlucena/knative-lambda-operator/internal/deploy"
	"github.com/brunovlucena/knative-lambda-operator/internal/eventing"
	"github.com/brunovlucena/knative-lambda-operator/internal/events"
	"github.com/brunovlucena/knative-lambda-operator/internal/metrics"
	"github.com/brunovlucena/knative-lambda-operator/internal/observability"
	"github.com/brunovlucena/knative-lambda-operator/internal/webhook"
)

// Mode determines how the operator runs
type Mode string

const (
	ModeController Mode = "controller" // CRD controller mode (default)
	ModeReceiver   Mode = "receiver"   // CloudEvents receiver mode (Knative Service)
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(lambdav1alpha1.AddToScheme(scheme))
	utilruntime.Must(servingv1.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
}

func main() {
	// CLI flags
	var (
		// Mode selection
		mode string

		// Controller mode flags
		metricsAddr          string
		probeAddr            string
		enableLeaderElection bool
		enableEvents         bool
		brokerURL            string
		// Scale tuning
		maxConcurrentReconciles int
		// Namespace filtering for sharding
		watchNamespace string

		// Receiver mode flags (CloudEvents Knative Service)
		defaultNamespace string
		rateLimit        float64
		burstSize        int

		// OpenTelemetry tracing flags
		enableTracing       bool
		otlpEndpoint        string
		tracingSamplingRate float64
	)

	flag.StringVar(&mode, "mode", "controller", "Run mode: 'controller' (CRD controller) or 'receiver' (CloudEvents Knative Service)")

	// Controller mode flags
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.BoolVar(&enableEvents, "enable-events", true, "Enable CloudEvents emission to broker.")
	flag.StringVar(&brokerURL, "broker-url", "", "URL of the Knative broker for CloudEvents.")

	// Scale tuning flags - critical for handling 10K+ concurrent requests
	flag.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", 50, "Maximum concurrent reconciles. For 10K lambdas, use 50-100.")

	// Namespace sharding - run multiple operators with different namespace filters
	flag.StringVar(&watchNamespace, "watch-namespace", "", "Namespace to watch. Empty = all namespaces. Use for sharding.")

	// Receiver mode flags
	flag.StringVar(&defaultNamespace, "default-namespace", "knative-lambda", "Default namespace for Lambda functions.")
	flag.Float64Var(&rateLimit, "rate-limit", 50, "Rate limit for K8s API calls (requests/second)")
	flag.IntVar(&burstSize, "burst-size", 100, "Burst size for rate limiting")

	// OpenTelemetry tracing flags
	flag.BoolVar(&enableTracing, "enable-tracing", true, "Enable OpenTelemetry distributed tracing.")
	flag.StringVar(&otlpEndpoint, "otlp-endpoint", "", "OTLP collector endpoint (default: alloy.observability.svc:4317 or OTEL_EXPORTER_OTLP_ENDPOINT env var)")
	flag.Float64Var(&tracingSamplingRate, "tracing-sampling-rate", 1.0, "Trace sampling rate (0.0-1.0, default: 1.0 for 100%)")

	opts := zap.Options{
		Development: false, // Production mode - structured JSON logs
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Initialize OpenTelemetry tracing
	var otelProvider *observability.Provider
	if enableTracing {
		otelConfig := observability.DefaultConfig()
		if otlpEndpoint != "" {
			otelConfig.OTLPEndpoint = otlpEndpoint
		}
		if tracingSamplingRate >= 0 && tracingSamplingRate <= 1.0 {
			otelConfig.TracingSamplingRate = tracingSamplingRate
		}
		otelConfig.TracingEnabled = true
		otelConfig.MetricsEnabled = false // Use Prometheus metrics instead of OTEL metrics

		var err error
		otelProvider, err = observability.NewProvider(otelConfig)
		if err != nil {
			setupLog.Error(err, "Failed to initialize OpenTelemetry tracing, continuing without tracing")
		} else {
			observability.SetGlobalProvider(otelProvider)
			setupLog.Info("OpenTelemetry tracing initialized",
				"endpoint", otelConfig.OTLPEndpoint,
				"samplingRate", otelConfig.TracingSamplingRate,
				"podName", otelConfig.PodName,
				"podNamespace", otelConfig.PodNamespace,
			)

			// Ensure graceful shutdown of tracing on exit
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := otelProvider.Shutdown(shutdownCtx); err != nil {
					setupLog.Error(err, "Error shutting down OpenTelemetry tracing")
				}
			}()
		}
	}

	// Run in the selected mode
	if mode == "receiver" {
		runReceiverMode(defaultNamespace, rateLimit, burstSize, otelProvider)
		return
	}

	// Default: controller mode
	setupLog.Info("Starting knative-lambda-operator (controller mode)",
		"maxConcurrentReconciles", maxConcurrentReconciles,
		"watchNamespace", watchNamespace,
		"tracingEnabled", enableTracing,
	)

	// Register custom metrics
	metrics.Register()

	// Manager options for scale
	mgrOptions := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "lambdafunction.lambda.knative.io",
	}

	// Namespace filtering for horizontal sharding
	if watchNamespace != "" {
		mgrOptions.Cache = cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				watchNamespace: {},
			},
		}
		setupLog.Info("Operator will only watch namespace", "namespace", watchNamespace)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize managers with proper error handling
	buildManager, err := build.NewManager(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to create build manager")
		os.Exit(1)
	}

	deployManager := deploy.NewManager(mgr.GetClient(), mgr.GetScheme())

	// Initialize event manager (CloudEvents emission)
	var eventManager *events.Manager
	if enableEvents {
		eventConfig := events.Config{
			BrokerURL: brokerURL,
			Enabled:   enableEvents,
		}
		eventManager = events.NewManager(eventConfig)
	}

	// Initialize eventing manager (Brokers, Triggers, DLQ)
	eventingManager, err := eventing.NewManager(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to create eventing manager")
		os.Exit(1)
	}

	// Setup controller with scale tuning
	reconciler := &controllers.LambdaFunctionReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Log:             ctrl.Log.WithName("controllers").WithName("LambdaFunction"),
		BuildManager:    buildManager,
		DeployManager:   deployManager,
		EventManager:    eventManager,
		EventingManager: eventingManager,
		Metrics:         metrics.NewReconcilerMetrics(),
		OTELProvider:    otelProvider,
	}

	// Custom rate limiter for work queue - critical for scale
	rateLimiter := workqueue.NewTypedMaxOfRateLimiter(
		// Exponential backoff for failures
		workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			5*time.Millisecond, // base delay
			1000*time.Second,   // max delay
		),
		// Token bucket for overall rate limiting
		workqueue.NewTypedItemFastSlowRateLimiter[ctrl.Request](
			5*time.Millisecond, // fast delay
			30*time.Second,     // slow delay
			100,                // max fast attempts
		),
	)

	if err = reconciler.SetupWithManager(mgr, controllers.ReconcilerOptions{
		MaxConcurrentReconciles: maxConcurrentReconciles,
		RateLimiter:             rateLimiter,
	}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LambdaFunction")
		os.Exit(1)
	}

	// Setup LambdaAgent controller with eventing support
	agentReconciler := &controllers.LambdaAgentReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		EventingManager: eventingManager,
	}
	if err = agentReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LambdaAgent")
		os.Exit(1)
	}
	setupLog.Info("LambdaAgent controller registered with eventing support")

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

// runReceiverMode runs the operator as a CloudEvents receiver (Knative Service)
// This mode is designed for high-throughput event processing:
//   - Receives events via Knative Triggers from RabbitMQ-backed Broker
//   - Auto-scales based on queue depth (0 to N replicas)
//   - Rate limits K8s API calls to prevent overload
//   - Returns proper HTTP status for retry/DLQ handling
//   - Supports OpenTelemetry tracing for distributed request tracing
func runReceiverMode(defaultNamespace string, rateLimit float64, burstSize int, otelProvider *observability.Provider) {
	setupLog.Info("Starting knative-lambda-operator (receiver mode)",
		"defaultNamespace", defaultNamespace,
		"rateLimit", rateLimit,
		"burstSize", burstSize,
	)

	// Create minimal manager for client access only
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Disable controller features - we only need the client
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "",
		LeaderElection:         false,
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager for receiver mode")
		os.Exit(1)
	}

	// Configure receiver
	debug := os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"
	config := webhook.ReceiverConfig{
		Port:              8080, // Knative injects PORT, but defaults to 8080
		Path:              "/",  // Knative Triggers POST to root path
		DefaultNamespace:  defaultNamespace,
		RateLimit:         rateLimit,
		BurstSize:         burstSize,
		ProcessingTimeout: 30 * time.Second,
		Debug:             debug,
	}

	// Check for PORT env var (Knative sets this)
	if port := os.Getenv("PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			config.Port = p
		}
	}

	receiver, err := webhook.NewReceiver(
		mgr.GetClient(),
		ctrl.Log.WithName("receiver"),
		config,
		nil, // No event emission in receiver mode
	)
	if err != nil {
		setupLog.Error(err, "unable to create receiver")
		os.Exit(1)
	}

	// Start manager cache in background (needed for client to work)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := mgr.Start(ctx); err != nil {
			setupLog.Error(err, "manager error")
			os.Exit(1)
		}
	}()

	// Wait for cache to sync
	if !mgr.GetCache().WaitForCacheSync(ctx) {
		setupLog.Error(nil, "failed to sync cache")
		os.Exit(1)
	}

	// Run receiver (blocking)
	if err := receiver.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "receiver error")
		os.Exit(1)
	}
}
