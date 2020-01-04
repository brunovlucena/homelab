#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# local binaries
MINIKUBE=~/.local/bin/minikube
KIND=~/.local/bin/kind
K9S=~/.local/bin/k9s
KUBECTL=~/.local/bin/kubectl
HELM=~/.local/bin/helm
SQUASH=~/.local/bin/squash
KUBEDIFF=~/.local/bin/kubediff

# system binaries
WGET=$(which wget)
TAR=$(which tar)
GZIP=$(which gzip)
MV=$(which mv)
CP=$(which cp)
CHMOD=$(which chmod)
GIT=$(which git)

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
            [[ ! -f $KUBECTL ]] && $WGET "$PATH" -O "$KUBECTL" && $CHMOD +x $KUBECTL
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
            [[ ! -f $SQUASH  ]] && $WGET "$PATH" -O "$SQUASH"; $CHMOD +x "$SQUASH"
        ;;
        kubediff)
            if [[ ! -f $KUBEDIFF && ! -d /tmp/kubediff ]]
            then
                $GIT clone https://github.com/weaveworks/kubediff.git /tmp/kubediff
                $CP /tmp/kubediff/kubediff $KUBEDIFF
                $CP -R /tmp/kubediff/kubedifflib ~/.local/bin
            fi
        ;;
        go)
            if [[ ! -d /tmp/go.tar && ! -d /usr/local/go  ]]; then
                $WGET https://dl.google.com/go/go$VERSION.linux-amd64.tar.gz -O /tmp/go.tar.gz
                $GZIP -d /tmp/go.tar.gz
                sudo $TAR -xvf /tmp/go.tar -C /usr/local
                # tools
                go get -u github.com/go-delve/delve/cmd/dlv
                go get -u github.com/leighmcculloch/gochecknoglobals
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
bootstrap_cluster() {
    local CLUSTER_CPUS="$1"
    local CLUSTER_MEMORY="$2"
    local CLUSTER_DISK="$3"
    local CLUSTER_DISK_EXTRA=$4
    local CLUSTER_VERSION="$5"
    local CLUSTER_NAME="$6"
    local VM_DRIVER="$7"
    # clean before start for the first time
    [[ -d ~/.minikube/profiles/$CLUSTER_NAME/ ]] && clean_cluster "$CLUSTER_NAME"
    # start cluster
    start_cluster "$CLUSTER_CPUS" "$CLUSTER_MEMORY" "$CLUSTER_DISK" "$CLUSTER_DISK_EXTRA" "$CLUSTER_VERSION" "$CLUSTER_NAME" "$VM_DRIVER"
}

# removes cluster from minikube.
#
# Usage:
#  clean_cluster [param1]
# * param1: [cluster_name]
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
    if [[ $VM_DRIVER = "none" ]]; then
        # start
        sudo $KIND create cluster --name $CLUSTER_NAME
        kind_add_registry $CLUSTER_NAME
    else
	    $MINIKUBE start --cpus="$CLUSTER_CPUS" --memory="$CLUSTER_MEMORY" --disk-size="$CLUSTER_DISK" --kubernetes-version="$CLUSTER_VERSION" -p "$CLUSTER_NAME"
        ## NOTE: load rbd for ceph
        minikube ssh -p "$CLUSTER_NAME" "sudo modprobe rbd"
        # add a second disk for ceph
        add_disk "$CLUSTER_NAME" "$DISK_SIZE"
        # manage pluggins
        manage_cluster_pluggins
    fi
}

