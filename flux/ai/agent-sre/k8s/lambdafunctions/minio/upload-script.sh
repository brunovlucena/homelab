#!/bin/bash
set -e

echo "📤 Uploading agent-sre LambdaFunction code to MinIO..."

# Get MinIO credentials
MINIO_ACCESS_KEY=$(kubectl get secret minio-credentials -n minio -o jsonpath='{.data.access-key}' | base64 -d)
MINIO_SECRET_KEY=$(kubectl get secret minio-credentials -n minio -o jsonpath='{.data.secret-key}' | base64 -d)

# Upload using mc in a pod
kubectl run mc-upload-$(date +%s) --rm -i --restart=Never \
  --image=minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1 \
  --namespace=ai \
  --env="MINIO_ACCESS_KEY=$MINIO_ACCESS_KEY" \
  --env="MINIO_SECRET_KEY=$MINIO_SECRET_KEY" \
  -- sh -c '
    mc alias set minio http://minio.minio.svc.cluster.local:9000 "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY"
    mc mb -p minio/lambda-functions || true
    
    # Note: Files need to be provided via ConfigMap or mounted volume
    # This is a placeholder - actual upload requires file contents
    echo "✅ MinIO configured. Files need to be uploaded manually."
    echo "   Use: mc cp <file> minio/lambda-functions/agent-sre/<function-name>/main.py"
  '

echo "📝 To upload files manually, use:"
echo "   kubectl run mc --rm -i --restart=Never --image=minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1 --namespace=ai -- sh"
echo "   Then inside the pod:"
echo "   mc alias set minio http://minio.minio.svc.cluster.local:9000 <access-key> <secret-key>"
echo "   mc cp <local-file> minio/lambda-functions/agent-sre/<function>/main.py"
