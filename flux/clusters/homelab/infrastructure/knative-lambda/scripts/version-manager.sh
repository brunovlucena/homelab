#!/bin/bash
# Knative Lambda Version Manager
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
DEPLOY_DIR="$ROOT_DIR/deploy"
SIDECAR_DIR="$ROOT_DIR/sidecar"
METRICS_PUSHER_DIR="$ROOT_DIR/metrics-pusher"

# Files
CHART_FILE="$DEPLOY_DIR/Chart.yaml"
VALUES_FILE="$DEPLOY_DIR/values.yaml"
VERSION_FILE="$ROOT_DIR/VERSION"
SIDECAR_VERSION_FILE="$SIDECAR_DIR/VERSION"
METRICS_PUSHER_VERSION_FILE="$METRICS_PUSHER_DIR/VERSION"
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
    if [ -f "$CHART_FILE" ]; then
        grep '^version:' "$CHART_FILE" | awk '{print $2}'
    else
        echo "0.1.0"
    fi
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
    
    if [ ! -f "$CHART_FILE" ]; then
        error "Chart.yaml not found at $CHART_FILE"
    fi
    
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
    
    if [ ! -f "$VALUES_FILE" ]; then
        warning "values.yaml not found, skipping..."
        return
    fi
    
    # Update all image tags (builder, sidecar, metrics-pusher)
    # This assumes a structure like:
    # builder:
    #   image:
    #     tag: "x.y.z"
    sed -i.bak "s/tag: \".*\"/tag: \"$version\"/" "$VALUES_FILE"
    
    rm -f "$VALUES_FILE.bak"
    success "Updated values.yaml"
}

update_version_files() {
    local version=$1
    
    # Main service VERSION file
    info "Creating VERSION file for main service..."
    echo "$version" > "$VERSION_FILE"
    success "Created $VERSION_FILE"
    
    # Sidecar VERSION file
    if [ -d "$SIDECAR_DIR" ]; then
        info "Creating VERSION file for sidecar..."
        echo "$version" > "$SIDECAR_VERSION_FILE"
        success "Created $SIDECAR_VERSION_FILE"
    fi
    
    # Metrics Pusher VERSION file
    if [ -d "$METRICS_PUSHER_DIR" ]; then
        info "Creating VERSION file for metrics-pusher..."
        echo "$version" > "$METRICS_PUSHER_VERSION_FILE"
        success "Created $METRICS_PUSHER_VERSION_FILE"
    fi
}

update_changelog() {
    local version=$1
    local date=$(date +%Y-%m-%d)
    
    info "Updating CHANGELOG.md..."
    
    if [ ! -f "$CHANGELOG_FILE" ]; then
        warning "CHANGELOG.md not found, creating new one..."
        cat > "$CHANGELOG_FILE" << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [$version] - $date
### Added
- Initial release

EOF
        success "Created CHANGELOG.md"
        return
    fi
    
    # Add new version section after [Unreleased]
    if ! grep -q "\[Unreleased\]" "$CHANGELOG_FILE"; then
        error "CHANGELOG.md doesn't have [Unreleased] section"
    fi
    
    # Create temporary file with new version section
    awk -v version="$version" -v date="$date" '
        /\[Unreleased\]/ {
            print $0
            print ""
            print "## [" version "] - " date
            print "### Added"
            print "- "
            print ""
            print "### Changed"
            print "- "
            print ""
            print "### Fixed"
            print "- "
            print ""
            next
        }
        {print}
    ' "$CHANGELOG_FILE" > "$CHANGELOG_FILE.tmp"
    
    mv "$CHANGELOG_FILE.tmp" "$CHANGELOG_FILE"
    success "Updated CHANGELOG.md"
}

show_version_info() {
    local current_version=$(get_current_version)
    local git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local git_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  📊 Knative Lambda Version Information${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${GREEN}Chart Version:${NC}          $current_version"
    
    if [ -f "$VERSION_FILE" ]; then
        echo -e "${GREEN}Service Version:${NC}        $(cat $VERSION_FILE)"
    fi
    
    if [ -f "$SIDECAR_VERSION_FILE" ]; then
        echo -e "${GREEN}Sidecar Version:${NC}        $(cat $SIDECAR_VERSION_FILE)"
    fi
    
    if [ -f "$METRICS_PUSHER_VERSION_FILE" ]; then
        echo -e "${GREEN}Metrics Pusher Version:${NC} $(cat $METRICS_PUSHER_VERSION_FILE)"
    fi
    
    echo -e "${GREEN}Git Commit:${NC}             $git_commit"
    echo -e "${GREEN}Git Branch:${NC}             $git_branch"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
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
    update_version_files "$new_version"
    update_changelog "$new_version"
    
    echo ""
    success "Version bumped to $new_version successfully!"
    echo ""
    echo "📝 Next steps:"
    echo "  1. Review changes: git diff"
    echo "  2. Commit changes: git add . && git commit -m 'chore: bump version to $new_version'"
    echo "  3. Create tag: git tag -a knative-lambda-v$new_version -m 'Release $new_version'"
    echo "  4. Push: git push origin <branch> --tags"
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
    local current_version=$(get_current_version)
    local git_tag="knative-lambda-v$current_version"
    
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  📝 Release Notes for v$current_version${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    # Get the previous tag
    local prev_tag=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
    
    if [ -n "$prev_tag" ]; then
        echo -e "${GREEN}Changes since $prev_tag:${NC}"
        echo ""
        
        # Get commit messages
        git log $prev_tag..HEAD --pretty=format:"- %s" --no-merges
    else
        echo -e "${GREEN}All commits:${NC}"
        echo ""
        git log --pretty=format:"- %s" --no-merges
    fi
    
    echo ""
    echo ""
    echo -e "${GREEN}Docker Images:${NC}"
    echo "- 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-knative-lambda-builder:$current_version"
    echo "- 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-knative-lambda-sidecar:$current_version"
    echo "- 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-knative-lambda-metrics-pusher:$current_version"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

show_help() {
    cat << EOF

🏗️  Knative Lambda Version Manager

USAGE:
    $0 <command> [arguments]

COMMANDS:
    bump <version>          Bump version to specified version (e.g., 1.0.0)
    bump-major              Auto-bump major version (1.0.0 → 2.0.0)
    bump-minor              Auto-bump minor version (1.0.0 → 1.1.0)
    bump-patch              Auto-bump patch version (1.0.0 → 1.0.1)
    info                    Show current version information
    release-notes           Generate release notes from git commits
    help                    Show this help message

EXAMPLES:
    # Bump to specific version
    $0 bump 1.2.0

    # Auto-bump patch version
    $0 bump-patch

    # Show version info
    $0 info

    # Generate release notes
    $0 release-notes

EOF
}

# Main script
case "${1:-}" in
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
    info)
        show_version_info
        ;;
    release-notes)
        generate_release_notes
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Unknown command: ${1:-}. Use 'help' for usage information."
        ;;
esac

