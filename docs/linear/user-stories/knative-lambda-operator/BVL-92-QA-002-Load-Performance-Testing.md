# 🧪 QA-002: Load and Performance Testing

**Priority**: P1 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-254/qa-002-load-and-performance-testing

**Status:** ✅ Active  
**Priority:** P0  
**Story Points:** 13  
**Sprint:** Sprint 3  
**Team:** QA Engineering  

---

## 📋 Story

**As a** QA Engineer  
**I want** to validate system performance under high load conditions  
**So that** I can ensure the system meets SLOs and scales appropriately

---

## 🎯 Acceptance Criteria

### ✅ AC1: Build Event Load Testing
- [ ] Simulate 100-500 concurrent build events
- [ ] Measure build job creation rate (target: > 50/sec)
- [ ] Verify no event loss during load
- [ ] Validate autoscaling triggers appropriately
- [ ] Confirm system recovers after spike

**Load Scenarios:**
1. **Baseline:** 10 events/sec for 5 minutes
2. **Ramp-up:** 10 → 100 → 200 events/sec over 10 minutes
3. **Spike:** 500 events/sec burst for 2 minutes
4. **Sustained:** 100 events/sec for 30 minutes

### ✅ AC2: Parser Event Load Testing
- [ ] Simulate 500-1500 concurrent parser events
- [ ] Measure event processing latency (p95 < 3s)
- [ ] Verify lambda service autoscaling (0 → 10+ pods)
- [ ] Validate no dropped events under load
- [ ] Confirm proper backpressure handling

**Load Scenarios:**
1. **Blockchain burst:** 800 events/sec (new block simulation)
2. **Trading events:** 1200 events/sec sustained
3. **Multi-parser:** 3 parsers × 500 events/sec each
4. **Spike test:** 1500 events/sec burst

### ✅ AC3: HTTP Direct Load Testing
- [ ] Send direct HTTP requests to lambda services
- [ ] Measure HTTP request latency (p95 < 5s)
- [ ] Verify cold-start performance (< 10s)
- [ ] Validate concurrent request handling (8+ req/pod)
- [ ] Confirm proper error handling under load

**Load Scenarios:**
1. **Cold-start:** 20 concurrent requests to scaled-to-zero service
2. **Warm service:** 100 concurrent requests to running service
3. **Sustained load:** 50 req/sec for 15 minutes

### ✅ AC4: Stress Testing
- [ ] Push system beyond normal capacity
- [ ] Identify breaking points
- [ ] Measure graceful degradation
- [ ] Validate error rates under extreme load
- [ ] Confirm system recovery after stress

**Stress Scenarios:**
1. **Extreme burst:** 2000 events/sec
2. **Resource exhaustion:** Fill all node capacity
3. **Network saturation:** Max out broker connections
4. **Chaos testing:** Random pod deletions during load

### ✅ AC5: Endurance Testing
- [ ] Run sustained load for 8+ hours
- [ ] Monitor memory leaks
- [ ] Validate no resource degradation
- [ ] Confirm metrics remain stable
- [ ] Assert no connection pool exhaustion

---

## 🔧 Technical Implementation

