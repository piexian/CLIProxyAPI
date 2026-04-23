package executor

import (
	"context"
	"net/http"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/runtime/executor/helps"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/usage"
)

func newProxyAwareHTTPClient(ctx context.Context, cfg *config.Config, auth *cliproxyauth.Auth, timeout time.Duration) *http.Client {
	return helps.NewProxyAwareHTTPClient(ctx, cfg, auth, timeout)
}

func parseOpenAIUsage(data []byte) usage.Detail {
	return helps.ParseOpenAIUsage(data)
}

func parseOpenAIStreamUsage(line []byte) (usage.Detail, bool) {
	return helps.ParseOpenAIStreamUsage(line)
}

func parseOpenAIResponsesUsage(data []byte) usage.Detail {
	return helps.ParseOpenAIUsage(data)
}

func parseOpenAIResponsesStreamUsage(line []byte) (usage.Detail, bool) {
	return helps.ParseOpenAIStreamUsage(line)
}

type upstreamRequestLog = helps.UpstreamRequestLog

func recordAPIRequest(ctx context.Context, cfg *config.Config, info upstreamRequestLog) {
	helps.RecordAPIRequest(ctx, cfg, info)
}

func recordAPIResponseMetadata(ctx context.Context, cfg *config.Config, status int, headers http.Header) {
	helps.RecordAPIResponseMetadata(ctx, cfg, status, headers)
}

func recordAPIResponseError(ctx context.Context, cfg *config.Config, err error) {
	helps.RecordAPIResponseError(ctx, cfg, err)
}

func appendAPIResponseChunk(ctx context.Context, cfg *config.Config, chunk []byte) {
	helps.AppendAPIResponseChunk(ctx, cfg, chunk)
}

func payloadRequestedModel(opts cliproxyexecutor.Options, fallback string) string {
	return helps.PayloadRequestedModel(opts, fallback)
}

func applyPayloadConfigWithRoot(cfg *config.Config, model, protocol, root string, payload, original []byte, requestedModel string) []byte {
	return helps.ApplyPayloadConfigWithRoot(cfg, model, protocol, root, payload, original, requestedModel)
}

func summarizeErrorBody(contentType string, body []byte) string {
	return helps.SummarizeErrorBody(contentType, body)
}

type usageReporter struct {
	reporter *helps.UsageReporter
}

func newUsageReporter(ctx context.Context, provider, model string, auth *cliproxyauth.Auth) *usageReporter {
	return &usageReporter{reporter: helps.NewUsageReporter(ctx, provider, model, auth)}
}

func (r *usageReporter) publish(ctx context.Context, detail usage.Detail) {
	if r == nil || r.reporter == nil {
		return
	}
	r.reporter.Publish(ctx, detail)
}

func (r *usageReporter) publishFailure(ctx context.Context) {
	if r == nil || r.reporter == nil {
		return
	}
	r.reporter.PublishFailure(ctx)
}

func (r *usageReporter) trackFailure(ctx context.Context, errPtr *error) {
	if r == nil || r.reporter == nil {
		return
	}
	r.reporter.TrackFailure(ctx, errPtr)
}

func (r *usageReporter) ensurePublished(ctx context.Context) {
	if r == nil || r.reporter == nil {
		return
	}
	r.reporter.EnsurePublished(ctx)
}
