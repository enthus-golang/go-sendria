// This example shows how to monitor emails in real-time during development/testing
// It demonstrates email pattern detection and content analysis useful for testing
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/enthus-golang/sendria"
)

func main() {
	// Get Sendria URL from environment or use default
	baseURL := os.Getenv("SENDRIA_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1080"
	}

	fmt.Printf("📧 Email Test Monitor - Connected to %s\n", baseURL)
	fmt.Println("Monitoring for test emails with pattern detection...")
	fmt.Println("Detects: verification links, password resets, welcome emails, invoices")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("---")
	
	// Build options based on environment variables
	var opts []sendria.Option
	if username := os.Getenv("SENDRIA_USERNAME"); username != "" {
		password := os.Getenv("SENDRIA_PASSWORD")
		opts = append(opts, sendria.WithBasicAuth(username, password))
	}

	client := sendria.NewClient(baseURL, opts...)

	// Keep track of processed messages and statistics
	processedIDs := make(map[string]bool)
	stats := &EmailStats{
		total:         0,
		verification:  0,
		passwordReset: 0,
		welcome:       0,
		invoice:       0,
		other:         0,
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create a ticker for periodic checks
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Initial check
	checkNewMessages(client, processedIDs, stats)

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\nStopping email monitor...")
			// Show summary
			fmt.Println("\n=== Email Statistics ===")
			fmt.Printf("Total emails monitored: %d\n", stats.total)
			fmt.Printf("  Verification emails: %d\n", stats.verification)
			fmt.Printf("  Password resets: %d\n", stats.passwordReset)
			fmt.Printf("  Welcome emails: %d\n", stats.welcome)
			fmt.Printf("  Invoices: %d\n", stats.invoice)
			fmt.Printf("  Other: %d\n", stats.other)
			return
		case <-ticker.C:
			checkNewMessages(client, processedIDs, stats)
		}
	}
}

type EmailStats struct {
	total         int
	verification  int
	passwordReset int
	welcome       int
	invoice       int
	other         int
}

func checkNewMessages(client *sendria.Client, processedIDs map[string]bool, stats *EmailStats) {
	messages, err := client.ListMessages(1, 50)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		return
	}

	for _, msg := range messages.Messages {
		if !processedIDs[msg.ID] {
			processedIDs[msg.ID] = true
			stats.total++
			processNewMessage(client, msg, stats)
		}
	}
}

func processNewMessage(client *sendria.Client, msg sendria.Message, stats *EmailStats) {
	emailType := detectEmailType(msg, client)
	updateStats(emailType, stats)
	
	fmt.Printf("\n[%s] New %s email!\n", time.Now().Format("15:04:05"), emailType)
	fmt.Printf("  Subject: %s\n", msg.Subject)
	
	if len(msg.From) > 0 {
		fmt.Printf("  From: %s <%s>\n", msg.From[0].Name, msg.From[0].Email)
	}
	
	if len(msg.To) > 0 {
		fmt.Printf("  To: %s <%s>\n", msg.To[0].Name, msg.To[0].Email)
	}

	// Get full message details
	fullMsg, err := client.GetMessage(msg.ID)
	if err != nil {
		log.Printf("  Error getting message details: %v", err)
		return
	}

	// Analyze email content
	if plainText, err := client.GetMessagePlain(msg.ID); err == nil && plainText != "" {
		// Extract and display important information based on email type
		switch emailType {
		case "verification":
			if link := extractVerificationLink(plainText); link != "" {
				fmt.Printf("  ✓ Verification link: %s\n", link)
			}
		case "password-reset":
			if link := extractResetLink(plainText); link != "" {
				fmt.Printf("  🔐 Reset link: %s\n", link)
			}
			if token := extractResetToken(plainText); token != "" {
				fmt.Printf("  🔑 Reset token: %s\n", token)
			}
		case "invoice":
			if invoiceNum := extractInvoiceNumber(plainText); invoiceNum != "" {
				fmt.Printf("  📄 Invoice number: %s\n", invoiceNum)
			}
			if amount := extractAmount(plainText); amount != "" {
				fmt.Printf("  💰 Amount: %s\n", amount)
			}
		case "welcome":
			if username := extractUsername(plainText); username != "" {
				fmt.Printf("  👤 Username: %s\n", username)
			}
		}
		
		// Show content preview
		preview := plainText
		if len(preview) > 150 {
			preview = preview[:150] + "..."
		}
		fmt.Printf("  Preview: %s\n", strings.ReplaceAll(preview, "\n", " "))
	}

	// Show attachment info
	if len(fullMsg.Attachments) > 0 {
		fmt.Printf("  📎 Attachments: %d file(s)\n", len(fullMsg.Attachments))
		for _, att := range fullMsg.Attachments {
			fmt.Printf("     - %s (%s, %d bytes)\n", att.Filename, att.ContentType, att.Size)
		}
	}

	fmt.Println("  ---")
}

