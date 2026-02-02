"""Prometheus metrics for speech coach agent"""
from prometheus_client import Counter, Histogram, Gauge

# Request metrics
REQUESTS_TOTAL = Counter(
    "speech_coach_requests_total",
    "Total number of requests",
    ["status", "exercise_type"]
)

RESPONSE_DURATION = Histogram(
    "speech_coach_response_duration_seconds",
    "Response duration in seconds",
    ["model"]
)

LLM_INFERENCE_DURATION = Histogram(
    "speech_coach_llm_inference_duration_seconds",
    "LLM inference duration in seconds",
    ["model"]
)

TOKENS_USED = Counter(
    "speech_coach_tokens_used_total",
    "Total tokens used",
    ["model", "type"]
)

# Exercise metrics
EXERCISES_COMPLETED = Counter(
    "speech_coach_exercises_completed_total",
    "Total exercises completed",
    ["exercise_type", "difficulty"]
)

GAME_SESSIONS_TOTAL = Counter(
    "speech_coach_sessions_total",
    "Total game sessions",
    ["status"]
)

# Progress metrics
USER_PROGRESS = Gauge(
    "speech_coach_user_progress_points",
    "User progress points",
    ["user_id"]
)

# Build info
BUILD_INFO = Gauge(
    "speech_coach_build_info",
    "Build information",
    ["version", "commit"]
)


def init_build_info(version: str, commit: str):
    """Initialize build info metric"""
    BUILD_INFO.labels(version=version, commit=commit).set(1)
