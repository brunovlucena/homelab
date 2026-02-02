# Week 1 Execution Guide: Infrastructure Setup

**Step-by-step guide to execute Week 1 tasks from the launch plan**

---

## Day 1-2: Infrastructure Planning

### Task 1: Choose Hosting Location

**Options**:
- **Home**: If you have good internet (1Gbps fiber), static IP, reliable power
- **Colocation**: If you need better uptime, professional hosting
- **Cloud**: Backup plan (Hetzner, OVH, DigitalOcean)

**Recommendation**: Start with **home** (if internet is good) or **colocation** (if you need reliability)

**Action Items**:
- [ ] Evaluate your home internet (speed, reliability, static IP option)
- [ ] Research colocation providers in your area
- [ ] Compare costs: Home ($80/month internet) vs. Colocation ($100-200/month)
- [ ] **Decision**: [ ] Home [ ] Colocation [ ] Cloud

---

### Task 2: Order/Lease Server

**Specifications**:
- CPU: 16-core x86_64 (AMD EPYC or Intel Xeon)
- RAM: 64GB DDR4
- Storage: 2TB NVMe SSD
- Network: 1Gbps
- Cost: $1,500 (purchase) or $400/month (lease)

**Where to Buy/Lease**:
- **Hetzner** (Germany): Dedicated servers, €50-100/month
- **OVH** (France): Dedicated servers, €50-100/month
- **Online.net** (France): Dedicated servers, €50-100/month
- **Local Providers**: Check local data centers in your area

**Action Items**:
- [ ] Research server providers
- [ ] Compare prices and specs
- [ ] Order/lease server
- [ ] **Expected Delivery**: [Date]

**Providers to Check**:
- Hetzner: https://www.hetzner.com/dedicated-rootserver
- OVH: https://www.ovhcloud.com/en/dedicated-servers/
- Online.net: https://www.online.net/en/dedicated-server

---

### Task 3: Set Up Internet

**Requirements**:
- Speed: 1Gbps fiber minimum
- Static IP: Recommended (for remote access)
- Business Plan: Better SLA, static IP included

**Action Items**:
- [ ] Check current internet plan
- [ ] Upgrade to 1Gbps fiber (if needed)
- [ ] Request static IP (if available)
- [ ] Test speed: `speedtest-cli` or speedtest.net
- [ ] **Expected Setup**: [Date]

**If Home Hosting**:
- [ ] Set up port forwarding (ports 80, 443, 22)
- [ ] Configure router firewall
- [ ] Set up DDNS (if no static IP): DuckDNS, No-IP

---

### Task 4: Order Domain

**Domain Suggestions**:
- agenticplatform.com
- agenticsystems.com
- aiagents.io
- agentic.ai (if available)
- agenticplatform.io

**Providers**:
- **Namecheap**: $10-15/year
- **Google Domains**: $12/year
- **Cloudflare**: $8-10/year (recommended - includes DNS)

**Action Items**:
- [ ] Check domain availability
- [ ] Register domain
- [ ] Set up DNS (Cloudflare recommended)
- [ ] **Domain Registered**: [Domain Name]

---

### Task 5: Set Up Cloudflare

**Free Tier Includes**:
- DNS management
- DDoS protection
- SSL/TLS certificates
- CDN (optional)

**Action Items**:
- [ ] Create Cloudflare account (free)
- [ ] Add domain to Cloudflare
- [ ] Update nameservers at domain registrar
- [ ] Set up DNS records:
  - A record: @ → [Your Server IP]
  - A record: www → [Your Server IP]
- [ ] Enable SSL/TLS (Full mode)
- [ ] **Cloudflare Setup**: Complete

**Cloudflare Tunnel (Optional - for remote access)**:
- [ ] Install cloudflared on server
- [ ] Set up tunnel for remote access
- [ ] Configure routes

---

## Day 3-5: Server Setup

### Task 1: Install Ubuntu Server 22.04 LTS

**Download**: https://ubuntu.com/download/server

**Installation Steps**:
1. Download ISO
2. Create bootable USB (Rufus, Balena Etcher)
3. Boot server from USB
4. Install Ubuntu Server 22.04 LTS
5. Set up user account
6. Enable SSH

**Action Items**:
- [ ] Download Ubuntu Server 22.04 LTS ISO
- [ ] Create bootable USB
- [ ] Install on server
- [ ] Set up SSH access
- [ ] **Server Installed**: [Date]

**Post-Installation**:
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl wget git vim htop

# Set up firewall
sudo ufw allow 22/tcp  # SSH
sudo ufw allow 80/tcp  # HTTP
sudo ufw allow 443/tcp # HTTPS
sudo ufw enable
```

---

### Task 2: Install k3s (Kubernetes)

**k3s** is a lightweight Kubernetes distribution, perfect for single-server deployments.

**Installation**:
```bash
# Install k3s
curl -sfL https://get.k3s.io | sh -

# Check status
sudo systemctl status k3s

# Get kubeconfig
sudo cat /etc/rancher/k3s/k3s.yaml

