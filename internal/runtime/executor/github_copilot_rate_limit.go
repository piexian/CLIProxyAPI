package executor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

type githubCopilotRateLimiter struct {
	mu            sync.Mutex
	nextAllowedAt time.Time
}

var globalGitHubCopilotRateLimiter = &githubCopilotRateLimiter{}

func (l *githubCopilotRateLimiter) acquire(ctx context.Context, interval time.Duration, wait bool) error {
	if interval <= 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now()

	l.mu.Lock()
	if l.nextAllowedAt.After(now) {
		waitDuration := l.nextAllowedAt.Sub(now)
		if !wait {
			l.mu.Unlock()
			return statusErr{
				code: http.StatusTooManyRequests,
				msg:  githubCopilotRateLimitError(waitDuration),
			}
		}
		scheduledAt := l.nextAllowedAt
		l.nextAllowedAt = scheduledAt.Add(interval)
		l.mu.Unlock()
		return waitForGitHubCopilotSlot(ctx, scheduledAt.Sub(now))
	}
	l.nextAllowedAt = now.Add(interval)
	l.mu.Unlock()
	return nil
}

func githubCopilotRateLimitConfig(cfg *config.Config) (time.Duration, bool) {
	waitEnabled := true
	if cfg != nil && cfg.GitHubCopilotRateLimitWait != nil {
		waitEnabled = *cfg.GitHubCopilotRateLimitWait
	}
	if cfg == nil || cfg.GitHubCopilotRateLimitSeconds <= 0 {
		return 0, waitEnabled
	}
	return time.Duration(cfg.GitHubCopilotRateLimitSeconds) * time.Second, waitEnabled
}

func githubCopilotRateLimitError(waitDuration time.Duration) string {
	waitSeconds := int((waitDuration + time.Second - 1) / time.Second)
	if waitSeconds < 1 {
		waitSeconds = 1
	}
	return fmt.Sprintf(
		`{"error":{"code":"rate_limit_exceeded","message":"GitHub Copilot local rate limit reached, retry after %ds","type":"rate_limit_error"}}`,
		waitSeconds,
	)
}

func acquireGitHubCopilotRateLimit(ctx context.Context, cfg *config.Config) error {
	interval, waitEnabled := githubCopilotRateLimitConfig(cfg)
	return globalGitHubCopilotRateLimiter.acquire(ctx, interval, waitEnabled)
}

func githubCopilotQueueEnabled(cfg *config.Config) bool {
	_, waitEnabled := githubCopilotRateLimitConfig(cfg)
	return waitEnabled
}

func waitForGitHubCopilotSlot(ctx context.Context, waitDuration time.Duration) error {
	if waitDuration <= 0 {
		return nil
	}
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
