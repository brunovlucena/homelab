// FutBoss AI - TUI Styles (k9s-inspired)
// Author: Bruno Lucena (bruno@lucena.cloud)

package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#00D4AA")
	secondaryColor = lipgloss.Color("#7B61FF")
	accentColor    = lipgloss.Color("#FFD700")
	errorColor     = lipgloss.Color("#FF6B6B")
	successColor   = lipgloss.Color("#4ECB71")
	mutedColor     = lipgloss.Color("#626262")
	bgColor        = lipgloss.Color("#1A1A2E")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(bgColor).
			Foreground(lipgloss.Color("#FFFFFF"))

	// Title bar (k9s style)
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(lipgloss.Color("#0F0F1A")).
			Padding(0, 1).
			MarginBottom(1)

	// Header row for tables
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(mutedColor)

	// Selected item
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(primaryColor).
			Padding(0, 1)

	// Normal item
	ItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(lipgloss.Color("#0F0F1A")).
			Padding(0, 1)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Stats box
	StatsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1).
			MarginRight(1)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Score display (for matches)
	ScoreStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			Background(lipgloss.Color("#2A2A4A")).
			Padding(0, 2)

	// Player position styles
	GKStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	DEFStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4169E1"))
	MIDStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
	ATKStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500"))

	// Attribute bar
	AttrHighStyle = lipgloss.NewStyle().Foreground(successColor)
	AttrMidStyle  = lipgloss.NewStyle().Foreground(accentColor)
	AttrLowStyle  = lipgloss.NewStyle().Foreground(errorColor)
)

// GetPositionStyle returns style based on player position
func GetPositionStyle(position string) lipgloss.Style {
	switch position {
	case "GK":
		return GKStyle
	case "CB", "LB", "RB":
		return DEFStyle
	case "CDM", "CM", "CAM":
		return MIDStyle
	case "LW", "RW", "ST":
		return ATKStyle
	default:
		return ItemStyle
	}
}

// GetAttrStyle returns style based on attribute value
func GetAttrStyle(value int) lipgloss.Style {
	if value >= 80 {
		return AttrHighStyle
	} else if value >= 60 {
		return AttrMidStyle
	}
	return AttrLowStyle
}

// RenderAttrBar renders a visual bar for attribute value
func RenderAttrBar(value int) string {
	filled := value / 10
	empty := 10 - filled
	style := GetAttrStyle(value)
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	return style.Render(bar)
}
