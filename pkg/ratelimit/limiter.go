// Package ratelimit provides a token bucket rate limiter for API requests.
package ratelimit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// Limiter is a context-aware rate limiter using the token bucket algorithm.
type Limiter struct {
	limiter *rate.Limiter
	mu      sync.Mutex
}

// New creates a new rate limiter with the specified requests per minute.
// Default is 60 requests per minute (1 per second).
func New(requestsPerMinute int) *Limiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}

	// Convert requests per minute to requests per second
	rps := float64(requestsPerMinute) / 60.0
	burst := requestsPerMinute / 10 // Allow 10% burst capacity

	if burst < 1 {
		burst = 1
	}

	return &Limiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Wait blocks until the limiter permits an event or the context is cancelled.
// It returns an error if the context is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	return l.limiter.Wait(ctx)
}

// Allow reports whether an event may happen now.
// Use this for non-blocking checks.
func (l *Limiter) Allow() bool {
	return l.limiter.Allow()
}

// Reserve returns a Reservation that indicates how long the caller must wait.
// The reservation is valid for the specified duration.
func (l *Limiter) Reserve() *rate.Reservation {
	return l.limiter.Reserve()
}

// SetRate changes the rate limit to the new requests per minute.
func (l *Limiter) SetRate(requestsPerMinute int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}

	rps := float64(requestsPerMinute) / 60.0
	burst := requestsPerMinute / 10

	if burst < 1 {
		burst = 1
	}

	l.limiter.SetLimit(rate.Limit(rps))
	l.limiter.SetBurst(burst)
}

// WaitN blocks until the limiter permits n events or the context is cancelled.
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	return l.limiter.WaitN(ctx, n)
}

// AdaptiveLimiter is a rate limiter that adjusts its rate based on API responses.
type AdaptiveLimiter struct {
	limiter         *Limiter
	baseRate        int
	currentRate     int
	mu              sync.Mutex
	consecutiveOK   int
	consecutiveWait int
}

// NewAdaptive creates a new adaptive rate limiter.
func NewAdaptive(baseRequestsPerMinute int) *AdaptiveLimiter {
	if baseRequestsPerMinute <= 0 {
		baseRequestsPerMinute = 60
	}

	return &AdaptiveLimiter{
		limiter:     New(baseRequestsPerMinute),
		baseRate:    baseRequestsPerMinute,
		currentRate: baseRequestsPerMinute,
	}
}

// Wait blocks until the limiter permits an event.
func (al *AdaptiveLimiter) Wait(ctx context.Context) error {
	return al.limiter.Wait(ctx)
}

// RecordSuccess records a successful API call.
// After several consecutive successes, the rate may be increased.
func (al *AdaptiveLimiter) RecordSuccess() {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.consecutiveOK++
	al.consecutiveWait = 0

	// After 10 consecutive successes, try increasing the rate by 10%
	if al.consecutiveOK >= 10 && al.currentRate < al.baseRate*2 {
		newRate := int(float64(al.currentRate) * 1.1)
		if newRate > al.baseRate*2 {
			newRate = al.baseRate * 2
		}
		al.currentRate = newRate
		al.limiter.SetRate(newRate)
		al.consecutiveOK = 0
	}
}

// RecordRateLimit records a rate limit response from the API.
// The rate will be decreased to avoid hitting the limit.
func (al *AdaptiveLimiter) RecordRateLimit() {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.consecutiveWait++
	al.consecutiveOK = 0

	// Immediately reduce rate by 50%
	newRate := al.currentRate / 2
	if newRate < al.baseRate/4 {
		newRate = al.baseRate / 4 // Never go below 25% of base rate
	}

	al.currentRate = newRate
	al.limiter.SetRate(newRate)
}

// GetCurrentRate returns the current rate in requests per minute.
func (al *AdaptiveLimiter) GetCurrentRate() int {
	al.mu.Lock()
	defer al.mu.Unlock()
	return al.currentRate
}

// Reset resets the adaptive limiter to its base rate.
func (al *AdaptiveLimiter) Reset() {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.currentRate = al.baseRate
	al.consecutiveOK = 0
	al.consecutiveWait = 0
	al.limiter.SetRate(al.baseRate)
}
