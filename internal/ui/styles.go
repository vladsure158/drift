package ui

import "github.com/charmbracelet/lipgloss"

// ─── Color Scheme ───────────────────────────────

type ColorScheme struct {
	Name          string
	Accent        lipgloss.Color
	Green         lipgloss.Color
	Yellow        lipgloss.Color
	Red           lipgloss.Color
	Gray          lipgloss.Color
	DimGray       lipgloss.Color
	White         lipgloss.Color
	SelectedBg    lipgloss.Color
	DimSelectedBg lipgloss.Color
}

var Themes = []ColorScheme{
	{
		Name:          "cyan",
		Accent:        lipgloss.Color("6"),
		Green:         lipgloss.Color("2"),
		Yellow:        lipgloss.Color("3"),
		Red:           lipgloss.Color("1"),
		Gray:          lipgloss.Color("8"),
		DimGray:       lipgloss.Color("240"),
		White:         lipgloss.Color("15"),
		SelectedBg:    lipgloss.Color("237"),
		DimSelectedBg: lipgloss.Color("236"),
	},
	{
		Name:          "claude",
		Accent:        lipgloss.Color("173"), // warm terracotta
		Green:         lipgloss.Color("108"), // sage green
		Yellow:        lipgloss.Color("179"), // warm gold
		Red:           lipgloss.Color("167"), // coral
		Gray:          lipgloss.Color("245"),
		DimGray:       lipgloss.Color("240"),
		White:         lipgloss.Color("253"),
		SelectedBg:    lipgloss.Color("237"),
		DimSelectedBg: lipgloss.Color("236"),
	},
	{
		Name:          "green",
		Accent:        lipgloss.Color("34"),  // matrix green
		Green:         lipgloss.Color("10"),  // bright green
		Yellow:        lipgloss.Color("142"), // olive
		Red:           lipgloss.Color("1"),
		Gray:          lipgloss.Color("8"),
		DimGray:       lipgloss.Color("240"),
		White:         lipgloss.Color("15"),
		SelectedBg:    lipgloss.Color("237"),
		DimSelectedBg: lipgloss.Color("236"),
	},
	{
		Name:          "purple",
		Accent:        lipgloss.Color("135"), // medium purple
		Green:         lipgloss.Color("2"),
		Yellow:        lipgloss.Color("3"),
		Red:           lipgloss.Color("168"), // rose
		Gray:          lipgloss.Color("8"),
		DimGray:       lipgloss.Color("240"),
		White:         lipgloss.Color("15"),
		SelectedBg:    lipgloss.Color("237"),
		DimSelectedBg: lipgloss.Color("236"),
	},
	{
		Name:          "mono",
		Accent:        lipgloss.Color("252"), // bright gray
		Green:         lipgloss.Color("250"),
		Yellow:        lipgloss.Color("248"),
		Red:           lipgloss.Color("245"),
		Gray:          lipgloss.Color("243"),
		DimGray:       lipgloss.Color("238"),
		White:         lipgloss.Color("255"),
		SelectedBg:    lipgloss.Color("237"),
		DimSelectedBg: lipgloss.Color("236"),
	},
}

// ─── Style Variables (updated by ApplyTheme) ────

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

// ApplyTheme updates all style variables to match the given color scheme.
func ApplyTheme(t ColorScheme) {
	Cyan = t.Accent
	Green = t.Green
	Yellow = t.Yellow
	Red = t.Red
	Gray = t.Gray
	White = t.White
	DimGray = t.DimGray

	StatusColors = map[string]lipgloss.Color{
		"active":    t.Accent,
		"done":      t.Green,
		"idea":      t.Yellow,
		"paused":    t.Gray,
		"abandoned": t.DimGray,
	}

	Bold = lipgloss.NewStyle().Bold(true)
	Dim = lipgloss.NewStyle().Foreground(t.DimGray)
	Accent = lipgloss.NewStyle().Foreground(t.Accent)
	AccentB = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	GreenS = lipgloss.NewStyle().Foreground(t.Green)
	YellowS = lipgloss.NewStyle().Foreground(t.Yellow)
	RedS = lipgloss.NewStyle().Foreground(t.Red)
	GrayS = lipgloss.NewStyle().Foreground(t.Gray)
	DimGrayS = lipgloss.NewStyle().Foreground(t.DimGray)

	SelectedRow = lipgloss.NewStyle().
		Background(t.SelectedBg).
		Foreground(t.White).
		Bold(true)
	DimSelectedRow = lipgloss.NewStyle().
		Background(t.DimSelectedBg).
		Foreground(t.DimGray)

	ActiveItem = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	ListBorderActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent)
	ListBorderDim = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.DimGray)
	DetailBorderActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent)
	DetailBorderDim = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.DimGray)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	HelpKey = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	HelpLabel = lipgloss.NewStyle().Foreground(t.DimGray)

	TagStyle = lipgloss.NewStyle().Foreground(t.Accent)
	FlashStyle = lipgloss.NewStyle().Foreground(t.Green)
}
