package views

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"linc/internal/linear"
	"linc/internal/tui/messages"
	"linc/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditMode int

const (
	EditModeNone EditMode = iota
	EditModeRename
	EditModePriority
	EditModeStatus
)

type ListModel struct {
	issues        []linear.Issue
	allIssues     []linear.Issue
	myIssues      []linear.Issue
	states        []linear.State
	issuesByState map[string][]linear.Issue
	filtered      []linear.Issue
	cursor        int
	activeState   int
	filtering     bool
	filterInput   textinput.Model
	loading       bool
	showAllIssues bool // false = my issues, true = all issues
	err           error
	currentBranch string         // current git branch
	currentIssue  *linear.Issue  // issue matching current branch (if any)
	workingDir    string         // current working directory
	version       string         // app version

	// Edit mode state
	editMode      EditMode
	editInput     textinput.Model  // for renaming
	editCursor    int              // for priority/status selection
	editIssue     *linear.Issue    // issue being edited
}

func NewListModel() ListModel {
	ti := textinput.New()
	ti.Placeholder = "Filter issues..."
	ti.CharLimit = 100
	ti.Width = 40

	editTi := textinput.New()
	editTi.Placeholder = "New title..."
	editTi.CharLimit = 200
	editTi.Width = 60

	return ListModel{
		filterInput:   ti,
		editInput:     editTi,
		issuesByState: make(map[string][]linear.Issue),
		loading:       true,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) currentStateIssues() []linear.Issue {
	if len(m.states) == 0 || m.activeState >= len(m.states) {
		return nil
	}
	stateID := m.states[m.activeState].ID
	return m.issuesByState[stateID]
}

func (m *ListModel) applyFilter() {
	issues := m.currentStateIssues()
	filter := strings.ToLower(m.filterInput.Value())

	if filter == "" {
		m.filtered = issues
		return
	}

	m.filtered = make([]linear.Issue, 0)
	for _, issue := range issues {
		if strings.Contains(strings.ToLower(issue.Title), filter) ||
			strings.Contains(strings.ToLower(issue.Identifier), filter) ||
			strings.Contains(strings.ToLower(issue.Description), filter) {
			m.filtered = append(m.filtered, issue)
		}
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *ListModel) groupIssuesByState() {
	m.issuesByState = make(map[string][]linear.Issue)
	for _, issue := range m.issues {
		stateID := issue.State.ID
		m.issuesByState[stateID] = append(m.issuesByState[stateID], issue)
	}
	for stateID := range m.issuesByState {
		sort.Slice(m.issuesByState[stateID], func(i, j int) bool {
			pi, pj := m.issuesByState[stateID][i].Priority, m.issuesByState[stateID][j].Priority
			if pi == 0 {
				pi = 99
			}
			if pj == 0 {
				pj = 99
			}
			return pi < pj
		})
	}
}

func (m *ListModel) findDefaultState() {
	for i, state := range m.states {
		if strings.EqualFold(state.Name, "todo") {
			m.activeState = i
			return
		}
	}
	m.activeState = 0
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.StatesLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.states = msg.States
		m.findDefaultState()
		m.applyFilter()
		return m, nil

	case messages.IssuesLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		m.issues = msg.Issues
		m.loading = false
		m.groupIssuesByState()
		m.applyFilter()
		return m, nil

	case messages.IssueTitleUpdatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.updateIssueTitle(msg.IssueID, msg.NewTitle)
		}
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil

	case messages.IssuePriorityUpdatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.updateIssuePriority(msg.IssueID, msg.NewPriority)
		}
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil

	case messages.IssueStateUpdatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.updateIssueState(msg.IssueID, msg.NewStateID)
		}
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil

	case tea.KeyMsg:
		if m.editMode == EditModeRename {
			return m.handleRenameInput(msg)
		}
		if m.editMode == EditModePriority {
			return m.handlePriorityInput(msg)
		}
		if m.editMode == EditModeStatus {
			return m.handleStatusInput(msg)
		}

		if m.filtering {
			switch msg.String() {
			case "enter", "esc":
				m.filtering = false
				m.filterInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				m.applyFilter()
				return m, cmd
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "left", "h":
			if m.activeState > 0 {
				m.activeState--
				m.cursor = 0
				m.filterInput.SetValue("")
				m.applyFilter()
			}
		case "right", "l":
			if m.activeState < len(m.states)-1 {
				m.activeState++
				m.cursor = 0
				m.filterInput.SetValue("")
				m.applyFilter()
			}
		case "/":
			m.filtering = true
			m.filterInput.Focus()
			return m, textinput.Blink
		case "enter":
			if len(m.filtered) > 0 {
				return m, func() tea.Msg {
					return messages.SwitchToDetailMsg{Issue: m.filtered[m.cursor]}
				}
			}
		case "esc":
			if m.filterInput.Value() != "" {
				m.filterInput.SetValue("")
				m.applyFilter()
			} else {
				return m, func() tea.Msg {
					return messages.SwitchToTeamSelectMsg{}
				}
			}
		case "a":
			m = m.ToggleShowAll()
			return m, nil
		case "R":
			if len(m.filtered) > 0 {
				m.editMode = EditModeRename
				m.editIssue = &m.filtered[m.cursor]
				m.editInput.SetValue(m.editIssue.Title)
				m.editInput.Focus()
				m.editInput.CursorEnd()
				return m, textinput.Blink
			}
		case "p":
			if len(m.filtered) > 0 {
				m.editMode = EditModePriority
				m.editIssue = &m.filtered[m.cursor]
				m.editCursor = m.editIssue.Priority
			}
		case "s":
			if len(m.filtered) > 0 {
				m.editMode = EditModeStatus
				m.editIssue = &m.filtered[m.cursor]
				// Find current state index
				for i, state := range m.states {
					if state.ID == m.editIssue.State.ID {
						m.editCursor = i
						break
					}
				}
			}
		case ",":
			return m, func() tea.Msg {
				return messages.SwitchToSettingsMsg{}
			}
		}
	}

	return m, nil
}

