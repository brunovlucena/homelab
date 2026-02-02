#!/bin/bash
# Cleanup Red Team Exploit Artifacts
# ⚠️ AUTHORIZED TESTING ONLY

set -e

NAMESPACE="${1:-redteam-test}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║           🧹 Red Team Cleanup                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Confirm cleanup
read -p "⚠️  This will remove all exploit artifacts. Continue? (y/N) " confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    echo "Cleanup cancelled."
    exit 0
fi

echo ""

# Remove exploit namespace
echo -e "${YELLOW}[*] Removing test namespace...${NC}"
kubectl delete namespace "$NAMESPACE" --ignore-not-found
echo -e "${GREEN}[+] Namespace removed${NC}"

# Remove all labeled exploit resources
echo -e "${YELLOW}[*] Removing labeled exploit resources...${NC}"
kubectl delete lambdafunctions -A -l redteam=true --ignore-not-found
kubectl delete lambdaagents -A -l redteam=true --ignore-not-found
echo -e "${GREEN}[+] Labeled resources removed${NC}"

# Remove privilege escalation artifacts
echo -e "${YELLOW}[*] Removing privilege escalation artifacts...${NC}"
kubectl delete clusterrolebinding attacker-admin-binding --ignore-not-found
kubectl delete clusterrole attacker-admin --ignore-not-found
kubectl delete sa attacker-sa -n default --ignore-not-found
echo -e "${GREEN}[+] RBAC artifacts removed${NC}"

# Remove persistence mechanisms
echo -e "${YELLOW}[*] Removing persistence mechanisms...${NC}"
kubectl delete cronjob system-health-check -n default --ignore-not-found
kubectl delete daemonset node-monitor -n kube-system --ignore-not-found
echo -e "${GREEN}[+] Persistence mechanisms removed${NC}"

# Remove orphaned build jobs
echo -e "${YELLOW}[*] Removing orphaned build jobs...${NC}"
kubectl delete jobs -A -l lambda.knative.io/build=true --ignore-not-found
echo -e "${GREEN}[+] Build jobs removed${NC}"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                   Cleanup Complete                           ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}All red team artifacts have been removed.${NC}"
echo ""
echo "⚠️  Note: Check manually for any resources that may have been"
echo "   created with different names or in unexpected namespaces."
