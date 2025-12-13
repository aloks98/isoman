.PHONY: help lint fmt fmt-check test build run clean install-tools

# Go tools paths
GOPATH ?= $(shell go env GOPATH)
GOLANGCI_LINT := $(GOPATH)/bin/golangci-lint
GOFUMPT := $(GOPATH)/bin/gofumpt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install development tools (golangci-lint, gofumpt)
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing gofumpt..."
	@go install mvdan.cc/gofumpt@latest
	@echo "Tools installed successfully!"

lint: ## Run golangci-lint for linting and static analysis
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run ./...

fmt: ## Format code with gofumpt
	@echo "Formatting code with gofumpt..."
	@$(GOFUMPT) -l -w .

fmt-check: ## Check if code is formatted with gofumpt
	@echo "Checking code formatting..."
	@$(GOFUMPT) -l . | grep . && echo "Code is not formatted. Run 'make fmt' to fix." && exit 1 || echo "Code is formatted correctly."

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build the binary
	@echo "Building..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o server .
	@echo "Binary created: server"

run: ## Run the application
	@echo "Running application..."
	@go run main.go

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f server coverage.out coverage.html
	@echo "Clean complete!"

check: lint fmt-check test ## Run all checks (lint, format check, tests)

ci: ## CI pipeline: install tools, check formatting, lint, and test
	@make install-tools
	@make fmt-check
	@make lint
	@make test
