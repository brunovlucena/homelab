# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#	🏗️ KNATIVE LAMBDA BUILDER - SIMPLIFIED MAKEFILE
#
#	🎯 Purpose: Build, test, and deploy the Knative Lambda Builder service
#	💡 Features: DRY, KISS, simplified targets with essential functionality
#
#	🔧 USAGE:
#	  make              - Show help
#	  make build        - Build the service
#	  make test         - Run tests
#	  make deploy       - Deploy to cluster
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 CONFIGURATION VARIABLES                                             │
# └─────────────────────────────────────────────────────────────────────────┘

# 🏷️ Project Configuration
PROJECT_NAME := knative-lambda-new
SERVICE_NAME := knative-lambda-builder
VERSION ?= $(shell grep '^version:' deploy/Chart.yaml 2>/dev/null | awk '{print $$2}' || echo "0.1.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR := ./build
BIN_DIR := ./bin
BINARY_NAME := service
SIDECAR_BINARY := sidecar

# 🐳 Docker Configuration
DOCKER_REGISTRY := ghcr.io/brunovlucena
DOCKER_REPO := $(DOCKER_REGISTRY)/knative-lambda-$(SERVICE_NAME)
# Image tags based on branch
ifeq ($(GIT_BRANCH),main)
	DOCKER_TAG := $(VERSION)
else ifeq ($(GIT_BRANCH),develop)
	DOCKER_TAG := $(VERSION)-beta.$(shell date +%s)
else
	DOCKER_TAG := $(VERSION)-dev.$(GIT_COMMIT)
endif

# 🏠 Local Registry Configuration
LOCAL_REGISTRY := ghcr.io/brunovlucena
LOCAL_REPO := $(LOCAL_REGISTRY)/knative-lambda-builder
LOCAL_TAG := local

# 🐳 Sidecar Docker Configuration
SIDECAR_SERVICE_NAME := knative-lambda-sidecar
SIDECAR_DOCKER_REPO := $(DOCKER_REGISTRY)/knative-lambda-$(SIDECAR_SERVICE_NAME)
SIDECAR_DOCKER_TAG := $(DOCKER_TAG)

# 🐳 Metrics Pusher Docker Configuration
METRICS_PUSHER_SERVICE_NAME := knative-lambda-metrics-pusher
METRICS_PUSHER_DOCKER_REPO := $(DOCKER_REGISTRY)/knative-lambda-$(METRICS_PUSHER_SERVICE_NAME)
METRICS_PUSHER_DOCKER_TAG := $(DOCKER_TAG)

# ☸️ Kubernetes Configuration
ENV ?= dev
NAMESPACE ?= knative-lambda-$(ENV)

# 🎨 Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m
COLOR_RED := \033[31m

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 COMMON FUNCTIONS                                                    │
# └─────────────────────────────────────────────────────────────────────────┘

# 🔧 Print colored status messages
define print_status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)$(1)$(COLOR_RESET)"
endef

define print_success
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ $(1)$(COLOR_RESET)"
endef

define print_warning
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)⚠️ $(1)$(COLOR_RESET)"
endef

# 🔧 Check if command exists
define check_command
	@which $(1) > /dev/null || (echo "$(COLOR_YELLOW)Installing $(1)...$(COLOR_RESET)" && $(2))
endef

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 HELP TARGET                                                         │
# └─────────────────────────────────────────────────────────────────────────┘

.DEFAULT_GOAL := help

help: ## 📋 Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🏗️ Knative Lambda Builder - Available Commands$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)🔨 Build & Test:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep "🔨\|🧪\|🏗️" | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)🐳 Docker:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep "🐳\|🚀" | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)☸️ Kubernetes:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep "☸️\|🚀\|🔌" | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)🛠️ Development:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep "🛠️\|🔧" | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔨 BUILD TARGETS                                                       │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build build-all build-service clean clean-knative-lambda

build-all: ## 🔨 Build all binaries inside ./bin
	$(call print_status,🔨 Building all binaries...)
	@mkdir -p $(BIN_DIR)
	@echo "Building main service binary..."
	@CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/service
	@echo "Building sidecar binary..."
	@if [ -f sidecar/main.go ]; then \
		CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN_DIR)/$(SIDECAR_BINARY) ./sidecar 2>/dev/null || echo "⚠️ Sidecar build failed, skipping"; \
	elif [ -f sidecar/cmd/main.go ]; then \
		CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN_DIR)/$(SIDECAR_BINARY) ./sidecar/cmd 2>/dev/null || echo "⚠️ Sidecar build failed, skipping"; \
	else \
		echo "⚠️ Sidecar source not found, skipping sidecar build"; \
	fi
	$(call print_success,All binaries built in $(BIN_DIR)/)

