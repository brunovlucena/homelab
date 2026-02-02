"""
Unit tests for Alert Dispatcher.
"""
import pytest
from unittest.mock import AsyncMock, patch, MagicMock
import os

import sys
sys.path.insert(0, str(__file__).replace("/tests/unit/test_alert_dispatcher.py", "/src"))

from alert_dispatcher.handler import (
    AlertDispatcher,
    Alert,
    Severity,
    AlertChannel,
    SEVERITY_CHANNELS,
    TelegramNotifier,
    DiscordNotifier,
    GrafanaNotifier,
)


class TestAlert:
    """Tests for Alert dataclass."""
    
    def test_auto_timestamp(self):
        """Test that timestamp is auto-generated."""
        alert = Alert(
            severity=Severity.HIGH,
            title="Test",
            chain="ethereum",
            contract_address="0x1234",
            vulnerability_type="reentrancy",
        )
        
        assert alert.timestamp != ""
        assert "T" in alert.timestamp  # ISO format
    
    def test_to_dict(self):
        """Test serialization to dict."""
        alert = Alert(
            severity=Severity.CRITICAL,
            title="Critical Finding",
            chain="ethereum",
            contract_address="0x1234",
            vulnerability_type="reentrancy",
            profit_potential="10 ETH",
            exploit_validated=True,
        )
        
        d = alert.to_dict()
        assert d["severity"] == "critical"
        assert d["exploit_validated"] is True
        assert d["profit_potential"] == "10 ETH"


class TestSeverityRouting:
    """Tests for severity-based channel routing."""
    
    def test_critical_routes_to_all(self):
        """Critical alerts go to all channels."""
        channels = SEVERITY_CHANNELS[Severity.CRITICAL]
        
        assert AlertChannel.TELEGRAM in channels
        assert AlertChannel.DISCORD in channels
        assert AlertChannel.GRAFANA_INCIDENT in channels
        assert AlertChannel.EMAIL in channels
    
    def test_low_routes_to_grafana_only(self):
        """Low severity only goes to Grafana."""
        channels = SEVERITY_CHANNELS[Severity.LOW]
        
        assert AlertChannel.GRAFANA in channels
        assert AlertChannel.TELEGRAM not in channels
        assert AlertChannel.DISCORD not in channels
    
    def test_high_excludes_email(self):
        """High severity doesn't trigger email."""
        channels = SEVERITY_CHANNELS[Severity.HIGH]
        
        assert AlertChannel.EMAIL not in channels
        assert AlertChannel.TELEGRAM in channels


class TestTelegramNotifier:
    """Tests for Telegram notifications."""
    
    def test_disabled_when_no_config(self):
        """Telegram is disabled without credentials."""
        with patch.dict(os.environ, {}, clear=True):
            notifier = TelegramNotifier()
            assert notifier.enabled is False
    
    def test_enabled_with_config(self):
        """Telegram is enabled with credentials."""
        with patch.dict(os.environ, {
            "TELEGRAM_BOT_TOKEN": "test-token",
            "TELEGRAM_CHAT_ID": "123456",
        }):
            notifier = TelegramNotifier()
            assert notifier.enabled is True
    
    @pytest.mark.asyncio
    async def test_send_returns_false_when_disabled(self):
        """Send returns False when disabled."""
        with patch.dict(os.environ, {}, clear=True):
            notifier = TelegramNotifier()
            alert = Alert(
                severity=Severity.HIGH,
                title="Test",
                chain="ethereum",
                contract_address="0x1234",
                vulnerability_type="test",
            )
            result = await notifier.send(alert)
            assert result is False


class TestDiscordNotifier:
    """Tests for Discord notifications."""
    
    def test_disabled_when_no_config(self):
        """Discord is disabled without webhook URL."""
        with patch.dict(os.environ, {}, clear=True):
            notifier = DiscordNotifier()
            assert notifier.enabled is False
    
    def test_enabled_with_config(self):
        """Discord is enabled with webhook URL."""
        with patch.dict(os.environ, {
            "DISCORD_WEBHOOK_URL": "https://discord.com/api/webhooks/test",
        }):
            notifier = DiscordNotifier()
            assert notifier.enabled is True


class TestGrafanaNotifier:
    """Tests for Grafana notifications."""
    
    def test_disabled_when_no_api_key(self):
        """Grafana is disabled without API key."""
        with patch.dict(os.environ, {}, clear=True):
            notifier = GrafanaNotifier()
            assert notifier.enabled is False


class TestAlertDispatcher:
    """Tests for main AlertDispatcher class."""
    
    @pytest.mark.asyncio
    async def test_dispatch_routes_by_severity(self):
        """Test that alerts are routed to correct channels."""
        dispatcher = AlertDispatcher()
        
        # Mock all notifiers
        dispatcher.telegram = AsyncMock()
        dispatcher.telegram.send = AsyncMock(return_value=True)
        dispatcher.discord = AsyncMock()
        dispatcher.discord.send = AsyncMock(return_value=True)
        dispatcher.grafana = AsyncMock()
        dispatcher.grafana.send = AsyncMock(return_value=True)
        dispatcher.grafana.create_incident = AsyncMock(return_value=True)
        
        alert = Alert(
            severity=Severity.CRITICAL,
            title="Critical Test",
            chain="ethereum",
            contract_address="0x1234",
            vulnerability_type="reentrancy",
        )
        
        results = await dispatcher.dispatch(alert)
        
        # Critical should call telegram, discord, grafana incident
        dispatcher.telegram.send.assert_called_once()
        dispatcher.discord.send.assert_called_once()
        dispatcher.grafana.create_incident.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_dispatch_low_severity(self):
        """Test low severity only goes to Grafana."""
        dispatcher = AlertDispatcher()
        
        dispatcher.telegram = AsyncMock()
        dispatcher.telegram.send = AsyncMock(return_value=True)
        dispatcher.discord = AsyncMock()
        dispatcher.discord.send = AsyncMock(return_value=True)
        dispatcher.grafana = AsyncMock()
        dispatcher.grafana.send = AsyncMock(return_value=True)
        
        alert = Alert(
            severity=Severity.LOW,
            title="Low Test",
            chain="ethereum",
            contract_address="0x1234",
            vulnerability_type="info",
        )
        
        results = await dispatcher.dispatch(alert)
        
        # Low should only call grafana.send
        dispatcher.telegram.send.assert_not_called()
        dispatcher.discord.send.assert_not_called()
        dispatcher.grafana.send.assert_called_once()

