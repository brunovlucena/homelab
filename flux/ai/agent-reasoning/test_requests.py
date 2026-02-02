#!/usr/bin/env python3
"""
Interactive test script for Agent-Reasoning service.

Run this to test the service with various requests.
"""
import requests
import json
import sys
from typing import Dict, Any

BASE_URL = "http://localhost:8080"


def print_response(title: str, response: requests.Response):
    """Print formatted response."""
    print(f"\n{'='*60}")
    print(f"  {title}")
    print(f"{'='*60}")
    print(f"Status: {response.status_code}")
    try:
        data = response.json()
        print(json.dumps(data, indent=2))
    except:
        print(response.text)
    print()


def test_health():
    """Test health endpoint."""
    print("Testing /health endpoint...")
    response = requests.get(f"{BASE_URL}/health")
    print_response("Health Check", response)
    return response.status_code == 200


def test_root():
    """Test root endpoint."""
    print("Testing / endpoint...")
    response = requests.get(f"{BASE_URL}/")
    print_response("Root Endpoint", response)
    return response.status_code == 200


def test_reasoning(question: str, task_type: str = "general", max_steps: int = 6):
    """Test reasoning endpoint."""
    print(f"Testing /reason endpoint with question: '{question[:50]}...'")
    
    payload = {
        "question": question,
        "context": {
            "test": True,
            "source": "local_test"
        },
        "max_steps": max_steps,
        "task_type": task_type,
    }
    
    response = requests.post(
        f"{BASE_URL}/reason",
        json=payload,
        headers={"Content-Type": "application/json"}
    )
    
    print_response("Reasoning Response", response)
    return response.status_code == 200


def test_metrics():
    """Test metrics endpoint."""
    print("Testing /metrics endpoint...")
    response = requests.get(f"{BASE_URL}/metrics")
    print(f"\n{'='*60}")
    print("  Metrics (first 30 lines)")
    print(f"{'='*60}")
    lines = response.text.split('\n')[:30]
    for line in lines:
        print(line)
    print("...")
    return response.status_code == 200


def main():
    """Run all tests."""
    print("🧪 Agent-Reasoning Local Testing")
    print("=" * 60)
    
    # Check if service is running
    try:
        requests.get(f"{BASE_URL}/health", timeout=2)
    except requests.exceptions.ConnectionError:
        print("❌ Error: Service is not running!")
        print(f"   Start it with: docker run -p 8080:8080 agent-reasoning:latest")
        sys.exit(1)
    
    # Run tests
    tests = [
        ("Health Check", test_health),
        ("Root Endpoint", test_root),
        ("Metrics", test_metrics),
    ]
    
    for name, test_func in tests:
        try:
            success = test_func()
            if not success:
                print(f"⚠️  {name} returned non-200 status")
        except Exception as e:
            print(f"❌ {name} failed: {e}")
    
    # Test reasoning with different scenarios
    print("\n" + "="*60)
    print("  Reasoning Tests")
    print("="*60)
    
    reasoning_tests = [
        {
            "question": "How should I optimize my Kubernetes cluster?",
            "task_type": "optimization",
            "max_steps": 6
        },
        {
            "question": "Why is my service slow?",
            "task_type": "troubleshooting",
            "max_steps": 4
        },
        {
            "question": "How should I deploy this application?",
            "task_type": "planning",
            "max_steps": 6
        },
    ]
    
    for i, test in enumerate(reasoning_tests, 1):
        print(f"\n--- Test {i} ---")
        try:
            test_reasoning(**test)
        except Exception as e:
            print(f"❌ Reasoning test {i} failed: {e}")
    
    print("\n✅ All tests completed!")
    print("\nTo test with custom questions, modify this script or use curl:")
    print('  curl -X POST http://localhost:8080/reason \\')
    print('    -H "Content-Type: application/json" \\')
    print('    -d \'{"question": "Your question here", "max_steps": 6}\'')


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n⚠️  Tests interrupted by user")
        sys.exit(1)


