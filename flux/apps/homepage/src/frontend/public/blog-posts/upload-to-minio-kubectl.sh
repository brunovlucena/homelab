#!/bin/bash
# Upload three-scales-framework.png to MinIO using kubectl pod

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_FILE="$SCRIPT_DIR/three-scales-framework.png"
SVG_FILE="$SCRIPT_DIR/three-scales-framework.svg"

if [ ! -f "$IMAGE_FILE" ]; then
    echo "❌ Image file not found: $IMAGE_FILE"
    echo "💡 Run generate-three-scales-diagram.py first to generate the image"
    exit 1
fi

POD_NAME="upload-three-scales-$(date +%s)"
NAMESPACE="${NAMESPACE:-homepage}"

echo "🚀 Creating upload pod in namespace: $NAMESPACE..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: $POD_NAME
  namespace: $NAMESPACE
spec:
  restartPolicy: Never
  serviceAccountName: minio-access
  initContainers:
    - name: sync-secret
      image: localhost:5001/kubectl:v1.34.0
      command:
        - /bin/sh
        - -c
        - |
          set -e
          echo "🔐 Syncing minio-credentials secret from minio namespace..."
          
          # Wait for secret in minio namespace (source)
          MAX_WAIT=60
          WAITED=0
          while ! kubectl get secret minio-credentials -n minio &>/dev/null; do
            if [ $WAITED -ge $MAX_WAIT ]; then
              echo "❌ Timeout waiting for minio-credentials in minio namespace"
              exit 1
            fi
            echo "⏳ Waiting for minio-credentials in minio namespace... (${WAITED}s/${MAX_WAIT}s)"
            sleep 2
            WAITED=$((WAITED + 2))
          done
          
          echo "📋 Found secret in minio namespace, syncing to $NAMESPACE..."
          
          # Extract credentials from minio namespace
          ACCESS_KEY=\$(kubectl get secret minio-credentials -n minio -o jsonpath='{.data.access-key}' | base64 -d)
          SECRET_KEY=\$(kubectl get secret minio-credentials -n minio -o jsonpath='{.data.secret-key}' | base64 -d)
          ROOT_USER=\$(kubectl get secret minio-credentials -n minio -o jsonpath='{.data.root-user}' | base64 -d || echo "\$ACCESS_KEY")
          ROOT_PASSWORD=\$(kubectl get secret minio-credentials -n minio -o jsonpath='{.data.root-password}' | base64 -d || echo "\$SECRET_KEY")
          
          # Always sync/update secret in target namespace
          kubectl create secret generic minio-credentials \
            --from-literal=access-key="\$ACCESS_KEY" \
            --from-literal=secret-key="\$SECRET_KEY" \
            --from-literal=root-user="\$ROOT_USER" \
            --from-literal=root-password="\$ROOT_PASSWORD" \
            --namespace=$NAMESPACE \
            --dry-run=client -o yaml | kubectl apply -f -
          
          echo "✅ Secret minio-credentials synced to $NAMESPACE namespace"
  containers:
    - name: upload
      image: minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1
      command:
        - /bin/sh
        - -c
        - |
          echo "⏳ Waiting for files to be copied..."
          sleep 30
          echo "📤 Uploading three-scales-framework to MinIO..."
          
          # Configure MinIO client
          mc alias set minio http://minio.minio.svc.cluster.local:9000 "\$MINIO_ACCESS_KEY" "\$MINIO_SECRET_KEY"
          
          # Create bucket if not exists
          mc mb -p minio/homepage-blog || true
          
          # Set public read policy
          mc anonymous set download minio/homepage-blog/images/
          
          # Upload PNG
          if [ -f /tmp/three-scales-framework.png ]; then
            mc cp /tmp/three-scales-framework.png minio/homepage-blog/images/graphs/three-scales-framework.png
            echo "✅ Uploaded: three-scales-framework.png"
          else
            echo "❌ PNG file not found in /tmp/"
            exit 1
          fi
          
          # Upload SVG if exists
          if [ -f /tmp/three-scales-framework.svg ]; then
            mc cp /tmp/three-scales-framework.svg minio/homepage-blog/images/graphs/three-scales-framework.svg
            echo "✅ Uploaded: three-scales-framework.svg"
          fi
          
          echo "📋 Verifying upload..."
          mc ls minio/homepage-blog/images/graphs/three-scales-framework.*
          
          echo "🎉 Upload complete!"
          echo "🌐 Images available at: http://minio.minio.svc.cluster.local:9000/homepage-blog/images/graphs/three-scales-framework.png"
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
      volumeMounts:
        - name: tmp
          mountPath: /tmp
  volumes:
    - name: tmp
      emptyDir: {}
EOF

echo "⏳ Waiting for pod to be ready..."
kubectl wait --for=condition=Ready --timeout=60s pod/$POD_NAME -n $NAMESPACE || {
    echo "❌ Pod failed to start"
    kubectl describe pod/$POD_NAME -n $NAMESPACE || true
    exit 1
}

echo "📤 Copying image files to pod..."
kubectl cp "$IMAGE_FILE" $NAMESPACE/$POD_NAME:/tmp/three-scales-framework.png

if [ -f "$SVG_FILE" ]; then
    kubectl cp "$SVG_FILE" $NAMESPACE/$POD_NAME:/tmp/three-scales-framework.svg
fi

echo "⏳ Waiting for upload to complete..."
kubectl wait --for=condition=Ready=false --timeout=120s pod/$POD_NAME -n $NAMESPACE || true

echo "📋 Pod logs:"
kubectl logs $POD_NAME -n $NAMESPACE || true

# Check if pod completed successfully
if kubectl get pod/$POD_NAME -n $NAMESPACE -o jsonpath='{.status.phase}' | grep -q Succeeded; then
    echo "✅ Upload successful!"
else
    echo "⚠️  Pod did not complete successfully, checking logs..."
    kubectl logs $POD_NAME -n $NAMESPACE || true
fi

echo "🧹 Cleaning up..."
kubectl delete pod/$POD_NAME -n $NAMESPACE || true

echo "🎉 Done! Image is now in MinIO at homepage-blog/images/graphs/three-scales-framework.png"
