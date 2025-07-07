package sendria

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enthus-golang/go-sendria/internal/models"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected struct {
			baseURL string
			timeout time.Duration
		}
	}{
		{
			name:   "default values",
			config: Config{},
			expected: struct {
				baseURL string
				timeout time.Duration
			}{
				baseURL: "http://localhost:1025",
				timeout: 30 * time.Second,
			},
		},
		{
			name: "custom values",
			config: Config{
				BaseURL:  "http://sendria.example.com:8025",
				Username: "user",
				Password: "pass",
				Timeout:  60 * time.Second,
			},
			expected: struct {
				baseURL string
				timeout time.Duration
			}{
				baseURL: "http://sendria.example.com:8025",
				timeout: 60 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client.baseURL != tt.expected.baseURL {
				t.Errorf("expected baseURL %s, got %s", tt.expected.baseURL, client.baseURL)
			}
			if client.httpClient.Timeout != tt.expected.timeout {
				t.Errorf("expected timeout %v, got %v", tt.expected.timeout, client.httpClient.Timeout)
			}
			if tt.config.Username != "" && client.username != tt.config.Username {
				t.Errorf("expected username %s, got %s", tt.config.Username, client.username)
			}
			if tt.config.Password != "" && client.password != tt.config.Password {
				t.Errorf("expected password %s, got %s", tt.config.Password, client.password)
			}
		})
	}
}

func TestListMessages(t *testing.T) {
	expectedMessages := models.MessageList{
		Messages: []models.Message{
			{
				ID:      "1",
				Subject: "Test Email 1",
				To: []models.Recipient{
					{Name: "John Doe", Email: "john@example.com"},
				},
				From: []models.Recipient{
					{Name: "Jane Doe", Email: "jane@example.com"},
				},
				CreatedAt: time.Now(),
				Size:      1024,
				Type:      "text/plain",
			},
		},
		Total:   1,
		Page:    1,
		PerPage: 10,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/" {
			t.Errorf("expected path /api/messages/, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedMessages)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	messages, err := client.ListMessages(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages.Messages))
	}
	if messages.Messages[0].ID != "1" {
		t.Errorf("expected message ID 1, got %s", messages.Messages[0].ID)
	}
}

func TestGetMessage(t *testing.T) {
	expectedMessage := models.Message{
		ID:      "123",
		Subject: "Test Email",
		To: []models.Recipient{
			{Name: "John Doe", Email: "john@example.com"},
		},
		From: []models.Recipient{
			{Name: "Jane Doe", Email: "jane@example.com"},
		},
		CreatedAt: time.Now(),
		Size:      2048,
		Type:      "multipart/alternative",
		Parts: []models.Part{
			{
				Type:        "text/plain",
				ContentType: "text/plain; charset=utf-8",
				Body:        "Hello, World!",
				Size:        13,
			},
			{
				Type:        "text/html",
				ContentType: "text/html; charset=utf-8",
				Body:        "<p>Hello, World!</p>",
				Size:        20,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123.json" {
			t.Errorf("expected path /api/messages/123.json, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedMessage)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	message, err := client.GetMessage("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if message.ID != expectedMessage.ID {
		t.Errorf("expected message ID %s, got %s", expectedMessage.ID, message.ID)
	}
	if message.Subject != expectedMessage.Subject {
		t.Errorf("expected subject %s, got %s", expectedMessage.Subject, message.Subject)
	}
	if len(message.Parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(message.Parts))
	}
}

func TestGetMessagePlain(t *testing.T) {
	expectedBody := "Hello, this is a plain text email!"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123.plain" {
			t.Errorf("expected path /api/messages/123.plain, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	body, err := client.GetMessagePlain("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestGetMessageHTML(t *testing.T) {
	expectedBody := "<html><body><p>Hello, this is an HTML email!</p></body></html>"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123.html" {
			t.Errorf("expected path /api/messages/123.html, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	body, err := client.GetMessageHTML("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestGetMessageSource(t *testing.T) {
	expectedSource := `From: sender@example.com
To: recipient@example.com
Subject: Test Email

This is the email body.`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123.source" {
			t.Errorf("expected path /api/messages/123.source, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(expectedSource))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	source, err := client.GetMessageSource("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if source != expectedSource {
		t.Errorf("expected source %q, got %q", expectedSource, source)
	}
}

func TestGetMessageEML(t *testing.T) {
	expectedEML := []byte(`From: sender@example.com
To: recipient@example.com
Subject: Test Email

This is the email body.`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123.eml" {
			t.Errorf("expected path /api/messages/123.eml, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "message/rfc822")
		w.Header().Set("Content-Disposition", "attachment; filename=\"message.eml\"")
		w.Write(expectedEML)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	eml, err := client.GetMessageEML("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(eml) != string(expectedEML) {
		t.Errorf("expected EML %q, got %q", string(expectedEML), string(eml))
	}
}

func TestGetAttachment(t *testing.T) {
	expectedAttachment := []byte("This is the attachment content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123/parts/cid123" {
			t.Errorf("expected path /api/messages/123/parts/cid123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(expectedAttachment)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	attachment, err := client.GetAttachment("123", "cid123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(attachment) != string(expectedAttachment) {
		t.Errorf("expected attachment %q, got %q", string(expectedAttachment), string(attachment))
	}
}

func TestDeleteMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123" {
			t.Errorf("expected path /api/messages/123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected method DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	err := client.DeleteMessage("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAllMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/" {
			t.Errorf("expected path /api/messages/, got %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected method DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	err := client.DeleteAllMessages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth, but not present")
		}
		if username != "testuser" {
			t.Errorf("expected username testuser, got %s", username)
		}
		if password != "testpass" {
			t.Errorf("expected password testpass, got %s", password)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.MessageList{})
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	})
	_, err := client.ListMessages(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}