#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
MINIKUBE=~/.local/bin/minikube
OPERATOR=~/.local/bin/operator-sdk

# pre-install some basic components.
#
# Usage:
#  $ ./helper.sh param1 [param2]
# * param1: pre-install
# * param2: [minikube]
pre_install() {
    local ARG="$1"
    local VERSION="$2"
    case "$ARG" in
        minikube)
	        wget "https://github.com/kubernetes/minikube/releases/download/$VERSION/minikube-linux-amd64" -O "$MINIKUBE"
            chmod +x "$MINIKUBE"
        ;;
        operator-sdk)
            wget "https://github.com/operator-framework/operator-sdk/releases/download/$VERSION/operator-sdk-$VERSION-x86_64-linux-gnu" -O "$OPERATOR"
            chmod +x "$OPERATOR"
        ;;
    esac
}

# bootstraps a kubernetes cluster using  minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2] [param3] [param4] [param5]
# * param1: start-cluster
# * param2: cpus used
# * param3: memory used
# * param4: kubernetes version
# * param5: cluster name
bootstrap_cluster() {
    local CLUSTER_CPUS="$1"
    local CLUSTER_MEMORY="$2"
    local CLUSTER_DISK="$3"
    local CLUSTER_VERSION="$4"
    local CLUSTER_NAME="$5"
    # clean before start for the first time
    clean_cluster "$CLUSTER_NAME"
	# start cluster
	start_cluster "$CLUSTER_CPUS" "$CLUSTER_MEMORY" "$CLUSTER_DISK" "$CLUSTER_VERSION" "$CLUSTER_NAME"
    # add a second disk for ceph
    local DISK_SIZE=20000
    add_disk "$CLUSTER_NAME" "$DISK_SIZE"
    # manage pluggins
	manage_cluster_pluggins
}

# removes cluster from minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2]
# * param1: clean-cluster
# * param2: [cluster_name]
clean_cluster() {
    local CLUSTER_NAME="$1"
    $MINIKUBE stop "$CLUSTER_NAME" || true
    rm -r ~/.minikube/profiles/"$CLUSTER_NAME" || true
    rm -r ~/.minikube/machines/"$CLUSTER_NAME" || true
}

# stops a kubernetes cluster using minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2]
# * param1: stop-cluster
# * param2: [cluster_name]
stop_cluster() {
    local CLUSTER_NAME="$1"
	$MINIKUBE stop -p "$CLUSTER_NAME"
}

# starts a kubernetes cluster using minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2] [param3] [param4] [param5]
# * param1: start-cluster
# * param2: cpus used
# * param3: memory used
# * param4: kubernetes version
# * param5: cluster name
start_cluster() {
    local CLUSTER_CPUS="$1"
    local CLUSTER_MEMORY="$2"
    local CLUSTER_DISK="$3"
    local CLUSTER_VERSION="$4"
    local CLUSTER_NAME="$5"
	$MINIKUBE start --cpus="$CLUSTER_CPUS" --memory="$CLUSTER_MEMORY" --disk-size="$CLUSTER_DISK" --kubernetes-version="$CLUSTER_VERSION" -p "$CLUSTER_NAME"
}

# adds a second disk to minikube (to be used by ceph).
#
# Usage:
#  $ ./helper.sh param1 [param2]
# * param1: add_disk
# * param2: [cluster_name]
add_disk(){
    local CLUSTER_NAME="$1"
    local DISK_SIZE"$2"
	local VOLUME_PATH=~/.minikube/disks/rook-ceph-1.vdi
	VBoxManage createhd --filename "$VOLUME_PATH" --size "$DISK_SIZE" || true
	VBoxManage storageattach "$CLUSTER_NAME" \
                         --storagectl "SATA" \
                         --device 0 \
                         --port 2 \
                         --type hdd \
                         --medium "$VOLUME_PATH"
}

# manages minikube's cluster plugins.
#
# Usage:
#  $ ./helper.sh param1
# * param1: manage_cluster_pluggins
manage_cluster_pluggins() {
	# disable
	minikube addons disable helm-tiller # Helm 3.
    minikube addons disable registry-creds # Using local registry only.
	# enable
    minikube addons enable dashboard # Because dashboards are nice.
}

# creates a tunnel to registry.
#
# Usage:
#  $ ./helper.sh param1
# * param1: tunnel-registry
tunnel_registry() {
	kubectl port-forward $(kubectl get pod -l actual-registry=true -o jsonpath='{.items[0].metadata.name}' -n kube-system) 5000:5000 -n kube-system &
}

# installs prometheus-operator.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install-prometheus-operator
helm_install_prometheus_operator() {
    NAMESPACE=kube-system
    helm upgrade --install \
		prometheus-operator infra/charts/prometheus-operator -n "$NAMESPACE"
}

# installs kube-state-metrics.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install-kube-state-metrics
helm_install_kube_state_metrics() {
    NAMESPACE=kube-system
    helm upgrade --install \
        kube-state-metrics infra/charts/kube-state-metrics -n "$NAMESPACE"
}

main() {
  local ARG0="$1"
  local ARG1="$2"
  local ARG2="$3"
  local ARG3="$4"
  local ARG4="$5"
  local ARG5="$6"
  case "$ARG0" in
    pre-install)
		pre_install "$ARG1" "$ARG2" # [tool] [version]
	;;
    bootstrap-cluster)
        bootstrap_cluster "$ARG1" "$ARG2" "$ARG3" "$ARG4" "$ARG5"
    ;;
    start-cluster)
        start_cluster "$ARG1" "$ARG2" "$ARG3" "$ARG4" "$ARG5"
    ;;
    stop-cluster)
        stop_cluster "$ARG1"
    ;;
    tunnel-registry)
        tunnel_registry
    ;;
    helm-install)
		helm_install_prometheus_operator
		helm_install_kube_state_metrics
	;;
  esac
}

main "$@"
