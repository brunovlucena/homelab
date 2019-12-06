#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
ROOT=apps/app-example
OPERATOR_SDK=~/.local/bin/operator-sdk

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: build-myapp
build_myapp() {
    cd "$ROOT"/cmd/myapp
    go mod tidy
    go build -o ../../build/myapp
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: build-deploy-myapp
build_deploy_myapp() {
    cd "$ROOT"/cmd/myapp
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp
    local RELEASE="$1"
	docker rmi $REPOSITORY/$BUILD_NAME:$RELEASE || true
	docker build . -t $REPOSITORY/$BUILD_NAME:$RELEASE
    docker push $REPOSITORY/$BUILD_NAME:$RELEASE
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: run-myapp
run_myapp() {
    cd "$ROOT"/cmd/myapp
    go mod tidy
    go run main.go
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: deploy-test
deploy_test() {
    local NAMESPACE=dev
    kubectl create ns $NAMESPACE || true
    kubectl apply -f apps/app-example/deployments/deployment.yaml -n "$NAMESPACE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: deploy-operator-test
deploy_operator_test() {
    local NAMESPACE=dev
    kubectl create ns $NAMESPACE || true
    cd "$ROOT"/cmd/myapp-operator
    kubectl apply -f deploy/crds/app.example.com_v1alpha1_myappexample_cr.yaml -n "$NAMESPACE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: bootstrap-operator
bootstrap_operator(){
    cd "$ROOT"/cmd/
    rm -r myapp-operator
    $OPERATOR_SDK new myapp-operator --repo github.com/brunovlucena/mobimeo
    cd myapp-operator
    # Add a new API for the custom resource
    $OPERATOR_SDK add api --api-version=app.example.com/v1alpha1 --kind=MyAppExample
    # Add a new controller that watches
    $OPERATOR_SDK add controller --api-version=app.example.com/v1alpha1 --kind=MyAppExample
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: build-push-operator
build_deploy_operator() {
    local NAMESPACE=dev
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp-operator
    local RELEASE="$1"
    local IMAGE=$REPOSITORY/$BUILD_NAME:$RELEASE
    cd "$ROOT"/cmd/myapp-operator
    $OPERATOR_SDK build "$IMAGE"
    docker push "$IMAGE"
    local IMAGE=$REPOSITORY\/$BUILD_NAME\:$RELEASE
    sed -i "s|REPLACE_IMAGE|$IMAGE|g" deploy/operator.yaml
    # Setup Service Account
    kubectl apply -f deploy/service_account.yaml -n "$NAMESPACE"
    # Setup RBAC
    kubectl apply -f deploy/role.yaml -n "$NAMESPACE"
    kubectl apply -f deploy/role_binding.yaml -n "$NAMESPACE"
    # Setup the CRD
    kubectl apply -f deploy/crds/app.example.com_myappexamples_crd.yaml -n "$NAMESPACE"
    # Deploy the app-operator
    kubectl apply -f deploy/operator.yaml -n "$NAMESPACE"
}

main() {
  local ARG0="$1"
  local ARG1="$2"
  case "$ARG0" in
    bootstrap-operator)
        bootstrap_operator
    ;;
    build-deploy-myapp)
        build_deploy_myapp "$ARG1"
    ;;
    build-deploy-operator)
        build_deploy_operator "$ARG1"
    ;;
    build-myapp)
        build_myapp
    ;;
    run-myapp)
        run_myapp
    ;;
    deploy-test)
        deploy_test
    ;;
    deploy-operator-test)
        deploy_operator_test
    ;;
  esac
}

main "$@"
