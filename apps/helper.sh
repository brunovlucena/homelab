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
    cd "$MYAPP"/cmd/myapp
	# go get 4d63.com/gochecknoglobals
	gochecknoglobals ./... || true
	# go get -u github.com/360EntSecGroup-Skylar/goreporter
	# goreporter
    go mod tidy
    # One type of analysis the compiler performs is called escape analysis.
    # This produces optimizations and simplifications around memory management.
    # To Escape analysis and Inlining (-gcflags -m)
    #go build -gcflags -m -o ../../build/_output/bin/myapp
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: dissasemble
dissasemble(){
    cd "$MYAPP"/build/_output/bin
    go tool objdump main
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: build-deploy-myapp
build_push_myapp() {
    cd "$MYAPP"
    local RELEASE="$1"
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp
    local IMAGE="$REPOSITORY/$BUILD_NAME:$RELEASE"
	docker rmi "$IMAGE" || true
	docker build -f build/Dockerfile -t "$IMAGE" .
    docker push "$IMAGE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: deploy-operator-test
deploy_myapp_test() {
    local RELEASE="$1"
    build_push_myapp $RELEASE
    local NAMESPACE=dev
    kubectl create ns $NAMESPACE || true
    helm delete myapp -n "$NAMESPACE" || true
    helm install myapp  deploy/chart/myapp -n "$NAMESPACE"
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
    kubectl create ns $NAMESPACE || true
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp-operator
    local RELEASE="$1"
    local IMAGE=$REPOSITORY/$BUILD_NAME:$RELEASE
    cd "$MYAPP_OPERATOR"
    pwd
    # Run it after updates to regenerate files after changes
    $OPERATOR_SDK generate k8s
    # Update CRDs
    $OPERATOR_SDK generate openapi
    # Build images
    $OPERATOR_SDK build "$IMAGE"
    docker push "$IMAGE"
    local IMAGE=$REPOSITORY\/$BUILD_NAME\:$RELEASE
    sed -i "s|image:.*|image: $IMAGE|g" deploy/operator.yaml
    # Setup Service Account
    kubectl apply -f deploy/service_account.yaml -n "$NAMESPACE"
    # Setup RBAC
    kubectl apply -f deploy/role.yaml -n "$NAMESPACE"
    kubectl apply -f deploy/role_binding.yaml -n "$NAMESPACE"
    # Setup the CRD
    kubectl delete -f deploy/crds/myapp.com_myapps_crd.yaml -n "$NAMESPACE" || true
    kubectl apply -f deploy/crds/myapp.com_myapps_crd.yaml -n "$NAMESPACE"
    # Deploy the app-operator
    kubectl delete -f deploy/operator.yaml -n "$NAMESPACE" || true
    kubectl apply -f deploy/operator.yaml -n "$NAMESPACE"
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: run-skafoold
run_skaffold(){
    cd apps/app-example
    ENV=dev skaffold dev --cache-artifacts=false --watch-poll-interval=2000
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
#
test(){
	cd apps/app-example/cmd/myapp
    go mod tidy
	go test ./... || true
	# calculate the total of all packages
	#go test ./... --cover | awk '{if ($1 != "?") print $5; else print "0.0";}' | sed 's/\%//g' | awk '{s+=$1} END {printf "%.2f\n", s}'
	# total number of packages
	#go test ./... --cover | wc -l
	# 66.70/319.0 = 0.21
	# So 21 % of the package covered.
}

# x.
#
# Usage:
#  $ ./helper.sh param1
# * param1: test
#
# go get github.com/smartystreets/goconvey
test_gui(){
	cd apps/app-example/cmd/myapp
    # Then watch the test results display in your browser
    $GOPATH/bin/goconvey
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
# * param1: debug-myapp
#
# b apps/app-example/cmd/myapp/repository/postgres_test.go:32
# c
debug_myapp_tests(){
    cd apps/app-example/cmd/myapp/router/
    dlv test
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

# performas a load test.
#
# Usage:
#  $ ./helper.sh param1
# * param1: load-test
# NOTE:  sudo apt install maven
load_test() {
    cd apps
    k6 run -d 10m load.js
}

main() {
  local ARG0="$1"
  local ARG1="$2"
  case "$ARG0" in
    bootstrap-operator)
        bootstrap_operator
    ;;
    build-deploy-operator)
        build_deploy_operator "$ARG1"
    ;;
    deploy-myapp-test)
        deploy_myapp_test "$ARG1"
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
    debug-myapp-tests)
        debug_myapp_tests
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
    test-gui)
        test_gui
    ;;
    load-test)
        load_test
    ;;
  esac
}

main "$@"
