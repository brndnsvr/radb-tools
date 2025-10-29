package state

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bss/radb-client/internal/models"
)

// CleanupOptions configures snapshot cleanup behavior.
type CleanupOptions struct {
	// KeepCount is the number of most recent snapshots to keep
	KeepCount int

	// KeepAfter is a timestamp; snapshots after this time are kept
	KeepAfter time.Time

	// KeepByType allows different retention per snapshot type
	KeepByType map[models.SnapshotType]int

	// DryRun if true, only reports what would be deleted without actually deleting
	DryRun bool
}

// CleanupResult contains the results of a cleanup operation.
type CleanupResult struct {
	TotalSnapshots   int      `json:"total_snapshots"`
	Kept             int      `json:"kept"`
	Deleted          int      `json:"deleted"`
	DeletedIDs       []string `json:"deleted_ids,omitempty"`
	Errors           []string `json:"errors,omitempty"`
	DryRun           bool     `json:"dry_run"`
}

// Cleanup removes old snapshots based on retention policies.
func (m *FileManager) Cleanup(ctx context.Context, options CleanupOptions) (*CleanupResult, error) {
	m.logger.Info("Starting snapshot cleanup")

	result := &CleanupResult{
		DeletedIDs: make([]string, 0),
		Errors:     make([]string, 0),
		DryRun:     options.DryRun,
	}

	// List all snapshots
	snapshots, err := m.ListSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	result.TotalSnapshots = len(snapshots)

	// Sort snapshots by timestamp (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	// Group by type if per-type retention is specified
	var toDelete []string

	if len(options.KeepByType) > 0 {
		toDelete = m.cleanupByType(snapshots, options)
	} else if options.KeepCount > 0 {
		toDelete = m.cleanupByCount(snapshots, options.KeepCount)
	} else if !options.KeepAfter.IsZero() {
		toDelete = m.cleanupByAge(snapshots, options.KeepAfter)
	} else {
		return nil, fmt.Errorf("no cleanup criteria specified")
	}

	result.Deleted = len(toDelete)
	result.Kept = result.TotalSnapshots - result.Deleted
	result.DeletedIDs = toDelete

	// Delete snapshots if not a dry run
	if !options.DryRun {
		for _, id := range toDelete {
			if err := m.DeleteSnapshot(ctx, id); err != nil {
				errMsg := fmt.Sprintf("failed to delete snapshot %s: %v", id, err)
				result.Errors = append(result.Errors, errMsg)
				m.logger.Warn(errMsg)
			}
		}
	}

	m.logger.Infof("Cleanup completed: kept %d, deleted %d (dry_run=%v)",
		result.Kept, result.Deleted, result.DryRun)

	return result, nil
}

// cleanupByCount keeps the N most recent snapshots.
func (m *FileManager) cleanupByCount(snapshots []models.Snapshot, keepCount int) []string {
	if keepCount >= len(snapshots) {
		return []string{}
	}

	toDelete := make([]string, 0)
	for i := keepCount; i < len(snapshots); i++ {
		toDelete = append(toDelete, snapshots[i].ID)
	}

	return toDelete
}

// cleanupByAge keeps snapshots after a certain date.
func (m *FileManager) cleanupByAge(snapshots []models.Snapshot, keepAfter time.Time) []string {
	toDelete := make([]string, 0)

	for _, snap := range snapshots {
		if snap.Timestamp.Before(keepAfter) {
			toDelete = append(toDelete, snap.ID)
		}
	}

	return toDelete
}

// cleanupByType keeps different numbers of snapshots per type.
func (m *FileManager) cleanupByType(snapshots []models.Snapshot, options CleanupOptions) []string {
	// Group snapshots by type
	byType := make(map[models.SnapshotType][]models.Snapshot)
	for _, snap := range snapshots {
		byType[snap.Type] = append(byType[snap.Type], snap)
	}

	toDelete := make([]string, 0)

	// Apply retention policy per type
	for snapshotType, snaps := range byType {
		keepCount, ok := options.KeepByType[snapshotType]
		if !ok {
			// Use default if not specified
			keepCount = options.KeepCount
		}

		if keepCount < len(snaps) {
			for i := keepCount; i < len(snaps); i++ {
				toDelete = append(toDelete, snaps[i].ID)
			}
		}
	}

	return toDelete
}

// CleanupByAge is a convenience method that keeps snapshots after a certain age.
func (m *FileManager) CleanupByAge(ctx context.Context, maxAge time.Duration, dryRun bool) (*CleanupResult, error) {
	keepAfter := time.Now().Add(-maxAge)
	options := CleanupOptions{
		KeepAfter: keepAfter,
		DryRun:    dryRun,
	}
	return m.Cleanup(ctx, options)
}

// CleanupByCount is a convenience method that keeps the N most recent snapshots.
func (m *FileManager) CleanupByCount(ctx context.Context, keepCount int, dryRun bool) (*CleanupResult, error) {
	options := CleanupOptions{
		KeepCount: keepCount,
		DryRun:    dryRun,
	}
	return m.Cleanup(ctx, options)
}

// AutoCleanup runs cleanup based on default policies.
// Keeps 30 route snapshots, 10 contact snapshots, and 5 full snapshots.
func (m *FileManager) AutoCleanup(ctx context.Context, dryRun bool) (*CleanupResult, error) {
	m.logger.Info("Running auto-cleanup with default policies")

	options := CleanupOptions{
		KeepByType: map[models.SnapshotType]int{
			models.SnapshotTypeRoute:   30,
			models.SnapshotTypeContact: 10,
			models.SnapshotTypeFull:    5,
		},
		DryRun: dryRun,
	}

	return m.Cleanup(ctx, options)
}
