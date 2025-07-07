.PHONY: all test lint build clean coverage help

# Default target
all: lint test build

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

# Build the project
build:
	@echo "Building..."
	@go build -v ./...
	@cd examples && go build -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean
	@rm -f coverage.txt coverage.html
	@rm -f examples/basic_usage examples/integration_test examples/monitor_emails

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Install development tools
tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# Run security scan
security:
	@echo "Running security scan..."
	@gosec ./...

# Help
help:
	@echo "Available targets:"
	@echo "  all       - Run lint, test, and build"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  lint      - Run linters"
	@echo "  build     - Build the project"
	@echo "  clean     - Clean build artifacts"
	@echo "  deps      - Install/update dependencies"
	@echo "  tools     - Install development tools"
	@echo "  fmt       - Format code"
	@echo "  security  - Run security scan"
	@echo "  help      - Show this help message"