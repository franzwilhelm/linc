package echo

import (
	"fmt"

	"linc/internal/linear"
	"linc/internal/provider"
)

// Provider implements a simple echo provider for testing
// It just prints the prompt to stdout and returns
type Provider struct{}

// New creates a new Echo provider
func New() *Provider {
	return &Provider{}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "Echo (test)"
}

// Exec prints the prompt to stdout
func (p *Provider) Exec(issue linear.Issue, comment string, ctx *linear.IssueContext, planMode bool) error {
	prompt := provider.BuildPrompt(issue, comment, ctx)

	fmt.Println("=== Echo Provider Output ===")
	fmt.Printf("Plan Mode: %v\n", planMode)
	fmt.Println("=== Prompt Start ===")
	fmt.Println(prompt)
	fmt.Println("=== Prompt End ===")

	return nil
}
