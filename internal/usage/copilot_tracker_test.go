package usage

import (
	"testing"
	"time"

	copilotauth "github.com/router-for-me/CLIProxyAPI/v6/internal/auth/copilot"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func TestCopilotTrackerRecordAndQuotaSnapshot(t *testing.T) {
	tracker := NewCopilotTracker()
	auth := &coreauth.Auth{
		ID:       "copilot-auth-1",
		Provider: "github-copilot",
		Label:    "primary",
		Metadata: map[string]any{
			"email": "dev@example.com",
		},
	}

	firstRequestAt := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	secondRequestAt := firstRequestAt.Add(2 * time.Minute)

	tracker.RecordRequest(CopilotRequestRecord{
		Auth:         auth,
		Model:        "gpt-5.4",
		Endpoint:     "/responses",
		ResponseType: "streaming",
		RequestedAt:  firstRequestAt,
		Multiplier:   3,
	})
	tracker.RecordRequest(CopilotRequestRecord{
		Auth:         auth,
		Model:        "gpt-5.4",
		Endpoint:     "/responses",
		ResponseType: "streaming",
		RequestedAt:  secondRequestAt,
		Multiplier:   3,
		Failed:       true,
	})

	tracker.UpdateQuota(auth, &copilotauth.CopilotUsageResponse{
		CopilotPlan: "copilot-pro",
		QuotaSnapshots: copilotauth.QuotaSnapshots{
			PremiumInteractions: copilotauth.QuotaDetail{
				Entitlement:    300,
				Remaining:      297,
				QuotaID:        "premium",
				Unlimited:      false,
				QuotaRemaining: 297,
			},
		},
	})

	snapshot := tracker.Snapshot()
	account, ok := snapshot.Accounts[auth.EnsureIndex()]
	if !ok {
		t.Fatalf("missing account snapshot for auth index %q", auth.EnsureIndex())
	}
	if account.Requests != 2 {
		t.Fatalf("requests = %d, want 2", account.Requests)
	}
	if account.FailedRequests != 1 {
		t.Fatalf("failed requests = %d, want 1", account.FailedRequests)
	}
	if account.PremiumRequests != 6 {
		t.Fatalf("premium requests = %v, want 6", account.PremiumRequests)
	}
	if account.Account != "dev@example.com" {
		t.Fatalf("account = %q, want %q", account.Account, "dev@example.com")
	}
	if account.Quota == nil || account.Quota.CopilotPlan != "copilot-pro" {
		t.Fatalf("quota snapshot not stored")
	}

	model := account.Models["gpt-5.4"]
	if model.Requests != 2 {
		t.Fatalf("model requests = %d, want 2", model.Requests)
	}
	if model.FailedRequests != 1 {
		t.Fatalf("model failed requests = %d, want 1", model.FailedRequests)
	}
	if model.Endpoints["/responses"] != 2 {
		t.Fatalf("responses endpoint count = %d, want 2", model.Endpoints["/responses"])
	}
	if model.ResponseTypes["streaming"] != 2 {
		t.Fatalf("streaming response count = %d, want 2", model.ResponseTypes["streaming"])
	}
	if !model.LastRequestAt.Equal(secondRequestAt) {
		t.Fatalf("last request at = %s, want %s", model.LastRequestAt, secondRequestAt)
	}
}
