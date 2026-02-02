# 🔄 DEVOPS-005: Infrastructure as Code

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-237/devops-005-infrastructure-as-code

---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** manage all infrastructure as code using Helm and Kustomize  
**So that** infrastructure is version-controlled, auditable, and reproducible across environments

---

## 🎯 Acceptance Criteria

### ✅ Helm Chart Management
- [ ] Complete Helm chart for all components
- [ ] Parameterized values for different environments
- [ ] Chart versioning with semantic versioning
- [ ] Helm chart testing and validation
- [ ] Chart documentation and README
- [ ] Published to Helm repository

### ✅ Kustomize Overlays
- [ ] Base resources for common configuration
- [ ] Environment-specific overlays (dev/staging/prod)
- [ ] Patch strategies for customization
- [ ] ConfigMap and Secret generators
- [ ] Image transformation rules
- [ ] Resource quota definitions

### ✅ Template Best Practices
- [ ] No hardcoded values
- [ ] Proper label and annotation usage
- [ ] Resource limits and requests defined
- [ ] Health checks (liveness/readiness)
- [ ] Security context configuration
- [ ] Service account and RBAC

### ✅ Validation & Testing
- [ ] Helm lint passes
- [ ] kubeval schema validation
- [ ] Template dry-run successful
- [ ] Integration tests for deployments
- [ ] Chart upgrade testing
- [ ] Rollback verification

### ✅ Documentation
- [ ] values.yaml fully documented
- [ ] README with usage examples
- [ ] Architecture diagrams
- [ ] Troubleshooting guide
- [ ] Upgrade procedures

---

## 🏗️ Helm Chart Structure

```
deploy/
├── Chart.yaml                    # Chart metadata
├── values.yaml                   # Default values
├── values-dev.yaml              # Dev overrides
├── values-staging.yaml          # Staging overrides
├── values-prd.yaml              # Production overrides
├── README.md                    # Chart documentation
│
├── templates/
│   ├── _helpers.tpl             # Template helpers
│   ├── NOTES.txt               # Post-install notes
│   │
│   ├── namespace.yaml           # Namespace
│   ├── serviceaccount.yaml      # Service account
│   ├── rbac.yaml               # RBAC roles
│   │
│   ├── configmap.yaml          # Application config
│   ├── secret.yaml             # Secrets (sealed)
│   │
│   ├── deployment.yaml         # Builder deployment
│   ├── service.yaml            # Service
│   ├── servicemonitor.yaml     # Prometheus monitoring
│   │
│   ├── hpa.yaml                # Horizontal Pod Autoscaler
│   ├── pdb.yaml                # Pod Disruption Budget
│   ├── networkpolicy.yaml      # Network policies
│   └── resourcequota.yaml      # Resource quotas
│
├── charts/                      # Dependency charts
│   └── rabbitmq/               # RabbitMQ subchart
│
└── tests/
    ├── test-connection.yaml    # Helm test
    └── test-deployment.yaml    # Deployment test
```

---

## 🔧 Technical Implementation

### Chart.yaml

```yaml
apiVersion: v2
name: knative-lambda
description: Serverless function platform on Knative
type: application
version: 1.2.3              # Chart version
appVersion: "1.2.3"         # Application version

keywords:
  - knative
  - serverless
  - lambda
  - faas

maintainers:
  - name: Bruno Lucena
    email: bruno@homelab.io

home: https://github.com/brunolucena/homelab
sources:
  - https://github.com/brunolucena/homelab/tree/main/flux/clusters/homelab/infrastructure/knative-lambda

dependencies:
  - name: rabbitmq
    version: "12.5.0"
    repository: https://charts.bitnami.com/bitnami
    condition: rabbitmq.enabled
```

### values.yaml (Default Configuration)

