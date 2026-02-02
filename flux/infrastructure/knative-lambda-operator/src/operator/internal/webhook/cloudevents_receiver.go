package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/events"
	"github.com/brunovlucena/knative-lambda-operator/internal/metrics"
	"github.com/brunovlucena/knative-lambda-operator/internal/schema"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  🌐 CLOUDEVENTS RECEIVER - KNATIVE SERVICE ENDPOINT
//
//  This receiver is designed to be deployed as a Knative Service and receive
//  events via Knative Triggers from a RabbitMQ-backed Broker.
//
//  Architecture for 1000+ concurrent events:
//
//    External Services → Broker → RabbitMQ Queue → Trigger → This Service
//                                 (buffered)                   (auto-scaled)
//
//  Key features:
//  - Rate limiting to protect K8s API server
//  - Batch processing for efficiency
//  - Proper CloudEvents acknowledgment (202 vs 500)
//  - Metrics for observability
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ReceiverConfig configures the CloudEvents receiver
type ReceiverConfig struct {
	// Port to listen on (Knative injects PORT env var)
	Port int

	// Path for CloudEvents (Knative expects root path)
	Path string

	// Default namespace for Lambda functions
	DefaultNamespace string

	// Rate limiting for K8s API protection
	// K8s API server typically allows ~100 QPS per client
	RateLimit float64 // requests per second
	BurstSize int     // max burst

	// Worker pool size for concurrent processing
	WorkerPoolSize int

	// Processing timeout per event
	ProcessingTimeout time.Duration

	// Schema validation (enabled by default)
	EnableSchemaValidation bool

	// Debug mode - enables verbose logging for troubleshooting
	Debug bool
}

// DefaultReceiverConfig returns production-ready defaults
func DefaultReceiverConfig() ReceiverConfig {
	return ReceiverConfig{
		Port:                   8080, // Knative default
		Path:                   "/",  // Knative Trigger expects root
		DefaultNamespace:       "knative-lambda",
		RateLimit:              50,  // 50 QPS to K8s API (safe margin)
		BurstSize:              100, // Allow bursts up to 100
		WorkerPoolSize:         10,  // 10 concurrent workers
		ProcessingTimeout:      30 * time.Second,
		EnableSchemaValidation: true, // Schema validation enabled by default
	}
}

// Receiver handles CloudEvents for Lambda management
// Designed for high-throughput event processing via Knative Eventing
type Receiver struct {
	client          client.Client
	log             logr.Logger
	config          ReceiverConfig
	eventManager    *events.Manager
	schemaValidator *schema.Validator
	server          *http.Server

	// Rate limiter to protect K8s API
	rateLimiter *rate.Limiter

	// Metrics
	mu                      sync.RWMutex
	eventsReceived          int64
	eventsProcessed         int64
	eventsFailed            int64
	eventsSchemaValidFailed int64
}

// NewReceiver creates a CloudEvents receiver for Knative Service deployment
func NewReceiver(c client.Client, log logr.Logger, config ReceiverConfig, eventManager *events.Manager) (*Receiver, error) {
	r := &Receiver{
		client:       c,
		log:          log.WithName("cloudevents-receiver"),
		config:       config,
		eventManager: eventManager,
		rateLimiter:  rate.NewLimiter(rate.Limit(config.RateLimit), config.BurstSize),
	}

	// Initialize schema validator if enabled
	if config.EnableSchemaValidation {
		validator, err := schema.NewValidator()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize schema validator: %w", err)
		}
		r.schemaValidator = validator
		log.Info("Schema validation enabled", "registeredSchemas", len(validator.RegisteredEventTypes()))
	} else {
		log.Info("Schema validation DISABLED - CloudEvents will not be validated against schemas")
	}

	return r, nil
}

