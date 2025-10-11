#!/usr/bin/env python3
"""
Unit tests for core module
"""

import os
import sys
from unittest.mock import patch

import pytest

# Import the module to test
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src", "mcp-server"))

from core import JAMIE_SLACK_BOT_URL, SERVICE_NAME, logger, logfire


class TestCoreModule:
    """Test suite for core module"""

    def test_logger_exists(self):
        """Test that logger is properly initialized"""
        assert logger is not None
        assert hasattr(logger, "info")
        assert hasattr(logger, "error")
        assert hasattr(logger, "warning")
        assert hasattr(logger, "debug")

    def test_logfire_exists(self):
        """Test that logfire is properly initialized"""
        assert logfire is not None

    def test_environment_variables(self):
        """Test that expected environment variables are accessible"""
        # These should have defaults in core.py
        assert JAMIE_SLACK_BOT_URL is not None
        assert isinstance(JAMIE_SLACK_BOT_URL, str)
        assert SERVICE_NAME is not None
        assert isinstance(SERVICE_NAME, str)

    def test_service_name(self):
        """Test service name constant"""
        assert SERVICE_NAME == "jamie-mcp-server"

    @patch.dict(os.environ, {"JAMIE_SLACK_BOT_URL": "http://custom:8080"})
    def test_custom_environment_variables(self):
        """Test that environment variables can be customized"""
        # Re-import to get new env values
        import importlib
        import core as core_module

        importlib.reload(core_module)

        assert core_module.JAMIE_SLACK_BOT_URL == "http://custom:8080"

    def test_logger_output(self, caplog):
        """Test that logger actually logs"""
        import logging

        with caplog.at_level(logging.INFO, logger="jamie-mcp"):
            logger.info("Test message")
            assert "Test message" in caplog.text or len(caplog.records) > 0

    def test_logger_error(self, caplog):
        """Test that logger logs errors"""
        import logging

        with caplog.at_level(logging.ERROR, logger="jamie-mcp"):
            logger.error("Test error")
            assert "Test error" in caplog.text or "ERROR" in caplog.text or len(caplog.records) > 0


class TestLogging:
    """Test suite for logging functionality"""

    def test_logger_levels(self, caplog):
        """Test different log levels"""
        logger.debug("Debug message")
        logger.info("Info message")
        logger.warning("Warning message")
        logger.error("Error message")

        # At least some of these should be in the log
        assert any(
            msg in caplog.text
            for msg in ["Debug message", "Info message", "Warning message", "Error message"]
        )

    def test_logger_formatting(self, caplog):
        """Test logger message formatting"""
        import logging

        test_data = {"key": "value", "number": 42}
        with caplog.at_level(logging.INFO, logger="jamie-mcp"):
            logger.info(f"Test data: {test_data}")
            assert "Test data" in caplog.text or len(caplog.records) > 0

    def test_logger_exception(self, caplog):
        """Test logging exceptions"""
        try:
            raise ValueError("Test exception")
        except ValueError as e:
            logger.error(f"Caught exception: {e}")

        assert "exception" in caplog.text.lower() or "error" in caplog.text.lower()

