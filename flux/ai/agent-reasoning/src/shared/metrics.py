"""
Prometheus metrics for Agent-Reasoning service.
"""
from prometheus_client import Counter, Histogram, Gauge
import os


# Request metrics
REASONING_REQUESTS = Counter(
    "agent_reasoning_requests_total",
    "Total reasoning requests",
    ["task_type", "status"]
)

REASONING_DURATION = Histogram(
    "agent_reasoning_duration_seconds",
    "Reasoning task duration",
    ["task_type"],
    buckets=[0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0]
)

REASONING_STEPS = Histogram(
    "agent_reasoning_steps",
    "Number of reasoning steps used",
    ["task_type"],
    buckets=[1, 2, 3, 4, 5, 6, 8, 10, 12, 15, 20]
)

REASONING_CONFIDENCE = Histogram(
    "agent_reasoning_confidence",
    "Confidence score of reasoning result",
    ["task_type"],
    buckets=[0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
)

# System metrics
GPU_UTILIZATION = Gauge(
    "agent_reasoning_gpu_utilization",
    "GPU utilization percentage",
    ["gpu_id"]
)

GPU_MEMORY_USED = Gauge(
    "agent_reasoning_gpu_memory_used_bytes",
    "GPU memory used in bytes",
    ["gpu_id"]
)

MODEL_LOADED = Gauge(
    "agent_reasoning_model_loaded",
    "Whether the TRM model is loaded (1=loaded, 0=not loaded)"
)


def init_build_info(version: str, commit: str):
    """Initialize build info as a gauge."""
    BUILD_INFO = Gauge(
        "agent_reasoning_build_info",
        "Build information",
        ["version", "commit"]
    )
    BUILD_INFO.labels(version=version, commit=commit).set(1)


def init_metrics():
    """Initialize metrics with default values."""
    MODEL_LOADED.set(0)

