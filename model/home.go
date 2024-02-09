package model

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/request"
	"github.com/onfirebyte/chatt/signal"
)

type Home struct {
	inited bool

	selectedTab selectedTab

	userTab UserListTab
	roomTab RoomListTab
	chatTab *Chat
}

type selectedTab uint

const (
	userTab selectedTab = iota
	roomTab selectedTab = iota
	chatTab selectedTab = iota
)

var LeftTabWidth = 32

func NewHomeModel() Home {
	chat := NewChatModel("Chat")
	return Home{
		userTab:     NewUserListTabModel("Users", request.GetAllUsers),
		roomTab:     NewRoomListTabModel("Rooms", request.GetAllRooms),
		chatTab:     &chat,
		selectedTab: chatTab,
	}
}

func (m Home) Init() tea.Cmd {
	if m.inited {
		return nil
	}
	m.inited = true
	cmds := []tea.Cmd{}
	cmds = append(cmds,
		m.userTab.Init(),
		m.roomTab.Init(),
		m.chatTab.Init())
	return tea.Batch(cmds...)
}

func (m Home) Update(msg tea.Msg) (Home, tea.Cmd) {
	cmds := []tea.Cmd{}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		smallTabHeight := (msg.Height - 1) / 2
		extraHeight := (msg.Height - 1) % 2
		chatWidth := msg.Width - LeftTabWidth

		m.userTab, cmd = m.userTab.Update(signal.Size{Width: LeftTabWidth, Height: smallTabHeight})
		cmds = append(cmds, cmd)

		m.roomTab, cmd = m.roomTab.Update(signal.Size{Width: LeftTabWidth, Height: smallTabHeight + extraHeight})
		cmds = append(cmds, cmd)

		m.chatTab, cmd = m.chatTab.Update(signal.Size{Width: chatWidth, Height: msg.Height - 1})
		cmds = append(cmds, cmd)

	case signal.Connect:
		m.selectedTab = chatTab
		m.chatTab, cmd = m.chatTab.Update(signal.HomeTabSelected(true))
		cmds = append(cmds, cmd)

		m.userTab, cmd = m.userTab.Update(signal.HomeTabSelected(false))
		cmds = append(cmds, cmd)

		m.roomTab, cmd = m.roomTab.Update(signal.HomeTabSelected(false))
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.selectedTab = (m.selectedTab + 1) % 3

			m.userTab, cmd = m.userTab.Update(signal.HomeTabSelected(m.selectedTab == userTab))
			cmds = append(cmds, cmd)

			m.roomTab, cmd = m.roomTab.Update(signal.HomeTabSelected(m.selectedTab == roomTab))
			cmds = append(cmds, cmd)

			m.chatTab, cmd = m.chatTab.Update(signal.HomeTabSelected(m.selectedTab == chatTab))
			cmds = append(cmds, cmd)
		}
	}

	m.userTab, cmd = m.userTab.Update(msg)
	cmds = append(cmds, cmd)

	m.roomTab, cmd = m.roomTab.Update(msg)
	cmds = append(cmds, cmd)

	m.chatTab, cmd = m.chatTab.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Home) View() string {
	title := lip.NewStyle().Foreground(lip.Color("205")).Render(fmt.Sprintf("Welcome %s", common.UserName))
	leftTab := lip.JoinVertical(lip.Bottom,
		m.userTab.View(),
		m.roomTab.View())
	content := lip.JoinHorizontal(lip.Left, leftTab, m.chatTab.View())
	return lip.JoinVertical(lip.Top, title, content)
}
