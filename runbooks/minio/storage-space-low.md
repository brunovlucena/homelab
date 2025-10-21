# 🚨 Runbook: MinIO Storage Space Low

## Alert Information

**Alert Name:** `MinIOStorageSpaceLow`  
**Severity:** High  
**Component:** minio  
**Service:** object-storage

## Symptom

MinIO storage is approaching capacity. Disk usage is above threshold (typically >80%).

## Impact

- **User Impact:** MEDIUM - Uploads may start failing soon
- **Business Impact:** HIGH - Risk of service disruption
- **Data Impact:** MEDIUM - No new data can be written when full

## Diagnosis

### 1. Check Current Storage Usage

```bash
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- df -h /data
```

### 2. Check PVC Usage

```bash
kubectl get pvc -n minio
kubectl describe pvc -n minio
```

### 3. List Bucket Sizes

```bash
# Port forward to MinIO
kubectl port-forward -n minio svc/minio 9000:9000

# Use MinIO client to check bucket sizes
mc alias set local http://localhost:9000 <access-key> <secret-key>
mc du local/ --depth 1
```

### 4. Check for Large Objects

```bash
# Find largest objects in a bucket
mc ls --recursive local/<bucket-name> | sort -k4 -n -r | head -20
```

### 5. Check Node Disk Space

```bash
NODE=$(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].spec.nodeName}')
kubectl get node $NODE -o jsonpath='{.status.capacity.storage}'
```

## Resolution

### Scenario A: Old/Unused Data Cleanup

**Likely Cause:** Accumulation of old backups or temporary files

**Steps:**
1. Identify old/unused buckets:
   ```bash
   mc ls local/
   ```

2. Check bucket lifecycle policies:
   ```bash
   mc ilm ls local/<bucket-name>
   ```

3. Set up lifecycle policy to auto-delete old objects:
   ```bash
   # Delete objects older than 90 days
   mc ilm add local/<bucket-name> --expiry-days 90
   
   # Or delete incomplete multipart uploads older than 7 days
   mc ilm add local/<bucket-name> --expire-delete-marker --noncurrentversion-expiration-days 7
   ```

4. Manually delete old data if needed:
   ```bash
   # Be careful with this command!
   mc rm --recursive --force local/<bucket-name>/old-data/
   ```

### Scenario B: Increase Storage Capacity

**Likely Cause:** Legitimate growth in data volume

**Steps:**
1. Check if PV can be expanded:
   ```bash
   kubectl get storageclass
   # Look for "allowVolumeExpansion: true"
   ```

2. If expansion is allowed, edit PVC:
   ```bash
   kubectl edit pvc -n minio <pvc-name>
   # Increase spec.resources.requests.storage
   ```

   Or update in Flux:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   persistence:
     size: 200Gi  # Increase from current value
   ```

3. Commit and push changes:
   ```bash
   git add .
   git commit -m "Increase MinIO storage capacity"
   git push
   flux reconcile helmrelease -n minio minio
   ```

4. Monitor expansion:
   ```bash
   kubectl get pvc -n minio -w
   ```

### Scenario C: Enable Compression

**Likely Cause:** Storing compressible data without compression

**Steps:**
1. Enable compression in MinIO:
   ```bash
   mc admin config set local compression enable="on" extensions=".txt,.log,.csv,.json,.xml"
   mc admin service restart local
   ```

2. Or configure via environment variables:
   ```yaml
   # In HelmRelease
   environment:
     MINIO_COMPRESS: "on"
     MINIO_COMPRESS_EXTENSIONS: ".txt,.log,.csv,.json,.xml"
     MINIO_COMPRESS_MIME_TYPES: "text/*,application/json,application/xml"
   ```

### Scenario D: Set Up Multi-Tier Storage

**Likely Cause:** All data stored in hot storage

**Steps:**
1. Configure S3-compatible remote tier:
   ```bash
   mc admin tier add s3 local COLD-TIER \
     --endpoint https://s3.amazonaws.com \
     --access-key <access-key> \
     --secret-key <secret-key> \
     --bucket archive-bucket \
     --region us-east-1
   ```

2. Create lifecycle rule to transition old data:
   ```bash
   mc ilm add local/<bucket-name> \
     --transition-days 30 \
     --storage-class "COLD-TIER"
   ```

### Scenario E: Emergency Cleanup

**Likely Cause:** Storage critically full (>95%)

**Steps:**
1. Identify and delete temporary files:
   ```bash
   # Find .tmp files
   mc find local/ --name "*.tmp"
   
   # Delete them
   mc find local/ --name "*.tmp" --exec "mc rm {}"
   ```

2. Clear incomplete multipart uploads:
   ```bash
   mc rm --recursive --incomplete local/<bucket-name>/
   ```

3. Check and clear MinIO's internal .minio.sys:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   cd /data/.minio.sys
   du -sh *
   # Clean up old logs if safe
   ```

