# Notifi Homelab Implementation Plan

> **100% Self-Hosted Kubernetes Platform - Zero AWS Dependencies**
>
> **Target Environment**: K3s on Raspberry Pi + x86_64 nodes  
> **Estimated Timeline**: 4-6 weeks (phased approach)  
> **Cost**: ~$3,000 initial setup vs $42,000/year AWS  
> **Author**: Senior Cloud Architect  
> **Date**: November 2025

---

## 📋 Table of Contents

1. [Executive Summary](#executive-summary)
2. [Prerequisites](#prerequisites)
3. [Hardware Requirements](#hardware-requirements)
4. [Network Architecture](#network-architecture)
5. [Implementation Phases](#implementation-phases)
6. [Phase 1: Infrastructure Setup](#phase-1-infrastructure-setup)
7. [Phase 2: Platform Services](#phase-2-platform-services)
8. [Phase 3: Data Layer](#phase-3-data-layer)
9. [Phase 4: Messaging Infrastructure](#phase-4-messaging-infrastructure)
10. [Phase 5: Notifi Services](#phase-5-notifi-services)
11. [Phase 6: Observability & Monitoring](#phase-6-observability--monitoring)
12. [Phase 7: Production Readiness](#phase-7-production-readiness)
13. [Operations & Maintenance](#operations--maintenance)
14. [Troubleshooting Guide](#troubleshooting-guide)
15. [Migration from Cloud](#migration-from-cloud)

---

## 🎯 Executive Summary

### Objective
Deploy the complete Notifi notification infrastructure (37 microservices) on a self-hosted Kubernetes homelab cluster using 100% open-source tools with zero AWS dependencies.

### Key Benefits
- **93% cost savings**: $3,000/year vs $42,000/year AWS
- **Complete control**: Own your infrastructure and data
- **Privacy**: No data leaves your premises
- **Learning**: Hands-on cloud-native experience
- **Scalability**: Handles 1-3K RPS (sufficient for development/small production)

### Architecture Overview
```
┌─────────────────────────────────────────────────────────────────┐
│                    Edge Layer (Nginx/Traefik)                    │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│              K3s Cluster (5 nodes, ARM64 + x86_64)               │
│  ┌────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │ Platform   │  │ Data Layer   │  │ Notifi Services (37)   │  │
│  │ Services   │  │ (Self-Hosted)│  │ Microservices          │  │
│  │ - Flux CD  │  │ - PostgreSQL │  │ - API Gateways         │  │
│  │ - Linkerd  │  │ - Redis      │  │ - Blockchain Proxies   │  │
│  │ - Grafana  │  │ - MongoDB    │  │ - Event Processor      │  │
│  │ - Knative  │  │ - MinIO      │  │ - Messaging Workers    │  │
│  └────────────┘  └──────────────┘  └────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## ✅ Prerequisites

### Knowledge Requirements
- ✅ Kubernetes basics (pods, services, deployments)
- ✅ Docker & containerization
- ✅ Linux system administration
- ✅ Networking (DNS, routing, firewalls)
- ✅ Git & GitOps concepts
- ⚠️  Recommended: Service mesh, observability, HA databases

### Tools Required
```bash
# Install required CLI tools
brew install kubectl helm flux linkerd kubeseal
brew install k3sup  # K3s installer
brew install velero restic  # Backup tools
```

### Accounts & Access
- [ ] GitHub account (for GitOps)
- [ ] Domain name (for external access)
- [ ] Cloudflare account (optional, for DNS/CDN)
- [ ] Docker Hub account (or Harbor setup)
- [ ] SMTP relay service (for email notifications)
- [ ] Twilio account (for SMS)
- [ ] Telegram Bot API key

---

## 🖥️ Hardware Requirements

### Minimum Configuration (5 Nodes)

#### **Node 1-3: Raspberry Pi 4/5 (ARM64)**
```yaml
Model: Raspberry Pi 4B or 5
CPU: 4-core ARM Cortex-A72/A76
RAM: 8GB
Storage: 256GB NVMe SSD (via USB 3.0/PCIe)
Network: Gigabit Ethernet
Power: 15W per node
Role: Control plane + worker nodes
```

#### **Node 4: Studio (x86_64)**
```yaml
CPU: 8-core x86_64 (Intel/AMD)
RAM: 32GB DDR4
Storage: 512GB NVMe SSD
Network: Gigabit Ethernet
Power: 65W
Role: Heavy workloads (messaging, webhooks)
```

#### **Node 5: Forge (GPU Node)**
```yaml
CPU: 12-core x86_64
RAM: 64GB DDR4
GPU: RTX 4090 (optional, for analytics)
Storage: 1TB NVMe SSD
Network: Gigabit Ethernet
Power: 450W (with GPU)
Role: GPU workloads, analytics, ML
```

### Storage Requirements
```yaml
Node Storage: 256-512GB NVMe per node
NAS/MinIO: 2-4TB HDD/SSD (distributed across nodes)
Backup Storage: 1TB external USB/NAS

Total: ~5TB usable capacity
```

### Network Requirements
```yaml
Internal: 1Gbps LAN switch (managed, VLAN support)
Internet: 100Mbps-1Gbps fiber (symmetrical preferred)
Static IP: Recommended (or DDNS)
Router: Port forwarding capability
Firewall: Hardware firewall (pfSense/OPNsense recommended)
```

### Power & Cooling
```yaml
Total Power Draw: ~600W peak (with GPU)
UPS: 1000VA minimum (for graceful shutdown)
Cooling: Passive cooling for Pis, active for x86_64/GPU
Rack: Optional 12U rack or shelf setup
```

### Cost Breakdown
| Component | Quantity | Unit Price | Total |
|-----------|----------|------------|-------|
| Raspberry Pi 5 (8GB) | 3 | $80 | $240 |
| NVMe SSD (256GB) | 3 | $40 | $120 |
| Studio PC (refurbished) | 1 | $400 | $400 |
| Forge GPU Node | 1 | $1,500 | $1,500 |
| Network Switch (8-port) | 1 | $100 | $100 |
| NAS/Storage (2TB) | 1 | $200 | $200 |
| UPS (1000VA) | 1 | $150 | $150 |
| Cables & Accessories | - | - | $100 |
| **Total** | | | **$2,810** |

---

## 🌐 Network Architecture

### Network Topology
```
Internet
    │
    ▼
[Router/Firewall] (192.168.1.1)
    │
    ├─── [Edge Nginx] (192.168.1.10) - Reverse Proxy
    │
    ▼
[K3s Cluster Network]
    │
    ├─── Pi Node 1 (192.168.1.101) - Control Plane
    ├─── Pi Node 2 (192.168.1.102) - Worker
    ├─── Pi Node 3 (192.168.1.103) - Worker
    ├─── Studio (192.168.1.104) - Worker
    └─── Forge (192.168.1.105) - GPU Worker
    │
    ▼
[Cluster Services]
    ├─── Pod Network: 10.42.0.0/16
    └─── Service Network: 10.43.0.0/16
```

### DNS Configuration
```yaml
# External DNS (Cloudflare)
notifi.yourdomain.com → Public IP → Router → Nginx Edge

# Internal DNS (CoreDNS)
*.svc.cluster.local → K8s Services
management-gateway.notifi.svc.cluster.local → Management Gateway
dataplane-gateway.notifi.svc.cluster.local → Dataplane Gateway
```

### Port Mapping
```yaml
# External Access (through router)
80/TCP   → Nginx Edge → Traefik Ingress
443/TCP  → Nginx Edge → Traefik Ingress (TLS)
6443/TCP → K3s API Server

# Internal Services
5432/TCP  → PostgreSQL
6379/TCP  → Redis
5672/TCP  → RabbitMQ
9092/TCP  → Kafka
9000/TCP  → MinIO
80/TCP    → Harbor Registry
```

### Firewall Rules
```bash
# Allow from Internet to Edge Nginx
iptables -A FORWARD -p tcp --dport 80 -j ACCEPT
iptables -A FORWARD -p tcp --dport 443 -j ACCEPT

# Allow K3s API (restrict to known IPs)
iptables -A INPUT -p tcp --dport 6443 -s <your-ip> -j ACCEPT

# Allow cluster communication
iptables -A INPUT -s 192.168.1.0/24 -j ACCEPT

# Block everything else
iptables -A INPUT -j DROP
```

---

## 📅 Implementation Phases

### Timeline Overview
```
Week 1-2: Infrastructure Setup & K3s Deployment
Week 3: Platform Services (Flux, Linkerd, Observability)
Week 4: Data Layer (PostgreSQL, Redis, MongoDB, MinIO)
Week 5: Messaging (RabbitMQ, Kafka) & Notifi Core Services
Week 6: Notifi Application Services & Testing
Week 7: Production Hardening & Documentation
Week 8: Monitoring, Backup, & Operations Training
```

### Phase Dependencies
```mermaid
graph LR
    A[Phase 1: Infrastructure] --> B[Phase 2: Platform]
    B --> C[Phase 3: Data Layer]
    B --> D[Phase 4: Messaging]
    C --> E[Phase 5: Notifi Services]
    D --> E
    E --> F[Phase 6: Observability]
    F --> G[Phase 7: Production Ready]
```

---

## 🏗️ Phase 1: Infrastructure Setup

**Duration**: 3-5 days  
**Goal**: Bootstrap K3s cluster with basic networking

### Step 1.1: Prepare Nodes

```bash
# On each node (Pi 1-3, Studio, Forge)
# 1. Install Ubuntu Server 22.04 LTS
# 2. Update system
sudo apt update && sudo apt upgrade -y

# 3. Set static IP
cat << EOF | sudo tee /etc/netplan/01-static.yaml
network:
  version: 2
  ethernets:
    eth0:
      dhcp4: no
      addresses: [192.168.1.101/24]  # Change per node
      gateway4: 192.168.1.1
      nameservers:
        addresses: [8.8.8.8, 1.1.1.1]
EOF

sudo netplan apply

# 4. Set hostname
sudo hostnamectl set-hostname pi-node-1  # Change per node

# 5. Disable swap (required for K8s)
sudo swapoff -a
sudo sed -i '/swap/d' /etc/fstab

# 6. Enable cgroup memory (Raspberry Pi only)
sudo sed -i '$ s/$/ cgroup_enable=cpuset cgroup_enable=memory cgroup_memory=1/' /boot/firmware/cmdline.txt
sudo reboot
```

### Step 1.2: Install K3s Cluster

```bash
# On Pi Node 1 (Control Plane)
curl -sfL https://get.k3s.io | sh -s - server \
  --cluster-init \
  --disable traefik \
  --disable servicelb \
  --write-kubeconfig-mode 644 \
  --node-taint node-role.kubernetes.io/master:NoSchedule \
  --flannel-backend=host-gw \
  --kube-apiserver-arg='service-node-port-range=80-32767'

# Get node token
sudo cat /var/lib/rancher/k3s/server/node-token

# On Pi Node 2-3 (Workers - ARM64)
curl -sfL https://get.k3s.io | K3S_URL=https://192.168.1.101:6443 \
  K3S_TOKEN=<TOKEN_FROM_NODE1> sh -s - agent \
  --node-label topology.kubernetes.io/zone=arm64 \
  --node-label node.kubernetes.io/instance-type=pi

# On Studio Node (Worker - x86_64)
curl -sfL https://get.k3s.io | K3S_URL=https://192.168.1.101:6443 \
  K3S_TOKEN=<TOKEN_FROM_NODE1> sh -s - agent \
  --node-label topology.kubernetes.io/zone=x86 \
  --node-label node.kubernetes.io/instance-type=studio

# On Forge Node (GPU Worker)
curl -sfL https://get.k3s.io | K3S_URL=https://192.168.1.101:6443 \
  K3S_TOKEN=<TOKEN_FROM_NODE1> sh -s - agent \
  --node-label topology.kubernetes.io/zone=x86 \
  --node-label node.kubernetes.io/instance-type=forge \
  --node-label nvidia.com/gpu=true
```

### Step 1.3: Configure kubectl Access

```bash
# Copy kubeconfig from Pi Node 1
mkdir -p ~/.kube
scp ubuntu@192.168.1.101:/etc/rancher/k3s/k3s.yaml ~/.kube/config

# Update server address
sed -i 's/127.0.0.1/192.168.1.101/g' ~/.kube/config

# Test connectivity
kubectl get nodes
# Should show 5 nodes in Ready state
```

### Step 1.4: Install Local Path Provisioner

```bash
# K3s includes local-path provisioner by default
kubectl get storageclass
# Should show 'local-path' as default

# Test with a PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF

kubectl get pvc test-pvc
kubectl delete pvc test-pvc
```

### Step 1.5: Label Nodes for Workload Placement

```bash
# Label nodes for specific workloads
kubectl label nodes pi-node-1 workload-type=control-plane
kubectl label nodes pi-node-2 workload-type=data-services
kubectl label nodes pi-node-3 workload-type=data-services
kubectl label nodes studio workload-type=messaging
kubectl label nodes forge workload-type=analytics

# Verify labels
kubectl get nodes --show-labels
```

### Validation Checklist Phase 1
- [ ] All 5 nodes showing as Ready
- [ ] Node roles correctly assigned
- [ ] kubectl access working
- [ ] Local path provisioner installed
- [ ] Network connectivity between nodes
- [ ] DNS resolution working

---

## 🔧 Phase 2: Platform Services

**Duration**: 3-4 days  
**Goal**: Deploy GitOps, Service Mesh, and core platform tools

### Step 2.1: Bootstrap Flux CD

```bash
# Install Flux CLI
brew install flux

# Export GitHub token
export GITHUB_TOKEN=<your-github-pat>
export GITHUB_USER=<your-github-username>

# Bootstrap Flux
flux bootstrap github \
  --owner=$GITHUB_USER \
  --repository=homelab \
  --path=flux/clusters/studio \
  --personal \
  --components-extra=image-reflector-controller,image-automation-controller

# Verify Flux installation
kubectl get pods -n flux-system
flux check
```

### Step 2.2: Install Linkerd Service Mesh

```bash
# Install Linkerd CLI
brew install linkerd

# Generate certificates (for mTLS)
mkdir -p ~/linkerd-certs
cd ~/linkerd-certs

# Create root CA
step certificate create root.linkerd.cluster.local ca.crt ca.key \
  --profile root-ca --no-password --insecure

# Create issuer certificate
step certificate create identity.linkerd.cluster.local issuer.crt issuer.key \
  --profile intermediate-ca --not-after 8760h --no-password --insecure \
  --ca ca.crt --ca-key ca.key

# Install Linkerd CRDs
linkerd install --crds | kubectl apply -f -

# Install Linkerd control plane
linkerd install \
  --identity-trust-anchors-file ca.crt \
  --identity-issuer-certificate-file issuer.crt \
  --identity-issuer-key-file issuer.key \
  | kubectl apply -f -

# Install Linkerd Viz (observability)
linkerd viz install | kubectl apply -f -

# Verify installation
linkerd check
```

### Step 2.3: Install Sealed Secrets

```bash
# Install Sealed Secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Install kubeseal CLI
brew install kubeseal

# Test sealed secrets
kubectl create secret generic test-secret \
  --from-literal=password=mysecret \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > sealed-secret.yaml

kubectl apply -f sealed-secret.yaml
kubectl get secret test-secret
kubectl delete secret test-secret
rm sealed-secret.yaml
```

### Step 2.4: Install Cert Manager

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=available --timeout=300s deployment/cert-manager -n cert-manager

# Create Let's Encrypt ClusterIssuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
EOF
```

### Step 2.5: Install Traefik Ingress

```bash
# Create Traefik namespace
kubectl create namespace traefik

# Install Traefik via Helm
helm repo add traefik https://traefik.github.io/charts
helm repo update

helm install traefik traefik/traefik \
  --namespace traefik \
  --set ports.web.nodePort=30080 \
  --set ports.websecure.nodePort=30443 \
  --set service.type=NodePort \
  --set ingressRoute.dashboard.enabled=true

# Verify Traefik
kubectl get pods -n traefik
kubectl get svc -n traefik
```

### Step 2.6: Install Metrics Server

```bash
# Install metrics-server for HPA
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Patch for k3s (self-signed certs)
kubectl patch deployment metrics-server -n kube-system --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

# Test metrics
kubectl top nodes
kubectl top pods -A
```

### Validation Checklist Phase 2
- [ ] Flux CD reconciling from Git
- [ ] Linkerd service mesh running
- [ ] Sealed Secrets controller active
- [ ] Cert Manager issuing certificates
- [ ] Traefik ingress responding
- [ ] Metrics Server providing node/pod metrics

---

## 💾 Phase 3: Data Layer

**Duration**: 4-5 days  
**Goal**: Deploy HA databases (PostgreSQL, Redis, MongoDB, MinIO)

### Step 3.1: Deploy PostgreSQL with Patroni

Create namespace and configuration:

```yaml
# postgresql-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: data-services
  annotations:
    linkerd.io/inject: enabled  # Enable service mesh
```

```yaml
# postgresql-patroni.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgresql-patroni
  namespace: data-services
spec:
  serviceName: postgresql
  replicas: 3
  selector:
    matchLabels:
      app: postgresql
  template:
    metadata:
      labels:
        app: postgresql
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: postgresql
            topologyKey: kubernetes.io/hostname
      containers:
      - name: postgresql
        image: postgres:16-alpine
        ports:
        - containerPort: 5432
          name: postgresql
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgresql-secret
              key: password
        - name: POSTGRES_USER
          value: "notifi"
        - name: POSTGRES_DB
          value: "notifi"
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgresql
  namespace: data-services
spec:
  clusterIP: None
  ports:
  - port: 5432
    targetPort: 5432
  selector:
    app: postgresql
---
apiVersion: v1
kind: Service
metadata:
  name: postgresql-primary
  namespace: data-services
spec:
  ports:
  - port: 5432
    targetPort: 5432
  selector:
    app: postgresql
    role: master
```

Create secret and deploy:

```bash
# Create PostgreSQL password secret
kubectl create secret generic postgresql-secret \
  --from-literal=password=$(openssl rand -base64 32) \
  --namespace data-services \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > postgresql-sealed-secret.yaml

kubectl apply -f postgresql-sealed-secret.yaml
kubectl apply -f postgresql-namespace.yaml
kubectl apply -f postgresql-patroni.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgresql -n data-services --timeout=300s

# Test connection
kubectl run -it --rm psql-test --image=postgres:16-alpine --restart=Never -n data-services -- \
  psql -h postgresql-primary.data-services.svc.cluster.local -U notifi -d notifi
```

### Step 3.2: Deploy Redis Sentinel

```yaml
# redis-sentinel.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: data-services
spec:
  serviceName: redis
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: redis
              topologyKey: kubernetes.io/hostname
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
          name: redis
        command:
        - redis-server
        args:
        - --appendonly
        - "yes"
        - --maxmemory
        - "2gb"
        - --maxmemory-policy
        - "allkeys-lru"
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
      - name: sentinel
        image: redis:7-alpine
        ports:
        - containerPort: 26379
          name: sentinel
        command:
        - redis-sentinel
        args:
        - /etc/redis/sentinel.conf
        volumeMounts:
        - name: sentinel-config
          mountPath: /etc/redis
      volumes:
      - name: sentinel-config
        configMap:
          name: redis-sentinel-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: data-services
spec:
  clusterIP: None
  ports:
  - port: 6379
    name: redis
  - port: 26379
    name: sentinel
  selector:
    app: redis
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-sentinel-config
  namespace: data-services
data:
  sentinel.conf: |
    port 26379
    sentinel monitor mymaster redis-0.redis.data-services.svc.cluster.local 6379 2
    sentinel down-after-milliseconds mymaster 5000
    sentinel failover-timeout mymaster 10000
    sentinel parallel-syncs mymaster 1
```

Deploy Redis:

```bash
kubectl apply -f redis-sentinel.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod -l app=redis -n data-services --timeout=300s

# Test Redis
kubectl run -it --rm redis-test --image=redis:7-alpine --restart=Never -n data-services -- \
  redis-cli -h redis-0.redis.data-services.svc.cluster.local ping
```

### Step 3.3: Deploy MongoDB Replica Set

```yaml
# mongodb-replicaset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
  namespace: data-services
spec:
  serviceName: mongodb
  replicas: 3
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:7
        ports:
        - containerPort: 27017
          name: mongodb
        env:
        - name: MONGO_INITDB_ROOT_USERNAME
          value: "admin"
        - name: MONGO_INITDB_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mongodb-secret
              key: password
        command:
        - mongod
        args:
        - --replSet
        - rs0
        - --bind_ip_all
        volumeMounts:
        - name: data
          mountPath: /data/db
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: mongodb
  namespace: data-services
spec:
  clusterIP: None
  ports:
  - port: 27017
    targetPort: 27017
  selector:
    app: mongodb
```

Deploy and initialize MongoDB:

```bash
# Create MongoDB secret
kubectl create secret generic mongodb-secret \
  --from-literal=password=$(openssl rand -base64 32) \
  --namespace data-services \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > mongodb-sealed-secret.yaml

kubectl apply -f mongodb-sealed-secret.yaml
kubectl apply -f mongodb-replicaset.yaml

# Wait for MongoDB pods
kubectl wait --for=condition=ready pod -l app=mongodb -n data-services --timeout=300s

# Initialize replica set
kubectl exec -it mongodb-0 -n data-services -- mongosh --eval '
  rs.initiate({
    _id: "rs0",
    members: [
      { _id: 0, host: "mongodb-0.mongodb.data-services.svc.cluster.local:27017" },
      { _id: 1, host: "mongodb-1.mongodb.data-services.svc.cluster.local:27017" },
      { _id: 2, host: "mongodb-2.mongodb.data-services.svc.cluster.local:27017", arbiterOnly: true }
    ]
  })
'

# Verify replica set
kubectl exec -it mongodb-0 -n data-services -- mongosh --eval 'rs.status()'
```

### Step 3.4: Deploy MinIO (S3-Compatible Storage)

```yaml
# minio-distributed.yaml
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: data-services
spec:
  clusterIP: None
  ports:
  - port: 9000
    name: minio
  - port: 9001
    name: console
  selector:
    app: minio
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: minio
  namespace: data-services
spec:
  serviceName: minio
  replicas: 3
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args:
        - server
        - http://minio-{0...2}.minio.data-services.svc.cluster.local/data
        - --console-address
        - ":9001"
        env:
        - name: MINIO_ROOT_USER
          value: "admin"
        - name: MINIO_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: minio-secret
              key: password
        ports:
        - containerPort: 9000
          name: minio
        - containerPort: 9001
          name: console
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 500Gi
---
apiVersion: v1
kind: Service
metadata:
  name: minio-api
  namespace: data-services
spec:
  type: ClusterIP
  ports:
  - port: 9000
    targetPort: 9000
  selector:
    app: minio
---
apiVersion: v1
kind: Service
metadata:
  name: minio-console
  namespace: data-services
spec:
  type: ClusterIP
  ports:
  - port: 9001
    targetPort: 9001
  selector:
    app: minio
```

Deploy MinIO:

```bash
# Create MinIO secret
kubectl create secret generic minio-secret \
  --from-literal=password=$(openssl rand -base64 32) \
  --namespace data-services \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > minio-sealed-secret.yaml

kubectl apply -f minio-sealed-secret.yaml
kubectl apply -f minio-distributed.yaml

# Wait for MinIO to be ready
kubectl wait --for=condition=ready pod -l app=minio -n data-services --timeout=300s

# Access MinIO console (port-forward)
kubectl port-forward svc/minio-console 9001:9001 -n data-services
# Open http://localhost:9001 in browser
```

### Validation Checklist Phase 3
- [ ] PostgreSQL 3-node cluster running
- [ ] Redis Sentinel 3-node cluster active
- [ ] MongoDB replica set initialized
- [ ] MinIO distributed storage accessible
- [ ] All databases passing health checks
- [ ] Data persistence verified

---

## 📨 Phase 4: Messaging Infrastructure

**Duration**: 2-3 days  
**Goal**: Deploy RabbitMQ, Kafka, and Knative Eventing

### Step 4.1: Deploy RabbitMQ Cluster

```bash
# Install RabbitMQ Cluster Operator
kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml

# Wait for operator to be ready
kubectl wait --for=condition=available deployment/rabbitmq-cluster-operator --timeout=300s
```

```yaml
# rabbitmq-cluster.yaml
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: notifi-rabbitmq
  namespace: messaging
spec:
  replicas: 3
  image: rabbitmq:3.12-management-alpine
  persistence:
    storageClassName: local-path
    storage: 20Gi
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 2Gi
  rabbitmq:
    additionalConfig: |
      cluster_formation.peer_discovery_backend = rabbit_peer_discovery_k8s
      cluster_formation.k8s.host = kubernetes.default.svc.cluster.local
      cluster_formation.k8s.address_type = hostname
      cluster_partition_handling = autoheal
      queue_master_locator = min-masters
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: notifi-rabbitmq
          topologyKey: kubernetes.io/hostname
```

Deploy RabbitMQ:

```bash
kubectl create namespace messaging
kubectl apply -f rabbitmq-cluster.yaml

# Wait for RabbitMQ cluster
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=notifi-rabbitmq -n messaging --timeout=600s

# Get RabbitMQ credentials
kubectl get secret notifi-rabbitmq-default-user -n messaging -o jsonpath='{.data.username}' | base64 -d
kubectl get secret notifi-rabbitmq-default-user -n messaging -o jsonpath='{.data.password}' | base64 -d

# Access RabbitMQ Management UI
kubectl port-forward svc/notifi-rabbitmq 15672:15672 -n messaging
# Open http://localhost:15672
```

### Step 4.2: Deploy Kafka Cluster

```bash
# Install Strimzi Kafka Operator
kubectl create namespace kafka
kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka

# Wait for operator
kubectl wait --for=condition=available deployment/strimzi-cluster-operator -n kafka --timeout=300s
```

```yaml
# kafka-cluster.yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: notifi-kafka
  namespace: kafka
spec:
  kafka:
    version: 3.6.0
    replicas: 2
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
    config:
      offsets.topic.replication.factor: 2
      transaction.state.log.replication.factor: 2
      transaction.state.log.min.isr: 1
      default.replication.factor: 2
      min.insync.replicas: 1
      inter.broker.protocol.version: "3.6"
    storage:
      type: persistent-claim
      size: 100Gi
      deleteClaim: false
    resources:
      requests:
        memory: 2Gi
        cpu: 1000m
      limits:
        memory: 4Gi
        cpu: 2000m
  zookeeper:
    replicas: 3
    storage:
      type: persistent-claim
      size: 10Gi
      deleteClaim: false
    resources:
      requests:
        memory: 1Gi
        cpu: 500m
      limits:
        memory: 2Gi
        cpu: 1000m
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

Deploy Kafka:

```bash
kubectl apply -f kafka-cluster.yaml

# Wait for Kafka cluster
kubectl wait kafka/notifi-kafka --for=condition=Ready --timeout=600s -n kafka

# Create test topic
kubectl apply -f - <<EOF
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: test-events
  namespace: kafka
  labels:
    strimzi.io/cluster: notifi-kafka
spec:
  partitions: 3
  replicas: 2
  config:
    retention.ms: 604800000
    segment.bytes: 1073741824
EOF

# Test Kafka
kubectl run kafka-producer -ti --image=quay.io/strimzi/kafka:0.38.0-kafka-3.6.0 --rm=true --restart=Never -n kafka -- \
  bin/kafka-console-producer.sh --bootstrap-server notifi-kafka-kafka-bootstrap:9092 --topic test-events
```

### Step 4.3: Deploy Knative Eventing

```bash
# Install Knative Eventing CRDs
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.0/eventing-crds.yaml

# Install Knative Eventing core
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.12.0/eventing-core.yaml

# Install RabbitMQ Broker
kubectl apply -f https://github.com/knative-sandbox/eventing-rabbitmq/releases/download/knative-v1.12.0/rabbitmq-broker.yaml

# Wait for Knative Eventing
kubectl wait --for=condition=available deployment/eventing-controller -n knative-eventing --timeout=300s
```

```yaml
# knative-broker.yaml
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: notifi-broker
  namespace: notifi
  annotations:
    eventing.knative.dev/broker.class: RabbitMQBroker
spec:
  config:
    apiVersion: v1
    kind: ConfigMap
    name: rabbitmq-broker-config
    namespace: notifi
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rabbitmq-broker-config
  namespace: notifi
data:
  brokerClass: RabbitMQBroker
  rabbitmqURL: "amqp://notifi-rabbitmq.messaging.svc.cluster.local:5672"
```

Deploy Knative Broker:

```bash
kubectl create namespace notifi
kubectl apply -f knative-broker.yaml

# Verify broker
kubectl get broker -n notifi
```

### Validation Checklist Phase 4
- [ ] RabbitMQ 3-node cluster operational
- [ ] Kafka 2-broker cluster running
- [ ] ZooKeeper 3-node ensemble active
- [ ] Knative Eventing installed
- [ ] Knative Broker connected to RabbitMQ
- [ ] Test messages flowing through queues

---

## 🚀 Phase 5: Notifi Services

**Duration**: 5-7 days  
**Goal**: Deploy all 37 Notifi microservices

### Step 5.1: Create Notifi Namespace and Secrets

```bash
# Create namespace
kubectl create namespace notifi
kubectl label namespace notifi linkerd.io/inject=enabled

# Create database connection secrets
kubectl create secret generic notifi-db-credentials \
  --from-literal=connection-string="postgresql://notifi:PASSWORD@postgresql-primary.data-services.svc.cluster.local:5432/notifi" \
  --from-literal=redis-url="redis://redis-0.redis.data-services.svc.cluster.local:6379" \
  --from-literal=mongodb-url="mongodb://admin:PASSWORD@mongodb-0.mongodb.data-services.svc.cluster.local:27017" \
  --namespace notifi \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > notifi-db-sealed-secret.yaml

kubectl apply -f notifi-db-sealed-secret.yaml

# Create external API secrets
kubectl create secret generic notifi-external-apis \
  --from-literal=twilio-account-sid="YOUR_TWILIO_SID" \
  --from-literal=twilio-auth-token="YOUR_TWILIO_TOKEN" \
  --from-literal=sendgrid-api-key="YOUR_SENDGRID_KEY" \
  --from-literal=telegram-bot-token="YOUR_TELEGRAM_TOKEN" \
  --from-literal=fcm-server-key="YOUR_FCM_KEY" \
  --namespace notifi \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > notifi-apis-sealed-secret.yaml

kubectl apply -f notifi-apis-sealed-secret.yaml
```

### Step 5.2: Deploy Management Gateway

```yaml
# management-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: management-gateway
  namespace: notifi
  annotations:
    linkerd.io/inject: enabled
spec:
  replicas: 2
  selector:
    matchLabels:
      app: management-gateway
  template:
    metadata:
      labels:
        app: management-gateway
        version: v1
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: management-gateway
              topologyKey: kubernetes.io/hostname
      containers:
      - name: management-gateway
        image: your-registry/management-gateway:latest
        ports:
        - containerPort: 5000
          name: http
        - containerPort: 5001
          name: grpc
        env:
        - name: ASPNETCORE_ENVIRONMENT
          value: "Production"
        - name: ConnectionStrings__PostgreSQL
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: connection-string
        - name: ConnectionStrings__Redis
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: redis-url
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 5000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 5000
          initialDelaySeconds: 10
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: management-gateway
  namespace: notifi
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 5000
    name: http
  - port: 5001
    targetPort: 5001
    name: grpc
  selector:
    app: management-gateway
---
apiVersion: policy.linkerd.io/v1beta1
kind: Server
metadata:
  name: management-gateway-server
  namespace: notifi
spec:
  podSelector:
    matchLabels:
      app: management-gateway
  port: grpc
  proxyProtocol: gRPC
```

### Step 5.3: Deploy Dataplane Gateway

```yaml
# dataplane-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dataplane-gateway
  namespace: notifi
spec:
  replicas: 2
  selector:
    matchLabels:
      app: dataplane-gateway
  template:
    metadata:
      labels:
        app: dataplane-gateway
        version: v1
    spec:
      containers:
      - name: dataplane-gateway
        image: your-registry/dataplane-gateway:latest
        ports:
        - containerPort: 5000
          name: http
        - containerPort: 5001
          name: grpc
        env:
        - name: ASPNETCORE_ENVIRONMENT
          value: "Production"
        - name: ConnectionStrings__PostgreSQL
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: connection-string
        - name: ConnectionStrings__Redis
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: redis-url
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
---
apiVersion: v1
kind: Service
metadata:
  name: dataplane-gateway
  namespace: notifi
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 5000
    name: http
  - port: 5001
    targetPort: 5001
    name: grpc
  selector:
    app: dataplane-gateway
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: dataplane-gateway-hpa
  namespace: notifi
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: dataplane-gateway
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Step 5.4: Deploy Event Processor (Core Engine)

```yaml
# event-processor.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: event-processor
  namespace: notifi
spec:
  replicas: 3
  selector:
    matchLabels:
      app: event-processor
  template:
    metadata:
      labels:
        app: event-processor
        version: v1
    spec:
      containers:
      - name: event-processor
        image: your-registry/event-processor:latest
        ports:
        - containerPort: 5000
          name: http
        env:
        - name: ASPNETCORE_ENVIRONMENT
          value: "Production"
        - name: ConnectionStrings__PostgreSQL
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: connection-string
        - name: ConnectionStrings__Redis
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: redis-url
        - name: RabbitMQ__HostName
          value: "notifi-rabbitmq.messaging.svc.cluster.local"
        - name: RabbitMQ__Port
          value: "5672"
        - name: RabbitMQ__UserName
          value: "notifi"
        - name: RabbitMQ__Password
          valueFrom:
            secretKeyRef:
              name: rabbitmq-credentials
              key: password
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
---
apiVersion: v1
kind: Service
metadata:
  name: event-processor
  namespace: notifi
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 5000
  selector:
    app: event-processor
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: event-processor-hpa
  namespace: notifi
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: event-processor
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

### Step 5.5: Deploy Messaging Workers (Knative Services)

```yaml
# telegram-messenger.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: telegram-messenger
  namespace: notifi
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "5"
        autoscaling.knative.dev/target: "10"
    spec:
      containers:
      - image: your-registry/telegram-messenger:latest
        ports:
        - containerPort: 8080
        env:
        - name: TELEGRAM_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: notifi-external-apis
              key: telegram-bot-token
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
# sms-messenger.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: sms-messenger
  namespace: notifi
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "5"
    spec:
      containers:
      - image: your-registry/sms-messenger:latest
        ports:
        - containerPort: 8080
        env:
        - name: TWILIO_ACCOUNT_SID
          valueFrom:
            secretKeyRef:
              name: notifi-external-apis
              key: twilio-account-sid
        - name: TWILIO_AUTH_TOKEN
          valueFrom:
            secretKeyRef:
              name: notifi-external-apis
              key: twilio-auth-token
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
---
# webhook-sender.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: webhook-sender
  namespace: notifi
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "10"
    spec:
      containers:
      - image: your-registry/webhook-sender:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
```

### Step 5.6: Deploy Blockchain Proxies

```yaml
# solana-proxy.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: solana-proxy
  namespace: notifi
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "3"
    spec:
      containers:
      - image: your-registry/solana-proxy:latest
        ports:
        - containerPort: 8080
        env:
        - name: SOLANA_RPC_URL
          value: "https://api.mainnet-beta.solana.com"
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: redis-url
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
---
# evm-proxy.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: evm-proxy
  namespace: notifi
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "3"
    spec:
      containers:
      - image: your-registry/evm-proxy:latest
        ports:
        - containerPort: 8080
        env:
        - name: ETH_RPC_URL
          value: "https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY"
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: notifi-db-credentials
              key: redis-url
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
```

### Step 5.7: Create Ingress Routes

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: notifi-ingress
  namespace: notifi
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  ingressClassName: traefik
  tls:
  - hosts:
    - api.notifi.yourdomain.com
    secretName: notifi-tls
  rules:
  - host: api.notifi.yourdomain.com
    http:
      paths:
      - path: /management
        pathType: Prefix
        backend:
          service:
            name: management-gateway
            port:
              number: 80
      - path: /dataplane
        pathType: Prefix
        backend:
          service:
            name: dataplane-gateway
            port:
              number: 80
```

### Validation Checklist Phase 5
- [ ] Management Gateway accessible
- [ ] Dataplane Gateway responding
- [ ] Event Processor consuming messages
- [ ] Messaging workers scaling to zero
- [ ] Blockchain proxies caching queries
- [ ] All services healthy
- [ ] Ingress routes working
- [ ] TLS certificates issued

---

## 📊 Phase 6: Observability & Monitoring

**Duration**: 2-3 days  
**Goal**: Full stack observability with Grafana, Prometheus, Loki, Tempo

### Step 6.1: Install Prometheus Stack

```bash
# Add Prometheus Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install kube-prometheus-stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.retention=30d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=100Gi \
  --set grafana.enabled=true \
  --set grafana.adminPassword=<your-password>
```

### Step 6.2: Install Loki for Logs

```bash
# Install Loki
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set loki.persistence.enabled=true \
  --set loki.persistence.size=50Gi \
  --set promtail.enabled=true
```

### Step 6.3: Install Tempo for Tracing

```bash
# Install Tempo
helm install tempo grafana/tempo \
  --namespace monitoring \
  --set tempo.storage.trace.backend=s3 \
  --set tempo.storage.trace.s3.bucket=tempo-traces \
  --set tempo.storage.trace.s3.endpoint=minio-api.data-services.svc.cluster.local:9000
```

### Step 6.4: Configure Grafana Dashboards

```yaml
# grafana-dashboards-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: notifi-dashboards
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
data:
  notifi-overview.json: |
    {
      "dashboard": {
        "title": "Notifi Overview",
        "panels": [
          {
            "title": "API Gateway RPS",
            "targets": [
              {
                "expr": "sum(rate(http_requests_total{namespace=\"notifi\"}[5m]))"
              }
            ]
          },
          {
            "title": "Event Processor Queue Depth",
            "targets": [
              {
                "expr": "rabbitmq_queue_messages{namespace=\"messaging\"}"
              }
            ]
          },
          {
            "title": "Database Connections",
            "targets": [
              {
                "expr": "pg_stat_database_numbackends{namespace=\"data-services\"}"
              }
            ]
          }
        ]
      }
    }
```

### Step 6.5: Configure Alerting

```yaml
# prometheus-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: notifi-alerts
  namespace: monitoring
spec:
  groups:
  - name: notifi
    interval: 30s
    rules:
    - alert: HighErrorRate
      expr: |
        sum(rate(http_requests_total{namespace="notifi",status=~"5.."}[5m])) /
        sum(rate(http_requests_total{namespace="notifi"}[5m])) > 0.05
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High error rate detected"
        description: "Error rate is {{ $value | humanizePercentage }}"
    
    - alert: EventProcessorDown
      expr: up{job="event-processor"} == 0
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "Event Processor is down"
        description: "Event Processor has been down for more than 2 minutes"
    
    - alert: DatabaseHighConnections
      expr: pg_stat_database_numbackends > 80
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High database connections"
        description: "PostgreSQL has {{ $value }} active connections"
    
    - alert: RabbitMQQueueDepth
      expr: rabbitmq_queue_messages > 10000
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "High RabbitMQ queue depth"
        description: "Queue has {{ $value }} messages"
```

### Validation Checklist Phase 6
- [ ] Prometheus collecting metrics
- [ ] Grafana dashboards displaying data
- [ ] Loki aggregating logs
- [ ] Tempo tracing requests
- [ ] AlertManager sending alerts
- [ ] All service metrics visible

---

## 🔒 Phase 7: Production Readiness

**Duration**: 3-5 days  
**Goal**: Hardening, backups, documentation

### Step 7.1: Configure Backups with Velero

```bash
# Install Velero CLI
brew install velero

# Install Velero in cluster (using MinIO)
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.8.0 \
  --bucket velero-backups \
  --secret-file ./minio-credentials \
  --use-node-agent \
  --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://minio-api.data-services.svc.cluster.local:9000

# Create backup schedule
velero schedule create daily-backup \
  --schedule="0 2 * * *" \
  --ttl 720h0m0s

# Test backup
velero backup create test-backup --wait
velero backup describe test-backup
```

### Step 7.2: Configure Database Backups

```yaml
# postgresql-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgresql-backup
  namespace: data-services
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgresql-primary.data-services.svc.cluster.local \
                -U notifi -d notifi --no-password \
                | gzip > /backups/notifi-$(date +%Y%m%d).sql.gz
              
              # Upload to MinIO
              mc alias set minio http://minio-api.data-services.svc.cluster.local:9000 admin $MINIO_PASSWORD
              mc cp /backups/notifi-$(date +%Y%m%d).sql.gz minio/database-backups/
              
              # Cleanup old backups (keep 30 days)
              find /backups -name "notifi-*.sql.gz" -mtime +30 -delete
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgresql-secret
                  key: password
            - name: MINIO_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: minio-secret
                  key: password
            volumeMounts:
            - name: backup-storage
              mountPath: /backups
          restartPolicy: OnFailure
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
```

### Step 7.3: Implement Network Policies

```yaml
# network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: notifi-allow-ingress
  namespace: notifi
spec:
  podSelector:
    matchLabels:
      app: management-gateway
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: traefik
    ports:
    - protocol: TCP
      port: 5000
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: database-access
  namespace: data-services
spec:
  podSelector:
    matchLabels:
      app: postgresql
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: notifi
    ports:
    - protocol: TCP
      port: 5432
```

### Step 7.4: Resource Quotas

```yaml
# resource-quotas.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: notifi-quota
  namespace: notifi
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    persistentvolumeclaims: "20"
---
apiVersion: v1
kind: LimitRange
metadata:
  name: notifi-limits
  namespace: notifi
spec:
  limits:
  - max:
      cpu: "4"
      memory: 8Gi
    min:
      cpu: "100m"
      memory: 128Mi
    default:
      cpu: "500m"
      memory: 1Gi
    defaultRequest:
      cpu: "250m"
      memory: 512Mi
    type: Container
```

### Step 7.5: Create Operations Runbook

Create a comprehensive runbook covering:
- Service restart procedures
- Database failover process
- Backup and restore procedures
- Scaling guidelines
- Troubleshooting common issues
- Emergency contacts
- SLA definitions

### Validation Checklist Phase 7
- [ ] Automated backups running
- [ ] Backup restoration tested
- [ ] Network policies enforced
- [ ] Resource quotas in place
- [ ] Documentation complete
- [ ] Team trained on operations
- [ ] Disaster recovery plan tested

---

## 🛠️ Operations & Maintenance

### Daily Operations

```bash
# Check cluster health
kubectl get nodes
kubectl top nodes

# Check pod status
kubectl get pods -A | grep -v Running

# Check recent events
kubectl get events -A --sort-by='.lastTimestamp' | tail -20

# Check Linkerd health
linkerd check

# Check Flux sync status
flux get all
```

### Weekly Maintenance

```bash
# Update Flux
flux check --pre
flux install --export > flux-system.yaml
kubectl apply -f flux-system.yaml

# Check for security updates
kubectl get vulnerabilityreports -A

# Review backup status
velero backup get
velero restore get

# Check disk usage
kubectl exec -it postgresql-0 -n data-services -- df -h

# Review logs for errors
kubectl logs -n notifi -l app=event-processor --tail=1000 | grep ERROR
```

### Monthly Tasks

- [ ] Review and update resource quotas
- [ ] Rotate database passwords
- [ ] Update all container images
- [ ] Test disaster recovery procedures
- [ ] Review monitoring dashboards
- [ ] Clean up old backups
- [ ] Update documentation

---

## 🔧 Troubleshooting Guide

### Issue: Pod Stuck in Pending

```bash
# Check why pod is pending
kubectl describe pod <pod-name> -n <namespace>

# Common causes:
# 1. Insufficient resources
kubectl describe nodes | grep -A 5 "Allocated resources"

# 2. PVC not bound
kubectl get pvc -A

# 3. Node affinity not satisfied
kubectl get nodes --show-labels
```

### Issue: Service Not Accessible

```bash
# Check service endpoints
kubectl get endpoints <service-name> -n <namespace>

# Check if pods are ready
kubectl get pods -l app=<app-name> -n <namespace>

# Test service connectivity
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://<service-name>.<namespace>.svc.cluster.local
```

### Issue: Database Connection Errors

```bash
# Check PostgreSQL status
kubectl exec -it postgresql-0 -n data-services -- psql -U notifi -c "SELECT version();"

# Check connections
kubectl exec -it postgresql-0 -n data-services -- psql -U notifi -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Check replication lag
kubectl exec -it postgresql-0 -n data-services -- psql -U notifi -c \
  "SELECT * FROM pg_stat_replication;"
```

### Issue: High Memory Usage

```bash
# Check memory usage
kubectl top pods -A --sort-by=memory

# Identify memory leaks
kubectl logs <pod-name> -n <namespace> | grep -i "out of memory"

# Restart pod
kubectl rollout restart deployment/<deployment-name> -n <namespace>
```

---

## 🚀 Migration from Cloud

### Pre-Migration Checklist

- [ ] Data export from AWS RDS
- [ ] Object storage sync from S3
- [ ] DNS TTL reduced to 60 seconds
- [ ] Backup of all configurations
- [ ] Team notified of maintenance window
- [ ] Rollback plan documented

### Migration Steps

```bash
# 1. Export data from AWS
pg_dump -h aws-rds-endpoint -U postgres -d notifi > notifi-export.sql

# 2. Import to homelab
kubectl exec -it postgresql-0 -n data-services -- psql -U notifi -d notifi < notifi-export.sql

# 3. Sync S3 to MinIO
aws s3 sync s3://notifi-production minio://notifi-production --endpoint-url http://minio:9000

# 4. Update DNS
# Point notifi.yourdomain.com to new IP

# 5. Monitor traffic
kubectl logs -f -n notifi -l app=management-gateway

# 6. Verify functionality
curl https://api.notifi.yourdomain.com/health
```

---

## 📚 Additional Resources

### Documentation
- [K3s Documentation](https://docs.k3s.io/)
- [Linkerd Documentation](https://linkerd.io/2/overview/)
- [Flux Documentation](https://fluxcd.io/docs/)
- [Knative Documentation](https://knative.dev/docs/)

### Monitoring & Troubleshooting
- Grafana Dashboard: `http://<node-ip>:30300`
- Prometheus: `http://<node-ip>:30900`
- Linkerd Viz: `linkerd viz dashboard`
- RabbitMQ Management: `http://<node-ip>:31672`

### Community Support
- Kubernetes Slack
- Linkerd Slack
- Flux Discord
- /r/selfhosted

---

## 📝 Appendix

### A. Hardware Shopping List

See [Hardware Requirements](#hardware-requirements) section

### B. Network Diagram

```
[Internet]
    │
[Router/Firewall] 192.168.1.1
    │
    ├── [Nginx Edge] 192.168.1.10
    │
    ├── [Pi Node 1] 192.168.1.101 (Control Plane)
    ├── [Pi Node 2] 192.168.1.102 (Worker)
    ├── [Pi Node 3] 192.168.1.103 (Worker)
    ├── [Studio] 192.168.1.104 (Worker)
    └── [Forge] 192.168.1.105 (GPU Worker)
```

### C. Estimated Timeline

| Phase | Duration | Complexity |
|-------|----------|------------|
| Phase 1: Infrastructure | 3-5 days | Medium |
| Phase 2: Platform Services | 3-4 days | Medium |
| Phase 3: Data Layer | 4-5 days | High |
| Phase 4: Messaging | 2-3 days | Medium |
| Phase 5: Notifi Services | 5-7 days | High |
| Phase 6: Observability | 2-3 days | Medium |
| Phase 7: Production Ready | 3-5 days | High |
| **Total** | **4-6 weeks** | **High** |

### D. Cost Analysis

**One-Time Costs:**
- Hardware: $2,810
- Networking: $100
- Total: **$2,910**

**Monthly Costs:**
- Power (600W × 730hrs × $0.12/kWh): $53
- Internet (1Gbps fiber): $80
- Total: **$133/month** or **$1,596/year**

**Year 1 Total: $4,506**

**AWS Equivalent: $42,000+/year**

**Savings: 89%**

---

**End of Implementation Plan**

*This plan is a living document. Update as your deployment evolves.*

