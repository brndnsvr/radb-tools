// Package models defines the data structures used throughout the RADb client.
package models

import (
	"fmt"
	"strings"
	"time"
)

// RouteObject represents a route or route6 object in RADb.
// These objects map IP prefixes to origin ASNs.
type RouteObject struct {
	// Route is the IP prefix (IPv4 or IPv6 in CIDR notation)
	Route string `json:"route"`

	// Origin is the AS number that originates this prefix
	Origin string `json:"origin"`

	// Descr is a human-readable description of the route
	Descr []string `json:"descr,omitempty"`

	// MntBy lists the maintainer objects that control this route
	MntBy []string `json:"mnt_by"`

	// Source identifies the IRR database (typically "RADB")
	Source string `json:"source"`

	// Created is when the route was first registered (if available)
	Created *time.Time `json:"created,omitempty"`

	// LastModified is when the route was last updated (if available)
	LastModified *time.Time `json:"last_modified,omitempty"`

	// Remarks contains any additional comments
	Remarks []string `json:"remarks,omitempty"`

	// MemberOf lists route-set memberships
	MemberOf []string `json:"member_of,omitempty"`

	// Holes lists more-specific prefixes that should be excluded
	Holes []string `json:"holes,omitempty"`

	// RawAttributes stores any additional RPSL attributes
	RawAttributes map[string][]string `json:"raw_attributes,omitempty"`
}

// ID returns a unique identifier for this route object.
func (r *RouteObject) ID() string {
	return fmt.Sprintf("%s-%s", r.Route, r.Origin)
}

// Validate performs basic validation on the route object.
func (r *RouteObject) Validate() error {
	if r.Route == "" {
		return fmt.Errorf("route prefix is required")
	}

	if r.Origin == "" {
		return fmt.Errorf("origin ASN is required")
	}

	if !strings.HasPrefix(r.Origin, "AS") {
		return fmt.Errorf("origin must start with 'AS'")
	}

	if len(r.MntBy) == 0 {
		return fmt.Errorf("at least one mnt-by is required")
	}

	if r.Source == "" {
		return fmt.Errorf("source is required")
	}

	return nil
}

// ToRPSL converts the route object to RPSL format for submission to RADb.
func (r *RouteObject) ToRPSL() string {
	var b strings.Builder

	// Determine object class
	objectClass := "route"
	if strings.Contains(r.Route, ":") {
		objectClass = "route6"
	}

	b.WriteString(fmt.Sprintf("%s: %s\n", objectClass, r.Route))
	b.WriteString(fmt.Sprintf("origin: %s\n", r.Origin))

	// Description
	for _, desc := range r.Descr {
		b.WriteString(fmt.Sprintf("descr: %s\n", desc))
	}

	// Maintainers
	for _, mnt := range r.MntBy {
		b.WriteString(fmt.Sprintf("mnt-by: %s\n", mnt))
	}

	// Remarks
	for _, remark := range r.Remarks {
		b.WriteString(fmt.Sprintf("remarks: %s\n", remark))
	}

	// Member-of
	for _, memberOf := range r.MemberOf {
		b.WriteString(fmt.Sprintf("member-of: %s\n", memberOf))
	}

	// Holes
	for _, hole := range r.Holes {
		b.WriteString(fmt.Sprintf("holes: %s\n", hole))
	}

	// Source (always last before changed/created)
	b.WriteString(fmt.Sprintf("source: %s\n", r.Source))

	return b.String()
}

// RouteList is a collection of route objects.
type RouteList struct {
	Routes    []RouteObject `json:"routes"`
	Timestamp time.Time     `json:"timestamp"`
	Count     int           `json:"count"`
}

// NewRouteList creates a new route list with the current timestamp.
func NewRouteList(routes []RouteObject) *RouteList {
	return &RouteList{
		Routes:    routes,
		Timestamp: time.Now().UTC(),
		Count:     len(routes),
	}
}

// ByID returns a map of routes indexed by their ID for quick lookup.
func (rl *RouteList) ByID() map[string]*RouteObject {
	m := make(map[string]*RouteObject, len(rl.Routes))
	for i := range rl.Routes {
		route := &rl.Routes[i]
		m[route.ID()] = route
	}
	return m
}
