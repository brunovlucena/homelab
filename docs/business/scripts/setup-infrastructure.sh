#!/bin/bash
# Week 1 Infrastructure Setup Script
# Run this on your server after installing Ubuntu Server 22.04 LTS

set -e

echo "🚀 Starting Infrastructure Setup for AI Agentic Systems SaaS Platform"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
   echo "❌ Please run as regular user (not root)"
   exit 1
fi

echo -e "${GREEN}Step 1: System Update${NC}"
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git vim htop net-tools

echo -e "${GREEN}Step 2: Install k3s (Kubernetes)${NC}"
curl -sfL https://get.k3s.io | sh -
sudo systemctl enable k3s
sudo systemctl start k3s

# Wait for k3s to be ready
echo "Waiting for k3s to be ready..."
sleep 10

# Set up kubectl
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $USER:$USER ~/.kube/config
export KUBECONFIG=~/.kube/config

# Verify k3s
echo "Verifying k3s installation..."
kubectl get nodes

echo -e "${GREEN}Step 3: Install Flux (GitOps)${NC}"
curl -s https://fluxcd.io/install.sh | sudo bash
export PATH=$PATH:$HOME/.local/bin
flux install

# Wait for Flux
echo "Waiting for Flux to be ready..."
sleep 30
kubectl get pods -n flux-system

echo -e "${GREEN}Step 4: Install Linkerd (Service Mesh)${NC}"
curl -sL https://run.linkerd.io/install | sh
export PATH=$PATH:$HOME/.linkerd2/bin
linkerd install | kubectl apply -f -

# Wait for Linkerd
echo "Waiting for Linkerd to be ready..."
sleep 60
linkerd check

echo -e "${GREEN}Step 5: Install Knative (Serverless)${NC}"
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.14.0/serving-core.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.14.0/eventing-core.yaml
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.14.0/kourier.yaml

# Wait for Knative
echo "Waiting for Knative to be ready..."
sleep 60
kubectl get pods -n knative-serving
kubectl get pods -n knative-eventing

echo -e "${GREEN}Step 6: Install Observability Stack${NC}"

# Create monitoring namespace
kubectl create namespace monitoring || true

# Install Prometheus Operator
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml

# Wait for Prometheus Operator
echo "Waiting for Prometheus Operator..."
sleep 60
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus-operator -n default --timeout=300s || true

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

# Install Grafana
kubectl apply -f - <<EOF
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
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          value: "admin"
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
  type: ClusterIP
EOF

echo -e "${GREEN}Step 7: Configure Firewall${NC}"
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 6443/tcp
sudo ufw --force enable

echo -e "${GREEN}Step 8: Install cert-manager (SSL/TLS)${NC}"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager
echo "Waiting for cert-manager..."
sleep 60
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s || true

echo ""
echo -e "${GREEN}✅ Infrastructure Setup Complete!${NC}"
echo ""
echo "Next Steps:"
echo "1. Access Grafana: kubectl port-forward -n monitoring svc/grafana 3000:3000"
echo "2. Access Prometheus: kubectl port-forward -n monitoring svc/prometheus 9090:9090"
echo "3. Check all pods: kubectl get pods --all-namespaces"
echo ""
echo "Grafana credentials: admin / admin (change on first login)"
echo ""
