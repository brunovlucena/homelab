# 🐰 RabbitMQ Cluster

This directory contains the RabbitMQ cluster configuration using the RabbitMQ Cluster Operator CRD.

## 📦 Components

- **Namespace**: `rabbitmq-cluster`
- **RabbitmqCluster CR**: Defines a single-node RabbitMQ cluster
- **ServiceMonitor**: Prometheus monitoring integration

## ⚙️ Configuration

The cluster is configured with:

- 🔢 **Replicas**: 1 (single-node for development)
- 🐳 **Image**: `rabbitmq:3.12-management-alpine`
- 💾 **Persistence**: 1Gi storage with `standard` StorageClass
- 📊 **Management UI**: Enabled via `rabbitmq_management` plugin
- 🎯 **Node Tolerations**: Configured to run on nodes with `knative` taint

### Resource Limits

- **Requests**: 100m CPU, 512Mi memory
- **Limits**: 500m CPU, 512Mi memory

### Memory Configuration

RabbitMQ is configured with container-optimized memory settings:
- VM memory high watermark: 256MiB
- Total memory override: 512MiB (536870912 bytes)
- Disk free limit: 200MB
- Memory calculation strategy: `allocated`

### Topology Access

The cluster allows topology operations (queues, exchanges, bindings) from the following namespaces:
- `knative-lambda`
- `knative-eventing`
- `rabbitmq-cluster`

## 🔐 Access

### Management UI

The RabbitMQ management UI is available through the cluster service. To access it locally:

```bash
kubectl port-forward -n rabbitmq-cluster svc/rabbitmq-cluster 15672:15672
```

Then open http://localhost:15672

Default credentials are stored in the auto-generated secret: `rabbitmq-cluster-default-user`

```bash
kubectl get secret -n rabbitmq-cluster rabbitmq-cluster-default-user -o jsonpath='{.data.username}' | base64 -d
kubectl get secret -n rabbitmq-cluster rabbitmq-cluster-default-user -o jsonpath='{.data.password}' | base64 -d
```

### AMQP Connection

Applications can connect to RabbitMQ using the service endpoint:

```
amqp://rabbitmq-cluster.rabbitmq-cluster.svc.cluster.local:5672
```

## 🔍 Monitoring

The cluster exposes Prometheus metrics on port 15692. The ServiceMonitor is configured to scrape these metrics every 30 seconds.

## 🛠️ Operations

### Scale the Cluster

To scale to a multi-node cluster, update the `replicas` field in `cluster.yaml`:

```yaml
spec:
  replicas: 3
```

### View Logs

```bash
kubectl logs -n rabbitmq-cluster -l app.kubernetes.io/name=rabbitmq-cluster
```

### Check Cluster Status

```bash
kubectl get rabbitmqclusters -n rabbitmq-cluster
kubectl describe rabbitmqcluster -n rabbitmq-cluster rabbitmq-cluster
```

## 📚 References

- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [RabbitMQ Cluster Operator](https://www.rabbitmq.com/kubernetes/operator/operator-overview.html)
- [RabbitMQ Management Plugin](https://www.rabbitmq.com/management.html)

