#!/usr/bin/env python3
"""
Upload the three-scales-framework diagram to MinIO.

This script uploads the generated diagram to MinIO storage for use in the blog post.
"""

import os
import sys
import subprocess
from pathlib import Path

def get_minio_credentials():
    """Get MinIO credentials from Kubernetes secret or environment variables."""
    # Try environment variables first
    access_key = os.getenv("MINIO_ACCESS_KEY")
    secret_key = os.getenv("MINIO_SECRET_KEY")
    
    if access_key and secret_key:
        return access_key, secret_key
    
    # Try to get from Kubernetes secret
    try:
        print("🔐 Attempting to get MinIO credentials from Kubernetes...")
        result = subprocess.run(
            ["kubectl", "get", "secret", "minio-credentials", "-n", "minio", 
             "-o", "jsonpath={.data.access-key}"],
            capture_output=True,
            text=True,
            check=True
        )
        access_key_b64 = result.stdout.strip()
        
        result = subprocess.run(
            ["kubectl", "get", "secret", "minio-credentials", "-n", "minio",
             "-o", "jsonpath={.data.secret-key}"],
            capture_output=True,
            text=True,
            check=True
        )
        secret_key_b64 = result.stdout.strip()
        
        # Decode base64
        import base64
        access_key = base64.b64decode(access_key_b64).decode('utf-8')
        secret_key = base64.b64decode(secret_key_b64).decode('utf-8')
        
        print("✅ Got credentials from Kubernetes secret")
        return access_key, secret_key
    except (subprocess.CalledProcessError, FileNotFoundError) as e:
        print(f"⚠️  Could not get credentials from Kubernetes: {e}")
        print("💡 Set MINIO_ACCESS_KEY and MINIO_SECRET_KEY environment variables")
        sys.exit(1)

def upload_to_minio(file_path, bucket="homepage-blog", object_name=None):
    """Upload a file to MinIO using the Python client."""
    try:
        from minio import Minio
        from minio.error import S3Error
    except ImportError:
        print("❌ minio package not installed. Install with: pip install minio")
        sys.exit(1)
    
    # Get credentials
    access_key, secret_key = get_minio_credentials()
    
    # MinIO endpoint
    endpoint = os.getenv("MINIO_ENDPOINT", "minio.minio.svc.cluster.local:9000")
    secure = os.getenv("MINIO_SECURE", "false").lower() == "true"
    
    # Initialize MinIO client
    client = Minio(
        endpoint,
        access_key=access_key,
        secret_key=secret_key,
        secure=secure
    )
    
    # Set object name if not provided
    if object_name is None:
        object_name = f"images/graphs/{Path(file_path).name}"
    
    # Check if file exists
    if not os.path.exists(file_path):
        print(f"❌ File not found: {file_path}")
        sys.exit(1)
    
    try:
        # Create bucket if it doesn't exist
        found = client.bucket_exists(bucket)
        if not found:
            print(f"📦 Creating bucket: {bucket}")
            client.make_bucket(bucket)
        else:
            print(f"✅ Bucket exists: {bucket}")
        
        # Upload file
        print(f"📤 Uploading {file_path} to {bucket}/{object_name}...")
        client.fput_object(
            bucket,
            object_name,
            file_path,
            content_type="image/png"
        )
        print(f"✅ Successfully uploaded to {bucket}/{object_name}")
        print(f"🌐 URL: http://{endpoint}/{bucket}/{object_name}")
        
        # Also upload SVG if it exists
        svg_path = file_path.replace('.png', '.svg')
        if os.path.exists(svg_path):
            svg_object_name = object_name.replace('.png', '.svg')
            print(f"📤 Uploading {svg_path} to {bucket}/{svg_object_name}...")
            client.fput_object(
                bucket,
                svg_object_name,
                svg_path,
                content_type="image/svg+xml"
            )
            print(f"✅ Successfully uploaded SVG to {bucket}/{svg_object_name}")
        
    except S3Error as e:
        print(f"❌ MinIO error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    # Default file path
    script_dir = Path(__file__).parent
    png_file = script_dir / "three-scales-framework.png"
    
    if len(sys.argv) > 1:
        png_file = Path(sys.argv[1])
    
    if not png_file.exists():
        print(f"❌ Image file not found: {png_file}")
        print("💡 Run generate-three-scales-diagram.py first to generate the image")
        sys.exit(1)
    
    upload_to_minio(str(png_file))

