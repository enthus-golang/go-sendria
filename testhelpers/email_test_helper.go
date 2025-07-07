// Package testhelpers provides utilities for testing email functionality with Sendria.
package testhelpers

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/enthus-golang/sendria"
)

// EmailTestClient wraps Sendria client with test-friendly helpers
type EmailTestClient struct {
	*sendria.Client
	t *testing.T
}

// NewEmailTestClient creates a test-friendly email client with automatic cleanup
func NewEmailTestClient(t *testing.T) *EmailTestClient {
	t.Helper()

	// Get URL from environment or use default
	url := os.Getenv("SENDRIA_URL")
	if url == "" {
		url = "http://localhost:1080"
	}

	client := sendria.NewClient(url)

	// Clear messages at start
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}

	// Ensure cleanup after test
	t.Cleanup(func() {
		// If test failed, log all messages for debugging
		if t.Failed() {
			messages, _ := client.ListMessages(1, 100)
			t.Logf("=== Email messages at test end (%d total) ===", len(messages.Messages))
			for i, msg := range messages.Messages {
				fromEmail := "<empty>"
				if len(msg.From) > 0 {
					fromEmail = msg.From[0].Email
				}
				toEmail := "<empty>"
				if len(msg.To) > 0 {
					toEmail = msg.To[0].Email
				}
				t.Logf("[%d] From: %s, To: %s, Subject: %s",
					i+1, fromEmail, toEmail, msg.Subject)
			}
		}
		_ = client.DeleteAllMessages()
	})

	return &EmailTestClient{
		Client: client,
		t:      t,
	}
}