# Set up kubectl
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $USER:$USER ~/.kube/config

# Test
kubectl get nodes
```

**Action Items**:
- [ ] Install k3s
- [ ] Verify installation (`kubectl get nodes`)
- [ ] Set up kubectl access
- [ ] **k3s Installed**: [Date]

---

### Task 3: Install Flux (GitOps)

**Flux** manages Kubernetes deployments via Git.

**Installation**:
```bash
# Install Flux CLI
curl -s https://fluxcd.io/install.sh | sudo bash

# Install Flux on cluster
flux install

# Verify
kubectl get pods -n flux-system
```

**Action Items**:
- [ ] Install Flux CLI
- [ ] Install Flux on cluster
- [ ] Verify installation
- [ ] **Flux Installed**: [Date]

---

### Task 4: Install Linkerd (Service Mesh)

**Linkerd** provides service mesh, mTLS, and observability.

**Installation**:
```bash
# Install Linkerd CLI
curl -sL https://run.linkerd.io/install | sh

# Add to PATH
export PATH=$PATH:$HOME/.linkerd2/bin

# Install Linkerd
linkerd install | kubectl apply -f -

# Verify
linkerd check
```

**Action Items**:
- [ ] Install Linkerd CLI
- [ ] Install Linkerd on cluster
- [ ] Verify installation (`linkerd check`)
- [ ] **Linkerd Installed**: [Date]

---

### Task 5: Install Knative (Serverless)

**Knative** enables serverless workloads (scale-to-zero).

**Installation**:
```bash
# Install Knative Serving
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.14.0/serving-core.yaml

# Install Knative Eventing
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.14.0/eventing-core.yaml

# Install Kourier (Ingress)
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.14.0/kourier.yaml

# Verify
kubectl get pods -n knative-serving
kubectl get pods -n knative-eventing
```

**Action Items**:
- [ ] Install Knative Serving
- [ ] Install Knative Eventing
- [ ] Install Kourier Ingress
- [ ] Verify installation
- [ ] **Knative Installed**: [Date]

---

### Task 6: Install Observability Stack

**Prometheus** (Metrics):
```bash
# Install Prometheus Operator
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml

# Wait for operator
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus-operator -n default --timeout=300s

# Install Prometheus
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
---
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector: {}
  podMonitorSelector: {}
  ruleSelector: {}
EOF
```

**Grafana** (Dashboards):
```bash
# Install Grafana
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:latest
        ports:
        - containerPort: 3000
        volumeMounts:
        - name: grafana-storage
          mountPath: /var/lib/grafana
      volumes:
      - name: grafana-storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: monitoring
spec:
  selector:
    app: grafana
  ports:
  - port: 3000
    targetPort: 3000
  type: LoadBalancer
EOF
```

**Loki** (Logs):
```bash
# Install Loki (simplified)
kubectl apply -f https://raw.githubusercontent.com/grafana/loki/main/production/helm/loki-stack.yaml
```

**Action Items**:
- [ ] Install Prometheus
- [ ] Install Grafana
- [ ] Install Loki
- [ ] Access Grafana: `kubectl port-forward -n monitoring svc/grafana 3000:3000`
- [ ] **Observability Stack Installed**: [Date]

---

### Task 7: Set Up Backups

**Velero** (Backup Tool):
```bash
# Install Velero CLI
wget https://github.com/vmware-tanzu/velero/releases/download/v1.12.0/velero-v1.12.0-linux-amd64.tar.gz
tar -xzf velero-v1.12.0-linux-amd64.tar.gz
sudo mv velero-v1.12.0-linux-amd64/velero /usr/local/bin/

# Install Velero on cluster (requires S3/MinIO)
# For now, set up basic backup script
```

**Simple Backup Script**:
```bash
#!/bin/bash
# backup-k8s.sh

BACKUP_DIR="/backup/k8s"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

# Backup all resources
kubectl get all --all-namespaces -o yaml > $BACKUP_DIR/all-resources-$DATE.yaml

# Backup etcd (if accessible)
# kubectl exec -n kube-system etcd-<pod> -- etcdctl snapshot save /backup/etcd-$DATE.db

echo "Backup completed: $BACKUP_DIR/all-resources-$DATE.yaml"
```

**Action Items**:
- [ ] Install Velero (or create backup script)
- [ ] Set up backup storage (S3/MinIO)
- [ ] Schedule daily backups (cron job)
- [ ] Test restore process
- [ ] **Backups Configured**: [Date]

---

## Day 6-7: Security Hardening

### Task 1: Configure Firewall

```bash
# Allow SSH
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow Kubernetes API (if needed)
sudo ufw allow 6443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

**Action Items**:
- [ ] Configure UFW firewall
- [ ] Allow necessary ports only
- [ ] Test firewall rules
- [ ] **Firewall Configured**: [Date]

---

### Task 2: Set Up SSL/TLS

