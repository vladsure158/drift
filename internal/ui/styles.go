package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Cyan    = lipgloss.Color("6")
	Green   = lipgloss.Color("2")
	Yellow  = lipgloss.Color("3")
	Red     = lipgloss.Color("1")
	Gray    = lipgloss.Color("8")
	White   = lipgloss.Color("15")
	DimGray = lipgloss.Color("240")

	// Status colors
	StatusColors = map[string]lipgloss.Color{
		"active":    Cyan,
		"done":      Green,
		"idea":      Yellow,
		"paused":    Gray,
		"abandoned": DimGray,
	}

	StatusIcons = map[string]string{
		"active":    "●",
		"done":      "✓",
		"idea":      "○",
		"paused":    "◇",
		"abandoned": "✗",
	}

	// Base styles
	Bold     = lipgloss.NewStyle().Bold(true)
	Dim      = lipgloss.NewStyle().Foreground(DimGray)
	Accent   = lipgloss.NewStyle().Foreground(Cyan)
	AccentB  = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	GreenS   = lipgloss.NewStyle().Foreground(Green)
	YellowS  = lipgloss.NewStyle().Foreground(Yellow)
	RedS     = lipgloss.NewStyle().Foreground(Red)
	GrayS    = lipgloss.NewStyle().Foreground(Gray)
	DimGrayS = lipgloss.NewStyle().Foreground(DimGray)

	// Selected row
	SelectedRow = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(White).
			Bold(true)

	// Selected row when panel is dimmed
	DimSelectedRow = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(DimGray)

	// Active cursor in detail
	ActiveItem = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// Panel borders
	ListBorderActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Cyan)

	ListBorderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DimGray)

	DetailBorderActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Cyan)

	DetailBorderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DimGray)

	// Header
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// Footer / help
	HelpKey   = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	HelpLabel = lipgloss.NewStyle().Foreground(DimGray)

	// Tags
	TagStyle = lipgloss.NewStyle().Foreground(Cyan)

	// Flash message
	FlashStyle = lipgloss.NewStyle().Foreground(Green)
)
