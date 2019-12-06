#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# cluster
make pre-install
make bootstrap-cluster
make helm-install
# app
make tunnel-registry
make operator-build
make operator-deploy
make image-build-myapp
