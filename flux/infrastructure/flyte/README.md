# Flyte Installation

Flyte workflow orchestration platform for ML and data pipelines.

## Prerequisites

- PostgreSQL database (using existing `postgres.postgres.svc.cluster.local`)
- MinIO object storage (using existing `minio.minio.svc.cluster.local`)
- Kubernetes cluster (v1.19+)

## Setup

### 1. Create Flyte Database

Create a database for Flyte in Postgres:

```bash
kubectl exec -it -n postgres deployment/postgres -- psql -U postgres -c "CREATE DATABASE flyte;"
```

### 2. Create MinIO Bucket

Create a bucket for Flyte data:

```bash
# Port-forward MinIO console
kubectl port-forward -n minio svc/minio 9001:9001

# Or use mc client
mc alias set minio http://minio.minio.svc.cluster.local:9000 $(kubectl get secret -n minio minio-credentials -o jsonpath='{.data.access-key}' | base64 -d) $(kubectl get secret -n minio minio-credentials -o jsonpath='{.data.secret-key}' | base64 -d)
mc mb minio/flyte-data
```

### 3. Configure HelmRelease Values

The HelmRelease needs database password and MinIO credentials. You have two options:

#### Option A: Use ExternalSecrets (Recommended)

Create an ExternalSecret to sync credentials:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: flyte-credentials
  namespace: flyte
spec:
  secretStoreRef:
    name: kubernetes
    kind: SecretStore
  target:
    name: flyte-credentials
    creationPolicy: Owner
  data:
    - secretKey: db-password
      remoteRef:
        name: postgres
        namespace: postgres
        key: password
    - secretKey: minio-access-key
      remoteRef:
        name: minio-credentials
        namespace: minio
        key: access-key
    - secretKey: minio-secret-key
      remoteRef:
        name: minio-credentials
        namespace: minio
        key: secret-key
```

Then update `helmrelease.yaml` to reference these values.

#### Option B: Manual Configuration

1. Get the Postgres password:
```bash
kubectl get secret -n postgres postgres -o jsonpath='{.data.password}' | base64 -d
```

2. Get MinIO credentials:
```bash
kubectl get secret -n minio minio-credentials -o jsonpath='{.data.access-key}' | base64 -d
kubectl get secret -n minio minio-credentials -o jsonpath='{.data.secret-key}' | base64 -d
```

3. Update `helmrelease.yaml` with these values in the `userSettings` section.

### 4. Deploy

Flux will automatically reconcile the HelmRelease. Monitor with:

```bash
kubectl get helmrelease -n flyte
kubectl get pods -n flyte
```

## Access

Once deployed, access Flyte Console:

```bash
# Port-forward the console service (UI)
kubectl port-forward -n flyte svc/flyteconsole 8080:80

# In a separate terminal, port-forward the admin service (API)
kubectl port-forward -n flyte svc/flyteadmin 8088:80
```

Then open http://localhost:8080/console in your browser.

**Note:** The console needs access to the `flyteadmin` service. When accessing via port-forward, you need both services forwarded. The console will automatically connect to the admin service via Kubernetes DNS when running inside the cluster, but when accessing via port-forward from your browser, you need both services accessible.

**Projects:** The `homelab` project should already exist (created by the `flyte-create-project` job). If you don't see projects in the UI, check the browser console for API connection errors.

## Configuration

The HelmRelease is configured to use:
- **Database**: `postgres.postgres.svc.cluster.local:5432/flyte`
- **Storage**: MinIO at `minio.minio.svc.cluster.local:9000` bucket `flyte-data`

## Resources

- [Flyte Documentation](https://docs.flyte.org/)
- [Flyte Helm Charts](https://helm.flyte.org)
- [Flyte GitHub](https://github.com/flyteorg/flyte)
