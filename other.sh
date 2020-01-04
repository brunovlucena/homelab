#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# local binaries
SONOBUOY=~/.local/bin/sonobuoy

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

# pre-install some basic components.
#
# Usage:
#  $ ./security.sh param1 [param2]
# * param1: pre-install
# * param2: [minikube]
pre_install() {
    local ARG="$1"
    local VERSION="$2"
    case "$ARG" in
        sonobuoy)
            local PATH="https://github.com/vmware-tanzu/sonobuoy/releases/download/v"$VERSION"/sonobuoy_"$VERSION"_"$OS"_amd64.tar.gz"
            if [[ ! -f "$SONOBUOY"  ]]; then
    	        $WGET "$PATH" -O /tmp/sonobuoy.tar.gz
                cd /tmp
    	        $GZIP -d /tmp/sonobuoy.tar.gz
    	        $TAR -xzf /tmp/sonobuoy.tar -C /tmp
                $MV /tmp/$OS-amd64/sonobuoy $SONOBUOY
    	        $CHMOD +x $SONOBUOY
            fi
	    ;;
    esac
}

# sniffs pod
#
# Usage:
#  $ ./security.sh sniff
# * param1: it's the pod name
#
# Ksniff:
#  $ kubectl sniff <POD_NAME> [-n <NAMESPACE_NAME>] [-c <CONTAINER_NAME>] [-i <INTERFACE_NAME>]
#    [-o OUTPUT_FILE] [-l LOCAL_TCPDUMP_FILE] [-r REMOTE_TCPDUMP_FILE]
sniff() {
    local LABEL="$1"
    local NAMESPACE="$2"
	POD_NAME=$(kubectl get pod -l "$LABEL" -o jsonpath='{.items[0].metadata.name}' -n "$NAMESPACE")
	sudo kubectl sniff ${POD_NAME} -n default -c chart -o ksniff-dump.pcap -p
    # To support those containers as well, ksniff now ships with the "-p" (privileged) mode.
    # When executed with the -p flag, ksniff will create a new pod on the remote kubernetes cluster that
    # will have access to the node docker daemon.
}

# checks deployment security
#
# Usage:
#  $ ./security.sh check_pod_security param1
# * param1: it's the pod name
check_pod_security() {
    local POD_NAME="$1"
    local NAMESPACE="$2"
    kubectl kubesec-scan pod $POD_NAME -n "$NAMESPACE"
}

# kube-bench
#
# Usage:
#  $ ./debugger.sh param1
# * param1: kube_bench
kube_bench() {
    docker run --pid=host -v /etc:/etc:ro -v /var:/var:ro -t -v `pwd`/infra/kube-bench/config.yaml:/opt/kube-bench/cfg/config.yaml -v $(which kubectl):/usr/bin/kubectl -v ~/.kube:/.kube -e KUBECONFIG=/.kube/config aquasec/kube-bench:latest master
}

main() {
  local ARG0="$1"
  case "$ARG0" in
    kube-bench)
		kube_bench
	;;
    check-pod-security)
        check_pod_security myapp-pod "dev"
    ;;
    sniff)
        pod_sniff
    ;;
  esac
}

main "$@"
