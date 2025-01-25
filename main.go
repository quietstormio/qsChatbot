package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	titan "github.com/quietstormio/qsChatbot/bedrock"
)

const gap = "\n\n"

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

var titanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0097b2")).Bold(true) // Blue color

type (
	errMsg error
)

type responseMsg struct {
	response string
	err      error
}

type model struct {
	viewport    viewport.Model
	messages    []string
	answers     []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Titan Chat!
Enter a prompt and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		answers:     []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#f0aa8d")).Bold(true),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func processOutputCmd(userMessage string) tea.Cmd {
	return func() tea.Msg {
		response, err := titan.ProcessOutput(userMessage)
		return responseMsg{response: response, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			userMessage := m.textarea.Value()
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+userMessage)

			// Update the viewport content immediately
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()

			// Return the command to process the output
			return m, processOutputCmd(userMessage)
		}
	case responseMsg:
		if msg.err != nil {
			log.Fatal(msg.err)
		}
		m.messages = append(m.messages, titanStyle.Render("Titan: ")+msg.response)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.viewport.GotoBottom()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
