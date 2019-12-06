#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
ROOT=apps/app-example

build_myapp() {
    cd "$ROOT"/cmd/myapp
    go build -o ../../build/myapp
}

image_build_push_myapp() {
    cd "$ROOT"/cmd/myapp
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp
    local RELEASE="$1"
	docker rmi $REPOSITORY/$BUILD_NAME:$RELEASE || true
	docker build . -t $REPOSITORY/$BUILD_NAME:$RELEASE
    docker push $REPOSITORY/$BUILD_NAME:$RELEASE
}

run_myapp() {
    go run "$ROOT"/cmd/myapp/main.go
}

mod_tidy_myapp() {
    cd "$ROOT"/cmd/myapp && go mod tidy
}

create_helm_operator(){
    operator-sdk new myapp-helm-operator --api-version=example.com/v1alpha1 --kind=AppServiceHelm --helm-chart=deployments/chart --type=helm
}

run_test_deployment() {
    local NAMESPACE=dev
    kubectl create ns $NAMESPACE || true
    kubectl apply -f apps/app-example/deployments/deployment.yaml -n "$NAMESPACE"
}

main() {
  local ARG0="$1"
  local ARG1="$2"
  case "$ARG0" in
    image-build-push-myapp)
        image_build_push_myapp "$ARG1"
    ;;
    build-myapp)
        build_myapp
    ;;
    run-myapp)
        run_myapp
    ;;
    mod-tidy-myapp)
        mod_tidy_myapp
    ;;
    create-helm-operator)
        create_helm_operator
    ;;
    run-test-deployment)
        run_test_deployment
    ;;
  esac
}

main "$@"
