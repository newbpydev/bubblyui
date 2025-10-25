.PHONY: help test test-race test-cover lint fmt imports vet build clean install-tools

# Default target
help:
	@echo "BubblyUI Development Commands:"
	@echo "  make test         - Run tests"
	@echo "  make test-race    - Run tests with race detector"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format code"
	@echo "  make imports      - Fix imports"
	@echo "  make vet          - Run go vet"
	@echo "  make build        - Build all packages"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install-tools - Install development tools"

# Testing
test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Linting
lint:
	golangci-lint run

# Formatting
fmt:
	gofmt -s -w .

imports:
	goimports -w -local github.com/newbpydev/bubblyui .

vet:
	go vet ./...

# Building
build:
	go build ./...

# Cleanup
clean:
	go clean
	rm -f coverage.out coverage.html

# Tool installation
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
