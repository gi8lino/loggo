package tui

import "charm.land/lipgloss/v2"

var (
	statusStyle   = lipgloss.NewStyle().Reverse(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	matchStyle    = lipgloss.NewStyle().Background(lipgloss.Color("11")).Foreground(lipgloss.Color("0"))
	badgeStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11")).Padding(0, 1)
	headerStyle   = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("7"))
	liveStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("10")).Padding(0, 1)
	frozenStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11")).Padding(0, 1)
	eofStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("8")).Padding(0, 1)
)

// colorStyle returns a lipgloss style for a named color.
func colorStyle(color string) lipgloss.Style {
	switch color {
	case "black":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
	case "red":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	case "green":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case "yellow":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	case "blue":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	case "magenta":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	case "cyan":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	case "white":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	case "gray", "grey":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	case "bold":
		return lipgloss.NewStyle().Bold(true)
	case "dim":
		return lipgloss.NewStyle().Faint(true)
	default:
		return lipgloss.NewStyle()
	}
}
