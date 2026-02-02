#!/usr/bin/env python3
"""
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 Async Load Testing Framework

Purpose: High-performance async load testing for CloudEvents
User Story: QA-002 - Load and Performance Testing
Priority: P0 | Story Points: 13

Features:
- Asynchronous event generation
- Configurable concurrency and duration
- Real-time metrics collection
- Support for build, parser, and delete events

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"""

import asyncio
import aiohttp
import time
import json
import uuid
import os
import sys
from datetime import datetime, timezone
from dataclasses import dataclass
from typing import List

@dataclass
class LoadTestConfig:
    """Load test configuration"""
    broker_url: str = "http://0.0.0.0:8081"
    concurrent_requests: int = 100
    duration_seconds: int = 300
    rampup_seconds: int = 60
    event_type: str = "parser"  # build, parser, delete

@dataclass
class LoadTestResult:
    """Load test results"""
    total_requests: int
    successful_requests: int
    failed_requests: int
    avg_latency_ms: float
    p95_latency_ms: float
    p99_latency_ms: float
    requests_per_second: float
    duration_seconds: float

class AsyncLoadTester:
    """Asynchronous load testing framework"""
    
    def __init__(self, config: LoadTestConfig):
        self.config = config
        self.results: List[float] = []
        self.successful = 0
        self.failed = 0
        
    async def send_event(self, session: aiohttp.ClientSession, event: dict) -> bool:
        """Send a single CloudEvent"""
        start_time = time.time()
        
        try:
            headers = {'Content-Type': 'application/cloudevents+json'}
            async with session.post(
                self.config.broker_url,
                json=event,
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                latency_ms = (time.time() - start_time) * 1000
                self.results.append(latency_ms)
                
                if response.status in [200, 202]:
                    self.successful += 1
                    return True
                else:
                    self.failed += 1
                    return False
                    
        except Exception as e:
            self.failed += 1
            latency_ms = (time.time() - start_time) * 1000
            self.results.append(latency_ms)
            if self.failed % 100 == 0:  # Only print every 100th error
                print(f"❌ Error sending event: {e}")
            return False
    
    def create_build_event(self) -> dict:
        """Create a build CloudEvent"""
        third_party_id = f"load-test-{uuid.uuid4().hex[:16]}"
        parser_id = f"parser-{uuid.uuid4().hex[:8]}"
        
        return {
            "specversion": "1.0",
            "id": str(uuid.uuid4()),
            "source": f"network.notifi.{third_party_id}",
            "subject": parser_id,
            "type": "network.notifi.lambda.build.start",
            "time": datetime.now(timezone.utc).isoformat(),
            "data": {
                "third_party_id": third_party_id,
                "parser_id": parser_id,
            },
            "datacontenttype": "application/json"
        }
    
    def create_parser_event(self) -> dict:
        """Create a parser CloudEvent"""
        third_party_id = f"load-test-{uuid.uuid4().hex[:16]}"
        parser_id = f"parser-{uuid.uuid4().hex[:8]}"
        context_id = f"ctx-{uuid.uuid4().hex}"
        
        return {
            "specversion": "1.0",
            "id": str(uuid.uuid4()),
            "source": f"network.notifi.{third_party_id}",
            "subject": parser_id,
            "type": "network.notifi.lambda.parser.start",
            "time": datetime.now(timezone.utc).isoformat(),
            "data": {
                "contextId": context_id,
                "parameters": {
                    "blockId": str(659780 + int(time.time()) % 10000),
                    "blockchainType": 52,
                    "urlForBlob": "redis://ephemeralblock/LoadTest/999999",
                    "logIndices": [0]
                }
            },
            "datacontenttype": "application/json"
        }
    
    def create_delete_event(self) -> dict:
        """Create a delete CloudEvent"""
        third_party_id = f"load-test-{uuid.uuid4().hex[:16]}"
        parser_id = f"parser-{uuid.uuid4().hex[:8]}"
        
        return {
            "specversion": "1.0",
            "id": str(uuid.uuid4()),
            "source": f"network.notifi.{third_party_id}",
            "subject": parser_id,
            "type": "network.notifi.lambda.service.delete",
            "time": datetime.now(timezone.utc).isoformat(),
            "data": {
                "third_party_id": third_party_id,
                "parser_id": parser_id,
                "correlation_id": str(uuid.uuid4()),
                "reason": "Load test cleanup"
            },
            "datacontenttype": "application/json"
        }
    
    async def run_load_test(self) -> LoadTestResult:
        """Execute the load test"""
        print(f"🚀 Starting async load test...")
        print(f"   Target: {self.config.broker_url}")
        print(f"   Concurrent: {self.config.concurrent_requests}")
        print(f"   Duration: {self.config.duration_seconds}s")
        print(f"   Rampup: {self.config.rampup_seconds}s")
        print(f"   Event Type: {self.config.event_type}")
        print("=" * 80)
        
        start_time = time.time()
        
        # Create aiohttp session
        connector = aiohttp.TCPConnector(limit=self.config.concurrent_requests + 50)
        async with aiohttp.ClientSession(connector=connector) as session:
            
            # Create event generator based on type
            if self.config.event_type == "build":
                event_factory = self.create_build_event
            elif self.config.event_type == "parser":
                event_factory = self.create_parser_event
            elif self.config.event_type == "delete":
                event_factory = self.create_delete_event
            else:
                raise ValueError(f"Unknown event type: {self.config.event_type}")
            
            # Run load test with rampup
            tasks = []
            rampup_delay = self.config.rampup_seconds / self.config.concurrent_requests
            
            for i in range(self.config.concurrent_requests):
                # Rampup delay
                if i > 0 and self.config.rampup_seconds > 0:
                    await asyncio.sleep(rampup_delay)
                
                # Create continuous event sender
                task = asyncio.create_task(
                    self.continuous_sender(session, event_factory)
                )
                tasks.append(task)
                
                # Progress indicator
                if (i + 1) % 10 == 0:
                    print(f"⏳ Ramped up to {i + 1} workers...")
            
            print(f"🔥 All {self.config.concurrent_requests} workers active")
            
            # Wait for test duration
            await asyncio.sleep(self.config.duration_seconds)
            
            print(f"⏰ Test duration reached, shutting down workers...")
            
            # Cancel all tasks
            for task in tasks:
                task.cancel()
            
            # Wait for cancellations
            await asyncio.gather(*tasks, return_exceptions=True)
        
        duration = time.time() - start_time
        
        # Calculate statistics
        if not self.results:
            print("❌ No results collected!")
            sys.exit(1)
            
        self.results.sort()
        result = LoadTestResult(
            total_requests=self.successful + self.failed,
            successful_requests=self.successful,
            failed_requests=self.failed,
            avg_latency_ms=sum(self.results) / len(self.results),
            p95_latency_ms=self.results[int(len(self.results) * 0.95)],
            p99_latency_ms=self.results[int(len(self.results) * 0.99)],
            requests_per_second=(self.successful + self.failed) / duration,
            duration_seconds=duration
        )
        
        # Print results
        print("=" * 80)
        print(f"📊 LOAD TEST RESULTS:")
        print(f"   Total Requests: {result.total_requests:,}")
        print(f"   ✅ Successful: {result.successful_requests:,}")
        print(f"   ❌ Failed: {result.failed_requests:,}")
        print(f"   Success Rate: {(result.successful_requests / result.total_requests * 100):.2f}%")
        print(f"   Avg Latency: {result.avg_latency_ms:.2f}ms")
        print(f"   P95 Latency: {result.p95_latency_ms:.2f}ms")
        print(f"   P99 Latency: {result.p99_latency_ms:.2f}ms")
        print(f"   Requests/sec: {result.requests_per_second:.2f}")
        print(f"   Duration: {result.duration_seconds:.2f}s")
        print("=" * 80)
        
        return result
    
    async def continuous_sender(self, session: aiohttp.ClientSession, event_factory):
        """Continuously send events until cancelled"""
        try:
            while True:
                event = event_factory()
                await self.send_event(session, event)
                await asyncio.sleep(0.05)  # Small delay between events (20 req/sec per worker)
        except asyncio.CancelledError:
            pass

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🎯 Main Entry Point
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

if __name__ == "__main__":
    config = LoadTestConfig(
        broker_url=os.getenv("BROKER_URL", "http://0.0.0.0:8081"),
        concurrent_requests=int(os.getenv("CONCURRENT", "100")),
        duration_seconds=int(os.getenv("DURATION", "300")),
        rampup_seconds=int(os.getenv("RAMPUP", "60")),
        event_type=os.getenv("EVENT_TYPE", "parser")
    )
    
    tester = AsyncLoadTester(config)
    result = asyncio.run(tester.run_load_test())
    
    # Exit with error if success rate < 95%
    success_rate = result.successful_requests / result.total_requests
    if success_rate < 0.95:
        print(f"\n❌ Load test FAILED: Success rate {success_rate:.2%} < 95%")
        sys.exit(1)
    elif result.p95_latency_ms > 5000:
        print(f"\n⚠️ Load test WARNING: P95 latency {result.p95_latency_ms:.2f}ms > 5000ms")
        sys.exit(1)
    else:
        print(f"\n✅ Load test PASSED: Success rate {success_rate:.2%}, P95 {result.p95_latency_ms:.2f}ms")
        sys.exit(0)