func detectEmailType(msg sendria.Message, client *sendria.Client) string {
	// Check subject patterns
	subjectLower := strings.ToLower(msg.Subject)
	
	if strings.Contains(subjectLower, "verify") || strings.Contains(subjectLower, "confirm") {
		return "verification"
	}
	if strings.Contains(subjectLower, "password") || strings.Contains(subjectLower, "reset") {
		return "password-reset"
	}
	if strings.Contains(subjectLower, "welcome") || strings.Contains(subjectLower, "thanks for signing up") {
		return "welcome"
	}
	if strings.Contains(subjectLower, "invoice") || strings.Contains(subjectLower, "receipt") || strings.Contains(subjectLower, "payment") {
		return "invoice"
	}
	
	// Check content if subject doesn't match
	if plainText, err := client.GetMessagePlain(msg.ID); err == nil {
		contentLower := strings.ToLower(plainText)
		if strings.Contains(contentLower, "verify your email") || strings.Contains(contentLower, "confirm your email") {
			return "verification"
		}
		if strings.Contains(contentLower, "reset your password") || strings.Contains(contentLower, "forgot your password") {
			return "password-reset"
		}
		if strings.Contains(contentLower, "welcome to") || strings.Contains(contentLower, "thank you for joining") {
			return "welcome"
		}
	}
	
	return "other"
}

func updateStats(emailType string, stats *EmailStats) {
	switch emailType {
	case "verification":
		stats.verification++
	case "password-reset":
		stats.passwordReset++
	case "welcome":
		stats.welcome++
	case "invoice":
		stats.invoice++
	default:
		stats.other++
	}
}

func extractVerificationLink(content string) string {
	// Common verification link patterns
	patterns := []string{
		`https?://[^\s]+/verify[^\s]*`,
		`https?://[^\s]+/confirm[^\s]*`,
		`https?://[^\s]+/activate[^\s]*`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(content); match != "" {
			return match
		}
	}
	return ""
}

func extractResetLink(content string) string {
	// Common reset link patterns
	patterns := []string{
		`https?://[^\s]+/reset[^\s]*`,
		`https?://[^\s]+/password[^\s]*reset[^\s]*`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(content); match != "" {
			return match
		}
	}
	return ""
}

func extractResetToken(content string) string {
	// Look for reset tokens in various formats
	patterns := []string{
		`token=([a-zA-Z0-9\-_]+)`,
		`code:\s*([A-Z0-9]{6,8})`,
		`reset code:\s*([A-Z0-9]{6,8})`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func extractInvoiceNumber(content string) string {
	// Common invoice number patterns
	patterns := []string{
		`Invoice\s*#?\s*([A-Z0-9\-]+)`,
		`Order\s*#?\s*([A-Z0-9\-]+)`,
		`Receipt\s*#?\s*([A-Z0-9\-]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func extractAmount(content string) string {
	// Common amount patterns
	patterns := []string{
		`\$([0-9,]+\.?[0-9]*)`,
		`USD\s*([0-9,]+\.?[0-9]*)`,
		`Total:\s*\$?([0-9,]+\.?[0-9]*)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			return "$" + matches[1]
		}
	}
	return ""
}

func extractUsername(content string) string {
	// Common username patterns in welcome emails
	patterns := []string{
		`Hi\s+([^,\n]+)`,
		`Hello\s+([^,\n]+)`,
		`Dear\s+([^,\n]+)`,
		`Welcome\s+([^,\n]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			username := strings.TrimSpace(matches[1])
			// Filter out generic terms
			if username != "there" && username != "user" && username != "customer" {
				return username
			}
		}
	}
	return ""
}