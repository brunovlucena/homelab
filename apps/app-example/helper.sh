#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# Variables
ROOT=apps/app-example

build_myapp() {
    cd "$ROOT"/cmd/myapp
    go build -o ../../build/myapp
}

image_build_myapp() {
    cd "$ROOT"/cmd/myapp
    local REPOSITORY=localhost:5000
    local BUILD_NAME=myapp
    local RELEASE=stable #$(git rev-parse --short HEAD)
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

main() {
  local ARG0="$1"
  case "$ARG0" in
    image-build-myapp)
        image_build_myapp
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
  esac
}

main "$@"
