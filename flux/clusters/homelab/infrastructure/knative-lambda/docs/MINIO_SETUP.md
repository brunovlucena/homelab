# 🏠 MinIO Setup Guide for Knative Lambda

## 📋 Overview

This guide explains how to configure knative-lambda to use **MinIO** instead of AWS S3 for object storage in your homelab or local development environment.

MinIO is an S3-compatible object storage solution that can be self-hosted, making it perfect for:
- 🏠 **Homelab deployments** - Run everything locally without AWS costs
- 💻 **Local development** - Test without cloud dependencies
- 🔒 **On-premises** - Keep data in your own infrastructure

## 🎯 Quick Start

### Switch to MinIO (3 Steps)

1. **Set the storage provider** in `values.yaml`:
```yaml
env:
  storageProvider: "minio"  # Changed from "aws-s3"
```

2. **Configure MinIO credentials** in `values.yaml`:
```yaml
env:
  minioEndpoint: "minio.minio.svc.cluster.local:9000"
  minioAccessKey: "minioadmin"
  minioSecretKey: "minioadmin"
  minioSourceBucket: "knative-lambda-source"
  minioTempBucket: "knative-lambda-tmp"
```

3. **Deploy the updated configuration**:
```bash
helm upgrade --install knative-lambda ./deploy \
  --namespace knative-lambda-dev \
  --set env.storageProvider=minio
```

## 🔧 Detailed Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `STORAGE_PROVIDER` | Storage provider type | `aws-s3` | ✅ |
| `MINIO_ENDPOINT` | MinIO server endpoint | `minio.minio.svc.cluster.local:9000` | ✅ (if using MinIO) |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` | ✅ (if using MinIO) |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` | ✅ (if using MinIO) |
| `MINIO_USE_SSL` | Use HTTPS for MinIO | `false` | ❌ |
| `MINIO_REGION` | MinIO region | `us-east-1` | ❌ |
| `MINIO_SOURCE_BUCKET` | Source code bucket | `knative-lambda-source` | ✅ (if using MinIO) |
| `MINIO_TEMP_BUCKET` | Temporary build bucket | `knative-lambda-tmp` | ✅ (if using MinIO) |

### values.yaml Configuration

```yaml
env:
  # 💾 STORAGE PROVIDER - Switch between AWS S3 and local MinIO
  storageProvider: "minio"
  
  # MinIO Configuration (used when storageProvider: "minio")
  minioEndpoint: "minio.minio.svc.cluster.local:9000"
  minioAccessKey: "minioadmin"  # Override via Kubernetes secret
  minioSecretKey: "minioadmin"  # Override via Kubernetes secret
  minioUseSSL: "false"
  minioRegion: "us-east-1"
  minioSourceBucket: "knative-lambda-source"
  minioTempBucket: "knative-lambda-tmp"
```

## 🏗️ MinIO Installation

### Option 1: Helm Chart (Recommended)

```bash
# Add MinIO Helm repository
helm repo add minio https://charts.min.io/
helm repo update

# Install MinIO
helm upgrade --install minio minio/minio \
  --namespace minio \
  --create-namespace \
  --set rootUser=minioadmin \
  --set rootPassword=minioadmin \
  --set mode=standalone \
  --set replicas=1 \
  --set persistence.enabled=true \
  --set persistence.size=10Gi \
  --set resources.requests.memory=512Mi
```

### Option 2: Kubernetes Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: minio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args:
        - server
        - /data
        - --console-address
        - ":9001"
        env:
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
        ports:
        - containerPort: 9000
          name: api
        - containerPort: 9001
          name: console
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: minio-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: minio
spec:
  ports:
  - port: 9000
    targetPort: 9000
    name: api
  - port: 9001
    targetPort: 9001
    name: console
  selector:
    app: minio
```

## 🪣 Bucket Creation

### Using MinIO Client (mc)

```bash
# Install mc
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
sudo mv mc /usr/local/bin/

# Configure mc
mc alias set myminio http://localhost:9000 minioadmin minioadmin

# Create buckets
mc mb myminio/knative-lambda-source
mc mb myminio/knative-lambda-tmp

# Verify buckets
mc ls myminio
```

### Using MinIO Console

1. Access MinIO console at `http://localhost:9001`
2. Login with credentials (`minioadmin` / `minioadmin`)
3. Navigate to **Buckets** → **Create Bucket**
4. Create:
   - `knative-lambda-source` - for source code
   - `knative-lambda-tmp` - for build contexts

## 🔄 Switching Between S3 and MinIO

### For Development (MinIO)

```yaml
# values-dev.yaml
env:
  storageProvider: "minio"
  minioEndpoint: "minio.minio.svc.cluster.local:9000"
  minioAccessKey: "minioadmin"
  minioSecretKey: "minioadmin"
  minioSourceBucket: "knative-lambda-source"
  minioTempBucket: "knative-lambda-tmp"
```

