package sendria

import (
	"strings"
	"testing"
)

func TestParseMIMEMessage(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		expectedParts    int
		expectedAttachments int
		checkContent     bool
		expectedContent  string
	}{
		{
			name: "simple plain text email",
			source: `From: sender@example.com
To: recipient@example.com
Subject: Test Email
Content-Type: text/plain; charset=utf-8

This is a simple test email.`,
			expectedParts:    1,
			expectedAttachments: 0,
			checkContent:     true,
			expectedContent:  "This is a simple test email.",
		},
		{
			name: "multipart alternative email",
			source: `From: sender@example.com
To: recipient@example.com
Subject: Multipart Test
Content-Type: multipart/alternative; boundary="boundary1"

--boundary1
Content-Type: text/plain; charset=utf-8

Plain text version
--boundary1
Content-Type: text/html; charset=utf-8

<p>HTML version</p>
--boundary1--`,
			expectedParts:    2,
			expectedAttachments: 0,
		},
		{
			name: "email with attachment",
			source: `From: sender@example.com
To: recipient@example.com
Subject: Email with Attachment
Content-Type: multipart/mixed; boundary="boundary2"

--boundary2
Content-Type: text/plain; charset=utf-8

Email body text
--boundary2
Content-Type: application/pdf
Content-Disposition: attachment; filename="document.pdf"
Content-Transfer-Encoding: base64

JVBERi0xLjQKJeLjz9MKCg==
--boundary2--`,
			expectedParts:    1,
			expectedAttachments: 1,
		},
		{
			name: "quoted-printable encoding",
			source: `From: sender@example.com
To: recipient@example.com
Subject: Quoted Printable Test
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: quoted-printable

Hello=20World!=0D=0AThis=20is=20a=20test.`,
			expectedParts:    1,
			expectedAttachments: 0,
			checkContent:     true,
			expectedContent:  "Hello World!\r\nThis is a test.",
		},
		{
			name: "base64 encoded content",
			source: `From: sender@example.com
To: recipient@example.com
Subject: Base64 Test
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: base64

SGVsbG8gV29ybGQh`,
			expectedParts:    1,
			expectedAttachments: 0,
			checkContent:     true,
			expectedContent:  "Hello World!",
		},
		{
			name: "nested multipart",
			source: `From: sender@example.com
To: recipient@example.com
Subject: Nested Multipart
Content-Type: multipart/mixed; boundary="outer"

--outer
Content-Type: multipart/alternative; boundary="inner"

--inner
Content-Type: text/plain

Plain text
--inner
Content-Type: text/html

<p>HTML text</p>
--inner--
--outer
Content-Type: image/png
Content-Disposition: attachment; filename="image.png"
Content-ID: <image123>

[PNG data]
--outer--`,
			expectedParts:    2,
			expectedAttachments: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, attachments, err := parseMIMEMessage(tt.source)
			if err != nil {
				t.Fatalf("parseMIMEMessage() error = %v", err)
			}

			if len(parts) != tt.expectedParts {
				t.Errorf("Expected %d parts, got %d", tt.expectedParts, len(parts))
			}

			if len(attachments) != tt.expectedAttachments {
				t.Errorf("Expected %d attachments, got %d", tt.expectedAttachments, len(attachments))
			}

			if tt.checkContent && len(parts) > 0 {
				if strings.TrimSpace(parts[0].Body) != strings.TrimSpace(tt.expectedContent) {
					t.Errorf("Expected content %q, got %q", tt.expectedContent, parts[0].Body)
				}
			}
		})
	}
}

func TestParseMIMEMessage_ErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "invalid email format",
			source: "This is not a valid email",
		},
		{
			name:   "empty source",
			source: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseMIMEMessage(tt.source)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestDecodeContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		encoding string
		expected string
	}{
		{
			name:     "base64 decoding",
			content:  []byte("SGVsbG8gV29ybGQh"),
			encoding: "base64",
			expected: "Hello World!",
		},
		{
			name:     "quoted-printable decoding",
			content:  []byte("Hello=20World!"),
			encoding: "quoted-printable",
			expected: "Hello World!",
		},
		{
			name:     "no encoding",
			content:  []byte("Plain text"),
			encoding: "",
			expected: "Plain text",
		},
		{
			name:     "unknown encoding",
			content:  []byte("Some content"),
			encoding: "unknown",
			expected: "Some content",
		},
		{
			name:     "invalid base64",
			content:  []byte("Invalid!@#$"),
			encoding: "base64",
			expected: "Invalid!@#$", // Should return original on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeContent(tt.content, tt.encoding)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}