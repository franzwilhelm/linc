package tui

import (
	"os"
	"os/exec"
	"runtime"

	"linc/internal/config"
	"linc/internal/git"
	"linc/internal/linear"
	"linc/internal/tui/messages"
	"linc/internal/tui/styles"
	"linc/internal/tui/views"

	tea "github.com/charmbracelet/bubbletea"
)

var Version = "dev" // Set from main via SetVersion

type View int

const (
	ViewWorkspaceSelect View = iota
	ViewTeamSelect
	ViewList
	ViewDetail
	ViewStartWork
	ViewSettings
)

type RootModel struct {
	client          *linear.Client
	cfg             *config.Config
	workspace       *config.Workspace
	workspaces      []config.Workspace
	currentDir      string
	providers       []string
	currentView     View
	workspaceSelect views.WorkspaceSelectModel
	teamSelect      views.TeamSelectModel
	list            views.ListModel
	detail          views.DetailModel
	startWork       views.StartWorkModel
	settings        views.SettingsModel
	teams           []linear.Team
	selectedTeam    *linear.Team
	err             error
	quitting        bool
	startClaude     *messages.StartClaudeMsg
	addNewWorkspace bool
}

func NewRootModel(client *linear.Client, cfg *config.Config, workspace *config.Workspace, workspaces []config.Workspace, currentDir string, providers []string) RootModel {
	list := views.NewListModel()
	if branch := git.GetCurrentBranch(); branch != "" {
		list = list.SetCurrentBranch(branch)
	}
	if wd, err := os.Getwd(); err == nil {
		list = list.SetWorkingDir(wd)
	}
	list = list.SetVersion(Version)
	return RootModel{
		client:          client,
		cfg:             cfg,
		workspace:       workspace,
		workspaces:      workspaces,
		currentDir:      currentDir,
		providers:       providers,
		workspaceSelect: views.NewIntegratedWorkspaceSelectModel(workspaces),
		teamSelect:      views.NewTeamSelectModel(),
		list:            list,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.loadViewer
}

func (m RootModel) loadViewer() tea.Msg {
	viewer, err := m.client.GetViewer()
	if err != nil {
		return messages.ErrorMsg{Err: err}
	}
	return messages.ViewerLoadedMsg{Viewer: viewer}
}

func (m RootModel) loadStates(teamID string) tea.Cmd {
	return func() tea.Msg {
		states, err := m.client.GetTeamStates(teamID)
		return messages.StatesLoadedMsg{States: states, Err: err}
	}
}

func (m RootModel) loadIssues(teamID string) tea.Cmd {
	return func() tea.Msg {
		issues, err := m.client.GetAssignedIssues(teamID)
		return messages.IssuesLoadedMsg{Issues: issues, Err: err}
	}
}

func (m RootModel) loadAllIssues(teamID string) tea.Cmd {
	return func() tea.Msg {
		issues, err := m.client.GetAllTeamIssues(teamID)
		return messages.AllIssuesLoadedMsg{Issues: issues, Err: err}
	}
}

func (m RootModel) createComment(issueID, body string) tea.Cmd {
	return func() tea.Msg {
		comment, err := m.client.CreateComment(issueID, body)
		return messages.CommentCreatedMsg{Comment: comment, Err: err}
	}
}

func (m RootModel) updateIssueTitle(issueID, title string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.UpdateIssueTitle(issueID, title)
		return messages.IssueTitleUpdatedMsg{IssueID: issueID, NewTitle: title, Err: err, Completed: true}
	}
}

func (m RootModel) updateIssuePriority(issueID string, priority int) tea.Cmd {
	return func() tea.Msg {
		err := m.client.UpdateIssuePriority(issueID, priority)
		return messages.IssuePriorityUpdatedMsg{IssueID: issueID, NewPriority: priority, Err: err, Completed: true}
	}
}

func (m RootModel) updateIssueStateByID(issueID, stateID string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.UpdateIssueState(issueID, stateID)
		return messages.IssueStateUpdatedMsg{IssueID: issueID, NewStateID: stateID, Err: err, Completed: true}
	}
}

func (m RootModel) cancelIssue(issueID, teamID string) tea.Cmd {
	return func() tea.Msg {
		stateID, err := m.client.GetCanceledStateID(teamID)
		if err != nil {
			return messages.IssueStateUpdatedMsg{IssueID: issueID, Err: err, Completed: true}
		}
		if stateID == "" {
			return messages.IssueStateUpdatedMsg{IssueID: issueID, Err: nil, Completed: true}
		}
		err = m.client.UpdateIssueState(issueID, stateID)
		return messages.IssueStateUpdatedMsg{IssueID: issueID, NewStateID: stateID, Err: err, Completed: true}
	}
}

