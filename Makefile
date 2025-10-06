# AutoPDF Makefile
# Provides convenient commands for development and testing

.PHONY: help mocks test test-verbose clean build install

# Default target
help:
	@echo "AutoPDF Development Commands"
	@echo "=========================="
	@echo "mocks        - Generate mocks using Mockery"
	@echo "test         - Run all tests"
	@echo "test-verbose - Run tests with verbose output"
	@echo "clean        - Clean generated files"
	@echo "build        - Build the project"
	@echo "install      - Install dependencies"

# Generate mocks using Mockery
mocks:
	@echo "🔨 Generating mocks..."
	@./scripts/generate_mocks.sh

# Run all tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	@go test -v ./...

# Clean generated files
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -rf mocks/*
	@go clean

# Build the project
build:
	@echo "🔨 Building project..."
	@go build ./...

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy

# Run backward compatibility tests specifically
test-backward:
	@echo "🔄 Running backward compatibility tests..."
	@go test -v ./pkg/domain -run TestBackwardCompatibility

# Run tests with mocks
test-mocks:
	@echo "🎭 Running tests with mocks..."
	@go test -v ./pkg/domain -run TestBackwardCompatibilityWithMocks

# Run all domain tests
test-domain:
	@echo "🏗️ Running domain tests..."
	@go test -v ./pkg/domain/...

# Run integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	@go test -v ./test/...

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	@go vet ./...

# Full test suite
test-full: clean mocks test-verbose
	@echo "✅ Full test suite completed!"

# Development setup
dev-setup: install mocks test
	@echo "🚀 Development environment ready!"