clean: ## 🧹 Clean all build artifacts
	$(call print_status,🧹 Cleaning build artifacts...)
	@rm -rf $(BUILD_DIR) $(BIN_DIR)
	@rm -f coverage.out
	@go clean -cache
	@go clean -testcache

clean-knative-lambda: ## 🧹 Clean Knative Lambda resources
	$(call print_status,🧹 Cleaning Knative Lambda resources...)
	@kubectl delete kpa --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete revision --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete service --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete route --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete configuration --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete job --all -n $(NAMESPACE) 2>/dev/null || true
	@kubectl delete trigger --all -n $(NAMESPACE) 2>/dev/null || true
	$(call print_success,Knative Lambda resources cleaned)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧪 LINTING TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: lint fmt

lint: ## 🔍 Run linting
	$(call print_status,🔍 Running linting...)
	$(call check_command,golangci-lint,go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@timeout 60s golangci-lint run --timeout=5m ./... || \
		(echo "$(COLOR_YELLOW)⚠️ Full linting failed, running basic checks...$(COLOR_RESET)" && \
		 go vet ./... && \
		 gofmt -l . | (! grep .) && \
		 echo "$(COLOR_GREEN)✅ Basic linting completed$(COLOR_RESET)")
	$(call print_success,Linting completed)

fmt: ## 🎨 Format code
	$(call print_status,🎨 Formatting code...)
	@go fmt ./...
	$(call print_success,Code formatted)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🐳 DOCKER TARGETS                                                      │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: docker-build docker-push docker-login docker-build-service docker-build-sidecar docker-build-metrics-pusher docker-push-service docker-push-sidecar docker-push-metrics-pusher metrics-pusher docker-build-local docker-push-local

docker-build: docker-build-service docker-build-sidecar docker-build-metrics-pusher ## 🐳 Build all Docker images (service + sidecar + metrics-pusher)

docker-build-service: ## 🐳 Build service Docker image
	$(call print_status,🐳 Building service Docker image...)
	@docker build -t $(DOCKER_REPO):$(DOCKER_TAG) -t $(DOCKER_REPO):latest .
	$(call print_success,Service image built: $(DOCKER_REPO):$(DOCKER_TAG))

docker-build-sidecar: ## 🐳 Build sidecar Docker image
	$(call print_status,🐳 Building sidecar Docker image...)
	@if [ -f sidecar/Dockerfile ]; then \
		docker build -t $(SIDECAR_DOCKER_REPO):$(SIDECAR_DOCKER_TAG) -t $(SIDECAR_DOCKER_REPO):latest -f sidecar/Dockerfile .; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Sidecar image built: $(SIDECAR_DOCKER_REPO):$(SIDECAR_DOCKER_TAG)$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Sidecar Dockerfile not found, skipping sidecar build$(COLOR_RESET)"; \
	fi

docker-build-metrics-pusher: ## 🐳 Build metrics pusher Docker image
	$(call print_status,🐳 Building metrics pusher Docker image...)
	@if [ -f metrics-pusher/Dockerfile ]; then \
		docker build -t $(METRICS_PUSHER_DOCKER_REPO):$(METRICS_PUSHER_DOCKER_TAG) -t $(METRICS_PUSHER_DOCKER_REPO):latest -f metrics-pusher/Dockerfile metrics-pusher/; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Metrics pusher image built: $(METRICS_PUSHER_DOCKER_REPO):$(METRICS_PUSHER_DOCKER_TAG)$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Metrics pusher Dockerfile not found, skipping metrics pusher build$(COLOR_RESET)"; \
	fi

docker-push: docker-push-service docker-push-sidecar docker-push-metrics-pusher ## 🐳 Push all Docker images (service + sidecar + metrics-pusher)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🚀 BUILD AND PUSH ALL IMAGES                                          │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: build-and-push-all build-and-push-all-dev build-and-push-all-prd build-and-push-all-local

build-and-push-all: validate-env validate-docker validate-aws docker-login docker-build docker-push ## 🚀 Build and push all images (service + sidecar + metrics-pusher)

build-and-push-all-dev: ## 🚀 Build and push all images for dev environment
	$(call print_status,🚀 Building and pushing all images for dev environment...)
	@$(MAKE) ENV=dev build-and-push-all

build-and-push-all-prd: ## 🚀 Build and push all images for prd environment
	$(call print_status,🚀 Building and pushing all images for prd environment...)
	@$(MAKE) ENV=prd build-and-push-all

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔍 VALIDATION FUNCTIONS                                                │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: validate-env validate-docker validate-aws

docker-push-service: docker-build-service ## 🐳 Push service Docker image
	$(call print_status,🐳 Pushing service Docker image...)
	@docker push $(DOCKER_REPO):$(DOCKER_TAG)
	@docker push $(DOCKER_REPO):latest
	$(call print_success,Service image pushed: $(DOCKER_REPO):$(DOCKER_TAG))

docker-push-sidecar: docker-build-sidecar ## 🐳 Push sidecar Docker image
	$(call print_status,🐳 Pushing sidecar Docker image...)
	@if [ -f sidecar/Dockerfile ]; then \
		docker push $(SIDECAR_DOCKER_REPO):$(SIDECAR_DOCKER_TAG); \
		docker push $(SIDECAR_DOCKER_REPO):latest; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Sidecar image pushed: $(SIDECAR_DOCKER_REPO):$(SIDECAR_DOCKER_TAG)$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Sidecar Dockerfile not found, skipping sidecar push$(COLOR_RESET)"; \
	fi

docker-push-metrics-pusher: docker-build-metrics-pusher ## 🐳 Push metrics pusher Docker image
	$(call print_status,🐳 Pushing metrics pusher Docker image...)
	@if [ -f metrics-pusher/Dockerfile ]; then \
		docker push $(METRICS_PUSHER_DOCKER_REPO):$(METRICS_PUSHER_DOCKER_TAG); \
		docker push $(METRICS_PUSHER_DOCKER_REPO):latest; \
		echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Metrics pusher image pushed: $(METRICS_PUSHER_DOCKER_REPO):$(METRICS_PUSHER_DOCKER_TAG)$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ Metrics pusher Dockerfile not found, skipping metrics pusher push$(COLOR_RESET)"; \
	fi

metrics-pusher: docker-build-metrics-pusher docker-push-metrics-pusher ## 🐳 Build and push metrics pusher (convenience target)

test-github-token: ## 🔑 Test GitHub tokens from .zshrc
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)🔑 Testing GitHub Tokens$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Testing GITHUB_TOKEN...$(COLOR_RESET)"
	@if [ -n "$$GITHUB_TOKEN" ]; then \
		echo "$$GITHUB_TOKEN" | docker login ghcr.io -u brunovlucena --password-stdin && \
		echo "$(COLOR_GREEN)✅ GITHUB_TOKEN works!$(COLOR_RESET)" || \
		echo "$(COLOR_RED)❌ GITHUB_TOKEN failed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ GITHUB_TOKEN not set$(COLOR_RESET)"; \
	fi
	@echo ""
	@echo "$(COLOR_BOLD)Testing GITHUB_TOKEN_BRUNO...$(COLOR_RESET)"
	@if [ -n "$$GITHUB_TOKEN_BRUNO" ]; then \
		echo "$$GITHUB_TOKEN_BRUNO" | docker login ghcr.io -u brunovlucena --password-stdin && \
		echo "$(COLOR_GREEN)✅ GITHUB_TOKEN_BRUNO works!$(COLOR_RESET)" && \
		echo "$(COLOR_YELLOW)💡 Consider: export GITHUB_TOKEN=\$$GITHUB_TOKEN_BRUNO$(COLOR_RESET)" || \
		echo "$(COLOR_RED)❌ GITHUB_TOKEN_BRUNO failed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ GITHUB_TOKEN_BRUNO not set$(COLOR_RESET)"; \
	fi
	@echo ""
	@echo "$(COLOR_BOLD)Testing GITHUB_PERSONAL_ACCESS_TOKEN...$(COLOR_RESET)"
	@if [ -n "$$GITHUB_PERSONAL_ACCESS_TOKEN" ]; then \
		echo "$$GITHUB_PERSONAL_ACCESS_TOKEN" | docker login ghcr.io -u brunovlucena --password-stdin && \
		echo "$(COLOR_GREEN)✅ GITHUB_PERSONAL_ACCESS_TOKEN works!$(COLOR_RESET)" && \
		echo "$(COLOR_YELLOW)💡 Consider: export GITHUB_TOKEN=\$$GITHUB_PERSONAL_ACCESS_TOKEN$(COLOR_RESET)" || \
		echo "$(COLOR_RED)❌ GITHUB_PERSONAL_ACCESS_TOKEN failed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ GITHUB_PERSONAL_ACCESS_TOKEN not set$(COLOR_RESET)"; \
	fi
	@echo ""
	@echo "$(COLOR_YELLOW)💡 To use a working token, run:$(COLOR_RESET)"
	@echo "  export GITHUB_TOKEN=<working_token_value>"

