#!/bin/bash
# Simple upload script using base64 encoding

set -e

cd "$(dirname "$0")"

echo "📦 Encoding files..."
PNG_B64=$(base64 -i three-scales-framework.png)
SVG_B64=$(base64 -i three-scales-framework.svg)

echo "🚀 Creating upload pod..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: upload-three-scales
  namespace: default
spec:
  restartPolicy: Never
  containers:
    - name: upload
      image: minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1
      command:
        - /bin/sh
        - -c
        - |
          set -e
          echo "📥 Decoding files..."
          echo "$PNG_B64" | base64 -d > /tmp/three-scales-framework.png
          echo "$SVG_B64" | base64 -d > /tmp/three-scales-framework.svg
          echo "✅ Files decoded"
          
          echo "📤 Uploading to MinIO..."
          
          # Configure MinIO
          mc alias set minio http://minio.minio.svc.cluster.local:9000 "\$MINIO_ACCESS_KEY" "\$MINIO_SECRET_KEY" || \\
          mc alias set minio http://localhost:9000 "\$MINIO_ACCESS_KEY" "\$MINIO_SECRET_KEY" || \\
          (echo "❌ Could not connect to MinIO" && exit 1)
          
          # Create bucket
          mc mb -p minio/homepage-blog || true
          
          # Create directory structure
          touch /tmp/.placeholder
          mc cp /tmp/.placeholder minio/homepage-blog/images/graphs/.placeholder || true
          rm /tmp/.placeholder
          
          # Set public policy
          mc anonymous set download minio/homepage-blog/images/ || true
          
          # Upload files
          mc cp /tmp/three-scales-framework.png minio/homepage-blog/images/graphs/three-scales-framework.png
          echo "✅ PNG uploaded"
          
          mc cp /tmp/three-scales-framework.svg minio/homepage-blog/images/graphs/three-scales-framework.svg
          echo "✅ SVG uploaded"
          
          # Verify
          echo "📋 Verifying upload:"
          mc ls minio/homepage-blog/images/graphs/three-scales-framework.*
          
          echo "🎉 Upload complete!"
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
EOF

echo "⏳ Waiting for pod..."
kubectl wait --for=condition=Ready --timeout=30s pod/upload-three-scales -n default || true

echo "⏳ Waiting for upload to complete..."
sleep 15

echo "📋 Pod logs:"
kubectl logs upload-three-scales -n default || true

echo "🧹 Cleaning up..."
kubectl delete pod upload-three-scales -n default --ignore-not-found=true

echo "✅ Done!"

