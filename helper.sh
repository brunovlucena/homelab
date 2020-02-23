#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# local binaries
BIN=~/.local/bin
MINIKUBE=$BIN/minikube
KIND=$BIN/kind
K9S=$BIN/k9s
KUBECTL=$BIN/kubectl
HELM=$BIN/helm
SQUASH=$BIN/squash
KUBEDIFF=$BIN/kubediff
LINKERD=$BIN/linkerd
CALICOCTL=$BIN/calicoctl
KREW=~/.krew/bin/krew
SKAFFOLD=$BIN/skaffold

# system binaries
WGET=$(which wget)
TAR=$(which tar)
GZIP=$(which gzip)
MV=$(which mv)
CP=$(which cp)
CHMOD=$(which chmod)
GIT=$(which git)
MKDIR=$(which mkdir)
DOCKER=$(which docker)

# system variables
if [[ $OSTYPE == "linux-gnu" ]]; then
	OS="linux"
	OSLONG="linux-gnu"
elif [[ $OSTYPE == "darwin"* ]]; then
	OS="darwin"
	OSLONG="apple-darwin"
else
	echo "OS unknown. Exiting" && exit 1
fi

# pre-installs basic components.
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
            local PATH="https://github.com/kubernetes/minikube/releases/download/$VERSION/minikube-$OS-amd64"
            [[ ! -f $MINIKUBE ]] && $WGET "$PATH" -O "$MINIKUBE" ; $CHMOD +x "$MINIKUBE" ; $MINIKUBE version
        ;;
        kind)
            local PATH="https://github.com/kubernetes-sigs/kind/releases/download/$VERSION/kind-$(uname)-amd64"
            [[ ! -f $KIND ]] && $WGET "$PATH" -O "$KIND" ; $CHMOD +x "$KIND"
        ;;
        k9s)
            local PATH="https://github.com/derailed/k9s/releases/download/$VERSION/k9s_"$VERSION"_Linux_x86_64.tar.gz"
            if [[ ! -f $K9S && ! -f /tmp/k9s.tar ]]; then
                $WGET $PATH -O /tmp/k9s.tar.gz
                $GZIP -d /tmp/k9s.tar.gz
                $TAR xvf /tmp/k9s.tar
                $MV k9s $K9S
                $CHMOD +x "$K9S"
            fi
        ;;
        kubectl)
            local PATH="https://storage.googleapis.com/kubernetes-release/release/$VERSION/bin/$OS/amd64/kubectl"
            if [[ ! -f $KUBECTL ]]; then
                $WGET $PATH -O $KUBECTL && $CHMOD +x $KUBECTL
            fi
        ;;
        helm)
            local PATH="https://get.helm.sh/helm-$VERSION-$OS-amd64.tar.gz"
            if [[ ! -f $HELM && ! -f /tmp/helm.tar ]]; then
                $WGET "$PATH" -O /tmp/helm.tar.gz
                $GZIP -d /tmp/helm.tar.gz
                $TAR xvf /tmp/helm.tar -C /tmp
                $MV /tmp/$OS-amd64/helm $HELM
                $CHMOD +x "$HELM"
            fi
         ;;
        squash)
            local PATH="https://github.com/solo-io/squash/releases/download/$VERSION/squashctl-$OS"
            if [[ ! -f $SQUASH  ]]; then
                $WGET "$PATH" -O "$SQUASH"; $CHMOD +x "$SQUASH"
            fi
        ;;
        kubediff)
            if [[ ! -f $KUBEDIFF && ! -d /tmp/kubediff ]]
            then
                $GIT clone https://github.com/weaveworks/kubediff.git /tmp/kubediff
                $CP /tmp/kubediff/kubediff $KUBEDIFF
                $CP -R /tmp/kubediff/kubedifflib ~/.local/bin
            fi
        ;;
        linkerd)
            local PATH="https://github.com/linkerd/linkerd2/releases/download/stable-${VERSION}/linkerd2-cli-stable-${VERSION}-${OS}"
            if [[ ! -f $LINKERD ]]; then
                $WGET $PATH -O $LINKERD && $CHMOD +x $LINKERD
            fi
        ;;
        krew)
            local PATH="https://github.com/kubernetes-sigs/krew/releases/download/${VERSION}/krew.tar.gz"
            if [[ ! -f $KREW || ! -f /tmp/krew.tar ]]; then
                $WGET "$PATH" -O /tmp/krew.tar.gz
                $GZIP -d /tmp/krew.tar.gz
                $TAR xvf /tmp/krew.tar -C /tmp
                $MKDIR -p $HOME/.krew/bin
                $MV /tmp/krew-"$OS"_amd64 $KREW
            fi
        ;;
        skaffold)
            local PATH="https://github.com/GoogleContainerTools/skaffold/releases/download/$VERSION/skaffold-$OS-amd64"
            if [[ ! -f $SKAFFOLD ]]; then
                $WGET $PATH -O $SKAFFOLD && $CHMOD +x $SKAFFOLD
            fi
        ;;
    esac
}

