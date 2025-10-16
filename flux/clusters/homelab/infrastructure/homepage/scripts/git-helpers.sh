#!/bin/bash
# Git Branch Helper Functions for Homepage Versioning
# Source this file in your ~/.zshrc or ~/.bashrc:
# source ~/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/scripts/git-helpers.sh

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ========================================
# Branch Creation Functions
# ========================================

# Create a 'feature' branch from develop
gfeature() {
    if [ $# -eq 0 ]; then
        echo "Usage: gfeature <description>"
        echo "Example: gfeature add user authentication"
        return 1
    fi
    git checkout develop && git pull origin develop
    local description=$(echo "$*" | tr ' ' '-')
    local branch_name="feature/$(date +%Y-%m-%d)/$description"
    git checkout -b "$branch_name"
    echo -e "${GREEN}✅ Created feature branch: $branch_name${NC}"
    echo -e "${BLUE}💡 When done, create PR to develop${NC}"
}

# Create a 'bugfix' branch from develop
gbugfix() {
    if [ $# -eq 0 ]; then
        echo "Usage: gbugfix <description>"
        echo "Example: gbugfix fix api timeout"
        return 1
    fi
    git checkout develop && git pull origin develop
    local description=$(echo "$*" | tr ' ' '-')
    local branch_name="bugfix/$(date +%Y-%m-%d)/$description"
    git checkout -b "$branch_name"
    echo -e "${GREEN}✅ Created bugfix branch: $branch_name${NC}"
    echo -e "${BLUE}💡 When done, create PR to develop${NC}"
}

# Create a 'hotfix' branch from main (production)
ghotfix() {
    if [ $# -eq 0 ]; then
        echo "Usage: ghotfix <description>"
        echo "Example: ghotfix security patch"
        return 1
    fi
    git checkout main && git pull origin main
    local description=$(echo "$*" | tr ' ' '-')
    local branch_name="hotfix/$(date +%Y-%m-%d)/$description"
    git checkout -b "$branch_name"
    echo -e "${GREEN}✅ Created hotfix branch: $branch_name${NC}"
    echo -e "${YELLOW}⚠️  IMPORTANT: Hotfix must be merged to BOTH main AND develop${NC}"
    echo -e "${BLUE}💡 After merge, run: git checkout develop && git merge main${NC}"
}

# Create a 'release' branch from develop
grelease() {
    if [ $# -eq 0 ]; then
        echo "Usage: grelease <version>"
        echo "Example: grelease 1.1.0"
        return 1
    fi
    local version=$1
    git checkout develop && git pull origin develop
    local branch_name="release/v$version"
    git checkout -b "$branch_name"
    echo -e "${GREEN}✅ Created release branch: $branch_name${NC}"
    echo -e "${BLUE}💡 Next steps:${NC}"
    echo "  1. cd flux/clusters/homelab/infrastructure/homepage"
    echo "  2. make version-bump VERSION=$version"
    echo "  3. Update CHANGELOG.md"
    echo "  4. git commit -m 'chore: bump version to v$version'"
    echo "  5. Create PR to main"
}

# Create a 'task' branch from develop
gtask() {
    if [ $# -eq 0 ]; then
        echo "Usage: gtask <description>"
        echo "Example: gtask refactor api handlers"
        return 1
    fi
    git checkout develop && git pull origin develop
    local description=$(echo "$*" | tr ' ' '-')
    local branch_name="task/$(date +%Y-%m-%d)/$description"
    git checkout -b "$branch_name"
    echo -e "${GREEN}✅ Created task branch: $branch_name${NC}"
    echo -e "${BLUE}💡 When done, create PR to develop${NC}"
}

# ========================================
# Quick Commit Functions
# ========================================

# Quick commit with conventional commit format
gcommit() {
    if [ $# -lt 2 ]; then
        echo "Usage: gcommit <type> <message>"
        echo ""
        echo "Types:"
        echo "  feat     - New feature"
        echo "  fix      - Bug fix"
        echo "  docs     - Documentation"
        echo "  style    - Code style"
        echo "  refactor - Code refactoring"
        echo "  perf     - Performance improvement"
        echo "  test     - Tests"
        echo "  chore    - Build/tooling"
        echo "  ci       - CI/CD changes"
        echo ""
        echo "Examples:"
        echo "  gcommit feat add user authentication"
        echo "  gcommit fix resolve api timeout"
        echo "  gcommit docs update readme"
        return 1
    fi
    
    local type=$1
    shift
    local message="$*"
    
    # Validate type
    case $type in
        feat|fix|docs|style|refactor|perf|test|chore|ci)
            git add .
            git commit -m "$type: $message"
            echo -e "${GREEN}✅ Committed: $type: $message${NC}"
            ;;
        *)
            echo -e "${YELLOW}⚠️  Invalid type: $type${NC}"
            echo "Valid types: feat, fix, docs, style, refactor, perf, test, chore, ci"
            return 1
            ;;
    esac
}

# Quick commit with scope
gcommits() {
    if [ $# -lt 3 ]; then
        echo "Usage: gcommits <type> <scope> <message>"
        echo ""
        echo "Examples:"
        echo "  gcommits feat api add user authentication endpoint"
        echo "  gcommits fix frontend resolve memory leak"
        echo "  gcommits docs readme update installation steps"
        return 1
    fi
    
    local type=$1
    local scope=$2
    shift 2
    local message="$*"
    
    git add .
    git commit -m "$type($scope): $message"
    echo -e "${GREEN}✅ Committed: $type($scope): $message${NC}"
}

# ========================================
# Version Management Functions
# ========================================

# Show current version
gversion() {
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/chart/Chart.yaml" ]; then
        cd "$homepage_dir"
        make version
        cd - > /dev/null
    else
        echo "Error: Chart.yaml not found"
        return 1
    fi
}

# Bump version
gbump() {
    if [ $# -eq 0 ]; then
        echo "Usage: gbump <version> or gbump <type>"
        echo ""
        echo "Version: gbump 1.1.0"
        echo "Type:    gbump major|minor|patch"
        echo ""
        echo "Examples:"
        echo "  gbump 1.1.0       - Set specific version"
        echo "  gbump major       - Bump major version (1.0.0 → 2.0.0)"
        echo "  gbump minor       - Bump minor version (1.0.0 → 1.1.0)"
        echo "  gbump patch       - Bump patch version (1.0.0 → 1.0.1)"
        return 1
    fi
    
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/chart/Chart.yaml" ]; then
        cd "$homepage_dir"
        
        case $1 in
            major|minor|patch)
                ./scripts/version-manager.sh bump-$1
                ;;
            *)
                make version-bump VERSION=$1
                ;;
        esac
        
        cd - > /dev/null
    else
        echo "Error: Chart.yaml not found"
        return 1
    fi
}

# Generate release notes
grelnotes() {
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/chart/Chart.yaml" ]; then
        cd "$homepage_dir"
        make release-notes
        cd - > /dev/null
    else
        echo "Error: Chart.yaml not found"
        return 1
    fi
}

# ========================================
# Deployment Functions
# ========================================

# Build and push with version
gbuild() {
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/Makefile" ]; then
        cd "$homepage_dir"
        make build-push-version
        cd - > /dev/null
    else
        echo "Error: Makefile not found"
        return 1
    fi
}

# Trigger Flux reconciliation
greconcile() {
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/Makefile" ]; then
        cd "$homepage_dir"
        make reconcile
        cd - > /dev/null
    else
        echo "Error: Makefile not found"
        return 1
    fi
}

# ========================================
# Workflow Helpers
# ========================================

# Complete feature workflow
gfeature-complete() {
    echo -e "${BLUE}🔍 Checking current branch...${NC}"
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    
    if [[ ! $current_branch =~ ^(feature|bugfix|task)/ ]]; then
        echo -e "${YELLOW}⚠️  Not on a feature/bugfix/task branch${NC}"
        return 1
    fi
    
    echo -e "${BLUE}📋 Current branch: $current_branch${NC}"
    echo -e "${BLUE}🔄 Pulling latest changes...${NC}"
    git pull origin develop
    
    echo -e "${BLUE}🧪 Running tests...${NC}"
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/Makefile" ]; then
        cd "$homepage_dir"
        make test || {
            echo -e "${YELLOW}⚠️  Tests failed. Fix before pushing.${NC}"
            cd - > /dev/null
            return 1
        }
        cd - > /dev/null
    fi
    
    echo -e "${BLUE}🚀 Pushing to remote...${NC}"
    git push origin "$current_branch"
    
    echo -e "${GREEN}✅ Branch pushed!${NC}"
    echo -e "${BLUE}💡 Next: Create PR to develop on GitHub${NC}"
}

# Complete release workflow
grelease-complete() {
    local version=$1
    if [ -z "$version" ]; then
        echo "Usage: grelease-complete <version>"
        echo "Example: grelease-complete 1.1.0"
        return 1
    fi
    
    echo -e "${BLUE}🔄 Creating release branch...${NC}"
    grelease "$version"
    
    echo -e "${BLUE}📝 Bumping version...${NC}"
    gbump "$version"
    
    echo -e "${YELLOW}⚠️  MANUAL STEP REQUIRED:${NC}"
    echo "  1. Update CHANGELOG.md with release notes"
    echo "  2. Review changes: git diff"
    echo "  3. When ready, run: grelease-push"
}

# Push release branch
grelease-push() {
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    
    if [[ ! $current_branch =~ ^release/ ]]; then
        echo -e "${YELLOW}⚠️  Not on a release branch${NC}"
        return 1
    fi
    
    echo -e "${BLUE}📝 Committing changes...${NC}"
    git add .
    git commit -m "chore: release ${current_branch#release/v}"
    
    echo -e "${BLUE}🚀 Pushing to remote...${NC}"
    git push origin "$current_branch"
    
    echo -e "${GREEN}✅ Release branch pushed!${NC}"
    echo -e "${BLUE}💡 Next: Create PR to main on GitHub${NC}"
}

# Complete hotfix workflow
ghotfix-complete() {
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    
    if [[ ! $current_branch =~ ^hotfix/ ]]; then
        echo -e "${YELLOW}⚠️  Not on a hotfix branch${NC}"
        return 1
    fi
    
    echo -e "${BLUE}🚀 Pushing hotfix...${NC}"
    git push origin "$current_branch"
    
    echo -e "${GREEN}✅ Hotfix pushed!${NC}"
    echo -e "${YELLOW}⚠️  CRITICAL: Create PR to main${NC}"
    echo -e "${YELLOW}⚠️  REMEMBER: After merge, back-merge to develop!${NC}"
}

# ========================================
# Status and Info
# ========================================

# Show branch info
ginfo() {
    echo -e "${BLUE}📊 Current Repository Status${NC}"
    echo ""
    echo -e "${BLUE}Branch:${NC} $(git rev-parse --abbrev-ref HEAD)"
    echo -e "${BLUE}Commit:${NC} $(git rev-parse --short HEAD)"
    echo ""
    
    # Show version if in homepage
    local homepage_dir="flux/clusters/homelab/infrastructure/homepage"
    if [ -f "$homepage_dir/chart/Chart.yaml" ]; then
        echo -e "${BLUE}Version:${NC}"
        cd "$homepage_dir"
        make version
        cd - > /dev/null
        echo ""
    fi
    
    echo -e "${BLUE}Status:${NC}"
    git status -sb
    echo ""
    
    echo -e "${BLUE}Recent commits:${NC}"
    git log --oneline -5
}

# Show available commands
ghelp() {
    cat << 'EOF'
🚀 Git Helper Functions for Homepage

Branch Creation:
  gfeature <desc>     - Create feature branch from develop
  gbugfix <desc>      - Create bugfix branch from develop
  ghotfix <desc>      - Create hotfix branch from main
  grelease <version>  - Create release branch from develop
  gtask <desc>        - Create task branch from develop

Commit:
  gcommit <type> <message>           - Conventional commit
  gcommits <type> <scope> <message>  - Commit with scope

Version Management:
  gversion            - Show current version
  gbump <version>     - Bump to specific version
  gbump major|minor|patch - Auto-bump version
  grelnotes           - Generate release notes

Build & Deploy:
  gbuild              - Build and push with version
  greconcile          - Trigger Flux reconciliation

Workflows:
  gfeature-complete   - Complete and push feature
  grelease-complete <v> - Create complete release
  grelease-push       - Push release branch
  ghotfix-complete    - Complete and push hotfix

Info:
  ginfo               - Show repo status
  ghelp               - Show this help

Examples:
  gfeature add user auth
  gcommit feat add user authentication
  grelease 1.1.0
  gbump minor
  gfeature-complete

EOF
}

# Print success message
echo -e "${GREEN}✅ Git helper functions loaded!${NC}"
echo -e "${BLUE}💡 Run 'ghelp' for available commands${NC}"

