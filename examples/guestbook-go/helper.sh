#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

build() {
    local APP="$1"

    cd "cmd/$APP"

    # checks
    gochecknoglobals ./... || true
    go mod tidy

    go build -gcflags -m -o "build/bin/$APP"
}

main() {

  local ARG0="$1"
  local ARG1="$2"
  local ARG2="$3"
  local ARG3="$4"
  case "$ARG0" in
    build)
        build "$ARG1"
    ;;
    debug)
        debug "$ARG1"
    ;;
    debug-tests)
        debug_tests "$ARG1"
    ;;
    skaffold)
        run_skaffold
    ;;
    test)
        test "$ARG1"
    ;;
  esac
}

main "$@"
