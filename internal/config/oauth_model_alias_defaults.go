package config

import "strings"

// defaultGitHubCopilotAliases returns default oauth-model-alias entries for
// GitHub Copilot Claude models so hyphen-style client ids also work.
func defaultGitHubCopilotAliases() []OAuthModelAlias {
	return GitHubCopilotAliasesFromModels([]string{
		"claude-haiku-4.5",
		"claude-opus-4.1",
		"claude-opus-4.5",
		"claude-opus-4.6",
		"claude-opus-4.6-fast",
		"claude-sonnet-4.5",
		"claude-sonnet-4.6",
		"gemini-2.5-pro",
		"gemini-3.1-pro-preview",
		"gpt-5.1",
		"gpt-5.1-codex",
		"gpt-5.1-codex-mini",
		"gpt-5.1-codex-max",
		"gpt-5.2",
		"gpt-5.2-codex",
		"gpt-5.3-codex",
		"gpt-5.4",
		"gpt-5.4-mini",
	})
}

// DefaultGitHubCopilotAliases returns the built-in GitHub Copilot alias set used
// for request-time model resolution.
func DefaultGitHubCopilotAliases() []OAuthModelAlias {
	return defaultGitHubCopilotAliases()
}

// GitHubCopilotAliasesFromModels generates aliases for Copilot model ids that contain dots.
func GitHubCopilotAliasesFromModels(modelIDs []string) []OAuthModelAlias {
	var aliases []OAuthModelAlias
	seen := make(map[string]struct{})
	for _, id := range modelIDs {
		if !strings.Contains(id, ".") {
			continue
		}
		alias := strings.ReplaceAll(id, ".", "-")
		key := id + "->" + alias
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		aliases = append(aliases, OAuthModelAlias{Name: id, Alias: alias, Fork: true})
	}
	return aliases
}
