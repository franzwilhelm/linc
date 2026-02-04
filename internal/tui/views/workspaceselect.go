package views

import (
	"fmt"

	"linc/internal/config"
	"linc/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkspaceSelectedMsg struct {
	Workspace *config.Workspace
	AddNew    bool
}

type WorkspaceSelectModel struct {
	workspaces []config.Workspace
	cursor     int
	done       bool
	integrated bool // when true, emit messages instead of quitting
}

func NewWorkspaceSelectModel(workspaces []config.Workspace) WorkspaceSelectModel {
	return WorkspaceSelectModel{
		workspaces: workspaces,
	}
}

func NewIntegratedWorkspaceSelectModel(workspaces []config.Workspace) WorkspaceSelectModel {
	return WorkspaceSelectModel{
		workspaces: workspaces,
		integrated: true,
	}
}

func (m WorkspaceSelectModel) Init() tea.Cmd {
	return nil
}

func (m WorkspaceSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.workspaces) {
				m.cursor++
			}
		case "enter":
			m.done = true
			if m.integrated {
				if m.cursor == len(m.workspaces) {
					return m, func() tea.Msg {
						return WorkspaceSelectedMsg{AddNew: true}
					}
				}
				return m, func() tea.Msg {
					return WorkspaceSelectedMsg{Workspace: &m.workspaces[m.cursor]}
				}
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m WorkspaceSelectModel) View() string {
	if m.done {
		return ""
	}

	s := styles.TitleStyle.Render("Select a Linear workspace for this directory") + "\n\n"

	var rows []string
	for i, ws := range m.workspaces {
		isFirst := i == 0
		if m.cursor == i {
			rows = append(rows, styles.SelectedRowStyle.Render(ws.Name))
		} else if i == m.cursor-1 {
			if isFirst {
				rows = append(rows, styles.FirstRowAboveSelectedStyle.Render(ws.Name))
			} else {
				rows = append(rows, styles.RowAboveSelectedStyle.Render(ws.Name))
			}
		} else if isFirst {
			rows = append(rows, styles.FirstRowStyle.Render(ws.Name))
		} else {
			rows = append(rows, styles.RowStyle.Render(ws.Name))
		}
	}

	addNewContent := "+ Add new workspace"
	if m.cursor == len(m.workspaces) {
		rows = append(rows, styles.SelectedRowStyle.Render(addNewContent))
	} else {
		rows = append(rows, styles.RowStyle.Render(addNewContent))
	}

	s += lipgloss.JoinVertical(lipgloss.Left, rows...)
	s += styles.HelpStyle.Render("\n\nj/k: navigate • enter: select • q: quit")

	return s
}

func (m WorkspaceSelectModel) Selected() (*config.Workspace, bool) {
	if m.cursor == len(m.workspaces) {
		return nil, true
	}
	if m.cursor < len(m.workspaces) {
		return &m.workspaces[m.cursor], false
	}
	return nil, false
}

func (m WorkspaceSelectModel) SetWorkspaces(workspaces []config.Workspace) WorkspaceSelectModel {
	m.workspaces = workspaces
	m.done = false
	return m
}

func RunWorkspaceSelector(workspaces []config.Workspace) (*config.Workspace, bool, error) {
	model := NewWorkspaceSelectModel(workspaces)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}

	if m, ok := finalModel.(WorkspaceSelectModel); ok {
		ws, addNew := m.Selected()
		return ws, addNew, nil
	}

	return nil, false, fmt.Errorf("unexpected model type")
}
