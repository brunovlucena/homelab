"""
Command and Control Center Backend for Knative Lambda Operator

FastAPI backend for uploading files to MinIO for agents and lambdas.
Uses presigned URLs for direct uploads to MinIO.
"""
import os
import uuid
from datetime import timedelta
from typing import Optional

from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from minio import Minio
from minio.error import S3Error
from kubernetes import client, config
from kubernetes.client.rest import ApiException

app = FastAPI(
    title="Knative Lambda Operator C2",
    description="Command and Control Center for uploading files to MinIO",
    version="1.0.0"
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, restrict to frontend domain
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# MinIO client initialization
minio_client = Minio(
    os.getenv("MINIO_ENDPOINT", "minio.minio.svc.cluster.local:9000"),
    access_key=os.getenv("MINIO_ACCESS_KEY"),
    secret_key=os.getenv("MINIO_SECRET_KEY"),
    secure=False  # Set to True if using HTTPS
)

# Bucket names
LAMBDA_BUCKET = os.getenv("LAMBDA_BUCKET", "lambda-functions")
AGENT_BUCKET = os.getenv("AGENT_BUCKET", "agent-files")

# Kubernetes client initialization
try:
    # Try in-cluster config first (when running in K8s)
    config.load_incluster_config()
except:
    try:
        # Fall back to kubeconfig (for local development)
        config.load_kube_config()
    except:
        pass  # Kubernetes not available

custom_api = client.CustomObjectsApi()


# Request/Response models
class PresignedURLRequest(BaseModel):
    filename: str = Field(..., description="Name of the file to upload")
    mimeType: str = Field(..., description="MIME type of the file")
    size: int = Field(..., description="Size of the file in bytes")
    target: str = Field(..., description="Target: 'lambda' or 'agent'")
    path: Optional[str] = Field(None, description="Optional path within the bucket (e.g., 'my-function/')")


class PresignedURLResponse(BaseModel):
    uploadUrl: str
    fileId: str
    expiresIn: int
    objectPath: str


class UploadCompleteRequest(BaseModel):
    fileId: str
    objectPath: str


class FileListResponse(BaseModel):
    files: list[dict]


class ResourceItem(BaseModel):
    name: str
    namespace: str
    status: Optional[dict] = None
    labels: Optional[dict] = None


class ResourceListResponse(BaseModel):
    items: list[ResourceItem]
    total: int


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy"}


@app.get("/api/v1/lambdas", response_model=ResourceListResponse)
async def list_lambda_functions(namespace: Optional[str] = None):
    """
    List all LambdaFunctions from Kubernetes.
    
    Args:
        namespace: Optional namespace filter. If not provided, lists all namespaces.
    """
    try:
        group = "lambda.knative.io"
        version = "v1alpha1"
        plural = "lambdafunctions"
        
        if namespace:
            response = custom_api.list_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural
            )
        else:
            response = custom_api.list_cluster_custom_object(
                group=group,
                version=version,
                plural=plural
            )
        
        items = []
        for item in response.get("items", []):
            metadata = item.get("metadata", {})
            items.append(ResourceItem(
                name=metadata.get("name", ""),
                namespace=metadata.get("namespace", ""),
                status=item.get("status"),
                labels=metadata.get("labels", {})
            ))
        
        return ResourceListResponse(items=items, total=len(items))
    except ApiException as e:
        raise HTTPException(status_code=e.status, detail=f"Kubernetes API error: {e.reason}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list LambdaFunctions: {str(e)}")


@app.get("/api/v1/agents", response_model=ResourceListResponse)
async def list_lambda_agents(namespace: Optional[str] = None):
    """
    List all LambdaAgents from Kubernetes.
    
    Args:
        namespace: Optional namespace filter. If not provided, lists all namespaces.
    """
    try:
        group = "lambda.knative.io"
        version = "v1alpha1"
        plural = "lambdaagents"
        
        if namespace:
            response = custom_api.list_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural
            )
        else:
            response = custom_api.list_cluster_custom_object(
                group=group,
                version=version,
                plural=plural
            )
        
        items = []
        for item in response.get("items", []):
            metadata = item.get("metadata", {})
            items.append(ResourceItem(
                name=metadata.get("name", ""),
                namespace=metadata.get("namespace", ""),
                status=item.get("status"),
                labels=metadata.get("labels", {})
            ))
        
        return ResourceListResponse(items=items, total=len(items))
    except ApiException as e:
        raise HTTPException(status_code=e.status, detail=f"Kubernetes API error: {e.reason}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list LambdaAgents: {str(e)}")


