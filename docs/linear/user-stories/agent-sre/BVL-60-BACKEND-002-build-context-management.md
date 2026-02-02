# 📦 BACKEND-002: Build Context Management

**Linear URL**: https://linear.app/bvlucena/issue/BVL-197/backend-002-build-context-management  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to manage build contexts for LambdaFunction remediation actions  
**So that** remediation functions can be built and deployed automatically


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Agent-sre can create build contexts from source code
- [ ] Support for multiple runtime types (Python, Node.js, Go)
- [ ] Generate appropriate Dockerfiles based on runtime
- [ ] Package source code into tar.gz archives
- [ ] Store build contexts in ConfigMaps or S3
- [ ] Content-based hashing for unique image tags
- [ ] Clean up old build contexts (TTL-based)
- [ ] Handle build context creation failures gracefully

---

## 🔧 Implementation Details

### Build Context Creation

```python
# src/sre_agent/build_context.py
from typing import Dict, Any, Optional
import hashlib
import tarfile
import io

class BuildContextManager:
    """Manage build contexts for LambdaFunction remediation."""
    
    def __init__(self, k8s_client):
        self.k8s_client = k8s_client
    
    async def create_build_context(
        self,
        lambda_function: str,
        source_code: str,
        runtime: str = "python"
    ) -> Dict[str, Any]:
        """
        Create build context for LambdaFunction.
        
        Args:
            lambda_function: Name of the LambdaFunction
            source_code: Source code content
            runtime: Runtime type (python, nodejs, go)
            
        Returns:
            Build context metadata
        """
        # Generate Dockerfile
        dockerfile = self._generate_dockerfile(runtime)
        
        # Create tar.gz archive
        archive = self._create_tar_gz(source_code, dockerfile, runtime)
        
        # Compute content hash
        content_hash = self._compute_hash(archive)
        
        # Store in ConfigMap
        configmap_name = f"{lambda_function}-build-context"
        await self._store_in_configmap(configmap_name, archive)
        
        return {
            "configmap_name": configmap_name,
            "content_hash": content_hash,
            "runtime": runtime
        }
    
    def _generate_dockerfile(self, runtime: str) -> str:
        """Generate Dockerfile based on runtime."""
        if runtime == "python":
            return """FROM python:3.11-slim
WORKDIR /app
COPY remediation.py .
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
CMD ["python", "remediation.py"]
"""
        elif runtime == "nodejs":
            return """FROM node:20-alpine
WORKDIR /app
COPY remediation.js .
COPY package.json .
RUN npm install --production
CMD ["node", "remediation.js"]
"""
        elif runtime == "go":
            return """FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY remediation.go .
RUN go build -o remediation remediation.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/remediation .
CMD ["./remediation"]
"""
        else:
            raise ValueError(f"Unsupported runtime: {runtime}")
    
    def _create_tar_gz(
        self,
        source_code: str,
        dockerfile: str,
        runtime: str
    ) -> bytes:
        """Create tar.gz archive with source and Dockerfile."""
        buffer = io.BytesIO()
        
        with tarfile.open(fileobj=buffer, mode="w:gz") as tar:
            # Add source code
            source_filename = f"remediation.{self._get_extension(runtime)}"
            source_info = tarfile.TarInfo(name=source_filename)
            source_info.size = len(source_code.encode())
            tar.addfile(source_info, io.BytesIO(source_code.encode()))
            
            # Add Dockerfile
            dockerfile_info = tarfile.TarInfo(name="Dockerfile")
            dockerfile_info.size = len(dockerfile.encode())
            tar.addfile(dockerfile_info, io.BytesIO(dockerfile.encode()))
            
            # Add requirements/package.json if needed
            if runtime == "python":
                requirements = "requests\n"
                req_info = tarfile.TarInfo(name="requirements.txt")
                req_info.size = len(requirements.encode())
                tar.addfile(req_info, io.BytesIO(requirements.encode()))
            elif runtime == "nodejs":
                package_json = '{"name": "remediation", "version": "1.0.0", "dependencies": {}}\n'
                pkg_info = tarfile.TarInfo(name="package.json")
                pkg_info.size = len(package_json.encode())
                tar.addfile(pkg_info, io.BytesIO(package_json.encode()))
        
        buffer.seek(0)
        return buffer.read()
    
    def _compute_hash(self, data: bytes) -> str:
        """Compute SHA-256 hash of data."""
        return hashlib.sha256(data).hexdigest()
    
    def _get_extension(self, runtime: str) -> str:
        """Get file extension for runtime."""
        return {
            "python": "py",
            "nodejs": "js",
            "go": "go"
        }[runtime]
```

---

## 📚 Related Documentation

- [BACKEND-001: CloudEvents Processing](./BVL-59-BACKEND-001-cloudevents-processing.md)
- [Knative Lambda Operator Documentation](../../docs/knative/03-for-engineers/backend/README.md)

---

**Related Stories**:
- [SRE-001: Build Failure Investigation](./BVL-45-SRE-001-build-failure-investigation.md)


## 🧪 Test Scenarios

### Scenario 1: Python Build Context Creation
1. Create LambdaFunction with Python runtime
2. Trigger build context creation with Python source code
3. Verify build context created successfully
4. Verify Dockerfile generated correctly for Python
5. Verify source code packaged in tar.gz archive
6. Verify requirements.txt included
7. Verify content hash computed correctly
8. Verify build context stored in ConfigMap

### Scenario 2: Node.js Build Context Creation
1. Create LambdaFunction with Node.js runtime
2. Trigger build context creation with Node.js source code
3. Verify build context created successfully
4. Verify Dockerfile generated correctly for Node.js
5. Verify source code packaged in tar.gz archive
6. Verify package.json included
7. Verify content hash computed correctly
8. Verify build context stored in ConfigMap

### Scenario 3: Go Build Context Creation
1. Create LambdaFunction with Go runtime
2. Trigger build context creation with Go source code
3. Verify build context created successfully
4. Verify multi-stage Dockerfile generated correctly for Go
5. Verify source code packaged in tar.gz archive
6. Verify content hash computed correctly
7. Verify build context stored in ConfigMap

### Scenario 4: Build Context Content Hashing
1. Create build context with source code
2. Record content hash
3. Create build context with same source code
4. Verify same content hash generated (idempotency)
5. Modify source code slightly
6. Verify different content hash generated
7. Verify hash uniqueness for different contexts

### Scenario 5: Build Context Cleanup
1. Create multiple build contexts
2. Configure TTL for build contexts (7 days)
3. Wait for TTL to expire (simulate)
4. Trigger cleanup process
5. Verify old build contexts cleaned up
6. Verify active build contexts retained
7. Verify cleanup metrics recorded

### Scenario 6: Build Context Failure Handling
1. Trigger build context creation with invalid source code
2. Verify failure handled gracefully
3. Verify error logged with context
4. Verify no partial build context created
5. Verify retry logic works (if applicable)
6. Verify failure metrics recorded
7. Verify alert fires for repeated failures

### Scenario 7: Build Context High Load
1. Create 100+ build contexts simultaneously
2. Verify all build contexts created successfully
3. Verify no resource exhaustion
4. Verify build context creation performance acceptable (< 5 seconds per context)
5. Verify ConfigMap storage handles load
6. Verify cleanup works correctly under load

## 📊 Success Metrics

- **Build Context Creation Success Rate**: > 99%
- **Build Context Creation Time**: < 5 seconds per context (P95)
- **Content Hash Computation**: < 100ms per context (P95)
- **Build Context Storage**: < 1 second per context (P95)
- **Build Context Cleanup Rate**: 100% (old contexts cleaned)
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required