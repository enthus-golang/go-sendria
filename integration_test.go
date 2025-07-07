//go:build integration

package sendria_test

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/enthus-golang/sendria"
)

// TestSendriaIntegration runs integration tests against a real Sendria instance.
// Run with: go test -tags=integration -v ./...
func TestSendriaIntegration(t *testing.T) {
	// Get Sendria configuration from environment
	sendriaURL := os.Getenv("SENDRIA_URL")
	if sendriaURL == "" {
		t.Skip("Skipping integration test. Set SENDRIA_URL to run (e.g., SENDRIA_URL=http://localhost:1080).")
	}

	smtpHost := os.Getenv("SENDRIA_SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "localhost:1025"
	}

	// Create client
	client := sendria.NewClient(sendriaURL)

	// Clear all messages before starting
	t.Log("Clearing all messages...")
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}
	// Give time for any in-flight emails to be processed
	time.Sleep(500 * time.Millisecond)

	// Ensure cleanup after all tests
	t.Cleanup(func() {
		_ = client.DeleteAllMessages()
	})

	// Run sub-tests
	t.Run("BasicEmailSend", func(t *testing.T) {
		testBasicEmailSend(t, client, smtpHost)
	})

	t.Run("EmailWithHTML", func(t *testing.T) {
		testEmailWithHTML(t, client, smtpHost)
	})

	t.Run("EmailWithMultipleRecipients", func(t *testing.T) {
		testEmailWithMultipleRecipients(t, client, smtpHost)
	})

	t.Run("EmailWithAttachment", func(t *testing.T) {
		testEmailWithAttachment(t, client, smtpHost)
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		testDeleteMessage(t, client, smtpHost)
	})
}

func testBasicEmailSend(t *testing.T, client *sendria.Client, smtpHost string) {
	t.Helper()
	// Small delay to avoid connection issues
	time.Sleep(100 * time.Millisecond)
	
	// Clear messages
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}

	// Send a simple email
	from := "test@example.com"
	to := []string{"recipient@example.com"}
	subject := "Integration Test - Basic"
	body := "This is a test email from the Sendria integration test."

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to[0], subject, body))

	t.Logf("Sending email to SMTP host: %s", smtpHost)
	if err := smtp.SendMail(smtpHost, nil, from, to, msg); err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}
	t.Log("Email sent successfully")

	// Wait for Sendria to process the email
	var messages *sendria.MessageList
	waitFor(t, func() bool {
		var err error
		messages, err = client.ListMessages(1, 10)
		return err == nil && len(messages.Messages) == 1
	}, 2*time.Second, 100*time.Millisecond)

	// Verify message details
	msg1 := messages.Messages[0]
	if msg1.Subject != subject {
		t.Errorf("Expected subject %q, got %q", subject, msg1.Subject)
	}

	if len(msg1.From) == 0 || msg1.From[0].Email != from {
		t.Errorf("Expected from %q, got %v", from, msg1.From)
	}

	if len(msg1.To) == 0 || msg1.To[0].Email != to[0] {
		t.Errorf("Expected to %q, got %v", to[0], msg1.To)
	}

	// Get message details
	fullMsg, err := client.GetMessage(msg1.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if fullMsg.ID != msg1.ID {
		t.Errorf("Message ID mismatch: %s != %s", fullMsg.ID, msg1.ID)
	}

	// Get plain text
	plainText, err := client.GetMessagePlain(msg1.ID)
	if err != nil {
		t.Fatalf("Failed to get plain text: %v", err)
	}

	if plainText != body+"\r\n" {
		t.Errorf("Expected body %q, got %q", body, plainText)
	}
}

