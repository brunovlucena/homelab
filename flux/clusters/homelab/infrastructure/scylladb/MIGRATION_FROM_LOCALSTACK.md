# 🔄 Migrating from LocalStack DynamoDB to ScyllaDB Alternator

## Overview

This guide helps you migrate from LocalStack's DynamoDB emulation to ScyllaDB's Alternator (production-grade DynamoDB-compatible API).

## 🎯 Why Migrate?

| Feature | LocalStack DynamoDB | ScyllaDB Alternator |
|---------|-------------------|-------------------|
| **Purpose** | Development/testing emulation | Production-ready database |
| **Performance** | Limited (emulation overhead) | High-performance (native implementation) |
| **Persistence** | In-memory or basic file storage | Full distributed persistence |
| **Scalability** | Single instance | Distributed cluster (up to 1000s of nodes) |
| **Resource Usage** | Moderate | Configurable (low in dev mode) |
| **API Coverage** | ~80% of DynamoDB API | ~95% of DynamoDB API |
| **Production Use** | ❌ Not recommended | ✅ Production-ready |
| **Cost** | Free/Pro license | Open source (free) |

## 📋 Migration Checklist

- [ ] Deploy ScyllaDB with Alternator enabled
- [ ] Test ScyllaDB with your application
- [ ] Update service endpoints in your code
- [ ] Migrate data (if needed)
- [ ] Update monitoring dashboards
- [ ] Run integration tests
- [ ] Decommission LocalStack (optional)

## 🚀 Step-by-Step Migration

### Step 1: Deploy ScyllaDB Alongside LocalStack

ScyllaDB is already configured in the same deployment phase as LocalStack. Both can run simultaneously.

```bash
# Verify both are running
kubectl get pods -n localstack
kubectl get pods -n scylladb

# Check services
kubectl get svc -n localstack
kubectl get svc -n scylladb
```

### Step 2: Update Your Application Configuration

#### Before (LocalStack):
```python
import boto3
import os

dynamodb = boto3.client(
    'dynamodb',
    endpoint_url=os.getenv('DYNAMODB_ENDPOINT', 'http://localstack.localstack.svc.cluster.local:4566'),
    region_name='us-east-1',
    aws_access_key_id='test',
    aws_secret_access_key='test'
)
```

#### After (ScyllaDB):
```python
import boto3
import os

dynamodb = boto3.client(
    'dynamodb',
    endpoint_url=os.getenv('DYNAMODB_ENDPOINT', 'http://scylladb.scylladb.svc.cluster.local:8000'),
    region_name='us-east-1',
    aws_access_key_id='test',  # Can be any value
    aws_secret_access_key='test'  # Can be any value
)
```

**Key Changes:**
- ✅ Update `endpoint_url` from port 4566 to 8000
- ✅ Update service name from `localstack.localstack` to `scylladb.scylladb`
- ✅ Credentials can remain the same (any value works)

### Step 3: Environment Variables Approach (Recommended)

Use environment variables to make the switch seamless:

```yaml
# Kubernetes ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  # Switch between LocalStack and ScyllaDB by changing this value
  DYNAMODB_ENDPOINT: "http://scylladb.scylladb.svc.cluster.local:8000"
  # DYNAMODB_ENDPOINT: "http://localstack.localstack.svc.cluster.local:4566"
```

```python
# Application code
import boto3
import os

dynamodb = boto3.client(
    'dynamodb',
    endpoint_url=os.getenv('DYNAMODB_ENDPOINT'),
    region_name='us-east-1',
    aws_access_key_id='test',
    aws_secret_access_key='test'
)
```

### Step 4: Data Migration (If Needed)

If you have data in LocalStack that needs to be migrated:

