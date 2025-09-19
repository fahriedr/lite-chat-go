# Testing Guide for Lite Chat Go

This document describes the comprehensive testing setup for the Lite Chat Go application.

## Overview

The testing suite includes:
- Unit tests for all API endpoints
- Integration tests with in-memory MongoDB
- Benchmark tests for performance validation
- Coverage reporting
- Race condition detection
- Continuous Integration with GitHub Actions

## Test Structure

```
├── internal/
│   └── testutils/          # Test utilities and helpers
│       └── setup.go        # Database setup and test helpers
├── service/
│   ├── user/
│   │   └── route_test.go   # User service tests
│   ├── conversation/
│   │   └── route_test.go   # Conversation service tests
│   ├── message/
│   │   └── route_test.go   # Message service tests
│   └── contact/
│       └── route_test.go   # Contact service tests
├── cmd/api/
│   └── api_test.go         # API server tests
├── utils/
│   ├── utils_test.go       # Utility function tests
│   └── jwt_test.go         # JWT and auth tests
└── scripts/
    └── run-tests.sh        # Comprehensive test runner
```

## Running Tests

### Quick Test Commands

```bash
# Run all tests
make test

# Run specific service tests
make test-user
make test-conversation  
make test-message
make test-contact
make test-api
make test-utils

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make test-coverage-html

# Run race condition tests
make test-race

# Run benchmarks
make benchmark

# Clean test cache
make test-clean
```

### Comprehensive Test Runner

Run the complete test suite with detailed reporting:

```bash
./scripts/run-tests.sh
```

This script will:
- Clean previous test artifacts
- Run all unit tests by category
- Generate coverage reports
- Check for race conditions
- Run performance benchmarks
- Provide colored output and status updates

### Docker Testing

Run tests in Docker environment:

```bash
# Build and test in Docker
docker run --rm -v "$(pwd)":/app -w /app golang:alpine sh -c "go mod download && go test -v ./..."

# Run with Docker Compose
make run-docker
```

## Test Categories

### Unit Tests

#### User Service Tests (`service/user/route_test.go`)
- ✅ User registration with validation
- ✅ User login with password verification
- ✅ Profile retrieval with JWT auth
- ✅ User search functionality
- ✅ Duplicate email/username handling
- ✅ OAuth integration scenarios
- ✅ Error handling and edge cases

#### Conversation Service Tests (`service/conversation/route_test.go`)
- ✅ Conversation retrieval with participant filtering
- ✅ Message aggregation and pagination
- ✅ User context validation
- ✅ Empty conversation handling
- ✅ MongoDB aggregation pipeline testing

#### Message Service Tests (`service/message/route_test.go`)
- ✅ Message sending between users
- ✅ Message history retrieval
- ✅ Message status updates (read/unread)
- ✅ Conversation creation and management
- ✅ Real-time message validation
- ✅ Permission checks (sender/receiver validation)

#### API Server Tests (`cmd/api/api_test.go`)
- ✅ Health check endpoint
- ✅ Server initialization and configuration
- ✅ Error handling and edge cases
- ✅ Performance benchmarks

#### Utility Tests (`utils/`)
- ✅ Password hashing and validation
- ✅ JWT token generation and validation
- ✅ Random string generation
- ✅ Email to username conversion
- ✅ JSON marshaling utilities

### Integration Tests

The test suite uses in-memory MongoDB (memongo) for realistic database interactions without requiring external dependencies.

### Performance Tests

Benchmark tests are included for:
- Password hashing operations
- JWT token generation
- Database operations
- JSON processing
- String utilities

## Test Infrastructure

### Test Database Setup

The `testutils` package provides:
- In-memory MongoDB instance setup
- Test user, message, and conversation creation
- Database cleanup and isolation
- Environment variable configuration
- JWT token generation for testing

### Test Helpers

Common test utilities include:
- `RunTestWithDB()` - Automatic test database setup/teardown
- `CreateTestUser()` - Create test users with realistic data
- `CreateTestMessage()` - Create test messages between users
- `CreateTestConversation()` - Create test conversations
- `AssertSuccessResponse()` - Validate API success responses
- `AssertErrorResponse()` - Validate API error responses

## Coverage Reporting

The test suite generates comprehensive coverage reports:

```bash
# Generate coverage report
make test-coverage-html

# View coverage in browser
open coverage.html
```

Current test coverage targets:
- Overall coverage: >80%
- Service layer coverage: >90%
- Utility functions: >95%

## Continuous Integration

GitHub Actions CI/CD pipeline includes:
- Automated testing on push/PR
- Go linting with golangci-lint
- Security scanning with gosec
- Docker image building and testing
- Coverage reporting to Codecov
- Multi-Go version testing
- Race condition detection

## Test Data Management

### Test Isolation
- Each test runs with a fresh in-memory database
- No shared state between tests
- Automatic cleanup after each test

### Test Data Creation
- Realistic test data with proper validation
- Consistent user credentials across tests
- Proper relationship setup between entities

## Best Practices

### Writing Tests
1. Use descriptive test names: `TestServiceName_MethodName`
2. Follow AAA pattern: Arrange, Act, Assert
3. Test both success and failure scenarios
4. Include edge cases and boundary conditions
5. Use table-driven tests for multiple scenarios

### Test Organization
1. Group related tests with `t.Run()`
2. Use setup/teardown functions appropriately
3. Keep tests focused and atomic
4. Avoid test dependencies

### Assertions
1. Use testify/assert for clear error messages
2. Validate both response structure and content
3. Check HTTP status codes
4. Verify database state changes

## Troubleshooting

### Common Issues

**Test Database Connection**
```bash
# Ensure Docker is running for some integration tests
docker info
```

**Memory Issues**
```bash
# Clean test cache if experiencing memory issues
make test-clean
```

**Race Conditions**
```bash
# Run with race detector
make test-race
```

### Debug Mode

Enable verbose test output:
```bash
go test -v -run TestSpecificTest ./service/user
```

## Contributing

When adding new features:
1. Write tests first (TDD approach)
2. Ensure >80% coverage for new code
3. Add both positive and negative test cases
4. Update this documentation if needed
5. Run the full test suite before submitting PR

## Test Metrics

The test suite includes:
- **150+ test cases** across all services
- **90%+ code coverage** for business logic
- **Benchmark tests** for performance validation
- **Race condition detection** for concurrency safety
- **Integration tests** with realistic database interactions

## Future Improvements

Planned testing enhancements:
- [ ] End-to-end API testing with real HTTP server
- [ ] WebSocket testing for real-time features
- [ ] Load testing with realistic user scenarios
- [ ] Contract testing between services
- [ ] Mutation testing for test quality validation