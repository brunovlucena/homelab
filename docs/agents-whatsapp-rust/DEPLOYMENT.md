# Agent WhatsApp Rust - Deployment Guide

> **Version**: 1.0.0  
> **Status**: Production-Ready

## Prerequisites

- Kubernetes cluster (v1.24+)
- Knative Serving & Eventing installed
- MongoDB replica set (3 nodes)
- Redis (sentinel/cluster mode)
- MinIO/S3 for media storage
- Ingress controller with session affinity support

## Quick Start

```bash
# Clone repository
git clone <repo-url>
cd agents-whatsapp-rust

# Build Docker images
docker build -t messaging-service:latest ./services/messaging-service
docker build -t user-service:latest ./services/user-service
docker build -t agent-gateway:latest ./services/agent-gateway
docker build -t media-service:latest ./services/media-service

# Apply Kubernetes manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/mongodb/
kubectl apply -f k8s/redis/
kubectl apply -f k8s/services/
kubectl apply -f k8s/knative/
```

## Service Deployment

### Messaging Service

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: messaging-service
  namespace: homelab-services
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "2"
        autoscaling.knative.dev/maxScale: "10"
        autoscaling.knative.dev/target: "100"
    spec:
      containers:
      - image: messaging-service:latest
        env:
        - name: MONGODB_URI
          value: "mongodb://mongodb.homelab-services.svc.cluster.local:27017"
        - name: REDIS_URI
          value: "redis://redis.homelab-services.svc.cluster.local:6379"
        - name: RUST_LOG
          value: "info"
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2000m"
            memory: "2Gi"
```

### User Service

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: user-service
  namespace: homelab-services
spec:
  template:
    spec:
      containers:
      - image: user-service:latest
        env:
        - name: MONGODB_URI
          value: "mongodb://mongodb.homelab-services.svc.cluster.local:27017"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
```

### Agent Gateway

```yaml
apiVersion: lambda.knative.dev/v1alpha1
kind: Function
metadata:
  name: agent-gateway
  namespace: knative-lambda
spec:
  image: agent-gateway:latest
  env:
  - name: MONGODB_URI
    value: "mongodb://mongodb.homelab-services.svc.cluster.local:27017"
  - name: REDIS_URI
    value: "redis://redis.homelab-services.svc.cluster.local:6379"
  - name: KUBERNETES_NAMESPACE
    value: "ai-agents"
```

### Media Service

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: media-service
  namespace: homelab-services
spec:
  template:
    spec:
      containers:
      - image: media-service:latest
        env:
        - name: MINIO_ENDPOINT
          value: "minio.homelab-services.svc.cluster.local:9000"
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
```

## Ingress Configuration

### Session Affinity for WebSocket

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: messaging-ingress
  namespace: homelab-services
  annotations:
    nginx.ingress.kubernetes.io/affinity: "cookie"
    nginx.ingress.kubernetes.io/affinity-mode: "persistent"
    nginx.ingress.kubernetes.io/session-cookie-name: "messaging-session"
    nginx.ingress.kubernetes.io/session-cookie-expires: "3600"
    nginx.ingress.kubernetes.io/session-cookie-max-age: "3600"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - messaging.example.com
    secretName: messaging-tls
  rules:
  - host: messaging.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: messaging-service
            port:
              number: 80
```

## MongoDB Setup

### Replica Set Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mongodb-config
  namespace: homelab-services
data:
  mongod.conf: |
    replication:
      replSetName: "rs0"
    storage:
      wiredTiger:
        engineConfig:
          cacheSizeGB: 2
```

### Collections & Indexes

```javascript
// Connect to MongoDB
use messaging_app

// Create collections
db.createCollection("users")
db.createCollection("conversations")
db.createCollection("messages")
db.createCollection("idempotency_keys")
db.createCollection("sequence_numbers")

// Create indexes
db.messages.createIndex({ conversation_id: 1, sequence_number: 1 })
db.messages.createIndex({ message_id: 1 }, { unique: true })
db.messages.createIndex({ idempotency_key: 1 }, { unique: true })
db.idempotency_keys.createIndex({ idempotency_key: 1 }, { unique: true })
db.idempotency_keys.createIndex({ createdAt: 1 }, { expireAfterSeconds: 86400 })
db.sequence_numbers.createIndex({ conversation_id: 1 }, { unique: true })
```

## Redis Setup

### Sentinel Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-sentinel-config
  namespace: homelab-services
data:
  sentinel.conf: |
    sentinel monitor mymaster redis-master 6379 2
    sentinel down-after-milliseconds mymaster 5000
    sentinel failover-timeout mymaster 10000
```

## Knative Trigger Configuration

### Agent Response Trigger

```yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: messaging-service-agent-responses
  namespace: homelab-services
spec:
  broker: default
  filter:
    attributes:
      type: agent.response
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: messaging-service
```

## Health Checks

### Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Monitoring

### Prometheus ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: messaging-service
  namespace: homelab-services
spec:
  selector:
    matchLabels:
      app: messaging-service
  endpoints:
  - port: metrics
    interval: 30s
```

## Secrets Management

### JWT Secret

```bash
kubectl create secret generic jwt-secret \
  --from-literal=secret=$(openssl rand -base64 32) \
  -n homelab-services
```

### MinIO Credentials

```bash
kubectl create secret generic minio-credentials \
  --from-literal=access-key=minioadmin \
  --from-literal=secret-key=minioadmin \
  -n homelab-services
```

## Verification

### Check Services

```bash
# Check Knative services
kubectl get ksvc -n homelab-services

# Check pods
kubectl get pods -n homelab-services

# Check logs
kubectl logs -f deployment/messaging-service -n homelab-services
```

### Test WebSocket Connection

```bash
# Test WebSocket endpoint
wscat -c wss://messaging.example.com/ws
```

## Troubleshooting

### Connection Issues

- Check Redis connection registry: `redis-cli KEYS connection:*`
- Check MongoDB connectivity: `kubectl exec -it mongodb-0 -- mongosh`
- Check Ingress session affinity: `kubectl describe ingress messaging-ingress`

### Message Delivery Issues

- Check idempotency keys: `db.idempotency_keys.find()`
- Check sequence numbers: `db.sequence_numbers.find()`
- Check Redis Pub/Sub: `redis-cli PUBSUB CHANNELS`

### Performance Issues

- Check connection count: `kubectl top pods -n homelab-services`
- Check MongoDB indexes: `db.messages.getIndexes()`
- Check Redis memory: `redis-cli INFO memory`

## Status

🟢 **Ready for Deployment**
