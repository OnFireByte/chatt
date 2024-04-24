package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/design"
	"github.com/onfirebyte/chatt/dto"
	"github.com/onfirebyte/chatt/signal"
)

type RoomListResult signal.Result[[]dto.Room]

type RoomListTab struct {
	title   string
	width   int
	height  int
	focus   bool
	idx     int
	loading bool

	textInput         textinput.Model
	roomPasswordInput textinput.Model
	inputMode         bool

	offset int

	data      []dto.Room
	error     error
	fetchFunc func() ([]dto.Room, error)
}

func NewRoomListTabModel(name string, fetchFunc func() ([]dto.Room, error)) RoomListTab {
	ti := textinput.New()
	ti.Placeholder = "Create a room..."
	ti.Blur()
	ti.CharLimit = 16
	ti.Width = 14

	pi := textinput.New()
	pi.Placeholder = "Room Password..."
	pi.Blur()
	pi.CharLimit = 16
	pi.Width = 14

	return RoomListTab{
		title:             name,
		fetchFunc:         fetchFunc,
		textInput:         ti,
		roomPasswordInput: pi,
	}
}

func (m RoomListTab) Init() tea.Cmd {
	return func() tea.Msg {
		if m.fetchFunc == nil {
			return RoomListResult{
				Value: nil,
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
			cmd = m.textInput.Focus()
			m.roomPasswordInput.Blur()
			cmds = append(cmds, cmd)
		} else {
			m.inputMode = false
			m.textInput.Blur()
			m.textInput.SetValue("")
		}
	case tea.KeyMsg:

		switch msg.String() {
		case "r":
			if m.focus && m.fetchFunc != nil && !m.inputMode {
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
			if m.focus && !m.inputMode && !m.roomPasswordInput.Focused() {
				m.inputMode = true
				m.textInput.Blur()
			}

		case "esc":
			if m.focus && m.inputMode {
				m.inputMode = false
				m.textInput.Blur()
				m.textInput.SetValue("")
			} else if m.roomPasswordInput.Focused() {
				m.roomPasswordInput.Blur()
				m.roomPasswordInput.SetValue("")
			}
		case "enter":
			if !m.focus {
				break
			}
			var data dto.Room
			var roomName string
			if m.inputMode {
				roomName = m.textInput.Value()
				m.inputMode = false
				m.textInput.Blur()
				m.textInput.SetValue("")
			} else {
				data = m.data[m.idx+m.offset]
				roomName = data.Name
			}
			if !data.Lock {
				cmds = append(cmds,
					func() tea.Msg {
						return signal.Connect{
							IsRoom: true,
							Value:  roomName,
						}
					})
			} else if m.roomPasswordInput.Focused() {
				roomPassword := m.roomPasswordInput.Value()
				m.roomPasswordInput.Blur()
				m.roomPasswordInput.SetValue("")
				cmds = append(cmds,
					func() tea.Msg {
						return signal.Connect{
							IsRoom:   true,
							Value:    roomName,
							Password: roomPassword,
						}
					})
			} else {
				cmd = m.roomPasswordInput.Focus()
				cmds = append(cmds, cmd)
				m.inputMode = false
				m.textInput.Blur()
			}
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
		m.roomPasswordInput.Blur()
		cmds = append(cmds, cmd)
	}

	if m.roomPasswordInput.Focused() {
		m.roomPasswordInput, cmd = m.roomPasswordInput.Update(msg)
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
			data := m.data[i+m.offset]
			v := data.Name
			if data.Lock {
				v += " 🔒"
			}
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

	if m.roomPasswordInput.Focused() {
		items[len(items)-1] = m.roomPasswordInput.View()
	}

	return tabStyle.
		Width(m.width - 2).
		Height(m.height - 2).
		MaxHeight(m.height).
		Render(lip.JoinVertical(lip.Top, items...))
}
