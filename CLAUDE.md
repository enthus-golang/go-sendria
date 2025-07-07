# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go client library for Sendria, an SMTP server designed for testing email functionality. The library is primarily used for testing email features in applications without actually sending emails.

## Key Commands

### Development Setup
```bash
# Start Sendria server for local development
docker-compose up -d

# Run all tests (unit + integration + examples)
go test ./...

# Run only unit tests
go test ./...

# Run integration tests (requires Sendria running)
SENDRIA_URL=http://localhost:1080 go test -tags=integration -v ./...

# Run a specific test
go test -v -run TestPasswordResetEmail ./examples/testing

# Run tests with race detection
go test -race ./...

# Check test coverage
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

### Linting and Code Quality
```bash
# Run golangci-lint (must be installed)
golangci-lint run

# Run go vet
go vet ./...

# Format code
go fmt ./...
```

## Architecture and Code Structure

### Core Components

1. **Client (`client.go`)**: Main API client that communicates with Sendria's REST API
   - Uses functional options pattern for configuration
   - Handles connection pooling and HTTP transport
   - Parses Sendria's specific response format: `{code, data, meta}`

2. **Models (`models/message.go`)**: Data structures for Sendria API responses
   - `APIResponse`: Wrapper for all API responses
   - `Message`: Email message representation
   - `Attachment`: Email attachment metadata

3. **Test Helpers (`testhelpers/email_test_helper.go`)**: Testing utilities
   - `EmailTestClient`: Wrapper around Client with test-friendly methods
   - Automatic cleanup after tests
   - Retry logic for handling timing issues

### Sendria API Response Format

The Sendria API wraps all responses in a structure:
```json
{
  "code": "OK",
  "data": {...},
  "meta": {...}
}
```

This is handled by unmarshaling into `APIResponse` first, then extracting the actual data.

### Testing Patterns

1. **Integration Tests**: Tagged with `//go:build integration`
   - Require environment variables: `SENDRIA_URL` and `SENDRIA_SMTP_HOST`
   - Test against real Sendria instance
   - Located in `integration_test.go`

2. **Example Tests**: Show real-world usage patterns
   - Located in `examples/testing/email_test.go`
   - Demonstrate table-driven tests, email verification, etc.

3. **Test Isolation**: Important for CI/CD
   - Tests clear messages at start
   - Use delays between test suites to prevent interference
   - CI runs tests sequentially to avoid conflicts

### Connection Management

- HTTP client configured with connection pooling
- Response bodies must be read and closed to prevent EOF errors
- Small delays after operations help with connection stability

### Common Issues and Solutions

1. **EOF Errors**: Usually caused by improper response body handling
   - Always read response body with `io.ReadAll()`
   - Always close response body, even for DELETE requests

2. **Test Interference**: Multiple test suites running against same Sendria
   - CI workflow runs tests sequentially
   - Clear messages between test suites
   - Add delays after clearing messages

3. **Timing Issues**: Emails take time to process
   - Use retry logic with `waitFor` patterns
   - Add small delays after SMTP operations
   - Use longer timeouts in CI environments

### Important Notes

- The library focuses on testing, not production email sending
- All examples assume Sendria is running on localhost:1080
- The SMTP server runs on port 1025 by default
- Tests should always clean up after themselves