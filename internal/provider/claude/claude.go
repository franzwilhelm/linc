package claude

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"linc/internal/linear"
	"linc/internal/provider"
)

// Provider implements the Claude Code agent provider
type Provider struct{}

// New creates a new Claude provider
func New() *Provider {
	return &Provider{}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "Claude Code"
}

// Exec launches Claude Code with the issue context
// This replaces the current process with Claude
func (p *Provider) Exec(issue linear.Issue, comment string, ctx *linear.IssueContext, planMode bool) error {
	prompt := provider.BuildPrompt(issue, comment, ctx)

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	args := []string{"claude", prompt}
	if planMode {
		args = append(args, "--permission-mode", "plan")
	}

	return syscall.Exec(claudePath, args, os.Environ())
}
