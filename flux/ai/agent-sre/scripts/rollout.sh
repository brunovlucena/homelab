#!/bin/bash
# Agent-SRE Rollout Script
# Bumps version, commits, pushes, and builds on studio cluster

set -e

STUDIO_HOST="10.0.0.1"
STUDIO_USER="${STUDIO_USER:-bruno}"
AGENT_SRE_DIR="${AGENT_SRE_DIR:-~/workspace/bruno/repos/homelab/flux/ai/agent-sre}"

echo "🚀 Agent-SRE Rollout Script"
echo "=========================="
echo ""

# Step 1: Check git status
echo "📋 Checking git status..."
cd "$(dirname "$0")"
if [ -n "$(git status --porcelain)" ]; then
    echo "✅ Changes detected"
    git status --short
else
    echo "⚠️  No changes to commit"
    exit 0
fi

# Step 2: Stage all changes
echo ""
echo "📦 Staging all changes..."
git add -A

# Step 3: Get version
VERSION=$(cat VERSION | tr -d '[:space:]')
echo ""
echo "🏷️  Current version: v${VERSION}"

# Step 4: Commit
echo ""
echo "💾 Committing changes..."
git commit -m "feat: agent-sre v${VERSION} - Add all 4 training phases

- Phase 1: Function Calling with FunctionGemma
- Phase 2: Few-Shot Learning with example database
- Phase 3: RAG with vector similarity search
- Phase 4: Fine-tuning pipeline scripts
- Hybrid intelligent remediation selection
- Automatic success recording for continuous improvement"

# Step 5: Pull and Push
echo ""
echo "📥 Pulling latest changes..."
git pull --rebase origin main || git pull --rebase origin master

echo ""
echo "📤 Pushing to remote..."
git push origin main || git push origin master

echo ""
echo "✅ Changes committed and pushed!"
echo ""

# Step 6: SSH to studio and build
echo "🔨 Building on studio cluster (${STUDIO_HOST})..."
echo "   This will SSH to studio and run: make build-local"
echo ""

ssh ${STUDIO_USER}@${STUDIO_HOST} << EOF
    set -e
    echo "📂 Changing to agent-sre directory..."
    cd ${AGENT_SRE_DIR}
    
    echo "🔄 Pulling latest changes..."
    git pull origin main || git pull origin master
    
    echo "🐳 Building Docker image..."
    make build-local
    
    echo ""
    echo "✅ Build complete on studio!"
    echo "   Image: localhost:5001/agent-sre:v${VERSION}"
    echo "   Image: localhost:5001/agent-sre:latest"
EOF

echo ""
echo "🎉 Rollout complete!"
echo ""
echo "📋 Next steps:"
echo "   1. Flux will automatically reconcile and deploy the new version"
echo "   2. Monitor rollout: kubectl get lambdaagent agent-sre -n ai -w"
echo "   3. Check logs: kubectl logs -n ai -l app=agent-sre --tail=50 -f"

