package state

import (
	"context"
	"testing"
	"time"

	"github.com/bss/radb-client/internal/models"
)

func TestComputeDiff(t *testing.T) {
	ctx := context.Background()

	// Create test routes
	route1 := models.RouteObject{
		Route:  "192.0.2.0/24",
		Origin: "AS64496",
		MntBy:  []string{"MAINT-TEST"},
		Source: "RADB",
	}

	route2 := models.RouteObject{
		Route:  "198.51.100.0/24",
		Origin: "AS64497",
		MntBy:  []string{"MAINT-TEST"},
		Source: "RADB",
	}

	route1Modified := route1
	route1Modified.Descr = []string{"Modified description"}

	// Create snapshots
	snap1 := &models.Snapshot{
		ID:        "test-1",
		Timestamp: time.Now(),
		Type:      models.SnapshotTypeRoute,
		Routes: &models.RouteList{
			Routes:    []models.RouteObject{route1},
			Timestamp: time.Now(),
			Count:     1,
		},
	}

	snap2 := &models.Snapshot{
		ID:        "test-2",
		Timestamp: time.Now(),
		Type:      models.SnapshotTypeRoute,
		Routes: &models.RouteList{
			Routes:    []models.RouteObject{route1Modified, route2},
			Timestamp: time.Now(),
			Count:     2,
		},
	}

	// Compute diff
	diff, err := ComputeDiff(ctx, snap1, snap2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// Verify results
	if len(diff.Added) != 1 {
		t.Errorf("Expected 1 added route, got %d", len(diff.Added))
	}

	if len(diff.Modified) != 1 {
		t.Errorf("Expected 1 modified route, got %d", len(diff.Modified))
	}

	if len(diff.Removed) != 0 {
		t.Errorf("Expected 0 removed routes, got %d", len(diff.Removed))
	}

	// Check summary
	if diff.Summary.AddedCount != 1 || diff.Summary.ModifiedCount != 1 {
		t.Errorf("Summary counts incorrect: added=%d, modified=%d",
			diff.Summary.AddedCount, diff.Summary.ModifiedCount)
	}
}

func TestComputeDiffEmpty(t *testing.T) {
	ctx := context.Background()

	snap1 := &models.Snapshot{
		ID:        "test-1",
		Timestamp: time.Now(),
		Type:      models.SnapshotTypeRoute,
		Routes: &models.RouteList{
			Routes:    []models.RouteObject{},
			Timestamp: time.Now(),
			Count:     0,
		},
	}

	snap2 := &models.Snapshot{
		ID:        "test-2",
		Timestamp: time.Now(),
		Type:      models.SnapshotTypeRoute,
		Routes: &models.RouteList{
			Routes:    []models.RouteObject{},
			Timestamp: time.Now(),
			Count:     0,
		},
	}

	diff, err := ComputeDiff(ctx, snap1, snap2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	if !diff.IsEmpty() {
		t.Errorf("Expected empty diff")
	}
}

func TestDiffToChangeSet(t *testing.T) {
	route := &models.RouteObject{
		Route:  "192.0.2.0/24",
		Origin: "AS64496",
		MntBy:  []string{"MAINT-TEST"},
		Source: "RADB",
	}

	diff := models.NewDiffResult()
	diff.Added = append(diff.Added, route)
	diff.ComputeSummary()

	cs := DiffToChangeSet(diff, "snap1", "snap2")

	if len(cs.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(cs.Changes))
	}

	if cs.Changes[0].Type != models.ChangeTypeAdded {
		t.Errorf("Expected added change, got %s", cs.Changes[0].Type)
	}

	if cs.Summary[models.ChangeTypeAdded] != 1 {
		t.Errorf("Expected 1 added in summary")
	}
}
