# 🔄 DEVOPS-007: Cost Optimization

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 5
**Linear URL**: https://linear.app/bvlucena/issue/BVL-239/devops-007-cost-optimization

---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** optimize infrastructure costs through rightsizing, spot instances, and efficient resource usage  
**So that** we reduce cloud spending while maintaining performance and reliability

---

## 🎯 Acceptance Criteria

### ✅ Compute Cost Optimization
- [ ] Use Spot Instances for build jobs (60% savings)
- [ ] Implement cluster autoscaling
- [ ] Rightsize pod resource requests/limits
- [ ] Enable scale-to-zero for Knative functions
- [ ] Use Fargate Spot for batch workloads

### ✅ Storage Cost Optimization
- [ ] ECR image lifecycle policies
- [ ] S3 lifecycle policies (transition to Glacier)
- [ ] EBS volume optimization
- [ ] Snapshot cleanup automation
- [ ] Unused volume detection

### ✅ Network Cost Optimization
- [ ] VPC endpoints for AWS services
- [ ] Minimize cross-AZ traffic
- [ ] CloudFront for static assets
- [ ] Optimize data transfer costs
- [ ] NAT Gateway optimization

### ✅ Resource Monitoring
- [ ] Cost allocation tags
- [ ] Per-environment cost tracking
- [ ] Resource utilization dashboards
- [ ] Cost anomaly detection
- [ ] Budget alerts and limits

### ✅ Optimization Automation
- [ ] Automated rightsizing recommendations
- [ ] Unused resource cleanup
- [ ] Reserved Instance management
- [ ] Savings Plans recommendations
- [ ] Cost optimization reports

---

## 💰 Cost Breakdown & Savings Opportunities

```
┌─────────────────────────────────────────────────────────────────┐
│                    MONTHLY COST ANALYSIS                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  COMPUTE (EC2/EKS)                          $450/month          │
│  ├─ Build nodes (3x c5.xlarge)              $350               │
│  │  💡 OPTIMIZATION: Use Spot → SAVE $210 (60%)                │
│  │                                                              │
│  ├─ Control plane (managed EKS)              $73                │
│  │  ✅ Already optimized                                       │
│  │                                                              │
│  └─ Function pods (Knative)                  $27                │
│     ✅ Scale-to-zero enabled                                   │
│                                                                 │
│  STORAGE                                     $85/month          │
│  ├─ ECR (Docker images)                      $50                │
│  │  💡 OPTIMIZATION: Lifecycle policy → SAVE $25 (50%)         │
│  │                                                              │
│  ├─ S3 (parser code)                         $20                │
│  │  💡 OPTIMIZATION: Intelligent-Tiering → SAVE $8 (40%)       │
│  │                                                              │
│  └─ EBS (persistent volumes)                 $15                │
│     ✅ Already optimized (gp3)                                 │
│                                                                 │
│  NETWORK                                     $75/month          │
│  ├─ Data transfer                            $35                │
│  │  💡 OPTIMIZATION: VPC Endpoints → SAVE $15 (43%)            │
│  │                                                              │
│  ├─ Load balancers                           $25                │
│  │  ✅ Shared ALB across services                              │
│  │                                                              │
│  └─ NAT Gateway                              $15                │
│     💡 OPTIMIZATION: Single NAT → SAVE $30 (67%)               │
│                                                                 │
│  MONITORING                                  $50/month          │
│  ├─ CloudWatch Logs                          $30                │
│  │  💡 OPTIMIZATION: Retention policy → SAVE $12 (40%)         │
│  │                                                              │
│  └─ Prometheus storage                       $20                │
│     ✅ Local storage (no extra cost)                           │
│                                                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│  CURRENT TOTAL:                              $660/month         │
│  OPTIMIZED TOTAL:                            $360/month         │
│  💰 TOTAL SAVINGS:                           $300/month (45%)   │
│  💰 ANNUAL SAVINGS:                          $3,600/year        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Technical Implementation

### 1. Spot Instances for Build Jobs

**NodeGroup Configuration** (eksctl)
```yaml
# eksctl-config.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: homelab
  region: us-west-2

