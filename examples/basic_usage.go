package main

import (
	"fmt"
	"log"

	"github.com/enthus-golang/go-sendria"
)

func main() {
	// Create a new client with default settings
	client := sendria.NewClient(sendria.Config{})

	// List all messages
	fmt.Println("Fetching messages...")
	messages, err := client.ListMessages(1, 10)
	if err != nil {
		log.Fatalf("Error listing messages: %v", err)
	}

	fmt.Printf("Found %d messages (Total: %d)\n\n", len(messages.Messages), messages.Total)

	// Display each message
	for _, msg := range messages.Messages {
		fmt.Printf("Message ID: %s\n", msg.ID)
		fmt.Printf("Subject: %s\n", msg.Subject)
		
		if len(msg.From) > 0 {
			fmt.Printf("From: %s <%s>\n", msg.From[0].Name, msg.From[0].Email)
		}
		
		if len(msg.To) > 0 {
			fmt.Printf("To: %s <%s>\n", msg.To[0].Name, msg.To[0].Email)
		}
		
		fmt.Printf("Date: %s\n", msg.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Size: %d bytes\n", msg.Size)
		fmt.Printf("Type: %s\n", msg.Type)
		fmt.Println("---")
	}

	// Get details of the first message if any exist
	if len(messages.Messages) > 0 {
		firstMsg := messages.Messages[0]
		fmt.Printf("\nFetching details for message: %s\n", firstMsg.ID)

		// Get full message details
		fullMsg, err := client.GetMessage(firstMsg.ID)
		if err != nil {
			log.Printf("Error getting message details: %v", err)
			return
		}

		// Display parts
		fmt.Printf("\nMessage has %d parts:\n", len(fullMsg.Parts))
		for i, part := range fullMsg.Parts {
			fmt.Printf("  Part %d: %s (%d bytes)\n", i+1, part.ContentType, part.Size)
		}

		// Display attachments
		if len(fullMsg.Attachments) > 0 {
			fmt.Printf("\nMessage has %d attachments:\n", len(fullMsg.Attachments))
			for _, att := range fullMsg.Attachments {
				fmt.Printf("  - %s (%s, %d bytes)\n", att.Filename, att.ContentType, att.Size)
			}
		}

		// Get plain text content
		plainText, err := client.GetMessagePlain(firstMsg.ID)
		if err != nil {
			log.Printf("Error getting plain text: %v", err)
		} else {
			fmt.Printf("\nPlain text content:\n%s\n", plainText)
		}
	}
}