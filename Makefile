.PHONY: help secret-argocd pf-argocd bootstrap-flux-dev bootstrap-flux-prd flux-status flux-logs init-studio init-homelab up-studio up-homelab destroy-studio destroy-homelab cancel clean logs-dev logs-prd status-dev status-prd setup-env flux-refresh flux-refresh-bruno reconcile rollout flagger-status flagger-logs promote-canary rollback-canary istio-status istio-logs istio-proxy-status linkerd-install linkerd-install-clean linkerd-uninstall linkerd-status linkerd-dashboard linkerd-check

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# =============================================================================
# Pulumi Operations
# =============================================================================

init: ## Initialize Pulumi homelab stack
	cd pulumi && pulumi stack init homelab

up: ## Deploy homelab stack
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "Error: GITHUB_TOKEN environment variable is required"; \
		echo "Run 'make setup-env' for instructions"; \
		exit 1; \
	fi
	@if [ -z "$$CLOUDFLARE_TOKEN" ]; then \
		echo "Error: CLOUDFLARE_TOKEN environment variable is required"; \
		echo "Run 'make setup-env' for instructions"; \
		exit 1; \
	fi
	cd pulumi && pulumi stack select homelab && pulumi refresh --yes && pulumi up --yes
	@echo "🔐 Creating Kubernetes secrets..."
	./scripts/create-secrets.sh

destroy: ## Destroy homelab stack
	cd pulumi && pulumi stack select homelab && pulumi destroy --yes

cancel: ## Cancel ongoing Pulumi operation
	cd pulumi && pulumi stack select homelab && pulumi cancel

# =============================================================================
# Flux Operations
# =============================================================================

flux-refresh: ## Force refresh all Flux HelmRepositories, GitRepositories, and HelmReleases
	@echo "🔄 Forcing refresh of all Flux resources..."
	@echo "📦 Refreshing HelmRepositories..."
	kubectl annotate helmrepository --all -n flux-system --overwrite reconcile.fluxcd.io/requestedAt="$$(date +%s)"
	@echo "📚 Refreshing GitRepositories..."
	kubectl annotate gitrepository --all -n flux-system --overwrite reconcile.fluxcd.io/requestedAt="$$(date +%s)"
	@echo "🚀 Refreshing HelmReleases..."
	kubectl annotate helmrelease --all -n flux-system --overwrite reconcile.fluxcd.io/requestedAt="$$(date +%s)"
	@echo "✅ Flux refresh triggered for all resources"

reconcile-gitrepo: ## Reconcile GitRepository (usage: make reconcile-gitrepo homelab)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "❌ Error: GitRepository name is required"; \
		echo "Usage: make reconcile-gitrepo <gitrepo-name>"; \
		echo "Example: make reconcile-gitrepo homelab"; \
		exit 1; \
	fi
	@GITREPO="$(filter-out $@,$(MAKECMDGOALS))"; \
	echo "🔄 Reconciling GitRepository: $$GITREPO..."; \
	flux reconcile source git $$GITREPO -n flux-system; \
	echo "✅ GitRepository $$GITREPO reconciled successfully!"

