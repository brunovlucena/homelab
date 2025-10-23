# 📦 ScyllaDB Deployment with Official ScyllaDB Operator

## 🎯 Overview

This deployment uses the **official [ScyllaDB Operator](https://operator.docs.scylladb.com/)** to run ScyllaDB with **Alternator** (DynamoDB-compatible API) on Kubernetes.

> **Note**: This deployment has been migrated from the Bitnami Helm chart to the official ScyllaDB Operator due to [Bitnami's distribution changes](https://community.broadcom.com/blogs/beltran-rueda-borrego/2025/08/18/how-to-prepare-for-the-bitnami-changes-coming-soon) in September 2025.

### 🌟 Key Features

- ✅ **Official ScyllaDB Operator** - Managed by ScyllaDB team
- ✅ **Alternator API** - DynamoDB-compatible interface on port 8000
- ✅ **CQL API** - Cassandra-compatible interface on port 9042
- ✅ **Prometheus Metrics** - Automatic ServiceMonitor creation
- ✅ **Auto-healing** - Operator manages node failures
- ✅ **GitOps Ready** - Fully managed via Flux

## 📁 Structure

```
/repos/homelab/flux/clusters/homelab/infrastructure/scylladb/
├── namespace-operator.yaml      # Namespace for ScyllaDB Operator
├── helmrelease-operator.yaml    # ScyllaDB Operator deployment
├── namespace.yaml                # Namespace for ScyllaDB cluster
├── helmrelease-scylla.yaml      # ScyllaDB cluster deployment
├── kustomization.yaml            # Kustomize configuration
└── README.md                     # This file
```

## 🚀 Quick Start

### 1. Check Deployment Status

```bash
# Check ScyllaDB Operator
kubectl get pods -n scylla-operator

# Check ScyllaDB cluster
kubectl get pods -n scylladb
kubectl get scyllaclusters -n scylladb
```

### 2. Access ScyllaDB

**DynamoDB API (Alternator):**
```
Endpoint: http://scylla-client.scylladb.svc.cluster.local:8000
```

**CQL API (Cassandra):**
```
Endpoint: scylla-client.scylladb.svc.cluster.local:9042
```

**Prometheus Metrics:**
```
Endpoint: http://scylla-client.scylladb.svc.cluster.local:9180/metrics
```

### 3. Port Forwarding for Local Access

```bash
# DynamoDB API
kubectl port-forward -n scylladb svc/scylla-client 8000:8000

# CQL API
kubectl port-forward -n scylladb svc/scylla-client 9042:9042

# Prometheus Metrics
kubectl port-forward -n scylladb svc/scylla-client 9180:9180
```

## 🔌 Using Alternator (DynamoDB API)

### Python Example

```python
import boto3

# Create DynamoDB client pointing to ScyllaDB Alternator
dynamodb = boto3.resource(
    'dynamodb',
    endpoint_url='http://scylla-client.scylladb.svc.cluster.local:8000',
    region_name='us-east-1',
    aws_access_key_id='none',
    aws_secret_access_key='none'
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
table.put_item(Item={'user_id': '123', 'name': 'Alice'})

# Get item
response = table.get_item(Key={'user_id': '123'})
print(response['Item'])
```

## 📊 Monitoring

### Prometheus Metrics

The deployment automatically creates a `ServiceMonitor` for Prometheus to scrape ScyllaDB metrics:

```bash
# View metrics
kubectl port-forward -n scylladb svc/scylla-client 9180:9180
curl http://localhost:9180/metrics
```

### Important Metrics

- `scylla_database_total_writes` - Total write operations
- `scylla_database_total_reads` - Total read operations
- `scylla_storage_proxy_coordinator_write_latency` - Write latency
- `scylla_storage_proxy_coordinator_read_latency` - Read latency
- `scylla_alternator_*` - Alternator-specific metrics

### Grafana Dashboard

Import the official ScyllaDB dashboard:
- **Dashboard ID**: 17032 (ScyllaDB Overview)

## 🔧 Configuration

### Current Setup

- **Namespace**: `scylladb` (cluster) and `scylla-operator` (operator)
- **Datacenter**: `homelab-dc1`
- **Racks**: 1 rack with 1 node
- **Storage**: 20Gi per node
- **Resources**: 
  - CPU: 500m request, 2 limit
  - Memory: 2Gi request, 4Gi limit
- **ScyllaDB Version**: 6.2.3
- **Developer Mode**: Enabled (reduced resource requirements)

### Scaling

To scale up the cluster, edit `helmrelease-scylla.yaml`:

```yaml
racks:
  - name: rack1
    members: 3  # Change from 1 to 3
```

Then commit and push to let Flux apply the changes.

## 🛠️ Troubleshooting

### Check Operator Logs

```bash
kubectl logs -n scylla-operator -l app.kubernetes.io/name=scylla-operator
```

### Check ScyllaDB Cluster Status

```bash
# Get cluster status
kubectl get scyllaclusters -n scylladb

# Describe cluster
kubectl describe scyllacluster -n scylladb scylla-homelab-dc1

# Check pod logs
kubectl logs -n scylladb scylla-homelab-dc1-rack1-0 -c scylla
```

### Common Issues

**Pods not starting:**
```bash
# Check events
kubectl get events -n scylladb --sort-by='.lastTimestamp'

# Check pod description
kubectl describe pod -n scylladb <pod-name>
```

**Alternator not accessible:**
```bash
# Verify service
kubectl get svc -n scylladb scylla-client

# Test connectivity
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl -v http://scylla-client.scylladb.svc.cluster.local:8000
```

## 📚 Resources

- [ScyllaDB Operator Documentation](https://operator.docs.scylladb.com/)
- [ScyllaDB Operator Helm Charts](https://operator.docs.scylladb.com/stable/installation/helm.html)
- [ScyllaDB Alternator Documentation](https://docs.scylladb.com/stable/using-scylla/alternator/)
- [DynamoDB API Compatibility](https://docs.scylladb.com/stable/using-scylla/alternator/compatibility.html)
- [ScyllaDB University](https://university.scylladb.com/)
- [ScyllaDB GitHub Operator](https://github.com/scylladb/scylla-operator)

## 🔄 Migration from Bitnami

This deployment was migrated from the Bitnami Helm chart due to Bitnami's distribution changes. The official ScyllaDB Operator provides:

- ✅ Better support and active maintenance by ScyllaDB team
- ✅ Advanced features (auto-healing, rolling upgrades, multi-datacenter)
- ✅ Official Helm charts that are regularly updated
- ✅ Production-grade deployment best practices

### What Changed

1. **Helm Repository**: Changed from `oci://registry-1.docker.io/bitnamicharts` to `https://scylla-operator-charts.storage.googleapis.com/stable`
2. **Deployment Model**: Now uses ScyllaDB Operator + ScyllaCluster CRD
3. **Configuration**: Moved from Bitnami values to ScyllaDB Operator values
4. **Images**: Now uses official `scylladb/scylla` images

## ✅ Deployment Checklist

- [x] ScyllaDB Operator deployed in `scylla-operator` namespace
- [x] ScyllaDB cluster deployed in `scylladb` namespace
- [x] Alternator (DynamoDB API) enabled on port 8000
- [x] CQL API available on port 9042
- [x] Prometheus ServiceMonitor created
- [x] Resource limits configured for homelab

## 🎓 Next Steps

1. ✅ Verify cluster is running: `kubectl get pods -n scylladb`
2. ✅ Test DynamoDB API connectivity
3. ✅ Import Grafana dashboard for monitoring
4. ✅ Update applications to use new endpoint
5. ✅ Configure backups (see ScyllaDB Manager documentation)

---

**ScyllaDB Version**: 6.2.3  
**Operator Version**: >=1.0.0  
**Deployment Date**: October 23, 2025  
**Migration Reason**: Bitnami distribution changes
