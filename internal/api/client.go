package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPClient implements the Client interface using HTTP Basic Auth.
type HTTPClient struct {
	baseURL    string
	source     string
	timeout    time.Duration
	httpClient *http.Client
	logger     *logrus.Logger

	// Authentication state
	username string
	password string
	authenticated bool

	// Rate limiting
	rateLimiter *time.Ticker
}

// NewHTTPClient creates a new HTTP API client.
func NewHTTPClient(baseURL, source string, timeout int, logger *logrus.Logger) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		source:  source,
		timeout: time.Duration(timeout) * time.Second,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		logger:      logger,
		rateLimiter: time.NewTicker(time.Second), // Simple rate limiting
	}
}

// Login authenticates with the RADb API.
func (c *HTTPClient) Login(ctx context.Context, username, password string) error {
	// Store credentials
	c.username = username
	c.password = password
	c.authenticated = true

	c.logger.Debugf("Credentials stored for %s (length: %d)", username, len(password))
	c.logger.Debug("Credentials will be validated on first API request")

	// Note: We don't test auth here because most RADb API endpoints either:
	// 1. Don't require auth (like ASN validation)
	// 2. Require specific parameters (like search needs query-string)
	// 3. Need objects to exist (like route GET)
	//
	// The credentials will be validated when the user makes their first actual API call.
	// If invalid, they'll get a clear 401 Unauthorized error at that time.

	return nil
}

// Logout clears authentication state.
func (c *HTTPClient) Logout(ctx context.Context) error {
	c.username = ""
	c.password = ""
	c.authenticated = false
	c.logger.Info("Logged out")
	return nil
}

// IsAuthenticated returns whether the client is authenticated.
func (c *HTTPClient) IsAuthenticated() bool {
	return c.authenticated
}

// doRequest performs an HTTP request with retries and error handling.
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Rate limiting
	select {
	case <-c.rateLimiter.C:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if c.authenticated {
		req.SetBasicAuth(c.username, c.password)
		c.logger.Debugf("Set BasicAuth for request (user: %s)", c.username)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request with retries
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < maxRetries-1 {
			c.logger.Warnf("Request failed (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
	}

	return resp, nil
}

// Actual implementations are in routes.go, contacts.go, and search.go

// SetBaseURL updates the base URL.
func (c *HTTPClient) SetBaseURL(url string) {
	c.baseURL = url
}

// SetSource updates the source.
func (c *HTTPClient) SetSource(source string) {
	c.source = source
}

// SetTimeout updates the timeout.
func (c *HTTPClient) SetTimeout(seconds int) {
	c.timeout = time.Duration(seconds) * time.Second
	c.httpClient.Timeout = c.timeout
}
