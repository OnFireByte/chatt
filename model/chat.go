package model

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/design"
	"github.com/onfirebyte/chatt/signal"
)

type (
	chatResult signal.Result[[]string]
	chatError  error
	chatConn   *websocket.Conn
)

type chatMessage struct {
	User    string `json:"user"`
	Message string `json:"data"`
}

type Chat struct {
	title   string
	width   int
	height  int
	focus   bool
	loading bool

	offset int

	connection *websocket.Conn

	data  []chatMessage
	error error

	textInput textinput.Model
}

func NewChatModel(name string) Chat {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Blur()
	ti.CharLimit = 128

	return Chat{
		title:     name,
		textInput: ti,
		loading:   false,
	}
}

func ConnectWS(oldConn *websocket.Conn, data signal.Connect) tea.Cmd {
	return func() tea.Msg {
		u, err := url.Parse(common.URL)
		u.Scheme = "ws"
		u.Path = "/ws"

		if err != nil {
			return func() tea.Msg {
				return chatError(err)
			}
		}

		q := u.Query()
		q.Set("senderUserName", common.UserName)
		if data.IsRoom {
			roomName := data.Value
			if data.Password != "" {
				roomName = roomName + ":" + data.Password
			}
			q.Set("roomName", url.QueryEscape(roomName))
		} else {
			q.Set("recvUserName", url.QueryEscape(data.Value))
		}

		u.RawQuery = q.Encode()

		header := http.Header{}
		header.Set("Authorization", "Bearer "+common.Token)

		log.Printf("connecting to %s", u.String())
		c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
		if err != nil {
			log.Println("dial error:", err)
			return func() tea.Msg {
				return chatError(err)
			}
		}

		return chatConn(c)
	}
}

func (m *Chat) ReadMessage() tea.Msg {
	c := m.connection
	_, message, err := m.connection.ReadMessage()

	// It is possible that connection on model is changed while waiting
	if c != m.connection {
		return nil
	}

	if err != nil {
		c.Close()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			return nil
		}
		return chatError(err)
	}
	var data chatMessage
	err = json.Unmarshal(message, &data)
	if err != nil {
		return chatError(err)
	}

	return data
}

func (m *Chat) Init() tea.Cmd {
	return nil
}

func (m *Chat) Update(msg tea.Msg) (*Chat, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case signal.Size:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = m.width - 8
	case signal.HomeTabSelected:
		m.focus = bool(msg)
		if m.focus {
			m.textInput.Focus()
		} else {
			m.textInput.Blur()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if m.focus {
				m.offset--
				if m.offset < 0 {
					m.offset = 0
				}
			}
		case "up":
			if m.focus {
				m.offset++
			}
		case "enter":
			if m.focus && m.connection != nil && !m.loading {
				val := m.textInput.Value()
				if val == "" {
					break
				}
				m.textInput.SetValue("")
				err := m.connection.WriteMessage(websocket.TextMessage, []byte(val))
				if err != nil {
					cmds = append(cmds, func() tea.Msg {
						return chatError(err)
					})
				}
			}
		default:
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)

		}

	case signal.Connect:
		m.data = nil
		m.error = nil
		m.loading = true
		name := strings.Split(msg.Value, ":")[0]
		if msg.IsRoom {
			m.title = "Room: " + name
		} else {
			m.title = "Chat with: " + msg.Value
		}
		cmd := ConnectWS(m.connection, msg)
		cmds = append(cmds, cmd)

	case chatConn:
		m.loading = false
		m.connection = msg

		cmds = append(cmds,
			func() tea.Msg {
				return signal.Refetch("all")
			},
			m.ReadMessage)

	case chatMessage:
		m.data = append(m.data, msg)
		cmds = append(cmds, m.ReadMessage)
	case chatError:
		m.error = msg
	}

	return m, tea.Batch(cmds...)
}

func (m *Chat) View() string {
	var tabStyle lip.Style

	if m.focus {
		tabStyle = design.ActiveTab
	} else {
		tabStyle = design.Tab
	}

	text := []string{}

	prevUser := ""
	for _, v := range m.data {
		rendered := lip.NewStyle().
			Border(lip.RoundedBorder()).
			Padding(0, 1).
			MaxWidth(m.width - 4).
			Render(v.Message)

		if prevUser != v.User {
			rendered = lipgloss.JoinVertical(lip.Top,
				lip.NewStyle().Foreground(lip.Color("205")).Bold(true).Render(v.User),
				rendered,
			)
		}
		prevUser = v.User
		text = append(text, strings.Split(rendered, "\n")...)
	}

	contentHeight := m.height - 5
	if contentHeight < 0 {
		contentHeight = 0
	}

	if len(text) > contentHeight {
		if m.offset > len(text)-m.height+5 {
			m.offset = len(text) - m.height + 5
		}
		text = text[len(text)-m.height+5-m.offset : len(text)-m.offset]
	}

	if m.error != nil {
		text = []string{design.ErrorText.Render(m.error.Error())}
	}

	res := make([]string, contentHeight+2)
	title := m.title

	if m.loading {
		title = title + " " + common.Spinner.View()
	}

	res[0] = design.ListHeader.Width(m.width - 4).Render(
		title,
	)

	for i, v := range text {
		res[i+1] = v
	}

	if m.focus && !m.loading && m.connection != nil {
		res[len(res)-1] = m.textInput.View()
	}

	return tabStyle.
		Width(m.width - 2).
		Height(m.height - 2).
		MaxHeight(m.height).
		Render(lip.JoinVertical(lip.Top, res...))
}
