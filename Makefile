.PHONY: help

ARGS = $(filter-out $(firstword $(MAKECMDGOALS)), $(MAKECMDGOALS))

# variables
MINIKUBE_VERSION 	= v1.7.3
CLUSTER_CPUS 		= 6
CLUSTER_MEMORY 		= 4096mb
CLUSTER_DISK 		= 20GB
CLUSTER_DISK_EXTRA  = 15GB
CLUSTER_VERSION 	= v1.17.0
VM_DRIVER           = none # virtualbox|kind
KIND_VERSION		= v0.6.1
CLUSTER_NAME		= mobimeo

# components
CNI					= cilium
MESH				= disabled
BASIC				= enabled
MONITORING			= enabled
STORAGE				= disabled
CICD				= disabled
SECURITY			= disabled
TESTING				= disabled
ROOK_CEPH			= disabled
BACKUP				= disabled

# external k8s
PROMETHEUS_VERSION		= v2.15.2
NODE_EXPORTER_VERSION 	= v0.18.1

# tools
K9S_VERSION 		= 0.10.8
KUBECTL_VERSION 	= v1.17.0
HELM_VERSION 		= v3.0.1
SQUASH_VERSION 		= v0.5.18
SONOBUOY_VERSION	= 0.16.1
LINKERD_VERSION		= 2.6.1
KREW_VERSION		= v0.3.3
SKAFFOLD_VERSION	= v1.1.0

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
	@./helper.sh pre-install skaffold ${SKAFFOLD_VERSION}
	@./helper.sh pre-install kubediff
	@./helper.sh pre-install linkerd ${LINKERD_VERSION}
	@./helper.sh pre-install krew ${KREW_VERSION}

add-host-monitoring: ## add local monitoring for the host
	@./helper.sh add-host-monitoring prometheus ${PROMETHEUS_VERSION} ${CLUSTER_NAME}
	@./helper.sh add-host-monitoring node-exporter ${NODE_EXPORTER_VERSION} ${CLUSTER_NAME}
	@./helper.sh add-host-monitoring docker-hub-exporter latest ${CLUSTER_NAME}
	@./helper.sh add-host-monitoring github-exporter latest ${CLUSTER_NAME}

kind-add-registry: ## add local docker registry for cluster.
	@./helper.sh kind-add-registry ${CLUSTER_NAME}

bootstrap-cluster: ## Bootstraps a kubernetes cluster.
	@./helper.sh bootstrap-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_DISK_EXTRA} ${CLUSTER_VERSION} ${CLUSTER_NAME} ${VM_DRIVER} ${CNI} ${MESH}

start-cluster-minikube: ## Starts using minikube cluster.
	@./helper.sh start-cluster ${CLUSTER_CPUS} ${CLUSTER_MEMORY} ${CLUSTER_DISK} ${CLUSTER_DISK_EXTRA} ${CLUSTER_VERSION} ${CLUSTER_NAME} ${VM_DRIVER}

start-cluster-kind: ## Starts cluster using Kind
	@docker start ${CLUSTER_NAME}-control-plane
	@kind export kubeconfig --name ${CLUSTER_NAME}

stop-cluster: ## Stops cluster.
	@./helper.sh stop-cluster ${CLUSTER_NAME} ${VM_DRIVER}

tunnel: ## Creates a tunnel to cluster.
	@./helper.sh tunnel ${ARGS}

add-components: ## adds new components via helm upgrade  
	@./helper.sh add-basic ${BASIC}
	@./helper.sh add-monitoring ${MONITORING}
	@./helper.sh add-storage ${STORAGE}
	@./helper.sh add-cicd ${CICD}
	@./helper.sh add-security ${SECURITY}
	@./helper.sh add-testing ${TESTING}
	@./helper.sh add-rook-ceph ${ROOK_CEPH}
	@./helper.sh add-backup ${BACKUP}

# Security
sniff: ## Sniffs comunication (E.g. make sniff)
	@./other.sh sniff

check-pod-security: ## outputs infomation about the cluster
	@./other.sh check-pod-security

kube-bench: ##
	@./other.sh kube-bench

kube-diff: ##
	@./other.sh kube-diff
