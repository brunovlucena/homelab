# ScyllaDB with Alternator (DynamoDB Compatibility)

## 🎯 Overview

ScyllaDB is a high-performance NoSQL database compatible with Apache Cassandra. This deployment includes **Alternator**, which provides DynamoDB-compatible API, making it a production-grade alternative to LocalStack's DynamoDB emulation.

## ✨ Features

- **🔌 DynamoDB API Compatibility**: Use existing DynamoDB SDKs and code
- **🚀 High Performance**: C++ rewrite of Cassandra, optimized for modern hardware
- **💾 Persistent Storage**: Real database persistence (unlike LocalStack's in-memory mode)
- **📊 Prometheus Metrics**: Built-in monitoring integration
- **🔄 Dual API**: Both CQL (Cassandra) and DynamoDB (Alternator) protocols

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│         ScyllaDB Cluster                │
│                                         │
│  ┌─────────────┐    ┌─────────────┐   │
│  │   CQL API   │    │ Alternator  │   │
│  │  (Port 9042)│    │ (Port 8000) │   │
│  └─────────────┘    └─────────────┘   │
│         │                   │          │
│         └───────┬───────────┘          │
│                 │                      │
│         ┌───────▼────────┐            │
│         │  ScyllaDB Core │            │
│         │   (Developer   │            │
│         │     Mode)      │            │
│         └────────────────┘            │
└─────────────────────────────────────────┘
```

## 🔧 Configuration

### Key Settings

- **Alternator Port**: 8000 (DynamoDB-compatible API)
- **CQL Port**: 9042 (Cassandra-compatible API)
- **Metrics Port**: 9180 (Prometheus scraping)
- **Developer Mode**: Enabled (reduced resource requirements)
- **Storage**: 20Gi persistent volume

### Resource Allocation

```yaml
Resources:
  Requests:
    Memory: 2Gi
    CPU: 500m
  Limits:
    Memory: 4Gi
    CPU: 2000m
```

## 🚀 Quick Start

### 1. Access Alternator (DynamoDB API)

The Alternator API is accessible within the cluster at:

```
http://scylladb.scylladb.svc.cluster.local:8000
```

### 2. Port Forward for Local Access

```bash
# Port forward Alternator (DynamoDB API)
kubectl port-forward -n scylladb svc/scylladb 8000:8000

# Port forward CQL (Cassandra API)
kubectl port-forward -n scylladb svc/scylladb 9042:9042
```

### 3. Test with AWS CLI

```bash
# Configure AWS CLI to use ScyllaDB
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_DEFAULT_REGION=us-east-1

# Create a table
aws dynamodb create-table \
    --table-name TestTable \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000

# Put an item
aws dynamodb put-item \
    --table-name TestTable \
    --item '{"id": {"S": "test-1"}, "name": {"S": "ScyllaDB"}}' \
    --endpoint-url http://localhost:8000

# Get the item
aws dynamodb get-item \
    --table-name TestTable \
    --key '{"id": {"S": "test-1"}}' \
    --endpoint-url http://localhost:8000
```

### 4. Test with Python (boto3)

```python
import boto3

# Create DynamoDB client pointing to ScyllaDB Alternator
dynamodb = boto3.client(
    'dynamodb',
    endpoint_url='http://localhost:8000',
    region_name='us-east-1',
    aws_access_key_id='dummy',
    aws_secret_access_key='dummy'
)

# Create table
table = dynamodb.create_table(
    TableName='users',
    KeySchema=[
        {'AttributeName': 'user_id', 'KeyType': 'HASH'}
    ],
    AttributeDefinitions=[
        {'AttributeName': 'user_id', 'AttributeType': 'S'}
    ],
    BillingMode='PAY_PER_REQUEST'
)

# Put item
dynamodb.put_item(
    TableName='users',
    Item={
        'user_id': {'S': 'user-123'},
        'name': {'S': 'Bruno'},
        'email': {'S': 'bruno@example.com'}
    }
)

# Get item
response = dynamodb.get_item(
    TableName='users',
    Key={'user_id': {'S': 'user-123'}}
)
print(response['Item'])
```

## 🔄 Migration from LocalStack

If you're currently using LocalStack's DynamoDB, ScyllaDB with Alternator is a drop-in replacement:

### Before (LocalStack):
```python
dynamodb = boto3.client(
    'dynamodb',
    endpoint_url='http://localstack.localstack.svc.cluster.local:4566',
    region_name='us-east-1'
)
```

### After (ScyllaDB Alternator):
```python
dynamodb = boto3.client(
    'dynamodb',
    endpoint_url='http://scylladb.scylladb.svc.cluster.local:8000',
    region_name='us-east-1'
)
```

## 📊 Monitoring

ScyllaDB exports Prometheus metrics on port 9180. A ServiceMonitor is automatically created for Prometheus Operator integration.

### Key Metrics

- `scylla_database_total_writes`: Total write operations
- `scylla_database_total_reads`: Total read operations
- `scylla_storage_proxy_coordinator_write_latency`: Write latency
- `scylla_storage_proxy_coordinator_read_latency`: Read latency
- `scylla_alternator_*`: Alternator-specific metrics

### Access Grafana Dashboards

ScyllaDB has official Grafana dashboards available. Import dashboard ID: `17032` (ScyllaDB Overview)

## 🔍 Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n scylladb
kubectl describe pod -n scylladb scylladb-0
```

### View Logs

```bash
kubectl logs -n scylladb scylladb-0 --follow
```

### Exec into Pod

```bash
kubectl exec -it -n scylladb scylladb-0 -- bash

# Inside pod, use cqlsh
cqlsh localhost 9042
```

### Check Alternator Status

```bash
# From inside the cluster
curl http://scylladb.scylladb.svc.cluster.local:8000/
```

## 🎛️ Advanced Configuration

### Scaling the Cluster

To scale up to 3 nodes:

```yaml
cluster:
  replicaCount: 3
```

### Adjusting Resources

For production workloads, increase resources:

```yaml
resources:
  requests:
    memory: "8Gi"
    cpu: "2000m"
  limits:
    memory: "16Gi"
    cpu: "4000m"
```

### Disable Developer Mode

For production:

```yaml
scylladb:
  developerMode: false
  extraFlags:
    - "--smp 4"
    - "--memory 8G"
```

## 📚 Additional Resources

- [ScyllaDB Documentation](https://docs.scylladb.com/)
- [Alternator Documentation](https://docs.scylladb.com/stable/using-scylla/alternator/)
- [DynamoDB API Compatibility](https://docs.scylladb.com/stable/using-scylla/alternator/compatibility.html)
- [ScyllaDB University](https://university.scylladb.com/)

## 🆚 ScyllaDB vs LocalStack DynamoDB

| Feature | LocalStack DynamoDB | ScyllaDB Alternator |
|---------|-------------------|-------------------|
| **Performance** | Limited (emulation) | High (native) |
| **Persistence** | In-memory or basic | Full persistence |
| **Production Use** | Development only | Production-ready |
| **Scalability** | Single instance | Distributed cluster |
| **Monitoring** | Limited | Full Prometheus |
| **License** | Free/Pro | Open Source |
| **API Coverage** | ~80% | ~95% |

## 🎯 Use Cases

- **Development/Testing**: DynamoDB-compatible local development environment
- **Production**: High-performance NoSQL database with DynamoDB compatibility
- **Migration**: Gradual migration from AWS DynamoDB to self-hosted
- **Cost Optimization**: Reduce AWS DynamoDB costs by hosting in-house
- **Multi-Model**: Use both CQL and DynamoDB APIs on the same data

