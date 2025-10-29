package models

import (
	"encoding/json"
	"time"
)

// ChangelogEntry represents a single entry in the changelog file.
// Changelog entries are stored in JSONL format (one JSON object per line).
type ChangelogEntry struct {
	// Timestamp is when this change was recorded
	Timestamp time.Time `json:"timestamp"`

	// ChangeType indicates the type of change (added, removed, modified)
	ChangeType ChangeType `json:"change_type"`

	// ObjectType indicates what kind of object changed (route, contact)
	ObjectType string `json:"object_type"`

	// ObjectID uniquely identifies the changed object
	ObjectID string `json:"object_id"`

	// SnapshotID references the snapshot where this change was detected
	SnapshotID string `json:"snapshot_id"`

	// Before contains the object state before the change (for modified/removed)
	// Stored as JSON for flexibility
	Before json.RawMessage `json:"before,omitempty"`

	// After contains the object state after the change (for added/modified)
	// Stored as JSON for flexibility
	After json.RawMessage `json:"after,omitempty"`

	// FieldChanges lists which fields were modified (for modified changes)
	FieldChanges []string `json:"field_changes,omitempty"`

	// Note is an optional user-provided annotation
	Note string `json:"note,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewChangelogEntry creates a new changelog entry from a Change.
func NewChangelogEntry(change Change, snapshotID string) (*ChangelogEntry, error) {
	entry := &ChangelogEntry{
		Timestamp:  change.Timestamp,
		ChangeType: change.Type,
		ObjectType: change.ObjectType,
		ObjectID:   change.ObjectID,
		SnapshotID: snapshotID,
		Metadata:   make(map[string]string),
	}

	// Serialize before/after states
	if change.Before != nil {
		beforeJSON, err := json.Marshal(change.Before)
		if err != nil {
			return nil, err
		}
		entry.Before = beforeJSON
	}

	if change.After != nil {
		afterJSON, err := json.Marshal(change.After)
		if err != nil {
			return nil, err
		}
		entry.After = afterJSON
	}

	// Extract field changes from details
	if fieldChanges, ok := change.Details["field_changes"].([]string); ok {
		entry.FieldChanges = fieldChanges
	}

	return entry, nil
}

// ToChange converts a changelog entry back to a Change object.
func (e *ChangelogEntry) ToChange() Change {
	change := Change{
		Type:       e.ChangeType,
		ObjectType: e.ObjectType,
		ObjectID:   e.ObjectID,
		Timestamp:  e.Timestamp,
		Details:    make(map[string]interface{}),
	}

	// Deserialize before/after as raw JSON
	if len(e.Before) > 0 {
		var before interface{}
		json.Unmarshal(e.Before, &before)
		change.Before = before
	}

	if len(e.After) > 0 {
		var after interface{}
		json.Unmarshal(e.After, &after)
		change.After = after
	}

	// Add field changes to details
	if len(e.FieldChanges) > 0 {
		change.Details["field_changes"] = e.FieldChanges
	}

	if e.Note != "" {
		change.Details["note"] = e.Note
	}

	if e.SnapshotID != "" {
		change.Details["snapshot_id"] = e.SnapshotID
	}

	return change
}
