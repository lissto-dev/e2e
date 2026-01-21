#!/bin/bash
set -euo pipefail

# Deploy Lissto via Helm chart

HELM_REF="${HELM_REF:-main}"
API_TAG="${API_TAG:-main}"
CONTROLLER_TAG="${CONTROLLER_TAG:-main}"
USE_HELM_VERSIONS="${USE_HELM_VERSIONS:-false}"
NAMESPACE="${NAMESPACE:-lissto-system}"
RELEASE_NAME="${RELEASE_NAME:-lissto}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$(dirname "$SCRIPT_DIR")/fixtures"

echo "🚀 Deploying Lissto..."
echo "   Helm ref: $HELM_REF"
echo "   API tag: $API_TAG"
echo "   Controller tag: $CONTROLLER_TAG"
echo "   Use Helm versions: $USE_HELM_VERSIONS"

# Clone helm chart at specific ref
HELM_TEMP_DIR=$(mktemp -d)
trap "rm -rf $HELM_TEMP_DIR" EXIT

echo "📦 Fetching Helm chart from lissto-dev/lissto-helm@$HELM_REF..."
git clone --depth 1 --branch "$HELM_REF" https://github.com/lissto-dev/lissto-helm.git "$HELM_TEMP_DIR/chart" 2>/dev/null || \
git clone --depth 1 https://github.com/lissto-dev/lissto-helm.git "$HELM_TEMP_DIR/chart"

# If specific ref requested (not main), checkout
if [ "$HELM_REF" != "main" ]; then
    cd "$HELM_TEMP_DIR/chart"
    git fetch --depth 1 origin "$HELM_REF" 2>/dev/null && git checkout FETCH_HEAD || true
    cd - > /dev/null
fi

# Build values override
VALUES_OVERRIDE="$HELM_TEMP_DIR/values-override.yaml"

cat > "$VALUES_OVERRIDE" << EOF
# E2E Test Configuration

# API configuration
api:
  image:
    tag: "${API_TAG}"

# Controller configuration  
controller:
  image:
    tag: "${CONTROLLER_TAG}"

# Test configuration
config:
  data:
    # Configure test repository
    repos:
      e2e-test:
        url: "https://github.com/lissto-dev/e2e"
        name: "E2E Test Repository"
        branches: ["main"]

    # Test API keys
    api_keys:
      - role: admin
        api_key: "e2e-test-admin-key-abc123"
        name: "e2e-admin"
      - role: user
        api_key: "e2e-test-user-key-xyz789"
        name: "e2e-user"

    # Namespace configuration
    namespaces:
      global: "lissto-global"
      developerPrefix: "dev-"

    # Stack image configuration (use public registry for e2e)
    stacks:
      images:
        registry: ""
        repositoryPrefix: ""
EOF

# If USE_HELM_VERSIONS is true, don't override image tags
if [ "$USE_HELM_VERSIONS" = "true" ]; then
    echo "📝 Using image versions from Helm chart defaults"
    # Remove image tag overrides
    sed -i '/^api:/,/^controller:/{/tag:/d}' "$VALUES_OVERRIDE"
    sed -i '/^controller:/,/^config:/{/tag:/d}' "$VALUES_OVERRIDE"
fi

echo "📝 Values override:"
cat "$VALUES_OVERRIDE"

# Ensure namespace exists
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Install or upgrade Helm release
echo "🔧 Installing Helm release..."
helm upgrade --install "$RELEASE_NAME" "$HELM_TEMP_DIR/chart" \
    --namespace "$NAMESPACE" \
    --values "$VALUES_OVERRIDE" \
    --wait \
    --timeout 5m

echo "✅ Lissto deployed successfully!"
echo ""
echo "Pods:"
kubectl get pods -n "$NAMESPACE"
