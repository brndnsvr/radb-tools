# Phase 2 Integration Guide

This document provides guidance for integrating Phase 2 components with Phase 1 and Phase 3.

## Integration with Phase 1

### Required from Phase 1

Phase 2 expects these interfaces and implementations from Phase 1:

#### 1. HTTP Client (internal/api/client.go)

```go
type Client struct {
    baseURL string
    source  string
    // Add your Phase 1 fields:
    // httpClient *http.Client
    // auth       Authenticator
    // logger     *logrus.Logger
}

// These methods must be implemented:
func (c *Client) Get(ctx context.Context, endpoint string, result interface{}) error
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}, result interface{}) error
func (c *Client) Put(ctx context.Context, endpoint string, body interface{}, result interface{}) error
func (c *Client) Delete(ctx context.Context, endpoint string) error
```

**Integration Steps:**
1. Implement the four HTTP methods in `internal/api/client.go`
2. Handle authentication (Basic Auth + API key)
3. Add retry logic with exponential backoff
4. Add rate limiting
5. Parse API responses and populate result structs

#### 2. State Manager (internal/state/manager.go)

Phase 1 already implemented `FileManager`. Phase 2 enhances it with:
- Diff generation (diff.go)
- History management (history.go)

**Integration:**
- ✅ Already compatible
- Phase 2 uses existing `FileManager.SaveSnapshot()` and `LoadSnapshot()`
- Phase 2 adds complementary diff and history features

#### 3. Models

Phase 1 created base models. Phase 2 added:
- `internal/models/diff.go` - Diff structures

**Minor alignment needed:**
- Phase 2's diff.go may reference old field names (e.g., `Route` vs `RouteObject`)
- Update diff.go to use `RouteObject.ID()` instead of `RouteObject.Key()`

### Integration Checklist

- [ ] Implement `Client.Get()`, `Client.Post()`, `Client.Put()`, `Client.Delete()`
- [ ] Update diff.go to use `RouteObject` instead of `Route`
- [ ] Ensure model field names match (Descr vs Description, etc.)
- [ ] Test API client with Phase 2 route operations
- [ ] Verify snapshot compatibility

## Integration with Phase 3

Phase 3 will add CLI commands. Here's how to wire up Phase 2 functionality:

### Route Commands (internal/cli/route.go)

```go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/bss/radb-client/internal/api"
    "github.com/bss/radb-client/internal/cli"
)

var routeCmd = &cobra.Command{
    Use:   "route",
    Short: "Manage route objects",
}

var routeListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all routes",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Get API client from context/config
        client := getAPIClient(cmd)
        routeAPI := api.NewRouteAPI(client)

        // 2. Call Phase 2 route API
        routes, err := routeAPI.List(cmd.Context(), nil)
        if err != nil {
            return err
        }

        // 3. Use Phase 2 formatter
        formatter := cli.NewFormatter(getOutputOptions(cmd))
        return formatter.FormatRoutes(routes)
    },
}

var routeShowCmd = &cobra.Command{
    Use:   "show <prefix> <asn>",
    Short: "Show a specific route",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        client := getAPIClient(cmd)
        routeAPI := api.NewRouteAPI(client)

        route, err := routeAPI.Get(cmd.Context(), args[0], args[1])
        if err != nil {
            return err
        }

        formatter := cli.NewFormatter(getOutputOptions(cmd))
        return formatter.FormatRoute(route)
    },
}

// Add create, update, delete commands similarly...
```

### Diff Commands (internal/cli/diff.go)

