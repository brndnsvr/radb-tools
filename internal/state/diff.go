package state

import (
	"context"
	"fmt"

	"github.com/bss/radb-client/internal/models"
)

// ComputeDiff calculates the differences between two snapshots using an O(n) algorithm.
// It uses hash maps for efficient comparison and detects added, removed, and modified items.
func ComputeDiff(ctx context.Context, from, to *models.Snapshot) (*models.DiffResult, error) {
	if from == nil || to == nil {
		return nil, fmt.Errorf("both snapshots must be non-nil")
	}

	result := models.NewDiffResult()

	// Compare routes if present in both snapshots
	if from.Routes != nil && to.Routes != nil {
		routeDiff := compareRoutes(from.Routes, to.Routes)
		result.Added = append(result.Added, routeDiff.Added...)
		result.Removed = append(result.Removed, routeDiff.Removed...)
		result.Modified = append(result.Modified, routeDiff.Modified...)
	} else if from.Routes == nil && to.Routes != nil {
		// All routes in 'to' are new
		for i := range to.Routes.Routes {
			result.Added = append(result.Added, &to.Routes.Routes[i])
		}
	} else if from.Routes != nil && to.Routes == nil {
		// All routes in 'from' were removed
		for i := range from.Routes.Routes {
			result.Removed = append(result.Removed, &from.Routes.Routes[i])
		}
	}

	// Compare contacts if present in both snapshots
	if from.Contacts != nil && to.Contacts != nil {
		contactDiff := compareContacts(from.Contacts, to.Contacts)
		result.Added = append(result.Added, contactDiff.Added...)
		result.Removed = append(result.Removed, contactDiff.Removed...)
		result.Modified = append(result.Modified, contactDiff.Modified...)
	} else if from.Contacts == nil && to.Contacts != nil {
		// All contacts in 'to' are new
		for i := range to.Contacts.Contacts {
			result.Added = append(result.Added, &to.Contacts.Contacts[i])
		}
	} else if from.Contacts != nil && to.Contacts == nil {
		// All contacts in 'from' were removed
		for i := range from.Contacts.Contacts {
			result.Removed = append(result.Removed, &from.Contacts.Contacts[i])
		}
	}

	// Compute summary statistics
	result.ComputeSummary()

	return result, nil
}

// compareRoutes performs an O(n) comparison of two route lists.
func compareRoutes(from, to *models.RouteList) *models.DiffResult {
	result := models.NewDiffResult()

	// Build hash maps for O(1) lookup
	fromMap := from.ByID()
	toMap := to.ByID()

	// Find added and modified routes
	for id, toRoute := range toMap {
		fromRoute, existsInFrom := fromMap[id]
		if !existsInFrom {
			// Route was added
			result.Added = append(result.Added, toRoute)
		} else {
			// Check if route was modified
			if !routesEqual(fromRoute, toRoute) {
				fieldChanges := models.DetectFieldChanges(fromRoute, toRoute)
				modified := models.ModifiedItem{
					ID:           id,
					ObjectType:   "route",
					Before:       fromRoute,
					After:        toRoute,
					FieldChanges: fieldChanges,
				}
				result.Modified = append(result.Modified, modified)
			}
		}
	}

	// Find removed routes
	for id, fromRoute := range fromMap {
		if _, existsInTo := toMap[id]; !existsInTo {
			// Route was removed
			result.Removed = append(result.Removed, fromRoute)
		}
	}

	return result
}

// compareContacts performs an O(n) comparison of two contact lists.
func compareContacts(from, to *models.ContactList) *models.DiffResult {
	result := models.NewDiffResult()

	// Build hash maps for O(1) lookup
	fromMap := from.ByID()
	toMap := to.ByID()

	// Find added and modified contacts
	for id, toContact := range toMap {
		fromContact, existsInFrom := fromMap[id]
		if !existsInFrom {
			// Contact was added
			result.Added = append(result.Added, toContact)
		} else {
			// Check if contact was modified
			if !contactsEqual(fromContact, toContact) {
				fieldChanges := models.DetectFieldChanges(fromContact, toContact)
				modified := models.ModifiedItem{
					ID:           id,
					ObjectType:   "contact",
					Before:       fromContact,
					After:        toContact,
					FieldChanges: fieldChanges,
				}
				result.Modified = append(result.Modified, modified)
			}
		}
	}

	// Find removed contacts
	for id, fromContact := range fromMap {
		if _, existsInTo := toMap[id]; !existsInTo {
			// Contact was removed
			result.Removed = append(result.Removed, fromContact)
		}
	}

	return result
}

// routesEqual checks if two routes are equal.
// We use a simple comparison here; could be optimized further.
func routesEqual(a, b *models.RouteObject) bool {
	// Quick checks
	if a.Route != b.Route || a.Origin != b.Origin || a.Source != b.Source {
		return false
	}

	// Compare string slices
	if !stringSliceEqual(a.Descr, b.Descr) {
		return false
	}
	if !stringSliceEqual(a.MntBy, b.MntBy) {
		return false
	}
	if !stringSliceEqual(a.Remarks, b.Remarks) {
		return false
	}
	if !stringSliceEqual(a.MemberOf, b.MemberOf) {
		return false
	}
	if !stringSliceEqual(a.Holes, b.Holes) {
		return false
	}

	return true
}

// contactsEqual checks if two contacts are equal.
func contactsEqual(a, b *models.Contact) bool {
	if a.ID != b.ID || a.Name != b.Name || a.Email != b.Email {
		return false
	}
	if a.Phone != b.Phone || a.Role != b.Role || a.Organization != b.Organization {
		return false
	}
	if !stringSliceEqual(a.Address, b.Address) {
		return false
	}

	return true
}

// stringSliceEqual checks if two string slices are equal.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// DiffToChangeSet converts a DiffResult to a ChangeSet.
func DiffToChangeSet(diff *models.DiffResult, fromID, toID string) *models.ChangeSet {
	cs := models.NewChangeSet(fromID, toID)

	// Convert added items to changes
	for _, item := range diff.Added {
		change := models.Change{
			Type:       models.ChangeTypeAdded,
			Timestamp:  cs.Timestamp,
			After:      item,
		}

		// Determine object type and ID
		switch v := item.(type) {
		case *models.RouteObject:
			change.ObjectType = "route"
			change.ObjectID = v.ID()
		case *models.Contact:
			change.ObjectType = "contact"
			change.ObjectID = v.ID
		}

		cs.AddChange(change)
	}

	// Convert removed items to changes
	for _, item := range diff.Removed {
		change := models.Change{
			Type:       models.ChangeTypeRemoved,
			Timestamp:  cs.Timestamp,
			Before:     item,
		}

		// Determine object type and ID
		switch v := item.(type) {
		case *models.RouteObject:
			change.ObjectType = "route"
			change.ObjectID = v.ID()
		case *models.Contact:
			change.ObjectType = "contact"
			change.ObjectID = v.ID
		}

		cs.AddChange(change)
	}

	// Convert modified items to changes
	for _, item := range diff.Modified {
		fieldNames := make([]string, len(item.FieldChanges))
		for i, fc := range item.FieldChanges {
			fieldNames[i] = fc.Field
		}

		change := models.Change{
			Type:       models.ChangeTypeModified,
			ObjectType: item.ObjectType,
			ObjectID:   item.ID,
			Timestamp:  cs.Timestamp,
			Before:     item.Before,
			After:      item.After,
			Details: map[string]interface{}{
				"field_changes": fieldNames,
			},
		}

		cs.AddChange(change)
	}

	return cs
}
