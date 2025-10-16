"""
🧪 Tests for Sift functionality
"""

import json
import sys
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

# Add deployments directory to path
sys.path.insert(0, str(Path(__file__).parent.parent / "deployments"))

from sift.analyzers import ErrorPatternAnalyzer, SlowRequestAnalyzer
from sift.investigation import Analysis, AnalysisType, Investigation, InvestigationStatus
from sift.loki_client import LokiClient
from sift.sift_core import SiftCore
from sift.storage import InvestigationStorage
from sift.tempo_client import TempoClient


@pytest.fixture
def temp_storage(tmp_path):
    """Create temporary storage"""
    db_path = tmp_path / "test_sift.db"
    return InvestigationStorage(str(db_path))


@pytest.fixture
def error_analyzer():
    """Create error pattern analyzer"""
    return ErrorPatternAnalyzer()


@pytest.fixture
def slow_request_analyzer():
    """Create slow request analyzer"""
    return SlowRequestAnalyzer()


class TestInvestigation:
    """Test investigation models"""

    def test_create_investigation(self):
        """Test creating an investigation"""
        inv = Investigation(
            name="Test Investigation",
            labels={"cluster": "prod", "namespace": "api"},
        )

        assert inv.id is not None
        assert inv.name == "Test Investigation"
        assert inv.labels == {"cluster": "prod", "namespace": "api"}
        assert inv.status == InvestigationStatus.PENDING
        assert len(inv.analyses) == 0

    def test_investigation_to_dict(self):
        """Test converting investigation to dict"""
        inv = Investigation(
            name="Test Investigation",
            labels={"cluster": "prod"},
        )

        data = inv.to_dict()

        assert data["id"] == inv.id
        assert data["name"] == "Test Investigation"
        assert data["labels"] == {"cluster": "prod"}
        assert data["status"] == "pending"

    def test_investigation_from_dict(self):
        """Test creating investigation from dict"""
        data = {
            "id": "test-id",
            "name": "Test",
            "labels": {"cluster": "prod"},
            "start_time": datetime.utcnow().isoformat(),
            "end_time": None,
            "status": "pending",
            "analyses": [],
            "created_at": datetime.utcnow().isoformat(),
            "updated_at": datetime.utcnow().isoformat(),
            "metadata": {},
        }

        inv = Investigation.from_dict(data)

        assert inv.id == "test-id"
        assert inv.name == "Test"
        assert inv.labels == {"cluster": "prod"}

    def test_add_analysis(self):
        """Test adding analysis to investigation"""
        inv = Investigation(name="Test")
        analysis = Analysis(type=AnalysisType.ERROR_PATTERN)

        inv.add_analysis(analysis)

        assert len(inv.analyses) == 1
        assert inv.analyses[0] == analysis


class TestStorage:
    """Test investigation storage"""

    def test_save_and_get_investigation(self, temp_storage):
        """Test saving and retrieving investigation"""
        inv = Investigation(
            name="Test Investigation",
            labels={"cluster": "prod"},
        )

        temp_storage.save_investigation(inv)
        retrieved = temp_storage.get_investigation(inv.id)

        assert retrieved is not None
        assert retrieved.id == inv.id
        assert retrieved.name == inv.name
        assert retrieved.labels == inv.labels

    def test_list_investigations(self, temp_storage):
        """Test listing investigations"""
        inv1 = Investigation(name="Test 1")
        inv2 = Investigation(name="Test 2")

        temp_storage.save_investigation(inv1)
        temp_storage.save_investigation(inv2)

        investigations = temp_storage.list_investigations()

        assert len(investigations) >= 2

    def test_delete_investigation(self, temp_storage):
        """Test deleting investigation"""
        inv = Investigation(name="Test")

        temp_storage.save_investigation(inv)
        deleted = temp_storage.delete_investigation(inv.id)

        assert deleted is True
        assert temp_storage.get_investigation(inv.id) is None


class TestErrorPatternAnalyzer:
    """Test error pattern analyzer"""

    def test_extract_patterns(self, error_analyzer):
        """Test extracting error patterns"""
        logs = [
            {"line": "ERROR: Database connection failed"},
            {"line": "ERROR: Database connection failed"},
            {"line": "INFO: Request processed"},
            {"line": "ERROR: Timeout occurred"},
        ]

        patterns = error_analyzer._extract_patterns(logs)

        assert len(patterns) > 0
        # Should have normalized patterns

    def test_normalize_log_line(self, error_analyzer):
        """Test log line normalization"""
        line = "2024-01-01T10:00:00.123Z ERROR: Connection to 192.168.1.1 failed with code 500"

        normalized = error_analyzer._normalize_log_line(line)

        # Should remove timestamp and IP
        assert "2024-01-01" not in normalized
        assert "192.168.1.1" not in normalized
        assert "ERROR" in normalized

    def test_is_error_pattern(self, error_analyzer):
        """Test error pattern detection"""
        assert error_analyzer._is_error_pattern("ERROR: something failed") is True
        assert error_analyzer._is_error_pattern("Exception occurred") is True
        assert error_analyzer._is_error_pattern("INFO: all good") is False

    def test_analyze(self, error_analyzer):
        """Test full analysis"""
        current_logs = [{"line": "ERROR: Failed"} for _ in range(10)]
        baseline_logs = [{"line": "ERROR: Failed"} for _ in range(2)]

        result = error_analyzer.analyze(current_logs, baseline_logs, threshold_multiplier=2.0)

        assert result["status"] == "completed"
        assert result["total_current_logs"] == 10
        assert result["total_baseline_logs"] == 2


