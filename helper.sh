#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
MINIKUBE=~/.local/bin/minikube

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
            local PATH="https://github.com/kubernetes/minikube/releases/download/$VERSION/minikube-linux-amd64"
            [[ ! -f $MINIKUBE ]] && /usr/bin/wget "$PATH" -O "$MINIKUBE" ; /bin/chmod +x "$MINIKUBE"
        ;;
        operator-sdk)
            local PATH="https://github.com/operator-framework/operator-sdk/releases/download/$VERSION/operator-sdk-$VERSION-x86_64-linux-gnu"
            [[ ! -f $OPERATOR ]] && /usr/bin/wget "$PATH" -O "$OPERATOR" ; /bin/chmod +x "$OPERATOR"
        ;;
        kubediff)
            [[ ! -f $OPERATOR  ]] &&
            git clone https://github.com/weaveworks/kubediff.git /tmp/kubediff
            cp /temp/kubediff/kubediff ~./local/bin/kubediff
            cp -R kubedifflib ~/.local/usr/local/bin
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
    [[ -d ~/.minikube/profiles/mobimeo/ ]] && clean_cluster "$CLUSTER_NAME"
    # start cluster
    start_cluster "$CLUSTER_CPUS" "$CLUSTER_MEMORY" "$CLUSTER_DISK" "$CLUSTER_VERSION" "$CLUSTER_NAME"
    # add a second disk for ceph
    local DISK_SIZE=20000
    add_disk "$CLUSTER_NAME" "$DISK_SIZE"
    # manage pluggins
    manage_cluster_pluggins
    # load rbd for ceph
    minikube ssh -p "$CLUSTER_NAME" "sudo modprobe rbd"
}

# removes cluster from minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2]
# * param1: clean-cluster
# * param2: [cluster_name]
clean_cluster() {
    local CLUSTER_NAME="$1"
    # just in case of a new bootstrap with the same name
    rm -r ~/.minikube/profiles/"$CLUSTER_NAME" || true
    # Stop
    vboxmanage controlvm mobimeo poweroff || true
    # Remove from virtualbox
    vboxmanage unregistervm --delete "$CLUSTER_NAME" || true
    # Remove volume because We need a new disk without partitions and filesystem.
    vboxmanage closemedium disk /home/bvl/.minikube/disks/rook-ceph-1.vdi --delete || true
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
	$MINIKUBE stop -p "$CLUSTER_NAME" || true
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
    ## load rbd for ceph
    minikube ssh -p "$CLUSTER_NAME" "sudo modprobe rbd"
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
    # create new volume
	local VOLUME_PATH=~/.minikube/disks/rook-ceph-1.vdi
	VBoxManage createhd --filename "$VOLUME_PATH" --size "$DISK_SIZE"
    # attach to the vm
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
	kubectl port-forward "$(kubectl get pod -l actual-registry=true -o jsonpath='{.items[0].metadata.name}' -n kube-system)" 5000:5000 -n kube-system &
}

# installs prometheus-operator.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install
helm_install_prometheus_operator() {
    NAMESPACE=monitoring
    kubectl create ns "$NAMESPACE" || true
    helm upgrade --install \
		prometheus-operator infra/charts/prometheus-operator -n "$NAMESPACE"
}

# installs kube-state-metrics.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install
helm_install_kube_state_metrics() {
    NAMESPACE=kube-system
    helm upgrade --install \
        kube-state-metrics infra/charts/kube-state-metrics -n "$NAMESPACE"
}

# installs rook-ceph.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install
helm_install_rook_ceph() {
    NAMESPACE=rook-ceph
    kubectl create ns "$NAMESPACE" || true
    helm upgrade --install --wait \
        rook-ceph infra/charts/rook-ceph -n "$NAMESPACE"
}

# installs velero.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install
helm_install_velero() {
    NAMESPACE=backup
    kubectl create ns "$NAMESPACE" || true
    helm upgrade --install --wait \
        velero infra/charts/velero -n "$NAMESPACE"
}


# installs postgres.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install
helm_install_postgres() {
    NAMESPACE=postgres
    kubectl create ns "$NAMESPACE" || true
    helm upgrade --install --wait \
        postgres infra/charts/postgres -n "$NAMESPACE"
}

# installs postgres.
#
# Usage:
#  $ ./helper.sh param1
# * param1: helm-install
helm_install_efk() {
    NAMESPACE=monitoring
    kubectl create ns "$NAMESPACE" || true
    helm upgrade --install --wait \
        es infra/charts/efk/charts/es -n "$NAMESPACE"
    helm upgrade --install --wait \
        fluentd infra/charts/efk/charts/fluentd -n "$NAMESPACE"
    helm upgrade --install --wait \
        kibana infra/charts/efk/charts/kibana -n "$NAMESPACE"
}

wait() {
    secs=$(("$1" * 60))
    while [ $secs -gt 0 ]; do
       echo -ne "$secs\033[0K\r"
       sleep 1
       : $((secs--))
    done
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
    clean-cluster)
        stop_cluster "$ARG1"
        clean_cluster "$ARG1"
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
        helm_install_rook_ceph
        if [[ ! -n $(helm ls -n rook-ceph) ]]; then
            echo -e "🙏 waiting for OSD before continuing..."
            wait 2 # minutes moreless in my environment
        fi
        helm_install_velero
        helm_install_efk
        helm_install_postgres
	;;
  esac
}

main "$@"
