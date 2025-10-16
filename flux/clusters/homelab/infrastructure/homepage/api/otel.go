package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var prometheusHandler http.Handler

// InitOTel initializes OpenTelemetry and returns a shutdown function
// Sends traces to Alloy → Alloy forwards to Logfire + Tempo
// Exports metrics in Prometheus format for scraping
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
			semconv.ServiceName("bruno-site"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// ═══════════════════════════════════════════════════════════════════════
	// 📊 METRICS: Prometheus exporter for /metrics endpoint
	// ═══════════════════════════════════════════════════════════════════════
	var err2 error
	prometheusExporter, err2 := otelprometheus.New()
	if err2 != nil {
		log.Printf("⚠️  WARNING: Failed to create Prometheus exporter: %v", err2)
		prometheusHandler = promhttp.Handler() // Fallback to default Prometheus handler
	} else {
		// 📊 Configure explicit histogram buckets for API latency
		// Using standard Prometheus histogram buckets optimized for API response times
		histogramView := sdkmetric.NewView(
			sdkmetric.Instrument{
				Kind: sdkmetric.InstrumentKindHistogram,
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: []float64{
						0.005, // 5ms
						0.01,  // 10ms
						0.025, // 25ms
						0.05,  // 50ms
						0.1,   // 100ms
						0.25,  // 250ms
						0.5,   // 500ms
						1,     // 1s
						2.5,   // 2.5s
						5,     // 5s
						10,    // 10s
					},
				},
			},
		)

		// Create meter provider with Prometheus exporter and histogram views
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(prometheusExporter),
			sdkmetric.WithView(histogramView),
		)
		otel.SetMeterProvider(mp)

		// Use the OpenTelemetry Prometheus exporter as the HTTP handler
		// Native histograms are disabled at Prometheus level, so standard text format works
		// Disable compression to avoid gzip parsing issues
		prometheusHandler = promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				EnableOpenMetrics: false, // Disable OpenMetrics format, use classic Prometheus text format
				DisableCompression: true,  // Disable gzip compression
			},
		)

		log.Println("✅ OpenTelemetry Metrics → Prometheus exporter initialized (with explicit histogram buckets)")
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

	log.Println("✅ OpenTelemetry Tracing → Alloy → Logfire")

	// Return shutdown function
	return tp.Shutdown, nil
}

// PrometheusHandler returns the HTTP handler for the Prometheus metrics endpoint
func PrometheusHandler() http.Handler {
	if prometheusHandler != nil {
		return prometheusHandler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Prometheus exporter not initialized"))
	})
}
