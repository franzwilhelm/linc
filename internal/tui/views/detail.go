package views

import (
	"fmt"
	"strings"
	"time"

	"linc/internal/linear"
	"linc/internal/tui/messages"
	"linc/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	issue        linear.Issue
	activeButton int // 0 = Open in Browser, 1 = Start Working
}

func NewDetailModel(issue linear.Issue) DetailModel {
	return DetailModel{
		issue:        issue,
		activeButton: 1, // Default to Start Working
	}
}

func (m DetailModel) Init() tea.Cmd {
	return nil
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			return m, func() tea.Msg {
				return messages.NextIssueMsg{}
			}
		case "k", "up":
			return m, func() tea.Msg {
				return messages.PrevIssueMsg{}
			}
		case "left", "h":
			if m.activeButton > 0 {
				m.activeButton--
			}
		case "right", "l":
			if m.activeButton < 1 {
				m.activeButton++
			}
		case "tab":
			m.activeButton = (m.activeButton + 1) % 2
		case "o":
			return m, func() tea.Msg {
				return messages.OpenBrowserMsg{URL: m.issue.URL}
			}
		case "s", "enter":
			if m.activeButton == 0 {
				return m, func() tea.Msg {
					return messages.OpenBrowserMsg{URL: m.issue.URL}
				}
			}
			return m, func() tea.Msg {
				return messages.SwitchToStartWorkMsg{Issue: m.issue}
			}
		case "esc":
			return m, func() tea.Msg {
				return messages.SwitchToListMsg{}
			}
		}
	}

	return m, nil
}

func (m DetailModel) View() string {
	var s strings.Builder

	// Header row (same format as list)
	s.WriteString(m.renderHeaderRow() + "\n\n")

	// Metadata
	s.WriteString(m.renderField("Status", m.renderStateIcon()+" "+m.issue.State.Name) + "\n")
	s.WriteString(m.renderField("Priority", m.renderPriority()+" "+m.priorityLabel(m.issue.Priority)) + "\n")

	if m.issue.Assignee != nil {
		s.WriteString(m.renderField("Assignee", m.renderAssignee()) + "\n")
	} else {
		s.WriteString(m.renderField("Assignee", styles.SubtitleStyle.Render("Unassigned")) + "\n")
	}

	if m.issue.Cycle != nil {
		cycleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
		s.WriteString(m.renderField("Cycle", cycleStyle.Render(fmt.Sprintf("#%d %s", m.issue.Cycle.Number, m.issue.Cycle.Name))) + "\n")
	}

	if m.issue.CreatedAt != "" {
		s.WriteString(m.renderField("Created", m.formatDate(m.issue.CreatedAt)) + "\n")
	}

	if len(m.issue.Labels) > 0 {
		var labelParts []string
		for _, label := range m.issue.Labels {
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color(hexToAnsiDetail(label.Color))).
				Padding(0, 1)
			labelParts = append(labelParts, labelStyle.Render(label.Name))
		}
		s.WriteString(m.renderField("Labels", strings.Join(labelParts, " ")) + "\n")
	}

	// Description
	s.WriteString("\n" + styles.DetailLabelStyle.Render("Description") + "\n")
	if m.issue.Description == "" {
		s.WriteString(styles.DetailDescriptionStyle.Render("(No description)") + "\n")
	} else {
		s.WriteString(formatMarkdown(m.issue.Description) + "\n")
	}

	// Buttons
	s.WriteString("\n")
	openBtn := styles.ButtonStyle.Render("Open in Browser (o)")
	startBtn := styles.ButtonStyle.Render("Start Working (s)")

	if m.activeButton == 0 {
		openBtn = styles.ActiveButtonStyle.Render("Open in Browser (o)")
	} else {
		startBtn = styles.ActiveButtonStyle.Render("Start Working (s)")
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, openBtn, startBtn))

	// Help
	s.WriteString(styles.HelpStyle.Render("\n\nj/k: prev/next issue • h/l: switch button • enter: activate • o: open • s: start • esc: back"))

	return s.String()
}

func (m DetailModel) renderHeaderRow() string {
	prio := m.renderPriority()
	identifier := styles.IssueIdentifierStyle.Render(m.issue.Identifier)
	stateIcon := m.renderStateIcon()
	title := styles.DetailTitleStyle.Render(m.issue.Title)
	return fmt.Sprintf("%s  %s %s %s", prio, identifier, stateIcon, title)
}

func (m DetailModel) renderPriority() string {
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	switch m.issue.Priority {
	case 1: // Urgent
		urgentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return urgentStyle.Render("[!]")
	case 2: // High
		return orangeStyle.Render("▂▄▆")
	case 3: // Medium
		return orangeStyle.Render("▂▄") + dimStyle.Render("▆")
	case 4: // Low
		return orangeStyle.Render("▂") + dimStyle.Render("▄▆")
	default:
		return dimStyle.Render("---")
	}
}

func (m DetailModel) renderStateIcon() string {
	color := "241"
	if m.issue.State.Color != "" {
		color = hexToAnsiDetail(m.issue.State.Color)
	}
	icon := "○"
	switch m.issue.State.Type {
	case "started":
		icon = "◐"
	case "completed":
		icon = "●"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(icon)
}

func (m DetailModel) renderAssignee() string {
	if m.issue.Assignee == nil {
		return ""
	}
	initials := getInitialsDetail(m.issue.Assignee.Name)
	assigneeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("62")).
		Padding(0, 1)
	return assigneeStyle.Render(initials) + " " + m.issue.Assignee.Name
}

func (m DetailModel) formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Jan 2, 2006")
}

func hexToAnsiDetail(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "241"
	}
	return "#" + hex
}

func getInitialsDetail(name string) string {
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

func (m DetailModel) renderField(label, value string) string {
	return styles.DetailLabelStyle.Render(label+":") + " " + styles.DetailValueStyle.Render(value)
}

func (m DetailModel) priorityLabel(priority int) string {
	switch priority {
	case 0:
		return "No priority"
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return fmt.Sprintf("Priority %d", priority)
	}
}

func (m DetailModel) Issue() linear.Issue {
	return m.issue
}

// formatMarkdown renders markdown with rich terminal formatting using glamour
func formatMarkdown(text string) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return text
	}

	rendered, err := renderer.Render(text)
	if err != nil {
		return text
	}

	return strings.TrimSpace(rendered)
}
