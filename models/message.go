package models

import "time"

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