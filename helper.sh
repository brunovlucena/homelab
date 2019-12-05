#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

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
	        wget https://github.com/kubernetes/minikube/releases/download/${VERSION}/minikube-linux-amd64 -O ~/.local/bin/minikube
            chmod +x ~/.local/bin/minikube
        ;;
    esac
}

main() {
  local ARG0="$1"
  local ARG1="$2"
  local ARG2="$3"
  case "$ARG0" in
    pre-install)
		pre_install "$ARG1" "$ARG2" # [minikube] [version]
	;;
  esac
}

main "$@"