```yaml
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🌍 GLOBAL CONFIGURATION
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
global:
  # Environment: dev, staging, prd
  environment: dev
  
  # Domain configuration
  domain: knative-lambda.homelab
  
  # AWS configuration
  aws:
    region: us-west-2
    accountId: "339954290315"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🚀 BUILDER SERVICE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
builder:
  # Replica count
  replicaCount: 1
  
  # Image configuration
  image:
    repository: 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-builder
    pullPolicy: IfNotPresent
    tag: ""  # Defaults to chart appVersion
  
  # Image pull secrets
  imagePullSecrets: []
  
  # Service account
  serviceAccount:
    create: true
    annotations:
      eks.amazonaws.com/role-arn: ""  # Set per environment
    name: knative-lambda-builder
  
  # Pod annotations (use PodMonitor/ServiceMonitor for Prometheus scraping)
  podAnnotations: {}
  
  # Security context
  podSecurityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 1000
  
  securityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    capabilities:
      drop:
      - ALL
  
  # Service configuration
  service:
    type: ClusterIP
    port: 8080
    targetPort: 8080
    annotations: {}
  
  # Resources
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"
  
  # Autoscaling
  autoscaling:
    enabled: false
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80
  
  # Pod Disruption Budget
  podDisruptionBudget:
    enabled: false
    minAvailable: 1
  
  # Health checks
  livenessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
  
  readinessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 10
    periodSeconds: 5
    timeoutSeconds: 3
    failureThreshold: 3
  
  # Environment variables
  env:
    - name: LOG_LEVEL
      value: "info"
    - name: PORT
      value: "8080"
    - name: ENVIRONMENT
      value: "dev"
  
  # Configuration from ConfigMap/Secret
  envFrom:
    - configMapRef:
        name: knative-lambda-config
    - secretRef:
        name: knative-lambda-secrets

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📝 APPLICATION CONFIG
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
config:
  # S3 configuration
  s3:
    bucket: knative-lambda-fusion-code
    region: us-west-2
  
  # ECR configuration
  ecr:
    registry: 339954290315.dkr.ecr.us-west-2.amazonaws.com
    repository: knative-lambdas-dev
  
  # Job configuration
  jobs:
    maxConcurrent: 10
    backoffLimit: 3
    ttlSecondsAfterFinished: 3600
  
  # Rate limiting
  rateLimiting:
    globalRPS: 100
    globalBurst: 200
    buildStartRPS: 50
    buildStartBurst: 100

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🐰 RABBITMQ (Subchart)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
rabbitmq:
  enabled: true
  auth:
    username: admin
    existingPasswordSecret: rabbitmq-password
  
  clustering:
    enabled: true
    replicaCount: 3
  
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📊 MONITORING
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
monitoring:
  enabled: true
  
  # ServiceMonitor (Prometheus Operator)
  serviceMonitor:
    enabled: true
    namespace: prometheus
    interval: 30s
    scrapeTimeout: 10s
  
  # Grafana dashboards
  grafana:
    dashboards:
      enabled: true

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔐 NETWORK POLICIES
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 💾 RESOURCE QUOTAS
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
resourceQuota:
  enabled: false
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "20"
```

### templates/deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "knative-lambda.fullname" . }}-builder
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "knative-lambda.labels" . | nindent 4 }}
    app.kubernetes.io/component: builder
