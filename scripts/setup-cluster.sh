#!/bin/bash
set -euo pipefail

# Setup k3d cluster for e2e testing

CLUSTER_NAME="${CLUSTER_NAME:-lissto-e2e}"
K3S_IMAGE="${K3S_IMAGE:-rancher/k3s:v1.31.4-k3s1}"

echo "🚀 Creating k3d cluster: $CLUSTER_NAME"

# Check if cluster already exists
if k3d cluster list | grep -q "^${CLUSTER_NAME}"; then
    echo "⚠️  Cluster $CLUSTER_NAME already exists, deleting..."
    k3d cluster delete "$CLUSTER_NAME"
fi

# Create cluster with:
# - 1 server node
# - Port mappings for ingress (80, 443)
# - Disable traefik (we'll use nginx-ingress for consistency)
k3d cluster create "$CLUSTER_NAME" \
    --image "$K3S_IMAGE" \
    --servers 1 \
    --agents 0 \
    --port "8080:80@loadbalancer" \
    --port "8443:443@loadbalancer" \
    --k3s-arg "--disable=traefik@server:0" \
    --wait

# Wait for cluster to be ready
echo "⏳ Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=120s

# Install nginx ingress controller
echo "📦 Installing nginx ingress controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.0/deploy/static/provider/cloud/deploy.yaml

# Wait for ingress controller to be ready
echo "⏳ Waiting for ingress controller..."
kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=120s

# Create lissto-system namespace
kubectl create namespace lissto-system --dry-run=client -o yaml | kubectl apply -f -

echo "✅ Cluster $CLUSTER_NAME is ready!"
echo ""
echo "Cluster info:"
kubectl cluster-info
echo ""
echo "Nodes:"
kubectl get nodes
