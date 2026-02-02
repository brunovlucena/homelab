# 🔄 DEVOPS-010: Multi-Storage Backend Support

**Epic**: Storage Backend Flexibility  
**Priority**: P1 | **Status**: Ready for Development**Points**: 8  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-242/devops-010-multi-storage-backend-support

**Sprint**: v1.1.0  

---


## 📋 User Story

**As a** DevOps Engineer  
**I want to** multi-storage backend support  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] *AC1**: MinIO deploys via Helm chart
- [ ] *AC2**: S3 client auto-detects MinIO vs AWS S3 from endpoint
- [ ] *AC3**: All S3 operations work (upload, download, delete, exists)
- [ ] *AC4**: MinIO credentials configured via env vars
- [ ] *AC5**: Buckets created automatically or via make target
- [ ] *AC6**: Configuration supports `STORAGE_PROVIDER=minio`
- [ ] *Tests:**
- [ ] -namespace minio \
- [ ] -create-namespace \
- [ ] -set rootUser=minioadmin \

---


## 🎯 Goal

Enable knative-lambda to use multiple storage backends (AWS S3, MinIO local/remote, future GCS) with automatic selection based on environment, eliminating AWS dependency for local development.

---

## 📖 User Stories

### **Story 1: MinIO Support for Local Development**

**As a** developer  
**I want** to use MinIO as local S3-compatible storage  
**So that** I can develop and test without AWS dependencies

**Acceptance Criteria:**

✅ **AC1**: MinIO deploys via Helm chart  
✅ **AC2**: S3 client auto-detects MinIO vs AWS S3 from endpoint  
✅ **AC3**: All S3 operations work (upload, download, delete, exists)  
✅ **AC4**: MinIO credentials configured via env vars  
✅ **AC5**: Buckets created automatically or via make target  
✅ **AC6**: Configuration supports `STORAGE_PROVIDER=minio`  

**Tests:**

```bash
# Test 1: Deploy MinIO to cluster
test_minio_deployment() {
  # Deploy MinIO
  helm repo add minio https://charts.min.io/
  helm install minio minio/minio \
    --namespace minio \
    --create-namespace \
    --set rootUser=minioadmin \
    --set rootPassword=minioadmin \
    --set mode=standalone
  
  # Wait for ready
  kubectl wait --for=condition=ready pod -l app=minio \
    -n minio --timeout=2m
  
  assert_success
}

# Test 2: Create MinIO buckets
test_minio_buckets() {
  # Port forward
  kubectl port-forward -n minio svc/minio 9000:9000 &
  PF_PID=$!
  sleep 3
  
  # Configure mc client
  mc alias set local http://localhost:9000 minioadmin minioadmin
  
  # Create buckets
  mc mb local/knative-lambda-source
  mc mb local/knative-lambda-tmp
  
  # Verify
  mc ls local/ | grep knative-lambda-source
  mc ls local/ | grep knative-lambda-tmp
  
  # Cleanup
  kill $PF_PID
  assert_success
}

# Test 3: Upload to MinIO
test_minio_upload() {
  export STORAGE_PROVIDER=minio
  export S3_ENDPOINT=http://minio.minio.svc.cluster.local:9000
  export STORAGE_ACCESS_KEY_ID=minioadmin
  export STORAGE_SECRET_ACCESS_KEY=minioadmin
  export STORAGE_USE_SSL=false
  export STORAGE_USE_IAM=false
  
  # Build test binary
  go build -o test-upload ./tests/storage/upload
  
  # Run upload test
  ./test-upload --bucket knative-lambda-tmp --key test/file.txt --data "test content"
  
  # Verify via mc
  kubectl port-forward -n minio svc/minio 9000:9000 &
  PF_PID=$!
  sleep 2
  
  mc cat local/knative-lambda-tmp/test/file.txt | grep "test content"
  
  kill $PF_PID
  assert_success
}

# Test 4: Download from MinIO
test_minio_download() {
  export STORAGE_PROVIDER=minio
  export S3_ENDPOINT=http://minio.minio.svc.cluster.local:9000
  
  # Upload test file first
  mc cp /tmp/test.txt local/knative-lambda-tmp/test.txt
  
  # Build test binary
  go build -o test-download ./tests/storage/download
  
  # Download
  content=$(./test-download --bucket knative-lambda-tmp --key test.txt)
  
  # Verify
  echo "$content" | grep "expected content"
  assert_success
}

# Test 5: Full build pipeline with MinIO
test_minio_build_pipeline() {
  # Deploy knative-lambda with MinIO config
  helm upgrade --install knative-lambda-builder ./deploy \
    -f deploy/overlays/local/values-local.yaml \
    --namespace knative-lambda \
    --create-namespace
  
  # Trigger build
  cd tests && ENV=local uv run --python 3.9 python create-event-builder.py
  
  # Wait for build completion
  sleep 30
  
  # Verify build artifacts in MinIO
  mc ls local/knative-lambda-tmp/ | grep build-context
  
  assert_success
}
```