// Start starts the receiver server
func (r *Receiver) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Main CloudEvents endpoint (Knative sends to root path)
	mux.HandleFunc(r.config.Path, r.handleCloudEvent)

	// Alternative CloudEvents endpoint (optional, for direct HTTP access)
	// POST /events also accepts CloudEvents for compatibility
	if r.config.Path != "/events" {
		mux.HandleFunc("/events", r.handleCloudEvent)
	}

	// Health endpoints for Knative probes
	mux.HandleFunc("/health", r.healthHandler)
	mux.HandleFunc("/ready", r.readyHandler)

	// Metrics endpoint
	mux.HandleFunc("/metrics", r.metricsHandler)

	r.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", r.config.Port),
		Handler:      mux,
		ReadTimeout:  r.config.ProcessingTimeout,
		WriteTimeout: r.config.ProcessingTimeout,
	}

	r.log.Info("Starting CloudEvents receiver (Knative Service mode)",
		"port", r.config.Port,
		"path", r.config.Path,
		"rateLimit", r.config.RateLimit,
		"burstSize", r.config.BurstSize,
	)

	// Start server
	errCh := make(chan error, 1)
	go func() {
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context or error
	select {
	case <-ctx.Done():
		r.log.Info("Shutting down CloudEvents receiver")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return r.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// handleCloudEvent processes incoming CloudEvents with rate limiting
func (r *Receiver) handleCloudEvent(w http.ResponseWriter, req *http.Request) {
	r.incrementReceived()

	// Only accept POST
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get or generate correlation ID for distributed tracing
	correlationID := req.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	// Always set correlation ID in response headers
	w.Header().Set("X-Correlation-ID", correlationID)

	// Debug: Log request details before parsing
	if r.config.Debug {
		contentType := req.Header.Get("Content-Type")
		bodyBytes := []byte{}
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		bodyPreview := string(bodyBytes)
		if len(bodyPreview) > 500 {
			bodyPreview = bodyPreview[:500] + "..."
		}
		r.log.Info("DEBUG: HTTP request before parsing",
			"contentType", contentType,
			"bodyLength", len(bodyBytes),
			"bodyPreview", bodyPreview,
			"headers", fmt.Sprintf("%v", req.Header),
			"correlationId", correlationID)
	}

	// Parse CloudEvent
	event, err := cehttp.NewEventFromHTTPRequest(req)
	if err != nil {
		r.log.Error(err, "Failed to parse CloudEvent", "correlationId", correlationID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "error",
			"error":         "invalid_cloudevent",
			"message":       fmt.Sprintf("Invalid CloudEvent: %v", err),
			"correlationId": correlationID,
		})
		r.incrementFailed()
		return
	}

	// Log all received CloudEvents with full context for observability
	r.log.Info("CloudEvent received",
		"eventId", event.ID(),
		"eventType", event.Type(),
		"eventSource", event.Source(),
		"eventSubject", event.Subject(),
		"correlationId", correlationID,
		"dataContentType", event.DataContentType(),
		"hasData", event.Data() != nil,
	)

	// Apply rate limiting - wait or reject
	ctx, cancel := context.WithTimeout(req.Context(), r.config.ProcessingTimeout)
	defer cancel()

	if err := r.rateLimiter.Wait(ctx); err != nil {
		r.log.Error(err, "Rate limit exceeded", "eventId", event.ID(), "correlationId", correlationID)
		// Return 429 so Knative/RabbitMQ will retry later
		w.Header().Set("Retry-After", "5")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "error",
			"error":         "rate_limit_exceeded",
			"message":       "Rate limit exceeded, retry later",
			"eventId":       event.ID(),
			"correlationId": correlationID,
		})
		return
	}

	// Process the event
	if err := r.processEvent(ctx, event); err != nil {
		r.log.Error(err, "Failed to process CloudEvent",
			"id", event.ID(),
			"type", event.Type(),
			"correlationId", correlationID,
		)
		r.incrementFailed()

		// Return appropriate status for retry logic
		status := http.StatusInternalServerError
		if IsSchemaValidationError(err) {
			// 🔐 Schema validation errors - don't retry, bad payload
			status = http.StatusBadRequest
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":        "error",
				"eventId":       event.ID(),
				"eventType":     event.Type(),
				"error":         "schema_validation_failed",
				"message":       err.Error(),
				"correlationId": correlationID,
			})
			return
		} else if apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err) {
			// Don't retry conflicts - return success
			status = http.StatusAccepted
		} else if apierrors.IsInvalid(err) {
			// Don't retry validation errors
			status = http.StatusBadRequest
		}
		// 5xx will trigger retry by Knative/RabbitMQ

		if status >= 500 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":        "error",
				"eventId":       event.ID(),
				"eventType":     event.Type(),
				"error":         "processing_failed",
				"message":       err.Error(),
				"correlationId": correlationID,
			})
			return
		}
	}

	r.incrementProcessed()

	r.log.Info("CloudEvent processed successfully",
		"eventId", event.ID(),
		"eventType", event.Type(),
		"correlationId", correlationID,
	)

	// Return 202 Accepted - event processed successfully
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "accepted",
		"eventId":       event.ID(),
		"correlationId": correlationID,
	})
}

