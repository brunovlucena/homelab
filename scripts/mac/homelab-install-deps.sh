#!/bin/bash
# =============================================================================
# 📦 Homelab Install Dependencies
# =============================================================================
# This script updates Go dependencies and installs matching Pulumi plugins
#
# Usage: ./scripts/homelab-install-deps.sh
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PULUMI_DIR="${SCRIPT_DIR}/../../pulumi"

MODE="all"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --test-deps-only)
            MODE="test-only"
            shift
            ;;
        *)
            echo "❌ Unknown option: $1"
            echo "   Usage: $0 [--test-deps-only]"
            exit 1
            ;;
    esac
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [[ "${MODE}" == "test-only" ]]; then
    echo "📦 Homelab Test Dependency Installation"
else
    echo "📦 Homelab Install Dependencies"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

install_test_deps() {
    local step_label
    if [[ "${MODE}" == "test-only" ]]; then
        step_label="🧪 Step 1/1"
    else
        step_label="🧪 Step 7/7"
    fi
    echo "${step_label}: Installing test dependencies (BATS, etc.)..."
    if command -v brew >/dev/null 2>&1; then
        echo "   Installing BATS via Homebrew..."
        brew list bats-core >/dev/null 2>&1 || brew install bats-core
        brew tap kaos/shell 2>/dev/null || true
        brew list bats-assert >/dev/null 2>&1 || brew install bats-assert
        brew list bats-support >/dev/null 2>&1 || brew install bats-support
        echo "   ✅ BATS installed"
    else
        echo "   ⚠️  Homebrew not found. Please install BATS manually:"
        echo "      https://bats-core.readthedocs.io/en/stable/installation.html"
    fi
    echo "   ✅ Test dependencies ready"
    echo ""
}

if [[ "${MODE}" != "test-only" ]]; then
    # Step 1: Update Go dependencies (use direct proxy to avoid TLS issues)
    echo "📥 Step 1/6: Updating Go dependencies..."
    cd "${PULUMI_DIR}"
    export GOPROXY="direct"
    export GOSUMDB="off"
    go get -u ./...
    go mod tidy
    echo "   ✅ Go dependencies updated"
    echo ""

    # Step 2: Extract plugin versions from go.mod
    echo "🔍 Step 2/6: Detecting required plugin versions from go.mod..."
    KUBERNETES_VERSION=$(grep 'github.com/pulumi/pulumi-kubernetes/sdk' go.mod | awk '{print $2}' | sed 's/v4\.//')
    COMMAND_VERSION=$(grep 'github.com/pulumi/pulumi-command/sdk' go.mod | awk '{print $2}' | sed 's/v//')

    echo "   Required versions:"
    echo "   • kubernetes: v${KUBERNETES_VERSION}"
    echo "   • command: v${COMMAND_VERSION}"
    echo ""

    # Step 3: Install matching Pulumi plugins
    echo "🔌 Step 3/6: Installing Pulumi plugins..."
    pulumi plugin install resource kubernetes "v${KUBERNETES_VERSION}"
    pulumi plugin install resource command "v${COMMAND_VERSION}"
    echo "   ✅ Pulumi plugins installed"
    echo ""

    # Step 4: Install kubeseal
    echo "🔐 Step 4/6: Installing kubeseal..."
    if command -v kubeseal &> /dev/null; then
        echo "   ✅ kubeseal already installed ($(kubeseal --version 2>&1 | head -n1))"
    else
        if command -v brew &> /dev/null; then
            brew install kubeseal
            echo "   ✅ kubeseal installed"
        else
            echo "   ⚠️  Homebrew not found, please install kubeseal manually"
        fi
    fi
    echo ""

    # Step 5: Install telepresence
    echo "🌐 Step 5/6: Installing telepresence..."
    if command -v telepresence &> /dev/null; then
        echo "   ✅ telepresence already installed ($(telepresence version 2>&1 | head -n1))"
    else
        if command -v brew &> /dev/null; then
            brew install datawire/blackbird/telepresence
            echo "   ✅ telepresence installed"
        else
            echo "   ⚠️  Homebrew not found, please install telepresence manually"
        fi
    fi
    # Step 6: Install act CLI for local GitHub Actions testing
    echo ""
    echo "🤖 Step 6/6: Installing act CLI..."
    if command -v act &> /dev/null; then
        echo "   ✅ act already installed ($(act --version 2>&1 | head -n1))"
    else
        if command -v brew &> /dev/null; then
            brew install act
            echo "   ✅ act installed"
        else
            echo "   ⚠️  Homebrew not found, please install act manually:"
            echo "      https://github.com/nektos/act#installation"
        fi
    fi
    echo ""
fi

install_test_deps

if [[ "${MODE}" != "test-only" ]]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ Dependencies updated and plugins synced successfully!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📋 Installed plugins:"
    pulumi plugin ls | grep -E 'kubernetes|command'
    echo ""
    echo "🔐 Installed tools:"
    echo "   • kubeseal: $(kubeseal --version 2>&1 | head -n1 || echo 'not installed')"
    echo "   • telepresence: $(telepresence version 2>&1 | head -n1 || echo 'not installed')"
    echo "   • act: $(act --version 2>&1 | head -n1 || echo 'not installed')"
    echo ""
fi

echo "🧪 Test dependencies:"
if command -v brew >/dev/null 2>&1; then
    echo "   • bats-core: $(brew list --versions bats-core 2>/dev/null || echo 'not installed')"
    echo "   • bats-assert: $(brew list --versions bats-assert 2>/dev/null || echo 'not installed')"
    echo "   • bats-support: $(brew list --versions bats-support 2>/dev/null || echo 'not installed')"
else
    echo "   Homebrew not installed; test dependencies not detected"
fi

