package executor

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/thinking"
)

func TestQwenExecutorParseSuffix(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantBase  string
		wantLevel string
	}{
		{"no suffix", "qwen-max", "qwen-max", ""},
		{"with level suffix", "qwen-max(high)", "qwen-max", "high"},
		{"with budget suffix", "qwen-max(16384)", "qwen-max", "16384"},
		{"complex model name", "qwen-plus-latest(medium)", "qwen-plus-latest", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thinking.ParseSuffix(tt.model)
			if result.ModelName != tt.wantBase {
				t.Errorf("ParseSuffix(%q).ModelName = %q, want %q", tt.model, result.ModelName, tt.wantBase)
			}
		})
	}
}

func TestIsQwenQuotaError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			"insufficient_quota code",
			`{"error":{"code":"insufficient_quota","message":"You exceeded your current quota","type":"insufficient_quota"}}`,
			true,
		},
		{
			"quota_exceeded code",
			`{"error":{"code":"quota_exceeded","message":"Quota exceeded","type":"quota_exceeded"}}`,
			true,
		},
		{
			"allocated quota exceeded in message",
			`{"error":{"code":"some_code","message":"Allocated quota exceeded, please check your plan","type":"some_type"}}`,
			true,
		},
		{
			"insufficient_quota in message only",
			`{"error":{"code":"unknown","message":"insufficient_quota reached","type":"unknown"}}`,
			true,
		},
		{
			"rate limit error should NOT match quota",
			`{"error":{"code":"rate_limit_exceeded","message":"Requests rate limit exceeded","type":"rate_limit_exceeded"}}`,
			false,
		},
		{
			"empty body",
			`{}`,
			false,
		},
		{
			"unrelated error",
			`{"error":{"code":"invalid_request","message":"Bad request","type":"invalid_request"}}`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isQwenQuotaError([]byte(tt.body))
			if got != tt.want {
				t.Errorf("isQwenQuotaError(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestIsQwenDailyQuotaError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			"generic insufficient_quota is not daily",
			`{"error":{"code":"insufficient_quota","message":"You exceeded your current quota","type":"insufficient_quota"}}`,
			false,
		},
		{
			"daily quota message matches",
			`{"error":{"code":"insufficient_quota","message":"Daily quota exceeded, please try again tomorrow","type":"insufficient_quota"}}`,
			true,
		},
		{
			"daily limit wording matches",
			`{"error":{"code":"quota_exceeded","message":"Daily limit reached for today","type":"quota_exceeded"}}`,
			true,
		},
		{
			"rate limit error stays false",
			`{"error":{"code":"rate_limit_exceeded","message":"Requests rate limit exceeded","type":"rate_limit_exceeded"}}`,
			false,
		},
		{
			"empty body",
			`{}`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isQwenDailyQuotaError([]byte(tt.body))
			if got != tt.want {
				t.Errorf("isQwenDailyQuotaError(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestIsQwenRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			"requests rate limit exceeded",
			`{"error":{"code":"rate_limit_exceeded","message":"Requests rate limit exceeded, please try again later","type":"rate_limit_exceeded"}}`,
			true,
		},
		{
			"you exceeded your current requests",
			`{"error":{"code":"rate_limit","message":"You exceeded your current requests list","type":"rate_limit"}}`,
			true,
		},
		{
			"request rate increased too quickly",
			`{"error":{"code":"stability_protection","message":"Request rate increased too quickly, please slow down","type":"rate_limit"}}`,
			true,
		},
		{
			"quota error should NOT match rate limit",
			`{"error":{"code":"insufficient_quota","message":"You exceeded your current quota","type":"insufficient_quota"}}`,
			false,
		},
		{
			"empty body",
			`{}`,
			false,
		},
		{
			"unrelated error",
			`{"error":{"code":"invalid_request","message":"Bad request","type":"invalid_request"}}`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isQwenRateLimitError([]byte(tt.body))
			if got != tt.want {
				t.Errorf("isQwenRateLimitError(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestWrapQwenErrorDifferentiation(t *testing.T) {
	ctx := context.Background()

	t.Run("RPM rate limit gets short cooldown", func(t *testing.T) {
		body := []byte(`{"error":{"code":"rate_limit_exceeded","message":"Requests rate limit exceeded","type":"rate_limit_exceeded"}}`)
		errCode, retryAfter := wrapQwenError(ctx, http.StatusTooManyRequests, body)

		if errCode != http.StatusTooManyRequests {
			t.Errorf("errCode = %d, want %d", errCode, http.StatusTooManyRequests)
		}
		if retryAfter == nil {
			t.Fatal("retryAfter is nil, want non-nil")
		}
		if *retryAfter != qwenRateCooldown {
			t.Errorf("retryAfter = %v, want %v", *retryAfter, qwenRateCooldown)
		}
	})

	t.Run("burst rate error gets short cooldown", func(t *testing.T) {
		body := []byte(`{"error":{"code":"stability","message":"Request rate increased too quickly","type":"rate_limit"}}`)
		errCode, retryAfter := wrapQwenError(ctx, http.StatusTooManyRequests, body)

		if errCode != http.StatusTooManyRequests {
			t.Errorf("errCode = %d, want %d", errCode, http.StatusTooManyRequests)
		}
		if retryAfter == nil {
			t.Fatal("retryAfter is nil, want non-nil")
		}
		if *retryAfter != qwenRateCooldown {
			t.Errorf("retryAfter = %v, want %v", *retryAfter, qwenRateCooldown)
		}
	})

	t.Run("generic quota signal gets short cooldown", func(t *testing.T) {
		body := []byte(`{"error":{"code":"insufficient_quota","message":"You exceeded your current quota","type":"insufficient_quota"}}`)
		errCode, retryAfter := wrapQwenError(ctx, http.StatusForbidden, body)

		if errCode != http.StatusTooManyRequests {
			t.Errorf("errCode = %d, want %d", errCode, http.StatusTooManyRequests)
		}
		if retryAfter == nil {
			t.Fatal("retryAfter is nil, want non-nil")
		}
		if *retryAfter != qwenRateCooldown {
			t.Errorf("retryAfter = %v, want %v", *retryAfter, qwenRateCooldown)
		}
	})

	t.Run("daily quota exhaustion gets next-day cooldown", func(t *testing.T) {
		body := []byte(`{"error":{"code":"insufficient_quota","message":"Daily quota exceeded, please try again tomorrow","type":"insufficient_quota"}}`)
		errCode, retryAfter := wrapQwenError(ctx, http.StatusForbidden, body)

		if errCode != http.StatusTooManyRequests {
			t.Errorf("errCode = %d, want %d", errCode, http.StatusTooManyRequests)
		}
		if retryAfter == nil {
			t.Fatal("retryAfter is nil, want non-nil")
		}
		if *retryAfter <= qwenRateCooldown {
			t.Errorf("retryAfter = %v, expected to be significantly longer than %v for daily quota", *retryAfter, qwenRateCooldown)
		}
	})

	t.Run("unknown 429 gets short cooldown", func(t *testing.T) {
		body := []byte(`{"error":{"code":"unknown","message":"some unknown error","type":"unknown"}}`)
		errCode, retryAfter := wrapQwenError(ctx, http.StatusTooManyRequests, body)

		if errCode != http.StatusTooManyRequests {
			t.Errorf("errCode = %d, want %d", errCode, http.StatusTooManyRequests)
		}
		if retryAfter == nil {
			t.Fatal("retryAfter is nil, want non-nil")
		}
		if *retryAfter != qwenRateCooldown {
			t.Errorf("retryAfter = %v, want %v", *retryAfter, qwenRateCooldown)
		}
	})

	t.Run("non-429 non-403 is not modified", func(t *testing.T) {
		body := []byte(`{"error":{"code":"bad_request","message":"Invalid input","type":"invalid_request"}}`)
		errCode, retryAfter := wrapQwenError(ctx, http.StatusBadRequest, body)

		if errCode != http.StatusBadRequest {
			t.Errorf("errCode = %d, want %d", errCode, http.StatusBadRequest)
		}
		if retryAfter != nil {
			t.Errorf("retryAfter = %v, want nil", *retryAfter)
		}
	})
}

func TestCheckQwenDailyLimit(t *testing.T) {
	// Clean up counters before test
	qwenDailyCounter.Lock()
	qwenDailyCounter.counts = make(map[string]*qwenDailyCount)
	qwenDailyCounter.Unlock()

	authID := "test-daily-limit-auth"

	t.Run("allows requests under limit", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			if err := checkQwenDailyLimit(authID); err != nil {
				t.Fatalf("request %d: unexpected error: %v", i, err)
			}
			// Simulate successful request by recording it
			recordQwenDailyRequest(authID)
		}
	})

	t.Run("check only does not increment counter", func(t *testing.T) {
		// Reset counters
		qwenDailyCounter.Lock()
		qwenDailyCounter.counts = make(map[string]*qwenDailyCount)
		qwenDailyCounter.Unlock()

		testID := "test-check-only"
		// Call check many times without recording
		for i := 0; i < 2000; i++ {
			if err := checkQwenDailyLimit(testID); err != nil {
				t.Fatalf("check-only request %d: unexpected error: %v (check should not increment)", i, err)
			}
		}
		// Counter should still be 0
		qwenDailyCounter.Lock()
		dc := qwenDailyCounter.counts[testID]
		qwenDailyCounter.Unlock()
		if dc != nil && dc.count != 0 {
			t.Errorf("counter = %d after check-only calls, want 0", dc.count)
		}
	})

	t.Run("empty authID is always allowed", func(t *testing.T) {
		if err := checkQwenDailyLimit(""); err != nil {
			t.Fatalf("unexpected error for empty authID: %v", err)
		}
	})

	t.Run("blocks at daily limit", func(t *testing.T) {
		markQwenDailyExhausted(authID)
		err := checkQwenDailyLimit(authID)
		if err == nil {
			t.Fatal("expected error at daily limit, got nil")
		}
		se, ok := err.(statusErr)
		if !ok {
			t.Fatalf("expected statusErr, got %T", err)
		}
		if se.code != http.StatusTooManyRequests {
			t.Errorf("code = %d, want %d", se.code, http.StatusTooManyRequests)
		}
		if se.retryAfter == nil {
			t.Fatal("retryAfter is nil")
		}
	})

	t.Run("resets on new day", func(t *testing.T) {
		qwenDailyCounter.Lock()
		// Simulate yesterday's exhausted counter
		qwenDailyCounter.counts[authID] = &qwenDailyCount{
			date:  time.Now().In(qwenBeijingLoc).Add(-24 * time.Hour).Format("2006-01-02"),
			count: qwenDailyLimit,
		}
		qwenDailyCounter.Unlock()

		// Should be allowed since the date has changed
		if err := checkQwenDailyLimit(authID); err != nil {
			t.Fatalf("expected reset on new day, got error: %v", err)
		}
	})
}

func TestRecordQwenDailyRequest(t *testing.T) {
	qwenDailyCounter.Lock()
	qwenDailyCounter.counts = make(map[string]*qwenDailyCount)
	qwenDailyCounter.Unlock()

	authID := "test-record-daily"

	// Record some requests
	for i := 0; i < 5; i++ {
		recordQwenDailyRequest(authID)
	}

	// Verify counter
	qwenDailyCounter.Lock()
	dc := qwenDailyCounter.counts[authID]
	qwenDailyCounter.Unlock()
	if dc == nil || dc.count != 5 {
		t.Fatalf("count = %v, want 5", dc)
	}

	// Empty authID should be a no-op
	recordQwenDailyRequest("")

	// Fill to limit via recording
	qwenDailyCounter.Lock()
	qwenDailyCounter.counts[authID].count = qwenDailyLimit
	qwenDailyCounter.Unlock()

	// Check should now block
	err := checkQwenDailyLimit(authID)
	if err == nil {
		t.Fatal("expected error at daily limit, got nil")
	}
}

func TestMarkQwenDailyExhausted(t *testing.T) {
	qwenDailyCounter.Lock()
	qwenDailyCounter.counts = make(map[string]*qwenDailyCount)
	qwenDailyCounter.Unlock()

	authID := "test-mark-exhausted"

	// First request should succeed
	if err := checkQwenDailyLimit(authID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mark exhausted
	markQwenDailyExhausted(authID)

	// Subsequent request should fail
	err := checkQwenDailyLimit(authID)
	if err == nil {
		t.Fatal("expected error after markQwenDailyExhausted, got nil")
	}

	// Empty authID should be a no-op
	markQwenDailyExhausted("")
}