# bootstraps cluster using minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2] [param3] [param4] [param5]
# * param1: start-cluster
# * param2: cpus used
# * param3: memory used
# * param4: kubernetes version
# * param5: cluster name
# * param6: vm driver
# * param7: cni
# * param8: mesh
bootstrap_cluster() {
    local CLUSTER_CPUS="$1"
    local CLUSTER_MEMORY="$2"
    local CLUSTER_DISK="$3"
    local CLUSTER_DISK_EXTRA=$4
    local CLUSTER_VERSION="$5"
    local CLUSTER_NAME="$6"
    local VM_DRIVER="$7"
    # clean before start for the first time
    clean_cluster "$CLUSTER_NAME" "$VM_DRIVER"
    # start cluster
    start_cluster "$CLUSTER_CPUS" "$CLUSTER_MEMORY" "$CLUSTER_DISK" "$CLUSTER_DISK_EXTRA" "$CLUSTER_VERSION" "$CLUSTER_NAME" "$VM_DRIVER"
    # networking
    local CNI="$8"
    [[ $CNI = "calico" ]] && kubectl apply -f manifests/calico
    # CAVEAT: ./manifests/calico/kind-iptables-fix.sh
    [[ $CNI = "kube-router" ]] && kubectl apply -f manifests/kube-router
    # mesh
    local MESH="$9"
    [[ $MESH = "linkerd" ]] && kubectl apply -f manifests/linkerd
}

# removes cluster from minikube.
#
# Usage:
#  clean_cluster [param1]
# * param1: [cluster_name]
clean_cluster() {
    local CLUSTER_NAME="$1"
    local VM_DRIVER="$2"
    if [[ $VM_DRIVER == "none" ]]; then
        $KIND delete cluster --name "$CLUSTER_NAME"
    else
        if [[ -d ~/.minikube/profiles/$CLUSTER_NAME/ ]]; then
            rm -r ~/.minikube/profiles/"$CLUSTER_NAME"
            # Stop
            vboxmanage controlvm mobimeo poweroff
            # Remove from virtualbox
            vboxmanage unregistervm --delete "$CLUSTER_NAME"
            # Remove volume because We need a new disk without partitions and filesystem.
            vboxmanage closemedium disk /home/bvl/.minikube/disks/rook-ceph-1.vdi --delete
            rm -r ~/.minikube/machines/"$CLUSTER_NAME"
        fi
    fi
}

