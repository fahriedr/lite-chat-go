#!/bin/bash

# Test runner script for lite-chat-go
# This script runs all tests and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Lite Chat Go - Test Suite ===${NC}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean previous test artifacts
print_status "Cleaning previous test artifacts..."
go clean -testcache
rm -f coverage.out coverage.html

# Check if Docker is running (needed for some tests)
if ! docker info > /dev/null 2>&1; then
    print_warning "Docker is not running. Some integration tests may fail."
fi

# Run tests by category
echo -e "\n${BLUE}=== Running Unit Tests ===${NC}"

print_status "Testing utilities..."
if go test -v ./utils/...; then
    print_status "✓ Utils tests passed"
else
    print_error "✗ Utils tests failed"
    exit 1
fi

print_status "Testing user service..."
if go test -v ./service/user/...; then
    print_status "✓ User service tests passed"
else
    print_error "✗ User service tests failed"
    exit 1
fi

print_status "Testing conversation service..."
if go test -v ./service/conversation/...; then
    print_status "✓ Conversation service tests passed"
else
    print_error "✗ Conversation service tests failed"
    exit 1
fi

print_status "Testing message service..."
if go test -v ./service/message/...; then
    print_status "✓ Message service tests passed"
else
    print_error "✗ Message service tests failed"
    exit 1
fi

print_status "Testing contact service..."
if go test -v ./service/contact/...; then
    print_status "✓ Contact service tests passed"
else
    print_error "✗ Contact service tests failed"
    exit 1
fi

print_status "Testing API server..."
if go test -v ./cmd/api/...; then
    print_status "✓ API server tests passed"
else
    print_error "✗ API server tests failed"
    exit 1
fi

# Run all tests with coverage
echo -e "\n${BLUE}=== Generating Coverage Report ===${NC}"
print_status "Running tests with coverage..."

if go test -coverprofile=coverage.out ./...; then
    print_status "✓ All tests passed with coverage"
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    print_status "Coverage report generated: coverage.html"
    
    # Show coverage summary
    echo -e "\n${BLUE}=== Coverage Summary ===${NC}"
    go tool cover -func=coverage.out
else
    print_error "✗ Some tests failed"
    exit 1
fi

# Run race condition tests
echo -e "\n${BLUE}=== Running Race Condition Tests ===${NC}"
print_status "Checking for race conditions..."

if go test -race ./...; then
    print_status "✓ No race conditions detected"
else
    print_warning "⚠ Race conditions detected - check the output above"
fi

# Run benchmarks
echo -e "\n${BLUE}=== Running Benchmarks ===${NC}"
print_status "Running performance benchmarks..."

go test -bench=. -benchmem ./... | grep -E "(Benchmark|PASS|FAIL)"

echo -e "\n${GREEN}=== Test Suite Complete ===${NC}"
print_status "All tests completed successfully!"
print_status "Coverage report: coverage.html"
print_status "You can open the coverage report in your browser to see detailed coverage information."