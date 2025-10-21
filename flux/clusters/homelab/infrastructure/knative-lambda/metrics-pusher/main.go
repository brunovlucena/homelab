// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📊 METRIC PUSHER - Prometheus Remote Write Service
//
//	🎯 Purpose: Collect and push metrics to Prometheus via remote write
//	💡 Features: Metric collection, remote write, structured metrics
//
//	🏛️ ARCHITECTURE:
//	📊 Metric Collection - Collect metrics from builder and jobs
//	📤 Remote Write - Push metrics to Prometheus via HTTP
//	⏱️ Timing Metrics - Track build duration and performance
//	🔄 Periodic Pushing - Push metrics at configurable intervals
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 MAIN FUNCTION - "Metric pusher application entry point"             │
// └─────────────────────────────────────────────────────────────────────────┘

// main is the entry point for the metric pusher service
// This function orchestrates metric collection and remote write to Prometheus
func main() {
	// Load configuration from environment variables
	cfg, err := LoadMetricPusherConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup structured logging with configurable log level
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create JSON logger for structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Log metric pusher startup with key configuration details
	slog.Info("Starting metric pusher service",
		"remote_write_url", cfg.RemoteWriteURL,
		"push_interval", cfg.PushInterval,
		"timeout", cfg.Timeout)

	// Create context for graceful shutdown handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create metric pusher
	metricPusher, err := NewMetricPusher(PusherConfig{
		RemoteWriteURL:   cfg.RemoteWriteURL,
		Timeout:          cfg.Timeout,
		Logger:           logger,
		FailureTolerance: cfg.FailureTolerance,
		Enabled:          cfg.Enabled,
	})
	if err != nil {
		slog.Error("Failed to create metric pusher", "error", err)
		os.Exit(1)
	}

	// Note: HTTP server disabled for sidecar mode
	// The metrics pusher runs as a sidecar and doesn't need to expose health check endpoints
	// Health checks are handled by the main builder container

	// Start metric collection and pushing in a separate goroutine
	done := make(chan error, 1)
	go func() {
		defer close(done)
		err := metricPusher.Start(ctx, cfg.PushInterval)
		if err != nil {
			slog.Error("Metric pusher failed", "error", err)
		}
		done <- err
	}()

	// Wait for either completion or shutdown signal
	select {
	case err := <-done:
		// Metric pusher completed (successfully or with error)
		if err != nil {
			slog.Error("Metric pusher completed with error", "error", err)
			os.Exit(1)
		}
		slog.Info("Metric pusher completed successfully")

	case sig := <-sigChan:
		// Received shutdown signal
		slog.Info("Received shutdown signal", "signal", sig.String())

		// Cancel context to signal metric pusher to stop
		cancel()

		// Wait for metric pusher to complete shutdown
		select {
		case err := <-done:
			if err != nil {
				slog.Error("Metric pusher failed during shutdown", "error", err)
				os.Exit(1)
			}
			slog.Info("Metric pusher shutdown completed successfully")

		case <-time.After(30 * time.Second):
			// Force shutdown after 30 seconds
			slog.Warn("Metric pusher shutdown timed out, forcing exit")
			os.Exit(1)
		}
	}
}