managedNodeGroups:
- name: build-spot
  instanceTypes:
    - c5.large
    - c5.xlarge
    - c5a.large
  spot: true  # Enable Spot Instances
  minSize: 0
  maxSize: 10
  desiredCapacity: 2
  
  labels:
    workload-type: builds
    spot: "true"
  
  taints:
  - key: spot
    value: "true"
    effect: NoSchedule
  
  tags:
    k8s.io/cluster-autoscaler/enabled: "true"
    k8s.io/cluster-autoscaler/homelab: "owned"
```

**Job Configuration for Spot**
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: kaniko-build
spec:
  template:
    spec:
      # Tolerate spot instance taint
      tolerations:
      - key: spot
        operator: Equal
        value: "true"
        effect: NoSchedule
      
      # Node affinity for spot instances
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: spot
                operator: In
                values:
                - "true"
      
      # Handle spot interruption
      terminationGracePeriodSeconds: 120
      
      containers:
      - name: kaniko
        image: gcr.io/kaniko-project/executor:latest
```

**Spot Interruption Handler**
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: spot-interrupt-handler
spec:
  selector:
    matchLabels:
      app: spot-interrupt-handler
  template:
    spec:
      hostNetwork: true
      containers:
      - name: handler
        image: amazon/aws-node-termination-handler:latest
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
```

### 2. Cluster Autoscaler

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cluster-autoscaler
  template:
    spec:
      serviceAccountName: cluster-autoscaler
      containers:
      - name: cluster-autoscaler
        image: k8s.gcr.io/autoscaling/cluster-autoscaler:v1.27.0
        command:
        - ./cluster-autoscaler
        - --v=4
        - --cloud-provider=aws
        - --skip-nodes-with-local-storage=false
        - --expander=least-waste
        - --node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/homelab
        - --balance-similar-node-groups
        - --skip-nodes-with-system-pods=false
        - --scale-down-enabled=true
        - --scale-down-delay-after-add=10m
        - --scale-down-unneeded-time=10m
```

### 3. ECR Lifecycle Policies

```json
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Keep only last 10 dev images",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": ["dev-"],
        "countType": "imageCountMoreThan",
        "countNumber": 10
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 2,
      "description": "Keep staging images for 30 days",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": ["staging-"],
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 30
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 3,
      "description": "Keep prod images for 90 days",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": ["prd-"],
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 90
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 4,
      "description": "Remove untagged images after 1 day",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 1
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
```

**Apply Policy**
```bash
aws ecr put-lifecycle-policy \
  --repository-name knative-lambdas-dev \
  --lifecycle-policy-text file://ecr-lifecycle-policy.json
```

### 4. S3 Lifecycle Policies

```json
{
  "Rules": [
    {
      "Id": "TransitionToIA",
      "Status": "Enabled",
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "STANDARD_IA"
        },
        {
          "Days": 90,
          "StorageClass": "GLACIER"
        }
      ],
      "NoncurrentVersionTransitions": [
        {
          "NoncurrentDays": 30,
          "StorageClass": "GLACIER"
        }
      ]
    },
    {
      "Id": "DeleteOldVersions",
      "Status": "Enabled",
      "NoncurrentVersionExpiration": {
        "NoncurrentDays": 90
      }
    },
    {
      "Id": "CleanupIncompleteUploads",
      "Status": "Enabled",
      "AbortIncompleteMultipartUpload": {
        "DaysAfterInitiation": 7
      }
    }
  ]
}
```

**Apply Policy**
```bash
aws s3api put-bucket-lifecycle-configuration \
  --bucket knative-lambda-fusion-code \
  --lifecycle-configuration file://s3-lifecycle-policy.json
```

### 5. VPC Endpoints (Reduce Data Transfer Costs)

