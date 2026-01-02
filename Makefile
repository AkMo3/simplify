.PHONY: build run test test-unit test-integration test-coverage clean

BINARY_NAME=simplify
BUILD_TAGS=remote exclude_graphdriver_btrfs btrfs_noversion exclude_graphdriver_devicemapper containers_image_openpgp

build:
	go build -tags "$(BUILD_TAGS)" -o bin/$(BINARY_NAME) ./cmd/simplify

run: build
	./bin/$(BINARY_NAME)

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

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
