package models

import (
	"encoding/json"
	"reflect"
)

// DiffResult contains the results of comparing two snapshots.
type DiffResult struct {
	// Added contains objects that exist in the new snapshot but not the old
	Added []interface{} `json:"added"`

	// Removed contains objects that exist in the old snapshot but not the new
	Removed []interface{} `json:"removed"`

	// Modified contains objects that exist in both but have changed
	Modified []ModifiedItem `json:"modified"`

	// Summary provides statistics about the diff
	Summary DiffSummary `json:"summary"`
}

// ModifiedItem represents an object that was modified between snapshots.
type ModifiedItem struct {
	// ID is the unique identifier for the object
	ID string `json:"id"`

	// ObjectType indicates what kind of object this is (route, contact)
	ObjectType string `json:"object_type"`

	// Before is the state of the object in the old snapshot
	Before interface{} `json:"before"`

	// After is the state of the object in the new snapshot
	After interface{} `json:"after"`

	// FieldChanges lists which fields were modified
	FieldChanges []FieldChange `json:"field_changes"`
}

// FieldChange represents a change to a specific field.
type FieldChange struct {
	// Field is the name of the field that changed
	Field string `json:"field"`

	// OldValue is the previous value (as JSON for flexibility)
	OldValue json.RawMessage `json:"old_value,omitempty"`

	// NewValue is the new value (as JSON for flexibility)
	NewValue json.RawMessage `json:"new_value,omitempty"`
}

// DiffSummary provides statistics about a diff.
type DiffSummary struct {
	// AddedCount is the number of added objects
	AddedCount int `json:"added_count"`

	// RemovedCount is the number of removed objects
	RemovedCount int `json:"removed_count"`

	// ModifiedCount is the number of modified objects
	ModifiedCount int `json:"modified_count"`

	// TotalChanges is the sum of all changes
	TotalChanges int `json:"total_changes"`

	// ByType breaks down changes by object type
	ByType map[string]TypeSummary `json:"by_type,omitempty"`
}

// TypeSummary provides change statistics for a specific object type.
type TypeSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
}

// NewDiffResult creates a new empty diff result.
func NewDiffResult() *DiffResult {
	return &DiffResult{
		Added:    make([]interface{}, 0),
		Removed:  make([]interface{}, 0),
		Modified: make([]ModifiedItem, 0),
		Summary: DiffSummary{
			ByType: make(map[string]TypeSummary),
		},
	}
}

// IsEmpty returns true if there are no differences.
func (dr *DiffResult) IsEmpty() bool {
	return len(dr.Added) == 0 && len(dr.Removed) == 0 && len(dr.Modified) == 0
}

// ComputeSummary calculates the summary statistics.
func (dr *DiffResult) ComputeSummary() {
	dr.Summary.AddedCount = len(dr.Added)
	dr.Summary.RemovedCount = len(dr.Removed)
	dr.Summary.ModifiedCount = len(dr.Modified)
	dr.Summary.TotalChanges = dr.Summary.AddedCount + dr.Summary.RemovedCount + dr.Summary.ModifiedCount
}

// DetectFieldChanges compares two objects and returns the list of changed fields.
func DetectFieldChanges(before, after interface{}) []FieldChange {
	changes := make([]FieldChange, 0)

	// Use reflection to compare fields
	beforeVal := reflect.ValueOf(before)
	afterVal := reflect.ValueOf(after)

	// Handle pointers
	if beforeVal.Kind() == reflect.Ptr {
		beforeVal = beforeVal.Elem()
	}
	if afterVal.Kind() == reflect.Ptr {
		afterVal = afterVal.Elem()
	}

	// Only works for structs
	if beforeVal.Kind() != reflect.Struct || afterVal.Kind() != reflect.Struct {
		return changes
	}

	// Compare each field
	beforeType := beforeVal.Type()
	for i := 0; i < beforeVal.NumField(); i++ {
		field := beforeType.Field(i)
		beforeField := beforeVal.Field(i)
		afterField := afterVal.Field(i)

		// Skip unexported fields
		if !beforeField.CanInterface() {
			continue
		}

		// Compare field values
		if !reflect.DeepEqual(beforeField.Interface(), afterField.Interface()) {
			// Serialize to JSON for storage
			oldJSON, _ := json.Marshal(beforeField.Interface())
			newJSON, _ := json.Marshal(afterField.Interface())

			changes = append(changes, FieldChange{
				Field:    field.Name,
				OldValue: oldJSON,
				NewValue: newJSON,
			})
		}
	}

	return changes
}