docker-login: ## 🔐 Login to GitHub Container Registry
	$(call print_status,🔐 Logging into GitHub Container Registry...)
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "$(COLOR_RED)❌ ERROR: GITHUB_TOKEN is not set$(COLOR_RESET)"; \
		echo ""; \
		echo "$(COLOR_YELLOW)Please run 'make test-github-token' to test your tokens$(COLOR_RESET)"; \
		echo ""; \
		echo "$(COLOR_YELLOW)Or set your GitHub token:$(COLOR_RESET)"; \
		echo "  export GITHUB_TOKEN=ghp_your_token_here"; \
		echo ""; \
		echo "$(COLOR_YELLOW)Generate a new token at:$(COLOR_RESET)"; \
		echo "  https://github.com/settings/tokens/new"; \
		echo "  Required scopes: write:packages, read:packages"; \
		exit 1; \
	fi
	@echo "$$GITHUB_TOKEN" | docker login ghcr.io -u brunovlucena --password-stdin || \
		(echo "$(COLOR_RED)❌ Docker login failed. Token may be expired or lack permissions.$(COLOR_RESET)" && \
		 echo "$(COLOR_YELLOW)Run 'make test-github-token' to test all available tokens$(COLOR_RESET)" && \
		 echo "$(COLOR_YELLOW)Or generate a new token with write:packages scope at:$(COLOR_RESET)" && \
		 echo "  https://github.com/settings/tokens/new" && \
		 exit 1)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ☸️ KUBERNETES TARGETS                                                  │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: status logs pf-argo pf-rabbitmq pf-rabbitmq-admin pf-broker debug-service fix-service trigger-service

