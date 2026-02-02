# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🌐 AI Agents - Web Command Centers Build Orchestrator
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Usage:
#   make -f Makefile.web help              # Show available commands
#   make -f Makefile.web docker-build-all  # Build all web command centers
#   make -f Makefile.web docker-build-local-all  # Build all for local registry
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Web Command Centers
WEB_DIRS := \
	agent-restaurant/web \
	agent-pos-edge/web-gas-station \
	agent-pos-edge/web-mcdonalds \
	agent-chat/web-command-center \
	agent-store-multibrands/web-command-center

# Configuration
REGISTRY ?= ghcr.io/brunovlucena
LOCAL_REGISTRY ?= localhost:5001

.PHONY: help docker-build-all docker-push-all docker-build-local-all dev-install-all build-all clean-all status

help: ## Show this help
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🌐 AI Agents - Web Command Centers Build Orchestrator"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Web Command Centers:"
	@for dir in $(WEB_DIRS); do echo "  • $$dir"; done
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Docker Build Commands
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

docker-build-all: ## Build all web command center Docker images
	@echo "🚀 Building all web command centers..."
	@for dir in $(WEB_DIRS); do \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "📦 Building $$dir..."; \
		$(MAKE) -C $$dir docker-build REGISTRY=$(REGISTRY) || exit 1; \
	done
	@echo ""
	@echo "✅ All web command centers built successfully!"

docker-push-all: ## Build and push all Docker images to registry
	@echo "🚀 Building and pushing all web command centers..."
	@for dir in $(WEB_DIRS); do \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "📤 Building and pushing $$dir..."; \
		$(MAKE) -C $$dir docker-push REGISTRY=$(REGISTRY) || exit 1; \
	done
	@echo ""
	@echo "✅ All web command centers pushed to $(REGISTRY)!"

docker-build-local-all: ## Build and push all to local registry
	@echo "🚀 Building all web command centers for local registry..."
	@for dir in $(WEB_DIRS); do \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "📦 Building $$dir for local registry..."; \
		$(MAKE) -C $$dir docker-build-local LOCAL_REGISTRY=$(LOCAL_REGISTRY) || exit 1; \
	done
	@echo ""
	@echo "✅ All web command centers pushed to $(LOCAL_REGISTRY)!"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Individual Builds
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

docker-build-restaurant: ## Build restaurant command center
	$(MAKE) -C agent-restaurant/web docker-build REGISTRY=$(REGISTRY)

docker-build-gas-station: ## Build gas station command center
	$(MAKE) -C agent-pos-edge/web-gas-station docker-build REGISTRY=$(REGISTRY)

docker-build-mcdonalds: ## Build McDonald's command center
	$(MAKE) -C agent-pos-edge/web-mcdonalds docker-build REGISTRY=$(REGISTRY)

docker-build-chat: ## Build AgentChat command center
	$(MAKE) -C agent-chat/web-command-center docker-build REGISTRY=$(REGISTRY)

docker-build-store: ## Build Multi-Brand Store command center
	$(MAKE) -C agent-store-multibrands/web-command-center docker-build REGISTRY=$(REGISTRY)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Development
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

dev-install-all: ## Install npm dependencies for all web apps
	@echo "📦 Installing dependencies for all web command centers..."
	@for dir in $(WEB_DIRS); do \
		echo "Installing dependencies in $$dir..."; \
		$(MAKE) -C $$dir install || exit 1; \
	done
	@echo "✅ All dependencies installed!"

build-all: ## Build all web apps (npm build)
	@echo "🔨 Building all web command centers..."
	@for dir in $(WEB_DIRS); do \
		echo "Building $$dir..."; \
		$(MAKE) -C $$dir build || exit 1; \
	done
	@echo "✅ All web apps built!"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Utilities
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

clean-all: ## Clean all build artifacts
	@echo "🧹 Cleaning all web command centers..."
	@for dir in $(WEB_DIRS); do \
		echo "Cleaning $$dir..."; \
		$(MAKE) -C $$dir clean || true; \
	done
	@echo "✅ All cleaned!"

status: ## Show build status summary
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🌐 Web Command Centers Status"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📁 Directories:"
	@for dir in $(WEB_DIRS); do \
		if [ -f "$$dir/Dockerfile" ] && [ -f "$$dir/Makefile" ]; then \
			echo "  ✅ $$dir (Dockerfile + Makefile)"; \
		elif [ -f "$$dir/Dockerfile" ]; then \
			echo "  ⚠️  $$dir (Dockerfile only)"; \
		elif [ -f "$$dir/Makefile" ]; then \
			echo "  ⚠️  $$dir (Makefile only)"; \
		else \
			echo "  ❌ $$dir (missing)"; \
		fi \
	done
	@echo ""
	@echo "📌 Registries:"
	@echo "  • GHCR: $(REGISTRY)"
	@echo "  • Local: $(LOCAL_REGISTRY)"
	@echo ""
	@echo "🔧 Standard Commands (available in each web app):"
	@echo "  • make dev               - Run local development server"
	@echo "  • make install           - Install npm dependencies"
	@echo "  • make build             - Build for production"
	@echo "  • make docker-build      - Build Docker image"
	@echo "  • make docker-push       - Push to registry"
	@echo "  • make docker-build-local - Build & push to local registry"
	@echo "  • make docker-run        - Run container locally"
	@echo "  • make clean             - Clean build artifacts"
