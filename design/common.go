package design

import lip "github.com/charmbracelet/lipgloss"

var (
	Subtle    = lip.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lip.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	Special   = lip.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	Error     = lip.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}

	ErrorText = lip.NewStyle().Foreground(Error)

	Tab = lip.NewStyle().
		Border(lip.RoundedBorder(), true).
		BorderForeground(Subtle).
		Padding(0, 1)

	ActiveTab = Tab.Copy().Bold(true).BorderForeground(Highlight)

	ListHeader = lip.NewStyle().
			BorderStyle(lip.NormalBorder()).
			BorderBottom(true).
			BorderForeground(Subtle)
)
