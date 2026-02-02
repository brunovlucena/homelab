# 🛠️ Installation Guide

Complete step-by-step guide to installing Knative Lambda Builder in your environment.

---

## 📋 Prerequisites

### Required
- ✅ Kubernetes cluster (v1.34+)
- ✅ `kubectl` configured and authenticated
- ✅ Knative Serving installed
- ✅ RabbitMQ cluster deployed
- ✅ AWS ECR access (for image registry)

### Recommended
- ✅ Prometheus Operator (for monitoring)
- ✅ Grafana (for dashboards)
- ✅ Tempo (for distributed tracing)
- ✅ Flux CD or ArgoCD (for GitOps)

---

## 🚀 Quick Install (Recommended)

### Option 1: Flux CD + Makefile (GitOps)

```bash
# Clone repository
git clone https://github.com/brunovlucena/homelab
cd homelab

# Setup environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export AWS_ACCESS_KEY_ID="your_access_key"
export AWS_SECRET_ACCESS_KEY="your_secret_key"
export AWS_REGION="us-west-2"

# Deploy using Flux (automated)
make up

# Verify deployment
flux get kustomizations -n flux-system | grep knative-lambda
kubectl get pods -n knative-lambda
```

### Option 2: ArgoCD (GitOps)

```bash
# Apply ArgoCD application
kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: knative-lambda
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/brunovlucena/homelab
    targetRevision: main
    path: flux/clusters/homelab/infrastructure/knative-lambda/k8s/chart
    helm:
      releaseName: knative-lambda
      values: |
        environment: dev
        namespace:
          create: true
          name: knative-lambda
        image:
          registry: "339954290315.dkr.ecr.us-west-2.amazonaws.com"
          repository: "knative-lambda/builder"
          tag: "dev"
        sidecar:
          image:
            registry: "339954290315.dkr.ecr.us-west-2.amazonaws.com"
            repository: "knative-lambda/sidecar"
            tag: "dev"
  destination:
    server: https://kubernetes.default.svc
    namespace: knative-lambda
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
EOF

# Verify deployment
kubectl get application knative-lambda -n argocd
kubectl get pods -n knative-lambda
```

### Option 3: Manual Helm (for testing)

```bash
# Navigate to knative-lambda directory
cd flux/clusters/homelab/infrastructure/knative-lambda

# Build and push images
make docker-login
make docker-build
make docker-push

# Deploy using Helm
helm install knative-lambda k8s/chart/ \
  --namespace knative-lambda \
  --create-namespace \
  --set environment=dev \
  --set image.registry=339954290315.dkr.ecr.us-west-2.amazonaws.com \
  --set image.repository=knative-lambda/builder \
  --set image.tag=dev

# Verify deployment
kubectl get pods -n knative-lambda
kubectl get ksvc -n knative-lambda
```

---

## 📝 Detailed Installation Steps

### Step 1: Setup Environment Variables

Add required environment variables to your `~/.zshrc`:

```bash
# AWS Configuration
export AWS_ACCESS_KEY_ID="your_access_key"
export AWS_SECRET_ACCESS_KEY="your_secret_key"
export AWS_REGION="us-west-2"
export AWS_ACCOUNT_ID="339954290315"

# GitHub (for Flux)
export GITHUB_TOKEN="ghp_your_token_here"
export GITHUB_USERNAME="your-github-username"

# ECR Registry
export ECR_REGISTRY="339954290315.dkr.ecr.us-west-2.amazonaws.com"
export ECR_REPOSITORY="knative-lambdas"

# Environment
export ENV="dev"  # or "prd" for production
```

Then reload your shell:
```bash
source ~/.zshrc
```

### Step 2: Deploy Prerequisites

#### Install Knative Serving

```bash
# Install Knative Serving
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.15.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.15.0/serving-core.yaml

# Install networking layer (Istio)
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.15.0/net-istio.yaml

# Verify installation
kubectl get pods -n knative-serving
```

#### Deploy RabbitMQ