pf-rabbitmq: ## 🔌 Port forward RabbitMQ to 0.0.0.0:5672

	$(call print_status,🔌 Port forwarding RabbitMQ to 0.0.0.0:5672...)
	@kubectl port-forward --address 0.0.0.0 svc/rabbitmq-cluster-$(ENV) 5672:5672 -n rabbitmq-$(ENV)

pf-rabbitmq-admin: ## 🔌 Port forward RabbitMQ Management UI to 0.0.0.0:15672
	$(call print_status,🔌 Port forwarding RabbitMQ Management UI to 0.0.0.0:15672...)
	@kubectl port-forward --address 0.0.0.0 svc/rabbitmq-cluster-$(ENV) 15672:15672 -n rabbitmq-$(ENV)

pf-broker: ## 🔌 Port forward Knative broker to 0.0.0.0:8081
	$(call print_status,🔌 Port forwarding Knative broker to 0.0.0.0:8081...)
	@kubectl port-forward --address 0.0.0.0 svc/knative-lambda-broker-$(ENV)-broker-ingress 8081:80 -n knative-lambda-$(ENV)

pf-argocd: ## 🔌 Port forward ArgoCD to 0.0.0.0:8000
	$(call print_status,🔌 Port forwarding ArgoCD to 0.0.0.0:8000...)
	@kubectl port-forward --address 0.0.0.0 svc/argocd-server 8000:443 -n argocd

pf-prometheus-mcp: ## 🔌 Port forward prometheus-mcp to 0.0.0.0:8001
	$(call print_status,🔌 Port forwarding prometheus-mcp to 0.0.0.0:8001...)
	@kubectl port-forward --address 0.0.0.0 svc/prometheus-mcp 8001:8000 -n prometheus-mcp

pf-tempo-mcp: ## 🔌 Port forward tempo-mcp to 0.0.0.0:8002
	$(call print_status,🔌 Port forwarding tempo-mcp to 0.0.0.0:8002...)
	@kubectl port-forward --address 0.0.0.0 svc/tempo-mcp 8002:8000 -n tempo-mcp

