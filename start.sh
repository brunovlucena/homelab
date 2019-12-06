#!/usr/bin/env bash
set encoding=utf-8
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

clean(){
    docker rmi -f $(docker images -f "dangling=true" -q)
    docker images | grep local | awk '{print $1}' | xargs -I {} docker rmi {}
}

# cluster
icon=🚀
echo -e "$icon Pre-install..."
#make pre-install
echo -e "$icon Bootstrapping..."
#make bootstrap-cluster
echo -e "$icon Intalling Components via Helm Charts..."
#make helm-install
# app
#echo -e "🌉 Creating tunnel..."
#make tunnel-registry
#echo -e "😁  cleaning local images..."
#clean
echo -e "🚛 Deploying myapp-operator..."
make build-deploy-operator-test
echo -e "🚀 Deploying myapp..."
make build-deploy-test
