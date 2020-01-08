#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

build() {
    local APP="$1"

    cd "cmd/$APP"

    go mod tidy

    go build -gcflags -m -o "../../build/bin/$APP"
}

check() {
    local APP="$1"

    cd "cmd/$APP"

    # syscalls
    strace -c "../../build/bin/$APP"
}

# Runs App.
#
# Usage:
#  $ ./helper.sh param1
# * param1: [api|repository]
run_app() {
    local APP="$1"

    cd "cmd/$APP"

    go mod tidy
    go run main.go
}

# Runs skaffold.
#
# Usage:
#  $ ./helper.sh skaffold param1
# * param1: [api|repository]
run_skaffold() {
    ENV=dev skaffold dev --cache-artifacts=false --watch-poll-interval=2000
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
    check)
        check "$ARG1"
    ;;
    debug)
        debug "$ARG1"
    ;;
    debug-tests)
        debug_tests "$ARG1"
    ;;
    run)
        run_app "$ARG1"
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