func (m ListModel) handleRenameInput(msg tea.KeyMsg) (ListModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		newTitle := m.editInput.Value()
		if newTitle != "" && m.editIssue != nil {
			issueID := m.editIssue.ID
			m.editInput.Blur()
			return m, func() tea.Msg {
				return messages.IssueTitleUpdatedMsg{
					IssueID:  issueID,
					NewTitle: newTitle,
				}
			}
		}
		m.editMode = EditModeNone
		m.editIssue = nil
		m.editInput.Blur()
		return m, nil
	case "esc":
		m.editMode = EditModeNone
		m.editIssue = nil
		m.editInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}
}

func (m ListModel) handlePriorityInput(msg tea.KeyMsg) (ListModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.editCursor > 0 {
			m.editCursor--
		}
	case "down", "j":
		if m.editCursor < 4 {
			m.editCursor++
		}
	case "enter":
		if m.editIssue != nil {
			issueID := m.editIssue.ID
			priority := m.editCursor
			return m, func() tea.Msg {
				return messages.IssuePriorityUpdatedMsg{
					IssueID:     issueID,
					NewPriority: priority,
				}
			}
		}
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil
	case "esc":
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil
	case "0", "1", "2", "3", "4":
		priority := int(msg.String()[0] - '0')
		if m.editIssue != nil {
			issueID := m.editIssue.ID
			return m, func() tea.Msg {
				return messages.IssuePriorityUpdatedMsg{
					IssueID:     issueID,
					NewPriority: priority,
				}
			}
		}
	}
	return m, nil
}

func (m ListModel) handleStatusInput(msg tea.KeyMsg) (ListModel, tea.Cmd) {
	maxCursor := len(m.states) + 1

	switch msg.String() {
	case "up", "k":
		if m.editCursor > 0 {
			m.editCursor--
		}
	case "down", "j":
		if m.editCursor < maxCursor {
			m.editCursor++
		}
	case "enter":
		if m.editIssue != nil {
			issueID := m.editIssue.ID
			teamID := m.editIssue.Team.ID

			if m.editCursor == len(m.states) {
				return m, func() tea.Msg {
					return messages.CancelIssueMsg{
						IssueID: issueID,
						TeamID:  teamID,
					}
				}
			} else if m.editCursor == len(m.states)+1 {
				return m, func() tea.Msg {
					return messages.MarkDuplicateMsg{
						IssueID: issueID,
						TeamID:  teamID,
					}
				}
			} else if m.editCursor < len(m.states) {
				stateID := m.states[m.editCursor].ID
				return m, func() tea.Msg {
					return messages.IssueStateUpdatedMsg{
						IssueID:    issueID,
						NewStateID: stateID,
					}
				}
			}
		}
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil
	case "esc":
		m.editMode = EditModeNone
		m.editIssue = nil
		return m, nil
	}
	return m, nil
}

