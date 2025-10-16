# 🚀 Setup Guide

## Initial Setup for Homepage Development

This guide will help you set up your development environment for the Homepage application with the new versioning and branching workflow.

## Prerequisites

- **Git** (2.0+)
- **Docker** & **Docker Compose**
- **kubectl** (for Kubernetes deployments)
- **Flux CLI** (for GitOps)
- **Node.js** (24+) - for frontend
- **Go** (1.23+) - for API
- **make** - for build automation

## Step 1: Clone Repository

```bash
cd ~/workspace
git clone https://github.com/brunovlucena/homelab.git
cd homelab
```

## Step 2: Install Git Helper Functions

Add to your `~/.zshrc` (or `~/.bashrc`):

```bash
# Git Helper Functions for Homepage
source ~/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/scripts/git-helpers.sh
```

Reload your shell:

```bash
source ~/.zshrc  # or source ~/.bashrc
```

Verify installation:

```bash
ghelp
# Should display available commands
```

## Step 3: Set Up Development Environment

### Navigate to Homepage Directory

```bash
cd flux/clusters/homelab/infrastructure/homepage
```

### Install Dependencies

**API (Go):**
```bash
cd api
go mod download
cd ..
```

**Frontend (Node.js):**
```bash
cd frontend
npm install --legacy-peer-deps
cd ..
```

### Make Scripts Executable

```bash
chmod +x scripts/*.sh
```

## Step 4: Configure Git

### Set Up Git User (if not already done)

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### Set Up GPG Signing (Recommended)

```bash
# Generate GPG key
gpg --full-generate-key

# List keys
gpg --list-secret-keys --keyid-format=long

# Configure Git to use GPG
git config --global user.signingkey YOUR_KEY_ID
git config --global commit.gpgsign true
```

## Step 5: Start Local Development

### Using Docker Compose

```bash
# Start all services
make up

# View logs
make logs

# Access services:
# - Frontend: http://localhost:3000
# - API: http://localhost:8080
# - Prometheus: http://localhost:9090
```

### Or Run Services Individually

**API:**
```bash
make run-api
# Runs on http://localhost:8080
```

**Frontend:**
```bash
make dev-frontend
# Runs on http://localhost:5173 (Vite dev server)
```

## Step 6: Verify Setup

### Run Tests

```bash
# All tests
make test

# API tests only
make test-api

# Frontend tests only
make test-frontend
```

### Check Version

```bash
# Using make
make version

# Using helper function
gversion

# Detailed info
ginfo
```

## Step 7: Create Your First Branch

### Feature Branch

```bash
gfeature my first feature
# Creates: feature/2025-10-16/my-first-feature
```

### Make Changes and Commit

```bash
# Edit some files
# ...

# Commit with conventional commit
gcommit docs update setup guide

# Or with scope
gcommits docs setup update installation steps
```

### Push Your Work

```bash
gfeature-complete
# Runs tests, pushes branch, and reminds you to create PR
```

## Step 8: GitHub Setup (Optional)

### Set Up GitHub CLI

```bash
# Install GitHub CLI
brew install gh  # macOS
# or: https://cli.github.com/

# Authenticate
gh auth login

# Create PR from command line
gh pr create --base develop --title "docs: update setup guide"
```

### Set Up SSH Key for GitHub

```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your.email@example.com"

# Add to SSH agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519

# Copy public key
cat ~/.ssh/id_ed25519.pub

# Add to GitHub: https://github.com/settings/keys
```

## Step 9: Kubernetes Setup (for Deployments)

### Set Up kubeconfig

```bash
# Set KUBECONFIG environment variable
export KUBECONFIG=~/.kube/config

# Or add to ~/.zshrc
echo 'export KUBECONFIG=~/.kube/config' >> ~/.zshrc
```

### Verify Kubernetes Access

```bash
kubectl get nodes
kubectl get pods -n homepage
```

### Install Flux CLI

```bash
# macOS
brew install fluxcd/tap/flux

# Or: https://fluxcd.io/docs/installation/

# Verify
flux version
```

## Step 10: Test Deployment Workflow

### Local Build and Test

```bash
# Build images locally
make build-all

# Push to registry (requires permissions)
make build-push-version

# Trigger Flux reconciliation
greconcile
```

## Quick Reference

### Daily Commands

```bash
# Create feature
gfeature <description>

# Check status
ginfo

# Commit
gcommit <type> <message>

# Complete and push
gfeature-complete

# Check version
gversion
```

### Common Workflows

**Feature Development:**
```bash
gfeature add search
gcommit feat add search functionality
gfeature-complete
```

**Bug Fix:**
```bash
gbugfix api timeout
gcommit fix increase timeout to 30s
gfeature-complete
```

**Release:**
```bash
grelease-complete 1.1.0
# Update CHANGELOG.md
grelease-push
```

## Troubleshooting

### Docker Issues

```bash
# Clean up Docker
make clean
docker system prune -a

# Restart Docker Desktop (macOS)
```

### Port Already in Use

```bash
# Find and kill process
lsof -ti:8080 | xargs kill -9  # API port
lsof -ti:3000 | xargs kill -9  # Frontend port
```

### Permission Denied

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Fix Docker permissions
sudo usermod -aG docker $USER
# Log out and back in
```

### Tests Failing

```bash
# Clean and reinstall dependencies
make clean
cd api && go mod tidy && cd ..
cd frontend && rm -rf node_modules && npm install --legacy-peer-deps && cd ..
```

### Git Helper Functions Not Loading

```bash
# Check if sourced in shell config
cat ~/.zshrc | grep git-helpers

# Source manually
source ~/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/scripts/git-helpers.sh

# Verify
ghelp
```

## Next Steps

1. **Read Documentation**
   - [Quick Start Guide](./QUICK_START_VERSIONING.md)
   - [Branching Guide](./BRANCHING_GUIDE.md)
   - [Git Helpers Guide](./GIT_HELPERS_GUIDE.md)

2. **Explore Codebase**
   - API: `api/`
   - Frontend: `frontend/`
   - Helm Chart: `chart/`

3. **Join Development**
   - Create your first feature
   - Submit a pull request
   - Review others' code

## Useful Links

- **Repository**: https://github.com/brunovlucena/homelab
- **Documentation**: `docs/`
- **Issues**: https://github.com/brunovlucena/homelab/issues
- **PRs**: https://github.com/brunovlucena/homelab/pulls

## Support

- **Documentation**: Check `docs/` directory
- **Quick Help**: Run `ghelp` or `make help`
- **Issues**: Create GitHub issue
- **Email**: bruno@lucena.cloud

---

**You're all set! Start developing with `gfeature <description>` 🚀**

