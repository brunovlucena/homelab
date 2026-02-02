"""
macOS Automation Client
Python client for interacting with macOS automation service
Can be used from Kubernetes agents to control macOS applications
"""

import logging
from typing import Any, Dict, Optional

import httpx

logger = logging.getLogger(__name__)


class AutomationError(Exception):
    """Exception raised when automation fails"""
    pass


class MacOSAutomationClient:
    """
    Client for macOS automation service
    
    Usage from Kubernetes agents:
        client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")
        result = await client.navigate("https://example.com")
        js_result = await client.execute_javascript("document.title")
    """
    
    def __init__(
        self,
        base_url: str = "http://host.docker.internal:8080",
        timeout: float = 30.0
    ):
        """
        Initialize macOS automation client
        
        Args:
            base_url: Base URL of macOS automation service
                     Use "http://host.docker.internal:8080" from Kubernetes
                     Use "http://localhost:8080" from local macOS
            timeout: Request timeout in seconds
        """
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)
    
    async def health(self) -> Dict[str, Any]:
        """Check if automation service is healthy"""
        try:
            response = await self.client.get(f"{self.base_url}/health")
            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            raise AutomationError(f"Service unavailable: {e}")
    
    async def navigate(
        self,
        url: str,
        application: str = "Safari",
        wait: Optional[float] = None
    ) -> Dict[str, Any]:
        """
        Navigate browser to URL
        
        Args:
            url: URL to navigate to
            application: Application to control (default: Safari)
            wait: Wait time in seconds after navigation
        
        Returns:
            Automation response with success status
        """
        payload = {
            "action": "navigate",
            "application": application,
            "url": url,
            "wait": wait
        }
        
        return await self._execute(payload)
    
    async def execute_javascript(
        self,
        javascript: str,
        application: str = "Safari"
    ) -> Dict[str, Any]:
        """
        Execute JavaScript in browser
        
        Args:
            javascript: JavaScript code to execute
            application: Application to control (default: Safari)
        
        Returns:
            Automation response with JavaScript result
        """
        payload = {
            "action": "execute_js",
            "application": application,
            "javascript": javascript
        }
        
        return await self._execute(payload)
    
    async def execute_applescript(
        self,
        applescript: str
    ) -> Dict[str, Any]:
        """
        Execute raw AppleScript
        
        Args:
            applescript: AppleScript code to execute
        
        Returns:
            Automation response with AppleScript result
        """
        payload = {
            "action": "applescript",
            "applescript": applescript
        }
        
        return await self._execute(payload)
    
    async def get_info(
        self,
        application: str = "Safari"
    ) -> Dict[str, Any]:
        """
        Get current state of application
        
        Args:
            application: Application to query (default: Safari)
        
        Returns:
            Application state information
        """
        payload = {
            "action": "info",
            "application": application
        }
        
        return await self._execute(payload)
    
    async def _execute(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Execute automation request"""
        try:
            response = await self.client.post(
                f"{self.base_url}/v1/automation/execute",
                json=payload
            )
            response.raise_for_status()
            result = response.json()
            
            if not result.get("success"):
                error = result.get("error", "Unknown error")
                raise AutomationError(f"Automation failed: {error}")
            
            return result
        
        except httpx.HTTPError as e:
            logger.error(f"HTTP error: {e}")
            raise AutomationError(f"HTTP error: {e}")
        except Exception as e:
            logger.error(f"Unexpected error: {e}")
            raise AutomationError(f"Unexpected error: {e}")
    
    async def send_cloudevent(
        self,
        event_type: str,
        event_source: str,
        data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Send CloudEvent to automation service
        
        Args:
            event_type: CloudEvent type
            event_source: CloudEvent source
            data: Event data payload
        
        Returns:
            CloudEvent response
        """
        try:
            headers = {
                "ce-type": event_type,
                "ce-source": event_source,
                "Content-Type": "application/json"
            }
            
            payload = {
                "type": event_type,
                "source": event_source,
                "data": data
            }
            
            response = await self.client.post(
                f"{self.base_url}/v1/events",
                json=payload,
                headers=headers
            )
            response.raise_for_status()
            return response.json()
        
        except Exception as e:
            logger.error(f"CloudEvent error: {e}")
            raise AutomationError(f"CloudEvent failed: {e}")
    
    async def close(self):
        """Close HTTP client"""
        await self.client.aclose()