# starts a cluster using minikube.
#
# Usage:
#  $ ./helper.sh param1 [param2] [param3] [param4] [param5] [param6] [param7]
# * param1: start-cluster
# * param2: cpus used
# * param3: memory used
# * param4: kubernetes version
# * param5: cluster name
# * param6: vm driver.
start_cluster() {
    local CLUSTER_CPUS="$1"
    local CLUSTER_MEMORY="$2"
    local CLUSTER_DISK="$3"
    local CLUSTER_DISK_EXTRA=$4
    local CLUSTER_VERSION="$5"
    local CLUSTER_NAME="$6"
    local VM_DRIVER="$7"
    if [[ $VM_DRIVER = "kind" ]]; then
        # start
        sudo $KIND create cluster --name $CLUSTER_NAME --config=./kind.yaml
        kind_add_registry $CLUSTER_NAME
        # get config
        $KIND export kubeconfig --name $CLUSTER_NAME
    elif [[ $VM_DRIVER = "none" ]]; then
	    sudo $MINIKUBE start --vm-driver="$VM_DRIVER" --cpus="$CLUSTER_CPUS" --memory="$CLUSTER_MEMORY" --disk-size="$CLUSTER_DISK" --kubernetes-version="$CLUSTER_VERSION"
        # manage pluggins
        manage_cluster_pluggins
    else
	    $MINIKUBE start --vm-driver="$VM_DRIVER" --cpus="$CLUSTER_CPUS" --memory="$CLUSTER_MEMORY" --disk-size="$CLUSTER_DISK" --kubernetes-version="$CLUSTER_VERSION" -p "$CLUSTER_NAME"
        ## NOTE: load rbd for ceph
        minikube ssh -p "$CLUSTER_NAME" "sudo modprobe rbd"
        # add a second disk for ceph
        add_disk "$CLUSTER_NAME" "$DISK_SIZE"
        # manage pluggins
        manage_cluster_pluggins
    fi
}

# stops the kubernetes cluster.
#
# Usage:
#  $ ./helper.sh param1 [param2]
# * param1: stop-cluster
# * param2: [cluster_name]
# * param3: [vm_driver]
stop_cluster() {
    local CLUSTER_NAME="$1"
    local VM_DRIVER="$2"
    if [[ $VM_DRIVER == "none" ]]; then
        $DOCKER stop $CLUSTER_NAME-control-plane
        # TODO: implement stop workers
        #kind get nodes --name homelab
        #$DOCKER stop $CLUSTER_NAME-worker
    else
	    $MINIKUBE stop -p "$CLUSTER_NAME" || true
    fi
}

kind_add_registry() {
    # create registry container unless it already exists
    local CLUSTER_NAME="$1"
    local REG_PORT='5000'
    local RUNNING="$(docker inspect -f '{{.State.Running}}' "${CLUSTER_NAME}-registry" 2>/dev/null || true)"
    if [[ ${RUNNING} != 'true' ]]; then
        docker run \
            -d --restart=always -p "${REG_PORT}:5000" --name "${CLUSTER_NAME}-registry" registry:2
    fi
    # add the registry to /etc/hosts on each node
    local IP="$(docker inspect -f '{{.NetworkSettings.IPAddress}}' "${CLUSTER_NAME}-registry" 2>/dev/null || true)"
    CMD="echo $IP registry >> /etc/hosts"
    for NODE in $(kind get nodes --name "${CLUSTER_NAME}"); do
        docker exec "${NODE}" sh -c "${CMD}"
    done
}

