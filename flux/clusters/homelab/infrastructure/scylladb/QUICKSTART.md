# 🚀 ScyllaDB Quick Start Guide

## What is ScyllaDB with Alternator?

ScyllaDB is a high-performance NoSQL database with **Alternator**, a DynamoDB-compatible API. This allows you to:

- ✅ Use AWS DynamoDB SDK without modifications
- ✅ Run DynamoDB-compatible applications locally or in your homelab
- ✅ Get production-grade performance and persistence (unlike LocalStack)
- ✅ Avoid AWS costs for development/testing environments

## 🎯 5-Minute Setup

### 1. Deploy ScyllaDB

The deployment is managed by Flux. Simply ensure the kustomization is included in your deployment phase:

```bash
# Verify the deployment
kubectl get helmrelease -n scylladb
kubectl get pods -n scylladb

# Wait for the pod to be ready
kubectl wait --for=condition=ready pod -n scylladb -l app.kubernetes.io/name=scylladb --timeout=5m
```

### 2. Port Forward (for local testing)

```bash
# Forward the Alternator (DynamoDB) port
kubectl port-forward -n scylladb svc/scylladb 8000:8000
```

### 3. Test with AWS CLI

```bash
# Set dummy credentials
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_DEFAULT_REGION=us-east-1

# Create a table
aws dynamodb create-table \
    --table-name users \
    --attribute-definitions AttributeName=user_id,AttributeType=S \
    --key-schema AttributeName=user_id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --endpoint-url http://localhost:8000

# Put an item
aws dynamodb put-item \
    --table-name users \
    --item '{"user_id": {"S": "123"}, "name": {"S": "Bruno"}, "email": {"S": "bruno@example.com"}}' \
    --endpoint-url http://localhost:8000

# Get the item
aws dynamodb get-item \
    --table-name users \
    --key '{"user_id": {"S": "123"}}' \
    --endpoint-url http://localhost:8000
```

### 4. Run Automated Tests

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/scylladb
./test-scylladb.sh
```

This script tests:
- ✅ Table creation and deletion
- ✅ CRUD operations (Create, Read, Update, Delete)
- ✅ Scan and Query operations
- ✅ Performance benchmarks
- ✅ Metrics endpoint

## 🐍 Python Example

```python
import boto3

# Create DynamoDB client
dynamodb = boto3.client(
    'dynamodb',
    endpoint_url='http://localhost:8000',  # or use the cluster service URL
    region_name='us-east-1',
    aws_access_key_id='dummy',
    aws_secret_access_key='dummy'
)

# Create table
dynamodb.create_table(
    TableName='products',
    KeySchema=[
        {'AttributeName': 'product_id', 'KeyType': 'HASH'}
    ],
    AttributeDefinitions=[
        {'AttributeName': 'product_id', 'AttributeType': 'S'}
    ],
    BillingMode='PAY_PER_REQUEST'
)

# Insert data
dynamodb.put_item(
    TableName='products',
    Item={
        'product_id': {'S': 'PROD-001'},
        'name': {'S': 'Laptop'},
        'price': {'N': '999.99'},
        'in_stock': {'BOOL': True}
    }
)

# Query data
response = dynamodb.get_item(
    TableName='products',
    Key={'product_id': {'S': 'PROD-001'}}
)
print(response['Item'])
```

## 🔄 Migrating from LocalStack

If you're currently using LocalStack's DynamoDB, migration is simple:

### Before (LocalStack):
```python
endpoint_url='http://localstack.localstack.svc.cluster.local:4566'
```

### After (ScyllaDB):
```python
endpoint_url='http://scylladb.scylladb.svc.cluster.local:8000'
```

**That's it!** Your code doesn't need any other changes.

## 📊 Monitoring

ScyllaDB exports Prometheus metrics automatically:

```bash
# Port forward metrics
kubectl port-forward -n scylladb svc/scylladb 9180:9180

# View metrics
curl http://localhost:9180/metrics | grep scylla
```

### Key Metrics to Watch:

- `scylla_database_total_writes` - Total write operations
- `scylla_database_total_reads` - Total read operations  
- `scylla_storage_proxy_coordinator_write_latency_bucket` - Write latency
- `scylla_storage_proxy_coordinator_read_latency_bucket` - Read latency

## 🔍 Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n scylladb
kubectl describe pod -n scylladb scylladb-0

# Check logs
kubectl logs -n scylladb scylladb-0 --tail=100
```

### Connection Issues

```bash
# Verify service
kubectl get svc -n scylladb

# Test from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://scylladb.scylladb.svc.cluster.local:8000
```

### Performance Issues

If you're experiencing slow performance:

1. **Check resources**: Ensure the pod has enough CPU/memory
2. **Disable developer mode**: Edit helmrelease.yaml and set `developerMode: false`
3. **Scale up**: Increase CPU/memory limits
4. **Add nodes**: Scale to 3 replicas for production

## 🎛️ Configuration Options

### Enable Production Mode

Edit `helmrelease.yaml`:

```yaml
scylladb:
  developerMode: false
  extraFlags:
    - "--smp 4"
    - "--memory 8G"
```

### Scale to 3 Nodes

```yaml
cluster:
  replicaCount: 3
```

### Increase Storage

```yaml
persistence:
  size: 50Gi  # Increase from default 20Gi
```

## 📚 Learn More

- **Full Documentation**: See [README.md](./README.md)
- **ScyllaDB Docs**: https://docs.scylladb.com/
- **Alternator Guide**: https://docs.scylladb.com/stable/using-scylla/alternator/
- **DynamoDB API**: https://docs.aws.amazon.com/dynamodb/

## 🎯 Next Steps

1. ✅ Test the basic setup with `./test-scylladb.sh`
2. ✅ Update your applications to use ScyllaDB endpoint
3. ✅ Set up Grafana dashboards for monitoring
4. ✅ Configure backups (ScyllaDB supports automated backups)
5. ✅ Consider scaling for production workloads

## 💡 Pro Tips

- **Development**: Use `developerMode: true` for lower resource usage
- **Production**: Disable developer mode and increase resources
- **Testing**: Use PAY_PER_REQUEST billing mode (no capacity planning needed)
- **Monitoring**: Import Grafana dashboard #17032 for ScyllaDB metrics
- **Dual API**: You can use both CQL (Cassandra) and DynamoDB APIs on the same data!

---

**Questions or Issues?** Check the [README.md](./README.md) for detailed documentation.