```bash
# Install RabbitMQ Operator
kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml

# Create RabbitMQ cluster
kubectl apply -f - <<EOF
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: rabbitmq-cluster-dev
  namespace: rabbitmq-dev
spec:
  replicas: 1
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
EOF

# Wait for ready
kubectl wait -n rabbitmq-dev \
  --for=condition=ready pod \
  -l app.kubernetes.io/name=rabbitmq-cluster-dev \
  --timeout=300s
```

### Step 3: Deploy Knative Lambda

#### Using Flux CD (Recommended)

```bash
# Deploy entire homelab infrastructure (includes knative-lambda)
make up

# Or force reconcile just knative-lambda
flux reconcile kustomization knative-lambda -n flux-system
```

#### Using ArgoCD

```bash
# Apply the ArgoCD application (see Option 2 above)
kubectl apply -f knative-lambda-argocd-app.yaml

# Check sync status
argocd app get knative-lambda
argocd app sync knative-lambda
```

#### Using Makefile (Development)

```bash
# Navigate to knative-lambda directory
cd flux/clusters/homelab/infrastructure/knative-lambda

# Build and deploy
make docker-login
make docker-build
make docker-push
make deploy ENV=dev

# Check status
kubectl get pods -n knative-lambda
kubectl get ksvc -n knative-lambda
```

### Step 4: Verify Installation

```bash
# Check all resources
kubectl get all -n knative-lambda

# Check Knative services
kubectl get ksvc -n knative-lambda

# Check builder logs
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=100

# Check sidecar logs
kubectl logs -n knative-lambda -l app=knative-lambda-sidecar --tail=100
```

**Expected Output**:
```
✅ Pod running (1/1 Ready)
✅ Knative service created
✅ RabbitMQ connection established
✅ ECR authentication successful
✅ Logs show: "Knative Lambda Builder ready"
```

---

## 🔧 Configuration

### Environment Variables

Edit `k8s/chart/values.yaml` to customize:

```yaml
# Environment configuration
environment: dev  # dev, prd, local

# Image configuration
image:
  registry: "339954290315.dkr.ecr.us-west-2.amazonaws.com"
  repository: "knative-lambda/builder"
  tag: "dev"
  pullPolicy: IfNotPresent

# Sidecar configuration
sidecar:
  image:
    registry: "339954290315.dkr.ecr.us-west-2.amazonaws.com"
    repository: "knative-lambda/sidecar"
    tag: "dev"
    pullPolicy: IfNotPresent

# Builder service configuration
builder:
  name: knative-lambda-service-builder
  minScale: 0
  maxScale: 10
  targetConcurrency: 5
  scaleToZeroGracePeriod: "30s"

# RabbitMQ configuration
rabbitmq:
  clusterName: "rabbitmq-cluster-dev"
  namespace: "rabbitmq-dev"
  connectionSecretName: "rabbitmq-connection"

# AWS configuration
aws:
  accountId: "339954290315"
  region: "us-west-2"
```

### Resource Limits

Adjust based on your cluster size:

```yaml
resources:
  requests:
    cpu: 100m      # Minimal for scale-to-zero
    memory: 256Mi
  limits:
    cpu: 1000m     # Burst for heavy builds
    memory: 2Gi
```

---

## 🧪 Post-Installation Testing

### Test 1: Health Check

```bash
# Port-forward to service
kubectl port-forward -n knative-lambda svc/knative-lambda-builder 8080:8080

# Check health endpoint
curl http://localhost:8080/health
```

**Expected**: `{"status": "healthy", "version": "1.0.0"}`

### Test 2: Deploy Test Function

```bash
# Create a simple Python function
cat > test-function.py <<EOF
import json

def handler(event):
    return {
        'statusCode': 200,
        'body': json.dumps({
            'message': 'Hello from Knative Lambda!',
            'event': event
        })
    }
EOF

# Upload to S3
export PARSER_ID="test-function-$(uuidgen | tr '[:upper:]' '[:lower:]')"
aws s3 cp test-function.py s3://knative-lambda-fusion-modules-tmp/global/parser/${PARSER_ID}

# Trigger build event
cd tests
ENV=dev uv run --python 3.9 python create-event-builder.py

# Monitor build
kubectl get jobs -n knative-lambda -w
```

