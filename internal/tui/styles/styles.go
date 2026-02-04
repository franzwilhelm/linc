package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	PrimaryColor   = lipgloss.Color("208") // Orange
	secondaryColor = lipgloss.Color("241") // Gray
	successColor   = lipgloss.Color("42")  // Green
	errorColor     = lipgloss.Color("196") // Red

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginBottom(1)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(PrimaryColor).
				Bold(true)

	// Row styles with borders
	// Unselected: bottom + left/right borders (invisible) for consistent spacing
	RowStyle = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder()).
			BorderTop(false).
			Padding(0, 1)

	// First row (unselected): top + bottom + left/right borders (invisible)
	FirstRowStyle = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder()).
			Padding(0, 1)

	// Row above selected: left/right only (no bottom, selected row's top provides separation)
	RowAboveSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder()).
				BorderTop(false).
				BorderBottom(false).
				Padding(0, 1)

	// First row above selected: top + left/right (no bottom)
	FirstRowAboveSelectedStyle = lipgloss.NewStyle().
					Border(lipgloss.HiddenBorder()).
					BorderBottom(false).
					Padding(0, 1)

	// Selected: all borders visible
	SelectedRowStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("250")). // Lighter gray
				Padding(0, 1)

	CursorStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Underline(true).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(secondaryColor).
			MarginBottom(1)

	// Issue styles
	IssueIdentifierStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	IssueTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	IssueStateStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	IssueLabelStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	// Detail view styles
	DetailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(PrimaryColor).
				MarginBottom(1)

	DetailLabelStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Width(12)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

	DetailDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				MarginTop(1).
				MarginBottom(1)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1).
			Background(secondaryColor).
			Foreground(lipgloss.Color("255"))

	ActiveButtonStyle = lipgloss.NewStyle().
				Padding(0, 2).
				MarginRight(1).
				Background(PrimaryColor).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Padding(0, 1)

	// Checkbox styles
	CheckboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	CheckboxCheckedStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginTop(1)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Filter styles
	FilterPromptStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	FilterInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

	// Branch box styles
	BranchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1).
			MarginBottom(1)

	BranchLabelStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	BranchNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	// Logo styles
	LogoSwordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // Green like the logo

	LogoTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)
)

func StateStyle(color string) lipgloss.Style {
	return IssueStateStyle.Background(lipgloss.Color(color))
}

func LabelStyle(color string) lipgloss.Style {
	return IssueLabelStyle.Background(lipgloss.Color(color))
}

func PrimaryColorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(PrimaryColor)
}
