package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enthus-golang/sendria"
)

func main() {
	// Create client with custom configuration if needed
	baseURL := os.Getenv("SENDRIA_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1080"
	}
	
	// Build options based on environment variables
	var opts []sendria.Option
	if username := os.Getenv("SENDRIA_USERNAME"); username != "" {
		password := os.Getenv("SENDRIA_PASSWORD")
		opts = append(opts, sendria.WithBasicAuth(username, password))
	}

	client := sendria.NewClient(baseURL, opts...)

	// Keep track of processed messages
	processedIDs := make(map[string]bool)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Starting email monitor...")
	fmt.Println("Press Ctrl+C to stop")

	// Create a ticker for periodic checks
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\nShutting down email monitor...")
			return
		case <-ticker.C:
			checkNewMessages(client, processedIDs)
		}
	}
}

func checkNewMessages(client *sendria.Client, processedIDs map[string]bool) {
	messages, err := client.ListMessages(1, 50)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		return
	}

	for _, msg := range messages.Messages {
		if !processedIDs[msg.ID] {
			processedIDs[msg.ID] = true
			processNewMessage(client, msg)
		}
	}
}

func processNewMessage(client *sendria.Client, msg sendria.Message) {
	fmt.Printf("\n[%s] New message received!\n", time.Now().Format("15:04:05"))
	fmt.Printf("  ID: %s\n", msg.ID)
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

	// Display content preview
	if plainText, err := client.GetMessagePlain(msg.ID); err == nil && plainText != "" {
		preview := plainText
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("  Preview: %s\n", preview)
	}

	// Show attachment info
	if len(fullMsg.Attachments) > 0 {
		fmt.Printf("  Attachments: %d file(s)\n", len(fullMsg.Attachments))
		for _, att := range fullMsg.Attachments {
			fmt.Printf("    - %s (%s, %d bytes)\n", att.Filename, att.ContentType, att.Size)
		}
	}

	fmt.Println("  ---")
}