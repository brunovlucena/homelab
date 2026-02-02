#!/bin/bash
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Agent Makefile Validation Script
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Validates that Makefiles follow best practices:
# - Uses VERSION_FILE variable
# - Has version-bump target
# - Has release-patch/minor/major targets
#
# Usage: ./scripts/validate-makefile.sh <Makefile-path>
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set -euo pipefail

MAKEFILE="${1:-}"

if [ -z "$MAKEFILE" ] || [ ! -f "$MAKEFILE" ]; then
    echo "Error: Makefile not found: $MAKEFILE"
    exit 1
fi

ISSUES=0

# Check for VERSION_FILE variable
if ! grep -q "VERSION_FILE" "$MAKEFILE"; then
    echo "❌ Missing VERSION_FILE variable"
    ((ISSUES++))
fi

# Check for version-bump target
if ! grep -q "version-bump:" "$MAKEFILE"; then
    echo "❌ Missing version-bump target"
    ((ISSUES++))
fi

# Check for release targets
if ! grep -q "release-patch:" "$MAKEFILE"; then
    echo "⚠️  Missing release-patch target (recommended)"
fi

if ! grep -q "release-minor:" "$MAKEFILE"; then
    echo "⚠️  Missing release-minor target (recommended)"
fi

if ! grep -q "release-major:" "$MAKEFILE"; then
    echo "⚠️  Missing release-major target (recommended)"
fi

if [ $ISSUES -gt 0 ]; then
    echo ""
    echo "❌ Makefile validation failed. Please fix the issues above."
    exit 1
fi

echo "✅ Makefile validation passed"
exit 0
