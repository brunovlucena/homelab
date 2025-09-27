# 🏃‍♂️ GitHub Actions Self-Hosted Runner

> **Self-hosted GitHub Actions runners for your home infrastructure**

This component provides self-hosted GitHub Actions runners running in your Kubernetes cluster, ensuring your CI/CD data stays within your infrastructure and is not used for GitHub's training purposes.

## 🎯 Overview

The GitHub Actions runner setup includes:
- **Actions Runner Controller**: Manages runner lifecycle in Kubernetes
- **Runner Deployments**: Configurable runner instances with resource limits
- **Sealed Secrets**: Secure storage of GitHub tokens and configuration
- **Monitoring**: Integration with Prometheus for metrics collection
- **Security**: Non-root containers with proper security contexts

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GitHub        │    │   Kubernetes    │    │   Self-Hosted   │
│   Actions       │───▶│   Cluster       │───▶│   Runners       │
│   Workflows     │    │   (Home)        │    │   (Ephemeral)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Workflow      │    │   Runner        │    │   Build         │
│   Execution     │    │   Controller    │    │   Artifacts     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster with Flux GitOps
- GitHub token with `repo` and `admin:org` permissions
- Sealed Secrets controller installed
- kubectl and kubeseal CLI tools

### Setup Steps

1. **Create GitHub Token**:
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Create a token with `repo` and `admin:org` permissions
   - Copy the token for the next step

2. **Create Runner Secrets**:
   ```bash
   cd flux/clusters/studio/infrastructure/github-actions-runner
   chmod +x create-secrets.sh
   ./create-secrets.sh
   ```

3. **Apply the Configuration**:
   ```bash
   # Apply the sealed secret
   kubectl apply -f github-actions-runner-sealed-secret.yaml
   
   # The HelmRelease and RunnerDeployment will be applied by Flux
   ```

4. **Verify Installation**:
   ```bash
   # Check controller status
   kubectl get pods -n github-actions-runner
   
   # Check runner status
   kubectl get runners -n github-actions-runner
   
   # View runner logs
   kubectl logs -n github-actions-runner -l app.kubernetes.io/name=github-actions-runner
   ```

## 🔧 Configuration

### Runner Configuration

The runners are configured with:
- **Image**: `summerwind/actions-runner:ubuntu-22.04`
- **Resources**: 100m CPU, 128Mi memory (requests), 2000m CPU, 4Gi memory (limits)
- **Security**: Non-root user (1000:1000)
- **Volumes**: Ephemeral storage for work and temp directories
- **Docker**: Access to host Docker socket for container builds

### Labels and Groups

Default configuration:
- **Labels**: `self-hosted,linux,x64,home-infrastructure`
- **Group**: `default`
- **Repository**: Configurable via secret

### Resource Limits

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 2000m
    memory: 4Gi
```

## 🔒 Security

### Security Context

- **Non-root execution**: All containers run as user 1000
- **Read-only root filesystem**: Prevents unauthorized modifications
- **No privilege escalation**: Security context prevents privilege escalation
- **Capability dropping**: All Linux capabilities are dropped

### Secret Management

- **Sealed Secrets**: GitHub tokens are encrypted using Sealed Secrets
- **Namespace isolation**: Secrets are scoped to the runner namespace
- **RBAC**: Proper role-based access control is configured

### Network Security

- **Pod security standards**: Enforces restricted pod security
- **Network policies**: Can be configured for additional isolation
- **Service mesh**: Compatible with Istio for advanced networking

## 📊 Monitoring

### Metrics

The runner controller exposes Prometheus metrics:
- Runner registration status
- Job execution metrics
- Resource utilization
- Error rates and failures

### Grafana Dashboard

A Grafana dashboard is available to monitor:
- Active runners
- Job execution times
- Resource usage
- Error rates
- Runner lifecycle events

### Logs

Runner logs are collected and can be viewed via:
```bash
# Controller logs
kubectl logs -n github-actions-runner -l app.kubernetes.io/name=actions-runner-controller

