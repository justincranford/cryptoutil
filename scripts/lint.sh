#!/bin/bash
# Format and lint Go code for cryptoutil project

echo "🔧 Running gofumpt (stricter gofmt)..."
gofumpt -l -w .

echo "📦 Running goimports (import organization)..."
goimports -l -w .

echo "🏗️ Running go vet (static analysis)..."
go vet ./...

echo "🔍 Running go build (compilation check)..."
go build ./...

echo "✅ Code formatting and basic linting complete!"

echo ""
echo "Note: golangci-lint requires Go 1.24 but project uses Go 1.25"
echo "Consider using golangci-lint built with Go 1.25+ for full linting"
echo ""
echo "Manual checks you can run:"
echo "- go test ./... -cover  # Run tests with coverage"
echo "- go mod tidy          # Clean up dependencies"
echo "- go generate ./...    # Regenerate code"