reconcile: ## Reconcile HelmRelease(s) managed by Flux (usage: make reconcile [service-name] or make reconcile for all)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "🔄 Reconciling all HelmReleases..."; \
		echo "📦 Reconciling alloy..."; \
		flux reconcile helmrelease alloy -n alloy; \
		echo "🔐 Reconciling cert-manager..."; \
		flux reconcile helmrelease cert-manager -n cert-manager; \
		echo "🖥️  Reconciling headlamp..."; \
		flux reconcile helmrelease headlamp -n headlamp; \
		echo "🏠 Reconciling homepage..."; \
		flux reconcile helmrelease homepage -n homepage; \
		echo "⚙️  Reconciling knative-operator..."; \
		flux reconcile helmrelease knative-operator -n knative-operator; \
		echo "📝 Reconciling loki..."; \
		flux reconcile helmrelease loki -n loki; \
		echo "📊 Reconciling metrics-server..."; \
		flux reconcile helmrelease metrics-server -n metrics-server; \
		echo "🍃 Reconciling mongodb..."; \
		flux reconcile helmrelease mongodb -n mongodb; \
		echo "🔥 Reconciling prometheus-operator..."; \
		flux reconcile helmrelease prometheus-operator -n prometheus; \
		echo "⏱️  Reconciling tempo..."; \
		flux reconcile helmrelease tempo -n tempo; \
		echo "✅ All HelmReleases reconciled successfully!"; \
	else \
		SERVICE="$(filter-out $@,$(MAKECMDGOALS))"; \
		echo "🔄 Reconciling HelmRelease: $$SERVICE..."; \
		NAMESPACE=$$(kubectl get helmrelease -A | grep -w "$$SERVICE" | awk '{print $$1}' | head -n 1); \
		if [ -z "$$NAMESPACE" ]; then \
			echo "❌ Error: HelmRelease '$$SERVICE' not found in any namespace"; \
			echo "Available HelmReleases:"; \
			kubectl get helmrelease -A | awk 'NR>1 {print "  - " $$2 " (namespace: " $$1 ")"}' | sort; \
			exit 1; \
		fi; \
		echo "📍 Found HelmRelease in namespace: $$NAMESPACE"; \
		flux reconcile helmrelease $$SERVICE -n $$NAMESPACE; \
		echo "✅ HelmRelease $$SERVICE reconciled successfully!"; \
	fi

rollout: ## Rollout restart a deployment (usage: make rollout homepage-api)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "❌ Error: Deployment name is required"; \
		echo "Usage: make rollout <deployment-name>"; \
		echo "Example: make rollout homepage-api"; \
		echo ""; \
		echo "Available deployments:"; \
		kubectl get deployments -A | awk 'NR>1 {print "  - " $$2 " (namespace: " $$1 ")"}' | sort; \
		exit 1; \
	fi
	@DEPLOY="$(filter-out $@,$(MAKECMDGOALS))"; \
	echo "🔄 Rolling out deployment: $$DEPLOY..."; \
	NAMESPACE=$$(kubectl get deployment -A | grep -w "$$DEPLOY" | awk '{print $$1}' | head -n 1); \
	if [ -z "$$NAMESPACE" ]; then \
		echo "❌ Error: Deployment '$$DEPLOY' not found in any namespace"; \
		echo "Available deployments:"; \
		kubectl get deployments -A | awk 'NR>1 {print "  - " $$2 " (namespace: " $$1 ")"}' | sort; \
		exit 1; \
	fi; \
	echo "📍 Found deployment in namespace: $$NAMESPACE"; \
	kubectl rollout restart deployment/$$DEPLOY -n $$NAMESPACE; \
	echo "✅ Deployment $$DEPLOY rolled out successfully!"

# Catch-all target to prevent "No rule to make target" errors when passing deployment names
%:
	@:

# =============================================================================
# Linkerd Operations
# =============================================================================

linkerd-install: ## Install Linkerd service mesh (manual installation)
	@echo "🚀 Installing Linkerd service mesh..."
	scripts/install-linkerd.sh homelab

linkerd-status: ## Check Linkerd status
	@echo "📊 Linkerd Status:"
	linkerd check --context kind-homelab

linkerd-viz-install: ## Install Linkerd Viz programmatically
	@echo "📊 Installing Linkerd Viz programmatically..."
	scripts/install-linkerd-viz.sh homelab

linkerd-viz-status: ## Check Linkerd Viz status
	@echo "📊 Linkerd Viz Status:"
	linkerd viz check --context kind-homelab

linkerd-dashboard: ## Access Linkerd dashboard
	@echo "🌐 Opening Linkerd dashboard..."
	@echo "Dashboard will be available at: http://localhost:8084"
	linkerd viz dashboard --context kind-homelab --port 8084