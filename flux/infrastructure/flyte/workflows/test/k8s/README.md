# Flyte Workflow Registration (Kubernetes Job)

This directory contains Kubernetes resources to automatically register Flyte workflows from within the cluster, eliminating the need for port-forwards or local scripts.

## How It Works

1. **ConfigMap Generation**: Kustomize automatically generates a ConfigMap from the workflow source file (`../workflows/agent_training.py`)
2. **Job Execution**: A Kubernetes Job runs inside the cluster and:
   - Installs flytekit
   - Mounts the workflow code from the ConfigMap
   - Connects to Flyte admin via internal service DNS (`flyteadmin.flyte.svc.cluster.local:8089`)
   - Connects to MinIO via internal service DNS (`minio.minio.svc.cluster.local:9000`)
   - Registers the workflow using `pyflyte register`

## Automatic Updates

- When the workflow code changes, Kustomize generates a new ConfigMap with a different hash
- The Job annotation `workflow/hash` can be updated to trigger recreation
- Flux will automatically recreate the Job when the manifest changes (due to `kustomize.toolkit.fluxcd.io/force: "enabled"`)

## Manual Trigger

To manually trigger workflow registration:

```bash
# Delete the existing job to force recreation
kubectl delete job register-agent-training-workflow -n flyte

# Flux will automatically recreate it, or apply manually:
kubectl apply -k .
```

## Requirements

- `minio-credentials` secret must exist in the `flyte` namespace with:
  - `access-key`: MinIO access key
  - `secret-key`: MinIO secret key
- Flyte admin service must be accessible at `flyteadmin.flyte.svc.cluster.local:8089`
- MinIO service must be accessible at `minio.minio.svc.cluster.local:9000`

## Viewing Job Status

```bash
# Check job status
kubectl get job register-agent-training-workflow -n flyte

# View job logs
kubectl logs -n flyte -l app=flyte-workflow-registration,workflow=agent-training --tail=100

# View specific pod logs
kubectl get pods -n flyte -l app=flyte-workflow-registration,workflow=agent-training
kubectl logs <pod-name> -n flyte
```

## Troubleshooting

### Job fails with connection errors
- Verify Flyte admin is running: `kubectl get pods -n flyte -l app.kubernetes.io/name=flyteadmin`
- Verify MinIO is running: `kubectl get pods -n minio`
- Check service DNS resolution from within cluster

### ConfigMap not found
- Ensure the workflow file exists at `../workflows/agent_training.py`
- Run `kubectl get configmap -n flyte` to see generated ConfigMaps

### Workflow registration fails
- Check that the Docker image exists: `ghcr.io/brunovlucena/flyte-sandbox-training:latest`
- Verify Flyte admin is accessible from the job pod
- Check job logs for detailed error messages
