"""
🔬 Analysis algorithms
Implements error pattern detection and slow request analysis
"""

import logging
import re
from collections import Counter, defaultdict
from datetime import datetime
from typing import Any, Dict, List

logger = logging.getLogger(__name__)


class ErrorPatternAnalyzer:
    """Analyzes logs to find elevated error patterns"""

    def __init__(self, baseline_window: int = 24):
        """
        Initialize analyzer

        Args:
            baseline_window: Hours to use for baseline comparison (default: 24)
        """
        self.baseline_window = baseline_window

    def analyze(
        self,
        current_logs: List[Dict[str, Any]],
        baseline_logs: List[Dict[str, Any]],
        threshold_multiplier: float = 2.0,
    ) -> Dict[str, Any]:
        """
        Analyze logs for elevated error patterns

        Args:
            current_logs: Logs from the investigation period
            baseline_logs: Logs from the baseline period
            threshold_multiplier: Multiplier for determining elevated patterns (default: 2.0)

        Returns:
            Analysis results with detected patterns
        """
        logger.info(f"🔬 Analyzing {len(current_logs)} current logs vs {len(baseline_logs)} baseline logs")

        # Extract log patterns
        current_patterns = self._extract_patterns(current_logs)
        baseline_patterns = self._extract_patterns(baseline_logs)

        # Find elevated patterns
        elevated_patterns = []
        for pattern, current_count in current_patterns.items():
            baseline_count = baseline_patterns.get(pattern, 0)

            # Calculate rate per hour
            current_rate = current_count
            baseline_rate = baseline_count / self.baseline_window if baseline_count > 0 else 0

            # Check if elevated
            if current_rate > baseline_rate * threshold_multiplier and current_count > 5:
                elevation_factor = current_rate / baseline_rate if baseline_rate > 0 else float("inf")
                elevated_patterns.append(
                    {
                        "pattern": pattern,
                        "current_count": current_count,
                        "baseline_count": baseline_count,
                        "current_rate": round(current_rate, 2),
                        "baseline_rate": round(baseline_rate, 2),
                        "elevation_factor": round(elevation_factor, 2),
                        "severity": self._determine_severity(elevation_factor, current_count),
                    }
                )

        # Sort by severity and elevation factor
        elevated_patterns.sort(key=lambda x: (x["severity"], x["elevation_factor"]), reverse=True)

        return {
            "status": "completed",
            "total_current_logs": len(current_logs),
            "total_baseline_logs": len(baseline_logs),
            "unique_current_patterns": len(current_patterns),
            "unique_baseline_patterns": len(baseline_patterns),
            "elevated_patterns_count": len(elevated_patterns),
            "elevated_patterns": elevated_patterns[:10],  # Top 10
            "analysis_timestamp": datetime.utcnow().isoformat(),
        }

    def _extract_patterns(self, logs: List[Dict[str, Any]]) -> Counter:
        """Extract error patterns from logs"""
        patterns = Counter()

        for log in logs:
            # Extract log line
            line = log.get("line", "") if isinstance(log, dict) else str(log)

            # Extract error patterns
            pattern = self._normalize_log_line(line)
            if pattern and self._is_error_pattern(line):
                patterns[pattern] += 1

        return patterns

    def _normalize_log_line(self, line: str) -> str:
        """Normalize log line to extract pattern"""
        # Remove timestamps
        line = re.sub(r"\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(\.\d+)?([+-]\d{2}:?\d{2})?", "", line)

        # Remove IPs
        line = re.sub(r"\b(?:\d{1,3}\.){3}\d{1,3}\b", "IP", line)

        # Remove UUIDs
        line = re.sub(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "UUID", line)

        # Remove numbers (but keep error codes)
        line = re.sub(r"(?<![A-Z])\b\d+\b(?![A-Z])", "N", line)

        # Remove quoted strings
        line = re.sub(r'"[^"]*"', "STRING", line)
        line = re.sub(r"'[^']*'", "STRING", line)

        # Normalize whitespace
        line = " ".join(line.split())

        return line.strip()

    def _is_error_pattern(self, line: str) -> bool:
        """Check if log line contains error indicators"""
        error_keywords = [
            "error",
            "exception",
            "failed",
            "failure",
            "fatal",
            "panic",
            "critical",
            "emergency",
            "alert",
            "denied",
            "timeout",
            "refused",
            "unreachable",
            "unavailable",
        ]

        line_lower = line.lower()
        return any(keyword in line_lower for keyword in error_keywords)

    def _determine_severity(self, elevation_factor: float, count: int) -> str:
        """Determine severity based on elevation factor and count"""
        if elevation_factor >= 10 or count >= 100:
            return "critical"
        elif elevation_factor >= 5 or count >= 50:
            return "high"
        elif elevation_factor >= 3 or count >= 20:
            return "medium"
        else:
            return "low"