```yaml
# Terraform configuration
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.us-west-2.s3"
  
  route_table_ids = [
    aws_route_table.private.id
  ]
  
  tags = {
    Name = "s3-vpc-endpoint"
  }
}

resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.us-west-2.ecr.api"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]
  
  private_dns_enabled = true
  
  tags = {
    Name = "ecr-api-vpc-endpoint"
  }
}

resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.us-west-2.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]
  
  private_dns_enabled = true
  
  tags = {
    Name = "ecr-dkr-vpc-endpoint"
  }
}
```

### 6. Resource Rightsizing

**Vertical Pod Autoscaler** (recommendations)
```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: knative-lambda-builder-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-builder
  updatePolicy:
    updateMode: "Off"  # Recommendation only
  resourcePolicy:
    containerPolicies:
    - containerName: builder
      minAllowed:
        cpu: 100m
        memory: 128Mi
      maxAllowed:
        cpu: 2
        memory: 4Gi
```

**Get Recommendations**
```bash
kubectl describe vpa knative-lambda-builder-vpa
```

---

## 📊 Cost Tracking & Monitoring

### Cost Allocation Tags

```yaml
# Add to all resources
tags:
  Environment: prd
  Project: knative-lambda
  Team: platform
  CostCenter: engineering
  ManagedBy: terraform
```

### Kubecost Installation

```bash
# Install Kubecost
helm repo add kubecost https://kubecost.github.io/cost-analyzer/
helm install kubecost kubecost/cost-analyzer \
  --namespace kubecost --create-namespace \
  --set kubecostToken="${KUBECOST_TOKEN}"

# Access dashboard
kubectl port-forward -n kubecost svc/kubecost-cost-analyzer 9090:9090
```

### Cost Dashboard (Grafana)

```json
{
  "dashboard": {
    "title": "Knative Lambda - Cost Analysis",
    "panels": [
      {
        "title": "Monthly Cost by Environment",
        "targets": [{
          "expr": "sum by (environment) (node_cpu_hourly_cost * on (node) group_left (label_environment) kube_node_labels)"
        }]
      },
      {
        "title": "Cost per Build",
        "targets": [{
          "expr": "sum(increase(build_cost_total[1h])) / sum(increase(builds_total[1h]))"
        }]
      },
      {
        "title": "Spot Instance Savings",
        "targets": [{
          "expr": "(sum(node_spot_savings_total) / sum(node_total_cost)) * 100"
        }]
      }
    ]
  }
}
```

---

## 🧪 Cost Optimization Testing

### Test 1: Spot Instance Savings

```bash
# Deploy build job to spot
kubectl apply -f build-job-spot.yaml

# Monitor cost
kubectl get nodes -l spot=true -o json | \
  jq '.items[] | {name: .metadata.name, instanceType: .metadata.labels["node.kubernetes.io/instance-type"]}'

# Calculate savings
SPOT_COST=$(kubectl get nodes -l spot=true --no-headers | wc -l | awk '{print $1 * 0.034}')
ON_DEMAND_COST=$(kubectl get nodes -l spot=true --no-headers | wc -l | awk '{print $1 * 0.085}')
SAVINGS=$(echo "scale=2; ($ON_DEMAND_COST - $SPOT_COST) / $ON_DEMAND_COST * 100" | bc)
echo "Savings: ${SAVINGS}%"
```

### Test 2: ECR Lifecycle Policy

```bash
# List images before cleanup
aws ecr describe-images \
  --repository-name knative-lambdas-dev \
  --query 'sort_by(imageDetails,& imagePushedAt)[*]' \
  --output table

# Wait for lifecycle policy execution (daily)
# Check again after 24 hours
```

### Test 3: Cluster Autoscaling

```bash
# Scale down test
kubectl scale deployment knative-lambda-builder --replicas=0

# Wait 10 minutes
sleep 600

# Check node count (should scale down)
kubectl get nodes
```

---

## 📈 Cost Optimization Metrics

### Key Metrics