```go
var diffCmd = &cobra.Command{
    Use:   "diff [snapshot1] [snapshot2]",
    Short: "Show differences between snapshots",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Load snapshots using Phase 1 FileManager
        manager := getStateManager(cmd)

        var old, new *models.Snapshot
        var err error

        if len(args) == 0 {
            // Compare last two snapshots
            snapshots, err := manager.ListSnapshots(cmd.Context())
            if err != nil {
                return err
            }
            if len(snapshots) < 2 {
                return fmt.Errorf("need at least 2 snapshots")
            }
            old = &snapshots[1]
            new = &snapshots[0]
        } else {
            // Load specific snapshots
            old, err = manager.LoadSnapshot(cmd.Context(), args[0])
            if err != nil {
                return err
            }
            new, err = manager.LoadSnapshot(cmd.Context(), args[1])
            if err != nil {
                return err
            }
        }

        // 2. Generate diff using Phase 2 DiffGenerator
        diffGen := state.NewDiffGenerator(models.DiffOptions{})

        var diff interface{}
        if old.Type == models.SnapshotTypeRoute {
            diff, err = diffGen.DiffRoutes(old, new)
        } else {
            diff, err = diffGen.DiffContacts(old, new)
        }

        if err != nil {
            return err
        }

        // 3. Format using Phase 2 Formatter
        formatter := cli.NewFormatter(getOutputOptions(cmd))
        return formatter.FormatDiff(diff)
    },
}
```

### Auto-Snapshotting on List

Automatically create snapshots when listing routes/contacts:

```go
var routeListCmd = &cobra.Command{
    Use:   "list",
    RunE: func(cmd *cobra.Command, args []string) error {
        client := getAPIClient(cmd)
        routeAPI := api.NewRouteAPI(client)

        // Fetch routes
        routes, err := routeAPI.List(cmd.Context(), nil)
        if err != nil {
            return err
        }

        // Auto-snapshot (if enabled in config)
        if shouldAutoSnapshot(cmd) {
            manager := getStateManager(cmd)
            snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "auto")
            snapshot.Routes = routes

            if err := manager.SaveSnapshot(cmd.Context(), snapshot); err != nil {
                // Log warning but don't fail
                logWarning(cmd, "Failed to save snapshot: %v", err)
            }
        }

        // Format output
        formatter := cli.NewFormatter(getOutputOptions(cmd))
        return formatter.FormatRoutes(routes)
    },
}
```

## Testing Integration

### Mock API Client

Create a mock client for testing Phase 2 components:

```go
// internal/api/mock_client_test.go
type MockClient struct {
    GetFunc    func(ctx context.Context, endpoint string, result interface{}) error
    PostFunc   func(ctx context.Context, endpoint string, body interface{}, result interface{}) error
    PutFunc    func(ctx context.Context, endpoint string, body interface{}, result interface{}) error
    DeleteFunc func(ctx context.Context, endpoint string) error
}

func (m *MockClient) Get(ctx context.Context, endpoint string, result interface{}) error {
    if m.GetFunc != nil {
        return m.GetFunc(ctx, endpoint, result)
    }
    return nil
}

// Similar for Post, Put, Delete...
```

### End-to-End Test Example

```go
func TestRouteOperations_EndToEnd(t *testing.T) {
    // Setup
    mockClient := &MockClient{
        PostFunc: func(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
            // Simulate successful creation
            route := result.(*models.RouteObject)
            route.Created = &now
            return nil
        },
    }

    routeAPI := api.NewRouteAPI(mockClient)

    // Test
    route := &models.RouteObject{
        Route:  "192.0.2.0/24",
        Origin: "AS64500",
        MntBy:  []string{"MAINT-AS64500"},
        Source: "RADB",
    }

    err := routeAPI.Create(context.Background(), route)
    assert.NoError(t, err)
    assert.NotNil(t, route.Created)
}
```

## Configuration

### Add Phase 2 settings to config.yaml

```yaml
api:
  base_url: https://api.radb.net
  source: RADB
  format: json
  timeout: 30

state:
  auto_snapshot: true           # Auto-snapshot on list operations
  snapshot_dir: ~/.radb-client/snapshots
  history_dir: ~/.radb-client/history

history:
  compression_enabled: true      # Use gzip for old snapshots
  compression_age_days: 7        # Compress snapshots older than 7 days
  retention_count: 30            # Keep last 30 snapshots
  changelog_enabled: true        # Enable JSONL changelog

output:
  default_format: table          # table, json, yaml, raw
  colored: true                  # Enable colored output
  show_headers: true             # Show table headers
```

## Error Handling Integration

Phase 2 provides rich error types. Handle them in CLI:

```go
func handleError(cmd *cobra.Command, err error) {
    formatter := cli.NewFormatter(getOutputOptions(cmd))

    switch e := err.(type) {
    case *api.ValidationError:
        formatter.FormatError(err)
        if e.Suggestion != "" {
            formatter.FormatInfo(fmt.Sprintf("Suggestion: %s", e.Suggestion))
        }
        os.Exit(1)

    case *api.NotFoundError:
        formatter.FormatError(err)
        if e.Suggestion != "" {
            formatter.FormatInfo(fmt.Sprintf("Suggestion: %s", e.Suggestion))
        }
        os.Exit(1)

    case *api.ConflictError:
        formatter.FormatError(err)
        if e.Suggestion != "" {
            formatter.FormatInfo(fmt.Sprintf("Suggestion: %s", e.Suggestion))
        }
        os.Exit(1)

    default:
        formatter.FormatError(err)
        os.Exit(1)
    }
}
```

## Performance Considerations

### Large Datasets

Phase 2 is optimized for large datasets:

```go
// Use streaming for large result sets
if routeCount > 1000 {
    // Enable chunked processing
    opts := &api.ListOptions{
        Limit:  100,  // Fetch in batches of 100
        Offset: 0,
    }

    for {
        routes, err := routeAPI.List(ctx, opts)
        if err != nil {
            return err
        }

        // Process chunk
        processRoutes(routes)

        if len(routes.Routes) < opts.Limit {
            break  // Last page
        }

        opts.Offset += opts.Limit
    }
}
```

### Compression

Enable compression for historical snapshots:

```go
// In background goroutine or cron job
func cleanupOldSnapshots(ctx context.Context) {
    manager := getStateManager()
    histManager := state.NewHistoryManager(manager)

    // Compress snapshots older than 7 days
    age := 7 * 24 * time.Hour
    if err := histManager.CompressOldSnapshots(age); err != nil {
        logError("Failed to compress snapshots: %v", err)
    }

    // Keep only last 30 snapshots
    if err := histManager.CleanupOldSnapshots(models.SnapshotTypeRoute, 30); err != nil {
        logError("Failed to cleanup snapshots: %v", err)
    }
}
```

## Monitoring & Observability

### Add Metrics

```go
// Track API operation metrics
type Metrics struct {
    routeListCalls    prometheus.Counter
    routeCreateCalls  prometheus.Counter
    diffGenerations   prometheus.Counter
    snapshotsSaved    prometheus.Counter
}

// In route list command
metrics.routeListCalls.Inc()

// In snapshot save
metrics.snapshotsSaved.Inc()

// In diff generation
metrics.diffGenerations.Inc()
```

### Logging

Phase 2 uses structured logging:

```go
import "github.com/sirupsen/logrus"

logger := logrus.New()
logger.Infof("Listing routes with limit=%d offset=%d", opts.Limit, opts.Offset)
logger.Debugf("Snapshot saved: %s", snapshot.ID)
logger.Warnf("Snapshot integrity check failed: %v", err)
logger.Errorf("API call failed: %v", err)
```

## Troubleshooting

### Common Integration Issues

1. **Model field mismatch**
   - Symptom: nil pointer or missing data in diffs
   - Fix: Ensure Phase 1 and Phase 2 use same field names
   - Check: `RouteObject.ID()` vs `Route.Key()`

2. **HTTP client not implemented**
   - Symptom: "client not implemented by Phase 1 yet" error
   - Fix: Implement Client.Get/Post/Put/Delete in Phase 1

3. **Snapshot format incompatible**
   - Symptom: Checksum verification fails
   - Fix: Ensure Phase 1 and Phase 2 use same Snapshot structure

4. **Diff returns empty results**
   - Symptom: No changes detected when there should be
   - Fix: Verify RouteObject.ID() returns consistent keys

## Next Steps

1. **Complete Phase 1** - Implement HTTP client methods
2. **Align Models** - Ensure field names match between phases
3. **Wire CLI** - Create command files in Phase 3
4. **Integration Tests** - Test full flow end-to-end
5. **Documentation** - Update user docs with examples

## Questions?

If you encounter issues integrating Phase 2:

1. Check model alignment (route.go, contact.go, snapshot.go)
2. Verify HTTP client implementation
3. Test with mock client first
4. Review PHASE2_IMPLEMENTATION.md for details
5. Check error messages for suggestions

---

**Phase 2 Integration Team**
