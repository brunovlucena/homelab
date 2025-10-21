// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔍 KNATIVE LAMBDA SIDECAR - BUILD MONITOR & EVENT PUBLISHER
//
//	🎯 Purpose: Monitor Kaniko build process and publish completion events
//	💡 Features: CloudEvents HTTP integration to Knative broker
//
//	🏛️ ARCHITECTURE:
//	🔍 Build Monitoring - Monitor Kaniko pod status and logs
//	📨 Event Publishing - Publish CloudEvents to Knative broker
//	⏱️ Timeout Handling - Handle build timeouts and failures
//	🔄 Graceful Shutdown - Clean shutdown on termination signals
//
//	🔧 COMPONENTS:
//	📊 Build Monitor - Monitors Kaniko build process
//	📨 Event Publisher - Publishes CloudEvents to broker
//	⚙️ Configuration - Sidecar-specific configuration
//	📝 Logging - Structured logging for observability
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

	"knative-lambda-new/sidecar/internal/config"
	"knative-lambda-new/sidecar/internal/monitor"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 MAIN FUNCTION - "Sidecar application entry point"                   │
// └─────────────────────────────────────────────────────────────────────────┘

// main is the entry point for the Knative Lambda sidecar
// This function orchestrates the build monitoring process and handles
// graceful shutdown when the sidecar receives termination signals
func main() {
	// Load sidecar-specific configuration from environment variables
	// This includes Kaniko pod details, broker URL, and monitoring settings
	cfg, err := config.LoadSidecarConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup structured logging with configurable log level
	// This provides consistent logging format for observability
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

	// Log sidecar startup with key configuration details
	slog.Info("Starting Kaniko build monitor sidecar",
		"namespace", cfg.KanikoNamespace, // Kubernetes namespace containing Kaniko pod
		"pod", cfg.KanikoPodName, // Name of the Kaniko pod to monitor
		"container", cfg.KanikoContainerName, // Name of the Kaniko container
		"job_name", cfg.JobName, // Associated Kubernetes job name
		"broker_url", cfg.BrokerURL) // Knative broker URL for event publishing

	// Create context for graceful shutdown handling
	// This context will be cancelled when shutdown signals are received
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled when main exits

	// Setup signal handling for graceful shutdown
	// Listen for SIGINT (Ctrl+C) and SIGTERM (termination signal)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create build monitor configuration with all necessary parameters
	// This configuration is passed to the build monitor for initialization
	monitorConfig := monitor.BuildMonitorConfig{
		KanikoNamespace:     cfg.KanikoNamespace,     // Namespace containing Kaniko pod
		KanikoPodName:       cfg.KanikoPodName,       // Name of Kaniko pod to monitor
		KanikoContainerName: cfg.KanikoContainerName, // Name of Kaniko container
		PollInterval:        cfg.PollInterval,        // How often to check pod status
		BuildTimeout:        cfg.BuildTimeout,        // Maximum time to wait for build completion
		JobName:             cfg.JobName,             // Associated Kubernetes job name
		ImageURI:            cfg.ImageURI,            // Expected image URI after successful build
		ThirdPartyID:        cfg.ThirdPartyID,        // Third party identifier for event correlation
		ParserID:            cfg.ParserID,            // Parser identifier for event correlation
		ContentHash:         cfg.ContentHash,         // Content hash for unique image tagging
		CorrelationID:       cfg.CorrelationID,       // Correlation ID for request tracing
		BrokerURL:           cfg.BrokerURL,           // Knative broker URL for event publishing
	}

	// Create and initialize the build monitor
	// This component will handle the actual monitoring of the Kaniko build process
	buildMonitor, err := monitor.NewBuildMonitor(monitorConfig)
	if err != nil {
		slog.Error("Failed to create build monitor", "error", err)
		os.Exit(1)
	}
	defer buildMonitor.Close() // Ensure build monitor is cleaned up on exit

	// Start build monitoring in a separate goroutine
	// This allows the main goroutine to handle shutdown signals
	done := make(chan error, 1)
	go func() {
		defer close(done)
		// Start monitoring the Kaniko build process
		// This will block until the build completes or fails
		err := buildMonitor.MonitorBuild(ctx)
		if err != nil {
			slog.Error("Build monitoring failed", "error", err)
		}
		done <- err
	}()

	// Wait for either build completion or shutdown signal
	// This implements graceful shutdown handling
	select {
	case err := <-done:
		// Build monitoring completed (successfully or with error)
		if err != nil {
			slog.Error("Build monitoring completed with error", "error", err)
			os.Exit(1)
		}
		slog.Info("Build monitoring completed successfully")

	case sig := <-sigChan:
		// Received shutdown signal
		slog.Info("Received shutdown signal", "signal", sig.String())

		// Cancel context to signal build monitor to stop
		cancel()

		// Wait for build monitor to complete shutdown
		select {
		case err := <-done:
			if err != nil {
				slog.Error("Build monitoring failed during shutdown", "error", err)
				os.Exit(1)
			}
			slog.Info("Build monitoring shutdown completed successfully")

		case <-time.After(30 * time.Second):
			// Force shutdown after 30 seconds
			slog.Warn("Build monitoring shutdown timed out, forcing exit")
			os.Exit(1)
		}
	}
}
