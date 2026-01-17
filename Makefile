.PHONY: build run test test-unit test-integration test-coverage lint fmt clean install bundle

BINARY_NAME=simplify
BUILD_TAGS=remote exclude_graphdriver_btrfs btrfs_noversion exclude_graphdriver_devicemapper containers_image_openpgp

# Version information (can be overridden during build)
VERSION ?= dev
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X github.com/AkMo3/simplify/internal/cli.Version=$(VERSION) \
                  -X github.com/AkMo3/simplify/internal/cli.GitCommit=$(GIT_COMMIT) \
                  -X github.com/AkMo3/simplify/internal/cli.BuildDate=$(BUILD_DATE)"

build:
	go build -tags "$(BUILD_TAGS)" $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/simplify

# Build with optimizations for release
build-release:
	CGO_ENABLED=0 go build -tags "$(BUILD_TAGS)" $(LDFLAGS) -ldflags "-s -w" -o bin/$(BINARY_NAME) ./cmd/simplify

# Cross-compile for Linux
.PHONY: build-linux build-linux-amd64 build-linux-arm64

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags "$(BUILD_TAGS)" $(LDFLAGS) -ldflags "-s -w" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/simplify

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags "$(BUILD_TAGS)" $(LDFLAGS) -ldflags "-s -w" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/simplify

build-linux: build-linux-amd64 build-linux-arm64
	@echo "Built Linux binaries: bin/$(BINARY_NAME)-linux-amd64, bin/$(BINARY_NAME)-linux-arm64"

# Build all release binaries
build-release-all: build-linux build-release
	@echo "Built all release binaries"

run: build
	./bin/$(BINARY_NAME)

# Run the server
run-server: build
	./bin/$(BINARY_NAME) server

# Frontend targets
.PHONY: build-frontend build-frontend-prod

# Build frontend for development
build-frontend:
	cd web && npm install && npm run build

# Build frontend for production (clean install)
build-frontend-prod:
	cd web && npm ci && npm run build

# Package frontend for release (creates web-dist.zip)
package-frontend: build-frontend-prod
	cd web && zip -r ../bin/web-dist.zip dist
	@echo "Created bin/web-dist.zip"

# Build everything for production
build-all: build-frontend-prod build-release package-frontend
	@echo "Built frontend, backend, and packaged web-dist.zip"

# Build all release artifacts (binaries + frontend zip)
release: build-linux package-frontend
	@echo "Release artifacts ready in bin/"
	@ls -la bin/

# Bundle release assets into dist/
bundle: release
	@echo "Bundling release assets..."
	@rm -rf dist && mkdir -p dist
	@cp bin/$(BINARY_NAME)-linux-amd64 dist/
	@cp bin/$(BINARY_NAME)-linux-arm64 dist/
	@cp bin/web-dist.zip dist/
	@echo "Generating checksums..."
	@cd dist && (which sha256sum > /dev/null && sha256sum * || shasum -a 256 *) > SHA256SUMS
	@echo "Bundle created in dist/:"
	@ls -la dist/

# Run with hot reload (requires air)
dev:
	air -c air.toml

# Install to GOPATH/bin
install:
	go install -tags "$(BUILD_TAGS)" $(LDFLAGS) ./cmd/simplify

# Run all tests
test:
	go test -tags "$(BUILD_TAGS)" -v ./...

# Run only unit tests (skip integration tests)
test-unit:
	SKIP_INTEGRATION=1 go test -tags "$(BUILD_TAGS)" -v ./...

# Run only integration tests (requires Podman)
test-integration:
	go test -tags "$(BUILD_TAGS)" -v -run Integration ./...

# Run tests with coverage report
test-coverage:
	go test -tags "$(BUILD_TAGS)" -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
	go test -tags "$(BUILD_TAGS)" -v -race ./...

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Check formatting
fmt:
	@echo "Checking formatting..."
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files need formatting:"; \
		echo "$$unformatted"; \
		echo ""; \
		echo "Run 'make fmt-fix' to fix."; \
		exit 1; \
	fi
	@echo "All files are formatted correctly."

# Fix formatting
fmt-fix:
	gofmt -w .
	@echo "Formatting fixed."

# Run all checks (useful before committing)
check: fmt lint test-unit
	@echo "All checks passed!"

# Verify go.mod is tidy
mod-tidy:
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod or go.sum not tidy, run 'go mod tidy'" && exit 1)

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Development helpers
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  build-release   - Build optimized binary for release"
	@echo "  run             - Build and run the binary"
	@echo "  run-server      - Build and run the server"
	@echo "  install         - Install to GOPATH/bin"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-race       - Run tests with race detector"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Check formatting"
	@echo "  fmt-fix         - Fix formatting"
	@echo "  check           - Run all checks (fmt, lint, test-unit)"
	@echo "  mod-tidy        - Verify go.mod is tidy"
	@echo "  clean           - Remove build artifacts"
	@echo "  dev-deps        - Install development dependencies"