### K6 Load Test Framework
```javascript
// tests/load/k6/builder-load-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const errorRate = new Rate('builder_error_rate');
const eventCounter = new Counter('builder_events_total');
const publishDuration = new Trend('builder_publish_duration');

// Load test configuration
export const options = {
  scenarios: {
    // Scenario 1: Baseline load
    baseline: {
      executor: 'constant-arrival-rate',
      rate: 10,                    // 10 events/sec
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 10,
      maxVUs: 50,
      tags: { scenario: 'baseline' },
    },
    
    // Scenario 2: Ramp-up test
    rampup: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      stages: [
        { duration: '3m', target: 50 },
        { duration: '3m', target: 100 },
        { duration: '3m', target: 200 },
        { duration: '3m', target: 100 },
        { duration: '2m', target: 10 },
      ],
      preAllocatedVUs: 100,
      maxVUs: 300,
      tags: { scenario: 'rampup' },
      startTime: '5m',  // Start after baseline
    },
    
    // Scenario 3: Spike test
    spike: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      stages: [
        { duration: '1m', target: 500 },  // Spike to 500 events/sec
        { duration: '2m', target: 500 },  // Sustain spike
        { duration: '1m', target: 100 },  // Ramp down
      ],
      preAllocatedVUs: 200,
      maxVUs: 600,
      tags: { scenario: 'spike' },
      startTime: '20m',  // Start after rampup
    },
    
    // Scenario 4: Sustained load
    sustained: {
      executor: 'constant-arrival-rate',
      rate: 100,                   // 100 events/sec
      timeUnit: '1s',
      duration: '30m',
      preAllocatedVUs: 100,
      maxVUs: 200,
      tags: { scenario: 'sustained' },
      startTime: '25m',  // Start after spike
    },
  },
  
  thresholds: {
    // HTTP-level thresholds
    'http_req_duration': ['p(95)<5000'],        // 95% under 5s
    'http_req_duration{scenario:baseline}': ['p(95)<2000'],  // Baseline tighter
    'http_req_failed': ['rate<0.01'],           // < 1% failures
    
    // Custom metric thresholds
    'builder_error_rate': ['rate<0.01'],
    'builder_publish_duration': ['p(95)<3000'],
    'builder_events_total': ['count>10000'],    // Should send 10k+ events
  },
};

// Test data
const testEvents = [
  { third_party_id: "0307ea43639b4616b044d190310a26bd", parser_id: "0197ad6c10b973b2b854a0e652155b7e" },
  { third_party_id: "0307ea43639b4616b044d190310a26bd", parser_id: "c42d2e6ca3214f4b8d28a2cab47beecf" },
  { third_party_id: "0307ea43639b4616b044d190310a26bd", parser_id: "e0a711bde5d748009a995432acbf590b" },
  { third_party_id: "1234567890abcdef1234567890abcdef", parser_id: "enterprise-parser-001" },
  { third_party_id: "fedcba0987654321fedcba0987654321", parser_id: "trading-bot-alpha" },
];

const RABBITMQ_URL = __ENV.RABBITMQ_URL | | 'http://notifi:notifi@rabbitmq-cluster-dev.rabbitmq-dev:15672';
const EXCHANGE_NAME = 'cloud-events';
const ROUTING_KEY = 'network.notifi.lambda.build.start';

function createCloudEvent(thirdPartyId, parserId) {
  return {
    specversion: "1.0",
    id: randomString(32),
    source: `network.notifi.parsers.${thirdPartyId}.${parserId}`,
    type: "network.notifi.lambda.build.start",
    time: new Date().toISOString(),
    data: {
      third_party_id: thirdPartyId,
      parser_id: parserId,
    },
    datacontenttype: "application/json"
  };
}

export default function () {
  // Select random test event
  const testEvent = testEvents[Math.floor(Math.random() * testEvents.length)];
  const cloudEvent = createCloudEvent(testEvent.third_party_id, testEvent.parser_id);
  
  // Publish to RabbitMQ
  const publishUrl = `${RABBITMQ_URL}/api/exchanges/%2F/${EXCHANGE_NAME}/publish`;
  const publishPayload = {
    properties: {
      content_type: "application/cloudevents+json",
      delivery_mode: 2
    },
    routing_key: ROUTING_KEY,
    payload: JSON.stringify(cloudEvent),
    payload_encoding: "string"
  };
  
  const startTime = Date.now();
  const response = http.post(publishUrl, JSON.stringify(publishPayload), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });
  const duration = Date.now() - startTime;
  
  publishDuration.add(duration);
  
  const success = check(response, {
    'event published successfully': (r) => r.status === 200,
    'response time < 3s': (r) => r.timings.duration < 3000,
  });
  
  if (success) {
    eventCounter.add(1);
  } else {
    errorRate.add(1);
    console.error(`Failed to publish event: ${response.status} - ${response.body}`);
  }
  
  // Small random delay to avoid thundering herd
  sleep(Math.random() * 0.3);
}

export function handleSummary(data) {
  const timestamp = Date.now();
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    [`/tmp/k6-builder-load-${timestamp}.json`]: JSON.stringify(data),
  };
}
```

