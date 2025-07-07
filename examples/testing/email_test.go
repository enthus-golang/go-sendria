// Package testing_example shows how to use sendria for testing email functionality
package testing_example

import (
	"fmt"
	"net/smtp"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/enthus-golang/sendria/testhelpers"
)

// Example application functions that send emails
// In a real app, these would be in your application code

func SendWelcomeEmail(email string) error {
	// Validate email address
	if email == "" {
		return fmt.Errorf("email address cannot be empty")
	}
	
	from := "noreply@example.com"
	to := []string{email}
	subject := "Welcome to Our App!"
	body := fmt.Sprintf(`Hi there!

Welcome to our application. We're excited to have you on board.

To get started, please verify your email by clicking the link below:
https://example.com/verify?token=abc123def456&email=%s

If you have any questions, feel free to reach out to our support team.

Best regards,
The Team`, email)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to[0], subject, body))

	return smtp.SendMail("localhost:1025", nil, from, to, msg)
}

func SendPasswordResetEmail(email, resetToken string) error {
	from := "security@example.com"
	to := []string{email}
	subject := "Password Reset Request"
	body := fmt.Sprintf(`Hello,

We received a request to reset your password. If you didn't make this request, please ignore this email.

To reset your password, click the link below:
https://example.com/reset-password?token=%s

This link will expire in 24 hours.

For security reasons, this request was made from IP: 192.168.1.1

Best regards,
Security Team`, resetToken)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to[0], subject, body))

	return smtp.SendMail("localhost:1025", nil, from, to, msg)
}

func SendInvoiceEmail(email string, invoiceNumber string, amount float64) error {
	from := "billing@example.com"
	to := []string{email}
	subject := fmt.Sprintf("Invoice %s - $%.2f", invoiceNumber, amount)
	
	// Create multipart message with HTML
	boundary := "boundary42"
	body := fmt.Sprintf(`Your invoice %s for $%.2f is ready.
	
View it online: https://example.com/invoices/%s
	
Thank you for your business!`, invoiceNumber, amount, invoiceNumber)

	htmlBody := fmt.Sprintf(`<html>
<body>
<h2>Invoice %s</h2>
<p>Amount due: <strong>$%.2f</strong></p>
<p><a href="https://example.com/invoices/%s">View Invoice</a></p>
<p>Thank you for your business!</p>
</body>
</html>`, invoiceNumber, amount, invoiceNumber)

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
		"--%s--\r\n", from, to[0], subject, boundary, boundary, body, boundary, htmlBody, boundary))

	return smtp.SendMail("localhost:1025", nil, from, to, msg)
}

// Actual tests start here

func TestWelcomeEmail(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	// Send welcome email
	err := SendWelcomeEmail("newuser@example.com")
	if err != nil {
		t.Fatalf("Failed to send welcome email: %v", err)
	}

	// Verify email was sent
	msg := client.AssertEmailSent("newuser@example.com", "Welcome to Our App!")

	// Verify sender
	if msg.From[0].Email != "noreply@example.com" {
		t.Errorf("Expected from noreply@example.com, got %s", msg.From[0].Email)
	}

	// Verify content
	client.AssertEmailContent(msg,
		"Welcome to our application",
		"verify your email",
		"https://example.com/verify?token=",
	)

	// Extract and verify verification link
	body, _ := client.GetMessagePlain(msg.ID)
	verifyLinkRegex := regexp.MustCompile(`https://example\.com/verify\?token=([a-zA-Z0-9]+)&email=([^&\s]+)`)
	matches := verifyLinkRegex.FindStringSubmatch(body)
	
	if len(matches) != 3 {
		t.Fatal("Verification link not found or invalid format")
	}
	
	token := matches[1]
	emailParam := matches[2]
	
	if token == "" {
		t.Error("Verification token is empty")
	}
	
	if emailParam != "newuser@example.com" {
		t.Errorf("Email parameter in link incorrect: %s", emailParam)
	}
}

func TestPasswordResetEmail(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	resetToken := "reset-token-123456"
	
	// Send password reset email
	err := SendPasswordResetEmail("user@example.com", resetToken)
	if err != nil {
		t.Fatalf("Failed to send password reset email: %v", err)
	}

	// Verify email
	msg := client.AssertEmailSent("user@example.com", "Password Reset Request")

	// Verify it's from security team
	if msg.From[0].Email != "security@example.com" {
		t.Errorf("Expected from security@example.com, got %s", msg.From[0].Email)
	}

	// Verify content
	body, _ := client.GetMessagePlain(msg.ID)
	
	expectedTexts := []string{
		"reset your password",
		"This link will expire in 24 hours",
		"IP: 192.168.1.1",
		fmt.Sprintf("token=%s", resetToken),
	}
	
	for _, text := range expectedTexts {
		if !strings.Contains(body, text) {
			t.Errorf("Missing expected text: %s", text)
		}
	}

	// Verify the reset link
	if !strings.Contains(body, fmt.Sprintf("https://example.com/reset-password?token=%s", resetToken)) {
		t.Error("Reset link not found or incorrect")
	}
}

