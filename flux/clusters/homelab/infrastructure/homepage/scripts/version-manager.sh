#!/bin/bash
# Homepage Version Manager
# This script manages versioning across all components

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$ROOT_DIR/chart"
API_DIR="$ROOT_DIR/api"
FRONTEND_DIR="$ROOT_DIR/frontend"

# Files
CHART_FILE="$CHART_DIR/Chart.yaml"
VALUES_FILE="$CHART_DIR/values.yaml"
PACKAGE_JSON="$FRONTEND_DIR/package.json"
VERSION_FILE="$API_DIR/VERSION"
CHANGELOG_FILE="$ROOT_DIR/CHANGELOG.md"

# Functions
error() {
    echo -e "${RED}❌ ERROR: $1${NC}" >&2
    exit 1
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

get_current_version() {
    grep '^version:' "$CHART_FILE" | awk '{print $2}'
}

validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        error "Invalid version format: $version (expected: X.Y.Z or X.Y.Z-prerelease)"
    fi
}

update_chart_yaml() {
    local version=$1
    info "Updating $CHART_FILE..."
    
    # Update version
    sed -i.bak "s/^version:.*/version: $version/" "$CHART_FILE"
    
    # Update appVersion
    sed -i.bak "s/^appVersion:.*/appVersion: \"$version\"/" "$CHART_FILE"
    
    rm -f "$CHART_FILE.bak"
    success "Updated Chart.yaml"
}

update_values_yaml() {
    local version=$1
    info "Updating $VALUES_FILE..."
    
    # Update image tags (both API and frontend)
    sed -i.bak "s/tag: .*/tag: \"$version\"/" "$VALUES_FILE"
    
    rm -f "$VALUES_FILE.bak"
    success "Updated values.yaml"
}

update_frontend_version() {
    local version=$1
    info "Updating frontend package.json..."
    
    cd "$FRONTEND_DIR"
    npm version "$version" --no-git-tag-version --allow-same-version || true
    cd - > /dev/null
    
    success "Updated frontend package.json"
}

update_api_version() {
    local version=$1
    info "Creating API VERSION file..."
    
    echo "$version" > "$VERSION_FILE"
    
    success "Created API VERSION file"
}

update_changelog() {
    local version=$1
    local date=$(date +%Y-%m-%d)
    
    info "Updating CHANGELOG.md..."
    
    # Check if changelog exists
    if [ ! -f "$CHANGELOG_FILE" ]; then
        warning "CHANGELOG.md not found, skipping..."
        return
    fi
    
    # Add new version entry after [Unreleased]
    sed -i.bak "/## \[Unreleased\]/a\\
\\
## [$version] - $date\\
\\
### Added\\
- Version bump to $version\\
" "$CHANGELOG_FILE"
    
    rm -f "$CHANGELOG_FILE.bak"
    success "Updated CHANGELOG.md"
}

show_current_version() {
    local version=$(get_current_version)
    echo ""
    echo "📋 Current Version Information"
    echo "=============================="
    echo ""
    echo "Chart Version: $version"
    echo "Git Branch: $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'N/A')"
    echo "Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
    echo ""
}

bump_version() {
    local new_version=$1
    
    if [ -z "$new_version" ]; then
        error "Version is required"
    fi
    
    validate_version "$new_version"
    
    local current_version=$(get_current_version)
    
    echo ""
    info "Version Bump: $current_version → $new_version"
    echo ""
    
    # Update all version files
    update_chart_yaml "$new_version"
    update_values_yaml "$new_version"
    update_frontend_version "$new_version"
    update_api_version "$new_version"
    update_changelog "$new_version"
    
    echo ""
    success "Version bumped to $new_version successfully!"
    echo ""
    echo "📝 Next steps:"
    echo "  1. Review changes: git diff"
    echo "  2. Commit changes: git add . && git commit -m 'chore: bump version to $new_version'"
    echo "  3. Create tag: git tag -a homepage-v$new_version -m 'Release $new_version'"
    echo "  4. Push: git push origin main --tags"
    echo ""
}

auto_bump_version() {
    local bump_type=$1
    local current_version=$(get_current_version)
    
    # Remove any pre-release suffix
    current_version=$(echo "$current_version" | sed 's/-.*$//')
    
    # Split version into parts
    IFS='.' read -r -a parts <<< "$current_version"
    local major=${parts[0]}
    local minor=${parts[1]}
    local patch=${parts[2]}
    
    case $bump_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            error "Invalid bump type: $bump_type (expected: major, minor, or patch)"
            ;;
    esac
    
    local new_version="$major.$minor.$patch"
    bump_version "$new_version"
}

generate_release_notes() {
    local version=$(get_current_version)
    local previous_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    echo ""
    echo "📝 Release Notes for $version"
    echo "=============================="
    echo ""
    
    if [ -z "$previous_tag" ]; then
        echo "## Initial Release"
        echo ""
        echo "First stable release of Homepage application."
    else
        echo "## Changes since $previous_tag"
        echo ""
        git log "$previous_tag"..HEAD --pretty=format:"- %s (%h)" --no-merges
    fi
    
    echo ""
    echo ""
    echo "## Docker Images"
    echo ""
    echo "- \`ghcr.io/brunovlucena/homelab/homepage-api:$version\`"
    echo "- \`ghcr.io/brunovlucena/homelab/homepage-frontend:$version\`"
    echo ""
}

show_help() {
    cat << EOF
Homepage Version Manager

Usage: $0 <command> [options]

Commands:
    show                    Show current version information
    bump <version>          Bump to specific version (e.g., 1.0.0)
    bump-major              Bump major version (X.0.0)
    bump-minor              Bump minor version (x.Y.0)
    bump-patch              Bump patch version (x.y.Z)
    release-notes           Generate release notes
    help                    Show this help message

Examples:
    $0 show
    $0 bump 1.0.0
    $0 bump-major
    $0 bump-minor
    $0 bump-patch
    $0 release-notes

EOF
}

# Main
case "${1:-help}" in
    show)
        show_current_version
        ;;
    bump)
        bump_version "$2"
        ;;
    bump-major)
        auto_bump_version "major"
        ;;
    bump-minor)
        auto_bump_version "minor"
        ;;
    bump-patch)
        auto_bump_version "patch"
        ;;
    release-notes)
        generate_release_notes
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac

