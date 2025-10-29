package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bss/radb-client/internal/models"
	"github.com/bss/radb-client/pkg/validator"
	"github.com/gofrs/flock"
	"github.com/sirupsen/logrus"
)

// FileManager implements the Manager interface with file-based storage.
type FileManager struct {
	stateDir string
	logger   *logrus.Logger
	lock     *flock.Flock
}

// NewFileManager creates a new file-based state manager.
func NewFileManager(stateDir string, logger *logrus.Logger) (*FileManager, error) {
	// Validate path
	if err := validator.ValidatePath(stateDir); err != nil {
		return nil, fmt.Errorf("invalid state directory: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	// Initialize file lock
	lockPath := filepath.Join(stateDir, ".lock")
	lock := flock.New(lockPath)

	return &FileManager{
		stateDir: stateDir,
		logger:   logger,
		lock:     lock,
	}, nil
}

// SaveSnapshot saves a snapshot to disk with file locking and checksumming.
func (fm *FileManager) SaveSnapshot(ctx context.Context, snapshot *models.Snapshot) error {
	// Acquire lock
	locked, err := fm.lock.TryLockContext(ctx, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return errors.New("could not acquire lock: timeout")
	}
	defer fm.lock.Unlock()

	// Validate snapshot
	if err := snapshot.Validate(); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	// Compute checksum
	if err := snapshot.ComputeChecksum(); err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Write atomically
	filename := fmt.Sprintf("%s.json", snapshot.ID)
	path := filepath.Join(fm.stateDir, filename)
	tmpPath := path + ".tmp"

	// Write to temp file
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	fm.logger.Infof("Saved snapshot %s (%d bytes)", snapshot.ID, len(data))
	return nil
}

// LoadSnapshot loads a snapshot from disk and verifies its integrity.
func (fm *FileManager) LoadSnapshot(ctx context.Context, id string) (*models.Snapshot, error) {
	// Acquire read lock
	locked, err := fm.lock.TryRLockContext(ctx, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return nil, errors.New("could not acquire lock: timeout")
	}
	defer fm.lock.Unlock()

	filename := fmt.Sprintf("%s.json", id)
	path := filepath.Join(fm.stateDir, filename)

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("snapshot not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	// Unmarshal
	var snapshot models.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	// Verify checksum
	if err := snapshot.VerifyChecksum(); err != nil {
		fm.logger.Warnf("Snapshot %s failed integrity check: %v", id, err)
		return nil, fmt.Errorf("snapshot integrity check failed: %w", err)
	}

	fm.logger.Debugf("Loaded snapshot %s", id)
	return &snapshot, nil
}

// GetLatestSnapshot retrieves the most recent snapshot of a given type.
func (fm *FileManager) GetLatestSnapshot(ctx context.Context, snapshotType models.SnapshotType) (*models.Snapshot, error) {
	snapshots, err := fm.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by type and sort by timestamp
	var filtered []models.Snapshot
	for _, s := range snapshots {
		if s.Type == snapshotType {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no snapshots found of type %s", snapshotType)
	}

	// Sort by timestamp descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Load the latest one
	return fm.LoadSnapshot(ctx, filtered[0].ID)
}

// ListSnapshots lists all available snapshots.
func (fm *FileManager) ListSnapshots(ctx context.Context) ([]models.Snapshot, error) {
	entries, err := os.ReadDir(fm.stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var snapshots []models.Snapshot
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Load snapshot metadata only
		path := filepath.Join(fm.stateDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fm.logger.Warnf("Failed to read %s: %v", entry.Name(), err)
			continue
		}

		var snapshot models.Snapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			fm.logger.Warnf("Failed to unmarshal %s: %v", entry.Name(), err)
			continue
		}

		snapshots = append(snapshots, snapshot)
	}

	// Sort by timestamp
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	return snapshots, nil
}

// DeleteSnapshot deletes a snapshot from disk.
func (fm *FileManager) DeleteSnapshot(ctx context.Context, id string) error {
	// Acquire lock
	locked, err := fm.lock.TryLockContext(ctx, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return errors.New("could not acquire lock: timeout")
	}
	defer fm.lock.Unlock()

	filename := fmt.Sprintf("%s.json", id)
	path := filepath.Join(fm.stateDir, filename)

	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("snapshot not found: %s", id)
		}
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	fm.logger.Infof("Deleted snapshot %s", id)
	return nil
}

// ComputeChanges computes the differences between two snapshots.
func (fm *FileManager) ComputeChanges(ctx context.Context, from, to *models.Snapshot) (*models.ChangeSet, error) {
	changeset := models.NewChangeSet(from.ID, to.ID)

	// Compare routes if present
	if from.Routes != nil && to.Routes != nil {
		fm.compareRoutes(from.Routes, to.Routes, changeset)
	}

	// Compare contacts if present
	if from.Contacts != nil && to.Contacts != nil {
		fm.compareContacts(from.Contacts, to.Contacts, changeset)
	}

	fm.logger.Debugf("Computed %d changes between %s and %s", len(changeset.Changes), from.ID, to.ID)
	return changeset, nil
}

// compareRoutes detects route changes.
func (fm *FileManager) compareRoutes(from, to *models.RouteList, changeset *models.ChangeSet) {
	fromMap := from.ByID()
	toMap := to.ByID()

	// Check for removed and modified routes
	for id, oldRoute := range fromMap {
		if newRoute, exists := toMap[id]; exists {
			// Route exists in both - check if modified
			if !routesEqual(oldRoute, newRoute) {
				changeset.AddChange(models.Change{
					Type:       models.ChangeTypeModified,
					ObjectType: "route",
					ObjectID:   id,
					Timestamp:  time.Now().UTC(),
					Before:     oldRoute,
					After:      newRoute,
				})
			}
		} else {
			// Route removed
			changeset.AddChange(models.Change{
				Type:       models.ChangeTypeRemoved,
				ObjectType: "route",
				ObjectID:   id,
				Timestamp:  time.Now().UTC(),
				Before:     oldRoute,
			})
		}
	}

	// Check for added routes
	for id, newRoute := range toMap {
		if _, exists := fromMap[id]; !exists {
			changeset.AddChange(models.Change{
				Type:       models.ChangeTypeAdded,
				ObjectType: "route",
				ObjectID:   id,
				Timestamp:  time.Now().UTC(),
				After:      newRoute,
			})
		}
	}
}

// compareContacts detects contact changes.
func (fm *FileManager) compareContacts(from, to *models.ContactList, changeset *models.ChangeSet) {
	fromMap := from.ByID()
	toMap := to.ByID()

	// Similar logic to compareRoutes
	for id, oldContact := range fromMap {
		if newContact, exists := toMap[id]; exists {
			if !contactsEqual(oldContact, newContact) {
				changeset.AddChange(models.Change{
					Type:       models.ChangeTypeModified,
					ObjectType: "contact",
					ObjectID:   id,
					Timestamp:  time.Now().UTC(),
					Before:     oldContact,
					After:      newContact,
				})
			}
		} else {
			changeset.AddChange(models.Change{
				Type:       models.ChangeTypeRemoved,
				ObjectType: "contact",
				ObjectID:   id,
				Timestamp:  time.Now().UTC(),
				Before:     oldContact,
			})
		}
	}

	for id, newContact := range toMap {
		if _, exists := fromMap[id]; !exists {
			changeset.AddChange(models.Change{
				Type:       models.ChangeTypeAdded,
				ObjectType: "contact",
				ObjectID:   id,
				Timestamp:  time.Now().UTC(),
				After:      newContact,
			})
		}
	}
}

// Cleanup implementation is in cleanup.go

// Close releases resources.
func (fm *FileManager) Close() error {
	if fm.lock != nil {
		return fm.lock.Unlock()
	}
	return nil
}
