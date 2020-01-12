#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

GO_VERSION=1.13.5

WGET=$(which wget)
TAR=$(which tar)
GZIP=$(which gzip)

install(){
    if [[ ! -d /tmp/go.tar && ! -d /usr/local/go  ]]; then
         $WGET https://dl.google.com/go/go$VERSION.linux-amd64.tar.gz -O /tmp/go.tar.gz
         $GZIP -d /tmp/go.tar.gz
         sudo $TAR -xvf /tmp/go.tar -C /usr/local
    fi
}

get_libs() {
    go get -u github.com/go-delve/delve/cmd/dlv
    go get -u github.com/gorilla/mux
    go get -u github.com/go-sql-driver/mysql
    go get -u github.com/getsentry/sentry-go
    go get -u github.com/stretchr/testify/assert
}

install
get_libs