```python
#!/usr/bin/env python3
"""
Migrate data from LocalStack DynamoDB to ScyllaDB Alternator
"""

import boto3

# Source (LocalStack)
localstack_client = boto3.client(
    'dynamodb',
    endpoint_url='http://localhost:4566',  # Port forward LocalStack
    region_name='us-east-1',
    aws_access_key_id='test',
    aws_secret_access_key='test'
)

# Destination (ScyllaDB)
scylladb_client = boto3.client(
    'dynamodb',
    endpoint_url='http://localhost:8000',  # Port forward ScyllaDB
    region_name='us-east-1',
    aws_access_key_id='test',
    aws_secret_access_key='test'
)

def migrate_table(table_name):
    """Migrate a single table from LocalStack to ScyllaDB"""
    
    print(f"Migrating table: {table_name}")
    
    # Get table schema from LocalStack
    response = localstack_client.describe_table(TableName=table_name)
    table_desc = response['Table']
    
    # Create table in ScyllaDB with same schema
    create_params = {
        'TableName': table_name,
        'KeySchema': table_desc['KeySchema'],
        'AttributeDefinitions': table_desc['AttributeDefinitions'],
    }
    
    # Use BillingMode if available
    if 'BillingModeSummary' in table_desc:
        create_params['BillingMode'] = table_desc['BillingModeSummary']['BillingMode']
    else:
        create_params['ProvisionedThroughput'] = {
            'ReadCapacityUnits': table_desc['ProvisionedThroughput']['ReadCapacityUnits'],
            'WriteCapacityUnits': table_desc['ProvisionedThroughput']['WriteCapacityUnits']
        }
    
    try:
        scylladb_client.create_table(**create_params)
        print(f"  ✅ Created table {table_name}")
    except Exception as e:
        if 'ResourceInUseException' in str(e):
            print(f"  ⚠️  Table {table_name} already exists")
        else:
            raise
    
    # Scan and copy all items
    paginator = localstack_client.get_paginator('scan')
    page_iterator = paginator.paginate(TableName=table_name)
    
    item_count = 0
    for page in page_iterator:
        items = page.get('Items', [])
        
        # Batch write items to ScyllaDB
        if items:
            with scylladb_client.batch_writer(TableName=table_name) as batch:
                for item in items:
                    batch.put_item(Item=item)
                    item_count += 1
    
    print(f"  ✅ Migrated {item_count} items")

def migrate_all_tables():
    """Migrate all tables from LocalStack to ScyllaDB"""
    
    # Get list of tables from LocalStack
    response = localstack_client.list_tables()
    tables = response.get('TableNames', [])
    
    print(f"Found {len(tables)} tables to migrate")
    
    for table_name in tables:
        try:
            migrate_table(table_name)
        except Exception as e:
            print(f"  ❌ Error migrating {table_name}: {e}")
    
    print("\n✅ Migration complete!")

if __name__ == '__main__':
    # Port forward both services before running:
    # kubectl port-forward -n localstack svc/localstack 4566:4566
    # kubectl port-forward -n scylladb svc/scylladb 8000:8000
    
    migrate_all_tables()
```

### Step 5: Testing

#### A. Test with AWS CLI

```bash
# Port forward both services
kubectl port-forward -n localstack svc/localstack 4566:4566 &
kubectl port-forward -n scylladb svc/scylladb 8000:8000 &

# Test LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

aws dynamodb list-tables --endpoint-url http://localhost:4566

# Test ScyllaDB
aws dynamodb list-tables --endpoint-url http://localhost:8000
```

#### B. Run Integration Tests

```bash
# Test ScyllaDB
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/scylladb
./test-scylladb.sh
```

#### C. Application-Specific Tests

Run your application's test suite against ScyllaDB:

```bash
# Set environment to use ScyllaDB
export DYNAMODB_ENDPOINT=http://localhost:8000

# Run your tests
pytest tests/
# or
npm test
# or
go test ./...
```

### Step 6: Gradual Rollout (Blue-Green Deployment)

For safer migration in production-like environments:

#### Option A: Environment Variables per Service

```yaml
# Service A (using ScyllaDB)
apiVersion: v1
kind: Pod
metadata:
  name: service-a
spec:
  containers:
  - name: app
    env:
    - name: DYNAMODB_ENDPOINT
      value: "http://scylladb.scylladb.svc.cluster.local:8000"
```

```yaml
# Service B (still using LocalStack)
apiVersion: v1
kind: Pod
metadata:
  name: service-b
spec:
  containers:
  - name: app
    env:
    - name: DYNAMODB_ENDPOINT
      value: "http://localstack.localstack.svc.cluster.local:4566"
```

#### Option B: Feature Flags

```python
import os
import boto3

USE_SCYLLADB = os.getenv('USE_SCYLLADB', 'false').lower() == 'true'

if USE_SCYLLADB:
    endpoint_url = 'http://scylladb.scylladb.svc.cluster.local:8000'
else:
    endpoint_url = 'http://localstack.localstack.svc.cluster.local:4566'

dynamodb = boto3.client(
    'dynamodb',
    endpoint_url=endpoint_url,
    region_name='us-east-1',
    aws_access_key_id='test',
    aws_secret_access_key='test'
)
```

### Step 7: Update Monitoring

#### Prometheus Queries

Replace LocalStack metrics with ScyllaDB metrics:

**Before (LocalStack):**
LocalStack has limited metrics

**After (ScyllaDB):**
```promql
# Write throughput
rate(scylla_database_total_writes[5m])

# Read throughput
rate(scylla_database_total_reads[5m])

# Write latency (p99)
histogram_quantile(0.99, rate(scylla_storage_proxy_coordinator_write_latency_bucket[5m]))

# Read latency (p99)
histogram_quantile(0.99, rate(scylla_storage_proxy_coordinator_read_latency_bucket[5m]))
```

#### Grafana Dashboards

Import ScyllaDB dashboard:
- Dashboard ID: `17032` (ScyllaDB Overview)

### Step 8: Decommission LocalStack (Optional)

Once you've fully migrated and validated:

```yaml
# Comment out LocalStack in phase5-mocks/kustomization.yaml
resources:
  - ../../infrastructure/notifi-test
  - ../../infrastructure/alerts/mocks
  # - ../../infrastructure/localstack  # Disabled, using ScyllaDB
  - ../../infrastructure/scylladb
```

Or keep both for different purposes:
- **LocalStack**: For other AWS service emulation (S3, SQS, SNS, etc.)
- **ScyllaDB**: For DynamoDB workloads

## 🎯 Common Migration Patterns

### Pattern 1: Config-Based Switch

```python
# config.py
import os

DATABASES = {
    'localstack': {
        'endpoint': 'http://localstack.localstack.svc.cluster.local:4566',
        'type': 'emulation'
    },
    'scylladb': {
        'endpoint': 'http://scylladb.scylladb.svc.cluster.local:8000',
        'type': 'production'
    }
}

DYNAMODB_BACKEND = os.getenv('DYNAMODB_BACKEND', 'scylladb')
DYNAMODB_CONFIG = DATABASES[DYNAMODB_BACKEND]
```

### Pattern 2: Dual-Write for Safety

```python
class DualWriteDynamoDB:
    """Write to both LocalStack and ScyllaDB during migration"""
    
    def __init__(self):
        self.localstack = boto3.client(
            'dynamodb',
            endpoint_url='http://localstack.localstack.svc.cluster.local:4566',
            region_name='us-east-1',
            aws_access_key_id='test',
            aws_secret_access_key='test'
        )
        
        self.scylladb = boto3.client(
            'dynamodb',
            endpoint_url='http://scylladb.scylladb.svc.cluster.local:8000',
            region_name='us-east-1',
            aws_access_key_id='test',
            aws_secret_access_key='test'
        )
        
        self.read_from = os.getenv('DYNAMODB_READ_FROM', 'scylladb')
    
    def put_item(self, **kwargs):
        """Write to both databases"""
        try:
            self.scylladb.put_item(**kwargs)
        except Exception as e:
            print(f"ScyllaDB write failed: {e}")
        
        try:
            self.localstack.put_item(**kwargs)
        except Exception as e:
            print(f"LocalStack write failed: {e}")
    
    def get_item(self, **kwargs):
        """Read from configured database"""
        if self.read_from == 'scylladb':
            return self.scylladb.get_item(**kwargs)
        else:
            return self.localstack.get_item(**kwargs)
```

## 🔍 Validation Checklist

After migration, verify:

- [ ] All tables exist in ScyllaDB
- [ ] Data is correctly migrated (if applicable)
- [ ] Application can read/write to ScyllaDB
- [ ] Performance meets requirements
- [ ] Metrics are being collected
- [ ] Backups are configured
- [ ] Integration tests pass
- [ ] No errors in application logs

## 📊 Performance Comparison

Run benchmarks to compare:

```bash
# LocalStack benchmark
time bash -c 'for i in {1..1000}; do \
  aws dynamodb put-item \
    --table-name test \
    --item "{\"id\": {\"S\": \"$i\"}}" \
    --endpoint-url http://localhost:4566 \
    >/dev/null 2>&1; \
done'

# ScyllaDB benchmark
time bash -c 'for i in {1..1000}; do \
  aws dynamodb put-item \
    --table-name test \
    --item "{\"id\": {\"S\": \"$i\"}}" \
    --endpoint-url http://localhost:8000 \
    >/dev/null 2>&1; \
done'
```

## 🎓 Best Practices

1. **Start Small**: Migrate one table/service at a time
2. **Test Thoroughly**: Run integration tests before full migration
3. **Monitor Closely**: Watch metrics during and after migration
4. **Keep LocalStack**: Use for other AWS services (S3, SQS, etc.)
5. **Document Changes**: Update architecture diagrams and documentation
6. **Gradual Rollout**: Use feature flags or environment variables
7. **Backup Data**: Export LocalStack data before migration (if needed)

## ❓ Troubleshooting

### Issue: Application can't connect to ScyllaDB

```bash
# Test connectivity
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://scylladb.scylladb.svc.cluster.local:8000

# Check service
kubectl get svc -n scylladb
kubectl get endpoints -n scylladb
```

### Issue: Different behavior than LocalStack

ScyllaDB Alternator has better DynamoDB API coverage. Check:
- [API Compatibility Matrix](https://docs.scylladb.com/stable/using-scylla/alternator/compatibility.html)
- Ensure you're using supported operations

### Issue: Performance is slower than expected

1. Check if developer mode is enabled (reduce resources)
2. Increase CPU/memory limits
3. Disable developer mode for production workloads

## 📚 Additional Resources

- [ScyllaDB Alternator Documentation](https://docs.scylladb.com/stable/using-scylla/alternator/)
- [DynamoDB API Reference](https://docs.aws.amazon.com/dynamodb/)
- [ScyllaDB Performance Guide](https://docs.scylladb.com/stable/operating-scylla/performance/)

---

**Need Help?** Check the [README.md](./README.md) or [QUICKSTART.md](./QUICKSTART.md) for more information.