Deploy:
```bash
helm upgrade --install knative-lambda ./deploy \
  -f values-dev.yaml \
  --namespace knative-lambda-dev
```

### For Production (AWS S3)

```yaml
# values-prd.yaml
env:
  storageProvider: "aws-s3"
  awsRegion: "us-west-2"
  awsAccountId: "339954290315"
  s3SourceBucket: "notifi-uw2-prd-fusion-modules"
  s3TmpBucket: "knative-lambda-prd-context-tmp"
  useEksPodIdentity: "true"
```

Deploy:
```bash
helm upgrade --install knative-lambda ./deploy \
  -f values-prd.yaml \
  --namespace knative-lambda-prd
```

## 🔐 Security Best Practices

### 1. Use Kubernetes Secrets

Don't hardcode MinIO credentials in values.yaml:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-credentials
  namespace: knative-lambda-dev
type: Opaque
stringData:
  access-key: "your-secure-access-key"
  secret-key: "your-secure-secret-key"
```

Reference in deployment:
```yaml
env:
  - name: MINIO_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: minio-credentials
        key: access-key
  - name: MINIO_SECRET_KEY
    valueFrom:
      secretKeyRef:
        name: minio-credentials
        key: secret-key
```

### 2. Enable SSL/TLS

For production MinIO deployments:

```yaml
env:
  minioUseSSL: "true"
  minioEndpoint: "minio.example.com:9000"
```

Configure MinIO with TLS certificates:
```bash
mc admin update myminio \
  --cert /path/to/public.crt \
  --key /path/to/private.key
```

### 3. Network Policies

Restrict MinIO access to knative-lambda namespace:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-knative-lambda-to-minio
  namespace: minio
spec:
  podSelector:
    matchLabels:
      app: minio
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: knative-lambda-dev
    ports:
    - protocol: TCP
      port: 9000
```

## 🐛 Troubleshooting

### Connection Issues

**Problem**: Cannot connect to MinIO
```
Error: failed to upload build context: connection refused
```

**Solution**: Verify MinIO service is running and accessible
```bash
kubectl get svc -n minio
kubectl get pods -n minio
kubectl logs -n minio deployment/minio
```

### Bucket Not Found

**Problem**: Bucket doesn't exist
```
Error: NoSuchBucket: The specified bucket does not exist
```

**Solution**: Create the required buckets
```bash
mc mb myminio/knative-lambda-source
mc mb myminio/knative-lambda-tmp
```

### Authentication Failed

**Problem**: Invalid credentials
```
Error: Access Denied
```

**Solution**: Verify credentials match MinIO configuration
```bash
mc alias ls myminio
kubectl get secret minio-credentials -n knative-lambda-dev -o yaml
```

### SSL/TLS Issues

**Problem**: Certificate verification failed
```
Error: x509: certificate signed by unknown authority
```

**Solution**: Either disable SSL verification or add CA certificate
```yaml
env:
  minioUseSSL: "false"  # For development only
```

## 📊 Monitoring

### Check Storage Provider

```bash
kubectl logs -n knative-lambda-dev deployment/knative-lambda-builder | grep "storage_provider"
```

### Verify Bucket Usage

```bash
mc du myminio/knative-lambda-source
mc du myminio/knative-lambda-tmp
```

### Monitor Build Context Uploads

```bash
mc event add myminio/knative-lambda-tmp arn:minio:sqs::_:webhook --event put
```

## 🎓 Architecture Notes

### How Kaniko Uses MinIO

1. **Build Context Upload**: Build context manager uploads `context.tar.gz` to MinIO
2. **Kaniko Job**: Kaniko container downloads context from MinIO using S3-compatible API
3. **Environment Variables**: MinIO credentials and endpoint are passed to Kaniko via env vars
4. **S3 API Compatibility**: Kaniko uses `s3://` URLs but points to MinIO via `S3_ENDPOINT`

### Storage Abstraction

The knative-lambda service uses a **storage abstraction layer** that provides a unified interface:

```go
type ObjectStorage interface {
    UploadObject(ctx, bucket, key string, reader io.Reader, contentType string, size int64) error
    GetObject(ctx, bucket, key string) (io.ReadCloser, ObjectMetadata, error)
    ObjectExists(ctx, bucket, key string) (bool, error)
    DeleteObject(ctx, bucket, key string) error
}
```

Both S3 and MinIO implement this interface, allowing seamless switching between providers.

## 📚 References

- [MinIO Documentation](https://min.io/docs/minio/kubernetes/upstream/index.html)
- [MinIO Client (mc) Guide](https://min.io/docs/minio/linux/reference/minio-mc.html)
- [Kaniko S3 Support](https://github.com/GoogleContainerTools/kaniko#pushing-to-amazon-ecr)
- [knative-lambda Storage Abstraction](../internal/storage/interface.go)

## 🤝 Contributing

If you encounter issues or have improvements for MinIO support, please:
1. Check existing issues
2. Create a detailed issue with logs and configuration
3. Submit a PR with fixes or enhancements

