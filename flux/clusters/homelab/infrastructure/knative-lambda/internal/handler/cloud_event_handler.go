// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🌐 CLOUD EVENT HANDLER - Focused CloudEvent HTTP handling
//
//	🎯 Purpose: Handle CloudEvent HTTP requests and process them
//	💡 Features: CloudEvent parsing, validation, processing, response generation
//
//	🏛️ ARCHITECTURE:
//	🌐 HTTP Request Handling - Parse and validate incoming HTTP requests
//	📥 CloudEvent Processing - Process CloudEvents through the event handler
//	📊 Response Generation - Generate appropriate HTTP responses
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"knative-lambda-new/internal/observability"
	"knative-lambda-new/pkg/builds"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// 🌐 CloudEventHandlerImpl - "Focused CloudEvent HTTP handling"
type CloudEventHandlerImpl struct {
	obs       *observability.Observability
	container ComponentContainer
}

// 🌐 CloudEventHandlerConfig - "Configuration for creating CloudEvent handler"
type CloudEventHandlerConfig struct {
	Observability *observability.Observability
	Container     ComponentContainer
}

// 🏗️ NewCloudEventHandler - "Create new CloudEvent handler with dependencies"
func NewCloudEventHandler(config CloudEventHandlerConfig) (CloudEventHandler, error) {
	if config.Observability == nil {
		return nil, fmt.Errorf("observability cannot be nil")
	}

	if config.Container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}

	return &CloudEventHandlerImpl{
		obs:       config.Observability,
		container: config.Container,
	}, nil
}

// HandleCloudEvent handles CloudEvent ingestion with comprehensive tracing
func (h *CloudEventHandlerImpl) HandleCloudEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Extract trace context from request headers
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

	// Create span for the entire CloudEvent processing
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "handle_cloud_event", map[string]string{
		"http.method": r.Method,
		"http.path":   r.URL.Path,
		"endpoint":    "cloud_event",
	})
	defer span.End()

	// Add correlation ID to response headers for trace propagation
	correlationID := observability.GetCorrelationID(ctx)
	if correlationID != "" {
		w.Header().Set("X-Correlation-ID", correlationID)
	}

	// Set trace headers for propagation
	w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())
	w.Header().Set("X-Span-ID", span.SpanContext().SpanID().String())

	h.obs.Info(ctx, "Processing CloudEvent request",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"content_type", r.Header.Get("Content-Type"))

	// Parse CloudEvent with tracing
	ctx, parseSpan := h.obs.StartSpan(ctx, "parse_cloud_event")
	event, err := h.parseCloudEvent(r)
	if err != nil {
		parseSpan.End()
		span.SetStatus(codes.Error, fmt.Sprintf("Failed to parse CloudEvent: %v", err))
		h.obs.Error(ctx, err, "Failed to parse CloudEvent",
			"method", r.Method,
			"path", r.URL.Path,
			"error_details", err.Error())
		h.sendErrorResponse(w, "Failed to parse CloudEvent", err, http.StatusBadRequest)
		return
	}
	parseSpan.SetAttributes(
		attribute.String("event.type", event.Type()),
		attribute.String("event.source", event.Source()),
		attribute.String("event.id", event.ID()),
		attribute.String("event.subject", event.Subject()),
	)
	parseSpan.End()

	h.obs.Info(ctx, "CloudEvent parsed successfully",
		"event_type", event.Type(),
		"event_source", event.Source(),
		"event_id", event.ID(),
		"event_subject", event.Subject())

	// Get event handler from container with tracing
	ctx, handlerSpan := h.obs.StartSpan(ctx, "get_event_handler")
	eventHandler := h.container.GetEventHandler()
	if eventHandler == nil {
		handlerSpan.End()
		span.SetStatus(codes.Error, "Event handler not available")
		h.obs.Error(ctx, fmt.Errorf("event handler not available"), "Event handler not available")
		h.sendErrorResponse(w, "Event handler not available", fmt.Errorf("event handler not available"), http.StatusServiceUnavailable)
		return
	}
	handlerSpan.End()

	h.obs.Info(ctx, "Event handler retrieved successfully",
		"event_type", event.Type(),
		"event_source", event.Source(),
		"event_id", event.ID())

	// Process the CloudEvent with tracing
	ctx, processSpan := h.obs.StartSpan(ctx, "process_cloud_event")
	h.obs.Info(ctx, "Processing CloudEvent",
		"event_type", event.Type(),
		"event_source", event.Source(),
		"event_id", event.ID())

	response, err := eventHandler.ProcessCloudEvent(ctx, event)
	if err != nil {
		processSpan.End()
		span.SetStatus(codes.Error, fmt.Sprintf("Failed to process CloudEvent: %v", err))
		h.obs.Error(ctx, err, "Failed to process CloudEvent",
			"event_type", event.Type(),
			"event_source", event.Source(),
			"event_id", event.ID(),
			"error_details", err.Error())
		h.sendErrorResponse(w, "Failed to process CloudEvent", err, http.StatusInternalServerError)
		return
	}
	processSpan.SetAttributes(
		attribute.String("response.status", response.Status),
		attribute.String("response.message", response.Message),
		attribute.String("response.correlation_id", response.CorrelationID),
	)
	processSpan.End()

	// Record processing duration
	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("processing.duration_ms", float64(duration.Milliseconds())))

	h.obs.Info(ctx, "CloudEvent processed successfully",
		"event_type", event.Type(),
		"event_source", event.Source(),
		"event_id", event.ID(),
		"response_status", response.Status,
		"response_message", response.Message,
		"processing_duration_ms", duration.Milliseconds())

	// Send success response
	h.sendSuccessResponse(w, response)
}