pf-prometheus: ## 🔌 Port forward prometheus to 0.0.0.0:9090
	$(call print_status,🔌 Port forwarding prometheus to 0.0.0.0:9090...)
	@kubectl port-forward --address 0.0.0.0 svc/prometheus-kube-prometheus-prometheus 9090:9090 -n prometheus

pf-registry: ## 🔌 Port forward registry to 0.0.0.0:5000
	@echo "🔌 Port forwarding registry to 0.0.0.0:5000..."
	kubectl port-forward --address 0.0.0.0 -n registry service/registry 5000:5000

rabbitmq-status: ## 📊 Print RabbitMQ queue status and statistics
	$(call print_status,📊 Checking RabbitMQ queue message counts...)
	@echo "$(COLOR_BOLD)🔍 RabbitMQ Queue Messages for $(ENV) environment$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)📦 Finding RabbitMQ pods...$(COLOR_RESET)"
	@RABBITMQ_POD=$$(kubectl get pods -l app.kubernetes.io/name=rabbitmq-cluster-$(ENV) -n rabbitmq-$(ENV) -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -n "$$RABBITMQ_POD" ]; then \
		echo "✅ Found RabbitMQ pod: $$RABBITMQ_POD"; \
		echo ""; \
		echo "$(COLOR_BOLD)📊 Queue Message Counts:$(COLOR_RESET)"; \
		echo "Queue Name | Messages Ready | Messages Unacknowledged | Total Messages | Consumers"; \
		echo "-----------|---------------|-------------------------|----------------|----------"; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl list_queues name messages_ready messages_unacknowledged messages consumers 2>/dev/null | tail -n +2 | while IFS='\t' read -r name ready unack total consumers; do \
			echo "$$name | $$ready | $$unack | $$total | $$consumers"; \
		done || echo "$(COLOR_YELLOW)⚠️ Could not retrieve queue information$(COLOR_RESET)"; \
		echo ""; \
		echo "$(COLOR_BOLD)📈 Summary:$(COLOR_RESET)"; \
		TOTAL_READY=$$(kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl list_queues messages_ready 2>/dev/null | tail -n +2 | awk '{sum += $$1} END {print sum+0}'); \
		TOTAL_UNACK=$$(kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl list_queues messages_unacknowledged 2>/dev/null | tail -n +2 | awk '{sum += $$1} END {print sum+0}'); \
		TOTAL_MESSAGES=$$(kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl list_queues messages 2>/dev/null | tail -n +2 | awk '{sum += $$1} END {print sum+0}'); \
		echo "Total Messages Ready: $$TOTAL_READY"; \
		echo "Total Messages Unacknowledged: $$TOTAL_UNACK"; \
		echo "Total Messages in All Queues: $$TOTAL_MESSAGES"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ No RabbitMQ pods found in rabbitmq-$(ENV) namespace$(COLOR_RESET)"; \
		echo ""; \
		echo "$(COLOR_BOLD)🔍 Troubleshooting:$(COLOR_RESET)"; \
		echo "1. Check if RabbitMQ is deployed: kubectl get rabbitmqcluster -n rabbitmq-$(ENV)"; \
		echo "2. Check pod status: kubectl get pods -n rabbitmq-$(ENV)"; \
		echo "3. Check if namespace exists: kubectl get namespace rabbitmq-$(ENV)"; \
	fi

