package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bss/radb-client/pkg/validator"
)

// SearchResult contains search results from the API.
type SearchResult struct {
	Results   []map[string]interface{} `json:"results"`
	Count     int                      `json:"count"`
	Query     string                   `json:"query"`
	Type      string                   `json:"type,omitempty"`
	NextToken string                   `json:"next_token,omitempty"`
}

// Search performs a general search query on the RADb.
// The objectType parameter can be "route", "contact", "as-set", "mntner", etc.
func (c *HTTPClient) Search(ctx context.Context, query string, objectType string) (interface{}, error) {
	c.logger.Debugf("Search called with query=%s type=%s", query, objectType)

	if !c.authenticated {
		return nil, fmt.Errorf("not authenticated: please login first")
	}

	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Build query parameters per RADb API requirements
	params := url.Values{}
	params.Add("query-string", query)
	if objectType != "" {
		params.Add("type", objectType)
	}

	// Use lowercase source name in path
	sourceLower := "radb"  // API requires lowercase
	path := fmt.Sprintf("/%s/search?%s", sourceLower, params.Encode())
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	c.logger.Debugf("[DEBUG] Response body (first 500 chars): %s", string(body[:min(500, len(body))]))

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Debugf("[DEBUG] JSON decode failed, body might be RPSL format")
		// Return raw text as a simple result
		return map[string]interface{}{
			"raw_response": string(body),
			"format":       "rpsl",
		}, nil
	}

	c.logger.Infof("Search returned %d results", result.Count)
	return &result, nil
}

// ValidateASN validates an ASN with the RADb API.
// It checks if the ASN exists and returns true if it's valid.
func (c *HTTPClient) ValidateASN(ctx context.Context, asn string) (bool, error) {
	c.logger.Debugf("ValidateASN called for %s", asn)

	if !c.authenticated {
		return false, fmt.Errorf("not authenticated: please login first")
	}

	// First, validate ASN format locally
	if err := validator.ValidateASN(asn); err != nil {
		return false, fmt.Errorf("invalid ASN format: %w", err)
	}

	// Query the API to check if the ASN exists
	params := url.Values{}
	params.Add("asn", asn)

	path := fmt.Sprintf("/%s/validate/asn?%s", c.source, params.Encode())
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return false, fmt.Errorf("ASN validation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("ASN validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var validationResult struct {
		Valid bool   `json:"valid"`
		ASN   string `json:"asn"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&validationResult); err != nil {
		return false, fmt.Errorf("failed to decode validation response: %w", err)
	}

	c.logger.Infof("ASN %s validation result: %v", asn, validationResult.Valid)
	return validationResult.Valid, nil
}

// SearchRoutesByPrefix searches for routes matching a specific prefix.
func (c *HTTPClient) SearchRoutesByPrefix(ctx context.Context, prefix string) (interface{}, error) {
	if err := validator.ValidatePrefix(prefix); err != nil {
		return nil, fmt.Errorf("invalid prefix: %w", err)
	}
	return c.Search(ctx, prefix, "route")
}

// SearchRoutesByASN searches for routes originated by a specific ASN.
func (c *HTTPClient) SearchRoutesByASN(ctx context.Context, asn string) (interface{}, error) {
	if err := validator.ValidateASN(asn); err != nil {
		return nil, fmt.Errorf("invalid ASN: %w", err)
	}
	return c.Search(ctx, asn, "route")
}
