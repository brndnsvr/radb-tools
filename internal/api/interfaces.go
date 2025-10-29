// Package api provides the HTTP client for interacting with the RADb API.
package api

import (
	"context"

	"github.com/bss/radb-client/internal/models"
)

// Client defines the interface for RADb API operations.
type Client interface {
	// Authentication
	Login(ctx context.Context, username, password string) error
	Logout(ctx context.Context) error
	IsAuthenticated() bool

	// Route operations
	ListRoutes(ctx context.Context, filters map[string]string) (*models.RouteList, error)
	GetRoute(ctx context.Context, prefix, asn string) (*models.RouteObject, error)
	CreateRoute(ctx context.Context, route *models.RouteObject) error
	UpdateRoute(ctx context.Context, route *models.RouteObject) error
	DeleteRoute(ctx context.Context, prefix, asn string) error

	// Contact operations
	ListContacts(ctx context.Context) (*models.ContactList, error)
	GetContact(ctx context.Context, id string) (*models.Contact, error)
	CreateContact(ctx context.Context, contact *models.Contact) error
	UpdateContact(ctx context.Context, contact *models.Contact) error
	DeleteContact(ctx context.Context, id string) error

	// Search operations
	Search(ctx context.Context, query string, objectType string) (interface{}, error)
	ValidateASN(ctx context.Context, asn string) (bool, error)

	// Configuration
	SetBaseURL(url string)
	SetSource(source string)
	SetTimeout(seconds int)
}
