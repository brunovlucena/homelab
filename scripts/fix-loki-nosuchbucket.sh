#!/bin/bash
# 🔧 Script: Fix Loki "NoSuchBucket" Storage Issues

set -euo pipefail

NAMESPACE="loki"

echo "🔧 Fixing Loki NoSuchBucket errors..."
echo ""
echo "🪣 Creating bucket creation job..."

cat <<'EOF' | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: loki-minio-make-buckets
  namespace: loki
spec:
  ttlSecondsAfterFinished: 300
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - name: minio-mc
        image: quay.io/minio/mc:latest
        command:
        - /bin/sh
        - -c
        - |
          set -e
          echo "🔧 Configuring MinIO client..."
          mc alias set myminio http://loki-minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
          
          echo "🪣 Creating buckets..."
          for bucket in loki chunks ruler admin; do
            if mc ls myminio/$bucket >/dev/null 2>&1; then
              echo "✅ Bucket '$bucket' already exists"
            else
              echo "🆕 Creating bucket '$bucket'..."
              mc mb myminio/$bucket
              echo "✅ Created bucket '$bucket'"
            fi
          done
          
          echo ""
          echo "📋 All buckets:"
          mc ls myminio
          echo "✅ Done!"
        env:
        - name: MINIO_ROOT_USER
          valueFrom:
            secretKeyRef:
              name: loki-minio
              key: rootUser
        - name: MINIO_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: loki-minio
              key: rootPassword
EOF

echo "✅ Job created"
echo ""

kubectl wait --for=condition=complete --timeout=120s job/loki-minio-make-buckets -n loki || {
    echo "📋 Job logs:"
    kubectl logs -n loki job/loki-minio-make-buckets
    exit 1
}

echo "✅ Job completed!"
echo ""
kubectl logs -n loki job/loki-minio-make-buckets
echo ""

echo "🔄 Restarting Loki write..."
kubectl rollout restart statefulset -n loki loki-write

echo "⏳ Waiting 30s..."
sleep 30

ERROR_COUNT=$(kubectl logs -n loki -l app.kubernetes.io/component=write --since=30s 2>/dev/null | grep "NoSuchBucket" | wc -l | tr -d ' ')

if [ "$ERROR_COUNT" -eq 0 ]; then
    echo "✅ No new NoSuchBucket errors! 🎉"
else
    echo "⚠️  Still seeing $ERROR_COUNT NoSuchBucket errors"
fi
