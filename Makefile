# Makefile for md2nsx

# Binary name
BINARY_NAME=md2nsx

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Build for macOS
.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf temp_nsx_output/
	rm -f *.nsx

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Format code and check if files need formatting
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(go fmt ./...)" ]; then \
		echo "Code formatting issues found. Run 'make fmt' to fix."; \
		exit 1; \
	fi
	@echo "Code formatting is correct."

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run --timeout=5m

# Lint code with auto-fix
.PHONY: lint-fix
lint-fix:
	@echo "Linting code with auto-fix..."
	golangci-lint run --fix --timeout=5m

# Run go vet for static analysis
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run go vet with all analyzers
.PHONY: vet-all
vet-all:
	@echo "Running go vet with all analyzers..."
	go vet -all ./...

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	govulncheck ./...

# Check for outdated dependencies
.PHONY: deps-check
deps-check:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Run staticcheck for additional static analysis
.PHONY: staticcheck
staticcheck:
	@echo "Running staticcheck..."
	staticcheck ./...

# Run all code quality checks
.PHONY: quality
quality: fmt-check vet lint staticcheck

# Run all checks (format, vet, lint)
.PHONY: check
check: fmt-check vet lint

# Run comprehensive checks including security
.PHONY: check-all
check-all: fmt-check vet lint staticcheck security

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-all      - Build for all platforms (Linux, Windows, macOS)"
	@echo "  deps           - Install dependencies"
	@echo "  deps-check     - Check for outdated dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  fmt-check      - Check code formatting"
	@echo "  lint           - Lint code"
	@echo "  lint-fix       - Lint code with auto-fix"
	@echo "  vet            - Run go vet"
	@echo "  vet-all        - Run go vet with all analyzers"
	@echo "  staticcheck    - Run staticcheck"
	@echo "  security       - Check for security vulnerabilities"
	@echo "  quality        - Run all code quality checks"
	@echo "  check          - Run all checks (fmt, vet, lint)"
	@echo "  check-all      - Run comprehensive checks including security"
	@echo "  help           - Show this help message" 