func (m *ListModel) updateIssueTitle(issueID, newTitle string) {
	for i := range m.issues {
		if m.issues[i].ID == issueID {
			m.issues[i].Title = newTitle
			break
		}
	}
	for i := range m.myIssues {
		if m.myIssues[i].ID == issueID {
			m.myIssues[i].Title = newTitle
			break
		}
	}
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			m.allIssues[i].Title = newTitle
			break
		}
	}
	m.groupIssuesByState()
	m.applyFilter()
}

func (m *ListModel) updateIssuePriority(issueID string, newPriority int) {
	for i := range m.issues {
		if m.issues[i].ID == issueID {
			m.issues[i].Priority = newPriority
			break
		}
	}
	for i := range m.myIssues {
		if m.myIssues[i].ID == issueID {
			m.myIssues[i].Priority = newPriority
			break
		}
	}
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			m.allIssues[i].Priority = newPriority
			break
		}
	}
	m.groupIssuesByState()
	m.applyFilter()
}

func (m *ListModel) updateIssueState(issueID, newStateID string) {
	var newState linear.State
	stateFound := false
	for _, state := range m.states {
		if state.ID == newStateID {
			newState = state
			stateFound = true
			break
		}
	}

	if !stateFound {
		m.removeIssue(issueID)
		return
	}

	for i := range m.issues {
		if m.issues[i].ID == issueID {
			m.issues[i].State = newState
			break
		}
	}
	for i := range m.myIssues {
		if m.myIssues[i].ID == issueID {
			m.myIssues[i].State = newState
			break
		}
	}
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			m.allIssues[i].State = newState
			break
		}
	}
	m.groupIssuesByState()
	m.applyFilter()
}

func (m *ListModel) removeIssue(issueID string) {
	for i := range m.issues {
		if m.issues[i].ID == issueID {
			m.issues = append(m.issues[:i], m.issues[i+1:]...)
			break
		}
	}
	for i := range m.myIssues {
		if m.myIssues[i].ID == issueID {
			m.myIssues = append(m.myIssues[:i], m.myIssues[i+1:]...)
			break
		}
	}
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			m.allIssues = append(m.allIssues[:i], m.allIssues[i+1:]...)
			break
		}
	}
	if m.cursor >= len(m.filtered)-1 && m.cursor > 0 {
		m.cursor--
	}
	m.groupIssuesByState()
	m.applyFilter()
}

