#!/usr/bin/env bash
set encoding=utf-8
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# cluster
icon=
echo -e "$icon Pre-install..."
make pre-install
echo -e "$icon Bootstrapping..."
make bootstrap-cluster
echo -e "$icon Waiting..."
make add

# helper
espera() {
    secs=$(("$1" * 60))
    while [ $secs -gt 0 ]; do
       echo -ne "$secs\033[0K\r"
       sleep 1
       : $((secs--))
    done
}
