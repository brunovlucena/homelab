package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitOTel initializes OpenTelemetry tracing and returns a shutdown function
// Sends traces to Alloy → Alloy forwards to Logfire + Tempo
// Metrics are now handled natively by Prometheus (see metrics package)
func InitOTel(ctx context.Context) (func(context.Context) error, error) {
	// Alloy OTLP endpoint
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "alloy.alloy.svc.cluster.local:4317"
	}

	log.Printf("📊 OpenTelemetry → Alloy: %s", endpoint)

	// Resource with service info
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName("homepage"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// ═══════════════════════════════════════════════════════════════════════
	// 🔍 TRACES: OTLP trace exporter → Alloy
	// ═══════════════════════════════════════════════════════════════════════
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Println("✅ OpenTelemetry Tracing → Alloy → Logfire + Tempo")

	// Return shutdown function
	return tp.Shutdown, nil
}

// PrometheusHandler returns the HTTP handler for the Prometheus metrics endpoint
// Uses native Prometheus metrics from the metrics package
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
