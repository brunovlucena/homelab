#!/bin/bash
# Test script for agents-whatsapp-rust project structure

set -e

echo "🧪 Testing Agents WhatsApp Rust Project Structure"
echo "=================================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check project structure
echo "📁 Testing project structure..."
MISSING_FILES=0

check_file() {
    if [ ! -f "$1" ]; then
        echo -e "${RED}❌ Missing: $1${NC}"
        MISSING_FILES=$((MISSING_FILES + 1))
    else
        echo -e "${GREEN}✅ Found: $1${NC}"
    fi
}

check_file "Cargo.toml"
check_file "VERSION"
check_file "Makefile"
check_file "README.md"
check_file "DEPLOYMENT.md"
check_file "shared/Cargo.toml"
check_file "shared/src/lib.rs"
check_file "messaging-service/Cargo.toml"
check_file "messaging-service/Dockerfile"
check_file "user-service/Cargo.toml"
check_file "user-service/Dockerfile"
check_file "agent-gateway/Cargo.toml"
check_file "agent-gateway/Dockerfile"
check_file "message-storage-service/Cargo.toml"
check_file "message-storage-service/Dockerfile"
check_file "k8s/base/kustomization.yaml"
check_file "k8s/base/namespace.yaml"
check_file "k8s/base/messaging-service.yaml"
check_file "k8s/base/user-service.yaml"
check_file "k8s/base/agent-gateway.yaml"
check_file "k8s/base/message-storage-service.yaml"
check_file "k8s/base/broker-trigger.yaml"

if [ $MISSING_FILES -gt 0 ]; then
    echo -e "${RED}❌ Found $MISSING_FILES missing files${NC}"
    exit 1
fi

echo ""

# Test 2: Validate kustomize
echo "🔧 Testing kustomize manifests..."
if kubectl kustomize k8s/base > /dev/null 2>&1; then
    echo -e "${GREEN}✅ kustomize base validates${NC}"
else
    echo -e "${RED}❌ kustomize base validation failed${NC}"
    kubectl kustomize k8s/base 2>&1 | head -20
    exit 1
fi

if kubectl kustomize k8s/overlays/studio > /dev/null 2>&1; then
    echo -e "${GREEN}✅ kustomize studio overlay validates${NC}"
else
    echo -e "${RED}❌ kustomize studio overlay validation failed${NC}"
    kubectl kustomize k8s/overlays/studio 2>&1 | head -20
    exit 1
fi

echo ""

# Test 3: Validate Makefile
echo "🔧 Testing Makefile..."
if make help > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Makefile works${NC}"
else
    echo -e "${RED}❌ Makefile test failed${NC}"
    exit 1
fi

echo ""

# Test 4: Check version file
echo "🏷️  Testing version file..."
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION)
    if [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${GREEN}✅ Version format valid: $VERSION${NC}"
    else
        echo -e "${YELLOW}⚠️  Version format may be invalid: $VERSION${NC}"
    fi
else
    echo -e "${RED}❌ VERSION file not found${NC}"
    exit 1
fi

echo ""

# Test 5: Check Dockerfiles
echo "🐳 Testing Dockerfiles..."
for service in messaging-service user-service agent-gateway message-storage-service; do
    if [ -f "$service/Dockerfile" ]; then
        if grep -q "FROM rust" "$service/Dockerfile" && grep -q "cargo build" "$service/Dockerfile"; then
            echo -e "${GREEN}✅ $service/Dockerfile looks valid${NC}"
        else
            echo -e "${YELLOW}⚠️  $service/Dockerfile may have issues${NC}"
        fi
    fi
done

echo ""

# Test 6: Check Cargo.toml files
echo "📦 Testing Cargo.toml files..."
for toml in Cargo.toml shared/Cargo.toml messaging-service/Cargo.toml user-service/Cargo.toml agent-gateway/Cargo.toml message-storage-service/Cargo.toml; do
    if [ -f "$toml" ]; then
        if grep -q "\[package\]\|\[workspace\]" "$toml"; then
            echo -e "${GREEN}✅ $toml looks valid${NC}"
        else
            echo -e "${YELLOW}⚠️  $toml may have issues${NC}"
        fi
    fi
done

echo ""
echo -e "${GREEN}✅ All structure tests passed!${NC}"
echo ""
echo "📝 Next steps:"
echo "  1. Install Rust: https://rustup.rs/"
echo "  2. Test compilation: cargo check --workspace"
echo "  3. Build images: make build-images-local"
echo "  4. Deploy: make deploy"
