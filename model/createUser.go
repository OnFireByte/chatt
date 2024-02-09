package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/request"
	"github.com/onfirebyte/chatt/signal"
)

type CreateUser struct {
	textInput textinput.Model
	err       error
	loading   bool
}

type createUserStatus struct {
	error error
}

var errorMsgStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF0000")).
	Bold(true)

func NewCreateUserModel() CreateUser {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return CreateUser{
		textInput: ti,
		err:       nil,
	}
}

func (m CreateUser) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateUser) Update(msg tea.Msg) (CreateUser, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.loading {
				if m.textInput.Value() == "" {
					m.err = fmt.Errorf("name cannot be empty")
				} else {
					m.loading = true
					cmds = append(cmds, func() tea.Msg {
						return createUserStatus{
							error: request.CreateUser(m.textInput.Value()),
						}
					})
				}
			}
		}

		// We handle errors just like any other message
	case createUserStatus:
		m.loading = false
		if msg.error != nil {
			m.err = msg.error
		} else {
			cmds = append(cmds, func() tea.Msg {
				return signal.UserName(m.textInput.Value())
			})
		}
	}

	textInput, cmd := m.textInput.Update(msg)
	m.textInput = textInput
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m CreateUser) View() string {
	var errMessage string
	if m.err != nil {
		errMessage = errorMsgStyle.Render(m.err.Error())
	}
	var input string
	if m.loading {
		input = common.Spinner.View()
	} else {
		input = m.textInput.View()
	}

	return fmt.Sprintf(
		"What's your name?\n\n%s\n\n%s\n\n%s",
		input,
		"(esc to quit)",
		errMessage,
	) + "\n"
}
