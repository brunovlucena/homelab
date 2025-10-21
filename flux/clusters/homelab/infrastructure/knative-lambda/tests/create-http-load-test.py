#!/usr/bin/env python3
"""
HTTP Load Test for Knative Lambda Services
Sends concurrent HTTP requests directly to lambda services to trigger autoscaling
"""

import asyncio
import aiohttp
import time
import json
import uuid
from concurrent.futures import ThreadPoolExecutor
import os

async def send_http_request(session, url, payload, request_id):
    """Send a single HTTP request to a lambda service"""
    try:
        # Do not specify ce-* headers, else CloudEvent libraries assume binary mode instead of structured
        headers = {
            'Content-Type': 'application/cloudevents+json',
        }
        
        async with session.post(url, json=payload, headers=headers, timeout=30) as response:
            status = response.status
            text = await response.text()
            print(f"✅ Request {request_id}: Status {status}")
            return True, status, text
            
    except asyncio.TimeoutError:
        print(f"⏰ Request {request_id}: Timeout")
        return False, "timeout", "Request timed out"
    except Exception as e:
        print(f"❌ Request {request_id}: Error - {e}")
        return False, "error", str(e)

async def run_load_test(service_url, concurrent_requests=20, request_delay=0.1):
    """Run concurrent HTTP load test against a lambda service"""
    
    print(f"🚀 Starting HTTP load test against: {service_url}")
    print(f"📊 Concurrent requests: {concurrent_requests}")
    print(f"🎯 This should trigger scaling at 8+ concurrent requests per pod")
    print(f"⚡ Expected result: Multiple pods should spin up")
    print("=" * 60)
    
    # Create payload
    payload = {
        "contextId": "load-test-context-12345",
        "parameters": {
            "blockId": "999999",
            "blockchainType": 52,
            "urlForBlob": "redis://ephemeralblock/LoadTest/999999",
            "logIndices": [0]
        }
    }
    
    successful_requests = 0
    failed_requests = 0
    responses = []
    
    start_time = time.time()
    
    # Create aiohttp session with connection limits
    connector = aiohttp.TCPConnector(limit=concurrent_requests + 10)
    timeout = aiohttp.ClientTimeout(total=60)
    
    async with aiohttp.ClientSession(connector=connector, timeout=timeout) as session:
        # Create tasks for concurrent execution
        tasks = []
        for i in range(concurrent_requests):
            request_id = f"load-test-{i+1:03d}"
            task = send_http_request(session, service_url, payload, request_id)
            tasks.append(task)
            
            # Add small delay between task creation to spread out the load
            if i > 0 and i % 5 == 0:
                await asyncio.sleep(request_delay)
        
        # Wait for all requests to complete
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Process results
        for result in results:
            if isinstance(result, Exception):
                print(f"❌ Exception: {result}")
                failed_requests += 1
            else:
                success, status, text = result
                responses.append((success, status, text))
                if success:
                    successful_requests += 1
                else:
                    failed_requests += 1
    
    end_time = time.time()
    duration = end_time - start_time
    
    print("=" * 60)
    print(f"📈 HTTP LOAD TEST RESULTS:")
    print(f"   ✅ Successful requests: {successful_requests}")
    print(f"   ❌ Failed requests: {failed_requests}")
    print(f"   ⏱️  Total duration: {duration:.2f} seconds")
    print(f"   📊 Request rate: {concurrent_requests/duration:.1f} req/sec")
    print()
    print(f"🔍 MONITORING COMMANDS:")
    print(f"   kubectl get pods -n knative-lambda-dev -w")
    print(f"   kubectl get ksvc -n knative-lambda-dev")
    print(f"   kubectl get pa -n knative-lambda-dev")
    print("=" * 60)
    
    return successful_requests, failed_requests, duration

def get_service_url(env, third_party_id, parser_id):
    """Get the service URL for a lambda function"""
    service_name = f"lambda-{third_party_id[:16]}-{parser_id[:15]}"
    
    if env == "local":
        # For local development, assume port-forward
        return f"http://localhost:8080"
    else:
        # For cluster access, use cluster-internal URL
        namespace = f"knative-lambda-{env}"
        return f"http://{service_name}.{namespace}.svc.cluster.local"

async def main():
    """Main load test function"""
    # Get environment from ENV variable, default to dev
    env = os.getenv("ENV", "dev")
    
    # Get number of concurrent requests
    concurrent_requests = int(os.getenv("CONCURRENT_REQUESTS", "20"))
    
    # Get target service parameters
    third_party_id = "0307ea43639b4616b044d190310a26bd"
    parser_id = "0197ad6c10b973b2b854a0e652155b7e"  # Focus on first parser
    
    service_url = get_service_url(env, third_party_id, parser_id)
    
    print(f"Environment: {env}")
    print(f"Service URL: {service_url}")
    print(f"Concurrent requests: {concurrent_requests}")
    print()
    
    # Run the load test
    try:
        await run_load_test(service_url, concurrent_requests)
    except KeyboardInterrupt:
        print("\n🛑 Load test interrupted by user")
    except Exception as e:
        print(f"\n❌ Load test failed: {e}")

if __name__ == "__main__":
    # Run async main
    asyncio.run(main())
