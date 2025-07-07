package sendria

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enthus-golang/sendria/models"
)

type clientConfig struct {
	baseURL  string
	timeout  time.Duration
	username string
	password string
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name     string
		baseURL  string
		options  []Option
		expected clientConfig
	}{
		{
			name:    "default values",
			baseURL: "",
			options: nil,
			expected: clientConfig{
				baseURL: "http://localhost:1080",
				timeout: 30 * time.Second,
			},
		},
		{
			name:    "custom URL only",
			baseURL: "http://sendria.example.com:8025",
			options: nil,
			expected: clientConfig{
				baseURL: "http://sendria.example.com:8025",
				timeout: 30 * time.Second,
			},
		},
		{
			name:    "with basic auth",
			baseURL: "http://localhost:1080",
			options: []Option{
				WithBasicAuth("user", "pass"),
			},
			expected: clientConfig{
				baseURL:  "http://localhost:1080",
				timeout:  30 * time.Second,
				username: "user",
				password: "pass",
			},
		},
		{
			name:    "with custom timeout",
			baseURL: "http://localhost:1080",
			options: []Option{
				WithTimeout(60 * time.Second),
			},
			expected: clientConfig{
				baseURL: "http://localhost:1080",
				timeout: 60 * time.Second,
			},
		},
		{
			name:    "with all options",
			baseURL: "http://sendria.example.com:8025",
			options: []Option{
				WithBasicAuth("user", "pass"),
				WithTimeout(60 * time.Second),
			},
			expected: clientConfig{
				baseURL:  "http://sendria.example.com:8025",
				timeout:  60 * time.Second,
				username: "user",
				password: "pass",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			client := NewClient(tt.baseURL, tt.options...)
			if client.baseURL != tt.expected.baseURL {
				t.Errorf("expected baseURL %s, got %s", tt.expected.baseURL, client.baseURL)
			}
			if client.httpClient.Timeout != tt.expected.timeout {
				t.Errorf("expected timeout %v, got %v", tt.expected.timeout, client.httpClient.Timeout)
			}
			if client.username != tt.expected.username {
				t.Errorf("expected username %s, got %s", tt.expected.username, client.username)
			}
			if client.password != tt.expected.password {
				t.Errorf("expected password %s, got %s", tt.expected.password, client.password)
			}
		})
	}
}


func TestListMessages(t *testing.T) {
	t.Parallel()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/" {
			t.Errorf("expected path /api/messages/, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		// Convert to API format
		apiMessages := []models.APIMessage{
			{
				ID:                  1,
				Subject:             "Test Email 1",
				SenderMessage:       "jane@example.com",
				RecipientsMessageTo: []string{"john@example.com"},
				CreatedAt:           time.Now().Format("2006-01-02T15:04:05"),
				Size:                1024,
				Type:                "text/plain",
			},
		}
		data, _ := json.Marshal(apiMessages)
		resp := models.APIResponse{
			Code: "OK",
			Data: json.RawMessage(data),
			Meta: &models.APIMeta{PagesTotal: 1},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
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
	t.Parallel()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123.json" {
			t.Errorf("expected path /api/messages/123.json, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		// Convert to API format
		apiMessage := models.APIMessage{
			ID:                  123,
			Subject:             "Test Email",
			SenderMessage:       "jane@example.com",
			RecipientsMessageTo: []string{"john@example.com"},
			CreatedAt:           time.Now().Format("2006-01-02T15:04:05"),
			Size:                2048,
			Type:                "multipart/alternative",
		}
		data, _ := json.Marshal(apiMessage)
		resp := models.APIResponse{
			Code: "OK",
			Data: json.RawMessage(data),
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	message, err := client.GetMessage("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if message.ID != "123" {
		t.Errorf("expected message ID %s, got %s", "123", message.ID)
	}
	if message.Subject != "Test Email" {
		t.Errorf("expected subject %s, got %s", "Test Email", message.Subject)
	}
}

func TestGetMessageContent(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name         string
		method       func(*Client, string) (string, error)
		path         string
		contentType  string
		expectedBody string
	}{
		{
			name: "plain text",
			method: (*Client).GetMessagePlain,
			path: "/api/messages/123.plain",
			contentType: "text/plain",
			expectedBody: "Hello, this is a plain text email!",
		},
		{
			name: "HTML",
			method: (*Client).GetMessageHTML,
			path: "/api/messages/123.html",
			contentType: "text/html",
			expectedBody: "<html><body><p>Hello, this is an HTML email!</p></body></html>",
		},
		{
			name: "source",
			method: (*Client).GetMessageSource,
			path: "/api/messages/123.source",
			contentType: "text/plain",
			expectedBody: `From: sender@example.com
To: recipient@example.com
Subject: Test Email

This is the email body.`,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", tt.contentType)
				if _, err := w.Write([]byte(tt.expectedBody)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			body, err := tt.method(client, "123")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestGetMessageEML(t *testing.T) {
	t.Parallel()
	
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
		if _, err := w.Write(expectedEML); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	eml, err := client.GetMessageEML("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(eml) != string(expectedEML) {
		t.Errorf("expected EML %q, got %q", string(expectedEML), string(eml))
	}
}

func TestGetAttachment(t *testing.T) {
	t.Parallel()
	
	expectedAttachment := []byte("This is the attachment content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/messages/123/parts/cid123" {
			t.Errorf("expected path /api/messages/123/parts/cid123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(expectedAttachment); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	attachment, err := client.GetAttachment("123", "cid123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(attachment) != string(expectedAttachment) {
		t.Errorf("expected attachment %q, got %q", string(expectedAttachment), string(attachment))
	}
}

func TestDeleteMessage(t *testing.T) {
	t.Parallel()
	
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

	client := NewClient(server.URL)
	err := client.DeleteMessage("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAllMessages(t *testing.T) {
	t.Parallel()
	
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

	client := NewClient(server.URL)
	err := client.DeleteAllMessages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBasicAuth(t *testing.T) {
	t.Parallel()
	
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
		resp := models.APIResponse{
			Code: "OK",
			Data: json.RawMessage("[]"),
			Meta: &models.APIMeta{PagesTotal: 0},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, WithBasicAuth("testuser", "testpass"))
	_, err := client.ListMessages(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}