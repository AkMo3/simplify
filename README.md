# Simplify

A self-hosted Platform-as-a-Service (PaaS) for deploying and managing,
containerized applications across distributed infrastructure.

> ⚠️ **Development Status**: This project is currently in active development
> and is not yet ready for production use. APIs and features may change
> without notice.

## About

Simplify aims to be a modern alternative to platforms like Coolify, with key differentiators:

- **Native A/B and Canary Deployments** — Built-in support for advanced deployment
  strategies without external tooling
- **VPN-Based Multi-Server Coordination** — Secure communication between nodes using
  WireGuard mesh
- **Security-First Architecture** — Only port 443 exposed by default, with network
  namespace isolation per application
- **Unified Node Architecture** — Every node runs the same binary with Raft consensus
  for leader election

## Current Progress

### ✅ Completed (Stage 1)

- CLI framework with Cobra
- Container management via Podman (run, stop, rm, ps, logs)
- Port mapping and environment variable support
- Structured logging with Zap
- Configuration management with Viper

### 🚧 Planned

- Stage 2: HTTP API Server & BoltDB persistence
- Stage 3: Raft consensus integration
- Stage 4: Multi-node clustering & WireGuard VPN
- Stage 5: Traefik reverse proxy & TLS
- Stage 6: Rolling, Blue-Green, Canary, and A/B deployments

## Requirements

- Go 1.22+
- Podman

## Installation

```bash
# Clone the repository
git clone https://github.com/AkMo3/simplify.git
cd simplify

# Build
make build

# The binary will be at ./bin/simplify
```

## Usage

```bash
# Run a container
./bin/simplify run --name web --image nginx:latest --port 8080:80

# List running containers
./bin/simplify ps

# View container logs
./bin/simplify logs web
./bin/simplify logs web --follow

# Stop a container
./bin/simplify stop web

# Remove a container
./bin/simplify rm web
./bin/simplify rm --force web
```

## Configuration

Configuration is stored at `/etc/simplify/config.yaml`:

```yaml
# Environment: development | production
env: development
```

Override with a custom path:

```bash
./bin/simplify --config ~/.simplify/config.yaml run --name web --image nginx:latest
```

## Development

```bash
# Run tests
make test

# Run unit tests only (no Podman required)
make test-unit

# Run integration tests (requires Podman)
make test-integration

# Run linter
golangci-lint run

# Generate coverage report
make test-coverage
```

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please note that this project is in early development,
so major changes are expected.
