#!/bin/bash
set -euo pipefail

# Setup CLI contexts for e2e testing
# Creates three contexts: e2e-admin, e2e-deploy, and e2e-user

NAMESPACE="${NAMESPACE:-lissto-system}"
SERVICE_NAME="${SERVICE_NAME:-lissto-api}"

# API keys (must match what's configured in Helm values)
ADMIN_API_KEY="${ADMIN_API_KEY:-e2e-test-admin-key-abc123}"
DEPLOY_API_KEY="${DEPLOY_API_KEY:-e2e-test-deploy-key-deploy456}"
USER_API_KEY="${USER_API_KEY:-e2e-test-user-key-xyz789}"

echo "🔧 Setting up CLI contexts..."

# Get current k8s context
KUBE_CONTEXT=$(kubectl config current-context)
echo "   K8s context: $KUBE_CONTEXT"

# Create config directory
CONFIG_DIR="${HOME}/.config/lissto"
mkdir -p "$CONFIG_DIR"

# Get API URL - try ingress first, then port-forward
API_URL=""

# Check for ingress
INGRESS_HOST=$(kubectl get ingress -n "$NAMESPACE" -o jsonpath='{.items[0].spec.rules[0].host}' 2>/dev/null || true)
if [ -n "$INGRESS_HOST" ]; then
    API_URL="https://$INGRESS_HOST"
    echo "   Using ingress: $API_URL"
fi

# If no ingress, use port-forward approach (CLI will handle this)
if [ -z "$API_URL" ]; then
    echo "   No ingress found, CLI will use port-forward"
fi

# Create config file with all contexts
cat > "$CONFIG_DIR/config.yaml" << EOF
current-context: e2e-admin
current-env: ""
contexts:
  - name: e2e-admin
    kube-context: ${KUBE_CONTEXT}
    service-name: ${SERVICE_NAME}
    service-namespace: ${NAMESPACE}
    api-key: ${ADMIN_API_KEY}
    api-url: "${API_URL}"
  - name: e2e-deploy
    kube-context: ${KUBE_CONTEXT}
    service-name: ${SERVICE_NAME}
    service-namespace: ${NAMESPACE}
    api-key: ${DEPLOY_API_KEY}
    api-url: "${API_URL}"
  - name: e2e-user
    kube-context: ${KUBE_CONTEXT}
    service-name: ${SERVICE_NAME}
    service-namespace: ${NAMESPACE}
    api-key: ${USER_API_KEY}
    api-url: "${API_URL}"
settings:
  update-check: false
EOF

echo "✅ CLI contexts created!"
echo ""
echo "Available contexts:"
echo "  - e2e-admin  (admin role  - for read/delete operations)"
echo "  - e2e-deploy (deploy role - for global blueprint creation)"
echo "  - e2e-user   (user role   - for stack and user blueprint operations)"
echo ""
echo "Current context: e2e-admin"
echo ""
echo "Switch contexts with: lissto context use <name>"

# Test connectivity
echo ""
echo "🔍 Testing API connectivity..."

# Use admin context to test
if lissto context use e2e-admin 2>/dev/null; then
    if lissto blueprint list 2>/dev/null; then
        echo "✅ API is accessible"
    else
        echo "⚠️  API test failed - this is expected if components are still starting"
    fi
else
    echo "⚠️  Context switch failed - CLI may not be fully configured"
fi
