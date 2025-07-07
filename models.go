package sendria

import (
	"github.com/enthus-golang/go-sendria/internal/models"
)

// Re-export models for public API
type (
	Message     = models.Message
	MessageList = models.MessageList
	Recipient   = models.Recipient
	Part        = models.Part
	Attachment  = models.Attachment
)