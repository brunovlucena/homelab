# 🐰 RabbitMQ Cluster Operator

This directory contains the official RabbitMQ Cluster Operator installation using `kubectl krew` plugin, which enables managing RabbitMQ clusters as Kubernetes custom resources.

## 📦 Components

- **Namespace**: `rabbitmq-operator`
- **Installation Method**: kubectl krew plugin (`kubectl rabbitmq install-cluster-operator`)
- **Custom Image**: `ghcr.io/brunovlucena/homelab/kubectl:v1.31.0` (kubectl + krew + rabbitmq plugin)

## 🚀 Installation

The operator is installed via a Kubernetes Job that:

1. Uses a custom kubectl image with krew and the rabbitmq plugin pre-installed
2. Runs `kubectl rabbitmq install-cluster-operator` to deploy the official operator
3. Creates the operator in the `rabbitmq-system` namespace (as per RabbitMQ defaults)

## 🐳 Docker Image

The custom kubectl image includes:
- kubectl v1.31.0
- krew plugin manager
- RabbitMQ kubectl plugin

Build the image:
```bash
cd docker/
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/brunovlucena/homelab/kubectl:v1.31.0 \
  --push .
```

## 🔧 Usage

Once deployed, the operator watches for `RabbitmqCluster` custom resources in the cluster and manages RabbitMQ clusters accordingly.

## 📚 References

- [Official RabbitMQ Cluster Operator Installation](https://www.rabbitmq.com/kubernetes/operator/install-operator)
- [RabbitMQ kubectl Plugin](https://www.rabbitmq.com/kubernetes/operator/kubectl-plugin)
- [RabbitMQ Operator GitHub](https://github.com/rabbitmq/cluster-operator)

