package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	subtleColor    = lipgloss.Color("#6B7280")
	textColor      = lipgloss.Color("#F9FAFB")
	dimColor       = lipgloss.Color("#9CA3AF")
)

var (
	appStyle      = lipgloss.NewStyle().Padding(1, 2)
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(primaryColor).PaddingBottom(1)
	subtitleStyle = lipgloss.NewStyle().Foreground(dimColor)
	statusStyle   = lipgloss.NewStyle().Foreground(dimColor).PaddingTop(1)
	helpStyle     = lipgloss.NewStyle().Foreground(subtleColor)
	errorStyle    = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(textColor)

	dirStyle          = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	fileStyle         = lipgloss.NewStyle().Foreground(textColor)
	selectedFileStyle = lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
	cursorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)

	progressBarStyle  = lipgloss.NewStyle().Foreground(secondaryColor)
	progressTextStyle = lipgloss.NewStyle().Foreground(dimColor)

	searchStyle       = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(primaryColor).Padding(0, 1)
	searchPromptStyle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	checkboxChecked   = lipgloss.NewStyle().Foreground(secondaryColor).Render("[x]")
	checkboxUnchecked = lipgloss.NewStyle().Foreground(subtleColor).Render("[ ]")
)

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return lipgloss.NewStyle().Foreground(subtleColor).Render(fmt.Sprintf("%d B", bytes))
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return lipgloss.NewStyle().Foreground(subtleColor).Render(
		fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp]),
	)
}
