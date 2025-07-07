// Package models contains the data structures used by the Sendria API.
package models

import (
	"encoding/json"
	"time"
)

// Message represents an email message in Sendria
type Message struct {
	ID          string       `json:"id"`
	Subject     string       `json:"subject"`
	To          []Recipient  `json:"to"`
	From        []Recipient  `json:"from"`
	CreatedAt   time.Time    `json:"created_at"`
	Size        int          `json:"size"`
	Type        string       `json:"type"`
	Source      string       `json:"source,omitempty"`
	Parts       []Part       `json:"parts,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Recipient represents an email recipient
type Recipient struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Part represents a message part (plain text, HTML, etc.)
type Part struct {
	Type        string `json:"type"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
	Size        int    `json:"size"`
}

// Attachment represents an email attachment
type Attachment struct {
	CID         string `json:"cid"`
	Type        string `json:"type"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

// MessageList represents a paginated list of messages
type MessageList struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	Code string          `json:"code"`
	Data json.RawMessage `json:"data"`
	Meta *APIMeta        `json:"meta,omitempty"`
}

// APIMeta represents metadata in API responses
type APIMeta struct {
	PagesTotal int `json:"pages_total"`
}

// APIMessage represents a message in the API response
type APIMessage struct {
	ID                   int       `json:"id"`
	SenderEnvelope       string    `json:"sender_envelope"`
	SenderMessage        string    `json:"sender_message"`
	RecipientsEnvelope   []string  `json:"recipients_envelope"`
	RecipientsMessageTo  []string  `json:"recipients_message_to"`
	RecipientsMessageCC  []string  `json:"recipients_message_cc"`
	RecipientsMessageBCC []string  `json:"recipients_message_bcc"`
	Subject              string    `json:"subject"`
	Source               string    `json:"source"`
	Size                 int       `json:"size"`
	Type                 string    `json:"type"`
	Peer                 string    `json:"peer"`
	CreatedAt            string    `json:"created_at"`
}