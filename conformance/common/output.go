package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for output
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginTop(1).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	GroupStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			MarginTop(1)

	PassStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	FailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Italic(true)

	SummaryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11")).
			MarginTop(1)

	DetailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
)

// ShouldRunGroup determines if a test group should run based on the filter
func ShouldRunGroup(groupName, filter string) bool {
	if filter == "" || filter == "all" {
		return true
	}

	filters := strings.Split(filter, ",")
	for _, f := range filters {
		if strings.TrimSpace(f) == groupName {
			return true
		}
	}

	return false
}