func (m RootModel) markDuplicate(issueID, teamID string) tea.Cmd {
	return func() tea.Msg {
		stateID, err := m.client.GetDuplicateStateID(teamID)
		if err != nil {
			return messages.IssueStateUpdatedMsg{IssueID: issueID, Err: err, Completed: true}
		}
		if stateID == "" {
			stateID, err = m.client.GetCanceledStateID(teamID)
			if err != nil {
				return messages.IssueStateUpdatedMsg{IssueID: issueID, Err: err, Completed: true}
			}
		}
		if stateID == "" {
			return messages.IssueStateUpdatedMsg{IssueID: issueID, Err: nil, Completed: true}
		}
		err = m.client.UpdateIssueState(issueID, stateID)
		return messages.IssueStateUpdatedMsg{IssueID: issueID, NewStateID: stateID, Err: err, Completed: true}
	}
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || (msg.String() == "q" && m.currentView != ViewStartWork && m.currentView != ViewSettings) {
			m.quitting = true
			return m, tea.Quit
		}

	case messages.ViewerLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}

		m.teams = msg.Viewer.Viewer.Teams.Nodes

		if m.workspace != nil && m.workspace.DefaultTeamID != "" {
			for _, team := range m.teams {
				if team.ID == m.workspace.DefaultTeamID {
					m.selectedTeam = &team
					m.currentView = ViewList
					return m, tea.Batch(
						m.loadStates(team.ID),
						m.loadIssues(team.ID),
						m.loadAllIssues(team.ID),
					)
				}
			}
		}

		if len(m.teams) == 1 {
			m.selectedTeam = &m.teams[0]
			m.currentView = ViewList
			return m, tea.Batch(
				m.loadStates(m.teams[0].ID),
				m.loadIssues(m.teams[0].ID),
				m.loadAllIssues(m.teams[0].ID),
			)
		}

		m.currentView = ViewTeamSelect
		m.teamSelect = m.teamSelect.SetTeams(m.teams)
		return m, nil

	case messages.TeamSelectedMsg:
		m.selectedTeam = &msg.Team
		if msg.SetAsDefault && m.workspace != nil {
			_ = m.cfg.SetDefaultTeam(m.workspace.ID, msg.Team.ID)
		}
		m.currentView = ViewList
		return m, tea.Batch(
			m.loadStates(msg.Team.ID),
			m.loadIssues(msg.Team.ID),
			m.loadAllIssues(msg.Team.ID),
		)

	case messages.StatesLoadedMsg:
		if msg.Err != nil {
			m.list = m.list.SetError(msg.Err)
		} else {
			m.list = m.list.SetStates(msg.States)
		}
		return m, nil

	case messages.IssuesLoadedMsg:
		if msg.Err != nil {
			m.list = m.list.SetError(msg.Err)
		} else {
			m.list = m.list.SetMyIssues(msg.Issues)
		}
		return m, nil

	case messages.AllIssuesLoadedMsg:
		if msg.Err != nil {
			m.list = m.list.SetError(msg.Err)
		} else {
			m.list = m.list.SetAllIssues(msg.Issues)
		}
		return m, nil

	case messages.SwitchToListMsg:
		m.currentView = ViewList
		return m, nil

	case messages.SwitchToTeamSelectMsg:
		m.currentView = ViewTeamSelect
		m.teamSelect = m.teamSelect.SetTeams(m.teams)
		return m, nil

	case messages.SwitchToWorkspaceSelectMsg:
		m.currentView = ViewWorkspaceSelect
		m.workspaceSelect = m.workspaceSelect.SetWorkspaces(m.workspaces)
		return m, nil

	case views.WorkspaceSelectedMsg:
		if msg.AddNew {
			m.addNewWorkspace = true
			m.quitting = true
			return m, tea.Quit
		}
		if msg.Workspace != nil && (m.workspace == nil || msg.Workspace.ID != m.workspace.ID) {
			// Switch to the selected workspace
			m.workspace = msg.Workspace
			m.client = linear.NewClient(msg.Workspace.APIKey)
			// Reset list model for new workspace
			m.list = views.NewListModel()
			if branch := git.GetCurrentBranch(); branch != "" {
				m.list = m.list.SetCurrentBranch(branch)
			}
			if wd, err := os.Getwd(); err == nil {
				m.list = m.list.SetWorkingDir(wd)
			}
			m.list = m.list.SetVersion(Version)
			m.selectedTeam = nil
			m.teams = nil
			return m, m.loadViewer
		}
		m.currentView = ViewTeamSelect
		m.teamSelect = m.teamSelect.SetTeams(m.teams)
		return m, nil

	case messages.SwitchToDetailMsg:
		m.detail = views.NewDetailModel(msg.Issue)
		m.currentView = ViewDetail
		return m, nil

	case messages.NextIssueMsg:
		if next := m.list.GetNextIssue(); next != nil {
			m.list = m.list.MoveCursorNext()
			m.detail = views.NewDetailModel(*next)
		}
		return m, nil

	case messages.PrevIssueMsg:
		if prev := m.list.GetPrevIssue(); prev != nil {
			m.list = m.list.MoveCursorPrev()
			m.detail = views.NewDetailModel(*prev)
		}
		return m, nil

	case messages.SwitchToStartWorkMsg:
		m.startWork = views.NewStartWorkModel(msg.Issue)
		m.currentView = ViewStartWork
		return m, nil

	case messages.SwitchToSettingsMsg:
		m.settings = views.NewSettingsModel(m.cfg, m.workspace, m.providers)
		m.currentView = ViewSettings
		return m, nil

	case messages.OpenBrowserMsg:
		openBrowser(msg.URL)
		return m, nil

	case messages.StartClaudeMsg:
		if msg.Comment != "" {
			m.startClaude = &msg
			return m, m.createComment(msg.Issue.ID, msg.Comment)
		}
		m.startClaude = &msg
		m.quitting = true
		return m, tea.Quit

	case messages.CommentCreatedMsg:
		if m.startClaude != nil {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case messages.ErrorMsg:
		m.err = msg.Err
		return m, nil

	case messages.IssueTitleUpdatedMsg:
		if !msg.Completed {
			return m, m.updateIssueTitle(msg.IssueID, msg.NewTitle)
		}
		m.list, _ = m.list.Update(msg)
		return m, nil

	case messages.IssuePriorityUpdatedMsg:
		if !msg.Completed {
			return m, m.updateIssuePriority(msg.IssueID, msg.NewPriority)
		}
		m.list, _ = m.list.Update(msg)
		return m, nil

	case messages.IssueStateUpdatedMsg:
		if !msg.Completed {
			return m, m.updateIssueStateByID(msg.IssueID, msg.NewStateID)
		}
		m.list, _ = m.list.Update(msg)
		return m, nil

	case messages.CancelIssueMsg:
		return m, m.cancelIssue(msg.IssueID, msg.TeamID)

	case messages.MarkDuplicateMsg:
		return m, m.markDuplicate(msg.IssueID, msg.TeamID)
	}

	var cmd tea.Cmd
	switch m.currentView {
	case ViewWorkspaceSelect:
		var wsModel tea.Model
		wsModel, cmd = m.workspaceSelect.Update(msg)
		m.workspaceSelect = wsModel.(views.WorkspaceSelectModel)
	case ViewTeamSelect:
		m.teamSelect, cmd = m.teamSelect.Update(msg)
	case ViewList:
		m.list, cmd = m.list.Update(msg)
	case ViewDetail:
		m.detail, cmd = m.detail.Update(msg)
	case ViewStartWork:
		m.startWork, cmd = m.startWork.Update(msg)
	case ViewSettings:
		m.settings, cmd = m.settings.Update(msg)
	}

	return m, cmd
}

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return styles.ErrorStyle.Render(m.err.Error())
	}

	switch m.currentView {
	case ViewWorkspaceSelect:
		return m.workspaceSelect.View()
	case ViewTeamSelect:
		return m.teamSelect.View()
	case ViewList:
		return m.list.View()
	case ViewDetail:
		return m.detail.View()
	case ViewStartWork:
		return m.startWork.View()
	case ViewSettings:
		return m.settings.View()
	}

	return "Loading..."
}

func (m RootModel) ShouldStartClaude() *messages.StartClaudeMsg {
	return m.startClaude
}

func (m RootModel) ShouldAddNewWorkspace() bool {
	return m.addNewWorkspace
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}

	_ = cmd.Start()
}
