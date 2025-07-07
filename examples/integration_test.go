package main

import (
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/enthus-golang/go-sendria"
)

func main() {
	// Create Sendria client
	client := sendria.NewClient(sendria.Config{})

	// Clear all messages before starting
	fmt.Println("Clearing all messages...")
	if err := client.DeleteAllMessages(); err != nil {
		log.Fatalf("Error clearing messages: %v", err)
	}

	// Send a test email
	fmt.Println("Sending test email...")
	if err := sendTestEmail(); err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	// Wait for Sendria to receive the email
	time.Sleep(500 * time.Millisecond)

	// Check if email was received
	fmt.Println("Checking for received email...")
	messages, err := client.ListMessages(1, 10)
	if err != nil {
		log.Fatalf("Error listing messages: %v", err)
	}

	if len(messages.Messages) == 0 {
		log.Fatal("No messages received - is Sendria running?")
	}

	fmt.Printf("Success! Received %d message(s)\n", len(messages.Messages))

	// Verify the email content
	msg := messages.Messages[0]
	fmt.Printf("\nVerifying email content:\n")
	fmt.Printf("  Subject: %s\n", msg.Subject)
	fmt.Printf("  From: %s\n", msg.From[0].Email)
	fmt.Printf("  To: %s\n", msg.To[0].Email)

	// Get and display the plain text content
	plainText, err := client.GetMessagePlain(msg.ID)
	if err != nil {
		log.Printf("Error getting plain text: %v", err)
	} else {
		fmt.Printf("  Body: %s\n", plainText)
	}

	// Cleanup - delete the test message
	fmt.Printf("\nCleaning up - deleting message %s...\n", msg.ID)
	if err := client.DeleteMessage(msg.ID); err != nil {
		log.Printf("Error deleting message: %v", err)
	} else {
		fmt.Println("Message deleted successfully")
	}
}

// sendTestEmail sends a test email to Sendria
func sendTestEmail() error {
	from := "test@example.com"
	to := []string{"recipient@example.com"}
	subject := "Test Email from go-sendria"
	body := "This is a test email sent to Sendria for integration testing."

	// Format the email
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to[0])
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n" + body

	// Connect to Sendria SMTP server
	auth := smtp.PlainAuth("", "", "", "localhost")
	err := smtp.SendMail("localhost:1025", auth, from, to, []byte(message))
	return err
}