package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/bss/radb-client/internal/models"
	"github.com/bss/radb-client/pkg/ratelimit"
)

// BulkResult contains the results of a bulk operation.
type BulkResult struct {
	Total     int           `json:"total"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
	Errors    []BulkError   `json:"errors,omitempty"`
}

// BulkError represents an error from a bulk operation.
type BulkError struct {
	Index   int    `json:"index"`
	ID      string `json:"id"`
	Error   string `json:"error"`
}

// BatchCreateRoutes creates multiple routes in parallel with rate limiting.
func (c *HTTPClient) BatchCreateRoutes(ctx context.Context, routes []*models.RouteObject, workers int) (*BulkResult, error) {
	c.logger.Infof("Starting batch create for %d routes with %d workers", len(routes), workers)

	if workers <= 0 {
		workers = 5 // Default to 5 workers
	}

	result := &BulkResult{
		Total:  len(routes),
		Errors: make([]BulkError, 0),
	}

	// Create rate limiter
	limiter := ratelimit.New(60) // 60 requests per minute

	// Create worker pool
	jobs := make(chan workJob, len(routes))
	results := make(chan workResult, len(routes))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Wait for rate limiter
				if err := limiter.Wait(ctx); err != nil {
					results <- workResult{
						Index: job.Index,
						ID:    job.ID,
						Error: err,
					}
					continue
				}

				// Execute create
				err := c.CreateRoute(ctx, job.Route)
				results <- workResult{
					Index: job.Index,
					ID:    job.ID,
					Error: err,
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for i, route := range routes {
			jobs <- workJob{
				Index: i,
				ID:    route.ID(),
				Route: route,
			}
		}
		close(jobs)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var mu sync.Mutex
	for res := range results {
		mu.Lock()
		if res.Error != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkError{
				Index: res.Index,
				ID:    res.ID,
				Error: res.Error.Error(),
			})
		} else {
			result.Succeeded++
		}
		mu.Unlock()
	}

	c.logger.Infof("Batch create completed: %d succeeded, %d failed", result.Succeeded, result.Failed)
	return result, nil
}

// BatchUpdateRoutes updates multiple routes in parallel with rate limiting.
func (c *HTTPClient) BatchUpdateRoutes(ctx context.Context, routes []*models.RouteObject, workers int) (*BulkResult, error) {
	c.logger.Infof("Starting batch update for %d routes with %d workers", len(routes), workers)

	if workers <= 0 {
		workers = 5
	}

	result := &BulkResult{
		Total:  len(routes),
		Errors: make([]BulkError, 0),
	}

	limiter := ratelimit.New(60)
	jobs := make(chan workJob, len(routes))
	results := make(chan workResult, len(routes))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := limiter.Wait(ctx); err != nil {
					results <- workResult{Index: job.Index, ID: job.ID, Error: err}
					continue
				}

				err := c.UpdateRoute(ctx, job.Route)
				results <- workResult{Index: job.Index, ID: job.ID, Error: err}
			}
		}()
	}

	go func() {
		for i, route := range routes {
			jobs <- workJob{Index: i, ID: route.ID(), Route: route}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var mu sync.Mutex
	for res := range results {
		mu.Lock()
		if res.Error != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkError{
				Index: res.Index,
				ID:    res.ID,
				Error: res.Error.Error(),
			})
		} else {
			result.Succeeded++
		}
		mu.Unlock()
	}

	c.logger.Infof("Batch update completed: %d succeeded, %d failed", result.Succeeded, result.Failed)
	return result, nil
}

// BatchDeleteRoutes deletes multiple routes in parallel with rate limiting.
func (c *HTTPClient) BatchDeleteRoutes(ctx context.Context, routes []RouteIdentifier, workers int) (*BulkResult, error) {
	c.logger.Infof("Starting batch delete for %d routes with %d workers", len(routes), workers)

	if workers <= 0 {
		workers = 5
	}

	result := &BulkResult{
		Total:  len(routes),
		Errors: make([]BulkError, 0),
	}

	limiter := ratelimit.New(60)
	jobs := make(chan deleteJob, len(routes))
	results := make(chan workResult, len(routes))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := limiter.Wait(ctx); err != nil {
					results <- workResult{Index: job.Index, ID: job.ID, Error: err}
					continue
				}

				err := c.DeleteRoute(ctx, job.Prefix, job.ASN)
				results <- workResult{Index: job.Index, ID: job.ID, Error: err}
			}
		}()
	}

	go func() {
		for i, route := range routes {
			jobs <- deleteJob{
				Index:  i,
				ID:     fmt.Sprintf("%s-%s", route.Prefix, route.ASN),
				Prefix: route.Prefix,
				ASN:    route.ASN,
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var mu sync.Mutex
	for res := range results {
		mu.Lock()
		if res.Error != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkError{
				Index: res.Index,
				ID:    res.ID,
				Error: res.Error.Error(),
			})
		} else {
			result.Succeeded++
		}
		mu.Unlock()
	}

	c.logger.Infof("Batch delete completed: %d succeeded, %d failed", result.Succeeded, result.Failed)
	return result, nil
}

// RouteIdentifier identifies a route for deletion.
type RouteIdentifier struct {
	Prefix string
	ASN    string
}

// workJob represents a work item for the worker pool.
type workJob struct {
	Index   int
	ID      string
	Route   *models.RouteObject
	Contact *models.Contact
}

// deleteJob represents a delete work item.
type deleteJob struct {
	Index  int
	ID     string
	Prefix string
	ASN    string
}

// workResult represents the result of a work item.
type workResult struct {
	Index int
	ID    string
	Error error
}
