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

# Install directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

print_info "EngineerDNA Development Installer"
echo ""
echo "Install directory: ${INSTALL_DIR}"
echo ""

# Build
print_info "Building EngineerDNA..."
if ! make build; then
    print_error "Build failed"
    exit 1
fi

# Verify binary exists
if [ ! -f "bin/engineerdna" ]; then
    print_error "Binary not found at bin/engineerdna after build"
    exit 1
fi

# Install
print_info "Installing to ${INSTALL_DIR}..."

if [ -w "$INSTALL_DIR" ]; then
  cp bin/engineerdna "$INSTALL_DIR/"
  chmod +x "$INSTALL_DIR/engineerdna"
else
  echo "Permission required to install to ${INSTALL_DIR}"
  if ! sudo cp bin/engineerdna "$INSTALL_DIR/"; then
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

echo ""
print_success "Next steps:"
echo "  1. Initialize: engineerdna init"
echo "  2. Start server: engineerdna serve"
echo "  3. Open http://localhost:3847"
echo ""
echo "For development, use: ./scripts/restart-dev-server.sh"
