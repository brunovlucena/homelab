"""
Example usage of macOS Automation Client from Kubernetes agents
"""

import asyncio
import logging
from macos_automation import MacOSAutomationClient, AutomationError

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


async def example_navigate():
    """Example: Navigate Safari to a URL"""
    client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")
    
    try:
        # Check service health
        health = await client.health()
        logger.info(f"Service health: {health}")
        
        # Navigate to URL
        result = await client.navigate("https://lucena.cloud")
        logger.info(f"Navigation result: {result}")
        
    except AutomationError as e:
        logger.error(f"Automation failed: {e}")
    finally:
        await client.close()


async def example_execute_javascript():
    """Example: Execute JavaScript in Safari"""
    client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")
    
    try:
        # Get page title
        result = await client.execute_javascript("document.title")
        title = result['result']['output']
        logger.info(f"Page title: {title}")
        
        # Get page URL
        result = await client.execute_javascript("window.location.href")
        url = result['result']['output']
        logger.info(f"Page URL: {url}")
        
    except AutomationError as e:
        logger.error(f"Automation failed: {e}")
    finally:
        await client.close()


async def example_get_info():
    """Example: Get Safari state information"""
    client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")
    
    try:
        info = await client.get_info()
        logger.info(f"Safari info: {info}")
        logger.info(f"Current URL: {info['result']['url']}")
        logger.info(f"Current title: {info['result']['title']}")
        logger.info(f"Window count: {info['result']['window_count']}")
        
    except AutomationError as e:
        logger.error(f"Automation failed: {e}")
    finally:
        await client.close()


async def example_cloudevent():
    """Example: Send CloudEvent to automation service"""
    client = MacOSAutomationClient(base_url="http://host.docker.internal:8080")
    
    try:
        result = await client.send_cloudevent(
            event_type="io.homelab.macos.automation.request",
            event_source="/example-agent/automation",
            data={
                "action": "navigate",
                "url": "https://example.com"
            }
        )
        logger.info(f"CloudEvent result: {result}")
        
    except AutomationError as e:
        logger.error(f"CloudEvent failed: {e}")
    finally:
        await client.close()


if __name__ == "__main__":
    # Run examples
    asyncio.run(example_navigate())
    asyncio.run(example_execute_javascript())
    asyncio.run(example_get_info())
    asyncio.run(example_cloudevent())

