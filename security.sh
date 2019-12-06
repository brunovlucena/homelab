#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
SONOBUOY=~/.local/bin/sonobuoy

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
    	    local VERSION=0.16.1
            local OS=linux
            local SOURCE="https://github.com/vmware-tanzu/sonobuoy/releases/download/v${VERSION}/sonobuoy_${VERSION}_${OS}_amd64.tar.gz"
            if [[ ! -f "$SONOBUOY"  ]]; then
    	        curl -L "$SOURCE" --output /tmp/sonobuoy.tar.gz
    	        tar -xzf /tmp/sonobuoy.tar.gz -C ~/.local/bin
    	        chmod +x ~/.local/bin/sonobuoy
            fi
	    ;;
    esac
}
# checks deployment security
#
# Usage:
#  $ ./security.sh check_pod_security param1
# * param1: it's the pod name
check_pod_security() {
    local LABEL="$1"
    local NAMESPACE="$2"
    POD_NAME=$(kubectl get pod -l "$LABEL" -o jsonpath='{.items[0].metadata.name}' -n "$NAMESPACE")
    kubectl kubesec-scan pod $POD_NAME -n "$NAMESPACE"
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
	kubectl sniff ${POD_NAME} -n default -c chart -o ksniff-dump.pcap -p
    # To support those containers as well, ksniff now ships with the "-p" (privileged) mode.
    # When executed with the -p flag, ksniff will create a new pod on the remote kubernetes cluster that
    # will have access to the node docker daemon.
}

main() {
  local ARG0="$1"
  local ARG1="$2"
  local ARG2="$3"
  case "$ARG0" in
    check-pod-security)
        check_pod_security "component=myapp" "dev"
        check_pod_security "name=myapp-operator" "dev"
    ;;
    sniff)
        pod_sniff
    ;;
  esac
}

main "$@"