// WaitForEmails waits for expected number of emails to arrive
func (c *EmailTestClient) WaitForEmails(count int, timeout time.Duration) []sendria.Message {
	c.t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		messages, err := c.ListMessages(1, count+10) // Get a few extra in case
		if err != nil {
			c.t.Fatalf("Failed to list messages: %v", err)
		}

		if len(messages.Messages) >= count {
			return messages.Messages[:count]
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Timeout - show what we have
	messages, _ := c.ListMessages(1, 100)
	c.t.Fatalf("Timeout waiting for %d emails, got %d", count, len(messages.Messages))
	return nil
}

// AssertEmailSent verifies an email was sent to recipient with subject
func (c *EmailTestClient) AssertEmailSent(to, subject string) *sendria.Message {
	c.t.Helper()

	// Wait for the specific email to appear, checking periodically
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		messages, err := c.ListMessages(1, 10)
		if err != nil {
			c.t.Fatalf("Failed to list messages: %v", err)
		}
		
		for _, msg := range messages.Messages {
			// Check if this message matches
			recipientMatch := false
			for _, recipient := range msg.To {
				if recipient.Email == to {
					recipientMatch = true
					break
				}
			}
			
			if recipientMatch && msg.Subject == subject {
				return &msg
			}
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	// Not found - show what we have
	messages, _ := c.ListMessages(1, 100)
	c.t.Errorf("No email found with recipient=%s and subject=%s", to, subject)
	c.t.Logf("Available messages:")
	for _, msg := range messages.Messages {
		c.t.Logf("  - To: %v, Subject: %s", msg.To, msg.Subject)
	}
	c.t.FailNow()
	return nil
}

// AssertEmailContent verifies email contains expected text
func (c *EmailTestClient) AssertEmailContent(msg *sendria.Message, expectedTexts ...string) {
	c.t.Helper()

	body, err := c.GetMessagePlain(msg.ID)
	if err != nil {
		c.t.Fatalf("Failed to get message content: %v", err)
	}

	for _, text := range expectedTexts {
		if !strings.Contains(body, text) {
			c.t.Errorf("Email missing expected text: %q", text)
			c.t.Logf("Email body:\n%s", body)
		}
	}
}

// AssertNoEmailsSent verifies no emails were sent
func (c *EmailTestClient) AssertNoEmailsSent(waitTime time.Duration) {
	c.t.Helper()

	time.Sleep(waitTime)
	
	messages, err := c.ListMessages(1, 10)
	if err != nil {
		c.t.Fatalf("Failed to list messages: %v", err)
	}

	if len(messages.Messages) > 0 {
		c.t.Errorf("Expected no emails, but found %d", len(messages.Messages))
		for _, msg := range messages.Messages {
			c.t.Logf("  - To: %v, Subject: %s", msg.To, msg.Subject)
		}
	}
}

// GetLatestEmail returns the most recent email
func (c *EmailTestClient) GetLatestEmail() *sendria.Message {
	c.t.Helper()

	messages := c.WaitForEmails(1, 2*time.Second)
	if len(messages) == 0 {
		c.t.Fatal("No emails found")
		return nil
	}

	// Messages are returned newest first
	return &messages[0]
}

// ClearMessages deletes all messages (useful between test phases)
func (c *EmailTestClient) ClearMessages() {
	c.t.Helper()

	// Retry clearing messages up to 3 times to handle connection issues
	for i := 0; i < 3; i++ {
		if err := c.DeleteAllMessages(); err != nil {
			if i == 2 { // Last attempt
				c.t.Fatalf("Failed to clear messages after %d attempts: %v", i+1, err)
			}
			// Wait a bit before retrying
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}
	
	// Small delay to allow connection to stabilize
	time.Sleep(50 * time.Millisecond)
}

// CountEmails returns the current number of emails
func (c *EmailTestClient) CountEmails() int {
	c.t.Helper()

	messages, err := c.ListMessages(1, 100)
	if err != nil {
		c.t.Fatalf("Failed to count messages: %v", err)
	}

	return len(messages.Messages)
}

// FindEmail searches for an email by recipient and/or subject
func (c *EmailTestClient) FindEmail(to, subject string) *sendria.Message {
	c.t.Helper()

	messages, err := c.ListMessages(1, 100)
	if err != nil {
		c.t.Fatalf("Failed to list messages: %v", err)
	}

	for _, msg := range messages.Messages {
		// Match recipient if specified
		if to != "" {
			recipientMatch := false
			for _, recipient := range msg.To {
				if recipient.Email == to {
					recipientMatch = true
					break
				}
			}
			if !recipientMatch {
				continue
			}
		}

		// Match subject if specified
		if subject != "" && msg.Subject != subject {
			continue
		}

		return &msg
	}

	return nil
}

// ExtractLink extracts a URL matching the pattern from email body
func (c *EmailTestClient) ExtractLink(msg *sendria.Message, urlPattern string) string {
	c.t.Helper()

	body, err := c.GetMessagePlain(msg.ID)
	if err != nil {
		c.t.Fatalf("Failed to get message content: %v", err)
	}

	// Find URLs in the body
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.Contains(line, urlPattern) {
			// Extract the URL (simple approach - you might need regex for complex cases)
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "http") {
				// Find where URL ends (space or newline)
				endIdx := strings.IndexAny(line, " \t\r\n")
				if endIdx == -1 {
					return line
				}
				return line[:endIdx]
			}
		}
	}

	c.t.Errorf("No link found matching pattern: %s", urlPattern)
	c.t.Logf("Email body:\n%s", body)
	return ""
}

// DebugPrintEmail prints email details for debugging
func (c *EmailTestClient) DebugPrintEmail(msg *sendria.Message) {
	c.t.Helper()

	c.t.Logf("=== Email Debug ===")
	c.t.Logf("ID: %s", msg.ID)
	c.t.Logf("From: %v", msg.From)
	c.t.Logf("To: %v", msg.To)
	c.t.Logf("Subject: %s", msg.Subject)
	c.t.Logf("Created: %s", msg.CreatedAt)
	
	if body, err := c.GetMessagePlain(msg.ID); err == nil {
		c.t.Logf("Plain Body:\n%s", body)
	}
	
	if html, err := c.GetMessageHTML(msg.ID); err == nil && html != "" {
		c.t.Logf("HTML Body:\n%s", html)
	}
	
	c.t.Logf("==================")
}

// WaitFor is a generic helper to wait for a condition
func WaitFor(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// CreateTestEmail creates a test email with unique content
func CreateTestEmail(testName string) (subject, body string) {
	timestamp := time.Now().Unix()
	subject = fmt.Sprintf("Test Email - %s - %d", testName, timestamp)
	body = fmt.Sprintf("This is a test email for %s\nTimestamp: %d\nTest ID: %s",
		testName, timestamp, fmt.Sprintf("%s-%d", testName, timestamp))
	return subject, body
}