#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
OPERATOR_SDK=~/.local/bin/operator-sdk
MYAPP=apps/app-example
MYAPP_OPERATOR=apps/myapp-operator
# App Variables
export DATABASE_TYPE=postgres
export DATABASE_HOST=0.0.0.0
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASS=postgres
export DATABASE_NAME=myapp
export API_CONTAINER_PORT=8000

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: build-myapp
build_myapp() {
    cd "$MYAPP"
    go mod tidy
    cd cmd/myapp
    go build -o ../../build/_output/bin/myapp
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: build-deploy-myapp
build_deploy_myapp() {
    cd "$MYAPP"
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp
    local RELEASE="$1"
    local IMAGE="$REPOSITORY/$BUILD_NAME:$RELEASE"
	docker rmi "$IMAGE" || true
	docker build build -t "$IMAGE"
    docker push "$IMAGE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: deploy-test
deploy_test() {
    local NAMESPACE=dev
    kubectl create ns $NAMESPACE || true
    cd "$MYAPP"
    kubectl apply -f "$MYAPP"/deploy/deployment.yaml -n "$NAMESPACE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: deploy-operator-test
deploy_operator_test() {
    local NAMESPACE=dev
    kubectl create ns $NAMESPACE || true
    kubectl apply -f "$MYAPP_OPERATOR"/deploy/crds/app.example.com_v1alpha1_myappexample_cr.yaml -n "$NAMESPACE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: bootstrap-operator
bootstrap_operator(){
    cd "$ROOT"/cmd/
    rm -r "$MYAPP_OPERATOR"
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
    cd "$MYAPP_OPERATOR"
    pwd
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

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: run-skafoold
run_skaffold(){
    cd apps/app-example
    ENV=dev skaffold dev
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: run-my-app
run_myapp(){
    cd apps/app-example/cmd/myapp
    go mod tidy
	go run main.go
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: test
test(){
	cd apps/app-example/cmd/myapp/repository
	go mod tidy
    go test
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: debug-myapp
#
# b apps/app-example/cmd/myapp/repository/postgres.go:68
# c
debug_myapp(){
    cd apps/app-example/cmd/myapp
    dlv debug
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: run-postgres-local
run_postgres_local(){
    # start postgres
    docker stop postgres-myapp || true
    echo "Starting Container..."
    docker run --rm -d --network=host \
          -e "POSTGRES_DB=$DATABASE_NAME" \
          -e "POSTGRES_USER=$DATABASE_USER" \
          -e "POSTGRES_PASSWORD=$DATABASE_PASS" \
          --name postgres-myapp postgres:12 || true
     # create table on database myapp
     sleep 3
     echo "Creating examples..."
     local CONN=postgresql://$DATABASE_USER:$DATABASE_PASS@$DATABASE_HOST:$DATABASE_PORT/$DATABASE_NAME?sslmode=disable
     echo "Connecting to $CONN"
    #/usr/bin/psql "$CONN" < apps/examples.sql
    /usr/bin/psql "$CONN" < infra/charts/postgres/data.sql
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
    debug-myapp)
        debug_myapp
    ;;
    run-postgres-local)
        run_postgres_local
    ;;
    deploy-operator-host)
        deploy_operator_host
    ;;
    skaffold)
        run_skaffold
    ;;
    go-tidy)
        cd apps/app-example/cmd/myapp && go mod tidy
    ;;
    test)
        test
    ;;
  esac
}

main "$@"
