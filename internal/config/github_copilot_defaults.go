package config

import "strings"

func defaultGitHubCopilotPremiumModelMultipliers() map[string]float64 {
	return map[string]float64{
		"claude-haiku-4.5":       0.33,
		"claude-sonnet-4":        1,
		"claude-sonnet-4.5":      1,
		"claude-sonnet-4.6":      3,
		"claude-opus-4.5":        3,
		"claude-opus-4.6":        3,
		"claude-opus-4.6-fast":   3,
		"gemini-2.5-pro":         1,
		"gemini-3-flash-preview": 0.33,
		"gemini-3.1-pro-preview": 1,
		"grok-code-fast-1":       0.25,
		"gpt-5.1":                1,
		"gpt-5.2":                1,
		"gpt-5.2-codex":          1,
		"gpt-5.3-codex":          1,
		"gpt-5.4-mini":           0.33,
		"gpt-5.4":                3,
	}
}

func normalizeGitHubCopilotModelKey(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	if model == "" {
		return ""
	}
	if open := strings.LastIndex(model, "("); open >= 0 && strings.HasSuffix(model, ")") {
		model = strings.TrimSpace(model[:open])
	}
	return model
}

// GitHubCopilotPremiumModelMultiplier returns the local premium-usage multiplier for a Copilot model.
// When the model is unknown, the multiplier defaults to 0.
func GitHubCopilotPremiumModelMultiplier(cfg *Config, model string) float64 {
	key := normalizeGitHubCopilotModelKey(model)
	if key == "" {
		return 0
	}

	if cfg != nil && len(cfg.GitHubCopilotPremiumModelMultipliers) > 0 {
		if multiplier, ok := cfg.GitHubCopilotPremiumModelMultipliers[key]; ok && multiplier >= 0 {
			return multiplier
		}
	}

	return defaultGitHubCopilotPremiumModelMultipliers()[key]
}
