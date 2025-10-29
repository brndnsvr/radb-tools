package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	limiter := New(60)
	if limiter == nil {
		t.Fatal("Failed to create limiter")
	}
}

func TestLimiterWait(t *testing.T) {
	limiter := New(600) // 600 requests per minute = 10 per second

	ctx := context.Background()

	start := time.Now()

	// First request should be immediate
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > 10*time.Millisecond {
		t.Errorf("First request took too long: %v", elapsed)
	}

	// Multiple requests should be rate limited
	for i := 0; i < 5; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}
}

func TestLimiterAllow(t *testing.T) {
	limiter := New(60)

	// First request should be allowed
	if !limiter.Allow() {
		t.Error("First request should be allowed")
	}
}

func TestLimiterContextCancellation(t *testing.T) {
	limiter := New(1) // Very slow rate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// First wait succeeds
	limiter.Wait(context.Background())

	// Second wait should fail due to context timeout
	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

func TestAdaptiveLimiter(t *testing.T) {
	limiter := NewAdaptive(60)

	if limiter.GetCurrentRate() != 60 {
		t.Errorf("Expected initial rate of 60, got %d", limiter.GetCurrentRate())
	}

	// Record success multiple times
	for i := 0; i < 20; i++ {
		limiter.RecordSuccess()
	}

	// Rate should have increased
	if limiter.GetCurrentRate() <= 60 {
		t.Errorf("Expected rate to increase after successes")
	}

	// Record rate limit
	limiter.RecordRateLimit()

	// Rate should have decreased
	if limiter.GetCurrentRate() >= 60 {
		t.Errorf("Expected rate to decrease after rate limit")
	}

	// Reset should restore base rate
	limiter.Reset()
	if limiter.GetCurrentRate() != 60 {
		t.Errorf("Expected rate to reset to 60, got %d", limiter.GetCurrentRate())
	}
}

func TestSetRate(t *testing.T) {
	limiter := New(60)

	limiter.SetRate(120)

	// Rate should be updated (we can't directly test the internal rate,
	// but we can verify the limiter still works)
	ctx := context.Background()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("Wait failed after SetRate: %v", err)
	}
}