// parseCloudEvent parses a CloudEvent from an HTTP request
func (h *CloudEventHandlerImpl) parseCloudEvent(r *http.Request) (*cloudevents.Event, error) {
	// Create a new CloudEvent
	event := cloudevents.NewEvent()

	// Set basic CloudEvent attributes from headers
	if specVersion := r.Header.Get("Ce-Specversion"); specVersion != "" {
		event.SetSpecVersion(specVersion)
	} else {
		event.SetSpecVersion(cloudevents.VersionV1)
	}

	if eventID := r.Header.Get("Ce-Id"); eventID != "" {
		event.SetID(eventID)
	}

	if eventType := r.Header.Get("Ce-Type"); eventType != "" {
		event.SetType(eventType)
	}

	if eventSource := r.Header.Get("Ce-Source"); eventSource != "" {
		event.SetSource(eventSource)
	}

	if eventTime := r.Header.Get("Ce-Time"); eventTime != "" {
		if t, err := time.Parse(time.RFC3339, eventTime); err == nil {
			event.SetTime(t)
		}
	}

	if subject := r.Header.Get("Ce-Subject"); subject != "" {
		event.SetSubject(subject)
	}

	// Parse the request body as JSON data
	var data interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}

	// Set the data
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	return &event, nil
}

// sendSuccessResponse sends a successful response
func (h *CloudEventHandlerImpl) sendSuccessResponse(w http.ResponseWriter, response *builds.HandlerResponse) {
	// Set timestamp if not already set
	if response.Timestamp.IsZero() {
		response.Timestamp = time.Now().UTC()
	}

	// Set status code based on response status
	statusCode := http.StatusOK
	switch response.Status {
	case "started", "service_created", "acknowledged", "processed":
		statusCode = http.StatusOK
	case "running", "skipped":
		statusCode = http.StatusAccepted
	default:
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the response, we can't send an error response either
		// Just log the error
		fmt.Printf("Failed to encode success response: %v\n", err)
	}
}

// sendErrorResponse sends an error response
func (h *CloudEventHandlerImpl) sendErrorResponse(w http.ResponseWriter, message string, err error, statusCode int) {
	errorResponse := map[string]interface{}{
		"error":     message,
		"details":   err.Error(),
		"timestamp": time.Now().UTC(),
		"status":    "error",
	}

	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		// If we can't encode the error response, just write a simple error
		http.Error(w, message, statusCode)
	}
}
