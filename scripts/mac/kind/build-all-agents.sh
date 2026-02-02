#!/bin/bash
# Build all agent images after version bump
# Usage: ./scripts/build-all-agents.sh [registry] [push]

set -e

REGISTRY="${1:-ghcr.io/brunovlucena}"
PUSH="${2:-false}"

echo "🏗️  Building all agent images..."
echo "Registry: $REGISTRY"
echo "Push: $PUSH"
echo ""

# List of agents to build
AGENTS=(
    "agent-bruno"
    "agent-redteam"
    "agent-blueteam"
    "agent-contracts"
    "agent-tools"
    "agent-restaurant"
    "agent-pos-edge"
    "agent-chat"
    "agent-store-multibrands"
    "agent-rpg"
)

FAILED=()
SUCCESS=()

for agent in "${AGENTS[@]}"; do
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📦 Building $agent..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    AGENT_DIR="flux/ai/$agent"
    
    if [ ! -d "$AGENT_DIR" ]; then
        echo "⚠️  Directory not found: $AGENT_DIR"
        FAILED+=("$agent")
        continue
    fi
    
    # Check if Makefile exists
    if [ ! -f "$AGENT_DIR/Makefile" ]; then
        echo "⚠️  No Makefile found for $agent, skipping..."
        FAILED+=("$agent")
        continue
    fi
    
    # Read version
    VERSION_FILE="$AGENT_DIR/VERSION"
    if [ ! -f "$VERSION_FILE" ]; then
        echo "⚠️  No VERSION file found for $agent"
        FAILED+=("$agent")
        continue
    fi
    
    VERSION=$(cat "$VERSION_FILE" | tr -d 'v' | tr -d '\n')
    echo "Version: $VERSION"
    
    # Build using Makefile
    cd "$AGENT_DIR"
    
    if make build 2>&1; then
        echo "✅ $agent built successfully (v$VERSION)"
        SUCCESS+=("$agent")
        
        if [ "$PUSH" = "true" ]; then
            echo "📤 Pushing $agent..."
            if make push 2>&1; then
                echo "✅ $agent pushed successfully"
            else
                echo "❌ Failed to push $agent"
                FAILED+=("$agent (push failed)")
            fi
        fi
    else
        echo "❌ Failed to build $agent"
        FAILED+=("$agent")
    fi
    
    cd - > /dev/null
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Build Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Successful: ${#SUCCESS[@]}"
for agent in "${SUCCESS[@]}"; do
    echo "   - $agent"
done

if [ ${#FAILED[@]} -gt 0 ]; then
    echo ""
    echo "❌ Failed: ${#FAILED[@]}"
    for agent in "${FAILED[@]}"; do
        echo "   - $agent"
    done
    exit 1
fi

echo ""
echo "🎉 All agents built successfully!"
