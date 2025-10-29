#!/bin/bash
# Install RADb client

set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Installing RADb client for ${OS}/${ARCH}..."

# Set installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Download latest release
LATEST_RELEASE=$(curl -s https://api.github.com/repos/bss/radb-client/releases/latest | grep "tag_name" | cut -d '"' -f 4)

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to fetch latest release information"
    exit 1
fi

echo "Latest version: $LATEST_RELEASE"

# Construct download URL
BINARY_NAME="radb-client-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

DOWNLOAD_URL="https://github.com/bss/radb-client/releases/download/${LATEST_RELEASE}/${BINARY_NAME}"

echo "Downloading from: $DOWNLOAD_URL"

# Download binary
curl -L -o "/tmp/radb-client" "$DOWNLOAD_URL"

# Make executable
chmod +x "/tmp/radb-client"

# Move to installation directory
echo "Installing to ${INSTALL_DIR}/radb-client..."
sudo mv "/tmp/radb-client" "${INSTALL_DIR}/radb-client"

echo ""
echo "Installation complete!"
echo "Run 'radb-client --help' to get started"
