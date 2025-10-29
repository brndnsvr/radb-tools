package api

import (
	"context"
	"fmt"

	"github.com/bss/radb-client/internal/models"
)

// RouteStream provides an iterator for streaming routes in batches.
type RouteStream struct {
	client    *HTTPClient
	ctx       context.Context
	batchSize int
	filters   map[string]string
	offset    int
	buffer    []models.RouteObject
	bufferPos int
	done      bool
	err       error
}

// StreamRoutes creates a new route stream for memory-efficient processing.
func (c *HTTPClient) StreamRoutes(ctx context.Context, filters map[string]string, batchSize int) *RouteStream {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &RouteStream{
		client:    c,
		ctx:       ctx,
		batchSize: batchSize,
		filters:   filters,
		buffer:    make([]models.RouteObject, 0, batchSize),
	}
}

// Next advances to the next route and returns true if a route is available.
// Returns false when there are no more routes or an error occurred.
func (s *RouteStream) Next() bool {
	if s.done {
		return false
	}

	// If we have routes in the buffer, return the next one
	if s.bufferPos < len(s.buffer) {
		s.bufferPos++
		return true
	}

	// Need to fetch the next batch
	s.bufferPos = 0
	s.buffer = s.buffer[:0]

	// Add pagination to filters
	filters := make(map[string]string)
	for k, v := range s.filters {
		filters[k] = v
	}
	filters["offset"] = fmt.Sprintf("%d", s.offset)
	filters["limit"] = fmt.Sprintf("%d", s.batchSize)

	// Fetch next batch
	routeList, err := s.client.ListRoutes(s.ctx, filters)
	if err != nil {
		s.err = err
		s.done = true
		return false
	}

	// Check if we got any routes
	if len(routeList.Routes) == 0 {
		s.done = true
		return false
	}

	// Update buffer and offset
	s.buffer = routeList.Routes
	s.offset += len(routeList.Routes)

	// If we got fewer routes than requested, we're done after this batch
	if len(routeList.Routes) < s.batchSize {
		s.done = true
	}

	s.bufferPos = 1 // Move to first item
	return true
}

// Route returns the current route. Only valid after Next() returns true.
func (s *RouteStream) Route() *models.RouteObject {
	if s.bufferPos == 0 || s.bufferPos > len(s.buffer) {
		return nil
	}
	return &s.buffer[s.bufferPos-1]
}

// Err returns any error that occurred during streaming.
func (s *RouteStream) Err() error {
	return s.err
}

// Close cleans up resources used by the stream.
func (s *RouteStream) Close() error {
	s.done = true
	s.buffer = nil
	return nil
}

// ContactStream provides an iterator for streaming contacts in batches.
type ContactStream struct {
	client    *HTTPClient
	ctx       context.Context
	batchSize int
	offset    int
	buffer    []models.Contact
	bufferPos int
	done      bool
	err       error
}

// StreamContacts creates a new contact stream for memory-efficient processing.
func (c *HTTPClient) StreamContacts(ctx context.Context, batchSize int) *ContactStream {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &ContactStream{
		client:    c,
		ctx:       ctx,
		batchSize: batchSize,
		buffer:    make([]models.Contact, 0, batchSize),
	}
}

// Next advances to the next contact and returns true if a contact is available.
func (s *ContactStream) Next() bool {
	if s.done {
		return false
	}

	if s.bufferPos < len(s.buffer) {
		s.bufferPos++
		return true
	}

	// Fetch next batch
	s.bufferPos = 0
	s.buffer = s.buffer[:0]

	contactList, err := s.client.ListContacts(s.ctx)
	if err != nil {
		s.err = err
		s.done = true
		return false
	}

	// For contacts, we might not have pagination, so we load all at once
	// In a real implementation, this would support pagination
	if len(contactList.Contacts) == 0 {
		s.done = true
		return false
	}

	s.buffer = contactList.Contacts
	s.done = true // All contacts loaded
	s.bufferPos = 1

	return true
}

// Contact returns the current contact. Only valid after Next() returns true.
func (s *ContactStream) Contact() *models.Contact {
	if s.bufferPos == 0 || s.bufferPos > len(s.buffer) {
		return nil
	}
	return &s.buffer[s.bufferPos-1]
}

// Err returns any error that occurred during streaming.
func (s *ContactStream) Err() error {
	return s.err
}

// Close cleans up resources used by the stream.
func (s *ContactStream) Close() error {
	s.done = true
	s.buffer = nil
	return nil
}