func (m ListModel) View() string {
	if m.loading {
		return m.renderLogo() + "\n\n" + styles.TitleStyle.Render("Loading issues...")
	}

	if m.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.editMode == EditModeRename {
		return m.renderRenameView()
	}
	if m.editMode == EditModePriority {
		return m.renderPriorityView()
	}
	if m.editMode == EditModeStatus {
		return m.renderStatusView()
	}

	var s strings.Builder

	s.WriteString(m.renderLogo() + "\n")

	if m.currentBranch != "" {
		s.WriteString(m.renderBranchBox() + "\n")
	}

	modeLabel := "My Issues"
	if m.showAllIssues {
		modeLabel = "All Issues"
	}
	s.WriteString(styles.TitleStyle.Render(modeLabel) + "\n\n")

	for i, state := range m.states {
		count := len(m.issuesByState[state.ID])
		icon := renderStateIcon(state)
		text := fmt.Sprintf("%s (%d)", state.Name, count)

		if i == m.activeState {
			s.WriteString(icon + styles.ActiveTabStyle.Render(text))
		} else {
			s.WriteString(icon + styles.InactiveTabStyle.Render(text))
		}
	}
	s.WriteString("\n\n")

	if m.filtering {
		s.WriteString(styles.FilterPromptStyle.Render("/") + " " + m.filterInput.View() + "\n")
	} else if m.filterInput.Value() != "" {
		s.WriteString(styles.FilterPromptStyle.Render("Filter: ") + m.filterInput.Value() + "\n")
	}

	if len(m.filtered) == 0 {
		s.WriteString(styles.SubtitleStyle.Render("No issues"))
	} else {
		const maxVisible = 10
		total := len(m.filtered)

		start := 0
		end := total
		if total > maxVisible {
			start = m.cursor - maxVisible/2
			if start < 0 {
				start = 0
			}
			end = start + maxVisible
			if end > total {
				end = total
				start = end - maxVisible
			}
		}

		if start > 0 {
			s.WriteString(styles.SubtitleStyle.Render(fmt.Sprintf("  ↑ %d more above", start)) + "\n")
		}

		var rows []string
		for i := start; i < end; i++ {
			isSelected := m.cursor == i
			isAboveSelected := i == m.cursor-1
			isFirst := i == start
			rows = append(rows, m.renderIssueRow(m.filtered[i], isSelected, isAboveSelected, isFirst))
		}
		s.WriteString(lipgloss.JoinVertical(lipgloss.Left, rows...))

		if end < total {
			s.WriteString("\n" + styles.SubtitleStyle.Render(fmt.Sprintf("  ↓ %d more below", total-end)))
		}
	}

	s.WriteString(styles.HelpStyle.Render("\nh/l: status • j/k: navigate • R: rename • p: priority • s: status • a: my/all • /: filter • ,: settings • enter: select • q: quit"))

	return s.String()
}

func (m ListModel) renderRenameView() string {
	var s strings.Builder
	s.WriteString(styles.TitleStyle.Render("Rename Issue") + "\n\n")
	if m.editIssue != nil {
		s.WriteString(styles.IssueIdentifierStyle.Render(m.editIssue.Identifier) + "\n\n")
	}
	s.WriteString(m.editInput.View() + "\n\n")
	s.WriteString(styles.HelpStyle.Render("enter: save • esc: cancel"))
	return s.String()
}

func (m ListModel) renderPriorityView() string {
	var s strings.Builder
	s.WriteString(styles.TitleStyle.Render("Set Priority") + "\n\n")
	if m.editIssue != nil {
		s.WriteString(styles.IssueIdentifierStyle.Render(m.editIssue.Identifier) + " " + m.editIssue.Title + "\n\n")
	}

	priorities := []struct {
		value int
		label string
		icon  string
	}{
		{0, "No priority", "---"},
		{1, "Urgent", "[!]"},
		{2, "High", "▂▄▆"},
		{3, "Medium", "▂▄ "},
		{4, "Low", "▂  "},
	}

	for _, p := range priorities {
		cursor := "  "
		if m.editCursor == p.value {
			cursor = styles.CursorStyle.Render("> ")
		}
		label := fmt.Sprintf("%s %d: %s %s", cursor, p.value, p.icon, p.label)
		if m.editIssue != nil && m.editIssue.Priority == p.value {
			label += " (current)"
		}
		s.WriteString(label + "\n")
	}

	s.WriteString(styles.HelpStyle.Render("\nj/k: navigate • 0-4: quick select • enter: save • esc: cancel"))
	return s.String()
}

func (m ListModel) renderStatusView() string {
	var s strings.Builder
	s.WriteString(styles.TitleStyle.Render("Set Status") + "\n\n")
	if m.editIssue != nil {
		s.WriteString(styles.IssueIdentifierStyle.Render(m.editIssue.Identifier) + " " + m.editIssue.Title + "\n\n")
	}

	for i, state := range m.states {
		cursor := "  "
		if m.editCursor == i {
			cursor = styles.CursorStyle.Render("> ")
		}
		stateIcon := renderStateIcon(state)
		label := fmt.Sprintf("%s%s %s", cursor, stateIcon, state.Name)
		if m.editIssue != nil && m.editIssue.State.ID == state.ID {
			label += " (current)"
		}
		s.WriteString(label + "\n")
	}

	s.WriteString("\n")

	cancelCursor := "  "
	if m.editCursor == len(m.states) {
		cancelCursor = styles.CursorStyle.Render("> ")
	}
	cancelIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("○")
	s.WriteString(fmt.Sprintf("%s%s Canceled\n", cancelCursor, cancelIcon))

	dupCursor := "  "
	if m.editCursor == len(m.states)+1 {
		dupCursor = styles.CursorStyle.Render("> ")
	}
	dupIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("○")
	s.WriteString(fmt.Sprintf("%s%s Duplicate\n", dupCursor, dupIcon))

	s.WriteString(styles.HelpStyle.Render("\nj/k: navigate • enter: save • esc: cancel"))
	return s.String()
}