### Python Async Load Test
```python
# tests/load/python/async_load_test.py

import asyncio
import aiohttp
import time
import json
import uuid
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
            print(f"Error sending event: {e}")
            return False
    
    def create_build_event(self) -> dict:
        """Create a build CloudEvent"""
        return {
            "specversion": "1.0",
            "id": str(uuid.uuid4()),
            "source": "network.notifi.load-test",
            "type": "network.notifi.lambda.build.start",
            "time": datetime.now(timezone.utc).isoformat(),
            "data": {
                "third_party_id": "load-test-123",
                "parser_id": f"parser-{uuid.uuid4().hex[:8]}",
            },
            "datacontenttype": "application/json"
        }
    
    def create_parser_event(self) -> dict:
        """Create a parser CloudEvent"""
        return {
            "specversion": "1.0",
            "id": str(uuid.uuid4()),
            "source": "network.notifi.load-test",
            "subject": "parser-load-test",
            "type": "network.notifi.lambda.parser.start",
            "time": datetime.now(timezone.utc).isoformat(),
            "data": {
                "contextId": f"ctx-{uuid.uuid4().hex}",
                "parameters": {
                    "blockId": "999999",
                    "blockchainType": 52,
                    "urlForBlob": "redis://ephemeralblock/LoadTest/999999",
                    "logIndices": [0]
                }
            },
            "datacontenttype": "application/json"
        }
    
    async def run_load_test(self) -> LoadTestResult:
        """Execute the load test"""
        print(f"🚀 Starting load test...")
        print(f"   Target: {self.config.broker_url}")
        print(f"   Concurrent: {self.config.concurrent_requests}")
        print(f"   Duration: {self.config.duration_seconds}s")
        print(f"   Event Type: {self.config.event_type}")
        print("=" * 60)
        
        start_time = time.time()
        
        # Create aiohttp session
        connector = aiohttp.TCPConnector(limit=self.config.concurrent_requests + 50)
        async with aiohttp.ClientSession(connector=connector) as session:
            
            # Create event generator based on type
            if self.config.event_type == "build":
                event_factory = self.create_build_event
            elif self.config.event_type == "parser":
                event_factory = self.create_parser_event
            else:
                raise ValueError(f"Unknown event type: {self.config.event_type}")
            
            # Run load test with rampup
            tasks = []
            rampup_delay = self.config.rampup_seconds / self.config.concurrent_requests
            
            for i in range(self.config.concurrent_requests):
                # Rampup delay
                if i > 0:
                    await asyncio.sleep(rampup_delay)
                
                # Create continuous event sender
                task = asyncio.create_task(
                    self.continuous_sender(session, event_factory)
                )
                tasks.append(task)
            
            # Wait for test duration
            await asyncio.sleep(self.config.duration_seconds)
            
            # Cancel all tasks
            for task in tasks:
                task.cancel()
            
            # Wait for cancellations
            await asyncio.gather(*tasks, return_exceptions=True)
        
        duration = time.time() - start_time
        
        # Calculate statistics
        self.results.sort()
        result = LoadTestResult(
            total_requests=self.successful + self.failed,
            successful_requests=self.successful,
            failed_requests=self.failed,
            avg_latency_ms=sum(self.results) / len(self.results) if self.results else 0,
            p95_latency_ms=self.results[int(len(self.results) * 0.95)] if self.results else 0,
            p99_latency_ms=self.results[int(len(self.results) * 0.99)] if self.results else 0,
            requests_per_second=(self.successful + self.failed) / duration,
            duration_seconds=duration
        )
        
        # Print results
        print("=" * 60)
        print(f"📊 LOAD TEST RESULTS:")
        print(f"   Total Requests: {result.total_requests}")
        print(f"   ✅ Successful: {result.successful_requests}")
        print(f"   ❌ Failed: {result.failed_requests}")
        print(f"   Success Rate: {(result.successful_requests / result.total_requests * 100):.2f}%")
        print(f"   Avg Latency: {result.avg_latency_ms:.2f}ms")
        print(f"   P95 Latency: {result.p95_latency_ms:.2f}ms")
        print(f"   P99 Latency: {result.p99_latency_ms:.2f}ms")
        print(f"   Requests/sec: {result.requests_per_second:.2f}")
        print(f"   Duration: {result.duration_seconds:.2f}s")
        print("=" * 60)
        
        return result
    
    async def continuous_sender(self, session: aiohttp.ClientSession, event_factory):
        """Continuously send events until cancelled"""
        try:
            while True:
                event = event_factory()
                await self.send_event(session, event)
                await asyncio.sleep(0.1)  # Small delay between events
        except asyncio.CancelledError:
            pass

# Example usage
if __name__ == "__main__":
    import sys
    
    config = LoadTestConfig(
        broker_url=os.getenv("BROKER_URL", "http://0.0.0.0:8081"),
        concurrent_requests=int(os.getenv("CONCURRENT", "100")),
        duration_seconds=int(os.getenv("DURATION", "300")),
        event_type=os.getenv("EVENT_TYPE", "parser")
    )
    
    tester = AsyncLoadTester(config)
    result = asyncio.run(tester.run_load_test())
    
    # Exit with error if success rate < 95%
    success_rate = result.successful_requests / result.total_requests
    if success_rate < 0.95:
        print(f"❌ Load test failed: Success rate {success_rate:.2%} < 95%")
        sys.exit(1)
    else:
        print(f"✅ Load test passed: Success rate {success_rate:.2%}")
        sys.exit(0)
```

