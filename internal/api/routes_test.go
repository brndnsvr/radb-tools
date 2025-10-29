package api

import (
	"testing"

	"github.com/bss/radb-client/internal/models"
	"github.com/sirupsen/logrus"
)

func TestRouteValidation(t *testing.T) {
	route := &models.RouteObject{
		Route:  "192.0.2.0/24",
		Origin: "AS64496",
		MntBy:  []string{"MAINT-TEST"},
		Source: "RADB",
	}

	if err := route.Validate(); err != nil {
		t.Errorf("Valid route failed validation: %v", err)
	}

	// Test missing route
	invalidRoute := &models.RouteObject{
		Origin: "AS64496",
		MntBy:  []string{"MAINT-TEST"},
		Source: "RADB",
	}

	if err := invalidRoute.Validate(); err == nil {
		t.Errorf("Expected validation error for missing route")
	}

	// Test missing origin
	invalidRoute2 := &models.RouteObject{
		Route:  "192.0.2.0/24",
		MntBy:  []string{"MAINT-TEST"},
		Source: "RADB",
	}

	if err := invalidRoute2.Validate(); err == nil {
		t.Errorf("Expected validation error for missing origin")
	}

	// Test missing maintainer
	invalidRoute3 := &models.RouteObject{
		Route:  "192.0.2.0/24",
		Origin: "AS64496",
		Source: "RADB",
	}

	if err := invalidRoute3.Validate(); err == nil {
		t.Errorf("Expected validation error for missing maintainer")
	}
}

func TestHTTPClientCreation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewHTTPClient("https://api.example.com", "RADB", 30, logger)
	if client == nil {
		t.Fatal("Failed to create HTTP client")
	}

	if !client.IsAuthenticated() {
		// Client should not be authenticated initially
	} else {
		t.Error("New client should not be authenticated")
	}
}

func TestRouteID(t *testing.T) {
	route := &models.RouteObject{
		Route:  "192.0.2.0/24",
		Origin: "AS64496",
	}

	id := route.ID()
	expected := "192.0.2.0/24-AS64496"

	if id != expected {
		t.Errorf("Expected ID %s, got %s", expected, id)
	}
}
