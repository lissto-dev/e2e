#!/bin/bash
set -euo pipefail

# Download Lissto CLI

CLI_REF="${1:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
OS="${OS:-linux}"
ARCH="${ARCH:-amd64}"

echo "📦 Downloading Lissto CLI..."
echo "   Version: $CLI_REF"
echo "   OS: $OS"
echo "   Arch: $ARCH"

# Determine if we should download release or build from source
if [ "$CLI_REF" = "latest" ] || [[ "$CLI_REF" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    # Download from GitHub releases
    echo "🔽 Downloading from GitHub releases..."
    
    if [ "$CLI_REF" = "latest" ]; then
        RELEASE_URL="https://api.github.com/repos/lissto-dev/cli/releases/latest"
    else
        RELEASE_URL="https://api.github.com/repos/lissto-dev/cli/releases/tags/$CLI_REF"
    fi
    
    # Get download URL for the appropriate asset
    DOWNLOAD_URL=$(curl -sL "$RELEASE_URL" | \
        grep -o "https://github.com/lissto-dev/cli/releases/download/[^\"]*${OS}_${ARCH}[^\"]*" | \
        head -1)
    
    if [ -z "$DOWNLOAD_URL" ]; then
        echo "❌ Failed to find release asset for $OS/$ARCH"
        exit 1
    fi
    
    echo "   URL: $DOWNLOAD_URL"
    
    # Download and extract
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    curl -sL "$DOWNLOAD_URL" | tar xz -C "$TEMP_DIR"
    
    # Find the binary (might be named 'lissto' or 'cli')
    CLI_BINARY=$(find "$TEMP_DIR" -type f -name "lissto" -o -name "cli" | head -1)
    
    if [ -z "$CLI_BINARY" ]; then
        echo "❌ CLI binary not found in archive"
        ls -la "$TEMP_DIR"
        exit 1
    fi
    
    # Install
    chmod +x "$CLI_BINARY"
    sudo mv "$CLI_BINARY" "$INSTALL_DIR/lissto" || mv "$CLI_BINARY" "$INSTALL_DIR/lissto"
    
else
    # Build from source
    echo "🔨 Building from source (branch: $CLI_REF)..."
    
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    git clone --depth 1 --branch "$CLI_REF" https://github.com/lissto-dev/cli.git "$TEMP_DIR/cli" 2>/dev/null || \
    git clone https://github.com/lissto-dev/cli.git "$TEMP_DIR/cli"
    
    cd "$TEMP_DIR/cli"
    
    # Try to checkout specific ref if not main
    if [ "$CLI_REF" != "main" ]; then
        git fetch --depth 1 origin "$CLI_REF" 2>/dev/null && git checkout FETCH_HEAD || true
    fi
    
    # Build using make for consistent build tags
    make build
    
    # Install
    chmod +x lissto
    sudo mv lissto "$INSTALL_DIR/lissto" || mv lissto "$INSTALL_DIR/lissto"
    
    cd - > /dev/null
fi

# Verify installation
echo ""
echo "✅ CLI installed successfully!"
lissto --version || lissto version || echo "Version command not available"
