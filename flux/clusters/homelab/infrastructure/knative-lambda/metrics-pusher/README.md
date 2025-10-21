# Metrics Pusher

A sidecar service that collects and pushes metrics to Prometheus via remote write.

## Purpose

The metrics pusher is responsible for:
- Running as a sidecar alongside the knative-lambda builder service
- Collecting system metrics from the metrics pusher itself
- Pushing metrics to Prometheus via remote write
- Providing health check endpoints for Kubernetes

## Features

- **System Metrics Collection**: Collects basic system metrics (goroutines, memory usage)
- **Remote Write**: Pushes metrics to Prometheus via HTTP POST
- **Health Checks**: Provides `/health` and `/ready` endpoints
- **Structured Logging**: JSON logging with configurable log levels
- **Graceful Shutdown**: Handles shutdown signals properly

## Configuration

The service is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PROMETHEUS_REMOTE_WRITE_URL` | Prometheus remote write URL | Required |
| `PUSH_INTERVAL` | How often to push metrics | `30s` |
| `TIMEOUT` | HTTP timeout for requests | `30s` |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `NAMESPACE` | Kubernetes namespace | `knative-lambda` |

## Metrics

The service exposes the following metrics:

### System Metrics
- `knative_lambda_metric_pusher_system_goroutines` - Number of goroutines
- `knative_lambda_metric_pusher_system_memory_alloc_bytes` - Allocated memory
- `knative_lambda_metric_pusher_system_memory_total_alloc_bytes` - Total allocated memory
- `knative_lambda_metric_pusher_system_memory_sys_bytes` - System memory

### Service Metrics
- `knative_lambda_metric_pusher_heartbeat` - Service heartbeat timestamp
- `knative_lambda_metric_pusher_uptime_seconds` - Service uptime

## Deployment

The service is deployed as a sidecar container alongside the knative-lambda builder service with:

- Resource limits and requests
- Automatic restart with the builder pod
- No external port exposure (sidecar mode)

## Health Endpoints

**Note**: Health endpoints are disabled in sidecar mode. The metrics pusher runs as a sidecar container and doesn't expose HTTP endpoints. Health checks are handled by the main builder container.

## Building

```bash
# Build the Docker image
docker build -t knative-lambda-metrics-pusher .

# Run locally
docker run -e PROMETHEUS_REMOTE_WRITE_URL=http://localhost:9090/api/v1/write knative-lambda-metrics-pusher

# Using Makefile (from parent directory)
make docker-build-metrics-pusher
make metrics-pusher  # Build and push
```

## Architecture

The metrics pusher is designed to be simple and focused:

1. **Sidecar Pattern**: Runs alongside the builder service in the same pod
2. **Main Loop**: Runs metric collection and pushing at configurable intervals
3. **HTTP Server**: Serves health checks and metrics endpoint
4. **Graceful Shutdown**: Handles SIGTERM/SIGINT signals
5. **Error Handling**: Logs errors but continues operation

The service doesn't scrape metrics from other services - it only collects its own system metrics and pushes them to Prometheus. As a sidecar, it automatically scales with the builder service and shares the same lifecycle. 