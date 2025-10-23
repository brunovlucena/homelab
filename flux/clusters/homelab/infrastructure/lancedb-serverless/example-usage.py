"""
🔍 Example Usage of LanceDB Serverless API

This script demonstrates how to use the LanceDB Serverless API
for vector storage and search operations.
"""

import requests
import numpy as np
from typing import List

# Configuration
SERVICE_URL = "http://lancedb-serverless.lancedb-serverless.svc.cluster.local"
# Or use the external URL from: kubectl get ksvc -n lancedb-serverless


class LanceDBClient:
    """Simple client for LanceDB Serverless API"""
    
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')
    
    def health_check(self):
        """Check if the service is healthy"""
        response = requests.get(f"{self.base_url}/health")
        response.raise_for_status()
        return response.json()
    
    def list_tables(self):
        """List all tables"""
        response = requests.get(f"{self.base_url}/tables")
        response.raise_for_status()
        return response.json()
    
    def create_table(self, name: str, data: List[dict], mode: str = "overwrite"):
        """Create a new table with data"""
        payload = {
            "name": name,
            "data": data,
            "mode": mode
        }
        response = requests.post(f"{self.base_url}/tables", json=payload)
        response.raise_for_status()
        return response.json()
    
    def search(self, table: str, query_vector: List[float], limit: int = 10, filter_expr: str = None):
        """Perform vector search"""
        payload = {
            "table": table,
            "query_vector": query_vector,
            "limit": limit
        }
        if filter_expr:
            payload["filter"] = filter_expr
        
        response = requests.post(f"{self.base_url}/search", json=payload)
        response.raise_for_status()
        return response.json()
    
    def delete_table(self, table: str):
        """Delete a table"""
        response = requests.delete(f"{self.base_url}/tables/{table}")
        response.raise_for_status()
        return response.json()


def example_document_search():
    """Example: Document search with embeddings"""
    print("📚 Example: Document Search")
    print("=" * 60)
    
    client = LanceDBClient(SERVICE_URL)
    
    # 1. Check health
    print("\n1️⃣ Health Check")
    health = client.health_check()
    print(f"   Status: {health['status']}")
    
    # 2. Create a table with document embeddings
    print("\n2️⃣ Creating 'documents' table")
    
    # Simulated embeddings (in reality, use a model like OpenAI, Sentence Transformers, etc.)
    documents = [
        {
            "id": 1,
            "vector": np.random.rand(384).tolist(),
            "text": "Introduction to Machine Learning",
            "category": "tech",
            "date": "2025-01-15"
        },
        {
            "id": 2,
            "vector": np.random.rand(384).tolist(),
            "text": "Deep Learning Fundamentals",
            "category": "tech",
            "date": "2025-01-20"
        },
        {
            "id": 3,
            "vector": np.random.rand(384).tolist(),
            "text": "Climate Change Impact",
            "category": "science",
            "date": "2025-01-18"
        },
        {
            "id": 4,
            "vector": np.random.rand(384).tolist(),
            "text": "Quantum Computing Basics",
            "category": "tech",
            "date": "2025-01-22"
        },
    ]
    
    result = client.create_table("documents", documents)
    print(f"   Created table with {result['rows']} documents")
    
    # 3. List tables
    print("\n3️⃣ Listing all tables")
    tables = client.list_tables()
    print(f"   Tables: {tables['tables']}")
    
    # 4. Search for similar documents
    print("\n4️⃣ Searching for similar documents")
    query_vector = np.random.rand(384).tolist()
    
    results = client.search("documents", query_vector, limit=3)
    print(f"   Found {results['count']} results:")
    for i, doc in enumerate(results['results'], 1):
        print(f"   {i}. {doc.get('text', 'N/A')} (category: {doc.get('category', 'N/A')})")
    
    # 5. Search with filter
    print("\n5️⃣ Searching with filter (category='tech')")
    results = client.search("documents", query_vector, limit=3, filter_expr="category = 'tech'")
    print(f"   Found {results['count']} tech documents:")
    for i, doc in enumerate(results['results'], 1):
        print(f"   {i}. {doc.get('text', 'N/A')}")
    
    print("\n" + "=" * 60)
    print("✅ Example completed successfully!")


def example_image_similarity():
    """Example: Image similarity search"""
    print("\n🖼️ Example: Image Similarity Search")
    print("=" * 60)
    
    client = LanceDBClient(SERVICE_URL)
    
    # 1. Create a table with image embeddings
    print("\n1️⃣ Creating 'images' table")
    
    images = [
        {
            "id": "img_001",
            "vector": np.random.rand(512).tolist(),
            "filename": "cat_001.jpg",
            "label": "cat",
            "timestamp": "2025-01-15T10:00:00"
        },
        {
            "id": "img_002",
            "vector": np.random.rand(512).tolist(),
            "filename": "dog_001.jpg",
            "label": "dog",
            "timestamp": "2025-01-15T11:00:00"
        },
        {
            "id": "img_003",
            "vector": np.random.rand(512).tolist(),
            "filename": "cat_002.jpg",
            "label": "cat",
            "timestamp": "2025-01-15T12:00:00"
        },
    ]
    
    result = client.create_table("images", images)
    print(f"   Created table with {result['rows']} images")
    
    # 2. Search for similar images
    print("\n2️⃣ Searching for similar images")
    query_vector = np.random.rand(512).tolist()
    
    results = client.search("images", query_vector, limit=2)
    print(f"   Found {results['count']} similar images:")
    for i, img in enumerate(results['results'], 1):
        print(f"   {i}. {img.get('filename', 'N/A')} - {img.get('label', 'N/A')}")
    
    print("\n" + "=" * 60)
    print("✅ Example completed successfully!")


def example_cleanup():
    """Cleanup: Delete example tables"""
    print("\n🗑️ Cleanup: Deleting example tables")
    print("=" * 60)
    
    client = LanceDBClient(SERVICE_URL)
    
    tables_to_delete = ["documents", "images"]
    
    for table in tables_to_delete:
        try:
            result = client.delete_table(table)
            print(f"   ✅ Deleted table: {table}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print(f"   ⚠️ Table '{table}' not found, skipping...")
            else:
                raise
    
    print("\n" + "=" * 60)
    print("✅ Cleanup completed!")


if __name__ == "__main__":
    try:
        # Run examples
        example_document_search()
        example_image_similarity()
        
        # Uncomment to cleanup after running
        # example_cleanup()
        
    except requests.exceptions.RequestException as e:
        print(f"\n❌ Error: {e}")
        print("\nMake sure the service is running and accessible.")
        print("Get the service URL with: kubectl get ksvc -n lancedb-serverless")