func TestInvoiceEmail(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	invoiceNumber := "INV-2024-001"
	amount := 299.99

	// Send invoice email
	err := SendInvoiceEmail("customer@example.com", invoiceNumber, amount)
	if err != nil {
		t.Fatalf("Failed to send invoice email: %v", err)
	}

	// Verify email
	expectedSubject := fmt.Sprintf("Invoice %s - $%.2f", invoiceNumber, amount)
	msg := client.AssertEmailSent("customer@example.com", expectedSubject)

	// Check plain text version
	plainText, _ := client.GetMessagePlain(msg.ID)
	if !strings.Contains(plainText, fmt.Sprintf("invoice %s for $%.2f", invoiceNumber, amount)) {
		t.Error("Plain text missing invoice details")
	}

	// Check HTML version
	html, _ := client.GetMessageHTML(msg.ID)
	if !strings.Contains(html, fmt.Sprintf("<h2>Invoice %s</h2>", invoiceNumber)) {
		t.Error("HTML missing invoice header")
	}
	if !strings.Contains(html, fmt.Sprintf("<strong>$%.2f</strong>", amount)) {
		t.Error("HTML missing formatted amount")
	}

	// Verify link in both versions
	expectedLink := fmt.Sprintf("https://example.com/invoices/%s", invoiceNumber)
	if !strings.Contains(plainText, expectedLink) {
		t.Error("Plain text missing invoice link")
	}
	if !strings.Contains(html, fmt.Sprintf(`href="%s"`, expectedLink)) {
		t.Error("HTML missing invoice link")
	}
}

func TestEmailNotSentOnError(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	// Simulate a scenario where email should NOT be sent
	// For example, invalid email address
	err := SendWelcomeEmail("")
	if err == nil {
		t.Error("Expected error for empty email address")
	}

	// Verify no email was sent
	client.AssertNoEmailsSent(100 * time.Millisecond)
}

func TestBulkEmailScenario(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	users := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
	}

	// Send welcome emails to multiple users
	for _, email := range users {
		err := SendWelcomeEmail(email)
		if err != nil {
			t.Errorf("Failed to send email to %s: %v", email, err)
		}
	}

	// Wait for all emails
	messages := client.WaitForEmails(len(users), 2*time.Second)

	// Verify each user got their email
	receivedEmails := make(map[string]bool)
	for _, msg := range messages {
		if msg.Subject == "Welcome to Our App!" && len(msg.To) > 0 {
			receivedEmails[msg.To[0].Email] = true
		}
	}

	for _, email := range users {
		if !receivedEmails[email] {
			t.Errorf("User %s did not receive welcome email", email)
		}
	}

	// Verify total count
	if count := client.CountEmails(); count != len(users) {
		t.Errorf("Expected %d emails, found %d", len(users), count)
	}
}

func TestEmailFlow(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	// Test a complete user flow
	userEmail := "testuser@example.com"

	// 1. User signs up - welcome email sent
	err := SendWelcomeEmail(userEmail)
	if err != nil {
		t.Fatal(err)
	}

	welcomeMsg := client.AssertEmailSent(userEmail, "Welcome to Our App!")
	
	// Extract verification token (in real app, you'd verify the email here)
	body, _ := client.GetMessagePlain(welcomeMsg.ID)
	if !strings.Contains(body, "verify?token=") {
		t.Fatal("Verification link not found")
	}

	// 2. User forgets password
	client.ClearMessages() // Clear previous emails
	
	err = SendPasswordResetEmail(userEmail, "reset-abc-123")
	if err != nil {
		t.Fatal(err)
	}

	resetMsg := client.AssertEmailSent(userEmail, "Password Reset Request")
	
	// Verify reset email content
	resetBody, _ := client.GetMessagePlain(resetMsg.ID)
	if !strings.Contains(resetBody, "reset-abc-123") {
		t.Error("Reset token not found in password reset email")
	}

	// 3. User makes a purchase
	client.ClearMessages()
	
	err = SendInvoiceEmail(userEmail, "INV-001", 99.99)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for and verify invoice email
	invoiceMsg := client.AssertEmailSent(userEmail, "Invoice INV-001 - $99.99")
	
	// Verify invoice content
	invoiceBody, _ := client.GetMessagePlain(invoiceMsg.ID)
	if !strings.Contains(invoiceBody, "invoice INV-001") {
		t.Error("Invoice number not found in email")
	}
	if !strings.Contains(invoiceBody, "$99.99") {
		t.Error("Invoice amount not found in email")
	}
}

// Table-driven test example
func TestEmailTemplates(t *testing.T) {
	client := testhelpers.NewEmailTestClient(t)

	tests := []struct {
		name     string
		sendFunc func() error
		to       string
		subject  string
		contains []string
		from     string
	}{
		{
			name: "welcome email",
			sendFunc: func() error {
				return SendWelcomeEmail("test@example.com")
			},
			to:      "test@example.com",
			subject: "Welcome to Our App!",
			contains: []string{
				"Welcome to our application",
				"verify your email",
				"support team",
			},
			from: "noreply@example.com",
		},
		{
			name: "password reset",
			sendFunc: func() error {
				return SendPasswordResetEmail("test@example.com", "token123")
			},
			to:      "test@example.com",
			subject: "Password Reset Request",
			contains: []string{
				"reset your password",
				"expire in 24 hours",
				"token=token123",
			},
			from: "security@example.com",
		},
		{
			name: "invoice",
			sendFunc: func() error {
				return SendInvoiceEmail("test@example.com", "INV-123", 50.00)
			},
			to:      "test@example.com",
			subject: "Invoice INV-123 - $50.00",
			contains: []string{
				"invoice INV-123",
				"$50.00",
				"View it online",
			},
			from: "billing@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear messages for each test
			client.ClearMessages()

			// Send email
			if err := tt.sendFunc(); err != nil {
				t.Fatal(err)
			}

			// Verify email
			msg := client.AssertEmailSent(tt.to, tt.subject)

			// Check sender
			if msg.From[0].Email != tt.from {
				t.Errorf("Expected from %s, got %s", tt.from, msg.From[0].Email)
			}

			// Check content
			client.AssertEmailContent(msg, tt.contains...)
		})
	}
}