@app.post("/api/v1/files/presigned-url", response_model=PresignedURLResponse)
async def generate_presigned_url(request: PresignedURLRequest):
    """
    Generate a presigned URL for direct upload to MinIO.
    
    Flow:
    1. Client requests presigned URL
    2. Client uploads directly to MinIO using presigned URL
    3. Client calls /complete endpoint to notify backend
    """
    # Validate file size (max 100 MB)
    max_size = 100 * 1024 * 1024  # 100 MB
    if request.size > max_size:
        raise HTTPException(status_code=400, detail=f"File too large (max {max_size / 1024 / 1024} MB)")
    
    # Validate target
    if request.target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    # Select bucket
    bucket = LAMBDA_BUCKET if request.target == "lambda" else AGENT_BUCKET
    
    # Ensure bucket exists
    try:
        if not minio_client.bucket_exists(bucket):
            minio_client.make_bucket(bucket)
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to access bucket: {str(e)}")
    
    # Generate unique file ID
    file_id = str(uuid.uuid4())
    
    # Construct object path
    if request.path:
        # Use provided path, ensure it ends with /
        path = request.path.rstrip("/") + "/"
        object_name = f"{path}{file_id}/{request.filename}"
    else:
        # Use default structure: {target}/{file_id}/{filename}
        object_name = f"{request.target}/{file_id}/{request.filename}"
    
    try:
        # Generate presigned URL (5 min expiry)
        presigned_url = minio_client.presigned_put_object(
            bucket,
            object_name,
            expires=timedelta(minutes=5)
        )
        
        return PresignedURLResponse(
            uploadUrl=presigned_url,
            fileId=file_id,
            expiresIn=300,  # 5 minutes
            objectPath=object_name
        )
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to generate presigned URL: {str(e)}")


@app.post("/api/v1/files/{file_id}/complete")
async def complete_upload(file_id: str, request: UploadCompleteRequest):
    """
    Notify backend that file upload is complete.
    This endpoint verifies the file exists in MinIO.
    """
    # Determine bucket from object path
    bucket = LAMBDA_BUCKET if request.objectPath.startswith("lambda/") else AGENT_BUCKET
    
    try:
        # Verify file exists
        stat = minio_client.stat_object(bucket, request.objectPath)
        
        return {
            "fileId": file_id,
            "status": "uploaded",
            "size": stat.size,
            "objectPath": request.objectPath,
            "bucket": bucket,
            "minioUrl": f"s3://{bucket}/{request.objectPath}"
        }
    except S3Error as e:
        raise HTTPException(status_code=404, detail=f"File not found: {str(e)}")


@app.get("/api/v1/files/list")
async def list_files(target: str = "lambda", prefix: Optional[str] = None):
    """
    List files in MinIO bucket.
    
    Args:
        target: 'lambda' or 'agent'
        prefix: Optional prefix to filter files (e.g., 'my-function/')
    """
    if target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    bucket = LAMBDA_BUCKET if target == "lambda" else AGENT_BUCKET
    
    try:
        objects = minio_client.list_objects(bucket, prefix=prefix or "", recursive=True)
        
        files = []
        for obj in objects:
            files.append({
                "name": obj.object_name,
                "size": obj.size,
                "lastModified": obj.last_modified.isoformat() if obj.last_modified else None,
                "etag": obj.etag
            })
        
        return FileListResponse(files=files)
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to list files: {str(e)}")


@app.get("/api/v1/files/{target}/{path:path}")
async def get_file(target: str, path: str):
    """
    Get file contents from MinIO.
    
    Args:
        target: 'lambda' or 'agent'
        path: Path to the file in the bucket
    """
    if target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    bucket = LAMBDA_BUCKET if target == "lambda" else AGENT_BUCKET
    
    try:
        # Get file object
        response = minio_client.get_object(bucket, path)
        content = response.read()
        response.close()
        response.release_conn()
        
        # Try to decode as text (for code files)
        try:
            text_content = content.decode('utf-8')
            return {
                "path": path,
                "bucket": bucket,
                "content": text_content,
                "isText": True,
                "size": len(content)
            }
        except UnicodeDecodeError:
            # Binary file - return as base64
            import base64
            return {
                "path": path,
                "bucket": bucket,
                "content": base64.b64encode(content).decode('utf-8'),
                "isText": False,
                "size": len(content)
            }
    except S3Error as e:
        raise HTTPException(status_code=404, detail=f"File not found: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to read file: {str(e)}")


