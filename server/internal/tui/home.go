package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HomeModel struct {
	loading bool
	message string
	ready   bool
	width   int
	height  int
}

func NewHomeModel() HomeModel {
	return HomeModel{
		loading: false,
		message: "Ready!",
		ready:   true,
		width:   80,
		height:  24,
	}
}

func (h *HomeModel) UpdateSize(width, height int) {
	h.width = width
	h.height = height
}

// ASCII art for sslcerttop
const asciiArt = ` ██████╗ ██████╗ ██╗      ██████╗███████╗██████╗ ████████╗████████╗ ██████╗ ██████╗ 
██╔════╝██╔════╝██║     ██╔════╝██╔════╝██╔══██╗╚══██╔══╝╚══██╔══╝██╔═══██╗██╔══██╗
╚█████╗ ╚█████╗ ██║     ██║     █████╗  ██████╔╝   ██║      ██║   ██║   ██║██████╔╝
 ╚═══██╗ ╚═══██╗██║     ██║     ██╔══╝  ██╔══██╗   ██║      ██║   ██║   ██║██╔═══╝ 
██████╔╝██████╔╝███████╗╚██████╗███████╗██║  ██║   ██║      ██║   ╚██████╔╝██║     
╚═════╝ ╚═════╝ ╚══════╝ ╚═════╝╚══════╝╚═╝  ╚═╝   ╚═╝      ╚═╝    ╚═════╝ ╚═╝     `

const asciiArtCompact = ` ███████╗███████╗██╗      ██████╗███████╗██████╗ ████████╗████████╗ ██████╗ ██████╗ 
██╔════╝██╔════╝██║     ██╔════╝██╔════╝██╔══██╗╚══██╔══╝╚══██╔══╝██╔═══██╗██╔══██╗
███████╗███████╗██║     ██║     █████╗  ██████╔╝   ██║      ██║   ██║   ██║██████╔╝
╚════██║╚════██║██║     ██║     ██╔══╝  ██╔══██╗   ██║      ██║   ██║   ██║██╔═══╝ 
███████║███████║███████╗╚██████╗███████╗██║  ██║   ██║      ██║   ╚██████╔╝██║     
╚══════╝╚══════╝╚══════╝ ╚═════╝╚══════╝╚═╝  ╚═╝   ╚═╝      ╚═╝    ╚═════╝ ╚═╝     `

const asciiArtMinimal = `sslcerttop`

func (h HomeModel) RenderSplash() string {
	var art string

	if h.width < 84 {
		art = asciiArtMinimal
	} else if h.width < 95 {
		art = asciiArtCompact
	} else {
		art = asciiArt
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff88")).
		Bold(true).
		Width(h.width).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Width(h.width).
		Align(lipgloss.Center)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Width(h.width).
		Align(lipgloss.Center)

	var content strings.Builder

	if h.height > 20 {
		content.WriteString("\n\n")
	}

	if h.width < 84 {
		bigTitleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff88")).
			Bold(true).
			Width(h.width).
			Align(lipgloss.Center).
			Padding(1, 0)
		content.WriteString(bigTitleStyle.Render("🔒 sslcerttop 🔒"))
	} else {
		content.WriteString(titleStyle.Render(art))
	}
	content.WriteString("\n\n")

	subtitle := "🔒 SSL Certificate Monitor v1.0"
	if h.width < 84 {
		subtitle = "SSL Certificate Monitor"
	}
	content.WriteString(subtitleStyle.Render(subtitle))
	content.WriteString("\n\n")

	if h.loading {
		content.WriteString(messageStyle.Render(h.message))
		content.WriteString("\n")
		content.WriteString(messageStyle.Render("⣾⣽⣻⢿⡿⣟⣯⣟"))
	} else {
		content.WriteString(messageStyle.Render("Ready!"))
		content.WriteString("\n")
		instructionText := "Press any key to continue"
		if h.width < 84 {
			instructionText = "Press any key"
		}
		instructionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00bfff")).
			Width(h.width).
			Align(lipgloss.Center)
		content.WriteString(instructionStyle.Render(instructionText))
	}

	if h.height > 20 {
		content.WriteString("\n\n")
	}

	return lipgloss.Place(h.width, h.height, lipgloss.Center, lipgloss.Center, content.String())
}
