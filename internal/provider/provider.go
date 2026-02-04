package provider

import (
	"linc/internal/linear"
)

// Provider is the interface for code agent providers (Claude, opencode, echo, etc.)
type Provider interface {
	// Name returns the provider's display name
	Name() string

	// Exec launches the provider with the given issue context
	// This may replace the current process (syscall.Exec) or return after completion
	Exec(issue linear.Issue, comment string, ctx *linear.IssueContext, planMode bool) error
}

// BuildPrompt creates a standardized prompt from issue data
// This can be used by any provider that needs a text prompt
func BuildPrompt(issue linear.Issue, comment string, ctx *linear.IssueContext) string {
	return buildPromptInternal(issue, comment, ctx)
}