// processEvent handles the actual event processing
func (r *Receiver) processEvent(ctx context.Context, event *cloudevents.Event) error {
	// 🔐 SCHEMA VALIDATION - Validate event payload against JSON schema
	if r.schemaValidator != nil {
		var eventData interface{}
		if err := event.DataAs(&eventData); err != nil {
			r.log.Error(err, "Schema validation: failed to parse event data",
				"eventId", event.ID(),
				"eventType", event.Type(),
				"dataContentType", event.DataContentType(),
				"dataType", fmt.Sprintf("%T", event.Data()),
				"dataLength", len(event.Data()),
			)
			if event.Data() != nil && len(event.Data()) > 0 {
				// Log first 500 chars of raw data to see what we're dealing with
				dataPreview := string(event.Data())
				if len(dataPreview) > 500 {
					dataPreview = dataPreview[:500] + "..."
				}
				r.log.Info("Raw event data preview",
					"eventId", event.ID(),
					"dataPreview", dataPreview)
			}
			r.incrementSchemaValidationFailed()
			return &SchemaValidationError{
				EventID:   event.ID(),
				EventType: event.Type(),
				Message:   fmt.Sprintf("failed to parse event data: %v", err),
			}
		}

		if err := r.schemaValidator.Validate(event.Type(), eventData); err != nil {
			r.incrementSchemaValidationFailed()
			r.log.Error(err, "Schema validation failed",
				"eventId", event.ID(),
				"eventType", event.Type(),
			)
			return &SchemaValidationError{
				EventID:   event.ID(),
				EventType: event.Type(),
				Message:   err.Error(),
			}
		}
		r.log.V(2).Info("Schema validation passed", "eventType", event.Type())
	}

	switch event.Type() {
	// Command events - Lambda lifecycle management
	case events.EventTypeCommandFunctionDeploy:
		return r.handleFunctionDeploy(ctx, event)
	case events.EventTypeCommandServiceCreate:
		return r.handleFunctionDeploy(ctx, event) // Alias
	case events.EventTypeCommandServiceUpdate:
		return r.handleFunctionDeploy(ctx, event) // Alias
	case events.EventTypeCommandServiceDelete:
		return r.handleServiceDelete(ctx, event)
	case events.EventTypeCommandBuildStart:
		return r.handleBuildStart(ctx, event)
	case events.EventTypeCommandBuildRetry:
		return r.handleBuildStart(ctx, event) // Alias
	case events.EventTypeCommandBuildCancel:
		return r.handleBuildCancel(ctx, event)
	case events.EventTypeCommandFunctionRollback:
		return r.handleFunctionRollback(ctx, event)
	// Response events - Function invocation metrics (RED metrics)
	case events.EventTypeResponseSuccess:
		return r.handleResponseSuccess(ctx, event)
	case events.EventTypeResponseError:
		return r.handleResponseError(ctx, event)
	default:
		r.log.V(1).Info("Ignoring unhandled event type", "type", event.Type())
		return nil // Don't fail on unknown types
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 EVENT DATA STRUCTURES                                               │
// └─────────────────────────────────────────────────────────────────────────┘

// FunctionDeployData represents the payload for function.deploy events
type FunctionDeployData struct {
	Metadata FunctionMetadata                  `json:"metadata"`
	Spec     lambdav1alpha1.LambdaFunctionSpec `json:"spec"`
}

// FunctionMetadata represents metadata for function events
type FunctionMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ServiceDeleteData represents the payload for service.delete events
type ServiceDeleteData struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// BuildCommandData represents the payload for build command events
type BuildCommandData struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace,omitempty"`
	ForceRebuild bool   `json:"forceRebuild,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 COMMAND HANDLERS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func (r *Receiver) handleFunctionDeploy(ctx context.Context, event *cloudevents.Event) error {
	var data FunctionDeployData

	// Debug: Log raw event data BEFORE parsing
	if r.config.Debug {
		if event.Data() != nil {
			dataBytes := event.Data()
			previewLen := len(dataBytes)
			if previewLen > 500 {
				previewLen = 500
			}
			r.log.Info("DEBUG: Raw event data before parse",
				"eventId", event.ID(),
				"dataContentType", event.DataContentType(),
				"dataType", fmt.Sprintf("%T", event.Data()),
				"dataLength", len(dataBytes),
				"dataPreview", string(dataBytes[:previewLen]))
		} else {
			r.log.Info("DEBUG: Raw event data is nil",
				"eventId", event.ID(),
				"dataContentType", event.DataContentType())
		}
	}

	if err := event.DataAs(&data); err != nil {
		r.log.Error(err, "Failed to parse event data",
			"eventId", event.ID(),
			"eventType", event.Type(),
			"dataContentType", event.DataContentType(),
			"dataType", fmt.Sprintf("%T", event.Data()),
			"dataLength", len(event.Data()),
		)
		if event.Data() != nil && len(event.Data()) > 0 {
			dataPreview := string(event.Data())
			if len(dataPreview) > 500 {
				dataPreview = dataPreview[:500] + "..."
			}
			r.log.Info("Raw event data preview",
				"eventId", event.ID(),
				"dataPreview", dataPreview)
		}
		return fmt.Errorf("failed to parse function deploy data: %w", err)
	}

	// Debug: Log parsed data
	if r.config.Debug {
		r.log.Info("DEBUG: Parsed function deploy data",
			"eventId", event.ID(),
			"metadataName", data.Metadata.Name,
			"metadataNamespace", data.Metadata.Namespace,
			"sourceType", data.Spec.Source.Type,
			"hasMinIO", data.Spec.Source.MinIO != nil,
			"hasS3", data.Spec.Source.S3 != nil,
			"hasGit", data.Spec.Source.Git != nil,
			"hasInline", data.Spec.Source.Inline != nil,
			"hasImage", data.Spec.Source.Image != nil)
	}

	if data.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	// Validate source type is set
	if data.Spec.Source.Type == "" {
		r.log.Error(nil, "Source type is empty after parsing",
			"eventId", event.ID(),
			"metadataName", data.Metadata.Name)
		if r.config.Debug {
			var rawData interface{}
			if event.Data() != nil {
				json.Unmarshal(event.Data(), &rawData)
				rawDataJSON, _ := json.MarshalIndent(rawData, "", "  ")
				r.log.Info("DEBUG: Full event payload when source.type was empty",
					"eventId", event.ID(),
					"payload", string(rawDataJSON))
			}
		}
		return fmt.Errorf("spec.source.type is required but was empty after parsing event data")
	}

	namespace := data.Metadata.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	// Check if exists
	existing := &lambdav1alpha1.LambdaFunction{}
	err := r.client.Get(ctx, client.ObjectKey{Name: data.Metadata.Name, Namespace: namespace}, existing)

	if err == nil {
		// Update existing
		existing.Spec = data.Spec
		if existing.Labels == nil {
			existing.Labels = make(map[string]string)
		}
		for k, v := range data.Metadata.Labels {
			existing.Labels[k] = v
		}
		if existing.Annotations == nil {
			existing.Annotations = make(map[string]string)
		}
		for k, v := range data.Metadata.Annotations {
			existing.Annotations[k] = v
		}
		existing.Annotations["lambda.knative.io/last-cloudevent-id"] = event.ID()
		if err := r.client.Update(ctx, existing); err != nil {
			return err
		}
		r.log.Info("LambdaFunction updated",
			"name", existing.Name,
			"namespace", existing.Namespace,
			"eventId", event.ID(),
		)
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing lambda: %w", err)
	}

	// Create new
	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Metadata.Name,
			Namespace: namespace,
			Labels:    data.Metadata.Labels,
			Annotations: map[string]string{
				"lambda.knative.io/created-by-cloudevent": "true",
				"lambda.knative.io/cloudevent-id":         event.ID(),
				"lambda.knative.io/cloudevent-source":     event.Source(),
			},
		},
		Spec: data.Spec,
	}
	for k, v := range data.Metadata.Annotations {
		lambda.Annotations[k] = v
	}

	if err := r.client.Create(ctx, lambda); err != nil {
		return err
	}

	r.log.Info("LambdaFunction created",
		"name", lambda.Name,
		"namespace", lambda.Namespace,
		"eventId", event.ID(),
		"sourceType", lambda.Spec.Source.Type,
	)
	return nil
}

func (r *Receiver) handleServiceDelete(ctx context.Context, event *cloudevents.Event) error {
	var data ServiceDeleteData
	if err := event.DataAs(&data); err != nil {
		return fmt.Errorf("failed to parse delete data: %w", err)
	}

	if data.Name == "" {
		data.Name = event.Subject()
	}
	if data.Name == "" {
		return fmt.Errorf("name is required")
	}

	namespace := data.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: namespace,
		},
	}
	return r.client.Delete(ctx, lambda)
}

func (r *Receiver) handleBuildStart(ctx context.Context, event *cloudevents.Event) error {
	var data BuildCommandData
	if err := event.DataAs(&data); err != nil {
		return fmt.Errorf("failed to parse build data: %w", err)
	}

	if data.Name == "" {
		data.Name = event.Subject()
	}
	if data.Name == "" {
		return fmt.Errorf("name is required")
	}

	namespace := data.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	lambda := &lambdav1alpha1.LambdaFunction{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: data.Name, Namespace: namespace}, lambda); err != nil {
		return err
	}

	if lambda.Annotations == nil {
		lambda.Annotations = make(map[string]string)
	}
	lambda.Annotations["lambda.knative.io/rebuild-requested"] = time.Now().Format(time.RFC3339)

	if data.ForceRebuild && lambda.Spec.Build != nil {
		lambda.Spec.Build.ForceRebuild = true
	}

	lambda.Status.Phase = lambdav1alpha1.PhasePending
	lambda.Status.BuildStatus = nil

	if err := r.client.Update(ctx, lambda); err != nil {
		return err
	}
	return r.client.Status().Update(ctx, lambda)
}

func (r *Receiver) handleBuildCancel(ctx context.Context, event *cloudevents.Event) error {
	var data BuildCommandData
	if err := event.DataAs(&data); err != nil {
		return fmt.Errorf("failed to parse cancel data: %w", err)
	}

	if data.Name == "" {
		data.Name = event.Subject()
	}
	if data.Name == "" {
		return fmt.Errorf("name is required")
	}

	namespace := data.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	lambda := &lambdav1alpha1.LambdaFunction{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: data.Name, Namespace: namespace}, lambda); err != nil {
		return err
	}

	if lambda.Status.Phase != lambdav1alpha1.PhaseBuilding {
		return nil // Nothing to cancel
	}

	lambda.Status.Phase = lambdav1alpha1.PhaseFailed
	if lambda.Status.BuildStatus != nil {
		lambda.Status.BuildStatus.Error = fmt.Sprintf("Cancelled: %s", data.Reason)
	}

	return r.client.Status().Update(ctx, lambda)
}

func (r *Receiver) handleFunctionRollback(ctx context.Context, event *cloudevents.Event) error {
	var data struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
		Revision  string `json:"revision,omitempty"`
		Reason    string `json:"reason,omitempty"`
	}
	if err := event.DataAs(&data); err != nil {
		return fmt.Errorf("failed to parse rollback data: %w", err)
	}

	if data.Name == "" {
		data.Name = event.Subject()
	}
	if data.Name == "" {
		return fmt.Errorf("name is required")
	}

	namespace := data.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	lambda := &lambdav1alpha1.LambdaFunction{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: data.Name, Namespace: namespace}, lambda); err != nil {
		return err
	}

	if lambda.Annotations == nil {
		lambda.Annotations = make(map[string]string)
	}
	lambda.Annotations["lambda.knative.io/rollback-requested"] = time.Now().Format(time.RFC3339)
	lambda.Annotations["lambda.knative.io/rollback-reason"] = data.Reason
	if data.Revision != "" {
		lambda.Annotations["lambda.knative.io/rollback-target"] = data.Revision
	}

	return r.client.Update(ctx, lambda)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 RESPONSE EVENT HANDLERS - RED METRICS                              │
// │                                                                         │
// │  These handlers process response events from Lambda runtime wrappers   │
// │  and populate the knative_lambda_function_* Prometheus metrics:        │
// │  - invocations_total (Rate)                                            │
// │  - duration_seconds (Duration)                                         │
// │  - errors_total (Errors)                                               │
// │  - cold_starts_total                                                   │
// └─────────────────────────────────────────────────────────────────────────┘

// ResponseEventData represents data from Lambda runtime wrappers
// This structure matches what the Python/Go/Node.js runtime wrappers emit
type ResponseEventData struct {
	FunctionName  string               `json:"functionName"`
	Namespace     string               `json:"namespace"`
	InvocationID  string               `json:"invocationId,omitempty"`
	CorrelationID string               `json:"correlationId,omitempty"`
	Result        *ResponseResultData  `json:"result,omitempty"`
	Error         *ResponseErrorData   `json:"error,omitempty"`
	Metrics       *ResponseMetricsData `json:"metrics,omitempty"`
}

// ResponseResultData represents successful execution result
type ResponseResultData struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body,omitempty"`
}

// ResponseErrorData represents error details from Lambda
type ResponseErrorData struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Stack     string `json:"stack,omitempty"`
}