rabbitmq-purge-lambda-queues: ## 🧹 Purge lambda event queues for environment (uses ENV variable)
	$(call print_status,🧹 Purging lambda event queues for $(ENV) environment...)
	@echo "$(COLOR_BOLD)🔍 Finding RabbitMQ pods...$(COLOR_RESET)"
	@RABBITMQ_POD=$$(kubectl get pods -l app.kubernetes.io/name=rabbitmq-cluster-$(ENV) -n rabbitmq-$(ENV) -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -n "$$RABBITMQ_POD" ]; then \
		echo "✅ Found RabbitMQ pod: $$RABBITMQ_POD"; \
		echo ""; \
		echo "$(COLOR_BOLD)🧹 Purging lambda event queues...$(COLOR_RESET)"; \
		echo "Purging queue: lambda-build-events-$(ENV)"; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl purge_queue lambda-build-events-$(ENV) 2>/dev/null || echo "$(COLOR_YELLOW)⚠️ Queue lambda-build-events-$(ENV) not found or already empty$(COLOR_RESET)"; \
		echo "Purging queue: lambda-service-events-$(ENV)"; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl purge_queue lambda-service-events-$(ENV) 2>/dev/null || echo "$(COLOR_YELLOW)⚠️ Queue lambda-service-events-$(ENV) not found or already empty$(COLOR_RESET)"; \
		echo "Purging queue: parser-results-$(ENV)"; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl purge_queue parser-results-$(ENV) 2>/dev/null || echo "$(COLOR_YELLOW)⚠️ Queue parser-results-$(ENV) not found or already empty$(COLOR_RESET)"; \
		echo ""; \
		echo "$(COLOR_BOLD)📊 Checking queue status after purge...$(COLOR_RESET)"; \
		kubectl exec -n rabbitmq-$(ENV) $$RABBITMQ_POD -- rabbitmqctl list_queues name messages 2>/dev/null | grep -E "(lambda-build-events-$(ENV)|lambda-service-events-$(ENV)|parser-results-$(ENV))" || echo "$(COLOR_GREEN)✅ All lambda event queues are empty$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_RED)❌ No RabbitMQ pods found in rabbitmq-$(ENV) namespace$(COLOR_RESET)"; \
		echo ""; \
		echo "$(COLOR_BOLD)🔍 Troubleshooting:$(COLOR_RESET)"; \
		echo "1. Check if RabbitMQ is deployed: kubectl get rabbitmqcluster -n rabbitmq-$(ENV)"; \
		echo "2. Check pod status: kubectl get pods -n rabbitmq-$(ENV)"; \
		echo "3. Check if namespace exists: kubectl get namespace rabbitmq-$(ENV)"; \
		exit 1; \
	fi
	$(call print_success,Lambda event queues purged for $(ENV) environment)

rabbitmq-purge-lambda-queues-dev: ## 🧹 Purge lambda event queues for dev environment
	$(call print_status,🧹 Purging lambda event queues for dev environment...)
	@$(MAKE) ENV=dev rabbitmq-purge-lambda-queues

rabbitmq-purge-lambda-queues-prd: ## 🧹 Purge lambda event queues for prd environment
	$(call print_status,🧹 Purging lambda event queues for prd environment...)
	@$(MAKE) ENV=prd rabbitmq-purge-lambda-queues

argocd-secret: ## 🤫 Show ArgoCD initial admin secret
	$(call print_status,🤫 Retrieving ArgoCD initial admin secret...)
	@echo "ArgoCD initial admin password:"
	@kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo

trigger-lambda: ## 🚀 Trigger a lambda parser event (uses ENV variable)
	@cd tests && ENV=$(ENV) uv run --python 3.9 python create-event-lambda.py

trigger-lambda-dev: ## 🚀 Trigger a lambda parser event for dev environment
	@cd tests && ENV=dev uv run --python 3.9 python create-event-lambda.py

trigger-lambda-prd: ## 🚀 Trigger a lambda parser event for prd environment
	@cd tests && ENV=prd uv run --python 3.9 python create-event-lambda.py

trigger-build-clean: clean-knative-lambda ## 🚀 Trigger a lambda build event (cleans up first, uses ENV variable)
	@cd tests && ENV=$(ENV) uv run --python 3.9 python create-event-builder.py

trigger-build: ## 🚀 Trigger a lambda build event (uses ENV variable)
	@cd tests && ENV=$(ENV) uv run --python 3.9 python create-event-builder.py

trigger-build-local: ## 🚀 Trigger a lambda build event for local environment
	@cd tests && ENV=local uv run --python 3.9 python create-event-builder.py

trigger-build-dev: ## 🚀 Trigger a lambda build event for dev environment
	@cd tests && ENV=dev uv run --python 3.9 python create-event-builder.py

trigger-build-prd: ## 🚀 Trigger a lambda build event for prd environment
	@cd tests && ENV=prd uv run --python 3.9 python create-event-builder.py

trigger-delete-dev: ## 🗑️ Trigger a lambda delete event for dev environment
	@cd tests && ENV=dev uv run --python 3.9 python create-event-delete.py

trigger-delete-prd: ## 🗑️ Trigger a lambda delete event for prd environment
	@cd tests && ENV=prd uv run --python 3.9 python create-event-delete.py

trigger-delete-local: ## 🗑️ Trigger a lambda delete event for local environment
	@cd tests && ENV=local uv run --python 3.9 python create-event-delete.py

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🛠️ DEVELOPMENT TARGETS                                                │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: run dev setup deps-check upload-parser upload-parser-dev upload-parser-prd

