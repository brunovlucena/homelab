.PHONY: help

ARGS = $(filter-out $(firstword $(MAKECMDGOALS)), $(MAKECMDGOALS))

# Tool Variables
MINIKUBE_VERSION = v1.5.2
# Cluster Specification
CLUSTER_CPUS = 4
CLUSTER_MEMORY = 8192mb
CLUSTER_DISK = 20GB
CLUSTER_VERSION = v1.16.2
CLUSTER_NAME = mobimeo

help: ## Help. 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

pre-install: ## Pre-Installs tools (E.g: $ make pre-install).
	@./helper.sh pre-install minikube ${MINIKUBE_VERSION}

bootstrap-cluster: ## Bootstraps cluster (E.g. make bootstrap).
	@./helper.sh bootstrap-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_VERSION} ${CLUSTER_NAME}

clean-cluster: ## Cleans Minikube (E.g. make clean-cluster).
	@./helper.sh clean-cluster ${CLUSTER_NAME}

operator-build: ## Builds operator (E.g. make operator-build).
	 @./apps/my-k8s-operator/helper.sh operator-build

operator-deploy: ## Deploys operator (E.g. make operator-deploy).
	 @./apps/my-k8s-operator/helper.sh operator-deploy

test-operator: ## Tests operator (E.g. make test-operator).
	 @./apps/my-k8s-operator/helper.sh test-operator
