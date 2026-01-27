#!/bin/bash
set -euo pipefail

# Teardown k3d cluster

CLUSTER_NAME="${CLUSTER_NAME:-lissto-e2e}"

echo "🗑️  Deleting k3d cluster: $CLUSTER_NAME"

if k3d cluster list | grep -q "^${CLUSTER_NAME}"; then
    k3d cluster delete "$CLUSTER_NAME"
    echo "✅ Cluster $CLUSTER_NAME deleted"
else
    echo "⚠️  Cluster $CLUSTER_NAME does not exist"
fi