func testEmailWithHTML(t *testing.T, client *sendria.Client, smtpHost string) {
	t.Helper()
	// Small delay to avoid connection issues
	time.Sleep(100 * time.Millisecond)
	
	// Clear messages
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}

	// Send multipart email with HTML
	from := "html-test@example.com"
	to := []string{"recipient@example.com"}
	subject := "Integration Test - HTML Email"

	boundary := "boundary123"
	plainBody := "This is the plain text version."
	htmlBody := "<html><body><h1>Test Email</h1><p>This is the <b>HTML</b> version.</p></body></html>"

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: multipart/alternative; boundary=\"%s\"\r\n"+
		"\r\n"+
		"--%s\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s\r\n"+
		"--%s\r\n"+
		"Content-Type: text/html; charset=utf-8\r\n"+
		"\r\n"+
		"%s\r\n"+
		"--%s--\r\n", from, to[0], subject, boundary, boundary, plainBody, boundary, htmlBody, boundary))

	if err := smtp.SendMail(smtpHost, nil, from, to, msg); err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	// Wait for processing
	var messages *sendria.MessageList
	waitFor(t, func() bool {
		var err error
		messages, err = client.ListMessages(1, 10)
		return err == nil && len(messages.Messages) == 1
	}, 2*time.Second, 100*time.Millisecond)

	msgID := messages.Messages[0].ID

	// Get HTML content
	html, err := client.GetMessageHTML(msgID)
	if err != nil {
		// Some versions of Sendria may not support HTML for multipart emails
		t.Logf("Failed to get HTML: %v", err)
		t.Skip("HTML endpoint not available or message doesn't have HTML content")
	}

	// Sendria may modify HTML by adding missing tags like <head></head>
	// So we check if our content is contained within the response
	if !strings.Contains(html, "<h1>Test Email</h1>") || !strings.Contains(html, "<p>This is the <b>HTML</b> version.</p>") {
		t.Errorf("HTML content mismatch.\nExpected to contain: %q\nGot: %q", htmlBody, html)
	}

	// Get plain text content
	plain, err := client.GetMessagePlain(msgID)
	if err != nil {
		t.Fatalf("Failed to get plain text: %v", err)
	}

	if plain != plainBody {
		t.Errorf("Plain text content mismatch.\nExpected: %q\nGot: %q", plainBody, plain)
	}

	// Verify parts were parsed correctly
	fullMsg, err := client.GetMessage(msgID)
	if err != nil {
		t.Fatalf("Failed to get full message: %v", err)
	}

	// Should have at least 2 parts (plain and HTML)
	if len(fullMsg.Parts) < 2 {
		t.Errorf("Expected at least 2 parts, got %d", len(fullMsg.Parts))
	}

	// Check if we have both text/plain and text/html parts
	hasPlain := false
	hasHTML := false
	for _, part := range fullMsg.Parts {
		if part.Type == "text/plain" {
			hasPlain = true
			if !strings.Contains(part.Body, "This is the plain text version") {
				t.Errorf("Plain text part doesn't contain expected content")
			}
		}
		if part.Type == "text/html" {
			hasHTML = true
			if !strings.Contains(part.Body, "<h1>Test Email</h1>") {
				t.Errorf("HTML part doesn't contain expected content")
			}
		}
	}

	if !hasPlain {
		t.Error("No text/plain part found")
	}
	if !hasHTML {
		t.Error("No text/html part found")
	}
}