// ResponseMetricsData represents runtime metrics from Lambda
type ResponseMetricsData struct {
	DurationMs   int64 `json:"durationMs"`
	ColdStart    bool  `json:"coldStart"`
	MemoryUsedMb int64 `json:"memoryUsedMb,omitempty"`
}

// handleResponseSuccess processes io.knative.lambda.response.success events
// and updates function RED metrics for successful invocations
func (r *Receiver) handleResponseSuccess(ctx context.Context, event *cloudevents.Event) error {
	var data ResponseEventData
	if err := event.DataAs(&data); err != nil {
		r.log.Error(err, "Failed to parse response success data",
			"eventId", event.ID(),
			"eventType", event.Type())
		return nil // Don't fail on parse errors - just skip metrics update
	}

	// Extract function name and namespace
	functionName := data.FunctionName
	namespace := data.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	// Validate we have function name
	if functionName == "" {
		functionName = event.Subject()
	}
	if functionName == "" {
		r.log.V(1).Info("Response event missing function name, skipping metrics",
			"eventId", event.ID())
		return nil
	}

	// Update metrics
	r.recordFunctionMetrics(functionName, namespace, "success", data.Metrics)

	r.log.V(2).Info("Recorded success metrics for function invocation",
		"function", functionName,
		"namespace", namespace,
		"invocationId", data.InvocationID,
		"durationMs", getDurationMs(data.Metrics),
		"coldStart", getColdStart(data.Metrics))

	return nil
}

