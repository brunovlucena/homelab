# 🌐 BACKEND-002: Build Context Management

**Priority**: P0 | **Status**: ✅ Implemented  | **Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-215/backend-002-build-context-management

---

## 📋 User Story

**As a** Backend Developer  
**I want to** create and manage build contexts from S3 source code  
**So that** Kaniko can build Docker images with the correct code and dependencies

---

## 🎯 Acceptance Criteria

### ✅ S3 Source Management
- [ ] Download parser code from S3 bucket (`s3://knative-lambda-{env}-fusion-modules-tmp/global/parser/{parser_id}`)
- [ ] Support multiple source file formats (Python, Node.js, Go)
- [ ] Validate source file exists before build
- [ ] Handle S3 access errors gracefully
- [ ] Support content-based hashing for unique image tags

### ✅ Build Context Creation
- [ ] Generate appropriate Dockerfile based on runtime
- [ ] Package source code into tar.gz archive
- [ ] Upload build context to S3 temp bucket
- [ ] Generate unique build context key: `build-context/{parser_id}/context.tar.gz`
- [ ] Set proper S3 permissions for Kaniko access
- [ ] Clean up old build contexts (TTL-based)

### ✅ Dockerfile Generation
- [ ] Support Node.js runtime (default: nodejs20)
- [ ] Support Python runtime (default: python3.11)
- [ ] Support Go runtime (default: go1.21)
- [ ] Include runtime-specific dependencies
- [ ] Configure npm/pip/go modules properly
- [ ] Optimize for layer caching

### ✅ Content Hashing
- [ ] Generate SHA-256 hash of source content
- [ ] Use content hash for Docker image tagging
- [ ] Track content hash in BuildRequest
- [ ] Enable content-addressable builds
- [ ] Prevent duplicate builds for same content

### ✅ Error Handling
- [ ] Handle S3 access denied errors
- [ ] Handle source file not found errors
- [ ] Handle Dockerfile generation failures
- [ ] Handle tar.gz compression failures
- [ ] Provide detailed error messages for debugging

---

## 🔧 Technical Implementation

### File: `internal/handler/build_context_manager.go`

```go
// Build Context Manager Interface
type BuildContextManager interface {
    CreateBuildContext(ctx context.Context, buildRequest *builds.BuildRequest) (string, error)
}

// Implementation
func (b *BuildContextManagerImpl) CreateBuildContext(ctx context.Context, buildRequest *builds.BuildRequest) (string, error) {
    ctx, span := b.obs.StartSpan(ctx, "create_build_context")
    defer span.End()
    
    // 1. Download source from S3
    sourceCode, err := b.downloadSourceFromS3(ctx, buildRequest)
    if err != nil {
        return "", fmt.Errorf("failed to download source: %w", err)
    }
    
    // 2. Generate content hash for unique tagging
    contentHash := b.generateContentHash(sourceCode)
    buildRequest.ContentHash = contentHash
    
    // 3. Generate Dockerfile based on runtime
    dockerfile, err := b.generateDockerfile(buildRequest)
    if err != nil {
        return "", fmt.Errorf("failed to generate dockerfile: %w", err)
    }
    
    // 4. Create tar.gz archive with source + Dockerfile
    tarGz, err := b.createTarGz(sourceCode, dockerfile)
    if err != nil {
        return "", fmt.Errorf("failed to create tar.gz: %w", err)
    }
    
    // 5. Upload build context to S3
    buildContextKey := fmt.Sprintf("build-context/%s/context.tar.gz", buildRequest.ParserID)
    err = b.uploadBuildContext(ctx, buildContextKey, tarGz)
    if err != nil {
        return "", fmt.Errorf("failed to upload build context: %w", err)
    }
    
    b.obs.Info(ctx, "Build context created successfully",
        "parser_id", buildRequest.ParserID,
        "content_hash", contentHash,
        "build_context_key", buildContextKey)
    
    return buildContextKey, nil
}
```

### Dockerfile Templates

