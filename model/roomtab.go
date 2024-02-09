package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/design"
	"github.com/onfirebyte/chatt/signal"
)

type RoomListResult signal.Result[[]string]

type RoomListTab struct {
	title   string
	width   int
	height  int
	focus   bool
	idx     int
	loading bool

	textInput textinput.Model
	inputMode bool

	offset int

	data      []string
	error     error
	fetchFunc func() ([]string, error)
}

func NewRoomListTabModel(name string, fetchFunc func() ([]string, error)) RoomListTab {
	ti := textinput.New()
	ti.Placeholder = "Create a room..."
	ti.Blur()
	ti.CharLimit = 16
	ti.Width = 14

	return RoomListTab{
		title:     name,
		fetchFunc: fetchFunc,
		textInput: ti,
	}
}

func (m RoomListTab) Init() tea.Cmd {
	return func() tea.Msg {
		if m.fetchFunc == nil {
			return RoomListResult{
				Value: []string{},
				Err:   nil,
			}
		}

		res, err := m.fetchFunc()
		return RoomListResult{
			Value: res,
			Err:   err,
		}
	}
}

func (m RoomListTab) Update(msg tea.Msg) (RoomListTab, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case signal.Size:
		m.width = msg.Width
		m.height = msg.Height
	case signal.HomeTabSelected:
		m.focus = bool(msg)
		if m.focus {
			m.textInput.Focus()
		} else {
			m.inputMode = false
			m.textInput.Blur()
			m.textInput.SetValue("")
		}
	case tea.KeyMsg:
		if msg.String() == "r" && m.focus && m.fetchFunc != nil {
			m.loading = true
			return m, func() tea.Msg {
				res, err := m.fetchFunc()
				return RoomListResult{
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
					return RoomListResult{
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
		case "n":
			if m.focus && !m.inputMode {
				m.inputMode = true
				m.textInput.Blur()
			}

		case "esc":
			if m.focus && m.inputMode {
				m.inputMode = false
				m.textInput.Blur()
				m.textInput.SetValue("")
			}
		case "enter":
			if !m.focus {
				break
			}
			var val string
			if m.inputMode {
				val = m.textInput.Value()
				m.inputMode = false
				m.textInput.Blur()
				m.textInput.SetValue("")
			} else {
				val = m.data[m.idx+m.offset]
			}

			cmds = append(cmds,
				func() tea.Msg {
					return signal.Connect{
						IsRoom: true,
						Value:  val,
					}
				})
		}

	case RoomListResult:

		m.loading = false
		m.data = msg.Value
		m.error = msg.Err

	case signal.Refetch:
		if msg == "all" && m.fetchFunc != nil {
			m.loading = true
			cmds = append(cmds, func() tea.Msg {
				res, err := m.fetchFunc()
				return RoomListResult{
					Value: res,
					Err:   err,
				}
			})
		}
	}

	if m.inputMode {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		cmd = m.textInput.Focus()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m RoomListTab) View() string {
	var tabStyle lip.Style

	if m.focus {
		tabStyle = design.ActiveTab
	} else {
		tabStyle = design.Tab
	}

	items := make([]string, max(m.height-3, 2))

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

	if m.inputMode {
		if len(items) == 1 {
			items = append(items, m.textInput.View())
		} else {
			items[len(items)-1] = m.textInput.View()
		}
	}

	return tabStyle.
		Width(m.width - 2).
		Height(m.height - 2).
		MaxHeight(m.height).
		Render(lip.JoinVertical(lip.Top, items...))
}
