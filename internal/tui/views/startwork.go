package views

import (
	"fmt"
	"strings"

	"linc/internal/linear"
	"linc/internal/tui/messages"
	"linc/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type StartWorkModel struct {
	issue         linear.Issue
	commentInput  textinput.Model
	useBranchName bool
	planMode      bool
	focusIndex    int // 0 = comment, 1 = useBranch, 2 = planMode, 3 = start button, 4 = checkout only
	err           error
}

func NewStartWorkModel(issue linear.Issue) StartWorkModel {
	ti := textinput.New()
	ti.Placeholder = "Add a comment (optional, syncs to Linear)"
	ti.CharLimit = 500
	ti.Width = 60
	ti.Focus()

	return StartWorkModel{
		issue:         issue,
		commentInput:  ti,
		useBranchName: true,
		planMode:      true,
		focusIndex:    0,
	}
}

func (m StartWorkModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m StartWorkModel) Update(msg tea.Msg) (StartWorkModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle text input first when focused on comment field
		if m.focusIndex == 0 {
			switch msg.String() {
			case "tab", "down":
				m.focusIndex = 1
				m.commentInput.Blur()
				return m, nil
			case "esc":
				return m, func() tea.Msg {
					return messages.SwitchToDetailMsg{Issue: m.issue}
				}
			default:
				var cmd tea.Cmd
				m.commentInput, cmd = m.commentInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 5
			if m.focusIndex == 0 {
				m.commentInput.Focus()
			}
			return m, nil
		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = 4
			}
			if m.focusIndex == 0 {
				m.commentInput.Focus()
			}
			return m, nil
		case "enter", " ":
			switch m.focusIndex {
			case 1:
				m.useBranchName = !m.useBranchName
			case 2:
				m.planMode = !m.planMode
			case 3:
				return m, func() tea.Msg {
					return messages.StartClaudeMsg{
						Issue:     m.issue,
						Comment:   m.commentInput.Value(),
						UseBranch: m.useBranchName,
						PlanMode:  m.planMode,
					}
				}
			case 4:
				return m, func() tea.Msg {
					return messages.StartClaudeMsg{
						Issue:        m.issue,
						Comment:      m.commentInput.Value(),
						UseBranch:    true,
						CheckoutOnly: true,
					}
				}
			}
			return m, nil
		case "esc":
			return m, func() tea.Msg {
				return messages.SwitchToDetailMsg{Issue: m.issue}
			}
		}
	}

	return m, nil
}

func (m StartWorkModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(styles.TitleStyle.Render("Start Working on "+m.issue.Identifier) + "\n\n")
	s.WriteString(styles.SubtitleStyle.Render(m.issue.Title) + "\n\n")

	// Comment input
	commentStyle := styles.InputStyle
	if m.focusIndex == 0 {
		commentStyle = styles.FocusedInputStyle
	}
	s.WriteString(styles.DetailLabelStyle.Render("Comment:") + "\n")
	s.WriteString(commentStyle.Render(m.commentInput.View()) + "\n\n")

	// Checkboxes
	s.WriteString(m.renderCheckbox("Use Linear branch name", m.useBranchName, m.focusIndex == 1))
	if m.issue.BranchName != "" {
		s.WriteString(styles.SubtitleStyle.Render(fmt.Sprintf("  (%s)", m.issue.BranchName)))
	}
	s.WriteString("\n")
	s.WriteString(m.renderCheckbox("Start in plan mode", m.planMode, m.focusIndex == 2) + "\n\n")

	// Buttons
	startBtnStyle := styles.ButtonStyle
	if m.focusIndex == 3 {
		startBtnStyle = styles.ActiveButtonStyle
	}
	checkoutBtnStyle := styles.ButtonStyle
	if m.focusIndex == 4 {
		checkoutBtnStyle = styles.ActiveButtonStyle
	}
	s.WriteString(startBtnStyle.Render("Start Claude") + "  " + checkoutBtnStyle.Render("Checkout Only") + "\n")

	// Error
	if m.err != nil {
		s.WriteString("\n" + styles.ErrorStyle.Render(m.err.Error()))
	}

	// Help
	s.WriteString(styles.HelpStyle.Render("\ntab/arrows: navigate • space/enter: toggle/select • esc: back"))

	return s.String()
}

func (m StartWorkModel) renderCheckbox(label string, checked bool, focused bool) string {
	checkbox := "[ ]"
	if checked {
		checkbox = styles.CheckboxCheckedStyle.Render("[x]")
	}

	labelStyle := styles.CheckboxStyle
	if focused {
		labelStyle = styles.SelectedItemStyle
	}

	cursor := "  "
	if focused {
		cursor = styles.CursorStyle.Render("> ")
	}

	return cursor + checkbox + " " + labelStyle.Render(label)
}

func (m StartWorkModel) Issue() linear.Issue {
	return m.issue
}

func (m StartWorkModel) Comment() string {
	return m.commentInput.Value()
}

func (m StartWorkModel) UseBranchName() bool {
	return m.useBranchName
}

func (m StartWorkModel) PlanMode() bool {
	return m.planMode
}
