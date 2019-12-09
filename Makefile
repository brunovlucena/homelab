.PHONY: help

ARGS = $(filter-out $(firstword $(MAKECMDGOALS)), $(MAKECMDGOALS))

# Tool Variables
MINIKUBE_VERSION = v1.5.2
OPERATOR_VERSION = v0.12.0
# Cluster Specification
CLUSTER_CPUS = 4
CLUSTER_MEMORY = 8192mb
CLUSTER_DISK = 20GB
CLUSTER_VERSION = v1.16.2
CLUSTER_NAME = mobimeo

help: ## Help. 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Cluster
pre-install: ## Pre-Installs tools (E.g: $ make pre-install).
	@./helper.sh pre-install minikube ${MINIKUBE_VERSION}
	@./helper.sh pre-install operator-sdk ${OPERATOR_VERSION}
	@./helper.sh pre-install k6

bootstrap-cluster: ## Bootstraps cluster (E.g. make bootstrap).
	@./helper.sh bootstrap-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_VERSION} ${CLUSTER_NAME}

start-cluster: ## Starts cluster.
	@./helper.sh start-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_VERSION} ${CLUSTER_NAME}

stop-cluster: ## Stops cluster.
	@./helper.sh stop-cluster ${CLUSTER_NAME}

clean-cluster: ## Cleans Minikube (E.g. make clean-cluster).
	@./helper.sh clean-cluster ${CLUSTER_NAME}

tunnel-registry: ## Creates a tunnel to minikube's registry (E.g. make tunnel-registry).
	@./helper.sh tunnel-registry

helm-install: ## Installs components via helm charts.
	@./helper.sh helm-install

# MyAppOperator
bootstrap-operator: ## Builds operator (E.g. make bootstrap-operator).
	 @./apps/helper.sh bootstrap-operator

build-deploy-operator: ## Deploys operator (E.g. make build-deploy-operator).
	 @./apps/helper.sh build-deploy-operator stable

build-deploy-operator-test: ## Tests MyAppOperator (E.g. make build-deploy-test). 
	 @./apps/helper.sh build-deploy-operator dev
	 @./apps/helper.sh deploy-operator-test

# MyApp
run-myapp: ## Runs app example on host (E.g. make run-myapp).
	 @./apps/helper.sh run-myapp

run-postgres-local: ## Runs postgres on host (E.g. make run-postgres-local).
	 @./apps/helper.sh run-postgres-local

build-myapp: ## Builds binary app example (E.g. make build-myapp).
	 @./apps/helper.sh build-myapp

build-deploy-myapp: ## Builds image for app example (E.g. make build-push-myapp latest).
	 @./apps/helper.sh build-deploy-myapp stable

# Dev
skaffold: ## Uses skaffold during the development
	@./apps/helper.sh go-tidy
	@./apps/helper.sh skaffold

debug-myapp: ## Runs dlv (E.g. make debug-myapp).
	 @./apps/helper.sh debug-myapp

test: ## Run Go Tests
	 @./apps/helper.sh test

load-test: ## Run Go Tests
	 @./apps/helper.sh load-test

checks:
	@shellcheck helper.sh || true
	@shellcheck apps/helper.sh || true

# Security
sniff: ## Sniffs comunication (E.g. make sniff)
	@./security.sh sniff

check-pod-security: ## outputs infomation about the cluster
	@./security.sh check-pod-security

kube-bench: ##
	@./debugger.sh kube-bench

kube-diff: ##
	@./debugger.sh kube-diff

cluster-rights: ##
	@./debugger.sh cluster-rights

cluster-roles: ##
	@./debugger.sh cluster-roles
