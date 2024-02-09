package common

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var URL string

var Spinner = spinner.New(
	spinner.WithSpinner(spinner.MiniDot),
)

var UserName string

type Model[T any] interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (T, tea.Cmd)
	View() string
}
