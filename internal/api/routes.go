package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bss/radb-client/internal/models"
	"github.com/bss/radb-client/pkg/validator"
)

// ListRoutes retrieves all routes matching the given filters with pagination support.
// Filters can include: prefix, origin (ASN), mnt-by, etc.
func (c *HTTPClient) ListRoutes(ctx context.Context, filters map[string]string) (*models.RouteList, error) {
	c.logger.Debug("ListRoutes called")

	if !c.authenticated {
		return nil, fmt.Errorf("not authenticated: please login first")
	}

	// Build query parameters
	params := url.Values{}
	for key, value := range filters {
		params.Add(key, value)
	}

	path := fmt.Sprintf("/%s/route", c.source)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list routes failed with status %d: %s", resp.StatusCode, string(body))
	}

	var routes []models.RouteObject
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, fmt.Errorf("failed to decode routes response: %w", err)
	}

	c.logger.Infof("Retrieved %d routes", len(routes))
	return models.NewRouteList(routes), nil
}

// GetRoute retrieves a specific route object by prefix and origin ASN.
func (c *HTTPClient) GetRoute(ctx context.Context, prefix, asn string) (*models.RouteObject, error) {
	c.logger.Debugf("GetRoute called for %s AS%s", prefix, asn)

	if !c.authenticated {
		return nil, fmt.Errorf("not authenticated: please login first")
	}

	// Validate inputs
	if err := validator.ValidatePrefix(prefix); err != nil {
		return nil, fmt.Errorf("invalid prefix: %w", err)
	}
	if err := validator.ValidateASN(asn); err != nil {
		return nil, fmt.Errorf("invalid ASN: %w", err)
	}

	// Ensure ASN has AS prefix
	if !strings.HasPrefix(asn, "AS") {
		asn = "AS" + asn
	}

	// Build path - use prefix and origin as identifier
	path := fmt.Sprintf("/%s/route/%s/%s", c.source, url.PathEscape(prefix), asn)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("route not found: %s %s", prefix, asn)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get route failed with status %d: %s", resp.StatusCode, string(body))
	}

	var route models.RouteObject
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, fmt.Errorf("failed to decode route response: %w", err)
	}

	c.logger.Infof("Retrieved route %s", route.ID())
	return &route, nil
}

// CreateRoute creates a new route object in RADb.
func (c *HTTPClient) CreateRoute(ctx context.Context, route *models.RouteObject) error {
	c.logger.Debugf("CreateRoute called for %s", route.ID())

	if !c.authenticated {
		return fmt.Errorf("not authenticated: please login first")
	}

	// Validate the route object
	if err := route.Validate(); err != nil {
		return fmt.Errorf("route validation failed: %w", err)
	}

	// Additional validation using pkg/validator
	if err := validator.ValidatePrefix(route.Route); err != nil {
		return fmt.Errorf("invalid prefix %s: %w", route.Route, err)
	}
	if err := validator.ValidateASN(route.Origin); err != nil {
		return fmt.Errorf("invalid origin ASN %s: %w", route.Origin, err)
	}

	// Set source if not provided
	if route.Source == "" {
		route.Source = c.source
	}

	path := fmt.Sprintf("/%s/route", c.source)
	resp, err := c.doRequest(ctx, "POST", path, route)
	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create route failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Infof("Successfully created route %s", route.ID())
	return nil
}

// UpdateRoute updates an existing route object in RADb.
func (c *HTTPClient) UpdateRoute(ctx context.Context, route *models.RouteObject) error {
	c.logger.Debugf("UpdateRoute called for %s", route.ID())

	if !c.authenticated {
		return fmt.Errorf("not authenticated: please login first")
	}

	// Validate the route object
	if err := route.Validate(); err != nil {
		return fmt.Errorf("route validation failed: %w", err)
	}

	// Additional validation
	if err := validator.ValidatePrefix(route.Route); err != nil {
		return fmt.Errorf("invalid prefix %s: %w", route.Route, err)
	}
	if err := validator.ValidateASN(route.Origin); err != nil {
		return fmt.Errorf("invalid origin ASN %s: %w", route.Origin, err)
	}

	// Ensure ASN has AS prefix
	asn := route.Origin
	if !strings.HasPrefix(asn, "AS") {
		asn = "AS" + asn
	}

	path := fmt.Sprintf("/%s/route/%s/%s", c.source, url.PathEscape(route.Route), asn)
	resp, err := c.doRequest(ctx, "PUT", path, route)
	if err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("route not found: %s", route.ID())
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update route failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Infof("Successfully updated route %s", route.ID())
	return nil
}

// DeleteRoute deletes a route object from RADb.
func (c *HTTPClient) DeleteRoute(ctx context.Context, prefix, asn string) error {
	c.logger.Debugf("DeleteRoute called for %s AS%s", prefix, asn)

	if !c.authenticated {
		return fmt.Errorf("not authenticated: please login first")
	}

	// Validate inputs
	if err := validator.ValidatePrefix(prefix); err != nil {
		return fmt.Errorf("invalid prefix: %w", err)
	}
	if err := validator.ValidateASN(asn); err != nil {
		return fmt.Errorf("invalid ASN: %w", err)
	}

	// Ensure ASN has AS prefix
	if !strings.HasPrefix(asn, "AS") {
		asn = "AS" + asn
	}

	path := fmt.Sprintf("/%s/route/%s/%s", c.source, url.PathEscape(prefix), asn)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("route not found: %s %s", prefix, asn)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete route failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Infof("Successfully deleted route %s-%s", prefix, asn)
	return nil
}