@app.put("/api/v1/files/{target}/{path:path}")
async def update_file(target: str, path: str, request: dict):
    """
    Update file contents in MinIO.
    
    Args:
        target: 'lambda' or 'agent'
        path: Path to the file in the bucket
        request: JSON body with 'content' field
    """
    if target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    if "content" not in request:
        raise HTTPException(status_code=400, detail="Missing 'content' field in request body")
    
    bucket = LAMBDA_BUCKET if target == "lambda" else AGENT_BUCKET
    
    try:
        # Check if file exists
        try:
            minio_client.stat_object(bucket, path)
        except S3Error:
            raise HTTPException(status_code=404, detail="File not found")
        
        # Handle text vs binary content
        if request.get("isText", True):
            content_bytes = request["content"].encode('utf-8')
        else:
            import base64
            content_bytes = base64.b64decode(request["content"])
        
        # Upload updated content
        from io import BytesIO
        from minio.commonconfig import CopySource
        
        file_data = BytesIO(content_bytes)
        minio_client.put_object(
            bucket,
            path,
            file_data,
            length=len(content_bytes),
            content_type=request.get("contentType", "text/plain")
        )
        
        return {
            "status": "updated",
            "path": path,
            "bucket": bucket,
            "size": len(content_bytes)
        }
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to update file: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to update file: {str(e)}")


@app.patch("/api/v1/files/{target}/{path:path}")
async def rename_file(target: str, path: str, request: dict):
    """
    Rename/move a file in MinIO.
    
    Args:
        target: 'lambda' or 'agent'
        path: Current path to the file
        request: JSON body with 'newPath' field
    """
    if target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    if "newPath" not in request:
        raise HTTPException(status_code=400, detail="Missing 'newPath' field in request body")
    
    bucket = LAMBDA_BUCKET if target == "lambda" else AGENT_BUCKET
    new_path = request["newPath"]
    
    try:
        from minio.commonconfig import CopySource
        
        # Copy file to new location
        copy_source = CopySource(bucket, path)
        minio_client.copy_object(bucket, new_path, copy_source)
        
        # Delete old file
        minio_client.remove_object(bucket, path)
        
        return {
            "status": "renamed",
            "oldPath": path,
            "newPath": new_path,
            "bucket": bucket
        }
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to rename file: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to rename file: {str(e)}")


@app.post("/api/v1/files/{target}/{path:path}/copy")
async def copy_file(target: str, path: str, request: dict):
    """
    Copy a file to a new location.
    
    Args:
        target: 'lambda' or 'agent'
        path: Current path to the file
        request: JSON body with 'destinationPath' field
    """
    if target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    if "destinationPath" not in request:
        raise HTTPException(status_code=400, detail="Missing 'destinationPath' field in request body")
    
    bucket = LAMBDA_BUCKET if target == "lambda" else AGENT_BUCKET
    destination = request["destinationPath"]
    
    try:
        from minio.commonconfig import CopySource
        
        copy_source = CopySource(bucket, path)
        minio_client.copy_object(bucket, destination, copy_source)
        
        return {
            "status": "copied",
            "sourcePath": path,
            "destinationPath": destination,
            "bucket": bucket
        }
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to copy file: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to copy file: {str(e)}")


@app.delete("/api/v1/files/{target}/{path:path}")
async def delete_file(target: str, path: str):
    """
    Delete a file from MinIO.
    
    Args:
        target: 'lambda' or 'agent'
        path: Path to the file in the bucket
    """
    if target not in ["lambda", "agent"]:
        raise HTTPException(status_code=400, detail="Target must be 'lambda' or 'agent'")
    
    bucket = LAMBDA_BUCKET if target == "lambda" else AGENT_BUCKET
    
    try:
        minio_client.remove_object(bucket, path)
        return {"status": "deleted", "path": path, "bucket": bucket}
    except S3Error as e:
        raise HTTPException(status_code=500, detail=f"Failed to delete file: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
