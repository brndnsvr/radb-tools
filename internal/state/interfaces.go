// Package state provides state management with file locking and integrity checks.
package state

import (
	"context"

	"github.com/bss/radb-client/internal/models"
)

// Manager defines the interface for state management operations.
type Manager interface {
	// Snapshot operations
	SaveSnapshot(ctx context.Context, snapshot *models.Snapshot) error
	LoadSnapshot(ctx context.Context, id string) (*models.Snapshot, error)
	GetLatestSnapshot(ctx context.Context, snapshotType models.SnapshotType) (*models.Snapshot, error)
	ListSnapshots(ctx context.Context) ([]models.Snapshot, error)
	DeleteSnapshot(ctx context.Context, id string) error

	// Change detection
	ComputeChanges(ctx context.Context, from, to *models.Snapshot) (*models.ChangeSet, error)

	// Maintenance
	Cleanup(ctx context.Context, options CleanupOptions) (*CleanupResult, error)
	Close() error
}
