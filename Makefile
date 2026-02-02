.PHONY: help bootstrap

# Set help as the default target
.DEFAULT_GOAL := help

# =============================================================================
# Environment Configuration
# =============================================================================
# Auto-discover clusters from flux/clusters/ directory
CLUSTERS_DIR := flux/clusters
CLUSTERS := $(notdir $(wildcard $(CLUSTERS_DIR)/*))

# Detect host OS to pick the right scripts directory (default: mac)
UNAME_S := $(shell uname -s 2>/dev/null || echo "")
SCRIPT_OS ?=
ifeq ($(strip $(SCRIPT_OS)),)
  ifeq ($(UNAME_S),Darwin)
    SCRIPT_OS := mac
  else ifeq ($(UNAME_S),Linux)
    SCRIPT_OS := linux
  else
    SCRIPT_OS := mac
  endif
endif
SCRIPTS_ROOT := scripts/$(SCRIPT_OS)
KIND_SCRIPTS := $(SCRIPTS_ROOT)/kind

# Auto-detect environment from kubectl context if not explicitly set
KUBECTL_CONTEXT := $(shell kubectl config current-context 2>/dev/null || echo "")
DETECTED_ENV := $(shell echo "$(KUBECTL_CONTEXT)" | sed -n 's/^kind-\(.*\)/\1/p')

# Set ENV with proper defaults (only if not set by user)
ifeq ($(origin ENV),undefined)
  # ENV not set by user, try to detect or use default
  ifeq ($(filter $(DETECTED_ENV),$(CLUSTERS)),$(DETECTED_ENV))
    # Detected env is valid
    ENV := $(DETECTED_ENV)
  else
    # No valid detected env, use studio as default
    ENV := studio
  endif
endif

# Only validate ENV if it's being used (will be checked in targets)
# Default to bvlucena if not set
ifeq ($(strip $(PULUMI_ORG)),)
  PULUMI_ORG := bvlucena
endif

# Machine detection for multi-client/multi-machine support
# Extract hostname and normalize it (remove .local, lowercase, sanitize)
MACHINE_HOSTNAME := $(shell hostname | tr '[:upper:]' '[:lower:]' | sed 's/\.local$$//' | sed 's/[^a-z0-9-]/-/g')
# Detect machine type from hostname (studio, pro, or use hostname as fallback)
ifeq ($(shell echo "$(MACHINE_HOSTNAME)" | grep -q "studio" && echo "studio"),studio)
  MACHINE_ID := studio
else ifeq ($(shell echo "$(MACHINE_HOSTNAME)" | grep -q "pro" && echo "pro"),pro)
  MACHINE_ID := pro
else
  # Use sanitized hostname as machine ID for other machines/clients
  MACHINE_ID := $(MACHINE_HOSTNAME)
endif

# Stack naming: Include machine ID to prevent cross-machine conflicts
# Format: org/project/env-machine (e.g., bvlucena/homelab/studio-studio, bvlucena/homelab/pro-pro, bvlucena/homelab/air-air)
# This ensures each machine has its own stack even for the same ENV
PULUMI_PROJECT := homelab
PULUMI_STACK := $(PULUMI_ORG)/$(PULUMI_PROJECT)/$(ENV)-$(MACHINE_ID)
CLUSTER_NAME := $(ENV)

# Validation function to be called by targets
define validate_env
	@if [ -z "$(ENV)" ] || ! echo "$(CLUSTERS)" | grep -qw "$(ENV)"; then \
		echo "❌ Error: Invalid ENV=$(ENV). Must be one of: $(CLUSTERS)"; \
		exit 1; \
	fi
endef

# Helper to get context names (DRY) - using recursive expansion for runtime evaluation
KIND_CONTEXT_NAME = kind-$(CLUSTER_NAME)
HOMELAB_CONTEXT_NAME = homelab-$(CLUSTER_NAME)

define ensure_cluster_context
	$(call validate_env)
	@kubectl config use-context $(KIND_CONTEXT_NAME) >/dev/null 2>&1 || (echo "❌ Failed to switch to context $(KIND_CONTEXT_NAME)" && exit 1)
endef

define ensure_homelab_context
	@if kubectl config get-contexts $(HOMELAB_CONTEXT_NAME) >/dev/null 2>&1; then \
		echo "   • Homelab context $(HOMELAB_CONTEXT_NAME) already exists"; \
	else \
		if ! kubectl config get-contexts $(KIND_CONTEXT_NAME) >/dev/null 2>&1; then \
			echo "❌ Base context $(KIND_CONTEXT_NAME) not found in kubeconfig"; \
			echo "   Run 'make bootstrap ENV=$(ENV)' to create the Kind substrate first."; \
			exit 1; \
		fi; \
		kubectl config set-context $(HOMELAB_CONTEXT_NAME) --cluster=$(KIND_CONTEXT_NAME) --user=$(KIND_CONTEXT_NAME) >/dev/null; \
		echo "   • Created homelab context $(HOMELAB_CONTEXT_NAME)"; \
	fi
endef

.PHONY: ensure-context
ensure-context: ## 🔐 Ensure homelab-* kube context alias exists for current ENV
	$(call validate_env)
	@echo "🔐 Ensuring kube context alias for $(CLUSTER_NAME)..."
	$(call ensure_homelab_context)
	@echo "   ✓ Context ready"

help: ## 📖 Show this help message
	@echo ''
	@echo '╔═══════════════════════════════════════════════════════════════════╗'
	@echo '║               🏗️  Homelab Infrastructure Manager                   ║'
	@echo '╚═══════════════════════════════════════════════════════════════════╝'
	@echo ''
	@echo '🎯 Current Environment: $(ENV) $(if $(DETECTED_ENV),(auto-detected from kubectl context: $(KUBECTL_CONTEXT)),(default))'
	@echo '🌍 Available Clusters: $(CLUSTERS)'
	@echo ''
	@echo '💡 Environment auto-detects from kubectl context or use ENV=<cluster>'
	@echo '   Examples: make up            # uses auto-detected/default'
	@echo '             make up ENV=pro     # explicit override'
	@echo '             kubectl config use-context kind-pro && make up  # auto-detects pro'
	@echo ''
	@echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
	@echo '🔧 Infrastructure Management'
	@echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''

env: ## 🎯 Show current environment and context information
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🎯 Environment Information"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Current Environment:  $(ENV)"
	@echo "Pulumi Stack:         $(PULUMI_STACK)"
	@echo "Cluster Name:         $(CLUSTER_NAME)"
	@echo "Machine ID:          $(MACHINE_ID)"
	@echo "Hostname:            $(MACHINE_HOSTNAME)"
	@echo "Kubectl Context:      $(KUBECTL_CONTEXT)"
	@echo "Detected from:        $(if $(DETECTED_ENV),kubectl context ($(DETECTED_ENV)),default fallback)"
	@echo ""
	@echo "Available Clusters:   $(CLUSTERS)"
	@echo ""
	@echo "Local Kind Clusters:"
	@kind get clusters 2>/dev/null | sed 's/^/   • /' || echo "   (none)"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "💡 Tips:"
	@echo "   • Switch context: kubectl config use-context kind-<cluster>"
	@echo "   • Override: make <command> ENV=<cluster>"
	@echo "   • List contexts: kubectl config get-contexts"
	@echo "   • Each machine has its own Pulumi stack (machine-specific)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""

init-homelab-deps: ## 🚀 Initialize complete environment (deps + registry + bootstrap)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🚀 Initializing Complete Environment"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(MAKE) init-deps
	@$(MAKE) init-registry

init-deps: ## 🔄 Initialize all dependencies
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔄 Initializing Dependencies"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./$(SCRIPTS_ROOT)/homelab-install-deps.sh

init-registry: ## 📦 Initialize local registry (shared across ALL clusters)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📦 Initializing Registry & Pre-warming Images (shared across ALL clusters)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./$(KIND_SCRIPTS)/setup-local-registry.sh
	@$(MAKE) prewarm-images

# =============================================================================
# Build Optimization
# =============================================================================
PULUMI_BIN := pulumi/.pulumi-bin
GO_CACHE := $(HOME)/.cache/go-build

build-pulumi: ## 🔨 Pre-compile Pulumi program (speeds up deployment ~30s)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔨 Pre-compiling Pulumi program"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd pulumi && GOCACHE="$(GO_CACHE)" go build -buildvcs=false -o .pulumi-bin .
	@echo "   ✓ Binary compiled to pulumi/.pulumi-bin"
	@echo "   ✓ Go build cache at $(GO_CACHE)"

install-pulumi-plugins: ## 🔌 Pre-install Pulumi plugins (speeds up compilation)
	@echo "🔌 Installing Pulumi plugins..."
	@cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi plugin install --yes
	@echo "   ✓ Plugins installed"

init-pulumi: ## 📦 Initialize Pulumi stack (usage: make init ENV=studio|pro|air)
	@echo "📦 Initializing Pulumi stack: homelab"
	
	@cd pulumi && \
		if [ -f ~/.zshrc ]; then . ~/.zshrc 2>/dev/null || true; \
		elif [ -f ~/.bashrc ]; then . ~/.bashrc 2>/dev/null || true; fi && \
		if ! PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive 2>/dev/null; then \
			PULUMI_CONFIG_PASSPHRASE="" pulumi stack init --stack $(PULUMI_STACK); \
		fi && \
		PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive 2>/dev/null && \
		echo "📦 Creating/updating Pulumi secrets..." && \
		CONFIGS="githubToken:GITHUB_TOKEN \
		homepagePostgresPassword:HOMEPAGE_POSTGRES_PASSWORD \
		homepageRedisPassword:HOMEPAGE_REDIS_PASSWORD \
		grafanaAdmin:GRAFANA_ADMIN \
		grafanaPassword:GRAFANA_PASSWORD \
		grafanaApiKey:GRAFANA_API_KEY \
		pagerdutyUrl:PAGERDUTY_URL \
		pagerdutyServiceKey:PAGERDUTY_SERVICE_KEY \
		logfireToken:LOGFIRE_TOKEN \
		slackWebhookUrl:SLACK_WEBHOOK_URL \
		cloudflareEmail:CLOUDFLARE_EMAIL \
		cloudflareApiKey:CLOUDFLARE_API_KEY \
		cloudflareApiToken:CLOUDFLARE_API_TOKEN \
		cloudflareWarpToken:CLOUDFLARE_WARP_TOKEN \
		cloudflareTunnelToken:CLOUDFLARE_TUNNEL_TOKEN \
		twingateApiToken:TWINGATE_API_TOKEN \
		twingateNetwork:TWINGATE_NETWORK \
		twingateProAccessToken:TWINGATE_PRO_ACCESS_TOKEN \
		twingateProRefreshToken:TWINGATE_PRO_REFRESH_TOKEN \
		twingateStudioAccessToken:TWINGATE_STUDIO_ACCESS_TOKEN \
		twingateStudioRefreshToken:TWINGATE_STUDIO_REFRESH_TOKEN \
		pulumiAccessToken:PULUMI_ACCESS_TOKEN \
		linearApiKey:LINEAR_API_KEY \
		jiraUrl:JIRA_URL \
		jiraEmail:JIRA_EMAIL \
		jiraApiToken:JIRA_API_TOKEN"; \
		TEMP_SECRET=$$(mktemp); \
		trap "rm -f $$TEMP_SECRET" EXIT; \
		for config in $$CONFIGS; do \
			KEY=$$(echo $$config | cut -d: -f1); \
			ENV_VAR=$$(echo $$config | cut -d: -f2); \
			VALUE=$$(eval echo \$$$$ENV_VAR); \
			if [ -z "$$VALUE" ]; then \
				echo "   ⚠️  Warning: $$ENV_VAR is empty, skipping $$KEY" >&2; \
			fi; \
		done; \
		for config in $$CONFIGS; do \
			KEY=$$(echo $$config | cut -d: -f1); \
			ENV_VAR=$$(echo $$config | cut -d: -f2); \
			VALUE=$$(eval echo \$$$$ENV_VAR); \
			if [ -n "$$VALUE" ]; then \
				printf '%s' "$$VALUE" > $$TEMP_SECRET; \
				pulumi config set $$KEY --secret < $$TEMP_SECRET >/dev/null 2>&1; \
			fi; \
		done; \
		PROVIDER_CONFIGS="twingate:apiToken:TWINGATE_API_TOKEN \
		twingate:network:TWINGATE_NETWORK"; \
		for config in $$PROVIDER_CONFIGS; do \
			KEY=$$(echo $$config | cut -d: -f1-2); \
			ENV_VAR=$$(echo $$config | cut -d: -f3); \
			VALUE=$$(eval echo \$$$$ENV_VAR); \
			if [ -n "$$VALUE" ]; then \
				printf '%s' "$$VALUE" > $$TEMP_SECRET; \
				if [ "$$KEY" = "twingate:apiToken" ]; then \
					pulumi config set $$KEY --secret < $$TEMP_SECRET >/dev/null 2>&1; \
				else \
					pulumi config set $$KEY < $$TEMP_SECRET >/dev/null 2>&1; \
				fi; \
			fi; \
		done && echo "   ✓ All secrets set successfully" >&2 || echo "   ❌ Error: Some secrets failed to set" >&2; \
		echo "   ✓ Pulumi secrets created/updated"

select: ## 📦 Select Pulumi stack (usage: make select ENV=studio|pro|air)
	@echo "📦 Selecting Pulumi stack: $(PULUMI_STACK)"
	cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK)

list-stacks: ## 📋 List all Pulumi stacks (across all machines/clients)

mac-sync: ## 🔄 Sync kubeconfigs and configs between all Macs (studio, pro, air)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔄 Syncing kubeconfigs and configs between Macs"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔌 Installing Pulumi plugins and Go dependencies..."
	@cd pulumi/mac-sync && PULUMI_CONFIG_PASSPHRASE="" pulumi plugin install --yes >/dev/null 2>&1 || echo "   ℹ️  Plugin installation skipped"
	@cd pulumi/mac-sync && GOCACHE="$(GO_CACHE)" go mod download >/dev/null 2>&1 && echo "   ✓ Go dependencies cached" || echo "   ℹ️  Go dependency download skipped"
	@echo "🔨 Building Pulumi program..."
	@cd pulumi/mac-sync && GOCACHE="$(GO_CACHE)" go build -buildvcs=false -o .pulumi-bin . || (echo "❌ Failed to build Pulumi program" && exit 1)
	@echo "   ✓ Binary compiled to pulumi/mac-sync/.pulumi-bin"
	@echo ""
	@cd pulumi/mac-sync && \
		MAC_SYNC_STACK="$(PULUMI_ORG)/mac-sync" && \
		if ! PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $$MAC_SYNC_STACK &>/dev/null; then \
			echo "📦 Stack '$$MAC_SYNC_STACK' not found, initializing..."; \
			PULUMI_CONFIG_PASSPHRASE="" pulumi stack init $$MAC_SYNC_STACK; \
		fi && \
		PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $$MAC_SYNC_STACK && \
		PULUMI_CONFIG_PASSPHRASE="" pulumi up --yes
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📋 All Pulumi Stacks (Multi-Machine/Client Support)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Current Machine: $(MACHINE_HOSTNAME) ($(MACHINE_ID))"
	@echo ""
	@echo "All Stacks:"
	@cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi stack ls --json 2>/dev/null | \
		python3 -c "import sys, json; \
		stacks = json.load(sys.stdin); \
		print('   Stack Name'.ljust(40) + 'Last Update'.ljust(20) + 'Resources'); \
		print('   ' + '-'*38 + ' ' + '-'*18 + ' ' + '-'*10); \
		[print(f\"   {s['name']:<38} {s.get('lastUpdate', 'N/A'):<18} {s.get('resourceCount', 0)}\") for s in stacks]" || \
		cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi stack ls
	@echo ""
	@echo "💡 Stack naming format: env-machine"
	@echo "   Example: studio-studio (studio env on studio machine)"
	@echo "           studio-pro    (studio env on pro machine)"
	@echo "           air-air       (air env on air machine)"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Dynamically generate targets for each cluster
# Creates: up-pro, cancel-pro, destroy-pro, etc.
# And shorthand: pro → up-pro, studio → up-studio, etc.
# Note: bootstrap is done internally by 'up', so no separate bootstrap-$(cluster) target
define CLUSTER_TARGETS
.PHONY: up-$(1) cancel-$(1) destroy-$(1) $(1)

up-$(1): ## 🚀 Deploy $(1) cluster (usage: make up-$(1))
	@$$(MAKE) up ENV=$(1)

cancel-$(1): ## ⏸️  Cancel $(1) operation (usage: make cancel-$(1))
	@$$(MAKE) cancel ENV=$(1)

destroy-$(1): ## 💥 Destroy $(1) cluster (usage: make destroy-$(1))
	@$$(MAKE) destroy ENV=$(1)

$(1): ## 📦 Shorthand to deploy $(1) (alias for: make up-$(1))
	@$$(MAKE) up-$(1)
endef
bootstrap: ## 🏗️  Bootstrap Kind substrate (registry + cluster only)
	$(call validate_env)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🏗️  Bootstrapping Kind substrate for $(ENV)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   Cluster:      $(CLUSTER_NAME)"
	@echo "   Kind context: kind-$(CLUSTER_NAME)"
	@echo "   Kind config:  flux/clusters/$(CLUSTER_NAME)/kind.yaml"
	@echo ""
	@SCRIPT_PATH="./$(KIND_SCRIPTS)/bootstrap-cluster.sh"; \
	CONFIG_PATH="flux/clusters/$(CLUSTER_NAME)/kind.yaml"; \
	KIND_CONTEXT="kind-$(CLUSTER_NAME)"; \
	if [ ! -x "$$SCRIPT_PATH" ]; then \
		echo "❌ Bootstrap script not found at $$SCRIPT_PATH"; \
		exit 1; \
	fi; \
	if [ ! -f "$$CONFIG_PATH" ]; then \
		echo "❌ Kind config not found at $$CONFIG_PATH"; \
		exit 1; \
	fi; \
	NEEDS_BOOTSTRAP=1; \
	if [ -z "$$FORCE_BOOTSTRAP" ] && kubectl config get-contexts "$$KIND_CONTEXT" >/dev/null 2>&1 && \
		kubectl get nodes --context "$$KIND_CONTEXT" --no-headers >/dev/null 2>&1; then \
		echo "ℹ️  Kind substrate already present (context $$KIND_CONTEXT)."; \
		echo "    Use FORCE_BOOTSTRAP=1 make bootstrap ENV=$(ENV) to re-run anyway."; \
		NEEDS_BOOTSTRAP=0; \
	fi; \
	if [ "$$NEEDS_BOOTSTRAP" -eq 1 ]; then \
		"$$SCRIPT_PATH" "$(CLUSTER_NAME)" "$$CONFIG_PATH" "$$KIND_CONTEXT"; \
	fi
	@$(MAKE) ensure-context ENV=$(ENV)
	@echo ""
	@echo "✅ Kind substrate ready. Next: run 'make up ENV=$(ENV)' to apply Pulumi."
	@echo ""

# Generate targets for each discovered cluster
$(foreach cluster,$(CLUSTERS),$(eval $(call CLUSTER_TARGETS,$(cluster))))

up: ## 🚀 Deploy cluster (usage: make up ENV=studio|pro|air or make studio|pro|air)
	$(call validate_env)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🚀 Deploying $(ENV) cluster"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   Stack:        $(PULUMI_STACK)"
	@echo "   Cluster:      $(CLUSTER_NAME)"
	@echo "   Machine ID:   $(MACHINE_ID)"
	@echo "   Hostname:     $(MACHINE_HOSTNAME)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "ℹ️  Note: This will create/update a machine-specific Pulumi stack"
	@echo "   Each machine (client) has its own stack to prevent conflicts."
	@echo ""
	@echo "📦 Step 1/5: Initializing Pulumi stack..."
	@$(MAKE) init-pulumi ENV=$(ENV)
	@echo ""
	@echo "🏗️  Step 2/5: Bootstrapping Kind Cluster..."
	@$(MAKE) bootstrap ENV=$(ENV)
	@echo ""
	@echo "🛑 Step 3/5: Canceling any pending Pulumi operations..."
	@cd pulumi && \
		if PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) 2>/dev/null; then \
			PULUMI_CONFIG_PASSPHRASE="" pulumi cancel --yes 2>/dev/null || true; \
		else \
			echo "   ℹ️  Stack '$(PULUMI_STACK)' not found yet, will be created in next step"; \
		fi
	@echo ""
	@echo "🔌 Step 3.5/5: Ensuring Pulumi plugins and Go dependencies..."
	@cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi plugin install --yes >/dev/null 2>&1 || echo "   ℹ️  Plugin installation skipped or already complete"
	@cd pulumi && GOCACHE="$(GO_CACHE)" go mod download >/dev/null 2>&1 && echo "   ✓ Go dependencies cached" || echo "   ℹ️  Go dependency download skipped"
	@echo ""
	@echo "🔨 Step 4/5: Ensuring Pulumi binary is up-to-date..."
	@if [ ! -f pulumi/.pulumi-bin ] || [ pulumi/main.go -nt pulumi/.pulumi-bin ]; then \
		echo "   Rebuilding binary (main.go changed or binary missing)..."; \
		cd pulumi && GOCACHE="$(GO_CACHE)" go build -buildvcs=false -o .pulumi-bin main.go; \
	else \
		echo "   ✓ Binary is up-to-date, skipping compilation"; \
	fi
	@echo ""
	@echo "🚀 Step 5/6: Deploying with Pulumi..."
	@cd pulumi && \
		if ! PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive 2>/dev/null; then \
			echo "   📦 Stack '$(PULUMI_STACK)' not found, creating it..."; \
			PULUMI_CONFIG_PASSPHRASE="" pulumi stack init --stack $(PULUMI_STACK); \
			PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive; \
		fi
	@cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi up --yes
	@echo ""
	@echo "⏳ Step 6/6: Waiting for cluster to be ready..."
	@kubectl config use-context kind-$(CLUSTER_NAME) || (echo "❌ Failed to switch to cluster context" && exit 1)
	@echo "   ✓ Switched to context: kind-$(CLUSTER_NAME)"
	@echo ""
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ $(ENV) cluster deployment complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📊 Next steps:"
	@echo "   • Watch Flux: make observe"
	@echo "   • Check status: kubectl get nodes,pods -A"
	@echo "   • View resources: flux get all -A"
	@echo ""
	@echo "🔗 Useful commands:"
	@echo "   make reconcile-all       # Reconcile all Flux resources"
	@echo "   make watch-flux          # Continuous monitoring"
	@echo "   make fast-refresh        # Quick update without recreating cluster"
	@echo "   make destroy-$(ENV)      # Destroy this cluster"
	@echo ""
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔒 Twingate Connector"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🐳 Ensuring Twingate connector Docker container is running..."
	@$(SCRIPTS_ROOT)/twingate-connector.sh "$(CLUSTER_NAME)" || echo "   ⚠️  Failed to start container (check script output above)"
	@echo ""

# =============================================================================
# Fast Operations (Optimized for Speed)
# =============================================================================
fast-refresh: ## ⚡ Fast refresh: update cluster without recreating (saves ~5-8min)
	$(call validate_env)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "⚡ Fast Refresh for $(ENV) cluster"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if kind get clusters 2>/dev/null | grep -qw "$(CLUSTER_NAME)"; then \
		echo "♻️  Cluster exists, applying changes only..."; \
		echo ""; \
		kubectl config use-context kind-$(CLUSTER_NAME) >/dev/null 2>&1 || exit 1; \
		echo "📦 Step 1/3: Refreshing Pulumi state..."; \
		cd pulumi && \
			if ! PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive 2>/dev/null; then \
				echo "   📦 Stack '$(PULUMI_STACK)' not found, creating it..."; \
				PULUMI_CONFIG_PASSPHRASE="" pulumi stack init --stack $(PULUMI_STACK); \
				PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive; \
			fi; \
		cd pulumi && PULUMI_CONFIG_PASSPHRASE="" GOCACHE="$(GO_CACHE)" GOFLAGS="-buildvcs=false" pulumi refresh --yes 2>/dev/null || true; \
		echo ""; \
		echo "🚀 Step 2/3: Applying changes with Pulumi..."; \
		cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi up --yes; \
		echo ""; \
		echo "🔄 Step 3/3: Triggering Flux reconciliation..."; \
		flux reconcile source git homelab -n flux-system 2>/dev/null || true; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "✅ Fast refresh complete!"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	else \
		echo "⚠️  Cluster $(CLUSTER_NAME) not found, falling back to full deployment..."; \
		$(MAKE) up ENV=$(ENV); \
	fi

fast-up: ## ⚡ Alias for fast-refresh
	@$(MAKE) fast-refresh ENV=$(ENV)


cancel: ## ⏸️  Cancel operation (usage: make cancel ENV=studio|pro|air or make cancel-studio|pro|air)
	@echo "🛑 Cancelling Pulumi operation for cluster: $(ENV)"
	@cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) && PULUMI_CONFIG_PASSPHRASE="" pulumi cancel --yes

destroy: ## 💥 Destroy cluster (usage: make destroy ENV=studio|pro|air or make destroy-studio|pro|air)
	$(call validate_env)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "⚠️  WARNING: Destroying $(ENV) cluster"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   Stack:        $(PULUMI_STACK)"
	@echo "   Cluster:      $(CLUSTER_NAME)"
	@echo "   Machine ID:   $(MACHINE_ID)"
	@echo "   Hostname:     $(MACHINE_HOSTNAME)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🔒 Safety Checks:"
	@echo ""
	@# Check 1: Verify Kind cluster exists locally
	@if ! kind get clusters 2>/dev/null | grep -qw "$(CLUSTER_NAME)"; then \
		echo "❌ ERROR: Kind cluster '$(CLUSTER_NAME)' does NOT exist on this machine ($(MACHINE_HOSTNAME))"; \
		echo ""; \
		echo "   This prevents accidental destruction of clusters on other machines."; \
		echo "   If you want to destroy a cluster, it must exist locally first."; \
		echo ""; \
		echo "   Available clusters on this machine:"; \
		kind get clusters 2>/dev/null | sed 's/^/     • /' || echo "     (none)"; \
		echo ""; \
		echo "   To destroy a cluster, you must run this command from the machine"; \
		echo "   where the cluster is actually running."; \
		exit 1; \
	fi
	@echo "   ✓ Cluster '$(CLUSTER_NAME)' exists locally"
	@# Check 2: Verify we can connect to the cluster
	@if ! kubectl cluster-info --context kind-$(CLUSTER_NAME) >/dev/null 2>&1; then \
		echo "   ⚠️  Warning: Cannot connect to cluster (may be starting up)"; \
	else \
		echo "   ✓ Cluster is accessible"; \
	fi
	@# Check 3: Verify Pulumi stack matches this machine
	@cd pulumi && \
		if PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) --non-interactive >/dev/null 2>&1; then \
			echo "   ✓ Pulumi stack '$(PULUMI_STACK)' exists and matches this machine"; \
		else \
			echo "   ⚠️  Warning: Pulumi stack '$(PULUMI_STACK)' not found (will be created on next 'make up')"; \
		fi
	@echo ""
	@if [ -z "$(SKIP_CONFIRM)" ]; then \
		echo "⚠️  This action will:"; \
		echo "   • Destroy Pulumi stack: $(PULUMI_STACK)"; \
		echo "   • Delete Kind cluster: $(CLUSTER_NAME)"; \
		echo "   • Remove all Docker containers"; \
		echo "   • Remove all Docker networks"; \
		echo "   • Clean up kubeconfig entries"; \
		echo "   • Prune unused Docker volumes"; \
		echo ""; \
		echo "   Machine: $(MACHINE_HOSTNAME) ($(MACHINE_ID))"; \
		echo ""; \
		echo "❗ This action CANNOT be undone!"; \
		echo ""; \
		read -p "Type 'yes' to confirm destruction of $(ENV) cluster on $(MACHINE_HOSTNAME): " confirm && [ "$$confirm" = "yes" ] || (echo "❌ Destruction cancelled." && exit 1); \
	fi
	@echo ""
	@echo "⚡ Fast cleanup (skipping K8s resource cleanup & backups)..."
	@echo ""
	@echo "🗑️  Step 1/6: Force deleting Kind cluster..."
	@kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true
	@echo "   ✓ Kind cluster deletion attempted"
	@echo ""
	@echo "🐳 Step 2/6: Force cleaning Docker containers..."
	@docker ps -a --filter "label=io.x-k8s.kind.cluster=$(CLUSTER_NAME)" -q 2>/dev/null | xargs docker rm -f 2>/dev/null || true
	@docker ps -a --format '{{.Names}}' 2>/dev/null | grep "^$(CLUSTER_NAME)-" | xargs docker rm -f 2>/dev/null || true
	@echo "   ✓ All containers force removed"
	@echo ""
	@echo "🔗 Step 3/6: Force cleaning kubeconfig..."
	@kubectl config delete-context kind-$(CLUSTER_NAME) 2>/dev/null || true
	@kubectl config delete-cluster kind-$(CLUSTER_NAME) 2>/dev/null || true
	@echo "   ✓ Kubeconfig cleaned"
	@echo ""
	@echo "🌐 Step 4/6: Force cleaning Docker networks..."
	@docker network ls --filter "label=io.x-k8s.kind.cluster=$(CLUSTER_NAME)" -q 2>/dev/null | xargs docker network rm 2>/dev/null || true
	@docker network ls --format '{{.Name}}' 2>/dev/null | grep "^$(CLUSTER_NAME)" | xargs docker network rm 2>/dev/null || true
	@echo "   ✓ All networks force removed"
	@echo ""
	@echo "💥 Step 5/6: Cleaning Pulumi stack..."
	@cd pulumi && PULUMI_CONFIG_PASSPHRASE="" pulumi stack select $(PULUMI_STACK) 2>/dev/null && PULUMI_CONFIG_PASSPHRASE="" pulumi stack rm $(PULUMI_STACK) --yes --force 2>/dev/null || true
	@echo "   ✓ Pulumi stack removed"
	@echo ""
	@echo "🧹 Step 6/6: Pruning Docker volumes..."
	@docker volume prune -f 2>/dev/null || true
	@echo "   ✓ All volumes force pruned"
	@echo ""
	@echo ""
	@echo "✅ Cleanup complete!"
	@if ls "$(CURDIR)/flux/infrastructure/bootstrap/sealed-secrets/$(CLUSTER_NAME)-sealed-secrets-key"*.yaml >/dev/null 2>&1; then \
		echo "   ✓ SealedSecrets keys backed up and ready for restore"; \
	else \
		echo "   ⚠️  No backup found (may be first-time setup)"; \
	fi
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ $(ENV) cluster cleanup complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "💡 SealedSecrets keys are backed up automatically"
	@echo "   Next 'make up ENV=$(ENV)' will restore them automatically"
	@echo ""

reconcile-all: ## 🔄 Reconcile all resources managed by Flux
	$(call ensure_cluster_context)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔄 Reconciling all Flux resources"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   Environment: $(ENV)"
	@echo "   Cluster:     $(CLUSTER_NAME)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📚 Step 1/3: Reconciling GitRepositories..."
	@for repo in $$(flux get sources git -A --no-header | awk '{print $$1 ":" $$2}'); do \
		namespace=$$(echo $$repo | cut -d: -f1); \
		name=$$(echo $$repo | cut -d: -f2); \
		echo "   Reconciling $$name in $$namespace..."; \
		flux reconcile source git $$name -n $$namespace; \
	done
	@echo "   ✓ GitRepositories reconciled"
	@echo ""
	@echo "📦 Step 2/3: Reconciling Kustomizations..."
	@for kustomization in $$(flux get kustomizations -A --no-header | awk '{print $$1 ":" $$2}'); do \
		namespace=$$(echo $$kustomization | cut -d: -f1); \
		name=$$(echo $$kustomization | cut -d: -f2); \
		echo "   Reconciling $$name in $$namespace..."; \
		flux reconcile kustomization $$name -n $$namespace; \
	done
	@echo "   ✓ Kustomizations reconciled"
	@echo ""
	@echo "⎈ Step 3/3: Reconciling HelmReleases..."
	@for helmrelease in $$(flux get helmreleases -A --no-header | awk '{print $$1 ":" $$2}'); do \
		namespace=$$(echo $$helmrelease | cut -d: -f1); \
		name=$$(echo $$helmrelease | cut -d: -f2); \
		echo "   Reconciling $$name in $$namespace..."; \
		flux reconcile helmrelease $$name -n $$namespace; \
	done
	@echo "   ✓ HelmReleases reconciled"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ All Flux resources reconciled successfully!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "💡 Tip: Use 'make observe' to watch the status in real-time"
	@echo ""

reconcile-helm: ## ⎈ Reconcile all HelmReleases
	$(call ensure_cluster_context)
	@echo "⎈ Reconciling all HelmReleases for $(ENV)..."
	@for helmrelease in $$(flux get helmreleases -A --no-header | awk '{print $$1 ":" $$2}'); do \
		namespace=$$(echo $$helmrelease | cut -d: -f1); \
		name=$$(echo $$helmrelease | cut -d: -f2); \
		echo "   Reconciling $$name in $$namespace..."; \
		flux reconcile helmrelease $$name -n $$namespace; \
	done
	@echo "✅ All HelmReleases reconciled!"

reconcile-git: ## 🔀 Reconcile all GitRepositories
	$(call ensure_cluster_context)
	@echo "📚 Reconciling all GitRepositories for $(ENV)..."
	@for repo in $$(flux get sources git -A --no-header | awk '{print $$1 ":" $$2}'); do \
		namespace=$$(echo $$repo | cut -d: -f1); \
		name=$$(echo $$repo | cut -d: -f2); \
		echo "   Reconciling $$name in $$namespace..."; \
		flux reconcile source git $$name -n $$namespace; \
	done
	@echo "✅ All GitRepositories reconciled!"

reconcile-all-force: ## 📦 Force reconcile all Kustomizations (with source)
	$(call ensure_cluster_context)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📦 Force Reconciling all Kustomizations for $(ENV)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@for kustomization in $$(flux get kustomizations -A --no-header | awk '{print $$1 ":" $$2}'); do \
		namespace=$$(echo $$kustomization | cut -d: -f1); \
		name=$$(echo $$kustomization | cut -d: -f2); \
		echo "   🔄 Force reconciling $$name in $$namespace..."; \
		flux reconcile kustomization $$name -n $$namespace --with-source || true; \
	done
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ All Kustomizations force reconciled!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""

rollout: ## 🔁 Rollout restart a deployment (usage: make rollout <deployment-name>)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "❌ Error: Deployment name is required"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "Usage: make rollout <deployment-name>"; \
		echo "Example: make rollout homepage"; \
		echo ""; \
		echo "Available deployments:"; \
		kubectl get deployments -A 2>/dev/null | awk 'NR>1 {printf "   • %-30s (namespace: %s)\n", $$2, $$1}' | sort || echo "   No deployments found"; \
		echo ""; \
		echo "Available daemonsets:"; \
		kubectl get daemonsets -A 2>/dev/null | awk 'NR>1 {printf "   • %-30s (namespace: %s)\n", $$2, $$1}' | sort || echo "   No daemonsets found"; \
		echo ""; \
		echo "Available statefulsets:"; \
		kubectl get statefulsets -A 2>/dev/null | awk 'NR>1 {printf "   • %-30s (namespace: %s)\n", $$2, $$1}' | sort || echo "   No statefulsets found"; \
		echo ""; \
		exit 1; \
	fi
	@RESOURCE="$(filter-out $@,$(MAKECMDGOALS))"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "🔁 Rolling out restart: $$RESOURCE"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "🔍 Searching for resource..."; \
	DEPLOYMENT=$$(kubectl get deployment -A 2>/dev/null | grep -w "$$RESOURCE" | awk '{print $$1 " " $$2}' | head -n 1); \
	DAEMONSET=$$(kubectl get daemonset -A 2>/dev/null | grep -w "$$RESOURCE" | awk '{print $$1 " " $$2}' | head -n 1); \
	STATEFULSET=$$(kubectl get statefulset -A 2>/dev/null | grep -w "$$RESOURCE" | awk '{print $$1 " " $$2}' | head -n 1); \
	if [ -n "$$DEPLOYMENT" ]; then \
		NAMESPACE=$$(echo $$DEPLOYMENT | awk '{print $$1}'); \
		NAME=$$(echo $$DEPLOYMENT | awk '{print $$2}'); \
		echo "   Found deployment: $$NAME in namespace: $$NAMESPACE"; \
		echo ""; \
		echo "🔄 Restarting deployment..."; \
		kubectl rollout restart deployment/$$NAME -n $$NAMESPACE; \
		echo ""; \
		echo "⏳ Waiting for rollout to complete..."; \
		kubectl rollout status deployment/$$NAME -n $$NAMESPACE --timeout=300s; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "✅ Deployment $$NAME rolled out successfully!"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	elif [ -n "$$DAEMONSET" ]; then \
		NAMESPACE=$$(echo $$DAEMONSET | awk '{print $$1}'); \
		NAME=$$(echo $$DAEMONSET | awk '{print $$2}'); \
		echo "   Found daemonset: $$NAME in namespace: $$NAMESPACE"; \
		echo ""; \
		echo "🔄 Restarting daemonset..."; \
		kubectl rollout restart daemonset/$$NAME -n $$NAMESPACE; \
		echo ""; \
		echo "⏳ Waiting for rollout to complete..."; \
		kubectl rollout status daemonset/$$NAME -n $$NAMESPACE --timeout=300s; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "✅ DaemonSet $$NAME rolled out successfully!"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	elif [ -n "$$STATEFULSET" ]; then \
		NAMESPACE=$$(echo $$STATEFULSET | awk '{print $$1}'); \
		NAME=$$(echo $$STATEFULSET | awk '{print $$2}'); \
		echo "   Found statefulset: $$NAME in namespace: $$NAMESPACE"; \
		echo ""; \
		echo "🔄 Restarting statefulset..."; \
		kubectl rollout restart statefulset/$$NAME -n $$NAMESPACE; \
		echo ""; \
		echo "⏳ Waiting for rollout to complete..."; \
		kubectl rollout status statefulset/$$NAME -n $$NAMESPACE --timeout=300s; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "✅ StatefulSet $$NAME rolled out successfully!"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	else \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "❌ Error: Resource '$$RESOURCE' not found"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "Available resources:"; \
		echo ""; \
		echo "Deployments:"; \
		kubectl get deployments -A 2>/dev/null | awk 'NR>1 {printf "   • %-30s (namespace: %s)\n", $$2, $$1}' | sort || echo "   No deployments found"; \
		echo ""; \
		echo "DaemonSets:"; \
		kubectl get daemonsets -A 2>/dev/null | awk 'NR>1 {printf "   • %-30s (namespace: %s)\n", $$2, $$1}' | sort || echo "   No daemonsets found"; \
		echo ""; \
		echo "StatefulSets:"; \
		kubectl get statefulsets -A 2>/dev/null | awk 'NR>1 {printf "   • %-30s (namespace: %s)\n", $$2, $$1}' | sort || echo "   No statefulsets found"; \
		echo ""; \
		exit 1; \
	fi
	@echo ""

# Catch-all target to prevent "No rule to make target" errors when passing resource names
%:
	@:

observe: ## 👀 Watch Kustomizations and HelmReleases status in real-time
	$(call ensure_cluster_context)
	@echo "👀 Observing Flux resources in real-time..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   Environment: $(ENV)"
	@echo "   Cluster:     $(CLUSTER_NAME)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📦 Kustomizations:"
	@flux get kustomizations -A
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🚀 HelmReleases:"
	@flux get helmreleases -A
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "💡 Tip: Use 'watch -n 2 make observe' for continuous monitoring"

watch-flux: ## 🔄 Continuously watch Flux resources (auto-refreshing every 2s)
	$(call ensure_cluster_context)
	@watch -n 2 "echo '👀 Observing Flux Resources ($(ENV))' && echo '' && echo '📦 Kustomizations:' && echo 'NAMESPACE        NAME                     READY   MESSAGE' && flux get kustomizations -A --no-header | awk '{printf \"%-15s %-23s %-7s\", \$$1, \$$2, \$$4; for(i=5;i<=NF;i++) printf \" %s\", \$$i; print \"\"}' | cut -c1-80 && echo '' && FIRST_FAILED=\$$(flux get kustomizations -A --no-header | grep 'False' | head -n 1) && if [ -n \"\$$FIRST_FAILED\" ]; then NS=\$$(echo \"\$$FIRST_FAILED\" | awk '{print \$$1}') && NAME=\$$(echo \"\$$FIRST_FAILED\" | awk '{print \$$2}') && echo '' && echo '╔═══════════════════════════════════════════════════════════╗' && echo '║  ❌ FAILED: '\$$NAME' (ns: '\$$NS')' && echo '╚═══════════════════════════════════════════════════════════╝' && kubectl get kustomization \$$NAME -n \$$NS -o jsonpath='{.status.conditions[?(@.type==\"Ready\")].message}' | fold -w 60 -s && echo '' && echo 'Last Applied: ' && kubectl get kustomization \$$NAME -n \$$NS -o jsonpath='{.status.lastAppliedRevision}' && echo '' && echo ''; fi && echo '🚀 HelmReleases:' && flux get helmreleases -A"

# =============================================================================
# Testing
# =============================================================================
test: ## 🧪 Run all IaC tests (unit + integration + scripts)
	@echo "🧪 Running all tests..."
	@cd tests && ./run-all-tests.sh

test-unit: ## 🧪 Run IaC Go unit tests only
	@echo "🧪 Running Go unit tests..."
	@cd tests/pulumi && go test -v ./...

test-scripts: ## 🧪 Run shell script for IaC tests (BATS)
	@echo "🧪 Running shell script tests..."
	@cd tests && bats scripts/setup-local-registry.bats

test-integration: ## 🧪 Run integration tests (requires Docker)
	@echo "🧪 Running integration tests..."
	@cd tests && bats integration/cluster-provisioning.bats

test-coverage: ## 📊 Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@cd tests && ./run-all-tests.sh --coverage
	@echo "📊 Coverage report generated at tests/pulumi/coverage.html"

# =============================================================================
# Port Forwarding
# =============================================================================
pf-grafana: ## 🌐 Port-forward to Grafana (localhost:3000)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🌐 Port-forwarding to Grafana"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   URL: http://localhost:3000"
	@echo "   Username: admin"
	@echo "   Password: check sealed secrets or GRAFANA_PASSWORD env var"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@kubectl port-forward -n prometheus svc/kube-prometheus-stack-grafana 3000:80

pf-grafana-secret: ## 🔑 Show Grafana admin password
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔑 Grafana Admin Credentials"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   Username: admin"
	@printf "   Password: "
	@kubectl get secret grafana-secrets -n prometheus -o jsonpath="{.data.GRAFANA_PASSWORD}" 2>/dev/null | base64 -d && echo "" || echo "❌ Secret not found"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

grafana-mcp-setup: ## 🔧 Setup Grafana MCP server (fully automated)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔧 Setting up Grafana MCP Server (Fully Automated)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@POD=$$(kubectl get pod -n prometheus -l app.kubernetes.io/name=grafana -o jsonpath='{.items[0].metadata.name}') && \
		GRAFANA_PASSWORD=$$(kubectl get secret -n prometheus prometheus -o jsonpath='{.data.grafana-password}' | base64 -d) && \
		echo "🔄 Resetting Grafana admin password to match secret..." && \
		kubectl exec -n prometheus $$POD -c grafana -- grafana cli admin reset-admin-password "$$GRAFANA_PASSWORD" >/dev/null 2>&1 && \
		echo "✅ Password reset complete" && \
		echo "🚀 Running MCP setup..." && \
		GRAFANA_PASSWORD="$$GRAFANA_PASSWORD" python3 flux/infrastructure/prometheus-operator/scripts/setup_grafana_mcp.py

pf-prometheus: ## 📊 Port-forward to Prometheus (localhost:9090)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📊 Port-forwarding to Prometheus"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   URL: http://localhost:9090"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@kubectl port-forward -n prometheus svc/kube-prometheus-stack-prometheus 9090:9090

pf-alertmanager: ## 🚨 Port-forward to Alertmanager (localhost:9093)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🚨 Port-forwarding to Alertmanager"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   URL: http://localhost:9093"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@kubectl port-forward -n prometheus svc/kube-prometheus-stack-alertmanager 9093:9093

pf-rabbitmq: ## 🐰 Port-forward to RabbitMQ Management UI (localhost:15672)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🐰 Port-forwarding to RabbitMQ Management UI"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   URL: http://localhost:15672"
	@echo "   Default credentials: guest/guest"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@kubectl port-forward -n knative-lambda svc/rabbitmq-cluster-knative-lambda 15672:15672

rabbitmq-stats: ## 🐰 Show RabbitMQ statistics (queues, exchanges, connections, nodes)
	$(call ensure_cluster_context)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🐰 RabbitMQ Statistics"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@RABBITMQ_POD=$$(kubectl get pods -n knative-lambda -l app.kubernetes.io/name=knative-lambda -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -z "$$RABBITMQ_POD" ]; then \
		RABBITMQ_POD=$$(kubectl get pods -n knative-lambda -l app=rabbitmq -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	fi; \
	if [ -z "$$RABBITMQ_POD" ]; then \
		echo "❌ No RabbitMQ pods found in knative-lambda namespace"; \
		echo ""; \
		echo "Available pods:"; \
		kubectl get pods -n knative-lambda 2>/dev/null || echo "   No pods found"; \
		exit 1; \
	fi; \
	echo "✅ Found RabbitMQ pod: $$RABBITMQ_POD"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "📊 Overview"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	kubectl exec -n knative-lambda $$RABBITMQ_POD -- rabbitmqctl status 2>/dev/null | head -20 || echo "   ⚠️ Could not retrieve status"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "📬 Queues"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	kubectl exec -n knative-lambda $$RABBITMQ_POD -- rabbitmqctl list_queues name messages_ready messages_unacknowledged messages consumers memory state 2>/dev/null || echo "   ⚠️ Could not retrieve queue information"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "📤 Exchanges"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	kubectl exec -n knative-lambda $$RABBITMQ_POD -- rabbitmqctl list_exchanges name type durable auto_delete 2>/dev/null | grep -E "(knative|lambda|broker|^name)" || echo "   ⚠️ Could not retrieve exchange information"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "🔗 Connections"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	kubectl exec -n knative-lambda $$RABBITMQ_POD -- rabbitmqctl list_connections name peer_host peer_port state 2>/dev/null || echo "   ⚠️ Could not retrieve connection information"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "👥 Channels"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	kubectl exec -n knative-lambda $$RABBITMQ_POD -- rabbitmqctl list_channels name connection number 2>/dev/null | head -10 || echo "   ⚠️ Could not retrieve channel information"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "💡 Tip: Use 'make pf-rabbitmq' to access the Management UI at http://localhost:15672"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# =============================================================================
# Docker Images
# =============================================================================
REGISTRY_PORT ?= 5001
REGISTRY_HOST := localhost:$(REGISTRY_PORT)
DOCKER_DIR := docker
BUILDER_NAME := homelab-builder

# Generic Docker build function
# Args: IMAGE_NAME, DOCKERFILE, BUILD_CONTEXT, EXTRA_TAG, BUILD_ARGS
define build_docker_image
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🐳 Building $(1) Docker Image"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if [ "$(shell docker inspect -f '{{.State.Running}}' kind-registry 2>/dev/null || echo false)" != 'true' ]; then \
		echo "❌ Local registry 'kind-registry' is not running!"; \
		echo "   Run: make init-registry"; \
		exit 1; \
	fi
	@ARCH=$$(uname -m); \
	case "$$ARCH" in \
		x86_64) PLATFORM="linux/amd64"; TARGETARCH="amd64" ;; \
		arm64|aarch64) PLATFORM="linux/arm64"; TARGETARCH="arm64" ;; \
		*) echo "❌ Unsupported architecture: $$ARCH"; exit 1 ;; \
	esac; \
	if ! docker buildx ls | grep -q "$(BUILDER_NAME)"; then \
		echo "🔧 Creating builder with host network access..."; \
		docker buildx create --name "$(BUILDER_NAME)" --driver docker-container --driver-opt network=host --bootstrap || true; \
	fi; \
	docker buildx use "$(BUILDER_NAME)" >/dev/null 2>&1 || true; \
	echo "🏗️  Building $(1) image for platform: $$PLATFORM..."; \
	EXTRA_TAG="$(strip $(4))"; \
	BUILD_ARGS="$(strip $(5))"; \
	IMAGE_NAME="$(strip $(1))"; \
	REGISTRY="$(strip $(REGISTRY_HOST))"; \
	TAG_ARGS="-t $$REGISTRY/$$IMAGE_NAME:latest"; \
	if [ -n "$$EXTRA_TAG" ] && [ "$$EXTRA_TAG" != "" ]; then TAG_ARGS="$$TAG_ARGS -t $$REGISTRY/$$IMAGE_NAME:$$EXTRA_TAG"; fi; \
	BUILD_CTX="$(strip $(3))"; \
	DOCKERFILE="$(strip $(2))"; \
	if [ -z "$$BUILD_CTX" ]; then \
		echo "❌ Error: BUILD_CTX is empty!"; \
		exit 1; \
	fi; \
	if [ -n "$$BUILD_ARGS" ]; then \
		docker buildx build \
			--builder "$(BUILDER_NAME)" \
			--platform "$$PLATFORM" \
			$$BUILD_ARGS \
			$$TAG_ARGS \
			-f "$$DOCKERFILE" \
			--cache-from type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache \
			--cache-to type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache,mode=max \
			--push \
			"$$BUILD_CTX" || \
		( \
			echo "⚠️  Build failed, retrying with credential helper disabled..."; \
			DOCKER_CONFIG_TMP=$$(mktemp -d); \
			echo '{}' > "$$DOCKER_CONFIG_TMP/config.json"; \
			DOCKER_CONFIG="$$DOCKER_CONFIG_TMP" docker buildx build \
				--builder "$(BUILDER_NAME)" \
				--platform "$$PLATFORM" \
				$$BUILD_ARGS \
				$$TAG_ARGS \
				-f "$$DOCKERFILE" \
				--cache-from type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache \
				--cache-to type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache,mode=max \
				--push \
				"$$BUILD_CTX" && \
			rm -rf "$$DOCKER_CONFIG_TMP" || exit 1 \
		); \
	else \
		docker buildx build \
			--builder "$(BUILDER_NAME)" \
			--platform "$$PLATFORM" \
			$$TAG_ARGS \
			-f "$$DOCKERFILE" \
			--cache-from type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache \
			--cache-to type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache,mode=max \
			--push \
			"$$BUILD_CTX" || \
		( \
			echo "⚠️  Build failed, retrying with credential helper disabled..."; \
			DOCKER_CONFIG_TMP=$$(mktemp -d); \
			echo '{}' > "$$DOCKER_CONFIG_TMP/config.json"; \
			DOCKER_CONFIG="$$DOCKER_CONFIG_TMP" docker buildx build \
				--builder "$(BUILDER_NAME)" \
				--platform "$$PLATFORM" \
				$$TAG_ARGS \
				-f "$$DOCKERFILE" \
				--cache-from type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache \
				--cache-to type=registry,ref=$$REGISTRY/$$IMAGE_NAME:buildcache,mode=max \
				--push \
				"$$BUILD_CTX" && \
			rm -rf "$$DOCKER_CONFIG_TMP" || exit 1 \
		); \
	fi; \
	echo ""; \
	echo "✅ $(1) image built and pushed successfully!"
endef

prewarm-images: ## 🐳 Pre-warm images (shared across ALL clusters)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🐳 Pre-warming images (shared across ALL clusters)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(MAKE) build-docker-kubectl
	@$(MAKE) build-docker-garak
	@./$(KIND_SCRIPTS)/prewarm-images.sh

cleanup-registry: ## 🧹 Clean up unused images from local registry
	@./$(KIND_SCRIPTS)/cleanup-registry.sh

build-docker-kubectl: ## 🐳 Build kubectl Docker image
	$(call build_docker_image, \
		kubectl, \
		$(DOCKER_DIR)/kubectl.Dockerfile, \
		$(DOCKER_DIR), \
		v1.34.0, \
		--build-arg KUBECTL_VERSION=v1.34.0 \
		--build-arg KREW_VERSION=v0.4.4 \
		--build-arg TARGETARCH=$$TARGETARCH \
		--build-arg LINKERD_VERSION=edge-25.11.1)

build-docker-garak: ## 🐳 Build garak Docker image
	@if [ ! -d "flux/infrastructure/garak/src" ]; then \
		echo "❌ garak src directory not found at flux/infrastructure/garak/src"; \
		exit 1; \
	fi
	$(call build_docker_image, \
		garak, \
		$(DOCKER_DIR)/garak.Dockerfile, \
		flux/infrastructure/garak, \
		, \
		)

# =============================================================================
# AI Agents
# =============================================================================
AI_AGENTS_DIR := flux/ai

build-ai-agents: build-agent-bruno build-agent-contracts ## 🤖 Build all AI agent images

build-agent-bruno: ## 🤖 Build agent-bruno chatbot image
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🤖 Building agent-bruno"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd $(AI_AGENTS_DIR)/agent-bruno && $(MAKE) build REGISTRY=$(REGISTRY_HOST)

build-agent-contracts: ## 🛡️ Build agent-contracts images (all services)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🛡️ Building agent-contracts"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd $(AI_AGENTS_DIR)/agent-contracts && $(MAKE) build REGISTRY=$(REGISTRY_HOST)

push-ai-agents: push-agent-bruno push-agent-contracts ## 🤖 Push all AI agent images

push-agent-bruno: build-agent-bruno ## 🤖 Push agent-bruno image
	@echo "📤 Pushing agent-bruno..."
	@docker push $(REGISTRY_HOST)/agent-bruno/chatbot:latest
	@VERSION=$$(cat $(AI_AGENTS_DIR)/agent-bruno/VERSION 2>/dev/null || echo "0.1.0"); \
		docker push $(REGISTRY_HOST)/agent-bruno/chatbot:$$VERSION 2>/dev/null || true

push-agent-contracts: build-agent-contracts ## 🛡️ Push agent-contracts images
	@echo "📤 Pushing agent-contracts..."
	@cd $(AI_AGENTS_DIR)/agent-contracts && $(MAKE) push REGISTRY=$(REGISTRY_HOST)

deploy-ollama: ## 🦙 Deploy Ollama LLM server
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🦙 Deploying Ollama to ai-inference namespace"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@kubectl apply -k flux/infrastructure/ollama/
	@echo "⏳ Waiting for Ollama to be ready..."
	@kubectl rollout status deployment/ollama -n ai-inference --timeout=300s || true
	@echo "✅ Ollama deployed!"

deploy-agent-bruno: ## 🤖 Deploy agent-bruno chatbot
	$(call validate_env)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🤖 Deploying agent-bruno to $(ENV)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@OVERLAY="studio"; \
	if echo "$(ENV)" | grep -q "pro"; then OVERLAY="pro"; fi; \
	kubectl apply -k $(AI_AGENTS_DIR)/agent-bruno/k8s/kustomize/$$OVERLAY/
	@echo "✅ agent-bruno deployed!"

deploy-agent-contracts: ## 🛡️ Deploy agent-contracts
	$(call validate_env)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🛡️ Deploying agent-contracts to $(ENV)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@OVERLAY="studio"; \
	if echo "$(ENV)" | grep -q "pro"; then OVERLAY="pro"; fi; \
	kubectl apply -k $(AI_AGENTS_DIR)/agent-contracts/k8s/kustomize/$$OVERLAY/
	@echo "✅ agent-contracts deployed!"

deploy-ai-agents: deploy-ollama deploy-agent-bruno deploy-agent-contracts ## 🤖 Deploy all AI agents (Ollama + Bruno + Contracts)
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ All AI agents deployed!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🔗 Services:"
	@echo "   • Ollama:          http://ollama.ai-inference.svc.cluster.local:11434"
	@echo "   • Agent-Bruno:     http://agent-bruno.agent-bruno.svc.cluster.local"
	@echo "   • Agent-Contracts: (Knative services in agent-contracts namespace)"
	@echo ""
	@echo "💡 Pull models with:"
	@echo "   kubectl exec -n ai-inference deploy/ollama -- ollama pull llama3.2:3b"
	@echo ""

pull-ollama-models: ## 🦙 Pull default Ollama models
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🦙 Pulling Ollama models"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📥 Pulling llama3.2:3b (for agent-bruno)..."
	@kubectl exec -n ai-inference deploy/ollama -- ollama pull llama3.2:3b
	@echo ""
	@echo "📥 Pulling deepseek-coder-v2:16b (for agent-contracts)..."
	@kubectl exec -n ai-inference deploy/ollama -- ollama pull deepseek-coder-v2:16b
	@echo ""
	@echo "✅ Models pulled!"
	@kubectl exec -n ai-inference deploy/ollama -- ollama list

ai-status: ## 🤖 Show AI agents status
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🤖 AI Agents Status"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🦙 Ollama (ai-inference):"
	@kubectl get pods,svc -n ai-inference 2>/dev/null || echo "   Not deployed"
	@echo ""
	@echo "🤖 Agent-Bruno:"
	@kubectl get ksvc,pods -n agent-bruno 2>/dev/null || echo "   Not deployed"
	@echo ""
	@echo "🛡️ Agent-Contracts:"
	@kubectl get ksvc,pods -n agent-contracts 2>/dev/null || echo "   Not deployed"
	@echo ""
	@echo "📦 Ollama Models:"
	@kubectl exec -n ai-inference deploy/ollama -- ollama list 2>/dev/null || echo "   Ollama not running"
	@echo ""