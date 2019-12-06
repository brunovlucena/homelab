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
echo -e "🚛 Building operator..."
make operator-build
echo -e "🚀 Deploying deploy..."
make operator-deploy
echo -e "🚛 Building App Example image..."
make image-build-myapp
docker images | grep local
