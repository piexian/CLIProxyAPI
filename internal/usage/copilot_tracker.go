package usage

import (
	"strings"
	"sync"
	"time"

	copilotauth "github.com/router-for-me/CLIProxyAPI/v6/internal/auth/copilot"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

// CopilotRequestRecord describes one GitHub Copilot request for local accounting.
type CopilotRequestRecord struct {
	Auth         *coreauth.Auth
	Model        string
	Endpoint     string
	ResponseType string
	RequestedAt  time.Time
	Failed       bool
	Multiplier   float64
}

// CopilotSnapshot is an immutable view of the in-memory GitHub Copilot usage tracker.
type CopilotSnapshot struct {
	Accounts map[string]CopilotAccountSnapshot `json:"accounts"`
}

// CopilotAccountSnapshot summarizes one Copilot auth entry.
type CopilotAccountSnapshot struct {
	AuthID          string                            `json:"auth_id,omitempty"`
	AuthIndex       string                            `json:"auth_index,omitempty"`
	Label           string                            `json:"label,omitempty"`
	AccountType     string                            `json:"account_type,omitempty"`
	Account         string                            `json:"account,omitempty"`
	Requests        int64                             `json:"requests"`
	FailedRequests  int64                             `json:"failed_requests"`
	PremiumRequests float64                           `json:"premium_requests"`
	LastRequestAt   time.Time                         `json:"last_request_at,omitempty"`
	Quota           *copilotauth.CopilotUsageResponse `json:"quota,omitempty"`
	QuotaFetchedAt  time.Time                         `json:"quota_fetched_at,omitempty"`
	Models          map[string]CopilotModelSnapshot   `json:"models,omitempty"`
}

// CopilotModelSnapshot summarizes request counts for a specific Copilot model.
type CopilotModelSnapshot struct {
	Requests        int64            `json:"requests"`
	FailedRequests  int64            `json:"failed_requests"`
	PremiumRequests float64          `json:"premium_requests"`
	LastRequestAt   time.Time        `json:"last_request_at,omitempty"`
	Endpoints       map[string]int64 `json:"endpoints,omitempty"`
	ResponseTypes   map[string]int64 `json:"response_types,omitempty"`
}

type copilotAccountState struct {
	AuthID          string
	AuthIndex       string
	Label           string
	AccountType     string
	Account         string
	Requests        int64
	FailedRequests  int64
	PremiumRequests float64
	LastRequestAt   time.Time
	Quota           *copilotauth.CopilotUsageResponse
	QuotaFetchedAt  time.Time
	Models          map[string]*copilotModelState
}

type copilotModelState struct {
	Requests        int64
	FailedRequests  int64
	PremiumRequests float64
	LastRequestAt   time.Time
	Endpoints       map[string]int64
	ResponseTypes   map[string]int64
}

// CopilotTracker maintains in-memory per-auth Copilot request and quota summaries.
type CopilotTracker struct {
	mu       sync.RWMutex
	accounts map[string]*copilotAccountState
}

var defaultCopilotTracker = NewCopilotTracker()

// GetCopilotTracker returns the shared Copilot usage tracker.
func GetCopilotTracker() *CopilotTracker { return defaultCopilotTracker }

// NewCopilotTracker constructs an empty Copilot usage tracker.
func NewCopilotTracker() *CopilotTracker {
	return &CopilotTracker{
		accounts: make(map[string]*copilotAccountState),
	}
}

// RecordRequest records one GitHub Copilot request in memory.
func (t *CopilotTracker) RecordRequest(record CopilotRequestRecord) {
	if t == nil {
		return
	}

	key := copilotTrackerKey(record.Auth)
	if key == "" {
		key = "unknown"
	}
	when := record.RequestedAt
	if when.IsZero() {
		when = time.Now()
	}
	modelKey := strings.TrimSpace(record.Model)
	if modelKey == "" {
		modelKey = "unknown"
	}
	endpoint := strings.TrimSpace(record.Endpoint)
	if endpoint == "" {
		endpoint = "unknown"
	}
	responseType := strings.TrimSpace(record.ResponseType)
	if responseType == "" {
		responseType = "unknown"
	}
	multiplier := record.Multiplier
	if multiplier < 0 {
		multiplier = 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	account := t.ensureAccountLocked(key, record.Auth)
	account.Requests++
	if record.Failed {
		account.FailedRequests++
	}
	account.PremiumRequests += multiplier
	if when.After(account.LastRequestAt) {
		account.LastRequestAt = when
	}

	modelState := account.Models[modelKey]
	if modelState == nil {
		modelState = &copilotModelState{
			Endpoints:     make(map[string]int64),
			ResponseTypes: make(map[string]int64),
		}
		account.Models[modelKey] = modelState
	}
	modelState.Requests++
	if record.Failed {
		modelState.FailedRequests++
	}
	modelState.PremiumRequests += multiplier
	if when.After(modelState.LastRequestAt) {
		modelState.LastRequestAt = when
	}
	modelState.Endpoints[endpoint]++
	modelState.ResponseTypes[responseType]++
}

// UpdateQuota stores the latest Copilot quota snapshot for an auth entry.
func (t *CopilotTracker) UpdateQuota(auth *coreauth.Auth, quota *copilotauth.CopilotUsageResponse) {
	if t == nil || quota == nil {
		return
	}

	key := copilotTrackerKey(auth)
	if key == "" {
		key = "unknown"
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	account := t.ensureAccountLocked(key, auth)
	account.Quota = quota.Clone()
	account.QuotaFetchedAt = time.Now().UTC()
}

// Snapshot returns a deep copy of the tracker state.
func (t *CopilotTracker) Snapshot() CopilotSnapshot {
	result := CopilotSnapshot{
		Accounts: make(map[string]CopilotAccountSnapshot),
	}
	if t == nil {
		return result
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	result.Accounts = make(map[string]CopilotAccountSnapshot, len(t.accounts))
	for key, account := range t.accounts {
		if account == nil {
			continue
		}
		accountSnapshot := CopilotAccountSnapshot{
			AuthID:          account.AuthID,
			AuthIndex:       account.AuthIndex,
			Label:           account.Label,
			AccountType:     account.AccountType,
			Account:         account.Account,
			Requests:        account.Requests,
			FailedRequests:  account.FailedRequests,
			PremiumRequests: account.PremiumRequests,
			LastRequestAt:   account.LastRequestAt,
			QuotaFetchedAt:  account.QuotaFetchedAt,
			Models:          make(map[string]CopilotModelSnapshot, len(account.Models)),
		}
		if account.Quota != nil {
			accountSnapshot.Quota = account.Quota.Clone()
		}
		for model, modelState := range account.Models {
			if modelState == nil {
				continue
			}
			modelSnapshot := CopilotModelSnapshot{
				Requests:        modelState.Requests,
				FailedRequests:  modelState.FailedRequests,
				PremiumRequests: modelState.PremiumRequests,
				LastRequestAt:   modelState.LastRequestAt,
			}
			if len(modelState.Endpoints) > 0 {
				modelSnapshot.Endpoints = make(map[string]int64, len(modelState.Endpoints))
				for endpoint, count := range modelState.Endpoints {
					modelSnapshot.Endpoints[endpoint] = count
				}
			}
			if len(modelState.ResponseTypes) > 0 {
				modelSnapshot.ResponseTypes = make(map[string]int64, len(modelState.ResponseTypes))
				for responseType, count := range modelState.ResponseTypes {
					modelSnapshot.ResponseTypes[responseType] = count
				}
			}
			accountSnapshot.Models[model] = modelSnapshot
		}
		result.Accounts[key] = accountSnapshot
	}

	return result
}

func (t *CopilotTracker) ensureAccountLocked(key string, auth *coreauth.Auth) *copilotAccountState {
	account := t.accounts[key]
	if account == nil {
		account = &copilotAccountState{
			Models: make(map[string]*copilotModelState),
		}
		t.accounts[key] = account
	}
	populateCopilotAccountIdentity(account, auth)
	return account
}

func populateCopilotAccountIdentity(account *copilotAccountState, auth *coreauth.Auth) {
	if account == nil || auth == nil {
		return
	}
	if authID := strings.TrimSpace(auth.ID); authID != "" {
		account.AuthID = authID
	}
	if authIndex := auth.EnsureIndex(); authIndex != "" {
		account.AuthIndex = authIndex
	}
	if label := strings.TrimSpace(auth.Label); label != "" {
		account.Label = label
	}
	if kind, value := auth.AccountInfo(); value != "" {
		account.AccountType = strings.TrimSpace(kind)
		account.Account = strings.TrimSpace(value)
	}
}

func copilotTrackerKey(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	if index := auth.EnsureIndex(); index != "" {
		return index
	}
	if authID := strings.TrimSpace(auth.ID); authID != "" {
		return "id:" + authID
	}
	return ""
}
