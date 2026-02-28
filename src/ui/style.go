package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	catOrange = lipgloss.Color("#FF9D00")
	catBrown  = lipgloss.Color("#8B5A2B")
	dangerRed = lipgloss.Color("#FF0000")
	highlight = lipgloss.Color("#00FF00")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(catOrange).
			Bold(true)

	TargetStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)

	ExplanationStyle = lipgloss.NewStyle().
				Foreground(catBrown)

	DangerStyle = lipgloss.NewStyle().
			Foreground(dangerRed).
			Bold(true)

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(catOrange).
				Bold(true)
)