add_host_monitoring() {
    local ARG="$1"
    case "$ARG" in
        prometheus)
            local VERSION="$2"
            local CLUSTER_NAME="$3"
            local RUNNING="$(docker inspect -f '{{.State.Running}}' "${CLUSTER_NAME}-prometheus" 2>/dev/null || true)"
            if [[ ${RUNNING} != 'true' ]]; then
                docker run -v  /home/user/Workspace/homelab/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
                    -d --restart=always --network=host --name "${CLUSTER_NAME}-prometheus" prom/prometheus:${VERSION}
                docker run \
                    -d --restart=always --network=host --name "${CLUSTER_NAME}-grafana" grafana/grafana:6.5.2
            fi
        ;;
        node-exporter)
            local VERSION="$2"
            local CLUSTER_NAME="$3"
            #local REG_PORT='9100'
            local RUNNING="$(docker inspect -f '{{.State.Running}}' "${CLUSTER_NAME}-node-exporter" 2>/dev/null || true)"
            if [[ ${RUNNING} != 'true' ]]; then
                docker run \
                    -d --restart=always --network=host --name "${CLUSTER_NAME}-node-exporter" prom/node-exporter:${VERSION}
            fi
        ;;
        docker-hub-exporter)
            local VERSION=latest
            local CLUSTER_NAME="$3"
            #local REG_PORT='9170'
            local RUNNING="$(docker inspect -f '{{.State.Running}}' "${CLUSTER_NAME}-docker-hub-exporter" 2>/dev/null || true)"
            if [[ ${RUNNING} != 'true' ]]; then
                local IMAGES="prom/node-exporter,prom/prometheus,kindest/node,golang,busybox"
                docker run \
                    -d --restart=always --network=host --name "${CLUSTER_NAME}-docker-hub-exporter" infinityworks/docker-hub-exporter:"${VERSION}" -listen-address=:9170 -images="${IMAGES}"
            fi
        ;;
        github-exporter)
            local VERSION=latest
            local CLUSTER_NAME="$3"
            #local REG_PORT='9171'
            local RUNNING="$(docker inspect -f '{{.State.Running}}' "${CLUSTER_NAME}-github-exporter" 2>/dev/null || true)"
            if [[ ${RUNNING} != 'true' ]]; then
                local GITHUB_TOKEN=ae7928a60bb8caa3402cc0bef5ca2d3853be36df
                local API_URL=https://api.github.com/graphql
                local REPOS="prometheus/node_exporter, prometheus/prometheus, kubernetes/minikube, kubernetes-sigs/kind derailed/k9s, golang/go"
                docker run \
                    -d --restart=always --network=host --name "${CLUSTER_NAME}-github-exporter" -e API_URL="$API_URL" -e GITHUB_TOKEN="$GITHUB_TOKEN" -e REPOS="$REPOS" infinityworks/github-exporter:"${VERSION}" -listen-address=:9170 -images="${IMAGES}"
            fi
        ;;

    esac
}

# adds a second disk to minikube.
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

# manages minikube's plugins.
manage_cluster_pluggins() {
	# disable
	minikube addons disable helm-tiller
    minikube addons disable registry-creds
	# enable
    minikube addons enable registry
    minikube addons enable ingress
    minikube addons enable ingress-dns
    minikube addons enable dashboard
}

# add helm repos
add_helm_repos() {
    helm repo add stable https://kubernetes-charts.storage.googleapis.com
}

# add kube-system components.
add_basic() {
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        kubectl apply -f manifests/dashboard.yaml || true
        local NAMESPACE=kube-system
        helm upgrade --install \
            nginx-ingress helm/kube-system/nginx-ingress -n "$NAMESPACE"
        helm upgrade --install \
           kube-state-metrics helm/kube-system/kube-state-metrics -n "$NAMESPACE"
    fi
}


# add ci/cd components.
add_cicd() {
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=cicd
        kubectl create ns $NAMESPACE || true
        helm upgrade --install \
            gocd helm/cicd/gocd -n "$NAMESPACE"
        helm upgrade --install \
            gogs helm/cicd/gogs -n "$NAMESPACE"
    fi
}

# add monitoring components.
add_monitoring() {
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=monitoring
        kubectl create ns $NAMESPACE || true
        # order is important (prometheus first)
        helm upgrade --install \
            prometheus-operator helm/monitoring/prometheus-operator -n "$NAMESPACE"
        helm upgrade --install \
            kibana helm/monitoring/efk/charts/kibana -n "$NAMESPACE"
        helm upgrade --install \
            es helm/monitoring/efk/charts/es -n "$NAMESPACE"
        helm upgrade --install \
            fluentd helm/monitoring/efk/charts/fluentd -n "$NAMESPACE"
    fi
}