---

### **Story 2: Storage Abstraction Layer**

**As a** platform engineer  
**I want** a storage abstraction layer  
**So that** I can switch backends without code changes

**Acceptance Criteria:**

✅ **AC1**: `Storage` interface defines all operations  
✅ **AC2**: Factory creates client based on `STORAGE_PROVIDER`  
✅ **AC3**: Supports: `aws-s3`, `minio`, `gcs` (future)  
✅ **AC4**: Credentials from env vars or IAM roles  
✅ **AC5**: Consistent error handling across backends  
✅ **AC6**: All callers use interface, not concrete types  

**Tests:**

```go
// internal/storage/factory_test.go

func TestStorageFactory_MinIO(t *testing.T) {
    config := &config.StorageConfig{
        Type:            "minio",
        Endpoint:        "http://minio:9000",
        Bucket:          "test-bucket",
        AccessKeyID:     "minioadmin",
        SecretAccessKey: "minioadmin",
        UseSSL:          false,
        UseIAM:          false,
    }
    
    factory := NewStorageFactory(config, nil)
    storage, err := factory.GetStorage(context.Background())
    
    require.NoError(t, err)
    assert.NotNil(t, storage)
    
    // Verify it's S3-compatible client
    s3Storage, ok := storage.(*S3Storage)
    require.True(t, ok)
    assert.Equal(t, "http://minio:9000", s3Storage.endpoint)
}

func TestStorageFactory_S3(t *testing.T) {
    config := &config.StorageConfig{
        Type:   "aws-s3",
        Region: "us-west-2",
        Bucket: "test-bucket",
        UseIAM: true,
    }
    
    factory := NewStorageFactory(config, nil)
    storage, err := factory.GetStorage(context.Background())
    
    require.NoError(t, err)
    assert.NotNil(t, storage)
}

func TestStorageFactory_InvalidType(t *testing.T) {
    config := &config.StorageConfig{
        Type: "invalid",
    }
    
    factory := NewStorageFactory(config, nil)
    _, err := factory.GetStorage(context.Background())
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unsupported storage type")
}

func TestStorage_Upload(t *testing.T) {
    tests := []struct {
        name     string
        provider string
        config   *config.StorageConfig
    }{
        {
            name:     "MinIO upload",
            provider: "minio",
            config: &config.StorageConfig{
                Type:            "minio",
                Endpoint:        "http://localhost:9000",
                AccessKeyID:     "minioadmin",
                SecretAccessKey: "minioadmin",
            },
        },
        {
            name:     "S3 upload",
            provider: "aws-s3",
            config: &config.StorageConfig{
                Type:   "aws-s3",
                Region: "us-west-2",
                UseIAM: true,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            factory := NewStorageFactory(tt.config, nil)
            storage, err := factory.GetStorage(context.Background())
            require.NoError(t, err)
            
            // Upload test
            data := strings.NewReader("test content")
            err = storage.Upload(context.Background(), "test/file.txt", data)
            assert.NoError(t, err)
            
            // Verify exists
            exists, err := storage.Exists(context.Background(), "test/file.txt")
            assert.NoError(t, err)
            assert.True(t, exists)
            
            // Download and verify
            reader, err := storage.Download(context.Background(), "test/file.txt")
            require.NoError(t, err)
            defer reader.Close()
            
            content, err := io.ReadAll(reader)
            require.NoError(t, err)
            assert.Equal(t, "test content", string(content))
            
            // Cleanup
            err = storage.Delete(context.Background(), "test/file.txt")
            assert.NoError(t, err)
        })
    }
}
```

---

### **Story 3: Environment-Based Storage Selection**

**As a** DevOps engineer  
**I want** automatic storage backend selection per environment  
**So that** local uses MinIO, dev/prd use S3 without manual changes

**Acceptance Criteria:**