## Verification

### 1. Check Storage Usage After Cleanup

```bash
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- df -h /data
```

Should show reduced usage, ideally below 70%.

### 2. Verify MinIO is Functioning

```bash
mc ls local/
# Should list buckets successfully
```

### 3. Test Upload

```bash
echo "test" > test.txt
mc cp test.txt local/<bucket-name>/test.txt
mc rm local/<bucket-name>/test.txt
rm test.txt
```

### 4. Check Lifecycle Policies

```bash
mc ilm ls local/<bucket-name>
# Should show configured policies
```

## Prevention

1. **Set up storage monitoring:**
   ```yaml
   # Prometheus alert for storage >80%
   - alert: MinIOStorageSpaceLow
     expr: |
       (minio_cluster_disk_total_bytes - minio_cluster_disk_free_bytes) 
       / minio_cluster_disk_total_bytes * 100 > 80
     for: 5m
     annotations:
       summary: "MinIO storage usage above 80%"
   ```

2. **Configure lifecycle policies proactively:**
   - Set expiration for temporary buckets
   - Auto-delete old backups
   - Transition old data to cold storage

3. **Regular capacity planning:**
   - Review storage trends monthly
   - Project growth for next 6 months
   - Plan capacity increases in advance

4. **Enable compression for appropriate data:**
   - Text files
   - Logs
   - JSON/XML

5. **Set up automated cleanup jobs:**
   ```yaml
   # CronJob to clean old data
   apiVersion: batch/v1
   kind: CronJob
   metadata:
     name: minio-cleanup
     namespace: minio
   spec:
     schedule: "0 2 * * 0"  # Weekly at 2 AM Sunday
     jobTemplate:
       spec:
         template:
           spec:
             containers:
             - name: cleanup
               image: minio/mc:latest
               command:
               - /bin/sh
               - -c
               - |
                 mc alias set local http://minio:9000 $ACCESS_KEY $SECRET_KEY
                 mc rm --recursive --force --older-than 90d local/temp-bucket/
             restartPolicy: OnFailure
   ```

## Metrics to Monitor

```promql
# Current storage usage percentage
(minio_cluster_disk_total_bytes - minio_cluster_disk_free_bytes) / minio_cluster_disk_total_bytes * 100

# Storage usage rate (bytes per hour)
rate(minio_cluster_disk_total_bytes[1h])

# Number of objects
minio_bucket_objects_size_bytes

# Available free space
minio_cluster_disk_free_bytes
```

## Related Alerts

- `MinIODown`
- `MinIOHighErrorRate`
- `MinIODiskFull` (>95%)
- `MinIOPVCExpansionFailed`

## Escalation

**When to escalate:**
- Storage >95% and cleanup efforts ineffective
- PVC expansion failing
- Storage growth rate unsustainable
- Application requirements exceed infrastructure capacity

**Escalation Path:**
1. Senior SRE Team
2. Storage Infrastructure Team
3. Capacity Planning Team
4. Finance (for budget approval if expansion needed)

## Additional Resources

- [MinIO Lifecycle Management](https://min.io/docs/minio/linux/administration/object-management/object-lifecycle-management.html)
- [MinIO Compression](https://min.io/docs/minio/linux/operations/server-side-encryption.html#compression)
- [MinIO Tiering](https://min.io/docs/minio/linux/operations/data-recovery/tiering.html)
- Internal Wiki: Storage Capacity Planning
- Slack: #sre-storage

