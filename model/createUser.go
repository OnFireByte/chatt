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
	userInput     textinput.Model
	passwordInput textinput.Model
	focusOnUser   bool
	err           error
	loading       bool
}

type createUserStatus struct {
	token string
	error error
}

var errorMsgStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF0000")).
	Bold(true)

func NewCreateUserModel() CreateUser {
	userInput := textinput.New()
	userInput.Placeholder = "Username"
	userInput.Focus()
	userInput.CharLimit = 156
	userInput.Width = 20

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password (Optional)"
	passwordInput.CharLimit = 156
	passwordInput.Width = 20
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'

	return CreateUser{
		userInput:     userInput,
		passwordInput: passwordInput,
		err:           nil,
		focusOnUser:   true,
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
				if m.userInput.Value() == "" {
					m.err = fmt.Errorf("name cannot be empty")
				} else {
					m.loading = true
					cmds = append(cmds, func() tea.Msg {
						token, err := request.CreateUser(m.userInput.Value(), m.passwordInput.Value())
						return createUserStatus{
							token: token,
							error: err,
						}
					})
				}
			}
		case tea.KeyTab:
			if m.focusOnUser {
				m.userInput.Blur()
				m.passwordInput.Focus()
			} else {
				m.passwordInput.Blur()
				m.userInput.Focus()
			}
			m.focusOnUser = !m.focusOnUser
		}

		// We handle errors just like any other message
	case createUserStatus:
		m.loading = false
		if msg.error != nil {
			m.err = msg.error
		} else {
			cmds = append(cmds, func() tea.Msg {
				return signal.UserInfo{
					Name:  m.userInput.Value(),
					Token: msg.token,
				}
			})
		}
	}

	textInput, cmd := m.userInput.Update(msg)
	cmds = append(cmds, cmd)
	m.userInput = textInput
	passwordInput, cmd := m.passwordInput.Update(msg)
	cmds = append(cmds, cmd)
	m.passwordInput = passwordInput
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
		input = m.userInput.View()
	}

	var passwordInput string
	if m.loading {
		passwordInput = common.Spinner.View()
	} else {
		passwordInput = m.passwordInput.View()
	}

	return fmt.Sprintf(
		"What's your name?\n\n%s\n%s\n\n%s\n\n%s",
		input,
		passwordInput,
		"(esc to quit)",
		errMessage,
	) + "\n"
}