# add rook.
add_rook_ceph() {
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=rook-ceph
        kubectl create ns $NAMESPACE || true
        helm upgrade --install --wait \
            rook-ceph helm/storage/rook-ceph -n "$NAMESPACE" || true
    fi
}

# add backup capability.
add_backup() {
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=rook-ceph
        kubectl create ns $NAMESPACE || true
        helm upgrade -V-install \
            velero helm/storage/velero -n "$NAMESPACE"
    fi
}

# add storage components.
add_storage() {
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=storage
        kubectl create ns $NAMESPACE || true
        helm upgrade --install \
           postgres helm/storage/postgres -n "$NAMESPACE" || true
        helm upgrade --install \
           rabbitmq helm/storage/rabbitmq -n "$NAMESPACE" || true
        helm upgrade --install \
           mysql helm/storage/mysql -n "$NAMESPACE" || true
        helm upgrade --install \
           redis helm/storage/redis -n "$NAMESPACE" || true
    fi
}

# add security components
add_security(){
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=security
        kubectl create ns $NAMESPACE || true
        helm upgrade --install \
           vault helm/security/vault -n "$NAMESPACE"
    fi
}

# add testing components
add_testing(){
    local ARG0="$1"
    if [[ $ARG0 = "enabled" ]]; then
        local NAMESPACE=testing
        kubectl create ns $NAMESPACE || true
        helm upgrade --install \
           kube-monkey helm/testing/kube-monkey -n "$NAMESPACE" || true
    fi
}

# Creates a tunnel to registry.
#
# Usage:
#  $ ./helper.sh param1
# * param1: tunnel
tunnel() {
    local ARG0="$1"
    case "$ARG0" in
        registry)
	        kubectl port-forward "$(kubectl get pod -l actual-registry=true -o jsonpath='{.items[0].metadata.name}' -n kube-system)" 5000:5000 -n kube-system
        ;;
    esac
}

main() {
  local ARG0="$1" # [task]
  local ARG1="$2" # tool|cpus|cluster_name
  local ARG2="$3" # version|memory|vm_driver
  local ARG3="$4" # cluster disk
  local ARG4="$5" # disk extra
  local ARG5="$6" # cluster version
  local ARG6="$7" # cluster name
  local ARG7="$8" # vm-driver
  local ARG8="$9" # cni
  local ARG9="${10}" # mesh
  case "$ARG0" in
    pre-install)
		pre_install "$ARG1" "$ARG2"
	;;
    add-host-monitoring)
        add_host_monitoring "$ARG1" "$ARG2" "$ARG3"
    ;;
    kind-add-registry)
        kind_add_registry "$ARG1"
    ;;
    bootstrap-cluster)
        bootstrap_cluster "$ARG1" "$ARG2" "$ARG3" "$ARG4" "$ARG5" "$ARG6" "$ARG7" "$ARG8" "$ARG9"
    ;;
    start-cluster)
        start_cluster "$ARG1" "$ARG2" "$ARG3" "$ARG4" "$ARG5" "$ARG6" "$ARG7"
    ;;
    stop-cluster)
        stop_cluster "$ARG1" "$ARG2"
    ;;
    clean-cluster)
        clean_cluster "$ARG1" "$ARG2"
    ;;
    tunnel)
        tunnel "$ARG1"
    ;;
    add-basic)
        add_basic "$ARG1"
    ;;
    add-cicd)
        add_cicd "$ARG1"
    ;;
    add-monitoring)
        add_monitoring "$ARG1"
    ;;
    add-storage)
        add_storage "$ARG1"
    ;;
    add-security)
        add_security "$ARG1"
    ;;
    add-testing)
        add_testing "$ARG1"
    ;;
    add-rook-ceph)
        add_rook_ceph "$ARG1"
    ;;
    add-backup)
        add_backup "$ARG1"
    ;;
  esac
}

main "$@"
