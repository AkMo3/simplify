# Simplify Deployment Guide

## Prerequisites

- **Podman** - [Install Podman](https://podman.io/getting-started/installation)
- Linux x86_64 (Ubuntu 20.04+, Debian 11+, Fedora 36+, RHEL 8+)

Verify Podman is working:
```bash
podman run --rm alpine echo "Podman is ready!"
```

## Quick Install

Download and run the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/AkMo3/simplify/main/scripts/install.sh | bash
```

Or download manually:

```bash
# Download binary
curl -Lo simplify https://github.com/AkMo3/simplify/releases/latest/download/simplify-linux-amd64
chmod +x simplify
sudo mv simplify /usr/local/bin/

# Start server
simplify server
```

## Usage

After installation, access the web UI at **http://localhost** (or the configured port).

### Configuration (Optional)

Create `~/.config/simplify/config.yaml` to customize:

```yaml
server:
  port: 8080

caddy:
  enabled: true
  http_port: 80
  https_port: 443
```

### Run as a Service

```bash
# Enable and start service (if installed via script)
sudo systemctl enable --now simplify
```

## Troubleshooting

```bash
# Check status
simplify --version
podman ps

# View logs
journalctl -u simplify -f
```

## Uninstall

```bash
sudo systemctl stop simplify
sudo rm /usr/local/bin/simplify
sudo rm -rf /var/lib/simplify
```
