package views

import (
	"fmt"
	"sort"
	"strings"

	"linc/internal/config"
	"linc/internal/tui/messages"
	"linc/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

type SettingsModel struct {
	cfg              *config.Config
	workspace        *config.Workspace
	providers        []string
	currentProvider  string
	providerCursor   int
	editingProvider  bool
	saved            bool
	err              error
}

func NewSettingsModel(cfg *config.Config, workspace *config.Workspace, providers []string) SettingsModel {
	// Sort providers for consistent display
	sort.Strings(providers)

	currentProvider := cfg.GetProvider()

	// Find cursor position for current provider
	cursorIdx := 0
	for i, p := range providers {
		if p == currentProvider {
			cursorIdx = i
			break
		}
	}

	return SettingsModel{
		cfg:             cfg,
		workspace:       workspace,
		providers:       providers,
		currentProvider: currentProvider,
		providerCursor:  cursorIdx,
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return nil
}

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle provider editing mode
		if m.editingProvider {
			return m.handleProviderInput(msg)
		}

		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg {
				return messages.SwitchToListMsg{}
			}

		case "p":
			// Enter provider selection mode
			m.editingProvider = true
			m.saved = false
			return m, nil
		}

	case messages.SettingsSavedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.saved = true
			m.err = nil
		}
		return m, nil
	}

	return m, nil
}

func (m SettingsModel) handleProviderInput(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.providerCursor > 0 {
			m.providerCursor--
		}
	case "down", "j":
		if m.providerCursor < len(m.providers)-1 {
			m.providerCursor++
		}
	case "enter":
		if len(m.providers) > 0 {
			selectedProvider := m.providers[m.providerCursor]
			m.currentProvider = selectedProvider
			m.editingProvider = false
			// Auto-save on selection
			return m, m.saveSettings(selectedProvider)
		}
		m.editingProvider = false
		return m, nil
	case "esc":
		// Cancel - restore cursor to current provider
		for i, p := range m.providers {
			if p == m.currentProvider {
				m.providerCursor = i
				break
			}
		}
		m.editingProvider = false
		return m, nil
	}
	return m, nil
}

func (m SettingsModel) saveSettings(provider string) tea.Cmd {
	return func() tea.Msg {
		if err := m.cfg.SetProvider(provider); err != nil {
			return messages.SettingsSavedMsg{Err: err}
		}
		return messages.SettingsSavedMsg{}
	}
}

func (m SettingsModel) View() string {
	// Provider selection mode - show dedicated view
	if m.editingProvider {
		return m.renderProviderSelectView()
	}

	var s strings.Builder

	// Title
	s.WriteString(styles.TitleStyle.Render("Settings") + "\n\n")

	// Current Workspace section
	s.WriteString(styles.DetailLabelStyle.Render("Workspace") + "\n")
	if m.workspace != nil {
		s.WriteString(fmt.Sprintf("  Name: %s\n", styles.DetailValueStyle.Render(m.workspace.Name)))
		s.WriteString(fmt.Sprintf("  ID:   %s\n", styles.SubtitleStyle.Render(m.workspace.ID)))
		if m.workspace.DefaultTeamID != "" {
			s.WriteString(fmt.Sprintf("  Default Team: %s\n", styles.SubtitleStyle.Render(m.workspace.DefaultTeamID)))
		}
	} else {
		s.WriteString(styles.SubtitleStyle.Render("  No workspace selected") + "\n")
	}
	s.WriteString("\n")

	// Provider section
	s.WriteString(styles.DetailLabelStyle.Render("AI Provider") + "\n")
	s.WriteString(fmt.Sprintf("  %s\n", styles.DetailValueStyle.Render(m.currentProvider)))
	s.WriteString("\n")

	// Config file location
	s.WriteString(styles.DetailLabelStyle.Render("Config File") + "\n")
	s.WriteString(styles.SubtitleStyle.Render("  ~/.linc/config.json") + "\n")

	// Status messages
	if m.saved {
		s.WriteString("\n" + styles.CheckboxCheckedStyle.Render("Settings saved!"))
	}
	if m.err != nil {
		s.WriteString("\n" + styles.ErrorStyle.Render(m.err.Error()))
	}

	// Help
	s.WriteString(styles.HelpStyle.Render("\np: change provider • esc/q: back"))

	return s.String()
}

func (m SettingsModel) renderProviderSelectView() string {
	var s strings.Builder

	s.WriteString(styles.TitleStyle.Render("Select AI Provider") + "\n\n")

	for i, provider := range m.providers {
		cursor := "  "
		if m.providerCursor == i {
			cursor = styles.CursorStyle.Render("> ")
		}

		label := provider
		if provider == m.currentProvider {
			label += " (current)"
		}

		if m.providerCursor == i {
			s.WriteString(cursor + styles.SelectedItemStyle.Render(label) + "\n")
		} else {
			s.WriteString(cursor + label + "\n")
		}
	}

	s.WriteString(styles.HelpStyle.Render("\nj/k: navigate • enter: select • esc: cancel"))

	return s.String()
}
