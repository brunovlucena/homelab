#!/usr/bin/env python3
"""
Test script for Jamie -> Agent-SRE -> MCP flow
Tests that Jamie can successfully call agent-sre's MCP tools
"""

import asyncio
import aiohttp
import json
from datetime import datetime

# Configuration
AGENT_SRE_MCP_URL = "http://sre-agent-mcp-server-service.agent-sre:30120/mcp"
# For local testing:
# AGENT_SRE_MCP_URL = "http://localhost:30120/mcp"

async def test_mcp_initialize():
    """Test MCP initialize handshake"""
    print("🔧 Testing MCP initialize...")
    
    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "jamie-test-client",
                "version": "1.0.0"
            }
        }
    }
    
    async with aiohttp.ClientSession() as session:
        async with session.post(AGENT_SRE_MCP_URL, json=payload) as response:
            if response.status == 200:
                result = await response.json()
                print(f"✅ Initialize successful: {json.dumps(result, indent=2)}")
                return True
            else:
                error = await response.text()
                print(f"❌ Initialize failed: {error}")
                return False

async def test_mcp_tools_list():
    """Test listing available MCP tools"""
    print("\n📋 Testing tools/list...")
    
    payload = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {}
    }
    
    async with aiohttp.ClientSession() as session:
        async with session.post(AGENT_SRE_MCP_URL, json=payload) as response:
            if response.status == 200:
                result = await response.json()
                print(f"✅ Tools list successful:")
                tools = result.get("result", {}).get("tools", [])
                for tool in tools:
                    print(f"   - {tool['name']}: {tool['description']}")
                return True
            else:
                error = await response.text()
                print(f"❌ Tools list failed: {error}")
                return False

async def test_golden_signals(service_name="homepage"):
    """Test check_golden_signals tool"""
    print(f"\n📊 Testing check_golden_signals for service: {service_name}...")
    
    payload = {
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "check_golden_signals",
            "arguments": {
                "service_name": service_name,
                "namespace": "default"
            }
        }
    }
    
    async with aiohttp.ClientSession() as session:
        async with session.post(AGENT_SRE_MCP_URL, json=payload, timeout=aiohttp.ClientTimeout(total=30)) as response:
            if response.status == 200:
                result = await response.json()
                print(f"✅ Golden signals check successful:")
                if "result" in result and "content" in result["result"]:
                    content = result["result"]["content"]
                    if content and len(content) > 0 and "text" in content[0]:
                        data = json.loads(content[0]["text"])
                        print(f"   Service: {data.get('service_name')}")
                        print(f"   Overall Status: {data.get('overall_status')}")
                        print(f"   Signals:")
                        for signal_name, signal_data in data.get('signals', {}).items():
                            print(f"      - {signal_name}: {signal_data.get('value')} ({signal_data.get('status')})")
                return True
            else:
                error = await response.text()
                print(f"❌ Golden signals check failed: {error}")
                return False

async def test_sre_chat():
    """Test sre_chat tool"""
    print("\n💬 Testing sre_chat tool...")
    
    payload = {
        "jsonrpc": "2.0",
        "id": 4,
        "method": "tools/call",
        "params": {
            "name": "sre_chat",
            "arguments": {
                "message": "What are the best practices for monitoring web applications?"
            }
        }
    }
    
    async with aiohttp.ClientSession() as session:
        async with session.post(AGENT_SRE_MCP_URL, json=payload, timeout=aiohttp.ClientTimeout(total=30)) as response:
            if response.status == 200:
                result = await response.json()
                print(f"✅ SRE chat successful:")
                if "result" in result and "content" in result["result"]:
                    content = result["result"]["content"]
                    if content and len(content) > 0 and "text" in content[0]:
                        print(f"   Response: {content[0]['text'][:200]}...")
                return True
            else:
                error = await response.text()
                print(f"❌ SRE chat failed: {error}")
                return False

async def test_prometheus_query():
    """Test query_prometheus tool"""
    print("\n🔍 Testing query_prometheus tool...")
    
    payload = {
        "jsonrpc": "2.0",
        "id": 5,
        "method": "tools/call",
        "params": {
            "name": "query_prometheus",
            "arguments": {
                "query": "up"
            }
        }
    }
    
    async with aiohttp.ClientSession() as session:
        async with session.post(AGENT_SRE_MCP_URL, json=payload, timeout=aiohttp.ClientTimeout(total=30)) as response:
            if response.status == 200:
                result = await response.json()
                print(f"✅ Prometheus query successful:")
                if "result" in result and "content" in result["result"]:
                    content = result["result"]["content"]
                    if content and len(content) > 0 and "text" in content[0]:
                        print(f"   Result: {content[0]['text'][:200]}...")
                return True
            else:
                error = await response.text()
                print(f"❌ Prometheus query failed: {error}")
                return False

async def main():
    """Run all tests"""
    print("🚀 Starting Jamie -> Agent-SRE MCP Flow Tests\n")
    print(f"Target: {AGENT_SRE_MCP_URL}\n")
    
    tests = [
        ("Initialize", test_mcp_initialize),
        ("List Tools", test_mcp_tools_list),
        ("Golden Signals", lambda: test_golden_signals("homepage")),
        ("SRE Chat", test_sre_chat),
        ("Prometheus Query", test_prometheus_query),
    ]
    
    results = []
    for test_name, test_func in tests:
        try:
            success = await test_func()
            results.append((test_name, success))
        except Exception as e:
            print(f"❌ Test '{test_name}' failed with exception: {e}")
            results.append((test_name, False))
    
    # Summary
    print("\n" + "="*60)
    print("📊 Test Summary")
    print("="*60)
    
    for test_name, success in results:
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} - {test_name}")
    
    total = len(results)
    passed = sum(1 for _, success in results if success)
    print(f"\nTotal: {passed}/{total} tests passed")
    print("="*60)

if __name__ == "__main__":
    asyncio.run(main())

