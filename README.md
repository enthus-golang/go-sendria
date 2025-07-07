# sendria

[![Go Reference](https://pkg.go.dev/badge/github.com/enthus-golang/sendria.svg)](https://pkg.go.dev/github.com/enthus-golang/sendria)
[![CI](https://github.com/enthus-golang/sendria/actions/workflows/ci.yml/badge.svg)](https://github.com/enthus-golang/sendria/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/enthus-golang/sendria)](https://goreportcard.com/report/github.com/enthus-golang/sendria)
[![codecov](https://codecov.io/gh/enthus-golang/sendria/branch/main/graph/badge.svg)](https://codecov.io/gh/enthus-golang/sendria)
[![License](https://img.shields.io/github/license/enthus-golang/sendria)](https://github.com/enthus-golang/sendria/blob/main/LICENSE)
[![Release](https://img.shields.io/github/release/enthus-golang/sendria.svg)](https://github.com/enthus-golang/sendria/releases/latest)

Go client library for [Sendria](https://github.com/msztolcman/sendria) REST API - A package for integration testing with the Sendria SMTP development server.

## Overview

Sendria is an SMTP server designed for development and testing environments that catches emails and displays them in a web interface instead of sending them to real recipients. This Go package provides a client library for interacting with Sendria's REST API, making it perfect for integration testing of email functionality in your applications.

## Features

- Full support for Sendria REST API endpoints
- List and retrieve email messages
- Access email content in different formats (JSON, plain text, HTML, source, EML)
- Download attachments
- Delete individual messages or clear all messages
- Support for password-protected Sendria instances
- Configurable timeout and base URL
- Comprehensive test coverage

## Installation

```bash
go get github.com/enthus-golang/sendria
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/enthus-golang/sendria"
)

func main() {
    // Create a new client with default settings (http://localhost:1080)
    client := sendria.NewClient("")
    
    // List all messages
    messages, err := client.ListMessages(1, 10)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, msg := range messages.Messages {
        fmt.Printf("Message: %s from %s\n", msg.Subject, msg.From[0].Email)
    }
}
```

## Configuration

The client can be configured with custom settings using functional options:

```go
// Basic usage with default URL
client := sendria.NewClient("")

// Custom URL
client := sendria.NewClient("http://sendria.example.com:8025")

// With basic authentication
client := sendria.NewClient("http://localhost:1080", 
    sendria.WithBasicAuth("admin", "secret"))

// With custom timeout
client := sendria.NewClient("http://localhost:1080", 
    sendria.WithTimeout(60 * time.Second))

// With multiple options
client := sendria.NewClient("http://sendria.example.com:8025",
    sendria.WithBasicAuth("admin", "secret"),
    sendria.WithTimeout(60 * time.Second))

```

## API Reference

### Client Methods

#### NewClient(baseURL string, opts ...Option) *Client
Creates a new Sendria API client with the specified base URL and options. If baseURL is empty, defaults to "http://localhost:1080".

#### ListMessages(page, perPage int) (*MessageList, error)
Retrieves a paginated list of messages.

```go
messages, err := client.ListMessages(1, 20) // Page 1, 20 messages per page
```

#### GetMessage(id string) (*Message, error)
Retrieves detailed information about a specific message.

```go
message, err := client.GetMessage("message-id-123")
```

#### GetMessagePlain(id string) (string, error)
Retrieves the plain text content of a message.

```go
plainText, err := client.GetMessagePlain("message-id-123")
```

#### GetMessageHTML(id string) (string, error)
Retrieves the HTML content of a message.

```go
htmlContent, err := client.GetMessageHTML("message-id-123")
```

#### GetMessageSource(id string) (string, error)
Retrieves the raw source of a message.

```go
source, err := client.GetMessageSource("message-id-123")
```

#### GetMessageEML(id string) ([]byte, error)
Downloads the message as an EML file.

```go
emlData, err := client.GetMessageEML("message-id-123")
// Save to file
err = os.WriteFile("message.eml", emlData, 0644)
```

#### GetAttachment(messageID, cid string) ([]byte, error)
Downloads a specific attachment by its Content-ID.

```go
attachmentData, err := client.GetAttachment("message-id-123", "attachment-cid")
```

#### DeleteMessage(id string) error
Deletes a specific message.

```go
err := client.DeleteMessage("message-id-123")
```

#### DeleteAllMessages() error
Deletes all messages from Sendria.

```go
err := client.DeleteAllMessages()
```

## Data Structures

### Message
```go
type Message struct {
    ID          string       // Unique message identifier
    Subject     string       // Email subject
    To          []Recipient  // Recipients
    From        []Recipient  // Senders
    CreatedAt   time.Time    // When the message was received
    Size        int          // Message size in bytes
    Type        string       // MIME type
    Source      string       // Raw message source (optional)
    Parts       []Part       // Message parts (plain, HTML, etc.)
    Attachments []Attachment // File attachments
}
```

### Recipient
```go
type Recipient struct {
    Name  string // Display name
    Email string // Email address
}
```

### Part
```go
type Part struct {
    Type        string // Part type (e.g., "text/plain", "text/html")
    ContentType string // Full content type with charset
    Body        string // Part content
    Size        int    // Size in bytes
}
```

### Attachment
```go
type Attachment struct {
    CID         string // Content-ID for downloading
    Type        string // Attachment type
    Filename    string // Original filename
    ContentType string // MIME type
    Size        int    // Size in bytes
}
```

## Usage Examples

### Integration Testing Example

```go
package main

import (
    "testing"
    "time"
    
    "github.com/enthus-golang/sendria"
)

func TestEmailSending(t *testing.T) {
    // Create Sendria client
    client := sendria.NewClient("")
    
    // Clear all messages before test
    client.DeleteAllMessages()
    
    // Your application sends an email here
    sendTestEmail("test@example.com", "Test Subject", "Test Body")
    
    // Wait a moment for Sendria to receive the email
    time.Sleep(100 * time.Millisecond)
    
    // Check if email was received
    messages, err := client.ListMessages(1, 10)
    if err != nil {
        t.Fatal(err)
    }
    
    if len(messages.Messages) == 0 {
        t.Fatal("No messages received")
    }
    
    // Verify email content
    msg := messages.Messages[0]
    if msg.Subject != "Test Subject" {
        t.Errorf("Expected subject 'Test Subject', got '%s'", msg.Subject)
    }
    
    // Get and verify plain text content
    plainText, err := client.GetMessagePlain(msg.ID)
    if err != nil {
        t.Fatal(err)
    }
    
    if plainText != "Test Body" {
        t.Errorf("Expected body 'Test Body', got '%s'", plainText)
    }
}
```

### Working with Attachments

```go
// Get message with attachments
message, err := client.GetMessage("message-id")
if err != nil {
    log.Fatal(err)
}

// Download each attachment
for _, attachment := range message.Attachments {
    data, err := client.GetAttachment(message.ID, attachment.CID)
    if err != nil {
        log.Printf("Failed to download %s: %v", attachment.Filename, err)
        continue
    }
    
    // Save attachment to disk
    err = os.WriteFile(attachment.Filename, data, 0644)
    if err != nil {
        log.Printf("Failed to save %s: %v", attachment.Filename, err)
    }
}
```

### Monitoring Incoming Emails

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/enthus-golang/sendria"
)

func main() {
    client := sendria.NewClient(sendria.Config{})
    
    // Keep track of processed messages
    processedIDs := make(map[string]bool)
    
    for {
        messages, err := client.ListMessages(1, 50)
        if err != nil {
            log.Printf("Error fetching messages: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        for _, msg := range messages.Messages {
            if !processedIDs[msg.ID] {
                fmt.Printf("New message: %s from %s\n", msg.Subject, msg.From[0].Email)
                processedIDs[msg.ID] = true
                
                // Process the message
                processMessage(client, msg)
            }
        }
        
        time.Sleep(2 * time.Second)
    }
}

func processMessage(client *sendria.Client, msg sendria.Message) {
    // Get full message details
    fullMsg, err := client.GetMessage(msg.ID)
    if err != nil {
        log.Printf("Error getting message details: %v", err)
        return
    }
    
    // Process as needed
    fmt.Printf("Processing message with %d parts and %d attachments\n", 
        len(fullMsg.Parts), len(fullMsg.Attachments))
}
```

## Running Sendria

To use this library, you need a running Sendria instance:

```bash
# Install Sendria
pip install sendria

# Run Sendria on default port (1025 for SMTP, 1080 for web/API)
sendria

# Run with custom ports
sendria --smtp-port 2525 --http-port 8080

# Run with authentication
sendria --http-auth user:password
```

## Testing

### Unit Tests

Run the unit test suite:

```bash
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run with race detection
go test -v -race ./...
```

### Integration Tests

The package includes comprehensive integration tests that run against a real Sendria instance.

#### Using Docker Compose (Recommended)

The easiest way to run integration tests locally:

```bash
# Start Sendria and run tests
./scripts/integration-test.sh

# Or manually:
docker-compose up -d
export SENDRIA_URL=http://localhost:1080
go test -tags=integration -v ./...
docker-compose down
```

#### Manual Setup

1. Install and run Sendria:
```bash
pip install sendria
sendria --db /tmp/sendria.db
```

2. Run integration tests:
```bash
export SENDRIA_URL=http://localhost:1080        # Required to enable integration tests
export SENDRIA_SMTP_HOST=localhost:1025          # Optional, defaults to localhost:1025
go test -tags=integration -v ./...
```

#### What's Tested

The integration tests verify:
- Sending emails via SMTP and retrieving them via the API
- Plain text and HTML email content
- Multiple recipients
- Email attachments
- Message deletion (individual and bulk)
- Pagination
- Error handling

## Building Examples

The examples directory contains several demonstration programs. Build them individually:

```bash
# Basic usage example
go build -o basic ./examples/basic/

# Integration test example
go build -o integration ./examples/integration/

# Email monitor example
go build -o monitor ./examples/monitor/
```

## Examples

Complete working examples are available in the `examples/` directory:

- **[basic/](examples/basic/)** - Simple example showing how to list and read messages
- **[integration/](examples/integration/)** - Integration test example with SMTP sending
- **[monitor/](examples/monitor/)** - Real-time email monitoring example

To run an example:
```bash
go run examples/basic/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Sendria](https://github.com/msztolcman/sendria) - The SMTP server this client interacts with
- Built for integration testing and email development workflows