#!/bin/bash
set -e

# linc installer
# Usage: curl -fsSL https://raw.githubusercontent.com/franzwilhelm/linc/main/install.sh | bash

REPO="franzwilhelm/linc"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="linc"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}==>${NC} $1"; }
error() { echo -e "${RED}error:${NC} $1"; exit 1; }

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        ;;
esac

case "$OS" in
    darwin|linux)
        ;;
    *)
        error "Unsupported OS: $OS"
        ;;
esac

info "Detecting system: $OS/$ARCH"

# Try to download pre-built binary from GitHub Releases
info "Fetching latest release..."
LATEST_RELEASE=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

if [ -n "$LATEST_RELEASE" ]; then
    VERSION="${LATEST_RELEASE#v}"
    # GoReleaser naming: linc_0.1.0_darwin_arm64.tar.gz
    ARCHIVE_NAME="linc_${VERSION}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$ARCHIVE_NAME"

    info "Downloading $BINARY_NAME $LATEST_RELEASE..."

    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"; then
        info "Extracting..."
        tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

        if [ -w "$INSTALL_DIR" ]; then
            mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        else
            warn "Need sudo to install to $INSTALL_DIR"
            sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        fi

        chmod +x "$INSTALL_DIR/$BINARY_NAME"

        echo ""
        info "linc $LATEST_RELEASE installed successfully!"
        echo "    Location: $INSTALL_DIR/$BINARY_NAME"
        echo ""
        echo "Run 'linc' to get started."
        exit 0
    fi

    warn "Failed to download pre-built binary"
fi

# Fallback: Try go install
if command -v go &> /dev/null; then
    info "Installing via go install..."
    go install "github.com/$REPO@latest"

    GOBIN=$(go env GOBIN)
    if [ -z "$GOBIN" ]; then
        GOBIN="$(go env GOPATH)/bin"
    fi

    if [ -f "$GOBIN/$BINARY_NAME" ]; then
        echo ""
        info "linc installed successfully!"
        echo "    Location: $GOBIN/$BINARY_NAME"
        echo ""
        if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
            warn "Add $GOBIN to your PATH:"
            echo "    export PATH=\"\$PATH:$GOBIN\""
            echo ""
        fi
        echo "Run 'linc' to get started."
        exit 0
    fi
fi

# Final fallback: Build from source
info "Building from source..."

if ! command -v go &> /dev/null; then
    error "Go is required. Install from https://go.dev/dl/"
fi

if ! command -v git &> /dev/null; then
    error "Git is required to build from source."
fi

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

git clone --depth 1 "https://github.com/$REPO.git" "$TMP_DIR"
cd "$TMP_DIR"
go build -o "$BINARY_NAME" .

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    warn "Need sudo to install to $INSTALL_DIR"
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

echo ""
info "linc installed successfully!"
echo "    Location: $INSTALL_DIR/$BINARY_NAME"
echo ""
echo "Run 'linc' to get started."
