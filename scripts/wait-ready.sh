#!/bin/bash
set -euo pipefail

# Wait for all Lissto components to be ready

NAMESPACE="${NAMESPACE:-lissto-system}"
TIMEOUT="${TIMEOUT:-300}"

echo "⏳ Waiting for Lissto components to be ready..."
echo "   Namespace: $NAMESPACE"
echo "   Timeout: ${TIMEOUT}s"

# Wait for CRDs to be established
echo ""
echo "📋 Waiting for CRDs..."
CRDS=(
    "blueprints.env.lissto.dev"
    "stacks.env.lissto.dev"
    "envs.env.lissto.dev"
    "lisstovariables.env.lissto.dev"
    "lisstosecrets.env.lissto.dev"
)

for crd in "${CRDS[@]}"; do
    echo "   Waiting for CRD: $crd"
    kubectl wait --for=condition=Established crd/"$crd" --timeout="${TIMEOUT}s" 2>/dev/null || {
        echo "   ⚠️  CRD $crd not found, might be installed differently"
    }
done

# Wait for deployments
echo ""
echo "🚀 Waiting for deployments..."

# Get all deployments in the namespace
DEPLOYMENTS=$(kubectl get deployments -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true)

if [ -z "$DEPLOYMENTS" ]; then
    echo "   ⚠️  No deployments found in $NAMESPACE"
else
    for deploy in $DEPLOYMENTS; do
        echo "   Waiting for deployment: $deploy"
        kubectl rollout status deployment/"$deploy" -n "$NAMESPACE" --timeout="${TIMEOUT}s"
    done
fi

# Wait for pods to be ready
echo ""
echo "🐳 Waiting for pods..."
kubectl wait --for=condition=Ready pods --all -n "$NAMESPACE" --timeout="${TIMEOUT}s" 2>/dev/null || {
    echo "   ⚠️  Some pods may not be ready yet"
    kubectl get pods -n "$NAMESPACE"
}

# Test API endpoint
echo ""
echo "🔍 Testing API endpoint..."

# Try to reach the API via port-forward
API_POD=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=api -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || \
          kubectl get pods -n "$NAMESPACE" -l app=lissto-api -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

if [ -n "$API_POD" ]; then
    # Start port-forward in background
    kubectl port-forward -n "$NAMESPACE" "$API_POD" 18080:8080 &
    PF_PID=$!
    sleep 2
    
    # Test health endpoint
    if curl -s http://localhost:18080/health > /dev/null 2>&1; then
        echo "   ✅ API health check passed"
    else
        echo "   ⚠️  API health check failed"
    fi
    
    # Kill port-forward
    kill $PF_PID 2>/dev/null || true
else
    echo "   ⚠️  API pod not found"
fi

echo ""
echo "✅ Lissto components check complete!"
echo ""
echo "Pod status:"
kubectl get pods -n "$NAMESPACE"
