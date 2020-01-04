.PHONY: help

ARGS = $(filter-out $(firstword $(MAKECMDGOALS)), $(MAKECMDGOALS))

# minikube variables
MINIKUBE_VERSION 	= v1.5.2
CLUSTER_CPUS 		= 6
CLUSTER_MEMORY 		= 4096mb
CLUSTER_DISK 		= 20GB
CLUSTER_DISK_EXTRA  = 15GB
CLUSTER_VERSION 	= v1.17.0
#VM_DRIVER           = virtualbox 
# kind variables
KIND_VERSION		= v0.6.1
VM_DRIVER			= none
# other
CLUSTER_NAME		= homelab

# tools
K9S_VERSION 		= 0.10.8
KUBECTL_VERSION 	= v1.17.0
HELM_VERSION 		= v3.0.1
SQUASH_VERSION 		= v0.5.18
SONOBUOY_VERSION	= 0.16.1
GO_VERSION			= 1.13.5

help: ## helper. 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Cluster
pre-install: ## pre-installs all nescessary tools to bootstrap and manage cluster.
	@./helper.sh pre-install go ${GO_VERSION}
	@./helper.sh pre-install minikube ${MINIKUBE_VERSION}
	@./helper.sh pre-install kind ${KIND_VERSION}
	@./helper.sh pre-install k9s ${K9S_VERSION}
	@./helper.sh pre-install helm ${HELM_VERSION}
	@./helper.sh pre-install kubectl ${KUBECTL_VERSION}
	@./helper.sh pre-install squash ${SQUASH_VERSION}
	@./helper.sh pre-install kubediff

bootstrap-cluster: ## Bootstraps a kubernetes cluster.
	@./helper.sh bootstrap-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_DISK_EXTRA} ${CLUSTER_VERSION} ${CLUSTER_NAME} ${VM_DRIVER}

start-cluster-minikube: ## Starts using minikube cluster.
	@./helper.sh start-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_DISK_EXTRA} ${CLUSTER_VERSION} ${CLUSTER_NAME} ${VM_DRIVER}

start-cluster-kind: ## Starts cluster using Kind
	@docker start ${CLUSTER_NAME}-control-plane
	@kind export kubeconfig --name ${CLUSTER_NAME}

stop-cluster: ## Stops cluster.
	@./helper.sh stop-cluster ${CLUSTER_NAME} ${VM_DRIVER}

tunnel: ## Creates a tunnel to minikube's registry (E.g. make tunnel service=registry).
	@./helper.sh tunnel ${ARGS}

add: ## Adds components to the cluster.
	@./helper.sh add-kube-system
	@./helper.sh add-monitoring
	@./helper.sh add-storage
	@./helper.sh add-cicd
	@./helper.sh add-security
	@./helper.sh add-testing

# Security
sniff: ## Sniffs comunication (E.g. make sniff)
	@./other.sh sniff

check-pod-security: ## outputs infomation about the cluster
	@./other.sh check-pod-security

kube-bench: ##
	@./other.sh kube-bench

kube-diff: ##
	@./other.sh kube-diff