---

## 📊 Test Execution

### Prerequisites
```bash
# 1. Install k6
brew install k6  # macOS
# or
curl https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz -L | tar xvz

# 2. Install Python dependencies
cd tests/load/python
uv sync

# 3. Ensure cluster access
kubectl get nodes

# 4. Port-forward broker
make pf-broker ENV=dev &
```

### Run Load Tests
```bash
# Run K6 load tests
make test-load-k6 ENV=dev

# Run Python async load tests
make test-load-python ENV=dev EVENT_TYPE=parser CONCURRENT=100 DURATION=300

# Run specific scenarios
make test-load-baseline ENV=dev
make test-load-spike ENV=dev
make test-load-sustained ENV=dev

# Run stress tests
make test-stress ENV=dev
```

---

## 🎯 Performance Targets (SLOs) | Metric | Target | Critical Threshold | |-------- | -------- | ------------------- | | Event Processing Latency (p95) | < 3s | < 10s | | Event Publishing Success Rate | > 99% | > 95% | | Requests per Second | > 100 | > 50 | | Build Job Creation Rate | > 50/sec | > 20/sec | | Service Autoscaling Time | < 30s | < 60s | | Cold Start Latency | < 10s | < 30s | | System Recovery Time | < 5min | < 15min | ---

## 🔍 Observability

### Load Test Dashboard
- Real-time metrics during load tests
- Success/failure rates
- Latency distributions (p50, p95, p99)
- Resource utilization (CPU, memory, network)
- Pod scaling events

### Alerting
- Alert if success rate < 95%
- Alert if p95 latency > 10s
- Alert if system doesn't recover within 15min
- Alert on resource exhaustion

---

## 🚀 CI/CD Integration

```yaml
# .github/workflows/load-tests.yml
name: Load Tests

on:
  schedule:
    - cron: '0 3 * * *'  # Nightly at 3 AM
  workflow_dispatch:
    inputs:
      environment:
        description: 'Test environment'
        required: true
        default: 'dev'

jobs:
  load-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install k6
        run: | curl https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz -L | tar xvz
          sudo mv k6 /usr/local/bin/
          
      - name: Setup kubeconfig
        run: | mkdir -p ~/.kube
          echo "${{ secrets.KUBECONFIG }}" > ~/.kube/config
          
      - name: Port-forward broker
        run: | make pf-broker ENV=${{ github.event.inputs.environment | | 'dev' }} &
          sleep 10
          
      - name: Run load tests
        run: make test-load-all ENV=${{ github.event.inputs.environment | | 'dev' }}
        
      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: load-test-results
          path: tests/load/results/
```

---

## 📚 Related Stories

- **QA-001:** E2E CloudEvents Testing (validates correctness)
- **BACKEND-004:** Async Job Processing (provides async handling)
- **BACKEND-005:** Rate Limiting (provides backpressure)
- **DEVOPS-001:** Observability (provides metrics)
- **SRE-001:** Capacity Planning (uses load test data)

---

## 📝 Notes

- Load tests should run in **isolated environment** (dev/staging)
- **Do NOT** run load tests in production
- Monitor cluster resources during load tests
- Coordinate with SRE team for large-scale tests
- Load test results inform capacity planning

