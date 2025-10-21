# 🐰 RabbitMQ Cluster Operator

This directory contains the Flux HelmRelease for the RabbitMQ Cluster Operator, which enables managing RabbitMQ clusters as Kubernetes custom resources.

## 📦 Components

- **Namespace**: `rabbitmq-operator`
- **Helm Chart**: `bitnami/rabbitmq-cluster-operator`
- **Version**: `>=4.4.34`

## ⚙️ Configuration

The operator is configured with:

- ✅ **Cert-Manager Integration**: Enabled for TLS certificate management
- 🎯 **Node Tolerations**: Configured to run on nodes with `knative` taint
- 📊 **Metrics**: ServiceMonitor enabled for Prometheus scraping
- 🔧 **Resources**: 
  - Requests: 100m CPU, 128Mi memory
  - Limits: 500m CPU, 512Mi memory

## 🚀 Usage

Once deployed, the operator watches for `RabbitmqCluster` custom resources in the cluster and manages RabbitMQ clusters accordingly.

## 🔍 Monitoring

The operator exposes metrics that are scraped by Prometheus through the ServiceMonitor configured in the operator namespace.

## 📚 References

- [RabbitMQ Cluster Operator](https://www.rabbitmq.com/kubernetes/operator/operator-overview.html)
- [Bitnami Helm Chart](https://github.com/bitnami/charts/tree/main/bitnami/rabbitmq-cluster-operator)

