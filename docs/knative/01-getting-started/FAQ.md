# ❓ Frequently Asked Questions (FAQ)

**Common questions and answers about Knative Lambda**

---

## 📑 Table of Contents

- [General Questions](#general-questions)
- [Architecture & Design](#architecture--design)
- [Deployment & Operations](#deployment--operations)
- [Development](#development)
- [Performance & Scaling](#performance--scaling)
- [Troubleshooting](#troubleshooting)
- [Security](#security)
- [Cost & Pricing](#cost--pricing)

---

## General Questions

### What is Knative Lambda?

Knative Lambda is an open-source serverless platform that automatically builds, deploys, and scales containerized functions on Kubernetes. Upload code to S3, get a running auto-scaling function—no Dockerfiles or manual deployments needed.

**Key features:**
- Automatic container builds (Kaniko)
- Auto-scaling (Knative Serving)
- Event-driven (CloudEvents + RabbitMQ)
- Multi-language support (Python, Node.js, Go)

→ **[Overview](OVERVIEW.md)**

---

### How is it different from AWS Lambda?

| Feature | Knative Lambda | AWS Lambda |
|---------|----------------|------------|
| **Infrastructure** | Your Kubernetes cluster | AWS-managed |
| **Vendor lock-in** | ❌ None | ✅ AWS |
| **Cold start** | <5s | <1s |
| **Pricing** | Cluster costs only | Per-invocation |
| **Control** | Full customization | Limited config |
| **Multi-cloud** | ✅ Yes | ❌ AWS only |
| **Open source** | ✅ Yes | ❌ Proprietary |

**Use Knative Lambda if:**
- You want infrastructure control
- Multi-cloud or hybrid deployments
- Avoiding vendor lock-in
- Custom runtime requirements

**Use AWS Lambda if:**
- Fastest cold start critical
- No Kubernetes expertise
- Prefer managed services
- AWS-native architecture

---

### What languages are supported?

**Out-of-the-box:**
- **Python** (3.8, 3.9, 3.10, 3.11)
- **Node.js** (16, 18, 20)
- **Go** (1.20, 1.21, 1.22)

**Custom runtimes:**
Any language with a Dockerfile-based build. See [Multi-Language Strategy](../07-decisions/MULTI_LANGUAGE_STRATEGY.md).

---

### Do I need to know Kubernetes?

**Basic use**: No. Just upload code and trigger builds.

**Advanced use**: Yes, for:
- Debugging deployment issues
- Custom networking/security policies
- Resource optimization
- Production operations

**Recommended knowledge:**
- kubectl basics
- Pod concepts
- Service/Ingress fundamentals

→ **[SRE Guide](../03-for-engineers/sre/README.md)**

---

## Architecture & Design

### How does the build process work?

```
1. Upload code to S3
   ↓
2. Send CloudEvent (build.start) to RabbitMQ
   ↓
3. Builder Service receives event
   ↓
4. Creates Kaniko Job in Kubernetes
   ↓
5. Kaniko builds container image (no Docker daemon)
   ↓
6. Pushes image to ECR
   ↓
7. Creates Knative Service
   ↓
8. Function deployed and auto-scaling!
```

→ **[Build Pipeline](../04-architecture/BUILD_PIPELINE.md)**

---

### Why Kaniko instead of Docker?

**Security & Compatibility:**
- ✅ No Docker daemon (more secure)
- ✅ Runs in Kubernetes (no privileged containers)
- ✅ Consistent builds (reproducible)
- ✅ Multi-platform support

→ **[Why Kaniko?](../07-decisions/WHY_KANIKO.md)**

---

### Why CloudEvents?

**Standards-based event processing:**
- ✅ Vendor-neutral format
- ✅ Multi-cloud portability
- ✅ Rich ecosystem support
- ✅ Better debugging/tracing

→ **[Why CloudEvents?](../07-decisions/WHY_CLOUDEVENTS.md)**

---

### Why RabbitMQ instead of Kafka?

**Simplicity & Features:**
- ✅ Easier to operate
- ✅ Lower resource footprint
- ✅ Better message routing
- ✅ Native CloudEvents support

For high-throughput streaming, Kafka might be better.

→ **[Message Queue Comparison](../07-decisions/WHY_RABBITMQ.md)**

---

## Deployment & Operations

### How do I deploy to production?

**Recommended approach:**

1. **Use Helm** with environment-specific values
2. **GitOps (Flux CD)** for automated deployments
3. **Semantic versioning** for images
4. **Canary deployments** with Flagger

```bash
# Production deployment
helm install knative-lambda deploy/ \
  --namespace knative-lambda \
  --values deploy/overlays/prd/values-prd.yaml
```

→ **[Deployment Guide](../03-for-engineers/devops/DEPLOYMENT.md)**

---

### Can I run this locally?

**Yes!** Use kind or Docker Desktop:

```bash
# Create local kind cluster
kind create cluster --name knative-lambda

# Install dependencies
make install-deps-local

# Deploy locally
helm install knative-lambda deploy/ \
  --values deploy/overlays/local/values-local.yaml
```

→ **[Local Development](../06-development/LOCAL_DEVELOPMENT.md)**

---

### How do I handle secrets?

**Options:**

1. **Kubernetes Secrets** (built-in)
2. **External Secrets Operator** (recommended)
3. **AWS Secrets Manager** (via IRSA)
4. **HashiCorp Vault**

```yaml
# Example: External Secret
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: parser-secrets
spec:
  secretStoreRef:
    name: aws-secrets-manager
  target:
    name: parser-secrets
  data:
    - secretKey: api-key
      remoteRef:
        key: prod/parser/api-key
```

→ **[Secrets Management](../05-operations/SECRETS_MANAGEMENT.md)**

---

### How do I monitor functions?

**Built-in observability:**

**Metrics** (Prometheus):
- `knative_lambda_build_duration_seconds`
- `serving_revision_request_count`
- `serving_revision_request_latencies`

**Logs** (Structured JSON):
```bash
kubectl logs -l serving.knative.dev/service=my-function
```

**Traces** (OpenTelemetry):
- Distributed tracing via Tempo
- End-to-end request tracking

**Dashboards** (Grafana):
- Pre-built dashboards in `dashboards/`

→ **[Observability](../04-architecture/OBSERVABILITY.md)**

---

## Development

### How do I test functions locally?

**Option 1: Direct Python execution**
```bash
# Test function logic directly
python3 -c "
from parser import handler
result = handler({'type': 'test', 'data': {}})
print(result)
"
```

**Option 2: Local container**
```bash
# Build and run locally
docker build -t my-function .
docker run -p 8080:8080 my-function

# Test
curl http://localhost:8080
```

**Option 3: Full Knative Serving**
```bash
# Deploy to local cluster
kn service create my-function \
  --image my-function:latest \
  --scale-min 1
```

→ **[Testing Strategy](../06-development/TESTING_STRATEGY.md)**

---

### Can I use private Python packages?

**Yes!** Configure pip credentials:

**Option 1: Build-time secret**
```dockerfile
# Custom Dockerfile (advanced)
ARG PYPI_TOKEN
RUN pip install --extra-index-url https://${PYPI_TOKEN}@pypi.company.com/simple/ -r requirements.txt
```

**Option 2: Vendor dependencies**
```bash
# Download deps locally, upload to S3
pip download -r requirements.txt -d ./vendor/
aws s3 cp vendor/ s3://bucket/parser/${PARSER_ID}/vendor/ --recursive
```

---

### How do I debug build failures?

```bash
# Get Kaniko job logs
kubectl logs job/kaniko-<job-name> -n knative-lambda

# Common issues:
# - Syntax errors: Fix in parser.py
# - Missing deps: Check requirements.txt
# - S3 access denied: Verify IAM permissions
# - ECR push failed: Check ECR policy
```

**Enable debug logging:**
```yaml
# values-dev.yaml
builderService:
  env:
    - name: LOG_LEVEL
      value: "debug"
```

→ **[Troubleshooting](../05-operations/TROUBLESHOOTING.md)**

---

## Performance & Scaling

### How fast is the cold start?

**Typical cold starts:**
- **Python**: 3-5 seconds
- **Node.js**: 2-4 seconds
- **Go**: 1-3 seconds

**Factors affecting cold start:**
- Image size (use slim base images)
- Dependencies (minimize requirements)
- Resource limits (more CPU = faster)
- Image registry location (use regional)

**Optimization tips:**
```yaml
# Keep warm with min instances
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: my-function
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "1"  # Keep 1 pod warm
```

---

### How many requests can it handle?

**Depends on configuration:**

**Default (conservative):**
- 100 concurrent requests per pod
- Auto-scales to 10 pods
- **Capacity**: ~1,000 concurrent requests

**Optimized (production):**
- 200 concurrent requests per pod
- Auto-scales to 50 pods
- **Capacity**: ~10,000 concurrent requests

**Configuration:**
```yaml
# Increase capacity
apiVersion: serving.knative.dev/v1
kind: Service
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/target: "200"
        autoscaling.knative.dev/max-scale: "50"
```

→ **[Autoscaling Optimization](../03-for-engineers/sre/user-stories/SRE-005-autoscaling-optimization.md)**

---

### Does it support bursting?

**Yes!** Knative Serving scales rapidly:

```
Traffic spike: 0 → 1000 req/s
  ↓
Scale: 0 pods → 1 pod (5s)
  ↓
Scale: 1 pod → 5 pods (10s)
  ↓
Scale: 5 pods → 10 pods (15s)
  ↓
Stabilize: 10 pods handling load
```

**Burst configuration:**
```yaml
autoscaling.knative.dev/activation-scale: "10"  # Initial burst
autoscaling.knative.dev/scale-down-delay: "30s"  # Wait before scaling down
```

---

## Troubleshooting

### My function isn't scaling

**Check autoscaler configuration:**
```bash
# View autoscaler config
kubectl get cm config-autoscaler -n knative-serving -o yaml

# View service autoscaling status
kubectl get kpa -n knative-lambda
```

**Common issues:**
- Resource limits too low
- HPA not enabled
- Metrics not reporting
- Network latency

→ **[SRE Runbooks](../03-for-engineers/sre/RUNBOOKS.md)**

---

### Build is slow

**Optimize build time:**

1. **Use smaller base images:**
   ```python
   # python:3.9-slim instead of python:3.9
   ```

2. **Minimize dependencies:**
   ```txt
   # requirements.txt - only what you need
   requests==2.31.0
   # Don't include: pandas, numpy (large)
   ```

3. **Increase Kaniko resources:**
   ```yaml
   kanikoJob:
     resources:
       requests:
         cpu: "1000m"  # More CPU = faster builds
   ```

4. **Use build caching:** (coming soon)

---

### Function returns errors

**Debugging steps:**

```bash
# 1. Check function logs
kubectl logs -l serving.knative.dev/service=my-function --tail=100

# 2. Check for crashes
kubectl get pods -l serving.knative.dev/service=my-function

# 3. Describe pod for events
kubectl describe pod <pod-name>

# 4. Port-forward for direct access
kubectl port-forward svc/my-function 8080:80

# 5. Test directly
curl http://localhost:8080 -H "ce-type: test"
```

---

## Security

### Is this secure for production?

**Yes**, with proper configuration:

✅ **Build security:**
- Kaniko (no Docker daemon)
- Non-root containers
- Image scanning (Trivy)

✅ **Runtime security:**
- RBAC (least privilege)
- Network policies
- Pod security standards
- Resource quotas

✅ **Data security:**
- TLS/mTLS communication
- Secrets encryption at rest
- IRSA for AWS access (no static creds)

→ **[Security Assessment](../08-assessments/SECURITY_ASSESSMENT.md)**

---

### How do I restrict access?

**Network policies:**
```yaml
# Only allow from specific namespaces
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: function-access
spec:
  podSelector:
    matchLabels:
      app: my-function
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: allowed-namespace
```

**AuthN/AuthZ:**
- OAuth2 Proxy
- Istio AuthorizationPolicy
- OPA (Open Policy Agent)

---

## Cost & Pricing

### How much does it cost?

**Infrastructure costs only:**

**Small deployment (dev):**
- 3 nodes (t3.medium): ~$90/month
- RabbitMQ: ~2 GB RAM
- **Total**: ~$90/month + data transfer

**Production deployment:**
- 10 nodes (t3.large): ~$600/month
- RabbitMQ cluster: ~8 GB RAM
- **Total**: ~$600/month + data transfer

**Compare to AWS Lambda:**
- 10M requests/month: ~$200
- But locked to AWS

**Savings with scale-to-zero:**
- Idle functions: $0/hour
- Only pay for cluster, not per-function

---

### Is scale-to-zero really $0?

**Almost:**
- No pod costs when idle
- Still pay for:
  - Cluster control plane
  - Knative Serving (minimal)
  - RabbitMQ (minimal)

**Cost breakdown:**
```
Scale-to-zero function:
- Active (10% of time): 0.1 * pod_cost
- Idle (90% of time): $0
- Total: ~10% of always-on cost
```

---

## Still Have Questions?

| Channel | Best For |
|---------|----------|
| **[Slack `#knative-lambda`](https://slack.company.com)** | Quick questions |
| **[GitHub Issues](https://github.com/brunovlucena/homelab/issues)** | Bug reports, feature requests |
| **[Documentation](README.md)** | In-depth guides |

---

**Didn't find your answer?**

- **Troubleshooting**: [Operations Guide](../05-operations/TROUBLESHOOTING.md)
- **Architecture**: [System Design](../04-architecture/SYSTEM_DESIGN.md)
- **Development**: [Development Guide](../06-development/README.md)

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