upload-parser: ## 🚀 Upload parser file to S3 (uses PARSER_ID variable)
	$(call print_status,🚀 Uploading parser to S3...)
	@if [ -z "$(PARSER_ID)" ]; then \
		echo "$(COLOR_RED)❌ PARSER_ID variable is required. Use: make upload-parser PARSER_ID=0197ad6c10b973b2b854a0e652155b7e$(COLOR_RESET)"; \
		exit 1; \
	fi
	@if [ ! -f "tests/$(PARSER_ID)" ]; then \
		echo "$(COLOR_RED)❌ Parser file not found: tests/$(PARSER_ID)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)📤 Uploading parser $(PARSER_ID) to S3...$(COLOR_RESET)"
	@aws s3 cp tests/$(PARSER_ID) s3://knative-lambda-$(ENV)-fusion-modules-tmp/global/parser/$(PARSER_ID)
	@echo "$(COLOR_GREEN)✅ Parser uploaded successfully to s3://knative-lambda-$(ENV)-fusion-modules-tmp/global/parser/$(PARSER_ID)$(COLOR_RESET)"

upload-parser-dev: ## 🚀 Upload parser file to dev S3 (uses PARSER_ID variable)
	$(call print_status,🚀 Uploading parser to dev S3...)
	@$(MAKE) ENV=dev upload-parser

upload-parser-prd: ## 🚀 Upload parser file to prd S3 (uses PARSER_ID variable)
	$(call print_status,🚀 Uploading parser to prd S3...)
	@$(MAKE) ENV=prd upload-parser

run: build ## 🛠️ Run the service locally on port 8080
	$(call print_status,🛠️ Running service locally on port 8080...)
	@echo "$(COLOR_YELLOW)💡 Service will be available at http://localhost:8080$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)💡 Press Ctrl+C to stop the service$(COLOR_RESET)"
	@PORT=8080 \
	AWS_REGION=us-west-2 \
	ECR_REGISTRY=$(DOCKER_REGISTRY) \
	KANIKO_NAMESPACE=default \
	KANIKO_IMAGE=gcr.io/kaniko-project/executor:latest \
	LOG_LEVEL=info \
	./$(BIN_DIR)/$(BINARY_NAME)

setup: ## 🔧 Setup development environment
	$(call print_status,🔧 Setting up development environment...)
	@go mod download
	@go mod tidy
	$(call print_success,Development environment ready)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧪 TESTING TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: test-api test-build test-rabbitmq test-deps test-unit test-integration test-benchmark test-chaos test-coverage test-quality test preflight

test: lint test-unit ## 🧪 Run all tests (linting + unit tests)
	$(call print_status,🧪 Running comprehensive test suite...)
	$(call print_success,All tests completed successfully)

test-unit: ## 🧪 Run Go unit tests
	$(call print_status,🧪 Running Go unit tests...)
	@go test -race -timeout=30s ./internal/aws/... ./internal/config/... ./internal/constants/... ./internal/errors/... ./internal/handler/... ./internal/observability/... ./internal/resilience/... ./internal/security/... ./internal/storage/... ./internal/templates/... 2>&1 | grep -v "LC_DYSYMTAB" | grep -v "^ld: warning" | grep -v "^# knative-lambda-new.*\.test$$"
	$(call print_success,Unit tests completed)

test-k6: ## 🚀 Run k6 load tests
	$(call print_status,🚀 Running k6 load tests...)
	@if [ ! -f tests/k6/builder-tests.yaml ]; then \
		echo "$(COLOR_RED)❌ k6 test files not found$(COLOR_RESET)"; \
		exit 1; \
	fi
	@which k6 > /dev/null 2>&1 || (echo "$(COLOR_YELLOW)⚠️ k6 not installed, installing...$(COLOR_RESET)" && make install-k6)
	@echo "$(COLOR_BLUE)📋 Running builder tests...$(COLOR_RESET)"
	@k6 run tests/k6/builder-tests.yaml --quiet || echo "$(COLOR_YELLOW)⚠️ Builder tests completed with warnings$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)📋 Running lambda tests...$(COLOR_RESET)"
	@k6 run tests/k6/lambda-tests.yaml --quiet || echo "$(COLOR_YELLOW)⚠️ Lambda tests completed with warnings$(COLOR_RESET)"
	$(call print_success,k6 load tests completed)