func (m ListModel) renderIssueRow(issue linear.Issue, selected bool, aboveSelected bool, isFirst bool) string {
	const (
		colPrio       = 3
		colIdentifier = 10
		colState      = 2
		colTitle      = 120
		colCycle      = 6
		colEstimate   = 4
		colAssignee   = 4
		colDate       = 7
	)

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	var prio string
	switch issue.Priority {
	case 1:
		urgentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		prio = urgentStyle.Render("[!]")
	case 2:
		prio = whiteStyle.Render("▂▄▆")
	case 3:
		prio = whiteStyle.Render("▂▄") + dimStyle.Render("▆")
	case 4:
		prio = whiteStyle.Render("▂") + dimStyle.Render("▄▆")
	default:
		prio = dimStyle.Render("---")
	}

	identifierStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	identifier := identifierStyle.Render(issue.Identifier)
	identifier = padRightStyled(identifier, colIdentifier)

	stateIcon := renderStateIcon(issue.State)

	title := issue.Title
	if len(title) > colTitle-3 {
		title = title[:colTitle-3] + "..."
	}
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	title = titleStyle.Render(title)
	title = padRightStyled(title, colTitle)

	var cycleStr string
	if issue.Cycle != nil {
		cycleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		cycleStr = cycleStyle.Render(fmt.Sprintf("▶ %d", issue.Cycle.Number))
	}
	cycleStr = padRightStyled(cycleStr, colCycle)

	var estStr string
	if issue.Estimate != nil {
		estStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		estStr = estStyle.Render(fmt.Sprintf("%.0f", *issue.Estimate))
	}
	estStr = padRightStyled(estStr, colEstimate)

	var assigneeStr string
	if issue.Assignee != nil {
		initials := getInitials(issue.Assignee.Name)
		assigneeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)
		assigneeStr = assigneeStyle.Render(initials)
	}
	assigneeStr = padRightStyled(assigneeStr, colAssignee)

	dateStr := formatShortDate(issue.CreatedAt)
	if dateStr != "" {
		dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		dateStr = dateStyle.Render(dateStr)
	}
	dateStr = padRightStyled(dateStr, colDate)

	row := fmt.Sprintf("%s %s %s %s %s %s %s %s",
		prio, identifier, stateIcon, title, cycleStr, estStr, assigneeStr, dateStr)

	if selected {
		return styles.SelectedRowStyle.Render(row)
	}
	if aboveSelected {
		if isFirst {
			return styles.FirstRowAboveSelectedStyle.Render(row)
		}
		return styles.RowAboveSelectedStyle.Render(row)
	}
	if isFirst {
		return styles.FirstRowStyle.Render(row)
	}
	return styles.RowStyle.Render(row)
}

func (m ListModel) renderBranchBox() string {
	var content strings.Builder

	content.WriteString(styles.BranchLabelStyle.Render("branch: "))
	content.WriteString(styles.BranchNameStyle.Render(m.currentBranch))

	if m.currentIssue != nil {
		content.WriteString("\n")
		content.WriteString(styles.BranchLabelStyle.Render("issue:  "))
		stateIcon := renderStateIcon(m.currentIssue.State)
		identifier := styles.IssueIdentifierStyle.Render(m.currentIssue.Identifier)
		title := m.currentIssue.Title
		content.WriteString(fmt.Sprintf("%s %s %s", stateIcon, identifier, title))
	}

	return styles.BranchBoxStyle.Render(content.String())
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padRightStyled(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

func renderStateIcon(state linear.State) string {
	color := "241"
	if state.Color != "" {
		color = hexToAnsi(state.Color)
	}

	icon := "○"
	switch state.Type {
	case "started":
		icon = "◐"
	case "completed":
		icon = "●"
	case "canceled":
		icon = "○"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(icon)
}

func hexToAnsi(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "241"
	}
	return "#" + hex
}

func getInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	if len(parts) == 1 {
		if len(parts[0]) >= 2 {
			return strings.ToUpper(parts[0][:2])
		}
		return strings.ToUpper(parts[0])
	}
	return strings.ToUpper(string(parts[0][0]) + string(parts[len(parts)-1][0]))
}

func formatShortDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return ""
	}
	return t.Format("Jan 2")
}

