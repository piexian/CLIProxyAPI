package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// QuotaDetail represents one quota bucket from GitHub Copilot usage.
type QuotaDetail struct {
	Entitlement      float64 `json:"entitlement"`
	OverageCount     float64 `json:"overage_count"`
	OveragePermitted bool    `json:"overage_permitted"`
	PercentRemaining float64 `json:"percent_remaining"`
	QuotaID          string  `json:"quota_id"`
	QuotaRemaining   float64 `json:"quota_remaining"`
	Remaining        float64 `json:"remaining"`
	Unlimited        bool    `json:"unlimited"`
}

// QuotaSnapshots contains current Copilot quota state across usage buckets.
type QuotaSnapshots struct {
	Chat                QuotaDetail `json:"chat"`
	Completions         QuotaDetail `json:"completions"`
	PremiumInteractions QuotaDetail `json:"premium_interactions"`
}

// CopilotUsageResponse mirrors GET /copilot_internal/user.
type CopilotUsageResponse struct {
	AccessTypeSKU         string         `json:"access_type_sku"`
	AnalyticsTrackingID   string         `json:"analytics_tracking_id"`
	AssignedDate          string         `json:"assigned_date"`
	CanSignupForLimited   bool           `json:"can_signup_for_limited"`
	ChatEnabled           bool           `json:"chat_enabled"`
	CopilotPlan           string         `json:"copilot_plan"`
	OrganizationLoginList []interface{}  `json:"organization_login_list"`
	OrganizationList      []interface{}  `json:"organization_list"`
	QuotaResetDate        string         `json:"quota_reset_date"`
	QuotaSnapshots        QuotaSnapshots `json:"quota_snapshots"`
}

// Clone returns a detached copy of the usage response for safe in-memory caching.
func (u *CopilotUsageResponse) Clone() *CopilotUsageResponse {
	if u == nil {
		return nil
	}
	clone := *u
	if len(u.OrganizationLoginList) > 0 {
		clone.OrganizationLoginList = append([]interface{}(nil), u.OrganizationLoginList...)
	}
	if len(u.OrganizationList) > 0 {
		clone.OrganizationList = append([]interface{}(nil), u.OrganizationList...)
	}
	return &clone
}

// PremiumUsed returns the used premium-interaction quota when the quota is finite.
func (u *CopilotUsageResponse) PremiumUsed() float64 {
	if u == nil {
		return 0
	}
	premium := u.QuotaSnapshots.PremiumInteractions
	if premium.Unlimited {
		return 0
	}
	return premium.Entitlement - premium.Remaining
}

// GetUsageWithGitHubToken fetches GitHub Copilot usage information using the GitHub access token.
func (c *CopilotAuth) GetUsageWithGitHubToken(ctx context.Context, githubAccessToken string) (*CopilotUsageResponse, error) {
	if githubAccessToken == "" {
		return nil, fmt.Errorf("copilot: github access token is required for usage lookup")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/copilot_internal/user", nil)
	if err != nil {
		return nil, fmt.Errorf("copilot: failed to create usage request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+githubAccessToken)
	req.Header.Set("User-Agent", copilotUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("copilot: usage request failed: %w", err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("copilot usage: close body error: %v", errClose)
		}
	}()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxModelsResponseSize))
	if err != nil {
		return nil, fmt.Errorf("copilot: failed to read usage response: %w", err)
	}
	if !isHTTPSuccess(resp.StatusCode) {
		return nil, fmt.Errorf("copilot: usage request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var usage CopilotUsageResponse
	if err = json.Unmarshal(bodyBytes, &usage); err != nil {
		return nil, fmt.Errorf("copilot: failed to parse usage response: %w", err)
	}
	return &usage, nil
}
