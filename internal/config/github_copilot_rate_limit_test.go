package config

import "testing"

func TestSanitizeGitHubCopilotRateLimit_DefaultsQueueToEnabled(t *testing.T) {
	var cfg Config

	cfg.SanitizeGitHubCopilotRateLimit()

	if cfg.GitHubCopilotRateLimitSeconds != 0 {
		t.Fatalf("GitHubCopilotRateLimitSeconds = %d, want 0", cfg.GitHubCopilotRateLimitSeconds)
	}
	if cfg.GitHubCopilotRateLimitWait == nil {
		t.Fatalf("GitHubCopilotRateLimitWait = nil, want non-nil")
	}
	if !*cfg.GitHubCopilotRateLimitWait {
		t.Fatalf("GitHubCopilotRateLimitWait = false, want true")
	}
}