class TestSlowRequestAnalyzer:
    """Test slow request analyzer"""

    def test_extract_durations(self, slow_request_analyzer):
        """Test extracting durations by operation"""
        traces = [
            {"spanName": "GET /api/users", "duration": 150},
            {"spanName": "GET /api/users", "duration": 200},
            {"spanName": "POST /api/orders", "duration": 500},
        ]

        durations = slow_request_analyzer._extract_durations_by_operation(traces)

        assert "GET /api/users" in durations
        assert len(durations["GET /api/users"]) == 2
        assert "POST /api/orders" in durations

    def test_percentile(self, slow_request_analyzer):
        """Test percentile calculation"""
        values = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100]

        p50 = slow_request_analyzer._percentile(values, 50)
        p95 = slow_request_analyzer._percentile(values, 95)

        assert p50 == 50
        assert p95 == 95

    def test_analyze(self, slow_request_analyzer):
        """Test full analysis"""
        # Current traces are slower
        current_traces = [{"spanName": "GET /api", "duration": 5000} for _ in range(10)]
        # Baseline traces are faster
        baseline_traces = [{"spanName": "GET /api", "duration": 1000} for _ in range(10)]

        result = slow_request_analyzer.analyze(current_traces, baseline_traces)

        assert result["status"] == "completed"
        assert result["total_current_traces"] == 10
        assert result["total_baseline_traces"] == 10


@pytest.mark.asyncio
class TestLokiClient:
    """Test Loki client"""

    @patch("aiohttp.ClientSession.get")
    async def test_query_range_success(self, mock_get):
        """Test successful Loki query"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={
                "status": "success",
                "data": {
                    "resultType": "streams",
                    "result": [
                        {
                            "stream": {"namespace": "default"},
                            "values": [[1234567890, "log line 1"]],
                        }
                    ],
                },
            }
        )
        mock_get.return_value.__aenter__.return_value = mock_response

        client = LokiClient("http://loki:3100")
        result = await client.query_range('{namespace="default"}')

        assert result["status"] == "success"
        assert "result" in result


@pytest.mark.asyncio
class TestTempoClient:
    """Test Tempo client"""

    @patch("aiohttp.ClientSession.get")
    async def test_search_traces_success(self, mock_get):
        """Test successful Tempo search"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={
                "traces": [
                    {
                        "traceID": "abc123",
                        "rootServiceName": "api",
                        "rootTraceName": "GET /users",
                        "durationMs": 150,
                    }
                ]
            }
        )
        mock_get.return_value.__aenter__.return_value = mock_response

        client = TempoClient("http://tempo:3100")
        result = await client.search_traces(tags={"service.name": "api"})

        assert result["status"] == "success"
        assert "result" in result


@pytest.mark.asyncio
class TestSiftCore:
    """Test Sift core orchestration"""

    @pytest.fixture
    async def sift_core(self, tmp_path):
        """Create Sift core instance"""
        db_path = tmp_path / "test_sift.db"
        core = SiftCore(
            loki_url="http://loki:3100",
            tempo_url="http://tempo:3100",
            storage_path=str(db_path),
        )
        return core

    async def test_create_investigation(self, sift_core):
        """Test creating investigation"""
        inv = await sift_core.create_investigation(
            name="Test Investigation",
            labels={"cluster": "prod", "namespace": "api"},
        )

        assert inv is not None
        assert inv.id is not None
        assert inv.name == "Test Investigation"
        assert inv.labels == {"cluster": "prod", "namespace": "api"}

    async def test_get_investigation(self, sift_core):
        """Test getting investigation"""
        inv = await sift_core.create_investigation(
            name="Test",
            labels={"cluster": "prod"},
        )

        retrieved = await sift_core.get_investigation(inv.id)

        assert retrieved is not None
        assert retrieved.id == inv.id

    async def test_list_investigations(self, sift_core):
        """Test listing investigations"""
        await sift_core.create_investigation("Test 1", {"cluster": "prod"})
        await sift_core.create_investigation("Test 2", {"cluster": "prod"})

        investigations = await sift_core.list_investigations()

        assert len(investigations) >= 2

    def test_build_log_query(self, sift_core):
        """Test building LogQL query"""
        labels = {"namespace": "default", "pod": "test-pod"}

        query = sift_core._build_log_query(labels)

        assert "namespace" in query
        assert "pod" in query
        assert "default" in query

    def test_build_trace_tags(self, sift_core):
        """Test building trace tags"""
        labels = {"namespace": "default", "service": "api"}

        tags = sift_core._build_trace_tags(labels)

        assert "namespace" in tags
        assert tags["service.name"] == "api"

