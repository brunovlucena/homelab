#!/usr/bin/env python3
"""
Quick test script for the C2 backend API
"""
import requests
import json

BASE_URL = "http://localhost:8080"

def test_health():
    """Test health endpoint"""
    print("🏥 Testing health endpoint...")
    response = requests.get(f"{BASE_URL}/health")
    print(f"   Status: {response.status_code}")
    print(f"   Response: {response.json()}")
    return response.status_code == 200

def test_presigned_url():
    """Test presigned URL generation"""
    print("\n📤 Testing presigned URL generation...")
    payload = {
        "filename": "test-runbook.md",
        "mimeType": "text/markdown",
        "size": 1024,
        "target": "agent",
        "path": "agent-bruno"
    }
    response = requests.post(
        f"{BASE_URL}/api/v1/files/presigned-url",
        json=payload
    )
    print(f"   Status: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"   ✅ Presigned URL generated")
        print(f"   File ID: {data['fileId']}")
        print(f"   Object Path: {data['objectPath']}")
        print(f"   Expires In: {data['expiresIn']}s")
        return data
    else:
        print(f"   ❌ Error: {response.text}")
        return None

def test_list_files():
    """Test file listing"""
    print("\n📋 Testing file list...")
    response = requests.get(f"{BASE_URL}/api/v1/files/list?target=agent&prefix=agent-bruno/")
    print(f"   Status: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"   ✅ Found {len(data['files'])} files")
        for file in data['files'][:5]:  # Show first 5
            print(f"      - {file['name']} ({file['size']} bytes)")
        return True
    else:
        print(f"   ❌ Error: {response.text}")
        return False

if __name__ == "__main__":
    print("=" * 60)
    print("🧪 Testing C2 Backend API")
    print("=" * 60)
    
    try:
        # Test health
        if not test_health():
            print("\n❌ Health check failed. Is the server running?")
            exit(1)
        
        # Test presigned URL
        presigned_data = test_presigned_url()
        
        # Test file list
        test_list_files()
        
        print("\n" + "=" * 60)
        print("✅ All tests completed!")
        print("=" * 60)
        
    except requests.exceptions.ConnectionError:
        print("\n❌ Cannot connect to server. Make sure it's running:")
        print("   cd src/c2-backend")
        print("   uvicorn main:app --reload --host 0.0.0.0 --port 8080")
    except Exception as e:
        print(f"\n❌ Error: {e}")
