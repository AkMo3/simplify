.PHONY: build run test clean

BINARY_NAME=simplify

build:
	go build -o bin/$(BINARY_NAME) ./cmd/simplify

run: build
	./bin/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -rf bin/