**cert-manager** (Let's Encrypt):
```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for installation
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s

# Create Let's Encrypt issuer
kubectl apply -f - <<EOF
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
          class: nginx
EOF
```

**Action Items**:
- [ ] Install cert-manager
- [ ] Create Let's Encrypt issuer
- [ ] Test certificate generation
- [ ] **SSL/TLS Configured**: [Date]

---

### Task 3: Set Up Vault (Secret Management)

**HashiCorp Vault**:
```bash
# Install Vault (simplified - dev mode for testing)
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: vault
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vault
  namespace: vault
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vault
  template:
    metadata:
      labels:
        app: vault
    spec:
      containers:
      - name: vault
        image: hashicorp/vault:latest
        command: ["vault", "server", "-dev"]
        ports:
        - containerPort: 8200
        env:
        - name: VAULT_DEV_ROOT_TOKEN_ID
          value: "root"
        - name: VAULT_DEV_LISTEN_ADDRESS
          value: "0.0.0.0:8200"
EOF
```

**Action Items**:
- [ ] Install Vault
- [ ] Initialize Vault
- [ ] Create secrets for agents
- [ ] **Vault Configured**: [Date]

---

### Task 4: Configure RBAC

```bash
# Create namespace for agents
kubectl create namespace ai-agents

# Create ServiceAccount
kubectl create serviceaccount agent-sa -n ai-agents

# Create Role
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: agent-role
  namespace: ai-agents
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
EOF

# Create RoleBinding
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: agent-rolebinding
  namespace: ai-agents
subjects:
- kind: ServiceAccount
  name: agent-sa
  namespace: ai-agents
roleRef:
  kind: Role
  name: agent-role
  apiGroup: rbac.authorization.k8s.io
EOF
```

**Action Items**:
- [ ] Create namespaces
- [ ] Create ServiceAccounts
- [ ] Create Roles and RoleBindings
- [ ] Test RBAC
- [ ] **RBAC Configured**: [Date]

---

### Task 5: Set Up Monitoring & Alerting

**AlertManager**:
```bash
# Install AlertManager
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: alertmanager
  template:
    metadata:
      labels:
        app: alertmanager
    spec:
      containers:
      - name: alertmanager
        image: prom/alertmanager:latest
        ports:
        - containerPort: 9093
EOF
```

**Slack Integration** (Optional):
- [ ] Create Slack webhook
- [ ] Configure AlertManager to send to Slack
- [ ] Test alerts

**Action Items**:
- [ ] Install AlertManager
- [ ] Configure alert rules
- [ ] Set up Slack/PagerDuty (optional)
- [ ] Test alerts
- [ ] **Monitoring Configured**: [Date]

---

### Task 6: Run Security Scans

**Trivy** (Container Scanning):
```bash
# Install Trivy
sudo apt install -y wget apt-transport-https gnupg lsb-release
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt update
sudo apt install -y trivy

# Scan images
trivy image your-image:tag
```

**Falco** (Runtime Security):
```bash
# Install Falco (optional - more advanced)
# Follow: https://falco.org/docs/installation/
```

**Action Items**:
- [ ] Install Trivy
- [ ] Scan container images
- [ ] Fix critical vulnerabilities
- [ ] **Security Scans Complete**: [Date]

---

## Week 1 Completion Checklist

### Infrastructure
- [ ] Server ordered/leased
- [ ] Internet set up (1Gbps fiber)
- [ ] Domain registered
- [ ] Cloudflare configured

### Server Setup
- [ ] Ubuntu Server 22.04 LTS installed
- [ ] k3s installed and verified
- [ ] Flux installed and verified
- [ ] Linkerd installed and verified
- [ ] Knative installed and verified
- [ ] Observability stack installed (Prometheus, Grafana, Loki)
- [ ] Backups configured

### Security
- [ ] Firewall configured
- [ ] SSL/TLS set up (cert-manager)
- [ ] Vault installed
- [ ] RBAC configured
- [ ] Monitoring & alerting set up
- [ ] Security scans completed

### Verification
- [ ] All services running: `kubectl get pods --all-namespaces`
- [ ] Grafana accessible: `kubectl port-forward -n monitoring svc/grafana 3000:3000`
- [ ] Prometheus accessible: `kubectl port-forward -n monitoring svc/prometheus 9090:9090`
- [ ] Knative working: Deploy test service

**Week 1 Complete!** ✅

**Next**: Week 2 - Product Packaging

---

## Troubleshooting

### k3s Not Starting
```bash
# Check logs
sudo journalctl -u k3s -f

# Restart
sudo systemctl restart k3s
```

### Pods Not Starting
```bash
# Check pod status
kubectl get pods --all-namespaces

# Check pod logs
kubectl logs <pod-name> -n <namespace>

# Check events
kubectl get events --all-namespaces --sort-by='.lastTimestamp'
```

### Network Issues
```bash
# Check network policies
kubectl get networkpolicies --all-namespaces

# Check services
kubectl get svc --all-namespaces
```

---

**Ready to execute Week 1?** Start with Day 1 tasks! 🚀
