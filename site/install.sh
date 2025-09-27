#!/bin/bash
set -e

# Tira Installation Script
# Usage: curl -fsSL https://tira.sh | sh

echo "DOWNLOADING: Tira - Universal package manager for the curl | bash era"
echo

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="aarch64" ;;
    i?86) ARCH="i686" ;;
    armv7l) ARCH="armv7" ;;
    *)
        echo "ERROR: Unsupported architecture: $ARCH"
        echo "       Supported: x86_64, aarch64, i686"
        exit 1
        ;;
esac

case $OS in
    linux|darwin) ;;
    *)
        echo "ERROR: Unsupported OS: $OS"
        echo "       Supported: Linux, macOS (Darwin)"
        exit 1
        ;;
esac

BINARY="tira-${ARCH}-${OS}"
if [ "$OS" = "windows" ]; then
    BINARY="${BINARY}.exe"
fi

echo "DETECTED: Platform ${ARCH}-${OS}"
echo

# Set installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ] && [ "$(id -u)" -ne 0 ]; then
    echo "WARNING: $INSTALL_DIR is not writable. Trying with sudo..."
    SUDO="sudo"
else
    SUDO=""
fi

# Download URL
GITHUB_REPO="oheriko/tira"  # Update this to your actual GitHub username/repo
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY}"

echo "FETCHING: Tira from $DOWNLOAD_URL"

# Create temporary file
TMP_FILE=$(mktemp)
trap "rm -f $TMP_FILE" EXIT

# Download binary
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
else
    echo "ERROR: curl or wget is required"
    exit 1
fi

# Verify download
if [ ! -s "$TMP_FILE" ]; then
    echo "ERROR: Download failed or file is empty"
    exit 1
fi

# Make executable
chmod +x "$TMP_FILE"

# Install binary
echo "INSTALLING: Tira to $INSTALL_DIR/tira"
$SUDO mv "$TMP_FILE" "$INSTALL_DIR/tira"

# Verify installation
if ! command -v tira >/dev/null 2>&1; then
    echo "ERROR: tira not found in PATH after installation"
    echo "       You may need to add $INSTALL_DIR to your PATH"
    exit 1
fi

echo "SUCCESS: Tira installed successfully!"
echo

# Show version
TIRA_VERSION=$(tira --version 2>/dev/null || echo "unknown")
echo "READY: Tira $TIRA_VERSION is ready"
echo

# The meta moment - have Tira track its own installation!
echo "ANCHORING: Registering Tira installation with itself..."
if [ "${TIRA_INSTALLING:-}" = "true" ]; then
    echo "SUCCESS: Tira self-installation registered (nested execution)!"
elif tira install https://tira.sh/install.sh --name tira 2>/dev/null; then
    echo "SUCCESS: Tira is now tracking its own installation!"
else
    echo "NOTE: Could not register self-installation (this is normal for first-time setup)"
fi

echo
echo "GET STARTED:"
echo "  tira install https://ollama.com/install.sh"
echo "  tira install https://get.docker.com"
echo "  tira list"
echo
echo "Learn more: https://github.com/${GITHUB_REPO}"
