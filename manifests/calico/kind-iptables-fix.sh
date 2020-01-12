#!/usr/bin/env bash
[[ "$DEBUG" ]] && set -x # Print commands and their arguments as they are executed.

set -e # Exit immediately if a command exits with a non-zero status.

# -A POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE
sudo iptables -D POSTROUTING 1 -t nat
# -A POSTROUTING -o vif+ -j ACCEPT
# -A POSTROUTING -o lo -j ACCEPT
# -A POSTROUTING -j MASQUERADE
sudo iptables -D POSTROUTING 3 -t nat
sudo iptables -I POSTROUTING ! -o docker0 -j MASQUERADE -t nat  # SNAT
