package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DomainModel struct {
	textInput textinput.Model
	err       error
	adding    bool
	width     int
	height    int
}

func NewDomainModel() DomainModel {
	ti := textinput.New()
	ti.Placeholder = "Enter domain name (e.g., example.com)"
	ti.Focus()
	ti.CharLimit = 253
	ti.Width = 50

	return DomainModel{
		textInput: ti,
		width:     80,
		height:    24,
	}
}

func (m DomainModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m DomainModel) Update(msg tea.Msg) (DomainModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return m, func() tea.Msg { return "back_to_main" }
		case tea.KeyEnter:
			if m.textInput.Value() != "" && !m.adding {
				m.adding = true
				return m, func() tea.Msg {
					return AddDomainMsg{domain: m.textInput.Value()}
				}
			}
		}
	case DomainAddedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.adding = false
		} else {
			return m, func() tea.Msg { return "back_to_main" }
		}
	}

	// Update text input
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *DomainModel) UpdateSize(width, height int) {
	m.width = width
	m.height = height

	inputWidth := 30
	if width > 40 {
		inputWidth = 50
	}
	if width < 60 {
		inputWidth = width - 10
	}
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.textInput.Width = inputWidth
}

func (m DomainModel) View() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff88")).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(headerStyle.Render("sslcerttop 🔒 Add New Domain"))
	b.WriteString("\n")

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Width(m.width).
		Align(lipgloss.Center)

	if m.width < 84 {
		b.WriteString(separatorStyle.Render("- - - - - - - - - - - - - - - -"))
	} else {
		separatorWidth := m.width - 4
		if separatorWidth < 20 {
			separatorWidth = 20
		}
		if separatorWidth > 80 {
			separatorWidth = 80
		}
		b.WriteString(separatorStyle.Render(strings.Repeat("═", separatorWidth)))
	}
	b.WriteString("\n\n")

	formContentHeight := 4
	if m.err != nil {
		formContentHeight += 2
	}

	topPadding := 1
	if (m.height-formContentHeight-6)/2 > 1 {
		topPadding = (m.height - formContentHeight - 6) / 2
	}
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00bfff")).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center)

	instruction := "Enter a domain name to monitor its SSL certificate:"
	if m.width < 60 {
		instruction = "Enter domain name:"
	}
	b.WriteString(instructionStyle.Render(instruction))
	b.WriteString("\n\n")

	inputStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	var inputSection string
	if m.adding {
		inputSection = "⏳ Adding domain..."
	} else {
		inputSection = m.textInput.View()
	}
	b.WriteString(inputStyle.Render(inputSection))

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff4444")).
			Bold(true).
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("❌ Error: " + m.err.Error()))
	}

	bottomPadding := 0
	if m.height-topPadding-formContentHeight-2 > 0 {
		bottomPadding = m.height - topPadding - formContentHeight - 2
	}
	for i := 0; i < bottomPadding; i++ {
		b.WriteString("\n")
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Width(m.width).
		Align(lipgloss.Center)

	footerText := "[Enter] Add Domain  [Esc] Back  [Alt+Enter] Toggle Screen  [q] Quit"
	if m.width < 80 {
		footerText = "[Enter] Add  [Esc] Back  [q] Quit"
	}
	b.WriteString(footerStyle.Render(footerText))

	return b.String()
}

// Message types for domain operations
type AddDomainMsg struct {
	domain string
}

type DomainAddedMsg struct {
	err error
}
