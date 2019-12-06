#!/usr/bin/env bash
set encoding=utf-8
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# cluster
icon=🚀
echo -e "$icon Pre-install..."
make pre-install
echo -e "$icon Bootstrapping..."
make bootstrap-cluster
echo -e "$icon Intalling Components via Helm Charts..."
make helm-install
# app
make tunnel-registry
make operator-build
make operator-deploy
make image-build-myapp