### Test 3: Function Invocation

```bash
# Get function URL
FUNCTION_URL=$(kubectl get ksvc -n knative-lambda -o jsonpath='{.items[0].status.url}')

# Test function
curl -X POST $FUNCTION_URL \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

---

## 🔄 Integration Setup

### RabbitMQ Integration

Configure RabbitMQ queues for event processing:

```bash
# Check queue status
make rabbitmq-status ENV=dev

# Purge queues if needed
make rabbitmq-purge ENV=dev
```

### AWS ECR Integration

Ensure ECR authentication:

```bash
# Login to ECR
make docker-login

# Verify access
aws ecr describe-repositories --region us-west-2
```

### Monitoring Integration

Configure Prometheus monitoring:

```bash
# Port-forward to Prometheus
make pf-prometheus

# Check metrics
curl http://localhost:9090/api/v1/query?query=knative_lambda_builds_total
```

---

## ⚙️ Advanced Configuration

### Multi-Environment Setup

Deploy to different environments:

```bash
# Development
make deploy ENV=dev

# Production
make deploy ENV=prd

# Local testing
make deploy ENV=local
```

### High Availability

Run multiple replicas with leader election:

```yaml
replicas: 3

env:
  - name: LEADER_ELECTION_ENABLED
    value: "true"
  - name: LEADER_ELECTION_NAMESPACE
    value: "knative-lambda"
```

### Custom Storage

Use different S3 buckets per environment:

```yaml
env:
  - name: S3_SOURCE_BUCKET
    value: "knative-lambda-$(ENV)-fusion-modules-tmp"
  - name: S3_TMP_BUCKET
    value: "knative-lambda-$(ENV)-context-tmp"
```

---

## 🔍 Troubleshooting

### Pod CrashLoopBackOff

```bash
# Check logs
kubectl logs -n knative-lambda -l app=knative-lambda-builder --previous

# Common causes:
# - Missing AWS credentials
# - ECR authentication failed
# - RabbitMQ connection failed
# - Invalid S3 bucket configuration
```

### Build Jobs Failing

```bash
# Check Kaniko job logs
kubectl logs -n knative-lambda -l job-name=kaniko --tail=100

# Check build events
kubectl get events -n knative-lambda --sort-by='.lastTimestamp'
```

### Flux Not Reconciling

```bash
# Force reconcile all Flux resources
make flux-refresh

# Reconcile specific component
flux reconcile kustomization knative-lambda -n flux-system
```

### ArgoCD Sync Issues

```bash
# Check application status
argocd app get knative-lambda

# Force sync
argocd app sync knative-lambda

# Check logs
argocd app logs knative-lambda
```

---

## 🗑️ Uninstallation

### Flux CD

```bash
# Delete via Kustomize
kubectl delete -k flux/clusters/homelab/infrastructure/knative-lambda

# Or delete namespace (removes everything)
kubectl delete namespace knative-lambda
```

### ArgoCD

```bash
# Delete application
kubectl delete application knative-lambda -n argocd

# Or delete namespace
kubectl delete namespace knative-lambda
```

### Manual Helm

```bash
# Uninstall Helm release
helm uninstall knative-lambda -n knative-lambda

# Delete namespace
kubectl delete namespace knative-lambda
```

**Note**: Flux/ArgoCD will automatically try to recreate deleted resources. To permanently remove knative-lambda, remove it from the GitOps manifests.

---

## 📚 Next Steps

✅ **Installation complete!** Now:

1. **[First Steps](FIRST_STEPS.md)** - Deploy your first function
2. **[SRE Workflows](../03-for-engineers/sre/WORKFLOWS.md)** - Learn monitoring and debugging
3. **[Development Guide](../03-for-engineers/backend/README.md)** - Build custom functions

---

## 💬 Need Help?

- **Slack**: `#knative-lambda`
- **Issues**: [GitHub Issues](https://github.com/brunovlucena/homelab/issues)
- **Docs**: [Full documentation](../)

---

**Next**: [First Steps](FIRST_STEPS.md) →
