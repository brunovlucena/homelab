#!/bin/bash

# GitHub Actions Runner Secret Creation Script
# This script helps create the necessary secrets for GitHub Actions runners

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v kubeseal &> /dev/null; then
        print_error "kubeseal is not installed or not in PATH"
        exit 1
    fi
    
    print_success "All dependencies are available"
}

# Get user input
get_user_input() {
    print_status "Gathering configuration information..."
    
    # GitHub token
    if [ -z "$GITHUB_TOKEN" ]; then
        echo -n "Enter your GitHub token (with repo and admin:org permissions): "
        read -s GITHUB_TOKEN
        echo
    fi
    
    if [ -z "$GITHUB_TOKEN" ]; then
        print_error "GitHub token is required"
        exit 1
    fi
    
    # Repository or organization
    echo -n "Enter repository name (e.g., 'owner/repo') or organization name: "
    read REPOSITORY
    
    if [ -z "$REPOSITORY" ]; then
        print_error "Repository name is required"
        exit 1
    fi
    
    # Runner labels
    echo -n "Enter runner labels (comma-separated, default: 'self-hosted,linux,x64,homelab-infrastructure'): "
    read LABELS
    LABELS=${LABELS:-"self-hosted,linux,x64,homelab-infrastructure"}
    
    # Runner group
    echo -n "Enter runner group (default: 'default'): "
    read RUNNER_GROUP
    RUNNER_GROUP=${RUNNER_GROUP:-"default"}
    
    print_success "Configuration gathered successfully"
}

# Create the secret
create_secret() {
    print_status "Creating GitHub Actions runner secret..."
    
    # Create temporary secret file
    cat > /tmp/github-actions-runner-secret.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: github-actions-runner-secret
  namespace: github-actions-runner
  labels:
    app.kubernetes.io/name: github-actions-runner
    app.kubernetes.io/instance: github-actions-runner
    app.kubernetes.io/component: ci-cd
    app.kubernetes.io/part-of: homelab-infrastructure
    app.kubernetes.io/managed-by: flux
type: Opaque
data:
  github_token: $(echo -n "$GITHUB_TOKEN" | base64)
  repository: $(echo -n "$REPOSITORY" | base64)
  labels: $(echo -n "$LABELS" | base64)
  runner_group: $(echo -n "$RUNNER_GROUP" | base64)
EOF
    
    # Seal the secret
    kubeseal --format=yaml --cert=public.pem < /tmp/github-actions-runner-secret.yaml > github-actions-runner-sealed-secret.yaml
    
    # Clean up temporary file
    rm /tmp/github-actions-runner-secret.yaml
    
    print_success "Sealed secret created: github-actions-runner-sealed-secret.yaml"
}

# Update RunnerDeployment with repository information
update_runner_deployment() {
    print_status "Updating RunnerDeployment with repository information..."
    
    # Create a patch for the RunnerDeployment
    cat > /tmp/runner-deployment-patch.yaml << EOF
spec:
  template:
    spec:
      repository: "$REPOSITORY"
      labels: "$LABELS"
      runnerGroup: "$RUNNER_GROUP"
EOF
    
    # Apply the patch
    kubectl patch runnerdeployment github-actions-runner-deployment -n github-actions-runner --patch-file=/tmp/runner-deployment-patch.yaml
    
    # Clean up
    rm /tmp/runner-deployment-patch.yaml
    
    print_success "RunnerDeployment updated successfully"
}

# Main execution
main() {
    print_status "Starting GitHub Actions runner secret creation..."
    
    check_dependencies
    get_user_input
    create_secret
    update_runner_deployment
    
    print_success "GitHub Actions runner setup completed!"
    print_status "Next steps:"
    echo "1. Apply the sealed secret: kubectl apply -f github-actions-runner-sealed-secret.yaml"
    echo "2. Check runner status: kubectl get runners -n github-actions-runner"
    echo "3. View runner logs: kubectl logs -n github-actions-runner -l app.kubernetes.io/name=github-actions-runner"
    echo "4. Monitor runner metrics in Prometheus/Grafana"
}

# Run main function
main "$@"