install-k6: ## 📦 Install k6 load testing tool
	$(call print_status,📦 Installing k6...)
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install k6 2>/dev/null || echo "$(COLOR_YELLOW)⚠️ Homebrew not available$(COLOR_RESET)"; \
	else \
		curl -fsSL https://github.com/grafana/k6/releases/download/v0.48.0/k6-v0.48.0-linux-amd64.tar.gz | \
		sudo tar -xz --strip-components=1 -C /usr/local/bin k6-v0.48.0-linux-amd64/k6; \
	fi
	@k6 version
	$(call print_success,k6 installed)

install-test-deps: ## 📦 Install all test dependencies
	$(call print_status,📦 Installing test dependencies...)
	@echo "$(COLOR_BLUE)Installing Go test dependencies...$(COLOR_RESET)"
	@go get -t ./...
	@go install github.com/stretchr/testify@latest
	@echo "$(COLOR_BLUE)Installing k6...$(COLOR_RESET)"
	@make install-k6
	$(call print_success,Test dependencies installed)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  ℹ️ INFO TARGETS                                                        │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: info

info: ## ℹ️ Show build information
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)ℹ️ Build Information$(COLOR_RESET)"
	@echo "Project:         $(PROJECT_NAME)"
	@echo "Service:         $(SERVICE_NAME)"
	@echo "Version:         $(VERSION)"
	@echo "Environment:     $(ENV)"
	@echo "Namespace:       $(NAMESPACE)"
	@echo ""
	@echo "$(COLOR_BOLD)🐳 Docker Images:$(COLOR_RESET)"
	@echo "Service Repo:    $(DOCKER_REPO)"
	@echo "Service Tag:     $(DOCKER_TAG)"
	@echo "Sidecar Repo:    $(SIDECAR_DOCKER_REPO)"
	@echo "Sidecar Tag:     $(SIDECAR_DOCKER_TAG)"
	@echo "Metrics Pusher:  $(METRICS_PUSHER_DOCKER_REPO)"
	@echo "Metrics Tag:     $(METRICS_PUSHER_DOCKER_TAG)"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🏷️ VERSIONING TARGETS                                                  │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: version version-info version-bump version-bump-major version-bump-minor version-bump-patch release-notes

version: version-info ## 🏷️ Show current version (alias for version-info)

version-info: ## 🏷️ Show detailed version information
	@chmod +x scripts/version-manager.sh 2>/dev/null || true
	@./scripts/version-manager.sh info

version-bump: ## 🏷️ Bump version (usage: make version-bump VERSION=1.0.0)
	@if [ -z "$(VERSION_NEW)" ]; then \
		echo "$(COLOR_RED)❌ ERROR: VERSION_NEW is required$(COLOR_RESET)"; \
		echo "Usage: make version-bump VERSION_NEW=1.0.0"; \
		exit 1; \
	fi
	@chmod +x scripts/version-manager.sh 2>/dev/null || true
	@./scripts/version-manager.sh bump $(VERSION_NEW)
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ Version bumped to $(VERSION_NEW) successfully!$(COLOR_RESET)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review changes: git diff"
	@echo "  2. Commit changes: git add . && git commit -m 'chore: bump version to $(VERSION_NEW)'"
	@echo "  3. Create tag: git tag -a knative-lambda-v$(VERSION_NEW) -m 'Release $(VERSION_NEW)'"
	@echo "  4. Push: git push origin $$(git rev-parse --abbrev-ref HEAD) --tags"

version-bump-major: ## 🏷️ Auto-bump major version (1.0.0 → 2.0.0)
	@chmod +x scripts/version-manager.sh 2>/dev/null || true
	@./scripts/version-manager.sh bump-major

version-bump-minor: ## 🏷️ Auto-bump minor version (1.0.0 → 1.1.0)
	@chmod +x scripts/version-manager.sh 2>/dev/null || true
	@./scripts/version-manager.sh bump-minor

version-bump-patch: ## 🏷️ Auto-bump patch version (1.0.0 → 1.0.1)
	@chmod +x scripts/version-manager.sh 2>/dev/null || true
	@./scripts/version-manager.sh bump-patch

release-notes: ## 🏷️ Generate release notes from git commits
	@chmod +x scripts/version-manager.sh 2>/dev/null || true
	@./scripts/version-manager.sh release-notes 