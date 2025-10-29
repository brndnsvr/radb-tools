package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bss/radb-client/internal/models"
	"github.com/bss/radb-client/pkg/validator"
)

// ListContacts retrieves all contacts with optional role-based filtering.
func (c *HTTPClient) ListContacts(ctx context.Context) (*models.ContactList, error) {
	c.logger.Debug("ListContacts called")

	if !c.authenticated {
		return nil, fmt.Errorf("not authenticated: please login first")
	}

	path := fmt.Sprintf("/%s/contact", c.source)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list contacts failed with status %d: %s", resp.StatusCode, string(body))
	}

	var contacts []models.Contact
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts response: %w", err)
	}

	c.logger.Infof("Retrieved %d contacts", len(contacts))
	return models.NewContactList(contacts), nil
}

// GetContact retrieves a specific contact by ID.
func (c *HTTPClient) GetContact(ctx context.Context, id string) (*models.Contact, error) {
	c.logger.Debugf("GetContact called for %s", id)

	if !c.authenticated {
		return nil, fmt.Errorf("not authenticated: please login first")
	}

	if id == "" {
		return nil, fmt.Errorf("contact ID is required")
	}

	path := fmt.Sprintf("/%s/contact/%s", c.source, url.PathEscape(id))
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("contact not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get contact failed with status %d: %s", resp.StatusCode, string(body))
	}

	var contact models.Contact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return nil, fmt.Errorf("failed to decode contact response: %w", err)
	}

	c.logger.Infof("Retrieved contact %s", contact.ID)
	return &contact, nil
}

// CreateContact creates a new contact.
func (c *HTTPClient) CreateContact(ctx context.Context, contact *models.Contact) error {
	c.logger.Debugf("CreateContact called for %s", contact.ID)

	if !c.authenticated {
		return fmt.Errorf("not authenticated: please login first")
	}

	// Validate the contact
	if err := contact.Validate(); err != nil {
		return fmt.Errorf("contact validation failed: %w", err)
	}

	// Validate email format
	if err := validator.ValidateEmail(contact.Email); err != nil {
		return fmt.Errorf("invalid email %s: %w", contact.Email, err)
	}

	path := fmt.Sprintf("/%s/contact", c.source)
	resp, err := c.doRequest(ctx, "POST", path, contact)
	if err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create contact failed with status %d: %s", resp.StatusCode, string(body))
	}

	// If the response includes the created contact with ID, update it
	if resp.StatusCode == http.StatusCreated {
		var created models.Contact
		if err := json.NewDecoder(resp.Body).Decode(&created); err == nil && created.ID != "" {
			contact.ID = created.ID
		}
	}

	c.logger.Infof("Successfully created contact %s", contact.ID)
	return nil
}

// UpdateContact updates an existing contact.
func (c *HTTPClient) UpdateContact(ctx context.Context, contact *models.Contact) error {
	c.logger.Debugf("UpdateContact called for %s", contact.ID)

	if !c.authenticated {
		return fmt.Errorf("not authenticated: please login first")
	}

	if contact.ID == "" {
		return fmt.Errorf("contact ID is required for update")
	}

	// Validate the contact
	if err := contact.Validate(); err != nil {
		return fmt.Errorf("contact validation failed: %w", err)
	}

	// Validate email format
	if err := validator.ValidateEmail(contact.Email); err != nil {
		return fmt.Errorf("invalid email %s: %w", contact.Email, err)
	}

	path := fmt.Sprintf("/%s/contact/%s", c.source, url.PathEscape(contact.ID))
	resp, err := c.doRequest(ctx, "PUT", path, contact)
	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("contact not found: %s", contact.ID)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update contact failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Infof("Successfully updated contact %s", contact.ID)
	return nil
}

// DeleteContact deletes a contact.
func (c *HTTPClient) DeleteContact(ctx context.Context, id string) error {
	c.logger.Debugf("DeleteContact called for %s", id)

	if !c.authenticated {
		return fmt.Errorf("not authenticated: please login first")
	}

	if id == "" {
		return fmt.Errorf("contact ID is required")
	}

	path := fmt.Sprintf("/%s/contact/%s", c.source, url.PathEscape(id))
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("contact not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete contact failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Infof("Successfully deleted contact %s", id)
	return nil
}