✅ **AC1**: `local` environment uses MinIO  
✅ **AC2**: `dev` environment uses AWS S3  
✅ **AC3**: `prd` environment uses AWS S3  
✅ **AC4**: Storage config in environment-specific values  
✅ **AC5**: Credentials managed appropriately (IAM vs static)  
✅ **AC6**: Bucket names configurable per environment  

**Tests:**

```bash
# Test 1: Local uses MinIO
test_local_uses_minio() {
  grep "storageProvider: \"minio\"" deploy/overlays/local/values-local.yaml
  grep "s3Endpoint: \"http://minio" deploy/overlays/local/values-local.yaml
  grep "storageUseIam: \"false\"" deploy/overlays/local/values-local.yaml
  assert_success
}

# Test 2: Dev uses S3
test_dev_uses_s3() {
  grep "storageProvider: \"aws-s3\"" deploy/overlays/dev/values-dev.yaml
  grep "storageUseIam: \"true\"" deploy/overlays/dev/values-dev.yaml
  assert_success
}

# Test 3: Production uses S3
test_prd_uses_s3() {
  grep "storageProvider: \"aws-s3\"" deploy/overlays/prd/values-prd.yaml
  grep "storageUseIam: \"true\"" deploy/overlays/prd/values-prd.yaml
  grep "s3SourceBucket: \"notifi-uw2-prd-fusion-modules\"" deploy/overlays/prd/values-prd.yaml
  assert_success
}

# Test 4: Bucket names per environment
test_bucket_names() {
  # Local
  grep "s3TmpBucket: \"knative-lambda-tmp\"" deploy/overlays/local/values-local.yaml
  
  # Dev
  grep "s3TmpBucket: \"knative-lambda-context-tmp\"" deploy/overlays/dev/values-dev.yaml
  
  # Prd
  grep "s3TmpBucket: \"knative-lambda-context-tmp\"" deploy/overlays/prd/values-prd.yaml
  
  assert_success
}

# Test 5: Credentials configuration
test_credentials() {
  # Local should have static credentials
  kubectl get deployment knative-lambda-builder -n knative-lambda -o yaml | \
    grep "STORAGE_ACCESS_KEY_ID"
  
  # Prd should use IAM (no credentials in env)
  ! kubectl get deployment knative-lambda-builder -n knative-lambda -o yaml | \
    grep "STORAGE_ACCESS_KEY_ID"
  
  assert_success
}
```

---

### **Story 4: GCS Support (Future)**

**As a** platform engineer  
**I want** to support Google Cloud Storage  
**So that** we can deploy to GCP environments

**Acceptance Criteria:**

✅ **AC1**: GCS client implements `Storage` interface  
✅ **AC2**: Configuration supports `STORAGE_PROVIDER=gcs`  
✅ **AC3**: Authentication via service account or workload identity  
✅ **AC4**: All operations work identically to S3  
✅ **AC5**: Migration path from S3 to GCS documented  

**Tests:**

```go
// internal/storage/gcs_test.go

func TestGCSStorage_Upload(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping GCS integration test")
    }
    
    config := &config.StorageConfig{
        Type:              "gcs",
        GCSBucket:         "test-bucket",
        GCSProject:        "my-project",
        GCSCredentialsPath: "/path/to/credentials.json",
    }
    
    storage, err := NewGCSStorage(config, nil)
    require.NoError(t, err)
    
    // Upload
    data := strings.NewReader("test content")
    err = storage.Upload(context.Background(), "test/file.txt", data)
    assert.NoError(t, err)
    
    // Download
    reader, err := storage.Download(context.Background(), "test/file.txt")
    require.NoError(t, err)
    defer reader.Close()
    
    content, err := io.ReadAll(reader)
    require.NoError(t, err)
    assert.Equal(t, "test content", string(content))
    
    // Cleanup
    err = storage.Delete(context.Background(), "test/file.txt")
    assert.NoError(t, err)
}
```

---

## 🧪 Integration Tests

### Test Suite: Multi-Storage End-to-End

