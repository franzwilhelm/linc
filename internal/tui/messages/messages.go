package messages

import (
	"linc/internal/linear"
)

// View switching messages
type SwitchToListMsg struct{}
type SwitchToDetailMsg struct {
	Issue linear.Issue
}
type SwitchToStartWorkMsg struct {
	Issue linear.Issue
}
type SwitchToTeamSelectMsg struct{}
type SwitchToSettingsMsg struct{}
type SwitchToWorkspaceSelectMsg struct{}
type NextIssueMsg struct{}
type PrevIssueMsg struct{}

// Data loading messages
type TeamsLoadedMsg struct {
	Teams []linear.Team
	Err   error
}

type IssuesLoadedMsg struct {
	Issues []linear.Issue
	Err    error
}

type AllIssuesLoadedMsg struct {
	Issues []linear.Issue
	Err    error
}

type StatesLoadedMsg struct {
	States []linear.State
	Err    error
}

type ViewerLoadedMsg struct {
	Viewer *linear.ViewerResponse
	Err    error
}

// Action messages
type TeamSelectedMsg struct {
	Team         linear.Team
	SetAsDefault bool
}

type CommentCreatedMsg struct {
	Comment *linear.Comment
	Err     error
}

type StateUpdatedMsg struct {
	Err error
}

type IssueTitleUpdatedMsg struct {
	IssueID   string
	NewTitle  string
	Err       error
	Completed bool // true when API call is done
}

type IssuePriorityUpdatedMsg struct {
	IssueID     string
	NewPriority int
	Err         error
	Completed   bool
}

type IssueStateUpdatedMsg struct {
	IssueID    string
	NewStateID string
	Err        error
	Completed  bool
}

type CancelIssueMsg struct {
	IssueID string
	TeamID  string
}

type MarkDuplicateMsg struct {
	IssueID string
	TeamID  string
}

type OpenBrowserMsg struct {
	URL string
}

type StartClaudeMsg struct {
	Issue        linear.Issue
	Comment      string
	UseBranch    bool
	PlanMode     bool
	CheckoutOnly bool
}

type ErrorMsg struct {
	Err error
}

type SettingsSavedMsg struct {
	Err error
}

type QuitMsg struct{}
