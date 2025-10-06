#!/bin/bash

# AutoPDF Mockery Setup Script
# This script generates mocks for all interfaces in the AutoPDF project

set -e

echo "🔧 AutoPDF Mockery Setup"
echo "========================="

# Check if mockery is installed
if ! command -v mockery &> /dev/null; then
    echo "❌ Mockery is not installed. Installing Mockery v3..."
    go install github.com/vektra/mockery/v3@v3.5.5
fi

echo "✅ Mockery version: $(mockery version)"

# Clean existing mocks
echo "🧹 Cleaning existing mocks..."
rm -rf mocks/*

# Generate mocks
echo "🔨 Generating mocks..."
mockery

# Check if mocks were generated successfully
if [ -f "mocks/mocks.go" ]; then
    echo "✅ Mocks generated successfully!"
    echo "📁 Generated files:"
    find mocks -name "*.go" | head -10
    if [ $(find mocks -name "*.go" | wc -l) -gt 10 ]; then
        echo "   ... and $(($(find mocks -name "*.go" | wc -l) - 10)) more files"
    fi
else
    echo "❌ Failed to generate mocks"
    exit 1
fi

# Format generated code
echo "🎨 Formatting generated code..."
go fmt ./mocks/...

# Run tests to ensure mocks work
echo "🧪 Running tests to verify mocks..."
go test ./pkg/domain -v -run TestBackwardCompatibilityWithMocks

echo "✅ Mockery setup complete!"
echo ""
echo "📖 Usage:"
echo "  - Run 'make mocks' to regenerate mocks"
echo "  - Use mocks.NewMockTemplateEngine(t) in your tests"
echo "  - Set expectations with mock.EXPECT().Method().Return()"
echo "  - Mocks are automatically cleaned up after each test"
