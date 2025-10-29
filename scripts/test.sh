#!/bin/bash
# Run the full test suite for RADb client

set -e

echo "Running RADb client test suite..."
echo "=================================="

# Run unit tests with coverage
echo ""
echo "Running unit tests with coverage..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Calculate coverage percentage
echo ""
echo "Coverage report:"
go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $3}'

# Run benchmarks
echo ""
echo "Running benchmarks..."
go test -bench=. -benchmem -run=^$ ./internal/state/... ./pkg/ratelimit/...

# Run go vet
echo ""
echo "Running go vet..."
go vet ./...

# Check formatting
echo ""
echo "Checking code formatting..."
gofmt_output=$(gofmt -l .)
if [ -n "$gofmt_output" ]; then
    echo "ERROR: The following files need formatting:"
    echo "$gofmt_output"
    exit 1
fi
echo "All files are properly formatted"

echo ""
echo "=================================="
echo "All tests passed!"