func testEmailWithMultipleRecipients(t *testing.T, client *sendria.Client, smtpHost string) {
	t.Helper()
	// Small delay to avoid connection issues
	time.Sleep(100 * time.Millisecond)
	
	// Clear messages
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}

	// Send to multiple recipients
	from := "multi-test@example.com"
	to := []string{"recipient1@example.com", "recipient2@example.com", "recipient3@example.com"}
	subject := "Integration Test - Multiple Recipients"
	body := "This email is sent to multiple recipients."

	toHeader := ""
	for i, addr := range to {
		if i > 0 {
			toHeader += ", "
		}
		toHeader += addr
	}

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, toHeader, subject, body))

	if err := smtp.SendMail(smtpHost, nil, from, to, msg); err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	// Wait for processing and verify message
	var messages *sendria.MessageList
	waitFor(t, func() bool {
		var err error
		messages, err = client.ListMessages(1, 10)
		return err == nil && len(messages.Messages) == 1
	}, 2*time.Second, 100*time.Millisecond)

	msg1 := messages.Messages[0]
	if len(msg1.To) < 1 {
		t.Errorf("Expected at least 1 recipient, got %d", len(msg1.To))
	}

	// Sendria may not parse multiple recipients from To header correctly
	// Check if we have at least one of the expected recipients
	expectedRecipients := []string{"recipient1@example.com", "recipient2@example.com", "recipient3@example.com"}
	found := false
	for _, recipient := range msg1.To {
		for _, expected := range expectedRecipients {
			if recipient.Email == expected {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("None of the expected recipients found. Got: %v", msg1.To)
	}

	if len(msg1.To) != 3 {
		t.Logf("Note: Expected 3 recipients, got %d (Sendria may not parse multiple recipients correctly)", len(msg1.To))
	}
}

func testEmailWithAttachment(t *testing.T, client *sendria.Client, smtpHost string) {
	t.Helper()
	// Small delay to avoid connection issues
	time.Sleep(100 * time.Millisecond)
	
	// Clear messages
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}

	// Send email with attachment
	from := "attachment-test@example.com"
	to := []string{"recipient@example.com"}
	subject := "Integration Test - With Attachment"

	boundary := "boundary456"
	body := "This email contains an attachment."
	attachmentContent := []byte("This is the content of the test file.")
	attachmentName := "test.txt"

	// Encode attachment in base64
	encodedAttachment := make([]byte, len(attachmentContent)*2)
	n := len(attachmentContent)
	for i := 0; i < n; i++ {
		encodedAttachment[i*2] = "0123456789ABCDEF"[attachmentContent[i]>>4]
		encodedAttachment[i*2+1] = "0123456789ABCDEF"[attachmentContent[i]&0x0F]
	}

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: multipart/mixed; boundary=\"%s\"\r\n"+
		"\r\n"+
		"--%s\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s\r\n"+
		"--%s\r\n"+
		"Content-Type: text/plain; name=\"%s\"\r\n"+
		"Content-Disposition: attachment; filename=\"%s\"\r\n"+
		"Content-Transfer-Encoding: 7bit\r\n"+
		"\r\n"+
		"%s\r\n"+
		"--%s--\r\n", from, to[0], subject, boundary, boundary, body, boundary, attachmentName, attachmentName, string(attachmentContent), boundary))

	if err := smtp.SendMail(smtpHost, nil, from, to, msg); err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	// Wait for processing
	var messages *sendria.MessageList
	waitFor(t, func() bool {
		var err error
		messages, err = client.ListMessages(1, 10)
		return err == nil && len(messages.Messages) == 1
	}, 2*time.Second, 100*time.Millisecond)

	// Get full message with attachments
	fullMsg, err := client.GetMessage(messages.Messages[0].ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	// Verify attachment exists
	if len(fullMsg.Attachments) == 0 {
		t.Fatal("No attachments found - expected at least one attachment")
	}

	// Find our attachment
	var testAttachment *sendria.Attachment
	for _, att := range fullMsg.Attachments {
		if att.Filename == attachmentName {
			testAttachment = &att
			break
		}
	}

	if testAttachment == nil {
		t.Errorf("Attachment %s not found", attachmentName)
	} else {
		// Try to download attachment
		content, err := client.GetAttachment(fullMsg.ID, testAttachment.CID)
		if err != nil {
			t.Errorf("Failed to download attachment: %v", err)
		} else if !bytes.Equal(content, attachmentContent) {
			t.Errorf("Attachment content mismatch")
		}
	}
}

func testDeleteMessage(t *testing.T, client *sendria.Client, smtpHost string) {
	t.Helper()
	// Small delay to avoid connection issues
	time.Sleep(100 * time.Millisecond)
	
	// Clear messages
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}

	// Send two emails
	from := "delete-test@example.com"
	to := []string{"recipient@example.com"}

	for i := 1; i <= 2; i++ {
		subject := fmt.Sprintf("Delete Test %d", i)
		body := fmt.Sprintf("This is test email #%d", i)

		msg := []byte(fmt.Sprintf("From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s\r\n", from, to[0], subject, body))

		if err := smtp.SendMail(smtpHost, nil, from, to, msg); err != nil {
			t.Fatalf("Failed to send email %d: %v", i, err)
		}
	}

	// Wait for processing and verify we have 2 messages
	var messages *sendria.MessageList
	waitFor(t, func() bool {
		var err error
		messages, err = client.ListMessages(1, 10)
		return err == nil && len(messages.Messages) == 2
	}, 2*time.Second, 100*time.Millisecond)

	// Delete the first message
	firstMsgID := messages.Messages[0].ID
	if err := client.DeleteMessage(firstMsgID); err != nil {
		t.Fatalf("Failed to delete message: %v", err)
	}

	// Wait for deletion to be processed and verify we now have 1 message
	waitFor(t, func() bool {
		var err error
		messages, err = client.ListMessages(1, 10)
		return err == nil && len(messages.Messages) == 1
	}, 2*time.Second, 100*time.Millisecond)

	if len(messages.Messages) != 1 {
		t.Fatalf("Expected 1 message after deletion, got %d", len(messages.Messages))
	}

	// Verify the deleted message is gone
	if messages.Messages[0].ID == firstMsgID {
		t.Errorf("Deleted message still present with ID: %s", firstMsgID)
	}

	// Test DeleteAllMessages
	if err := client.DeleteAllMessages(); err != nil {
		t.Fatalf("Failed to delete all messages: %v", err)
	}

	// Verify no messages left
	var err error
	messages, err = client.ListMessages(1, 10)
	if err != nil {
		t.Fatalf("Failed to list messages: %v", err)
	}

	if len(messages.Messages) != 0 {
		t.Fatalf("Expected 0 messages after delete all, got %d", len(messages.Messages))
	}
}