#### Node.js Runtime
```dockerfile
ARG NODE_BASE_IMAGE=node:20-alpine
FROM ${NODE_BASE_IMAGE}

WORKDIR /app

# Copy parser code
COPY parser.js .
COPY package.json .

# Install dependencies with retry logic
RUN npm install --production

# Set environment
ENV NODE_ENV=production
ENV HTTP_PORT=8080

# Expose port
EXPOSE 8080

# Run parser
CMD ["node", "parser.js"]
```

#### Python Runtime
```dockerfile
ARG PYTHON_BASE_IMAGE=python:3.11-slim
FROM ${PYTHON_BASE_IMAGE}

WORKDIR /app

# Copy parser code
COPY parser.py .
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Set environment
ENV PYTHONUNBUFFERED=1
ENV HTTP_PORT=8080

# Expose port
EXPOSE 8080

# Run parser
CMD ["python", "parser.py"]
```

---

## 📊 Build Context Structure

```
build-context/
└── {parser_id}/
    └── context.tar.gz
        ├── Dockerfile
        ├── parser.{js | py | go}
        ├── package.json (Node.js)
        ├── requirements.txt (Python)
        └── go.mod (Go)
```

---

## 🧪 Testing Scenarios

### 1. Node.js Parser Build Context
```bash
# Upload parser to S3
aws s3 cp parser.js \
  s3://knative-lambda-fusion-modules-tmp/global/parser/test-parser-123

# Trigger build
make trigger-build-dev PARSER_ID=test-parser-123
```

**Expected**:
- Build context created at `s3://knative-lambda-fusion-modules-tmp/build-context/test-parser-123/context.tar.gz`
- Content hash generated and tracked
- Kaniko job created with correct context

### 2. Python Parser Build Context
```bash
# Upload parser to S3
aws s3 cp parser.py \
  s3://knative-lambda-fusion-modules-tmp/global/parser/test-parser-456

# Trigger build
make trigger-build-dev PARSER_ID=test-parser-456
```

**Expected**:
- Python Dockerfile generated
- Dependencies installed via pip
- Build completes successfully

### 3. Source File Not Found
```bash
# Trigger build for non-existent parser
make trigger-build-dev PARSER_ID=non-existent-parser
```

**Expected**:
- Error logged: "source file not found"
- Build job not created
- Clear error message returned

---

## 📈 Performance Requirements

- **Context Creation**: < 5s for typical parser (< 1MB)
- **S3 Download**: < 2s for source code download
- **Tar.gz Creation**: < 1s for compression
- **S3 Upload**: < 2s for context upload
- **Total Latency**: < 10s end-to-end

---

## 🔍 Monitoring & Alerts

### Metrics
- `build_context_creation_total` - Total context creations
- `build_context_creation_duration_seconds` - Creation latency
- `build_context_s3_errors_total` - S3 access errors
- `build_context_size_bytes` - Context archive size

### Alerts
- **Context Creation Failures**: Alert if > 5% failure rate
- **S3 Access Errors**: Alert on any S3 access denied errors
- **Slow Context Creation**: Alert if p95 > 15s

---

## 🏗️ Code References

**Main Files**:
- `internal/handler/build_context_manager.go` - Build context creation
- `internal/aws/client.go` - S3 client operations
- `internal/templates/templates.go` - Dockerfile templates
- `pkg/builds/request.go` - Build request structures

**Configuration**:
- `internal/config/aws.go` - S3 bucket configuration
- `internal/config/build.go` - Runtime and build settings

---

## 📚 Related Documentation

- [BACKEND-001: CloudEvents Processing](BACKEND-001-cloudevents-processing.md)
- [BACKEND-003: Kubernetes Job Lifecycle](BACKEND-003-kubernetes-job-lifecycle.md)
- AWS S3 SDK: https://aws.github.io/aws-sdk-go-v2/docs/
- Kaniko Documentation: https://github.com/GoogleContainerTools/kaniko

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: Production Ready

