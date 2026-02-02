#!/bin/bash
# Twingate Connector Docker Container Creation Script
# 
# Creates a Twingate connector Docker container to bridge the machine with Twingate.
#
# Usage: ./twingate-connector.sh <cluster-name>
#
# Environment variables:
#   TWINGATE_NETWORK - Twingate network name (default: bvlucena)
#   TWINGATE_{CLUSTER}_ACCESS_TOKEN - Access token for the cluster
#   TWINGATE_{CLUSTER}_REFRESH_TOKEN - Refresh token for the cluster

CLUSTER_NAME="${1:-studio}"
NETWORK="${TWINGATE_NETWORK:-bvlucena}"
HOSTNAME="${HOSTNAME:-$(hostname)}"
CONTAINER_NAME="twingate-${CLUSTER_NAME}"

# Get tokens from environment variables
# Convert cluster name to uppercase (macOS compatible)
CLUSTER_NAME_UPPER=$(echo "${CLUSTER_NAME}" | tr '[:lower:]' '[:upper:]')
ACCESS_TOKEN_VAR="TWINGATE_${CLUSTER_NAME_UPPER}_ACCESS_TOKEN"
REFRESH_TOKEN_VAR="TWINGATE_${CLUSTER_NAME_UPPER}_REFRESH_TOKEN"
ACCESS_TOKEN="${!ACCESS_TOKEN_VAR}"
REFRESH_TOKEN="${!REFRESH_TOKEN_VAR}"

if [ -z "${ACCESS_TOKEN}" ] || [ -z "${REFRESH_TOKEN}" ]; then
  echo "Error: Twingate tokens not found in environment"
  echo "Set ${ACCESS_TOKEN_VAR} and ${REFRESH_TOKEN_VAR} environment variables"
  exit 1
fi

# Stop and remove existing container if it exists
docker stop ${CONTAINER_NAME} >/dev/null 2>&1 || true
docker rm ${CONTAINER_NAME} >/dev/null 2>&1 || true

# Create container
docker run -d \
  --sysctl net.ipv4.ping_group_range="0 2147483647" \
  --env TWINGATE_NETWORK="${NETWORK}" \
  --env TWINGATE_ACCESS_TOKEN="${ACCESS_TOKEN}" \
  --env TWINGATE_REFRESH_TOKEN="${REFRESH_TOKEN}" \
  --env TWINGATE_LABEL_HOSTNAME="${HOSTNAME}" \
  --env TWINGATE_LABEL_DEPLOYED_BY="docker" \
  --name "${CONTAINER_NAME}" \
  --restart=unless-stopped \
  --pull=always \
  twingate/connector:1
