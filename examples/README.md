# Sendria Go Client Examples

This directory contains examples showing how to use the Sendria Go client library.

## Examples

### 1. Basic Usage (`basic/`)

Simple example showing basic API usage:
- Connecting to Sendria
- Listing messages
- Reading message content
- Deleting messages

Run:
```bash
go run examples/basic/main.go
```

### 2. Testing Example (`testing/`)

**Most Important Example** - Shows how to use Sendria for testing email functionality:
- Test helper utilities
- Testing welcome emails
- Testing password reset flows
- Testing HTML emails
- Table-driven tests
- Complete user flow testing

Run tests:
```bash
# Start Sendria first
docker run -d -p 1025:1025 -p 1080:1080 msztolcman/sendria

# Run the example tests
cd examples/testing
go test -v
```

### 3. Integration Example (`integration/`)

Shows integration with a real SMTP client:
- Sending emails via SMTP
- Retrieving them via Sendria API
- Verifying delivery

Run:
```bash
go run examples/integration/main.go
```

### 4. Monitor Example (`monitor/`)

Real-time email monitoring:
- Continuously polls for new messages
- Processes emails as they arrive
- Useful for debugging or live monitoring

Run:
```bash
go run examples/monitor/main.go
```

## Prerequisites

All examples require a running Sendria instance:

```bash
# Using Docker
docker run -p 1025:1025 -p 1080:1080 msztolcman/sendria

# Or using Python
pip install sendria
sendria
```

## Environment Variables

- `SENDRIA_URL` - Sendria API URL (default: `http://localhost:1080`)
- `SENDRIA_SMTP_HOST` - SMTP server address (default: `localhost:1025`)

## Most Useful for Testing

The **testing example** (`testing/email_test.go`) is the most valuable as it shows real-world patterns for testing email functionality in your Go applications. It includes:

- Reusable test helpers
- Common email testing scenarios
- Best practices for test isolation
- Examples of extracting and verifying email content

Copy the patterns from this example into your own test suite!