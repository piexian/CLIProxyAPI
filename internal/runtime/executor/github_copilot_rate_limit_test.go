package executor

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGitHubCopilotRateLimiterFailFastReturns429(t *testing.T) {
	limiter := &githubCopilotRateLimiter{}
	if err := limiter.acquire(context.Background(), 50*time.Millisecond, false); err != nil {
		t.Fatalf("first acquire error = %v, want nil", err)
	}

	err := limiter.acquire(context.Background(), 50*time.Millisecond, false)
	if err == nil {
		t.Fatalf("second acquire error = nil, want 429")
	}
	status, ok := err.(statusErr)
	if !ok {
		t.Fatalf("error type = %T, want statusErr", err)
	}
	if status.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("status code = %d, want %d", status.StatusCode(), http.StatusTooManyRequests)
	}
	if !strings.Contains(status.Error(), "rate_limit_exceeded") {
		t.Fatalf("error body = %q, want rate_limit_exceeded", status.Error())
	}
}

func TestGitHubCopilotRateLimiterQueueWaits(t *testing.T) {
	limiter := &githubCopilotRateLimiter{}
	if err := limiter.acquire(context.Background(), 40*time.Millisecond, true); err != nil {
		t.Fatalf("first acquire error = %v, want nil", err)
	}

	start := time.Now()
	if err := limiter.acquire(context.Background(), 40*time.Millisecond, true); err != nil {
		t.Fatalf("second acquire error = %v, want nil", err)
	}
	if elapsed := time.Since(start); elapsed < 30*time.Millisecond {
		t.Fatalf("queue wait = %v, want at least 30ms", elapsed)
	}
}