kind_add_registry() {
    # create registry container unless it already exists
    local CLUSTER_NAME="$1"
    local REG_PORT='5000'
    local RUNNING="$(docker inspect -f '{{.State.Running}}' "${CLUSTER_NAME}-registry" 2>/dev/null || true)"
    if [[ ${RUNNING} != 'true' ]]; then
        docker run \
            -d --restart=always -p "${REG_PORT}:5000" --name "${CLUSTER_NAME}-registry" \
            registry:2
    fi
    # add the registry to /etc/hosts on each node
    IP='{{.NetworkSettings.IPAddress}}'
    CMD="echo $(docker inspect -f "${IP}" "${CLUSTER_NAME}-registry") registry >> /etc/hosts"
    for NODE in $(kind get nodes --name "${KIND_CLUSTER_NAME}"); do
        docker exec "${NODE}" sh -c "${CMD}"
    done
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
add_kube_system() {
    NAMESPACE=kube-system
    helm upgrade --install \
        nginx-ingress helm/kube-system/nginx-ingress -n "$NAMESPACE"
    kubectl apply -f manifests
    helm upgrade --install \
       kube-state-metrics helm/kube-system/kube-state-metrics -n "$NAMESPACE"
}
# add ci/cd components.
add_cicd() {
    NAMESPACE=cicd
    kubectl create ns $NAMESPACE || true
    #helm upgrade --install \
        #gocd helm/cicd/gocd -n "$NAMESPACE"
    helm upgrade --install \
        gogs helm/cicd/gogs -n "$NAMESPACE"
}

# add monitoring components.
add_monitoring() {
    NAMESPACE=monitoring
    kubectl create ns $NAMESPACE || true
    # order is important (prometheus first)
    helm upgrade --install \
        prometheus-operator helm/monitoring/prometheus-operator -n "$NAMESPACE"
    helm upgrade --install \
        kibana helm/monitoring/efk/charts/kibana -n "$NAMESPACE"
    helm upgrade --install --wait \
        es helm/monitoring/efk/charts/es -n "$NAMESPACE"
    helm upgrade --install \
        fluentd helm/monitoring/efk/charts/fluentd -n "$NAMESPACE"
}

# add storage components.
add_storage() {
    NAMESPACE=storage
    kubectl create ns $NAMESPACE || true
    #helm upgrade --install --wait \
        #rook-ceph helm/storage/rook-ceph -n "$NAMESPACE" || true
    #helm upgrade --install \
        #velero helm/storage/velero -n "$NAMESPACE"
    helm upgrade --install \
       postgres helm/storage/postgres -n "$NAMESPACE" || true
    helm upgrade --install \
       rabbitmq helm/storage/rabbitmq -n "$NAMESPACE" || true
    helm upgrade --install \
       mysql helm/storage/mysql -n "$NAMESPACE" || true
    helm upgrade --install \
       redis helm/storage/redis -n "$NAMESPACE" || true
}

# add security components
add_security(){
    NAMESPACE=security
    kubectl create ns $NAMESPACE || true
    helm upgrade --install \
       vault helm/security/vault -n "$NAMESPACE"
}

# add testing components
add_testing(){
    NAMESPACE=testing
    kubectl create ns $NAMESPACE || true
    helm upgrade --install \
       kube-monkey helm/testing/kube-monkey -n "$NAMESPACE" || true
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
        $KIND delete cluster --name "$CLUSTER_NAME"
    else
	    $MINIKUBE stop -p "$CLUSTER_NAME" || true
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
  case "$ARG0" in
    pre-install)
		pre_install "$ARG1" "$ARG2" #
	;;
    bootstrap-cluster)
        bootstrap_cluster "$ARG1" "$ARG2" "$ARG3" "$ARG4" "$ARG5" "$ARG6" "$ARG7"
    ;;
    start-cluster)
        start_cluster "$ARG1" "$ARG2" "$ARG3" "$ARG4" "$ARG5" "$ARG6" "$ARG7"
    ;;
    stop-cluster)
        stop_cluster "$ARG1" "$ARG2"
    ;;
    tunnel)
        tunnel "$ARG1"
    ;;
    add-kube-system)
        add_kube_system
    ;;
    add-cicd)
        add_cicd
    ;;
    add-monitoring)
        add_monitoring
    ;;
    add-storage)
        add_storage
    ;;
    add-security)
        add_security
    ;;
    add-testing)
        add_testing
    ;;
  esac
}

main "$@"