```promql
# Cost per environment
sum by (environment) (node_cpu_hourly_cost * on (node) group_left kube_node_labels)

# Spot instance coverage
count(kube_node_labels{label_spot="true"}) / count(kube_node_labels) * 100

# Unused resources
sum(kube_pod_container_resource_requests_cpu_cores) - 
sum(rate(container_cpu_usage_seconds_total[5m]))

# Storage costs
sum(kube_persistentvolumeclaim_resource_requests_storage_bytes) * 0.10 / 1024 / 1024 / 1024

# Idle resources (waste)
(sum(kube_pod_container_resource_requests_cpu_cores) - 
 sum(rate(container_cpu_usage_seconds_total[5m]))) / 
sum(kube_pod_container_resource_requests_cpu_cores) * 100
```

### Savings Tracking

```promql
# Total monthly savings
sum(spot_instance_savings_total) +
sum(ecr_lifecycle_savings_total) +
sum(s3_lifecycle_savings_total) +
sum(vpc_endpoint_savings_total)
```

---

## 💡 Cost Optimization Checklist

### Quick Wins (Implement First)
- [x] Enable Spot Instances for build jobs (60% savings)
- [x] ECR lifecycle policies (50% savings)
- [x] S3 lifecycle policies (40% savings)
- [x] Cluster autoscaler (20% savings)
- [x] VPC endpoints (15% savings)

### Medium-Term Optimizations
- [ ] Reserved Instances for stable workloads
- [ ] Savings Plans for compute
- [ ] CloudFront for static assets
- [ ] Optimize EBS volume types (gp3)
- [ ] Right-size RabbitMQ instances

### Long-Term Strategies
- [ ] Multi-region cost optimization
- [ ] Serverless migration where applicable
- [ ] Custom instance types (Graviton)
- [ ] Advanced autoscaling policies
- [ ] FinOps culture and practices

---

## 🚨 Cost Anomaly Alerts

```yaml
groups:
- name: cost-alerts
  rules:
  - alert: CostAnomalyDetected
    expr: | (sum(rate(node_cpu_hourly_cost[1h])) - 
       sum(rate(node_cpu_hourly_cost[1h] offset 24h))) / 
       sum(rate(node_cpu_hourly_cost[1h] offset 24h)) > 0.25
    for: 1h
    labels:
      severity: warning
    annotations:
      summary: "Cost increased by > 25%"
  
  - alert: UnusedResourcesHigh
    expr: | (sum(kube_pod_container_resource_requests_cpu_cores) - 
       sum(rate(container_cpu_usage_seconds_total[5m]))) / 
      sum(kube_pod_container_resource_requests_cpu_cores) > 0.50
    for: 6h
    labels:
      severity: warning
    annotations:
      summary: "Over 50% unused CPU resources"
  
  - alert: MonthlyBudgetExceeded
    expr: | sum(monthly_cost_total) > 700
    labels:
      severity: critical
    annotations:
      summary: "Monthly budget of $700 exceeded"
```

---

## 📚 Cost Optimization Resources

### AWS Cost Tools
- AWS Cost Explorer
- AWS Budgets
- AWS Cost Anomaly Detection
- AWS Compute Optimizer
- AWS Trusted Advisor

### Third-Party Tools
- Kubecost (Kubernetes cost visibility)
- Infracost (IaC cost estimation)
- CloudHealth (multi-cloud FinOps)
- Spot.io (automated spot instance management)

---

## 📈 Performance Requirements

- **Cost Tracking Latency**: < 1 hour
- **Rightsizing Recommendations**: Daily
- **Spot Instance Interruption Handling**: < 2 minutes
- **Autoscaler Response Time**: < 5 minutes
- **Cost Report Generation**: < 30 seconds

---

## 📚 Related Documentation

- [DEVOPS-003: Multi-Environment Management](DEVOPS-003-multi-environment.md)
- [DEVOPS-005: Infrastructure as Code](DEVOPS-005-infrastructure-as-code.md)
- AWS Cost Optimization: https://aws.amazon.com/aws-cost-management/
- Kubecost Documentation: https://docs.kubecost.com/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: Production Ready