spec:
  {{- if not .Values.builder.autoscaling.enabled }}
  replicas: {{ .Values.builder.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "knative-lambda.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: builder
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
        {{- with .Values.builder.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      labels:
        {{- include "knative-lambda.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: builder
    spec:
      {{- with .Values.builder.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "knative-lambda.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.builder.podSecurityContext | nindent 8 }}
      containers:
      - name: builder
        securityContext:
          {{- toYaml .Values.builder.securityContext | nindent 12 }}
        image: "{{ .Values.builder.image.repository }}:{{ .Values.builder.image.tag | default .Chart.AppVersion }}"
        imagePullPolicy: {{ .Values.builder.image.pullPolicy }}
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        livenessProbe:
          {{- toYaml .Values.builder.livenessProbe | nindent 12 }}
        readinessProbe:
          {{- toYaml .Values.builder.readinessProbe | nindent 12 }}
        resources:
          {{- toYaml .Values.builder.resources | nindent 12 }}
        env:
          {{- toYaml .Values.builder.env | nindent 12 }}
        envFrom:
          {{- toYaml .Values.builder.envFrom | nindent 12 }}
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: cache
          mountPath: /.cache
      volumes:
      - name: tmp
        emptyDir: {}
      - name: cache
        emptyDir: {}
      {{- with .Values.builder.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.builder.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.builder.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

### templates/_helpers.tpl

```yaml
{{/*
Expand the name of the chart.
*/}}
{{- define "knative-lambda.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "knative-lambda.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "knative-lambda.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "knative-lambda.labels" -}}
helm.sh/chart: {{ include "knative-lambda.chart" . }}
{{ include "knative-lambda.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
environment: {{ .Values.global.environment }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "knative-lambda.selectorLabels" -}}
app.kubernetes.io/name: {{ include "knative-lambda.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "knative-lambda.serviceAccountName" -}}
{{- if .Values.builder.serviceAccount.create }}
{{- default (include "knative-lambda.fullname" .) .Values.builder.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.builder.serviceAccount.name }}
{{- end }}
{{- end }}
```

---

## 🧪 Helm Chart Testing

### Lint and Validate

```bash
# Helm lint
helm lint deploy/

# Template validation
helm template knative-lambda deploy/ \
  --values deploy/values-dev.yaml \
  --debug

# Dry-run install
helm install knative-lambda deploy/ \
  --namespace knative-lambda \
  --create-namespace \
  --dry-run \
  --debug

# kubeval schema validation
helm template knative-lambda deploy/ | kubeval --strict
```

### Chart Testing

```bash
# Install ct (chart-testing)
brew install chart-testing

# Lint charts
ct lint --chart-dirs deploy/ --all

# Install and test
ct install --chart-dirs deploy/
```

### Helm Unit Tests

```bash
# Install helm-unittest
helm plugin install https://github.com/helm-unittest/helm-unittest

# Run tests
helm unittest deploy/
```

**File**: `deploy/tests/deployment_test.yaml`
```yaml
suite: test deployment
templates:
  - deployment.yaml
tests:
  - it: should create deployment
    asserts:
      - isKind:
          of: Deployment
      - equal:
          path: metadata.name
          value: RELEASE-NAME-knative-lambda-builder
  
  - it: should set replicas
    set:
      builder.replicaCount: 3
    asserts:
      - equal:
          path: spec.replicas
          value: 3
  
  - it: should have correct image
    set:
      builder.image.tag: "v1.2.3"
    asserts:
      - equal:
          path: spec.template.spec.containers[0].image
          value: "339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-builder:v1.2.3"
```

---

## 🚀 Deployment Commands

### Install Chart

```bash
# Install to dev
helm install knative-lambda deploy/ \
  --namespace knative-lambda \
  --create-namespace \
  --values deploy/values-dev.yaml

# Install to production
helm install knative-lambda deploy/ \
  --namespace knative-lambda \
  --create-namespace \
  --values deploy/values-prd.yaml
```

### Upgrade Chart

```bash
# Upgrade dev
helm upgrade knative-lambda deploy/ \
  --namespace knative-lambda \
  --values deploy/values-dev.yaml \
  --reuse-values

# Upgrade with version bump
helm upgrade knative-lambda deploy/ \
  --namespace knative-lambda \
  --values deploy/values-prd.yaml \
  --set builder.image.tag=v1.2.4
```

### Rollback

```bash
# List revisions
helm history knative-lambda -n knative-lambda

# Rollback to previous
helm rollback knative-lambda -n knative-lambda

# Rollback to specific revision
helm rollback knative-lambda 3 -n knative-lambda
```

### Uninstall

```bash
# Uninstall chart
helm uninstall knative-lambda -n knative-lambda
```

---

## 📦 Chart Versioning

### Version Bumping

```bash
# Bump chart version
yq eval '.version = "1.2.4"' -i deploy/Chart.yaml

# Bump app version
yq eval '.appVersion = "1.2.4"' -i deploy/Chart.yaml

# Update dependencies
helm dependency update deploy/
```

### Packaging

```bash
# Package chart
helm package deploy/

# This creates: knative-lambda-1.2.3.tgz
```

### Publishing to Helm Repository

```bash
# Upload to S3 bucket
aws s3 cp knative-lambda-1.2.3.tgz s3://helm-charts.homelab/

# Update index
helm repo index . --url https://helm-charts.homelab
aws s3 cp index.yaml s3://helm-charts.homelab/

# Add repo (for users)
helm repo add homelab https://helm-charts.homelab
helm repo update
helm search repo knative-lambda
```

---

## 📊 Monitoring Infrastructure Changes

### Detect Drift

```bash
# Compare desired vs actual state
helm diff upgrade knative-lambda deploy/ \
  --namespace knative-lambda \
  --values deploy/values-prd.yaml

# Using kubectl diff
kubectl diff -f <(helm template knative-lambda deploy/ --values deploy/values-prd.yaml)
```

### Track Changes

```promql
# Helm releases
count(kube_configmap_labels{label_name="MANAGED_BY",label_value="Helm"})

# Chart version tracking
kube_configmap_labels{
  label_name="app_kubernetes_io_version",
  namespace=~"knative-lambda-.*"
}
```

---

## 💡 Pro Tips

### 1. Use Helm Secrets

```bash
# Encrypt secrets
helm secrets encrypt deploy/secrets.yaml

# Install with secrets
helm secrets install knative-lambda deploy/ \
  --values deploy/values-prd.yaml \
  --values deploy/secrets.prd.yaml
```

### 2. Post-Renderer (for Kustomize integration)

```bash
# Combine Helm + Kustomize
helm install knative-lambda deploy/ \
  --post-renderer ./kustomize
```

### 3. Helm Hooks

```yaml
# Pre-install hook
apiVersion: batch/v1
kind: Job
metadata:
  annotations:
    "helm.sh/hook": pre-install
    "helm.sh/hook-weight": "-5"
    "helm.sh/hook-delete-policy": hook-succeeded
```

---

## 📈 Performance Requirements

- **Chart Install Time**: < 2 minutes
- **Chart Upgrade Time**: < 3 minutes
- **Rollback Time**: < 1 minute
- **Template Rendering**: < 5 seconds
- **Chart Size**: < 50KB (compressed)

---

## 📚 Related Documentation

- [DEVOPS-002: GitOps Deployment](DEVOPS-002-gitops-deployment.md)
- [DEVOPS-003: Multi-Environment Management](DEVOPS-003-multi-environment.md)
- [DEVOPS-004: CI/CD Pipeline](DEVOPS-004-cicd-pipeline.md)
- Helm Best Practices: https://helm.sh/docs/chart_best_practices/
- Kustomize Documentation: https://kustomize.io/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: Production Ready

