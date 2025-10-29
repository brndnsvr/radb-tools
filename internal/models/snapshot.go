package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// SnapshotType defines the type of snapshot
type SnapshotType string

const (
	// SnapshotTypeRoute indicates a route object snapshot
	SnapshotTypeRoute SnapshotType = "route"

	// SnapshotTypeContact indicates a contact snapshot
	SnapshotTypeContact SnapshotType = "contact"

	// SnapshotTypeFull indicates a full snapshot of all data
	SnapshotTypeFull SnapshotType = "full"
)

// Snapshot represents a point-in-time capture of data.
// Snapshots are used for change detection and history tracking.
type Snapshot struct {
	// ID is a unique identifier for this snapshot
	ID string `json:"id"`

	// Timestamp is when this snapshot was created
	Timestamp time.Time `json:"timestamp"`

	// Type indicates what kind of data this snapshot contains
	Type SnapshotType `json:"type"`

	// Note is an optional user-provided description
	Note string `json:"note,omitempty"`

	// Checksum is a SHA-256 hash of the data for integrity verification
	Checksum string `json:"checksum"`

	// Version is the snapshot format version
	Version int `json:"version"`

	// Routes contains route objects (if Type is SnapshotTypeRoute or SnapshotTypeFull)
	Routes *RouteList `json:"routes,omitempty"`

	// Contacts contains contacts (if Type is SnapshotTypeContact or SnapshotTypeFull)
	Contacts *ContactList `json:"contacts,omitempty"`

	// Metadata contains additional snapshot information
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewSnapshot creates a new snapshot with the current timestamp.
func NewSnapshot(snapshotType SnapshotType, note string) *Snapshot {
	now := time.Now().UTC()
	return &Snapshot{
		ID:        fmt.Sprintf("%s-%d", snapshotType, now.Unix()),
		Timestamp: now,
		Type:      snapshotType,
		Note:      note,
		Version:   1,
		Metadata:  make(map[string]string),
	}
}

// ComputeChecksum calculates and updates the checksum for this snapshot.
// The checksum is computed over the data content (routes/contacts).
func (s *Snapshot) ComputeChecksum() error {
	// Create a consistent representation of the data
	data := struct {
		Routes   *RouteList   `json:"routes,omitempty"`
		Contacts *ContactList `json:"contacts,omitempty"`
	}{
		Routes:   s.Routes,
		Contacts: s.Contacts,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for checksum: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	s.Checksum = hex.EncodeToString(hash[:])

	return nil
}

// VerifyChecksum verifies the integrity of the snapshot.
func (s *Snapshot) VerifyChecksum() error {
	if s.Checksum == "" {
		return fmt.Errorf("no checksum present")
	}

	originalChecksum := s.Checksum
	s.Checksum = "" // Clear for recomputation

	if err := s.ComputeChecksum(); err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	if s.Checksum != originalChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", originalChecksum, s.Checksum)
	}

	return nil
}

// Validate performs basic validation on the snapshot.
func (s *Snapshot) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("snapshot ID is required")
	}

	if s.Timestamp.IsZero() {
		return fmt.Errorf("snapshot timestamp is required")
	}

	if s.Type == "" {
		return fmt.Errorf("snapshot type is required")
	}

	switch s.Type {
	case SnapshotTypeRoute:
		if s.Routes == nil {
			return fmt.Errorf("route snapshot must contain routes")
		}
	case SnapshotTypeContact:
		if s.Contacts == nil {
			return fmt.Errorf("contact snapshot must contain contacts")
		}
	case SnapshotTypeFull:
		if s.Routes == nil && s.Contacts == nil {
			return fmt.Errorf("full snapshot must contain at least routes or contacts")
		}
	default:
		return fmt.Errorf("invalid snapshot type: %s", s.Type)
	}

	return nil
}

// ChangeType represents the type of change detected
type ChangeType string

const (
	// ChangeTypeAdded indicates a new object was added
	ChangeTypeAdded ChangeType = "added"

	// ChangeTypeRemoved indicates an object was removed
	ChangeTypeRemoved ChangeType = "removed"

	// ChangeTypeModified indicates an object was modified
	ChangeTypeModified ChangeType = "modified"
)

// Change represents a detected change between snapshots.
type Change struct {
	// Type is the kind of change (added, removed, modified)
	Type ChangeType `json:"type"`

	// ObjectType is what kind of object changed (route, contact)
	ObjectType string `json:"object_type"`

	// ObjectID uniquely identifies the changed object
	ObjectID string `json:"object_id"`

	// Timestamp is when the change was detected
	Timestamp time.Time `json:"timestamp"`

	// Before contains the object state before the change (for modified/removed)
	Before interface{} `json:"before,omitempty"`

	// After contains the object state after the change (for added/modified)
	After interface{} `json:"after,omitempty"`

	// Details contains additional information about the change
	Details map[string]interface{} `json:"details,omitempty"`
}

// ChangeSet represents a collection of changes between two snapshots.
type ChangeSet struct {
	// FromSnapshot is the ID of the older snapshot
	FromSnapshot string `json:"from_snapshot"`

	// ToSnapshot is the ID of the newer snapshot
	ToSnapshot string `json:"to_snapshot"`

	// Timestamp is when this changeset was computed
	Timestamp time.Time `json:"timestamp"`

	// Changes is the list of detected changes
	Changes []Change `json:"changes"`

	// Summary provides counts of changes by type
	Summary map[ChangeType]int `json:"summary"`
}

// NewChangeSet creates a new changeset between two snapshots.
func NewChangeSet(fromID, toID string) *ChangeSet {
	return &ChangeSet{
		FromSnapshot: fromID,
		ToSnapshot:   toID,
		Timestamp:    time.Now().UTC(),
		Changes:      make([]Change, 0),
		Summary:      make(map[ChangeType]int),
	}
}

// AddChange adds a change to the changeset and updates the summary.
func (cs *ChangeSet) AddChange(change Change) {
	cs.Changes = append(cs.Changes, change)
	cs.Summary[change.Type]++
}

// IsEmpty returns true if there are no changes in this changeset.
func (cs *ChangeSet) IsEmpty() bool {
	return len(cs.Changes) == 0
}
