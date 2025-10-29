package state

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bss/radb-client/internal/models"
	"github.com/sirupsen/logrus"
)

func TestFileManager(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "radb-state-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mgr, err := NewFileManager(tmpDir, logger)
	if err != nil {
		t.Fatalf("NewFileManager() failed: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	t.Run("SaveAndLoadSnapshot", func(t *testing.T) {
		// Create a snapshot
		snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "test snapshot")
		snapshot.Routes = models.NewRouteList([]models.RouteObject{
			{
				Route:  "192.0.2.0/24",
				Origin: "AS64500",
				MntBy:  []string{"MAINT-TEST"},
				Source: "RADB",
			},
		})

		// Save
		if err := mgr.SaveSnapshot(ctx, snapshot); err != nil {
			t.Fatalf("SaveSnapshot() failed: %v", err)
		}

		// Load
		loaded, err := mgr.LoadSnapshot(ctx, snapshot.ID)
		if err != nil {
			t.Fatalf("LoadSnapshot() failed: %v", err)
		}

		if loaded.ID != snapshot.ID {
			t.Errorf("Expected ID %s, got %s", snapshot.ID, loaded.ID)
		}

		if loaded.Routes == nil || len(loaded.Routes.Routes) != 1 {
			t.Error("Routes not loaded correctly")
		}
	})

	t.Run("ListSnapshots", func(t *testing.T) {
		// Create multiple snapshots
		for i := 0; i < 3; i++ {
			snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "test")
			snapshot.Routes = models.NewRouteList([]models.RouteObject{})
			time.Sleep(time.Millisecond * 10) // Ensure different timestamps
			if err := mgr.SaveSnapshot(ctx, snapshot); err != nil {
				t.Fatalf("SaveSnapshot() failed: %v", err)
			}
		}

		snapshots, err := mgr.ListSnapshots(ctx)
		if err != nil {
			t.Fatalf("ListSnapshots() failed: %v", err)
		}

		if len(snapshots) < 3 {
			t.Errorf("Expected at least 3 snapshots, got %d", len(snapshots))
		}
	})

	t.Run("DeleteSnapshot", func(t *testing.T) {
		snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "to delete")
		snapshot.Routes = models.NewRouteList([]models.RouteObject{})

		if err := mgr.SaveSnapshot(ctx, snapshot); err != nil {
			t.Fatal(err)
		}

		if err := mgr.DeleteSnapshot(ctx, snapshot.ID); err != nil {
			t.Fatalf("DeleteSnapshot() failed: %v", err)
		}

		_, err := mgr.LoadSnapshot(ctx, snapshot.ID)
		if err == nil {
			t.Error("Expected error loading deleted snapshot")
		}
	})

	t.Run("ComputeChanges", func(t *testing.T) {
		// Create two snapshots with differences
		snap1 := models.NewSnapshot(models.SnapshotTypeRoute, "snapshot 1")
		snap1.Routes = models.NewRouteList([]models.RouteObject{
			{
				Route:  "192.0.2.0/24",
				Origin: "AS64500",
				MntBy:  []string{"MAINT-TEST"},
				Source: "RADB",
			},
		})

		snap2 := models.NewSnapshot(models.SnapshotTypeRoute, "snapshot 2")
		snap2.Routes = models.NewRouteList([]models.RouteObject{
			{
				Route:  "192.0.2.0/24",
				Origin: "AS64500",
				MntBy:  []string{"MAINT-TEST"},
				Source: "RADB",
			},
			{
				Route:  "198.51.100.0/24",
				Origin: "AS64501",
				MntBy:  []string{"MAINT-TEST"},
				Source: "RADB",
			},
		})

		changeset, err := mgr.ComputeChanges(ctx, snap1, snap2)
		if err != nil {
			t.Fatalf("ComputeChanges() failed: %v", err)
		}

		if len(changeset.Changes) != 1 {
			t.Errorf("Expected 1 change, got %d", len(changeset.Changes))
		}

		if changeset.Changes[0].Type != models.ChangeTypeAdded {
			t.Errorf("Expected added change, got %s", changeset.Changes[0].Type)
		}
	})
}

func TestSnapshotIntegrity(t *testing.T) {
	snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "test")
	snapshot.Routes = models.NewRouteList([]models.RouteObject{
		{
			Route:  "192.0.2.0/24",
			Origin: "AS64500",
			MntBy:  []string{"MAINT-TEST"},
			Source: "RADB",
		},
	})

	// Compute checksum
	if err := snapshot.ComputeChecksum(); err != nil {
		t.Fatalf("ComputeChecksum() failed: %v", err)
	}

	// Verify checksum
	if err := snapshot.VerifyChecksum(); err != nil {
		t.Fatalf("VerifyChecksum() failed: %v", err)
	}

	// Modify data and verify it fails
	snapshot.Routes.Routes[0].Origin = "AS64999"
	if err := snapshot.VerifyChecksum(); err == nil {
		t.Error("Expected checksum verification to fail after modification")
	}
}