class SlowRequestAnalyzer:
    """Analyzes traces to find slow requests"""

    def __init__(self, baseline_window: int = 24):
        """
        Initialize analyzer

        Args:
            baseline_window: Hours to use for baseline comparison (default: 24)
        """
        self.baseline_window = baseline_window

    def analyze(
        self,
        current_traces: List[Dict[str, Any]],
        baseline_traces: List[Dict[str, Any]],
        percentile: float = 95.0,
    ) -> Dict[str, Any]:
        """
        Analyze traces for slow requests

        Args:
            current_traces: Traces from the investigation period
            baseline_traces: Traces from the baseline period
            percentile: Percentile to use for slow request threshold (default: 95.0)

        Returns:
            Analysis results with slow requests
        """
        logger.info(f"🔬 Analyzing {len(current_traces)} current traces vs {len(baseline_traces)} baseline traces")

        # Extract durations by service/operation
        current_durations = self._extract_durations_by_operation(current_traces)
        baseline_durations = self._extract_durations_by_operation(baseline_traces)

        # Find slow operations
        slow_operations = []
        for operation, current_durs in current_durations.items():
            baseline_durs = baseline_durations.get(operation, [])

            if not current_durs:
                continue

            # Calculate percentiles
            current_p95 = self._percentile(current_durs, percentile)
            baseline_p95 = self._percentile(baseline_durs, percentile) if baseline_durs else 0

            # Check if slow
            if baseline_p95 > 0 and current_p95 > baseline_p95 * 1.5:  # 50% slower
                slowdown_factor = current_p95 / baseline_p95
                slow_operations.append(
                    {
                        "operation": operation,
                        "current_p95_ms": round(current_p95, 2),
                        "baseline_p95_ms": round(baseline_p95, 2),
                        "slowdown_factor": round(slowdown_factor, 2),
                        "current_request_count": len(current_durs),
                        "baseline_request_count": len(baseline_durs),
                        "severity": self._determine_severity(slowdown_factor, current_p95),
                    }
                )

        # Sort by severity and slowdown factor
        slow_operations.sort(key=lambda x: (x["severity"], x["slowdown_factor"]), reverse=True)

        return {
            "status": "completed",
            "total_current_traces": len(current_traces),
            "total_baseline_traces": len(baseline_traces),
            "unique_operations": len(current_durations),
            "slow_operations_count": len(slow_operations),
            "slow_operations": slow_operations[:10],  # Top 10
            "analysis_timestamp": datetime.utcnow().isoformat(),
        }

    def _extract_durations_by_operation(self, traces: List[Dict[str, Any]]) -> Dict[str, List[float]]:
        """Extract durations grouped by service/operation"""
        durations = defaultdict(list)

        for trace in traces:
            # Extract operation name and duration
            if isinstance(trace, dict):
                # Handle different trace formats
                operation = trace.get("spanName") or trace.get("operationName") or trace.get("name", "unknown")
                duration = trace.get("duration") or trace.get("durationMs", 0)

                # Convert to milliseconds if needed
                if duration > 1000000:  # Likely nanoseconds
                    duration = duration / 1000000
                elif duration > 1000:  # Likely microseconds
                    duration = duration / 1000

                durations[operation].append(duration)

        return dict(durations)

    def _percentile(self, values: List[float], percentile: float) -> float:
        """Calculate percentile of values"""
        if not values:
            return 0.0

        sorted_values = sorted(values)
        index = int(len(sorted_values) * (percentile / 100))
        index = min(index, len(sorted_values) - 1)
        return sorted_values[index]

    def _determine_severity(self, slowdown_factor: float, duration_ms: float) -> str:
        """Determine severity based on slowdown factor and duration"""
        if slowdown_factor >= 5 or duration_ms >= 5000:
            return "critical"
        elif slowdown_factor >= 3 or duration_ms >= 2000:
            return "high"
        elif slowdown_factor >= 2 or duration_ms >= 1000:
            return "medium"
        else:
            return "low"