// handleResponseError processes io.knative.lambda.response.error events
// and updates function RED metrics for failed invocations
func (r *Receiver) handleResponseError(ctx context.Context, event *cloudevents.Event) error {
	var data ResponseEventData
	if err := event.DataAs(&data); err != nil {
		r.log.Error(err, "Failed to parse response error data",
			"eventId", event.ID(),
			"eventType", event.Type())
		return nil // Don't fail on parse errors - just skip metrics update
	}

	// Extract function name and namespace
	functionName := data.FunctionName
	namespace := data.Namespace
	if namespace == "" {
		namespace = r.config.DefaultNamespace
	}

	// Validate we have function name
	if functionName == "" {
		functionName = event.Subject()
	}
	if functionName == "" {
		r.log.V(1).Info("Response event missing function name, skipping metrics",
			"eventId", event.ID())
		return nil
	}

	// Determine error type (use bounded set to avoid high cardinality)
	errorType := normalizeErrorType(data.Error)

	// Update metrics
	r.recordFunctionMetrics(functionName, namespace, "error", data.Metrics)

	// Also record the error counter with error type
	metrics.FunctionErrorsTotal.WithLabelValues(functionName, namespace, errorType).Inc()

	r.log.V(2).Info("Recorded error metrics for function invocation",
		"function", functionName,
		"namespace", namespace,
		"invocationId", data.InvocationID,
		"errorType", errorType,
		"durationMs", getDurationMs(data.Metrics),
		"coldStart", getColdStart(data.Metrics))

	return nil
}

