#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
KUBEDIFF=~/.local/bin/kubediff

# pre-install some basic components.
#
# Usage:
#  $ ./debugger.sh param1 [param2]
# * param1: pre-install
# * param2: [minikube]
pre_install() {
    local ARG="$1"
    local VERSION="$2"
    case "$ARG" in
        kubediff)
            if [[ ! -f $KUBEDIFF  ]]; then
                git clone https://github.com/weaveworks/kubediff.git /tmp/kubediff
                cp /temp/kubediff/kubediff "$KUBEDIFF"
                cp -R kubedifflib ~/.local/usr/local/bin
            fi
        ;;
    esac
}
# cluster roles
#
# Usage:
#  $ ./debugger.sh param1
# * param1: cluster_roles
cluster_roles() {
    echo "------------------------------------"
    echo -e "\n==== RBAC(default): ====\n"
    kubectl rbac-lookup default -o wide
    echo -e "\n==== RBAC(kube-system): ====\n"
    kubectl rbac-lookup kube-system -o wide
    echo "------------------------------------"
}

# cluster rights
#
# Usage:
#  $ ./debugger.sh param1
# * param1: cluster_rights
# kubectl krew install rbac-lookup who-can
cluster_rights() {
    echo "------------------------------------"
    echo -e "\n==== List who can get customresourcedefinitions: ====\n"
    kubectl who-can get customresourcedefinitions
    echo -e "\n==== List who can create services: ====\n"
    kubectl who-can create services
    echo -e "\n==== List who can create pods: ====\n"
    kubectl who-can create pods
    echo -e "\n==== List who can read pod logs: ====\n"
    kubectl who-can get pods --subresource=log
    echo -e "\n==== List who can access the URL /api: ====\n"
    kubectl who-can get /api
    echo "------------------------------------"
}

# kubediff
#
# Usage:
#  $ ./debugger.sh kubediff
kubediff() {
	# postgres
	helm template charts/postgres > kubediff/postgres.yaml
	sed -i "s/RELEASE-NAME/postgres/g" kubediff/postgres.yaml
    # operator

    # myapp
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
  local ARG1="$2"
  local ARG2="$3"
  local ARG3="$4"
  local ARG4="$5"
  case "$ARG0" in
    cluster-rights)
	    cluster_rights
	;;
    cluster-roles)
	    cluster_roles
	;;
    kubediff)
	    kubediff
	;;
    kube-bench)
		kube_bench
	;;
  esac
}

main "$@"
