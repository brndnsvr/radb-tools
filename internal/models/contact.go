package models

import (
	"fmt"
	"time"
)

// ContactRole defines the type of contact
type ContactRole string

const (
	// ContactRoleAdmin represents an administrative contact
	ContactRoleAdmin ContactRole = "admin"

	// ContactRoleTech represents a technical contact
	ContactRoleTech ContactRole = "tech"

	// ContactRoleBilling represents a billing contact
	ContactRoleBilling ContactRole = "billing"

	// ContactRoleAbuse represents an abuse contact
	ContactRoleAbuse ContactRole = "abuse"
)

// Contact represents an account contact in RADb.
type Contact struct {
	// ID is the unique identifier for this contact
	ID string `json:"id"`

	// Name is the full name of the contact
	Name string `json:"name"`

	// Email is the contact's email address
	Email string `json:"email"`

	// Phone is the contact's phone number (optional)
	Phone string `json:"phone,omitempty"`

	// Role is the contact's role (admin, tech, billing, abuse)
	Role ContactRole `json:"role"`

	// Organization is the contact's organization (optional)
	Organization string `json:"organization,omitempty"`

	// Address contains the contact's physical address (optional)
	Address []string `json:"address,omitempty"`

	// Created is when the contact was created
	Created *time.Time `json:"created,omitempty"`

	// LastModified is when the contact was last updated
	LastModified *time.Time `json:"last_modified,omitempty"`

	// RawAttributes stores any additional attributes
	RawAttributes map[string][]string `json:"raw_attributes,omitempty"`
}

// Validate performs basic validation on the contact.
func (c *Contact) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("contact name is required")
	}

	if c.Email == "" {
		return fmt.Errorf("contact email is required")
	}

	if c.Role == "" {
		return fmt.Errorf("contact role is required")
	}

	// Validate role
	switch c.Role {
	case ContactRoleAdmin, ContactRoleTech, ContactRoleBilling, ContactRoleAbuse:
		// Valid role
	default:
		return fmt.Errorf("invalid contact role: %s", c.Role)
	}

	return nil
}

// ContactList is a collection of contacts.
type ContactList struct {
	Contacts  []Contact `json:"contacts"`
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// NewContactList creates a new contact list with the current timestamp.
func NewContactList(contacts []Contact) *ContactList {
	return &ContactList{
		Contacts:  contacts,
		Timestamp: time.Now().UTC(),
		Count:     len(contacts),
	}
}

// ByID returns a map of contacts indexed by their ID for quick lookup.
func (cl *ContactList) ByID() map[string]*Contact {
	m := make(map[string]*Contact, len(cl.Contacts))
	for i := range cl.Contacts {
		contact := &cl.Contacts[i]
		m[contact.ID] = contact
	}
	return m
}