# Runner logs
kubectl logs -n github-actions-runner -l app.kubernetes.io/name=github-actions-runner
```

## 🛠️ Usage in Workflows

### Basic Usage

```yaml
name: CI/CD Pipeline
on: [push, pull_request]

jobs:
  build:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v4
      - name: Build
        run: make build
      - name: Test
        run: make test
```

### Advanced Usage with Labels

```yaml
name: Advanced CI/CD
on: [push, pull_request]

jobs:
  build:
    runs-on: [self-hosted, linux, x64, home-infrastructure]
    steps:
      - uses: actions/checkout@v4
      - name: Build with Docker
        run: |
          docker build -t myapp .
          docker run myapp
```

### Matrix Strategy

```yaml
name: Matrix Build
on: [push, pull_request]

jobs:
  test:
    runs-on: self-hosted
    strategy:
      matrix:
        node-version: [16, 18, 20]
    steps:
      - uses: actions/checkout@v4
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
      - name: Test
        run: npm test
```

## 🔄 Maintenance

### Scaling Runners

To scale the number of runners:
```bash
kubectl scale runnerdeployment github-actions-runner-deployment -n github-actions-runner --replicas=5
```

### Updating Runner Image

Update the runner image in the HelmRelease values:
```yaml
runner:
  image:
    tag: "ubuntu-24.04"  # Update to newer version
```

### Token Rotation

To rotate GitHub tokens:
1. Create a new token in GitHub
2. Update the secret using the create-secrets.sh script
3. Apply the new sealed secret
4. Delete the old token in GitHub

## 🐛 Troubleshooting

### Common Issues

1. **Runner Not Registering**:
   ```bash
   # Check secret configuration
   kubectl get secret github-actions-runner-secret -n github-actions-runner -o yaml
   
   # Check runner logs
   kubectl logs -n github-actions-runner -l app.kubernetes.io/name=github-actions-runner
   ```

2. **Permission Denied**:
   ```bash
   # Check security context
   kubectl describe pod -n github-actions-runner -l app.kubernetes.io/name=github-actions-runner
   ```

3. **Resource Limits**:
   ```bash
   # Check resource usage
   kubectl top pods -n github-actions-runner
   
   # Adjust limits in HelmRelease if needed
   ```

### Useful Commands

```bash
# Check runner status
kubectl get runners -n github-actions-runner

# View runner events
kubectl get events -n github-actions-runner --sort-by='.lastTimestamp'

# Check controller logs
kubectl logs -n github-actions-runner -l app.kubernetes.io/name=actions-runner-controller -f

# Force runner restart
kubectl delete pods -n github-actions-runner -l app.kubernetes.io/name=github-actions-runner
```

## 📈 Performance Tuning

### Resource Optimization

- **CPU**: Adjust based on workload (default: 2000m limit)
- **Memory**: Increase for memory-intensive builds (default: 4Gi limit)
- **Storage**: Use faster storage classes for better I/O performance

### Scaling Strategy

- **Horizontal scaling**: Increase replica count for more parallel jobs
- **Vertical scaling**: Increase resource limits for larger builds
- **Auto-scaling**: Configure HPA for dynamic scaling based on demand

## 🔗 Integration

### With Existing Services

The runners integrate with your existing home infrastructure:
- **Prometheus**: Metrics collection and monitoring
- **Grafana**: Visualization and alerting
- **Loki**: Centralized logging
- **Tempo**: Distributed tracing
- **Istio**: Service mesh networking

### With External Services

- **Docker Registry**: Push/pull container images
- **Artifact Storage**: Store build artifacts
- **Notification**: Slack/email notifications for build status

## 📄 License

This component is part of the home infrastructure project and is licensed under the MIT License.

---

**Built with ❤️ for privacy-focused CI/CD**
