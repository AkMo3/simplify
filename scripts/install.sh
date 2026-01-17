#!/usr/bin/env bash
#
# Simplify Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/AkMo3/simplify/main/scripts/install.sh | bash
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Configuration
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/simplify"
CONFIG_DIR="/etc/simplify"
SERVICE_USER="simplify"
REPO="AkMo3/simplify"
VERSION=""

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -t|--tag) VERSION="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$OS" != "linux" ]]; then
    log_error "This script only supports Linux. For macOS, use: brew install simplify"
    exit 1
fi

# Check for Podman
check_podman() {
    if ! command -v podman &> /dev/null; then
        log_error "Podman is not installed. Please install Podman first:"
        echo "  - Ubuntu/Debian: sudo apt install podman"
        echo "  - Fedora: sudo dnf install podman"
        echo "  - RHEL/CentOS: sudo yum install podman"
        echo "  - See: https://podman.io/getting-started/installation"
        exit 1
    fi
    log_info "Podman found: $(podman --version)"
}

# Download latest release
download_binary() {
    log_info "Downloading Simplify for $OS-$ARCH..."
    
    # Get release URL
    if [ -n "$VERSION" ]; then
        log_info "Using version: $VERSION"
        BASE_URL="https://github.com/$REPO/releases/download/$VERSION"
    else
        log_info "Using latest release"
        BASE_URL="https://github.com/$REPO/releases/latest/download"
    fi
    BINARY_URL="$BASE_URL/simplify-$OS-$ARCH"
    WEB_DIST_URL="$BASE_URL/web-dist.zip"
    
    # Download binary
    TMP_FILE=$(mktemp)
    if ! curl -fsSL "$BINARY_URL" -o "$TMP_FILE"; then
        log_error "Failed to download binary from $BINARY_URL"
        exit 1
    fi
    
    chmod +x "$TMP_FILE"
    
    # Install binary
    log_info "Installing binary to $INSTALL_DIR/simplify..."
    sudo mv "$TMP_FILE" "$INSTALL_DIR/simplify"

    # Download and install frontend
    log_info "Downloading frontend assets..."
    TMP_ZIP=$(mktemp)
    if ! curl -fsSL "$WEB_DIST_URL" -o "$TMP_ZIP"; then
        log_warn "Failed to download frontend assets from $WEB_DIST_URL"
        log_warn "Web UI may not work correctly"
    else
        WEB_DIR="/opt/simplify/web"
        log_info "Installing frontend to $WEB_DIR..."
        
        # Create web directory
        sudo mkdir -p "$WEB_DIR"
        
        # Install unzip if needed
        if ! command -v unzip &> /dev/null; then
            log_warn "unzip command not found, attempting to install..."
            if command -v apt-get &> /dev/null; then
                sudo apt-get update && sudo apt-get install -y unzip
            elif command -v dnf &> /dev/null; then
                sudo dnf install -y unzip
            elif command -v yum &> /dev/null; then
                sudo yum install -y unzip
            else
                log_error "Could not install unzip. Please install unzip manually to deploy frontend."
            fi
        fi

        if command -v unzip &> /dev/null; then
            # Unzip to temp dir first
            TMP_EXTRACT=$(mktemp -d)
            unzip -q "$TMP_ZIP" -d "$TMP_EXTRACT"
            
            # Move dist folder to destination
            if [ -d "$TMP_EXTRACT/dist" ]; then
                sudo rm -rf "$WEB_DIR/dist"
                sudo mv "$TMP_EXTRACT/dist" "$WEB_DIR/"
                log_info "Frontend installed successfully"
            else
                log_warn "web-dist.zip did not contain 'dist' folder"
            fi
            
            rm -rf "$TMP_EXTRACT"
        fi
        
        rm -f "$TMP_ZIP"
    fi
}

# Create directories and config
setup_dirs() {
    log_info "Creating directories..."
    sudo mkdir -p "$DATA_DIR"
    sudo mkdir -p "$CONFIG_DIR"
    
    # Create default config if not exists
    if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
        sudo tee "$CONFIG_DIR/config.yaml" > /dev/null <<EOF
env: production

server:
  port: 8080

database:
  path: $DATA_DIR/data.db

caddy:
  enabled: true
  data_dir: $DATA_DIR
  frontend_path: /opt/simplify/web/dist
  dashboard_domain: ""
  http_port: 80
  https_port: 443
EOF
        log_info "Created default config at $CONFIG_DIR/config.yaml"
    fi
}

# Create systemd service
install_service() {
    log_info "Creating systemd service..."
    
    # For simplicity, run as root to access system Podman socket
    # Alternative: configure Podman socket for the service user
    
    # Enable and start Podman socket (system-wide)
    if systemctl list-unit-files | grep -q podman.socket; then
        log_info "Enabling Podman system socket..."
        sudo systemctl enable --now podman.socket || true
    fi
    
    # Create systemd service (runs as root for Podman access)
    sudo tee /etc/systemd/system/simplify.service > /dev/null <<EOF
[Unit]
Description=Simplify Container Orchestrator
After=network.target podman.socket
Requires=podman.socket

[Service]
Type=simple
ExecStart=$INSTALL_DIR/simplify server --config $CONFIG_DIR/config.yaml
Restart=always
RestartSec=5
Environment=CONTAINER_HOST=unix:///run/podman/podman.sock

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    log_info "Systemd service created"
}

# Main installation
main() {
    echo ""
    echo "================================"
    echo "  Simplify Installer"
    echo "================================"
    echo ""
    
    check_podman
    download_binary
    setup_dirs
    install_service
    
    echo ""
    log_info "Installation complete!"
    echo ""
    echo "To start Simplify:"
    echo "  sudo systemctl enable --now simplify"
    echo ""
    echo "Or run manually:"
    echo "  simplify server"
    echo ""
    echo "Access the web UI at: http://localhost"
    echo ""
}

main "$@"
