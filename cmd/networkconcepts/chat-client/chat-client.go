package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
	"github.com/vinh0604/go-network-concepts/internal/chatutils"
)

type globalState struct {
	nick string
	host string
	port int
	sock *net.Conn
}

func main() {
	var err error

	args := flag.Args()
	host := "localhost"
	port := 8080
	if len(args) > 2 {
		host = args[0]
		port, err = strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}
	}

	state := globalState{
		nick: "",
		host: host,
		port: port,
	}
	p := tea.NewProgram(initNickInputModel(&state))

	go func() {
		for {
			if state.sock == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			readBuf := chatutils.ReadBuffer{}
			for {
				payload, err := chatutils.ReadNextMessage(*state.sock, &readBuf)
				if err != nil {
					p.Send(errMsg{err: fmt.Errorf("error reading message from server: %s", err.Error())})
					break
				}

				switch payload.MsgType {
				case chatmodels.MsgTypeChat:
					p.Send(recvMsg{msg: fmt.Sprint(*payload.Nick, ": ", *payload.Msg), isSys: false})
				case chatmodels.MsgTypeJoin:
					p.Send(recvMsg{msg: fmt.Sprint("[", *payload.Nick, " joined the chat]"), isSys: true})
				default:
					p.Send(errMsg{err: fmt.Errorf("unknown message type: %s", payload.MsgType)})
				}
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initNickInputModel(state *globalState) tea.Model {
	return nickInputModel{
		state: state,
		nick:  "",
	}
}

func initialChatViewModel(state *globalState) chatViewModel {
	var err error
	sock, err := net.Dial("tcp", fmt.Sprintf("%s:%d", state.host, state.port))
	if err != nil {
		panic(fmt.Sprint("Error connecting to server: ", err))
	}
	state.sock = &sock

	helloPayload := chatmodels.Payload{
		MsgType: chatmodels.MsgTypeHello,
		Nick:    &state.nick,
	}
	err = sendChat(&sock, helloPayload)
	if err != nil {
		panic(fmt.Sprint("Error sending hello message to server: ", err))
	}

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "| "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return chatViewModel{
		state:         state,
		textarea:      ta,
		messages:      []string{},
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		announceStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		err:           nil,
	}
}

type errMsg struct {
	err error
}

type recvMsg struct {
	msg   string
	isSys bool
}

type chatViewModel struct {
	state         *globalState
	viewport      viewport.Model
	messages      []string
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	receiverStyle lipgloss.Style
	announceStyle lipgloss.Style
	err           error
}

func (m chatViewModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m chatViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.state != nil && m.state.sock != nil {
				(*m.state.sock).Close()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textarea.Value() == "" {
				return m, nil
			}

			msg := m.textarea.Value()
			chatPayload := chatmodels.Payload{
				MsgType: chatmodels.MsgTypeChat,
				Msg:     &msg,
			}
			err := sendChat(m.state.sock, chatPayload)
			if err != nil {
				m.err = fmt.Errorf("error sending chat message: %s", err.Error())
				return m, nil
			}

			m.messages = append(m.messages, m.senderStyle.Render(fmt.Sprint(m.state.nick, ": ", m.textarea.Value())))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case recvMsg:
		if msg.isSys {
			m.messages = append(m.messages, m.announceStyle.Render(msg.msg))
		} else {
			m.messages = append(m.messages, m.receiverStyle.Render(msg.msg))
		}
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil
	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m chatViewModel) View() string {
	return fmt.Sprintf("%s\n\n%s", m.viewport.View(), m.textarea.View()) + "\n\n"
}

type nickInputModel struct {
	state *globalState
	nick  string
	err   error
}

func (m nickInputModel) Init() tea.Cmd {
	return nil
}

func (m nickInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.nick == "" {
				m.err = fmt.Errorf("nick cannot be empty")
			} else {
				m.state.nick = m.nick
				return initialChatViewModel(m.state), nil
			}
		case tea.KeyRunes:
			m.nick += string(msg.Runes)
		case tea.KeyBackspace:
			if len(m.nick) > 0 {
				m.nick = m.nick[:len(m.nick)-1]
			}
		}
	}

	return m, nil
}

func (m nickInputModel) View() string {
	return fmt.Sprintf("Enter your nickname:\n%s\n%s", m.nick, dipslayError(m.err))
}

func dipslayError(err error) string {
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return ""
}

func sendChat(sock *net.Conn, payload chatmodels.Payload) error {
	out, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	outLen := len(out)
	outLenBytes := []byte{
		byte(outLen >> 8),
		byte(outLen & 0xFF),
	}

	_, err = (*sock).Write(append(outLenBytes, out...))
	if err != nil {
		return err
	}
	return nil
}
