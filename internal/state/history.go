package state

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bss/radb-client/internal/models"
	"github.com/sirupsen/logrus"
)

// HistoryManager manages the changelog file for tracking changes over time.
type HistoryManager struct {
	changelogPath string
	logger        *logrus.Logger
}

// NewHistoryManager creates a new history manager.
func NewHistoryManager(stateDir string, logger *logrus.Logger) *HistoryManager {
	return &HistoryManager{
		changelogPath: filepath.Join(stateDir, "changelog.jsonl"),
		logger:        logger,
	}
}

// AppendChanges appends a changeset to the changelog file in JSONL format.
// Each change is written as a separate JSON line for efficient append operations.
func (h *HistoryManager) AppendChanges(ctx context.Context, changeset *models.ChangeSet) error {
	if changeset == nil || changeset.IsEmpty() {
		h.logger.Debug("Skipping empty changeset")
		return nil
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(h.changelogPath), 0700); err != nil {
		return fmt.Errorf("failed to create changelog directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(h.changelogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open changelog file: %w", err)
	}
	defer file.Close()

	// Write each change as a JSON line
	encoder := json.NewEncoder(file)
	for _, change := range changeset.Changes {
		entry, err := models.NewChangelogEntry(change, changeset.ToSnapshot)
		if err != nil {
			h.logger.Warnf("Failed to create changelog entry: %v", err)
			continue
		}

		if err := encoder.Encode(entry); err != nil {
			return fmt.Errorf("failed to write changelog entry: %w", err)
		}
	}

	h.logger.Infof("Appended %d changes to changelog", len(changeset.Changes))
	return nil
}

// QueryChanges retrieves changes from the changelog within a time range.
func (h *HistoryManager) QueryChanges(ctx context.Context, from, to time.Time, objectType string) ([]models.ChangelogEntry, error) {
	if _, err := os.Stat(h.changelogPath); os.IsNotExist(err) {
		// No changelog file yet
		return []models.ChangelogEntry{}, nil
	}

	file, err := os.Open(h.changelogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open changelog file: %w", err)
	}
	defer file.Close()

	var entries []models.ChangelogEntry
	scanner := bufio.NewScanner(file)

	// Read each JSON line
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var entry models.ChangelogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			h.logger.Warnf("Failed to parse changelog entry: %v", err)
			continue
		}

		// Apply filters
		if entry.Timestamp.Before(from) || entry.Timestamp.After(to) {
			continue
		}

		if objectType != "" && entry.ObjectType != objectType {
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading changelog: %w", err)
	}

	return entries, nil
}

// GetChangesSince retrieves all changes since a specific timestamp.
func (h *HistoryManager) GetChangesSince(ctx context.Context, since time.Time) ([]models.ChangelogEntry, error) {
	return h.QueryChanges(ctx, since, time.Now(), "")
}

// GetRecentChanges retrieves the most recent N changes.
func (h *HistoryManager) GetRecentChanges(ctx context.Context, limit int) ([]models.ChangelogEntry, error) {
	if _, err := os.Stat(h.changelogPath); os.IsNotExist(err) {
		return []models.ChangelogEntry{}, nil
	}

	file, err := os.Open(h.changelogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open changelog file: %w", err)
	}
	defer file.Close()

	var allEntries []models.ChangelogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry models.ChangelogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			h.logger.Warnf("Failed to parse changelog entry: %v", err)
			continue
		}
		allEntries = append(allEntries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading changelog: %w", err)
	}

	// Return the last N entries
	if len(allEntries) <= limit {
		return allEntries, nil
	}

	return allEntries[len(allEntries)-limit:], nil
}

// GetStatistics computes statistics about changes in the changelog.
func (h *HistoryManager) GetStatistics(ctx context.Context, from, to time.Time) (*HistoryStatistics, error) {
	entries, err := h.QueryChanges(ctx, from, to, "")
	if err != nil {
		return nil, err
	}

	stats := &HistoryStatistics{
		TotalChanges: len(entries),
		ByType:       make(map[models.ChangeType]int),
		ByObjectType: make(map[string]int),
		TimeRange: TimeRange{
			From: from,
			To:   to,
		},
	}

	for _, entry := range entries {
		stats.ByType[entry.ChangeType]++
		stats.ByObjectType[entry.ObjectType]++

		if stats.FirstChange.IsZero() || entry.Timestamp.Before(stats.FirstChange) {
			stats.FirstChange = entry.Timestamp
		}
		if stats.LastChange.IsZero() || entry.Timestamp.After(stats.LastChange) {
			stats.LastChange = entry.Timestamp
		}
	}

	return stats, nil
}

// HistoryStatistics provides aggregate statistics about changes.
type HistoryStatistics struct {
	TotalChanges int                         `json:"total_changes"`
	ByType       map[models.ChangeType]int   `json:"by_type"`
	ByObjectType map[string]int              `json:"by_object_type"`
	FirstChange  time.Time                   `json:"first_change"`
	LastChange   time.Time                   `json:"last_change"`
	TimeRange    TimeRange                   `json:"time_range"`
}

// TimeRange represents a time range for queries.
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// Compact removes old entries from the changelog (for maintenance).
// This should be used carefully as it permanently removes historical data.
func (h *HistoryManager) Compact(ctx context.Context, keepAfter time.Time) error {
	if _, err := os.Stat(h.changelogPath); os.IsNotExist(err) {
		// No file to compact
		return nil
	}

	// Read all entries
	file, err := os.Open(h.changelogPath)
	if err != nil {
		return fmt.Errorf("failed to open changelog file: %w", err)
	}

	var keptEntries []models.ChangelogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry models.ChangelogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			h.logger.Warnf("Failed to parse changelog entry during compact: %v", err)
			continue
		}

		// Keep entries after the specified time
		if entry.Timestamp.After(keepAfter) {
			keptEntries = append(keptEntries, entry)
		}
	}

	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading changelog during compact: %w", err)
	}

	// Write back the kept entries
	tempPath := h.changelogPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	encoder := json.NewEncoder(tempFile)
	for _, entry := range keptEntries {
		if err := encoder.Encode(entry); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return fmt.Errorf("failed to write entry during compact: %w", err)
		}
	}

	tempFile.Close()

	// Replace original with compacted version
	if err := os.Rename(tempPath, h.changelogPath); err != nil {
		return fmt.Errorf("failed to replace changelog file: %w", err)
	}

	h.logger.Infof("Compacted changelog: kept %d entries, removed older entries", len(keptEntries))

	return nil
}