```bash
#!/bin/bash
# tests/integration/test-multi-storage.sh

set -e

echo "🧪 Testing Multi-Storage Backend Support"

# Test 1: MinIO complete flow
echo "Test 1: MinIO complete flow..."

# Deploy MinIO
helm install minio minio/minio \
  --namespace minio \
  --create-namespace \
  --set rootUser=minioadmin \
  --set rootPassword=minioadmin \
  --wait

# Create buckets
kubectl port-forward -n minio svc/minio 9000:9000 &
PF_PID=$!
sleep 3

mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/knative-lambda-source
mc mb local/knative-lambda-tmp

kill $PF_PID

# Deploy knative-lambda with MinIO
helm upgrade --install knative-lambda-builder ./deploy \
  -f deploy/overlays/local/values-local.yaml \
  --namespace knative-lambda \
  --create-namespace \
  --wait

# Trigger build
cd tests && ENV=local uv run --python 3.9 python create-event-builder.py

# Wait and verify
sleep 30
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=100 | \
  grep "Successfully uploaded to MinIO"

echo "✅ MinIO flow complete"

# Test 2: S3 complete flow
echo "Test 2: S3 complete flow..."

# Deploy knative-lambda with S3
helm upgrade --install knative-lambda-builder ./deploy \
  -f deploy/overlays/dev/values-dev.yaml \
  --namespace knative-lambda \
  --wait

# Trigger build
cd tests && ENV=dev uv run --python 3.9 python create-event-builder.py

# Wait and verify
sleep 30
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=100 | \
  grep "Successfully uploaded to S3"

echo "✅ S3 flow complete"

# Test 3: Storage interface abstraction
echo "Test 3: Storage interface abstraction..."

# Run unit tests
go test -v ./internal/storage/... -race

echo "✅ Storage abstraction tests complete"

# Test 4: Environment-based selection
echo "Test 4: Environment-based storage selection..."

for env in local dev prd; do
  echo "Validating $env storage config..."
  
  # Dry-run deployment
  helm upgrade --install knative-lambda-builder ./deploy \
    -f deploy/overlays/$env/values-$env.yaml \
    --namespace knative-lambda-$env \
    --dry-run
  
  echo "✅ $env storage config valid"
done

echo "🎉 All multi-storage tests passed!"
```

---

## 📝 Implementation Checklist

### Phase 1: MinIO Support (Week 1)

- [ ] Add MinIO Helm dependency to `deploy/Chart.yaml`
- [ ] Update `internal/storage/s3.go` to detect MinIO endpoint
- [ ] Add custom endpoint support to S3 client
- [ ] Add SSL/TLS toggle for MinIO
- [ ] Create `deploy/overlays/local/values-local.yaml` with MinIO config
- [ ] Add `make minio-deploy` target to Makefile
- [ ] Add `make minio-create-buckets` target
- [ ] Test MinIO deployment
- [ ] Test upload/download to MinIO
- [ ] Write MinIO unit tests
- [ ] Write MinIO integration tests

### Phase 2: Storage Abstraction (Week 2)

- [ ] Enhance `internal/storage/interface.go`
- [ ] Create `internal/storage/factory.go`
- [ ] Update `internal/storage/s3.go` for abstraction
- [ ] Add storage type validation in config
- [ ] Update all callers to use Storage interface
- [ ] Add credential loading logic (IAM vs static)
- [ ] Add comprehensive error handling
- [ ] Write factory unit tests
- [ ] Write storage interface tests
- [ ] Update documentation

### Phase 3: Environment-Based Selection (Week 3)

- [ ] Update all environment overlays with storage config
- [ ] Add MinIO config to `deploy/overlays/local/values-local.yaml`
- [ ] Update S3 config in `deploy/overlays/dev/values-dev.yaml`
- [ ] Update S3 config in `deploy/overlays/prd/values-prd.yaml`
- [ ] Add storage provider validation
- [ ] Test each environment configuration
- [ ] Write environment selection tests
- [ ] Add migration guide

### Phase 4: GCS Support (Future - Week 4)

- [ ] Add GCS SDK dependency
- [ ] Create `internal/storage/gcs.go`
- [ ] Implement GCS authentication
- [ ] Add GCS to factory
- [ ] Create `deploy/overlays/gcp/values-gcp.yaml`
- [ ] Test GCS upload/download
- [ ] Write GCS tests
- [ ] Document GCS setup

---

## 🎯 Definition of Done

- ✅ All user stories implemented
- ✅ All acceptance criteria met
- ✅ All tests passing (unit + integration + e2e)
- ✅ MinIO works for local development
- ✅ S3 continues to work (dev/prd)
- ✅ Storage abstraction layer complete
- ✅ Environment-based selection works
- ✅ Can switch backends without code changes
- ✅ Zero production downtime
- ✅ Code reviewed and approved
- ✅ Documentation updated

---

**Estimated Effort**: 3 weeks (MinIO + Abstraction + Selection)  
**Dependencies**: None  
**Risk Level**: Low (additive changes, no breaking changes to S3)

