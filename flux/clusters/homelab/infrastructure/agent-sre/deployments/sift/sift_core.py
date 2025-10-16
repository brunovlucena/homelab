"""
🧠 Sift Core
Main orchestration layer for Sift investigations
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional

from .analyzers import ErrorPatternAnalyzer, SlowRequestAnalyzer
from .investigation import Analysis, AnalysisType, Investigation, InvestigationStatus
from .loki_client import LokiClient
from .storage import InvestigationStorage
from .tempo_client import TempoClient

logger = logging.getLogger(__name__)


class SiftCore:
    """Core orchestration for Sift investigations"""

    def __init__(
        self,
        loki_url: str,
        tempo_url: str,
        storage_path: str = "/tmp/sift_investigations.db",
    ):
        """Initialize Sift core"""
        self.loki_client = LokiClient(loki_url)
        self.tempo_client = TempoClient(tempo_url)
        self.storage = InvestigationStorage(storage_path)
        self.error_analyzer = ErrorPatternAnalyzer()
        self.slow_request_analyzer = SlowRequestAnalyzer()
        logger.info("🧠 Initialized Sift Core")

    async def create_investigation(
        self,
        name: str,
        labels: Dict[str, str],
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
    ) -> Investigation:
        """
        Create a new investigation

        Args:
            name: Investigation name
            labels: Labels to scope the investigation (e.g., {"cluster": "prod", "namespace": "api"})
            start_time: Start time (defaults to 30 minutes ago)
            end_time: End time (defaults to now)

        Returns:
            Created investigation
        """
        if not start_time:
            start_time = datetime.utcnow() - timedelta(minutes=30)
        if not end_time:
            end_time = datetime.utcnow()

        investigation = Investigation(
            name=name,
            labels=labels,
            start_time=start_time,
            end_time=end_time,
            status=InvestigationStatus.PENDING,
        )

        self.storage.save_investigation(investigation)
        logger.info(f"🆕 Created investigation: {investigation.id} - {name}")

        return investigation

    async def run_error_pattern_analysis(
        self,
        investigation_id: str,
        log_query: Optional[str] = None,
    ) -> Analysis:
        """
        Run error pattern analysis for an investigation

        Args:
            investigation_id: Investigation ID
            log_query: Optional LogQL query (will be built from labels if not provided)

        Returns:
            Analysis result
        """
        investigation = self.storage.get_investigation(investigation_id)
        if not investigation:
            raise ValueError(f"Investigation {investigation_id} not found")

        # Create analysis
        analysis = Analysis(
            type=AnalysisType.ERROR_PATTERN,
            status=InvestigationStatus.RUNNING,
        )
        investigation.add_analysis(analysis)
        investigation.update_status(InvestigationStatus.RUNNING)
        self.storage.save_investigation(investigation)

        try:
            # Build log query from labels if not provided
            if not log_query:
                log_query = self._build_log_query(investigation.labels)

            logger.info(f"🔬 Running error pattern analysis for investigation {investigation_id}")

            # Query current period logs
            current_result = await self.loki_client.query_range(
                query=log_query,
                start=investigation.start_time,
                end=investigation.end_time,
                limit=1000,
            )

            # Query baseline period logs (24 hours before current period)
            baseline_start = investigation.start_time - timedelta(hours=24)
            baseline_end = investigation.start_time
            baseline_result = await self.loki_client.query_range(
                query=log_query,
                start=baseline_start,
                end=baseline_end,
                limit=1000,
            )

            # Extract log lines
            current_logs = self._extract_log_lines(current_result)
            baseline_logs = self._extract_log_lines(baseline_result)

            # Run analysis
            analysis_result = self.error_analyzer.analyze(current_logs, baseline_logs)

            # Update analysis
            analysis.status = InvestigationStatus.COMPLETED
            analysis.end_time = datetime.utcnow()
            analysis.result = analysis_result
            analysis.metadata = {
                "log_query": log_query,
                "current_period": f"{investigation.start_time} to {investigation.end_time}",
                "baseline_period": f"{baseline_start} to {baseline_end}",
            }

            # Update investigation
            investigation.update_status(InvestigationStatus.COMPLETED)
            self.storage.save_investigation(investigation)

            logger.info(f"✅ Completed error pattern analysis for investigation {investigation_id}")

            return analysis

        except Exception as e:
            logger.error(f"❌ Error pattern analysis failed: {e}", exc_info=True)
            analysis.status = InvestigationStatus.FAILED
            analysis.error = str(e)
            analysis.end_time = datetime.utcnow()
            investigation.update_status(InvestigationStatus.FAILED)
            self.storage.save_investigation(investigation)
            raise

    async def run_slow_request_analysis(
        self,
        investigation_id: str,
        trace_tags: Optional[Dict[str, str]] = None,
    ) -> Analysis:
        """
        Run slow request analysis for an investigation

        Args:
            investigation_id: Investigation ID
            trace_tags: Optional trace tags (will be built from labels if not provided)

        Returns:
            Analysis result
        """
        investigation = self.storage.get_investigation(investigation_id)
        if not investigation:
            raise ValueError(f"Investigation {investigation_id} not found")

        # Create analysis
        analysis = Analysis(
            type=AnalysisType.SLOW_REQUEST,
            status=InvestigationStatus.RUNNING,
        )
        investigation.add_analysis(analysis)
        investigation.update_status(InvestigationStatus.RUNNING)
        self.storage.save_investigation(investigation)

        try:
            # Build trace tags from labels if not provided
            if not trace_tags:
                trace_tags = self._build_trace_tags(investigation.labels)

            logger.info(f"🔬 Running slow request analysis for investigation {investigation_id}")

            # Query current period traces
            current_result = await self.tempo_client.search_traces(
                tags=trace_tags,
                start=investigation.start_time,
                end=investigation.end_time,
                limit=100,
            )

            # Query baseline period traces (24 hours before current period)
            baseline_start = investigation.start_time - timedelta(hours=24)
            baseline_end = investigation.start_time
            baseline_result = await self.tempo_client.search_traces(
                tags=trace_tags,
                start=baseline_start,
                end=baseline_end,
                limit=100,
            )

            # Extract traces
            current_traces = self._extract_traces(current_result)
            baseline_traces = self._extract_traces(baseline_result)

            # Run analysis
            analysis_result = self.slow_request_analyzer.analyze(current_traces, baseline_traces)

            # Update analysis
            analysis.status = InvestigationStatus.COMPLETED
            analysis.end_time = datetime.utcnow()
            analysis.result = analysis_result
            analysis.metadata = {
                "trace_tags": trace_tags,
                "current_period": f"{investigation.start_time} to {investigation.end_time}",
                "baseline_period": f"{baseline_start} to {baseline_end}",
            }

            # Update investigation
            investigation.update_status(InvestigationStatus.COMPLETED)
            self.storage.save_investigation(investigation)

            logger.info(f"✅ Completed slow request analysis for investigation {investigation_id}")

            return analysis

        except Exception as e:
            logger.error(f"❌ Slow request analysis failed: {e}", exc_info=True)
            analysis.status = InvestigationStatus.FAILED
            analysis.error = str(e)
            analysis.end_time = datetime.utcnow()
            investigation.update_status(InvestigationStatus.FAILED)
            self.storage.save_investigation(investigation)
            raise

    async def get_investigation(self, investigation_id: str) -> Optional[Investigation]:
        """Get an investigation by ID"""
        return self.storage.get_investigation(investigation_id)

    async def list_investigations(self, limit: int = 10) -> List[Investigation]:
        """List recent investigations"""
        return self.storage.list_investigations(limit=limit)

    def _build_log_query(self, labels: Dict[str, str]) -> str:
        """Build LogQL query from labels"""
        if not labels:
            return '{job="varlogs"}'

        # Build label selector
        selectors = [f'{key}="{value}"' for key, value in labels.items()]
        return "{" + ", ".join(selectors) + "}"

    def _build_trace_tags(self, labels: Dict[str, str]) -> Dict[str, str]:
        """Build trace tags from labels"""
        # Map common Kubernetes labels to trace tags
        tag_mapping = {
            "namespace": "namespace",
            "cluster": "cluster",
            "service": "service.name",
            "app": "service.name",
        }

        tags = {}
        for label_key, label_value in labels.items():
            tag_key = tag_mapping.get(label_key, label_key)
            tags[tag_key] = label_value

        return tags

    def _extract_log_lines(self, query_result: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Extract log lines from Loki query result"""
        logs = []

        if query_result.get("status") != "success":
            return logs

        result = query_result.get("result", {})
        result_type = result.get("resultType", "")

        if result_type == "streams":
            streams = result.get("result", [])
            for stream in streams:
                values = stream.get("values", [])
                for value in values:
                    if len(value) >= 2:
                        logs.append({"timestamp": value[0], "line": value[1], "labels": stream.get("stream", {})})

        return logs

    def _extract_traces(self, search_result: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Extract traces from Tempo search result"""
        traces = []

        if search_result.get("status") != "success":
            return traces

        result = search_result.get("result", {})
        trace_list = result.get("traces", [])

        for trace in trace_list:
            traces.append(
                {
                    "traceID": trace.get("traceID"),
                    "rootServiceName": trace.get("rootServiceName"),
                    "rootTraceName": trace.get("rootTraceName"),
                    "duration": trace.get("durationMs", 0),
                    "startTimeUnixNano": trace.get("startTimeUnixNano"),
                    "spanName": trace.get("rootTraceName"),
                }
            )

        return traces

