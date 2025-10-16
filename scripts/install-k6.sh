#!/bin/bash
set -e

# 🎯 Install k6 Load Testing Tool
# Based on: https://k6.io/docs/getting-started/installation/

K6_VERSION=${1:-v0.48.0}
OS=${2:-linux}
ARCH=${3:-amd64}

echo "🎯 Installing k6 Load Testing Tool"
echo "📍 Version: ${K6_VERSION}"
echo "📍 OS: ${OS}"
echo "📍 Architecture: ${ARCH}"

# Check if k6 is already installed
if command -v k6 >/dev/null 2>&1; then
    CURRENT_VERSION=$(k6 version --short 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
    echo "✅ k6 is already installed (version: ${CURRENT_VERSION})"
    
    if [ "${CURRENT_VERSION}" = "${K6_VERSION}" ]; then
        echo "✅ k6 is already at the requested version, skipping installation..."
        exit 0
    else
        echo "🔄 Updating k6 from ${CURRENT_VERSION} to ${K6_VERSION}..."
    fi
fi

# Detect OS and architecture if not provided
if [ "${OS}" = "auto" ]; then
    case "$(uname -s)" in
        Linux*)     OS=linux;;
        Darwin*)    OS=macos;;
        CYGWIN*)    OS=windows;;
        MINGW*)     OS=windows;;
        *)          echo "❌ Unsupported operating system"; exit 1;;
    esac
fi

if [ "${ARCH}" = "auto" ]; then
    case "$(uname -m)" in
        x86_64)     ARCH=amd64;;
        arm64)      ARCH=arm64;;
        aarch64)    ARCH=arm64;;
        armv7l)     ARCH=armv7;;
        *)          echo "❌ Unsupported architecture"; exit 1;;
    esac
fi

# Download URL
DOWNLOAD_URL="https://github.com/grafana/k6/releases/download/${K6_VERSION}/k6-${K6_VERSION}-${OS}-${ARCH}.tar.gz"

echo "📥 Downloading k6 from: ${DOWNLOAD_URL}"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
cd "${TEMP_DIR}"

# Download and extract k6
if ! curl -L "${DOWNLOAD_URL}" | tar xz; then
    echo "❌ Failed to download k6"
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Find the k6 binary
K6_BINARY=$(find . -name "k6" -type f | head -1)

if [ -z "${K6_BINARY}" ]; then
    echo "❌ k6 binary not found in downloaded archive"
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Make it executable
chmod +x "${K6_BINARY}"

# Install to /usr/local/bin (requires sudo)
echo "🔧 Installing k6 to /usr/local/bin..."
if ! sudo mv "${K6_BINARY}" /usr/local/bin/; then
    echo "❌ Failed to install k6 to /usr/local/bin"
    echo "💡 You may need to run with sudo or install to a different location"
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Clean up
rm -rf "${TEMP_DIR}"

# Verify installation
echo "🔍 Verifying k6 installation..."
if k6 version; then
    echo "✅ k6 installation completed successfully!"
    echo ""
    echo "🎉 k6 is now installed and ready to use!"
    echo ""
    echo "📚 Quick start:"
    echo "  k6 run script.js"
    echo "  k6 run --vus 10 --duration 30s script.js"
    echo ""
    echo "📖 Documentation: https://k6.io/docs/"
else
    echo "❌ k6 installation verification failed"
    exit 1
fi