// recordFunctionMetrics updates the common function metrics
func (r *Receiver) recordFunctionMetrics(functionName, namespace, status string, responseMetrics *ResponseMetricsData) {
	// Increment invocation counter
	metrics.FunctionInvocationsTotal.WithLabelValues(functionName, namespace, status).Inc()

	// Record duration if available
	if responseMetrics != nil {
		durationSeconds := float64(responseMetrics.DurationMs) / 1000.0
		metrics.FunctionDuration.WithLabelValues(functionName, namespace).Observe(durationSeconds)

		// Record cold start
		if responseMetrics.ColdStart {
			metrics.FunctionColdStartsTotal.WithLabelValues(functionName, namespace).Inc()
		}
	}
}

// normalizeErrorType converts error codes to a bounded set of values
// to prevent high cardinality in Prometheus metrics
func normalizeErrorType(errData *ResponseErrorData) string {
	if errData == nil || errData.Code == "" {
		return "unknown"
	}

	// Map to bounded set of error types
	code := errData.Code
	switch code {
	case "RuntimeError", "runtime_error":
		return "runtime"
	case "TimeoutError", "timeout", "Timeout":
		return "timeout"
	case "MemoryError", "memory", "OutOfMemory", "OOM":
		return "memory"
	case "HandlerError", "handler_error", "handler":
		return "handler"
	case "ImportError", "ModuleNotFound", "module_not_found":
		return "import"
	case "ValueError", "TypeError", "validation", "ValidationError":
		return "validation"
	case "ConnectionError", "NetworkError", "network":
		return "network"
	case "PermissionError", "AccessDenied", "permission":
		return "permission"
	case "ConfigError", "ConfigurationError", "config":
		return "config"
	case "ReadBodyError":
		return "request"
	default:
		// Check for common patterns
		switch {
		case containsAny(code, "timeout", "Timeout", "TIMEOUT"):
			return "timeout"
		case containsAny(code, "memory", "Memory", "OOM"):
			return "memory"
		case containsAny(code, "import", "Import", "module"):
			return "import"
		case containsAny(code, "network", "Network", "connection"):
			return "network"
		default:
			return "other"
		}
	}
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// getDurationMs safely extracts duration from metrics
func getDurationMs(m *ResponseMetricsData) int64 {
	if m == nil {
		return 0
	}
	return m.DurationMs
}

// getColdStart safely extracts cold start flag from metrics
func getColdStart(m *ResponseMetricsData) bool {
	if m == nil {
		return false
	}
	return m.ColdStart
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 HEALTH & METRICS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func (r *Receiver) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (r *Receiver) readyHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func (r *Receiver) metricsHandler(w http.ResponseWriter, _ *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"events_received":  r.eventsReceived,
		"events_processed": r.eventsProcessed,
		"events_failed":    r.eventsFailed,
	})
}

func (r *Receiver) incrementReceived() {
	r.mu.Lock()
	r.eventsReceived++
	r.mu.Unlock()
}

func (r *Receiver) incrementProcessed() {
	r.mu.Lock()
	r.eventsProcessed++
	r.mu.Unlock()
}

func (r *Receiver) incrementFailed() {
	r.mu.Lock()
	r.eventsFailed++
	r.mu.Unlock()
}

func (r *Receiver) incrementSchemaValidationFailed() {
	r.mu.Lock()
	r.eventsSchemaValidFailed++
	r.mu.Unlock()
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚨 SCHEMA VALIDATION ERROR                                             │
// └─────────────────────────────────────────────────────────────────────────┘

// SchemaValidationError represents a schema validation failure
type SchemaValidationError struct {
	EventID   string `json:"eventId"`
	EventType string `json:"eventType"`
	Message   string `json:"message"`
}

func (e *SchemaValidationError) Error() string {
	return fmt.Sprintf("schema validation failed for event %s (type: %s): %s", e.EventID, e.EventType, e.Message)
}

// IsSchemaValidationError returns true if the error is a schema validation error
func IsSchemaValidationError(err error) bool {
	_, ok := err.(*SchemaValidationError)
	return ok
}
