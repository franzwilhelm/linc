package views

import (
	"fmt"

	"linc/internal/linear"
	"linc/internal/tui/messages"
	"linc/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TeamSelectModel struct {
	teams        []linear.Team
	cursor       int
	setAsDefault bool
	loading      bool
	err          error
}

func NewTeamSelectModel() TeamSelectModel {
	return TeamSelectModel{
		loading: true,
	}
}

func (m TeamSelectModel) Init() tea.Cmd {
	return nil
}

func (m TeamSelectModel) Update(msg tea.Msg) (TeamSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.TeamsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.teams = msg.Teams
		return m, nil

	case tea.KeyMsg:
		if m.loading || m.err != nil {
			return m, nil
		}

		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return messages.SwitchToWorkspaceSelectMsg{}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.teams)-1 {
				m.cursor++
			}
		case "d":
			m.setAsDefault = !m.setAsDefault
		case "enter":
			if len(m.teams) > 0 {
				return m, func() tea.Msg {
					return messages.TeamSelectedMsg{
						Team:         m.teams[m.cursor],
						SetAsDefault: m.setAsDefault,
					}
				}
			}
		}
	}

	return m, nil
}

func (m TeamSelectModel) View() string {
	if m.loading {
		return styles.TitleStyle.Render("Loading teams...")
	}

	if m.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error loading teams: %v", m.err))
	}

	if len(m.teams) == 0 {
		return styles.SubtitleStyle.Render("No teams found")
	}

	s := styles.TitleStyle.Render("Select a Team") + "\n\n"

	var rows []string
	for i, team := range m.teams {
		content := fmt.Sprintf("%s (%s)", team.Name, team.Key)
		isFirst := i == 0
		if m.cursor == i {
			rows = append(rows, styles.SelectedRowStyle.Render(content))
		} else if i == m.cursor-1 {
			if isFirst {
				rows = append(rows, styles.FirstRowAboveSelectedStyle.Render(content))
			} else {
				rows = append(rows, styles.RowAboveSelectedStyle.Render(content))
			}
		} else if isFirst {
			rows = append(rows, styles.FirstRowStyle.Render(content))
		} else {
			rows = append(rows, styles.RowStyle.Render(content))
		}
	}
	s += lipgloss.JoinVertical(lipgloss.Left, rows...)

	s += "\n\n"

	checkbox := "[ ]"
	if m.setAsDefault {
		checkbox = styles.CheckboxCheckedStyle.Render("[x]")
	}
	s += fmt.Sprintf("%s Set as default team (d to toggle)\n", checkbox)

	s += styles.HelpStyle.Render("\nj/k: navigate • enter: select • d: toggle default • esc: back • q: quit")

	return s
}

func (m TeamSelectModel) SetTeams(teams []linear.Team) TeamSelectModel {
	m.teams = teams
	m.loading = false
	return m
}

func (m TeamSelectModel) SetError(err error) TeamSelectModel {
	m.err = err
	m.loading = false
	return m
}
