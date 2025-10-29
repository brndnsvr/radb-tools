package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bss/radb-client/internal/models"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// OutputFormat defines the output format type.
type OutputFormat string

const (
	// OutputFormatTable renders output as a table
	OutputFormatTable OutputFormat = "table"

	// OutputFormatJSON renders output as JSON
	OutputFormatJSON OutputFormat = "json"

	// OutputFormatYAML renders output as YAML
	OutputFormatYAML OutputFormat = "yaml"
)

// Outputter handles formatting and rendering output.
type Outputter struct {
	format OutputFormat
	writer io.Writer
	color  bool
}

// NewOutputter creates a new outputter.
func NewOutputter(format OutputFormat, writer io.Writer, enableColor bool) *Outputter {
	if writer == nil {
		writer = os.Stdout
	}
	return &Outputter{
		format: format,
		writer: writer,
		color:  enableColor,
	}
}

// RenderRoutes renders a list of routes.
func (o *Outputter) RenderRoutes(routes *models.RouteList) error {
	switch o.format {
	case OutputFormatJSON:
		return o.renderJSON(routes)
	case OutputFormatYAML:
		return o.renderYAML(routes)
	case OutputFormatTable:
		return o.renderRoutesTable(routes.Routes)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

// RenderContacts renders a list of contacts.
func (o *Outputter) RenderContacts(contacts *models.ContactList) error {
	switch o.format {
	case OutputFormatJSON:
		return o.renderJSON(contacts)
	case OutputFormatYAML:
		return o.renderYAML(contacts)
	case OutputFormatTable:
		return o.renderContactsTable(contacts.Contacts)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

// RenderSnapshots renders a list of snapshots.
func (o *Outputter) RenderSnapshots(snapshots []models.Snapshot) error {
	switch o.format {
	case OutputFormatJSON:
		return o.renderJSON(snapshots)
	case OutputFormatYAML:
		return o.renderYAML(snapshots)
	case OutputFormatTable:
		return o.renderSnapshotsTable(snapshots)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

// RenderDiff renders a diff result with color highlighting.
func (o *Outputter) RenderDiff(diff *models.DiffResult) error {
	switch o.format {
	case OutputFormatJSON:
		return o.renderJSON(diff)
	case OutputFormatYAML:
		return o.renderYAML(diff)
	case OutputFormatTable:
		return o.renderDiffTable(diff)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

// renderJSON renders data as JSON.
func (o *Outputter) renderJSON(data interface{}) error {
	encoder := json.NewEncoder(o.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// renderYAML renders data as YAML.
func (o *Outputter) renderYAML(data interface{}) error {
	encoder := yaml.NewEncoder(o.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// renderRoutesTable renders routes as a table.
func (o *Outputter) renderRoutesTable(routes []models.RouteObject) error {
	table := tablewriter.NewWriter(o.writer)
	table.Header("Route", "Origin", "Maintainer", "Description")

	for _, route := range routes {
		descr := strings.Join(route.Descr, ", ")
		if len(descr) > 50 {
			descr = descr[:47] + "..."
		}
		mntBy := strings.Join(route.MntBy, ", ")
		if len(mntBy) > 30 {
			mntBy = mntBy[:27] + "..."
		}

		table.Append(route.Route, route.Origin, mntBy, descr)
	}

	return table.Render()
}

// renderContactsTable renders contacts as a table.
func (o *Outputter) renderContactsTable(contacts []models.Contact) error {
	table := tablewriter.NewWriter(o.writer)
	table.Header("ID", "Name", "Email", "Role", "Organization")

	for _, contact := range contacts {
		table.Append(contact.ID, contact.Name, contact.Email, string(contact.Role), contact.Organization)
	}

	return table.Render()
}

// renderSnapshotsTable renders snapshots as a table.
func (o *Outputter) renderSnapshotsTable(snapshots []models.Snapshot) error {
	table := tablewriter.NewWriter(o.writer)
	table.Header("ID", "Type", "Timestamp", "Note", "Items")

	for _, snap := range snapshots {
		items := 0
		if snap.Routes != nil {
			items += snap.Routes.Count
		}
		if snap.Contacts != nil {
			items += snap.Contacts.Count
		}

		table.Append(snap.ID, string(snap.Type), snap.Timestamp.Format("2006-01-02 15:04:05"), snap.Note, fmt.Sprintf("%d", items))
	}

	return table.Render()
}

// renderDiffTable renders a diff as a table with color.
func (o *Outputter) renderDiffTable(diff *models.DiffResult) error {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	if !o.color {
		color.NoColor = true
	}

	// Summary
	fmt.Fprintf(o.writer, "Summary:\n")
	fmt.Fprintf(o.writer, "  Added:    %s\n", green.Sprintf("%d", diff.Summary.AddedCount))
	fmt.Fprintf(o.writer, "  Removed:  %s\n", red.Sprintf("%d", diff.Summary.RemovedCount))
	fmt.Fprintf(o.writer, "  Modified: %s\n", yellow.Sprintf("%d", diff.Summary.ModifiedCount))
	fmt.Fprintf(o.writer, "  Total:    %d\n\n", diff.Summary.TotalChanges)

	// Added items
	if len(diff.Added) > 0 {
		fmt.Fprintf(o.writer, "%s:\n", green.Sprint("Added"))
		table := tablewriter.NewWriter(o.writer)
		table.Header("Type", "ID", "Details")

		for _, item := range diff.Added {
			typeStr, id, details := formatDiffItem(item)
			table.Append(typeStr, id, details)
		}
		table.Render()
		fmt.Fprintln(o.writer)
	}

	// Removed items
	if len(diff.Removed) > 0 {
		fmt.Fprintf(o.writer, "%s:\n", red.Sprint("Removed"))
		table := tablewriter.NewWriter(o.writer)
		table.Header("Type", "ID", "Details")

		for _, item := range diff.Removed {
			typeStr, id, details := formatDiffItem(item)
			table.Append(typeStr, id, details)
		}
		table.Render()
		fmt.Fprintln(o.writer)
	}

	// Modified items
	if len(diff.Modified) > 0 {
		fmt.Fprintf(o.writer, "%s:\n", yellow.Sprint("Modified"))
		table := tablewriter.NewWriter(o.writer)
		table.Header("Type", "ID", "Changed Fields")

		for _, item := range diff.Modified {
			fields := make([]string, len(item.FieldChanges))
			for i, fc := range item.FieldChanges {
				fields[i] = fc.Field
			}
			table.Append(item.ObjectType, item.ID, strings.Join(fields, ", "))
		}
		table.Render()
	}

	return nil
}

// formatDiffItem extracts information from a diff item for display.
func formatDiffItem(item interface{}) (typeStr, id, details string) {
	switch v := item.(type) {
	case *models.RouteObject:
		return "route", v.ID(), fmt.Sprintf("%s -> %s", v.Route, v.Origin)
	case *models.Contact:
		return "contact", v.ID, fmt.Sprintf("%s <%s>", v.Name, v.Email)
	default:
		return "unknown", "unknown", "N/A"
	}
}

// RenderChangeHistory renders changelog entries.
func (o *Outputter) RenderChangeHistory(entries []models.ChangelogEntry) error {
	switch o.format {
	case OutputFormatJSON:
		return o.renderJSON(entries)
	case OutputFormatYAML:
		return o.renderYAML(entries)
	case OutputFormatTable:
		return o.renderChangeHistoryTable(entries)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

// renderChangeHistoryTable renders changelog entries as a table.
func (o *Outputter) renderChangeHistoryTable(entries []models.ChangelogEntry) error {
	table := tablewriter.NewWriter(o.writer)
	table.Header("Timestamp", "Type", "Object Type", "Object ID", "Fields")

	for _, entry := range entries {
		fields := strings.Join(entry.FieldChanges, ", ")
		if len(fields) > 40 {
			fields = fields[:37] + "..."
		}

		table.Append(entry.Timestamp.Format("2006-01-02 15:04:05"), string(entry.ChangeType), entry.ObjectType, entry.ObjectID, fields)
	}

	return table.Render()
}
