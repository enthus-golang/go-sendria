package sendria

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/enthus-golang/sendria/models"
)

// parseMIMEMessage parses the raw email source into parts and attachments
func parseMIMEMessage(source string) ([]models.Part, []models.Attachment, error) {
	// Parse the email message
	msg, err := mail.ReadMessage(strings.NewReader(source))
	if err != nil {
		return nil, nil, fmt.Errorf("parsing email message: %w", err)
	}

	// Get content type
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		// Simple message with no MIME parts
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("reading message body: %w", err)
		}

		part := models.Part{
			Type:        "text/plain",
			ContentType: "text/plain",
			Body:        string(body),
			Size:        len(body),
		}

		return []models.Part{part}, nil, nil
	}

	// Parse the content type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing content type: %w", err)
	}

	var parts []models.Part
	var attachments []models.Attachment

	if strings.HasPrefix(mediaType, "multipart/") {
		// Handle multipart messages
		mr := multipart.NewReader(msg.Body, params["boundary"])
		if err := parseMultipart(mr, &parts, &attachments); err != nil {
			return nil, nil, fmt.Errorf("parsing multipart message: %w", err)
		}
	} else {
		// Single part message
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("reading message body: %w", err)
		}

		// Decode if needed
		encoding := msg.Header.Get("Content-Transfer-Encoding")
		content := decodeContent(body, encoding)

		part := models.Part{
			Type:        mediaType,
			ContentType: contentType,
			Body:        content,
			Size:        len(content),
		}

		parts = append(parts, part)
	}

	return parts, attachments, nil
}

// parseMultipart recursively parses multipart messages
func parseMultipart(mr *multipart.Reader, parts *[]models.Part, attachments *[]models.Attachment) error {
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading part: %w", err)
		}

		contentType := p.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/plain"
		}

		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			// Default to text/plain if parsing fails
			mediaType = "text/plain"
			params = make(map[string]string)
		}

		// Read the part content
		partContent, err := io.ReadAll(p)
		if err != nil {
			return fmt.Errorf("reading part content: %w", err)
		}

		// Handle nested multipart
		if strings.HasPrefix(mediaType, "multipart/") {
			nestedReader := multipart.NewReader(bytes.NewReader(partContent), params["boundary"])
			if err := parseMultipart(nestedReader, parts, attachments); err != nil {
				return fmt.Errorf("parsing nested multipart: %w", err)
			}
			continue
		}

		// Get content disposition
		disposition := p.Header.Get("Content-Disposition")
		filename := p.FileName()
		contentID := p.Header.Get("Content-ID")

		// Clean up Content-ID (remove < and >)
		if contentID != "" {
			contentID = strings.Trim(contentID, "<>")
		}

		// Check if it's an attachment
		if strings.HasPrefix(disposition, "attachment") || filename != "" {
			attachment := models.Attachment{
				CID:         contentID,
				Type:        mediaType,
				Filename:    filename,
				ContentType: contentType,
				Size:        len(partContent),
			}
			*attachments = append(*attachments, attachment)
		} else {
			// It's a message part - decode content
			encoding := p.Header.Get("Content-Transfer-Encoding")
			decodedContent := decodeContent(partContent, encoding)
			part := models.Part{
				Type:        mediaType,
				ContentType: contentType,
				Body:        decodedContent,
				Size:        len(decodedContent),
			}

			*parts = append(*parts, part)
		}
	}

	return nil
}

// decodeContent decodes content based on transfer encoding
func decodeContent(content []byte, encoding string) string {
	switch strings.ToLower(encoding) {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(string(content))
		if err != nil {
			// Return original if decoding fails
			return string(content)
		}
		return string(decoded)
	case "quoted-printable":
		reader := quotedprintable.NewReader(bytes.NewReader(content))
		decoded, err := io.ReadAll(reader)
		if err != nil {
			// Return original if decoding fails
			return string(content)
		}
		return string(decoded)
	default:
		return string(content)
	}
}