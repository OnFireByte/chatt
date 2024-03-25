package model

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/design"
	"github.com/onfirebyte/chatt/signal"
)

type UserListResult signal.Result[[]string]

type UserListTab struct {
	title   string
	width   int
	height  int
	focus   bool
	idx     int
	loading bool

	offset int

	data      []string
	error     error
	fetchFunc func() ([]string, error)
}

func NewUserListTabModel(name string, fetchFunc func() ([]string, error)) UserListTab {
	return UserListTab{
		title:     name,
		fetchFunc: fetchFunc,
	}
}

func (m UserListTab) Init() tea.Cmd {
	return func() tea.Msg {
		if m.fetchFunc == nil {
			return UserListResult{
				Value: []string{},
				Err:   nil,
			}
		}

		res, err := m.fetchFunc()
		return UserListResult{
			Value: res,
			Err:   err,
		}
	}
}

func (m UserListTab) Update(msg tea.Msg) (UserListTab, tea.Cmd) {
	switch msg := msg.(type) {
	case signal.Size:
		m.width = msg.Width
		m.height = msg.Height
	case signal.HomeTabSelected:
		m.focus = bool(msg)
	case tea.KeyMsg:
		if msg.String() == "r" && m.focus && m.fetchFunc != nil {
			m.loading = true
			return m, func() tea.Msg {
				res, err := m.fetchFunc()
				return UserListResult{
					Value: res,
					Err:   err,
				}
			}
		}

		switch msg.String() {
		case "r":
			if m.focus && m.fetchFunc != nil {
				return m, func() tea.Msg {
					res, err := m.fetchFunc()
					return UserListResult{
						Value: res,
						Err:   err,
					}
				}
			}
		case "down":
			if m.focus {
				if m.idx < min(len(m.data)-1, m.height-5) {
					m.idx++
				} else if m.idx+m.offset < len(m.data)-1 {
					m.offset++
				}
			}
		case "up":
			if m.focus {
				if m.idx > 0 {
					m.idx--
				} else if m.offset > 0 {
					m.offset--
				}
			}
		case "enter":
			if m.focus && m.idx >= 0 && m.idx < len(m.data) {
				return m, func() tea.Msg {
					return signal.Connect{
						IsRoom: false,
						Value:  m.data[m.idx+m.offset],
					}
				}
			}

		}

	case UserListResult:
		m.loading = false
		m.data = msg.Value
		m.error = msg.Err

	case signal.Refetch:
		if msg == "all" && m.fetchFunc != nil {
			m.loading = true
			return m, func() tea.Msg {
				res, err := m.fetchFunc()
				return UserListResult{
					Value: res,
					Err:   err,
				}
			}
		}
	}

	return m, nil
}

func (m UserListTab) View() string {
	var tabStyle lip.Style

	if m.focus {
		tabStyle = design.ActiveTab
	} else {
		tabStyle = design.Tab
	}

	items := make([]string, max(min(len(m.data)+1, m.height-3), 2))

	title := m.title
	if m.loading {
		title = fmt.Sprintf("%s %s", title, common.Spinner.View())
	}

	items[0] = design.ListHeader.Width(m.width - 4).Render(title)
	maxLen := min(len(m.data), m.height-4)
	if m.error != nil {
		items[1] = design.ErrorText.Render(m.error.Error())
	} else {
		for i := 0; i < maxLen; i++ {
			v := m.data[i+m.offset]
			if i == m.idx && m.focus {
				v = lip.NewStyle().Foreground(design.Special).Bold(true).Render(fmt.Sprintf("▶ %s", v))
			}
			items[i+1] = v
		}
	}

	return tabStyle.
		Width(m.width - 2).
		Height(m.height - 2).
		MaxHeight(m.height).
		Render(lip.JoinVertical(lip.Top, items...))
}
