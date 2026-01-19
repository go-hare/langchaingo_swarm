.PHONY: test build clean install lint fmt help

# Variables
BINARY_NAME=langgraphgo_swarm
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod

# Default target
all: test build

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  make test        - Run all tests"
	@echo "  make build       - Build example binaries"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make install     - Download dependencies"
	@echo "  make lint        - Run linters"
	@echo "  make fmt         - Format code"
	@echo "  make coverage    - Run tests with coverage"
	@echo "  make examples    - Run all examples"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./swarm/...

## test-coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./swarm/...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## build: Build example binaries
build:
	@echo "Building examples..."
	$(GOBUILD) -o bin/basic ./examples/basic/main.go
	$(GOBUILD) -o bin/customer_support ./examples/customer_support/main.go
	@echo "Binaries built in bin/"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

## install: Download dependencies
install:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## lint: Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

## examples: Run all examples (requires OPENAI_API_KEY)
examples:
	@echo "Running basic example..."
	@cd examples/basic && $(GO) run main.go
	@echo "\nRunning customer support example..."
	@cd examples/customer_support && $(GO) run main.go

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./swarm/...

## doc: Generate and serve documentation
doc:
	@echo "Serving documentation on http://localhost:6060"
	godoc -http=:6060

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"
