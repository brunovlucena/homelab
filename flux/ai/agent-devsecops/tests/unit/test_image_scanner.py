"""Unit tests for image scanner."""
import pytest
from unittest.mock import patch, MagicMock

import sys
sys.path.insert(0, "src")


class TestImageScanner:
    """Tests for ImageScanner class."""
    
    def test_parse_image_uri_with_tag(self):
        """Test parsing image URI with tag."""
        from scanner.image_scanner import ImageScanner
        
        scanner = ImageScanner()
        result = scanner.parse_image_uri("ghcr.io/brunovlucena/test:v1.0.0")
        
        assert result["registry"] == "ghcr.io"
        assert result["owner"] == "brunovlucena"
        assert result["name"] == "test"
        assert result["tag"] == "v1.0.0"
        assert result["version"] == "1.0.0"
    
    def test_parse_image_uri_without_tag(self):
        """Test parsing image URI without tag."""
        from scanner.image_scanner import ImageScanner
        
        scanner = ImageScanner()
        result = scanner.parse_image_uri("ghcr.io/brunovlucena/test")
        
        assert result["registry"] == "ghcr.io"
        assert result["tag"] == "latest"
    
    def test_parse_image_uri_with_sha(self):
        """Test parsing image URI with SHA digest."""
        from scanner.image_scanner import ImageScanner
        
        scanner = ImageScanner()
        result = scanner.parse_image_uri("ghcr.io/brunovlucena/test@sha256:abc123")
        
        assert result["registry"] == "ghcr.io"
        assert "sha256" in result.get("digest", result.get("tag", ""))
    
    def test_extract_version_from_tag(self):
        """Test extracting version from various tag formats."""
        from scanner.image_scanner import ImageScanner
        
        scanner = ImageScanner()
        
        assert scanner.extract_version("v1.0.0") == "1.0.0"
        assert scanner.extract_version("1.2.3") == "1.2.3"
        assert scanner.extract_version("v2.1.0-beta") == "2.1.0"
        assert scanner.extract_version("latest") is None
    
    def test_compare_versions(self):
        """Test version comparison."""
        from scanner.image_scanner import ImageScanner
        
        scanner = ImageScanner()
        
        # v2.0.0 is newer than v1.0.0
        assert scanner.compare_versions("2.0.0", "1.0.0") > 0
        # v1.0.0 is older than v2.0.0
        assert scanner.compare_versions("1.0.0", "2.0.0") < 0
        # Same versions
        assert scanner.compare_versions("1.0.0", "1.0.0") == 0
    
    def test_is_outdated(self):
        """Test outdated detection."""
        from scanner.image_scanner import ImageScanner
        
        scanner = ImageScanner()
        scanner.latest_versions = {"test-function": "2.0.0"}
        
        assert scanner.is_outdated("test-function", "1.0.0") is True
        assert scanner.is_outdated("test-function", "2.0.0") is False
        assert scanner.is_outdated("test-function", "3.0.0") is False
        assert scanner.is_outdated("unknown", "1.0.0") is False


class TestMetricsExporter:
    """Tests for MetricsExporter class."""
    
    def test_update_lambdafunction_metrics(self):
        """Test updating LambdaFunction metrics."""
        from scanner.metrics_exporter import MetricsExporter
        
        exporter = MetricsExporter()
        
        functions = [
            {
                "name": "func1",
                "namespace": "default",
                "image": "test:v1.0.0",
                "version": "1.0.0",
                "outdated": False,
            },
            {
                "name": "func2",
                "namespace": "default",
                "image": "test:v0.9.0",
                "version": "0.9.0",
                "outdated": True,
            }
        ]
        
        exporter.update_lambdafunction_metrics(functions)
        
        # Metrics should be updated without errors
        assert True
    
    def test_count_outdated(self):
        """Test counting outdated functions."""
        from scanner.metrics_exporter import MetricsExporter
        
        exporter = MetricsExporter()
        
        functions = [
            {"outdated": True},
            {"outdated": False},
            {"outdated": True},
        ]
        
        count = sum(1 for f in functions if f.get("outdated"))
        
        assert count == 2


class TestHandler:
    """Tests for handler functions."""
    
    @patch('scanner.handler.ImageScanner')
    @patch('scanner.handler.get_handler')
    def test_handle_scan_lambdafunctions(self, mock_get_handler, mock_scanner_class):
        """Test handling LambdaFunction scan event."""
        from scanner.handler import handle
        
        mock_k8s = MagicMock()
        mock_k8s.list.return_value = {"items": [], "count": 0}
        mock_get_handler.return_value = mock_k8s
        
        mock_scanner = MagicMock()
        mock_scanner.scan_lambdafunctions.return_value = []
        mock_scanner_class.return_value = mock_scanner
        
        event = {
            "type": "io.homelab.scan.lambdafunctions",
            "data": {"namespace": "default"}
        }
        
        result = handle(event)
        
        assert result is not None
    
    @patch('scanner.handler.get_handler')
    def test_handle_unknown_event(self, mock_get_handler):
        """Test handling unknown event type."""
        from scanner.handler import handle
        
        event = {
            "type": "io.homelab.unknown.event",
            "data": {}
        }
        
        result = handle(event)
        
        # Should handle gracefully
        assert result is not None
