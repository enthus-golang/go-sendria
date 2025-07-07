package sendria

import (
	"github.com/enthus-golang/sendria/models"
)

// Re-export models for public API
type (
	Message     = models.Message
	MessageList = models.MessageList
	Recipient   = models.Recipient
	Part        = models.Part
	Attachment  = models.Attachment
)