#!/bin/sh
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Print colored message
print_error() {
    printf "${RED}ERROR: %s${NC}\n" "$1" >&2
}

print_success() {
    printf "${GREEN}%s${NC}\n" "$1"
}

print_info() {
    printf "${YELLOW}%s${NC}\n" "$1"
}

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    print_error "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    print_error "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Version (default: latest)
VERSION="${ENGINEERDNA_VERSION:-latest}"

# Construct download URL
if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/engineerdna/engineerdna/releases/latest/download/engineerdna-${OS}-${ARCH}.tar.gz"
else
  DOWNLOAD_URL="https://github.com/engineerdna/engineerdna/releases/download/${VERSION}/engineerdna-${OS}-${ARCH}.tar.gz"
fi

# Install directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

print_info "EngineerDNA Installer"
echo ""
echo "Platform: ${OS}-${ARCH}"
echo "Version: ${VERSION}"
echo "Install directory: ${INSTALL_DIR}"
echo ""

# Download
print_info "Downloading EngineerDNA..."
if ! curl -fL "$DOWNLOAD_URL" -o /tmp/engineerdna.tar.gz; then
    print_error "Failed to download EngineerDNA from $DOWNLOAD_URL"
    print_error "Please check that the release exists and your network connection is working."
    exit 1
fi

# Extract
print_info "Extracting..."
if ! tar -xzf /tmp/engineerdna.tar.gz -C /tmp; then
    print_error "Failed to extract archive"
    exit 1
fi

# Install
print_info "Installing to ${INSTALL_DIR}..."

if [ -w "$INSTALL_DIR" ]; then
  mv /tmp/engineerdna "$INSTALL_DIR/"
  chmod +x "$INSTALL_DIR/engineerdna"
else
  echo "Permission required to install to ${INSTALL_DIR}"
  if ! sudo mv /tmp/engineerdna "$INSTALL_DIR/"; then
    print_error "Failed to install binary to ${INSTALL_DIR}"
    exit 1
  fi
  sudo chmod +x "$INSTALL_DIR/engineerdna"
fi

# Verify
if command -v engineerdna >/dev/null 2>&1; then
  VERSION_OUTPUT=$(engineerdna version)
  print_success "Installation successful!"
  echo ""
  echo "$VERSION_OUTPUT"
else
  print_error "Installation complete, but 'engineerdna' not in PATH"
  echo "Add ${INSTALL_DIR} to your PATH or run: ${INSTALL_DIR}/engineerdna"
  exit 1
fi

# Cleanup
rm -f /tmp/engineerdna.tar.gz /tmp/engineerdna

# Next steps
echo ""
print_success "Next steps:"
echo "  1. Initialize: engineerdna init"
echo "  2. Start server: engineerdna serve"
echo "  3. Open http://localhost:3847"
echo ""
echo "Documentation: https://github.com/engineerdna/engineerdna"
echo "For help: engineerdna help"
