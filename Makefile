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

# Operator
operator-build: ## Builds operator (E.g. make operator-build).
	 @./apps/my-k8s-operator/helper.sh operator-build

operator-deploy: ## Deploys operator (E.g. make operator-deploy).
	 @./apps/my-k8s-operator/helper.sh operator-deploy

test-operator: ## Tests operator (E.g. make test-operator).
	 @./apps/my-k8s-operator/helper.sh test-operator

# App
run-myapp: ## Runs app example (E.g. make run-myapp).
	 @./apps/app-example/helper.sh run-myapp

mod-tidy-myapp: ## Runs app example (E.g. make mod-tidy-myapp).
	 @./apps/app-example/helper.sh mod-tidy-myapp

build-myapp: ## Builds app example (E.g. make build-myapp).
	 @./apps/app-example/helper.sh build-myapp

image-build-myapp: ## Builds image for app example (E.g. make image-build-push-myapp).
	 @./apps/app-example/helper.sh image-build-push-myapp stable

# Dev
update: ## updates everything. 
	@./apps/app-example/helper.sh image-build-push-myapp dev
	@./apps/app-example/helper.sh run-test-deployment
