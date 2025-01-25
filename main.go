package main

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

// Define styles for the Titan response messages
var titanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0097b2")).Bold(true) // Blue color

// Define custom error message type
type (
	errMsg error
)

// Define a message type for the response from the Titan model
type responseMsg struct {
	response string
	err      error
}

// Define the model for the Bubble Tea program
type model struct {
	viewport    viewport.Model
	messages    []string
	answers     []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

// Initialize the model with default values
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

// Initialize the program
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Command to process the output from the Titan model
func processOutputCmd(userMessage string) tea.Cmd {
	return func() tea.Msg {
		response, err := titan.ProcessOutput(userMessage)
		return responseMsg{response: response, err: err}
	}
}

// Update the model based on incoming messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	// Update the textarea and viewport based on the message
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust the viewport and textarea sizes based on the window size
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
			// Quit the program on Ctrl+C or Esc
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
			// Will always return a responseMsg (or Titan response in other words)
			return m, processOutputCmd(userMessage)
		}
	case responseMsg:
		// Handle the response from the Titan model
		if msg.err != nil {
			log.Fatal(msg.err)
		}
		m.messages = append(m.messages, titanStyle.Render("Titan: ")+msg.response)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.viewport.GotoBottom()

	// Handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

// View for our model
func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func main() {
	// Initialize and run the Bubble Tea program
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
