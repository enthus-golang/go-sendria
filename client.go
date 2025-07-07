package sendria

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/enthus-golang/go-sendria/internal/models"
)

// Client represents a Sendria API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// Config holds the configuration for the Sendria client
type Config struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

// NewClient creates a new Sendria API client
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:1025"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		username: config.Username,
		password: config.Password,
	}
}

// doRequest performs an HTTP request with optional basic auth
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}

	return resp, nil
}

// ListMessages retrieves a paginated list of messages
func (c *Client) ListMessages(page, perPage int) (*models.MessageList, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}

	path := "/api/messages/"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var messageList models.MessageList
	if err := json.NewDecoder(resp.Body).Decode(&messageList); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &messageList, nil
}

// GetMessage retrieves a specific message by ID
func (c *Client) GetMessage(id string) (*models.Message, error) {
	path := fmt.Sprintf("/api/messages/%s.json", id)

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var message models.Message
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &message, nil
}

// GetMessagePlain retrieves the plain text part of a message
func (c *Client) GetMessagePlain(id string) (string, error) {
	path := fmt.Sprintf("/api/messages/%s.plain", id)

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(body), nil
}

// GetMessageHTML retrieves the HTML part of a message
func (c *Client) GetMessageHTML(id string) (string, error) {
	path := fmt.Sprintf("/api/messages/%s.html", id)

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(body), nil
}

// GetMessageSource retrieves the raw source of a message
func (c *Client) GetMessageSource(id string) (string, error) {
	path := fmt.Sprintf("/api/messages/%s.source", id)

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(body), nil
}

// GetMessageEML retrieves the message as an EML file
func (c *Client) GetMessageEML(id string) ([]byte, error) {
	path := fmt.Sprintf("/api/messages/%s.eml", id)

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetAttachment downloads a message attachment by CID
func (c *Client) GetAttachment(messageID, cid string) ([]byte, error) {
	path := fmt.Sprintf("/api/messages/%s/parts/%s", messageID, cid)

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DeleteMessage deletes a specific message
func (c *Client) DeleteMessage(id string) error {
	path := fmt.Sprintf("/api/messages/%s", id)

	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DeleteAllMessages deletes all messages
func (c *Client) DeleteAllMessages() error {
	path := "/api/messages/"

	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}