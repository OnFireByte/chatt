package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/model"
	"github.com/onfirebyte/chatt/signal"
)

// SessionState is used to track which model is focused
type SessionState uint

const (
	createUserState SessionState = iota
	mainMenuState   SessionState = iota
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	spinnerStyle = lip.NewStyle().Foreground(lip.Color("69"))
	helpStyle    = lip.NewStyle().Foreground(lip.Color("241"))
)

type mainModel struct {
	state           SessionState
	createUserModel model.CreateUser
	homeModel       model.Home
}

func newModel() mainModel {
	m := mainModel{
		state:           createUserState,
		createUserModel: model.NewCreateUserModel(),
		homeModel:       model.NewHomeModel(),
	}

	return m
}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start

	return tea.Batch(
		common.Spinner.Tick,
		m.createUserModel.Init(),
		m.homeModel.Init(),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c":
			m.homeModel.Update(tea.QuitMsg{})
			return m, tea.Quit
		}
		switch m.state {
		case createUserState:
			m.createUserModel, cmd = m.createUserModel.Update(msg)
			cmds = append(cmds, cmd)
		case mainMenuState:
			m.homeModel, cmd = m.homeModel.Update(msg)
			cmds = append(cmds, cmd)
		}

	case spinner.TickMsg:
		s, cmd := common.Spinner.Update(msg)
		common.Spinner = s
		cmds = append(cmds, cmd)

	case signal.UserInfo:
		common.UserName = msg.Name
		common.Token = msg.Token

		m.state = mainMenuState
		m.homeModel, cmd = m.homeModel.Update(signal.Refetch("all"))
		cmds = append(cmds, cmd)

	}

	if _, ok := msg.(tea.KeyMsg); !ok {
		m.createUserModel, cmd = m.createUserModel.Update(msg)
		cmds = append(cmds, cmd)

		m.homeModel, cmd = m.homeModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	switch m.state {
	case createUserState:
		return m.createUserModel.View()
	case mainMenuState:
		return m.homeModel.View()
	}
	return ""
}

func (m mainModel) currentFocusedModel() string {
	return "spinner"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a URL")
		os.Exit(1)
	}
	rawURL := os.Args[1]
	if rawURL == "" {
		log.Fatal("Please provide a URL")
	}

	if rawURL[:4] != "http" {
		rawURL = "http://" + rawURL
	}

	common.URL = rawURL

	if os.Getenv("DEBUG") != "" {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		log.SetOutput(ioutil.Discard)
	}

	p := tea.NewProgram(newModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
