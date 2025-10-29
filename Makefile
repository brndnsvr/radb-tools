.PHONY: build test clean install help

# Binary name
BINARY=radb-client

# Build directory
BUILD_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

all: test build

build: ## Build the binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/radb-client

build-all: ## Build for all platforms
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/radb-client
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/radb-client
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/radb-client
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/radb-client

test: ## Run tests
	$(GOTEST) -v -cover ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -v -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt coverage.html

install: ## Install the binary
	$(GOCMD) install ./cmd/radb-client

deps: ## Download and tidy dependencies
	$(GOMOD) download
	$(GOMOD) tidy

lint: ## Run linter
	golangci-lint run ./...

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