func (m ListModel) renderLogo() string {
	brown := lipgloss.NewStyle().Foreground(lipgloss.Color("94"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	silver := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	white := styles.LogoTextStyle
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	version := m.version
	if version == "" {
		version = "dev"
	}

	wd := m.workingDir
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(wd, home) {
		wd = "~" + strings.TrimPrefix(wd, home)
	}

	lines := []string{
		brown.Render("  ▄█▄  "),
		green.Render("   █   "),
		brown.Render(" ▄███▄ "),
		silver.Render("   █   "),
		silver.Render("  ▐█▌  ") + white.Render(" LINC ") + dim.Render(version),
		silver.Render("  ▐█▌  ") + dim.Render(" "+wd),
		silver.Render("  ▐█▌  "),
		silver.Render("  ▐█▌  "),
		silver.Render("   ▀   "),
	}

	return strings.Join(lines, "\n")
}

func (m ListModel) SetMyIssues(issues []linear.Issue) ListModel {
	m.myIssues = issues
	if !m.showAllIssues {
		m.issues = issues
		m.groupIssuesByState()
		m.applyFilter()
	}
	m.loading = false
	m.findCurrentIssue()
	return m
}

func (m ListModel) SetAllIssues(issues []linear.Issue) ListModel {
	m.allIssues = issues
	if m.showAllIssues {
		m.issues = issues
		m.groupIssuesByState()
		m.applyFilter()
	}
	m.findCurrentIssue()
	return m
}

func (m ListModel) ToggleShowAll() ListModel {
	m.showAllIssues = !m.showAllIssues
	if m.showAllIssues {
		m.issues = m.allIssues
	} else {
		m.issues = m.myIssues
	}
	m.cursor = 0
	m.groupIssuesByState()
	m.applyFilter()
	return m
}

func (m ListModel) ShowingAllIssues() bool {
	return m.showAllIssues
}

func (m ListModel) GetCurrentIssue() *linear.Issue {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		return &m.filtered[m.cursor]
	}
	return nil
}

func (m ListModel) GetNextIssue() *linear.Issue {
	if m.cursor+1 < len(m.filtered) {
		return &m.filtered[m.cursor+1]
	}
	return nil
}

func (m ListModel) GetPrevIssue() *linear.Issue {
	if m.cursor > 0 {
		return &m.filtered[m.cursor-1]
	}
	return nil
}

func (m ListModel) MoveCursorNext() ListModel {
	if m.cursor+1 < len(m.filtered) {
		m.cursor++
	}
	return m
}

func (m ListModel) MoveCursorPrev() ListModel {
	if m.cursor > 0 {
		m.cursor--
	}
	return m
}

func (m ListModel) SetStates(states []linear.State) ListModel {
	m.states = states
	m.findDefaultState()
	m.applyFilter()
	return m
}

func (m ListModel) SetError(err error) ListModel {
	m.err = err
	m.loading = false
	return m
}

func (m ListModel) IsLoading() bool {
	return m.loading
}

func (m ListModel) SetCurrentBranch(branch string) ListModel {
	m.currentBranch = branch
	m.findCurrentIssue()
	return m
}

func (m *ListModel) findCurrentIssue() {
	if m.currentBranch == "" {
		m.currentIssue = nil
		return
	}

	allIssues := append(m.myIssues, m.allIssues...)
	for i := range allIssues {
		issue := &allIssues[i]
		if strings.Contains(m.currentBranch, issue.Identifier) {
			m.currentIssue = issue
			return
		}
		if issue.BranchName != "" && m.currentBranch == issue.BranchName {
			m.currentIssue = issue
			return
		}
	}
	m.currentIssue = nil
}

func (m ListModel) SetWorkingDir(dir string) ListModel {
	m.workingDir = dir
	return m
}

func (m ListModel) SetVersion(version string) ListModel {
	m.version = version
	return m
}
