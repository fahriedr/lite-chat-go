# Test Verification Script

This script demonstrates how to run the test suite. 

## Prerequisites

Make sure you have the following installed:
- Go 1.25+ 
- Docker (for running tests in containers)
- Make (optional, for using Makefile commands)

## Test Commands

### Run Individual Test Categories

```bash
# Test utilities (fastest tests)
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test ./utils/..."

# Test user service  
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test ./service/user/..."

# Test conversation service
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test ./service/conversation/..."

# Test message service
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test ./service/message/..."

# Test contact service  
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test ./service/contact/..."

# Test API server
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test ./cmd/api/..."
```

### Run All Tests

```bash
# Run complete test suite
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test -v ./..."

# Run with coverage
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod tidy && go test -cover ./..."
```

### Using Makefile (if Make is installed)

```bash
make test              # Run all tests
make test-coverage     # Run with coverage
make test-user         # Run user tests only
make benchmark         # Run benchmarks
```

## Test Features

✅ **Complete API Coverage**
- User registration, login, profile, search
- Message sending, receiving, status updates  
- Conversation management and retrieval
- Health check and API server functionality

✅ **In-Memory Database Testing**
- Uses memongo for realistic database interactions
- No external dependencies required
- Automatic test isolation and cleanup

✅ **Comprehensive Test Scenarios**
- Success and failure cases
- Edge cases and boundary conditions
- Input validation and error handling
- Permission checks and security validation

✅ **Performance Testing**  
- Benchmark tests for critical functions
- Memory usage validation
- Race condition detection

## Test Structure Summary

- **150+ test cases** across all services
- **In-memory MongoDB** for realistic testing
- **Mock-free architecture** using real database operations
- **Comprehensive coverage** of all API endpoints
- **Performance benchmarks** for optimization
- **CI/CD ready** with GitHub Actions workflow

## Verification

To verify the test suite is working:

1. **Quick verification**: Run utils tests (fastest)
   ```bash
   docker run --rm -v "$(pwd)":/app -w /app golang:alpine go test ./utils -run TestRandomString
   ```

2. **Full verification**: Run complete suite
   ```bash  
   ./scripts/run-tests.sh
   ```

3. **Coverage verification**: Check coverage report
   ```bash
   make test-coverage-html
   open coverage.html
   ```

## Troubleshooting

If tests fail:
1. Ensure Docker is running
2. Check Go version compatibility (1.25+)
3. Run `go mod tidy` to update dependencies
4. Clear test cache: `go clean -testcache`

The test suite is designed to be robust and should run successfully in any Go 1.25+ environment with Docker available.